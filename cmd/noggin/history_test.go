package main

import (
	"path/filepath"
	"testing"

	"noggin/internal/tree"
)

// C4/M6.1's own acceptance criterion, verbatim: each Document Viewer keeps
// its own session-only back/forward history with browser semantics —
// activating after going back discards forward entries, the same document
// may recur but never adjacently, the $EDITOR handoff and Refresh create no
// entries, traversal never changes the Active Root, and DV2's history dies
// with DV2. viewer.pushHistory/goBack/goForward carry the mechanics
// (viewer.go); these tests drive them through the real key bindings
// (keys.HistoryBack "[", keys.HistoryForward "]") and model.navigate's call
// sites, the same way every other S-level acceptance test in this codebase
// exercises a real m.Update rather than calling the mechanism directly.

// threeDocFixture is a single Root with three files, opened and returned
// with Focus on the Files Nav, top row selected — the common setup every
// test below activates from in some sequence via m.files.move + Activate.
func threeDocFixture(t *testing.T) (m model, dir string, docA, docB, docC string) {
	t.Helper()
	dir = t.TempDir()
	docA = writeFileContent(t, filepath.Join(dir, "a.md"), "doc a\n")
	docB = writeFileContent(t, filepath.Join(dir, "b.md"), "doc b\n")
	docC = writeFileContent(t, filepath.Join(dir, "c.md"), "doc c\n")

	node, err := tree.Walk(dir, tree.Options{Filter: tree.FilterMarkdown})
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	// A resolvable document store (docsPath) is what lets recordViewed's own
	// withDocRecord write succeed on every Activate below — without it,
	// Activate still opens the document (recordViewed and navigate are
	// independent calls), but leaves an unrelated "cannot save document
	// store" message behind that TestEditHandoffCreatesNoHistoryEntry's own
	// "no message" assertion would otherwise trip over.
	docsPath := filepath.Join(dir, "documents")
	m = sized(t, newModel().withDocs(docsPath, loadDocStore(docsPath)), 80, 24)
	m = m.openFolder(node)
	if got := m.files.selectedNode(); got == nil || got.Path != docA {
		t.Fatalf("setup: selected node = %+v, want %q", got, docA)
	}
	return m, dir, docA, docB, docC
}

// activateSelected presses Activate (dv1Pane target) on whatever the Files
// Nav currently has selected, first returning Focus there — Activate (like
// ActivateDV2 below) is matched only in Update's filesPane case, so it is a
// no-op once a previous Activate has already moved Focus onto a viewer, the
// same way every other Files-Nav-only key is.
func activateSelected(t *testing.T, m model) model {
	t.Helper()
	m = press(t, m, keys.FocusFiles.Keys()[0])
	return press(t, m, keys.Activate.Keys()[0])
}

// activateDV2Selected is activateSelected's ActivateDV2 counterpart.
func activateDV2Selected(t *testing.T, m model) model {
	t.Helper()
	m = press(t, m, keys.FocusFiles.Keys()[0])
	return press(t, m, keys.ActivateDV2.Keys()[0])
}

func historyPaths(v viewer) []string { return v.history }

