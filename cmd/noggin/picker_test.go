package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"noggin/internal/tree"
)

// TestPickerOpensFromEveryFocusAndNavigatorHidden is M2.1's "reachable from
// every Focus including while the Navigator is hidden" (S1): pressing the
// Picker binding opens it regardless of which pane currently has focus, and
// still opens it when the Navigator has been toggled away (so no Navigator
// pane could possibly have focus).
func TestPickerOpensFromEveryFocusAndNavigatorHidden(t *testing.T) {
	for _, p := range []pane{rootsPane, filesPane, dv1Pane, dv2Pane} {
		m := openedFixture(t, 80, 24)
		m.focus = p

		m = press(t, m, keys.Picker.Keys()[0])

		if !m.picker.open {
			t.Errorf("focus %v: picker.open = false after the Picker key, want true", p)
		}
	}

	m := openedFixture(t, 80, 24)
	m = press(t, m, keys.Navigator.Keys()[0])
	if !m.navigatorHidden {
		t.Fatal("Navigator toggle did not hide it; setup is broken")
	}

	m = press(t, m, keys.Picker.Keys()[0])
	if !m.picker.open {
		t.Error("navigator hidden: picker.open = false after the Picker key, want true")
	}
}

// TestPickerAlwaysOpensAsFileActiveRootEmptyPrompt is S1's fixed opening
// state, checked from a picker that (if memory existed) would say
// otherwise: Grep mode, All Roots Scope, a typed Prompt. Opening always
// resets to File Picker / Active Source / empty Prompt regardless, for the
// (here, default) case where a Root, not Wikipedia, is the Active Source —
// S2/C6/M3.2's Wikipedia-active exception to "always File Picker" has its
// own tests (TestOpenPickerLandsOnWikipediaModeWithWikipediaActive,
// picker_wikipedia_mode_test.go).
// Scope stays activeRootScope regardless of which mode is landed on: it is
// simply unread and unrendered while Wikipedia mode is active (S2).
func TestPickerAlwaysOpensAsFileActiveRootEmptyPrompt(t *testing.T) {
	m := openedFixture(t, 80, 24)
	m.picker = picker{open: false, mode: grepMode, scope: allRootsScope, input: newPicker().input}
	m.picker.input.SetValue("stale query")

	next, _ := m.openPicker()

	if next.picker.mode != fileMode {
		t.Errorf("mode = %v, want fileMode", next.picker.mode)
	}
	if next.picker.scope != activeRootScope {
		t.Errorf("scope = %v, want activeRootScope", next.picker.scope)
	}
	if next.picker.input.Value() != "" {
		t.Errorf("prompt = %q, want empty", next.picker.input.Value())
	}
	if !next.picker.open {
		t.Error("open = false, want true")
	}
}

// TestPickerDismissRestoresThePreOpenFrameByteForByte is M2.1's own
// acceptance criterion: opening the Picker, changing what it would show
// (mode, Scope, Prompt) and dismissing it leaves the session exactly as it
// was found — the same Active Document, Active Source and scroll position,
// and (since M2.1 predates any Picker rendering, M2.2 onward's job) the
// exact same rendered frame.
func TestPickerDismissRestoresThePreOpenFrameByteForByte(t *testing.T) {
	m := openedFixture(t, 80, 24)
	m = press(t, m, "down") // move the viewer's scroll position off zero
	beforeFrame := m.View().Content
	beforeDoc := m.dv1.doc.path
	beforeRoot := m.roots.active
	beforeOffset := m.dv1.offset()

	m, _ = m.openPicker()
	if !m.picker.open {
		t.Fatal("openPicker did not open the picker; setup is broken")
	}
	// Perturb the Picker's own state (mode/Scope/Prompt) so dismissal is
	// proven to discard it, not merely leave it alone because nothing
	// changed.
	m.picker.mode = grepMode
	m.picker.scope = allRootsScope
	m.picker.input.SetValue("something typed")

	m = press(t, m, keys.Cancel.Keys()[0])

	if m.picker.open {
		t.Fatal("picker.open = true after Cancel, want false")
	}
	if got := m.View().Content; got != beforeFrame {
		t.Errorf("frame after dismissal does not match the pre-open frame")
	}
	if m.dv1.doc.path != beforeDoc {
		t.Errorf("active document = %q, want %q", m.dv1.doc.path, beforeDoc)
	}
	if m.roots.active != beforeRoot {
		t.Errorf("active root = %d, want %d", m.roots.active, beforeRoot)
	}
	if m.dv1.offset() != beforeOffset {
		t.Errorf("scroll offset = %d, want %d", m.dv1.offset(), beforeOffset)
	}
	// A reopen must not resurrect the discarded mode/Scope/Prompt (S8).
	m, _ = m.openPicker()
	if m.picker.mode != fileMode || m.picker.scope != activeRootScope || m.picker.input.Value() != "" {
		t.Errorf("reopened picker = %+v, want File/ActiveRoot/empty", m.picker)
	}
}

// TestPickerReopenAfterDismissKeepsPreviewSize is M3.2a's own acceptance
// criterion (D48/TRIAGE T6): dismissPicker must carry the Preview
// viewport's Width/Height over its reset the same way openPicker already
// does, so a dismiss immediately followed by a reopen — with no
// tea.WindowSizeMsg (and so no model.resize call) in between, the ordinary
// case since dismissing the Picker never resizes the terminal — does not
// leave the reopened Preview at newPicker()'s fresh 0x0 zero value.
func TestPickerReopenAfterDismissKeepsPreviewSize(t *testing.T) {
	m := openedFixture(t, 80, 24)
	m, _ = m.openPicker()
	wantWidth, wantHeight := m.picker.preview.Width(), m.picker.preview.Height()
	if wantWidth == 0 || wantHeight == 0 {
		t.Fatalf("setup is broken: preview size is %dx%d before dismissal, want nonzero", wantWidth, wantHeight)
	}

	m = press(t, m, keys.Cancel.Keys()[0])
	if m.picker.open {
		t.Fatal("Cancel did not dismiss the picker; setup is broken")
	}

	m, _ = m.openPicker()

	if got := m.picker.preview.Width(); got != wantWidth {
		t.Errorf("preview width after reopen (no resize in between) = %d, want %d unchanged", got, wantWidth)
	}
	if got := m.picker.preview.Height(); got != wantHeight {
		t.Errorf("preview height after reopen (no resize in between) = %d, want %d unchanged", got, wantHeight)
	}
}

// TestPickerOwnsKeyboardWhileOpen is the same exclusivity handleFindInput
// and handleHelpInput already give their own modals (model.go): while the
// Picker is open, a key bound elsewhere (here, Quit's "q") must not act on
// its usual binding.
func TestPickerOwnsKeyboardWhileOpen(t *testing.T) {
	m := openedFixture(t, 80, 24)
	m, _ = m.openPicker()

	next, _ := m.Update(keyPress("q"))
	got := concrete(t, next)
	if !got.picker.open {
		t.Error("picker.open = false after \"q\" while open, want true (q must not quit)")
	}
}

// TestPickerTypesEveryPrintableKeyIncludingBindingsElsewhere is M2.3's own
// acceptance criterion, verbatim: a string containing characters that are
// bindings elsewhere ("j"/"k" move selection on every Nav pane, "g" is
// Top, "/" starts Find) must land entirely in the Prompt, with none of
// those side effects firing — the same "always typing" guarantee
// updateFindQuery already gives Find, now given to the Picker (S2).
func TestPickerTypesEveryPrintableKeyIncludingBindingsElsewhere(t *testing.T) {
	m := openedFixture(t, 80, 24)
	m, _ = m.openPicker()
	beforeDoc := m.dv1.doc.path
	beforeOffset := m.dv1.offset()

	const query = "jog/report"
	m = typeText(t, m, query)

	if got := m.picker.input.Value(); got != query {
		t.Errorf("prompt = %q, want %q", got, query)
	}
	if !m.picker.open {
		t.Error("picker.open = false after typing, want true")
	}
	if m.dv1.doc.path != beforeDoc {
		t.Errorf("active document changed to %q while typing, want unchanged (%q)", m.dv1.doc.path, beforeDoc)
	}
	if m.dv1.offset() != beforeOffset {
		t.Errorf("viewer scroll offset changed to %d while typing, want unchanged (%d)", m.dv1.offset(), beforeOffset)
	}
	if m.dv1.findEditing() {
		t.Error("\"/\" in the query started a Find, want it to only type")
	}
	if m.helpOpen {
		t.Error("typing opened the Help Modal, want it to stay closed")
	}
}

// TestPickerNonTypingKeysDoNotType is S2's other half: Cancel, Activate,
// ActivateDV2 (C3/M1.5), PickerUp/PickerDown, the Preview scroll keys and
// PickerMode/PickerScope are all non-printable (D2's "always typing" only
// ever binds printable keystrokes), so none of them should ever reach the
// Prompt — checked here for every one but Cancel, which this task's earlier
// acceptance criterion (M2.1's dismiss-on-Cancel) already covers by making
// the Picker close instead.
func TestPickerNonTypingKeysDoNotType(t *testing.T) {
	nonTyping := []key.Binding{
		keys.Activate, keys.ActivateDV2, keys.PickerUp, keys.PickerDown,
		keys.PageUp, keys.PageDown, keys.HalfPageUp, keys.HalfPageDown,
		keys.PickerMode, keys.PickerScope,
	}
	for _, b := range nonTyping {
		for _, k := range b.Keys() {
			m := openedFixture(t, 80, 24)
			m, _ = m.openPicker()

			m = press(t, m, k)

			if !m.picker.open {
				t.Errorf("key %q closed the picker, want it to stay open", k)
				continue
			}
			if got := m.picker.input.Value(); got != "" {
				t.Errorf("key %q typed %q into the prompt, want it to type nothing", k, got)
			}
		}
	}
}

// TestPickerCancelDoesNotType is Cancel's own version of the check above,
// kept separate since Cancel's effect is dismissal (m.picker.open becomes
// false) rather than "stays open, nothing typed".
func TestPickerCancelDoesNotType(t *testing.T) {
	m := openedFixture(t, 80, 24)
	m, _ = m.openPicker()

	m = press(t, m, keys.Cancel.Keys()[0])

	if m.picker.open {
		t.Error("picker.open = true after Cancel, want false")
	}
}

