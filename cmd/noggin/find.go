package main

import (
	"strings"
	"unicode"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// find is the viewer's S8 search state: the query prompt, the matches it
// currently produces and which one is indicated. It is the viewer's alone
// (D4's viewport-scrolling-only shape extends naturally to this: search
// moves the same viewport, it does not add a cursor).
//
// A match's position is computed once, directly, rather than through
// viewport.Model's own SetHighlights/HighlightNext/HighlightPrevious
// (findings/viewport-search-ansi-offsets.md): those walk byte offsets against
// the *styled* content viewport.Model.GetContent() returns, and that walk
// silently drifts once any ANSI escape precedes the match, which every
// Glamour-styled document does. EnsureVisible and StyleLineFunc, the two
// primitives that actually move and mark the viewport, take a line number
// directly and are not implicated \u2014 findMatches below computes that line
// number correctly instead, against the escape-stripped content, and this
// file drives EnsureVisible/StyleLineFunc with it.
type find struct {
	// editing is true while the query has the keyboard: every key but
	// Cancel and Activate (see model.handleFindInput) types into it rather
	// than acting on its usual binding.
	editing bool
	input   textinput.Model

	// matches are every occurrence of input.Value() in the rendered
	// document, in on-screen order. current indexes the one EnsureVisible
	// and StyleLineFunc are pointed at; -1 when there are none (no query,
	// or a query with no hits).
	matches []findMatch
	current int
}

// findMatch is one occurrence of a query, located in the rendered content's
// own line space \u2014 the viewport's SoftWrap is off (viewer.go's render already
// wraps to the pane width), so this is directly the viewport's line index,
// not a further wrapping of it.
type findMatch struct {
	line, colStart, colEnd int
}

// newFind is the no-search-in-progress state: no query, no matches, no
// keyboard capture.
func newFind() find {
	in := textinput.New()
	in.Prompt = "/"
	return find{current: -1, input: in}
}

// startFind begins a new query, focusing the input (S8's "/-style"). Any
// previous query is discarded \u2014 like vim's "/", a fresh search starts empty
// rather than reusing the last one.
func (v viewer) startFind() (viewer, tea.Cmd) {
	v.find = newFind()
	v.find.editing = true
	cmd := v.find.input.Focus()
	return v, cmd
}

// updateFindQuery routes one keypress into the query input, then recomputes
// matches and re-applies the current one \u2014 the "incremental" of "incremental
// search" (S8).
func (v viewer) updateFindQuery(msg tea.KeyPressMsg) (viewer, tea.Cmd) {
	next, cmd := v.find.input.Update(msg)
	v.find.input = next
	v.find.matches = findMatches(v.vp.GetContent(), v.find.input.Value())
	v.find.current = nearestMatch(v.find.matches, v.vp.YOffset())
	return v.applyCurrentMatch(), cmd
}

// commitFind ends editing but keeps the matches and the current one
// indicated, so FindNext/FindPrevious keep working after Enter.
func (v viewer) commitFind() viewer {
	v.find.editing = false
	v.find.input.Blur()
	return v
}

// cancelFind dismisses the search entirely: no query, no matches, no
// indication. The viewport itself is untouched \u2014 S8's "dismissal that
// leaves the view where it is" \u2014 whatever EnsureVisible last did to it stands.
func (v viewer) cancelFind() viewer {
	v.find = newFind()
	v.vp.StyleLineFunc = nil
	return v
}

// findEditing reports whether the query currently has the keyboard.
func (v viewer) findEditing() bool { return v.find.editing }

// hasFindMatches reports whether FindNext/FindPrevious have anything to do.
func (v viewer) hasFindMatches() bool { return len(v.find.matches) > 0 }

// findNext and findPrevious step the indicated match, wrapping past either
// end (S8).
func (v viewer) findNext() viewer {
	if len(v.find.matches) == 0 {
		return v
	}
	v.find.current = (v.find.current + 1) % len(v.find.matches)
	return v.applyCurrentMatch()
}

func (v viewer) findPrevious() viewer {
	if len(v.find.matches) == 0 {
		return v
	}
	v.find.current = (v.find.current - 1 + len(v.find.matches)) % len(v.find.matches)
	return v.applyCurrentMatch()
}

// applyCurrentMatch scrolls the current match on screen (EnsureVisible) and
// marks its line (StyleLineFunc) \u2014 S8's "the view scrolls so the match's
// line is on screen and that line is indicated." Line-level indication, not
// the character-exact kind S8 calls desirable-but-not-required: the whole
// line is reverse-videoed rather than just the matched span, which sidesteps
// viewport.Model's own column highlighting (SetHighlights) and the ANSI-offset
// drift it would reintroduce (see the find type's doc comment).
func (v viewer) applyCurrentMatch() viewer {
	if v.find.current < 0 || v.find.current >= len(v.find.matches) {
		v.vp.StyleLineFunc = nil
		return v
	}
	m := v.find.matches[v.find.current]
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

// findView renders the query prompt at width, for the footer to show in
// place of key hints while editing (model.renderFooter).
func (v viewer) findView(width int) string {
	v.find.input.SetWidth(width)
	return v.find.input.View()
}

// findMatches locates every occurrence of query in content, in line order.
// content is the viewport's own GetContent() \u2014 rendered, ANSI-styled markdown
// \u2014 but matching happens against ansi.Strip(content): searching the styled
// bytes directly risks a query matching inside an SGR escape by coincidence
// (an escape sequence ends in the literal byte 'm'), and would in any case
// produce the same byte-offset-drift problem described on the find type.
//
// Matching is a plain substring search, not a regular expression: S8 asks
// for incremental search over arbitrary typed text, and a query containing
// regex metacharacters (parentheses are common in prose) must still match
// literally.
//
// Case sensitivity is smart-case per S8: an all-lowercase query matches
// either case, any uppercase in it makes the match case-sensitive.
func findMatches(content, query string) []findMatch {
	if query == "" {
		return nil
	}
	stripped := ansi.Strip(content)
	smartCase := query == strings.ToLower(query)

	// The case-insensitive path searches a folded copy of stripped rather
	// than stripped itself, so found byte offsets land in the *folded*
	// string's byte space, not stripped's -- folding is not always
	// byte-length-preserving (the Kelvin sign U+212A, three bytes, folds to
	// the single-byte 'k'), so treating those offsets as if they were
	// stripped's own would misattribute a match once such a rune precedes it
	// on the same line (T29). toOrig maps each folded byte offset back to
	// stripped's, so start/end below are always in stripped's space
	// regardless of which path ran.
	haystack, needle := stripped, query
	var toOrig []int
	if smartCase {
		haystack, toOrig = foldLower(stripped)
		needle = strings.ToLower(query)
	}

	var starts, ends []int
	for pos := 0; ; {
		idx := strings.Index(haystack[pos:], needle)
		if idx < 0 {
			break
		}
		hStart := pos + idx
		hEnd := hStart + len(needle)
		start, end := hStart, hEnd
		if smartCase {
			start, end = toOrig[hStart], toOrig[hEnd]
		}
		starts = append(starts, start)
		ends = append(ends, end)
		pos = hEnd // non-overlapping, in haystack's own byte space
	}
	if len(starts) == 0 {
		return nil
	}

	// One linear pass over stripped, converting each (already-sorted) byte
	// offset into a line and column by counting newlines up to it \u2014 done
	// once here rather than per match, and done against stripped itself (via
	// start/end above, already translated back from haystack's space when
	// they differ), which is what keeps this correct where viewport.Model's
	// own SetHighlights is not.
	matches := make([]findMatch, len(starts))
	line, lineStart, pos := 0, 0, 0
	for i, start := range starts {
		for pos < start {
			if stripped[pos] == '\n' {
				line++
				lineStart = pos + 1
			}
			pos++
		}
		matches[i] = findMatch{
			line:     line,
			colStart: start - lineStart,
			colEnd:   ends[i] - lineStart,
		}
	}
	return matches
}

// foldLower lower-cases s rune by rune for findMatches' smart-case path,
// returning alongside it a mapping (toOrig) from each byte offset in the
// folded result back to the byte offset in s it came from, with a trailing
// sentinel entry at len(folded) mapping to len(s) so a match ending at the
// folded string's end still resolves. See findMatches' own comment for why
// this mapping is necessary rather than optional (T29).
func foldLower(s string) (folded string, toOrig []int) {
	var b strings.Builder
	toOrig = make([]int, 0, len(s)+1)
	for i, r := range s {
		before := b.Len()
		b.WriteRune(unicode.ToLower(r))
		for n := b.Len() - before; n > 0; n-- {
			toOrig = append(toOrig, i)
		}
	}
	toOrig = append(toOrig, len(s))
	return b.String(), toOrig
}

// nearestMatch picks the first match at or after fromLine \u2014 incremental
// search moves forward as you type, like vim's \u2014 wrapping to the first
// match overall if none is. Returns -1 for no matches at all.
func nearestMatch(matches []findMatch, fromLine int) int {
	for i, m := range matches {
		if m.line >= fromLine {
			return i
		}
	}
	if len(matches) > 0 {
		return 0
	}
	return -1
}
