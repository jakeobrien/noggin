package main

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"

	"noggin/internal/tree"
)

// fixtureTree is a small tree shaped to exercise the pane: two nested folders
// (one with content, one empty), files at the root, so top-level entries,
// expansion and an empty folder can all be checked. Children are already
// dirs-first then by name, matching what tree.Walk produces.
func fixtureTree() *tree.Node {
	return &tree.Node{
		Name: "root", Path: "/root", IsDir: true,
		Children: []*tree.Node{
			{Name: "docs", Path: "/root/docs", IsDir: true, Children: []*tree.Node{
				{Name: "guide.md", Path: "/root/docs/guide.md"},
				{Name: "ref.md", Path: "/root/docs/ref.md"},
			}},
			{Name: "empty", Path: "/root/empty", IsDir: true},
			{Name: "README.md", Path: "/root/README.md"},
			{Name: "notes.txt", Path: "/root/notes.txt"},
		},
	}
}

// fixtureTreeB is a second, differently-Path'd tree (M3.1): distinct from
// fixtureTree's Path so filesNav.setRoot treats switching to/from it as a
// different Root to remember state against, not the same one reshown.
func fixtureTreeB() *tree.Node {
	return &tree.Node{
		Name: "root-b", Path: "/root-b", IsDir: true,
		Children: []*tree.Node{
			{Name: "top.md", Path: "/root-b/top.md"},
			{Name: "other.md", Path: "/root-b/other.md"},
		},
	}
}

// rowNames is the Name of each currently visible row, top to bottom.
func rowNames(f filesNav) []string {
	var names []string
	for _, r := range f.visibleRows() {
		names = append(names, r.node.Name)
	}
	return names
}

func newFiles(root *tree.Node) filesNav {
	return newFilesNav().setRoot(root)
}

func TestFilesOpensWithTopLevelCollapsed(t *testing.T) {
	f := newFiles(fixtureTree())
	got := rowNames(f)
	want := []string{"docs", "empty", "README.md", "notes.txt"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("opening rows = %v, want %v (only top-level, all collapsed)", got, want)
	}
	// A nested file must not be visible until its folder is opened.
	for _, r := range f.visibleRows() {
		if r.node.Name == "guide.md" {
			t.Fatal("guide.md visible before docs was expanded")
		}
	}
}

func TestFilesEmptyFolderIsListed(t *testing.T) {
	f := newFiles(fixtureTree())
	if !containsRow(f, "empty") {
		t.Fatal("empty folder not listed among top-level entries")
	}
	// Expanding it reveals nothing but leaves it listed (S6).
	f = f.move(1) // select "empty"
	if f.selectedNode().Name != "empty" {
		t.Fatalf("selected = %q, want empty", f.selectedNode().Name)
	}
	f = f.expand()
	got := rowNames(f)
	want := []string{"docs", "empty", "README.md", "notes.txt"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("rows after expanding empty folder = %v, want %v", got, want)
	}
}

func TestFilesExpandAndCollapseInPlace(t *testing.T) {
	f := newFiles(fixtureTree()) // selection on "docs"
	if f.selectedNode().Name != "docs" {
		t.Fatalf("selection starts on %q, want docs", f.selectedNode().Name)
	}

	f = f.expand()
	got := rowNames(f)
	want := []string{"docs", "guide.md", "ref.md", "empty", "README.md", "notes.txt"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("rows after expanding docs = %v, want %v", got, want)
	}
	// Expansion is in place: docs is still the selected row, at the same index.
	if f.selected != 0 || f.selectedNode().Name != "docs" {
		t.Errorf("selection after expand = index %d (%q), want 0 (docs)", f.selected, f.selectedNode().Name)
	}

	f = f.collapse()
	got = rowNames(f)
	want = []string{"docs", "empty", "README.md", "notes.txt"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("rows after collapsing docs = %v, want %v", got, want)
	}
	if f.selectedNode().Name != "docs" {
		t.Errorf("selection after collapse = %q, want docs", f.selectedNode().Name)
	}
}

func TestFilesSelectionMovesAcrossExpandedSet(t *testing.T) {
	f := newFiles(fixtureTree()).expand() // docs open; six visible rows

	// Walk the whole expanded set top to bottom.
	want := []string{"docs", "guide.md", "ref.md", "empty", "README.md", "notes.txt"}
	for i, name := range want {
		if f.selectedNode().Name != name {
			t.Errorf("row %d selection = %q, want %q", i, f.selectedNode().Name, name)
		}
		if i < len(want)-1 {
			f = f.move(1)
		}
	}

	// Movement clamps at the bottom.
	f = f.move(1)
	if f.selectedNode().Name != "notes.txt" {
		t.Errorf("moving down past the end = %q, want it to clamp on notes.txt", f.selectedNode().Name)
	}

	// And back up to the top, clamping there too.
	for range want {
		f = f.move(-1)
	}
	f = f.move(-1)
	if f.selected != 0 || f.selectedNode().Name != "docs" {
		t.Errorf("moving up past the top = index %d (%q), want 0 (docs)", f.selected, f.selectedNode().Name)
	}
}