// TestPickerUpDownMoveTheResultsSelection is M3.1's own wiring of
// PickerUp/PickerDown (reserved as non-typing, no-op keys by M2.3): once
// there are Results to move over, they shift picker.selected exactly like
// filesNav.move/rootsNav.move already shift their own cursors, clamped to
// the result count rather than wrapping or going out of bounds.
func TestPickerUpDownMoveTheResultsSelection(t *testing.T) {
	requireRipgrep(t)
	m, _ := pickerRoot(t)

	m, seqs := typePickerQuery(t, m, "repor")
	m, searches := deliverPickerTicks(t, m, seqs)
	if searches != 1 || len(m.picker.results.files) < 2 {
		t.Fatalf("setup is broken: %d searches, %d results", searches, len(m.picker.results.files))
	}
	if m.picker.selected != 0 {
		t.Fatalf("selected = %d after results landed, want 0", m.picker.selected)
	}

	m = press(t, m, keys.PickerDown.Keys()[0])
	if m.picker.selected != 1 {
		t.Errorf("selected = %d after PickerDown, want 1", m.picker.selected)
	}

	// Clamped at the bottom: one more PickerDown than there are rows below
	// must not run selected past the last result.
	for range m.picker.results.files {
		m = press(t, m, keys.PickerDown.Keys()[0])
	}
	if want := len(m.picker.results.files) - 1; m.picker.selected != want {
		t.Errorf("selected = %d after over-shooting PickerDown, want clamped to %d", m.picker.selected, want)
	}

	m = press(t, m, keys.PickerUp.Keys()[0])
	if want := len(m.picker.results.files) - 2; m.picker.selected != want {
		t.Errorf("selected = %d after PickerUp, want %d", m.picker.selected, want)
	}

	// Neither key types into the Prompt (S2's always-typing set already
	// covers PickerUp/PickerDown as non-printable; this pins that a working
	// action still does not type).
	if got := m.picker.input.Value(); got != "repor" {
		t.Errorf("prompt = %q after PickerUp/PickerDown, want unchanged %q", got, "repor")
	}
}

// TestPickerSelectionResetsOnFreshResults is the reset half of M3.1's
// selection cursor: an index into a candidate list that no longer exists
// (a new query landed a wholly different result set) must not be left
// pointing past the new list's end, or at some mid-list row that means
// nothing for the new query — it goes back to the top.
func TestPickerSelectionResetsOnFreshResults(t *testing.T) {
	requireRipgrep(t)
	m, _ := pickerRoot(t)

	m, seqs := typePickerQuery(t, m, "repor")
	m, _ = deliverPickerTicks(t, m, seqs)
	if len(m.picker.results.files) < 2 {
		t.Fatalf("setup is broken: %d results", len(m.picker.results.files))
	}
	m = press(t, m, keys.PickerDown.Keys()[0])
	if m.picker.selected == 0 {
		t.Fatal("PickerDown did not move the selection; setup is broken")
	}

	m, seqs = typePickerQuery(t, m, "t") // "report" now, a fresh generation
	m, _ = deliverPickerTicks(t, m, seqs)

	if m.picker.selected != 0 {
		t.Errorf("selected = %d after fresh results landed, want reset to 0", m.picker.selected)
	}
}

// TestPickerActivateOpensResultAsActiveDocumentOnly is M3.3's own
// acceptance criterion, verbatim: activating a Result belonging to a
// non-active Root makes it the Active Document while the Active Root, the
// Files Nav tree's expansion and its selection are all unchanged — S10's
// "does not change the Active Root, and does not move, expand or re-select
// the Files Nav tree", checked against a Result whose Root is deliberately
// not the active one.
func TestPickerActivateOpensResultAsActiveDocumentOnly(t *testing.T) {
	rootA := t.TempDir()
	rootB := t.TempDir()
	writeFileContent(t, filepath.Join(rootA, "existing.md"), "root A doc\n")
	writeFileContent(t, filepath.Join(rootA, "other.md"), "root A other\n")
	subDir := filepath.Join(rootA, "sub")
	if err := os.Mkdir(subDir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", subDir, err)
	}
	writeFileContent(t, filepath.Join(subDir, "child.md"), "root A sub child\n")
	fileB := filepath.Join(rootB, "report.md")
	writeFileContent(t, fileB, "root B report\n")

	nodeA, err := tree.Walk(rootA, tree.Options{Filter: tree.FilterMarkdown})
	if err != nil {
		t.Fatalf("walking %s: %v", rootA, err)
	}

	m := openedFixture(t, 80, 24).withRipgrep(ripgrepAvailability{available: true})
	m.roots = m.roots.setRoots([]root{{path: rootA}, {path: rootB}}, 0)
	m = m.openFolder(nodeA)
	// Expand "sub" (TRIAGE T4/D56): without at least one expanded
	// subdirectory, m.files.expanded is empty both before and after
	// Activate, so the reflect.DeepEqual check below would pass even if
	// Activate started clearing or replacing expansion state — an empty
	// map equals an empty map regardless of what produced it.
	m.files = m.files.withExpanded(subDir, true)
	m.files = m.files.move(1) // off the tree's initial selection
	wantSelected := m.files.selected
	// A cloned copy, not just m.files.expanded itself (still less its
	// count): the whole key set must match afterwards (M3.3b/D49/TRIAGE
	// T9), not merely its cardinality — a same-count substitution (e.g.
	// Activate collapsing rootA's node but expanding a different one) would
	// pass a count-only check.
	wantExpanded := make(map[string]bool, len(m.files.expanded))
	for k, v := range m.files.expanded {
		wantExpanded[k] = v
	}
	if len(wantExpanded) == 0 {
		t.Fatal("setup is broken: expanded set is empty")
	}

	m, _ = m.openPicker()
	m.picker.input.SetValue("report")
	m.picker.results = pickerResults{
		query: "report",
		mode:  fileMode,
		scope: activeRootScope,
		files: []pickerFile{{root: root{path: rootB}, path: fileB}},
	}
	m.picker.selected = 0

	got := press(t, m, keys.Activate.Keys()[0])

	if got.picker.open {
		t.Error("picker.open = true after Activate, want dismissed")
	}
	if got.dv1.doc.path != fileB {
		t.Errorf("Active Document = %q, want %q", got.dv1.doc.path, fileB)
	}
	if got.roots.active != 0 {
		t.Errorf("Active Root index = %d, want unchanged 0 (rootA)", got.roots.active)
	}
	if got.files.selected != wantSelected {
		t.Errorf("Files Nav selection = %d, want unchanged %d", got.files.selected, wantSelected)
	}
	if !reflect.DeepEqual(got.files.expanded, wantExpanded) {
		t.Errorf("Files Nav expansion = %v, want unchanged %v", got.files.expanded, wantExpanded)
	}
	if got.files.root.Path != rootA {
		t.Errorf("Files Nav root = %q, want unchanged %q", got.files.root.Path, rootA)
	}
}

// TestPickerActivateWithAReadErrorLeavesSessionRunning is M3.3a's
// read-error case (TRIAGE T7), mirroring
// TestActivateWithAReadErrorLeavesSessionRunning's shape (files_test.go)
// for the Picker's own Activate path: a Result whose file can no longer be
// read (removed between listing and Activate, or any other read failure)
// must not open it, crash, or leave the session inconsistent — the failure
// surfaces as a footer message (T15, activatePicker's own comment), and the
// Picker still closes since it is "dismissed either way".
func TestPickerActivateWithAReadErrorLeavesSessionRunning(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.md")

	m := openedFixture(t, 80, 24).withRipgrep(ripgrepAvailability{available: true})
	m, _ = m.openPicker()
	m.picker.input.SetValue("report")
	m.picker.results = pickerResults{
		query: "report",
		mode:  fileMode,
		scope: activeRootScope,
		files: []pickerFile{{path: missing}},
	}
	m.picker.selected = 0
	beforeDoc := m.dv1.doc.path

	got := press(t, m, keys.Activate.Keys()[0])

	if got.picker.open {
		t.Error("picker.open = true after a failed Activate, want dismissed")
	}
	if got.dv1.doc.path != beforeDoc {
		t.Errorf("Active Document = %q after a failed Activate, want unchanged %q", got.dv1.doc.path, beforeDoc)
	}
	if got.message == "" {
		t.Error("message is empty after a failed Activate, want a footer message")
	}
}

// TestPickerActivateGrepResultOpensDocumentScrolledToMatchedLine is M4.3's
// own acceptance criterion (S10, S11): activating a Grep search whose match
// sits deep in a long file makes that file the Active Document, with the
// Document Viewer scrolled so the matched line is on screen and indicated
// — the rendered-content re-derivation D34 calls for, not rg's raw-file
// line number directly (which the fixture's own reflow, mirroring
// TestPickerGrepPreviewOpensScrolledToMatchedLineWithHighlight's setup, is
// exactly why the assertion below checks by content/offset rather than
// pinning the exact rendered line index the way that Preview-side test can
// for raw text). This single-match fixture alone does not isolate D34's
// ordinal-resolution bridge (resolveGrepLanding) from a plain "land
// somewhere near the match" bridge: with only one occurrence, disabling
// resolveGrepLanding's ordinal lookup would still happen to land in the
// right place (M7.1b/TRIAGE T1). Only paired with the sibling test below,
// TestPickerActivateGrepResultLandsOnTheSelectedOccurrenceNotJustTheFirst
// (two occurrences, activating the second), is the criterion actually
// covered end to end.
func TestPickerActivateGrepResultOpensDocumentScrolledToMatchedLine(t *testing.T) {
	requireRipgrep(t)
	dir := t.TempDir()
	const matchLine = 87
	path := numberedLinesFixture(t, dir, "long.md", 100, matchLine, "the target report line")

	m := openedFixture(t, 80, 24).withRipgrep(detectRipgrep())
	m.roots = m.roots.setRoots([]root{{path: dir}}, 0)
	m, _ = m.openPicker()
	m.picker.mode = grepMode

	m, seqs := typePickerQuery(t, m, "report")
	m, searches := deliverPickerTicks(t, m, seqs)
	if searches != 1 || len(m.picker.results.grep) != 1 {
		t.Fatalf("setup is broken: %d searches, %d grep results", searches, len(m.picker.results.grep))
	}

	got := press(t, m, keys.Activate.Keys()[0])

	if got.picker.open {
		t.Error("picker.open = true after Activate, want dismissed")
	}
	if got.dv1.doc.path != path {
		t.Errorf("Active Document = %q, want %q", got.dv1.doc.path, path)
	}
	if !got.dv1.hasFindMatches() {
		t.Fatal("Activate did not leave the Document Viewer with find matches for the grep query")
	}
	if got.dv1.offset() == 0 {
		t.Error("Document Viewer did not scroll away from the top to reach the matched line")
	}
	view := got.View().Content
	if !strings.Contains(view, "target report line") {
		t.Errorf("view does not show the matched line's text after activating the Grep Result\n%s", view)
	}
	// hasReverse reads cell styles, not a literal "\x1b[7m" (see the same
	// assertion in TestFindLocatesAPhraseNearTheEndOfTheDocument).
	if !hasReverse(view) {
		t.Error("matched line is not indicated (no reverse-video escape in the view)")
	}
}

