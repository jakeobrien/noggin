package main

import (
	"regexp"
	"strings"
	"testing"
)

// TestStripChromeRemovesNonArticleElements is M2.2's acceptance check on a
// minimal Parsoid-shaped fixture: navboxes, sidebars, infoboxes, citation
// superscripts, reference-list sections, metadata tables and images are
// gone; the content table and article prose survive.
func TestStripChromeRemovesNonArticleElements(t *testing.T) {
	const input = `<html><body>
<section data-mw-section-id="0"><p>Intro prose with a claim<sup typeof="mw:Extension/ref" class="mw-ref"><a href="#cite_note-1">[1]</a></sup>.</p>
<p>A dubious claim<sup class="noprint Inline-Template Template-Fact"><span typeof="mw:Entity">[</span><a href="./Wikipedia:Citation_needed">citation needed</a><span typeof="mw:Entity">]</span></sup> and a page-referenced one<sup class="reference nowrap" typeof="mw:Transclusion" data-mw='{"parts":[{"template":{"target":{"wt":"rp"}}}]}'><span>: 25</span></sup>.</p></section>
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

	for _, want := range []string{"Intro prose with a claim", "A dubious claim", "and a page-referenced one", "Col A", "wikitable", "Real section prose follows the sidebar"} {
		if !strings.Contains(got, want) {
			t.Errorf("stripped output lost article content %q:\n%s", want, got)
		}
	}
	for _, unwanted := range []string{
		"cite_note-1", "[1]", // citation superscript
		"citation needed", ": 25", // maintenance-template / page-locator sup (TRIAGE T1, M3.9)
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

// TestSisterProjectInlineLineIgnoresMidSentenceMentions is M3.7's safety-
// property test for sisterProjectInlineLine (TRIAGE T3): ordinary prose that
// merely mentions a sister project in passing, rather than rendering one of
// the whole-line sister-project-inline templates, must survive stripChrome
// untouched. Only the removal path (Major Arcana's Commons category-inline)
// had a test before this; the survival half of the same regex's contract
// (anchored `^...$`, not a substring search) was never independently
// exercised.
func TestSisterProjectInlineLineIgnoresMidSentenceMentions(t *testing.T) {
	const input = `<html><body><section>
<p>The photograph was later donated to Wikimedia Commons by the photographer.</p>
<li>Wikiquote is a sister project that collects quotations, not covered here.</li>
<p>See also the Wiktionary entry, which the author consulted while drafting this passage.</p>
</section></body></html>`

	got, err := stripChrome(input)
	if err != nil {
		t.Fatalf("stripChrome: %v", err)
	}

	for _, want := range []string{
		"later donated to Wikimedia Commons by the photographer",
		"Wikiquote is a sister project that collects quotations",
		"the Wiktionary entry, which the author consulted",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("stripChrome removed a mid-sentence sister-project mention %q:\n%s", want, got)
		}
	}
}

// rpRemnant matches the {{rp|<page>}} page-locator template's leaked
// output (e.g. ": 25 ") once its surrounding citation-superscript brackets
// are gone — a bare colon-space-digits run with no digit immediately
// before or after (so it doesn't false-match an ordinary "12:30" or similar
// in article prose), used by TestStripChromeLiveSamples's M3.9 regression
// guard.
var rpRemnant = regexp.MustCompile(`[^0-9]: [0-9]+ `)

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
			// Regression guards for M3.9 (TRIAGE T1): the maintenance-template
			// family and the {{rp}} page-locator template used to leak as
			// literal bracketed (or, for {{rp}}, unbracketed) text since
			// neither carries the typeof="mw:Extension/ref" the pre-M3.9
			// selector matched — found live in these exact three samples
			// during M4.1's fresh regen.
			if title == "Major Arcana" || title == "List of programming languages by type" {
				for _, leak := range []string{"citation needed", "clarification needed", "dubious", "disputed"} {
					if strings.Contains(markdown, leak) {
						t.Errorf("%q: markdown leaks a maintenance-template marker (%q)", title, leak)
					}
				}
			}
			if title == "Why is there anything at all?" && rpRemnant.MatchString(markdown) {
				t.Errorf("%q: markdown leaks a {{rp}} page-locator remnant", title)
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
			// Regression guard for M3.7 (TRIAGE T1): removing that inline
			// link used to leave "## External links" as a dangling,
			// content-free heading through EOF. This subtest only runs
			// stripChrome+htmlToMarkdown (not the full pipeline, so not
			// cleanupMarkdown's dropEmptyHeadings), so the check is against
			// the full pipeline's output instead, reusing this subtest's
			// cached fetch.
			if full := liveMarkdown(t, title); dropEmptyHeadings(full) != full {
				t.Errorf("%q: full pipeline still leaves a heading emptied by stripping:\n%s", title, full)
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