// TestFilesExpansionIsValueSemantic guards the immutable-update convention: a
// retained copy of the pane must not see a later copy's expansion.
func TestFilesExpansionIsValueSemantic(t *testing.T) {
	before := newFiles(fixtureTree())
	after := before.expand()
	if containsRow(before, "guide.md") {
		t.Error("expanding a later copy leaked into an earlier one (shared map)")
	}
	if !containsRow(after, "guide.md") {
		t.Error("expand did not open the folder on the copy that asked for it")
	}
}

// TestFilesNavRestoresPerRootStateAcrossSwitches is M3.1's own acceptance
// criterion at the filesNav level: expanding/collapsing nodes and selecting
// one in Root A, switching to Root B and back to Root A, shows the same
// expansion and selection as left. TestActivateRootRestoresFilesNavPerRoot
// (rootsnav_test.go) checks the same thing driven through model.activateRoot,
// the path a user's keypresses actually take.
func TestFilesNavRestoresPerRootStateAcrossSwitches(t *testing.T) {
	rootA, rootB := fixtureTree(), fixtureTreeB()

	f := newFilesNav().setRoot(rootA)
	f = f.expand() // open docs
	f = f.move(1)  // select guide.md, inside it

	wantExpanded := len(f.expanded)
	wantSelected := f.selectedNode().Name
	if wantSelected != "guide.md" {
		t.Fatalf("setup: selection = %q, want guide.md", wantSelected)
	}

	// Root B has never been shown: setRoot starts it fresh, not carrying A's
	// expansion/selection over by accident.
	f = f.setRoot(rootB)
	if len(f.expanded) != 0 || f.selected != 0 {
		t.Fatalf("Root B's first visit = expanded %v selected %d, want fresh (empty, 0)", f.expanded, f.selected)
	}

	// Back to Root A: restored exactly as left, not reset like B's first visit.
	f = f.setRoot(rootA)
	if len(f.expanded) != wantExpanded {
		t.Errorf("Root A expansion after the A\u2192B\u2192A round trip = %v, want %d folder(s) open as left", f.expanded, wantExpanded)
	}
	if got := f.selectedNode(); got == nil || got.Name != wantSelected {
		t.Errorf("Root A selection after the round trip = %v, want %q", got, wantSelected)
	}
}

// TestFilesViewFillsBoxAndMarksSelection checks the row list, one row below
// S5's own mode/scope row (M4.1) — lines[0] is always that status line now,
// so the rows checked here are shifted down by one from before M4.1.
func TestFilesViewFillsBoxAndMarksSelection(t *testing.T) {
	f := newFiles(fixtureTree()).setSize(20, 6)
	view := f.view()
	lines := strings.Split(view, "\n")
	if len(lines) != 6 {
		t.Fatalf("view has %d lines, want 6 (the box height)", len(lines))
	}
	for i, line := range lines {
		if w := lipgloss.Width(line); w != 20 {
			t.Errorf("line %d width = %d, want 20", i, w)
		}
	}
	if !strings.Contains(lines[0], "Tree") || !strings.Contains(lines[0], "-") {
		t.Errorf("mode/scope row = %q, want it to show Tree and its no-scope dash", lines[0])
	}
	// The selected row (now lines[1], one below the mode/scope row) carries
	// the reverse-video SGR; a later row does not, so the selection is
	// visibly distinct.
	if !strings.Contains(lines[1], "docs") {
		t.Errorf("first file row = %q, want it to name docs", lines[1])
	}
	if !strings.Contains(lines[1], "\x1b[7m") {
		t.Errorf("selected row = %q, want reverse-video SGR", lines[1])
	}
	if strings.Contains(lines[2], "\x1b[7m") {
		t.Errorf("unselected row = %q, want no reverse-video SGR", lines[2])
	}
}

func TestFilesEmptyStateWhenNoRoot(t *testing.T) {
	f := newFilesNav().setSize(20, 6)
	lines := strings.Split(f.view(), "\n")
	if len(lines) != 6 {
		t.Fatalf("empty-state view has %d lines, want 6", len(lines))
	}
	// The mode/scope row (M4.1) still shows above the empty-state message,
	// which is one row shorter than the whole box for the same reason the
	// file list is (S5: "costing the file list one row of height in every
	// mode" applies with no root open too, files.go's view doc comment).
	if !strings.Contains(lines[0], "Tree") {
		t.Errorf("mode/scope row = %q, want it to show Tree", lines[0])
	}
	if !strings.Contains(f.view(), "No folder open") {
		t.Errorf("empty-state view = %q, want the no-folder message", f.view())
	}
}