// TestPickerActivateGrepResultDoesNotChangeFilesNav is M7.1b's grep-mode
// counterpart to TestPickerActivateOpensResultAsActiveDocumentOnly
// (M3.3, M7.1b/TRIAGE T2): S10 ("does not change the Active Root, and does
// not move, expand or re-select the Files Nav tree") applies to Activate
// regardless of Picker mode, since both paths end at the same
// model.openDocument (which never touches m.files or m.roots) — but only
// the file-mode path had a direct assertion for it. Activating a Grep
// Result belonging to a non-active Root must leave the Active Root, the
// Files Nav tree's expansion and its selection exactly as they were.
func TestPickerActivateGrepResultDoesNotChangeFilesNav(t *testing.T) {
	rootA := t.TempDir()
	rootB := t.TempDir()
	writeFileContent(t, filepath.Join(rootA, "existing.md"), "root A doc\n")
	subDir := filepath.Join(rootA, "sub")
	if err := os.Mkdir(subDir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", subDir, err)
	}
	writeFileContent(t, filepath.Join(subDir, "child.md"), "root A sub child\n")
	fileB := filepath.Join(rootB, "report.md")
	writeFileContent(t, fileB, "a report line\n")

	nodeA, err := tree.Walk(rootA, tree.Options{Filter: tree.FilterMarkdown})
	if err != nil {
		t.Fatalf("walking %s: %v", rootA, err)
	}

	m := openedFixture(t, 80, 24).withRipgrep(ripgrepAvailability{available: true})
	m.roots = m.roots.setRoots([]root{{path: rootA}, {path: rootB}}, 0)
	m = m.openFolder(nodeA)
	m.files = m.files.withExpanded(subDir, true)
	m.files = m.files.move(1) // off the tree's initial selection
	wantSelected := m.files.selected
	wantExpanded := make(map[string]bool, len(m.files.expanded))
	for k, v := range m.files.expanded {
		wantExpanded[k] = v
	}
	if len(wantExpanded) == 0 {
		t.Fatal("setup is broken: expanded set is empty")
	}

	m, _ = m.openPicker()
	m.picker.mode = grepMode
	m.picker.input.SetValue("report")
	m.picker.results = pickerResults{
		query: "report",
		mode:  grepMode,
		scope: activeRootScope,
		grep:  []grepResult{{file: pickerFile{root: root{path: rootB}, path: fileB}, line: 1, text: "a report line"}},
	}
	m.picker.selected = 0

	got := press(t, m, keys.Activate.Keys()[0])

	if got.picker.open {
		t.Error("picker.open = true after Activate, want dismissed")
	}
	if got.dv1.doc.path != fileB {
		t.Errorf("Active Document = %q, want %q", got.dv1.doc.path, fileB)
	}
	if got.roots.active != 0 {
		t.Errorf("Active Root index = %d, want unchanged 0 (rootA)", got.roots.active)
	}
	if got.files.selected != wantSelected {
		t.Errorf("Files Nav selection = %d, want unchanged %d", got.files.selected, wantSelected)
	}
	if !reflect.DeepEqual(got.files.expanded, wantExpanded) {
		t.Errorf("Files Nav expansion = %v, want unchanged %v", got.files.expanded, wantExpanded)
	}
	if got.files.root.Path != rootA {
		t.Errorf("Files Nav root = %q, want unchanged %q", got.files.root.Path, rootA)
	}
}

// TestPickerActivateGrepResultWithAReadErrorLeavesSessionRunning is M4.3's
// read-error case, mirroring TestPickerActivateWithAReadErrorLeavesSessionRunning's
// fileMode shape (T15, M3.3a/TRIAGE T7): a Grep Result whose file can no
// longer be read must not open it, crash, or leave the session
// inconsistent — the failure surfaces as a footer message, and the Picker
// still closes.
func TestPickerActivateGrepResultWithAReadErrorLeavesSessionRunning(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.md")

	m := openedFixture(t, 80, 24).withRipgrep(ripgrepAvailability{available: true})
	m, _ = m.openPicker()
	m.picker.mode = grepMode
	m.picker.input.SetValue("report")
	m.picker.results = pickerResults{
		query: "report",
		mode:  grepMode,
		scope: activeRootScope,
		grep:  []grepResult{{file: pickerFile{path: missing}, line: 1, text: "report"}},
	}
	m.picker.selected = 0
	beforeDoc := m.dv1.doc.path

	got := press(t, m, keys.Activate.Keys()[0])

	if got.picker.open {
		t.Error("picker.open = true after a failed Activate, want dismissed")
	}
	if got.dv1.doc.path != beforeDoc {
		t.Errorf("Active Document = %q after a failed Activate, want unchanged %q", got.dv1.doc.path, beforeDoc)
	}
	if got.message == "" {
		t.Error("message is empty after a failed Activate, want a footer message")
	}
}

// TestPickerActivateGrepResultLandsOnTheSelectedOccurrenceNotJustTheFirst
// is the sibling half of
// TestPickerActivateGrepResultOpensDocumentScrolledToMatchedLine's pair
// (M7.1b/TRIAGE T1, that test's own doc comment) and guards against
// resolveGrepLanding's ordinal lookup silently degrading to
// the same "first match" landing a broken lookup would also produce
// (code-reviewer finding, C2): with two grep matches in the same file, far
// enough apart in the *rendered* document that they cannot both be on
// screen at once (200+ rendered lines from blank-line-separated entries,
// unlike numberedLinesFixture's single reflowed paragraph), activating the
// *second* Result must land on the second occurrence specifically, not
// silently fall back to the first the way an always-degrade bug would.
func TestPickerActivateGrepResultLandsOnTheSelectedOccurrenceNotJustTheFirst(t *testing.T) {
	requireRipgrep(t)
	dir := t.TempDir()
	var b strings.Builder
	for i := 1; i <= 100; i++ {
		switch i {
		case 10:
			b.WriteString("first report line\n\n")
		case 90:
			b.WriteString("second report line\n\n")
		default:
			fmt.Fprintf(&b, "filler line %d\n\n", i)
		}
	}
	path := writeFileContent(t, filepath.Join(dir, "long.md"), b.String())

	m := openedFixture(t, 80, 24).withRipgrep(detectRipgrep())
	m.roots = m.roots.setRoots([]root{{path: dir}}, 0)
	m, _ = m.openPicker()
	m.picker.mode = grepMode

	m, seqs := typePickerQuery(t, m, "report")
	m, searches := deliverPickerTicks(t, m, seqs)
	if searches != 1 || len(m.picker.results.grep) != 2 {
		t.Fatalf("setup is broken: %d searches, %d grep results", searches, len(m.picker.results.grep))
	}
	// grepBatch reports matches in the file's own line order, so index 1 is
	// the later (second) occurrence.
	if m.picker.results.grep[1].line <= m.picker.results.grep[0].line {
		t.Fatalf("setup is broken: grep results are not in ascending line order: %+v", m.picker.results.grep)
	}
	m.picker.selected = 1

	got := press(t, m, keys.Activate.Keys()[0])

	if got.dv1.doc.path != path {
		t.Fatalf("Active Document = %q, want %q", got.dv1.doc.path, path)
	}
	view := got.View().Content
	if strings.Contains(view, "first report line") {
		t.Errorf("view shows the first occurrence's text after activating the *second* Result, want the ordinal to distinguish them\n%s", view)
	}
	if !strings.Contains(view, "second report line") {
		t.Errorf("view does not show the second occurrence's text after activating it\n%s", view)
	}
	if !hasReverse(view) {
		t.Error("matched line is not indicated (no reverse-video escape in the view)")
	}
}

// TestPickerActivateGrepResultInsideStrippedHTMLCommentDegradesToPlainOpen
// is resolveGrepLanding's *other* explicit degrade trigger (D34's headline
// motivating example, "the query spanned syntax Glamour removed"): an HTML
// comment is dropped from Glamour's render entirely, so a grep match whose
// only occurrence sits inside one has no counterpart anywhere in the
// rendered content (ordinal >= len(rend), not the frontmatter-offset
// branch the sibling test below exercises) — and, with no other occurrence
// in the body either, commitQuery's own nearestMatch(matches, 0) has
// nothing to fall back to, exercising D34's full "if there are none, open
// it plain at the top" chain end to end.
func TestPickerActivateGrepResultInsideStrippedHTMLCommentDegradesToPlainOpen(t *testing.T) {
	requireRipgrep(t)
	dir := t.TempDir()
	src := "<!-- reporthidden -->\n\n# doc\n\nnothing else matches this query\n"
	path := writeFileContent(t, filepath.Join(dir, "hidden.md"), src)

	m := openedFixture(t, 80, 24).withRipgrep(detectRipgrep())
	m.roots = m.roots.setRoots([]root{{path: dir}}, 0)
	m, _ = m.openPicker()
	m.picker.mode = grepMode

	m, seqs := typePickerQuery(t, m, "reporthidden")
	m, searches := deliverPickerTicks(t, m, seqs)
	if searches != 1 || len(m.picker.results.grep) != 1 {
		t.Fatalf("setup is broken: %d searches, %d grep results", searches, len(m.picker.results.grep))
	}

	got := press(t, m, keys.Activate.Keys()[0])

	if got.picker.open {
		t.Error("picker.open = true after Activate, want dismissed")
	}
	if got.dv1.doc.path != path {
		t.Errorf("Active Document = %q, want %q", got.dv1.doc.path, path)
	}
	if got.dv1.hasFindMatches() {
		t.Error("degrade path found a find match, want none: the query's only raw occurrence is inside an HTML comment Glamour strips entirely")
	}
	if got.dv1.offset() != 0 {
		t.Errorf("Document Viewer offset = %d, want 0 (opened plain at the top per D34's full degrade)", got.dv1.offset())
	}
	// Scoped to the viewer pane's own rendering, not the whole frame (S5's
	// two-line Sources Nav rows use a bare Reverse(true) on an active row's
	// unbold path line, which would otherwise make this substring check a
	// false positive whenever a root is both active and selected — the
	// common single-root case, unrelated to whether the viewer has a find
	// match to indicate).
	view := got.dv1.view()
	if hasReverse(view) {
		t.Error("no line should be indicated when the query has no rendered matches at all")
	}
}

