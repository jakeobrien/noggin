package main

import (
	"path/filepath"
	"strings"
	"testing"

	"noggin/internal/tree"
)

// alt+h (hidden files) and alt+i (ignored files) are session-wide view
// toggles shared by the Files Nav Tree and the Picker, shown by a
// right-aligned indicator on the Files Nav status row.

// toggleRoot is a Root with a dotfile, a dot-directory, a gitignored file and
// folder, and a gitignored heavy folder:
//
//	.gitignore  README.md  .hidden.md  .notes/inner.md
//	ignored.md  build/out.md  node_modules/pkg.md
func toggleRoot(t *testing.T) string {
	t.Helper()
	dir := mkdir(t, filepath.Join(t.TempDir(), "root"))
	writeFileContent(t, filepath.Join(dir, ".gitignore"), "ignored.md\nbuild/\nnode_modules/\n")
	writeFileContent(t, filepath.Join(dir, "README.md"), "# readme\n")
	writeFileContent(t, filepath.Join(dir, ".hidden.md"), "# hidden\n")
	writeFileContent(t, filepath.Join(mkdir(t, filepath.Join(dir, ".notes")), "inner.md"), "# inner\n")
	writeFileContent(t, filepath.Join(dir, "ignored.md"), "# ignored\n")
	writeFileContent(t, filepath.Join(mkdir(t, filepath.Join(dir, "build")), "out.md"), "# out\n")
	writeFileContent(t, filepath.Join(mkdir(t, filepath.Join(dir, "node_modules")), "pkg.md"), "# pkg\n")
	return dir
}

func TestFilesNavHiddenToggleHidesDotfilesAndUpdatesIndicator(t *testing.T) {
	m := filterModel(t, toggleRoot(t))
	if !containsRow(m.files, ".hidden.md") || !containsRow(m.files, ".notes") {
		t.Fatalf("setup: rows = %v, want dotfiles shown by default", rowNames(m.files))
	}
	if got := statusRow(m); !strings.HasSuffix(got, "h    md") {
		t.Errorf("default status row = %q, want it to end h    md", got)
	}

	m = press(t, m, keys.ToggleHidden.Keys()[0])
	if containsRow(m.files, ".hidden.md") || containsRow(m.files, ".notes") {
		t.Errorf("rows = %v after alt+h, want no dotfiles", rowNames(m.files))
	}
	if !m.files.hideDotfiles || !m.searchBounds().hideDotfiles {
		t.Error("alt+h in the Files Nav did not set the shared hidden-files flag")
	}
	if got := statusRow(m); !strings.HasSuffix(got, "     md") || strings.HasSuffix(got, "h    md") {
		t.Errorf("status row = %q, want the h slot blank", got)
	}
}

func TestFilesNavIgnoredToggleRevealsIgnoredButNotHeavyDirs(t *testing.T) {
	m := filterModel(t, toggleRoot(t))
	if containsRow(m.files, "ignored.md") || containsRow(m.files, "build") {
		t.Fatalf("setup: rows = %v, want gitignored entries hidden by default", rowNames(m.files))
	}

	m = press(t, m, keys.ToggleIgnored.Keys()[0])
	if !containsRow(m.files, "ignored.md") || !containsRow(m.files, "build") {
		t.Errorf("rows = %v after alt+i, want gitignored entries shown", rowNames(m.files))
	}
	if containsRow(m.files, "node_modules") {
		t.Errorf("rows = %v, want node_modules still excluded", rowNames(m.files))
	}
	for _, r := range m.files.visibleRows() {
		if want := r.node.Name == "ignored.md" || r.node.Name == "build"; r.node.Ignored != want {
			t.Errorf("%s Ignored = %v, want %v", r.node.Name, r.node.Ignored, want)
		}
	}
	if got := statusRow(m); !strings.HasSuffix(got, "h i  md") {
		t.Errorf("status row = %q, want it to end h i  md", got)
	}
	if !m.searchBounds().showIgnored {
		t.Error("the Picker's bounds do not see the Files Nav's alt+i")
	}
}

// TestHiddenToggleMovesSelectionToNearestSurvivingAncestor: a selection the
// toggle hides falls back the way a File Filter toggle's does.
func TestHiddenToggleMovesSelectionToNearestSurvivingAncestor(t *testing.T) {
	m := filterModel(t, toggleRoot(t))
	m.files = m.files.withExpanded(filepath.Join(m.files.root.Path, ".notes"), true)
	for i, r := range m.files.visibleRows() {
		if r.node.Name == "inner.md" {
			m.files.selected = i
		}
	}
	if m.files.selectedNode().Name != "inner.md" {
		t.Fatalf("setup: selection = %q", m.files.selectedNode().Name)
	}
	m = press(t, m, keys.ToggleHidden.Keys()[0])
	if n := m.files.selectedNode(); n == nil || strings.HasPrefix(n.Name, ".") || n.Name == "inner.md" {
		t.Errorf("selection after hiding its dot-directory = %v, want a surviving row", n)
	}
}

