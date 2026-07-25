package generator

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestBuildRendersLocalesRoutesAndHashedAssets(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "site.json", `{"title":"Test","baseURL":"https://example.test","defaultLocale":"de","locales":["de","en"]}`)
	writeTestFile(t, root, "i18n/de.json", `{"greeting":"Hallo"}`)
	writeTestFile(t, root, "i18n/en.json", `{"greeting":"Hello"}`)
	writeTestFile(t, root, "static/css/site.css", "body { color: green; }")
	writeTestFile(t, root, "pages/index.gohtml", `<h1>{{ t "greeting" }}</h1><link href="{{ asset "css/site.css" }}"><a href="{{ url "about" }}">About</a>`)
	writeTestFile(t, root, "pages/index.md", "# {{ t \"greeting\" }}\n{{ range .Locales }}[{{ languageName . }}]({{ languageMarkdownURL . }}) {{ end }}")
	writeTestFile(t, root, "pages/llms.txt", "LLMS: {{ t \"greeting\" }}")
	writeTestFile(t, root, "pages/about.gohtml", `<html lang="{{ .Locale }}"><a href="{{ languageURL "de" }}">Deutsch</a></html>`)

	output := filepath.Join(root, "dist")
	if err := Build(root, output); err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	css := []byte("body { color: green; }")
	hash := releaseHashForTest(map[string][]byte{"css/site.css": css})
	assertFileContains(t, filepath.Join(output, "de/index.html"), "<h1>Hallo</h1>")
	assertFileContains(t, filepath.Join(output, "en/index.html"), "<h1>Hello</h1>")
	assertFileContains(t, filepath.Join(output, "en/index.html"), `/static/`+hash+`/css/site.css`)
	assertFileContains(t, filepath.Join(output, "de/index.html"), `href="/de/about/"`)
	assertFileContains(t, filepath.Join(output, "en/about/index.html"), `href="/de/about/"`)
	assertFileContains(t, filepath.Join(output, "de/index.md"), "# Hallo")
	assertFileContains(t, filepath.Join(output, "en/index.md"), `[Deutsch](https://example.test/de/index.md)`)
	assertFileContains(t, filepath.Join(output, "index.md"), "# Hallo")
	assertFileContains(t, filepath.Join(output, "llms.txt"), "LLMS: Hallo")
	assertFileContains(t, filepath.Join(output, "index.html"), `url=/de/`)
	assertFileContains(t, filepath.Join(output, "index.html"), `var supported = ["de","en"];`)
	assertFileContains(t, filepath.Join(output, "index.html"), `var fallback = "de";`)
	assertFileContains(t, filepath.Join(output, "index.html"), `navigator.languages`)
	assertFileContains(t, filepath.Join(output, "sitemap.xml"), `<loc>https://example.test/en/about/</loc>`)
	assertFileContains(t, filepath.Join(output, "sitemap.xml"), `hreflang="x-default" href="https://example.test/de/about/"`)
	assertFileContains(t, filepath.Join(output, "robots.txt"), `Sitemap: https://example.test/sitemap.xml`)
	if contents, err := os.ReadFile(filepath.Join(output, "static", hash, "css/site.css")); err != nil {
		t.Fatalf("read copied asset: %v", err)
	} else if string(contents) != string(css) {
		t.Errorf("copied asset = %q, want %q", contents, css)
	}
}

func TestBuildWithLoggerReportsProgressAndSummary(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "site.json", `{"defaultLocale":"en","locales":["en"]}`)
	writeTestFile(t, root, "i18n/en.json", `{}`)
	writeTestFile(t, root, "static/site.css", "body {}")
	writeTestFile(t, root, "pages/index.gohtml", `<p>Test</p>`)

	var records bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&records, &slog.HandlerOptions{Level: slog.LevelDebug}))
	if err := BuildWithLogger(root, filepath.Join(root, "dist"), logger); err != nil {
		t.Fatalf("BuildWithLogger() error = %v", err)
	}
	output := records.String()
	for _, expected := range []string{
		`msg="Building site"`,
		`msg="Loaded site configuration"`,
		`msg="Discovered templates"`,
		`msg="Site built"`,
		"pages=1",
		"assets=1",
		"locales=1",
	} {
		if !strings.Contains(output, expected) {
			t.Errorf("build log does not contain %q: %s", expected, output)
		}
	}
}

func TestBuildFailsWhenLocaleKeysDiffer(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "site.json", `{"defaultLocale":"de","locales":["de","en"]}`)
	writeTestFile(t, root, "i18n/de.json", `{"only.default":"Standardwert"}`)
	writeTestFile(t, root, "i18n/en.json", `{}`)
	writeTestFile(t, root, "pages/index.gohtml", `{{ t "only.default" }}`)

	err := Build(root, filepath.Join(root, "dist"))
	if err == nil || !strings.Contains(err.Error(), "translations for en do not match de keys") {
		t.Fatalf("Build() error = %v, want locale key mismatch", err)
	}
}

