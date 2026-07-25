package generator

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	texttemplate "text/template"
	"time"
)

const configFile = "site.json"

// Config describes the site and its available languages.
type Config struct {
	Title          string   `json:"title"`
	BaseURL        string   `json:"baseURL"`
	TwitterCreator string   `json:"twitterCreator"`
	DefaultLocale  string   `json:"defaultLocale"`
	Locales        []string `json:"locales"`
}

// Page contains information about the page currently being rendered.
type Page struct {
	Route string
}

// Data is available as the dot value in every page template.
type Data struct {
	Site    Config
	Page    Page
	Locale  string
	Locales []string
}

type buildContext struct {
	root         string
	output       string
	config       Config
	translations map[string]map[string]string
	assets       map[string]string
	shared       []string
}

// Build renders the site below source into output without emitting progress
// records. Existing output is replaced only after a complete successful build.
func Build(source, output string) error {
	return BuildWithLogger(source, output, slog.New(slog.DiscardHandler))
}

// BuildWithLogger renders the site and reports structured build progress to
// logger. Existing output is replaced only after a complete successful build.
func BuildWithLogger(source, output string, logger *slog.Logger) error {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	startedAt := time.Now()
	logger.Info("Building site", "status", "build", "source", source, "output", output)

	root, err := filepath.Abs(source)
	if err != nil {
		return fmt.Errorf("resolve source directory: %w", err)
	}
	dist, err := filepath.Abs(output)
	if err != nil {
		return fmt.Errorf("resolve output directory: %w", err)
	}
	if root == dist {
		return errors.New("source and output directory must be different")
	}
	logger.Debug("Resolved build directories", "source", root, "output", dist)

	cfg, err := loadConfig(filepath.Join(root, configFile))
	if err != nil {
		return err
	}
	logger.Debug("Loaded site configuration", "default_locale", cfg.DefaultLocale, "locales", len(cfg.Locales))
	translations, err := loadTranslations(root, cfg)
	if err != nil {
		return err
	}
	logger.Debug("Validated translations", "locales", len(translations), "keys", len(translations[cfg.DefaultLocale]))
	pages, err := filesWithExtension(filepath.Join(root, "pages"), ".gohtml", true)
	if err != nil {
		return err
	}
	if len(pages) == 0 {
		return errors.New("no pages found in pages/")
	}
	shared, err := filesWithExtension(filepath.Join(root, "templates"), ".gohtml", false)
	if err != nil {
		return err
	}
	logger.Debug("Discovered templates", "pages", len(pages), "shared_templates", len(shared))

	parent := filepath.Dir(dist)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return fmt.Errorf("create output parent: %w", err)
	}
	tmp, err := os.MkdirTemp(parent, ".site-build-*")
	if err != nil {
		return fmt.Errorf("create temporary output: %w", err)
	}
	defer os.RemoveAll(tmp)

	ctx := &buildContext{
		root:         root,
		output:       tmp,
		config:       cfg,
		translations: translations,
		shared:       shared,
	}
	ctx.assets, err = copyAssets(filepath.Join(root, "static"), tmp)
	if err != nil {
		return err
	}
	logger.Debug("Copied static release", "assets", len(ctx.assets))
	if err := ctx.renderPages(pages); err != nil {
		return err
	}
	localizedPages := len(pages) * len(cfg.Locales)
	logger.Debug("Rendered localized pages", "pages", localizedPages)
	markdownTemplate := filepath.Join(root, "pages", "index.md")
	if err := ctx.renderMarkdownIndex(markdownTemplate); err != nil {
		return err
	}
	llmsTemplate := filepath.Join(root, "pages", "llms.txt")
	if err := ctx.renderLLMSText(llmsTemplate); err != nil {
		return err
	}
	if err := writeSitemap(tmp, cfg, pages, root); err != nil {
		return err
	}
	if err := writeRobots(tmp, cfg.BaseURL); err != nil {
		return err
	}
	if err := writeRedirect(tmp, cfg.DefaultLocale, cfg.Locales); err != nil {
		return err
	}

	logger.Debug("Publishing generated output", "output", dist)
	if err := os.RemoveAll(dist); err != nil {
		return fmt.Errorf("remove old output: %w", err)
	}
	if err := os.Rename(tmp, dist); err != nil {
		return fmt.Errorf("publish output: %w", err)
	}
	logger.Info("Site built",
		"status", "success",
		"duration", time.Since(startedAt).Round(time.Millisecond),
		"pages", localizedPages,
		"assets", len(ctx.assets),
		"locales", len(cfg.Locales),
		"output", output,
	)
	return nil
}

