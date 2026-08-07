package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"noggin/internal/tree"
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
	if !strings.Contains(v.view(), emptyViewerText) {
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

	node, err := tree.Walk(root)
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
	if total := m.viewer.lineCount(); total <= m.viewer.height() {
		t.Fatalf("doc.md renders %d lines into a %d-line viewport; nothing to preserve a scroll over", total, m.viewer.height())
	}
	m.viewer = m.viewer.withOffset(10)
	wantOffset := m.viewer.offset()

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
	if got := m.viewer.offset(); got != wantOffset {
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

	node, err := tree.Walk(root)
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	d, err := loadDocument(docPath)
	if err != nil {
		t.Fatalf("loadDocument: %v", err)
	}
	m := sized(t, newModel().openFolder(node), 80, 24)
	m = m.openDocument(d)
	if !m.viewer.doc.isOpen() {
		t.Fatalf("setup: document not open")
	}

	if err := os.Remove(docPath); err != nil {
		t.Fatalf("remove doc.md: %v", err)
	}
	m = press(t, m, keys.Refresh.Keys()[0])

	if m.viewer.doc.isOpen() {
		t.Errorf("viewer still reports a document open after its file was deleted")
	}
	if !strings.Contains(m.viewer.view(), emptyViewerText) {
		t.Errorf("viewer view = %q, want the empty state", m.viewer.view())
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
	dir := newGitRepo(t) // tracked.txt (M), sub/untracked.txt (A), clean.txt unmarked

	node, err := tree.Walk(dir)
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	m := sized(t, newModel().openFolder(node), 80, 24)
	if strings.Contains(lineContaining(strings.Split(m.files.view(), "\n"), "clean.txt"), "M clean.txt") {
		t.Fatalf("setup: clean.txt already marked before it was dirtied")
	}

	writeFileContent(t, filepath.Join(dir, "clean.txt"), "now dirty\n")
	m = press(t, m, keys.Refresh.Keys()[0])

	view := m.files.view()
	if !strings.Contains(lineContaining(strings.Split(view, "\n"), "clean.txt"), "M clean.txt") {
		t.Errorf("Files Nav view after refresh = %q, want clean.txt marked M", view)
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
	node, err := tree.Walk(root)
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
	if m.viewer.findEditing() {
		t.Fatal("test setup: enter did not commit the find")
	}
	if !m.viewer.hasFindMatches() {
		t.Fatal("test setup: committed find has no matches to lose")
	}

	m = press(t, m, keys.Refresh.Keys()[0])
	if m.viewer.hasFindMatches() {
		t.Fatal("test setup: refresh did not actually invalidate the committed find")
	}
	if !strings.Contains(m.message, "search") {
		t.Errorf("message = %q, want it to explain the cleared search", m.message)
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
	if m.viewer.doc.isOpen() {
		t.Errorf("viewer document after refreshing with none = %v, want none still", m.viewer.doc)
	}
	if m.message != "" {
		t.Errorf("refresh with nothing to refresh produced a message %q, want silence", m.message)
	}
}

// TestRefreshWorksWithFocusOnRootsPane is TRIAGE T4: keys.Refresh is matched
// in model.Update's global switch, unconditional on m.focus (unlike
// FocusRoots/FocusFiles' every-pane table in focus_test.go, there is no
// existing test that drives Refresh specifically with focus on rootsPane —
// only filesPane and viewerPane were exercised. This pins S2's "reachable
// from any pane" for the one pane the existing suite skipped.
func TestRefreshWorksWithFocusOnRootsPane(t *testing.T) {
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	writeFile(t, filepath.Join(root, "keep.md"))

	node, err := tree.Walk(root)
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
// model.refresh only ever assigns to m.files/m.viewer — but nothing
// regression-proofs it. With two roots configured, deletes the inactive
// root's directory (which would flip its Roots Nav broken mark if it were
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
