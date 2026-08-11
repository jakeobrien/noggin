package main

import (
	"path/filepath"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"noggin/internal/tree"
)

// emptyFilesText is what the Files Nav pane says when there is no root to
// browse (a bare launch, or a file opened without a folder). M3 gives the pane
// a root to show; until then this is its resting state.
const emptyFilesText = "No folder open."

// filesNav is the Files Nav pane: an expansion/selection model layered over an
// immutable tree.Node. The tree itself is static data (M2.1); which folders are
// open and which row is selected is UI state kept here, keyed by absolute Path
// so it survives independent of the tree's shape.
type filesNav struct {
	root *tree.Node

	// expanded holds the Paths of folders currently open. A folder absent from
	// the map is collapsed; the tree opens with every folder collapsed (S5), so
	// the empty map is the initial state.
	expanded map[string]bool

	// selected is an index into the flattened visible rows (visibleRows), not
	// into the tree — it addresses what the user sees, so it tracks the current
	// expansion rather than a fixed node.
	selected int

	// memory holds, per Root, the expansion set and selected node's Path last
	// left there (M3.1/S3), and — since M4.6 — the Browse Mode and per-mode
	// Browse Scope last left there too (S7), keyed by the Root's own Path —
	// the same value model.activateRoot re-walks into a brand new *tree.Node
	// each time a Root is (re)activated, even when nothing on disk changed, so
	// memory is keyed by Path rather than Node identity, the same way expanded
	// already is. A Root with no entry here has never been visited this
	// session.
	memory map[string]filesNavMemory

	// status holds S8's per-file git markers for the current root, keyed by
	// absolute Path (gitStatus's return shape, ready to match tree.Node.Path
	// directly). Unlike memory, it is never restored from a prior visit — the
	// caller (model.openFolder) recomputes it fresh on every load, since the
	// underlying git state can change between visits to the same Root. Nil
	// (the zero value) renders identically to an empty map: no markers.
	status map[string]gitFileStatus

	// starred holds the S1 document store's currently-Starred Paths (S2/S3),
	// membership only — a file's own row needs nothing more than "is it in
	// this set". Unlike status, this can change from *inside* the pane
	// itself (M2.1's toggleDocStar, a Files Nav action), so the model
	// refreshes it there too, not just at openFolder/refresh — see
	// model.filesStarredMarks' callers. Nil renders identically to an empty
	// map: no markers, the same degrade status already gives.
	starred map[string]bool

	// mode and scope are S5's Browse Mode/Browse Scope, cycled by
	// keys.FilesMode/keys.FilesScope (handleKey below) only while the Files
	// Nav has Focus. The zero values (treeBrowse, activeRootBrowseScope) are S5's
	// own starting state, so newFilesNav needs no explicit initialization —
	// a session always opens on Tree (S7), and Tree's scope reads "-"
	// (scopeLabel) regardless of the scope field's value. M4.6 gives these
	// per-Root (mode) and per-(Root,mode) (scope) memory via scopeByMode below
	// and setRoot/remember's own restore/capture.
	mode  browseMode
	scope browseScope

	// scopeByMode is the live, current-Root half of S7's Browse Scope memory:
	// each Browse Mode's own last-used scope while this Root has been active
	// this visit. cycleMode/remember both snapshot the outgoing mode's live
	// f.scope into it (withScopeRemembered) before moving off that mode, so
	// scopeForCurrentMode can always answer "what was this mode's scope last
	// time" rather than reaching for modeScopes' own first value every time —
	// which now only happens for a mode never yet visited this Root. setRoot
	// swaps the whole map out for whatever a different Root's own memory (or
	// nothing, on a first visit) holds, the same round-trip expanded already
	// gets.
	scopeByMode map[browseMode]browseScope

	// docs, sessionRoots and activeRootPath are the S1 store and S1 launch
	// state's own snapshot the pane's Recent (M4.2) and Starred (M4.3) flat
	// lists are computed from, installed by withRecentSource. Like status/
	// starred (their own doc comments), this is a caller-refreshed snapshot,
	// not something the pane keeps live itself — model.refreshFilesRecentSource
	// is the one place that recomputes and pushes it, at every point the
	// store or the session's roots could have changed. Unlike status/starred,
	// the rows built from it are never cached on filesNav either: recentRows
	// runs fresh from these three fields on every render (rowsView), the same
	// "no cache to go stale" shape M2.11's Picker-row star markers already
	// chose over M2.7's Files Nav marker cache.
	docs           map[string]docRecord
	sessionRoots   []root
	activeRootPath string

	// width and height are the content box the pane draws into, borders and
	// title already deducted (see computeLayout / rect.content).
	width  int
	height int
}

