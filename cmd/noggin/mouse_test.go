package main

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

// TestViewTakesOverTheTerminal is the regression test for the bug report:
// noggin used to leave every tea.View field at its zero value, so it never
// entered the alt screen buffer (nothing to restore on exit — the frame
// stayed in the normal scrollback after quitting) and never asked the
// terminal to report mouse events (wheel scroll fell through to the
// terminal's own native scrolling instead of scrolling a pane).
func TestViewTakesOverTheTerminal(t *testing.T) {
	m := sized(t, newModel(), 80, 24)
	v := m.View()
	if !v.AltScreen {
		t.Error("View().AltScreen = false, want true (full terminal takeover)")
	}
	if v.MouseMode != tea.MouseModeCellMotion {
		t.Errorf("View().MouseMode = %v, want MouseModeCellMotion (click/release/wheel)", v.MouseMode)
	}
}

// TestMouseWheelScrollsTheFilesNavUnderTheCursor is the routing half of the
// mouse fix: a wheel event lands wherever the cursor's cell is, not
// wherever Focus happens to be — paneAt maps the event's (X, Y) back to the
// pane whose region contains it.
func TestMouseWheelScrollsTheFilesNavUnderTheCursor(t *testing.T) {
	m := sized(t, newModel().openFolder(fixtureTree()), 80, 24)
	m.focus = dv1Pane // deliberately not the Files Nav, to prove routing is by cursor position, not Focus
	if m.files.rowCount() < 2 {
		t.Fatalf("fixture tree has %d rows, want at least 2 to scroll between", m.files.rowCount())
	}
	l := m.layout()
	x, y := 1, l.roots.height+1 // inside the Navigator column, below Sources Nav's rows: the Files Nav

	before := m.files.selected
	next, _ := m.Update(tea.MouseWheelMsg{X: x, Y: y, Button: tea.MouseWheelDown})
	m = concrete(t, next)
	if m.files.selected != before+1 {
		t.Errorf("files.selected after wheel-down over the Files Nav = %d, want %d", m.files.selected, before+1)
	}
	if m.focus != dv1Pane {
		t.Errorf("focus after a wheel event = %v, want unchanged (dv1Pane) — wheel scroll must not steal Focus", m.focus)
	}

	next, _ = m.Update(tea.MouseWheelMsg{X: x, Y: y, Button: tea.MouseWheelUp})
	m = concrete(t, next)
	if m.files.selected != before {
		t.Errorf("files.selected after wheel-up over the Files Nav = %d, want %d", m.files.selected, before)
	}
}

// TestMouseWheelScrollsTheDocumentViewerUnderTheCursor mirrors the Files Nav
// case above for the Document Viewer: a wheel event over DV1's region
// scrolls its viewport, regardless of which pane has Focus.
func TestMouseWheelScrollsTheDocumentViewerUnderTheCursor(t *testing.T) {
	m := openedFixture(t, 80, 24)
	m.focus = filesPane
	page := m.dv1.height()
	if total := m.dv1.lineCount(); total <= page {
		t.Fatalf("fixture renders %d lines into a %d-line viewport; nothing to scroll", total, page)
	}

	l := m.layout()
	x, y := l.navigatorWidth()+1, 1 // inside the Document Viewer's region

	next, _ := m.Update(tea.MouseWheelMsg{X: x, Y: y, Button: tea.MouseWheelDown})
	m2 := concrete(t, next)
	if got := m2.dv1.offset(); got != 1 {
		t.Errorf("dv1 offset after wheel-down over the Document Viewer = %d, want 1", got)
	}
	if m2.focus != filesPane {
		t.Errorf("focus after a wheel event = %v, want unchanged (filesPane)", m2.focus)
	}
}

// TestMouseWheelOverTheFooterIsANoOp confirms a wheel event outside every
// pane's region (the footer row, here) is silently ignored rather than
// falling back onto some default pane.
func TestMouseWheelOverTheFooterIsANoOp(t *testing.T) {
	m := sized(t, newModel().openFolder(fixtureTree()), 80, 24)
	before := m

	next, _ := m.Update(tea.MouseWheelMsg{X: 1, Y: m.height - 1, Button: tea.MouseWheelDown})
	m2 := concrete(t, next)

	if m2.files.selected != before.files.selected {
		t.Errorf("files.selected changed from a footer-row wheel event: %d -> %d", before.files.selected, m2.files.selected)
	}
	if m2.dv1.offset() != before.dv1.offset() {
		t.Errorf("dv1 offset changed from a footer-row wheel event: %d -> %d", before.dv1.offset(), m2.dv1.offset())
	}
}
