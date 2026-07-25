// Package devserver rebuilds and serves the site during local development.
package devserver

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"log/slog"
	"maps"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/nilsleifeld/fixyourmoney.space/internal/generator"
)

const (
	reloadEndpoint = "/__fixyourmoney_reload"
	watchInterval  = 150 * time.Millisecond
	watchDebounce  = 150 * time.Millisecond
)

var liveReloadTag = []byte(`<script>
const fixYourMoneyReloadSource = new EventSource("` + reloadEndpoint + `");
/**
 * Reloads the current page after a successful development build.
 * @returns {void}
 */
function reloadFixYourMoneyPage() {
  window.location.reload();
}
fixYourMoneyReloadSource.addEventListener("reload", reloadFixYourMoneyPage);
</script>`)

type fileState struct {
	size       int64
	modifiedAt int64
	mode       fs.FileMode
	digest     [sha256.Size]byte
}

type sourceState map[string]fileState

type sourceChange struct {
	path string
	kind string
}

type reloadHub struct {
	mu      sync.Mutex
	clients map[chan struct{}]struct{}
}

// Serve builds the site, watches its source files, and reloads connected
// browsers after every successful rebuild.
func Serve(source, output, address string) error {
	logger := slog.Default()
	root, err := filepath.Abs(source)
	if err != nil {
		return fmt.Errorf("resolve source directory: %w", err)
	}
	directory, err := filepath.Abs(output)
	if err != nil {
		return fmt.Errorf("resolve output directory: %w", err)
	}

	// Snapshot before building so an edit made during the initial build is not lost.
	initialState, err := snapshotSource(root, directory)
	if err != nil {
		return fmt.Errorf("inspect source files: %w", err)
	}
	if err := generator.BuildWithLogger(source, output, logger); err != nil {
		return err
	}

	hub := newReloadHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go watchSource(ctx, root, directory, initialState, func(changes []sourceChange) {
		logger.Info("Source changed", "status", "change", "changes", len(changes), "files", changeSummary(changes))
		for _, change := range changes {
			logger.Debug("Source file changed", "change", change.kind, "path", change.path)
		}
		if err := generator.BuildWithLogger(source, output, logger); err != nil {
			logger.Error("Rebuild failed; keeping previous output", "error", err, "output", directory)
			return
		}
		clients := hub.reload()
		if clients == 0 {
			logger.Debug("Rebuild complete; no browser is connected")
			return
		}
		logger.Info("Reloading connected browsers", "status", "reload", "browsers", clients)
	})

	server := &http.Server{
		Addr:    address,
		Handler: newDevelopmentHandler(directory, hub),
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", address, err)
	}
	defer listener.Close()
	logger.Info("Development server ready", "status", "serve", "url", developmentURL(address), "output", directory)
	logger.Info("Watching for changes", "status", "hint", "sources", "site.json, pages/, templates/, i18n/, static/", "stop", "Ctrl+C")
	return server.Serve(listener)
}

func watchSource(ctx context.Context, root, output string, previous sourceState, rebuild func([]sourceChange)) {
	ticker := time.NewTicker(watchInterval)
	defer ticker.Stop()

	var changedAt time.Time
	var batchStart sourceState
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			current, err := snapshotSource(root, output)
			if err != nil {
				slog.Warn("Unable to inspect watched source files", "error", err)
				continue
			}
			if !maps.Equal(previous, current) {
				if batchStart == nil {
					batchStart = previous
				}
				previous = current
				changedAt = now
				continue
			}
			if !changedAt.IsZero() && now.Sub(changedAt) >= watchDebounce {
				changes := sourceChanges(batchStart, current)
				changedAt = time.Time{}
				batchStart = nil
				if len(changes) > 0 {
					rebuild(changes)
				}
			}
		}
	}
}

func snapshotSource(root, output string) (sourceState, error) {
	state := make(sourceState)
	if err := addFileState(state, root, filepath.Join(root, "site.json")); err != nil {
		return nil, err
	}
	for _, filename := range []string{"index.md", "llms.txt"} {
		if err := addFileState(state, root, filepath.Join(root, "pages", filename)); err != nil {
			return nil, err
		}
	}

	watchedDirectories := []struct {
		name      string
		extension string
	}{
		{name: "pages", extension: ".gohtml"},
		{name: "templates", extension: ".gohtml"},
		{name: "i18n", extension: ".json"},
		{name: "static"},
	}
	for _, watched := range watchedDirectories {
		directory := filepath.Join(root, watched.name)
		err := filepath.WalkDir(directory, func(filename string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if insideDirectory(filename, output) {
				if entry.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if entry.IsDir() {
				return nil
			}
			if watched.extension != "" && !strings.EqualFold(filepath.Ext(filename), watched.extension) {
				return nil
			}
			return addFileState(state, root, filename)
		})
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}
	return state, nil
}

func addFileState(state sourceState, root, filename string) error {
	info, err := os.Lstat(filename)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	relative, err := filepath.Rel(root, filename)
	if err != nil {
		return err
	}
	var contents []byte
	if info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		contents, err = os.ReadFile(filename)
		if err != nil && info.Mode()&os.ModeSymlink != 0 {
			// A broken link must still enter the snapshot so the build can report it.
			linkTarget, linkErr := os.Readlink(filename)
			if linkErr != nil {
				return err
			}
			contents = []byte(linkTarget)
		} else if err != nil {
			return err
		}
	}
	state[relative] = fileState{
		size:       info.Size(),
		modifiedAt: info.ModTime().UnixNano(),
		mode:       info.Mode(),
		digest:     sha256.Sum256(contents),
	}
	return nil
}

