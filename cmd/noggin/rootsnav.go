package main

import (
	"os"
	"path/filepath"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// wikipediaLabel and wikipediaDomain are the Wikipedia Source's two-line
// row (S1, C6/M1.2): always the Sources Nav's last row, present with zero
// or many Roots — the "No roots…" empty state (T26's old guidance) no
// longer occurs, since there is always at least this one Source to show.
const (
	wikipediaLabel  = "Wikipedia"
	wikipediaDomain = "en.wikipedia.org"
)

// rootRow pairs a root with whether its path currently exists on disk (S3's
// "broken" mark). broken is computed once, when the pane's roots are set —
// like the Files Nav's tree, the list is a snapshot for the pane to browse,
// not something that live-tracks the filesystem underneath it.
type rootRow struct {
	root   root
	broken bool
}

// rootsNav is the Roots Nav pane: a selection cursor over the session's roots
// (S1's order — ad-hoc root first, then starred — is preserved as given,
// never reordered here), independent of which root is active. storePath is
// resolved once at startup (main.go's storePath(), threaded through
// model.withStorePath, not this type's constructor — a fresh pane never
// starts with one, so newRootsNav takes no argument) so a star toggle can
// persist without re-deriving dataDir() on every keypress; empty means the
// data directory could not be resolved, in which case a toggle still
// updates the pane's own display but has nothing to write through. That is
// an uncommon case — no $HOME or $XDG_DATA_HOME — matching loadStarredRoots'
// own degrade-to-empty behavior for the same failure.
//
// wikipedia is S1's Active Source flag for the Wikipedia Query Source
// (C6/M1.1): true when Wikipedia, not any Root, is the session's Active
// Source. Mutually exclusive with active >= 0 — activate and
// activateWikipedia each clear the other's claim on "active" — so every
// existing Root-only reader of (rows, active) keeps working unmodified:
// active < 0 already meant "nothing to walk/search" before Wikipedia
// existed (an empty session), and it means exactly the same thing now that
// Wikipedia being active is one more way to get there. This field only
// exists so the *reason* active is -1 can be told apart where it matters
// (Sources Nav rendering, M1.2; Footer/Help labels, M1.3; Files Nav modes,
// M1.4) — none of those exist yet in M1.1, which only introduces the flag
// and the Root-only tolerance it buys for free.
type rootsNav struct {
	rows      []rootRow
	active    int // index into rows of the active root, or -1
	wikipedia bool
	selected  int // index into rows the cursor sits on

	storePath string

	width, height int
}

func newRootsNav() rootsNav {
	return rootsNav{active: -1}
}

// setRoots installs a fresh roots list and which one is active, resetting the
// selection cursor to the active root (or the top when none is active) —
// mirrors the Files Nav's setRoot "fresh browse" contract. It always clears
// the Wikipedia Active Source flag: a launch resolution (this method's only
// caller, model.applyResolution) never activates Wikipedia itself (C6/M1.1
// leaves that to an explicit activation, same as any Root), so a fresh
// session's Active Source is exactly what the resolution says, never a
// leftover Wikipedia claim from a value this rootsNav happened to start
// from.
func (r rootsNav) setRoots(roots []root, active int) rootsNav {
	rows := make([]rootRow, len(roots))
	for i, rt := range roots {
		rows[i] = rootRow{root: rt, broken: pathMissing(rt.path)}
	}
	r.rows = rows
	r.active = active
	r.wikipedia = false
	r.selected = max(active, 0)
	return r
}

// activePath is the active root's own Path, or "" when none is active —
// the (roots, active) shape's own scalar form the Files Nav's Recent/
// Starred flat lists need for their own Active Root scope (M4.2/M4.3),
// mirroring resolution.activeRootPath's identical shape for the launch-time
// version of the same question.
func (r rootsNav) activePath() string {
	if r.active < 0 || r.active >= len(r.rows) {
		return ""
	}
	return r.rows[r.active].root.path
}

// rootList is the session's roots in pane order, without the pane's own
// per-row display state (the broken mark) — the (roots, active) shape the
// search layer takes (scope.go's liveRootsInScope). The pane is where the
// session's roots actually live after applyResolution, so a Picker search
// reads them from here rather than the model keeping a second copy that
// could drift. broken is deliberately dropped: the search layer re-resolves
// liveness itself at search time, since this pane's mark is a snapshot taken
// when the roots were set (see rootRow).
func (r rootsNav) rootList() []root {
	roots := make([]root, len(r.rows))
	for i, row := range r.rows {
		roots[i] = row.root
	}
	return roots
}

// pathMissing reports whether path cannot be stat'd — the pane's definition
// of "broken" (S3). It does not distinguish "does not exist" from other stat
// failures (e.g. permission denied): either way the root is unusable to
// browse, which is the only distinction S3 asks for.
func pathMissing(path string) bool {
	_, err := os.Stat(path)
	return err != nil
}

// setSize sets the content box the pane draws into.
func (r rootsNav) setSize(width, height int) rootsNav {
	r.width = max(width, 0)
	r.height = max(height, 0)
	return r
}

// selectedRow is the Root row under the selection cursor, or false when
// there are no Root rows or the cursor sits on the Wikipedia row instead
// (selectedIsWikipedia) — callers that only know how to act on a Root
// (model.activateRoot, model.toggleStar) already treat that "false" as a
// no-op, so they need a Wikipedia-aware guard in front of this, not a
// different return here.
func (r rootsNav) selectedRow() (rootRow, bool) {
	if r.selected < 0 || r.selected >= len(r.rows) {
		return rootRow{}, false
	}
	return r.rows[r.selected], true
}

// rowCount is the Sources Nav's total row count (S1, C6/M1.2): every Root
// plus the Wikipedia Source, which is always present — the pane never has
// zero rows, even when r.rows itself is empty. Selection movement and
// rendering both walk indices [0, rowCount()), with len(r.rows) itself
// being the one reserved for Wikipedia.
func (r rootsNav) rowCount() int {
	return len(r.rows) + 1
}

// selectedIsWikipedia reports whether the selection cursor sits on the
// Sources Nav's Wikipedia row (S1, C6/M1.2) rather than a Root row.
func (r rootsNav) selectedIsWikipedia() bool {
	return r.selected == len(r.rows)
}

// activate makes rows[i] the active root. The caller (model.activateRoot)
// checks broken and does the actual re-rooting; this just updates which
// index the pane marks active. It clears the Wikipedia Active Source flag
// (C6/M1.1) — a Root and Wikipedia are mutually exclusive as the one Active
// Source, so activating a Root always displaces Wikipedia if it was active.
func (r rootsNav) activate(i int) rootsNav {
	r.active = i
	r.wikipedia = false
	return r
}

// activateWikipedia makes the Wikipedia Query Source the Active Source
// (C6/S1, M1.1): the mirror of activate, for the one Source that is never a
// row index. It clears active to -1 rather than leaving whatever Root was
// last active — a Root and Wikipedia are mutually exclusive as the Active
// Source, and every existing Root-only reader of (rows, active) already
// treats active < 0 as "nothing to walk/search" (scope.go's
// liveRootsInScope, the tree/git/staleness/Refresh call sites gated on
// model.files.root/m.files.root != nil) — the same tolerance an empty
// session already relied on, now doing double duty for Wikipedia too.
func (r rootsNav) activateWikipedia() rootsNav {
	r.active = -1
	r.wikipedia = true
	return r
}

// activeIsWikipedia reports whether the Wikipedia Query Source, rather than
// any Root (or nothing), is the session's Active Source.
func (r rootsNav) activeIsWikipedia() bool {
	return r.wikipedia
}

// toggleStarred flips rows[i]'s starred mark. The caller (model.toggleStar)
// is responsible for persisting the change via the S2 store first — this
// only updates the pane's own display.
func (r rootsNav) toggleStarred(i int) rootsNav {
	if i < 0 || i >= len(r.rows) {
		return r
	}
	rows := make([]rootRow, len(r.rows))
	copy(rows, r.rows)
	rows[i].root.starred = !rows[i].root.starred
	r.rows = rows
	return r
}

// handleKey applies the Roots Nav binding msg is bound to, reporting whether
// it was one. Only selection movement lives here — Activate and Star have
// cross-cutting effects (re-rooting the Files Nav, persisting to disk) that
// belong to the model, mirroring how the Files Nav's own Activate is handled
// one level up (M2.3).
func (r rootsNav) handleKey(msg tea.KeyPressMsg) (rootsNav, bool) {
	switch {
	case key.Matches(msg, keys.LineUp):
		return r.move(-1), true
	case key.Matches(msg, keys.LineDown):
		return r.move(1), true
	}
	return r, false
}

// move shifts the selection cursor by delta rows, clamped to the row set
// — every Root plus the always-present Wikipedia row (rowCount).
func (r rootsNav) move(delta int) rootsNav {
	r.selected = moveSelection(r.selected, delta, r.rowCount())
	return r
}

// rootRowHeight is how many physical lines one Roots Nav entry occupies
// (S5): name-and-markers, then path. windowedViewRows uses this to window
// and scroll in whole entries, never splitting one across the screen edge.
const rootRowHeight = 2

var (
	rootsBrokenStyle = lipgloss.NewStyle().Faint(true)
	rootsNameStyle   = lipgloss.NewStyle().Bold(true)
	rootsPathStyle   = lipgloss.NewStyle()
)

// view renders the pane's content box at exactly its width x height, the same
// contract the Files Nav and viewer keep. Unlike before C6/M1.2, there is no
// empty state: the Wikipedia row (rowCount) means the Sources Nav always has
// at least one row to show, with zero or many Roots alike (S1).
func (r rootsNav) view() string {
	if r.width <= 0 || r.height <= 0 {
		return ""
	}
	return windowedViewRows(r.width, r.height, r.selected, r.rowCount(), rootRowHeight, r.renderRow)
}

// renderRow lays out one entry across its two lines (S5). Line 1 carries the
// active marker, the starred marker, the base name in bold (S3's
// "prominently") and the broken-root suffix. Line 2 carries the entry's full
// path alone, unbold and left-truncated the same way the single-line layout
// truncated its parent path, so same-named roots still read apart
// (`…/work/admin/project`) — markers never repeat on line 2, per S5. A
// broken root is dimmed on both lines, since it cannot be activated. The
// selection cursor draws both lines in reverse video, matching the Files
// Nav's bar; each line's style is fully specified (not a partial run), so
// reverse video never gets cancelled mid-line
// (lipgloss-v2-row-rendering-gotchas.md).
func (r rootsNav) renderRow(idx int) []string {
	if idx == len(r.rows) {
		return r.renderWikipediaRow()
	}
	row := r.rows[idx]

	activeMark := "  "
	if idx == r.active {
		activeMark = "● "
	}
	starMark := "  "
	if row.root.starred {
		starMark = "★ "
	}

	suffix := ""
	if row.broken {
		suffix = " (missing)"
	}
	line1 := activeMark + starMark + filepath.Base(row.root.path) + suffix
	// line2 is already protected by leftTruncate below; line1 needs the same
	// guard — a long base name/suffix combination would otherwise word-wrap
	// under Width().Render() instead of truncating
	// (lipgloss-v2-row-rendering-gotchas.md), the same class of bug fixed in
	// files.go's renderRow/renderFlatRow.
	line1 = rightTruncate(line1, r.width)
	line2 := leftTruncate(row.root.path, r.width)

	style1, style2 := rootsNameStyle, rootsPathStyle
	if row.broken {
		style1, style2 = rootsBrokenStyle, rootsBrokenStyle
	}
	if idx == r.selected {
		style1 = style1.Reverse(true)
		style2 = style2.Reverse(true)
	}
	return []string{
		style1.Width(r.width).MaxWidth(r.width).Render(line1),
		style2.Width(r.width).MaxWidth(r.width).Render(line2),
	}
}

// renderWikipediaRow lays out the Sources Nav's always-present, always-last
// Wikipedia row (S1, C6/M1.2): line 1 is the active marker (when Wikipedia
// is the Active Source) plus the label "Wikipedia", bold like every other
// row's name line; line 2 is its domain, de-emphasized the same way a
// Root's path line is — same rootsPathStyle, not a heavier dim, matching
// S1's "de-emphasized like a Root's path" exactly. Neither the starred nor
// broken marker ever applies here (S1: "nothing can add, remove or hide
// it"; it cannot be starred at all, so it is never Broken either) — unlike
// renderRow, there is no per-row state to read either from, only r.wikipedia
// and the selection cursor.
func (r rootsNav) renderWikipediaRow() []string {
	activeMark := "  "
	if r.wikipedia {
		activeMark = "● "
	}
	line1 := rightTruncate(activeMark+wikipediaLabel, r.width)
	line2 := leftTruncate(wikipediaDomain, r.width)

	style1, style2 := rootsNameStyle, rootsPathStyle
	if r.selectedIsWikipedia() {
		style1 = style1.Reverse(true)
		style2 = style2.Reverse(true)
	}
	return []string{
		style1.Width(r.width).MaxWidth(r.width).Render(line1),
		style2.Width(r.width).MaxWidth(r.width).Render(line2),
	}
}

// leftTruncate keeps s's suffix, replacing the front with an ellipsis when it
// does not fit in width, so the folder nearest the entry's name stays visible
// (`…/work/admin/`) — what distinguishes two same-named roots (S3).
func leftTruncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	if width == 1 {
		return "…"
	}
	runes := []rune(s)
	keep := width - 1
	if keep > len(runes) {
		keep = len(runes)
	}
	return "…" + string(runes[len(runes)-keep:])
}
