package wikipedia

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// typeofFilePrefix and typeofMathExtension are the two Parsoid `typeof`
// values this pipeline matches on repeatedly — stripChrome's catch-all
// image/math removal (strip.go), renderImages' image-placeholder pass
// (images.go) and renderMath's math-placeholder pass (math.go) each used to
// spell these out as their own string literal; named once here instead
// (TRIAGE T6, M3.6).
const (
	typeofFilePrefix    = "mw:File"
	typeofMathExtension = "mw:Extension/math"
	typeofRefExtension  = "mw:Extension/ref"

	// typeofTransclusion and classInlineTemplate identify two more
	// citation-like superscript shapes distinct from typeofRefExtension
	// (TRIAGE T1, M3.9): Wikipedia's maintenance-template family
	// ({{citation needed}}, {{clarification needed}}, {{dubious}},
	// {{disputed}} and similar) renders its visible "[citation needed]"
	// bracket as a plain <sup class="noprint Inline-Template .."> with no
	// typeof at all — the typeof="mw:Transclusion" marker Parsoid emits for
	// the template instead lands on a preceding, invisible <link
	// rel="mw:PageProp/Category"> sharing the same `about` id, not on the
	// visible sup — so this family is matched by class, not typeof. The
	// page-locator template ({{rp|<page>}}, e.g. a bare ": 25" after a
	// citation) is the opposite shape: its own <sup> carries
	// typeof="mw:Transclusion" directly (exact match — a real citation sup's
	// typeof is always the compound "mw:Transclusion mw:Extension/ref",
	// already caught by typeofRefExtension's substring match, so this exact
	// match never double-fires on it).
	typeofTransclusion  = "mw:Transclusion"
	classInlineTemplate = "Inline-Template"
)

// selectorFilePrefix, selectorMathExtension and selectorCitationLikeSup are
// the goquery CSS selectors built from the constants above — "starts with"
// for the file typeof (Parsoid suffixes it, e.g. "mw:File/Thumb"), exact
// match for the math extension typeof (Parsoid never suffixes it), and a
// combined selector for every citation-like superscript this pipeline
// removes: real citation-reference markers, the maintenance-template family,
// and the page-locator template (TRIAGE T1, M3.9). Reused by renderImages
// (M3.5a) so it can strip any of these out of a figcaption before
// stripChrome removes them elsewhere in the document — alt is a plain
// string attribute and can never carry a nested <sup> to begin with, so
// only the figcaption case is real.
var (
	selectorFilePrefix             = fmt.Sprintf(`[typeof^=%q]`, typeofFilePrefix)
	selectorMathExtension          = fmt.Sprintf(`[typeof=%q]`, typeofMathExtension)
	selectorCitationSup            = fmt.Sprintf(`sup[typeof*=%q]`, typeofRefExtension)
	selectorMaintenanceTemplateSup = fmt.Sprintf(`sup.%s`, classInlineTemplate)
	selectorPageLocatorSup         = fmt.Sprintf(`sup[typeof=%q]`, typeofTransclusion)
	selectorCitationLikeSup        = strings.Join([]string{selectorCitationSup, selectorMaintenanceTemplateSup, selectorPageLocatorSup}, ", ")
)

// transformHTML parses html into a goquery document, lets mutate rewrite it
// in place, and re-serializes the result — the parse/mutate/re-serialize
// shape stripChrome, stripLinks, renderMath and renderImages each repeated
// as their own boilerplate (TRIAGE T6, M3.6). stage names the calling
// pipeline step (e.g. "strip chrome") for error context only.
func transformHTML(html, stage string, mutate func(*goquery.Document)) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", fmt.Errorf("parsing HTML to %s: %w", stage, err)
	}
	mutate(doc)
	out, err := doc.Html()
	if err != nil {
		return "", fmt.Errorf("re-serializing HTML after %s: %w", stage, err)
	}
	return out, nil
}
