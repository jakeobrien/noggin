package main

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"noggin/internal/tree"
	"noggin/internal/wikipedia"
)

// --- filesNav.refreshRoot / nearestSurvivingIndex (S2, C2a/M5.1) -----------

// TestFilesNavRefreshRootPreservesExpansionAndSelectionWhenUnchanged is the
// refresh no-op case: re-walking a tree whose shape hasn't changed leaves
// expansion and selection exactly as they were, the same "nothing to lose"
// baseline setRoot's own restore-from-memory path already guarantees for a
// Root switch (M3.1) — refreshRoot must hold it for a live refresh too.
func TestFilesNavRefreshRootPreservesExpansionAndSelectionWhenUnchanged(t *testing.T) {
	f := newFiles(fixtureTree())
	f = f.expand() // open docs
	f = f.move(1)  // select guide.md, inside it
	wantExpanded, wantSelected := len(f.expanded), f.selectedNode().Name
	if wantSelected != "guide.md" {
		t.Fatalf("setup: selection = %q, want guide.md", wantSelected)
	}

	f = f.refreshRoot(fixtureTree()) // a structurally identical re-walk
	if len(f.expanded) != wantExpanded {
		t.Errorf("expansion after an unchanged refresh = %v, want %d folder(s) open as before", f.expanded, wantExpanded)
	}
	if got := f.selectedNode(); got == nil || got.Name != wantSelected {
		t.Errorf("selection after an unchanged refresh = %v, want %q", got, wantSelected)
	}
}

// TestFilesNavRefreshRootSelectsNearestSurvivingAncestorWhenNodeGone is S2's
// own rule: the selected node vanishing under a refresh moves selection to
// its nearest surviving ancestor, not to the top of the list — here, guide.md
// disappears but its parent docs (still expanded) survives.
func TestFilesNavRefreshRootSelectsNearestSurvivingAncestorWhenNodeGone(t *testing.T) {
	f := newFiles(fixtureTree())
	f = f.expand() // open docs
	f = f.move(1)  // select guide.md
	if f.selectedNode().Name != "guide.md" {
		t.Fatalf("setup: selection = %q, want guide.md", f.selectedNode().Name)
	}

	after := &tree.Node{
		Name: "root", Path: "/root", IsDir: true,
		Children: []*tree.Node{
			{Name: "docs", Path: "/root/docs", IsDir: true, Children: []*tree.Node{
				{Name: "ref.md", Path: "/root/docs/ref.md"}, // guide.md is gone
			}},
			{Name: "empty", Path: "/root/empty", IsDir: true},
			{Name: "README.md", Path: "/root/README.md"},
			{Name: "notes.txt", Path: "/root/notes.txt"},
		},
	}
	f = f.refreshRoot(after)
	if got := f.selectedNode(); got == nil || got.Name != "docs" {
		t.Errorf("selection after guide.md vanished = %v, want docs (its surviving parent)", got)
	}
	// docs must still be expanded — refreshRoot never touches the expansion
	// set itself, only where selection points into it.
	if !strings.Contains(strings.Join(rowNames(f), ","), "ref.md") {
		t.Errorf("rows after refresh = %v, want docs still expanded (ref.md visible)", rowNames(f))
	}
}

// TestFilesNavRefreshRootFallsBackToTopWhenWholeBranchGone checks the far
// end of the ancestor walk: when even the top-level entry the selection sat
// under is gone, there is no surviving ancestor at all, and selection falls
// back to the top of the list — the same fallback setRoot's fresh-Root case
// already uses.
func TestFilesNavRefreshRootFallsBackToTopWhenWholeBranchGone(t *testing.T) {
	f := newFiles(fixtureTree())
	f = f.expand() // open docs
	f = f.move(1)  // select guide.md

	after := &tree.Node{
		Name: "root", Path: "/root", IsDir: true,
		Children: []*tree.Node{
			{Name: "empty", Path: "/root/empty", IsDir: true},
			{Name: "README.md", Path: "/root/README.md"},
			{Name: "notes.txt", Path: "/root/notes.txt"},
		},
	}
	f = f.refreshRoot(after)
	if f.selected != 0 {
		t.Errorf("selected index after the whole docs branch vanished = %d, want 0", f.selected)
	}
	if got := f.selectedNode(); got == nil || got.Name != "empty" {
		t.Errorf("selection after the whole docs branch vanished = %v, want empty (top of the new list)", got)
	}
}

// --- viewer.refreshDoc / close (S2, C2a/M5.1) -------------------------------

// listContent builds n bullet list items, one per rendered line at any
// reasonable width, so a scroll offset set against it is deterministic —
// unlike prose, a markdown list never re-flows two items onto the same line.
func listContent(n int) string {
	var b strings.Builder
	for i := 1; i <= n; i++ {
		b.WriteString("- item ")
		b.WriteString(strconv.Itoa(i))
		b.WriteString("\n")
	}
	return b.String()
}

// TestViewerRefreshDocPreservesScrollPosition checks S2's scroll-preserving
// half directly: re-reading an unchanged document and re-rendering through
// refreshDoc leaves the offset exactly where it was, unlike open()'s own
// GotoTop().
func TestViewerRefreshDocPreservesScrollPosition(t *testing.T) {
	path := writeFileContent(t, filepath.Join(t.TempDir(), "doc.md"), listContent(60))
	d, err := loadDocument(path)
	if err != nil {
		t.Fatalf("loadDocument: %v", err)
	}
	v := newViewer().setSize(40, 10).open(d)
	if total := v.lineCount(); total <= v.height() {
		t.Fatalf("fixture renders %d lines into a %d-line viewport; nothing to preserve a scroll over", total, v.height())
	}
	v = v.withOffset(5)

	reread, err := loadDocument(path) // an unchanged re-read, same as refresh's own path
	if err != nil {
		t.Fatalf("loadDocument (re-read): %v", err)
	}
	v = v.refreshDoc(reread)
	if got := v.offset(); got != 5 {
		t.Errorf("offset after refreshDoc of an unchanged document = %d, want 5 (preserved)", got)
	}
}