func insideDirectory(filename, directory string) bool {
	relative, err := filepath.Rel(directory, filename)
	if err != nil {
		return false
	}
	return relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)))
}

func sourceChanges(previous, current sourceState) []sourceChange {
	changes := make([]sourceChange, 0)
	for filename, previousState := range previous {
		currentState, exists := current[filename]
		switch {
		case !exists:
			changes = append(changes, sourceChange{path: filepath.ToSlash(filename), kind: "removed"})
		case previousState != currentState:
			changes = append(changes, sourceChange{path: filepath.ToSlash(filename), kind: "modified"})
		}
	}
	for filename := range current {
		if _, exists := previous[filename]; !exists {
			changes = append(changes, sourceChange{path: filepath.ToSlash(filename), kind: "added"})
		}
	}
	slices.SortFunc(changes, func(left, right sourceChange) int {
		return strings.Compare(left.path, right.path)
	})
	return changes
}

func changeSummary(changes []sourceChange) string {
	parts := make([]string, 0, len(changes))
	for _, change := range changes {
		marker := "~"
		switch change.kind {
		case "added":
			marker = "+"
		case "removed":
			marker = "-"
		}
		parts = append(parts, marker+" "+change.path)
	}
	return strings.Join(parts, ", ")
}

func newReloadHub() *reloadHub {
	return &reloadHub{clients: make(map[chan struct{}]struct{})}
}

func (hub *reloadHub) subscribe() (<-chan struct{}, func()) {
	updates := make(chan struct{}, 1)
	hub.mu.Lock()
	hub.clients[updates] = struct{}{}
	hub.mu.Unlock()
	return updates, func() {
		hub.mu.Lock()
		delete(hub.clients, updates)
		hub.mu.Unlock()
	}
}

func (hub *reloadHub) reload() int {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	for client := range hub.clients {
		select {
		case client <- struct{}{}:
		default:
		}
	}
	return len(hub.clients)
}

func (hub *reloadHub) serveEvents(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		response.Header().Set("Allow", http.MethodGet)
		http.Error(response, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	flusher, ok := response.(http.Flusher)
	if !ok {
		http.Error(response, "streaming is not supported", http.StatusInternalServerError)
		return
	}

	updates, unsubscribe := hub.subscribe()
	defer unsubscribe()
	response.Header().Set("Content-Type", "text/event-stream")
	response.Header().Set("Cache-Control", "no-store")
	response.Header().Set("Connection", "keep-alive")
	_, _ = fmt.Fprint(response, ": connected\n\n")
	flusher.Flush()

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()
	for {
		select {
		case <-request.Context().Done():
			return
		case <-updates:
			_, _ = fmt.Fprint(response, "event: reload\ndata: reload\n\n")
			flusher.Flush()
		case <-heartbeat.C:
			_, _ = fmt.Fprint(response, ": heartbeat\n\n")
			flusher.Flush()
		}
	}
}

type developmentHandler struct {
	files http.Handler
	hub   *reloadHub
}

func newDevelopmentHandler(directory string, hub *reloadHub) http.Handler {
	return &developmentHandler{
		files: http.FileServer(http.Dir(directory)),
		hub:   hub,
	}
}

func (handler *developmentHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	if request.URL.Path == reloadEndpoint {
		handler.hub.serveEvents(response, request)
		return
	}

	// Ignore validators so rapid rebuilds within one timestamp second cannot
	// leave the browser on a stale generated page.
	uncachedRequest := request.Clone(request.Context())
	uncachedRequest.Header = request.Header.Clone()
	uncachedRequest.Header.Del("If-Modified-Since")
	uncachedRequest.Header.Del("If-None-Match")

	buffered := newBufferedResponse()
	handler.files.ServeHTTP(buffered, uncachedRequest)
	body := buffered.body.Bytes()
	if request.Method == http.MethodGet && buffered.status == http.StatusOK && strings.HasPrefix(buffered.header.Get("Content-Type"), "text/html") {
		body = injectLiveReload(body)
		buffered.header.Set("Content-Length", fmt.Sprintf("%d", len(body)))
	}
	buffered.header.Set("Cache-Control", "no-store")
	buffered.header.Del("Last-Modified")
	buffered.header.Del("ETag")
	copyHeaders(response.Header(), buffered.header)
	response.WriteHeader(buffered.status)
	if request.Method != http.MethodHead {
		_, _ = response.Write(body)
	}
}

type bufferedResponse struct {
	header http.Header
	body   bytes.Buffer
	status int
}

func newBufferedResponse() *bufferedResponse {
	return &bufferedResponse{header: make(http.Header), status: http.StatusOK}
}

func (response *bufferedResponse) Header() http.Header {
	return response.header
}

func (response *bufferedResponse) WriteHeader(status int) {
	response.status = status
}

func (response *bufferedResponse) Write(contents []byte) (int, error) {
	return response.body.Write(contents)
}

func injectLiveReload(document []byte) []byte {
	lowerDocument := bytes.ToLower(document)
	position := bytes.LastIndex(lowerDocument, []byte("</body>"))
	if position < 0 {
		position = bytes.LastIndex(lowerDocument, []byte("</html>"))
	}
	if position < 0 {
		position = len(document)
	}

	injected := make([]byte, 0, len(document)+len(liveReloadTag)+1)
	injected = append(injected, document[:position]...)
	injected = append(injected, liveReloadTag...)
	injected = append(injected, '\n')
	injected = append(injected, document[position:]...)
	return injected
}

func copyHeaders(target, source http.Header) {
	for name, values := range source {
		for _, value := range values {
			target.Add(name, value)
		}
	}
}

func developmentURL(address string) string {
	if strings.HasPrefix(address, ":") {
		return "http://localhost" + address
	}
	return "http://" + address
}
