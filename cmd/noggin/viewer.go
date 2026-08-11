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
// no selection (D4). findState is the S8 in-document search state, layered on top.
type viewer struct {
	vp  viewport.Model
	doc document

	// renderedWidth is the width doc was last rendered at, so a resize that
	// leaves the width alone does not pay to re-render.
	renderedWidth int

	// err is a render failure, shown in the pane instead of the document.
	err error

	findState find

	// peek is S7/M7.1's minimal preview state: a document shown in the pane
	// instead of doc, without becoming the Active Document. The zero value
	// (peekState{}, doc.isOpen() == false) means not Peeking, the same
	// sentinel document itself already uses — a pointer would work too, but
	// every other field on viewer is value-semantic (see withOffset's own
	// doc comment) and this keeps that true.
	peek peekState

	// history and historyPos are S9/C4-M6.1's per-viewer back/forward stack:
	// the document paths opened via model.navigate (Activate, the Picker's
	// Activate, and the launch-time document), in visit order, with
	// historyPos indexing the current entry. Zero value (nil, 0) means no
	// history yet, matching every viewer's fresh state — the pair is never
	// populated by open()/refreshDoc themselves (Peek's startPeek, the
	// $EDITOR handoff's re-open, and a Refresh re-render all leave it alone,
	// S9's own list of non-recording events); only pushHistory writes to it,
	// and only model.navigate calls that. A pointer-free, value-semantic pair
	// like every other viewer field (withOffset's own doc comment) — DV2's
	// history is simply whatever *m.dv2 carries, so it disappears the moment
	// closeDV2 drops the pointer, with nothing further to clean up.
	history    []string
	historyPos int
}

// peekState is one Peeked document: its own read-only viewport, kept
// entirely separate from vp so Peek can never move the Active Document's
// scroll position — clearing Peek must restore vp exactly as it was found.
// It never scrolls (S7), so GotoTop on every (re-)render is deliberate, not
// a no-op that happens not to matter.
type peekState struct {
	doc document
	vp  viewport.Model
	err error
}

func (p peekState) isOpen() bool { return p.doc.isOpen() }

func newViewer() viewer {
	vp := viewport.New(viewport.WithWidth(0), viewport.WithHeight(0))
	// The viewport's own KeyMap carries key literals; leaving it populated
	// would put bindings outside keymap.go (D2) and duplicate the ones below.
	vp.KeyMap = viewport.KeyMap{}
	// v2 stops padding short content out to the viewport's height unless asked;
	// the pane contract is a content box of exactly its size.
	vp.FillHeight = true
	return viewer{vp: vp, findState: newFind()}
}