func loadConfig(filename string) (Config, error) {
	var cfg Config
	contents, err := os.ReadFile(filename)
	if err != nil {
		return cfg, fmt.Errorf("read %s: %w", configFile, err)
	}
	if err := json.Unmarshal(contents, &cfg); err != nil {
		return cfg, fmt.Errorf("parse %s: %w", configFile, err)
	}
	if cfg.DefaultLocale == "" {
		return cfg, errors.New("site.json: defaultLocale is required")
	}
	if len(cfg.Locales) == 0 {
		return cfg, errors.New("site.json: at least one locale is required")
	}
	seen := make(map[string]bool)
	for _, locale := range cfg.Locales {
		if !validSegment(locale) {
			return cfg, fmt.Errorf("site.json: invalid locale %q", locale)
		}
		if seen[locale] {
			return cfg, fmt.Errorf("site.json: duplicate locale %q", locale)
		}
		seen[locale] = true
	}
	if !seen[cfg.DefaultLocale] {
		return cfg, errors.New("site.json: defaultLocale must occur in locales")
	}
	return cfg, nil
}

func validSegment(value string) bool {
	if value == "" || value == "." || value == ".." {
		return false
	}
	for _, char := range value {
		if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') && (char < '0' || char > '9') && char != '-' && char != '_' {
			return false
		}
	}
	return true
}

func loadTranslations(root string, cfg Config) (map[string]map[string]string, error) {
	all := make(map[string]map[string]string, len(cfg.Locales))
	for _, locale := range cfg.Locales {
		filename := filepath.Join(root, "i18n", locale+".json")
		contents, err := os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("read translations for %s: %w", locale, err)
		}
		values := make(map[string]string)
		if err := json.Unmarshal(contents, &values); err != nil {
			return nil, fmt.Errorf("parse translations for %s: %w", locale, err)
		}
		all[locale] = values
	}

	defaultValues := all[cfg.DefaultLocale]
	for _, locale := range cfg.Locales {
		if locale == cfg.DefaultLocale {
			continue
		}
		if err := compareTranslationKeys(defaultValues, all[locale], cfg.DefaultLocale, locale); err != nil {
			return nil, err
		}
	}
	return all, nil
}

func compareTranslationKeys(reference, candidate map[string]string, referenceLocale, candidateLocale string) error {
	var missing []string
	var extra []string
	for key := range reference {
		if _, exists := candidate[key]; !exists {
			missing = append(missing, key)
		}
	}
	for key := range candidate {
		if _, exists := reference[key]; !exists {
			extra = append(extra, key)
		}
	}
	if len(missing) == 0 && len(extra) == 0 {
		return nil
	}
	slices.Sort(missing)
	slices.Sort(extra)
	return fmt.Errorf("translations for %s do not match %s keys (missing: %v; extra: %v)", candidateLocale, referenceLocale, missing, extra)
}

func filesWithExtension(directory, extension string, required bool) ([]string, error) {
	var files []string
	err := filepath.WalkDir(directory, func(filename string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() && strings.EqualFold(filepath.Ext(filename), extension) {
			files = append(files, filename)
		}
		return nil
	})
	if err != nil {
		if !required && errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("scan %s: %w", filepath.Base(directory), err)
	}
	slices.Sort(files)
	return files, nil
}

