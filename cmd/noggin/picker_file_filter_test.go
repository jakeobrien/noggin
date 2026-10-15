package main

import (
	"path/filepath"
	"slices"
	"testing"

	"noggin/internal/tree"
)

// pickerFilterRoots makes two Roots, each with a Markdown file, a Text file
// and a binary named like a Text file, all containing "needle":
//
//	a/alpha.md  a/alpha.go  a/alpha.bin.txt (binary)
//	b/beta.md   b/beta.txt
func pickerFilterRoots(t *testing.T) (a, b string) {
	t.Helper()
	base := t.TempDir()
	a = mkdir(t, filepath.Join(base, "a"))
	b = mkdir(t, filepath.Join(base, "b"))
	writeFileContent(t, filepath.Join(a, "alpha.md"), "needle in alpha md\n")
	writeFileContent(t, filepath.Join(a, "alpha.go"), "needle in alpha go\n")
	writeFileContent(t, filepath.Join(a, "alpha.bin.txt"), "needle\x00binary\n")
	writeFileContent(t, filepath.Join(b, "beta.md"), "needle in beta md\n")
	writeFileContent(t, filepath.Join(b, "beta.txt"), "needle in beta txt\n")
	return a, b
}

// filterPickerModel has both Roots in the session with the first active, and
// the Files Nav focused, ready to open the Picker.
func filterPickerModel(t *testing.T) (m model, a, b string) {
	t.Helper()
	requireRipgrep(t)
	a, b = pickerFilterRoots(t)
	m = sized(t, newModel(), 100, 30).withRipgrep(detectRipgrep())
	m.roots = m.roots.setRoots([]root{{path: a}, {path: b}}, 0)
	return m.activateRoot(), a, b
}

func pressPickerFilter(t *testing.T, m model) (model, int) {
	t.Helper()
	next, cmd := m.Update(keyPress(keys.PickerFilter.Keys()[0]))
	m = concrete(t, next)
	if cmd == nil {
		return m, 0
	}
	msg, ok := cmd().(pickerSearchMsg)
	if !ok {
		t.Fatalf("alt+a's Cmd produced %T, want pickerSearchMsg", cmd())
	}
	return m, msg.seq
}

func searchPicker(t *testing.T, m model, query string) model {
	t.Helper()
	m, seqs := typePickerQuery(t, m, query)
	m, n := deliverPickerTicks(t, m, seqs)
	if n == 0 {
		t.Fatalf("no search was dispatched for %q", query)
	}
	return m
}

// researchPicker re-runs the current Prompt after an alt+a whose Cmd carried seq.
func researchPicker(t *testing.T, m model, seq int) model {
	t.Helper()
	m, n := deliverPickerTicks(t, m, []int{seq})
	if n == 0 {
		t.Fatal("alt+a did not re-run the search")
	}
	return m
}

func resultBases(m model) []string {
	var out []string
	for _, f := range m.picker.results.files {
		out = append(out, filepath.Base(f.path))
	}
	for _, g := range m.picker.results.grep {
		out = append(out, filepath.Base(g.file.path))
	}
	slices.Sort(out)
	return out
}

func TestPickerFileFilterTogglesFileResultsAndTheTitle(t *testing.T) {
	m, _, _ := filterPickerModel(t)
	m, _ = m.openPicker()
	if got := pickerModeScope(m.picker, m.searchBounds()); got != "File · Active Source · Markdown" {
		t.Errorf("title on open = %q", got)
	}
	m = searchPicker(t, m, "alpha")
	if got := resultBases(m); !slices.Equal(got, []string{"alpha.md"}) {
		t.Fatalf("results under Markdown = %v", got)
	}

	m, seq := pressPickerFilter(t, m)
	if got := pickerModeScope(m.picker, m.searchBounds()); got != "File · Active Source · All" {
		t.Errorf("title under All = %q", got)
	}
	m = researchPicker(t, m, seq)
	if got := resultBases(m); !slices.Equal(got, []string{"alpha.go", "alpha.md"}) {
		t.Errorf("results under All = %v, want Markdown and Text files, binary hidden", got)
	}

	m, seq = pressPickerFilter(t, m)
	m = researchPicker(t, m, seq)
	if got := resultBases(m); !slices.Equal(got, []string{"alpha.md"}) {
		t.Errorf("results after toggling back = %v", got)
	}
}

