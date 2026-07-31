package main

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"noggin/internal/tree"
)

// TestRootsNavOpensOnActiveRootAndPreservesOrder checks that setRoots keeps
// S1's order (ad-hoc first, then starred) untouched and starts the cursor on
// the active entry.
func TestRootsNavOpensOnActiveRootAndPreservesOrder(t *testing.T) {
	a := mkdir(t, filepath.Join(t.TempDir(), "a"))
	b := mkdir(t, filepath.Join(t.TempDir(), "b"))
	r := newRootsNav("").setRoots([]root{{path: a}, {path: b, starred: true}}, 1)

	if got := []string{r.rows[0].root.path, r.rows[1].root.path}; !reflect.DeepEqual(got, []string{a, b}) {
		t.Errorf("row order = %v, want %v (unreordered)", got, []string{a, b})
	}
	if r.active != 1 || r.selected != 1 {
		t.Errorf("active/selected = %d/%d, want 1/1 (cursor starts on the active root)", r.active, r.selected)
	}
}

// TestRootsNavDistinguishesSameNamedRootsByParent is M3.3's acceptance
// criterion: two roots sharing a base name must read apart by their parent
// path.
func TestRootsNavDistinguishesSameNamedRootsByParent(t *testing.T) {
	base := t.TempDir()
	a := mkdir(t, filepath.Join(base, "work", "admin", "project"))
	b := mkdir(t, filepath.Join(base, "other", "project"))

	r := newRootsNav("").setRoots([]root{{path: a}, {path: b}}, -1).setSize(60, 4)
	lines := strings.Split(r.view(), "\n")

	if !strings.Contains(lines[0], "project") || !strings.Contains(lines[1], "project") {
		t.Fatalf("rows = %q, want both to name project", lines[:2])
	}
	if lines[0] == lines[1] {
		t.Error("two same-named roots rendered identically, want their parent paths to distinguish them")
	}
	if !strings.Contains(lines[0], "admin") {
		t.Errorf("row 0 = %q, want it to show its parent path (admin)", lines[0])
	}
	if !strings.Contains(lines[1], "other") {
		t.Errorf("row 1 = %q, want it to show its parent path (other)", lines[1])
	}
}

// TestRootsNavMarksMissingPathBroken checks S3's "roots whose path is missing
// are marked broken".
func TestRootsNavMarksMissingPathBroken(t *testing.T) {
	ok := mkdir(t, filepath.Join(t.TempDir(), "here"))
	gone := filepath.Join(t.TempDir(), "gone")

	r := newRootsNav("").setRoots([]root{{path: ok}, {path: gone}}, 0).setSize(40, 4)
	if r.rows[0].broken {
		t.Error("an existing path was marked broken")
	}
	if !r.rows[1].broken {
		t.Error("a nonexistent path was not marked broken")
	}

	lines := strings.Split(r.view(), "\n")
	if strings.Contains(lines[0], "missing") {
		t.Errorf("row 0 (exists) = %q, want no broken marker", lines[0])
	}
	if !strings.Contains(lines[1], "missing") {
		t.Errorf("row 1 (gone) = %q, want a broken marker", lines[1])
	}
}

// TestRootsNavHighlightsActiveRoot checks S3's "the active root is
// highlighted" and that the bold marker moves with the active index, not the
// selection cursor.
func TestRootsNavHighlightsActiveRoot(t *testing.T) {
	a := mkdir(t, filepath.Join(t.TempDir(), "a"))
	b := mkdir(t, filepath.Join(t.TempDir(), "b"))

	r := newRootsNav("").setRoots([]root{{path: a}, {path: b}}, 1).setSize(40, 4)
	r = r.move(-1) // cursor back to row 0, active stays row 1

	lines := strings.Split(r.view(), "\n")
	if strings.Contains(lines[0], "\x1b[1m") {
		t.Errorf("inactive row 0 = %q, want no bold SGR", lines[0])
	}
	if !strings.Contains(lines[1], "\x1b[1m") {
		t.Errorf("active row 1 = %q, want bold SGR", lines[1])
	}
	// The selection cursor (row 0 after the move) is reverse-video regardless
	// of which row is active.
	if !strings.Contains(lines[0], "\x1b[7m") {
		t.Errorf("selected row 0 = %q, want reverse-video SGR", lines[0])
	}
}

// TestRootsNavStarredMarked checks S3's "starred roots are marked" and that
// toggleStarred flips only the targeted row.
func TestRootsNavStarredMarked(t *testing.T) {
	a := mkdir(t, filepath.Join(t.TempDir(), "a"))
	b := mkdir(t, filepath.Join(t.TempDir(), "b"))

	r := newRootsNav("").setRoots([]root{{path: a}, {path: b, starred: true}}, -1).setSize(40, 4)
	lines := strings.Split(r.view(), "\n")
	if strings.Contains(lines[0], "★") {
		t.Errorf("unstarred row 0 = %q, want no star mark", lines[0])
	}
	if !strings.Contains(lines[1], "★") {
		t.Errorf("starred row 1 = %q, want a star mark", lines[1])
	}

	r = r.toggleStarred(0)
	lines = strings.Split(r.view(), "\n")
	if !strings.Contains(lines[0], "★") {
		t.Errorf("row 0 after toggleStarred = %q, want a star mark", lines[0])
	}
	if !strings.Contains(lines[1], "★") {
		t.Error("toggling row 0 unstarred row 1 too")
	}
}