// TestViewerRefreshDocClampsScrollWhenDocumentShrank checks S2's other half:
// a document that shrank past the current offset clamps to the end rather
// than pointing past it.
func TestViewerRefreshDocClampsScrollWhenDocumentShrank(t *testing.T) {
	path := filepath.Join(t.TempDir(), "doc.md")
	writeFileContent(t, path, listContent(60))
	d, err := loadDocument(path)
	if err != nil {
		t.Fatalf("loadDocument: %v", err)
	}
	v := newViewer().setSize(40, 10).open(d)
	v = v.withOffset(30)
	if v.offset() != 30 {
		t.Fatalf("setup: offset = %d, want 30", v.offset())
	}

	writeFileContent(t, path, listContent(3)) // shrinks well past the old offset
	shrunk, err := loadDocument(path)
	if err != nil {
		t.Fatalf("loadDocument (shrunk): %v", err)
	}
	v = v.refreshDoc(shrunk)
	if !v.atBottom() {
		t.Errorf("offset after the document shrank = %d, want clamped to the end", v.offset())
	}
}

// TestViewerCloseEmptiesTheViewer checks refresh's other outcome for the
// Active Document (S2): a vanished/unreadable document leaves the viewer in
// its empty state, not showing stale content from before.
func TestViewerCloseEmptiesTheViewer(t *testing.T) {
	v := newViewer().setSize(60, 10).open(fixtureDocFor(t))
	if !v.doc.isOpen() {
		t.Fatalf("setup: doc not open")
	}
	v = v.close()
	if v.doc.isOpen() {
		t.Errorf("doc still open after close()")
	}
	if !strings.Contains(v.view(), emptyViewerMarker) {
		t.Errorf("view after close() = %q, want the empty state", v.view())
	}
}

func fixtureDocFor(t *testing.T) document {
	t.Helper()
	d, err := loadDocument(fixturePath)
	if err != nil {
		t.Fatalf("loadDocument(%s): %v", fixturePath, err)
	}
	return d
}

// --- model.refresh end to end (S2, C2a/M5.1's own acceptance criterion) ----

// TestRefreshShowsFilesystemChangesPreservingNavAndScrollPosition is the
// cycle's stated acceptance for S2/M5.1 verbatim: creating, editing and
// deleting files outside noggin under the Active Root, then pressing R,
// shows every change in both panes without losing tree expansion, selection
// (or its nearest surviving ancestor) or reading position.
func TestRefreshShowsFilesystemChangesPreservingNavAndScrollPosition(t *testing.T) {
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	sub := mkdir(t, filepath.Join(root, "sub"))
	writeFile(t, filepath.Join(sub, "keep.md"))
	doomedPath := writeFile(t, filepath.Join(sub, "doomed.md"))
	docPath := writeFileContent(t, filepath.Join(root, "doc.md"), listContent(60))

	node, err := tree.Walk(root, tree.Options{Filter: tree.FilterMarkdown})
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	m := sized(t, newModel().openFolder(node), 80, 10)

	// Expand sub, select doomed.md inside it (sub sorts first: dirs-first).
	if m.files.selectedNode().Name != "sub" {
		t.Fatalf("setup: initial selection = %q, want sub", m.files.selectedNode().Name)
	}
	m.files = m.files.expand()
	m.files = m.files.move(1) // sub -> doomed.md (sorted before keep.md)
	if m.files.selectedNode().Name != "doomed.md" {
		t.Fatalf("setup: selection = %q, want doomed.md", m.files.selectedNode().Name)
	}

	d, err := loadDocument(docPath)
	if err != nil {
		t.Fatalf("loadDocument: %v", err)
	}
	m = m.openDocument(d)
	if total := m.dv1.lineCount(); total <= m.dv1.height() {
		t.Fatalf("doc.md renders %d lines into a %d-line viewport; nothing to preserve a scroll over", total, m.dv1.height())
	}
	m.dv1 = m.dv1.withOffset(10)
	wantOffset := m.dv1.offset()

	// Outside noggin: delete doomed.md, add new.md, append to doc.md.
	if err := os.Remove(doomedPath); err != nil {
		t.Fatalf("remove doomed.md: %v", err)
	}
	writeFile(t, filepath.Join(sub, "new.md"))
	appendToFile(t, docPath, listContent(60)) // grows the document further, never shrinks past the offset

	m = press(t, m, keys.Refresh.Keys()[0])

	if m.message != "" {
		t.Errorf("refresh produced a message %q, want silence (nothing here should fail)", m.message)
	}
	rows := rowNames(m.files)
	if !strings.Contains(strings.Join(rows, ","), "new.md") {
		t.Errorf("rows after refresh = %v, want new.md picked up from disk", rows)
	}
	if strings.Contains(strings.Join(rows, ","), "doomed.md") {
		t.Errorf("rows after refresh = %v, want doomed.md gone", rows)
	}
	// doomed.md's own row is gone; selection must have moved to its nearest
	// surviving ancestor, sub, not reset to the top of the list.
	if got := m.files.selectedNode(); got == nil || got.Name != "sub" {
		t.Errorf("selection after refresh = %v, want sub (doomed.md's surviving parent)", got)
	}
	if got := m.dv1.offset(); got != wantOffset {
		t.Errorf("viewer offset after refresh = %d, want %d (preserved)", got, wantOffset)
	}
}

