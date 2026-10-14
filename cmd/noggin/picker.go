package main

import (
	"errors"
	"os"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"noggin/internal/tree"
	"noggin/internal/wikipedia"
)

// pickerMode selects which kind of search a Picker runs (S1, GLOSSARY.md
// "File Picker"/"Grep Picker"/"Wikipedia Picker"): fuzzy file-name matching,
// literal grep, or a Wikipedia title search. The Picker always opens in
// fileMode, unless the Active Source is Wikipedia, in which case it opens
// in wikipediaMode instead (S2, C6/M3.2 — revises S1's old "always File"
// default); the in-Picker switch (M5.1, C6/M3.2) cycles all three, File →
// Grep → Wikipedia, regardless of which Source is active.
type pickerMode int

const (
	fileMode pickerMode = iota
	grepMode
	wikipediaMode
)

// pickerSearchDebounce is how long a changed Prompt settles before its
// search is actually dispatched (M2.4), for File and Grep mode. A keystroke
// only ever schedules a timer; the search itself runs in a tea.Cmd goroutine
// once no newer keystroke has superseded it (dispatchPickerSearch's seq
// check), so fast typing issues far fewer searches than keystrokes and the
// Prompt never waits on rg.
//
// 120ms is a build-time judgment call, not a spec'd number: long enough
// that an average typing burst collapses into one search, short enough that
// a pause between words still feels immediate.
const pickerSearchDebounce = 120 * time.Millisecond

// pickerWikipediaSearchDebounce is Wikipedia mode's own, longer debounce
// (S13, C6/M3.1): "Wikipedia mode waits somewhat longer than local modes
// since each search is a network call", unlike File/Grep's local rg shell-
// out. Also a build-time judgment call, not a spec'd number.
const pickerWikipediaSearchDebounce = 300 * time.Millisecond

// pickerDebounceFor is the per-mode debounce queuePickerSearch schedules
// (S13, C6/M3.1): every mode but Wikipedia shares pickerSearchDebounce;
// Wikipedia gets its own, longer pickerWikipediaSearchDebounce. Pulled out
// as its own function, rather than inlined at the one call site, so the
// "Wikipedia waits longer" half of S13 is directly testable without driving
// a real tea.Tick.
func pickerDebounceFor(mode pickerMode) time.Duration {
	if mode == wikipediaMode {
		return pickerWikipediaSearchDebounce
	}
	return pickerSearchDebounce
}

// pickerQueryMinLength is the Picker's query threshold: neither Picker
// searches until at least this many characters are typed, and below it the
// Picker says so instead (pickerStatus, M2.5). Enforced at both search gates
// (queuePickerSearch and dispatchPickerSearch) for the same reason the
// empty-Prompt refusal it replaces was: whichever path scheduled a timer,
// no search may escape for a sub-threshold query.
//
// The number was originally S12/C2's own ("at least 3 characters", cycle C2);
// C6/S13 revises it down to 2 (C2/D31), on the same one-character-queries-
// are-too-broad reasoning, now weighed against 2 being the shortest useful
// Wikipedia title prefix.
const pickerQueryMinLength = 2

// pickerWikipediaResultLimit bounds a Wikipedia title search the same way
// S2 asks ("a bounded list"): unlike pickerResultCap (200, sized for a local
// fuzzy/grep match list that also reports "truncated"), this is deliberately
// small — a title-search Results list is meant to be scanned like Wikipedia's
// own search box, not paged through. A build-time judgment call, not a
// spec'd number, the same status pickerSearchDebounce's own comment gives.
const pickerWikipediaResultLimit = 20

// pickerWikipediaClient is the seam Wikipedia mode's search runs against
// (S2, C6/M3.3): whatever satisfies it — model's default, a real
// *wikipedia.Client held for the session so a later article fetch (M4.2)
// can share its cache, or a test fake — pickerSearchCmd's wikipediaMode
// case calls Search on. Kept to exactly what a title search needs, not the
// whole *wikipedia.Client, so a fake never has to also implement Fetch to
// satisfy it.
type pickerWikipediaClient interface {
	Search(query string, limit int) ([]wikipedia.SearchResult, error)
}

