package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"noggin/internal/tree"
)

// recentFixtureFile creates a real file at dir/name so pathMissing (S6's
// "a record whose file no longer exists is silently omitted") reads it as
// present \u2014 recentRows must never be exercised against fixture paths that
// merely look plausible, unlike fixtureTree's Tree-mode fixtures, since it
// stats the disk directly.
func recentFixtureFile(t *testing.T, dir, name string) string {
	t.Helper()
	return writeFileContent(t, filepath.Join(dir, name), "content\n")
}

// TestRecentRowsOrdersByTimestamp is M4.2's own acceptance wording, first
// clause: within one Root, rows come back newest-LastViewed-first.
func TestRecentRowsOrdersByTimestamp(t *testing.T) {
	dir := t.TempDir()
	rootDir := mkdir(t, filepath.Join(dir, "root"))
	oldest := recentFixtureFile(t, rootDir, "oldest.md")
	middle := recentFixtureFile(t, rootDir, "middle.md")
	newest := recentFixtureFile(t, rootDir, "newest.md")

	now := time.Now()
	docs := map[string]docRecord{
		oldest: {LastViewed: now.Add(-2 * time.Hour)},
		middle: {LastViewed: now.Add(-1 * time.Hour)},
		newest: {LastViewed: now},
	}
	roots := []root{{path: rootDir}}

	rows := recentRows(docs, roots, rootDir, activeRootBrowseScope)
	want := []string{newest, middle, oldest}
	if len(rows) != len(want) {
		t.Fatalf("recentRows returned %d rows, want %d", len(rows), len(want))
	}
	for i, w := range want {
		if rows[i].path != w {
			t.Errorf("row %d = %q, want %q (newest-first order)", i, rows[i].path, w)
		}
	}
}

// TestRecentRowsCapsAtOneHundredPerRoot is M4.2's per-Root cap: with more
// than recentPerRootCap documents under one Root, only the recentPerRootCap
// most recently viewed survive, oldest dropped first.
func TestRecentRowsCapsAtOneHundredPerRoot(t *testing.T) {
	dir := t.TempDir()
	rootDir := mkdir(t, filepath.Join(dir, "root"))

	const total = recentPerRootCap + 10
	docs := make(map[string]docRecord, total)
	now := time.Now()
	var newestPaths []string
	for i := 0; i < total; i++ {
		p := recentFixtureFile(t, rootDir, fmt.Sprintf("doc-%03d.md", i))
		// i=0 is the oldest (viewed longest ago), i=total-1 the newest.
		docs[p] = docRecord{LastViewed: now.Add(time.Duration(i) * time.Minute)}
		if i >= total-recentPerRootCap {
			newestPaths = append(newestPaths, p)
		}
	}
	roots := []root{{path: rootDir}}

	rows := recentRows(docs, roots, rootDir, activeRootBrowseScope)
	if len(rows) != recentPerRootCap {
		t.Fatalf("recentRows returned %d rows, want the cap %d", len(rows), recentPerRootCap)
	}
	got := map[string]bool{}
	for _, r := range rows {
		got[r.path] = true
	}
	for _, p := range newestPaths {
		if !got[p] {
			t.Errorf("newest document %q missing from the capped list; want only the oldest dropped", p)
		}
	}
}