// TestPickerActivateGrepResultInsideFrontmatterLandsOnIt: frontmatter is
// shown (D138), so a Grep Result on a frontmatter line has a rendered
// counterpart in the Frontmatter Card and lands on it, not on the body's
// later occurrence of the same query. (Under D5's stripping this was a
// degrade case.)
func TestPickerActivateGrepResultInsideFrontmatterLandsOnIt(t *testing.T) {
	requireRipgrep(t)
	dir := t.TempDir()
	src := "---\ntitle: report\n---\n# doc\n\nfiller\n\nthe body has a report too\n"
	path := writeFileContent(t, filepath.Join(dir, "fm.md"), src)

	m := openedFixture(t, 80, 24).withRipgrep(detectRipgrep())
	m.roots = m.roots.setRoots([]root{{path: dir}}, 0)
	m, _ = m.openPicker()
	m.picker.mode = grepMode

	m, seqs := typePickerQuery(t, m, "report")
	m, searches := deliverPickerTicks(t, m, seqs)
	if searches != 1 || len(m.picker.results.grep) != 2 {
		t.Fatalf("setup is broken: %d searches, %d grep results (want 2: frontmatter + body)", searches, len(m.picker.results.grep))
	}
	idx := -1
	for i, g := range m.picker.results.grep {
		if g.line == 2 { // the "title: report" frontmatter line
			idx = i
		}
	}
	if idx < 0 {
		t.Fatalf("setup is broken: no grep result at line 2 (the frontmatter line); results = %+v", m.picker.results.grep)
	}
	m.picker.selected = idx

	got := press(t, m, keys.Activate.Keys()[0])

	if got.dv1.doc.path != path {
		t.Errorf("Active Document = %q, want %q", got.dv1.doc.path, path)
	}
	if n := len(got.dv1.findState.matches); n != 2 {
		t.Fatalf("find matches = %d, want 2 (frontmatter + body)", n)
	}
	if got.dv1.findState.current != 0 {
		t.Errorf("current match = %d, want 0 (the frontmatter occurrence)", got.dv1.findState.current)
	}
	view := got.dv1.view()
	if !strings.Contains(view, "report") || !strings.Contains(view, "╭") {
		t.Error("frontmatter card is not shown in the rendered view (D138)")
	}
	if !hasReverse(view) {
		t.Error("no line indicated, want the frontmatter match highlighted")
	}
}

// TestPickerActivateGrepResultCommitsFindWithNextPreviousTraversal is
// M4.4's headline acceptance check: activateGrepPicker's hand-off to
// commitQuery (M4.3) is not just a one-shot landing but a real committed
// find — the same state a typed "/" search followed by Enter leaves
// (commitFind's doc comment) — so FindNext/FindPrevious (keys.FindNext,
// keys.FindPrevious; see TestFindTraversalWraps for the equivalent typed-
// search check) walk the document's *other* occurrences of the grep query
// after activation, not just the one it landed on.
func TestPickerActivateGrepResultCommitsFindWithNextPreviousTraversal(t *testing.T) {
	requireRipgrep(t)
	dir := t.TempDir()
	var b strings.Builder
	for i := 1; i <= 100; i++ {
		switch i {
		case 10, 50, 90:
			b.WriteString("a report line\n\n")
		default:
			fmt.Fprintf(&b, "filler line %d\n\n", i)
		}
	}
	path := writeFileContent(t, filepath.Join(dir, "long.md"), b.String())

	m := openedFixture(t, 80, 24).withRipgrep(detectRipgrep())
	m.roots = m.roots.setRoots([]root{{path: dir}}, 0)
	m, _ = m.openPicker()
	m.picker.mode = grepMode

	m, seqs := typePickerQuery(t, m, "report")
	m, searches := deliverPickerTicks(t, m, seqs)
	if searches != 1 || len(m.picker.results.grep) != 3 {
		t.Fatalf("setup is broken: %d searches, %d grep results (want 3)", searches, len(m.picker.results.grep))
	}
	m.picker.selected = 0

	got := press(t, m, keys.Activate.Keys()[0])

	if got.dv1.doc.path != path {
		t.Fatalf("Active Document = %q, want %q", got.dv1.doc.path, path)
	}
	if got := got.dv1.findState.input.Value(); got != "report" {
		t.Fatalf("committed find query = %q, want the grep query %q", got, "report")
	}
	n := len(got.dv1.findState.matches)
	if n != 3 {
		t.Fatalf("committed find has %d match(es), want 3 (one per grep result)", n)
	}
	start := got.dv1.findState.current
	if start < 0 {
		t.Fatal("committed find has no current match indicated after activation")
	}

	next := press(t, got, keys.FindNext.Keys()[0])
	if want := (start + 1) % n; next.dv1.findState.current != want {
		t.Errorf("FindNext after activation: current = %d, want %d", next.dv1.findState.current, want)
	}
	if next.picker.open {
		t.Error("FindNext after activation re-opened the Picker; want it to stay dismissed")
	}

	prev := press(t, next, keys.FindPrevious.Keys()[0])
	if prev.dv1.findState.current != start {
		t.Errorf("FindPrevious after a FindNext: current = %d, want back at %d", prev.dv1.findState.current, start)
	}
}

// TestPickerPreviewShowsSelectedFileRawFromTop is M3.2's own acceptance
// criterion, as D140 amends it: the Preview pane's content is the selected
// Result's raw file bytes — no Glamour styling — laid out by the Text
// renderer, and starts scrolled to the top.
func TestPickerPreviewShowsSelectedFileRawFromTop(t *testing.T) {
	requireRipgrep(t)
	m, _ := pickerRoot(t)

	m, seqs := typePickerQuery(t, m, "repor")
	m, searches := deliverPickerTicks(t, m, seqs)
	if searches != 1 || len(m.picker.results.files) == 0 {
		t.Fatalf("setup is broken: %d searches, %d results", searches, len(m.picker.results.files))
	}

	selected := m.picker.results.files[m.picker.selected]
	want, err := os.ReadFile(selected.path)
	if err != nil {
		t.Fatalf("reading %s: %v", selected.path, err)
	}

	wantContent := renderText(want, m.picker.preview.Width(), false)
	if got := m.picker.preview.GetContent(); got != wantContent {
		t.Errorf("preview content = %q, want the Text render of %s: %q", got, filepath.Base(selected.path), wantContent)
	}
	if !m.picker.preview.AtTop() {
		t.Error("preview did not start at the top")
	}
}

// TestPickerPreviewRetargetsOnSelectionChange is the "following selection
// changes" half of M3.2: moving the Results cursor loads the newly-selected
// Result's own bytes and scrolls back to the top, rather than leaving the
// previous selection's content (and scroll position) on screen.
func TestPickerPreviewRetargetsOnSelectionChange(t *testing.T) {
	requireRipgrep(t)
	m, _ := pickerRoot(t)

	m, seqs := typePickerQuery(t, m, "repor")
	m, _ = deliverPickerTicks(t, m, seqs)
	if len(m.picker.results.files) < 2 {
		t.Fatalf("setup is broken: %d results", len(m.picker.results.files))
	}
	firstContent := m.picker.preview.GetContent()
	firstPath := m.picker.results.files[m.picker.selected].path

	m = press(t, m, keys.PickerDown.Keys()[0])

	secondPath := m.picker.results.files[m.picker.selected].path
	if secondPath == firstPath {
		t.Fatal("selection did not move to a different file; setup is broken")
	}
	want, err := os.ReadFile(secondPath)
	if err != nil {
		t.Fatalf("reading %s: %v", secondPath, err)
	}
	wantContent := renderText(want, m.picker.preview.Width(), false)
	if got := m.picker.preview.GetContent(); got != wantContent {
		t.Errorf("preview content after selection moved = %q, want %s's Text render %q", got, filepath.Base(secondPath), wantContent)
	}
	if got := m.picker.preview.GetContent(); got == firstContent {
		t.Error("preview content unchanged after selection moved to a different file")
	}
	if !m.picker.preview.AtTop() {
		t.Error("preview did not scroll back to the top after selection moved")
	}
}

// TestPickerPreviewScrollsOnScrollKeys is the "scrollable" half of M3.2: with
// a file longer than the Preview pane's height loaded, the reused Document
// Viewer scroll bindings (PageDown then PageUp) move the Preview's own
// scroll position, independent of the Results cursor.
func TestPickerPreviewScrollsOnScrollKeys(t *testing.T) {
	requireRipgrep(t)
	dir := t.TempDir()
	var long strings.Builder
	for i := 1; i <= 200; i++ {
		fmt.Fprintf(&long, "line %d\n", i)
	}
	writeFileContent(t, filepath.Join(dir, "reportlong.md"), long.String())

	m := openedFixture(t, 80, 24).withRipgrep(detectRipgrep())
	m.roots = m.roots.setRoots([]root{{path: dir}}, 0)
	m, _ = m.openPicker()

	m, seqs := typePickerQuery(t, m, "repor")
	m, searches := deliverPickerTicks(t, m, seqs)
	if searches != 1 || len(m.picker.results.files) == 0 {
		t.Fatalf("setup is broken: %d searches, %d results", searches, len(m.picker.results.files))
	}
	if !m.picker.preview.AtTop() {
		t.Fatal("preview did not start at the top; setup is broken")
	}

	m = press(t, m, keys.PageDown.Keys()[0])
	afterDown := m.picker.preview.YOffset()
	if afterDown <= 0 {
		t.Fatalf("YOffset = %d after PageDown, want > 0", afterDown)
	}

	m = press(t, m, keys.PageUp.Keys()[0])
	if got := m.picker.preview.YOffset(); got >= afterDown {
		t.Errorf("YOffset = %d after PageUp, want < %d", got, afterDown)
	}
}

// TestPickerPreviewShowsUnavailableMessageOnReadError pins syncPreview's
// error-content fallback (TRIAGE T12): when the selected Result's file
// can't be read (removed, permission denied, etc.), the Preview shows the
// exact "preview unavailable: " + err.Error() string rather than stale
// content or an unexplained blank pane — matching this file's convention
// (picker_view_test.go) of pinning user-visible strings exactly.
func TestPickerPreviewShowsUnavailableMessageOnReadError(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.md")
	_, statErr := os.Stat(missing)
	if statErr == nil {
		t.Fatal("setup is broken: missing file actually exists")
	}

	p := newPicker()
	p.input.SetValue("report")
	p.results = pickerResults{
		query: "report",
		mode:  fileMode,
		scope: activeRootScope,
		files: []pickerFile{{path: missing}},
	}

	p = p.syncPreview(pickerBounds{})

	_, readErr := os.ReadFile(missing)
	want := "preview unavailable: " + readErr.Error()
	if got := p.preview.GetContent(); got != want {
		t.Errorf("preview content = %q, want %q", got, want)
	}
}

// numberedLinesFixture writes a file of n lines, "filler line 1" .. "filler
// line n", except line matchLine (1-based) which reads text instead — the
// shape M4.2's own acceptance criterion needs: "a match deep in a long
// file", with a file long enough that the matched line starts off-screen at
// the Preview's default (top, M3.2) scroll position.
func numberedLinesFixture(t *testing.T, dir, name string, n, matchLine int, text string) string {
	t.Helper()
	var b strings.Builder
	for i := 1; i <= n; i++ {
		if i == matchLine {
			b.WriteString(text)
			b.WriteByte('\n')
		} else {
			fmt.Fprintf(&b, "filler line %d\n", i)
		}
	}
	return writeFileContent(t, filepath.Join(dir, name), b.String())
}

