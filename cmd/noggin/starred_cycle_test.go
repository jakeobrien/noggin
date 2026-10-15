package main

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"noggin/internal/tree"
)

// starredCycleFixture is a Root, project/, holding (in folder order)
// sub/deep.md, sub/z.md, a.md, b.md, c.md, plus other/ (a second Root) with
// o.md. Every file but b.md is starred; DV1 holds Focus with b.md open.
func starredCycleFixture(t *testing.T) (m model, paths map[string]string) {
	t.Helper()
	dir := t.TempDir()
	rootDir := mkdir(t, filepath.Join(dir, "project"))
	mkdir(t, filepath.Join(rootDir, "sub"))
	otherDir := mkdir(t, filepath.Join(dir, "other"))
	paths = map[string]string{}
	for _, rel := range []string{"a.md", "b.md", "c.md", "sub/z.md", "sub/deep.md"} {
		paths[rel] = writeFileContent(t, filepath.Join(rootDir, rel), rel+"\n")
	}
	paths["other/o.md"] = writeFileContent(t, filepath.Join(otherDir, "o.md"), "o\n")

	docs := map[string]docRecord{}
	for rel, p := range paths {
		if rel != "b.md" {
			docs[p] = docRecord{Starred: true}
		}
	}
	docsPath := filepath.Join(dir, "documents")
	if err := saveDocStore(docsPath, docs); err != nil {
		t.Fatalf("saveDocStore: %v", err)
	}

	node, err := tree.Walk(rootDir, tree.Options{Filter: tree.FilterMarkdown})
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	m = sized(t, newModel().withDocs(docsPath, loadDocStore(docsPath)), 100, 24)
	m.roots = m.roots.setRoots([]root{{path: rootDir}, {path: otherDir}}, 0)
	m = m.openFolder(node)
	d, err := loadDocument(paths["b.md"])
	if err != nil {
		t.Fatalf("loadDocument: %v", err)
	}
	m = m.navigate(dv1Pane, d)
	return m, paths
}

func pressStarredNext(t *testing.T, m model) model {
	t.Helper()
	return press(t, m, keys.StarredNext.Keys()[0])
}

func pressStarredPrevious(t *testing.T, m model) model {
	t.Helper()
	return press(t, m, keys.StarredPrevious.Keys()[0])
}

func TestStarredNextFromAnUnstarredDocumentLoadsTheFirstInFolderOrderThenCycles(t *testing.T) {
	m, paths := starredCycleFixture(t)

	want := []string{"sub/deep.md", "sub/z.md", "a.md", "c.md", "sub/deep.md"}
	for _, rel := range want {
		m = pressStarredNext(t, m)
		if m.dv1.doc.path != paths[rel] {
			t.Fatalf("after s: DV1 = %q, want %q (message %q)", m.dv1.doc.path, paths[rel], m.message)
		}
	}
	if m.focus != dv1Pane {
		t.Errorf("focus = %v, want dv1Pane", m.focus)
	}
}

func TestStarredPreviousCyclesBackwardsAndFromOutsideStartsAtTheLast(t *testing.T) {
	m, paths := starredCycleFixture(t)

	want := []string{"c.md", "a.md", "sub/z.md", "sub/deep.md", "c.md"}
	for _, rel := range want {
		m = pressStarredPrevious(t, m)
		if m.dv1.doc.path != paths[rel] {
			t.Fatalf("after S: DV1 = %q, want %q (message %q)", m.dv1.doc.path, paths[rel], m.message)
		}
	}
}

func TestStarredCycleRecordsHistoryAndRecency(t *testing.T) {
	m, paths := starredCycleFixture(t)

	m = pressStarredNext(t, m)
	if got, want := m.dv1.history, []string{paths["b.md"], paths["sub/deep.md"]}; !slices.Equal(got, want) {
		t.Errorf("history = %v, want %v", got, want)
	}
	if m.docRecord(paths["sub/deep.md"]).LastViewed.IsZero() {
		t.Error("cycled-to document has no LastViewed, want recency recorded")
	}
	m = press(t, m, keys.HistoryBack.Keys()[0])
	if m.dv1.doc.path != paths["b.md"] {
		t.Errorf("after back: DV1 = %q, want %q", m.dv1.doc.path, paths["b.md"])
	}
}