func TestPickerFileFilterAppliesToGrep(t *testing.T) {
	m, _, _ := filterPickerModel(t)
	m, _ = m.openPicker()
	m = m.switchPickerModeForTest()
	m = searchPicker(t, m, "needle")
	if got := resultBases(m); !slices.Equal(got, []string{"alpha.md"}) {
		t.Fatalf("grep under Markdown = %v", got)
	}

	m, seq := pressPickerFilter(t, m)
	if got := pickerModeScope(m.picker, m.searchBounds()); got != "Grep · Active Root · All" {
		t.Errorf("title = %q", got)
	}
	m = researchPicker(t, m, seq)
	if got := resultBases(m); !slices.Equal(got, []string{"alpha.go", "alpha.md"}) {
		t.Errorf("grep under All = %v, want binary skipped", got)
	}
}

func TestPickerFileFilterIsTheFilesNavsPerRootState(t *testing.T) {
	m, a, b := filterPickerModel(t)
	m, _ = m.openPicker()
	m, _ = pressPickerFilter(t, m)

	if got := m.files.filterFor(a); got != tree.FilterAll {
		t.Errorf("Active Root's filter = %v, want All", got)
	}
	if got := m.files.filterFor(b); got != tree.FilterMarkdown {
		t.Errorf("other Root's filter = %v, want untouched", got)
	}
	if got := rowNames(m.files); !slices.Contains(got, "alpha.go") {
		t.Errorf("Tree rows = %v, want the Tree re-walked under All", got)
	}

	// A new Picker opens on the Root's filter, and the Files Nav's `a`
	// changes what it opens on.
	m = press(t, m, keys.Cancel.Keys()[0])
	m, _ = m.openPicker()
	if m.searchBounds().filter != tree.FilterAll {
		t.Errorf("reopened Picker filter = %v, want All", m.searchBounds().filter)
	}
	m = press(t, m, keys.Cancel.Keys()[0])
	m = pressFilter(t, m.setFocus(filesPane))
	m, _ = m.openPicker()
	if m.searchBounds().filter != tree.FilterMarkdown {
		t.Errorf("Picker filter after the Files Nav's `a` = %v, want Markdown", m.searchBounds().filter)
	}
}

func TestPickerAllRootsUsesTheActiveRootsFilterForEveryRoot(t *testing.T) {
	m, _, _ := filterPickerModel(t)
	m = pressFilter(t, m.setFocus(filesPane)) // Root a on All; Root b untouched
	m, _ = m.openPicker()
	next, _ := m.Update(keyPress(keys.PickerScope.Keys()[0]))
	m = concrete(t, next)
	if m.picker.scope != allRootsScope {
		t.Fatalf("scope = %v, want All Roots", m.picker.scope)
	}

	m = searchPicker(t, m, "beta")
	if got := resultBases(m); !slices.Equal(got, []string{"beta.md", "beta.txt"}) {
		t.Errorf("All Roots results = %v, want Root b's Text file too (Root a's filter applies)", got)
	}
	if got := pickerModeScope(m.picker, m.searchBounds()); got != "File · All Roots · All" {
		t.Errorf("title = %q", got)
	}

	m, seq := pressPickerFilter(t, m)
	m = researchPicker(t, m, seq)
	if got := resultBases(m); !slices.Equal(got, []string{"beta.md"}) {
		t.Errorf("All Roots results after toggling = %v", got)
	}
}

func TestPickerStarredScopeIgnoresTheFilter(t *testing.T) {
	m, a, _ := filterPickerModel(t)
	storePath := filepath.Join(t.TempDir(), "documents")
	m = m.withDocs(storePath, loadDocStore(storePath))
	m = m.withDocRecord(filepath.Join(a, "alpha.go"), docRecord{Starred: true})
	m, _ = m.openPicker()
	for i := 0; i < 2; i++ {
		next, _ := m.Update(keyPress(keys.PickerScope.Keys()[0]))
		m = concrete(t, next)
	}
	if m.picker.scope != starredScope {
		t.Fatalf("scope = %v, want Starred", m.picker.scope)
	}
	if got := pickerModeScope(m.picker, m.searchBounds()); got != "File · Starred" {
		t.Errorf("title = %q, want no filter named in Starred scope", got)
	}

	m = searchPicker(t, m, "alpha")
	if got := resultBases(m); !slices.Equal(got, []string{"alpha.go"}) {
		t.Fatalf("Starred results under Markdown = %v, want the starred Text file", got)
	}

	// Toggling flips the shared state but needs no new search: the rows on
	// screen stay current.
	m, seq := pressPickerFilter(t, m)
	if seq != 0 {
		t.Errorf("alt+a in Starred scope scheduled a search")
	}
	if !m.picker.resultsCurrent(m.searchBounds()) {
		t.Errorf("Starred results went stale after alt+a")
	}
	if got := m.files.filterFor(a); got != tree.FilterAll {
		t.Errorf("Active Root's filter = %v, want All", got)
	}
}

