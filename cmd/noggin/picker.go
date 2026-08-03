package main

import (
	"os"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
)

// pickerMode selects which kind of search a Picker runs (S1, GLOSSARY.md
// "File Picker"/"Grep Picker"): fuzzy file-name matching or literal grep.
// The Picker always opens in fileMode (S1's fixed default); only an
// in-Picker switch (M5.1) ever changes it.
type pickerMode int

const (
	fileMode pickerMode = iota
	grepMode
)

// pickerSearchDebounce is how long a changed Prompt settles before its
// search is actually dispatched (M2.4). A keystroke only ever schedules a
// timer; the search itself runs in a tea.Cmd goroutine once no newer
// keystroke has superseded it (dispatchPickerSearch's seq check), so fast
// typing issues far fewer searches than keystrokes and the Prompt never
// waits on rg.
//
// 120ms is a build-time judgment call, not a spec'd number: long enough
// that an average typing burst collapses into one search, short enough that
// a pause between words still feels immediate.
const pickerSearchDebounce = 120 * time.Millisecond

// pickerQueryMinLength is S12's query threshold: neither Picker searches
// until at least this many characters are typed, and below it the Picker
// says so instead (pickerStatus, M2.5). Enforced at both search gates
// (queuePickerSearch and dispatchPickerSearch) for the same reason the
// empty-Prompt refusal it replaces was: whichever path scheduled a timer,
// no search may escape for a sub-threshold query.
//
// The number is S12's own ("at least 3 characters"), not a build-time
// choice: one- and two-character queries match a large fraction of any real
// collection, so searching them spends an rg run over the whole Scope to
// produce a result list nobody wanted.
const pickerQueryMinLength = 3

// pickerResults is one completed search's answer, labelled with the query,
// mode and Scope it answers *for*. The labels are the point: results arrive
// asynchronously, so the only way the modal can render honestly is for a
// result set to carry what it was a search for rather than being implicitly
// assumed to match whatever the Prompt now says (S12's status/empty states,
// M2.5, read these labels; M5.1/M5.2 make mode and Scope change under an
// in-flight search).
//
// files and grep are the two modes' shapes; exactly one is ever populated,
// selected by mode. truncated is S12's cap flag, passed straight through
// from fuzzyFiles/grepFiles.
type pickerResults struct {
	query     string
	mode      pickerMode
	scope     pickerScope
	files     []pickerFile
	grep      []grepResult
	truncated bool
}

// picker is the S1 modal's state: whether it is open, which mode and Scope
// it is searching, its Prompt (the typed query, S2), and the async search
// state behind it (M2.4) — the current query generation, whether a search
// for it is in flight, and the last results to arrive.
//
// S8's "no memory" means every user-visible field here is discarded on
// dismiss: dismissPicker (below) replaces the whole struct with
// newPicker()'s zero state rather than clearing fields piecemeal, so
// nothing can survive by accident. seq is the one exception, and carries no
// user-visible information — see dismissPicker.
type picker struct {
	open  bool
	mode  pickerMode
	scope pickerScope
	input textinput.Model

	// seq identifies the current query generation: every change to what
	// would be searched (a keystroke, and later a mode/Scope switch)
	// increments it. A debounce timer and a search's results both carry the
	// seq they were started for, and both are dropped unless it still
	// matches — that single comparison is what debounces dispatch and what
	// discards results for a query the user has already typed past (M2.4).
	// Monotonic for the life of the model, never reset, so an in-flight
	// search can never be mistaken for a newer one that happens to share a
	// count.
	seq int

	// searching is whether a dispatched search has yet to report back: set
	// when a debounce timer matures into a real search, cleared when that
	// search's results land. It is the difference between "searched, found
	// nothing" and "still looking", which S12's empty states (M2.5) must
	// tell apart.
	searching bool

	// results is the last search to report back for the current generation
	// (see pickerResults). Zero value means nothing has been searched yet.
	results pickerResults

	// selected is the Results cursor (M3.1): an index into the current
	// mode's result rows (resultCount), the same "index into what's on
	// screen" shape filesNav.selected and rootsNav.selected already give
	// their own lists. It is reset to 0 whenever a fresh result set lands
	// (applyPickerResults) or the Prompt drops below threshold
	// (queuePickerSearch) — an index into a candidate list that no longer
	// exists has nothing sensible to preserve, unlike a Nav pane's
	// selection surviving its own list unchanged.
	selected int

	// preview is the Preview pane's own viewport (M3.2): raw, unrendered
	// file bytes — no Glamour, unlike the Document Viewer's own vp
	// (viewer.go) — scrolled independently of the Prompt/Results. Sized
	// from model.resize the same way m.viewer/m.files/m.roots already are
	// (computePickerLayout's preview rect), so PageUp/PageDown/HalfPageUp/
	// HalfPageDown (reserved non-typing by M2.3) have a real height to
	// page against.
	preview viewport.Model

	// previewPath is the path whose bytes are currently loaded into
	// preview, or "" when nothing is. Compared against the current
	// preview candidate (pickerPreviewCandidate) so a load only happens
	// when the candidate actually changed — not on every keystroke or
	// render — and is what "re-targets when the selection moves" (M3.2's
	// acceptance criterion) means concretely: a different path, a fresh
	// load, GotoTop.
	previewPath string
}