// TestPickerGrepPreviewOpensScrolledToMatchedLineWithHighlight is M4.2's
// own acceptance criterion (S9): activating a Grep search whose match sits
// deep in a long file leaves the Preview scrolled so that line is within
// the visible window, and carries the highlight style (D140: the exact span, not the whole line).
func TestPickerGrepPreviewOpensScrolledToMatchedLineWithHighlight(t *testing.T) {
	requireRipgrep(t)
	dir := t.TempDir()
	const matchLine = 87
	numberedLinesFixture(t, dir, "long.md", 100, matchLine, "the target report line")

	m := openedFixture(t, 80, 24).withRipgrep(detectRipgrep())
	m.roots = m.roots.setRoots([]root{{path: dir}}, 0)
	m, _ = m.openPicker()
	m.picker.mode = grepMode

	m, seqs := typePickerQuery(t, m, "report")
	m, searches := deliverPickerTicks(t, m, seqs)
	if searches != 1 || len(m.picker.results.grep) != 1 {
		t.Fatalf("setup is broken: %d searches, %d grep results", searches, len(m.picker.results.grep))
	}
	if got := m.picker.results.grep[0].line; got != matchLine {
		t.Fatalf("setup is broken: matched line = %d, want %d", got, matchLine)
	}

	idx := matchLine - 1 // short lines: one row each, so rows are 0-based line numbers
	yOff, height := m.picker.preview.YOffset(), m.picker.preview.Height()
	if idx < yOff || idx >= yOff+height {
		t.Fatalf("matched line %d (0-based) not within the visible window [%d, %d)", idx, yOff, yOff+height)
	}

	// D140: only the matched span is reversed, not the whole line.
	rows := strings.Split(m.picker.preview.GetContent(), "\n")
	want := gutterStyle.Render(fmt.Sprintf("%3d ", matchLine)) + "the target " + matchSpanStyle.Render("report") + " line"
	if rows[idx] != want {
		t.Errorf("matched row = %q, want %q", rows[idx], want)
	}
	if rows[idx-1] != gutterStyle.Render(fmt.Sprintf("%3d ", matchLine-1))+fmt.Sprintf("filler line %d", matchLine-1) {
		t.Errorf("row before the match = %q, want it plain", rows[idx-1])
	}
}

// TestPickerGrepPreviewRetargetsWithinTheSameFile is the "different match,
// same file" edge M3.2's own path-only re-targeting check
// (TestPickerPreviewRetargetsOnSelectionChange) cannot exercise, since two
// File Results are always two different files: moving the Results
// selection between two Grep matches inside one file must still re-scroll
// and re-highlight to the newly-selected match's own line, even though the
// Preview's loaded content (previewPath) does not change.
func TestPickerGrepPreviewRetargetsWithinTheSameFile(t *testing.T) {
	requireRipgrep(t)
	dir := t.TempDir()
	var b strings.Builder
	for i := 1; i <= 100; i++ {
		switch i {
		case 10:
			b.WriteString("first report line\n")
		case 90:
			b.WriteString("second report line\n")
		default:
			fmt.Fprintf(&b, "filler line %d\n", i)
		}
	}
	writeFileContent(t, filepath.Join(dir, "long.md"), b.String())

	m := openedFixture(t, 80, 24).withRipgrep(detectRipgrep())
	m.roots = m.roots.setRoots([]root{{path: dir}}, 0)
	m, _ = m.openPicker()
	m.picker.mode = grepMode

	m, seqs := typePickerQuery(t, m, "report")
	m, searches := deliverPickerTicks(t, m, seqs)
	if searches != 1 || len(m.picker.results.grep) != 2 {
		t.Fatalf("setup is broken: %d searches, %d grep results", searches, len(m.picker.results.grep))
	}

	firstIdx := m.picker.results.grep[m.picker.selected].line - 1
	if got := m.picker.preview.YOffset(); got != firstIdx {
		t.Fatalf("YOffset = %d after the first match landed, want %d; setup is broken", got, firstIdx)
	}

	m = press(t, m, keys.PickerDown.Keys()[0])

	secondIdx := m.picker.results.grep[m.picker.selected].line - 1
	if secondIdx == firstIdx {
		t.Fatal("selection did not move to a different line; setup is broken")
	}
	if got := m.picker.preview.YOffset(); got != secondIdx {
		t.Errorf("YOffset = %d after moving to the second match, want %d (re-targeted despite same file)", got, secondIdx)
	}
	rows := strings.Split(m.picker.preview.GetContent(), "\n")
	if !strings.Contains(rows[secondIdx], matchSpanStyle.Render("report")) {
		t.Errorf("row %d = %q, want the match span highlighted after re-targeting", secondIdx, rows[secondIdx])
	}
	if strings.Contains(rows[firstIdx], matchSpanStyle.Render("report")) {
		t.Errorf("row %d = %q (the previous match's line) is still highlighted", firstIdx, rows[firstIdx])
	}
}

// pickerRoot puts a fixture Root full of files whose names share a prefix
// behind an already-open Picker, so a query can be typed a character at a
// time and actually have something to find. Returns the model and the Root's
// path.
func pickerRoot(t *testing.T) (model, string) {
	t.Helper()
	dir := t.TempDir()
	writeFileContent(t, filepath.Join(dir, "report.md"), "quarterly report\n")
	writeFileContent(t, filepath.Join(dir, "reporting.md"), "reporting notes\n")
	writeFileContent(t, filepath.Join(dir, "other.md"), "unrelated\n")

	// The session must know rg is present: both search gates consult
	// m.ripgrep (S14, m.searchable), so a fixture without it would never
	// dispatch anything. Callers requireRipgrep, so detection here is true.
	m := openedFixture(t, 80, 24).withRipgrep(detectRipgrep())
	m.roots = m.roots.setRoots([]root{{path: dir}}, 0)
	m, _ = m.openPicker()
	return m, dir
}

// TestPickerFindsADotfileInBothModesWithNoExtraKeystroke is S4/M6.1's own
// acceptance criterion, verbatim: a dotfile visible in the Files Nav tree
// (tree.Walk applies no hidden-file rule of its own, internal/tree's own
// tests) is findable by both the File Picker and the Grep Picker with no
// extra keystroke — no toggle, no special syntax, just typing a query
// either mode would already take for a non-dotfile. File mode matches the
// dotfile's *name*; Grep mode separately matches its *content* — the two
// naturally take different queries (D31: File mode never looks at path or
// content), so this checks each mode's own natural match, not one query
// surviving a mode switch (TestPickerModeSwitchCarriesThePromptAndRerunsTheSearch,
// just below, already owns that).
func TestPickerFindsADotfileInBothModesWithNoExtraKeystroke(t *testing.T) {
	requireRipgrep(t)
	dir := t.TempDir()
	writeFileContent(t, filepath.Join(dir, ".config.md"), "dotfile needle content\n")

	m := openedFixture(t, 80, 24).withRipgrep(detectRipgrep())
	m.roots = m.roots.setRoots([]root{{path: dir}}, 0)

	// File mode (the Picker's default on open): the dotfile's own name is a
	// fuzzy match candidate like any other file.
	m, _ = m.openPicker()
	m, seqs := typePickerQuery(t, m, "config")
	m, searches := deliverPickerTicks(t, m, seqs)
	if searches == 0 {
		t.Fatal("no File search was dispatched; setup is broken")
	}
	if len(m.picker.results.files) != 1 || filepath.Base(m.picker.results.files[0].path) != ".config.md" {
		t.Fatalf("File mode results = %v, want the dotfile alone", m.picker.results.files)
	}

	// esc dismisses with no side effects (M2.3's own contract); reopening
	// fresh and switching to Grep before typing is the plain path to a Grep
	// search, no different from any other Grep query.
	m = press(t, m, keys.Cancel.Keys()[0])
	m, _ = m.openPicker()
	next, _ := m.Update(keyPress(keys.PickerMode.Keys()[0]))
	m = concrete(t, next)
	if m.picker.mode != grepMode {
		t.Fatalf("mode after PickerMode = %v, want grepMode", m.picker.mode)
	}
	m, seqs = typePickerQuery(t, m, "needle")
	m, searches = deliverPickerTicks(t, m, seqs)
	if searches == 0 {
		t.Fatal("no Grep search was dispatched; setup is broken")
	}
	if len(m.picker.results.grep) != 1 || filepath.Base(m.picker.results.grep[0].file.path) != ".config.md" {
		t.Fatalf("Grep mode results = %v, want one match inside the dotfile", m.picker.results.grep)
	}
}