// TestRefreshEmptiesTheViewerWhenTheActiveDocumentIsDeleted is S2's other
// stated outcome: the Active Document no longer existing after a refresh
// leaves the Document Viewer empty rather than showing stale content, and
// the session keeps running (a footer message, not a crash — T15's shape).
func TestRefreshEmptiesTheViewerWhenTheActiveDocumentIsDeleted(t *testing.T) {
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	docPath := writeFile(t, filepath.Join(root, "doc.md"))

	node, err := tree.Walk(root, tree.Options{Filter: tree.FilterMarkdown})
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	d, err := loadDocument(docPath)
	if err != nil {
		t.Fatalf("loadDocument: %v", err)
	}
	m := sized(t, newModel().openFolder(node), 80, 24)
	m = m.openDocument(d)
	if !m.dv1.doc.isOpen() {
		t.Fatalf("setup: document not open")
	}

	if err := os.Remove(docPath); err != nil {
		t.Fatalf("remove doc.md: %v", err)
	}
	m = press(t, m, keys.Refresh.Keys()[0])

	if m.dv1.doc.isOpen() {
		t.Errorf("viewer still reports a document open after its file was deleted")
	}
	if !strings.Contains(m.dv1.view(), emptyViewerMarker) {
		t.Errorf("viewer view = %q, want the empty state", m.dv1.view())
	}
	if m.message == "" {
		t.Error("message is empty after a refresh whose Active Document vanished, want a footer message")
	}
}

// TestRefreshRecomputesGitStatusMarkers is S8's own recompute promise,
// exercised through refresh rather than a Root switch (M4.1's own test
// covers that path): a file that becomes dirty after the initial load shows
// its marker only once R is pressed.
func TestRefreshRecomputesGitStatusMarkers(t *testing.T) {
	requireGit(t)
	dir := newGitRepo(t) // tracked.md (M), sub/untracked.md (A), clean.md unmarked

	node, err := tree.Walk(dir, tree.Options{Filter: tree.FilterMarkdown})
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	m := sized(t, newModel().openFolder(node), 80, 24)
	if strings.Contains(lineContaining(strings.Split(m.files.view(), "\n"), "clean.md"), "M   clean.md") {
		t.Fatalf("setup: clean.md already marked before it was dirtied")
	}

	writeFileContent(t, filepath.Join(dir, "clean.md"), "now dirty\n")
	m = press(t, m, keys.Refresh.Keys()[0])

	view := m.files.view()
	// Three spaces, not one, between M and the name: the status column's
	// own trailing space plus C4/M2.3's star marker column between it and
	// the name.
	if !strings.Contains(lineContaining(strings.Split(view, "\n"), "clean.md"), "M   clean.md") {
		t.Errorf("Files Nav view after refresh = %q, want clean.md marked M", view)
	}
}

// TestRefreshAfterCommittedFindSurfacesMessage is T43's rule (resize's own
// doc comment) applied to refresh: re-rendering the Active Document drops a
// committed find's matches the same way any other re-render already does
// (viewer.render's own doc comment), and losing that silently would be the
// same kind of confusion T43 was written against — so refresh must say so.
func TestRefreshAfterCommittedFindSurfacesMessage(t *testing.T) {
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	docPath := writeFile(t, filepath.Join(root, "doc.md")) // content is "# note\n"
	node, err := tree.Walk(root, tree.Options{Filter: tree.FilterMarkdown})
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	d, err := loadDocument(docPath)
	if err != nil {
		t.Fatalf("loadDocument: %v", err)
	}
	m := sized(t, newModel().openFolder(node), 80, 24)
	m = m.openDocument(d)

	m = press(t, m, "/")
	m = typeText(t, m, "note")
	m = press(t, m, "enter")
	if m.dv1.findEditing() {
		t.Fatal("test setup: enter did not commit the find")
	}
	if !m.dv1.hasFindMatches() {
		t.Fatal("test setup: committed find has no matches to lose")
	}

	m = press(t, m, keys.Refresh.Keys()[0])
	if m.dv1.hasFindMatches() {
		t.Fatal("test setup: refresh did not actually invalidate the committed find")
	}
	if !strings.Contains(m.message, "search") {
		t.Errorf("message = %q, want it to explain the cleared search", m.message)
	}
}

// --- model.refresh with DV2 (C3/M3.1's own acceptance criterion) -----------

