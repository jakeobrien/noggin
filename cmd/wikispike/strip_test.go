package main

import (
	"strings"
	"testing"
)

// TestStripChromeRemovesNonArticleElements is M2.2's acceptance check on a
// minimal Parsoid-shaped fixture: navboxes, sidebars, infoboxes, citation
// superscripts, reference-list sections, metadata tables and images are
// gone; the content table and article prose survive.
func TestStripChromeRemovesNonArticleElements(t *testing.T) {
	const input = `<html><body>
<section data-mw-section-id="0"><p>Intro prose with a claim<sup typeof="mw:Extension/ref" class="mw-ref"><a href="#cite_note-1">[1]</a></sup>.</p></section>
<table class="infobox"><tr><td>Born</td><td>1900</td></tr></table>
<section data-mw-section-id="1"><h2>Data</h2>
<table class="sidebar nomobile nowraplinks hlist"><tr><td>Part of a series on widgets</td></tr></table>
<p>Real section prose follows the sidebar.</p>
<table class="wikitable"><tr><th>Col A</th></tr><tr><td>1</td></tr></table>
</section>
<div role="navigation" class="navbox"><table class="navbox-inner"><tr><td>See also links</td></tr></table></div>
<table class="metadata"><tr><td>maintenance box</td></tr></table>
<figure typeof="mw:File/Thumb"><img src="x.png"/><figcaption>A caption</figcaption></figure>
<div class="portal-bar noprint"><ul><li><a href="#">Philosophy</a></li><li><a href="#">Physics</a></li></ul></div>
<div class="sistersitebox"><a href="#"><img src="commons.png"/>Wikimedia Commons has media related to Widgets.</a></div>
<section data-mw-section-id="3"><h2 id="External_links">External links</h2>
<ul><li id="mwA3c"><style data-mw-deduplicate="TemplateStyles:r1" typeof="mw:Extension/templatestyles mw:Transclusion" about="#mwt99">.sister-inline-image{display:inline}</style><span class="noviewer sister-inline-image" typeof="mw:File" about="#mwt99"><a href="./File:Commons-logo.svg" class="mw-file-description"><img src="commons2.png" alt="Wikimedia Commons logo"/></a></span><span about="#mwt99"> Media related to </span><a rel="mw:WikiLink/Interwiki" href="https://commons.wikimedia.org/wiki/Category:Widgets" about="#mwt99">Widgets</a><span about="#mwt99"> at Wikimedia Commons</span></li></ul>
</section>
<section data-mw-section-id="2"><h2 id="References">References</h2>
<div class="mw-references-wrap" typeof="mw:Extension/references"><ol><li>Some citation text</li></ol></div>
</section>
</body></html>`

	got, err := stripChrome(input)
	if err != nil {
		t.Fatalf("stripChrome: %v", err)
	}

	for _, want := range []string{"Intro prose with a claim", "Col A", "wikitable", "Real section prose follows the sidebar"} {
		if !strings.Contains(got, want) {
			t.Errorf("stripped output lost article content %q:\n%s", want, got)
		}
	}
	for _, unwanted := range []string{
		"cite_note-1", "[1]", // citation superscript
		"Born", "1900", // infobox
		"See also links", "navbox", // navbox
		"Part of a series on widgets", // sidebar, sitting between the heading and this section's real prose, same placement as the live bug
		"maintenance box",             // metadata table
		"img src", "A caption",        // image
		"References", "Some citation text", // reference-list section
		"Philosophy", "Physics", // portal bar (M3.3a)
		"Wikimedia Commons has media related to Widgets", // sister-project side-box (M3.3a)
		"Media related to", "at Wikimedia Commons", // inline sister-project link, e.g. Commons category-inline (M3.3b)
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
			// Regression guard for the live sidebar bug the sibling unit
			// fixture reproduces: `E (mathematical constant)`'s History
			// section used to open with this exact sidebar template's caption
			// text ("Part of a series of articles on") ahead of its real prose.
			if title == "E (mathematical constant)" && strings.Contains(markdown, "Part of a series of articles on") {
				t.Errorf("%q: markdown leaks the sidebar template", title)
			}
			// Regression guards for M3.3a's portal-bar/sister-project fix:
			// live samples that carried these boxes used to leak their link
			// text (or, post-M3.5, an [Image: ...] placeholder for their
			// icon) ahead of/after the article's real content.
			if title == "Why is there anything at all?" {
				if strings.Contains(markdown, "Portals:") {
					t.Errorf("%q: markdown leaks the portal-bar template", title)
				}
				if strings.Contains(markdown, "[Image: icon]") {
					t.Errorf("%q: markdown leaks a portal-bar icon placeholder", title)
				}
			}
			if title == "E (mathematical constant)" && strings.Contains(markdown, "Wikimedia Commons has media related to") {
				t.Errorf("%q: markdown leaks the sister-project side-box", title)
			}
			if title != "Major Arcana" {
				return
			}
			// Regression guard for M3.3b (TRIAGE T10/T12): this is the exact
			// sample PLAN.md cited as motivating the fix-up — a Commons
			// category-inline template rendered as a bare <li> in External
			// links, which M3.3a's div.portal-bar/div.sistersitebox selectors
			// never matched. T12 is folded in here: this is the first time
			// this live regression guard has actually exercised Major Arcana.
			if strings.Contains(markdown, "Media related to") || strings.Contains(markdown, "at Wikimedia Commons") {
				t.Errorf("%q: markdown leaks the inline sister-project link", title)
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