// TestRootsNavMoveClampsAtEnds checks selection movement, mirroring the Files
// Nav's own clamp behaviour.
func TestRootsNavMoveClampsAtEnds(t *testing.T) {
	a := mkdir(t, filepath.Join(t.TempDir(), "a"))
	b := mkdir(t, filepath.Join(t.TempDir(), "b"))
	r := newRootsNav("").setRoots([]root{{path: a}, {path: b}}, 0)

	r = r.move(-1)
	if r.selected != 0 {
		t.Errorf("moving up past the top = %d, want it to clamp at 0", r.selected)
	}
	r = r.move(1)
	if r.selected != 1 {
		t.Errorf("selected after move(1) = %d, want 1", r.selected)
	}
	r = r.move(1)
	if r.selected != 1 {
		t.Errorf("moving down past the end = %d, want it to clamp at 1", r.selected)
	}
}

// TestRootsNavEmptyState checks the no-roots resting state (a bare launch
// with no starred roots yet).
func TestRootsNavEmptyState(t *testing.T) {
	r := newRootsNav("").setSize(20, 4)
	lines := strings.Split(r.view(), "\n")
	if len(lines) != 4 {
		t.Fatalf("empty-state view has %d lines, want 4", len(lines))
	}
	if !strings.Contains(r.view(), "No roots") {
		t.Errorf("empty-state view = %q, want the no-roots message", r.view())
	}
}

// --- model wiring: activating and starring through the pane -----------------

// TestActivateRootRerootsFilesNavAndMovesFocus checks activateRoot's happy
// path: the Files Nav re-roots to the newly active root and focus follows,
// same shape as activate() for files (M2.3).
func TestActivateRootRerootsFilesNavAndMovesFocus(t *testing.T) {
	m := sized(t, newModel(), 80, 24)
	m.roots = m.roots.setRoots([]root{{path: "testdata/tree"}}, -1)

	m = m.activateRoot()
	if m.focus != filesPane {
		t.Fatalf("focus after activateRoot = %v, want filesPane", m.focus)
	}
	if m.roots.active != 0 {
		t.Errorf("active root index = %d, want 0", m.roots.active)
	}
	if !strings.Contains(m.files.view(), "README.md") {
		t.Errorf("files view = %q, want testdata/tree's entries", m.files.view())
	}
}

// TestActivateRootRefusesBroken checks S3/M3.3's acceptance criterion:
// activating a root whose path is missing is refused with a footer message,
// and neither the active root nor the Files Nav change.
func TestActivateRootRefusesBroken(t *testing.T) {
	ok := mkdir(t, filepath.Join(t.TempDir(), "here"))
	gone := filepath.Join(t.TempDir(), "gone")

	m := sized(t, newModel(), 80, 24)
	m.roots = m.roots.setRoots([]root{{path: ok}, {path: gone}}, 0)
	m.roots = m.roots.move(1) // cursor onto the broken entry; active stays 0

	m = m.activateRoot()
	if m.roots.active != 0 {
		t.Errorf("active root = %d after refusing a broken activation, want unchanged 0", m.roots.active)
	}
	if m.message == "" {
		t.Error("message is empty after refusing to activate a broken root, want a footer message")
	}
	if m.files.root != nil {
		t.Error("Files Nav was re-rooted despite the activation being refused")
	}
	if !strings.Contains(m.frame(), gone) {
		t.Errorf("frame does not surface the refusal message naming %s", gone)
	}
}

// TestToggleStarPersistsAndUpdatesDisplay is M3.3's other acceptance
// criterion: the star mark toggles and persists (through M3.2's store).
func TestToggleStarPersistsAndUpdatesDisplay(t *testing.T) {
	dir := t.TempDir()
	target := mkdir(t, filepath.Join(dir, "project"))
	storeFile := filepath.Join(dir, "config", "roots")

	m := sized(t, newModel().withStorePath(storeFile), 80, 24)
	m.roots = m.roots.setRoots([]root{{path: target}}, -1)

	m = m.toggleStar()
	row, ok := m.roots.selectedRow()
	if !ok || !row.root.starred {
		t.Fatalf("selected row after toggleStar: ok=%v starred=%v, want true/true", ok, row.root.starred)
	}
	if got := loadStore(storeFile); !reflect.DeepEqual(got, []string{target}) {
		t.Errorf("loadStore after toggleStar = %v, want [%s]", got, target)
	}

	// Toggling again unstars and removes it from the persisted file — the
	// restart-survival shape from TestToggleStorePersistsAcrossReload, driven
	// through the model action this time.
	m = m.toggleStar()
	row, _ = m.roots.selectedRow()
	if row.root.starred {
		t.Error("row still starred after toggling twice, want false")
	}
	if got := loadStore(storeFile); len(got) != 0 {
		t.Errorf("loadStore after unstar = %v, want empty", got)
	}
}

