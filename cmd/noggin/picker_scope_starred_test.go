package main

import (
	"path/filepath"
	"testing"
)

// TestSwitchPickerScopeCyclesThroughStarredAndBack is S8/M5.1's "a
// three-way cycle" acceptance wording: the scope key now visits Active
// Source, All Roots and Starred in that fixed order before returning to
// Active Source, the same "named cycle order" shape browse.go's modeScopes
// already gives Browse Scope rather than a pairwise toggle per value.
func TestSwitchPickerScopeCyclesThroughStarredAndBack(t *testing.T) {
	m := sized(t, newModel(), 80, 24).withRipgrep(ripgrepAvailability{available: true})
	m, _ = m.openPicker()
	if m.picker.scope != activeRootScope {
		t.Fatalf("scope on open = %v, want activeRootScope", m.picker.scope)
	}

	want := []pickerScope{allRootsScope, starredScope, activeRootScope}
	for i, w := range want {
		next, _ := m.Update(keyPress(keys.PickerScope.Keys()[0]))
		m = concrete(t, next)
		if m.picker.scope != w {
			t.Fatalf("scope after press %d = %v, want %v", i+1, m.picker.scope, w)
		}
	}
}

// TestSwitchPickerModeNeverRewritesScope is S8's own explicit guarantee,
// verbatim ("available in both the File and the Grep Picker so that
// switching mode never has to silently rewrite the scope"): leaving the
// Picker in Starred scope and switching mode all the way round the C6/M3.2
// three-way File → Grep → Wikipedia → File cycle must never move it off
// Starred, the same "only p.mode changes" contract switchPickerMode's own
// doc comment already states — including in Wikipedia mode, where
// switchPickerScope no-ops (S2) but p.scope itself must still read
// unchanged once File mode is reached again.
func TestSwitchPickerModeNeverRewritesScope(t *testing.T) {
	m := sized(t, newModel(), 80, 24).withRipgrep(ripgrepAvailability{available: true})
	m, _ = m.openPicker()
	m.picker.scope = starredScope

	for i, want := range []pickerMode{grepMode, wikipediaMode, fileMode} {
		next, _ := m.Update(keyPress(keys.PickerMode.Keys()[0]))
		m = concrete(t, next)
		if m.picker.mode != want {
			t.Fatalf("mode after switch %d = %v, want %v", i+1, m.picker.mode, want)
		}
		if m.picker.scope != starredScope {
			t.Fatalf("scope after mode switch %d = %v, want unchanged starredScope", i+1, m.picker.scope)
		}
	}
}

// starredScopeFixture puts a starred, un-Rooted ("orphan") document and a
// starred, Root-covered document behind an already-open Picker in Starred
// scope, plus an unstarred document under the Root that must never surface
// regardless of query — the file-set half of S8/M5.1's own acceptance
// wording: "a Grep and a File search under Starred return matches from a
// starred document under no session Root". storePath backs withDocRecord
// (model.go's own "no resolvable data directory" refusal otherwise), the
// same setup TestPickerStarTogglesSelectedFileResult already uses.
func starredScopeFixture(t *testing.T) (m model, orphan string) {
	t.Helper()
	dir := t.TempDir()
	rootDir := mkdir(t, filepath.Join(dir, "root"))
	orphanDir := mkdir(t, filepath.Join(dir, "elsewhere"))
	orphan = writeFileContent(t, filepath.Join(orphanDir, "orphan-report.md"), "an orphan needle\n")
	rooted := writeFileContent(t, filepath.Join(rootDir, "rooted-report.md"), "a rooted needle\n")
	writeFileContent(t, filepath.Join(rootDir, "unstarred.md"), "a needle nobody starred\n")

	storePath := filepath.Join(dir, "documents")
	m = sized(t, newModel().withDocs(storePath, loadDocStore(storePath)), 80, 24).
		withRipgrep(detectRipgrep())
	m.roots = m.roots.setRoots([]root{{path: rootDir}}, 0)
	m = m.withDocRecord(orphan, docRecord{Starred: true})
	if m.message != "" {
		t.Fatalf("starring the orphan: message = %q, want empty", m.message)
	}
	m = m.withDocRecord(rooted, docRecord{Starred: true})
	if m.message != "" {
		t.Fatalf("starring the rooted document: message = %q, want empty", m.message)
	}

	m, _ = m.openPicker()
	for i := 0; i < 2; i++ { // Active Source -> All Roots -> Starred
		next, _ := m.Update(keyPress(keys.PickerScope.Keys()[0]))
		m = concrete(t, next)
	}
	if m.picker.scope != starredScope {
		t.Fatalf("setup: scope = %v, want starredScope", m.picker.scope)
	}
	return m, orphan
}

// TestFileSearchUnderStarredScopeFindsOrphanDocument is S8/M5.1's own
// acceptance wording's File-mode half: a File Picker search under Starred
// scope reaches a starred document under no session Root by name, exactly
// as it reaches one under the active Root.
func TestFileSearchUnderStarredScopeFindsOrphanDocument(t *testing.T) {
	requireRipgrep(t)
	m, orphan := starredScopeFixture(t)

	m, seqs := typePickerQuery(t, m, "orphan-report")
	m, searches := deliverPickerTicks(t, m, seqs)
	if searches == 0 {
		t.Fatal("no File search was dispatched under Starred scope")
	}
	if len(m.picker.results.files) != 1 || m.picker.results.files[0].path != orphan {
		t.Fatalf("File search under Starred scope = %v, want exactly [%s]", m.picker.results.files, orphan)
	}
	if !m.picker.results.files[0].orphan {
		t.Errorf("matched pickerFile.orphan = false, want true for a document under no session Root")
	}
}

// TestGrepSearchUnderStarredScopeFindsOrphanDocument is S8/M5.1's own
// acceptance wording's Grep-mode half: a Grep Picker search under Starred
// scope reaches a starred document's content under no session Root, the
// same guarantee the File-mode test above gives file names. This is the
// half PLAN.md/M5.1 named as a pre-agreed cut in advance of building it;
// it fell out of the same starredPickerFiles/grepPickerFiles machinery the
// File-mode half needed, so it is exercised and asserted here rather than
// left unimplemented (see PROGRESS.md's own build note on this task).
func TestGrepSearchUnderStarredScopeFindsOrphanDocument(t *testing.T) {
	requireRipgrep(t)
	m, orphan := starredScopeFixture(t)
	m = m.switchPickerModeForTest()

	m, seqs := typePickerQuery(t, m, "orphan needle")
	m, searches := deliverPickerTicks(t, m, seqs)
	if searches == 0 {
		t.Fatal("no Grep search was dispatched under Starred scope")
	}
	if len(m.picker.results.grep) != 1 || m.picker.results.grep[0].file.path != orphan {
		t.Fatalf("Grep search under Starred scope = %v, want exactly one match in %s", m.picker.results.grep, orphan)
	}
}
