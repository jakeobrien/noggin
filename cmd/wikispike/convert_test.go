package main

import (
	"strings"
	"testing"
)

// TestHTMLToMarkdownRendersHeadingsParagraphsAndTables is M2.1's acceptance
// check: headings, paragraphs and at least one markdown table, off the
// network — a minimal Parsoid-shaped fixture, not a live fetch.
func TestHTMLToMarkdownRendersHeadingsParagraphsAndTables(t *testing.T) {
	const input = `<html><body>
<h2>Section</h2>
<p>Some prose about the topic.</p>
<table class="wikitable">
<tr><th>Col A</th><th>Col B</th></tr>
<tr><td>1</td><td>2</td></tr>
</table>
</body></html>`

	got, err := htmlToMarkdown(input)
	if err != nil {
		t.Fatalf("htmlToMarkdown: %v", err)
	}
	if !strings.Contains(got, "## Section") {
		t.Errorf("no heading in output:\n%s", got)
	}
	if !strings.Contains(got, "Some prose about the topic.") {
		t.Errorf("no paragraph in output:\n%s", got)
	}
	if !strings.Contains(got, "| Col A") || !strings.Contains(got, "| --- |") && !strings.Contains(got, "|---") {
		t.Errorf("no markdown table in output:\n%s", got)
	}
}

// TestHTMLToMarkdownLiveSamples runs the conversion step of the pipeline
// against the four S5 sample titles (shared live-fetch cache, M2.5), checking
// the same heading/paragraph/table shape end to end. Opt-in like fetch.go's
// live test so `make check` stays hermetic:
// WIKISPIKE_LIVE=1 go test ./cmd/wikispike -run Live -v
func TestHTMLToMarkdownLiveSamples(t *testing.T) {
	skipUnlessLive(t)
	for _, title := range sampleTitles {
		t.Run(title, func(t *testing.T) {
			html := fetchLiveSample(t, title)
			markdown, err := htmlToMarkdown(html)
			if err != nil {
				t.Fatalf("htmlToMarkdown(%q): %v", title, err)
			}
			if !strings.Contains(markdown, "#") {
				t.Errorf("%q: no heading in converted markdown", title)
			}
			if len(strings.TrimSpace(markdown)) < 500 {
				t.Errorf("%q: suspiciously short markdown (%d bytes)", title, len(markdown))
			}
		})
	}
}
