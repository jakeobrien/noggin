package main

import (
	"path/filepath"
	"strings"
	"testing"

	"noggin/internal/tree"
)

// peekFixtureScrollable is peekFixture (peek_test.go) with a long enough
// Active Document to actually have a mid-document scroll position to check
// for preservation across a clear (S7/M7.2's "restoring... at the scroll
// position it had before Peeking"). Returns the model already Peeking
// peeked.md, and the offset the Active Document was scrolled to beforehand.
func peekFixtureScrollable(t *testing.T) (m model, wantOffset int) {
	t.Helper()
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	activePath := writeFileContent(t, filepath.Join(root, "active.md"), listContent(60))
	writeFileContent(t, filepath.Join(root, "peeked.md"), "# Peeked\n")

	node, err := tree.Walk(root, tree.Options{Filter: tree.FilterMarkdown})
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	m = sized(t, newModel().openFolder(node), 80, 10)

	d, err := loadDocument(activePath)
	if err != nil {
		t.Fatalf("loadDocument(active): %v", err)
	}
	m = m.openDocument(d)
	if total := m.dv1.lineCount(); total <= m.dv1.height() {
		t.Fatalf("active.md renders %d lines into a %d-line viewport; nothing to preserve a scroll over", total, m.dv1.height())
	}
	m.dv1 = m.dv1.withOffset(7)
	wantOffset = m.dv1.offset()

	m = m.focusPane(filesPane)
	m.files = m.files.move(1) // active.md -> peeked.md
	if m.files.selectedNode().Name != "peeked.md" {
		t.Fatalf("setup: selection = %q, want peeked.md", m.files.selectedNode().Name)
	}
	m = press(t, m, keys.Peek.Keys()[0])
	if !m.dv1.peeking() {
		t.Fatalf("setup: Peek did not start")
	}
	return m, wantOffset
}

// assertPeekClearedRestoringScroll is the common shape every trigger test
// below checks: Peek is gone and the Active Document's own scroll position
// is exactly what it was before Peeking started (S7/M7.2).
func assertPeekClearedRestoringScroll(t *testing.T, m model, wantOffset int) {
	t.Helper()
	if m.dv1.peeking() {
		t.Error("still Peeking after the clearing trigger")
	}
	if got := m.dv1.offset(); got != wantOffset {
		t.Errorf("Active Document offset after Peek cleared = %d, want %d (preserved)", got, wantOffset)
	}
}

// TestPeekClearedBySelectionChange checks S7/M7.2's first trigger: moving
// the Files Nav selection while Peeking clears it.
func TestPeekClearedBySelectionChange(t *testing.T) {
	m, wantOffset := peekFixtureScrollable(t)
	m = press(t, m, keys.LineUp.Keys()[0]) // peeked.md -> active.md
	assertPeekClearedRestoringScroll(t, m, wantOffset)
}

// TestPeekClearedByFocusChange checks the second trigger, including the
// explicit "l/tab/f" examples S7 names (Tab-cycling itself is gone as of
// C3/M1.1, S5; FocusRoots is one of its direct-jump replacements): FocusRoots
// moves Focus off the Files Nav.
func TestPeekClearedByFocusChange(t *testing.T) {
	m, wantOffset := peekFixtureScrollable(t)
	m = press(t, m, keys.FocusRoots.Keys()[0])
	if m.focus != rootsPane {
		t.Fatalf("setup: FocusRoots did not move focus to rootsPane, got %v", m.focus)
	}
	assertPeekClearedRestoringScroll(t, m, wantOffset)
}

// TestPeekClearedByOpeningThePicker checks the third trigger.
func TestPeekClearedByOpeningThePicker(t *testing.T) {
	m, wantOffset := peekFixtureScrollable(t)
	m = press(t, m, keys.Picker.Keys()[0])
	if !m.picker.open {
		t.Fatalf("setup: picker did not open")
	}
	assertPeekClearedRestoringScroll(t, m, wantOffset)
}

// TestPeekClearedBySecondP checks the fourth trigger: pressing p again
// while already Peeking (S7's own "a second p" rule), not a fresh Peek of
// whatever happens to be selected.
func TestPeekClearedBySecondP(t *testing.T) {
	m, wantOffset := peekFixtureScrollable(t)
	m = press(t, m, keys.Peek.Keys()[0])
	assertPeekClearedRestoringScroll(t, m, wantOffset)
	if m.focus != filesPane {
		t.Errorf("focus after a second p = %v, want unchanged filesPane", m.focus)
	}
}

// TestPeekClearedByRefresh checks the fifth trigger.
func TestPeekClearedByRefresh(t *testing.T) {
	m, wantOffset := peekFixtureScrollable(t)
	m = press(t, m, keys.Refresh.Keys()[0])
	assertPeekClearedRestoringScroll(t, m, wantOffset)
}

// TestPeekClearedRestoresViewShowingTheActiveDocumentAgain checks the
// clearing triggers actually put the Active Document back on screen, not
// just internal state: the Peek header is gone and the Active Document's
// own title and content are back, the same document that was open before
// Peek started.
func TestPeekClearedRestoresViewShowingTheActiveDocumentAgain(t *testing.T) {
	m, _ := peekFixtureScrollable(t)
	if !strings.Contains(m.View().Content, "Peek: peeked.md") {
		t.Fatalf("setup: Peek header not showing before the clear")
	}

	m = press(t, m, keys.LineUp.Keys()[0])

	view := m.View().Content
	if strings.Contains(view, "Peek:") {
		t.Errorf("view after clearing Peek = %q, still shows a Peek header", view)
	}
	if !strings.Contains(view, "Document: active.md") {
		t.Errorf("view after clearing Peek = %q, want the Active Document's own title back", view)
	}
}

// TestPeekClearedWithNoActiveDocumentEmptiesTheViewer checks S7's other
// stated outcome: with no Active Document at all, clearing Peek reverts to
// the empty Viewer rather than anything else.
func TestPeekClearedWithNoActiveDocumentEmptiesTheViewer(t *testing.T) {
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	writeFileContent(t, filepath.Join(root, "peeked.md"), "# Peeked\n")
	node, err := tree.Walk(root, tree.Options{Filter: tree.FilterMarkdown})
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	m := sized(t, newModel().openFolder(node), 80, 24)
	if m.dv1.doc.isOpen() {
		t.Fatalf("setup: an Active Document is already open")
	}

	m = press(t, m, keys.Peek.Keys()[0])
	if !m.dv1.peeking() {
		t.Fatalf("setup: Peek did not start")
	}

	m = press(t, m, keys.Peek.Keys()[0]) // second p: clear
	if m.dv1.peeking() {
		t.Error("still Peeking after the second p")
	}
	if m.dv1.doc.isOpen() {
		t.Error("an Active Document appeared after clearing Peek with none open before")
	}
	if !strings.Contains(m.View().Content, emptyViewerMarker) {
		t.Errorf("view after clearing Peek with no Active Document = %q, want the empty state", m.View().Content)
	}
}