// filesNavMemory is one Root's remembered browse state: which folders were
// open, and which node was selected. The selection is remembered by Path,
// not the flattened-row index `selected` itself uses — an index is only
// meaningful against the expansion set that produced it, so setRoot
// recomputes the index against whatever the restored expansion set's visible
// rows turn out to be (indexOfPath), rather than trusting a stale int.
type filesNavMemory struct {
	expanded     map[string]bool
	selectedPath string

	// mode and scopeByMode are S7's own per-Root Browse Mode/Browse Scope
	// memory: mode is the Browse Mode the pane was showing when this Root was
	// last left; scopeByMode is that visit's own scopeByMode field (filesNav's
	// doc comment) — every mode's last-used scope while this Root was active,
	// not only the outgoing mode's, so switching Root and mode independently
	// each restores its own last value (S7).
	mode        browseMode
	scopeByMode map[browseMode]browseScope
}

func newFilesNav() filesNav {
	return filesNav{expanded: map[string]bool{}, memory: map[string]filesNavMemory{}}
}

// setRoot points the pane at root. A Root the pane has never shown this
// session starts a fresh browse — collapsed at the top level, selection at
// the top (S5), Tree mode at its own default scope (S7: "every session
// starts on Tree", which a never-visited Root gets exactly the same as a
// brand new session) — the same as before M3.1. A Root shown before instead
// restores exactly the expansion, selection, Browse Mode and per-mode Browse
// Scope it was left with, however many other Roots were visited in between
// (M3.1/S3, S7/M4.6). Either way, the outgoing root's own state (if there
// was one) is captured into memory first, so leaving it is round-trippable
// the next time it becomes active.
func (f filesNav) setRoot(root *tree.Node) filesNav {
	if f.root != nil {
		f = f.remember(f.root.Path)
	}

	f.root = root
	if saved, ok := f.memory[root.Path]; ok {
		f.expanded = saved.expanded
		f.selected = f.indexOfPath(saved.selectedPath)
		f.mode = saved.mode
		f.scopeByMode = saved.scopeByMode
		f.scope = f.scopeForCurrentMode()
	} else {
		f.expanded = map[string]bool{}
		f.selected = 0
		f.mode = treeBrowse
		f.scopeByMode = nil
		f.scope = activeRootBrowseScope
	}
	return f
}

// remember snapshots the pane's current expansion, selection, Browse Mode
// and per-mode Browse Scope into memory under path (the outgoing root's own
// Path) — setRoot's own call, just before switching away from it. The
// expansion map is stored by reference, not cloned: withExpanded never
// mutates a map in place, it always builds a fresh one, so the copy kept
// here can never be changed out from under it by a later expand/collapse on
// the pane that moved on. withScopeRemembered folds the current mode's live
// scope into scopeByMode first, so the mode being left behind right now is
// never missing from what gets stored — the same reason cycleMode calls it
// too.
func (f filesNav) remember(path string) filesNav {
	f = f.withScopeRemembered()
	selPath := ""
	if n := f.selectedNode(); n != nil {
		selPath = n.Path
	}
	memory := make(map[string]filesNavMemory, len(f.memory)+1)
	for k, v := range f.memory {
		memory[k] = v
	}
	memory[path] = filesNavMemory{
		expanded:     f.expanded,
		selectedPath: selPath,
		mode:         f.mode,
		scopeByMode:  f.scopeByMode,
	}
	f.memory = memory
	return f
}