// TestToggleStarWithNoStorePathDeclinesAndMessages checks the rare
// unresolvable-config-dir case (storePath ""): the toggle cannot persist, so
// it must not appear to flip either (T18) — the display stays put and a
// footer message explains why, the same as the explicit-write-error branch.
func TestToggleStarWithNoStorePathDeclinesAndMessages(t *testing.T) {
	target := mkdir(t, filepath.Join(t.TempDir(), "project"))
	m := sized(t, newModel(), 80, 24) // storePath defaults to ""
	m.roots = m.roots.setRoots([]root{{path: target}}, -1)

	m = m.toggleStar()
	row, ok := m.roots.selectedRow()
	if !ok || row.root.starred {
		t.Errorf("selected row after toggleStar with no store path: ok=%v starred=%v, want true/false (declined)", ok, row.root.starred)
	}
	if m.message == "" {
		t.Error("want a footer message explaining the decline, got none")
	}
}

// TestActivateRootLeavesOpenDocumentUntouched is M3.4's acceptance criterion:
// with a document open from root A, activating root B re-roots the Files Nav
// (collapsed at top level, selection reset) while the viewer keeps showing A's
// document at its previous scroll position, unmoved. This mostly falls out of
// M3.3's activateRoot reusing openFolder (which only ever touches
// m.files/m.focus) and filesNav.setRoot resetting expansion/selection — this
// test is what proves it rather than leaving it an assumption.
func TestActivateRootLeavesOpenDocumentUntouched(t *testing.T) {
	rootA, err := tree.Walk("testdata/tree")
	if err != nil {
		t.Fatalf("walk testdata/tree: %v", err)
	}
	rootBPath := mkdir(t, filepath.Join(t.TempDir(), "root-b"))
	writeFile(t, filepath.Join(rootBPath, "top.md"))
	sub := mkdir(t, filepath.Join(rootBPath, "sub"))
	writeFile(t, filepath.Join(sub, "nested.md"))

	m := openedFixture(t, 80, 24)
	if total := m.viewer.lineCount(); total <= m.viewer.height() {
		t.Fatalf("fixture renders %d lines into a %d-line viewport; nothing to preserve a scroll position over", total, m.viewer.height())
	}
	m.viewer.setOffset(3)
	wantOffset, wantView := m.viewer.offset(), m.viewer.view()

	// Root A open in the Files Nav, one folder expanded and selection off the
	// top, so re-rooting to B has something to reset.
	m = m.openFolder(rootA)
	var docsPath string
	for _, child := range rootA.Children {
		if child.IsDir && child.Name == "docs" {
			docsPath = child.Path
		}
	}
	if docsPath == "" {
		t.Fatalf("testdata/tree has no docs subfolder to expand")
	}
	m.files = m.files.withExpanded(docsPath, true)
	m.files = m.files.move(1)
	if len(m.files.expanded) == 0 || m.files.selected == 0 {
		t.Fatalf("setup did not leave root A expanded/selected off the top")
	}

	m.roots = m.roots.setRoots([]root{{path: rootA.Path}, {path: rootBPath}}, 0)
	m.roots = m.roots.move(1) // cursor onto root B; activating it makes it active

	m = m.activateRoot()

	if m.roots.active != 1 {
		t.Fatalf("active root after activateRoot = %d, want 1 (root B)", m.roots.active)
	}
	if len(m.files.expanded) != 0 {
		t.Errorf("Files Nav expansion after re-rooting = %v, want empty (collapsed at top level)", m.files.expanded)
	}
	if m.files.selected != 0 {
		t.Errorf("Files Nav selection after re-rooting = %d, want 0", m.files.selected)
	}
	view := m.files.view()
	if !strings.Contains(view, "top.md") || !strings.Contains(view, "sub") {
		t.Errorf("Files Nav view after re-rooting = %q, want root B's top-level entries", view)
	}
	if strings.Contains(view, "nested.md") {
		t.Errorf("Files Nav view after re-rooting = %q, want sub collapsed (nested.md not shown)", view)
	}

	if got := m.viewer.offset(); got != wantOffset {
		t.Errorf("viewer offset after activating a different root = %d, want unchanged %d", got, wantOffset)
	}
	if got := m.viewer.view(); got != wantView {
		t.Error("viewer content after activating a different root changed, want the open document unchanged")
	}
	if m.viewer.label() == "" {
		t.Error("viewer lost its open document after activating a different root")
	}
}
