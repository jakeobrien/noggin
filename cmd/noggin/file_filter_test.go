package main

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"noggin/internal/tree"
)

// filterRoot makes a Root with a Markdown file, Text files (one nested), and a
// binary named like a Text file:
//
//	docs/guide.md  docs/notes.txt  README.md  main.go  blob.txt (binary)
func filterRoot(t *testing.T) string {
	t.Helper()
	dir := mkdir(t, filepath.Join(t.TempDir(), "root"))
	sub := mkdir(t, filepath.Join(dir, "docs"))
	writeFileContent(t, filepath.Join(sub, "guide.md"), "# guide\n")
	writeFileContent(t, filepath.Join(sub, "notes.txt"), "notes\n")
	writeFileContent(t, filepath.Join(dir, "README.md"), "# readme\n")
	writeFileContent(t, filepath.Join(dir, "main.go"), "package main\n")
	writeFileContent(t, filepath.Join(dir, "blob.txt"), "a\x00b")
	return dir
}

// filterModel is a model with the Root at dir active and the Files Nav focused.
func filterModel(t *testing.T, dir string) model {
	t.Helper()
	m := sized(t, newModel(), 80, 24)
	m.roots = m.roots.setRoots([]root{{path: dir}}, -1)
	return m.activateRoot()
}

func pressFilter(t *testing.T, m model) model {
	t.Helper()
	return press(t, m, keys.FileFilter.Keys()[0])
}

func statusRow(m model) string {
	return strings.TrimSpace(ansi.Strip(strings.Split(m.files.view(), "\n")[0]))
}

func TestFileFilterDefaultsToMarkdownAndTheStatusRowSaysSo(t *testing.T) {
	m := filterModel(t, filterRoot(t))
	if got := rowNames(m.files); !slices.Equal(got, []string{"docs", "README.md"}) {
		t.Errorf("rows = %v, want Markdown files only", got)
	}
	if got := statusRow(m); !strings.Contains(got, "Tree \u00b7 - \u00b7 Markdown") {
		t.Errorf("status row = %q, want it to end Tree \u00b7 - \u00b7 Markdown", got)
	}
}

func TestFileFilterKeyTogglesBetweenMarkdownAndAll(t *testing.T) {
	m := filterModel(t, filterRoot(t))

	m = pressFilter(t, m)
	if got := rowNames(m.files); !slices.Equal(got, []string{"docs", "README.md", "main.go"}) {
		t.Errorf("rows under All = %v, want Markdown and Text files, binary hidden", got)
	}
	if got := statusRow(m); !strings.Contains(got, "Tree \u00b7 - \u00b7 All") {
		t.Errorf("status row under All = %q", got)
	}

	m = pressFilter(t, m)
	if got := rowNames(m.files); !slices.Equal(got, []string{"docs", "README.md"}) {
		t.Errorf("rows after toggling back = %v, want Markdown only", got)
	}
	if got := statusRow(m); !strings.Contains(got, "Tree \u00b7 - \u00b7 Markdown") {
		t.Errorf("status row after toggling back = %q", got)
	}
}

func TestFileFilterIsRememberedPerRootForTheSession(t *testing.T) {
	a := filterRoot(t)
	b := filterRoot(t)
	m := sized(t, newModel(), 80, 24)
	m.roots = m.roots.setRoots([]root{{path: a}, {path: b}}, -1)
	m = m.activateRoot() // A
	m = pressFilter(t, m)

	m.roots = m.roots.move(1)
	m = m.activateRoot() // B: never toggled
	if got := m.files.filterFor(b); got != tree.FilterMarkdown {
		t.Errorf("Root B filter = %v, want Markdown (the default)", got)
	}
	if containsRow(m.files, "main.go") {
		t.Errorf("Root B rows = %v, want Markdown only", rowNames(m.files))
	}

	m.roots = m.roots.move(-1)
	m = m.activateRoot() // back to A
	if !containsRow(m.files, "main.go") {
		t.Errorf("Root A rows after A\u2192B\u2192A = %v, want its All filter remembered", rowNames(m.files))
	}
	if got := statusRow(m); !strings.HasSuffix(got, "All") {
		t.Errorf("Root A status row = %q, want All", got)
	}
}

func TestFileFilterToggleMovesSelectionToNearestSurvivingAncestor(t *testing.T) {
	m := filterModel(t, filterRoot(t))
	m = pressFilter(t, m) // All
	m.files = m.files.expand()
	// Select docs/notes.txt, which only All shows.
	for i := 0; m.files.selectedNode().Name != "notes.txt"; i++ {
		if i > 10 {
			t.Fatalf("notes.txt not reachable; rows = %v", rowNames(m.files))
		}
		m.files = m.files.move(1)
	}

	m = pressFilter(t, m) // Markdown hides it
	if got := m.files.selectedNode(); got == nil || got.Name != "docs" {
		t.Errorf("selection after the toggle hid it = %v, want docs (nearest surviving ancestor)", got)
	}
	if !containsRow(m.files, "guide.md") {
		t.Errorf("rows = %v, want docs still expanded", rowNames(m.files))
	}
}