// indexOfPath is the visible-row index of the node at path against the
// pane's current root and expansion, or 0 when path is empty or no row
// matches. M3.1 only promises correctness when the tree shape underneath
// hasn't changed, in which case a remembered path always has a row; the
// zero fallback is defensive for the shape-changed case a refresh
// (refreshRoot/nearestSurvivingIndex, S2/C2a/M5.1) actually produces, but
// setRoot's own restore-from-memory call site never needs more than the
// fallback — M3.1's contract is exact-path-or-fresh, not ancestor-seeking.
func (f filesNav) indexOfPath(path string) int {
	if i := f.rowIndexOf(path); i >= 0 {
		return i
	}
	return 0
}

// rowIndexOf is the visible-row index of the node at path against the
// pane's current root and expansion, or -1 when path is empty or no row
// matches. The shared lookup indexOfPath and nearestSurvivingIndex both
// build on — the former stops at "no match", the latter keeps walking
// upward when it gets one.
func (f filesNav) rowIndexOf(path string) int {
	if path == "" {
		return -1
	}
	for i, row := range f.visibleRows() {
		if row.node.Path == path {
			return i
		}
	}
	return -1
}

// refreshRoot replaces the pane's tree with a freshly re-walked root (S2,
// C2a/M5.1), preserving the current expansion set as-is (a path for a
// folder that no longer exists is simply absent from visibleRows — dead
// weight no different from any other path this map has never pruned) and
// moving selection to the same node if it survives, or its nearest
// surviving ancestor otherwise (nearestSurvivingIndex). Unlike setRoot,
// this never touches memory: a refresh acts on the Active Root in place,
// not a Root switch, so there is nothing to remember or restore across.
func (f filesNav) refreshRoot(root *tree.Node) filesNav {
	selectedPath := ""
	if n := f.selectedNode(); n != nil {
		selectedPath = n.Path
	}
	f.root = root
	f.selected = f.nearestSurvivingIndex(selectedPath)
	return f
}

// nearestSurvivingIndex is refresh's own selection rule (S2, C2a/M5.1): the
// row for path if it still exists, else its parent, grandparent and so on
// until one does, else the top of the list. Every ancestor a visible row
// was ever selected under is, by construction, still expanded — a row is
// only visible because every ancestor between it and a top-level entry
// already is (renderRow/visibleRows) — so the walk only ever has to ask
// "does this path still exist", never "is it visible" separately.
func (f filesNav) nearestSurvivingIndex(path string) int {
	for path != "" && f.root != nil && path != f.root.Path {
		if i := f.rowIndexOf(path); i >= 0 {
			return i
		}
		parent := filepath.Dir(path)
		if parent == path {
			break
		}
		path = parent
	}
	return 0
}

// withStatus installs S8's per-file git status markers for the pane's
// current root. It is a separate setter from setRoot rather than a
// parameter to it: status is always freshly computed by the caller for
// every load (see the status field's own comment), never something setRoot
// itself would know how to produce or restore from memory.
func (f filesNav) withStatus(status map[string]gitFileStatus) filesNav {
	f.status = status
	return f
}

// withStarred installs S3's star-marker set (see the starred field's own
// comment) — the same "separate setter, caller recomputes it" shape
// withStatus already gives S8's markers.
func (f filesNav) withStarred(starred map[string]bool) filesNav {
	f.starred = starred
	return f
}

// withRecentSource installs the snapshot recentRows (M4.2) computes the
// pane's Recent-mode flat list from — see the docs/sessionRoots/
// activeRootPath fields' own comment for why this is a source snapshot
// rather than the rows themselves.
func (f filesNav) withRecentSource(docs map[string]docRecord, roots []root, activeRootPath string) filesNav {
	f.docs = docs
	f.sessionRoots = roots
	f.activeRootPath = activeRootPath
	return f
}