// TestPickerAltHTogglesDotfilesInBothModesAndSurvivesReopen is S4/M6.2's own
// acceptance criterion, verbatim: alt+h hides dotfiles from both modes'
// Results, pressing it again restores them, and the hidden/shown state is
// unchanged after closing and reopening the Picker in the same session.
func TestPickerAltHTogglesDotfilesInBothModesAndSurvivesReopen(t *testing.T) {
	requireRipgrep(t)
	dir := t.TempDir()
	writeFileContent(t, filepath.Join(dir, ".config.md"), "dotfile needle content\n")
	writeFileContent(t, filepath.Join(dir, "plain.md"), "plain needle content\n")

	m := openedFixture(t, 80, 24).withRipgrep(detectRipgrep())
	m.roots = m.roots.setRoots([]root{{path: dir}}, 0)
	m, _ = m.openPicker()

	if m.files.hideDotfiles {
		t.Fatal("hideDotfiles starts true; want false (S4's default: dotfiles shown)")
	}

	// File mode: the dotfile is a Result by default (M6.1's own contract).
	m, seqs := typePickerQuery(t, m, "config")
	m, searches := deliverPickerTicks(t, m, seqs)
	if searches == 0 {
		t.Fatal("no File search was dispatched; setup is broken")
	}
	if len(m.picker.results.files) != 1 {
		t.Fatalf("File mode results before alt+h = %v, want the dotfile alone", m.picker.results.files)
	}

	// alt+h hides it: a non-typing key (handlePickerInput), so the Prompt is
	// untouched and the toggle re-runs the same query through the narrowed
	// candidate set (files_search.go's listFiles), the same queuePickerSearch
	// path a mode/Scope switch already reuses for its own in-place switch.
	next, cmd := m.Update(keyPress(keys.ToggleHidden.Keys()[0]))
	m = concrete(t, next)
	if !m.files.hideDotfiles {
		t.Fatal("hideDotfiles after alt+h = false, want true")
	}
	if got := m.picker.input.Value(); got != "config" {
		t.Fatalf("Prompt after alt+h = %q, want unchanged %q", got, "config")
	}
	msg, ok := cmd().(pickerSearchMsg)
	if !ok {
		t.Fatalf("alt+h's Cmd produced %T, want pickerSearchMsg", cmd())
	}
	m, searches = deliverPickerTicks(t, m, []int{msg.seq})
	if searches == 0 {
		t.Fatal("no File search was dispatched after alt+h")
	}
	if len(m.picker.results.files) != 0 {
		t.Fatalf("File mode results after alt+h = %v, want none (dotfile hidden)", m.picker.results.files)
	}

	// A second alt+h restores it.
	next, cmd = m.Update(keyPress(keys.ToggleHidden.Keys()[0]))
	m = concrete(t, next)
	if m.files.hideDotfiles {
		t.Fatal("hideDotfiles after a second alt+h = true, want false")
	}
	msg, ok = cmd().(pickerSearchMsg)
	if !ok {
		t.Fatalf("second alt+h's Cmd produced %T, want pickerSearchMsg", cmd())
	}
	m, searches = deliverPickerTicks(t, m, []int{msg.seq})
	if searches == 0 {
		t.Fatal("no File search was dispatched after the second alt+h")
	}
	if len(m.picker.results.files) != 1 {
		t.Fatalf("File mode results after the second alt+h = %v, want the dotfile restored", m.picker.results.files)
	}

	// Hide again, then close and reopen the Picker (dismissPicker replaces
	// the whole picker struct, resetPicker) — the toggle must survive that
	// within the session (filesNav.hideDotfiles, not picker-owned) even though
	// mode/Scope/Prompt do not.
	next, cmd = m.Update(keyPress(keys.ToggleHidden.Keys()[0]))
	m = concrete(t, next)
	msg, _ = cmd().(pickerSearchMsg)
	m, _ = deliverPickerTicks(t, m, []int{msg.seq})

	m = press(t, m, keys.Cancel.Keys()[0])
	m, _ = m.openPicker()
	if !m.files.hideDotfiles {
		t.Fatal("hideDotfiles after closing and reopening the Picker = false, want true (unchanged)")
	}

	// Grep mode: switching to it and searching again confirms the toggle is
	// shared by both modes (S4), still narrowing grepPickerFiles' own
	// candidate set (files_search.go), not just fuzzyFiles'.
	next, _ = m.Update(keyPress(keys.PickerMode.Keys()[0]))
	m = concrete(t, next)
	m, seqs = typePickerQuery(t, m, "needle")
	m, searches = deliverPickerTicks(t, m, seqs)
	if searches == 0 {
		t.Fatal("no Grep search was dispatched; setup is broken")
	}
	if len(m.picker.results.grep) != 1 || filepath.Base(m.picker.results.grep[0].file.path) != "plain.md" {
		t.Fatalf("Grep mode results while hidden = %v, want the plain file's match alone", m.picker.results.grep)
	}

	// A second alt+h, still in Grep mode, restores the dotfile's own match.
	next, cmd = m.Update(keyPress(keys.ToggleHidden.Keys()[0]))
	m = concrete(t, next)
	if m.files.hideDotfiles {
		t.Fatal("hideDotfiles after alt+h in Grep mode = true, want false")
	}
	msg, ok = cmd().(pickerSearchMsg)
	if !ok {
		t.Fatalf("alt+h in Grep mode's Cmd produced %T, want pickerSearchMsg", cmd())
	}
	m, searches = deliverPickerTicks(t, m, []int{msg.seq})
	if searches == 0 {
		t.Fatal("no Grep search was dispatched after alt+h in Grep mode")
	}
	if len(m.picker.results.grep) != 2 {
		t.Fatalf("Grep mode results after restoring = %v, want both files' matches", m.picker.results.grep)
	}
}

// TestPickerModeSwitchCarriesThePromptAndRerunsTheSearch is M5.1's own
// acceptance criterion, verbatim: typing a query in File mode, switching to
// Grep, and switching back, at each step re-running the same Prompt's
// search in the new mode without the Prompt itself ever changing. C6/M3.2
// widens the File ↔ Grep two-way flip to the three-way File → Grep →
// Wikipedia → File cycle, so "switching back" from Grep now passes through
// Wikipedia mode first — exercised here (Prompt still unchanged, mode is
// wikipediaMode) without asserting on its search/results, since Wikipedia
// mode's own search dispatch is M3.3's, not this test's, contract.
func TestPickerModeSwitchCarriesThePromptAndRerunsTheSearch(t *testing.T) {
	requireRipgrep(t)
	m, _ := pickerRoot(t)

	const query = "report"
	m, seqs := typePickerQuery(t, m, query)
	m, searches := deliverPickerTicks(t, m, seqs)
	if searches == 0 {
		t.Fatal("no initial File search was dispatched; setup is broken")
	}
	if m.picker.mode != fileMode {
		t.Fatalf("mode = %v, want fileMode before any switch", m.picker.mode)
	}
	if got := len(m.picker.results.files); got == 0 {
		t.Fatal("File mode results are empty; setup is broken")
	}

	// Switch to Grep: the PickerMode key is swallowed as a non-typing key
	// (handlePickerInput), so it must not touch the Prompt.
	next, cmd := m.Update(keyPress(keys.PickerMode.Keys()[0]))
	m = concrete(t, next)
	if m.picker.input.Value() != query {
		t.Fatalf("Prompt after switching to Grep = %q, want unchanged %q", m.picker.input.Value(), query)
	}
	if m.picker.mode != grepMode {
		t.Fatalf("mode after PickerMode = %v, want grepMode", m.picker.mode)
	}
	if cmd == nil {
		t.Fatal("switching mode scheduled no debounce timer")
	}
	msg, ok := cmd().(pickerSearchMsg)
	if !ok {
		t.Fatalf("switching mode's Cmd produced %T, want pickerSearchMsg", cmd())
	}
	m, searches = deliverPickerTicks(t, m, []int{msg.seq})
	if searches == 0 {
		t.Fatal("no Grep search was dispatched after the mode switch")
	}
	if m.picker.results.mode != grepMode {
		t.Fatalf("results.mode after switch = %v, want grepMode", m.picker.results.mode)
	}
	if got := len(m.picker.results.grep); got == 0 {
		t.Fatal("Grep mode results are empty after the mode switch")
	}
	if m.picker.input.Value() != query {
		t.Fatalf("Prompt after Grep results landed = %q, want unchanged %q", m.picker.input.Value(), query)
	}

	// Switch to Wikipedia (C6/M3.2's third cycle position): the Prompt is
	// still untouched. Deliberately not delivering this leg's debounce tick
	// or asserting on results — Wikipedia mode's own search is M3.3's.
	next, cmd = m.Update(keyPress(keys.PickerMode.Keys()[0]))
	m = concrete(t, next)
	if m.picker.input.Value() != query {
		t.Fatalf("Prompt after switching to Wikipedia = %q, want unchanged %q", m.picker.input.Value(), query)
	}
	if m.picker.mode != wikipediaMode {
		t.Fatalf("mode after second PickerMode = %v, want wikipediaMode", m.picker.mode)
	}
	if cmd == nil {
		t.Fatal("switching to Wikipedia mode scheduled no debounce timer")
	}

	// Switch back to File: same Prompt, results become File results again.
	next, cmd = m.Update(keyPress(keys.PickerMode.Keys()[0]))
	m = concrete(t, next)
	if m.picker.input.Value() != query {
		t.Fatalf("Prompt after switching back to File = %q, want unchanged %q", m.picker.input.Value(), query)
	}
	if m.picker.mode != fileMode {
		t.Fatalf("mode after third PickerMode = %v, want fileMode", m.picker.mode)
	}
	msg, ok = cmd().(pickerSearchMsg)
	if !ok {
		t.Fatalf("switching mode back's Cmd produced %T, want pickerSearchMsg", cmd())
	}
	m, searches = deliverPickerTicks(t, m, []int{msg.seq})
	if searches == 0 {
		t.Fatal("no File search was dispatched after switching back")
	}
	if m.picker.results.mode != fileMode {
		t.Fatalf("results.mode after switching back = %v, want fileMode", m.picker.results.mode)
	}
	if got := len(m.picker.results.files); got == 0 {
		t.Fatal("File mode results are empty after switching back")
	}
}

// TestPickerScopeSwitchFindsNonActiveRootOnlyAfterSwitchingToAllRoots is
// M5.2's own acceptance criterion, verbatim: a query matching only in a
// non-active Root finds nothing under Active Source Scope, finds it once
// switched to All Roots, the Prompt is unchanged throughout, and a Broken
// Root in the session produces no error row (S6's "All Roots silently
// skips Broken Roots", already M1.2's contract at the listFiles layer;
// this is the same guarantee observed through the Picker's own state).
func TestPickerScopeSwitchFindsNonActiveRootOnlyAfterSwitchingToAllRoots(t *testing.T) {
	requireRipgrep(t)
	m := openedFixture(t, 80, 24).withRipgrep(detectRipgrep())

	activeDir := t.TempDir()
	writeFileContent(t, filepath.Join(activeDir, "other.md"), "unrelated\n")
	otherDir := t.TempDir()
	writeFileContent(t, filepath.Join(otherDir, "report.md"), "quarterly report\n")
	brokenDir := filepath.Join(t.TempDir(), "gone")

	m.roots = m.roots.setRoots([]root{{path: activeDir}, {path: brokenDir}, {path: otherDir}}, 0)
	m, _ = m.openPicker()

	const query = "report"
	m, seqs := typePickerQuery(t, m, query)
	m, searches := deliverPickerTicks(t, m, seqs)
	if searches == 0 {
		t.Fatal("no initial Active Source search was dispatched; setup is broken")
	}
	if m.picker.scope != activeRootScope {
		t.Fatalf("scope = %v, want activeRootScope before any switch", m.picker.scope)
	}
	if got := len(m.picker.results.files); got != 0 {
		t.Fatalf("Active Source results = %d files, want 0 (the match is only in a non-active Root)", got)
	}
	if m.message != "" {
		t.Fatalf("message after Active Source search = %q, want empty (no error row)", m.message)
	}

	// Switch to All Roots: the PickerScope key is swallowed as a non-typing
	// key (handlePickerInput), so it must not touch the Prompt.
	next, cmd := m.Update(keyPress(keys.PickerScope.Keys()[0]))
	m = concrete(t, next)
	if m.picker.input.Value() != query {
		t.Fatalf("Prompt after switching Scope = %q, want unchanged %q", m.picker.input.Value(), query)
	}
	if m.picker.scope != allRootsScope {
		t.Fatalf("scope after PickerScope = %v, want allRootsScope", m.picker.scope)
	}
	if cmd == nil {
		t.Fatal("switching Scope scheduled no debounce timer")
	}
	msg, ok := cmd().(pickerSearchMsg)
	if !ok {
		t.Fatalf("switching Scope's Cmd produced %T, want pickerSearchMsg", cmd())
	}
	m, searches = deliverPickerTicks(t, m, []int{msg.seq})
	if searches == 0 {
		t.Fatal("no All Roots search was dispatched after the Scope switch")
	}
	if m.picker.results.scope != allRootsScope {
		t.Fatalf("results.scope after switch = %v, want allRootsScope", m.picker.results.scope)
	}
	gotPaths := make([]string, 0, len(m.picker.results.files))
	for _, f := range m.picker.results.files {
		gotPaths = append(gotPaths, f.path)
	}
	want := filepath.Join(otherDir, "report.md")
	if !reflect.DeepEqual(gotPaths, []string{want}) {
		t.Fatalf("All Roots results = %v, want [%s]", gotPaths, want)
	}
	if m.picker.input.Value() != query {
		t.Fatalf("Prompt after All Roots results landed = %q, want unchanged %q", m.picker.input.Value(), query)
	}
	if m.message != "" {
		t.Fatalf("message after All Roots search = %q, want empty (Broken Root produced no error row)", m.message)
	}
}