func copyAssets(staticDir, output string) (map[string]string, error) {
	manifest := make(map[string]string)
	type staticAsset struct {
		key      string
		contents []byte
	}
	var assets []staticAsset
	err := filepath.WalkDir(staticDir, func(filename string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("static asset symlinks are not supported: %s", filename)
		}
		if entry.IsDir() {
			return nil
		}
		contents, err := os.ReadFile(filename)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(staticDir, filename)
		if err != nil {
			return err
		}
		assets = append(assets, staticAsset{key: filepath.ToSlash(relative), contents: contents})
		return nil
	})
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return manifest, nil
		}
		return nil, fmt.Errorf("copy static assets: %w", err)
	}

	slices.SortFunc(assets, func(a, b staticAsset) int { return strings.Compare(a.key, b.key) })
	releaseDigest := sha256.New()
	for _, asset := range assets {
		fmt.Fprintf(releaseDigest, "%d:%s:%d:", len(asset.key), asset.key, len(asset.contents))
		_, _ = releaseDigest.Write(asset.contents)
	}
	releaseHash := hex.EncodeToString(releaseDigest.Sum(nil))[:12]

	for _, asset := range assets {
		publicPath := path.Join("static", releaseHash, asset.key)
		target := filepath.Join(output, filepath.FromSlash(publicPath))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(target, asset.contents, 0o644); err != nil {
			return nil, err
		}
		manifest[asset.key] = "/" + publicPath
	}
	return manifest, nil
}

func (ctx *buildContext) renderPages(pages []string) error {
	outputs := make(map[string]string)
	for _, pageFile := range pages {
		route, err := pageRoute(filepath.Join(ctx.root, "pages"), pageFile)
		if err != nil {
			return err
		}
		for _, locale := range ctx.config.Locales {
			target := filepath.Join(ctx.output, locale, filepath.FromSlash(route), "index.html")
			if previous, exists := outputs[target]; exists {
				return fmt.Errorf("pages %s and %s have the same output path", previous, pageFile)
			}
			outputs[target] = pageFile
			if err := ctx.renderPage(pageFile, target, route, locale); err != nil {
				return err
			}
		}
	}
	return nil
}

func pageRoute(pagesDir, filename string) (string, error) {
	relative, err := filepath.Rel(pagesDir, filename)
	if err != nil {
		return "", err
	}
	route := strings.TrimSuffix(filepath.ToSlash(relative), filepath.Ext(relative))
	if path.Base(route) == "index" {
		route = path.Dir(route)
	}
	if route == "." {
		return "", nil
	}
	return strings.Trim(route, "/"), nil
}

func (ctx *buildContext) renderPage(pageFile, target, route, locale string) error {
	functions := template.FuncMap{
		"asset": func(name string) (string, error) {
			name = strings.TrimPrefix(path.Clean("/"+name), "/")
			publicPath, ok := ctx.assets[name]
			if !ok {
				return "", fmt.Errorf("unknown static asset %q", name)
			}
			return publicPath, nil
		},
		"t": func(key string) (string, error) {
			if value, ok := ctx.translations[locale][key]; ok {
				return value, nil
			}
			if value, ok := ctx.translations[ctx.config.DefaultLocale][key]; ok {
				return value, nil
			}
			return "", fmt.Errorf("translation %q is missing for %s and the default locale", key, locale)
		},
		"url": func(targetRoute string) string {
			return localizedURL(locale, targetRoute)
		},
		"languageURL": func(targetLocale string) (string, error) {
			if !slices.Contains(ctx.config.Locales, targetLocale) {
				return "", fmt.Errorf("unknown locale %q", targetLocale)
			}
			return localizedURL(targetLocale, route), nil
		},
		"canonicalURL": func() string {
			return absoluteURL(ctx.config.BaseURL, localizedURL(locale, route))
		},
		"siteCanonicalURL": func() string {
			return absoluteURL(ctx.config.BaseURL, localizedURL(locale, ""))
		},
		"languageCanonicalURL": func(targetLocale string) (string, error) {
			if !slices.Contains(ctx.config.Locales, targetLocale) {
				return "", fmt.Errorf("unknown locale %q", targetLocale)
			}
			return absoluteURL(ctx.config.BaseURL, localizedURL(targetLocale, route)), nil
		},
		"languageFlag": languageFlag,
		"languageName": languageName,
		"markdownURL": func() string {
			return absoluteURL(ctx.config.BaseURL, localizedMarkdownURL(locale))
		},
		"openGraphLocale": openGraphLocale,
		"textDirection": func() string {
			if locale == "ar" {
				return "rtl"
			}
			return "ltr"
		},
	}

	name := filepath.Base(pageFile)
	tmpl := template.New(name).Funcs(functions)
	var err error
	if len(ctx.shared) > 0 {
		tmpl, err = tmpl.ParseFiles(ctx.shared...)
		if err != nil {
			return fmt.Errorf("parse shared templates for %s: %w", pageFile, err)
		}
	}
	tmpl, err = tmpl.ParseFiles(pageFile)
	if err != nil {
		return fmt.Errorf("parse page %s: %w", pageFile, err)
	}

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("create page directory: %w", err)
	}
	file, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("create page %s: %w", target, err)
	}
	data := Data{Site: ctx.config, Page: Page{Route: route}, Locale: locale, Locales: slices.Clone(ctx.config.Locales)}
	renderErr := tmpl.ExecuteTemplate(file, name, data)
	closeErr := file.Close()
	if renderErr != nil {
		return fmt.Errorf("render %s for %s: %w", pageFile, locale, renderErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close page %s: %w", target, closeErr)
	}
	return nil
}

