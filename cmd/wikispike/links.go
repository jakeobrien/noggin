package main

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// stripLinks unwraps every remaining <a> element, keeping its content in
// place as ordinary prose (M2.3) — no markdown link syntax (`[text](url)`)
// and no bare URLs should reach the converted markdown. It runs after
// stripChrome, which has already removed whole link-bearing subtrees (image
// wrappers, citation superscripts, reference lists) — what is left here is
// prose and table-cell wikilinks.
func stripLinks(html string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", fmt.Errorf("parsing HTML to strip links: %w", err)
	}

	doc.Find("a").Each(func(_ int, a *goquery.Selection) {
		inner, err := a.Html()
		if err != nil {
			// A fragment that fails to re-serialize is vanishingly unlikely
			// (it already parsed once as part of doc), but falling back to
			// plain text beats crashing the spike over a single anchor.
			inner = a.Text()
		}
		a.ReplaceWithHtml(inner)
	})

	out, err := doc.Html()
	if err != nil {
		return "", fmt.Errorf("rendering delinked HTML: %w", err)
	}
	return out, nil
}
