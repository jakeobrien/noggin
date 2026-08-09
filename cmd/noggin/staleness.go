package main

import (
	"os"
	"time"

	tea "charm.land/bubbletea/v2"

	"noggin/internal/tree"
)

// staleCheckInterval is S10's periodic check cadence — the spec's own
// "~5-10s", picked at the low end so the indicator surfaces promptly.
const staleCheckInterval = 6 * time.Second

// staleMessage is S10's one and only wording: it never says what changed,
// only that something did, and where to act on it.
const staleMessage = "something changed on disk; press R to refresh"

// staleBaseline is the "known good" snapshot a periodic check compares the
// disk against (S10, C2a/M10.1; S7/C3/M5.2 extends it to both viewers):
// established fresh every time the Active Root's tree or a viewer's Active
// Document is (re)loaded (openFolder, openDocumentInto, refresh), never
// touched in between.
//
// dirModTimes is one os.Stat per directory the already-loaded tree knows
// about (root plus every directory node under it) — a lookup by path, never
// a directory listing, which is what keeps a periodic check from ever
// calling tree.Walk (S10's "without re-walking anything") no matter how many
// *files* a large Root holds: a notes vault's directory count is a small
// fraction of its file count, so this stays cheap exactly where the spec
// requires it. It is deliberately a heuristic, not exhaustive change
// detection: an in-place edit of a *non-Active* file that touches neither
// its own directory's entry list nor either viewer's Active Document goes
// unnoticed until the next refresh — catching that would mean statting
// every file, precisely the cost S10 forbids. A viewer's Active Document
// itself has no such gap: docs' own modTime/size are statted directly on
// every check, so any edit to it (in place or not) is always caught.
//
// docs holds at most two entries (DV1's, then DV2's if it exists) — a slice
// rather than a fixed DV1/DV2 pair so stale() can loop over "whichever
// documents exist" (S7's own phrase) without a per-viewer branch, and so a
// later increase in the number of viewers (out of scope this cycle, but not
// this type's problem) would not need a new field.
type staleBaseline struct {
	dirModTimes map[string]time.Time
	docs        []docSnapshot
}

// docSnapshot is one tracked document's on-disk identity at snapshot time:
// its path, mtime and size, the same trio staleBaseline tracked inline for
// the single Active Document before S7/C3/M5.2 gave it a name of its own.
type docSnapshot struct {
	path    string
	modTime time.Time
	size    int64
}

// snapshotStaleness builds a fresh baseline from the tree already loaded in
// memory (root, nil if no Root is open) and every non-empty docPath in docs
// (S7: DV1's and DV2's Active Document paths, whichever viewers exist and
// have one open — an absent DV2 or a viewer with nothing open contributes no
// path at all, never a placeholder empty one). Every Stat is against a path
// the caller already knows about — nothing here discovers a new path via
// os.ReadDir.
func snapshotStaleness(root *tree.Node, docPaths ...string) staleBaseline {
	b := staleBaseline{dirModTimes: map[string]time.Time{}}
	collectDirModTimes(root, b.dirModTimes)
	for _, docPath := range docPaths {
		if docPath == "" {
			continue
		}
		// A Stat failure right here (a narrow TOCTOU window right after the
		// caller's own successful read) leaves the document untracked rather
		// than tracked with a zero-value modTime/size, which would otherwise
		// compare unequal to the real file on every future check and pin
		// stale() permanently true until the next reset — the same "omit,
		// don't half-track" rule collectDirModTimes already applies to a
		// directory whose own Stat fails.
		if info, err := os.Stat(docPath); err == nil {
			b.docs = append(b.docs, docSnapshot{path: docPath, modTime: info.ModTime(), size: info.Size()})
		}
	}
	return b
}

// collectDirModTimes records n's own mtime (if n is a directory) and
// recurses into its Children — an in-memory traversal of the tree the
// caller already built, not a filesystem walk: no os.ReadDir call is made
// here at all, only os.Stat against paths the tree already names.
func collectDirModTimes(n *tree.Node, out map[string]time.Time) {
	if n == nil || !n.IsDir {
		return
	}
	if info, err := os.Stat(n.Path); err == nil {
		out[n.Path] = info.ModTime()
	}
	for _, c := range n.Children {
		collectDirModTimes(c, out)
	}
}

// withoutDoc returns a copy of b with path's tracked docSnapshot removed (a
// no-op, returning b unchanged, if path is empty or not tracked) — every
// other tracked doc and dirModTimes are left exactly as they were. Used by
// closeDV2 (C3/M2.1b, TRIAGE T12) to prune a viewer's former document from
// the baseline the instant it stops being visible, rather than waiting for
// the next resetStaleBaseline (openFolder/openDocumentInto/refresh) to
// happen to do it as a side effect of rebuilding the whole snapshot.
func (b staleBaseline) withoutDoc(path string) staleBaseline {
	if path == "" {
		return b
	}
	docs := make([]docSnapshot, 0, len(b.docs))
	for _, d := range b.docs {
		if d.path != path {
			docs = append(docs, d)
		}
	}
	b.docs = docs
	return b
}