// newPicker is the closed, no-search state: nothing open, nothing typed,
// File Picker mode and Active Root Scope — the zero values of pickerMode
// and pickerScope already are S1's fixed opening state, so there is nothing
// to set beyond a fresh Prompt.
func newPicker() picker {
	vp := viewport.New(viewport.WithWidth(0), viewport.WithHeight(0))
	// Same reason viewer.go's newViewer clears it: keys reach the Preview
	// through handlePickerInput's own non-typing set (keymap.go/D2), never
	// through the viewport's own KeyMap.
	vp.KeyMap = viewport.KeyMap{}
	vp.FillHeight = true
	return picker{input: textinput.New(), preview: vp}
}

// openPicker opens the Picker (S1): always File Picker mode, Active Root
// Scope, empty Prompt, regardless of what a previous open (in this or an
// earlier session) left behind — S1's fixed opening state and S8's "no
// memory" are the same requirement seen from two sides, so replacing the
// whole state with newPicker() satisfies both at once.
//
// Reachable from every Focus (S1), so unlike startFind/startEdit/activate,
// this never touches m.focus — opening the Picker is not "go to a pane",
// it's a modal layered on top of whichever pane already has focus, the same
// non-pane shape m.helpOpen has (model.go's openHelp).
func (m model) openPicker() (model, tea.Cmd) {
	seq := m.picker.seq + 1
	// The Preview viewport's width/height are a property of the terminal
	// (model.resize), not of any one Picker session, so they survive this
	// reset the same way m.viewer/m.files/m.roots's own sizes are untouched
	// by opening the Picker — carried over explicitly since newPicker()'s
	// fresh viewport.Model would otherwise start at 0x0 until the next
	// resize.
	previewWidth, previewHeight := m.picker.preview.Width(), m.picker.preview.Height()
	m.picker = newPicker()
	m.picker.open = true
	m.picker.seq = seq
	m.picker = m.picker.setPreviewSize(previewWidth, previewHeight)
	cmd := m.picker.input.Focus()
	return m, cmd
}

// dismissPicker closes the Picker, discarding its state entirely (S8): the
// pre-open frame — Active Document, Active Root, scroll positions — was
// never touched by an open Picker in the first place (nothing here reaches
// m.viewer, m.roots or m.files), so restoring it is just this: replace
// picker with its closed zero value.
// The one field that survives is seq, bumped rather than zeroed: a search
// dispatched before the dismissal is still running in its goroutine and
// will deliver its results afterwards, and resetting the counter would let
// those results be accepted by a later Picker that had counted its way back
// to the same number. It is a generation counter, not remembered state — it
// says nothing about mode, Scope or query, so S8 is untouched by carrying
// it.
func (m model) dismissPicker() model {
	seq := m.picker.seq + 1
	m.picker = newPicker()
	m.picker.seq = seq
	return m
}

