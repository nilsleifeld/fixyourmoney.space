package consolelog

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestLoggerWritesReadableStructuredText(t *testing.T) {
	var output bytes.Buffer
	logger := New(&output, false)

	logger.Info("Site built",
		"status", "success",
		"duration", 125*time.Millisecond,
		"pages", 28,
		"output", "dist folder",
	)

	line := output.String()
	if !strings.HasPrefix(line, "✓ Site built") {
		t.Errorf("output = %q, want success prefix", line)
	}
	for _, expected := range []string{"duration=125ms", "pages=28", `output="dist folder"`} {
		if !strings.Contains(line, expected) {
			t.Errorf("output = %q, want %q", line, expected)
		}
	}
	if strings.Contains(line, `"msg"`) || strings.Contains(line, `"status"`) {
		t.Errorf("output looks like JSON or exposes internal status: %q", line)
	}
}

func TestLoggerAddsGitHubActionsErrorAnnotation(t *testing.T) {
	t.Setenv("GITHUB_ACTIONS", "true")
	var output bytes.Buffer
	logger := New(&output, false)

	logger.Error("Build failed", "error", "bad template\nline 2")

	if annotation := "::error title=fixyourmoney.space::Build failed: bad template%0Aline 2"; !strings.Contains(output.String(), annotation) {
		t.Errorf("output = %q, want annotation %q", output.String(), annotation)
	}
}

func TestLoggerEnablesDebugOnlyInVerboseMode(t *testing.T) {
	var normalOutput bytes.Buffer
	New(&normalOutput, false).Debug("details")
	if normalOutput.Len() != 0 {
		t.Errorf("normal output = %q, want no debug record", normalOutput.String())
	}

	var verboseOutput bytes.Buffer
	New(&verboseOutput, true).Debug("details", slog.String("phase", "templates"))
	if output := verboseOutput.String(); !strings.Contains(output, "· details  phase=templates") {
		t.Errorf("verbose output = %q", output)
	}
}
