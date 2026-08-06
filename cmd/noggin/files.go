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

	// memory holds, per Root, the expansion set and selected node's Path last
	// left there (M3.1/S3), keyed by the Root's own Path — the same value
	// model.activateRoot re-walks into a brand new *tree.Node each time a Root
	// is (re)activated, even when nothing on disk changed, so memory is keyed
	// by Path rather than Node identity, the same way expanded already is. A
	// Root with no entry here has never been visited this session.
	memory map[string]filesNavMemory

	// status holds S8's per-file git markers for the current root, keyed by
	// absolute Path (gitStatus's return shape, ready to match tree.Node.Path
	// directly). Unlike memory, it is never restored from a prior visit — the
	// caller (model.openFolder) recomputes it fresh on every load, since the
	// underlying git state can change between visits to the same Root. Nil
	// (the zero value) renders identically to an empty map: no markers.
	status map[string]gitFileStatus

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
}

func newFilesNav() filesNav {
	return filesNav{expanded: map[string]bool{}, memory: map[string]filesNavMemory{}}
}

// setRoot points the pane at root. A Root the pane has never shown this
// session starts a fresh browse — collapsed at the top level, selection at
// the top (S5) — the same as before M3.1. A Root shown before instead
// restores exactly the expansion and selection it was left with, however
// many other Roots were visited in between (M3.1/S3). Either way, the
// outgoing root's own state (if there was one) is captured into memory
// first, so leaving it is round-trippable the next time it becomes active.
func (f filesNav) setRoot(root *tree.Node) filesNav {
	if f.root != nil {
		f = f.remember(f.root.Path)
	}

	f.root = root
	if saved, ok := f.memory[root.Path]; ok {
		f.expanded = saved.expanded
		f.selected = f.indexOfPath(saved.selectedPath)
	} else {
		f.expanded = map[string]bool{}
		f.selected = 0
	}
	return f
}

// remember snapshots the pane's current expansion and selection into memory
// under path (the outgoing root's own Path) — setRoot's own call, just
// before switching away from it. The expansion map is stored by reference,
// not cloned: withExpanded never mutates a map in place, it always builds a
// fresh one, so the copy kept here can never be changed out from under it by
// a later expand/collapse on the pane that moved on.
func (f filesNav) remember(path string) filesNav {
	selPath := ""
	if n := f.selectedNode(); n != nil {
		selPath = n.Path
	}
	memory := make(map[string]filesNavMemory, len(f.memory)+1)
	for k, v := range f.memory {
		memory[k] = v
	}
	memory[path] = filesNavMemory{expanded: f.expanded, selectedPath: selPath}
	f.memory = memory
	return f
}

// indexOfPath is the visible-row index of the node at path against the
// pane's current root and expansion, or 0 when path is empty or no row
// matches. M3.1 only promises correctness when the tree shape underneath
// hasn't changed, in which case a remembered path always has a row; the
// zero fallback is defensive for the shape-changed case M5's refresh will
// refine into "nearest surviving ancestor", not a case M3.1 itself produces.
func (f filesNav) indexOfPath(path string) int {
	if path == "" {
		return 0
	}
	for i, row := range f.visibleRows() {
		if row.node.Path == path {
			return i
		}
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
	f.selected = moveSelection(f.selected, delta, len(f.visibleRows()))
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
	emptyFilesStyle    = lipgloss.NewStyle().Foreground(dimColor)
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
	return windowedView(f.width, f.height, f.selected, len(rows), func(idx int) string {
		return f.renderRow(rows[idx], idx == f.selected)
	})
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
	line := strings.Repeat("  ", r.depth) + marker + status + r.node.Name

	style := lipgloss.NewStyle()
	if selected {
		style = filesSelectedStyle
	}
	return style.Width(f.width).MaxWidth(f.width).Render(line)
}
