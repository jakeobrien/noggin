package wikipedia

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// sisterProjectInlineLine matches a rendered sister-project link line
// (M3.3b, TRIAGE T10/T12) that div.portal-bar/div.sistersitebox above don't
// catch: Wikipedia's "-inline" template family (Commons category-inline,
// Wiktionary-inline, Wikiquote-inline and similar) renders as a bare <li>
// (or <p>) in the External-links section with no wrapping box element for a
// selector to key on — several sibling spans/an <a> sharing one Parsoid
// `about` id, e.g. "Media related to Major Arcana at Wikimedia Commons".
// Matched against the whole trimmed text of a candidate element (see below),
// not as a substring search, so real prose that merely mentions a sister
// project in passing is left alone. An optional leading image placeholder
// (renderImages, M3.5, already ran by the time stripChrome does) is allowed
// ahead of the phrase.
var sisterProjectInlineLine = regexp.MustCompile(`(?is)^(?:\[Image:[^\]]*\]\s*)?(?:` +
	`Media related to .+ at Wikimedia Commons|` +
	`Quotations related to .+ at Wikiquote|` +
	`(?:The dictionary definition of|Look up) .+ (?:at|in) Wiktionary(?:, the free dictionary)?|` +
	`Wikisource has (?:the )?original text related to this article:? .+|` +
	`Wikivoyage has a travel guide (?:for|to) .+|` +
	`Wikibooks has (?:a book on the topic of|more on this topic):? .+|` +
	`Wikinews has related news:? .+|` +
	`Wikispecies has information related to .+|` +
	`Wikidata has the property:? .+` +
	`)[.,]?$`)

// elementText is s.Text() with any <style>/<script> descendants excluded
// first. Parsoid inlines a templatestyles <style> tag's raw CSS as a text
// child of the same <li>/<p> a sister-project-inline template renders into
// (confirmed live on Major Arcana's Commons category-inline, M3.3b) —
// goquery's Text() does not skip style/script content the way a browser or
// this pipeline's own html-to-markdown step does, so a naive s.Text() call
// here would see that CSS prepended to the line's real wording and never
// match sisterProjectInlineLine's whole-string anchors.
func elementText(s *goquery.Selection) string {
	clone := s.Clone()
	clone.Find("style, script").Remove()
	return clone.Text()
}

// stripChrome removes everything from Parsoid article HTML that is not the
// article proper (M2.2): navboxes, sidebars, infoboxes, citation
// superscripts, reference-list sections, metadata tables and images.
// Edit-section links (.mw-editsection) are included defensively even though
// Parsoid's REST HTML never carries them. Content tables (table.wikitable)
// are untouched; link stripping is M2.3, blank-line/orphan cleanup is M2.4.
func stripChrome(html string) (string, error) {
	return transformHTML(html, "strip chrome", func(doc *goquery.Document) {
		stripChromeInPlace(doc)
	})
}

