package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"noggin/internal/tree"
)

// peekFixture builds a real temp Root with an Active Document (active.md,
// opened and focus moved to the viewer) and a second file (peeked.md)
// selected in the Files Nav, focus moved back to the Files Nav — the
// starting position every M7.1 test presses `p` from.
func peekFixture(t *testing.T) (m model, activePath, peekedPath string) {
	t.Helper()
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	activePath = writeFileContent(t, filepath.Join(root, "active.md"), "# Active\n")
	peekedPath = writeFileContent(t, filepath.Join(root, "peeked.md"), "# Peeked\n")

	node, err := tree.Walk(root, tree.FilterMarkdown)
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	m = sized(t, newModel().openFolder(node), 80, 24)

	d, err := loadDocument(activePath)
	if err != nil {
		t.Fatalf("loadDocument(active): %v", err)
	}
	m = m.openDocument(d) // Active Document set, focus -> dv1Pane

	m = m.focusPane(filesPane) // S6: back to the Files Nav
	// active.md, peeked.md sort alphabetically before active.md — walk
	// order is dirs-first then by name, both plain files here.
	if m.files.selectedNode().Name != "active.md" {
		t.Fatalf("setup: initial selection = %q, want active.md", m.files.selectedNode().Name)
	}
	m.files = m.files.move(1)
	if m.files.selectedNode().Name != "peeked.md" {
		t.Fatalf("setup: selection after move = %q, want peeked.md", m.files.selectedNode().Name)
	}
	return m, activePath, peekedPath
}

// TestPeekShowsSelectedDocumentWithoutChangingActiveDocumentOrFocus is
// M7.1's stated acceptance verbatim: pressing p on a non-Active file shows
// it with the Peek header while Focus stays in the Files Nav.
func TestPeekShowsSelectedDocumentWithoutChangingActiveDocumentOrFocus(t *testing.T) {
	m, activePath, _ := peekFixture(t)

	m = press(t, m, keys.Peek.Keys()[0])

	if m.focus != filesPane {
		t.Errorf("focus after Peek = %v, want filesPane (unchanged)", m.focus)
	}
	if !m.dv1.peeking() {
		t.Fatal("viewer not Peeking after p")
	}
	if m.dv1.doc.path != activePath {
		t.Errorf("Active Document path after Peek = %q, want unchanged %q", m.dv1.doc.path, activePath)
	}
	view := m.View().Content
	if !strings.Contains(view, "Peeked") {
		t.Errorf("view after Peek = %q, want the Peeked document's content", view)
	}
	if strings.Contains(view, "Active\n") {
		t.Errorf("view after Peek = %q, want the Active Document's own content hidden", view)
	}
	if !strings.Contains(view, "Peek: peeked.md") {
		t.Errorf("view after Peek = %q, want a header identifying the Peeked document", view)
	}
}

// TestPeekHeaderIsInvertedFromTheActiveDocuments checks S7's "inverted
// header" requirement is a real, distinct style from the plain title the
// Active Document's own view already uses — not just different text.
//
// The expected rows are built through titleRow, the same helper renderPane
// itself calls (M7.1), rather than the bare title text: DV1's title row now
// carries a right-aligned focus-key badge ("[o]") the raw title alone does
// not, so comparing against an un-badged render would never match.
func TestPeekHeaderIsInvertedFromTheActiveDocuments(t *testing.T) {
	m, _, _ := peekFixture(t)
	l := computeLayout(m.width, m.height, m.navigatorHidden, m.stale)
	dv1Rect, _ := l.splitViewer(m.dv2 != nil)
	width := dv1Rect.content().width

	// The title rows are compared by their cells' styles, not by re-rendering
	// the expected row: the frame's canvas pass (paintCanvas) re-serializes
	// every SGR, so a hand-rendered row never matches byte for byte.
	plainViewerTitle := ansi.Strip(titleRow(m.paneTitle(dv1Pane), dv1Pane.focusBadge(), width)) // "Document: active.md ... [o]", not inverted
	line, ok := styledLineWith(m.View().Content, strings.TrimSpace(plainViewerTitle))
	if !ok {
		t.Fatalf("setup: Active Document's own title not found before Peeking")
	}
	if hasReverse(line) {
		t.Fatalf("setup: Active Document's own title is inverted before Peeking: %q", line)
	}

	m = press(t, m, keys.Peek.Keys()[0])
	after := m.View().Content

	peekTitle := ansi.Strip(titleRow(m.paneTitle(dv1Pane), dv1Pane.focusBadge(), width)) // "Peek: peeked.md ... [o]", inverted
	peekLine, ok := styledLineWith(after, strings.TrimSpace(peekTitle))
	if !ok || !hasReverse(peekLine) {
		t.Errorf("Peek header not rendered inverted (peekTitleStyle); view = %q", after)
	}
	if _, still := styledLineWith(after, strings.TrimSpace(plainViewerTitle)); still {
		t.Errorf("view after Peek still shows the Active Document's own plain title; view = %q", after)
	}
}

