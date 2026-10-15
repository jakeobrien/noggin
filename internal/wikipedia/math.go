package wikipedia

import (
	"encoding/json"
	stdhtml "html"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/doug/termtex"
)

// renderMath replaces each Wikipedia formula (a span carrying
// typeof="mw:Extension/math") with plain text, so the formula survives the
// pipeline as ordinary prose instead of silently vanishing when stripChrome
// drops image/math markup — the S6a maths exception (M3.4).
//
// Parsoid marks a formula as display/block-style two ways, checked against a
// live fetch of `E (mathematical constant)`: the span carries
// `data-mw.attrs.display == "block"` and a `mwe-math-element-block` class
// (`mwe-math-element-inline` otherwise). That's exactly the signal Wikipedia
// itself uses to give a formula its own line visually (CSS `display: block`)
// even though Parsoid still nests the span inside the same running <p> as
// its surrounding prose — so this rewrite follows suit, per formula:
//
//   - Display/block formulas get their own line, not folded into the
//     sentence: rendered to a <pre> element. `<pre>` inside a `<p>` is not
//     valid nesting, so when the html-to-markdown converter's own parser
//     re-parses the serialized HTML, HTML5 parsing rules auto-close the
//     surrounding paragraph around it — the same mechanism that gives it a
//     line of its own in a browser. Rendering to `<pre>` also lifts M3.4's
//     single-line restriction: termtex's multi-line Unicode layout (stacked
//     fractions, sums, box-drawing) is fine once the formula is not fighting
//     a sentence for a line, so this path prefers Unicode even when
//     multi-line, falling back to raw TeX source only if termtex errors.
//   - Inline formulas keep M3.4's original two-path rule unchanged: Unicode
//     kept only if it stays on one line (otherwise it would tear the
//     sentence apart), else the raw TeX source as prose.
//
// Empirically, against a live fetch of all four S5 titles: see
// cycle/findings/maths-experiment.md for the per-title tally the original
// inline-only rule produced.
func renderMath(html string) (string, error) {
	return transformHTML(html, "render maths", func(doc *goquery.Document) {
		doc.Find(selectorMathExtension).Each(func(_ int, s *goquery.Selection) {
			text, isBlock := mathReplacementText(s)
			if text == "" {
				// No usable source (missing/unparseable data-mw): leave
				// the span in place for stripChrome's own math selector
				// to remove, same as before this experiment existed.
				return
			}
			escaped := stdhtml.EscapeString(text)
			if isBlock {
				s.ReplaceWithHtml("<pre>" + escaped + "</pre>")
				return
			}
			s.ReplaceWithHtml(escaped)
		})
	})
}

// mathDataMW is the slice of Parsoid's data-mw JSON this needs: the
// formula's raw TeX source, under body.extsrc, and whether Parsoid marked it
// display/block-style, under attrs.display — both present for every
// mw:Extension/math span Parsoid emits, with no OCR or extra fetch required.
type mathDataMW struct {
	Attrs struct {
		Display string `json:"display"`
	} `json:"attrs"`
	Body struct {
		Extsrc string `json:"extsrc"`
	} `json:"body"`
}

// isDisplayMath reports whether Parsoid marked this formula's span as
// display/block-style, checking both signals it uses (data-mw's declared
// attrs.display and the class Parsoid's own CSS keys "own line" rendering
// off) so a span missing one still resolves correctly off the other.
func isDisplayMath(s *goquery.Selection, data mathDataMW) bool {
	if data.Attrs.Display == "block" {
		return true
	}
	class, _ := s.Attr("class")
	return strings.Contains(class, "mwe-math-element-block")
}

// mathReplacementText picks a formula's replacement text and whether it
// should get its own line (display/block) or stay inline, per renderMath's
// rule. Returns ("", false) if no source could be recovered at all.
func mathReplacementText(s *goquery.Selection) (string, bool) {
	raw, ok := s.Attr("data-mw")
	if !ok {
		return "", false
	}
	var data mathDataMW
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return "", false
	}
	src := strings.TrimSpace(data.Body.Extsrc)
	if src == "" {
		return "", false
	}
	isBlock := isDisplayMath(s, data)

	rendered, err := termtex.Render(src, termtex.Style{})
	if err != nil {
		return src, isBlock
	}
	trimmed := strings.TrimRight(rendered, "\n")
	if isBlock || !strings.Contains(trimmed, "\n") {
		// Display formulas may keep multi-line Unicode output — they get
		// their own line via <pre>, so there is no sentence to tear apart.
		// Inline formulas keep M3.4's single-line-only restriction.
		return trimmed, isBlock
	}
	return src, isBlock
}