func (ctx *buildContext) renderMarkdownIndex(filename string) error {
	if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return fmt.Errorf("read Markdown index: %w", err)
	}

	for _, locale := range ctx.config.Locales {
		tmpl, err := texttemplate.New(filepath.Base(filename)).Funcs(ctx.plainTextFunctions(locale)).ParseFiles(filename)
		if err != nil {
			return fmt.Errorf("parse Markdown index: %w", err)
		}

		target := filepath.Join(ctx.output, locale, "index.md")
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fmt.Errorf("create Markdown index directory: %w", err)
		}
		file, err := os.Create(target)
		if err != nil {
			return fmt.Errorf("create Markdown index for %s: %w", locale, err)
		}
		data := Data{Site: ctx.config, Page: Page{}, Locale: locale, Locales: slices.Clone(ctx.config.Locales)}
		renderErr := tmpl.ExecuteTemplate(file, filepath.Base(filename), data)
		closeErr := file.Close()
		if renderErr != nil {
			return fmt.Errorf("render Markdown index for %s: %w", locale, renderErr)
		}
		if closeErr != nil {
			return fmt.Errorf("close Markdown index for %s: %w", locale, closeErr)
		}
	}

	defaultIndex := filepath.Join(ctx.output, ctx.config.DefaultLocale, "index.md")
	contents, err := os.ReadFile(defaultIndex)
	if err != nil {
		return fmt.Errorf("read default Markdown index: %w", err)
	}
	if err := os.WriteFile(filepath.Join(ctx.output, "index.md"), contents, 0o644); err != nil {
		return fmt.Errorf("write root Markdown index: %w", err)
	}
	return nil
}

func (ctx *buildContext) renderLLMSText(filename string) error {
	if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return fmt.Errorf("read llms.txt template: %w", err)
	}

	locale := ctx.config.DefaultLocale
	tmpl, err := texttemplate.New(filepath.Base(filename)).Funcs(ctx.plainTextFunctions(locale)).ParseFiles(filename)
	if err != nil {
		return fmt.Errorf("parse llms.txt template: %w", err)
	}
	file, err := os.Create(filepath.Join(ctx.output, "llms.txt"))
	if err != nil {
		return fmt.Errorf("create llms.txt: %w", err)
	}
	data := Data{Site: ctx.config, Page: Page{}, Locale: locale, Locales: slices.Clone(ctx.config.Locales)}
	renderErr := tmpl.ExecuteTemplate(file, filepath.Base(filename), data)
	closeErr := file.Close()
	if renderErr != nil {
		return fmt.Errorf("render llms.txt: %w", renderErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close llms.txt: %w", closeErr)
	}
	return nil
}

