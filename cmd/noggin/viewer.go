package main

import (
	"path/filepath"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// emptyViewerText is what the Document Viewer says when nothing is open (S7).
const emptyViewerText = "No document open. Pick a file in the Navigator."

// viewer is the Document Viewer: the active document, its rendered form and the
// scroll position within it. Navigation is viewport scrolling only — no cursor,
// no selection (D4). find is the S8 in-document search state, layered on top.
type viewer struct {
	vp  viewport.Model
	doc document

	// renderedWidth is the width doc was last rendered at, so a resize that
	// leaves the width alone does not pay to re-render.
	renderedWidth int

	// err is a render failure, shown in the pane instead of the document.
	err error

	find find
}

func newViewer() viewer {
	vp := viewport.New(viewport.WithWidth(0), viewport.WithHeight(0))
	// The viewport's own KeyMap carries key literals; leaving it populated
	// would put bindings outside keymap.go (D2) and duplicate the ones below.
	vp.KeyMap = viewport.KeyMap{}
	// v2 stops padding short content out to the viewport's height unless asked;
	// the pane contract is a content box of exactly its size.
	vp.FillHeight = true
	return viewer{vp: vp, find: newFind()}
}

// setSize sets the content box the viewer draws into, re-rendering (and so
// re-wrapping) only when the width actually changed.
func (v viewer) setSize(width, height int) viewer {
	v.vp.SetWidth(max(width, 0))
	v.vp.SetHeight(max(height, 0))
	if v.vp.Width() != v.renderedWidth {
		v = v.render()
	}
	return v
}

// open makes d the active document, rendered from the top.
func (v viewer) open(d document) viewer {
	v.doc = d
	v = v.render()
	v.vp.GotoTop()
	return v
}

// render rebuilds the viewport's content from the document at the current width.
//
// A re-render invalidates any find in progress (M5.1): a resize re-wraps the
// document, so every match's line number would be wrong, and a re-render also
// means a different document is open (or the same one just re-read from disk,
// M4.1) — either way, stale matches over content that no longer exists.
func (v viewer) render() viewer {
	v.renderedWidth = v.vp.Width()
	v.err = nil
	v.find = newFind()
	v.vp.StyleLineFunc = nil
	if !v.doc.isOpen() {
		v.vp.SetContent("")
		return v
	}
	out, err := renderMarkdown(v.doc.src, v.vp.Width())
	if err != nil {
		v.err = err
		v.vp.SetContent("")
		return v
	}
	// Glamour pads the document with trailing blank lines; they would be
	// scrollable emptiness past the end.
	v.vp.SetContent(strings.TrimRight(out, "\n"))
	return v
}

// offset is the viewport's current vertical scroll position, in rendered lines.
func (v viewer) offset() int { return v.vp.YOffset() }

// withOffset moves the viewport to a given scroll position directly, for
// tests that need to start from a known offset rather than scroll there one
// action at a time. A value receiver, like every other viewer method (T10) —
// the pointer receiver this used to carry bought nothing here, since
// SetYOffset already mutates the shared vp.Model by reference.
func (v viewer) withOffset(n int) viewer { v.vp.SetYOffset(n); return v }

// atBottom reports whether the viewport is scrolled to the end of the document.
func (v viewer) atBottom() bool { return v.vp.AtBottom() }

// lineCount is the number of rendered lines in the open document.
func (v viewer) lineCount() int { return v.vp.TotalLineCount() }

// height is the viewport's content height, in rows.
func (v viewer) height() int { return v.vp.Height() }

// label names the open document for the pane title, or is empty when none is.
func (v viewer) label() string {
	if !v.doc.isOpen() {
		return ""
	}
	return filepath.Base(v.doc.path)
}

var emptyViewerStyle = lipgloss.NewStyle().Foreground(dimColor)

// view renders the viewer's content box at exactly its width x height.
func (v viewer) view() string {
	if v.vp.Width() <= 0 || v.vp.Height() <= 0 {
		return ""
	}
	switch {
	case v.err != nil:
		return v.fill(emptyViewerStyle, v.err.Error())
	case !v.doc.isOpen():
		return v.fill(emptyViewerStyle, emptyViewerText)
	default:
		return v.vp.View()
	}
}

// fill lays text out over the whole content box, padding and truncating so the
// result is exactly the box's size — the same contract viewport.View() keeps.
func (v viewer) fill(style lipgloss.Style, text string) string {
	return style.
		Width(v.vp.Width()).Height(v.vp.Height()).
		MaxWidth(v.vp.Width()).MaxHeight(v.vp.Height()).
		Render(text)
}

// scrollActions pairs each scroll binding with the viewport move it performs.
// The five moves S7 allows are line, half-page, page, top and bottom. Update
// walks this list, and so can the help modal (S12), so a new scroll action is
// one entry here plus its binding in keymap.go.
var scrollActions = []struct {
	binding key.Binding
	move    func(*viewport.Model)
}{
	{keys.LineUp, func(vp *viewport.Model) { vp.ScrollUp(1) }},
	{keys.LineDown, func(vp *viewport.Model) { vp.ScrollDown(1) }},
	{keys.HalfPageUp, func(vp *viewport.Model) { vp.HalfPageUp() }},
	{keys.HalfPageDown, func(vp *viewport.Model) { vp.HalfPageDown() }},
	{keys.PageUp, func(vp *viewport.Model) { vp.PageUp() }},
	{keys.PageDown, func(vp *viewport.Model) { vp.PageDown() }},
	{keys.Top, func(vp *viewport.Model) { vp.GotoTop() }},
	{keys.Bottom, func(vp *viewport.Model) { vp.GotoBottom() }},
}

// scroll applies the scroll action msg is bound to, reporting whether it was
// one.
func (v viewer) scroll(msg tea.KeyPressMsg) (viewer, bool) {
	for _, action := range scrollActions {
		if key.Matches(msg, action.binding) {
			action.move(&v.vp)
			return v, true
		}
	}
	return v, false
}