// TestRefreshUpdatesBothViewersPreservingEachScrollPosition is M3.1's stated
// acceptance verbatim: with both DV1 and DV2 open on distinct documents,
// changing both on disk and pressing R updates the tree and both viewers'
// content in the same single action, preserving each viewer's own scroll
// position — the S6 independence guarantee M1.6 already gives Find/Edit/
// scroll must hold for a refresh too.
func TestRefreshUpdatesBothViewersPreservingEachScrollPosition(t *testing.T) {
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	docAPath := writeFileContent(t, filepath.Join(root, "a.md"), listContent(60))
	docBPath := writeFileContent(t, filepath.Join(root, "b.md"), listContent(60))

	node, err := tree.Walk(root, tree.Options{Filter: tree.FilterMarkdown})
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	m := sized(t, newModel().openFolder(node), 80, 10)

	docA, err := loadDocument(docAPath)
	if err != nil {
		t.Fatalf("loadDocument(a): %v", err)
	}
	docB, err := loadDocument(docBPath)
	if err != nil {
		t.Fatalf("loadDocument(b): %v", err)
	}
	m = m.openDocument(docA)              // DV1
	m = m.openDocumentInto(dv2Pane, docB) // DV2
	if total := m.dv1.lineCount(); total <= m.dv1.height() {
		t.Fatalf("a.md renders %d lines into a %d-line viewport; nothing to preserve a scroll over", total, m.dv1.height())
	}
	if total := m.dv2.lineCount(); total <= m.dv2.height() {
		t.Fatalf("b.md renders %d lines into a %d-line viewport; nothing to preserve a scroll over", total, m.dv2.height())
	}
	m.dv1 = m.dv1.withOffset(10)
	v := m.dv2.withOffset(20)
	m.dv2 = &v
	wantDV1Offset, wantDV2Offset := m.dv1.offset(), m.dv2.offset()
	beforeDV1Lines, beforeDV2Lines := m.dv1.lineCount(), m.dv2.lineCount()

	// Outside noggin: both documents grow (never shrinking past either
	// offset), plus a new file appears in the tree.
	appendToFile(t, docAPath, listContent(60))
	appendToFile(t, docBPath, listContent(60))
	writeFile(t, filepath.Join(root, "new.md"))

	m = press(t, m, keys.Refresh.Keys()[0])

	if m.message != "" {
		t.Errorf("refresh produced a message %q, want silence (nothing here should fail)", m.message)
	}
	rows := rowNames(m.files)
	if !strings.Contains(strings.Join(rows, ","), "new.md") {
		t.Errorf("rows after refresh = %v, want new.md picked up from disk", rows)
	}
	if got := m.dv1.lineCount(); got <= beforeDV1Lines {
		t.Errorf("DV1 line count after refresh = %d, want more than %d (grown content picked up)", got, beforeDV1Lines)
	}
	if got := m.dv2.lineCount(); got <= beforeDV2Lines {
		t.Errorf("DV2 line count after refresh = %d, want more than %d (grown content picked up)", got, beforeDV2Lines)
	}
	if got := m.dv1.offset(); got != wantDV1Offset {
		t.Errorf("DV1 offset after refresh = %d, want %d (preserved)", got, wantDV1Offset)
	}
	if got := m.dv2.offset(); got != wantDV2Offset {
		t.Errorf("DV2 offset after refresh = %d, want %d (preserved)", got, wantDV2Offset)
	}
}

// TestRefreshCombinesBothViewersDroppedFindIntoOneMessage is the two-viewer
// extension of TestRefreshAfterCommittedFindSurfacesMessage: a refresh that
// drops a committed find in *both* DV1 and DV2 at once must name both,
// mirroring resize's own dv1Drop/dv2Drop combination (model.go's resize) —
// S6's independence guarantee has to hold at this messaging layer too, not
// only in routing.
func TestRefreshCombinesBothViewersDroppedFindIntoOneMessage(t *testing.T) {
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	docAPath := writeFile(t, filepath.Join(root, "a.md")) // "# note\n"
	docBPath := writeFile(t, filepath.Join(root, "b.md"))

	node, err := tree.Walk(root, tree.Options{Filter: tree.FilterMarkdown})
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	m := sized(t, newModel().openFolder(node), 80, 24)
	docA, err := loadDocument(docAPath)
	if err != nil {
		t.Fatalf("loadDocument(a): %v", err)
	}
	docB, err := loadDocument(docBPath)
	if err != nil {
		t.Fatalf("loadDocument(b): %v", err)
	}
	m = m.openDocument(docA)              // focus -> dv1Pane
	m = m.openDocumentInto(dv2Pane, docB) // focus -> dv2Pane

	m = press(t, m, keys.FocusDV1.Keys()[0])
	m = press(t, m, "/")
	m = typeText(t, m, "note")
	m = press(t, m, "enter")
	if !m.dv1.hasFindMatches() {
		t.Fatal("test setup: DV1's find did not commit")
	}
	m = press(t, m, keys.FocusDV2.Keys()[0])
	m = press(t, m, "/")
	m = typeText(t, m, "note")
	m = press(t, m, "enter")
	if !m.dv2.hasFindMatches() {
		t.Fatal("test setup: DV2's find did not commit")
	}

	m = press(t, m, keys.Refresh.Keys()[0])

	if m.dv1.hasFindMatches() || m.dv2.hasFindMatches() {
		t.Fatal("test setup: refresh did not invalidate both committed finds")
	}
	if !strings.Contains(m.message, "dv1 search cleared") || !strings.Contains(m.message, "dv2 search cleared") {
		t.Errorf("message = %q, want both viewers' dropped search named", m.message)
	}
}

// TestRefreshDV1ReadFailureWinsOverDV2ReadFailure pins refresh's own stated
// priority (its doc comment) for the case both Active Documents vanish in
// the same refresh: DV1's read failure is the one reported, not DV2's,
// matching the "DV1 first" precedence resize's own dv1Drop/dv2Drop give the
// two viewers.
func TestRefreshDV1ReadFailureWinsOverDV2ReadFailure(t *testing.T) {
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	docAPath := writeFile(t, filepath.Join(root, "a.md"))
	docBPath := writeFile(t, filepath.Join(root, "b.md"))

	node, err := tree.Walk(root, tree.Options{Filter: tree.FilterMarkdown})
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	m := sized(t, newModel().openFolder(node), 80, 24)
	docA, err := loadDocument(docAPath)
	if err != nil {
		t.Fatalf("loadDocument(a): %v", err)
	}
	docB, err := loadDocument(docBPath)
	if err != nil {
		t.Fatalf("loadDocument(b): %v", err)
	}
	m = m.openDocument(docA)
	m = m.openDocumentInto(dv2Pane, docB)

	if err := os.Remove(docAPath); err != nil {
		t.Fatalf("remove a.md: %v", err)
	}
	if err := os.Remove(docBPath); err != nil {
		t.Fatalf("remove b.md: %v", err)
	}

	m = press(t, m, keys.Refresh.Keys()[0])

	if m.dv1.doc.isOpen() || m.dv2.doc.isOpen() {
		t.Error("a viewer still reports a document open after its file was deleted")
	}
	if !strings.Contains(m.message, docAPath) {
		t.Errorf("message = %q, want it naming DV1's path %q (DV1 wins the priority)", m.message, docAPath)
	}
	if strings.Contains(m.message, docBPath) {
		t.Errorf("message = %q, want it not also naming DV2's path %q", m.message, docBPath)
	}
}

