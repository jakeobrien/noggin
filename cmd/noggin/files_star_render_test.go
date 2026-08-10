package main

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestFilesNavRendersStarMarkers is M2.3's acceptance criterion: a starred
// file's Tree row carries the marker, an unstarred one doesn't, and a
// directory is never marked even if it happens to be a key in the set
// (mirrors TestFilesNavRendersGitStatusMarkers' identical directory-ignored
// check for S8's own per-file marker). The "absent from every row of a
// Starred list" clause is vacuous until M4.3 builds that mode; nothing to
// render yet that isn't Tree.
func TestFilesNavRendersStarMarkers(t *testing.T) {
	root := fixtureTree()
	f := newFiles(root).withStarred(map[string]bool{
		"/root/README.md": true,
		"/root/docs":      true, // a dir; must be ignored by render
	}).setSize(30, 6)

	view := f.view()
	lines := strings.Split(view, "\n")
	readmeLine := lineContaining(lines, "README.md")
	notesLine := lineContaining(lines, "notes.txt")
	docsLine := lineContaining(lines, "docs")

	if !strings.Contains(readmeLine, "★ README.md") {
		t.Errorf("README.md row = %q, want a star marker before the name", readmeLine)
	}
	if strings.Contains(notesLine, "★") {
		t.Errorf("notes.txt (unstarred) row = %q, want no star marker", notesLine)
	}
	if strings.Contains(docsLine, "★") {
		t.Errorf("docs (a directory) row = %q, want no star marker", docsLine)
	}
}

// TestToggleDocStarUpdatesTheRowMarkerImmediately confirms the marker isn't
// stale until the next openFolder/refresh: pressing the Files Nav's own
// Star key (M2.1) must show up in the very next render, and un-starring
// must remove it the same way.
func TestToggleDocStarUpdatesTheRowMarkerImmediately(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "documents")

	m := sized(t, newModel().withDocs(storePath, loadDocStore(storePath)), 80, 24)
	m = m.openFolder(fixtureTree())
	for m.files.selectedNode().Name != "README.md" {
		m.files = m.files.move(1)
	}

	before := lineContaining(strings.Split(m.files.view(), "\n"), "README.md")
	if strings.Contains(before, "★") {
		t.Fatalf("setup: README.md already marked before starring")
	}

	m = m.toggleDocStar()
	after := lineContaining(strings.Split(m.files.view(), "\n"), "README.md")
	if !strings.Contains(after, "★ README.md") {
		t.Errorf("README.md row after toggleDocStar = %q, want a star marker", after)
	}

	m = m.toggleDocStar()
	cleared := lineContaining(strings.Split(m.files.view(), "\n"), "README.md")
	if strings.Contains(cleared, "★") {
		t.Errorf("README.md row after un-starring = %q, want no star marker", cleared)
	}
}