func TestFileFilterToggleKeepsSelectionExpansionAndOpenDocuments(t *testing.T) {
	dir := filterRoot(t)
	m := filterModel(t, dir)
	m.files = m.files.expand() // docs
	m.files = m.files.move(1)  // docs/guide.md
	m = openPath(t, m, dv1Pane, filepath.Join(dir, "docs", "guide.md")).setFocus(filesPane)

	m = pressFilter(t, m)
	if got := m.files.selectedNode(); got == nil || got.Name != "guide.md" {
		t.Errorf("selection after toggling = %v, want guide.md kept", got)
	}
	if !containsRow(m.files, "notes.txt") || !containsRow(m.files, "guide.md") {
		t.Errorf("rows = %v, want docs still expanded", rowNames(m.files))
	}
	if m.dv1.doc.path != filepath.Join(dir, "docs", "guide.md") {
		t.Errorf("DV1 document = %q, want it kept open", m.dv1.doc.path)
	}
}

func TestFileFilterDoesNothingWithWikipediaActiveButSaysSo(t *testing.T) {
	dir := filterRoot(t)
	m := filterModel(t, dir)
	m = m.activateWikipediaSource().setFocus(filesPane)

	m = pressFilter(t, m)
	if !strings.Contains(m.message, "Wikipedia") || !strings.Contains(m.message, "File Filter") {
		t.Errorf("footer message = %q, want one saying Wikipedia has no File Filter", m.message)
	}
	if got := m.files.filterFor(dir); got != tree.FilterMarkdown {
		t.Errorf("filter of %s = %v, want it untouched", dir, got)
	}
}

func TestFileFilterWithNoRootIsAQuietNoOp(t *testing.T) {
	m := press(t, sized(t, newModel(), 80, 24).setFocus(filesPane), keys.FileFilter.Keys()[0])
	if m.message != "" {
		t.Errorf("message = %q, want none", m.message)
	}
}

func TestFileFilterIsOnlyShownForTreeAndRecentAndStarredIgnoreIt(t *testing.T) {
	for _, cycles := range []int{1, 2} { // Recent, Starred
		m := filterModel(t, filterRoot(t))
		for i := 0; i < cycles; i++ {
			m.files = m.files.cycleMode()
		}
		m = pressFilter(t, m)
		mode := m.files.mode.label()
		if got := m.files.filterFor(m.files.root.Path); got != tree.FilterAll {
			t.Errorf("%s: filter = %v, want All toggled from here too", mode, got)
		}
		if got := statusRow(m); strings.Contains(got, "All") || strings.Contains(got, "Markdown") {
			t.Errorf("%s status row = %q, want no filter shown", mode, got)
		}
		if !strings.Contains(m.message, "Recent and Starred ignore") {
			t.Errorf("%s: message = %q, want a note that Recent and Starred ignore the filter", mode, m.message)
		}
		if got := m.files.flatRows(); len(got) != 0 {
			t.Errorf("%s rows = %v, want the (empty) flat list unaffected by the filter", mode, got)
		}
	}
}

func TestFileFilterShowsAGitMarkerOnANewlyListedFile(t *testing.T) {
	requireGit(t)
	repo := newGitRepo(t)
	writeFileContent(t, filepath.Join(repo, "todo.txt"), "untracked text\n")
	m := filterModel(t, repo)
	if containsRow(m.files, "todo.txt") {
		t.Fatalf("setup: todo.txt shown under Markdown: %v", rowNames(m.files))
	}

	m = pressFilter(t, m)
	line := lineContaining(strings.Split(m.files.view(), "\n"), "todo.txt")
	if !strings.Contains(line, "A   todo.txt") {
		t.Errorf("todo.txt row = %q, want an A git marker like any shown file", line)
	}
}

// --- launch ---------------------------------------------------------------

func TestLaunchOnATextFileStartsItsAdHocRootOnAll(t *testing.T) {
	dir := filterRoot(t)
	m := sized(t, newModel().applyResolution(resolveFile(filepath.Join(dir, "main.go"), dir, nil)), 80, 24)
	if got := m.files.filterFor(dir); got != tree.FilterAll {
		t.Errorf("ad-hoc Root filter = %v, want All", got)
	}
	if !containsRow(m.files, "main.go") {
		t.Errorf("rows = %v, want the launched file in the Tree", rowNames(m.files))
	}
	if got := statusRow(m); !strings.HasSuffix(got, "All") {
		t.Errorf("status row = %q, want All", got)
	}
}

func TestLaunchOnAMarkdownFileOrADirectoryStaysOnMarkdown(t *testing.T) {
	dir := filterRoot(t)
	for name, res := range map[string]resolution{
		"markdown file": resolveFile(filepath.Join(dir, "README.md"), dir, nil),
		"directory":     prependAdhoc(dir, nil),
		"binary file":   resolveFile(filepath.Join(dir, "blob.txt"), dir, nil),
	} {
		m := newModel().applyResolution(res)
		if got := m.files.filterFor(dir); got != tree.FilterMarkdown {
			t.Errorf("%s: ad-hoc Root filter = %v, want Markdown", name, got)
		}
	}
}