// pickerSearchMsg is a matured debounce timer: the Prompt has been quiet for
// pickerSearchDebounce since generation seq was typed, so that generation's
// search may now be dispatched (M2.4). One is scheduled per change to the
// query, and all but the newest are dropped on arrival — see
// dispatchPickerSearch.
type pickerSearchMsg struct{ seq int }

// pickerResultsMsg is a completed search reporting back, carrying the
// generation it was dispatched for so a result set for a query the user has
// already typed past can be discarded (M2.4) rather than briefly displayed.
type pickerResultsMsg struct {
	seq     int
	results pickerResults
}

// queuePickerSearch is what any change to what would be searched (a
// keystroke now; a mode or Scope switch from M5.1/M5.2) calls: it opens a
// new generation and schedules that generation's debounce timer. It never
// searches — that is dispatchPickerSearch's job one debounce later — which
// is precisely how the Prompt stays responsive no matter how large the
// Scope (S1's "the Prompt never feels blocked").
//
// A sub-threshold Prompt is not a search (S12, pickerQueryMinLength): it
// clears results outright and schedules nothing, so deleting a query back
// below the threshold cannot leave the previous query's results on screen,
// and rg is never handed an empty pattern (which matches every line of every
// file — see grepFiles) nor a one-character one. A session without ripgrep
// (S14) is the same story: there is nothing to search on, so nothing is
// scheduled and the modal reports why instead (pickerStatus).
func (m model) queuePickerSearch() (model, tea.Cmd) {
	m.picker.seq++
	if !m.searchable() {
		m.picker.results = pickerResults{}
		m.picker.searching = false
		m.picker.selected = 0
		return m, nil
	}
	seq := m.picker.seq
	return m, tea.Tick(pickerSearchDebounce, func(time.Time) tea.Msg {
		return pickerSearchMsg{seq: seq}
	})
}

// dispatchPickerSearch runs a matured debounce timer's search, or drops the
// timer if its generation has been superseded (the user typed on) or the
// Picker has since been dismissed. This is the debounce: every keystroke
// schedules a timer, but only the last one of a burst still matches
// m.picker.seq when it matures, so a burst of n keystrokes dispatches one
// search rather than n (M2.4). The threshold and ripgrep gates (m.searchable)
// are re-checked here as well as in queuePickerSearch, so "no search for a
// sub-threshold query, none at all without rg" holds no matter which path
// scheduled the timer.
func (m model) dispatchPickerSearch(msg pickerSearchMsg) (model, tea.Cmd) {
	if !m.picker.open || msg.seq != m.picker.seq || !m.searchable() {
		return m, nil
	}
	m.picker.searching = true
	return m, pickerSearchCmd(m.picker.seq, m.picker.mode, m.picker.scope,
		m.picker.input.Value(), m.roots.rootList(), m.roots.active)
}

// resultCount is how many Result rows the current mode has to show right
// now (M3.1): the length of whichever of results.files/results.grep this
// mode populates. Used to clamp the selection cursor (move) and to size the
// Results content's windowedView — not what searchable/pickerStatus answer,
// which is about whether a search *should* run, not how many rows the last
// one that did produced.
func (p picker) resultCount() int {
	if p.mode == grepMode {
		return len(p.results.grep)
	}
	return len(p.results.files)
}

// move shifts the Results selection cursor by delta rows, clamped to the
// current mode's result count — the same moveSelection clamp filesNav.move
// and rootsNav.move already give their own lists (T16).
func (p picker) move(delta int) picker {
	p.selected = moveSelection(p.selected, delta, p.resultCount())
	return p
}