// trackedDocPaths is the paths of every document b actually tracks (S7:
// whichever of DV1's/DV2's Active Documents existed and stat'd successfully
// at snapshot time) — a read-only view for tests to assert against rather
// than reaching into the docs slice's shape directly.
func (b staleBaseline) trackedDocPaths() []string {
	paths := make([]string, len(b.docs))
	for i, d := range b.docs {
		paths[i] = d.path
	}
	return paths
}

// stale reports whether the disk still matches b: every known directory's
// own mtime unchanged, and every tracked document's (S7: whichever of DV1's/
// DV2's Active Documents were open at snapshot time) mtime and size both
// unchanged. A path that no longer stats at all (removed) counts as
// changed, matching S10's "something changed"; the indicator never says
// which case, or which document, applies — S9's "one combined, deliberately
// unspecific signal" — so this collapses all of it to one bool.
func (b staleBaseline) stale() bool {
	for path, want := range b.dirModTimes {
		info, err := os.Stat(path)
		if err != nil || !info.ModTime().Equal(want) {
			return true
		}
	}
	for _, d := range b.docs {
		info, err := os.Stat(d.path)
		if err != nil || !info.ModTime().Equal(d.modTime) || info.Size() != d.size {
			return true
		}
	}
	return false
}

// staleCheckMsg is a matured periodic timer — a pure "time to check again"
// signal with no payload. Unlike picker.go's pickerSearchMsg (which does
// carry a generation, since a superseded debounce timer must dispatch no
// search at all), a stale timer always dispatches a check: runStaleCheck
// reads whatever baseline/generation are current at that moment rather
// than one captured at schedule time, so there is nothing here for a
// generation to gate — see runStaleCheck's own doc comment.
type staleCheckMsg struct{}

// staleResultMsg is a completed disk check reporting back, carrying the
// generation it was dispatched for so a result computed against a baseline
// that has since been superseded is discarded rather than applied.
type staleResultMsg struct {
	gen   int
	stale bool
}

// scheduleStaleCheck arms the next periodic tick against the model's current
// generation. Called from Init (to start the cycle) and after every
// staleCheckMsg/staleResultMsg (to keep it going) — S10 runs for the life of
// the session, not once.
func (m model) scheduleStaleCheck() tea.Cmd {
	return tea.Tick(staleCheckInterval, func(time.Time) tea.Msg {
		return staleCheckMsg{}
	})
}

// resetStaleBaseline (re)establishes the "known good" snapshot from
// whatever Root and Active Document the model currently has, clears the
// indicator (the caller just (re)read from disk, so nothing is stale
// relative to that read), and opens a new generation — discarding any
// in-flight check that was still comparing against the old one. Called
// from openFolder, openDocument and refresh: every place that (re)reads the
// Active Root's tree or the Active Document from disk.
func (m model) resetStaleBaseline() model {
	// S7/C3/M5.2: both viewers' Active Documents, whichever exist — DV1 is
	// always present (an empty path if nothing is open there); DV2
	// contributes nothing at all while m.dv2 is nil, not a placeholder empty
	// path, matching snapshotStaleness's own "no path at all for an absent
	// viewer" contract.
	dv1Path := ""
	if m.dv1.doc.isOpen() {
		dv1Path = m.dv1.doc.path
	}
	if m.dv2 == nil {
		m.staleBaseline = snapshotStaleness(m.files.root, dv1Path)
	} else {
		dv2Path := ""
		if m.dv2.doc.isOpen() {
			dv2Path = m.dv2.doc.path
		}
		m.staleBaseline = snapshotStaleness(m.files.root, dv1Path, dv2Path)
	}
	m.stale = false
	m.staleGen++
	return m
}

// runStaleCheck handles a matured periodic timer by dispatching the actual
// Stat work as its own Cmd, against whichever baseline and generation are
// current right now — never ones captured back when the timer was
// scheduled. That is deliberate, not an oversight: a naive "drop this timer
// if resetStaleBaseline ran since it was scheduled" gate (mirroring
// picker.go's debounce discard) would mean opening a document, switching
// Root or refreshing while a timer is already in flight pushes the next
// real check out to roughly two intervals instead of the one M10.1's own
// acceptance criterion names — there is nothing stale about running a check
// against the model's current state; it is strictly the most correct thing
// to check.
//
// baseline is captured by value here (a struct copy; its map field is
// never mutated in place once built, only ever replaced wholesale by
// resetStaleBaseline, so sharing that reference across goroutines is safe —
// the same reasoning M3.1's `expanded` maps already rely on) so the check
// that runs reflects this exact moment even if the model moves on again
// before the Cmd returns; applyStaleResult's own generation compare is what
// catches that later case.
func (m model) runStaleCheck(staleCheckMsg) (model, tea.Cmd) {
	baseline, gen := m.staleBaseline, m.staleGen
	return m, func() tea.Msg {
		return staleResultMsg{gen: gen, stale: baseline.stale()}
	}
}

// applyStaleResult installs a completed check's verdict (only if its
// generation still matches — a superseded result is silently dropped, the
// baseline it was computed against no longer describes what "known good"
// means) and reschedules the next tick. Once set, m.stale stays set until
// resetStaleBaseline clears it (a refresh) — a periodic check that finds
// nothing new never un-sets an indicator a previous check already raised.
func (m model) applyStaleResult(msg staleResultMsg) (model, tea.Cmd) {
	if msg.gen == m.staleGen && msg.stale {
		m.stale = true
	}
	return m, m.scheduleStaleCheck()
}
