package main

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"noggin/internal/tree"
)

// TestPeekIntoRecentRowDV1 is M4.5's own acceptance wording, first clause:
// Peeking a Recent row loads it into DV1 — resolved against the flat row on
// screen (selectedFlatRow), not the unrelated tree row underneath
// (TRIAGE T21, togglePeekFlatRowInto).
func TestPeekIntoRecentRowDV1(t *testing.T) {
	dir := t.TempDir()
	rootDir := mkdir(t, filepath.Join(dir, "project"))
	target := recentFixtureFile(t, rootDir, "target.md")

	node, err := tree.Walk(rootDir, tree.Options{Filter: tree.FilterMarkdown})
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	docs := map[string]docRecord{target: {LastViewed: time.Now()}}
	m := sized(t, newModel().withDocs("", docs), 80, 24)
	m.roots = m.roots.setRoots([]root{{path: rootDir}}, 0)
	m = m.openFolder(node)

	m = press(t, m, keys.FilesMode.Keys()[0]) // Tree -> Recent

	m, _ = m.togglePeekInto(dv1Pane)
	if !m.dv1.peeking() || m.dv1.peek.doc.path != target {
		t.Fatalf("togglePeekInto(dv1Pane) did not peek %q in DV1: peek = %+v", target, m.dv1.peek.doc)
	}
	if m.dv1.doc.isOpen() {
		t.Errorf("Peek changed DV1's Active Document, want it untouched")
	}
}

// TestPeekIntoRecentRowDV2 is M4.5's own acceptance wording, second clause:
// the ctrl+ targeting variant lands a Recent row's Peek in DV2 instead,
// creating it first.
func TestPeekIntoRecentRowDV2(t *testing.T) {
	dir := t.TempDir()
	rootDir := mkdir(t, filepath.Join(dir, "project"))
	target := recentFixtureFile(t, rootDir, "target.md")

	node, err := tree.Walk(rootDir, tree.Options{Filter: tree.FilterMarkdown})
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	docs := map[string]docRecord{target: {LastViewed: time.Now()}}
	m := sized(t, newModel().withDocs("", docs), 80, 24)
	m.roots = m.roots.setRoots([]root{{path: rootDir}}, 0)
	m = m.openFolder(node)

	m = press(t, m, keys.FilesMode.Keys()[0]) // Tree -> Recent

	m, _ = m.togglePeekInto(dv2Pane)
	if m.dv2 == nil {
		t.Fatalf("togglePeekInto(dv2Pane) did not create DV2")
	}
	if !m.dv2.peeking() || m.dv2.peek.doc.path != target {
		t.Fatalf("togglePeekInto(dv2Pane) did not peek %q in DV2: peek = %+v", target, m.dv2.peek.doc)
	}
}

// TestActivateIntoRecentRowDV1AndDV2 confirms activateInto's flat-row
// resolution (already M4.3's fix) still works from Recent, not just
// Starred, into both viewers — the other half of M4.5's own acceptance
// wording ("tests activate and peek a Recent row into DV1 and into DV2").
func TestActivateIntoRecentRowDV1AndDV2(t *testing.T) {
	dir := t.TempDir()
	rootDir := mkdir(t, filepath.Join(dir, "project"))
	target := recentFixtureFile(t, rootDir, "target.md")

	node, err := tree.Walk(rootDir, tree.Options{Filter: tree.FilterMarkdown})
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	docs := map[string]docRecord{target: {LastViewed: time.Now()}}
	m := sized(t, newModel().withDocs("", docs), 80, 24)
	m.roots = m.roots.setRoots([]root{{path: rootDir}}, 0)
	m = m.openFolder(node)

	m = press(t, m, keys.FilesMode.Keys()[0]) // Tree -> Recent

	m, _ = m.activateInto(dv1Pane)
	if !m.dv1.doc.isOpen() || m.dv1.doc.path != target {
		t.Fatalf("activateInto(dv1Pane) did not open %q into DV1: doc = %+v", target, m.dv1.doc)
	}

	m, _ = m.activateInto(dv2Pane)
	if m.dv2 == nil || !m.dv2.doc.isOpen() || m.dv2.doc.path != target {
		t.Fatalf("activateInto(dv2Pane) did not open %q into DV2: dv2 = %+v", target, m.dv2)
	}
}

// TestPeekRefusesBrokenStarredRow mirrors TestActivateRefusesBrokenStarredRow
// (M4.3) for Peek (M4.5): a broken Starred row is refused with a footer
// message, the same "cannot be activated" treatment S3 gives Activate,
// rather than attempting the load.
func TestPeekRefusesBrokenStarredRow(t *testing.T) {
	dir := t.TempDir()
	rootDir := mkdir(t, filepath.Join(dir, "project"))
	gonePath := filepath.Join(rootDir, "gone.md")

	node, err := tree.Walk(rootDir, tree.Options{Filter: tree.FilterMarkdown})
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	docs := map[string]docRecord{gonePath: {Starred: true}}
	m := sized(t, newModel().withDocs("", docs), 80, 24)
	m.roots = m.roots.setRoots([]root{{path: rootDir}}, 0)
	m = m.openFolder(node)

	m = press(t, m, keys.FilesMode.Keys()[0])
	m = press(t, m, keys.FilesMode.Keys()[0]) // Tree -> Recent -> Starred
	if len(m.files.flatRows()) != 1 {
		t.Fatalf("setup: flatRows = %+v, want exactly the broken row", m.files.flatRows())
	}

	m, _ = m.togglePeekInto(dv1Pane)
	if m.dv1.peeking() {
		t.Errorf("togglePeekInto peeked the broken document, want refused")
	}
	if !strings.Contains(m.message, gonePath) || !strings.Contains(m.message, "not peeked") {
		t.Errorf("message = %q, want a refusal naming %q", m.message, gonePath)
	}
}

// TestExpandCollapseNoOpInFlatMode is M4.5's own acceptance wording, third
// clause: Expand/Collapse change nothing while the Files Nav is in a
// flat-list mode (Recent/Starred) — the keys must not silently mutate the
// hidden tree's expansion state via the unrelated row selectedNode() would
// otherwise resolve (TRIAGE T21).
func TestExpandCollapseNoOpInFlatMode(t *testing.T) {
	dir := t.TempDir()
	rootDir := mkdir(t, filepath.Join(dir, "project"))
	sub := mkdir(t, filepath.Join(rootDir, "sub"))
	target := recentFixtureFile(t, sub, "target.md")

	node, err := tree.Walk(rootDir, tree.Options{Filter: tree.FilterMarkdown})
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	docs := map[string]docRecord{target: {LastViewed: time.Now()}}
	m := sized(t, newModel().withDocs("", docs), 80, 24)
	m.roots = m.roots.setRoots([]root{{path: rootDir}}, 0)
	m = m.openFolder(node)

	m = press(t, m, keys.FilesMode.Keys()[0]) // Tree -> Recent
	before := m.files

	m = press(t, m, keys.Expand.Keys()[0])
	m = press(t, m, keys.Collapse.Keys()[0])

	if len(m.files.expanded) != len(before.expanded) {
		t.Errorf("expanded set changed in flat mode: before = %v, after = %v", before.expanded, m.files.expanded)
	}
	if m.files.mode != recentBrowse || m.files.selected != before.selected {
		t.Errorf("Expand/Collapse changed mode/selection in flat mode: before mode=%v selected=%d, after mode=%v selected=%d",
			before.mode, before.selected, m.files.mode, m.files.selected)
	}
}