func (ctx *buildContext) plainTextFunctions(locale string) texttemplate.FuncMap {
	return texttemplate.FuncMap{
		"t": func(key string) (string, error) {
			if value, ok := ctx.translations[locale][key]; ok {
				return value, nil
			}
			if value, ok := ctx.translations[ctx.config.DefaultLocale][key]; ok {
				return value, nil
			}
			return "", fmt.Errorf("translation %q is missing for %s and the default locale", key, locale)
		},
		"languageName": languageName,
		"languageMarkdownURL": func(targetLocale string) (string, error) {
			if !slices.Contains(ctx.config.Locales, targetLocale) {
				return "", fmt.Errorf("unknown locale %q", targetLocale)
			}
			return absoluteURL(ctx.config.BaseURL, localizedMarkdownURL(targetLocale)), nil
		},
		"pageURL": func() string {
			return absoluteURL(ctx.config.BaseURL, localizedURL(locale, ""))
		},
	}
}

func localizedMarkdownURL(locale string) string {
	return "/" + locale + "/index.md"
}

type sitemapURLSet struct {
	XMLName    xml.Name     `xml:"urlset"`
	XMLNS      string       `xml:"xmlns,attr"`
	XMLNSXHTML string       `xml:"xmlns:xhtml,attr"`
	URLs       []sitemapURL `xml:"url"`
}

type sitemapURL struct {
	Location   string             `xml:"loc"`
	Alternates []sitemapAlternate `xml:"xhtml:link"`
}

type sitemapAlternate struct {
	Rel      string `xml:"rel,attr"`
	Language string `xml:"hreflang,attr"`
	Href     string `xml:"href,attr"`
}

func writeSitemap(output string, cfg Config, pages []string, root string) error {
	urlSet := sitemapURLSet{
		XMLNS:      "http://www.sitemaps.org/schemas/sitemap/0.9",
		XMLNSXHTML: "http://www.w3.org/1999/xhtml",
	}
	for _, pageFile := range pages {
		route, err := pageRoute(filepath.Join(root, "pages"), pageFile)
		if err != nil {
			return fmt.Errorf("create sitemap route: %w", err)
		}
		alternates := make([]sitemapAlternate, 0, len(cfg.Locales)+1)
		for _, alternateLocale := range cfg.Locales {
			alternates = append(alternates, sitemapAlternate{
				Rel:      "alternate",
				Language: alternateLocale,
				Href:     absoluteURL(cfg.BaseURL, localizedURL(alternateLocale, route)),
			})
		}
		alternates = append(alternates, sitemapAlternate{
			Rel:      "alternate",
			Language: "x-default",
			Href:     absoluteURL(cfg.BaseURL, localizedURL(cfg.DefaultLocale, route)),
		})
		for _, locale := range cfg.Locales {
			urlSet.URLs = append(urlSet.URLs, sitemapURL{
				Location:   absoluteURL(cfg.BaseURL, localizedURL(locale, route)),
				Alternates: alternates,
			})
		}
	}

	contents, err := xml.MarshalIndent(urlSet, "", "  ")
	if err != nil {
		return fmt.Errorf("encode sitemap: %w", err)
	}
	contents = append([]byte(xml.Header), contents...)
	contents = append(contents, '\n')
	if err := os.WriteFile(filepath.Join(output, "sitemap.xml"), contents, 0o644); err != nil {
		return fmt.Errorf("write sitemap: %w", err)
	}
	return nil
}

func writeRobots(output, baseURL string) error {
	sitemap := absoluteURL(baseURL, "/sitemap.xml")
	contents := fmt.Sprintf("User-agent: *\nAllow: /\n\nSitemap: %s\n", sitemap)
	if err := os.WriteFile(filepath.Join(output, "robots.txt"), []byte(contents), 0o644); err != nil {
		return fmt.Errorf("write robots.txt: %w", err)
	}
	return nil
}

func absoluteURL(baseURL, targetPath string) string {
	if baseURL == "" {
		return targetPath
	}
	return strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(targetPath, "/")
}

