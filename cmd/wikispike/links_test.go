package main

import (
	"strings"
	"testing"
)

// TestStripLinksKeepsTextDropsMarkup is M2.3's acceptance check on a small
// fixture: link text survives as ordinary prose, the href does not.
func TestStripLinksKeepsTextDropsMarkup(t *testing.T) {
	const input = `<html><body>
<p>The <a href="./The_Fool_%28tarot_card%29" title="The Fool (tarot card)">Fool</a> is a
<a rel="mw:WikiLink" href="./Trump_(card_games)">trump card</a>, discussed at
<a rel="mw:ExtLink" href="http://example.com/article">an external site</a>.</p>
</body></html>`

	got, err := stripLinks(input)
	if err != nil {
		t.Fatalf("stripLinks: %v", err)
	}
	for _, want := range []string{"Fool", "trump card", "an external site"} {
		if !strings.Contains(got, want) {
			t.Errorf("link text %q lost:\n%s", want, got)
		}
	}
	for _, unwanted := range []string{"<a ", "href=", "example.com", "The_Fool"} {
		if strings.Contains(got, unwanted) {
			t.Errorf("link markup %q survived:\n%s", unwanted, got)
		}
	}
}

// TestArticleMarkdownHasNoLinkSyntax runs the full pipeline's HTML stages
// (strip, delink) but not conversion, on a fixture shaped like a Major Arcana
// table row, confirming the surrounding sentence still reads once the link
// is gone.
func TestStripLinksPreservesSurroundingProse(t *testing.T) {
	const input = `<p>See <a href="./Tarot_de_Marseille">Tarot de Marseille</a> for the older naming.</p>`
	got, err := stripLinks(input)
	if err != nil {
		t.Fatalf("stripLinks: %v", err)
	}
	md, err := htmlToMarkdown(got)
	if err != nil {
		t.Fatalf("htmlToMarkdown: %v", err)
	}
	want := "See Tarot de Marseille for the older naming."
	if strings.TrimSpace(md) != want {
		t.Errorf("markdown = %q, want %q", strings.TrimSpace(md), want)
	}
}

// TestPipelineLiveSamplesHaveNoLinkSyntax is M2.3's acceptance check against
// the four S5 sample titles (shared live-fetch cache, M2.5): no markdown
// link syntax and no bare URL syntax, with previously-linked phrases still
// present as prose. Opt-in like the other live tests so `make check` stays
// hermetic: WIKISPIKE_LIVE=1 go test ./cmd/wikispike -run Live -v
func TestPipelineLiveSamplesHaveNoLinkSyntax(t *testing.T) {
	skipUnlessLive(t)
	for _, title := range sampleTitles {
		t.Run(title, func(t *testing.T) {
			markdown := liveMarkdown(t, title)
			if strings.Contains(markdown, "](") {
				t.Errorf("%q: markdown link syntax survived", title)
			}
			if strings.Contains(markdown, "http://") || strings.Contains(markdown, "https://") {
				t.Errorf("%q: bare URL syntax survived", title)
			}
			if len(strings.TrimSpace(markdown)) < 500 {
				t.Errorf("%q: suspiciously short markdown (%d bytes) — delinking may have eaten prose", title, len(markdown))
			}
		})
	}
	// Major Arcana's content table is the sharpest test: its cells are
	// almost entirely wikilinks (card names). Confirm the card names
	// themselves are still there as plain text once delinked. Reuses the
	// shared cache — no extra fetch beyond the loop above's.
	markdown := liveMarkdown(t, "Major Arcana")
	for _, name := range []string{"The Fool", "The Magician", "The High Priestess"} {
		if !strings.Contains(markdown, name) {
			t.Errorf("Major Arcana: delinked table lost %q", name)
		}
	}
}
