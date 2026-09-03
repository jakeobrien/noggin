package main

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// stripChrome removes everything from Parsoid article HTML that is not the
// article proper (M2.2): navboxes, infoboxes, citation superscripts,
// reference-list sections, metadata tables and images. Edit-section links
// (.mw-editsection) are included defensively even though Parsoid's REST HTML
// never carries them — see .charm/cycle/PROGRESS.md's handoff. Content tables
// (table.wikitable) are untouched; link stripping is M2.3, blank-line/orphan
// cleanup is M2.4.
func stripChrome(html string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", fmt.Errorf("parsing HTML to strip chrome: %w", err)
	}

	// Navboxes and infoboxes: navigation/summary tables, not article prose.
	doc.Find("div.navbox").Remove()
	doc.Find("table.infobox, .infobox").Remove()

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
	// typeof starting with "mw:File". A final bare-<img> sweep catches
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
