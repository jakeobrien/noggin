package main

import (
	"bytes"
	"fmt"
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"noggin/internal/tree"
)

// renderMode is how a Document Viewer shows a document (D140): Markdown
// (rendered by Glamour) or Text (raw source, line-numbered). One per viewer;
// the zero value is Markdown, which is what every viewer starts on.
type renderMode int

const (
	modeMarkdown renderMode = iota
	modeText
)

func (r renderMode) String() string {
	if r == modeText {
		return "Text"
	}
	return "Markdown"
}

// toggled is the other mode.
func (r renderMode) toggled() renderMode {
	if r == modeText {
		return modeMarkdown
	}
	return modeText
}

// isNotMarkdownFile reports whether d is a Text file in the viewer's sense
// (D140): an open file that is not a Markdown file. It is always shown as text,
// whatever the viewer's mode. An article is never one: its converted markdown
// can be shown either way. (Whether a file is binary is tree.IsNotBinary's
// question, not this one's.)
func isNotMarkdownFile(d document) bool {
	return d.isOpen() && !d.isArticle() && !tree.IsMarkdown(filepath.Base(d.path))
}

// modeFor is the mode d is actually shown in by this viewer: its own mode,
// except that a Text file is always shown as Text. render resolves it once for
// the Active Document (viewer.docMode); every mode-specific behaviour of Find
// hangs off that (modeBehaviour), not off a fresh check.
func (v viewer) modeFor(d document) renderMode {
	if isNotMarkdownFile(d) {
		return modeText
	}
	return v.mode
}

// shownDoc is the document the pane shows: the Peeked one while Peeking, the
// Active Document otherwise.
func (v viewer) shownDoc() document {
	if v.peeking() {
		return v.peek.doc
	}
	return v.doc
}

// toggleMode flips the viewer's render mode and re-renders, carrying over what
// the reader was looking at (D140). A find (in progress or committed) keeps its
// query, is re-run in the new mode, and keeps the current match by occurrence
// ordinal, as D34 maps one; a Grep landing that fell back to the first match in
// Markdown lands on ripgrep's exact location in Text instead. With no find
// active the scroll position is kept as a proportion of the document. A Peek
// is re-rendered too, keeping its own proportional scroll position; the Active
// Document under it is carried as above.
func (v viewer) toggleMode() viewer {
	prev, was, frac := v.findState, v.docMode, v.scrollFraction()
	peekFrac := scrollFractionOf(v.peek.vp)
	v.mode = v.mode.toggled()
	v = v.render()
	// renderPeek lands on the top; restore the Peek's own proportional place.
	if v.peeking() {
		scrollToFraction(&v.peek.vp, peekFrac)
	}
	if !v.doc.isOpen() {
		return v
	}
	if prev.input.Value() == "" && !prev.editing {
		return v.withScrollFraction(frac)
	}
	v = v.withScrollFraction(frac)
	v.findState = prev
	v.findState.matches = v.findIn(prev.input.Value())
	v.findState.current = v.carriedMatch(prev, was)
	return v.applyCurrentMatch()
}

// carriedMatch is the index, in the matches just found in this viewer's new
// mode, of the match prev (the find state in the old mode, was saying which
// it was) had current; -1 when there are no matches. A match with no
// counterpart degrades to the first one, as D34 does; with no current match to
// carry (the query only matches in this mode), the nearest one to the scroll
// position is current.
func (v viewer) carriedMatch(prev find, was renderMode) int {
	matches := v.findState.matches
	if len(matches) == 0 {
		return -1
	}
	now := v.docMode
	if was == now {
		return min(max(prev.current, 0), len(matches)-1)
	}
	if prev.landing != nil {
		if i, ok := now.behaviour().carriedLanding(v.doc.src, matches, *prev.landing); ok {
			return i
		}
	}
	if prev.current < 0 || prev.current >= len(prev.matches) {
		return nearestMatch(matches, v.vp.YOffset())
	}
	// Frontmatter is counted apart from the body (D138): the Frontmatter Card
	// reshapes its text, so an occurrence the card drops must not shift every
	// body match after it.
	srcLines, renderedLines := frontmatterSpan(v.doc.src, v.vp.Width())
	fromHead := was.behaviour().headMatches(prev.matches, srcLines, renderedLines)
	toHead := now.behaviour().headMatches(matches, srcLines, renderedLines)
	switch cur := prev.current; {
	case cur >= fromHead && toHead+cur-fromHead < len(matches):
		return toHead + cur - fromHead
	case cur < fromHead && cur < toHead:
		return cur
	}
	return 0
}

// scrollFraction is how far down the document the viewport is, from 0 (top) to
// 1 (bottom); 0 when everything fits.
func (v viewer) scrollFraction() float64 { return scrollFractionOf(v.vp) }