func TestBuildUsesOneReleaseHashForRelativeAssetDependencies(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "site.json", `{"defaultLocale":"en","locales":["en"]}`)
	writeTestFile(t, root, "i18n/en.json", `{}`)
	writeTestFile(t, root, "pages/index.gohtml", `<link href="{{ asset "css/site.css" }}"><script src="{{ asset "js/site.js" }}"></script>`)
	assets := map[string][]byte{
		"css/site.css":     []byte(`@import "./tokens.css";`),
		"css/tokens.css":   []byte(`@font-face { src: url("../fonts/site.woff2"); }`),
		"fonts/site.woff2": []byte("font"),
		"js/site.js":       []byte(`import "./module.js";`),
		"js/module.js":     []byte(`export const ready = true;`),
	}
	for name, contents := range assets {
		writeTestFile(t, root, "static/"+name, string(contents))
	}

	output := filepath.Join(root, "dist")
	if err := Build(root, output); err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	hash := releaseHashForTest(assets)
	for name := range assets {
		if _, err := os.Stat(filepath.Join(output, "static", hash, filepath.FromSlash(name))); err != nil {
			t.Errorf("asset %s does not use release hash %s: %v", name, hash, err)
		}
	}
}

func TestBuildRendersMetadataAndRTLHelpers(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "site.json", `{"baseURL":"https://example.test","defaultLocale":"en","locales":["en","ar"]}`)
	writeTestFile(t, root, "i18n/en.json", `{}`)
	writeTestFile(t, root, "i18n/ar.json", `{}`)
	writeTestFile(t, root, "pages/index.gohtml", `<html dir="{{ textDirection }}"><link rel="canonical" href="{{ canonicalURL }}"><link hreflang="en" href="{{ languageCanonicalURL "en" }}"><meta property="og:locale" content="{{ openGraphLocale .Locale }}"><span>{{ siteCanonicalURL }}</span>{{ languageName .Locale }}</html>`)

	if err := Build(root, filepath.Join(root, "dist")); err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	assertFileContains(t, filepath.Join(root, "dist/ar/index.html"), `dir="rtl"`)
	assertFileContains(t, filepath.Join(root, "dist/ar/index.html"), `href="https://example.test/ar/"`)
	assertFileContains(t, filepath.Join(root, "dist/ar/index.html"), `content="ar_AR"`)
	assertFileContains(t, filepath.Join(root, "dist/ar/index.html"), `<span>https://example.test/ar/</span>`)
	assertFileContains(t, filepath.Join(root, "dist/ar/index.html"), "العربية")
}

func TestBuildFailsForUnknownAsset(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "site.json", `{"defaultLocale":"en","locales":["en"]}`)
	writeTestFile(t, root, "i18n/en.json", `{}`)
	writeTestFile(t, root, "pages/index.gohtml", `{{ asset "missing.css" }}`)

	err := Build(root, filepath.Join(root, "dist"))
	if err == nil || !strings.Contains(err.Error(), `unknown static asset "missing.css"`) {
		t.Fatalf("Build() error = %v, want unknown asset error", err)
	}
}

func TestBuildRejectsStaticAssetSymlinks(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "site.json", `{"defaultLocale":"en","locales":["en"]}`)
	writeTestFile(t, root, "i18n/en.json", `{}`)
	writeTestFile(t, root, "pages/index.gohtml", `ok`)
	writeTestFile(t, root, "outside.txt", `outside`)
	if err := os.MkdirAll(filepath.Join(root, "static"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "outside.txt"), filepath.Join(root, "static", "linked.txt")); err != nil {
		t.Fatal(err)
	}

	err := Build(root, filepath.Join(root, "dist"))
	if err == nil || !strings.Contains(err.Error(), "static asset symlinks are not supported") {
		t.Fatalf("Build() error = %v, want symlink error", err)
	}
}

func TestBuildFailsForUnknownTranslation(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "site.json", `{"defaultLocale":"de","locales":["de"]}`)
	writeTestFile(t, root, "i18n/de.json", `{}`)
	writeTestFile(t, root, "pages/index.gohtml", `{{ t "missing" }}`)

	err := Build(root, filepath.Join(root, "dist"))
	if err == nil || !strings.Contains(err.Error(), `translation "missing" is missing`) {
		t.Fatalf("Build() error = %v, want missing translation error", err)
	}
}

func writeTestFile(t *testing.T, root, name, contents string) {
	t.Helper()
	filename := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filename, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func releaseHashForTest(assets map[string][]byte) string {
	type asset struct {
		key      string
		contents []byte
	}
	ordered := make([]asset, 0, len(assets))
	for key, contents := range assets {
		ordered = append(ordered, asset{key: key, contents: contents})
	}
	slices.SortFunc(ordered, func(a, b asset) int { return strings.Compare(a.key, b.key) })
	digest := sha256.New()
	for _, item := range ordered {
		fmt.Fprintf(digest, "%d:%s:%d:", len(item.key), item.key, len(item.contents))
		_, _ = digest.Write(item.contents)
	}
	return hex.EncodeToString(digest.Sum(nil))[:12]
}

func assertFileContains(t *testing.T, filename, expected string) {
	t.Helper()
	contents, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("read %s: %v", filename, err)
	}
	if !strings.Contains(string(contents), expected) {
		t.Errorf("%s does not contain %q; contents: %q", filename, expected, contents)
	}
}
