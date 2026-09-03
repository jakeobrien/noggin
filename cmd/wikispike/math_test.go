package main

import (
	"strings"
	"testing"
)

// mathSpan builds a minimal Parsoid-shaped math span carrying the given TeX
// source in its data-mw JSON, matching the real shape (body.extsrc).
func mathSpan(extsrc string) string {
	return `<span typeof="mw:Extension/math" data-mw='{"name":"math","body":{"extsrc":"` + extsrc + `"}}'></span>`
}

// TestRenderMathUsesUnicodeWhenSingleLine covers M3.4's preferred path: a
// formula whose termtex rendering fits on one line (a lone Greek letter
// here) is substituted as that Unicode text, not its raw TeX source.
func TestRenderMathUsesUnicodeWhenSingleLine(t *testing.T) {
	input := `<p>denoted ` + mathSpan(`\\gamma`) + ` here</p>`

	got, err := renderMath(input)
	if err != nil {
		t.Fatalf("renderMath: %v", err)
	}
	if !strings.Contains(got, "γ") {
		t.Errorf("expected the Unicode Greek letter in output, got:\n%s", got)
	}
	if strings.Contains(got, `\gamma`) {
		t.Errorf("did not expect the raw TeX source to survive when Unicode was usable:\n%s", got)
	}
}

// TestRenderMathFallsBackToRawSourceWhenMultiLine covers M3.4's fallback
// path: a fraction renders as a multi-row box-drawing block in termtex,
// which would tear a sentence apart, so the raw TeX source is kept instead.
func TestRenderMathFallsBackToRawSourceWhenMultiLine(t *testing.T) {
	input := `<p>given by ` + mathSpan(`\\frac{1}{2}`) + ` here</p>`

	got, err := renderMath(input)
	if err != nil {
		t.Fatalf("renderMath: %v", err)
	}
	if !strings.Contains(got, `\frac{1}{2}`) {
		t.Errorf("expected the raw TeX source as fallback, got:\n%s", got)
	}
}

// TestRenderMathLeavesUnrecoverableSpanForStripChrome covers the edge case
// where a math span has no usable source at all (missing data-mw): renderMath
// makes no substitution, so it is left for stripChrome's own math selector
// to remove, matching pre-M3.4 behaviour.
func TestRenderMathLeavesUnrecoverableSpanForStripChrome(t *testing.T) {
	input := `<p>prose (<span typeof="mw:Extension/math">unreadable render</span>) continues</p>`

	got, err := renderMath(input)
	if err != nil {
		t.Fatalf("renderMath: %v", err)
	}
	if !strings.Contains(got, `typeof="mw:Extension/math"`) {
		t.Errorf("expected the untouched span to survive renderMath for stripChrome to remove, got:\n%s", got)
	}
}

// TestArticleMarkdownFromHTMLKeepsMathAsProse is the full-pipeline check:
// a formula that used to vanish without a trace (pre-D104) now survives as
// readable text and leaves no mid-sentence gap either way (Unicode or raw
// source), through fetch's real successor stages (strip, delink, convert,
// clean up).
func TestArticleMarkdownFromHTMLKeepsMathAsProse(t *testing.T) {
	input := `<html><body><section>
<p>The constant, denoted ` + mathSpan(`\\gamma`) + `, and the identity ` + mathSpan(`\\frac{1}{2}`) + ` continues the sentence.</p>
</section></body></html>`

	got, err := articleMarkdownFromHTML(input)
	if err != nil {
		t.Fatalf("articleMarkdownFromHTML: %v", err)
	}
	if !strings.Contains(got, "γ") {
		t.Errorf("expected the Unicode-rendered formula in the final markdown, got:\n%s", got)
	}
	if !strings.Contains(got, `1`) || !strings.Contains(got, `2`) {
		t.Errorf("expected the raw-source fallback formula's digits in the final markdown, got:\n%s", got)
	}
	if !strings.Contains(got, "continues the sentence") {
		t.Errorf("expected the sentence to survive intact around both formulae, got:\n%s", got)
	}
	if strings.Contains(got, "()") {
		t.Errorf("expected no empty-parens gap, got:\n%s", got)
	}
}
