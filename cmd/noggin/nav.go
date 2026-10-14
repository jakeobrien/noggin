package main

import "strings"

// moveSelection shifts a list-nav selection cursor by delta rows, clamped to
// [0, count-1], or leaves selected untouched when there is nothing to select.
// Shared by rootsNav.move and filesNav.move (T16), whose only difference was
// which slice they counted.
func moveSelection(selected, delta, count int) int {
	if count == 0 {
		return selected
	}
	return clamp(selected+delta, 0, count-1)
}

// scrollMargin is how many entries at each edge of a perScreen-entry window
// are the "margin zone": the cursor may roam freely between the two margin
// lines, and the window scrolls only when it enters a margin (tk-ew1). About
// a third of the pane, floored, and capped so the two margins never overlap
// (a pane of 1 or 2 entries has none).
func scrollMargin(perScreen int) int {
	return max(min(perScreen/3, (perScreen-1)/2), 0)
}

// scrollOffset is the sidebar panes' margin-scrolling rule (tk-ew1): given
// the window's remembered first-visible entry (offset), the cursor and the
// list/window sizes in entries, it returns the offset that keeps the cursor
// on screen, moved as little as possible and only when the cursor has entered
// a margin zone (scrollMargin). The result is clamped so no blank space
// shows past either end of a list taller than the window; a list that fits
// always starts at 0.
func scrollOffset(offset, selected, count, perScreen int) int {
	if perScreen <= 0 || count <= perScreen {
		return 0
	}
	maxOffset := count - perScreen
	selected = clamp(selected, 0, count-1)
	margin := scrollMargin(perScreen)
	offset = clamp(offset, 0, maxOffset)
	if top := selected - margin; selected < offset+margin {
		offset = top
	} else if bottom := selected - (perScreen - 1 - margin); selected > offset+perScreen-1-margin {
		offset = bottom
	}
	return clamp(offset, 0, maxOffset)
}

// edgePinnedOffset is the stateless scrolling the Picker's lists keep: the
// selection is scrolled just onto the last row and no further. The sidebar
// panes use scrollOffset instead.
func edgePinnedOffset(selected, perScreen int) int {
	if selected >= perScreen {
		return selected - perScreen + 1
	}
	return 0
}

// windowedView renders a scrollable row list's content box with the
// Picker's edge-pinned scrolling (see edgePinnedOffset): the selection stays
// on screen once the list outgrows height, and short lists pad to height
// with blank rows.
func windowedView(width, height, selected, count int, renderRow func(idx int) string) string {
	return windowedViewRows(width, height, edgePinnedOffset(selected, height), count, 1, func(idx int) []string {
		return []string{renderRow(idx)}
	})
}

// windowedViewRows renders count entries of rowHeight physical lines each
// (S5's two-line Sources Nav rows; 1 for single-line lists) starting at
// entry offset. Scrolling and the on-screen window are computed in entries,
// not lines, so an entry is never shown split across the top/bottom edge: at
// most height/rowHeight whole entries are drawn; any leftover lines that
// don't make a whole entry pad blank rather than showing a cut-off row.
// Choosing offset (so the selection is visible) is the caller's job — see
// scrollOffset. renderRow must return exactly rowHeight lines for a given idx.
func windowedViewRows(width, height, offset, count, rowHeight int, renderRow func(idx int) []string) string {
	// No forced minimum of 1 here (T18): a pane shorter than rowHeight must
	// show zero entries, not a whole row rendered then sliced down to a
	// truncated first line. perScreen=0 makes the loop below a no-op and
	// the pad loop fills the box with blank rows instead.
	perScreen := height / rowHeight
	lines := make([]string, 0, height)
	for i := 0; i < perScreen && offset+i < count; i++ {
		lines = append(lines, renderRow(offset+i)...)
	}
	for len(lines) < height {
		lines = append(lines, strings.Repeat(" ", width))
	}
	return strings.Join(lines[:height], "\n")
}
