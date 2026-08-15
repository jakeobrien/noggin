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
func cleanupMarkdown(markdown string) string {
	// Empty/whitespace-only parens or brackets: what is left when the whole
	// parenthetical's content was a single removed element (a formula, or a
	// lone citation superscript sitting alone inside "(...)").
	markdown = reEmptyParens.ReplaceAllString(markdown, "")
	markdown = reEmptyBrackets.ReplaceAllString(markdown, "")

	lines := strings.Split(markdown, "\n")
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
	markdown = strings.Join(lines, "\n")

	// Runs of 3+ blank lines: normalize any surviving run down to a single
	// blank line (standard paragraph spacing).
	markdown = reBlankRun.ReplaceAllString(markdown, "\n\n")

	return markdown
}

var (
	reEmptyParens      = regexp.MustCompile(`\(\s*\)`)
	reEmptyBrackets    = regexp.MustCompile(`\[\s*\]`)
	reMultiSpace       = regexp.MustCompile(` {2,}`)
	reSpaceBeforePunct = regexp.MustCompile(` +([.,;:!?])`)
	reEmptyListItem    = regexp.MustCompile(`^([-*+]|\d+\.)$`)
	reBlankRun         = regexp.MustCompile(`\n{3,}`)
)