func TestLaunchOnATextFileSwitchesAStarredCwdRootToAll(t *testing.T) {
	dir := filterRoot(t)
	starred := []root{{path: dir, starred: true}}
	m := sized(t, newModel().applyResolution(resolveFile(filepath.Join(dir, "main.go"), dir, starred)), 80, 24)
	if got := m.files.filterFor(dir); got != tree.FilterAll {
		t.Errorf("starred Root filter = %v, want All", got)
	}
	if got := m.files.fileFilter(); got != tree.FilterAll || m.files.lastFilter != tree.FilterAll {
		t.Errorf("most recently set filter = %v, want All", m.files.lastFilter)
	}
	if !containsRow(m.files, "main.go") {
		t.Fatalf("rows = %v, want the launched file in the Tree", rowNames(m.files))
	}
	if got := m.files.selectedNode(); got == nil || got.Name != "main.go" {
		t.Errorf("selection = %v, want the launched file", got)
	}
}

func TestLaunchOnANestedTextFileRevealsItInTheFilesNav(t *testing.T) {
	dir := filterRoot(t)
	for _, starred := range []bool{false, true} {
		roots := []root{{path: dir, starred: starred}}
		m := newModel().applyResolution(resolveFile(filepath.Join(dir, "docs", "notes.txt"), dir, roots[:boolToInt(starred)]))
		if got := m.files.selectedNode(); got == nil || got.Name != "notes.txt" {
			t.Errorf("starred=%v: selection = %v, want the launched nested file", starred, got)
		}
	}
}

func TestLaunchOnAMarkdownFileLeavesAStarredCwdRootOnMarkdown(t *testing.T) {
	dir := filterRoot(t)
	m := newModel().applyResolution(resolveFile(filepath.Join(dir, "README.md"), dir, []root{{path: dir, starred: true}}))
	if got := m.files.filterFor(dir); got != tree.FilterMarkdown {
		t.Errorf("starred Root filter = %v, want Markdown", got)
	}
}

// --- staleness (D139) -------------------------------------------------------

func TestStalenessUnderAllIgnoresHiddenBinariesButSeesTextFiles(t *testing.T) {
	for _, tc := range []struct {
		name, content string
		want          bool
	}{
		{"new.bin", "a\x00b", false},
		{"new.txt", "text\n", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := filterRoot(t)
			node, err := tree.Walk(root, tree.FilterAll)
			if err != nil {
				t.Fatal(err)
			}
			b := snapshotStaleness(node, tree.FilterAll)
			bumpMtimeForward(t, root)
			writeFileContent(t, filepath.Join(root, tc.name), tc.content)
			if got := b.stale(); got != tc.want {
				t.Errorf("stale() = %v after adding %s under All, want %v", got, tc.name, tc.want)
			}
		})
	}
}

func TestStalenessUnderMarkdownIgnoresTextFileChanges(t *testing.T) {
	root := filterRoot(t)
	b := snapshotStaleness(walkFor(t, root), tree.FilterMarkdown)
	bumpMtimeForward(t, root)
	writeFileContent(t, filepath.Join(root, "new.txt"), "text\n")
	if b.stale() {
		t.Error("stale() = true after adding a Text file under Markdown, want false")
	}
}

func TestFileFilterToggleRebasesTheTreeBaselineAndKeepsTheOpenDocumentTracked(t *testing.T) {
	dir := filterRoot(t)
	m := filterModel(t, dir)
	m = openPath(t, m, dv1Pane, filepath.Join(dir, "README.md")).setFocus(filesPane)
	gen := m.staleGen

	m = pressFilter(t, m) // All
	if m.staleGen == gen {
		t.Error("staleness generation did not move, so an in-flight check against the old filter would be applied")
	}
	if got := m.staleBaseline.trackedDocPaths(); !slices.Equal(got, []string{filepath.Join(dir, "README.md")}) {
		t.Errorf("tracked documents = %v, want the open one kept", got)
	}
	bumpMtimeForward(t, dir)
	writeFileContent(t, filepath.Join(dir, "later.txt"), "text\n")
	if !m.staleBaseline.stale() {
		t.Error("a new Text file after toggling to All did not raise the indicator")
	}
}

func TestFileFilterToggleKeepsAnAlreadyRaisedIndicator(t *testing.T) {
	m := filterModel(t, filterRoot(t))
	m.stale = true
	m = pressFilter(t, m)
	if !m.stale {
		t.Error("toggling the filter cleared the staleness indicator, but nothing was re-read")
	}
}

func TestRefreshKeepsTheRootsFileFilter(t *testing.T) {
	dir := filterRoot(t)
	m := pressFilter(t, filterModel(t, dir))
	m, _ = m.refresh()
	if !containsRow(m.files, "main.go") {
		t.Errorf("rows after refresh = %v, want the All filter still applied", rowNames(m.files))
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