// stripChromeInPlace holds stripChrome's actual removal steps, factored out
// so transformHTML (TRIAGE T6, M3.6) owns the parse/re-serialize
// boilerplate. Most steps are plain CSS-selector removal, but one
// (removeSisterProjectInlineLines, TRIAGE T5/M3.8) is a text-pattern match
// instead, because the chrome it targets has no selector to key on — see
// that function's own doc comment.
func stripChromeInPlace(doc *goquery.Document) {
	// Navboxes and infoboxes: navigation/summary tables, not article prose.
	doc.Find("div.navbox").Remove()
	doc.Find("table.infobox, .infobox").Remove()

	// Sidebars: a second, distinct navigation-box template Wikipedia floats
	// beside prose (e.g. "Part of a series of articles on ..." on
	// `E (mathematical constant)`) — CSS class "sidebar", not "navbox", so
	// the selector above misses it. Parsoid leaves the whole <table> sitting
	// as an ordinary child of the section it decorates (right after the
	// heading, before the section's real prose), with no structural marker
	// of the floated/boxed layout it renders as on the web — converted
	// as-is it reads as a disjointed run of paragraphs and lists ahead of
	// the section's actual content, not as a box. Dropped entirely, same as
	// a navbox, on the same M2.2 chrome-vs-content basis (S3a): what a
	// sidebar links to almost always exists in the article body too
	// (it is a summary/navigation aid, not a unique source), same
	// assumption already accepted for navbox and not separately re-verified
	// here.
	doc.Find("table.sidebar").Remove()

	// Portal bar / sister-project side-box: two more navigational chrome
	// boxes (TRIAGE T2, M3.3a) that predate this fix but were only visible
	// once M3.5's renderImages started replacing their icon images with
	// placeholders instead of silently deleting them — e.g. a trailing
	// "Philosophy"/"Physics"/"Astronomy"/"Religion"/"Science" list
	// (div.portal-bar) or a "Wikimedia Commons"/"Wikiquote" line
	// (div.sistersitebox). Same navbox/sidebar basis: navigational links to
	// content that exists (or is findable) elsewhere, not unique article
	// prose.
	doc.Find("div.portal-bar, div.sistersitebox").Remove()

	removeSisterProjectInlineLines(doc)

	// Edit-section links: a skin-rendered-page feature, not present in
	// Parsoid HTML, but harmless to also look for.
	doc.Find(".mw-editsection").Remove()

	// Citation-like superscripts: real reference markers (e.g. the "[12]"
	// after a claim), the maintenance-template family (e.g. "[citation
	// needed]", "[dubious – discuss]") and the page-locator template (e.g.
	// a bare ": 25" after a citation) — see selectorCitationLikeSup's own
	// doc comment for why these need three different selector shapes
	// (TRIAGE T1, M3.9). renderImages (M3.5a) already strips these out of
	// any image caption/alt it read before this stage runs; this is the
	// general sweep over the rest of the document.
	doc.Find(selectorCitationLikeSup).Remove()

	// Reference-list sections (both the unqualified "References" list and
	// named groups like "Notes") wrap their list inside their own <section>.
	// The typeof="mw:Extension/references" marker lands on the wrapping div
	// for some articles and on the inner <ol> for others, so also match on
	// the wrapper's stable class name. Drop the whole section so no empty
	// "References"/"Notes" heading is left behind.
	doc.Find(`.mw-references-wrap, [typeof*="mw:Extension/references"]`).Each(func(_ int, s *goquery.Selection) {
		if sec := s.Closest("section"); sec.Length() > 0 {
			sec.Remove()
		} else {
			s.Remove()
		}
	})

	// Metadata tables (e.g. ambox-style maintenance boxes); wikitable content
	// tables are deliberately not matched here.
	doc.Find("table.metadata, table.ambox").Remove()

	// Images: figures, thumbnails and inline file references all carry a
	// typeof starting with "mw:File". By the time stripChrome runs,
	// renderImages (M3.5) has already replaced every image it could
	// derive a caption/alt/filename description for with a plain-text
	// placeholder — this selector only catches the rare leftover
	// renderImages couldn't handle. A final bare-<img> sweep catches
	// anything else that slipped through.
	//
	// Rendered math formulae are also served as an <img> (an SVG render)
	// under typeof="mw:Extension/math", but by the time stripChrome runs,
	// renderMath (M3.4) has already replaced every formula it could
	// recover a source for with plain text — this selector only catches the
	// rare leftover renderMath couldn't handle (missing/unparseable data-mw).
	doc.Find(selectorFilePrefix + ", " + selectorMathExtension).Remove()
	doc.Find("img").Remove()
}

// removeSisterProjectInlineLines removes Wikipedia's "-inline" sister-
// project template family (Commons category-inline, Wiktionary-inline,
// Wikiquote-inline and similar, M3.3b, TRIAGE T10/T12). The boxed variants
// that stripChromeInPlace's div.portal-bar and div.sistersitebox selectors
// catch, above, don't catch this one — confirmed absent from the live
// Parsoid HTML as either selector. That is because the whole line here is
// inline content strung across sibling spans/an <a> inside a plain <li>
// (or <p>), with no structural marker except the wording itself, so this
// is a text-pattern match against each list item/paragraph's whole
// rendered text (see sisterProjectInlineLine) instead of a CSS selector
// like every other step in stripChromeInPlace.
func removeSisterProjectInlineLines(doc *goquery.Document) {
	doc.Find("li, p").Each(func(_ int, s *goquery.Selection) {
		if sisterProjectInlineLine.MatchString(strings.TrimSpace(elementText(s))) {
			s.Remove()
		}
	})
}