// TestRefreshReadFailureMessageWinsOverFindDroppedMessage is TRIAGE T11
// (C3/M3.1a): a document that fails to re-read while it had a
// committed Find used to lose its read-failure message to the find-dropped
// one, because refresh's message-combining switch had no "already set"
// guard, unlike refreshViewer's own read-failure branch (which does) and
// resize's near-identical switch (which never needed one, since resize
// never sets m.message before it runs) — refresh's own doc comment already
// promised the read failure wins; this pins that it actually does, in the
// single-viewer case.
func TestRefreshReadFailureMessageWinsOverFindDroppedMessage(t *testing.T) {
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	docPath := writeFile(t, filepath.Join(root, "doc.md")) // "# note\n"

	node, err := tree.Walk(root, tree.Options{Filter: tree.FilterMarkdown})
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	d, err := loadDocument(docPath)
	if err != nil {
		t.Fatalf("loadDocument: %v", err)
	}
	m := sized(t, newModel().openFolder(node), 80, 24)
	m = m.openDocument(d)

	m = press(t, m, "/")
	m = typeText(t, m, "note")
	m = press(t, m, "enter")
	if !m.dv1.hasFindMatches() {
		t.Fatal("test setup: committed find has no matches to lose")
	}

	if err := os.Remove(docPath); err != nil {
		t.Fatalf("remove doc.md: %v", err)
	}
	m = press(t, m, keys.Refresh.Keys()[0])

	if m.dv1.doc.isOpen() {
		t.Errorf("viewer still reports a document open after its file was deleted")
	}
	if !strings.Contains(m.message, "cannot open") || !strings.Contains(m.message, docPath) {
		t.Errorf("message = %q, want the read-failure message naming %q", m.message, docPath)
	}
	if strings.Contains(m.message, "search") {
		t.Errorf("message = %q, want the read failure to win outright, not share the slot with the dropped find", m.message)
	}
}

// TestRefreshReadFailureMessageWinsOverFindDroppedMessageWithDV2 is the
// dual-viewer extension of the test above: DV1's committed find is dropped
// by refreshViewer's own v.close() at the exact same moment DV1's read
// fails, while DV2 (untouched on disk) drops its own committed find via a
// perfectly ordinary successful re-render — two genuine find-drops "racing"
// one read failure. The read failure must still win the single m.message
// slot over both.
func TestRefreshReadFailureMessageWinsOverFindDroppedMessageWithDV2(t *testing.T) {
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	docAPath := writeFile(t, filepath.Join(root, "a.md")) // "# note\n"
	docBPath := writeFile(t, filepath.Join(root, "b.md"))

	node, err := tree.Walk(root, tree.Options{Filter: tree.FilterMarkdown})
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	docA, err := loadDocument(docAPath)
	if err != nil {
		t.Fatalf("loadDocument(a): %v", err)
	}
	docB, err := loadDocument(docBPath)
	if err != nil {
		t.Fatalf("loadDocument(b): %v", err)
	}
	m := sized(t, newModel().openFolder(node), 80, 24)
	m = m.openDocument(docA)              // focus -> dv1Pane
	m = m.openDocumentInto(dv2Pane, docB) // focus -> dv2Pane

	m = press(t, m, keys.FocusDV1.Keys()[0])
	m = press(t, m, "/")
	m = typeText(t, m, "note")
	m = press(t, m, "enter")
	if !m.dv1.hasFindMatches() {
		t.Fatal("test setup: DV1's find did not commit")
	}
	m = press(t, m, keys.FocusDV2.Keys()[0])
	m = press(t, m, "/")
	m = typeText(t, m, "note")
	m = press(t, m, "enter")
	if !m.dv2.hasFindMatches() {
		t.Fatal("test setup: DV2's find did not commit")
	}

	if err := os.Remove(docAPath); err != nil {
		t.Fatalf("remove a.md: %v", err)
	}
	m = press(t, m, keys.Refresh.Keys()[0])

	if m.dv1.doc.isOpen() {
		t.Errorf("DV1 still reports a document open after its file was deleted")
	}
	if m.dv1.hasFindMatches() || m.dv2.hasFindMatches() {
		t.Fatal("test setup: refresh did not drop both viewers' committed finds")
	}
	if !strings.Contains(m.message, "cannot open") || !strings.Contains(m.message, docAPath) {
		t.Errorf("message = %q, want the read-failure message naming %q", m.message, docAPath)
	}
	if strings.Contains(m.message, "search") {
		t.Errorf("message = %q, want the read failure to win outright over both viewers' dropped finds", m.message)
	}
}

// TestRefreshWithNoActiveRootIsANoOp checks the "nothing outside the Active
// Root's tree and the Active Document is touched" clause at its emptiest: a
// session with no root at all (no launch argument, no starred roots) has
// nothing for R to refresh, and pressing it must not panic or misbehave.
func TestRefreshWithNoActiveRootIsANoOp(t *testing.T) {
	m := sized(t, newModel(), 80, 24)
	m = press(t, m, keys.Refresh.Keys()[0])
	if m.files.root != nil {
		t.Errorf("files root after refreshing with none = %v, want nil still", m.files.root)
	}
	if m.dv1.doc.isOpen() {
		t.Errorf("viewer document after refreshing with none = %v, want none still", m.dv1.doc)
	}
	if m.message != "" {
		t.Errorf("refresh with nothing to refresh produced a message %q, want silence", m.message)
	}
}