func openGraphLocale(locale string) string {
	locales := map[string]string{
		"ar": "ar_AR", "cs": "cs_CZ", "de": "de_DE", "en": "en_GB",
		"es": "es_ES", "fr": "fr_FR", "hi": "hi_IN", "it": "it_IT",
		"ja": "ja_JP", "ko": "ko_KR", "pl": "pl_PL", "pt": "pt_PT",
		"ru": "ru_RU", "tr": "tr_TR",
	}
	if value, exists := locales[locale]; exists {
		return value
	}
	return strings.ReplaceAll(locale, "-", "_")
}

func languageName(locale string) string {
	names := map[string]string{
		"ar": "العربية", "cs": "Čeština", "de": "Deutsch", "en": "English",
		"es": "Español", "fr": "Français", "hi": "हिन्दी", "it": "Italiano",
		"ja": "日本語", "ko": "한국어", "pl": "Polski", "pt": "Português",
		"ru": "Русский", "tr": "Türkçe",
	}
	if name, exists := names[locale]; exists {
		return name
	}
	return locale
}

func languageFlag(locale string) string {
	flags := map[string]string{
		"ar": "🇸🇦", "cs": "🇨🇿", "de": "🇩🇪", "en": "🇬🇧",
		"es": "🇪🇸", "fr": "🇫🇷", "hi": "🇮🇳", "it": "🇮🇹",
		"ja": "🇯🇵", "ko": "🇰🇷", "pl": "🇵🇱", "pt": "🇵🇹",
		"ru": "🇷🇺", "tr": "🇹🇷",
	}
	if flag, exists := flags[locale]; exists {
		return flag
	}
	return "🌐"
}

func localizedURL(locale, route string) string {
	route = strings.Trim(path.Clean("/"+route), "/")
	if route == "" || route == "." {
		return "/" + locale + "/"
	}
	return "/" + locale + "/" + route + "/"
}

func writeRedirect(output, defaultLocale string, locales []string) error {
	localesJSON, err := json.Marshal(locales)
	if err != nil {
		return fmt.Errorf("encode locales for language redirect: %w", err)
	}
	defaultJSON, err := json.Marshal(defaultLocale)
	if err != nil {
		return fmt.Errorf("encode default locale for language redirect: %w", err)
	}

	// Client-side redirect: match navigator.languages against configured locales
	// (exact, then language prefix). Fall back to defaultLocale when unsupported.
	contents := fmt.Sprintf(`<!doctype html>
<html lang="%[1]s">
<head>
<meta charset="utf-8">
<meta name="robots" content="noindex">
<link rel="canonical" href="/%[1]s/">
<title>FixYourMoney</title>
<script>
(function () {
  var supported = %[2]s;
  var fallback = %[3]s;
  var supportedSet = Object.create(null);
  for (var i = 0; i < supported.length; i++) {
    supportedSet[supported[i].toLowerCase()] = supported[i];
  }
  function resolveLocale(tag) {
    if (!tag) return null;
    var normalized = String(tag).toLowerCase().replace(/_/g, "-");
    if (supportedSet[normalized]) return supportedSet[normalized];
    var language = normalized.split("-")[0];
    if (supportedSet[language]) return supportedSet[language];
    return null;
  }
  var candidates = [];
  if (typeof navigator !== "undefined") {
    if (Array.isArray(navigator.languages)) {
      candidates = candidates.concat(navigator.languages);
    }
    if (navigator.language) {
      candidates.push(navigator.language);
    }
  }
  var locale = fallback;
  for (var j = 0; j < candidates.length; j++) {
    var match = resolveLocale(candidates[j]);
    if (match) {
      locale = match;
      break;
    }
  }
  window.location.replace("/" + locale + "/");
})();
</script>
<noscript>
<meta http-equiv="refresh" content="0; url=/%[1]s/">
</noscript>
</head>
<body>
<p><a href="/%[1]s/">Continue</a></p>
</body>
</html>
`, defaultLocale, string(localesJSON), string(defaultJSON))

	if err := os.WriteFile(filepath.Join(output, "index.html"), []byte(contents), 0o644); err != nil {
		return fmt.Errorf("write language redirect: %w", err)
	}
	return nil
}
