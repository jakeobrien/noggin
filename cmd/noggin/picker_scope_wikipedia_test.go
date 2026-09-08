package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// TestSwitchPickerModeSkipsActiveRootScopeWithWikipediaActive is M3.5's own
// rule 1: switching from wikipediaMode back to File or Grep with Wikipedia
// still the Active Source must never land on Active Root Scope (there is
// no Active Root to search) — it lands on All Roots instead, S6's "no
// empty dead-end scope".
func TestSwitchPickerModeSkipsActiveRootScopeWithWikipediaActive(t *testing.T) {
	m := pickerWithWikipediaActive(t)
	if m.picker.mode != wikipediaMode || m.picker.scope != activeRootScope {
		t.Fatalf("setup: mode = %v, scope = %v, want wikipediaMode/activeRootScope (unused)", m.picker.mode, m.picker.scope)
	}

	next, _ := m.Update(keyPress(keys.PickerMode.Keys()[0]))
	m = concrete(t, next)
	if m.picker.mode != fileMode {
		t.Fatalf("mode after one switch = %v, want fileMode", m.picker.mode)
	}
	if m.picker.scope != allRootsScope {
		t.Errorf("scope after switching into fileMode with Wikipedia active = %v, want allRootsScope, not the dead-end activeRootScope", m.picker.scope)
	}

	next, _ = m.Update(keyPress(keys.PickerMode.Keys()[0]))
	m = concrete(t, next)
	if m.picker.mode != grepMode || m.picker.scope != allRootsScope {
		t.Errorf("mode/scope after switching into grepMode = %v/%v, want grepMode/allRootsScope", m.picker.mode, m.picker.scope)
	}
}

// TestSwitchPickerScopeCyclesOnlyAllRootsAndStarredWithWikipediaActive is
// M3.5's own rule 1, the Scope key's own half: with Wikipedia the Active
// Source, the Scope key flips only between All Roots and Starred — Active
// Root Scope is never one of the values it can land on, in File or Grep
// mode alike.
func TestSwitchPickerScopeCyclesOnlyAllRootsAndStarredWithWikipediaActive(t *testing.T) {
	m := pickerWithWikipediaActive(t)
	next, _ := m.Update(keyPress(keys.PickerMode.Keys()[0])) // wikipediaMode -> fileMode
	m = concrete(t, next)
	if m.picker.scope != allRootsScope {
		t.Fatalf("setup: scope = %v, want allRootsScope", m.picker.scope)
	}

	want := []pickerScope{starredScope, allRootsScope, starredScope}
	for i, w := range want {
		next, _ := m.Update(keyPress(keys.PickerScope.Keys()[0]))
		m = concrete(t, next)
		if m.picker.scope != w {
			t.Fatalf("scope after Scope press %d = %v, want %v", i+1, m.picker.scope, w)
		}
	}

	// Grep mode cycles the exact same two values, never Active Root.
	next, _ = m.Update(keyPress(keys.PickerMode.Keys()[0])) // fileMode -> grepMode
	m = concrete(t, next)
	if m.picker.scope != starredScope {
		t.Fatalf("setup: scope after switching to grepMode = %v, want unchanged starredScope", m.picker.scope)
	}
	next, _ = m.Update(keyPress(keys.PickerScope.Keys()[0]))
	m = concrete(t, next)
	if m.picker.scope != allRootsScope {
		t.Errorf("scope after one Scope press in grepMode = %v, want allRootsScope", m.picker.scope)
	}
}

// wikipediaScopeFixture is M3.5's own combined fixture: a starred file
// under a session Root and a starred article, both matchable by the same
// query fragment ("euler"), plus an unstarred file and an unstarred
// article that must never surface regardless of scope or mode — the same
// "only the starred one" shape starredScopeFixture (picker_scope_starred_
// test.go, C4/M5.1) already gives its own file-only fixture, widened to
// cover both classes at once (S4/M1.5's "the store's count cap ... applies
// to both together" is the same store this reuses, not a second one).
func wikipediaScopeFixture(t *testing.T) (m model, filePath string) {
	t.Helper()
	dir := t.TempDir()
	rootDir := mkdir(t, filepath.Join(dir, "root"))
	filePath = writeFileContent(t, filepath.Join(rootDir, "leonhard-euler-notes.md"), "a mathematician's needle\n")
	writeFileContent(t, filepath.Join(rootDir, "unstarred.md"), "a needle nobody starred\n")

	storePath := filepath.Join(dir, "documents")
	m = sized(t, newModel().withDocs(storePath, loadDocStore(storePath)), 80, 24).
		withRipgrep(detectRipgrep())
	m.roots = m.roots.setRoots([]root{{path: rootDir}}, 0)
	m = m.withDocRecord(filePath, docRecord{Starred: true})
	if m.message != "" {
		t.Fatalf("starring the file: message = %q, want empty", m.message)
	}
	articleKey := "https://en.wikipedia.org/wiki/Leonhard_Euler"
	m = m.withDocRecord(articleKey, docRecord{Starred: true, Article: true, Title: "Leonhard Euler"})
	if m.message != "" {
		t.Fatalf("starring the article: message = %q, want empty", m.message)
	}
	unstarredArticleKey := "https://en.wikipedia.org/wiki/Isaac_Newton"
	m = m.withDocRecord(unstarredArticleKey, docRecord{Article: true, Title: "Isaac Newton"})
	if m.message != "" {
		t.Fatalf("recording the unstarred article: message = %q, want empty", m.message)
	}
	return m, filePath
}