// TestModelRoutesKeysToFocusedFilesPane checks the wiring: with focus on the
// Files Nav, the pane's own keys move the selection and open a folder, and the
// rendered navigator reflects it.
func TestModelRoutesKeysToFocusedFilesPane(t *testing.T) {
	m := sized(t, newModel().openFolder(fixtureTree()), 80, 24)
	if m.focus != filesPane {
		t.Fatalf("focus = %v, want filesPane after openFolder", m.focus)
	}
	if m.files.selectedNode().Name != "docs" {
		t.Fatalf("selection = %q, want docs", m.files.selectedNode().Name)
	}

	// Expand docs, then move down onto its first child.
	m = press(t, m, keys.Expand.Keys()[1])   // "l"
	m = press(t, m, keys.LineDown.Keys()[1]) // "j"
	if got := m.files.selectedNode().Name; got != "guide.md" {
		t.Errorf("selection after expand+down = %q, want guide.md", got)
	}
	if !strings.Contains(m.frame(), "guide.md") {
		t.Error("navigator frame does not show guide.md after expanding docs")
	}

	// Collapse from the parent hides it again.
	m = press(t, m, keys.LineUp.Keys()[1])   // "k" back to docs
	m = press(t, m, keys.Collapse.Keys()[1]) // "h"
	if strings.Contains(m.frame(), "guide.md") {
		t.Error("navigator frame still shows guide.md after collapsing docs")
	}
}

// TestActivateOpensSelectedFileAndMovesFocus checks M2.3's wiring end to end:
// selecting a real fixture file and pressing Activate loads it into the
// viewer and moves focus there, and the viewer's own scroll keys then act on
// it. fixtureTree's nodes are not backed by real files, so this walks the
// checked-in testdata/tree fixture instead (loadDocument needs a real path).
func TestActivateOpensSelectedFileAndMovesFocus(t *testing.T) {
	root, err := tree.Walk("testdata/tree", tree.FilterMarkdown)
	if err != nil {
		t.Fatalf("tree.Walk(testdata/tree): %v", err)
	}
	m := sized(t, newModel().openFolder(root), 80, 24)

	// docs sorts first (dirs-first, per fixtureTree's comment); step down onto
	// README.md, a file, before activating.
	for m.files.selectedNode().Name != "README.md" {
		m = press(t, m, keys.LineDown.Keys()[1])
	}

	m = press(t, m, keys.Activate.Keys()[0])
	if m.focus != dv1Pane {
		t.Fatalf("focus after activating a file = %v, want dv1Pane", m.focus)
	}
	if !strings.Contains(m.dv1.view(), "Root Readme") {
		t.Errorf("viewer content = %q, want it to show README.md's heading", m.dv1.view())
	}
	if !strings.Contains(m.dv1.label(), "README.md") {
		t.Errorf("viewer label = %q, want it to name README.md", m.dv1.label())
	}

	// Scroll keys now act on the viewer, not the Files Nav.
	before := m.dv1.offset()
	m = press(t, m, keys.LineDown.Keys()[1])
	if m.dv1.offset() == before && !m.dv1.atBottom() {
		t.Errorf("scroll key after activate did not move the viewer (offset stayed %d)", before)
	}
}

// TestActivateOnAFolderIsANoOp checks that M2.3 keeps activation file-only —
// a folder's Activate key does nothing, per the plan's note that folder-enter
// is left for later.
func TestActivateOnAFolderIsANoOp(t *testing.T) {
	m := sized(t, newModel().openFolder(fixtureTree()), 80, 24)
	if !m.files.selectedNode().IsDir {
		t.Fatalf("selection = %q, want a directory to start", m.files.selectedNode().Name)
	}
	m = press(t, m, keys.Activate.Keys()[0])
	if m.focus != filesPane {
		t.Errorf("focus after activating a folder = %v, want filesPane unchanged", m.focus)
	}
	if m.dv1.label() != "" {
		t.Errorf("viewer label = %q, want empty (no document opened)", m.dv1.label())
	}
}

// TestActivateWithAReadErrorLeavesSessionRunning checks the watch called out
// in the M2.2 handoff: a file that fails to load must not exit or crash the
// session, focus stays put, and the failure surfaces as a footer message
// (T15/M4.2) rather than failing silently.
func TestActivateWithAReadErrorLeavesSessionRunning(t *testing.T) {
	root := &tree.Node{
		Name: "root", Path: "/root", IsDir: true,
		Children: []*tree.Node{
			{Name: "missing.md", Path: "/does/not/exist/missing.md"},
		},
	}
	m := sized(t, newModel().openFolder(root), 80, 24)
	m = press(t, m, keys.Activate.Keys()[0])
	if m.focus != filesPane {
		t.Errorf("focus after a failed activate = %v, want filesPane unchanged", m.focus)
	}
	if m.dv1.label() != "" {
		t.Errorf("viewer label = %q, want empty (nothing opened on error)", m.dv1.label())
	}
	if m.message == "" {
		t.Error("message is empty after a failed activate, want a footer message")
	}
}

func containsRow(f filesNav, name string) bool {
	for _, r := range f.visibleRows() {
		if r.node.Name == name {
			return true
		}
	}
	return false
}
