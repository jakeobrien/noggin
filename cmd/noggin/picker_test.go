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
	for _, p := range paneOrder {
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
// resets to File Picker / Active Root / empty Prompt regardless.
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
// was found — the same Active Document, Active Root and scroll position,
// and (since M2.1 predates any Picker rendering, M2.2 onward's job) the
// exact same rendered frame.
func TestPickerDismissRestoresThePreOpenFrameByteForByte(t *testing.T) {
	m := openedFixture(t, 80, 24)
	m = press(t, m, "j") // move the viewer's scroll position off zero
	beforeFrame := m.View().Content
	beforeDoc := m.viewer.doc.path
	beforeRoot := m.roots.active
	beforeOffset := m.viewer.offset()

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
	if m.viewer.doc.path != beforeDoc {
		t.Errorf("active document = %q, want %q", m.viewer.doc.path, beforeDoc)
	}
	if m.roots.active != beforeRoot {
		t.Errorf("active root = %d, want %d", m.roots.active, beforeRoot)
	}
	if m.viewer.offset() != beforeOffset {
		t.Errorf("scroll offset = %d, want %d", m.viewer.offset(), beforeOffset)
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
	beforeDoc := m.viewer.doc.path
	beforeOffset := m.viewer.offset()

	const query = "jog/report"
	m = typeText(t, m, query)

	if got := m.picker.input.Value(); got != query {
		t.Errorf("prompt = %q, want %q", got, query)
	}
	if !m.picker.open {
		t.Error("picker.open = false after typing, want true")
	}
	if m.viewer.doc.path != beforeDoc {
		t.Errorf("active document changed to %q while typing, want unchanged (%q)", m.viewer.doc.path, beforeDoc)
	}
	if m.viewer.offset() != beforeOffset {
		t.Errorf("viewer scroll offset changed to %d while typing, want unchanged (%d)", m.viewer.offset(), beforeOffset)
	}
	if m.viewer.findEditing() {
		t.Error("\"/\" in the query started a Find, want it to only type")
	}
	if m.helpOpen {
		t.Error("typing opened the Help Modal, want it to stay closed")
	}
}

// TestPickerNonTypingKeysDoNotType is S2's other half: Cancel, Activate,
// PickerUp/PickerDown, the Preview scroll keys and PickerMode/PickerScope
// are all non-printable (D2's "always typing" only ever binds printable
// keystrokes), so none of them should ever reach the Prompt — checked here
// for every one but Cancel, which this task's earlier acceptance criterion
// (M2.1's dismiss-on-Cancel) already covers by making the Picker close
// instead.
func TestPickerNonTypingKeysDoNotType(t *testing.T) {
	nonTyping := []key.Binding{
		keys.Activate, keys.PickerUp, keys.PickerDown,
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
	fileB := filepath.Join(rootB, "report.md")
	writeFileContent(t, fileB, "root B report\n")

	nodeA, err := tree.Walk(rootA)
	if err != nil {
		t.Fatalf("walking %s: %v", rootA, err)
	}

	m := openedFixture(t, 80, 24).withRipgrep(ripgrepAvailability{available: true})
	m.roots = m.roots.setRoots([]root{{path: rootA}, {path: rootB}}, 0)
	m = m.openFolder(nodeA)
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
	if got.viewer.doc.path != fileB {
		t.Errorf("Active Document = %q, want %q", got.viewer.doc.path, fileB)
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
	beforeDoc := m.viewer.doc.path

	got := press(t, m, keys.Activate.Keys()[0])

	if got.picker.open {
		t.Error("picker.open = true after a failed Activate, want dismissed")
	}
	if got.viewer.doc.path != beforeDoc {
		t.Errorf("Active Document = %q after a failed Activate, want unchanged %q", got.viewer.doc.path, beforeDoc)
	}
	if got.message == "" {
		t.Error("message is empty after a failed Activate, want a footer message")
	}
}

// TestPickerActivateIsANoOpInGrepMode is M3.3a's Grep-mode case (TRIAGE
// T8): M4.1 hasn't wired a Grep Result to Activate yet (activatePicker's
// own comment — "grepMode has no candidate yet"), so pressing Activate
// while mode = grepMode must leave everything exactly as it was: no
// document opened, no dismissal, no message, not even an unhandled-mode
// crash.
func TestPickerActivateIsANoOpInGrepMode(t *testing.T) {
	m := openedFixture(t, 80, 24).withRipgrep(ripgrepAvailability{available: true})
	m, _ = m.openPicker()
	m.picker.mode = grepMode
	m.picker.input.SetValue("report")
	m.picker.results = pickerResults{
		query: "report",
		mode:  grepMode,
		scope: activeRootScope,
		grep:  []grepResult{{file: pickerFile{path: "somefile.md"}, line: 1, text: "report"}},
	}
	m.picker.selected = 0
	beforeDoc := m.viewer.doc.path
	beforeMessage := m.message

	got := press(t, m, keys.Activate.Keys()[0])

	if !got.picker.open {
		t.Error("picker.open = false after Activate in Grep mode, want unchanged (still open)")
	}
	if got.viewer.doc.path != beforeDoc {
		t.Errorf("Active Document = %q after Activate in Grep mode, want unchanged %q", got.viewer.doc.path, beforeDoc)
	}
	if got.message != beforeMessage {
		t.Errorf("message = %q after Activate in Grep mode, want unchanged %q", got.message, beforeMessage)
	}
}

// TestPickerPreviewShowsSelectedFileRawFromTop is M3.2's own acceptance
// criterion: the Preview pane's content is the selected Result's raw file
// bytes, byte for byte — no Glamour styling — and starts scrolled to the
// top.
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

	if got := m.picker.preview.GetContent(); got != string(want) {
		t.Errorf("preview content = %q, want the raw bytes of %s: %q", got, filepath.Base(selected.path), want)
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
	if got := m.picker.preview.GetContent(); got != string(want) {
		t.Errorf("preview content after selection moved = %q, want %s's raw bytes %q", got, filepath.Base(secondPath), want)
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

	p = p.syncPreview()

	_, readErr := os.ReadFile(missing)
	want := "preview unavailable: " + readErr.Error()
	if got := p.preview.GetContent(); got != want {
		t.Errorf("preview content = %q, want %q", got, want)
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
// line of every file — see grepFiles).
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
// acceptance criterion: no search is issued for a 1- or 2-character query
// (S12's threshold), while the same query one character longer does search.
// Checked through the full keystroke-then-timer path, so it covers both
// gates: a sub-threshold keystroke schedules a timer that dispatches nothing.
func TestPickerIssuesNoSearchBelowTheThreshold(t *testing.T) {
	requireRipgrep(t)

	for _, query := range []string{"r", "re"} {
		m, _ := pickerRoot(t)
		m, seqs := typePickerQuery(t, m, query)
		m, searches := deliverPickerTicks(t, m, seqs)

		if searches != 0 {
			t.Errorf("query %q dispatched %d searches, want 0", query, searches)
		}
		if m.picker.results.query != "" {
			t.Errorf("query %q produced results for %q, want none", query, m.picker.results.query)
		}
		if m.picker.searching {
			t.Errorf("query %q left searching = true, want false", query)
		}
	}

	m, _ := pickerRoot(t)
	m, seqs := typePickerQuery(t, m, "rep")
	m, searches := deliverPickerTicks(t, m, seqs)
	if searches != 1 {
		t.Errorf("query %q dispatched %d searches, want 1 (the threshold is 3)", "rep", searches)
	}
	if m.picker.results.query != "rep" {
		t.Errorf("results answer %q, want %q", m.picker.results.query, "rep")
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