// setPreviewSize sizes the Preview pane's viewport, the same "pane content
// box" contract viewer.setSize gives m.viewer (layout.go/model.resize) — a
// value-receiver, T10-style method rather than a pointer one, even though
// viewport.Model's own Set* methods mutate through a pointer receiver
// (SetWidth/SetHeight take *Model): p is addressable here because it is a
// named parameter, the same reason promptView's SetWidth call works.
func (p picker) setPreviewSize(width, height int) picker {
	p.preview.SetWidth(max(width, 0))
	p.preview.SetHeight(max(height, 0))
	return p
}

// pickerPreviewCandidate is the Result the Preview pane should currently
// show (M3.2), or false when there is none: fileMode's selected row, gated
// by the same resultsCurrent check pickerStatus and pickerResultsContent
// already make — a stale generation's rows must not populate the Preview
// any more than they may flash in the Results column — and clamped against
// the live result count, since a fresh (shorter) result set can land
// between a selection move and the next render.
func (p picker) pickerPreviewCandidate() (pickerFile, bool) {
	if p.mode != fileMode || !p.resultsCurrent() {
		return pickerFile{}, false
	}
	if p.selected < 0 || p.selected >= len(p.results.files) {
		return pickerFile{}, false
	}
	return p.results.files[p.selected], true
}

// syncPreview loads the Preview pane's content for whichever Result
// pickerPreviewCandidate currently names, re-targeting (a fresh load,
// scrolled back to the top — M3.2's acceptance criterion, verbatim) only
// when that candidate's path actually changed since the last call: called
// after every mutation that can change the candidate (a Results move, a
// fresh or cleared result set landing), so the Preview always answers
// "what's selected right now" without re-reading a file on every keystroke
// or render that leaves the selection alone.
//
// The read is raw bytes, straight from disk — no loadDocument, no
// renderMarkdown: M3.2 is explicit that the Preview shows "raw, unrendered
// text", unlike the Document Viewer's own Glamour-rendered vp.
func (p picker) syncPreview() picker {
	f, ok := p.pickerPreviewCandidate()
	if !ok {
		if p.previewPath != "" {
			p.previewPath = ""
			p.preview.SetContent("")
		}
		return p
	}
	if f.path == p.previewPath {
		return p
	}
	src, err := os.ReadFile(f.path)
	if err != nil {
		src = []byte("preview unavailable: " + err.Error())
	}
	p.previewPath = f.path
	p.preview.SetContent(string(src))
	p.preview.GotoTop()
	return p
}

// resultsCurrent reports whether p.results still answers what the Prompt,
// mode and Scope now say — the same three-way label comparison pickerStatus
// makes to tell "searching" apart from a real empty result, pulled out here
// so the Results region's own content (picker_view.go) can make the exact
// same call: while a fresh search is in flight for a changed query/mode/
// Scope, the previous mode's stale rows must not flash on screen just
// because they are still sitting in p.results.
func (p picker) resultsCurrent() bool {
	return p.results.query == p.input.Value() && p.results.mode == p.mode && p.results.scope == p.scope
}

// searchable is whether the current Prompt may be searched at all: ripgrep
// present (S14 — it supplies both the file list and the text search, so
// without it there is no search to run) and the query at or above S12's
// threshold. Both search gates consult it, and pickerStatus reports the
// same two conditions to the user, so what the modal says and what the model
// does cannot drift apart.
func (m model) searchable() bool {
	return m.ripgrep.available && len([]rune(m.picker.input.Value())) >= pickerQueryMinLength
}

