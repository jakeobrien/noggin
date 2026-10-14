package main

import (
	"path/filepath"
	"testing"

	"noggin/internal/tree"
)

// TestOpenFolderRestoresFlatModeSelectionAgainstTheCorrectRoot is TRIAGE
// T38/M4.15's own regression: switching the Active Root back onto a Root
// previously left in a flat browse mode (Recent/Starred, Active Source scope)
// must restore M4.6's remembered row, not reset to row 0.
//
// The bug lived in filesNav.setRoot's own clampSelected() call: it runs
// before model.openFolder's later refreshFilesRecentSource re-points
// f.activeRootPath at the Root setRoot itself just switched to, so the
// restore's re-clamp (clampSelected -> rowCount -> flatRows -> recentRows/
// starredRows) reads Active Source scope's bucket keyed by the *outgoing*
// Root's still-stale activeRootPath instead of the incoming one — an empty
// or wrong-sized bucket, collapsing the restored index to 0 even though the
// incoming Root's own flat list is unchanged and the memory-restored index
// would otherwise be valid against it.
//
// This drives model.openFolder itself (not filesNav.setRoot in isolation,
// which TestSettingRootRestoringAFlatModeReClampsSelection already covers
// for a different shrink and explicitly defers this ordering hand-off —
// see that test's own doc comment) — two Roots, each with their own
// starred documents, so Root B's Active-Root-scope bucket is a different,
// smaller size than Root A's, exactly the mismatch a stale
// activeRootPath produces.
func TestOpenFolderRestoresFlatModeSelectionAgainstTheCorrectRoot(t *testing.T) {
	dir := t.TempDir()
	rootADir := mkdir(t, filepath.Join(dir, "root-a"))
	rootBDir := mkdir(t, filepath.Join(dir, "root-b"))
	a1 := starredFixtureFile(t, rootADir, "a1.md")
	a2 := starredFixtureFile(t, rootADir, "a2.md")
	a3 := starredFixtureFile(t, rootADir, "a3.md")
	_ = starredFixtureFile(t, rootBDir, "b1.md")

	nodeA, err := tree.Walk(rootADir, tree.FilterMarkdown)
	if err != nil {
		t.Fatalf("tree.Walk(rootA): %v", err)
	}
	nodeB, err := tree.Walk(rootBDir, tree.FilterMarkdown)
	if err != nil {
		t.Fatalf("tree.Walk(rootB): %v", err)
	}

	docs := map[string]docRecord{
		a1: {Starred: true},
		a2: {Starred: true},
		a3: {Starred: true},
		// Root B's own bucket is smaller (one row) than Root A's (three) —
		// the size mismatch a stale activeRootPath would read instead.
	}
	m := sized(t, newModel().withDocs("", docs), 80, 24)
	m.roots = m.roots.setRoots([]root{{path: rootADir}, {path: rootBDir}}, 0)
	m = m.openFolder(nodeA)

	m = press(t, m, keys.FilesMode.Keys()[0])
	m = press(t, m, keys.FilesMode.Keys()[0]) // Tree -> Recent -> Starred
	if rows := m.files.flatRows(); len(rows) != 3 || rows[2].path != a3 {
		t.Fatalf("setup: Root A Starred rows = %+v, want [a1.md, a2.md, a3.md]", rows)
	}
	m = press(t, m, keys.LineDown.Keys()[0])
	m = press(t, m, keys.LineDown.Keys()[0])
	if m.files.selected != 2 {
		t.Fatalf("setup: Root A Starred selection = %d, want 2 (a3.md)", m.files.selected)
	}

	// Switch to Root B (openFolder again, activating it first the same way
	// model.activateRoot does) — M4.6 memory now holds Root A's (Starred,
	// Active Source) pair with selectedPath = a3.md.
	m.roots = m.roots.activate(1)
	m = m.openFolder(nodeB)
	if m.files.root.Path != rootBDir {
		t.Fatalf("setup: active root after switch = %q, want %q", m.files.root.Path, rootBDir)
	}

	// Back to Root A: the memory-restored row (a3.md, index 2) must survive
	// against Root A's own three-row Starred/Active Source list, not collapse
	// to 0 against Root B's single-row bucket.
	m.roots = m.roots.activate(0)
	m = m.openFolder(nodeA)
	if m.files.mode != starredBrowse {
		t.Fatalf("mode after restoring Root A = %v, want starredBrowse", m.files.mode)
	}
	if rows := m.files.flatRows(); len(rows) != 3 {
		t.Fatalf("Root A Starred rows after restore = %+v, want 3", rows)
	}
	if m.files.selected != 2 {
		t.Errorf("selected after restoring Root A = %d, want 2 (memory-restored a3.md, not reset to 0)", m.files.selected)
	}
	if row, ok := m.files.selectedFlatRow(); !ok || row.path != a3 {
		t.Errorf("selectedFlatRow after restore = (%+v, %v), want (%q, true)", row, ok, a3)
	}
}