// TestHistoryBackForwardWalksBothViewersIndependently is M6.1's own "tests
// drive DV1 and DV2 through independent back/forward sequences" clause: DV1
// visits a, b, c and DV2 (via ActivateDV2) visits the same three documents
// in reverse order, and walking one viewer's history leaves the other's
// position and document completely alone.
func TestHistoryBackForwardWalksBothViewersIndependently(t *testing.T) {
	m, _, docA, docB, docC := threeDocFixture(t)

	// DV1: a -> b -> c.
	m = activateSelected(t, m) // a.md, selects row 0
	m.files = m.files.move(1)
	m = activateSelected(t, m) // b.md
	m.files = m.files.move(1)
	m = activateSelected(t, m) // c.md
	if m.dv1.doc.path != docC {
		t.Fatalf("setup: DV1 = %q, want %q", m.dv1.doc.path, docC)
	}

	// DV2: c -> b -> a, via the DV2-targeting modifier, from the top of the
	// same three rows.
	m = press(t, m, keys.FocusFiles.Keys()[0])
	m.files = m.files.move(-2) // back to a.md (row 0)
	m.files = m.files.move(2)  // c.md
	m = activateDV2Selected(t, m)
	m = press(t, m, keys.FocusFiles.Keys()[0])
	m.files = m.files.move(-1) // b.md
	m = activateDV2Selected(t, m)
	m = press(t, m, keys.FocusFiles.Keys()[0])
	m.files = m.files.move(-1) // a.md
	m = activateDV2Selected(t, m)
	if m.focus != dv2Pane || m.dv2.doc.path != docA {
		t.Fatalf("setup: focus=%v dv2=%q, want dv2Pane/%q", m.focus, m.dv2.doc.path, docA)
	}

	// Walk DV2 back twice: DV2's own history is [c, b, a] (visited in that
	// order), currently on a, so Back walks a -> b -> c.
	m = press(t, m, keys.HistoryBack.Keys()[0])
	if m.dv2.doc.path != docB {
		t.Errorf("DV2 after one Back = %q, want %q", m.dv2.doc.path, docB)
	}
	m = press(t, m, keys.HistoryBack.Keys()[0])
	if m.dv2.doc.path != docC {
		t.Errorf("DV2 after two Backs = %q, want %q", m.dv2.doc.path, docC)
	}
	// DV1 must not have moved at all.
	if m.dv1.doc.path != docC {
		t.Errorf("DV1 moved while walking DV2's history: got %q, want %q", m.dv1.doc.path, docC)
	}

	// Now walk DV1 back twice: c -> b -> a, independent of DV2's own
	// position (still sitting on b, mid-stack).
	m = press(t, m, keys.FocusDV1.Keys()[0])
	m = press(t, m, keys.HistoryBack.Keys()[0])
	if m.dv1.doc.path != docB {
		t.Errorf("DV1 after one Back = %q, want %q", m.dv1.doc.path, docB)
	}
	m = press(t, m, keys.HistoryBack.Keys()[0])
	if m.dv1.doc.path != docA {
		t.Errorf("DV1 after two Backs = %q, want %q", m.dv1.doc.path, docA)
	}
	if m.dv2.doc.path != docC {
		t.Errorf("DV2 moved while walking DV1's history: got %q, want %q", m.dv2.doc.path, docC)
	}

	// Forward on DV1: a -> b -> c.
	m = press(t, m, keys.HistoryForward.Keys()[0])
	if m.dv1.doc.path != docB {
		t.Errorf("DV1 after one Forward = %q, want %q", m.dv1.doc.path, docB)
	}
	m = press(t, m, keys.HistoryForward.Keys()[0])
	if m.dv1.doc.path != docC {
		t.Errorf("DV1 after two Forwards = %q, want %q", m.dv1.doc.path, docC)
	}
	// Off the end: a silent no-op, not an error.
	m = press(t, m, keys.HistoryForward.Keys()[0])
	if m.dv1.doc.path != docC || m.message != "" {
		t.Errorf("DV1 Forward past the end = path %q message %q, want %q/empty", m.dv1.doc.path, m.message, docC)
	}
}

// TestHistoryBackIsANoOpAtTheStart covers the other boundary: Back with
// nothing before the current entry (including no document open at all)
// leaves the model untouched and reports no message, the same "nothing to
// do" shape FindNext/FindPrevious already give an empty match list.
func TestHistoryBackIsANoOpAtTheStart(t *testing.T) {
	m, _, docA, _, _ := threeDocFixture(t)

	// No document open yet at all.
	m = press(t, m, keys.FocusDV1.Keys()[0])
	m = press(t, m, keys.HistoryBack.Keys()[0])
	if m.dv1.doc.isOpen() || m.message != "" {
		t.Errorf("Back with no document open = doc open %v message %q, want closed/empty", m.dv1.doc.isOpen(), m.message)
	}

	m = press(t, m, keys.FocusFiles.Keys()[0])
	m = activateSelected(t, m) // a.md, the only history entry
	m = press(t, m, keys.HistoryBack.Keys()[0])
	if m.dv1.doc.path != docA || m.message != "" {
		t.Errorf("Back at the start of history = path %q message %q, want %q/empty", m.dv1.doc.path, m.message, docA)
	}
}

// TestActivatingAfterGoingBackDiscardsForwardEntries is S9's own browser
// semantics clause: a -> b -> c, back twice to a, then activating a new
// document d must drop b and c from history entirely — Forward from d has
// nothing to go to.
func TestActivatingAfterGoingBackDiscardsForwardEntries(t *testing.T) {
	m, dir, docA, _, _ := threeDocFixture(t)
	docD := writeFileContent(t, filepath.Join(dir, "d.md"), "doc d\n")
	m, _ = m.refresh() // pick up d.md, written after the fixture's own tree.Walk

	m = activateSelected(t, m) // a
	m.files = m.files.move(1)
	m = activateSelected(t, m) // b
	m.files = m.files.move(1)
	m = activateSelected(t, m) // c

	m = press(t, m, keys.FocusFiles.Keys()[0])
	m.files = m.files.move(-2) // a.md (row 0)
	m = press(t, m, keys.FocusDV1.Keys()[0])
	m = press(t, m, keys.HistoryBack.Keys()[0]) // c -> b
	m = press(t, m, keys.HistoryBack.Keys()[0]) // b -> a
	if m.dv1.doc.path != docA {
		t.Fatalf("setup: DV1 = %q, want %q", m.dv1.doc.path, docA)
	}
	if !m.dv1.canGoForward() {
		t.Fatalf("setup: expected Forward entries (b, c) still pending")
	}

	m = press(t, m, keys.FocusFiles.Keys()[0])
	m.files = m.files.move(3) // d.md, the newly written fourth row
	if got := m.files.selectedNode(); got == nil || got.Path != docD {
		t.Fatalf("setup: selected node = %+v, want %q", got, docD)
	}
	m = activateSelected(t, m) // d: a fresh navigation

	if m.dv1.canGoForward() {
		t.Errorf("Forward entries survived a new Activate after Back: history=%v pos=%d",
			historyPaths(m.dv1), m.dv1.historyPos)
	}
	wantHistory := []string{docA, docD}
	if got := historyPaths(m.dv1); len(got) != len(wantHistory) || got[0] != wantHistory[0] || got[1] != wantHistory[1] {
		t.Errorf("DV1 history after the discarding Activate = %v, want %v", got, wantHistory)
	}
}

