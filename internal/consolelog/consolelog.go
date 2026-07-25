// Package consolelog provides a human-readable slog handler for the CLI.
package consolelog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
)

// New creates a structured text logger. Debug records are enabled when
// verbose is true, and colors are used only for an interactive terminal.
func New(writer io.Writer, verbose bool) *slog.Logger {
	minimumLevel := slog.LevelInfo
	if verbose {
		minimumLevel = slog.LevelDebug
	}
	return slog.New(&handler{
		writer:        writer,
		minimumLevel:  minimumLevel,
		color:         supportsColor(writer),
		githubActions: os.Getenv("GITHUB_ACTIONS") == "true",
		mu:            &sync.Mutex{},
	})
}

type handler struct {
	writer        io.Writer
	minimumLevel  slog.Level
	color         bool
	githubActions bool
	attrs         []namedValue
	groups        []string
	mu            *sync.Mutex
}

func (handler *handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= handler.minimumLevel
}

func (handler *handler) Handle(_ context.Context, record slog.Record) error {
	attributes := append([]namedValue(nil), handler.attrs...)
	record.Attrs(func(attribute slog.Attr) bool {
		attributes = appendAttribute(attributes, handler.groups, attribute)
		return true
	})

	status := ""
	visible := attributes[:0]
	for _, attribute := range attributes {
		if attribute.name == "status" {
			status = attribute.value
			continue
		}
		visible = append(visible, attribute)
	}

	symbol, color := recordStyle(record.Level, status)
	if handler.color {
		symbol = color + symbol + colorReset
	}
	var line strings.Builder
	fmt.Fprintf(&line, "%s %s", symbol, record.Message)
	for _, attribute := range visible {
		fmt.Fprintf(&line, "  %s=%s", attribute.name, quoteValue(attribute.value))
	}
	line.WriteByte('\n')
	if handler.githubActions && record.Level >= slog.LevelError {
		details := record.Message
		for _, attribute := range visible {
			if attribute.name == "error" {
				details += ": " + attribute.value
				break
			}
		}
		fmt.Fprintf(&line, "::error title=fixyourmoney.space::%s\n", escapeWorkflowCommand(details))
	}

	handler.mu.Lock()
	defer handler.mu.Unlock()
	_, err := io.WriteString(handler.writer, line.String())
	return err
}

func (handler *handler) WithAttrs(attributes []slog.Attr) slog.Handler {
	clone := *handler
	clone.attrs = append([]namedValue(nil), handler.attrs...)
	for _, attribute := range attributes {
		clone.attrs = appendAttribute(clone.attrs, handler.groups, attribute)
	}
	return &clone
}

func (handler *handler) WithGroup(name string) slog.Handler {
	if name == "" {
		return handler
	}
	clone := *handler
	clone.groups = append(append([]string(nil), handler.groups...), name)
	return &clone
}

type namedValue struct {
	name  string
	value string
}

func appendAttribute(values []namedValue, groups []string, attribute slog.Attr) []namedValue {
	attribute.Value = attribute.Value.Resolve()
	if attribute.Equal(slog.Attr{}) {
		return values
	}
	nameParts := append(append([]string(nil), groups...), attribute.Key)
	if attribute.Value.Kind() == slog.KindGroup {
		for _, child := range attribute.Value.Group() {
			values = appendAttribute(values, nameParts, child)
		}
		return values
	}
	return append(values, namedValue{
		name:  strings.Join(nameParts, "."),
		value: valueString(attribute.Value),
	})
}

func valueString(value slog.Value) string {
	switch value.Kind() {
	case slog.KindString:
		return value.String()
	case slog.KindTime:
		return value.Time().Format("2006-01-02T15:04:05Z07:00")
	case slog.KindDuration:
		return value.Duration().String()
	case slog.KindAny:
		return fmt.Sprint(value.Any())
	default:
		return value.String()
	}
}

func quoteValue(value string) string {
	if value == "" || strings.IndexFunc(value, unicode.IsSpace) >= 0 || strings.ContainsAny(value, `="`) {
		return strconv.Quote(value)
	}
	return value
}

func recordStyle(level slog.Level, status string) (string, string) {
	if level >= slog.LevelError {
		return "✖", colorRed
	}
	if level >= slog.LevelWarn {
		return "!", colorYellow
	}
	if level <= slog.LevelDebug {
		return "·", colorGray
	}
	switch status {
	case "success":
		return "✓", colorGreen
	case "build":
		return "◆", colorBlue
	case "change", "reload":
		return "↻", colorYellow
	case "serve":
		return "➜", colorCyan
	case "hint":
		return "→", colorGray
	default:
		return "●", colorCyan
	}
}

func escapeWorkflowCommand(value string) string {
	value = strings.ReplaceAll(value, "%", "%25")
	value = strings.ReplaceAll(value, "\r", "%0D")
	return strings.ReplaceAll(value, "\n", "%0A")
}

func supportsColor(writer io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	file, ok := writer.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
