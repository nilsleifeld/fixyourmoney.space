package devserver

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDevelopmentHandlerInjectsReloadClientIntoHTML(t *testing.T) {
	output := t.TempDir()
	writeFile(t, output, "en/index.html", "<!doctype html><html><body><h1>Test</h1></body></html>")

	request := httptest.NewRequest(http.MethodGet, "http://example.test/en/", nil)
	response := httptest.NewRecorder()
	newDevelopmentHandler(output, newReloadHub()).ServeHTTP(response, request)

	result := response.Result()
	body, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatal(err)
	}
	contents := string(body)
	if result.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", result.StatusCode, http.StatusOK)
	}
	if !strings.Contains(contents, `new EventSource("`+reloadEndpoint+`")`) {
		t.Errorf("HTML does not contain reload client: %s", contents)
	}
	if strings.Index(contents, "new EventSource") > strings.Index(contents, "</body>") {
		t.Errorf("reload client occurs after closing body: %s", contents)
	}
	if cacheControl := result.Header.Get("Cache-Control"); cacheControl != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", cacheControl)
	}
	if lastModified := result.Header.Get("Last-Modified"); lastModified != "" {
		t.Errorf("Last-Modified = %q, want empty", lastModified)
	}
}

func TestDevelopmentHandlerDoesNotModifyAssets(t *testing.T) {
	output := t.TempDir()
	writeFile(t, output, "site.css", "body { color: black; }")

	request := httptest.NewRequest(http.MethodGet, "http://example.test/site.css", nil)
	response := httptest.NewRecorder()
	newDevelopmentHandler(output, newReloadHub()).ServeHTTP(response, request)

	if body := response.Body.String(); body != "body { color: black; }" {
		t.Errorf("asset body = %q", body)
	}
}

func TestReloadHubNotifiesSubscribers(t *testing.T) {
	hub := newReloadHub()
	updates, unsubscribe := hub.subscribe()
	defer unsubscribe()

	hub.reload()
	select {
	case <-updates:
	case <-time.After(time.Second):
		t.Fatal("subscriber did not receive reload notification")
	}
}

func TestSnapshotSourceDetectsContentChangeWithStableMetadata(t *testing.T) {
	root := t.TempDir()
	output := filepath.Join(root, "dist")
	writeFile(t, root, "site.json", `{}`)
	writeFile(t, root, "pages/index.gohtml", "before")
	filename := filepath.Join(root, "pages", "index.gohtml")
	info, err := os.Stat(filename)
	if err != nil {
		t.Fatal(err)
	}
	before, err := snapshotSource(root, output)
	if err != nil {
		t.Fatal(err)
	}

	writeFile(t, root, "pages/index.gohtml", "differ")
	if err := os.Chtimes(filename, info.ModTime(), info.ModTime()); err != nil {
		t.Fatal(err)
	}
	after, err := snapshotSource(root, output)
	if err != nil {
		t.Fatal(err)
	}
	if before[filepath.Join("pages", "index.gohtml")] == after[filepath.Join("pages", "index.gohtml")] {
		t.Fatal("snapshot did not detect content change with unchanged size and timestamp")
	}
}

func TestWatchSourceRebuildsAfterRelevantChange(t *testing.T) {
	root := t.TempDir()
	output := filepath.Join(root, "dist")
	writeFile(t, root, "site.json", `{}`)
	writeFile(t, root, "pages/index.gohtml", "before")
	initial, err := snapshotSource(root, output)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	rebuilt := make(chan struct{}, 1)
	go watchSource(ctx, root, output, initial, func(changes []sourceChange) {
		if len(changes) != 1 || changes[0].path != "pages/index.gohtml" || changes[0].kind != "modified" {
			t.Errorf("changes = %#v", changes)
		}
		rebuilt <- struct{}{}
	})
	writeFile(t, root, "pages/index.gohtml", "after with a different size")

	select {
	case <-rebuilt:
	case <-time.After(2 * time.Second):
		t.Fatal("watcher did not request a rebuild")
	}
}

func TestSourceChangesReportsSortedAddedModifiedAndRemovedFiles(t *testing.T) {
	state := func(value byte) fileState {
		return fileState{size: int64(value)}
	}
	previous := sourceState{
		"pages/removed.gohtml": state(1),
		"pages/changed.gohtml": state(1),
	}
	current := sourceState{
		"pages/added.gohtml":   state(1),
		"pages/changed.gohtml": state(2),
	}

	changes := sourceChanges(previous, current)
	if summary := changeSummary(changes); summary != "+ pages/added.gohtml, ~ pages/changed.gohtml, - pages/removed.gohtml" {
		t.Errorf("summary = %q", summary)
	}
}

func TestSnapshotSourceIgnoresOutputAndUnrelatedPageFiles(t *testing.T) {
	root := t.TempDir()
	output := filepath.Join(root, "pages", "generated")
	writeFile(t, root, "site.json", `{}`)
	writeFile(t, root, "pages/index.gohtml", "page")
	writeFile(t, root, "pages/index.md", "agent page")
	writeFile(t, root, "pages/llms.txt", "LLM instructions")
	before, err := snapshotSource(root, output)
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := before[filepath.Join("pages", "index.md")]; !exists {
		t.Fatal("snapshot does not include Markdown index")
	}
	if _, exists := before[filepath.Join("pages", "llms.txt")]; !exists {
		t.Fatal("snapshot does not include llms.txt template")
	}

	writeFile(t, root, "pages/notes.txt", "not a page template")
	writeFile(t, output, "index.gohtml", "generated output")
	after, err := snapshotSource(root, output)
	if err != nil {
		t.Fatal(err)
	}
	if len(before) != len(after) {
		t.Fatalf("snapshot size changed from %d to %d", len(before), len(after))
	}
	for filename, state := range before {
		if after[filename] != state {
			t.Errorf("state for %s changed", filename)
		}
	}
}

func writeFile(t *testing.T, root, name, contents string) {
	t.Helper()
	filename := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filename, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}
