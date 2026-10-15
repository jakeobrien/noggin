package main

import (
	"path/filepath"
	"testing"
)

// C3/M1.6: DV1/DV2 independence (S2, S6) — once Focus is on a viewer, Find,
// Edit and scroll must act only on that viewer's own Active Document,
// leaving the other viewer's document, Find state and scroll position
// completely untouched. Much of the routing (focusedViewer/withFocusedViewer,
// M1.1) predates this task; these tests are the plan's own done-when: two
// distinct documents, distinct Find state, driven from both directions.

// crossViewerFixture opens docA (marked "alpha-marker") as DV1's Active
// Document and docB (marked "bravo-marker") as DV2's, both long enough to
// scroll in a 10-row viewport. Returns with Focus left on dv2Pane (the last
// one opened), the natural place to start driving DV2 first.
func crossViewerFixture(t *testing.T) (m model, pathA, pathB string) {
	t.Helper()
	dir := t.TempDir()
	pathA = writeFileContent(t, filepath.Join(dir, "a.md"), "alpha-marker\n"+listContent(60))
	pathB = writeFileContent(t, filepath.Join(dir, "b.md"), "bravo-marker\n"+listContent(60))

	docA, err := loadDocument(pathA)
	if err != nil {
		t.Fatalf("loadDocument(a): %v", err)
	}
	docB, err := loadDocument(pathB)
	if err != nil {
		t.Fatalf("loadDocument(b): %v", err)
	}

	m = sized(t, newModel(), 80, 10)
	m = m.openDocument(docA)              // DV1, focus -> dv1Pane
	m = m.openDocumentInto(dv2Pane, docB) // DV2, focus -> dv2Pane
	if m.focus != dv2Pane {
		t.Fatalf("setup: focus = %v, want dv2Pane", m.focus)
	}
	return m, pathA, pathB
}

// TestFindIsIndependentPerViewer drives a committed Find into DV2 first,
// then into DV1, and checks neither disturbs the other's query, matches or
// editing state, in either order.
func TestFindIsIndependentPerViewer(t *testing.T) {
	m, _, _ := crossViewerFixture(t)

	m = press(t, m, "/")
	m = typeText(t, m, "bravo-marker")
	m = press(t, m, "enter")
	if !m.dv2.hasFindMatches() {
		t.Fatal("setup: DV2's find did not commit any matches")
	}
	if m.dv1.findEditing() || m.dv1.hasFindMatches() {
		t.Error("DV1 gained find state from a query driven at DV2")
	}

	m = press(t, m, keys.FocusDV1.Keys()[0])
	m = press(t, m, "/")
	m = typeText(t, m, "alpha-marker")
	m = press(t, m, "enter")

	if !m.dv1.hasFindMatches() {
		t.Fatal("DV1's own find did not commit")
	}
	if got := m.dv1.findState.input.Value(); got != "alpha-marker" {
		t.Errorf("DV1 query = %q, want %q", got, "alpha-marker")
	}
	if !m.dv2.hasFindMatches() {
		t.Error("DV2's committed find was lost when DV1 started its own")
	}
	if got := m.dv2.findState.input.Value(); got != "bravo-marker" {
		t.Errorf("DV2 query = %q, want unchanged %q", got, "bravo-marker")
	}
}

// TestFindEditingIsScopedToTheFocusedViewer checks the in-progress (not yet
// committed) half: typing a query into whichever viewer holds Focus must not
// mark the other viewer as findEditing, and switching Focus mid-query must
// not leak the keystrokes across.
func TestFindEditingIsScopedToTheFocusedViewer(t *testing.T) {
	m, _, _ := crossViewerFixture(t) // focus on dv2Pane

	m = press(t, m, "/")
	m = typeText(t, m, "bra")

	if !m.dv2.findEditing() {
		t.Fatal("DV2 is not editing a find after / and typing")
	}
	if m.dv1.findEditing() {
		t.Error("DV1 became findEditing from a query typed at DV2")
	}
	if m.dv1.hasFindMatches() {
		t.Error("DV1 gained matches from a query typed at DV2")
	}
}

// TestScrollIsIndependentPerViewer scrolls DV2 then DV1 and checks each
// keeps the other's offset untouched, in both directions.
func TestScrollIsIndependentPerViewer(t *testing.T) {
	m, _, _ := crossViewerFixture(t) // focus on dv2Pane
	if total := m.dv1.lineCount(); total <= m.dv1.height() {
		t.Fatalf("docA renders %d lines into a %d-line viewport; nothing to scroll", total, m.dv1.height())
	}
	if total := m.dv2.lineCount(); total <= m.dv2.height() {
		t.Fatalf("docB renders %d lines into a %d-line viewport; nothing to scroll", total, m.dv2.height())
	}
	dv1Before := m.dv1.offset()

	m = press(t, m, "down")
	m = press(t, m, "down")
	m = press(t, m, "down")

	dv2After := m.dv2.offset()
	if dv2After == 0 {
		t.Fatal("scrolling DV2 did not move its own offset")
	}
	if got := m.dv1.offset(); got != dv1Before {
		t.Errorf("DV1 offset after scrolling DV2 = %d, want unchanged %d", got, dv1Before)
	}

	m = press(t, m, keys.FocusDV1.Keys()[0])
	m = press(t, m, "down")

	if got := m.dv1.offset(); got == dv1Before {
		t.Error("scrolling DV1 (now focused) did not move its own offset")
	}
	if got := m.dv2.offset(); got != dv2After {
		t.Errorf("DV2 offset after scrolling DV1 = %d, want unchanged %d", got, dv2After)
	}
}

// TestFinishEditReopensOnlyItsOwnPane is the Edit half: finishEdit's pane
// field (startEdit's own focusedViewerPane at the time, per its doc comment)
// must reopen the reloaded document into that pane alone, leaving the other
// viewer's Active Document untouched — the write-side counterpart to
// startEdit capturing m.focusedViewerPane() (its source, read-only and
// already covered by every FocusDV1/FocusDV2 test).
func TestFinishEditReopensOnlyItsOwnPane(t *testing.T) {
	m, pathA, pathB := crossViewerFixture(t)

	got := m.finishEdit(editFinishedMsg{path: pathB, pane: dv2Pane})

	if got.dv2.doc.path != pathB {
		t.Errorf("DV2 Active Document = %q, want %q", got.dv2.doc.path, pathB)
	}
	if got.dv1.doc.path != pathA {
		t.Errorf("DV1 Active Document = %q, want unchanged %q", got.dv1.doc.path, pathA)
	}
	if got.focus != dv2Pane {
		t.Errorf("focus = %v, want dv2Pane (the edited pane)", got.focus)
	}
}
