package main

import (
	"path/filepath"
	"testing"
	"time"

	"noggin/internal/tree"
)

// TestCyclingToAShorterFlatListReClampsSelection is TRIAGE T23/D82's first
// call site: selecting a lower Tree row, then cycling to a Recent/Starred
// list with fewer rows than that index, must leave f.selected pointing at a
// valid row of the new list rather than past its end — before clampSelected
// existed, cycleMode never touched f.selected at all, so selectedFlatRow()
// silently returned ok=false and Activate acted on nothing.
func TestCyclingToAShorterFlatListReClampsSelection(t *testing.T) {
	dir := t.TempDir()
	rootDir := mkdir(t, filepath.Join(dir, "project"))
	a := recentFixtureFile(t, rootDir, "a.md")
	_ = recentFixtureFile(t, rootDir, "b.md")
	_ = recentFixtureFile(t, rootDir, "c.md")

	node, err := tree.Walk(rootDir)
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	// Only a.md has a recency record, so Recent's Active Source scope lists
	// exactly one row — shorter than the three-row Tree list below it.
	docs := map[string]docRecord{a: {LastViewed: time.Now()}}
	m := sized(t, newModel().withDocs("", docs), 80, 24)
	m.roots = m.roots.setRoots([]root{{path: rootDir}}, 0)
	m = m.openFolder(node)

	if got := len(m.files.visibleRows()); got != 3 {
		t.Fatalf("setup: Tree rows = %d, want 3", got)
	}
	// Move selection to the last Tree row (index 2).
	m = press(t, m, keys.LineDown.Keys()[0])
	m = press(t, m, keys.LineDown.Keys()[0])
	if m.files.selected != 2 {
		t.Fatalf("setup: Tree selection = %d, want 2", m.files.selected)
	}

	m = press(t, m, keys.FilesMode.Keys()[0]) // Tree -> Recent
	if len(m.files.flatRows()) != 1 {
		t.Fatalf("setup: Recent rows = %+v, want exactly [a.md]", m.files.flatRows())
	}
	if m.files.selected != 0 {
		t.Errorf("selected after cycling to a shorter list = %d, want 0 (re-clamped)", m.files.selected)
	}

	m, _ = m.activateInto(dv1Pane)
	if !m.dv1.doc.isOpen() || m.dv1.doc.path != a {
		t.Errorf("activateInto after re-clamp did not open %q into DV1: doc = %+v (silent no-op bug)", a, m.dv1.doc)
	}
}

// TestUnstarringLastVisibleStarredRowReClampsSelection is TRIAGE T23/D82's
// second call site: un-starring the selected row when it is the last one
// visible in a Starred list shrinks that list out from under f.selected,
// which must re-clamp to the new last row rather than leave the cursor
// pointing past the end of what remains.
func TestUnstarringLastVisibleStarredRowReClampsSelection(t *testing.T) {
	dir := t.TempDir()
	rootDir := mkdir(t, filepath.Join(dir, "project"))
	a := starredFixtureFile(t, rootDir, "a.md")
	b := starredFixtureFile(t, rootDir, "b.md")
	storePath := filepath.Join(dir, "documents")

	node, err := tree.Walk(rootDir)
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	docs := map[string]docRecord{
		a: {Starred: true},
		b: {Starred: true},
	}
	m := sized(t, newModel().withDocs(storePath, docs), 80, 24)
	m.roots = m.roots.setRoots([]root{{path: rootDir}}, 0)
	m = m.openFolder(node)

	m = press(t, m, keys.FilesMode.Keys()[0])
	m = press(t, m, keys.FilesMode.Keys()[0]) // Tree -> Recent -> Starred
	if rows := m.files.flatRows(); len(rows) != 2 || rows[1].path != b {
		t.Fatalf("setup: Starred rows = %+v, want [a.md, b.md]", rows)
	}
	m = press(t, m, keys.LineDown.Keys()[0]) // select b.md, the last row
	if m.files.selected != 1 {
		t.Fatalf("setup: Starred selection = %d, want 1", m.files.selected)
	}

	m = m.toggleDocStar() // un-star b.md, dropping its row immediately
	if got := m.docRecord(b); got.Starred {
		t.Fatalf("docRecord(%s) after toggleDocStar = %+v, want Starred false", b, got)
	}
	if len(m.files.flatRows()) != 1 {
		t.Fatalf("Starred rows after un-starring = %+v, want exactly [a.md]", m.files.flatRows())
	}
	if m.files.selected != 0 {
		t.Errorf("selected after un-starring the last visible row = %d, want 0 (re-clamped)", m.files.selected)
	}

	// A silent no-op (the pre-fix bug) would leave a.md's own star
	// untouched (still true); a valid, re-clamped selection lets this
	// second Star actually flip it to false.
	m = m.toggleDocStar()
	if got := m.docRecord(a); got.Starred {
		t.Errorf("docRecord(%s) after second toggleDocStar = %+v, want Starred false (action taken, not a silent no-op)", a, got)
	}
}