// TestFileStarredScopeIncludesStarredArticleMatchedByTitle is M3.5's own
// rule 2: a File Picker search under Starred scope matches a starred
// article by title alongside a starred file, and the matched row carries
// the article's title (not any part of its URL) plus the article flag
// pickerFileRow renders as the "Wikipedia" label.
func TestFileStarredScopeIncludesStarredArticleMatchedByTitle(t *testing.T) {
	requireRipgrep(t)
	m, filePath := wikipediaScopeFixture(t)
	m, _ = m.openPicker()
	for i := 0; i < 2; i++ { // Active Source -> All Roots -> Starred
		next, _ := m.Update(keyPress(keys.PickerScope.Keys()[0]))
		m = concrete(t, next)
	}
	if m.picker.scope != starredScope {
		t.Fatalf("setup: scope = %v, want starredScope", m.picker.scope)
	}

	m, seqs := typePickerQuery(t, m, "euler")
	m, searches := deliverPickerTicks(t, m, seqs)
	if searches == 0 {
		t.Fatal("no File search was dispatched under Starred scope")
	}

	var gotFile, gotArticle bool
	for _, f := range m.picker.results.files {
		switch {
		case f.article:
			gotArticle = true
			if f.title != "Leonhard Euler" {
				t.Errorf("matched article title = %q, want %q", f.title, "Leonhard Euler")
			}
			if f.path != "https://en.wikipedia.org/wiki/Leonhard_Euler" {
				t.Errorf("matched article path (store key) = %q, want the canonical URL", f.path)
			}
		case f.path == filePath:
			gotFile = true
		}
	}
	if !gotFile {
		t.Errorf("File search under Starred scope for %q = %+v, missing the starred file", "euler", m.picker.results.files)
	}
	if !gotArticle {
		t.Errorf("File search under Starred scope for %q = %+v, missing the starred article", "euler", m.picker.results.files)
	}

	label := pickerLocationLabel(pickerFile{article: true, title: "Leonhard Euler"}, 40, "  Leonhard Euler")
	if label != "Wikipedia" {
		t.Errorf("pickerLocationLabel for an article = %q, want %q", label, "Wikipedia")
	}

	// End to end: the rendered Results list itself, not just the helper
	// function in isolation, carries the "Wikipedia" label and the title on
	// the same row (S6's "labelled Wikipedia").
	rendered := ansi.Strip(m.View().Content)
	var articleRow string
	for _, l := range strings.Split(rendered, "\n") {
		if strings.Contains(l, "Leonhard Euler") {
			articleRow = l
			break
		}
	}
	if articleRow == "" {
		t.Fatalf("rendered Results list has no row for the article\n%s", rendered)
	}
	if !strings.Contains(articleRow, "Wikipedia") {
		t.Errorf("rendered article row = %q, want it labelled %q", articleRow, "Wikipedia")
	}
}

// TestFileAllRootsScopeExcludesArticles is M3.5's own rule 3: File mode's
// All Roots scope is files only, even when a starred article's title would
// otherwise match the same query — Starred is the *only* File search that
// spans both classes (S6).
func TestFileAllRootsScopeExcludesArticles(t *testing.T) {
	requireRipgrep(t)
	m, filePath := wikipediaScopeFixture(t)
	m, _ = m.openPicker()
	next, _ := m.Update(keyPress(keys.PickerScope.Keys()[0])) // Active Source -> All Roots
	m = concrete(t, next)
	if m.picker.scope != allRootsScope {
		t.Fatalf("setup: scope = %v, want allRootsScope", m.picker.scope)
	}

	m, seqs := typePickerQuery(t, m, "euler")
	m, searches := deliverPickerTicks(t, m, seqs)
	if searches == 0 {
		t.Fatal("no File search was dispatched under All Roots scope")
	}

	if len(m.picker.results.files) != 1 || m.picker.results.files[0].path != filePath {
		t.Fatalf("File search under All Roots scope for %q = %+v, want exactly the file [%s]", "euler", m.picker.results.files, filePath)
	}
	if m.picker.results.files[0].article {
		t.Error("File search under All Roots scope returned an article result, want files only")
	}
}