// scrollFractionOf is scrollFraction for any viewport.
func scrollFractionOf(vp viewport.Model) float64 {
	if span := vp.TotalLineCount() - vp.Height(); span > 0 {
		return float64(vp.YOffset()) / float64(span)
	}
	return 0
}

// withScrollFraction scrolls to the given fraction of the document.
func (v viewer) withScrollFraction(f float64) viewer {
	scrollToFraction(&v.vp, f)
	return v
}

// scrollToFraction scrolls vp to the given fraction of its document.
func scrollToFraction(vp *viewport.Model, f float64) {
	span := max(vp.TotalLineCount()-vp.Height(), 0)
	vp.SetYOffset(int(math.Round(f * float64(span))))
}

// renderDoc renders d for width in the mode this viewer shows it in. muted
// selects the unfocused-pane variant (D136).
func (v viewer) renderDoc(d document, width int, muted bool) (string, error) {
	if v.modeFor(d) == modeText {
		return renderText(d.src, width, muted), nil
	}
	return renderMarkdownStyled(d.src, width, muted)
}

// toggleRenderMode is keys.RenderMode's action: flip the render mode of the
// viewer `/` and `e` would target (DV2 when it holds Focus, DV1 otherwise).
// A Text file is shown as text whatever the mode, so there is nothing to
// toggle there: refused with a Footer message.
func (m model) toggleRenderMode() model {
	p := m.focusedViewerPane()
	v := m.viewerFor(p)
	if shown := v.shownDoc(); isNotMarkdownFile(shown) {
		m.message = "cannot switch render mode: " + shown.label() + " is a text file"
		return m
	}
	return m.withViewerFor(p, v.toggleMode())
}

var (
	gutterStyle      = lipgloss.NewStyle().Foreground(theme.Subtle)
	mutedGutterStyle = lipgloss.NewStyle().Foreground(theme.UnfocusedSubtle)
	mutedTextStyle   = lipgloss.NewStyle().Foreground(theme.UnfocusedText)
	matchSpanStyle   = lipgloss.NewStyle().Reverse(true)
)

// textSpan is a byte range [start, end) within one display line of a Text
// render (textLayout.lines): what Find indicates as the current match.
type textSpan struct{ line, start, end int }

// textLayout is a document laid out as Text for one pane width: every source
// line, soft-wrapped into rows, with the row each line starts on. It is the
// one place that knows how a source line maps onto viewport rows, so Find can
// search source lines and still scroll to, and highlight on, the rows a match
// lands on. A layout is immutable once built.
type textLayout struct {
	// rows[i] are the plain (unstyled, gutterless) rows of display line i.
	// Concatenated they are exactly that line, so a byte offset in the line
	// lands in a definite row.
	rows [][]string
	// rowStart[i] is the viewport row display line i starts on; the final
	// entry is the total row count.
	rowStart []int
	// digits is the line-number gutter's digit count; gutter is its full
	// width in cells (0 in a pane too narrow to show one).
	digits, gutter int
}

// textLines is src as Text shows it, one display line per entry: tabs
// expanded, a leading BOM and each trailing carriage return dropped, and
// control characters replaced so a file can never drive the terminal.
func textLines(src []byte) []string {
	text := strings.TrimPrefix(string(expandTabs(src)), "\ufeff")
	if text == "" {
		return nil
	}
	lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
	for i, line := range lines {
		lines[i] = sanitizeControl(strings.TrimSuffix(line, "\r"))
	}
	return lines
}

// textDisplayCol translates a byte column within raw source line n (0-based)
// into the byte column of the same place in the display line textLines makes
// of it: tabs expanded, control characters replaced, a leading BOM dropped. The
// transform is local to a line and to the text before the column, so the
// prefix alone determines the result.
func textDisplayCol(src []byte, n, col int) int {
	for ; n > 0; n-- {
		i := bytes.IndexByte(src, '\n')
		if i < 0 {
			return 0
		}
		src = src[i+1:]
	}
	if i := bytes.IndexByte(src, '\n'); i >= 0 {
		src = src[:i]
	}
	prefix := sanitizeControl(string(expandTabs(src[:min(max(col, 0), len(src))])))
	return len(strings.TrimPrefix(prefix, "\ufeff"))
}

// layoutText lays src out for a pane width cells wide. A pane too narrow for
// a gutter and any text beside it gets the text alone, so the rows stay
// within the pane. A non-positive width does not wrap.
func layoutText(src []byte, width int) *textLayout {
	lines := textLines(src)
	l := &textLayout{
		rows:     make([][]string, len(lines)),
		rowStart: make([]int, len(lines)+1),
		digits:   len(strconv.Itoa(len(lines))),
	}
	l.gutter = l.digits + 1
	if width > 0 && width <= l.gutter+1 {
		l.gutter = 0
	}
	wrapAt := 0
	if width > 0 {
		wrapAt = width - l.gutter
	}
	for i, line := range lines {
		l.rows[i] = []string{line}
		if wrapAt > 0 {
			l.rows[i] = strings.Split(ansi.Hardwrap(line, wrapAt, true), "\n")
		}
		l.rowStart[i+1] = l.rowStart[i] + len(l.rows[i])
	}
	return l
}