// TestCyclingScopeToASmallerFlatListReClampsSelection is TRIAGE T37/D88's
// first gap: M4.7's own acceptance wording (D82) names cycleScope alongside
// cycleMode as a site that can shrink the row set out from under
// f.selected (clampSelected's own doc comment lists both), but only
// cycleMode had a regression test before this. Recent's own two scope
// values give the shrink directly: All Roots (this Root's one recent
// document plus another Root's three) cycling back down to Active Source
// (this Root's alone) — the same wrap TestFilesModeScopeKeysCycleFromFiles-
// NavFocus's own two-value cycle takes, just with real rows behind it here.
func TestCyclingScopeToASmallerFlatListReClampsSelection(t *testing.T) {
	dir := t.TempDir()
	rootADir := mkdir(t, filepath.Join(dir, "root-a"))
	rootBDir := mkdir(t, filepath.Join(dir, "root-b"))
	a := recentFixtureFile(t, rootADir, "a.md")
	b1 := recentFixtureFile(t, rootBDir, "b1.md")
	b2 := recentFixtureFile(t, rootBDir, "b2.md")
	b3 := recentFixtureFile(t, rootBDir, "b3.md")

	nodeA, err := tree.Walk(rootADir)
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}

	now := time.Now()
	docs := map[string]docRecord{
		a:  {LastViewed: now},
		b1: {LastViewed: now.Add(-1 * time.Minute)},
		b2: {LastViewed: now.Add(-2 * time.Minute)},
		b3: {LastViewed: now.Add(-3 * time.Minute)},
	}
	m := sized(t, newModel().withDocs("", docs), 80, 24)
	m.roots = m.roots.setRoots([]root{{path: rootADir}, {path: rootBDir}}, 0)
	m = m.openFolder(nodeA)

	m = press(t, m, keys.FilesMode.Keys()[0]) // Tree -> Recent, Active Source by default
	if got := len(m.files.flatRows()); got != 1 {
		t.Fatalf("setup: Recent/Active Source rows = %d, want 1", got)
	}

	m = press(t, m, keys.FilesScope.Keys()[0]) // Active Source -> All Roots
	if got := len(m.files.flatRows()); got != 4 {
		t.Fatalf("setup: Recent/All Roots rows = %d, want 4", got)
	}
	m = press(t, m, keys.LineDown.Keys()[0])
	m = press(t, m, keys.LineDown.Keys()[0])
	m = press(t, m, keys.LineDown.Keys()[0])
	if m.files.selected != 3 {
		t.Fatalf("setup: All Roots selection = %d, want 3 (last row)", m.files.selected)
	}

	m = press(t, m, keys.FilesScope.Keys()[0]) // All Roots -> Active Source (wraps: only two values)
	if got := len(m.files.flatRows()); got != 1 {
		t.Fatalf("setup: back on Active Source rows = %d, want 1", got)
	}
	if m.files.selected != 0 {
		t.Errorf("selected after cycling scope to a shorter list = %d, want 0 (re-clamped)", m.files.selected)
	}

	m, _ = m.activateInto(dv1Pane)
	if !m.dv1.doc.isOpen() || m.dv1.doc.path != a {
		t.Errorf("activateInto after re-clamp did not open %q into DV1: doc = %+v (silent no-op bug)", a, m.dv1.doc)
	}
}

