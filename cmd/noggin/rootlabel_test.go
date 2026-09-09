package main

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestRootDisplayNameIsTheOneRootLabelHelper is M7.3's own acceptance
// criterion, stated directly against the helper itself (roots.go's
// displayName, promoted by S8/M7.2/M7.3): a named Root reports its Name, an
// unnamed one falls back to the folder's base name — every other test in
// this file exercises a call site that must read from this helper rather
// than recomputing filepath.Base(<root path>) on its own.
func TestRootDisplayNameIsTheOneRootLabelHelper(t *testing.T) {
	named := root{path: "/tmp/project", name: "Work"}
	if got := named.displayName(); got != "Work" {
		t.Errorf("named root.displayName() = %q, want %q", got, "Work")
	}
	unnamed := root{path: "/tmp/project"}
	if got := unnamed.displayName(); got != "project" {
		t.Errorf("unnamed root.displayName() = %q, want the folder base name %q", got, "project")
	}
}

// TestRootsNavRowShowsRenamedRootLabel is M7.3's Sources-Nav-line-1 case:
// a renamed Root's row (rootsnav.go's renderRow) shows its Name, not the
// folder's base name, on line 1 — M7.2 wired renderRow to displayName
// already; this pins the outcome directly against a Root carrying a
// persisted Name (not one mid-rename, which rename_test.go already covers).
func TestRootsNavRowShowsRenamedRootLabel(t *testing.T) {
	dir := mkdir(t, filepath.Join(t.TempDir(), "project"))

	r := newRootsNav().setRoots([]root{{path: dir, starred: true, name: "Work"}}, 0).setSize(40, 2)
	lines := strings.Split(r.view(), "\n")
	if !strings.Contains(lines[0], "Work") {
		t.Errorf("row line 1 = %q, want the Root's Name (Work), not its folder base name", lines[0])
	}
	if strings.Contains(lines[0], "project") {
		t.Errorf("row line 1 = %q, want the folder base name (project) not to leak through when a Name is set", lines[0])
	}
}

// TestBuildBrowseRowUsesRenamedRootLabel is M7.3's cross-Source-Files-Nav-
// rows case: buildBrowseRow's rootLabel (the right-hand label Recent/
// Starred flat rows render, browse_list.go, files.go:817) is a covering
// Root's Name when it has one, not its folder's base name.
func TestBuildBrowseRowUsesRenamedRootLabel(t *testing.T) {
	dir := t.TempDir()
	rootDir := mkdir(t, filepath.Join(dir, "project"))
	doc := recentFixtureFile(t, rootDir, "a.md")

	row := buildBrowseRow(doc, []root{{path: rootDir, starred: true, name: "Work"}})
	if row.orphan {
		t.Fatalf("row = %+v, want orphan false (covered by a named Root)", row)
	}
	if row.rootLabel != "Work" {
		t.Errorf("rootLabel = %q, want the Root's Name %q, not its folder base name", row.rootLabel, "Work")
	}
}

// TestPickerRootLabelUsesRenamedRootLabel is M7.3's Picker-result-labels
// case: pickerRootLabel (picker_view.go, feeding pickerFileRow's Results
// rows) leads with a Root's Name when it has one, not its folder's base
// name — the same "prominent label plus disambiguating parent" layout it
// already gives an unnamed Root's base name.
func TestPickerRootLabelUsesRenamedRootLabel(t *testing.T) {
	f := pickerFile{
		root: root{path: "/home/user/work/admin/project", name: "Work"},
		path: "/home/user/work/admin/project/report.md",
	}
	got := pickerRootLabel(f, 40, "  report.md")
	if !strings.HasPrefix(got, "Work") {
		t.Errorf("pickerRootLabel() = %q, want it to lead with the Root's Name %q, not its folder base name", got, "Work")
	}
	if strings.Contains(got, "project ") || strings.HasPrefix(got, "project") {
		t.Errorf("pickerRootLabel() = %q, want the folder base name (project) not to leak through when a Name is set", got)
	}
}
