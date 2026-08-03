package main

import (
	"path/filepath"
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
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

	m := openedFixture(t, 80, 24)
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