// lineRows is display line i as styled viewport rows: a dim number on the
// first row and a blank gutter on the continuation rows, the body unstyled
// (or muted, for an unfocused pane, D136). span, when it is on line i, is
// drawn reversed.
func (l *textLayout) lineRows(i int, muted bool, span *textSpan) []string {
	numStyle, bodyStyle := gutterStyle, lipgloss.NewStyle()
	if muted {
		numStyle, bodyStyle = mutedGutterStyle, mutedTextStyle
	}
	body := func(s string) string {
		if s == "" || !muted {
			return s
		}
		return bodyStyle.Render(s)
	}
	out := make([]string, len(l.rows[i]))
	off := 0
	for j, row := range l.rows[i] {
		end := off + len(row)
		if span != nil && span.line == i && max(span.start, off) < min(span.end, end) {
			lo, hi := max(span.start, off)-off, min(span.end, end)-off
			row = body(row[:lo]) + matchSpanStyle.Render(row[lo:hi]) + body(row[hi:])
		} else {
			row = body(row)
		}
		off = end
		switch {
		case l.gutter == 0:
		case j == 0:
			row = numStyle.Render(fmt.Sprintf("%*d", l.digits, i+1)+" ") + row
		default:
			row = strings.Repeat(" ", l.gutter) + row
		}
		out[j] = row
	}
	return out
}

// render is the whole layout as viewport content.
func (l *textLayout) render(muted bool) string {
	rows := make([]string, 0, l.rowStart[len(l.rows)])
	for i := range l.rows {
		rows = append(rows, l.lineRows(i, muted, nil)...)
	}
	return strings.Join(rows, "\n")
}

// find locates every occurrence of query in the source lines, with the same
// smart-case rule as Markdown Find. A match lies within one source line, so it
// can span soft-wrapped rows. line is the viewport row the match starts on.
func (l *textLayout) find(query string) []findMatch {
	if query == "" || len(l.rows) == 0 {
		return nil
	}
	lines := make([]string, len(l.rows))
	for i, rows := range l.rows {
		lines[i] = strings.Join(rows, "")
	}
	found := findInText(strings.Join(lines, "\n"), query)
	for i, m := range found {
		row, off := 0, 0
		for row+1 < len(l.rows[m.line]) && off+len(l.rows[m.line][row]) <= m.colStart {
			off += len(l.rows[m.line][row])
			row++
		}
		endRow := row
		for off+len(l.rows[m.line][endRow]) < m.colEnd && endRow+1 < len(l.rows[m.line]) {
			off += len(l.rows[m.line][endRow])
			endRow++
		}
		found[i] = findMatch{
			line: l.rowStart[m.line] + row, colStart: m.colStart, colEnd: m.colEnd,
			src: m.line, endLine: l.rowStart[m.line] + endRow,
		}
	}
	return found
}

// renderText renders src as raw text for a pane width cells wide: every
// source line numbered in a dim gutter, tabs expanded, long lines soft-wrapped
// with a blank gutter on the continuation rows. There is no syntax colour; the
// only styling is the gutter's, and the muted variant of both for an
// unfocused pane (D136).
func renderText(src []byte, width int, muted bool) string {
	return layoutText(src, width).render(muted)
}

// withTextMatch indicates span (nil clears it) in the Text render: only the
// matched span is drawn reversed. Only the lines the old and new span sit on
// are re-rendered and spliced into the viewport content, so stepping through
// matches in a large file does not re-style the whole document.
func (v viewer) withTextMatch(span *textSpan) viewer {
	old := v.textMatch
	v.textMatch = span
	v.vp.SetContent(v.text.patch(v.vp.GetContent(), false, old, span))
	v.mutedContent = v.text.patch(v.mutedContent, true, old, span)
	return v
}

// patch is content (this layout rendered with old as its span, muted or not)
// re-rendered with span instead: only the lines old and span sit on are
// re-styled and spliced in.
func (l *textLayout) patch(content string, muted bool, old, span *textSpan) string {
	if content == "" || (old == nil && span == nil) {
		return content
	}
	rows := strings.Split(content, "\n")
	for _, s := range []*textSpan{old, span} {
		if s != nil {
			copy(rows[l.rowStart[s.line]:], l.lineRows(s.line, muted, span))
		}
	}
	return strings.Join(rows, "\n")
}

// sanitizeControl replaces control characters (ESC and friends) with U+FFFD so
// a file's contents can never drive the terminal. Tabs are already expanded.
func sanitizeControl(s string) string {
	if !strings.ContainsFunc(s, unicode.IsControl) {
		return s
	}
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return '\ufffd'
		}
		return r
	}, s)
}