// TestRefreshWorksWithFocusOnRootsPane is TRIAGE T4: keys.Refresh is matched
// in model.Update's global switch, unconditional on m.focus (unlike
// FocusRoots/FocusFiles' every-pane table in focus_test.go, there is no
// existing test that drives Refresh specifically with focus on rootsPane —
// only filesPane and dv1Pane were exercised. This pins S2's "reachable
// from any pane" for the one pane the existing suite skipped.
func TestRefreshWorksWithFocusOnRootsPane(t *testing.T) {
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	writeFile(t, filepath.Join(root, "keep.md"))

	node, err := tree.Walk(root, tree.Options{Filter: tree.FilterMarkdown})
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	m := sized(t, newModel().openFolder(node), 80, 24)
	m.focus = rootsPane

	writeFile(t, filepath.Join(root, "new.md")) // outside noggin, before refresh

	m = press(t, m, keys.Refresh.Keys()[0])

	if m.focus != rootsPane {
		t.Errorf("focus after Refresh = %v, want unchanged rootsPane", m.focus)
	}
	rows := rowNames(m.files)
	if !strings.Contains(strings.Join(rows, ","), "new.md") {
		t.Errorf("rows after refresh with focus on rootsPane = %v, want new.md picked up from disk", rows)
	}
}

// TestRefreshDoesNotTouchOtherRoots is TRIAGE T5: S2's negative clause
// ("nothing else is refreshed") holds today only by construction —
// model.refresh only ever assigns to m.files/m.dv1 — but nothing
// regression-proofs it. With two roots configured, deletes the inactive
// root's directory (which would flip its Sources Nav broken mark if it were
// ever recomputed) and adds a file under the Active Root, then refreshes:
// the Active Root's tree picks up the new file, while the inactive root's
// row is untouched down to its stale (pre-deletion) broken snapshot.
func TestRefreshDoesNotTouchOtherRoots(t *testing.T) {
	active := mkdir(t, filepath.Join(t.TempDir(), "active"))
	writeFile(t, filepath.Join(active, "keep.md"))
	inactive := mkdir(t, filepath.Join(t.TempDir(), "inactive"))

	m := sized(t, newModel(), 80, 24)
	m.roots = m.roots.setRoots([]root{{path: active}, {path: inactive}}, -1)
	m = m.activateRoot() // selection starts on row 0 (active)
	if m.files.root == nil || m.files.root.Path != active {
		t.Fatalf("setup: files root = %v, want %s active", m.files.root, active)
	}
	if m.roots.rows[1].broken {
		t.Fatalf("setup: inactive root already marked broken")
	}

	// Outside noggin: the inactive root's directory disappears, and a new
	// file appears under the Active Root.
	if err := os.RemoveAll(inactive); err != nil {
		t.Fatalf("remove inactive root: %v", err)
	}
	writeFile(t, filepath.Join(active, "new.md"))

	m = press(t, m, keys.Refresh.Keys()[0])

	rows := rowNames(m.files)
	if !strings.Contains(strings.Join(rows, ","), "new.md") {
		t.Errorf("Active Root rows after refresh = %v, want new.md picked up", rows)
	}
	if len(m.roots.rows) != 2 {
		t.Fatalf("roots after refresh = %d rows, want 2 (untouched)", len(m.roots.rows))
	}
	if m.roots.rows[1].root.path != inactive {
		t.Errorf("inactive root path after refresh = %q, want unchanged %q", m.roots.rows[1].root.path, inactive)
	}
	if m.roots.rows[1].broken {
		t.Errorf("inactive root's broken mark flipped true after a refresh that should never re-check it (its own snapshot, taken before deletion, must stand)")
	}
}

// --- model.refresh with a Wikipedia article (S3, C6/M4.3) ------------------

// TestRefreshRefetchesAnOpenArticleBypassingTheCache is S3/C6/M4.3's own
// acceptance clause verbatim: R re-fetches an open Wikipedia article,
// bypassing the session cache (Fetch's own refresh argument), rather than
// trying to re-read a URL from disk (loadDocument would fail loudly on
// that, a known landmine) or serving a stale cached copy. The UI
// stays responsive in the meantime — the same in-progress indicator
// Activate/Peek's own fetch already shows (M4.2) — and the fetched content
// replaces the article in place once it completes.
func TestRefreshRefetchesAnOpenArticleBypassingTheCache(t *testing.T) {
	updated := wikipedia.Article{Title: eulerArticle.Title, URL: eulerArticle.URL, Markdown: "# Leonhard Euler\n\nRe-fetched content.\n"}
	fetcher := &fakeWikipediaFetcher{articles: map[string]wikipedia.Article{euler.Title: updated}}
	m := sized(t, newModel(), 80, 24).withWikipediaFetcher(fetcher)
	m = m.navigate(dv1Pane, fixtureArticle())
	if got := m.dv1.view(); !strings.Contains(got, "Leonhard Euler was a Swiss") {
		t.Fatalf("setup: DV1 view = %q, want the original article content", got)
	}

	after, cmd := m.refresh()
	if cmd == nil {
		t.Fatal("refresh() returned a nil Cmd with an article open, want the re-fetch dispatched")
	}
	if !after.dv1.fetching() {
		t.Error("dv1.fetching() = false immediately after refresh, want the in-progress indicator up")
	}
	if after.dv1.doc.path != fixtureArticle().path {
		t.Errorf("DV1 Active Document path changed before the re-fetch completed: %q", after.dv1.doc.path)
	}
	if len(fetcher.calls) != 0 {
		t.Fatalf("Fetch calls = %v before the Cmd has run, want none yet", fetcher.calls)
	}

	got := runArticleFetch(t, after, cmd)

	if !strings.Contains(got.dv1.view(), "Re-fetched content") {
		t.Errorf("view after the refresh completed = %q, want the re-fetched content", got.dv1.view())
	}
	if got.dv1.fetching() {
		t.Error("dv1.fetching() = true after the re-fetch completed, want the indicator cleared")
	}
	if got.message != "" {
		t.Errorf("message = %q after a successful article refresh, want silence", got.message)
	}
	if len(fetcher.calls) != 1 || fetcher.calls[0] != euler.Title {
		t.Fatalf("Fetch calls = %v, want exactly one call for %q", fetcher.calls, euler.Title)
	}
	if !fetcher.refreshFlags[0] {
		t.Error("Fetch's own refresh argument = false, want true (the cache must be bypassed)")
	}
}