// cycleMode advances S5's Browse Mode one step (Tree -> Recent -> Starred ->
// Tree ...), keys.FilesMode's own action. The new mode's scope is whatever
// this Root last had it on (S7/M4.6: scopeForCurrentMode, fed by
// withScopeRemembered's snapshot of the outgoing mode's own scope just
// below) — a scope value only means something read against the mode that
// produced it (S5's own per-mode value sets), and modeScopes(mode)[0] (or
// the zero value for Tree, which scopeLabel ignores anyway) is
// scopeForCurrentMode's own fallback for a mode this Root has never shown.
func (f filesNav) cycleMode() filesNav {
	f = f.withScopeRemembered()
	switch f.mode {
	case treeBrowse:
		f.mode = recentBrowse
	case recentBrowse:
		f.mode = starredBrowse
	default:
		f.mode = treeBrowse
	}
	f.scope = f.scopeForCurrentMode()
	return f
}

// withScopeRemembered snapshots the current mode's live f.scope into
// scopeByMode (cloned, not mutated in place — the same value-semantic
// convention withExpanded already keeps) — the one point cycleMode and
// remember both pass through just before the pane moves off the mode it is
// currently showing, so that mode's scope is never lost to whichever comes
// next (S7/M4.6).
func (f filesNav) withScopeRemembered() filesNav {
	next := make(map[browseMode]browseScope, len(f.scopeByMode)+1)
	for k, v := range f.scopeByMode {
		next[k] = v
	}
	next[f.mode] = f.scope
	f.scopeByMode = next
	return f
}

// scopeForCurrentMode is the scope to show for f.mode right now: whatever
// scopeByMode remembers for it from this Root's current visit (S7/M4.6), or
// — a mode this Root has never shown yet — the mode's own first cycle value
// (modeScopes(mode)[0]), or the zero value for Tree (modeScopes returns
// none for it, and scopeLabel ignores the field regardless).
func (f filesNav) scopeForCurrentMode() browseScope {
	if s, ok := f.scopeByMode[f.mode]; ok {
		return s
	}
	if scopes := modeScopes(f.mode); len(scopes) > 0 {
		return scopes[0]
	}
	return activeRootBrowseScope
}

// cycleScope advances S5's Browse Scope one step within the current mode's
// own value set, wrapping; a no-op while the mode is Tree, whose scope key
// does nothing (S5: "the scope key is a no-op" for Tree) since
// modeScopes(treeBrowse) is empty.
func (f filesNav) cycleScope() filesNav {
	scopes := modeScopes(f.mode)
	if len(scopes) == 0 {
		return f
	}
	idx := 0
	for i, s := range scopes {
		if s == f.scope {
			idx = i
			break
		}
	}
	f.scope = scopes[(idx+1)%len(scopes)]
	return f
}

// scopeLabel is the current scope's own S5 display name for the mode/scope
// row: "-" for Tree regardless of the scope field's value (Tree has no
// scope to speak of), otherwise the scope's own label.
func (f filesNav) scopeLabel() string {
	if f.mode == treeBrowse {
		return "-"
	}
	return f.scope.label()
}

// setSize sets the content box the pane draws into.
func (f filesNav) setSize(width, height int) filesNav {
	f.width = max(width, 0)
	f.height = max(height, 0)
	return f
}

// fileRow is one rendered line: a node and how deep it sits, so the render can
// indent it. Rows are produced by flattening the tree to just its open part.
type fileRow struct {
	node  *tree.Node
	depth int
}

// visibleRows flattens the tree to the rows currently on screen: the root's
// children at depth 0, and a folder's children only when it is expanded. The
// root node itself is not a row — the pane shows what is *inside* the active
// root, titled by the pane, not the root as an entry.
func (f filesNav) visibleRows() []fileRow {
	if f.root == nil {
		return nil
	}
	var rows []fileRow
	var walk func(nodes []*tree.Node, depth int)
	walk = func(nodes []*tree.Node, depth int) {
		for _, n := range nodes {
			rows = append(rows, fileRow{node: n, depth: depth})
			if n.IsDir && f.expanded[n.Path] {
				walk(n.Children, depth+1)
			}
		}
	}
	walk(f.root.Children, 0)
	return rows
}

