package main

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// TestPickerStarTogglesSelectedFileResult is M2.5's own acceptance
// criterion for the File Picker: starring the selected Result persists via
// the S1 store the same way toggleDocStar/toggleActiveDocStar's own
// docRecord round-trip already does, and un-starring "leaves the Result in
// place with its marker cleared" (S2) — the Results list, the selection
// cursor and the Preview are all left byte-identical, since nothing here
// touches m.picker at all besides the document store.
func TestPickerStarTogglesSelectedFileResult(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "documents")
	path := writeFileContent(t, filepath.Join(dir, "report.md"), "report")

	m := sized(t, newModel().withDocs(storePath, loadDocStore(storePath)), 80, 24).
		withRipgrep(ripgrepAvailability{available: true})
	m, _ = m.openPicker()
	m.picker.input.SetValue("report")
	m.picker.results = pickerResults{
		query: "report",
		mode:  fileMode,
		scope: activeRootScope,
		files: []pickerFile{{root: root{path: dir}, path: path}},
	}
	m.picker.selected = 0
	before := m.picker

	m = press(t, m, keys.PickerStar.Keys()[0])

	if got := m.docRecord(path); !got.Starred {
		t.Errorf("docRecord(report.md) after Picker star = %+v, want Starred", got)
	}
	// Un-starring leaves the Result row and cursor position unchanged (S2):
	// picker state besides the store itself must be byte-identical to what
	// it was before either toggle.
	m = press(t, m, keys.PickerStar.Keys()[0])
	if got := m.docRecord(path); got.Starred {
		t.Errorf("docRecord(report.md) after un-starring = %+v, want cleared", got)
	}
	if !reflect.DeepEqual(m.picker.results, before.results) {
		t.Errorf("picker.results after star/un-star round-trip = %+v, want unchanged %+v", m.picker.results, before.results)
	}
	if m.picker.selected != before.selected {
		t.Errorf("picker.selected after star/un-star round-trip = %d, want unchanged %d", m.picker.selected, before.selected)
	}

	// Persists: a fresh model loading the same store path sees the final
	// (unstarred) state.
	restarted := newModel().withDocs(storePath, loadDocStore(storePath))
	if got := restarted.docRecord(path); got.Starred {
		t.Errorf("docRecord(report.md) after restart = %+v, want unstarred", got)
	}
}

// TestPickerStarTogglesSelectedGrepResult confirms the same action reaches
// a Grep Result too (S2 names "the selected Result" without restricting it
// to File mode; only the Grep Picker's own Starred *scope*, M5.1, is the
// pre-agreed cut, not this action).
func TestPickerStarTogglesSelectedGrepResult(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "documents")
	path := writeFileContent(t, filepath.Join(dir, "notes.md"), "alpha\nbravo\n")

	m := sized(t, newModel().withDocs(storePath, loadDocStore(storePath)), 80, 24).
		withRipgrep(ripgrepAvailability{available: true})
	m, _ = m.openPicker()
	m = m.switchPickerModeForTest()
	m.picker.input.SetValue("bravo")
	m.picker.results = pickerResults{
		query: "bravo",
		mode:  grepMode,
		scope: activeRootScope,
		grep:  []grepResult{{file: pickerFile{root: root{path: dir}, path: path}, line: 2, text: "bravo"}},
	}
	m.picker.selected = 0

	m = press(t, m, keys.PickerStar.Keys()[0])

	if got := m.docRecord(path); !got.Starred {
		t.Errorf("docRecord(notes.md) after Picker star (Grep mode) = %+v, want Starred", got)
	}
}

// TestPickerStarNoOpWithNothingSelected mirrors activatePickerInto's own
// "not ok" shape: no search has run yet (zero-value results), so there is
// no Result to act on — a silent no-op, not a message or a crash.
func TestPickerStarNoOpWithNothingSelected(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "documents")

	m := sized(t, newModel().withDocs(storePath, loadDocStore(storePath)), 80, 24).
		withRipgrep(ripgrepAvailability{available: true})
	m, _ = m.openPicker()

	m = press(t, m, keys.PickerStar.Keys()[0])

	if m.message != "" {
		t.Errorf("Picker star with nothing selected: want silent no-op, got message %q", m.message)
	}
	if len(m.docs) != 0 {
		t.Errorf("docs after Picker star with nothing selected = %+v, want untouched", m.docs)
	}
}

// TestPickerStarTogglingUpdatesTheRenderedMarkerLive is D72's resolution of
// TRIAGE T8, end to end: unlike the Files Nav's own m.files.starred cache
// (M2.7/D70, which needed an explicit resync when a star changed via a
// different surface), the Picker's Results rows read the S1 store straight
// off m.docs on every render (pickerResultsContent's own docs parameter),
// so there is nothing to go stale — toggling PickerStar must show the
// marker in the very next View(), with no separate refresh step.
func TestPickerStarTogglingUpdatesTheRenderedMarkerLive(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "documents")
	path := writeFileContent(t, filepath.Join(dir, "report.md"), "report")

	m := sized(t, newModel().withDocs(storePath, loadDocStore(storePath)), 80, 24).
		withRipgrep(ripgrepAvailability{available: true})
	m, _ = m.openPicker()
	m.picker.input.SetValue("report")
	m.picker.results = pickerResults{
		query: "report",
		mode:  fileMode,
		scope: activeRootScope,
		files: []pickerFile{{root: root{path: dir}, path: path}},
	}
	m.picker.selected = 0

	resultRow := func(content string) string {
		for _, l := range strings.Split(content, "\n") {
			if strings.Contains(l, "report.md") {
				return l
			}
		}
		return ""
	}

	before := ansi.Strip(m.View().Content)
	if row := resultRow(before); row == "" || strings.Contains(row, "\u2605") {
		t.Fatalf("Results row before starring = %q, want the report.md row present with no star marker\n%s", row, before)
	}

	m = press(t, m, keys.PickerStar.Keys()[0])

	after := ansi.Strip(m.View().Content)
	if row := resultRow(after); !strings.Contains(row, "\u2605") {
		t.Errorf("Results row after starring the selected Result = %q, want it to carry the \"\u2605\" marker\n%s", row, after)
	}

	// Un-starring clears the marker in the same frame, matching S2's "marker
	// cleared" wording literally now that there is a rendered marker to clear.
	m = press(t, m, keys.PickerStar.Keys()[0])
	cleared := ansi.Strip(m.View().Content)
	if row := resultRow(cleared); strings.Contains(row, "\u2605") {
		t.Errorf("Results row after un-starring = %q, want the star marker gone\n%s", row, cleared)
	}
}

// switchPickerModeForTest flips to Grep mode without going through
// queuePickerSearch's debounce/Cmd plumbing (irrelevant here — the test
// installs m.picker.results directly), the same shortcut
// TestPickerActivateOpensResultAsActiveDocumentOnly's own fixture takes.
func (m model) switchPickerModeForTest() model {
	m.picker.mode = grepMode
	return m
}