func TestPickerFileFilterWithWikipediaActiveUsesTheLastSetValue(t *testing.T) {
	requireRipgrep(t)
	a, b := pickerFilterRoots(t)
	m := sized(t, newModel(), 100, 30).withRipgrep(detectRipgrep())
	m.roots = m.roots.setRoots([]root{{path: a}, {path: b}}, 0)
	m = m.activateWikipediaSource()

	m, _ = m.openPicker()
	if m.files.fileFilter() != tree.FilterMarkdown {
		t.Fatalf("filter with nothing set = %v, want Markdown", m.files.fileFilter())
	}

	// alt+a toggles the last-set value; no Root's filter changes.
	m, _ = pressPickerFilter(t, m)
	if m.files.lastFilter != tree.FilterAll || m.files.fileFilter() != tree.FilterAll {
		t.Fatalf("lastFilter/fileFilter = %v/%v, want All/All", m.files.lastFilter, m.files.fileFilter())
	}
	for _, p := range []string{a, b} {
		if got := m.files.filterFor(p); got != tree.FilterMarkdown {
			t.Errorf("filter of %s = %v, want untouched", p, got)
		}
	}

	// File mode's All Roots scope (the one Wikipedia leaves) searches under it.
	next, _ := m.Update(keyPress(keys.PickerMode.Keys()[0]))
	m = concrete(t, next)
	if m.picker.mode != fileMode || m.picker.scope != allRootsScope {
		t.Fatalf("mode/scope = %v/%v, want File/All Roots", m.picker.mode, m.picker.scope)
	}
	m = searchPicker(t, m, "beta")
	if got := resultBases(m); !slices.Equal(got, []string{"beta.md", "beta.txt"}) {
		t.Errorf("results = %v, want All's", got)
	}
	if got := pickerModeScope(m.picker, m.searchBounds()); got != "File · All Roots · All" {
		t.Errorf("title = %q", got)
	}

	// The value survives closing and reopening.
	m = press(t, m, keys.Cancel.Keys()[0])
	m, _ = m.openPicker()
	if m.files.fileFilter() != tree.FilterAll {
		t.Errorf("reopened filter = %v, want the last set value All", m.files.fileFilter())
	}
}

func TestPickerFilterFollowsTheMostRecentlySetValueIntoWikipedia(t *testing.T) {
	m, _, _ := filterPickerModel(t)
	m = pressFilter(t, m.setFocus(filesPane)) // `a` sets All
	m = m.activateWikipediaSource()
	m, _ = m.openPicker()
	if m.files.fileFilter() != tree.FilterAll {
		t.Errorf("filter = %v, want the most recently set value All", m.files.fileFilter())
	}
}

func TestPickerFileFilterInWikipediaModeIsSaidInTheFooterNotTheTitle(t *testing.T) {
	m := pickerWithWikipediaActive(t)
	if got := pickerModeScope(m.picker, m.searchBounds()); got != "Wikipedia · -" {
		t.Errorf("title = %q", got)
	}
	m, seq := pressPickerFilter(t, m)
	if seq != 0 {
		t.Errorf("alt+a in Wikipedia mode scheduled a search")
	}
	if got := pickerModeScope(m.picker, m.searchBounds()); got != "Wikipedia · -" {
		t.Errorf("title after alt+a = %q", got)
	}
	if m.message != "File Filter: All" {
		t.Errorf("message after alt+a = %q, want it to name the new File Filter", m.message)
	}
	m, _ = pressPickerFilter(t, m)
	if m.message != "File Filter: Markdown" {
		t.Errorf("message after the second alt+a = %q", m.message)
	}
	m, _ = pressPickerFilter(t, m)
	if got := m.files.fileFilter(); got != tree.FilterAll {
		t.Errorf("filter the next Picker opened from Wikipedia uses = %v, want All", got)
	}
}
