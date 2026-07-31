package main

import (
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

	// width and height are the content box the pane draws into, borders and
	// title already deducted (see computeLayout / rect.content).
	width  int
	height int
}

func newFilesNav() filesNav {
	return filesNav{expanded: map[string]bool{}}
}

// setRoot points the pane at a new tree, collapsing everything and resetting
// the selection to the top — re-rooting is a fresh browse (M3.4 leans on this).
func (f filesNav) setRoot(root *tree.Node) filesNav {
	f.root = root
	f.expanded = map[string]bool{}
	f.selected = 0
	return f
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
	}
	return f, false
}

// move shifts the selection by delta rows, clamped to the visible set.
func (f filesNav) move(delta int) filesNav {
	rows := f.visibleRows()
	if len(rows) == 0 {
		return f
	}
	f.selected = clamp(f.selected+delta, 0, len(rows)-1)
	return f
}

// expand opens the selected folder in place. A file or an already-open folder
// is a no-op.
func (f filesNav) expand() filesNav {
	n := f.selectedNode()
	if n != nil && n.IsDir && !f.expanded[n.Path] {
		f = f.withExpanded(n.Path, true)
	}
	return f
}

// collapse closes the selected folder in place. A file or an already-closed
// folder is a no-op; the selection stays on the folder, which keeps its index
// valid as its children vanish beneath it.
func (f filesNav) collapse() filesNav {
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
	emptyFilesStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

// view renders the pane's content box at exactly its width x height, the same
// contract the viewer keeps. A window into the visible rows is drawn with the
// selection always on screen; short trees pad to the box height.
func (f filesNav) view() string {
	if f.width <= 0 || f.height <= 0 {
		return ""
	}
	rows := f.visibleRows()
	if len(rows) == 0 {
		return emptyFilesStyle.
			Width(f.width).Height(f.height).
			MaxWidth(f.width).MaxHeight(f.height).
			Render(emptyFilesText)
	}

	// Scroll just enough to keep the selection in view: the selected row sits
	// at the bottom edge once the list is taller than the box.
	offset := 0
	if f.selected >= f.height {
		offset = f.selected - f.height + 1
	}

	lines := make([]string, f.height)
	for i := range lines {
		idx := offset + i
		if idx < len(rows) {
			lines[i] = f.renderRow(rows[idx], idx == f.selected)
		} else {
			lines[i] = strings.Repeat(" ", f.width)
		}
	}
	return strings.Join(lines, "\n")
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
	line := strings.Repeat("  ", r.depth) + marker + r.node.Name

	style := lipgloss.NewStyle()
	if selected {
		style = filesSelectedStyle
	}
	return style.Width(f.width).MaxWidth(f.width).Render(line)
}
