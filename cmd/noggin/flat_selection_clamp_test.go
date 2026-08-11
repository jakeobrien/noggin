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
	// Only a.md has a recency record, so Recent's Active Root scope lists
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

	m = m.activateInto(dv1Pane)
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