// TestPickerActivateOnArticleRowInFileModeRefusesCleanlyForNow is C6/M3.5's
// own guard against a gap M3.5 itself opens: File mode's Starred scope now
// selects a real article pseudo-candidate (pickerFile.article), but
// fetching and rendering one is M4.2's job, not built yet — Activate must
// not attempt loadDocument against the article's URL (which would surface
// a raw "cannot open https://...: no such file or directory" read error,
// indistinguishable from a genuinely broken file) and must not crash or
// leave the session inconsistent, mirroring
// TestPickerActivateWithAReadErrorLeavesSessionRunning's own shape for a
// missing file.
func TestPickerActivateOnArticleRowInFileModeRefusesCleanlyForNow(t *testing.T) {
	m := sized(t, newModel(), 80, 24).withRipgrep(ripgrepAvailability{available: true})
	m, _ = m.openPicker()
	m.picker.scope = starredScope
	m.picker.input.SetValue("euler")
	m.picker.results = pickerResults{
		query: "euler",
		mode:  fileMode,
		scope: starredScope,
		files: []pickerFile{{path: "https://en.wikipedia.org/wiki/Leonhard_Euler", article: true, title: "Leonhard Euler"}},
	}
	m.picker.selected = 0
	beforeDoc := m.dv1.doc.path

	got := press(t, m, keys.Activate.Keys()[0])

	if got.picker.open {
		t.Error("picker.open = true after Activate on an article row, want dismissed")
	}
	if got.dv1.doc.path != beforeDoc {
		t.Errorf("Active Document = %q after Activate on an article row, want unchanged %q", got.dv1.doc.path, beforeDoc)
	}
	if got.message == "" {
		t.Error("message is empty after Activate on an article row, want a footer message")
	}
	if strings.Contains(got.message, "no such file or directory") {
		t.Errorf("message = %q, want a clean refusal rather than a raw filesystem error", got.message)
	}
}

// TestPickerPreviewOnArticleRowInFileModeShowsAPlaceholder is C6/M3.5's own
// Preview-side guard for the same gap: moving the Results cursor onto an
// article row must not attempt to read its URL as a file (which would show
// a raw OS error string in the live Preview pane, unlike
// TestPickerWikipediaModePreviewShowsPlaceholder's own wikipediaMode
// placeholder, which this mirrors for File mode's own article rows).
func TestPickerPreviewOnArticleRowInFileModeShowsAPlaceholder(t *testing.T) {
	m := sized(t, newModel(), 80, 24).withRipgrep(ripgrepAvailability{available: true})
	m, _ = m.openPicker()
	m.picker.scope = starredScope
	m.picker.input.SetValue("euler")
	m.picker.results = pickerResults{
		query: "euler",
		mode:  fileMode,
		scope: starredScope,
		files: []pickerFile{{path: "https://en.wikipedia.org/wiki/Leonhard_Euler", article: true, title: "Leonhard Euler"}},
	}
	m.picker.selected = 0
	m.picker = m.picker.syncPreview(m.hideDotfiles)

	content := pickerPreviewContent(m.picker)
	if strings.Contains(content, "no such file or directory") {
		t.Errorf("Preview content = %q, want a clean placeholder rather than a raw filesystem error", content)
	}
	if !strings.Contains(content, "Wikipedia articles") {
		t.Errorf("Preview content = %q, want a placeholder naming Wikipedia articles rather than blank/loaded content", content)
	}
}

// TestGrepExcludesArticlesInEveryScope is M3.5's own rule 4: "Grep never
// includes Wikipedia, in any scope" — checked across all three scopes a
// Grep search can run under, each seeded so the starred article's title
// would be the only thing to match if Grep ever looked at it.
func TestGrepExcludesArticlesInEveryScope(t *testing.T) {
	requireRipgrep(t)
	for _, scope := range []pickerScope{activeRootScope, allRootsScope, starredScope} {
		m, _ := wikipediaScopeFixture(t)
		m, _ = m.openPicker()
		m = m.switchPickerModeForTest()
		for m.picker.scope != scope {
			next, _ := m.Update(keyPress(keys.PickerScope.Keys()[0]))
			m = concrete(t, next)
		}

		m, seqs := typePickerQuery(t, m, "Leonhard Euler")
		m, searches := deliverPickerTicks(t, m, seqs)
		if searches == 0 {
			t.Fatalf("scope %v: no Grep search was dispatched", scope)
		}
		if len(m.picker.results.grep) != 0 {
			t.Errorf("scope %v: Grep search for the article's own title = %+v, want no results (Grep never reaches Wikipedia)", scope, m.picker.results.grep)
		}
	}
}