// TestRefreshFailedArticleFetchLeavesTheArticleUntouchedAndReportsMessage is
// S3's failure clause applied to Refresh ("a failed fetch ... leaves the
// viewer's current document untouched"): the article stays exactly as it
// was before R was pressed, and the failure surfaces as a Footer message,
// the same shape Activate's own failed fetch already gives (M4.2).
func TestRefreshFailedArticleFetchLeavesTheArticleUntouchedAndReportsMessage(t *testing.T) {
	fetcher := &fakeWikipediaFetcher{errs: map[string]error{euler.Title: errors.New("offline")}}
	m := sized(t, newModel(), 80, 24).withWikipediaFetcher(fetcher)
	m = m.navigate(dv1Pane, fixtureArticle())
	beforeView := m.dv1.view()

	after, cmd := m.refresh()
	got := runArticleFetch(t, after, cmd)

	if got.message == "" {
		t.Error("message is empty after a failed article refresh, want a Footer message")
	}
	if got.dv1.doc.path != fixtureArticle().path {
		t.Errorf("DV1 Active Document = %q after a failed refresh, want unchanged %q", got.dv1.doc.path, fixtureArticle().path)
	}
	if got.dv1.view() != beforeView {
		t.Errorf("DV1 view changed after a failed refresh, want unchanged content")
	}
	if got.dv1.fetching() {
		t.Error("dv1.fetching() = true after a failed refresh, want the indicator cleared")
	}
}

// TestRefreshFailedArticleFetchDoesNotClobberDV2sOwnRefreshMessage is a
// fresh-context review catch (C6/M4.3): refresh()'s own synchronous half
// (the file side) and the article's async re-fetch completion both write
// into the single m.message slot for the very same R press. Here DV2's own
// committed find is dropped by its ordinary synchronous re-render (setting
// "refreshed; dv2 search cleared", combineDropMessages' own precedent) at
// the same moment DV1's article re-fetch is dispatched; the article's fetch
// then fails. The "never clobber a message already set" rule
// (combineDropMessages) must hold across that sync/async boundary too, not
// just within refresh()'s own single call.
func TestRefreshFailedArticleFetchDoesNotClobberDV2sOwnRefreshMessage(t *testing.T) {
	fetcher := &fakeWikipediaFetcher{errs: map[string]error{euler.Title: errors.New("offline")}}
	dir := t.TempDir()
	filePath := writeFile(t, filepath.Join(dir, "note.md")) // "# note\n"

	m := sized(t, newModel(), 80, 24).withWikipediaFetcher(fetcher)
	m = m.navigate(dv1Pane, fixtureArticle())
	fileDoc, err := loadDocument(filePath)
	if err != nil {
		t.Fatalf("loadDocument: %v", err)
	}
	m = m.openDocumentInto(dv2Pane, fileDoc)

	m = press(t, m, keys.FocusDV2.Keys()[0])
	m = press(t, m, "/")
	m = typeText(t, m, "note")
	m = press(t, m, "enter")
	if !m.dv2.hasFindMatches() {
		t.Fatal("setup: DV2's find did not commit")
	}

	after, cmd := m.refresh()
	if !strings.Contains(after.message, "dv2 search cleared") {
		t.Fatalf("setup: message after the synchronous half of refresh = %q, want it to mention dv2's dropped search", after.message)
	}

	got := runArticleFetch(t, after, cmd)

	if !strings.Contains(got.message, "dv2 search cleared") {
		t.Errorf("message after the article's fetch failed = %q, want DV2's own refresh message preserved, not clobbered", got.message)
	}
}

