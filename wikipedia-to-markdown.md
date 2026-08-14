# Converting Wikipedia Articles to Markdown

Wikipedia has no native Markdown export. Its native formats are **wikitext** (source markup) and **parsed HTML**. Getting clean Markdown means fetching one of those and converting it — a well-trodden but not "one API call" path.

## 1. Fetching the content

### Option A — Pre-rendered HTML (recommended)
Use the Wikimedia REST API, available on every Wikimedia project at `/api/rest_v1/`:

```
GET https://en.wikipedia.org/api/rest_v1/page/html/{title}
```

- No auth token required — just set a descriptive `User-Agent` header.
- Returns MediaWiki's already-parsed HTML, which normalizes away most of wikitext's template/parser-function complexity.
- This is the easiest starting point for a converter pipeline.

### Option B — Raw wikitext
Use the newer Core REST API:

```
GET https://api.wikimedia.org/core/v1/wikipedia/en/page/{title}
```

- Requires a bearer token (register an app at api.wikimedia.org).
- Returns the raw wikitext source — you'll then need a wikitext-aware converter (see §3).

### Option C — Structured Contents (scale / commercial use)
Wikimedia Enterprise offers a **Structured Contents** endpoint that returns pre-parsed JSON broken into sections, paragraphs, lists, and tables — no HTML or wikitext parsing needed at all. Databricks used this to build a Markdown generator at scale. It's a paid/commercial-tier product, not free public API.

## 2. Converting HTML → Markdown in Go

**Library:** [`github.com/JohannesKaufmann/html-to-markdown`](https://github.com/JohannesKaufmann/html-to-markdown) (v2)

- Pure Go, built on a real HTML parser (not regex), which matters for Wikipedia's irregular markup.
- Plugin system, including GitHub-Flavored tables — useful since Wikipedia infoboxes and data tables are common.

### Minimal pipeline

```go
package main

import (
    "fmt"
    "io"
    "net/http"
    "strings"

    md "github.com/JohannesKaufmann/html-to-markdown"
    "github.com/JohannesKaufmann/html-to-markdown/plugin"
    "github.com/PuerkitoBio/goquery"
)

func main() {
    req, _ := http.NewRequest("GET",
        "https://en.wikipedia.org/api/rest_v1/page/html/Golang_(programming_language)", nil)
    req.Header.Set("User-Agent", "myapp/0.1 (me@example.com)")

    resp, _ := http.DefaultClient.Do(req)
    body, _ := io.ReadAll(resp.Body)

    // Strip Wikipedia cruft before converting
    doc, _ := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
    doc.Find(".navbox, .infobox, .reflist, sup.reference, .mw-editsection, table.metadata").Remove()
    cleanHTML, _ := doc.Html()

    converter := md.NewConverter("", true, nil)
    converter.Use(plugin.GitHubFlavored())

    markdown, _ := converter.ConvertString(cleanHTML)
    fmt.Println(markdown)
}
```

The `goquery` cleanup step matters in practice — skip it and you get navboxes, edit-section links, and citation superscripts cluttering the output. This is essentially what hosted "Wikipedia → Markdown" tools do under the hood.

## 3. Converting wikitext → Markdown (if you went with Option B)

- **Pandoc** is the trusted option: `pandoc -f mediawiki -t markdown`. Extremely well-tested, but it's an external binary, not a Go library — you'd shell out to it.
- **Go wikitext parsers** exist (e.g. `m-m-f/gowiki`) but are far less mature/maintained than Pandoc's MediaWiki reader or Python's `mwparserfromhell`. Not recommended for production-quality conversion.

## 4. Recommended approach for a Go project

```
REST API HTML fetch  →  goquery cleanup  →  html-to-markdown (GFM plugin)
```

This combination has the most battle-testing behind each piece and stays pure Go end-to-end.

## 5. If you'd rather not build it

Hosted converters exist for occasional/one-off use (web2md, inktomd, an Apify actor) — fine for casual use, but not something to depend on for anything serious or high-volume.

## Quick reference

| Need | Use |
|---|---|
| Article HTML, no auth | `en.wikipedia.org/api/rest_v1/page/html/{title}` |
| Article wikitext, needs token | `api.wikimedia.org/core/v1/wikipedia/en/page/{title}` |
| Pre-parsed JSON at scale | Wikimedia Enterprise Structured Contents (paid) |
| HTML → Markdown in Go | `JohannesKaufmann/html-to-markdown` |
| Wikitext → Markdown (any language) | Pandoc (`-f mediawiki -t markdown`) |
