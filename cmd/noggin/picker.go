package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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
// from fuzzyFiles/grepPickerFiles.
type pickerResults struct {
	query        string
	mode         pickerMode
	scope        pickerScope
	hideDotfiles bool
	files        []pickerFile
	grep         []grepResult
	truncated    bool
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

	// previewLine is the 1-based line syncPreview last scrolled/highlighted
	// the Preview to for a Grep Result (M4.2, S9's "opens scrolled to the
	// matched line with that line highlighted"), or 0 for a File Result
	// (M3.2's plain "from the top") or when nothing is loaded. Paired with
	// previewPath as syncPreview's re-targeting key: pickerPreviewCandidate
	// (File mode's Activate/Preview candidate) never changes without the
	// path also changing, but pickerPreviewTarget's Grep candidate can name a
	// *different match in the same file* — a plain path comparison would
	// then wrongly treat that as "nothing changed" and leave the Preview
	// scrolled to the previous match.
	previewLine int

	// previewPath is the path whose bytes are currently loaded into
	// preview, or "" when nothing is. Compared against the current preview
	// target (pickerPreviewTarget) so a load only happens when the target's
	// file actually changed — not on every keystroke or render — and is
	// what "re-targets when the selection moves" (M3.2's acceptance
	// criterion) means concretely: a different path, a fresh load, GotoTop
	// (File mode) or a scroll/highlight to previewLine (Grep mode, M4.2).
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

// resetPicker is the shared "bump seq, snapshot preview size, replace with
// newPicker(), restore seq, restore preview size" sequence every full Picker
// state reset needs (D51/M2.1a): openPicker and dismissPicker both discard
// the whole picker struct, but two fields must survive that regardless —
// see dismissPicker's doc comment for why seq and the Preview viewport's
// width/height are carried over rather than reset with everything else.
// Before this was extracted, both callers duplicated the sequence verbatim,
// which is exactly the shape that let one of them (dismissPicker) ship
// without the preview-size carry-over until D48/M3.2a caught it — a third
// reset path (anticipated around M4.x's grep preview) now has nowhere to
// independently regress that fix; it calls this instead of repeating the
// sequence a third time.
func (m model) resetPicker() model {
	seq := m.picker.seq + 1
	previewWidth, previewHeight := m.picker.preview.Width(), m.picker.preview.Height()
	m.picker = newPicker()
	m.picker.seq = seq
	m.picker = m.picker.setPreviewSize(previewWidth, previewHeight)
	return m
}

// openPicker opens the Picker (S1): always File Picker mode, Active Root
// Scope, empty Prompt, regardless of what a previous open (in this or an
// earlier session) left behind — S1's fixed opening state and S8's "no
// memory" are the same requirement seen from two sides, so replacing the
// whole state with newPicker() (via resetPicker) satisfies both at once.
//
// Reachable from every Focus (S1), so unlike startFind/startEdit/activate,
// this never touches m.focus — opening the Picker is not "go to a pane",
// it's a modal layered on top of whichever pane already has focus, the same
// non-pane shape m.helpOpen has (model.go's openHelp).
func (m model) openPicker() (model, tea.Cmd) {
	// S7/M7.1: opening the Picker is one of Peek's clearing triggers.
	m = m.clearPeek()
	m = m.resetPicker()
	m.picker.open = true
	cmd := m.picker.input.Focus()
	return m, cmd
}

// dismissPicker closes the Picker, discarding its state entirely (S8): the
// pre-open frame — Active Document, Active Root, scroll positions — was
// never touched by an open Picker in the first place (nothing here reaches
// m.viewer, m.roots or m.files), so restoring it is just this: replace
// picker with its closed zero value.
// Two fields survive the reset, neither of them remembered Picker state:
//   - seq, bumped rather than zeroed: a search dispatched before the
//     dismissal is still running in its goroutine and will deliver its
//     results afterwards, and resetting the counter would let those results
//     be accepted by a later Picker that had counted its way back to the
//     same number. It is a generation counter — it says nothing about mode,
//     Scope or query, so S8 is untouched by carrying it.
//   - the Preview viewport's width/height, carried over the same way and
//     for the same reason openPicker carries them into its own reset: they
//     are a property of the terminal (model.resize), not of the closed
//     Picker's state. Without this (D48/M3.2a), a session that opens the
//     Picker, dismisses it and reopens it with no tea.WindowSizeMsg in
//     between — the common case, since dismissing never resizes the
//     terminal — would render an empty 0x0 Preview until the next resize,
//     because newPicker()'s fresh viewport.Model starts at 0x0 and nothing
//     would set it again.
//
// Both fields are carried over by resetPicker (D51/M2.1a), the same helper
// openPicker uses; this is a bare call to it, not a duplicated sequence.
func (m model) dismissPicker() model {
	return m.resetPicker()
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
// file — see grepPickerFiles) nor a one-character one. A session without ripgrep
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

// toggleHiddenFiles flips the session-scoped alt+h dotfile toggle
// (m.hideDotfiles — see its own doc comment on model for why it lives
// there and not here) and re-searches the current query against the new
// candidate set (files_search.go's listFiles/isDotfilePath), the same
// queuePickerSearch path switchPickerMode/switchPickerScope already reuse
// for their own in-place switches: the Prompt is untouched, and
// resultsCurrent's hideDotfiles comparison stales the previous candidate
// set's rows immediately, exactly as it already does for a mode or Scope
// switch.
func (m model) toggleHiddenFiles() (model, tea.Cmd) {
	m.hideDotfiles = !m.hideDotfiles
	return m.queuePickerSearch()
}

// switchPickerMode flips the open Picker's mode (File ↔ Grep, S7/M5.1):
// the Prompt's text is untouched — only p.mode changes — and the switch
// goes through queuePickerSearch exactly like a keystroke would (its own
// doc comment already names "a mode or Scope switch from M5.1/M5.2" as a
// caller), so the new mode's search for the same query follows the same
// debounce/generation contract as everything else: the old mode's results
// go stale immediately (resultsCurrent compares p.mode against
// results.mode) even before the new search reports back, so nothing from
// the previous mode can flash on screen while the new one is in flight.
// selected is left alone here for the same reason a query edit leaves it
// alone: applyPickerResults resets it once fresh results actually land,
// and resultsCurrent already hides the stale index in the meantime.
func (m model) switchPickerMode() (model, tea.Cmd) {
	if m.picker.mode == grepMode {
		m.picker.mode = fileMode
	} else {
		m.picker.mode = grepMode
	}
	return m.queuePickerSearch()
}

// switchPickerScope cycles the open Picker's Scope (Active Root -> All
// Roots -> Starred -> Active Root, S7/S8/M5.2/C4-M5.1), the exact same
// shape switchPickerMode gives mode (above): only p.scope changes, the
// Prompt is untouched, and the switch goes through queuePickerSearch so
// the new Scope's search for the same query follows the same debounce/
// generation contract as everything else — resultsCurrent's scope
// comparison stales the old Scope's rows immediately, the same way it
// stales an old mode's. S8 adds the third value (Starred) to what used to
// be a two-way flip; the fixed cycle order matches modeScopes'/browse.go's
// own convention of naming a scope's cycle order once rather than as a
// pairwise toggle per value.
func (m model) switchPickerScope() (model, tea.Cmd) {
	switch m.picker.scope {
	case activeRootScope:
		m.picker.scope = allRootsScope
	case allRootsScope:
		m.picker.scope = starredScope
	default:
		m.picker.scope = activeRootScope
	}
	return m.queuePickerSearch()
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
		m.picker.input.Value(), m.roots.rootList(), m.roots.active, m.hideDotfiles, m.docs)
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

// pickerPreviewCandidate is the Result Activate would open (M3.3) and, for
// fileMode, the Result the Preview pane shows (M3.2): fileMode's selected
// row, gated by the same resultsCurrent check pickerStatus and
// pickerResultsContent already make — a stale generation's rows must not
// populate the Preview any more than they may flash in the Results list
// — and clamped against the live result count, since a fresh (shorter)
// result set can land between a selection move and the next render.
//
// grepMode has no candidate here — M4.1 gave the Results list its own
// grepMode rows and M4.2 gives the Preview pane its own grepMode target
// (pickerPreviewTarget, below), and M4.3's grepMode Activate
// (activateGrepPicker) needs D34's extra ordinal-resolution data a plain
// pickerFile can't carry, so it has its own candidate lookup
// (pickerGrepCandidate) instead of sharing this one; this function's own
// contract ("what fileMode's Activate would open") is why it must not
// start returning a grepMode candidate.
func (p picker) pickerPreviewCandidate(hideDotfiles bool) (pickerFile, bool) {
	if p.mode != fileMode || !p.resultsCurrent(hideDotfiles) {
		return pickerFile{}, false
	}
	if p.selected < 0 || p.selected >= len(p.results.files) {
		return pickerFile{}, false
	}
	return p.results.files[p.selected], true
}

// pickerPreviewTarget is what the Preview pane should currently show
// (M3.2, M4.2): the candidate file, plus — for a Grep Result only — the
// 1-based line syncPreview should scroll to and highlight (S9's "the Grep
// Picker opens scrolled to the matched line with that line highlighted").
// line is 0 for a File Result (previews from the top instead, M3.2's own
// contract, unchanged) and for "nothing selected". Unlike
// pickerPreviewCandidate, this covers both modes — the Preview pane itself
// has no reason to wait for M4.3's Activate hand-off the way Activate
// does, since showing the matched file's raw bytes never changes the
// Active Document (S9's own "never changes the Active Document").
func (p picker) pickerPreviewTarget(hideDotfiles bool) (f pickerFile, line int, ok bool) {
	if !p.resultsCurrent(hideDotfiles) {
		return pickerFile{}, 0, false
	}
	if p.mode == grepMode {
		if p.selected < 0 || p.selected >= len(p.results.grep) {
			return pickerFile{}, 0, false
		}
		g := p.results.grep[p.selected]
		return g.file, g.line, true
	}
	if p.selected < 0 || p.selected >= len(p.results.files) {
		return pickerFile{}, 0, false
	}
	return p.results.files[p.selected], 0, true
}

// syncPreview loads the Preview pane's content for whichever Result
// pickerPreviewTarget currently names, re-targeting only when that target
// actually changed since the last call: called after every mutation that
// can change it (a Results move, a fresh or cleared result set landing, a
// mode switch), so the Preview always answers "what's selected right now"
// without re-reading a file or re-scrolling on every keystroke or render
// that leaves the selection alone.
//
// previewPath and previewLine together are that "did it actually change"
// key: comparing path alone would miss a Grep selection moving to a
// *different match in the same file* (previewLine is what catches that;
// see previewLine's own doc comment), and a fresh load only happens when
// the path itself changed — a same-file, different-line Grep move re-uses
// the bytes already loaded and only re-scrolls/re-highlights.
//
// The read is raw bytes, straight from disk — no loadDocument, no
// renderMarkdown: M3.2 is explicit that the Preview shows "raw, unrendered
// text", unlike the Document Viewer's own Glamour-rendered vp.
func (p picker) syncPreview(hideDotfiles bool) picker {
	f, line, ok := p.pickerPreviewTarget(hideDotfiles)
	if !ok {
		if p.previewPath != "" {
			p.previewPath = ""
			p.previewLine = 0
			p.preview.SetContent("")
			p.preview.StyleLineFunc = nil
		}
		return p
	}
	if f.path == p.previewPath && line == p.previewLine {
		return p
	}
	if f.path != p.previewPath {
		src, err := os.ReadFile(f.path)
		if err != nil {
			src = []byte("preview unavailable: " + err.Error())
		}
		p.previewPath = f.path
		p.preview.SetContent(string(src))
	}
	p.previewLine = line
	if line <= 0 {
		p.preview.StyleLineFunc = nil
		p.preview.GotoTop()
		return p
	}
	// rg's line numbers are 1-based (S4/grepResult); the viewport's own line
	// indices, like every other line-addressing surface in this codebase
	// (find.go's findMatch), are 0-based.
	idx := line - 1
	p.preview.SetYOffset(idx)
	p.preview.StyleLineFunc = func(i int) lipgloss.Style {
		if i == idx {
			return lipgloss.NewStyle().Reverse(true)
		}
		return lipgloss.NewStyle()
	}
	return p
}

// resultsCurrent reports whether p.results still answers what the Prompt,
// mode and Scope now say — the same three-way label comparison pickerStatus
// makes to tell "searching" apart from a real empty result, pulled out here
// so the Results list's own content (picker_view.go) can make the exact
// same call: while a fresh search is in flight for a changed query/mode/
// Scope, the previous mode's stale rows must not flash on screen just
// because they are still sitting in p.results.
func (p picker) resultsCurrent(hideDotfiles bool) bool {
	return p.results.query == p.input.Value() && p.results.mode == p.mode && p.results.scope == p.scope &&
		p.results.hideDotfiles == hideDotfiles
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
// rg-shelling search layer (listFiles/fuzzyFiles, grepPickerFiles) run on
// Bubble Tea's own goroutine and reported back as a message. This is the codebase's
// first genuinely concurrent command (the $EDITOR handoff, model.go's
// startEdit, suspends the program instead), so it is deliberately shaped as
// a pure function of its arguments: everything it needs is copied in by
// value, and it touches no model state, because it runs while Update keeps
// processing keystrokes. docs is the one exception to "copied in by value"
// in spirit rather than letter (S8/C4-M5.1, starredScope's own candidate
// set, starredPickerFiles): the S1 docs map is passed by reference like
// roots is passed by slice header, but model.go's own copy-on-write
// discipline (withDocRecord always builds a fresh map before replacing
// m.docs, never mutates one in place) makes it as safe to read from a
// concurrent goroutine as roots already is.
func pickerSearchCmd(seq int, mode pickerMode, scope pickerScope, query string, roots []root, active int, hideDotfiles bool, docs map[string]docRecord) tea.Cmd {
	return func() tea.Msg {
		results := pickerResults{query: query, mode: mode, scope: scope, hideDotfiles: hideDotfiles}
		var files []pickerFile
		if scope == starredScope {
			// S8: the whole store, not a walk of any Root's tree —
			// starredPickerFiles resolves Scope on its own terms, so
			// listFiles' own roots-shaped contract (liveRootsInScope)
			// never sees starredScope at all.
			files = starredPickerFiles(docs, roots, hideDotfiles)
		} else {
			files = listFiles(roots, active, scope, hideDotfiles)
		}
		switch mode {
		case grepMode:
			results.grep, results.truncated = grepPickerFiles(files, query)
		default:
			results.files, results.truncated = fuzzyFiles(files, query)
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
	m.picker = m.picker.syncPreview(m.hideDotfiles)
	return m
}

// activatePicker is S10's Activate: the currently-previewed Result becomes
// the Active Document, the same and only path openDocument already gives
// activate() (Files Nav) and the launch-time resolution (applyResolution)
// — it touches m.viewer/m.focus alone, never m.roots or m.files, which is
// what "does not change the Active Root, and does not move, expand or
// re-select the Files Nav tree" means concretely: a Result's Root need not
// be the Active Root (S10), and this never calls openFolder or touches
// m.roots.active to make it so.
//
// grepMode is delegated to activateGrepPicker (M4.3): a Grep Result's
// activation additionally has to land the Document Viewer on the matched
// line (D34), which fileMode's plain openDocument has no reason to do.
func (m model) activatePicker() model {
	return m.activatePickerInto(dv1Pane)
}

// activatePickerInto is activatePicker's targeted form (C3/S3, M1.5): the
// Picker's own home for the `ctrl+` targeting modifier the Files Nav
// already has (M1.4) — keys.ActivateDV2 (ctrl+enter, ctrl+l fallback, D55)
// loads the previewed Result into DV2 instead of DV1, creating DV2 first if
// absent, and moves Focus there (openDocumentInto's own behavior, S3);
// plain Activate keeps targeting DV1 via the dv1Pane wrapper above. No
// Picker-specific carve-out is needed for handlePickerInput's always-typing
// Prompt (TRIAGE T3, closed structurally by D55): both of ActivateDV2's
// chords are raw control bytes, matched in handlePickerInput before the
// fallthrough to the Prompt, so a literal `l`/`L` keystroke there is
// unaffected.
func (m model) activatePickerInto(target pane) model {
	if m.picker.mode == grepMode {
		return m.activateGrepPickerInto(target)
	}
	// Reuses pickerPreviewCandidate rather than a second selection lookup —
	// "the Result the Preview is showing" and "the Result Activate would
	// open" are the same question (M3.2's own handoff note), so a fresh
	// generation's rows can no more be activated than they can leak into
	// the Preview.
	//
	// A read error surfaces as a footer message rather than refusing to
	// close the modal — activate()'s own "silence is not an error" shape
	// (T15) — the Picker is dismissed either way since there is nothing
	// further for it to do once a Result has been chosen.
	f, ok := m.picker.pickerPreviewCandidate(m.hideDotfiles)
	if !ok {
		return m
	}
	d, err := loadDocument(f.path)
	if err != nil {
		m.message = fmt.Sprintf("cannot open %s: %v", f.path, err)
		return m.dismissPicker()
	}
	// S4/M3.3: Picker Activate records recency the same as the Files Nav's
	// own activateInto (M3.1) — TRIAGE T12 found this call site bypassing
	// recordViewed entirely by going straight to openDocumentInto.
	m = m.recordViewed(f.path)
	// S9/C4/M6.1: the Picker's Activate is a real navigation, so it records
	// onto the target viewer's own back/forward stack the same way the
	// Files Nav's activateInto does (model.navigate).
	m = m.navigate(target, d)
	return m.dismissPicker()
}

// togglePickerStar flips the S1 document store's star mark for the
// Picker's currently selected Result (S2's third starring surface,
// M2.5) — whichever mode is open, fileMode's pickerPreviewCandidate or
// grepMode's pickerGrepCandidate, both already gated on resultsCurrent so
// a stale generation's row can no more be starred than it may be previewed
// or activated. Unlike the Files Nav's Starred-list row (M4.3, un-starring
// removes it immediately), S2 is explicit that un-starring from the Picker
// "leaves the Result in place with its marker cleared": this touches only
// the S1 document store via docRecord/withDocRecord, never m.picker itself,
// so the Results list, the selection cursor and the Preview are all left
// exactly where they were — there is nothing here that could move them.
// A no-op with nothing selected (empty Results, no search run yet, or a
// stale generation) is silent, the same shape activatePickerInto's own
// "not ok" case already gives.
func (m model) togglePickerStar() model {
	path, ok := m.pickerSelectedPath()
	if !ok {
		return m
	}
	rec := m.docRecord(path)
	rec.Starred = !rec.Starred
	return m.withDocRecord(path, rec)
}

// pickerSelectedPath is the path named by the Picker's currently selected
// Result, in whichever mode is open — the one piece togglePickerStar needs
// that pickerPreviewCandidate (fileMode) and pickerGrepCandidate (grepMode)
// do not already share directly, since Star (unlike Activate) does not
// otherwise care which mode produced the Result or need D34's extra
// ordinal-resolution data pickerGrepCandidate's grepResult also carries.
func (m model) pickerSelectedPath() (string, bool) {
	if m.picker.mode == grepMode {
		g, ok := m.picker.pickerGrepCandidate(m.hideDotfiles)
		if !ok {
			return "", false
		}
		return g.file.path, true
	}
	f, ok := m.picker.pickerPreviewCandidate(m.hideDotfiles)
	if !ok {
		return "", false
	}
	return f.path, true
}

// pickerGrepCandidate is the Grep Result Activate would open (M4.3): the
// selected row, gated the same way pickerPreviewCandidate gates fileMode's
// candidate — resultsCurrent (a stale generation's rows must not activate
// any more than they may preview or render) and the live result count (a
// fresh, shorter result set can land between a selection move and the next
// render).
func (p picker) pickerGrepCandidate(hideDotfiles bool) (grepResult, bool) {
	if p.mode != grepMode || !p.resultsCurrent(hideDotfiles) {
		return grepResult{}, false
	}
	if p.selected < 0 || p.selected >= len(p.results.grep) {
		return grepResult{}, false
	}
	return p.results.grep[p.selected], true
}

// activateGrepPicker is activatePicker's grepMode counterpart (M4.3, S10,
// S11): the matched file becomes the Active Document, and the Document
// Viewer lands on the matched line — re-derived in the rendered content's
// own line space rather than trusted from rg's raw-file line number, per
// D34: Glamour's rendering (re-flow, frontmatter stripping (D5), tab
// expansion, markdown syntax removed) does not preserve rg's line numbers
// across that boundary.
//
// The hand-off reuses find.go's own machinery (commitQuery, applyCurrentMatch)
// rather than a bespoke scroll — D34's own reasoning, and what lets M4.4's
// next/previous traversal fall out of this for free rather than needing its
// own wiring: commitQuery already puts the Document Viewer into exactly the
// "committed find" state FindNext/FindPrevious already know how to walk.
// commitQuery's own nearestMatch(matches, 0) default ("first match from the
// top, or none") is D34's degrade path verbatim, so the current match is
// only overridden here when resolveGrepLanding actually resolves an ordinal
// — never left in a half-set state, and never a failed or crashed
// activation either way.
//
// A read error surfaces as a footer message the same way activatePicker's
// fileMode half does (T15) — the Picker is dismissed either way.
func (m model) activateGrepPicker() model {
	return m.activateGrepPickerInto(dv1Pane)
}

// activateGrepPickerInto is activateGrepPicker's targeted form (C3/S3,
// M1.5), the grepMode half of activatePickerInto: the same DV2-vs-DV1
// choice, threaded through the pane-generic viewerFor/withViewerFor
// accessors (model.go) rather than m.dv1 directly — target is an explicit
// argument here, never "whichever pane has Focus", so the focused-viewer
// accessor (which answers that different question) still does not apply
// either way.
func (m model) activateGrepPickerInto(target pane) model {
	g, ok := m.picker.pickerGrepCandidate(m.hideDotfiles)
	if !ok {
		return m
	}
	query := m.picker.input.Value()
	d, err := loadDocument(g.file.path)
	if err != nil {
		m.message = fmt.Sprintf("cannot open %s: %v", g.file.path, err)
		return m.dismissPicker()
	}
	// S4/M3.3: same recordViewed fix as activatePickerInto's fileMode half,
	// for the grepMode path (TRIAGE T12).
	m = m.recordViewed(g.file.path)
	// S9/C4/M6.1: same history-recording navigation as fileMode's own
	// activatePickerInto above.
	m = m.navigate(target, d)
	v := m.viewerFor(target).commitQuery(query)
	if ordinal, ok := resolveGrepLanding(d.src, v.vp.GetContent(), query, g.line); ok {
		v.findState.current = ordinal
		v = v.applyCurrentMatch()
	}
	m = m.withViewerFor(target, v)
	return m.dismissPicker()
}

// resolveGrepLanding is D34's occurrence-ordinal bridge from a Grep
// Result's raw-file line number to a match index into
// findMatches(rendered, query): "the Nth literal occurrence of the query"
// is a property of the text itself, invariant to how it later gets wrapped
// onto lines, so the Nth occurrence in the raw source (after the same
// frontmatter-stripping renderMarkdown itself applies, D5 — line numbers
// past a stripped frontmatter block do not otherwise line up between the
// two spaces at all) is assumed to be the Nth occurrence in the rendered
// content too. ok is false — D34's explicit degrade trigger — when the
// target line sits inside the stripped frontmatter block itself (no
// corresponding occurrence in the rendered content can exist), or when the
// resolved ordinal has no counterpart at all in the rendered matches
// (Glamour having removed or reordered occurrences before it, e.g. inside
// stripped markdown syntax).
//
// "This match" resolves to the *first* occurrence of query on line: rg
// reports one Result per matching line, not per occurrence within it
// (grepBatch's own comment), so a line with several occurrences of query
// picks its first as the line's representative match — the same
// approximation D34 accepts for markdown-syntax-adjacent matches.
func resolveGrepLanding(src []byte, rendered, query string, line int) (ordinal int, ok bool) {
	stripped := stripFrontmatter(src)
	if len(stripped) != len(src) {
		removedLines := strings.Count(string(src[:len(src)-len(stripped)]), "\n")
		if line <= removedLines {
			return 0, false
		}
		line -= removedLines
	}
	raw := findMatches(string(stripped), query)
	ordinal = -1
	for i, m := range raw {
		if m.line+1 == line { // findMatch.line is 0-based; rg's line is 1-based
			ordinal = i
			break
		}
	}
	if ordinal < 0 {
		return 0, false
	}
	rend := findMatches(rendered, query)
	if ordinal >= len(rend) {
		return 0, false
	}
	return ordinal, true
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
// PickerDown move the Results selection cursor (M3.1, picker.move);
// PageUp/PageDown/HalfPageUp/HalfPageDown scroll the Preview (reused as-is
// from the Document Viewer's own scroll bindings — they are already
// non-printable, so nothing about "always typing" requires a second copy);
// Activate opens the previewed Result and dismisses the Picker (M3.3,
// activatePicker); keys.ActivateDV2 (ctrl+enter, ctrl+l fallback) is
// Activate's C3/S3/M1.5 targeting twin, routing the same open-and-dismiss
// into DV2 (activatePickerInto) — both its chords are raw control bytes
// (D55), so neither is ever mistaken for typed query text; PickerMode and
// PickerScope switch mode and Scope in place (M5.1 switchPickerMode, M5.2
// switchPickerScope). All are recognised and handled here rather than
// falling through to the Prompt.
func (m model) handlePickerInput(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Cancel):
		return m.dismissPicker(), nil
	case key.Matches(msg, keys.PickerUp):
		m.picker = m.picker.move(-1).syncPreview(m.hideDotfiles)
		return m, nil
	case key.Matches(msg, keys.PickerDown):
		m.picker = m.picker.move(1).syncPreview(m.hideDotfiles)
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
	case key.Matches(msg, keys.ActivateDV2):
		return m.activatePickerInto(dv2Pane), nil
	case key.Matches(msg, keys.Activate):
		return m.activatePicker(), nil
	case key.Matches(msg, keys.PickerMode):
		return m.switchPickerMode()
	case key.Matches(msg, keys.PickerScope):
		return m.switchPickerScope()
	case key.Matches(msg, keys.ToggleHidden):
		return m.toggleHiddenFiles()
	case key.Matches(msg, keys.PickerStar):
		return m.togglePickerStar(), nil
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