// pickerSearchCmd is the search itself, as a tea.Cmd: M1's blocking,
// rg-shelling search layer (listFiles/fuzzyFiles, grepFiles) run on Bubble
// Tea's own goroutine and reported back as a message. This is the codebase's
// first genuinely concurrent command (the $EDITOR handoff, model.go's
// startEdit, suspends the program instead), so it is deliberately shaped as
// a pure function of its arguments: everything it needs is copied in by
// value, and it touches no model state, because it runs while Update keeps
// processing keystrokes.
func pickerSearchCmd(seq int, mode pickerMode, scope pickerScope, query string, roots []root, active int) tea.Cmd {
	return func() tea.Msg {
		results := pickerResults{query: query, mode: mode, scope: scope}
		switch mode {
		case grepMode:
			results.grep, results.truncated = grepFiles(roots, active, scope, query)
		default:
			results.files, results.truncated = fuzzyFiles(listFiles(roots, active, scope), query)
		}
		return pickerResultsMsg{seq: seq, results: results}
	}
}

// applyPickerResults accepts a search's results only if they still answer
// the current generation — a stale set (the user typed on while rg was
// working) or one belonging to an already-dismissed Picker is dropped, never
// displayed (M2.4). Nothing else can filter these: a Cmd's goroutine cannot
// be cancelled once started, so discarding its message on arrival is the
// mechanism.
func (m model) applyPickerResults(msg pickerResultsMsg) model {
	if !m.picker.open || msg.seq != m.picker.seq {
		return m
	}
	m.picker.results = msg.results
	m.picker.searching = false
	m.picker.selected = 0
	m.picker = m.picker.syncPreview()
	return m
}

// handlePickerInput is where a keypress goes while the Picker is open
// (m.picker.open): S2's "always-typing" shape, the same one
// updateFindQuery already gives Find (find.go) — every key but a fixed set
// of non-typing ones falls through to m.picker.input.Update, typing the
// keystroke into the Prompt regardless of what it is bound to everywhere
// else (a query for "jog" must not move a selection or switch mode on its
// second letter).
//
// The non-typing set is exactly S2's list: Cancel dismisses; PickerUp/
// PickerDown now move the Results selection cursor (M3.1, picker.move);
// Activate, PageUp/PageDown/HalfPageUp/HalfPageDown (Preview scrolling,
// reused as-is from the Document Viewer's own scroll bindings — they are
// already non-printable, so nothing about "always typing" requires a second
// copy) and PickerMode/PickerScope (mode/Scope switch) are all recognised and
// swallowed here. Activate, Preview scrolling and mode/Scope switching still
// have no state to act on — M3.2/M3.3 give Activate and Preview their
// actions, M5.1/M5.2 give mode/Scope theirs — so today those keys do nothing
// beyond not typing (M2.3's original job, for the whole set).
func (m model) handlePickerInput(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Cancel):
		return m.dismissPicker(), nil
	case key.Matches(msg, keys.PickerUp):
		m.picker = m.picker.move(-1).syncPreview()
		return m, nil
	case key.Matches(msg, keys.PickerDown):
		m.picker = m.picker.move(1).syncPreview()
		return m, nil
	case key.Matches(msg, keys.PageUp):
		m.picker.preview.PageUp()
		return m, nil
	case key.Matches(msg, keys.PageDown):
		m.picker.preview.PageDown()
		return m, nil
	case key.Matches(msg, keys.HalfPageUp):
		m.picker.preview.HalfPageUp()
		return m, nil
	case key.Matches(msg, keys.HalfPageDown):
		m.picker.preview.HalfPageDown()
		return m, nil
	case key.Matches(msg, keys.Activate),
		key.Matches(msg, keys.PickerMode),
		key.Matches(msg, keys.PickerScope):
		return m, nil
	}
	before := m.picker.input.Value()
	next, cmd := m.picker.input.Update(msg)
	m.picker.input = next
	if m.picker.input.Value() == before {
		// A key the textinput ignored (or one that only moved its cursor)
		// changes nothing about what would be searched, so it must not open a
		// new generation — doing so would cancel an in-flight search for a
		// query that is still the current one.
		return m, cmd
	}
	m, searchCmd := m.queuePickerSearch()
	return m, tea.Batch(cmd, searchCmd)
}