// TestStaleArticleRefreshDroppedWhenTargetNavigatesAway mirrors M4.2's own
// "a completion whose target moved on ... is dropped" test
// (TestStaleWikipediaFetchDroppedWhenTargetActivatesElsewhere) for a
// Refresh-triggered re-fetch: DV1 navigates to something else while its own
// article re-fetch is still in flight, and the stale completion must not
// overwrite whatever DV1 has since moved on to.
func TestStaleArticleRefreshDroppedWhenTargetNavigatesAway(t *testing.T) {
	updated := wikipedia.Article{Title: eulerArticle.Title, URL: eulerArticle.URL, Markdown: "# Leonhard Euler\n\nRe-fetched content.\n"}
	fetcher := &fakeWikipediaFetcher{articles: map[string]wikipedia.Article{euler.Title: updated}}
	m := sized(t, newModel(), 80, 24).withWikipediaFetcher(fetcher)
	m = m.navigate(dv1Pane, fixtureArticle())

	after, cmd := m.refresh()
	if !after.dv1.fetching() {
		t.Fatal("setup: dv1 is not fetching after refresh")
	}

	// The target "moves on" before the re-fetch resolves: a real file lands
	// in DV1 via the ordinary open() path (navigate's own underlying call).
	moved := after.navigate(dv1Pane, document{path: "/tmp/other.md", src: []byte("# Other\n")})
	if moved.dv1.fetching() {
		t.Fatal("setup: dv1 still reports fetching after a real navigation, want it cleared")
	}

	got := runArticleFetch(t, moved, cmd)

	if got.dv1.doc.path != "/tmp/other.md" {
		t.Errorf("DV1 Active Document = %q after a stale refresh completion arrived, want the file it moved on to (%q)", got.dv1.doc.path, "/tmp/other.md")
	}
	if got.message != "" {
		t.Errorf("message = %q after a dropped stale refresh completion, want none", got.message)
	}
}

// TestRefreshUpdatesAnArticleInDV1BesideAFileInDV2 is M4.3's own stated
// done-when ("an article in DV1 beside a file in DV2"): a single R press
// re-fetches the article (async) and re-reads the file (synchronous, same
// as ever) in the same action, each independently, mirroring S6/C3/M3.1's
// two-viewer independence guarantee for this new mixed case.
func TestRefreshUpdatesAnArticleInDV1BesideAFileInDV2(t *testing.T) {
	updated := wikipedia.Article{Title: eulerArticle.Title, URL: eulerArticle.URL, Markdown: "# Leonhard Euler\n\nRe-fetched content.\n"}
	fetcher := &fakeWikipediaFetcher{articles: map[string]wikipedia.Article{euler.Title: updated}}
	dir := t.TempDir()
	filePath := writeFileContent(t, filepath.Join(dir, "note.md"), "# A file\n\nOriginal.\n")

	m := sized(t, newModel(), 80, 24).withWikipediaFetcher(fetcher)
	m = m.navigate(dv1Pane, fixtureArticle())
	fileDoc, err := loadDocument(filePath)
	if err != nil {
		t.Fatalf("loadDocument: %v", err)
	}
	m = m.openDocumentInto(dv2Pane, fileDoc)

	writeFileContent(t, filePath, "# A file\n\nEdited outside noggin.\n")

	after, cmd := m.refresh()
	if after.dv2.doc.path == "" || !strings.Contains(after.dv2.view(), "Edited outside noggin") {
		t.Errorf("DV2 view after refresh = %q, want the file's on-disk edit picked up immediately", after.dv2.view())
	}
	if !after.dv1.fetching() {
		t.Fatal("dv1.fetching() = false after refresh, want the article's re-fetch in flight")
	}

	got := runArticleFetch(t, after, cmd)

	if !strings.Contains(got.dv1.view(), "Re-fetched content") {
		t.Errorf("DV1 view after the re-fetch completed = %q, want the re-fetched article content", got.dv1.view())
	}
	if !strings.Contains(got.dv2.view(), "Edited outside noggin") {
		t.Errorf("DV2 view after the article's fetch completed = %q, want its own file edit still showing", got.dv2.view())
	}
}

// TestRefreshUpdatesAFileInDV1BesideAnArticleInDV2 is the mirror of
// TestRefreshUpdatesAnArticleInDV1BesideAFileInDV2: refreshViewer/
// refreshArticleFetchInto are written pane-symmetrically (dv1Pane/dv2Pane
// are just a parameter), so this pins that the same behaviour holds with
// the article on the *other* side, not only DV1.
func TestRefreshUpdatesAFileInDV1BesideAnArticleInDV2(t *testing.T) {
	updated := wikipedia.Article{Title: eulerArticle.Title, URL: eulerArticle.URL, Markdown: "# Leonhard Euler\n\nRe-fetched content.\n"}
	fetcher := &fakeWikipediaFetcher{articles: map[string]wikipedia.Article{euler.Title: updated}}
	dir := t.TempDir()
	filePath := writeFileContent(t, filepath.Join(dir, "note.md"), "# A file\n\nOriginal.\n")

	m := sized(t, newModel(), 80, 24).withWikipediaFetcher(fetcher)
	fileDoc, err := loadDocument(filePath)
	if err != nil {
		t.Fatalf("loadDocument: %v", err)
	}
	m = m.navigate(dv1Pane, fileDoc)
	m = m.navigate(dv2Pane, fixtureArticle())

	writeFileContent(t, filePath, "# A file\n\nEdited outside noggin.\n")

	after, cmd := m.refresh()
	if !strings.Contains(after.dv1.view(), "Edited outside noggin") {
		t.Errorf("DV1 view after refresh = %q, want the file's on-disk edit picked up immediately", after.dv1.view())
	}
	if !after.dv2.fetching() {
		t.Fatal("dv2.fetching() = false after refresh, want the article's re-fetch in flight")
	}

	got := runArticleFetch(t, after, cmd)

	if !strings.Contains(got.dv2.view(), "Re-fetched content") {
		t.Errorf("DV2 view after the re-fetch completed = %q, want the re-fetched article content", got.dv2.view())
	}
	if !strings.Contains(got.dv1.view(), "Edited outside noggin") {
		t.Errorf("DV1 view after the article's fetch completed = %q, want its own file edit still showing", got.dv1.view())
	}
}

func appendToFile(t *testing.T, path, more string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("open %s for append: %v", path, err)
	}
	defer f.Close()
	if _, err := f.WriteString(more); err != nil {
		t.Fatalf("append to %s: %v", path, err)
	}
}