// selectedNode is the node under the selection, or nil when there are no rows.
func (f filesNav) selectedNode() *tree.Node {
	rows := f.visibleRows()
	if f.selected < 0 || f.selected >= len(rows) {
		return nil
	}
	return rows[f.selected].node
}

// flatMode reports whether the pane's current Browse Mode shows a flat list
// (Recent, M4.2; Starred, M4.3) rather than the Tree — the one predicate
// flatRows/rowCount/rowsView/selectedFlatRow all branch on, so a third flat
// mode would only ever need to be added here.
func (f filesNav) flatMode() bool {
	return f.mode == recentBrowse || f.mode == starredBrowse
}

// flatRows is the pane's current flat list (M4.2/M4.3): Recent's or
// Starred's own rows, freshly computed from the withRecentSource snapshot
// and the current scope, while the mode is one of the two; nil for Tree (no
// flat list).
func (f filesNav) flatRows() []browseRow {
	switch f.mode {
	case recentBrowse:
		return recentRows(f.docs, f.sessionRoots, f.activeRootPath, f.scope)
	case starredBrowse:
		return starredRows(f.docs, f.sessionRoots, f.activeRootPath, f.scope)
	default:
		return nil
	}
}

// selectedFlatRow is the browseRow under the selection cursor while the
// pane is in a flat-list mode, or false when the mode is Tree or the cursor
// is out of range (e.g. an empty list). model.activateInto/toggleDocStar use
// this to resolve "the selected document" against whichever list is
// actually on screen (TRIAGE T21's own citation for the mismatch this
// closes for those two actions) rather than selectedNode's tree-based row
// underneath — Peek/Edit/DV2-targeting (M4.5) still go through selectedNode
// until that task gives them the same fix.
func (f filesNav) selectedFlatRow() (browseRow, bool) {
	rows := f.flatRows()
	if f.selected < 0 || f.selected >= len(rows) {
		return browseRow{}, false
	}
	return rows[f.selected], true
}

// rowCount is however many rows move/rowsView have to work with for the
// pane's current mode: flatRows' own length for Recent/Starred, visibleRows'
// otherwise (Tree) — the one place both a tree selection and a flat-list
// selection are counted the same way, so move clamps against whichever list
// is actually on screen.
func (f filesNav) rowCount() int {
	if f.flatMode() {
		return len(f.flatRows())
	}
	return len(f.visibleRows())
}

// handleKey applies the Files Nav binding msg is bound to, reporting whether it
// was one. Selection movement reuses the shared up/down bindings; expand and
// collapse are the pane's own.
func (f filesNav) handleKey(msg tea.KeyPressMsg) (filesNav, bool) {
	switch {
	case key.Matches(msg, keys.LineUp):
		return f.move(-1), true
	case key.Matches(msg, keys.LineDown):
		return f.move(1), true
	case key.Matches(msg, keys.Expand):
		return f.expand(), true
	case key.Matches(msg, keys.Collapse):
		return f.collapse(), true
	case key.Matches(msg, keys.FilesMode):
		return f.cycleMode(), true
	case key.Matches(msg, keys.FilesScope):
		return f.cycleScope(), true
	}
	return f, false
}

// move shifts the selection by delta rows, clamped to the current mode's row
// set (rowCount).
func (f filesNav) move(delta int) filesNav {
	f.selected = moveSelection(f.selected, delta, f.rowCount())
	return f
}

// expand opens the selected folder in place. A file or an already-open folder
// is a no-op — as is the whole call while the pane is in a flat-list mode
// (Recent/Starred, M4.5): flatMode's rows have no folders to expand, and
// f.selected there indexes flatRows, not visibleRows, so selectedNode()
// would otherwise resolve an unrelated tree row hidden behind the flat list
// on screen (TRIAGE T21) and mutate its expansion invisibly.
func (f filesNav) expand() filesNav {
	if f.flatMode() {
		return f
	}
	n := f.selectedNode()
	if n != nil && n.IsDir && !f.expanded[n.Path] {
		f = f.withExpanded(n.Path, true)
	}
	return f
}