// TestSameDocumentNeverAdjacentInHistory: activating the document already
// current is not a new entry — no adjacent duplicate is created, and Back
// still has nothing earlier to go to.
func TestSameDocumentNeverAdjacentInHistory(t *testing.T) {
	m, _, docA, docB, _ := threeDocFixture(t)

	m = activateSelected(t, m) // a
	m = activateSelected(t, m) // a again, same row still selected
	if got := historyPaths(m.dv1); len(got) != 1 || got[0] != docA {
		t.Fatalf("history after re-activating the current document = %v, want [%q]", got, docA)
	}
	if m.dv1.canGoBack() {
		t.Errorf("Back available after only a same-document re-activation")
	}

	m = press(t, m, keys.FocusFiles.Keys()[0])
	m.files = m.files.move(1)
	m = activateSelected(t, m) // b
	if got := historyPaths(m.dv1); len(got) != 2 || got[1] != docB {
		t.Fatalf("history after activating a different document = %v, want [.., %q]", got, docB)
	}
	m = press(t, m, keys.HistoryBack.Keys()[0])
	if m.dv1.doc.path != docA {
		t.Errorf("Back after b = %q, want %q", m.dv1.doc.path, docA)
	}
}

// TestEditHandoffCreatesNoHistoryEntry: the $EDITOR handoff (finishEdit)
// re-opens the same path after an edit, and must not push a second, adjacent
// entry for it — S9's own "the $EDITOR handoff ... do[es] not create
// entries".
func TestEditHandoffCreatesNoHistoryEntry(t *testing.T) {
	m, _, docA, docB, _ := threeDocFixture(t)

	m = activateSelected(t, m) // a
	m = press(t, m, keys.FocusFiles.Keys()[0])
	m.files = m.files.move(1)
	m = activateSelected(t, m) // b
	before := historyPaths(m.dv1)
	beforePos := m.dv1.historyPos

	m = m.finishEdit(editFinishedMsg{path: docB, pane: dv1Pane})
	if m.message != "" {
		t.Fatalf("finishEdit: unexpected message %q", m.message)
	}
	if got := historyPaths(m.dv1); len(got) != len(before) || got[len(got)-1] != before[len(before)-1] {
		t.Errorf("history after an edit handoff = %v, want unchanged %v", got, before)
	}
	if m.dv1.historyPos != beforePos {
		t.Errorf("historyPos after an edit handoff = %d, want unchanged %d", m.dv1.historyPos, beforePos)
	}
	m = press(t, m, keys.HistoryBack.Keys()[0])
	if m.dv1.doc.path != docA {
		t.Errorf("Back after an edit handoff = %q, want %q (history must still be [a, b])", m.dv1.doc.path, docA)
	}
}

// TestRefreshCreatesNoHistoryEntry: a Refresh re-render (refreshViewer, via
// refreshDoc not open) must not touch either viewer's history — S9's own
// "... and Refresh do not create entries".
func TestRefreshCreatesNoHistoryEntry(t *testing.T) {
	m, _, docA, docB, _ := threeDocFixture(t)

	m = activateSelected(t, m) // a
	m = press(t, m, keys.FocusFiles.Keys()[0])
	m.files = m.files.move(1)
	m = activateSelected(t, m) // b
	if m.dv1.doc.path != docB {
		t.Fatalf("setup: DV1 = %q, want %q", m.dv1.doc.path, docB)
	}
	before := historyPaths(m.dv1)
	beforePos := m.dv1.historyPos

	m, _ = m.refresh()
	if got := historyPaths(m.dv1); len(got) != len(before) || got[len(got)-1] != before[len(before)-1] {
		t.Errorf("history after refresh = %v, want unchanged %v", got, before)
	}
	if m.dv1.historyPos != beforePos {
		t.Errorf("historyPos after refresh = %d, want unchanged %d", m.dv1.historyPos, beforePos)
	}
	m = press(t, m, keys.HistoryBack.Keys()[0])
	if m.dv1.doc.path != docA {
		t.Errorf("Back after refresh = %q, want %q (history must still be [a, b])", m.dv1.doc.path, docA)
	}
}

