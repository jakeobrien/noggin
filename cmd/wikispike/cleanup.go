package main

import (
	"regexp"
	"strings"
)

// cleanupMarkdown closes the structural holes M2.2's chrome removal and
// M2.3's link unwrapping can leave behind (M2.4). It runs last in the
// pipeline, after conversion, because the holes are markdown-shaped (a bare
// "()" where a formula or lone citation used to sit inside parentheses, a
// run of blank lines where a block element vanished) rather than
// DOM-shaped: by the time an element is gone, what is left on either side is
// ordinary text nodes with no marker of what used to be between them.
//
// It deliberately does not try to repair prose that merely reads awkwardly
// because a formula was dropped (e.g. "is the limit an expression that
// arises...") — that is S6's mid-sentence-gap problem, scoped to M3.4, not
// this task. What it removes is markup residue: empty/whitespace-only
// parens or brackets, empty list items, and 3+ consecutive blank lines.
//
// Fenced code blocks (``` ... ```) are carved out and left completely
// untouched — not just by the per-line whitespace-collapse pass, but by
// every pass below (empty-parens/brackets, blank-run collapsing too):
// M3.4's display-formula path (renderMath) renders multi-line Unicode math
// as ASCII-art whose alignment depends on runs of 2+ spaces (box-drawing
// fraction bars, aligned exponents), on literal "()" (e.g. a formula
// containing "exp(1)"), and potentially on blank-ish rows within the art;
// running any whole-string pass over fence content risked corrupting it the
// same way the multi-space collapse originally did. Fence delimiter lines
// are matched by cleanupMarkdown's own fence regexp, which assumes
// htmlToMarkdown's default "```" style and well-formed (always-closed)
// fences — both true of every fence this pipeline currently produces
// (commonmark's renderBlockCode always emits a matching open/close pair);
// an unclosed trailing fence is treated as fence content through end of
// document rather than silently reverting to prose cleanup.
func cleanupMarkdown(markdown string) string {
	var out []string
	var prose []string
	inFence := false
	flushProse := func() {
		if len(prose) == 0 {
			return
		}
		out = append(out, cleanupProse(strings.Join(prose, "\n")))
		prose = nil
	}

	for _, line := range strings.Split(markdown, "\n") {
		if reFenceLine.MatchString(strings.TrimSpace(line)) {
			if !inFence {
				flushProse()
			}
			inFence = !inFence
			out = append(out, line)
			continue
		}
		if inFence {
			out = append(out, line)
			continue
		}
		prose = append(prose, line)
	}
	flushProse()

	return strings.Join(out, "\n")
}

// cleanupProse runs cleanupMarkdown's actual repair passes — empty/
// whitespace-only parens or brackets, per-line whitespace collapse, empty
// list items, and 3+ consecutive blank lines — on one non-fenced chunk of
// markdown. Split out so cleanupMarkdown can carve fenced code out first and
// call this only on the prose around it.
func cleanupProse(chunk string) string {
	// Empty/whitespace-only parens or brackets: what is left when the whole
	// parenthetical's content was a single removed element (a formula, or a
	// lone citation superscript sitting alone inside "(...)").
	chunk = reEmptyParens.ReplaceAllString(chunk, "")
	chunk = reEmptyBrackets.ReplaceAllString(chunk, "")

	lines := strings.Split(chunk, "\n")
	for i, line := range lines {
		trimmedRight := strings.TrimRight(line, " \t")
		leading := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
		rest := strings.TrimLeft(trimmedRight, " \t")

		// Collapse the double spaces and space-before-punctuation the empty
		// paren/bracket removal above can leave (e.g. "prose () continues"
		// -> "prose  continues" -> "prose continues"). Confined to a single
		// line and to text after any leading indent, so nested-list
		// indentation survives.
		rest = reMultiSpace.ReplaceAllString(rest, " ")
		rest = reSpaceBeforePunct.ReplaceAllString(rest, "$1")
		rest = strings.TrimRight(rest, " \t")

		if reEmptyListItem.MatchString(rest) {
			lines[i] = ""
			continue
		}
		lines[i] = leading + rest
	}
	chunk = strings.Join(lines, "\n")

	// Runs of 3+ blank lines: normalize any surviving run down to a single
	// blank line (standard paragraph spacing).
	chunk = reBlankRun.ReplaceAllString(chunk, "\n\n")

	return chunk
}

var (
	reEmptyParens      = regexp.MustCompile(`\(\s*\)`)
	reEmptyBrackets    = regexp.MustCompile(`\[\s*\]`)
	reMultiSpace       = regexp.MustCompile(` {2,}`)
	reSpaceBeforePunct = regexp.MustCompile(` +([.,;:!?])`)
	reEmptyListItem    = regexp.MustCompile(`^([-*+]|\d+\.)$`)
	reBlankRun         = regexp.MustCompile(`\n{3,}`)
	reFenceLine        = regexp.MustCompile("^```")
)