// collapse closes the selected folder in place. A file or an already-closed
// folder is a no-op; the selection stays on the folder, which keeps its index
// valid as its children vanish beneath it. Also a no-op in a flat-list mode
// (M4.5) — see expand's own comment for why.
func (f filesNav) collapse() filesNav {
	if f.flatMode() {
		return f
	}
	n := f.selectedNode()
	if n != nil && n.IsDir && f.expanded[n.Path] {
		f = f.withExpanded(n.Path, false)
	}
	return f
}

// withExpanded returns a copy with path's open state set, cloning the map so the
// update stays value-semantic (a stale copy of the pane keeps its own state).
func (f filesNav) withExpanded(path string, open bool) filesNav {
	next := make(map[string]bool, len(f.expanded)+1)
	for k, v := range f.expanded {
		next[k] = v
	}
	if open {
		next[path] = true
	} else {
		delete(next, path)
	}
	f.expanded = next
	return f
}

var (
	filesSelectedStyle = lipgloss.NewStyle().Reverse(true)
	emptyFilesStyle    = lipgloss.NewStyle().Foreground(dimColor)
	// modeScopeRowStyle is S5's mode/scope row (below): dim, like every
	// other pane's resting/empty text (dimColor's own doc comment), so it
	// reads as a status line rather than competing with the selected-row
	// reverse video or a git-status/star marker.
	modeScopeRowStyle = lipgloss.NewStyle().Foreground(dimColor)
)

// view renders the pane's content box at exactly its width x height, the same
// contract the viewer keeps. S5's mode/scope row (modeScopeRow) always takes
// the box's first line, in every mode and even with no root open (S5: "costing
// the file list one row of height in every mode" — the row itself, not just
// the modes it names, is what the pane always shows); the rest of the box
// (f.height-1) is the row list or the empty-state message exactly as before
// M4.1, just one row shorter.
func (f filesNav) view() string {
	if f.width <= 0 || f.height <= 0 {
		return ""
	}
	row := f.modeScopeRow()
	if f.height == 1 {
		// No room left for the list past the mode/scope row itself — join
		// would otherwise append an extra blank line (rowsView(0) == ""),
		// overshooting the exactly-f.height-lines contract.
		return row
	}
	return row + "\n" + f.rowsView(f.height-1)
}

// modeScopeRow renders S5's own status line: the current Browse Mode and
// Browse Scope, " \u00b7 "-joined the same way the footer joins its hints
// (footer.go's joinHints) so the two status lines read consistently.
func (f filesNav) modeScopeRow() string {
	text := f.mode.label() + " \u00b7 " + f.scopeLabel()
	return modeScopeRowStyle.Width(f.width).MaxWidth(f.width).Render(text)
}

// rowsView renders the file-list portion of the box at exactly height rows
// (the box height less modeScopeRow's own line) — the empty-state message
// when there is nothing to browse, or a scrollable window into the current
// mode's rows with the selection always on screen, otherwise. A height of 0
// (a pane too short to show anything past its mode/scope row) renders as an
// empty string, which "a\n" + "" still leaves as exactly f.height lines
// overall once split.
//
// Recent and Starred (M4.2/M4.3) render flatRows' own list here; Tree still
// renders visibleRows exactly as before — the empty-state message text is
// M4.4's own per-mode/scope-combination wording (flatEmptyText) for the
// former, emptyFilesText for the latter.
func (f filesNav) rowsView(height int) string {
	if height <= 0 {
		return ""
	}
	if f.flatMode() {
		return f.flatRowsView(height)
	}
	rows := f.visibleRows()
	if len(rows) == 0 {
		return emptyFilesStyle.
			Width(f.width).Height(height).
			MaxWidth(f.width).MaxHeight(height).
			Render(emptyFilesText)
	}
	return windowedView(f.width, height, f.selected, len(rows), func(idx int) string {
		return f.renderRow(rows[idx], idx == f.selected)
	})
}