// TestSettingRootRestoringAFlatModeReClampsSelection is TRIAGE T37/D88's
// second gap: setRoot's own doc comment names its flat-mode memory-restore
// branch as a shrink site too ("clampSelected's final pass is what makes
// this safe"), since the selectedPath it restores resolves against the
// *current* Tree (indexOfPath), meaningless for whatever Recent/Starred
// list the restored mode actually shows — but nothing exercised that
// restore branch shrinking directly. Root A is left on a three-row Starred
// list selected at its last row (a3.md, which also happens to sit at Tree
// row 2 here — alphabetical order coincides with Starred/Active Source's own
// name sort); a3.md is then un-starred while Root B is the pane's current
// root (Root A's Starred/Active Source list is now two rows, but a3.md's
// file itself — and so its Tree row — is untouched), before setRoot moves
// the pane back onto Root A.
//
// This drives setRoot directly (filesNav's own level, not a full model
// Root switch — browse_memory_test.go's style): a real Root switch also
// re-points activeRootPath (model.refreshFilesRecentSource, called right
// after setRoot by openFolder) in the same step, which would fold this
// test's shrink together with that separate hand-off. Keeping
// activeRootPath fixed at Root A throughout isolates the one thing this
// test is actually about — setRoot's own restored-index-vs-flat-list-size
// mismatch — from that unrelated ordering (logged instead as TRIAGE, not
// this task's concern).
func TestSettingRootRestoringAFlatModeReClampsSelection(t *testing.T) {
	dir := t.TempDir()
	rootADir := mkdir(t, filepath.Join(dir, "root-a"))
	rootBDir := mkdir(t, filepath.Join(dir, "root-b"))
	a1 := starredFixtureFile(t, rootADir, "a1.md")
	a2 := starredFixtureFile(t, rootADir, "a2.md")
	a3 := starredFixtureFile(t, rootADir, "a3.md")

	nodeA, err := tree.Walk(rootADir)
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	nodeB, err := tree.Walk(rootBDir)
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}

	docs := map[string]docRecord{
		a1: {Starred: true},
		a2: {Starred: true},
		a3: {Starred: true},
	}
	roots := []root{{path: rootADir}, {path: rootBDir}}

	f := newFilesNav().withRecentSource(docs, roots, rootADir)
	f = f.setRoot(nodeA)
	f = f.cycleMode()
	f = f.cycleMode() // Tree -> Recent -> Starred
	if rows := f.flatRows(); len(rows) != 3 || rows[2].path != a3 {
		t.Fatalf("setup: Starred rows = %+v, want [a1.md, a2.md, a3.md]", rows)
	}
	f = f.move(1)
	f = f.move(1)
	if f.selected != 2 {
		t.Fatalf("setup: Starred selection = %d, want 2 (last row, a3.md)", f.selected)
	}

	// Leave Root A: setRoot's remember() snapshots mode=starredBrowse and
	// selectedPath from the Tree's own row 2 (a3.md).
	f = f.setRoot(nodeB)

	// a3.md is un-starred while Root A is not the pane's current root —
	// Root A's Starred/Active Source list is now two rows, not the three
	// memory still remembers.
	docs = map[string]docRecord{a1: {Starred: true}, a2: {Starred: true}, a3: {Starred: false}}
	f = f.withRecentSource(docs, roots, rootADir)

	// Back to Root A: the restore resolves the remembered selectedPath
	// (a3.md) against the current Tree — still row 2, the file itself was
	// never removed — giving a Tree-shaped f.selected of 2 before
	// clampSelected re-resolves it against the now-two-row Starred list.
	nodeA2, err := tree.Walk(rootADir)
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	f = f.setRoot(nodeA2)

	if f.mode != starredBrowse {
		t.Fatalf("mode after restoring Root A = %v, want starredBrowse", f.mode)
	}
	rows := f.flatRows()
	if len(rows) != 2 {
		t.Fatalf("Starred rows after restore = %+v, want 2 (a3.md un-starred)", rows)
	}
	if f.selected != 1 {
		t.Errorf("selected after restoring a shrunk flat list = %d, want 1 (re-clamped)", f.selected)
	}

	if row, ok := f.selectedFlatRow(); !ok || row.path != a2 {
		t.Errorf("selectedFlatRow after re-clamp = (%+v, %v), want (%q, true) (silent no-op bug otherwise)", row, ok, a2)
	}
}