// TestPickerNoMemoryWithinAndAcrossSessions is M5.3's own acceptance
// criterion, verbatim: dismissing a Picker left in Grep mode / All Roots
// Scope with a query typed, then reopening it, finds File Picker / Active
// Root / empty Prompt — both "within a session" (a reopen after dismiss)
// and "across sessions" (a fresh model, model.go's newModel, starts its
// picker field from the exact same newPicker() a reopen resets to, so
// there is no separate cross-session reset path to prove independently).
// It also confirms no Picker field ever reaches the on-disk store: the
// only thing store.go's toggleStore writes is starred Roots (S2), and
// nothing in picker.go calls it, so an untouched store file after a full
// open/type/switch-mode/switch-scope/dismiss/reopen cycle is the concrete,
// regression-catching form of that guarantee — a future change that wired
// the Picker into persistence would flip this from pass to fail.
func TestPickerNoMemoryWithinAndAcrossSessions(t *testing.T) {
	requireRipgrep(t)
	m, dir := pickerRoot(t)

	storePath := filepath.Join(t.TempDir(), "roots")
	if err := os.WriteFile(storePath, []byte(dir+"\n"), 0o644); err != nil {
		t.Fatalf("seeding store: %v", err)
	}
	before, err := os.ReadFile(storePath)
	if err != nil {
		t.Fatalf("reading seeded store: %v", err)
	}
	m.roots.storePath = storePath

	m, seqs := typePickerQuery(t, m, "report")
	m, _ = deliverPickerTicks(t, m, seqs)

	next, cmd := m.Update(keyPress(keys.PickerMode.Keys()[0])) // -> grepMode
	m = concrete(t, next)
	if msg, ok := cmd().(pickerSearchMsg); ok {
		m, _ = deliverPickerTicks(t, m, []int{msg.seq})
	}
	next, cmd = m.Update(keyPress(keys.PickerScope.Keys()[0])) // -> allRootsScope
	m = concrete(t, next)
	if msg, ok := cmd().(pickerSearchMsg); ok {
		m, _ = deliverPickerTicks(t, m, []int{msg.seq})
	}

	if m.picker.mode != grepMode || m.picker.scope != allRootsScope || m.picker.input.Value() != "report" {
		t.Fatalf("setup did not leave the Picker in Grep/AllRoots/%q; got %+v", "report", m.picker)
	}

	m = press(t, m, keys.Cancel.Keys()[0])
	if m.picker.open {
		t.Fatal("picker still open after Cancel; setup is broken")
	}
	// Dismissal itself discards the state — checked before any reopen, so a
	// dismissPicker that closed the modal without actually resetting mode/
	// Scope/Prompt (leaving openPicker's own reset to paper over it on the
	// next open) would still fail here.
	if m.picker.mode != fileMode || m.picker.scope != activeRootScope || m.picker.input.Value() != "" {
		t.Fatalf("picker state right after dismissal = %+v, want File/ActiveRoot/empty already", m.picker)
	}

	// "Within a session": reopening in the same process.
	m, _ = m.openPicker()
	if m.picker.mode != fileMode || m.picker.scope != activeRootScope || m.picker.input.Value() != "" {
		t.Errorf("reopened picker = %+v, want File/ActiveRoot/empty", m.picker)
	}

	// "Across sessions": a fresh process's model starts from the same
	// newPicker() zero state (model.go's newModel), with no serialized
	// Picker state anywhere to have restored from in the first place.
	fresh := newModel()
	if fresh.picker.mode != fileMode || fresh.picker.scope != activeRootScope || fresh.picker.input.Value() != "" {
		t.Errorf("fresh model's picker = %+v, want File/ActiveRoot/empty", fresh.picker)
	}

	after, err := os.ReadFile(storePath)
	if err != nil {
		t.Fatalf("reading store after the Picker interaction: %v", err)
	}
	if string(after) != string(before) {
		t.Errorf("store file changed by the Picker interaction: before %q, after %q", before, after)
	}
}

// typePickerQuery types s into the open Picker one rune at a time, returning
// the model and the query generation (picker.seq) each keystroke left behind
// — i.e. the generation each keystroke's debounce timer will report when it
// matures. Deliberately does not run the returned Cmds: a real tea.Tick
// would sleep pickerSearchDebounce, and what the timers deliver is exactly
// these seqs (pickerSearchMsg), which deliverPickerTicks then plays back.
func typePickerQuery(t *testing.T, m model, s string) (model, []int) {
	t.Helper()
	var seqs []int
	for _, r := range s {
		next, _ := m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		m = concrete(t, next)
		seqs = append(seqs, m.picker.seq)
	}
	return m, seqs
}

// deliverPickerTicks plays back each scheduled debounce timer in order, as
// the real timers would fire, and runs any search they dispatch to
// completion (feeding its results message back into the model, as Bubble Tea
// would). Returns the model and how many searches were actually dispatched —
// a non-nil Cmd from a pickerSearchMsg is precisely "this timer issued a
// search", since dispatchPickerSearch returns nil for every superseded one.
func deliverPickerTicks(t *testing.T, m model, seqs []int) (model, int) {
	t.Helper()
	searches := 0
	for _, seq := range seqs {
		next, cmd := m.Update(pickerSearchMsg{seq: seq})
		m = concrete(t, next)
		if cmd == nil {
			continue
		}
		searches++
		msg, ok := cmd().(pickerResultsMsg)
		if !ok {
			t.Fatalf("search for seq %d produced %T, want pickerResultsMsg", seq, cmd())
		}
		next, _ = m.Update(msg)
		m = concrete(t, next)
	}
	return m, searches
}

// TestPickerSearchIsDebouncedAndShowsOnlyTheFinalQuery is M2.4's acceptance
// criterion, verbatim: fast successive queries (a query typed a character at
// a time, with no pause long enough for a debounce to mature) leave the
// displayed results corresponding to the final query only, and issue fewer
// searches than there were keystrokes. It also pins the "the Prompt never
// blocks" half: typing itself searches nothing at all — no results exist
// until a timer matures.
func TestPickerSearchIsDebouncedAndShowsOnlyTheFinalQuery(t *testing.T) {
	requireRipgrep(t)
	m, _ := pickerRoot(t)

	const query = "repor"
	m, seqs := typePickerQuery(t, m, query)

	if m.picker.input.Value() != query {
		t.Fatalf("prompt = %q, want %q; setup is broken", m.picker.input.Value(), query)
	}
	if m.picker.results.query != "" || m.picker.searching {
		t.Errorf("typing alone searched: results = %+v, searching = %v; want neither",
			m.picker.results, m.picker.searching)
	}

	m, searches := deliverPickerTicks(t, m, seqs)

	if searches != 1 {
		t.Errorf("dispatched %d searches for %d keystrokes, want exactly 1", searches, len(seqs))
	}
	if searches >= len(query) {
		t.Errorf("dispatched %d searches for %d keystrokes, want fewer", searches, len(query))
	}
	if m.picker.results.query != query {
		t.Errorf("displayed results answer %q, want the final query %q", m.picker.results.query, query)
	}
	if m.picker.searching {
		t.Error("searching = true after results landed, want false")
	}
	var names []string
	for _, f := range m.picker.results.files {
		names = append(names, filepath.Base(f.path))
	}
	want := []string{"report.md", "reporting.md"}
	if !equalStrings(sortedCopy(names), sortedCopy(want)) {
		t.Errorf("results = %v, want %v", names, want)
	}
}

// TestPickerDiscardsStaleResults is the discard half of M2.4 from the other
// side: a search that was genuinely dispatched (the user paused) but reports
// back after the query moved on must not be displayed, even for a frame.
func TestPickerDiscardsStaleResults(t *testing.T) {
	requireRipgrep(t)
	m, _ := pickerRoot(t)

	m, seqs := typePickerQuery(t, m, "repo")
	next, cmd := m.Update(pickerSearchMsg{seq: seqs[len(seqs)-1]})
	m = concrete(t, next)
	if cmd == nil {
		t.Fatal("the settled query dispatched no search; setup is broken")
	}
	if !m.picker.searching {
		t.Error("searching = false while a dispatched search is in flight, want true")
	}

	// The user types on while rg is still working, then its results land.
	m, _ = typePickerQuery(t, m, "rt")
	next, _ = m.Update(cmd())
	m = concrete(t, next)

	if m.picker.results.query != "" {
		t.Errorf("displayed stale results for %q; prompt is now %q",
			m.picker.results.query, m.picker.input.Value())
	}
}

// TestPickerDiscardsResultsFromAnEarlierOpen is why picker.seq survives
// dismissal (dismissPicker): a search dispatched by one Picker can report
// back after it is dismissed and another opened, and those results belong to
// neither.
func TestPickerDiscardsResultsFromAnEarlierOpen(t *testing.T) {
	requireRipgrep(t)
	m, _ := pickerRoot(t)

	m, seqs := typePickerQuery(t, m, "repo")
	next, cmd := m.Update(pickerSearchMsg{seq: seqs[len(seqs)-1]})
	m = concrete(t, next)
	if cmd == nil {
		t.Fatal("the settled query dispatched no search; setup is broken")
	}

	m = press(t, m, keys.Cancel.Keys()[0])
	m, _ = m.openPicker()
	next, _ = m.Update(cmd())
	m = concrete(t, next)

	if m.picker.results.query != "" {
		t.Errorf("results from before the dismissal survived into a fresh Picker: %+v", m.picker.results)
	}
	if m.picker.searching {
		t.Error("a fresh Picker reports searching = true, want false")
	}
}