// flatRowsView is rowsView's Recent/Starred-mode branch (M4.2/M4.3): the
// same empty-message-or-scrollable-window shape, over flatRows() instead of
// visibleRows(). Left as its own method (rather than folded into rowsView)
// so both flat modes share one copy of the windowedView plumbing.
func (f filesNav) flatRowsView(height int) string {
	rows := f.flatRows()
	if len(rows) == 0 {
		text := flatEmptyText(f.mode, f.scope)
		return emptyFilesStyle.
			Width(f.width).Height(height).
			MaxWidth(f.width).MaxHeight(height).
			Render(text)
	}
	return windowedView(f.width, height, f.selected, len(rows), func(idx int) string {
		return f.renderFlatRow(rows[idx], idx == f.selected)
	})
}

// renderFlatRow lays out one flat-list row (S6), the file name left-aligned
// and the right-hand label — the covering Root's base name, or the
// left-truncated parent directory for an orphan document — right-aligned
// with at least one gap column between the two. The two flat modes differ
// in the row's own left segment (S3): Recent gets the same star marker
// prefix Tree rows carry (Recent is not Starred, so membership alone does
// not already say it), while Starred omits it (membership already says it)
// and instead marks a broken row (S3's "broken starred document ... still
// listed and still marked, but cannot be activated") with a dimmed style and
// the same " (missing)" suffix rootsnav.go's renderRow gives a broken Root.
// The selected row draws in reverse video, the same bar convention
// renderRow's tree rows already use; each style call is fully specified
// across the row's own two segments (lipgloss-v2-row-rendering-gotchas.md)
// rather than applied to the joined line, so a wide label is never
// accidentally re-run through a second, un-reversed style call once joined.
func (f filesNav) renderFlatRow(r browseRow, selected bool) string {
	var left string
	broken := false
	if f.mode == starredBrowse {
		left = r.name
		if r.broken {
			left += " (missing)"
			broken = true
		}
	} else {
		star := "  "
		if f.starred[r.path] {
			star = "★ "
		}
		left = star + r.name
	}

	label := r.rootLabel
	if r.orphan {
		label = r.orphanDir
	}
	const gap = " "
	rightBudget := max(f.width-lipgloss.Width(left)-lipgloss.Width(gap), 0)
	right := leftTruncate(label, rightBudget)

	pad := f.width - lipgloss.Width(left) - lipgloss.Width(right)
	if pad < 0 {
		pad = 0
	}
	line := left + strings.Repeat(" ", pad) + right

	style := lipgloss.NewStyle()
	if broken {
		style = style.Faint(true)
	}
	if selected {
		style = style.Reverse(true)
	}
	return style.Width(f.width).MaxWidth(f.width).Render(line)
}

// renderRow lays out one row: indent by depth, an expand/collapse marker for a
// folder, then the name, padded and truncated to the box width. The selected
// row is drawn in reverse video so the whole line reads as a bar.
func (f filesNav) renderRow(r fileRow, selected bool) string {
	marker := "  "
	if r.node.IsDir {
		if f.expanded[r.node.Path] {
			marker = "▾ "
		} else {
			marker = "▸ "
		}
	}
	// Git status (S8) applies to files only — an expanded/collapsed folder
	// already occupies this pane's other marker column, and aggregating a
	// directory's descendants' statuses is out of scope.
	status := "  "
	if !r.node.IsDir {
		switch f.status[r.node.Path] {
		case gitStatusModified:
			status = "M "
		case gitStatusAdded:
			status = "A "
		}
	}
	// Star marker (S3), files only — a folder cannot itself be Starred (the
	// S1 store holds document records, not folders) — same "★ "/"  " pair the
	// Roots Nav's own renderRow already uses for row.root.starred, so the two
	// panes read the same mark for the same concept.
	star := "  "
	if !r.node.IsDir && f.starred[r.node.Path] {
		star = "★ "
	}
	line := strings.Repeat("  ", r.depth) + marker + status + star + r.node.Name

	style := lipgloss.NewStyle()
	if selected {
		style = filesSelectedStyle
	}
	return style.Width(f.width).MaxWidth(f.width).Render(line)
}