func TestStarredCycleActsOnTheFocusedViewerOnly(t *testing.T) {
	m, paths := starredCycleFixture(t)
	d, err := loadDocument(paths["c.md"])
	if err != nil {
		t.Fatalf("loadDocument: %v", err)
	}
	m = m.navigate(dv2Pane, d)

	m = pressStarredNext(t, m)
	if m.dv2.doc.path != paths["sub/deep.md"] {
		t.Errorf("DV2 = %q, want wrap from c.md to %q", m.dv2.doc.path, paths["sub/deep.md"])
	}
	if m.dv1.doc.path != paths["b.md"] {
		t.Errorf("DV1 = %q, want untouched %q", m.dv1.doc.path, paths["b.md"])
	}
}

func TestStarredCycleDoesNothingOutsideADocumentViewer(t *testing.T) {
	m, paths := starredCycleFixture(t)
	for _, focus := range []string{keys.FocusFiles.Keys()[0], keys.FocusRoots.Keys()[0]} {
		m = press(t, m, focus)
		m = pressStarredNext(t, m)
		m = pressStarredPrevious(t, m)
		if m.dv1.doc.path != paths["b.md"] {
			t.Errorf("focus %q: DV1 = %q, want untouched %q", focus, m.dv1.doc.path, paths["b.md"])
		}
	}
}

func TestStarredCycleFromAnEmptyViewerLoadsTheFirst(t *testing.T) {
	m, paths := starredCycleFixture(t)
	m = press(t, m, keys.CloseViewer.Keys()[0])
	if m.dv1.doc.isOpen() {
		t.Fatal("setup: DV1 still has a document")
	}
	m = pressStarredNext(t, m)
	if m.dv1.doc.path != paths["sub/deep.md"] {
		t.Errorf("DV1 = %q, want %q", m.dv1.doc.path, paths["sub/deep.md"])
	}
}

func TestStarredCycleSkipsBrokenStarsAndOtherRoots(t *testing.T) {
	m, paths := starredCycleFixture(t)
	if err := os.Remove(paths["sub/z.md"]); err != nil {
		t.Fatal(err)
	}
	var got []string
	for range 4 {
		m = pressStarredNext(t, m)
		got = append(got, m.dv1.doc.path)
	}
	want := []string{paths["sub/deep.md"], paths["a.md"], paths["c.md"], paths["sub/deep.md"]}
	if !slices.Equal(got, want) {
		t.Errorf("cycle = %v, want %v", got, want)
	}
}

func TestStarredCycleWithNoStarredDocumentsSaysSo(t *testing.T) {
	m, paths := starredCycleFixture(t)
	for rel, p := range paths {
		if rel != "other/o.md" {
			rec := m.docRecord(p)
			rec.Starred = false
			m = m.withDocRecord(p, rec)
		}
	}
	m = pressStarredNext(t, m)
	if m.dv1.doc.path != paths["b.md"] {
		t.Errorf("DV1 = %q, want untouched", m.dv1.doc.path)
	}
	if m.message != "no starred documents in project" {
		t.Errorf("message = %q", m.message)
	}
}

func TestStarredCycleWithWikipediaActiveSaysSo(t *testing.T) {
	m, paths := starredCycleFixture(t)
	m = m.activateWikipediaSource().setFocus(dv1Pane)
	m = pressStarredNext(t, m)
	if m.dv1.doc.path != paths["b.md"] {
		t.Errorf("DV1 = %q, want untouched", m.dv1.doc.path)
	}
	if m.message == "" {
		t.Error("no Footer message, want one saying Wikipedia is the Active Source")
	}
}

func TestFolderOrderLessMatchesTheTree(t *testing.T) {
	root := filepath.FromSlash("/r")
	in := []string{"b.md", "a/z.md", "a.md", "a/b/c.md", "B.md", "c/a.md"}
	var paths []string
	for _, rel := range in {
		paths = append(paths, filepath.Join(root, filepath.FromSlash(rel)))
	}
	slices.SortFunc(paths, func(x, y string) int {
		switch {
		case folderOrderLess(root, x, y):
			return -1
		case folderOrderLess(root, y, x):
			return 1
		}
		return 0
	})
	var got []string
	for _, p := range paths {
		rel, _ := filepath.Rel(root, p)
		got = append(got, filepath.ToSlash(rel))
	}
	want := []string{"a/b/c.md", "a/z.md", "c/a.md", "B.md", "a.md", "b.md"}
	if !slices.Equal(got, want) {
		t.Errorf("order = %v, want %v", got, want)
	}
}