// TestPickerEmptyPromptClearsResultsAndSearchesNothing: deleting a query back
// to nothing must not leave the old query's results on display, and must not
// dispatch a search for the empty string (which rg would match against every
// line of every file — see grepPickerFiles).
func TestPickerEmptyPromptClearsResultsAndSearchesNothing(t *testing.T) {
	requireRipgrep(t)
	m, _ := pickerRoot(t)

	m, seqs := typePickerQuery(t, m, "repor")
	m, searches := deliverPickerTicks(t, m, seqs)
	if searches != 1 || len(m.picker.results.files) == 0 {
		t.Fatalf("setup is broken: %d searches, %d results", searches, len(m.picker.results.files))
	}

	// Backspace is not a keymap.go binding (it belongs to the textinput), so
	// the keypress is built here rather than via keyPress.
	var deleted []int
	for range "repor" {
		next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
		m = concrete(t, next)
		deleted = append(deleted, m.picker.seq)
	}
	if m.picker.input.Value() != "" {
		t.Fatalf("prompt = %q after deleting the query, want empty", m.picker.input.Value())
	}
	if m.picker.results.query != "" || len(m.picker.results.files) != 0 {
		t.Errorf("results survived an emptied prompt: %+v", m.picker.results)
	}

	m, searches = deliverPickerTicks(t, m, deleted)
	if searches != 0 {
		t.Errorf("dispatched %d searches for an empty prompt, want 0", searches)
	}
}

// TestPickerIssuesNoSearchBelowTheThreshold is the second half of M2.5's own
// acceptance criterion, revised for S13/C6/M3.1's 2-character threshold: no
// search is issued for a 1-character query, while the same query one
// character longer (2 characters, S13's own "a 2-character File search
// runs") does search. Checked through the full keystroke-then-timer path,
// so it covers both gates: a sub-threshold keystroke schedules a timer that
// dispatches nothing.
func TestPickerIssuesNoSearchBelowTheThreshold(t *testing.T) {
	requireRipgrep(t)

	m, _ := pickerRoot(t)
	m, seqs := typePickerQuery(t, m, "r")
	m, searches := deliverPickerTicks(t, m, seqs)

	if searches != 0 {
		t.Errorf("query %q dispatched %d searches, want 0", "r", searches)
	}
	if m.picker.results.query != "" {
		t.Errorf("query %q produced results for %q, want none", "r", m.picker.results.query)
	}
	if m.picker.searching {
		t.Errorf("query %q left searching = true, want false", "r")
	}

	m, _ = pickerRoot(t)
	m, seqs = typePickerQuery(t, m, "re")
	m, searches = deliverPickerTicks(t, m, seqs)
	if searches != 1 {
		t.Errorf("query %q dispatched %d searches, want 1 (the threshold is 2)", "re", searches)
	}
	if m.picker.results.query != "re" {
		t.Errorf("results answer %q, want %q", m.picker.results.query, "re")
	}
}

// TestPickerWikipediaModeDebounceExceedsFileModeDebounce is S13/C6/M3.1's
// other half ("Wikipedia mode waits somewhat longer than local modes since
// each search is a network call"): the debounce queuePickerSearch schedules
// for wikipediaMode is strictly longer than the one it schedules for
// fileMode (grepMode shares fileMode's). Checked against pickerDebounceFor
// directly rather than by timing a real tea.Tick, the same reason every
// other Picker async test (typePickerQuery/deliverPickerTicks above) avoids
// sleeping on the real debounce: duration is a value, not something worth
// racing a clock over.
func TestPickerWikipediaModeDebounceExceedsFileModeDebounce(t *testing.T) {
	fileDebounce := pickerDebounceFor(fileMode)
	grepDebounce := pickerDebounceFor(grepMode)
	wikipediaDebounce := pickerDebounceFor(wikipediaMode)

	if wikipediaDebounce <= fileDebounce {
		t.Errorf("wikipediaMode debounce = %v, want strictly more than fileMode's %v", wikipediaDebounce, fileDebounce)
	}
	if grepDebounce != fileDebounce {
		t.Errorf("grepMode debounce = %v, want the same as fileMode's %v", grepDebounce, fileDebounce)
	}
}

// TestPickerIssuesNoSearchWithoutRipgrep is S14's other half in model terms:
// on a session where rg was not found, an above-threshold query still types
// but dispatches nothing (the rendered explanation is
// TestPickerStatusStates' "ripgrep unavailable" case).
func TestPickerIssuesNoSearchWithoutRipgrep(t *testing.T) {
	m, _ := pickerRoot(t)
	m = m.withRipgrep(ripgrepAvailability{reason: "ripgrep (rg) not found on PATH"})

	m, seqs := typePickerQuery(t, m, "report")
	m, searches := deliverPickerTicks(t, m, seqs)

	if m.picker.input.Value() != "report" {
		t.Errorf("prompt = %q, want %q (typing still works without rg)", m.picker.input.Value(), "report")
	}
	if searches != 0 {
		t.Errorf("dispatched %d searches with no rg available, want 0", searches)
	}
}

// TestPickerCursorKeysDoNotRestartTheSearch: a key the Prompt consumes
// without changing the query (a cursor move) must not open a new query
// generation, or it would discard the results of a search for a query that
// is still current.
func TestPickerCursorKeysDoNotRestartTheSearch(t *testing.T) {
	requireRipgrep(t)
	m, _ := pickerRoot(t)

	m, seqs := typePickerQuery(t, m, "repor")
	m, _ = deliverPickerTicks(t, m, seqs)
	before := m.picker.seq
	results := m.picker.results

	m = press(t, m, "left")
	m = press(t, m, "home")

	if m.picker.seq != before {
		t.Errorf("picker.seq = %d after cursor moves, want %d (unchanged)", m.picker.seq, before)
	}
	if m.picker.results.query != results.query || len(m.picker.results.files) != len(results.files) {
		t.Errorf("results changed on a cursor move: %+v, want %+v", m.picker.results, results)
	}
}

// TestPickerGrepModeWiresToGrepPickerFilesAndRendersRows is M4.1's own
// acceptance criterion end to end: switching to Grep mode and typing a
// query actually dispatches grepPickerFiles (M1.4) over the fixture Root
// through the same debounce/dispatch path M2.4 already gives File mode
// (TestPickerSearchIsDebouncedAndShowsOnlyTheFinalQuery's own shape), and
// the results land as one Results row per matching line with the correct
// (1-based, rg's own) line number — not merely a result set assembled by
// hand and handed straight to the renderer, the way
// TestPickerGrepResultsShowFileLineNumberAndText (picker_view_test.go)
// isolates the rendering half alone.
func TestPickerGrepModeWiresToGrepPickerFilesAndRendersRows(t *testing.T) {
	requireRipgrep(t)
	m, _ := pickerRoot(t)
	m.picker.mode = grepMode

	const query = "report"
	m, seqs := typePickerQuery(t, m, query)
	m, searches := deliverPickerTicks(t, m, seqs)

	if searches == 0 {
		t.Fatal("no grep search was dispatched; setup is broken")
	}
	if m.picker.results.mode != grepMode {
		t.Fatalf("results.mode = %v, want grepMode", m.picker.results.mode)
	}

	gotLines := make(map[string]int, len(m.picker.results.grep))
	for _, g := range m.picker.results.grep {
		gotLines[filepath.Base(g.file.path)] = g.line
	}
	wantLines := map[string]int{"report.md": 1, "reporting.md": 1}
	if !reflect.DeepEqual(gotLines, wantLines) {
		t.Fatalf("grep results (file -> line) = %v, want %v", gotLines, wantLines)
	}

	joined := strings.Join(plainLines(m.View().Content), "\n")
	if !strings.Contains(joined, "report.md:1") {
		t.Errorf("rendered modal is missing a report.md:1 row\n%s", joined)
	}
	if !strings.Contains(joined, "reporting.md:1") {
		t.Errorf("rendered modal is missing a reporting.md:1 row\n%s", joined)
	}
}

// TestPickerGrepPreviewMatchesTextViewerAfterActivation is tk-uld's
// acceptance criterion (D140): the Preview of a Grep Result is what a
// Text-mode viewer shows once the Result is activated — soft wrap, expanded
// tabs, the gutter and the exact match span — here on a line that wraps and
// carries a tab before the match.
func TestPickerGrepPreviewMatchesTextViewerAfterActivation(t *testing.T) {
	requireRipgrep(t)
	dir := t.TempDir()
	long := "\tintro " + strings.Repeat("word ", 30) + "needle tail " + strings.Repeat("more ", 10)
	writeFileContent(t, filepath.Join(dir, "wrap.md"), "first line\n"+long+"\nlast line\n")

	m := openedFixture(t, 80, 24).withRipgrep(detectRipgrep())
	m.dv1.mode = modeText
	m.roots = m.roots.setRoots([]root{{path: dir}}, 0)
	m, _ = m.openPicker()
	m.picker.mode = grepMode
	m, seqs := typePickerQuery(t, m, "needle")
	m, _ = deliverPickerTicks(t, m, seqs)
	if len(m.picker.results.grep) != 1 {
		t.Fatalf("setup is broken: %d grep results", len(m.picker.results.grep))
	}

	preview := m.picker.preview
	if rows := strings.Count(preview.GetContent(), "\n") + 1; rows < 5 {
		t.Fatalf("setup is broken: only %d rows, the line did not wrap", rows)
	}
	if !strings.Contains(preview.GetContent(), matchSpanStyle.Render("needle")) {
		t.Errorf("preview does not highlight the exact span: %q", preview.GetContent())
	}
	if strings.Contains(preview.GetContent(), "\t") {
		t.Errorf("preview still has a tab: %q", preview.GetContent())
	}

	// Give the Preview the viewer's width so the two are comparable.
	picker := m.picker.setPreviewSize(m.dv1.vp.Width(), m.picker.preview.Height())
	m = m.activateGrepPicker()
	if m.dv1.text == nil {
		t.Fatal("setup is broken: viewer is not in Text mode")
	}
	if got, want := picker.preview.GetContent(), m.dv1.vp.GetContent(); got != want {
		t.Errorf("preview differs from the Text viewer after activation:\n preview: %q\n viewer:  %q", got, want)
	}
}

// TestPickerPreviewRelayoutsOnResize pins that the Preview's soft wrap tracks
// its width: narrowing the pane wraps the loaded file again, and the Grep
// match stays highlighted and scrolled to.
func TestPickerPreviewRelayoutsOnResize(t *testing.T) {
	requireRipgrep(t)
	dir := t.TempDir()
	writeFileContent(t, filepath.Join(dir, "a.md"), "short\n"+strings.Repeat("abcdefghij ", 12)+"needle\n")

	m := openedFixture(t, 120, 24).withRipgrep(detectRipgrep())
	m.roots = m.roots.setRoots([]root{{path: dir}}, 0)
	m, _ = m.openPicker()
	m.picker.mode = grepMode
	m, seqs := typePickerQuery(t, m, "needle")
	m, _ = deliverPickerTicks(t, m, seqs)

	wide := strings.Count(m.picker.preview.GetContent(), "\n")
	p := m.picker.setPreviewSize(20, m.picker.preview.Height())
	if narrow := strings.Count(p.preview.GetContent(), "\n"); narrow <= wide {
		t.Errorf("rows after narrowing = %d, want more than %d", narrow+1, wide+1)
	}
	if !strings.Contains(p.preview.GetContent(), matchSpanStyle.Render("needle")) {
		t.Errorf("match span lost after resize: %q", p.preview.GetContent())
	}
}