func TestPickerTogglesShareStateWithTheFilesNavTree(t *testing.T) {
	requireRipgrep(t)
	m := filterModel(t, toggleRoot(t))
	m, _ = m.openPicker()
	if !m.picker.open {
		t.Fatal("setup: Picker did not open")
	}

	m = press(t, m, keys.ToggleIgnored.Keys()[0])
	if !m.files.showIgnored {
		t.Fatal("alt+i in the Picker did not set the shared ignored-files flag")
	}
	if !containsRow(m.files, "ignored.md") {
		t.Errorf("Tree rows = %v after the Picker's alt+i, want the Tree re-walked to show ignored.md", rowNames(m.files))
	}
	if got := pickerIndicator(m.picker, m.searchBounds()); got != "h i  md" {
		t.Errorf("Picker indicator = %q, want %q", got, "h i  md")
	}

	m = press(t, m, keys.ToggleHidden.Keys()[0])
	if containsRow(m.files, ".hidden.md") {
		t.Errorf("Tree rows = %v after the Picker's alt+h, want dotfiles hidden", rowNames(m.files))
	}
}

func TestViewToggleInAFlatModeSaysItIsIgnored(t *testing.T) {
	m := filterModel(t, toggleRoot(t))
	m = press(t, m, keys.FilesMode.Keys()[0]) // Tree -> Recent
	if !m.files.flatMode() {
		t.Fatal("setup: not in a flat mode")
	}
	m = press(t, m, keys.ToggleIgnored.Keys()[0])
	if !m.files.showIgnored {
		t.Error("alt+i in Recent did not flip the flag")
	}
	if !strings.Contains(m.message, "Recent and Starred ignore it") {
		t.Errorf("message = %q, want it to say Recent and Starred ignore it", m.message)
	}
}

func TestViewToggleWithWikipediaFlipsAndSaysSo(t *testing.T) {
	m := sized(t, newModel(), 80, 24)
	m = m.activateWikipediaSource()
	m.focus = filesPane
	m = press(t, m, keys.ToggleHidden.Keys()[0])
	if !m.files.hideDotfiles {
		t.Error("alt+h with Wikipedia active did not flip the flag")
	}
	if !strings.Contains(m.message, "Wikipedia") {
		t.Errorf("message = %q, want it to name the Wikipedia Source", m.message)
	}
}

func TestStatusLineDropsTheIndicatorBeforeTruncatingTheLabel(t *testing.T) {
	if got, want := statusLine("Tree · -", "h    md", 20), "Tree · -"+strings.Repeat(" ", 20-8-7)+"h    md"; got != want {
		t.Errorf("roomy = %q, want %q", got, want)
	}
	if got := statusLine("Tree · -", "h    md", 12); got != "Tree · -    " {
		t.Errorf("tight = %q, want the indicator dropped", got)
	}
	if got := statusLine("Tree · All Sources", "h    md", 6); got != "Tree …" {
		t.Errorf("tiny = %q, want the label truncated", got)
	}
}

func TestListFilesShowIgnoredUsesNoIgnoreVCSAndSkipsHeavyDirs(t *testing.T) {
	requireRipgrep(t)
	dir := toggleRoot(t)
	roots := []root{{path: dir}}

	names := func(files []pickerFile) map[string]bool {
		out := map[string]bool{}
		for _, f := range files {
			rel, _ := filepath.Rel(dir, f.path)
			out[filepath.ToSlash(rel)] = true
		}
		return out
	}
	off := names(listFiles(roots, 0, activeRootScope, pickerBounds{filter: tree.FilterMarkdown}))
	if off["ignored.md"] || off["build/out.md"] {
		t.Errorf("default listing = %v, want gitignored files excluded", off)
	}
	on := names(listFiles(roots, 0, activeRootScope, pickerBounds{showIgnored: true, filter: tree.FilterMarkdown}))
	if !on["ignored.md"] || !on["build/out.md"] {
		t.Errorf("showIgnored listing = %v, want gitignored files included", on)
	}
	if on["node_modules/pkg.md"] {
		t.Errorf("showIgnored listing = %v, want node_modules excluded", on)
	}
}
