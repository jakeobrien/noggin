package main

import (
	"strings"
	"testing"
)

// TestStripChromeRemovesNonArticleElements is M2.2's acceptance check on a
// minimal Parsoid-shaped fixture: navboxes, infoboxes, citation superscripts,
// reference-list sections, metadata tables and images are gone; the content
// table and article prose survive.
func TestStripChromeRemovesNonArticleElements(t *testing.T) {
	const input = `<html><body>
<section data-mw-section-id="0"><p>Intro prose with a claim<sup typeof="mw:Extension/ref" class="mw-ref"><a href="#cite_note-1">[1]</a></sup>.</p></section>
<table class="infobox"><tr><td>Born</td><td>1900</td></tr></table>
<section data-mw-section-id="1"><h2>Data</h2>
<table class="wikitable"><tr><th>Col A</th></tr><tr><td>1</td></tr></table>
</section>
<div role="navigation" class="navbox"><table class="navbox-inner"><tr><td>See also links</td></tr></table></div>
<table class="metadata"><tr><td>maintenance box</td></tr></table>
<figure typeof="mw:File/Thumb"><img src="x.png"/><figcaption>A caption</figcaption></figure>
<section data-mw-section-id="2"><h2 id="References">References</h2>
<div class="mw-references-wrap" typeof="mw:Extension/references"><ol><li>Some citation text</li></ol></div>
</section>
</body></html>`

	got, err := stripChrome(input)
	if err != nil {
		t.Fatalf("stripChrome: %v", err)
	}

	for _, want := range []string{"Intro prose with a claim", "Col A", "wikitable"} {
		if !strings.Contains(got, want) {
			t.Errorf("stripped output lost article content %q:\n%s", want, got)
		}
	}
	for _, unwanted := range []string{
		"cite_note-1", "[1]", // citation superscript
		"Born", "1900", // infobox
		"See also links", "navbox", // navbox
		"maintenance box",      // metadata table
		"img src", "A caption", // image
		"References", "Some citation text", // reference-list section
	} {
		if strings.Contains(got, unwanted) {
			t.Errorf("stripped output still contains chrome %q:\n%s", unwanted, got)
		}
	}
}

// TestStripChromeLiveSamples runs the strip step against the four S5 sample
// titles (shared live-fetch cache, M2.5) and checks M2.2's plan-stated
// acceptance shape: no "[edit]", no reference-list section, no image markup,
// no infobox/navbox table, while Major Arcana's content tables survive.
// Opt-in like the other live tests so `make check` stays hermetic:
// WIKISPIKE_LIVE=1 go test ./cmd/wikispike -run Live -v
func TestStripChromeLiveSamples(t *testing.T) {
	skipUnlessLive(t)
	for _, title := range sampleTitles {
		t.Run(title, func(t *testing.T) {
			html := fetchLiveSample(t, title)
			stripped, err := stripChrome(html)
			if err != nil {
				t.Fatalf("stripChrome(%q): %v", title, err)
			}
			markdown, err := htmlToMarkdown(stripped)
			if err != nil {
				t.Fatalf("htmlToMarkdown(%q): %v", title, err)
			}
			if strings.Contains(markdown, "[edit]") {
				t.Errorf("%q: markdown contains [edit]", title)
			}
			for _, heading := range []string{"## References", "## Notes", "# References", "# Notes"} {
				if strings.Contains(markdown, heading) {
					t.Errorf("%q: markdown still has a reference-list section (%q)", title, heading)
				}
			}
			if strings.Contains(markdown, "![") {
				t.Errorf("%q: markdown contains image markup", title)
			}
			if strings.Contains(strings.ToLower(markdown), "navbox") || strings.Contains(strings.ToLower(markdown), "infobox") {
				t.Errorf("%q: markdown leaks a navbox/infobox artifact", title)
			}
			if title != "Major Arcana" {
				return
			}
			// Major Arcana's content tables (table.wikitable) must survive
			// stripping — checked here, reusing this subtest's single fetch,
			// rather than a second fetch that risks the API's rate limit.
			if !strings.Contains(markdown, "| ---") && !strings.Contains(markdown, "|---") {
				t.Error("Major Arcana: content table did not survive stripping")
			}
		})
	}
}
