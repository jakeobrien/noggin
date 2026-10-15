package main

import (
	"path/filepath"
	"testing"

	"noggin/internal/tree"
)

// D141: Left in the Files Nav Tree collapses an expanded folder in place, and
// otherwise climbs: a file or collapsed folder inside a folder collapses that
// parent and selects it. At the top level, and in Recent/Starred, it does
// nothing.

func nestedTree() *tree.Node {
	return &tree.Node{
		Name: "root", Path: "/root", IsDir: true,
		Children: []*tree.Node{
			{Name: "docs", Path: "/root/docs", IsDir: true, Children: []*tree.Node{
				{Name: "api", Path: "/root/docs/api", IsDir: true, Children: []*tree.Node{
					{Name: "v1.md", Path: "/root/docs/api/v1.md"},
				}},
				{Name: "guide.md", Path: "/root/docs/guide.md"},
			}},
			{Name: "README.md", Path: "/root/README.md"},
		},
	}
}

// selectName moves the selection onto the visible row called name.
func selectName(t *testing.T, f filesNav, name string) filesNav {
	t.Helper()
	for i, r := range f.visibleRows() {
		if r.node.Name == name {
			f.selected = i
			return f
		}
	}
	t.Fatalf("no visible row %q in %v", name, rowNames(f))
	return f
}

func TestCollapseOnAnExpandedFolderCollapsesIt(t *testing.T) {
	f := newFiles(nestedTree()).withExpanded("/root/docs", true)
	f = selectName(t, f, "docs")
	f = f.collapse()
	if f.expanded["/root/docs"] {
		t.Error("docs still expanded")
	}
	if got := f.selectedNode().Name; got != "docs" {
		t.Errorf("selection = %q, want docs", got)
	}
}

func TestCollapseOnAFileClimbsToItsParent(t *testing.T) {
	f := newFiles(nestedTree()).withExpanded("/root/docs", true)
	f = selectName(t, f, "guide.md")
	f = f.collapse()
	if f.expanded["/root/docs"] {
		t.Error("docs still expanded after Left on guide.md")
	}
	if got := f.selectedNode().Name; got != "docs" {
		t.Errorf("selection = %q, want docs", got)
	}
}

func TestCollapseOnACollapsedFolderClimbsAndKeepsItsOwnState(t *testing.T) {
	f := newFiles(nestedTree()).withExpanded("/root/docs", true)
	f = selectName(t, f, "api") // collapsed folder inside docs
	f = f.collapse()
	if got := f.selectedNode().Name; got != "docs" {
		t.Fatalf("selection = %q, want docs", got)
	}
	if f.expanded["/root/docs"] {
		t.Error("docs still expanded")
	}

	// A nested expanded folder survives its parent collapsing.
	f = newFiles(nestedTree()).withExpanded("/root/docs", true).withExpanded("/root/docs/api", true)
	f = selectName(t, f, "v1.md")
	f = f.collapse() // v1.md -> api (collapses api)
	if got := f.selectedNode().Name; got != "api" || f.expanded["/root/docs/api"] {
		t.Fatalf("after Left on v1.md: selection %q, api expanded %v; want api, collapsed", got, f.expanded["/root/docs/api"])
	}
	f = f.withExpanded("/root/docs/api", true)
	f = selectName(t, f, "guide.md")
	f = f.collapse() // guide.md -> docs
	if !f.expanded["/root/docs/api"] {
		t.Error("collapsing docs forgot that api was expanded")
	}
}

func TestCollapseAtTheTopLevelIsANoOp(t *testing.T) {
	for _, name := range []string{"README.md", "docs"} {
		f := newFiles(nestedTree())
		f = selectName(t, f, name)
		before := f.selected
		f = f.collapse()
		if f.selected != before {
			t.Errorf("Left on top-level %s moved the selection to %d", name, f.selected)
		}
	}
}

func TestCollapseIsANoOpInFlatModes(t *testing.T) {
	f := newFiles(nestedTree()).withExpanded("/root/docs", true)
	f = selectName(t, f, "guide.md")
	f.mode = recentBrowse
	f = f.collapse()
	if !f.expanded["/root/docs"] {
		t.Error("Left in Recent collapsed a Tree folder")
	}
}

// TestCollapseToParentClearsPeek: climbing to the parent is a selection
// change, one of Peek's clearing triggers (S7).
func TestCollapseToParentClearsPeek(t *testing.T) {
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	mkdir(t, filepath.Join(root, "sub"))
	writeFileContent(t, filepath.Join(root, "sub", "inner.md"), "# Inner\n")
	node, err := tree.Walk(root, tree.Options{Filter: tree.FilterMarkdown})
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	m := sized(t, newModel().openFolder(node), 80, 24)
	m = press(t, m, keys.Expand.Keys()[0])
	m = press(t, m, keys.LineDown.Keys()[0])
	if m.files.selectedNode().Name != "inner.md" {
		t.Fatalf("setup: selection = %q, want inner.md", m.files.selectedNode().Name)
	}
	m = press(t, m, keys.Peek.Keys()[0])
	if !m.dv1.peeking() {
		t.Fatal("setup: Peek did not start")
	}

	m = press(t, m, keys.Collapse.Keys()[0])
	if m.files.selectedNode().Name != "sub" {
		t.Fatalf("selection = %q, want sub", m.files.selectedNode().Name)
	}
	if m.dv1.peeking() {
		t.Error("Peek survived the selection moving to the parent")
	}
}