// TestRecentRowsAllRootsMergesByTimestampAcrossRootsAndOrphans is M4.2's
// merged-ordering clause: All Roots interleaves two Roots' own per-Root
// lists, plus an orphan document under no session Root, into one
// timestamp-ordered list \u2014 also the case that exercises Recent's two
// right-hand label forms (S6), since Recent's scope set (S5) has no
// dedicated Orphans value the way Starred's does.
func TestRecentRowsAllRootsMergesByTimestampAcrossRootsAndOrphans(t *testing.T) {
	dir := t.TempDir()
	rootA := mkdir(t, filepath.Join(dir, "root-a"))
	rootB := mkdir(t, filepath.Join(dir, "root-b"))
	orphanDir := mkdir(t, filepath.Join(dir, "elsewhere"))

	aDoc := recentFixtureFile(t, rootA, "a.md")
	bDoc := recentFixtureFile(t, rootB, "b.md")
	orphanDoc := recentFixtureFile(t, orphanDir, "orphan.md")

	now := time.Now()
	docs := map[string]docRecord{
		aDoc:      {LastViewed: now.Add(-2 * time.Hour)}, // oldest
		orphanDoc: {LastViewed: now.Add(-1 * time.Hour)}, // middle
		bDoc:      {LastViewed: now},                     // newest
	}
	roots := []root{{path: rootA}, {path: rootB}}

	rows := recentRows(docs, roots, rootA, allRootsBrowseScope)
	wantOrder := []string{bDoc, orphanDoc, aDoc}
	if len(rows) != len(wantOrder) {
		t.Fatalf("recentRows returned %d rows, want %d", len(rows), len(wantOrder))
	}
	for i, w := range wantOrder {
		if rows[i].path != w {
			t.Errorf("row %d = %q, want %q", i, rows[i].path, w)
		}
	}

	// Right-hand label forms: a Root-owned row carries the Root's own base
	// name, never the orphan/left-truncated form, and vice versa.
	for _, r := range rows {
		switch r.path {
		case aDoc:
			if r.orphan || r.rootLabel != "root-a" {
				t.Errorf("a.md row = %+v, want rootLabel \"root-a\", orphan false", r)
			}
		case bDoc:
			if r.orphan || r.rootLabel != "root-b" {
				t.Errorf("b.md row = %+v, want rootLabel \"root-b\", orphan false", r)
			}
		case orphanDoc:
			if !r.orphan || r.orphanDir != orphanDir {
				t.Errorf("orphan.md row = %+v, want orphan true, orphanDir %q", r, orphanDir)
			}
		}
	}
}

// TestRecentRowsActiveRootScopeOnlyShowsThatRoot checks Active Root scope in
// isolation, against the same two-Root fixture shape as the All Roots test:
// only rootA's own document appears, never rootB's or the orphan's.
func TestRecentRowsActiveRootScopeOnlyShowsThatRoot(t *testing.T) {
	dir := t.TempDir()
	rootA := mkdir(t, filepath.Join(dir, "root-a"))
	rootB := mkdir(t, filepath.Join(dir, "root-b"))
	aDoc := recentFixtureFile(t, rootA, "a.md")
	bDoc := recentFixtureFile(t, rootB, "b.md")

	docs := map[string]docRecord{
		aDoc: {LastViewed: time.Now().Add(-time.Hour)},
		bDoc: {LastViewed: time.Now()},
	}
	roots := []root{{path: rootA}, {path: rootB}}

	rows := recentRows(docs, roots, rootA, activeRootBrowseScope)
	if len(rows) != 1 || rows[0].path != aDoc {
		t.Errorf("Active Root scope rows = %+v, want exactly [%q]", rows, aDoc)
	}
}

// TestRecentRowsOmitsDeletedFile is M4.2's own acceptance wording, last
// clause: a record whose file no longer exists on disk is silently dropped,
// never left in the Recent list as e.g. a broken entry (S6, unlike a broken
// star \u2014 M4.3's own concern).
func TestRecentRowsOmitsDeletedFile(t *testing.T) {
	dir := t.TempDir()
	rootDir := mkdir(t, filepath.Join(dir, "root"))
	present := recentFixtureFile(t, rootDir, "present.md")
	gone := filepath.Join(rootDir, "gone.md") // never created on disk

	docs := map[string]docRecord{
		present: {LastViewed: time.Now()},
		gone:    {LastViewed: time.Now().Add(time.Minute)}, // newer, but must still be dropped
	}
	roots := []root{{path: rootDir}}

	rows := recentRows(docs, roots, rootDir, activeRootBrowseScope)
	if len(rows) != 1 || rows[0].path != present {
		t.Errorf("recentRows = %+v, want exactly the present file, gone.md silently omitted", rows)
	}
}

