package main

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// TestCleanupMarkdownClosesRemovalHoles is M2.4's acceptance check on a
// fixture built to reproduce the holes M2.2/M2.3 leave: a formula and a
// lone citation each sitting alone inside parentheses (both stripped by
// stripChrome, leaving bare "()"), a list item that is only an image (its
// whole <li> is dropped, not left as an empty bullet), 3+ blank lines from a
// vanished block, and a stray empty bracket.
func TestCleanupMarkdownClosesRemovalHoles(t *testing.T) {
	const input = `Some prose () continues here.

A lone citation () trailing.

- Real item

Empty ref marker [] here.



Paragraph after a blank-line pileup.

-
`
	got := cleanupMarkdown(input)

	for _, want := range []string{
		"Some prose continues here.",
		"A lone citation trailing.",
		"- Real item",
		"Empty ref marker here.",
		"Paragraph after a blank-line pileup.",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("cleanup lost or mangled %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "()") {
		t.Errorf("empty parens survived cleanup:\n%s", got)
	}
	if strings.Contains(got, "[]") {
		t.Errorf("empty brackets survived cleanup:\n%s", got)
	}
	if m := regexp.MustCompile(`\n{3,}`).FindString(got); m != "" {
		t.Errorf("a run of 3+ blank lines survived cleanup:\n%s", got)
	}
	if regexp.MustCompile(`(?m)^[-*+]\s*$`).MatchString(got) {
		t.Errorf("an empty list item survived cleanup:\n%s", got)
	}
}

// TestCleanupMarkdownFullPipelineFixture runs the whole HTML->markdown
// pipeline (strip, delink, convert, clean up) on a fixture shaped like the
// adversarial cases found in real articles: a formula and a lone citation
// each alone inside parens, and a citation-only table cell. Confirms the
// holes are closed end to end, not just by cleanupMarkdown in isolation.
func TestCleanupMarkdownFullPipelineFixture(t *testing.T) {
	const input = `<html><body><section>
<p>Some prose (<span typeof="mw:Extension/math">x^2</span>) continues here.</p>
<p>A lone citation (<sup typeof="mw:Extension/ref" class="mw-ref"><a href="#cite_note-1">[1]</a></sup>) trailing.</p>
<table class="wikitable"><tr><th>Name</th><th>Note</th></tr>
<tr><td>Item A</td><td><sup typeof="mw:Extension/ref" class="mw-ref"><a href="#cite_note-2">[2]</a></sup></td></tr></table>
</section></body></html>`

	stripped, err := stripChrome(input)
	if err != nil {
		t.Fatalf("stripChrome: %v", err)
	}
	delinked, err := stripLinks(stripped)
	if err != nil {
		t.Fatalf("stripLinks: %v", err)
	}
	markdown, err := htmlToMarkdown(delinked)
	if err != nil {
		t.Fatalf("htmlToMarkdown: %v", err)
	}
	got := cleanupMarkdown(markdown)

	if strings.Contains(got, "()") {
		t.Errorf("empty parens survived the full pipeline:\n%s", got)
	}
	for _, want := range []string{"Some prose continues here.", "A lone citation trailing.", "Item A"} {
		if !strings.Contains(got, want) {
			t.Errorf("pipeline output lost %q:\n%s", want, got)
		}
	}
}

// TestPipelineLiveSamplesHaveNoStructuralHoles is M2.4's acceptance check
// against the four S5 sample titles fetched live: no run of 3+ consecutive
// blank lines, no empty list item, and no empty/whitespace-only parens or
// brackets left by a removal. Opt-in like the other live tests so `make
// check` stays hermetic: WIKISPIKE_LIVE=1 go test ./cmd/wikispike -run Live -v
func TestPipelineLiveSamplesHaveNoStructuralHoles(t *testing.T) {
	if os.Getenv("WIKISPIKE_LIVE") == "" {
		t.Skip("set WIKISPIKE_LIVE=1 to hit en.wikipedia.org")
	}
	blankRun := regexp.MustCompile(`\n{3,}`)
	emptyListItem := regexp.MustCompile(`(?m)^[-*+]\s*$|^\d+\.\s*$`)
	for _, title := range sampleTitles {
		t.Run(title, func(t *testing.T) {
			markdown, err := articleMarkdown(title)
			if err != nil {
				t.Fatalf("articleMarkdown(%q): %v", title, err)
			}
			if m := blankRun.FindString(markdown); m != "" {
				t.Errorf("%q: a run of 3+ consecutive blank lines survived", title)
			}
			if emptyListItem.MatchString(markdown) {
				t.Errorf("%q: an empty list item survived", title)
			}
			if strings.Contains(markdown, "()") {
				t.Errorf("%q: empty parens survived", title)
			}
			if strings.Contains(markdown, "[]") {
				t.Errorf("%q: empty brackets survived", title)
			}
		})
	}
}
