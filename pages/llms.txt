# {{ .Site.Title }}

> {{ t "meta.description" }}

<{{ pageURL }}>

## {{ t "introduction.title" }}

{{ t "hero.text" }}

## {{ t "currencies.headline" }}

*{{ t "currencies.subheading" }}*

{{ t "currencies.text" }}

- {{ t "currencies.points.1" }}
- {{ t "currencies.points.2" }}

## {{ t "assets.headline" }}

*{{ t "assets.subheading" }}*

{{ t "assets.text" }}

- {{ t "assets.points.1" }}
- {{ t "assets.points.2" }}
- {{ t "assets.points.3" }}

## {{ t "competition.headline" }}

*{{ t "competition.subheading" }}*

{{ t "competition.text" }}

### {{ t "competition.functions-headline" }}

{{ t "competition.functions-text" }}

#### {{ t "competition.functions.1.headline" }}

{{ t "competition.functions.1.text" }}

#### {{ t "competition.functions.2.headline" }}

{{ t "competition.functions.2.text" }}

#### {{ t "competition.functions.3.headline" }}

{{ t "competition.functions.3.text" }}

#### {{ t "competition.functions.4.headline" }}

{{ t "competition.functions.4.text" }}

#### {{ t "competition.functions.5.headline" }}

{{ t "competition.functions.5.text" }}

#### {{ t "competition.functions.6.headline" }}

{{ t "competition.functions.6.text" }}

#### {{ t "competition.functions.7.headline" }}

{{ t "competition.functions.7.text" }}

### {{ t "competition.confrontations.1.first" }} / {{ t "competition.confrontations.1.second" }}

{{ t "competition.confrontations.1.statement" }}

### {{ t "competition.confrontations.2.first" }} / {{ t "competition.confrontations.2.second" }}

{{ t "competition.confrontations.2.statement" }}

### {{ t "competition.confrontations.3.first" }} / {{ t "competition.confrontations.3.second" }}

{{ t "competition.confrontations.3.statement" }}

## {{ t "candidates.headline" }}

*{{ t "candidates.chart-headline" }}*

{{ t "candidates.text" }}

- {{ t "candidates.points.1" }}
- {{ t "candidates.points.2" }}
- {{ t "candidates.points.3" }}
- {{ t "candidates.points.4" }}
- {{ t "candidates.points.5" }}

## {{ t "impact.headline" }}

*{{ t "impact.subheading" }}*

{{ t "impact.text" }}

### {{ t "impact.cards.individual.headline" }}

{{ t "impact.cards.individual.text" }}

- {{ t "impact.cards.individual.points.1" }}
- {{ t "impact.cards.individual.points.2" }}
- {{ t "impact.cards.individual.points.3" }}

### {{ t "impact.cards.companies.headline" }}

{{ t "impact.cards.companies.text" }}

- {{ t "impact.cards.companies.points.1" }}
- {{ t "impact.cards.companies.points.2" }}
- {{ t "impact.cards.companies.points.3" }}

### {{ t "impact.cards.society.headline" }}

{{ t "impact.cards.society.text" }}

- {{ t "impact.cards.society.points.1" }}
- {{ t "impact.cards.society.points.2" }}
- {{ t "impact.cards.society.points.3" }}
- {{ t "impact.cards.society.points.4" }}
- {{ t "impact.cards.society.points.5" }}

## {{ t "vision.headline" }}

{{ t "vision.text" }}

## {{ t "closing.message" }}

## {{ t "nav.languages" }}

{{- range .Locales }}
- [{{ languageName . }}]({{ languageMarkdownURL . }})
{{- end }}