// TestRecentRowsIgnoresNeverViewedRecords checks S4's "a zero LastViewed
// never appears" directly: a starred-but-never-viewed record (LastViewed
// zero) is absent from every scope, matching S1's independent Starred/
// LastViewed fields.
func TestRecentRowsIgnoresNeverViewedRecords(t *testing.T) {
	dir := t.TempDir()
	rootDir := mkdir(t, filepath.Join(dir, "root"))
	neverViewed := recentFixtureFile(t, rootDir, "never.md")

	docs := map[string]docRecord{neverViewed: {Starred: true}}
	roots := []root{{path: rootDir}}

	if rows := recentRows(docs, roots, rootDir, activeRootBrowseScope); len(rows) != 0 {
		t.Errorf("Active Root scope rows = %+v, want none (zero LastViewed)", rows)
	}
	if rows := recentRows(docs, roots, rootDir, allRootsBrowseScope); len(rows) != 0 {
		t.Errorf("All Roots scope rows = %+v, want none (zero LastViewed)", rows)
	}
}

// TestRecentRowsActiveRootScopeWithNoActiveRootIsEmpty covers the "" active-
// root-path edge case explicitly: with no root active, Active Root scope has
// no bucket to read, not the orphan bucket by accident (both key on "").
func TestRecentRowsActiveRootScopeWithNoActiveRootIsEmpty(t *testing.T) {
	dir := t.TempDir()
	orphanDoc := recentFixtureFile(t, dir, "orphan.md")
	docs := map[string]docRecord{orphanDoc: {LastViewed: time.Now()}}

	rows := recentRows(docs, nil, "", activeRootBrowseScope)
	if len(rows) != 0 {
		t.Errorf("recentRows with no active root = %+v, want none", rows)
	}
}

// TestFilesNavRecentModeRendersFlatRows is M4.2's own model-level wiring
// check: switching the Files Nav to Recent mode (keys.FilesMode) actually
// shows the flat list — file names and each row's own right-hand label —
// rather than continuing to show the Tree's rows underneath the mode/scope
// row M4.1 already reserved.
func TestFilesNavRecentModeRendersFlatRows(t *testing.T) {
	dir := t.TempDir()
	rootDir := mkdir(t, filepath.Join(dir, "project"))
	viewed := recentFixtureFile(t, rootDir, "viewed.md")
	recentFixtureFile(t, rootDir, "unviewed.md")

	node, err := tree.Walk(rootDir)
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	docs := map[string]docRecord{viewed: {LastViewed: time.Now()}}
	m := sized(t, newModel().withDocs("", docs), 80, 24)
	m.roots = m.roots.setRoots([]root{{path: rootDir}}, 0)
	m = m.openFolder(node)

	m = press(t, m, keys.FilesMode.Keys()[0]) // Tree -> Recent
	if m.files.mode != recentBrowse {
		t.Fatalf("setup: mode = %v, want recentBrowse", m.files.mode)
	}

	view := m.files.view()
	if !strings.Contains(view, "viewed.md") {
		t.Errorf("Recent view does not show the viewed document:\n%s", view)
	}
	if strings.Contains(view, "unviewed.md") {
		t.Errorf("Recent view shows a never-viewed document:\n%s", view)
	}
	if !strings.Contains(view, "project") {
		t.Errorf("Recent view does not show the Root's base name as the row's label:\n%s", view)
	}
}

// TestFilesNavRecentModeEmptyShowsAMessage checks Recent mode's own resting
// state: with nothing recently viewed, the pane shows a message rather than
// a blank list (S6's general "nothing to show" rule, ahead of M4.4's own
// per-combination polish).
func TestFilesNavRecentModeEmptyShowsAMessage(t *testing.T) {
	dir := t.TempDir()
	rootDir := mkdir(t, filepath.Join(dir, "project"))
	recentFixtureFile(t, rootDir, "unviewed.md")

	node, err := tree.Walk(rootDir)
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	m := sized(t, newModel().withDocs("", map[string]docRecord{}), 80, 24)
	m.roots = m.roots.setRoots([]root{{path: rootDir}}, 0)
	m = m.openFolder(node)

	m = press(t, m, keys.FilesMode.Keys()[0]) // Tree -> Recent

	view := m.files.view()
	if !strings.Contains(view, recentEmptyText) {
		t.Errorf("empty Recent view = %q, want it to contain %q", view, recentEmptyText)
	}
}