// setSize sets the content box the viewer draws into, re-rendering (and so
// re-wrapping) only when the width actually changed. Peek's own viewport
// (S7, M7.1) always tracks the same box dimensions here, even on a
// height-only change render() would skip — render()/renderPeek only re-wrap
// the content on a width change, but a stale height would clip Peek wrong
// on a height-only resize if this didn't run unconditionally.
func (v viewer) setSize(width, height int) viewer {
	v.vp.SetWidth(max(width, 0))
	v.vp.SetHeight(max(height, 0))
	if v.peeking() {
		v.peek.vp.SetWidth(max(width, 0))
		v.peek.vp.SetHeight(max(height, 0))
	}
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

// refreshDoc replaces the open document with a freshly re-read d (S2,
// C2a/M5.1) and re-renders it in place, deliberately not following open()'s
// GotoTop(): render's SetContent call already clamps the offset down to the
// new end if the document shrank past it (viewport.SetContentLines), which
// is exactly S2's "clamped to the end" contract, so no extra clamping is
// needed here — leaving the offset alone is what preserves scroll position
// otherwise. Find state is dropped the same way any other re-render already
// drops it (T34/T43, render's own doc comment) — S2 allows this ("Find
// state may be lost").
func (v viewer) refreshDoc(d document) viewer {
	v.doc = d
	return v.render()
}

// close empties the viewer: refresh's path (S2, C2a/M5.1) for when the
// Active Document has vanished or become unreadable underneath it — the
// empty state open() never produces, since it is only ever called with a
// document that just loaded successfully.
func (v viewer) close() viewer {
	v.doc = document{}
	return v.render()
}

// pushHistory appends path as the S9 history's new current entry (C4/
// M6.1), discarding any forward entries left by an earlier goBack —
// browser semantics: "activating a new document after going back discards
// the forward entries". A no-op when path is already the current entry,
// which is what keeps the same document from ever landing in two adjacent
// entries (S9's own rule) — every other insertion only ever appends after
// truncating to historyPos, so this one check at push time is enough to
// hold that invariant across the whole stack, not just the latest push.
// The empty-history case is handled separately since slicing history[:1]
// against a nil/empty slice would panic.
func (v viewer) pushHistory(path string) viewer {
	if len(v.history) == 0 {
		v.history = []string{path}
		v.historyPos = 0
		return v
	}
	if v.history[v.historyPos] == path {
		return v
	}
	kept := append([]string{}, v.history[:v.historyPos+1]...)
	v.history = append(kept, path)
	v.historyPos = len(v.history) - 1
	return v
}

// canGoBack reports whether the S9 history has an earlier entry to move to.
func (v viewer) canGoBack() bool { return v.historyPos > 0 }

// canGoForward reports whether the S9 history has a later entry to move to
// — true only after goBack has moved off the most recent entry and nothing
// has pushed a fresh one since (pushHistory's own truncation).
func (v viewer) canGoForward() bool { return v.historyPos < len(v.history)-1 }

// goBack and goForward move the S9 history position one entry in their own
// direction, returning the path to open there — model.walkViewerHistory is
// what actually loads and opens it, since a viewer method has no access to
// disk. ok is false, v unchanged, at either end of history: there is
// nothing to move to, not a refusal, the same "nothing to do" shape
// FindNext/FindPrevious already give an empty match list.
func (v viewer) goBack() (viewer, string, bool) {
	if !v.canGoBack() {
		return v, "", false
	}
	v.historyPos--
	return v, v.history[v.historyPos], true
}

func (v viewer) goForward() (viewer, string, bool) {
	if !v.canGoForward() {
		return v, "", false
	}
	v.historyPos++
	return v, v.history[v.historyPos], true
}

// peeking reports whether a document is currently Peeked (S7, M7.1).
func (v viewer) peeking() bool { return v.peek.isOpen() }

// peekLabel names the Peeked document for the pane title, or is empty when
// not Peeking — the Peek equivalent of label().
func (v viewer) peekLabel() string {
	if !v.peeking() {
		return ""
	}
	return filepath.Base(v.peek.doc.path)
}

// startPeek shows d in the pane without disturbing doc or vp (S7, M7.1): a
// fresh viewport sized to match the Active Document's, rendered once at the
// current width. model.togglePeek is the one caller, and is what enforces
// the surrounding rules (not the Active Document, not already Peeking) —
// this method only ever renders whatever it is given.
func (v viewer) startPeek(d document) viewer {
	vp := viewport.New(viewport.WithWidth(v.vp.Width()), viewport.WithHeight(v.vp.Height()))
	vp.FillHeight = true
	v.peek = peekState{doc: d, vp: vp}
	return v.renderPeek()
}

// clearPeek ends Peek, restoring the Active Document's own view — nothing
// to restore on vp itself, since Peek never touched it (S7, M7.1).
func (v viewer) clearPeek() viewer {
	v.peek = peekState{}
	return v
}

// renderPeek (re-)renders the Peeked document into its own viewport, sized
// to the pane's current width/height — called from startPeek and from
// render() so a width change (setSize) reflows Peek exactly like the Active
// Document, and a no-op when not Peeking. GotoTop on every call is
// deliberate: Peek never scrolls (S7), so it is always shown from the top
// regardless of what a previous render left YOffset at.
func (v viewer) renderPeek() viewer {
	if !v.peeking() {
		return v
	}
	v.peek.vp.SetWidth(v.vp.Width())
	v.peek.vp.SetHeight(v.vp.Height())
	out, err := renderMarkdown(v.peek.doc.src, v.peek.vp.Width())
	if err != nil {
		v.peek.err = err
		v.peek.vp.SetContent("")
		return v
	}
	v.peek.err = nil
	v.peek.vp.SetContent(strings.TrimRight(out, "\n"))
	v.peek.vp.GotoTop()
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
	v.findState = newFind()
	v.vp.StyleLineFunc = nil
	if !v.doc.isOpen() {
		v.vp.SetContent("")
	} else if out, err := renderMarkdown(v.doc.src, v.vp.Width()); err != nil {
		v.err = err
		v.vp.SetContent("")
	} else {
		// Glamour pads the document with trailing blank lines; they would
		// be scrollable emptiness past the end.
		v.vp.SetContent(strings.TrimRight(out, "\n"))
	}
	// A width change reflows Peek (S7, M7.1) exactly like the Active
	// Document above — render() is the one path both go through on a
	// setSize; a doc-only re-render (open/refreshDoc/close) re-renders Peek
	// too, harmlessly, since renderPeek is a no-op when not Peeking.
	return v.renderPeek()
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
	// A Peeked document (S7, M7.1) takes over the content box entirely,
	// ahead of the Active Document's own err/empty/open cases below — those
	// describe vp, which Peek never touches or replaces.
	if v.peeking() {
		if v.peek.err != nil {
			return v.fill(emptyViewerStyle, v.peek.err.Error())
		}
		return v.peek.vp.View()
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
