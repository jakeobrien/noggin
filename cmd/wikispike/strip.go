package main

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// stripChrome removes everything from Parsoid article HTML that is not the
// article proper (M2.2): navboxes, sidebars, infoboxes, citation
// superscripts, reference-list sections, metadata tables and images.
// Edit-section links (.mw-editsection) are included defensively even though
// Parsoid's REST HTML never carries them — see .charm/cycle/PROGRESS.md's
// handoff. Content tables (table.wikitable) are untouched; link stripping is
// M2.3, blank-line/orphan cleanup is M2.4.
func stripChrome(html string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", fmt.Errorf("parsing HTML to strip chrome: %w", err)
	}

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

	// Edit-section links: a skin-rendered-page feature, not present in
	// Parsoid HTML, but harmless to also look for.
	doc.Find(".mw-editsection").Remove()

	// Citation superscripts, e.g. the "[12]" after a claim.
	doc.Find(`sup[typeof*="mw:Extension/ref"]`).Remove()

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
	// renderImages (M3.5, D104) has already replaced every image it could
	// derive a caption/alt/filename description for with a plain-text
	// placeholder — this selector only catches the rare leftover
	// renderImages couldn't handle. A final bare-<img> sweep catches
	// anything else that slipped through.
	//
	// Rendered math formulae are also served as an <img> (an SVG render)
	// under typeof="mw:Extension/math", but by the time stripChrome runs,
	// renderMath (M3.4, D104) has already replaced every formula it could
	// recover a source for with plain text — this selector only catches the
	// rare leftover renderMath couldn't handle (missing/unparseable data-mw).
	doc.Find(`[typeof^="mw:File"], [typeof="mw:Extension/math"]`).Remove()
	doc.Find("img").Remove()

	out, err := doc.Html()
	if err != nil {
		return "", fmt.Errorf("rendering stripped HTML: %w", err)
	}
	return out, nil
}
