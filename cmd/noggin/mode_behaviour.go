package main

import "charm.land/lipgloss/v2"

// modeBehaviour is what differs between the two Render Modes once a
// document is shown: how Find locates matches, how the current one is
// indicated, and how a Grep Result's location maps onto a match. render
// resolves the mode a document is shown in once (viewer.docMode) and
// renderMode.behaviour is the only place that turns it into behaviour, so
// callers never branch on the mode themselves.
type modeBehaviour interface {
	// find is every occurrence of query in the open document.
	find(v viewer, query string) []findMatch
	// indicate scrolls m (one of v's matches) on screen and marks it.
	indicate(v viewer, m findMatch) viewer
	// clearIndication removes any match indication.
	clearIndication(v viewer) viewer
	// landGrep makes the match at ripgrep's 1-based line and byte column col
	// the current one in v's committed find, or records the location when it
	// cannot be resolved here and a switch of mode could still use it.
	landGrep(v viewer, line, col int) viewer
	// carriedLanding is the index, in matches found in this mode, of the
	// location a Grep landing recorded in the other mode.
	carriedLanding(src []byte, matches []findMatch, l grepLanding) (int, bool)
	// headMatches counts the matches within the document's frontmatter, given
	// its extent in source lines and in rendered lines (frontmatterSpan).
	headMatches(matches []findMatch, srcLines, renderedLines int) int
}

// behaviour is r's modeBehaviour.
func (r renderMode) behaviour() modeBehaviour {
	if r == modeText {
		return textBehaviour{}
	}
	return markdownBehaviour{}
}

// markdownBehaviour: matches are found in the rendered (Glamour) content and
// the whole line of the current one is indicated.
type markdownBehaviour struct{}

func (markdownBehaviour) find(v viewer, query string) []findMatch {
	return findMatches(v.vp.GetContent(), query)
}

func (markdownBehaviour) indicate(v viewer, m findMatch) viewer {
	v.vp.EnsureVisible(m.line, m.colStart, m.colEnd)
	line := m.line
	v.vp.StyleLineFunc = func(i int) lipgloss.Style {
		if i == line {
			return lipgloss.NewStyle().Reverse(true)
		}
		return lipgloss.NewStyle()
	}
	return v
}

func (markdownBehaviour) clearIndication(v viewer) viewer {
	v.vp.StyleLineFunc = nil
	return v
}

// landGrep is the occurrence-ordinal bridge (ADR-0002); its degrade path
// remembers ripgrep's own location so a switch to Text can land exactly.
func (markdownBehaviour) landGrep(v viewer, line, col int) viewer {
	if ordinal, ok := resolveGrepLanding(v.doc.src, v.vp.GetContent(), v.findState.input.Value(), line, v.vp.Width()); ok {
		v.findState.current = ordinal
		return v.applyCurrentMatch()
	}
	v.findState.landing = &grepLanding{line: line, col: col}
	return v
}

// carriedLanding: only Text lands exactly from a recorded location.
func (markdownBehaviour) carriedLanding([]byte, []findMatch, grepLanding) (int, bool) {
	return 0, false
}

func (markdownBehaviour) headMatches(matches []findMatch, _, renderedLines int) int {
	n := 0
	for _, m := range matches {
		if m.line < renderedLines {
			n++
		}
	}
	return n
}

// textBehaviour: matches are found in the source lines (textLayout.find) and
// only the matched span is drawn.
type textBehaviour struct{}

func (textBehaviour) find(v viewer, query string) []findMatch {
	return v.text.find(query)
}

func (textBehaviour) indicate(v viewer, m findMatch) viewer {
	// Text has no ANSI in the body, so the matched span alone is drawn
	// (withTextMatch); the rows are already wrapped, so no horizontal scroll
	// is wanted.
	v.vp.EnsureVisible(m.line, 0, 0)
	if bottom := v.vp.YOffset() + v.vp.Height(); m.endLine >= bottom {
		// Show the rest of a match that wraps, but never at the cost of its
		// first row.
		v.vp.SetYOffset(min(m.line, m.endLine-v.vp.Height()+1))
	}
	return v.withTextMatch(&textSpan{line: m.src, start: m.colStart, end: m.colEnd})
}

func (textBehaviour) clearIndication(v viewer) viewer {
	v.vp.StyleLineFunc = nil
	return v.withTextMatch(nil)
}

// landGrep lands on ripgrep's own line and column, no occurrence-ordinal guess.
func (textBehaviour) landGrep(v viewer, line, col int) viewer {
	if ordinal, ok := textGrepLanding(v.doc.src, v.findState.matches, line, col); ok {
		v.findState.current = ordinal
		return v.applyCurrentMatch()
	}
	return v
}

func (textBehaviour) carriedLanding(src []byte, matches []findMatch, l grepLanding) (int, bool) {
	return textGrepLanding(src, matches, l.line, l.col)
}

func (textBehaviour) headMatches(matches []findMatch, srcLines, _ int) int {
	n := 0
	for _, m := range matches {
		if m.src < srcLines {
			n++
		}
	}
	return n
}
