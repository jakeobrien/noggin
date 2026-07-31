package main

import (
	"os"
	"path/filepath"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// emptyRootsText is what the Roots Nav pane says when the session has no
// roots at all — a bare launch with no starred roots yet (S2's "empty root
// list" outcome, unreached by anything else this cycle).
const emptyRootsText = "No roots."

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
// persist without re-deriving os.UserConfigDir() on every keypress; empty
// means the config directory could not be resolved, in which case a toggle
// still updates the pane's own display but has nothing to write through.
// That is an uncommon case — no $HOME or $XDG_CONFIG_HOME — matching
// loadStarredRoots' own degrade-to-empty behavior for the same failure.
type rootsNav struct {
	rows     []rootRow
	active   int // index into rows of the active root, or -1
	selected int // index into rows the cursor sits on

	storePath string

	width, height int
}

func newRootsNav() rootsNav {
	return rootsNav{active: -1}
}

// setRoots installs a fresh roots list and which one is active, resetting the
// selection cursor to the active root (or the top when none is active) —
// mirrors the Files Nav's setRoot "fresh browse" contract.
func (r rootsNav) setRoots(roots []root, active int) rootsNav {
	rows := make([]rootRow, len(roots))
	for i, rt := range roots {
		rows[i] = rootRow{root: rt, broken: pathMissing(rt.path)}
	}
	r.rows = rows
	r.active = active
	r.selected = max(active, 0)
	return r
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

// selectedRow is the row under the selection cursor, or false when there are
// no rows.
func (r rootsNav) selectedRow() (rootRow, bool) {
	if r.selected < 0 || r.selected >= len(r.rows) {
		return rootRow{}, false
	}
	return r.rows[r.selected], true
}

// activate makes rows[i] the active root. The caller (model.activateRoot)
// checks broken and does the actual re-rooting; this just updates which
// index the pane marks active.
func (r rootsNav) activate(i int) rootsNav {
	r.active = i
	return r
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

// move shifts the selection cursor by delta rows, clamped to the row set.
func (r rootsNav) move(delta int) rootsNav {
	r.selected = moveSelection(r.selected, delta, len(r.rows))
	return r
}

var (
	rootsActiveStyle = lipgloss.NewStyle().Bold(true)
	rootsBrokenStyle = lipgloss.NewStyle().Faint(true)
	emptyRootsStyle  = lipgloss.NewStyle().Foreground(dimColor)
)

// view renders the pane's content box at exactly its width x height, the same
// contract the Files Nav and viewer keep.
func (r rootsNav) view() string {
	if r.width <= 0 || r.height <= 0 {
		return ""
	}
	if len(r.rows) == 0 {
		return emptyRootsStyle.
			Width(r.width).Height(r.height).
			MaxWidth(r.width).MaxHeight(r.height).
			Render(emptyRootsText)
	}
	return windowedView(r.width, r.height, r.selected, len(r.rows), r.renderRow)
}

// renderRow lays out one entry: an active marker, a starred marker, the base
// name (S3's "prominently"), then its left-truncated parent path so
// same-named roots read apart (`…/work/admin/`). A broken root is marked in
// text (S3) and dimmed, since it cannot be activated. The selection cursor
// draws the whole line in reverse video, matching the Files Nav's bar.
func (r rootsNav) renderRow(idx int) string {
	row := r.rows[idx]

	activeMark := "  "
	if idx == r.active {
		activeMark = "● "
	}
	starMark := "  "
	if row.root.starred {
		starMark = "★ "
	}

	name := filepath.Base(row.root.path)
	parent := filepath.Dir(row.root.path) + string(filepath.Separator)

	const gap = "  "
	suffix := ""
	if row.broken {
		suffix = " (missing)"
	}
	prefixWidth := lipgloss.Width(activeMark) + lipgloss.Width(starMark)
	budget := max(r.width-prefixWidth-lipgloss.Width(name)-lipgloss.Width(gap)-lipgloss.Width(suffix), 0)
	parent = leftTruncate(parent, budget)

	line := activeMark + starMark + name + gap + parent + suffix

	style := lipgloss.NewStyle()
	switch {
	case row.broken:
		style = rootsBrokenStyle
	case idx == r.active:
		style = rootsActiveStyle
	}
	if idx == r.selected {
		style = style.Reverse(true)
	}
	return style.Width(r.width).MaxWidth(r.width).Render(line)
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