// pickerResults is one completed search's answer, labelled with the query,
// mode and Scope it answers *for*. The labels are the point: results arrive
// asynchronously, so the only way the modal can render honestly is for a
// result set to carry what it was a search for rather than being implicitly
// assumed to match whatever the Prompt now says (S12/C2's status/empty states,
// M2.5, read these labels; M5.1/M5.2 make mode and Scope change under an
// in-flight search).
//
// files, grep and wikipedia are the three modes' shapes; exactly one is
// ever populated, selected by mode (C6/M3.3 adds wikipedia to M4.1's
// original two). truncated is S12/C2's cap flag, passed straight through from
// fuzzyFiles/grepPickerFiles; a Wikipedia search has no such flag of its
// own (pickerWikipediaResultLimit's cap is silent, the way Wikipedia's own
// search box is).
type pickerResults struct {
	query        string
	mode         pickerMode
	scope        pickerScope
	hideDotfiles bool
	filter       tree.Filter
	files        []pickerFile
	grep         []grepResult
	wikipedia    []wikipedia.SearchResult
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
	// filter is the File Filter (D140) this Picker searches under, taken from
	// the model when it opens (model.pickerFilter) and changed only by alt+a
	// (togglePickerFilter), so the Picker can label and compare its Results
	// without being handed the model.
	filter tree.Filter
	input  textinput.Model

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
	// nothing" and "still looking", which S12/C2's empty states (M2.5) must
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

	// preview is the Preview pane's own viewport (M3.2): the file's raw
	// source laid out by the Text renderer (D140: soft wrap, line-number
	// gutter), whatever the file is — no Glamour, unlike the Document
	// Viewer's own vp (viewer.go) — scrolled independently of the Prompt/Results. Sized
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

	// previewCol and previewQuery are the rest of syncPreview's re-targeting
	// key for a Grep Result (previewLine's siblings): the byte column ripgrep
	// reported and the query that found it, which together pick the exact
	// span highlighted. Both are zero for a File Result.
	previewCol   int
	previewQuery string

	// previewSrc is the loaded file's bytes (or, with previewPlain, a short
	// explanatory message) and previewLayout their Text layout at the
	// Preview's current width, nil until built or after a resize
	// invalidates it (D140: the Preview uses the Text renderer). previewSpan
	// is the match span currently drawn reversed in it, so a move to another
	// match in the same file re-styles only two lines.
	previewSrc    []byte
	previewPlain  bool
	previewLayout *textLayout
	previewSpan   *textSpan

	// wikipediaErr is Wikipedia mode's own failure state (S2, C6/M3.3): the
	// most recent search's error text ("offline"/"service error"), or ""
	// when its last attempt (if any) succeeded. pickerStatus shows it in
	// place of the usual searching/no-results/truncated captions, and
	// applyPickerResults sets it instead of clearing p.results on a failed
	// search — S2's "leaves the last Results in place". queuePickerSearch
	// clears it on every new attempt (a keystroke, a mode/Scope switch), the
	// same "goes stale the instant something changes" rule results.query's
	// own mismatch already gives the Results list — so a leftover error
	// from a since-abandoned query is never shown once a fresh one is
	// underway.
	wikipediaErr string
}

// newPicker is the closed, no-search state: nothing open, nothing typed,
// File Picker mode and Active Source Scope (C6/M1.3's own label for File
// Picker's activeRootScope) — the zero values of pickerMode and
// pickerScope already are S1's fixed opening state, so there is nothing to
// set beyond a fresh Prompt.
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

// openPicker opens the Picker: File Picker mode and Active Source Scope,
// empty Prompt, regardless of what a previous open (in this or an earlier
// session) left behind — S1's fixed opening state and S8's "no memory" are
// the same requirement seen from two sides, so replacing the whole state
// with newPicker() (via resetPicker) satisfies both at once. The one
// exception is mode: S2/C6/M3.2 lands on wikipediaMode instead whenever
// Wikipedia is the Active Source, so opening the Picker never requires a
// manual mode switch to reach the Source that is already active. Scope
// stays Active Source Scope regardless — wikipediaMode's own Scope reads
// "-" (pickerModeScope) and ignores it.
//
// Reachable from every Focus (S1), so unlike startFind/startEdit/activate,
// this never touches m.focus — opening the Picker is not "go to a pane",
// it's a modal layered on top of whichever pane already has focus, the same
// non-pane shape m.helpOpen has (model.go's openHelp).
func (m model) openPicker() (model, tea.Cmd) {
	// S7/M7.1: opening the Picker is one of Peek's clearing triggers.
	m = m.clearPeek()
	m = m.resetPicker()
	if m.roots.activeIsWikipedia() {
		m.picker.mode = wikipediaMode
	}
	m.picker.filter = m.pickerFilter()
	m.picker.open = true
	cmd := m.picker.input.Focus()
	return m, cmd
}

// dismissPicker closes the Picker, discarding its state entirely (S8): the
// pre-open frame — Active Document, Active Source, scroll positions — was
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
//
// The one thing that does outlive a dismissal is m.lastPicker (see
// lastPickerState): a non-empty Prompt's query, mode and Scope are captured
// into it here, before the reset, so keys.PickerRestore (restorePicker) can
// reopen to them. A dismissal with an empty Prompt leaves it untouched.
func (m model) dismissPicker() model {
	if q := m.picker.input.Value(); q != "" {
		m.lastPicker = &lastPickerState{query: q, mode: m.picker.mode, scope: m.picker.scope}
	}
	return m.resetPicker()
}

// lastPickerState is what a dismissed Picker leaves behind for
// keys.PickerRestore: the typed query plus the mode and Scope it was
// searching. Scope is stored as its kind (activeRootScope/allRootsScope/
// starredScope), not a resolved path, so it re-resolves against whatever is
// the Active Source at restore time. Session-only: it lives on model, never
// persisted, and a plain openPicker neither reads nor clears it.
type lastPickerState struct {
	query string
	mode  pickerMode
	scope pickerScope
}