// TestPeekOnTheActiveDocumentIsANoOp is S7's own carve-out: pressing p on
// the row that is already the Active Document does nothing.
func TestPeekOnTheActiveDocumentIsANoOp(t *testing.T) {
	m, activePath, _ := peekFixture(t)
	m.files = m.files.move(-1) // back onto active.md
	if m.files.selectedNode().Name != "active.md" {
		t.Fatalf("setup: selection = %q, want active.md", m.files.selectedNode().Name)
	}

	m = press(t, m, keys.Peek.Keys()[0])

	if m.dv1.peeking() {
		t.Error("Peek started on the Active Document's own row, want a no-op")
	}
	if m.dv1.doc.path != activePath {
		t.Errorf("Active Document path = %q, want unchanged %q", m.dv1.doc.path, activePath)
	}
}

// TestPeekCannotBeScrolled checks the "cannot be scrolled" clause: a
// scroll-only binding (ctrl+d, not shared with any Files Nav binding)
// leaves Peek exactly as it was — it never even reaches the viewer, since
// Focus stays on the Files Nav throughout Peek.
func TestPeekCannotBeScrolled(t *testing.T) {
	m, _, _ := peekFixture(t)
	m = press(t, m, keys.Peek.Keys()[0])
	viewBefore := m.View().Content

	m = press(t, m, keys.HalfPageDown.Keys()[0])

	if !m.dv1.peeking() {
		t.Error("Peek cleared by a scroll-only key, want it to still be Peeking")
	}
	if got := m.View().Content; got != viewBefore {
		t.Errorf("view changed after a scroll attempt while Peeking:\nbefore=%q\nafter=%q", viewBefore, got)
	}
}

// TestPeekCannotBeFound checks the "cannot be searched with Find" clause:
// pressing Find while Peeking does nothing at all — no find starts (not
// even against the Active Document underneath), Peek stays up.
func TestPeekCannotBeFound(t *testing.T) {
	m, _, _ := peekFixture(t)
	m = press(t, m, keys.Peek.Keys()[0])

	m = press(t, m, keys.Find.Keys()[0])

	if !m.dv1.peeking() {
		t.Error("Peek cleared by Find, want it to still be Peeking")
	}
	if m.dv1.findEditing() {
		t.Error("Find started while Peeking, want a silent no-op")
	}
	if m.focus != filesPane {
		t.Errorf("focus after Find while Peeking = %v, want unchanged filesPane", m.focus)
	}
}

// TestPeekEditIsASilentNoOp checks S7's own explicit rule: `e` does nothing
// while Peeking, and produces no footer message either (unlike Edit's usual
// "no document open" refusal).
func TestPeekEditIsASilentNoOp(t *testing.T) {
	m, _, _ := peekFixture(t)
	m = press(t, m, keys.Peek.Keys()[0])

	m = press(t, m, keys.Edit.Keys()[0])

	if !m.dv1.peeking() {
		t.Error("Peek cleared by Edit, want it to still be Peeking")
	}
	if m.message != "" {
		t.Errorf("message after Edit while Peeking = %q, want silence", m.message)
	}
}

// TestPeekOnAFolderIsANoOp: Peek is documented as showing "the selected
// document" (S7) — a folder has none, so p on one must not panic or start
// a bogus Peek.
func TestPeekOnAFolderIsANoOp(t *testing.T) {
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	mkdir(t, filepath.Join(root, "sub"))
	node, err := tree.Walk(root, tree.FilterMarkdown)
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	m := sized(t, newModel().openFolder(node), 80, 24)
	if !m.files.selectedNode().IsDir {
		t.Fatalf("setup: selection is not a directory")
	}

	m = press(t, m, keys.Peek.Keys()[0])

	if m.dv1.peeking() {
		t.Error("Peek started on a folder, want a no-op")
	}
}
