package main

import (
	"os"
	"path/filepath"
	"slices"
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
// Document is (re)loaded (openFolder, openDocumentInto, refresh). The one
// sanctioned exception is closeDV2 (C3/M2.1b), which mutates it directly via
// withoutDoc to prune DV2's former document the instant it stops being
// visible, rather than waiting for the next full reset — otherwise never
// touched in between.
//
// dirs holds one dirSnapshot per directory the already-loaded tree knows
// about (root plus every directory node under it). A check is one os.Stat
// per directory (plus one per existing `.gitignore`), and only a directory
// whose mtime moved is re-listed (tree.ListDir) and compared against what
// the tree shows there (D139). A directory's mtime moves for entries the
// tree never shows — editor swap/probe files, .DS_Store, a non-Markdown or
// gitignored file, the temp file of an atomic save — so the re-list is what
// keeps those from raising the indicator, while still never calling
// tree.Walk (S10's "without re-walking anything"). It is deliberately a
// heuristic, not exhaustive change detection: an in-place edit of a
// *non-Active* file goes unnoticed (the tree would show nothing new for it
// anyway), and so do git status marker changes. A viewer's Active Document
// itself has no such gap: docs' own modTime/size are statted directly on
// every check, so any edit to it (in place or not) is always caught.
//
// A directory whose mtime moved but whose entries still match is re-listed
// again on every check until the next reset — a ReadDir per such directory,
// accepted over folding confirmed mtimes back into the baseline (which
// would make the baseline mutable across goroutines).
//
// docs holds at most two entries (DV1's, then DV2's if it exists) — a slice
// rather than a fixed DV1/DV2 pair so stale() can loop over "whichever
// documents exist" (S7's own phrase) without a per-viewer branch, and so a
// later increase in the number of viewers (out of scope this cycle, but not
// this type's problem) would not need a new field.
type staleBaseline struct {
	root string
	dirs map[string]dirSnapshot
	docs []docSnapshot

	// filter is the File Filter the tree was walked under, so re-listing a
	// directory compares like with like: a change to a file the filter hides
	// never raises the indicator (D139, D140).
	filter tree.Filter
}

// dirSnapshot is one known directory's state at snapshot time: its own
// mtime, the entries the tree shows in it (its Children, in Walk order —
// what tree.ListDir returns for an unchanged directory), and its
// `.gitignore`'s stamp (exists=false if it had none). A `.gitignore` is
// tracked on its own because it is never a tree entry, yet editing,
// creating or removing one can change what the tree shows anywhere beneath
// it — so any change to one counts as stale outright (D139).
type dirSnapshot struct {
	modTime   time.Time
	entries   []tree.Entry
	gitignore fileStamp
}

// fileStamp is a file's existence, mtime and size; the zero value means
// "does not exist".
type fileStamp struct {
	exists  bool
	modTime time.Time
	size    int64
}

func stampOf(path string) fileStamp {
	info, err := os.Stat(path)
	if err != nil {
		return fileStamp{}
	}
	return fileStamp{exists: true, modTime: info.ModTime(), size: info.Size()}
}

func (a fileStamp) equal(b fileStamp) bool {
	return a.exists == b.exists && a.modTime.Equal(b.modTime) && a.size == b.size
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
func snapshotStaleness(root *tree.Node, filter tree.Filter, docPaths ...string) staleBaseline {
	b := staleBaseline{dirs: map[string]dirSnapshot{}, filter: filter}
	if root != nil {
		b.root = root.Path
	}
	collectDirSnapshots(root, b.dirs)
	for _, docPath := range docPaths {
		if docPath == "" {
			continue
		}
		// A Stat failure right here (a narrow TOCTOU window right after the
		// caller's own successful read) leaves the document untracked rather
		// than tracked with a zero-value modTime/size, which would otherwise
		// compare unequal to the real file on every future check and pin
		// stale() permanently true until the next reset — the same "omit,
		// don't half-track" rule collectDirSnapshots already applies to a
		// directory whose own Stat fails.
		//
		// C6/M4.1 note for M4.3 ("the staleness indicator never fires for
		// [an article]", S3): once an article document's path can reach here
		// (a URL, isArticleKey), os.Stat fails on it exactly like any other
		// unreadable path and this same omit-don't-half-track rule already
		// keeps it untracked — accidentally satisfying that clause, not by a
		// deliberate URL-aware branch. Worth keeping in mind before "fixing"
		// this to Stat something else for a URL.
		if info, err := os.Stat(docPath); err == nil {
			b.docs = append(b.docs, docSnapshot{path: docPath, modTime: info.ModTime(), size: info.Size()})
		}
	}
	return b
}

// collectDirSnapshots records n's own dirSnapshot (if n is a directory) and
// recurses into its Children — an in-memory traversal of the tree the
// caller already built, not a filesystem walk: no os.ReadDir call is made
// here at all, only os.Stat against paths the tree already names.
func collectDirSnapshots(n *tree.Node, out map[string]dirSnapshot) {
	if n == nil || !n.IsDir {
		return
	}
	if info, err := os.Stat(n.Path); err == nil {
		entries := make([]tree.Entry, 0, len(n.Children))
		for _, c := range n.Children {
			entries = append(entries, tree.Entry{Name: c.Name, IsDir: c.IsDir})
		}
		out[n.Path] = dirSnapshot{
			modTime:   info.ModTime(),
			entries:   entries,
			gitignore: stampOf(filepath.Join(n.Path, ".gitignore")),
		}
	}
	for _, c := range n.Children {
		collectDirSnapshots(c, out)
	}
}

// withoutDoc returns a copy of b with path's tracked docSnapshot removed (a
// no-op, returning b unchanged, if path is empty or not tracked) — every
// other tracked doc and dirs are left exactly as they were. Used by
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

// stale reports whether the disk no longer matches b: a known directory
// whose mtime moved *and* whose re-listed entries differ from the tree's, a
// `.gitignore` that changed, appeared or vanished, or a tracked document
// (S7: whichever of DV1's/DV2's Active Documents were open at snapshot time)
// whose mtime or size changed. A path that no longer stats at all (removed)
// counts as changed, matching S10's "something changed"; the indicator never
// says which case, or which document, applies — S9's "one combined,
// deliberately unspecific signal" — so this collapses all of it to one bool.
func (b staleBaseline) stale() bool {
	for path, want := range b.dirs {
		if b.dirChanged(path, want) {
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

// dirChanged is stale()'s per-directory step. The `.gitignore` is statted
// only when one existed or the directory's mtime moved (a create or remove
// always moves it), so an untouched directory without one costs one Stat.
func (b staleBaseline) dirChanged(path string, want dirSnapshot) bool {
	info, err := os.Stat(path)
	if err != nil {
		return true
	}
	moved := !info.ModTime().Equal(want.modTime)
	if want.gitignore.exists || moved {
		if !stampOf(filepath.Join(path, ".gitignore")).equal(want.gitignore) {
			return true
		}
	}
	if !moved {
		return false
	}
	entries, err := tree.ListDir(b.root, path, b.filter)
	if err != nil {
		return true
	}
	return !slices.Equal(entries, want.entries)
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
// that has since been superseded is discarded rather than applied. periodic
// marks a check dispatched by the periodic timer, which applying the result
// re-arms; a focus-triggered check (checkStaleOnFocus) leaves it false so it
// never starts a second timer chain alongside the one already running.
type staleResultMsg struct {
	gen      int
	stale    bool
	periodic bool
}

// scheduleStaleCheck arms the next periodic tick against the model's current
// generation. Called from Init (to start the cycle) and after every
// periodic staleResultMsg (to keep it going) — S10 runs for the life of
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
	filter := m.files.fileFilter()
	if m.dv2 == nil {
		m.staleBaseline = snapshotStaleness(m.files.root, filter, dv1Path)
	} else {
		dv2Path := ""
		if m.dv2.doc.isOpen() {
			dv2Path = m.dv2.doc.path
		}
		m.staleBaseline = snapshotStaleness(m.files.root, filter, dv1Path, dv2Path)
	}
	m.stale = false
	m.staleGen++
	return m
}

// rebaseTreeStaleness re-snapshots the Root's tree after a File Filter toggle
// re-walked it, so staleness compares listings under the new filter (D139).
// Unlike resetStaleBaseline it keeps the tracked documents and any raised
// indicator: nothing re-read a document, so none is known-good again. The
// generation moves on so a check still running against the old filter is
// discarded.
func (m model) rebaseTreeStaleness() model {
	b := snapshotStaleness(m.files.root, m.files.fileFilter())
	b.docs = m.staleBaseline.docs
	m.staleBaseline = b
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
	return m, m.staleCheckCmd(true)
}

// checkStaleOnFocus runs a check the moment the terminal regains focus
// (tea.FocusMsg, D139) — the common "edited it in another window, came
// back" case — instead of waiting up to a full staleCheckInterval. It is
// in addition to the periodic timer, not a replacement: terminals that
// never report focus (or tmux without focus-events) still get the timer,
// and the result does not reschedule (staleResultMsg.periodic is false),
// so the one timer chain Init started stays the only one.
func (m model) checkStaleOnFocus() (model, tea.Cmd) {
	return m, m.staleCheckCmd(false)
}

// staleCheckCmd is the off-Update disk check both triggers dispatch,
// against the baseline and generation current right now (see
// runStaleCheck's doc comment for why the copy is safe).
func (m model) staleCheckCmd(periodic bool) tea.Cmd {
	baseline, gen := m.staleBaseline, m.staleGen
	return func() tea.Msg {
		return staleResultMsg{gen: gen, stale: baseline.stale(), periodic: periodic}
	}
}

// applyStaleResult installs a completed check's verdict (only if its
// generation still matches — a superseded result is silently dropped, the
// baseline it was computed against no longer describes what "known good"
// means) and, for a periodic check only, reschedules the next tick. Once set, m.stale stays set until
// resetStaleBaseline clears it (a refresh) — a periodic check that finds
// nothing new never un-sets an indicator a previous check already raised.
func (m model) applyStaleResult(msg staleResultMsg) (model, tea.Cmd) {
	if msg.gen == m.staleGen && msg.stale {
		m.stale = true
	}
	if !msg.periodic {
		return m, nil
	}
	return m, m.scheduleStaleCheck()
}
