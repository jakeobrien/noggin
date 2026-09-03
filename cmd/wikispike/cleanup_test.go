package main

import (
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

// TestCleanupMarkdownLeavesFencedCodeAlignmentAlone covers the interaction
// with renderMath's display-formula path: a fenced code block holding
// termtex's multi-line ASCII-art (box-drawing fraction bars, aligned
// exponents) depends on runs of 2+ spaces for its alignment. cleanupMarkdown
// still collapses that whitespace in ordinary prose around the fence.
func TestCleanupMarkdownLeavesFencedCodeAlignmentAlone(t *testing.T) {
	const input = "Prose  with   double  spaces.\n\n```\n     n\nlim   1 + ---\n```\n\nMore  prose.\n"
	got := cleanupMarkdown(input)

	if !strings.Contains(got, "     n\nlim   1 + ---") {
		t.Errorf("expected the fenced block's alignment untouched, got:\n%s", got)
	}
	if !strings.Contains(got, "Prose with double spaces.") {
		t.Errorf("expected prose whitespace still collapsed, got:\n%s", got)
	}
	if !strings.Contains(got, "More prose.") {
		t.Errorf("expected prose whitespace still collapsed, got:\n%s", got)
	}
}

// TestCleanupMarkdownLeavesFencedParensBracketsAndBlanksAlone covers what
// TestCleanupMarkdownLeavesFencedCodeAlignmentAlone's whitespace-only check
// missed: the empty-parens/brackets and blank-run passes ran on the whole
// string, not per-line, so a naive fence guard around only the whitespace
// pass still let them corrupt fence content (e.g. a formula containing
// literal "exp()" stripped to "exp", or blank rows inside ASCII-art
// collapsed) — found in review. cleanupMarkdown now carves fenced regions
// out before running any whole-string pass, so both are preserved exactly.
func TestCleanupMarkdownLeavesFencedParensBracketsAndBlanksAlone(t *testing.T) {
	const input = "Some prose () continues.\n\n```\nexp()  [top]\n\n\n\nrow two\n```\n\nMore prose () here.\n"
	got := cleanupMarkdown(input)

	if !strings.Contains(got, "exp()  [top]\n\n\n\nrow two") {
		t.Errorf("expected fence content's parens/brackets/blank rows untouched, got:\n%s", got)
	}
	if !strings.Contains(got, "Some prose continues.") || !strings.Contains(got, "More prose here.") {
		t.Errorf("expected prose empty-parens still stripped outside the fence, got:\n%s", got)
	}
}

// TestCleanupMarkdownHandlesUnclosedFence covers the degenerate case of an
// odd number of fence delimiter lines: rather than reverting to prose
// cleanup partway through (which would then apply whitespace/parens
// collapsing to what the writer intended as code), everything from the
// unmatched opening fence to end of document is treated as fence content.
func TestCleanupMarkdownHandlesUnclosedFence(t *testing.T) {
	const input = "Prose  before.\n\n```\nfence  content ()\nProse-shaped  line after an unclosed fence.\n"
	got := cleanupMarkdown(input)

	if !strings.Contains(got, "Prose before.") {
		t.Errorf("expected prose before the fence collapsed as usual, got:\n%s", got)
	}
	if !strings.Contains(got, "fence  content ()\nProse-shaped  line after an unclosed fence.") {
		t.Errorf("expected everything after the unmatched fence left untouched, got:\n%s", got)
	}
}

// TestPipelineLiveSamplesHaveNoStructuralHoles is M2.4's acceptance check
// against the four S5 sample titles (shared live-fetch cache, M2.5): no run
// of 3+ consecutive blank lines, no empty list item, and no empty/
// whitespace-only parens or brackets left by a removal. Opt-in like the
// other live tests so `make check` stays hermetic:
// WIKISPIKE_LIVE=1 go test ./cmd/wikispike -run Live -v
func TestPipelineLiveSamplesHaveNoStructuralHoles(t *testing.T) {
	skipUnlessLive(t)
	blankRun := regexp.MustCompile(`\n{3,}`)
	emptyListItem := regexp.MustCompile(`(?m)^[-*+]\s*$|^\d+\.\s*$`)
	for _, title := range sampleTitles {
		t.Run(title, func(t *testing.T) {
			markdown := liveMarkdown(t, title)
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
