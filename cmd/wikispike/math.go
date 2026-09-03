package main

import (
	"encoding/json"
	"fmt"
	stdhtml "html"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/doug/termtex"
)

// renderMath replaces each Wikipedia formula (a span carrying
// typeof="mw:Extension/math") with plain text, so the formula survives the
// pipeline as ordinary prose instead of silently vanishing when stripChrome
// drops image/math markup — the S6a maths exception (D104, M3.4).
//
// Two paths, per formula, cheapest first:
//
//   - Unicode approximation via termtex, kept only if the rendered result
//     stays on one line. termtex's layout for fractions, multi-glyph
//     superscripts without a single Unicode codepoint, \sum, \begin{align}
//     and similar produces a multi-row ASCII-art block (box-drawing
//     characters stacked over several lines) — fine for a formula that owns
//     its own line, but Parsoid's HTML puts these spans mid-sentence even
//     when marked "display: block" (checked against a live fetch of
//     `E (mathematical constant)`: none of that article's block-style
//     formulas sit inside their own <dl>/<dd> line), so a multi-line result
//     would tear the sentence apart rather than read as prose. Restricting
//     to single-line output keeps every substitution safely inline.
//   - Otherwise (termtex errors, or its result is multi-line, or there is no
//     usable source at all): fall back to the formula's raw TeX source
//     (already permitted, unchanged) so the sentence still reads, just with
//     LaTeX-flavoured text where the rendered formula used to be.
//
// Empirically, against a live fetch of all four S5 titles: see
// cycle/findings/maths-experiment.md for the per-title tally this rule
// produces.
func renderMath(html string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", fmt.Errorf("parsing HTML to render maths: %w", err)
	}

	doc.Find(`[typeof="mw:Extension/math"]`).Each(func(_ int, s *goquery.Selection) {
		text := mathReplacementText(s)
		if text == "" {
			// No usable source (missing/unparseable data-mw): leave the
			// span in place for stripChrome's own math selector to remove,
			// same as before this experiment existed.
			return
		}
		s.ReplaceWithHtml(stdhtml.EscapeString(text))
	})

	out, err := doc.Html()
	if err != nil {
		return "", fmt.Errorf("rendering maths-replaced HTML: %w", err)
	}
	return out, nil
}

// mathDataMW is the slice of Parsoid's data-mw JSON this needs: the
// formula's raw TeX source, under body.extsrc — present for every
// mw:Extension/math span Parsoid emits, with no OCR or extra fetch required
// (D104).
type mathDataMW struct {
	Body struct {
		Extsrc string `json:"extsrc"`
	} `json:"body"`
}

// mathReplacementText picks a formula's replacement text per renderMath's
// two-path rule. Returns "" if no source could be recovered at all.
func mathReplacementText(s *goquery.Selection) string {
	raw, ok := s.Attr("data-mw")
	if !ok {
		return ""
	}
	var data mathDataMW
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return ""
	}
	src := strings.TrimSpace(data.Body.Extsrc)
	if src == "" {
		return ""
	}

	if rendered, err := termtex.Render(src, termtex.Style{}); err == nil {
		if trimmed := strings.TrimRight(rendered, "\n"); !strings.Contains(trimmed, "\n") {
			return trimmed
		}
	}
	return src
}