// TestHistoryWalkNeverChangesActiveRoot is S9's own "History may walk into
// documents under other Roots and never changes the Active Root": DV1
// activates a document under Root A, the session then switches its Active
// Root to Root B and activates a document there too, and going Back into
// Root A's document must leave the Active Root at B.
func TestHistoryWalkNeverChangesActiveRoot(t *testing.T) {
	dir := t.TempDir()
	rootA := mkdir(t, filepath.Join(dir, "a-root"))
	rootB := mkdir(t, filepath.Join(dir, "b-root"))
	docA := writeFileContent(t, filepath.Join(rootA, "doc.md"), "doc a\n")
	docB := writeFileContent(t, filepath.Join(rootB, "doc.md"), "doc b\n")

	nodeA, err := tree.Walk(rootA, tree.Options{Filter: tree.FilterMarkdown})
	if err != nil {
		t.Fatalf("tree.Walk(a): %v", err)
	}
	nodeB, err := tree.Walk(rootB, tree.Options{Filter: tree.FilterMarkdown})
	if err != nil {
		t.Fatalf("tree.Walk(b): %v", err)
	}

	m := sized(t, newModel(), 80, 24)
	m.roots = m.roots.setRoots([]root{{path: rootA}, {path: rootB}}, 0)
	m = m.openFolder(nodeA)
	m = activateSelected(t, m) // docA into DV1, history=[docA]

	// Switch the Active Root to B, the same two steps model.activateRoot
	// itself performs (rootsnav's own activate plus a re-root).
	m.roots = m.roots.activate(1)
	m = m.openFolder(nodeB)
	m = activateSelected(t, m) // docB into DV1, history=[docA, docB]
	if m.dv1.doc.path != docB {
		t.Fatalf("setup: DV1 = %q, want %q", m.dv1.doc.path, docB)
	}
	if got := m.roots.activePath(); got != rootB {
		t.Fatalf("setup: Active Root = %q, want %q", got, rootB)
	}

	m = press(t, m, keys.FocusDV1.Keys()[0])
	m = press(t, m, keys.HistoryBack.Keys()[0])
	if m.dv1.doc.path != docA {
		t.Fatalf("Back = %q, want %q", m.dv1.doc.path, docA)
	}
	if got := m.roots.activePath(); got != rootB {
		t.Errorf("Active Root after walking history into another Root's document = %q, want unchanged %q", got, rootB)
	}
}

// TestDV2HistoryDiesWithDV2: closing DV2 and later re-creating it (an
// ordinary ActivateDV2) must start it with a fresh, empty history — nothing
// survives from before the close.
func TestDV2HistoryDiesWithDV2(t *testing.T) {
	m, _, docA, docB, docC := threeDocFixture(t)

	m = activateDV2Selected(t, m) // a into DV2
	m = press(t, m, keys.FocusFiles.Keys()[0])
	m.files = m.files.move(1)
	m = activateDV2Selected(t, m) // b into DV2
	if m.dv2.doc.path != docB || !m.dv2.canGoBack() {
		t.Fatalf("setup: DV2 = %q (canGoBack=%v), want %q/true", m.dv2.doc.path, m.dv2.canGoBack(), docB)
	}
	if got := historyPaths(*m.dv2); len(got) != 2 || got[0] != docA {
		t.Fatalf("setup: DV2 history = %v, want [%q, %q]", got, docA, docB)
	}

	m = press(t, m, keys.CloseViewer.Keys()[0])
	if m.dv2 != nil {
		t.Fatalf("setup: DV2 still exists after CloseViewer")
	}

	m = press(t, m, keys.FocusFiles.Keys()[0])
	m.files = m.files.move(1)
	m = activateDV2Selected(t, m) // c into a brand-new DV2
	if m.dv2.doc.path != docC {
		t.Fatalf("setup: DV2 = %q, want %q", m.dv2.doc.path, docC)
	}
	if m.dv2.canGoBack() {
		t.Errorf("re-created DV2 has Back available: history=%v pos=%d — the old DV2's history survived", historyPaths(*m.dv2), m.dv2.historyPos)
	}
	if got := historyPaths(*m.dv2); len(got) != 1 || got[0] != docC {
		t.Errorf("re-created DV2 history = %v, want [%q]", got, docC)
	}
}