// restorePicker opens the Picker to m.lastPicker's query, mode and Scope and
// searches it immediately (queuePickerSearch, so debounce/generation rules
// are the usual ones; the Results cursor starts at 0 and the Prompt cursor
// at the end of the query). With nothing saved it is exactly openPicker.
//
// The saved mode wins over openPicker's "land on Wikipedia mode when
// Wikipedia is the Active Source" rule. One fix-up mirrors
// switchPickerMode's: File/Grep mode with Wikipedia active must not carry
// activeRootScope, since there is no Active Root to search.
func (m model) restorePicker() (model, tea.Cmd) {
	m, cmd := m.openPicker()
	if m.lastPicker == nil {
		return m, cmd
	}
	last := *m.lastPicker
	m.picker.mode = last.mode
	m.picker.scope = last.scope
	if m.picker.mode != wikipediaMode && m.roots.activeIsWikipedia() && m.picker.scope == activeRootScope {
		m.picker.scope = allRootsScope
	}
	m.picker.input.SetValue(last.query)
	m.picker.input.CursorEnd()
	m, searchCmd := m.queuePickerSearch()
	return m, tea.Batch(cmd, searchCmd)
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
// err is Wikipedia mode's own addition (S2, C6/M3.3): a title search can
// fail (offline, service error) where a local File/Grep search never does,
// and results then carries only the labels (query/mode/scope/hideDotfiles)
// the failed search was for, never any rows — applyPickerResults's err
// branch is what keeps the previous Results in place instead.
type pickerResultsMsg struct {
	seq     int
	results pickerResults
	err     error
}

// queuePickerSearch is what any change to what would be searched (a
// keystroke now; a mode or Scope switch from M5.1/M5.2) calls: it opens a
// new generation and schedules that generation's debounce timer. It never
// searches — that is dispatchPickerSearch's job one debounce later — which
// is precisely how the Prompt stays responsive no matter how large the
// Scope (S1's "the Prompt never feels blocked").
//
// A sub-threshold Prompt is not a search (S13, pickerQueryMinLength): it
// clears results outright and schedules nothing, so deleting a query back
// below the threshold cannot leave the previous query's results on screen,
// and rg is never handed an empty pattern (which matches every line of every
// file — see grepPickerFiles) nor a one-character one. A session without ripgrep
// (S14) is the same story: there is nothing to search on, so nothing is
// scheduled and the modal reports why instead (pickerStatus).
func (m model) queuePickerSearch() (model, tea.Cmd) {
	m.picker.seq++
	// C6/M3.3: any new attempt — typed or a mode/Scope switch — invalidates
	// a leftover Wikipedia search error the same instant it invalidates the
	// previous Results labels (see wikipediaErr's own doc comment).
	m.picker.wikipediaErr = ""
	if !m.searchable() {
		m.picker.results = pickerResults{}
		m.picker.searching = false
		m.picker.selected = 0
		return m, nil
	}
	seq := m.picker.seq
	return m, tea.Tick(pickerDebounceFor(m.picker.mode), func(time.Time) tea.Msg {
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
//
// wikipediaMode is the one exception (TRIAGE T21): pickerSearchCmd's own
// wikipediaMode branch never calls listFiles, so hideDotfiles has no
// bearing on a Wikipedia search's results at all — queuePickerSearch would
// still fire a real, debounced network request with no visible effect. The
// toggle itself still flips (for whenever the Picker next switches to
// File/Grep), but instead of requerying, p.results.hideDotfiles is patched
// to match it directly — resultsCurrent's label comparison then still
// reports the existing, still-valid Wikipedia rows as current, so they
// stay on screen rather than reading as stale (which the view would show
// as "searching…" for a search that was never actually dispatched).
func (m model) toggleHiddenFiles() (model, tea.Cmd) {
	m.hideDotfiles = !m.hideDotfiles
	if m.picker.mode == wikipediaMode {
		m.picker.results.hideDotfiles = m.hideDotfiles
		return m, nil
	}
	return m.queuePickerSearch()
}

// togglePickerFilter is alt+a: it flips the File Filter (D140) and re-searches.
// With an Active Root it is that Root's filter (the Files Nav's `a` state, so
// the Tree is re-walked too); with Wikipedia active there is no Root and it
// flips the most recently set value instead. Wikipedia mode and the Starred
// scope ignore the filter, so like toggleHiddenFiles they relabel the rows on
// screen rather than fire a search whose results could not differ.
func (m model) togglePickerFilter() (model, tea.Cmd) {
	next := m.picker.filter.Toggled()
	if m.files.root == nil {
		m.lastFilter = next
	} else {
		var ok bool
		if m, ok = m.setActiveRootFilter(next); !ok {
			return m, nil
		}
	}
	return m.applyPickerFilter(next)
}

// applyPickerFilter puts next on the open Picker and refreshes its Results.
func (m model) applyPickerFilter(next tree.Filter) (model, tea.Cmd) {
	m.picker.filter = next
	if m.picker.mode == wikipediaMode || m.picker.scope == starredScope {
		m.picker.results.filter = next
		return m, nil
	}
	return m.queuePickerSearch()
}

// switchPickerMode cycles the open Picker's mode File → Grep → Wikipedia →
// File (S2, S7/M5.1; C6/M3.2 extends the old File ↔ Grep two-way flip to a
// three-way cycle, reachable from any Active Source — the cycle itself
// does not read m.roots at all): the Prompt's text is untouched — only
// p.mode changes — and the switch goes through queuePickerSearch exactly
// like a keystroke would (its own doc comment already names "a mode or
// Scope switch from M5.1/M5.2" as a caller), so the new mode's search for
// the same query follows the same debounce/generation contract as
// everything else: the old mode's results go stale immediately
// (resultsCurrent compares p.mode against results.mode) even before the
// new search reports back, so nothing from the previous mode can flash on
// screen while the new one is in flight. selected is left alone here for
// the same reason a query edit leaves it alone: applyPickerResults resets
// it once fresh results actually land, and resultsCurrent already hides
// the stale index in the meantime.
func (m model) switchPickerMode() (model, tea.Cmd) {
	switch m.picker.mode {
	case fileMode:
		m.picker.mode = grepMode
	case grepMode:
		m.picker.mode = wikipediaMode
	default:
		m.picker.mode = fileMode
	}
	// S6/C6/M3.5: landing back on File or Grep mode with Wikipedia still the
	// Active Source must never land on Active Root Scope — there is no
	// Active Root to search, and S6 explicitly rules out offering that
	// "empty dead-end scope" at all, not just making it escapable with one
	// more Scope press. The only way p.scope can still read activeRootScope
	// here is the fixed zero-value state a fresh openPicker leaves it in
	// (S1's own opening state, unconditional regardless of Active Source);
	// switchPickerScope's own Wikipedia-active branch is what keeps it out
	// from here on once this has run once.
	if m.picker.mode != wikipediaMode && m.roots.activeIsWikipedia() && m.picker.scope == activeRootScope {
		m.picker.scope = allRootsScope
	}
	return m.queuePickerSearch()
}

// switchPickerScope cycles the open Picker's Scope through the same fixed
// order regardless of mode (its first value read Active Root Scope in
// Grep mode, Active Source Scope in File mode, C6/M1.3) -> All Roots ->
// Starred -> back to the first, S7/S8/M5.2/C4-M5.1 — the exact same
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
	// S2, C6/M3.2: Wikipedia mode has no Scope — the Scope key is a no-op
	// (and pickerModeScope renders it as "-") — so there is nothing here to
	// change and no search to re-run.
	if m.picker.mode == wikipediaMode {
		return m, nil
	}
	// S6/C6/M3.5: with Wikipedia the Active Source there is no Active Root
	// to offer — File and Grep both skip it, cycling only between the
	// remaining two values rather than the full three-way order below, so
	// the Scope key never lands on the "no empty dead-end scope" S6 rules
	// out. switchPickerMode's own guard is what keeps p.scope out of
	// activeRootScope in the first place whenever this runs; this branch
	// only has to keep it out from here on.
	if m.roots.activeIsWikipedia() {
		if m.picker.scope == allRootsScope {
			m.picker.scope = starredScope
		} else {
			m.picker.scope = allRootsScope
		}
		return m.queuePickerSearch()
	}
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
		m.picker.input.Value(), m.roots.rootList(), m.roots.active, m.hideDotfiles, m.picker.filter, m.docs, m.wikipediaClient)
}

// resultCount is how many Result rows the current mode has to show right
// now (M3.1): the length of whichever of results.files/results.grep this
// mode populates. Used to clamp the selection cursor (move) and to size the
// Results content's windowedView — not what searchable/pickerStatus answer,
// which is about whether a search *should* run, not how many rows the last
// one that did produced.
func (p picker) resultCount() int {
	switch p.mode {
	case grepMode:
		return len(p.results.grep)
	case wikipediaMode:
		return len(p.results.wikipedia)
	default:
		return len(p.results.files)
	}
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
	widthChanged := p.preview.Width() != max(width, 0)
	p.preview.SetWidth(max(width, 0))
	p.preview.SetHeight(max(height, 0))
	if widthChanged && p.previewPath != "" {
		// Soft wrap depends on the width: lay the loaded file out again.
		p.previewLayout = nil
		return p.showPreview(p.preview.YOffset())
	}
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
// contract, unchanged) and for "nothing selected"; col is ripgrep's byte
// column of the match on that line (D140's exact span). Unlike
// pickerPreviewCandidate, this covers both modes — the Preview pane itself
// has no reason to wait for M4.3's Activate hand-off the way Activate
// does, since showing the matched file's raw bytes never changes the
// Active Document (S9's own "never changes the Active Document").
func (p picker) pickerPreviewTarget(hideDotfiles bool) (f pickerFile, line, col int, ok bool) {
	if !p.resultsCurrent(hideDotfiles) {
		return pickerFile{}, 0, 0, false
	}
	if p.mode == grepMode {
		if p.selected < 0 || p.selected >= len(p.results.grep) {
			return pickerFile{}, 0, 0, false
		}
		g := p.results.grep[p.selected]
		return g.file, g.line, g.col, true
	}
	if p.selected < 0 || p.selected >= len(p.results.files) {
		return pickerFile{}, 0, 0, false
	}
	return p.results.files[p.selected], 0, 0, true
}

// syncPreview loads the Preview pane's content for whichever Result
// pickerPreviewTarget currently names, re-targeting only when that target
// actually changed since the last call: called after every mutation that
// can change it (a Results move, a fresh or cleared result set landing, a
// mode switch), so the Preview always answers "what's selected right now"
// without re-reading a file or re-scrolling on every keystroke or render
// that leaves the selection alone.
//
// previewPath, previewLine, previewCol and previewQuery together are that
// "did it actually change" key: comparing path alone would miss a Grep
// selection moving to a *different match in the same file* (previewLine and
// previewCol are what catch that; see previewLine's own doc comment), and a fresh load only happens when
// the path itself changed — a same-file, different-line Grep move re-uses
// the bytes already loaded and only re-scrolls/re-highlights.
//
// The read is raw bytes, straight from disk — no loadDocument, no
// renderMarkdown: M3.2 is explicit that the Preview shows "raw, unrendered
// text", unlike the Document Viewer's own Glamour-rendered vp. D140 only
// changes how that text is drawn (showPreview), not what it is.
func (p picker) syncPreview(hideDotfiles bool) picker {
	f, line, col, ok := p.pickerPreviewTarget(hideDotfiles)
	if !ok {
		if p.previewPath != "" {
			p.previewPath = ""
			p.previewLine, p.previewCol, p.previewQuery = 0, 0, ""
			p.previewSrc, p.previewLayout, p.previewSpan = nil, nil, nil
			p.preview.SetContent("")
			p.preview.StyleLineFunc = nil
		}
		return p
	}
	query := ""
	if line > 0 {
		query = p.input.Value()
	}
	if f.path == p.previewPath && line == p.previewLine && col == p.previewCol && query == p.previewQuery {
		return p
	}
	if f.path != p.previewPath {
		var src []byte
		p.previewPlain = false
		if f.article {
			// C6/M3.5: an article pseudo-candidate's path is a URL, not a
			// file — reading it would only ever produce a raw OS "no such
			// file" error, the same misleading-as-broken text
			// activatePickerInto's own article branch avoids.
			//
			// C6/M4.2: Activate now does fetch this same row (D119's "every
			// network call is user-initiated" list: search, Activate, Peek,
			// Refresh — a Results-list selection move, which is what drives
			// this method, is deliberately not on that list), so the
			// placeholder is reworded to stay accurate about what is
			// specifically missing: a live preview of a not-yet-fetched
			// article, not the ability to open one at all.
			src = []byte("no preview — Wikipedia articles are not previewed from here")
			p.previewPlain = true
		} else {
			// D140: the Preview never reads a binary or oversize file
			// either; it says why instead.
			err := tree.CheckOpenable(f.path)
			if err == nil || !(errors.Is(err, tree.ErrBinary) || errors.Is(err, tree.ErrTooLarge)) {
				src, err = os.ReadFile(f.path) // I/O errors keep ReadFile's own wording
			}
			if err != nil {
				src = []byte("preview unavailable: " + err.Error())
				p.previewPlain = true
			}
		}
		p.previewPath = f.path
		p.previewSrc = src
		p.previewLayout, p.previewSpan = nil, nil
	}
	p.previewLine, p.previewCol, p.previewQuery = line, col, query
	return p.showPreview(0)
}

// showPreview draws the loaded file into the Preview viewport (D140): the
// Text renderer's soft wrap, expanded tabs and line-number gutter, so the
// Preview of a Grep Result matches what a Text-mode viewer shows once it is
// activated. A Grep Result's match is highlighted as the exact span ripgrep
// found, not its whole line, and scrolled to the top of the pane; any other
// target keeps yOffset (0 for a freshly loaded file). A message standing in
// for an unreadable file is wrapped but has no gutter.
func (p picker) showPreview(yOffset int) picker {
	p.preview.StyleLineFunc = nil
	width := p.preview.Width()
	if p.previewPlain {
		// The text can carry a file name, so it gets the same control-character
		// treatment as file content.
		msg := sanitizeControl(string(p.previewSrc))
		if width > 0 {
			msg = ansi.Wrap(msg, width, "")
		}
		p.preview.SetContent(msg)
		p.preview.GotoTop()
		return p
	}
	l := p.previewLayout
	content := ""
	old := p.previewSpan
	if l == nil {
		l = layoutText(p.previewSrc, width)
		p.previewLayout, old = l, nil
		content = l.render(false)
	} else {
		content = p.preview.GetContent()
	}
	row, span := yOffset, (*textSpan)(nil)
	if p.previewLine > 0 && len(l.rows) > 0 {
		// rg's line numbers are 1-based (S4/grepResult); the layout's
		// display lines are 0-based.
		row = l.rowStart[min(p.previewLine, len(l.rows))-1]
		matches := l.find(p.previewQuery)
		if i, ok := textGrepLanding(p.previewSrc, matches, p.previewLine, p.previewCol); ok {
			m := matches[i]
			row, span = m.line, &textSpan{line: m.src, start: m.colStart, end: m.colEnd}
		}
	}
	p.previewSpan = span
	p.preview.SetContent(l.patch(content, false, old, span))
	p.preview.SetYOffset(row)
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
		p.results.hideDotfiles == hideDotfiles && p.results.filter == p.filter
}

// searchable is whether the current Prompt may be searched at all: the
// query at or above S13's threshold, and — for File/Grep only — ripgrep
// present (S14 — it supplies both the file list and the text search, so
// without it there is no local search to run). Wikipedia mode needs no
// ripgrep at all (C6/M3.3): its search is a Wikimedia request, not a local
// shell-out, so a session with no rg on PATH can still search Wikipedia.
// Both search gates consult this, and pickerStatus reports the same
// conditions to the user, so what the modal says and what the model does
// cannot drift apart.
func (m model) searchable() bool {
	if len([]rune(m.picker.input.Value())) < pickerQueryMinLength {
		return false
	}
	if m.picker.mode == wikipediaMode {
		return true
	}
	return m.ripgrep.available
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
// wikipediaClient is nil only if a caller bypasses model.dispatchPickerSearch
// (which always passes m.wikipediaClient, never nil — newModel sets a real
// *wikipedia.Client and every test that reaches wikipediaMode injects a
// fake via withWikipediaClient); it is a parameter like every other value
// here rather than read from model for the same reason roots/docs/active
// are copied in, not read from m, in this function's own doc comment.
func pickerSearchCmd(seq int, mode pickerMode, scope pickerScope, query string, roots []root, active int, hideDotfiles bool, filter tree.Filter, docs map[string]docRecord, wikipediaClient pickerWikipediaClient) tea.Cmd {
	return func() tea.Msg {
		results := pickerResults{query: query, mode: mode, scope: scope, hideDotfiles: hideDotfiles, filter: filter}
		if mode == wikipediaMode {
			// S2: a title search, not a local file/grep one — no Root, no
			// Scope (wikipediaMode has none, C6/M3.2), no ripgrep. A
			// failure (offline, service error) reports back via err rather
			// than results.wikipedia, so applyPickerResults can leave the
			// previous Results in place instead of replacing them with an
			// empty list.
			items, err := wikipediaClient.Search(query, pickerWikipediaResultLimit)
			if err != nil {
				return pickerResultsMsg{seq: seq, results: results, err: err}
			}
			results.wikipedia = items
			return pickerResultsMsg{seq: seq, results: results}
		}
		var files []pickerFile
		if scope == starredScope {
			// S8: the whole store, not a walk of any Root's tree —
			// starredPickerFiles resolves Scope on its own terms, so
			// listFiles' own roots-shaped contract (liveRootsInScope)
			// never sees starredScope at all.
			files = starredPickerFiles(docs, roots, hideDotfiles)
		} else {
			files = listFiles(roots, active, scope, hideDotfiles, filter)
		}
		switch mode {
		case grepMode:
			// S6/C6/M3.5: "Grep never includes Wikipedia, in any scope" —
			// starredPickerFiles' own article pseudo-candidates (Starred
			// scope only; every other scope never produces one) are dropped
			// before rg ever sees them.
			results.grep, results.truncated = grepPickerFiles(filesOnly(files), query)
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
	m.picker.searching = false
	if msg.err != nil {
		// S2: "a failed search ... leaves the last Results in place." Only
		// the labels this search was for advance (so resultsCurrent still
		// recognizes the retained rows as answering the Prompt/mode/Scope);
		// the rows themselves — whichever of files/grep/wikipedia a
		// previous successful search set — are left untouched rather than
		// replaced by msg.results' own empty ones. wikipediaErr is what
		// pickerStatus shows in their place as the status line.
		m.picker.wikipediaErr = msg.err.Error()
		m.picker.results.query = msg.results.query
		m.picker.results.mode = msg.results.mode
		m.picker.results.scope = msg.results.scope
		m.picker.results.hideDotfiles = msg.results.hideDotfiles
		m.picker.results.filter = msg.results.filter
		return m
	}
	m.picker.wikipediaErr = ""
	m.picker.results = msg.results
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
func (m model) activatePicker() (model, tea.Cmd) {
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
//
// C6/M4.2 widens the return to (model, tea.Cmd): wikipediaMode's own Result
// and a starred article surfaced by File mode's Starred scope (f.article,
// below) both open asynchronously now (fetchArticleInto), unlike every
// other Result this method handles, which still opens synchronously with a
// nil Cmd.
func (m model) activatePickerInto(target pane) (model, tea.Cmd) {
	if m.picker.mode == grepMode {
		return m.activateGrepPickerInto(target), nil
	}
	// C6/M3.2/M3.3: wikipediaMode's own Results never populate
	// p.results.files, so pickerPreviewCandidate (fileMode-only, below)
	// always answers "not ok" for this mode — its own Result lookup is
	// pickerWikipediaCandidate instead, the same one togglePickerStar
	// already uses. A stale generation or an empty Results list is the same
	// silent no-op every other mode's own "not ok" gives (see
	// pickerWikipediaCandidate's own doc comment).
	if m.picker.mode == wikipediaMode {
		r, ok := m.picker.pickerWikipediaCandidate(m.hideDotfiles)
		if !ok {
			return m, nil
		}
		next, cmd := m.fetchArticleInto(target, r.Title, false)
		return next.dismissPicker(), cmd
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
		return m, nil
	}
	// C6/M3.5, fetch wired by C6/M4.2: a starred article surfaced by File
	// mode's Starred scope (starredPickerFiles) has no bytes on disk for
	// loadDocument to read — f.path is its canonical/derived URL, not a real
	// file — so this would otherwise fail as a plain "cannot open
	// https://..." read error, which reads like a broken file rather than
	// an unbuilt feature. M3.5 refused this outright pending M4.2; this now
	// fetches exactly like a wikipediaMode Result does, using f.title (the
	// starred record's own title, M3.4) rather than f.path — Fetch takes a
	// title, not a URL (wikipedia.Client.Fetch's own contract).
	if f.article {
		next, cmd := m.fetchArticleInto(target, f.title, false)
		return next.dismissPicker(), cmd
	}
	d, err := loadDocument(f.path)
	if err != nil {
		m.message = cannotOpen(f.path, err)
		return m.dismissPicker(), nil
	}
	// S4/M3.3: Picker Activate records recency the same as the Files Nav's
	// own activateInto (M3.1) — TRIAGE T12 found this call site bypassing
	// recordViewed entirely by going straight to openDocumentInto.
	m = m.recordViewed(f.path)
	// S9/C4/M6.1: the Picker's Activate is a real navigation, so it records
	// onto the target viewer's own back/forward stack the same way the
	// Files Nav's activateInto does (model.navigate).
	m = m.navigate(target, d)
	return m.dismissPicker(), nil
}

// togglePickerStar flips the S1 document store's star mark for the
// Picker's currently selected Result (S2's third starring surface,
// M2.5) — whichever mode is open, fileMode's pickerPreviewCandidate,
// grepMode's pickerGrepCandidate, or (C6/M3.4) wikipediaMode's
// pickerWikipediaCandidate, all three already gated on resultsCurrent so
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
//
// wikipediaMode is its own branch, not folded into pickerSelectedPath: a
// file/grep Result's key is already its path, but an article's store key
// is its canonical URL (M1.5/M4.1), which a title-search SearchResult never
// carries — only Fetch (M4.2) learns it. Rather than block starring on a
// fetch the Picker's own Star key was never asked to trigger, this derives
// a URL straight from the title (wikipedia.ArticleURL) the same way an
// already-canonical Fetch would, and stamps Article/Title so the record
// reads exactly like one M4.1 wrote after opening the article — for a
// title that turns out to be a redirect (rare: title search itself already
// matches real page titles), a later Fetch resolves the true canonical URL
// under a separate record; this one is not migrated or merged (see
// wikipedia.ArticleURL's own doc comment, and D122's precedent of accepting
// the narrower guarantee for a title noggin has not yet confirmed is
// canonical).
func (m model) togglePickerStar() model {
	if m.picker.mode == wikipediaMode {
		r, ok := m.picker.pickerWikipediaCandidate(m.hideDotfiles)
		if !ok {
			return m
		}
		key := wikipedia.ArticleURL(r.Title)
		rec := m.docRecord(key)
		rec.Starred = !rec.Starred
		rec.Article = true
		rec.Title = r.Title
		return m.withDocRecord(key, rec)
	}
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

// pickerWikipediaCandidate is the Result togglePickerStar (C6/M3.4) acts on
// in wikipediaMode: the selected row, gated the same way
// pickerPreviewCandidate/pickerGrepCandidate gate their own modes —
// resultsCurrent and the live result count, so a stale generation or a
// fresh, shorter result set landing between a selection move and the next
// render can no more be starred than it may be previewed or activated.
func (p picker) pickerWikipediaCandidate(hideDotfiles bool) (wikipedia.SearchResult, bool) {
	if p.mode != wikipediaMode || !p.resultsCurrent(hideDotfiles) {
		return wikipedia.SearchResult{}, false
	}
	if p.selected < 0 || p.selected >= len(p.results.wikipedia) {
		return wikipedia.SearchResult{}, false
	}
	return p.results.wikipedia[p.selected], true
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
// D34: Glamour's rendering (re-flow, tab expansion, markdown syntax
// removed) does not preserve rg's line numbers across that boundary.
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
		m.message = cannotOpen(g.file.path, err)
		return m.dismissPicker()
	}
	// S4/M3.3: same recordViewed fix as activatePickerInto's fileMode half,
	// for the grepMode path (TRIAGE T12).
	m = m.recordViewed(g.file.path)
	// S9/C4/M6.1: same history-recording navigation as fileMode's own
	// activatePickerInto above.
	m = m.navigate(target, d)
	v := m.viewerFor(target).commitQuery(query)
	if v.text != nil {
		// Text mode (D140): ripgrep's own line and column, no D34 guess.
		if ordinal, ok := textGrepLanding(d.src, v.findState.matches, g.line, g.col); ok {
			v.findState.current = ordinal
			v = v.applyCurrentMatch()
		}
	} else if ordinal, ok := resolveGrepLanding(d.src, v.vp.GetContent(), query, g.line, v.vp.Width()); ok {
		v.findState.current = ordinal
		v = v.applyCurrentMatch()
	} else {
		// D34's degrade path: remember ripgrep's own location so a switch to
		// Text can land exactly (D140).
		v.findState.landing = &grepLanding{line: g.line, col: g.col}
	}
	m = m.withViewerFor(target, v)
	return m.dismissPicker()
}

// resolveGrepLanding is D34's occurrence-ordinal bridge from a Grep
// Result's raw-file line number to a match index into
// findMatches(rendered, query): "the Nth literal occurrence of the query"
// is a property of the text itself, invariant to how it later gets wrapped
// onto lines, so the Nth occurrence in the raw source is assumed to be the
// Nth occurrence in the rendered content too. rendered is src as rendered
// at width.
//
// The Frontmatter Card (D138) reshapes its text ("title: x" is drawn as
// "title  x"), so frontmatter and body are counted as separate segments
// (frontmatterSpan): a raw frontmatter occurrence the card does not
// reproduce cannot shift every body match after it. ok is false — D34's
// explicit degrade trigger — when the resolved ordinal has no counterpart in
// its segment of the rendered matches (the card dropped it, or Glamour
// removed or reordered occurrences before it, e.g. inside stripped markdown
// syntax).
//
// "This match" resolves to the *first* occurrence of query on line: rg
// reports one Result per matching line, not per occurrence within it
// (grepBatch's own comment), so a line with several occurrences of query
// picks its first as the line's representative match — the same
// approximation D34 accepts for markdown-syntax-adjacent matches.
func resolveGrepLanding(src []byte, rendered, query string, line, width int) (ordinal int, ok bool) {
	raw := findMatches(string(src), query)
	target := -1
	for i, m := range raw {
		if m.line+1 == line { // findMatch.line is 0-based; rg's line is 1-based
			target = i
			break
		}
	}
	if target < 0 {
		return 0, false
	}
	rend := findMatches(rendered, query)
	sourceLines, renderedLines := frontmatterSpan(src, width)
	rawHead, rendHead := matchesBefore(raw, sourceLines), matchesBefore(rend, renderedLines)
	switch {
	case target >= rawHead:
		ordinal = rendHead + target - rawHead
	case target < rendHead:
		ordinal = target
	default:
		return 0, false
	}
	if ordinal >= len(rend) {
		return 0, false
	}
	return ordinal, true
}

// textGrepLanding is the index into a Text-mode find's matches of the Grep
// Result at ripgrep's 1-based line and byte column col (D140): the match on
// that source line starting at that column, once the column is translated from
// the raw line into the display line (tabs expanded, control characters and a
// BOM replaced or dropped, as textLines does). When no match starts exactly
// there — the file changed since ripgrep read it, or the display text no
// longer contains the query — the first match on the line stands in; ok is
// false when there is none.
func textGrepLanding(src []byte, matches []findMatch, line, col int) (int, bool) {
	displayCol := textDisplayCol(src, line-1, col)
	first := -1
	for i, m := range matches {
		if m.src != line-1 {
			continue
		}
		if m.colStart == displayCol {
			return i, true
		}
		if first < 0 {
			first = i
		}
	}
	if first < 0 {
		return 0, false
	}
	return first, true
}

// matchesBefore counts the matches on lines before line; matches are in
// on-screen order, so they are a prefix.
func matchesBefore(matches []findMatch, line int) int {
	n := 0
	for n < len(matches) && matches[n].line < line {
		n++
	}
	return n
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
	case key.Matches(msg, keys.PickerRestore):
		// Already open: nothing to restore, and never query text.
		return m, nil
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
		return m.activatePickerInto(dv2Pane)
	case key.Matches(msg, keys.Activate):
		return m.activatePicker()
	case key.Matches(msg, keys.PickerMode):
		return m.switchPickerMode()
	case key.Matches(msg, keys.PickerScope):
		return m.switchPickerScope()
	case key.Matches(msg, keys.ToggleHidden):
		return m.toggleHiddenFiles()
	case key.Matches(msg, keys.PickerFilter):
		return m.togglePickerFilter()
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
