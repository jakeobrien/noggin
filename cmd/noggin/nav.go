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

// windowedView renders a scrollable row list's content box: the selection
// stays on screen once the list outgrows height, scrolled just enough (never
// re-centred), and short lists pad to height with blank rows. Shared by
// filesNav.view (single-line rows) and, via windowedViewRows, rootsNav.view
// (S5's two-line rows) — a thin one-line special case of the general form.
func windowedView(width, height, selected, count int, renderRow func(idx int) string) string {
	return windowedViewRows(width, height, selected, count, 1, func(idx int) []string {
		return []string{renderRow(idx)}
	})
}

// windowedViewRows generalizes windowedView to entries that occupy more than
// one physical line each (rowHeight lines per entry — S5's two-line Roots
// Nav rows). Scrolling and the on-screen window are computed in entries, not
// lines, so an entry is never shown split across the top/bottom edge: at
// most height/rowHeight whole entries are drawn, the selected one kept fully
// on screen (scrolled just enough, never re-centred); any leftover lines
// that don't make a whole entry pad blank rather than showing a cut-off row.
// renderRow must return exactly rowHeight lines for a given idx.
func windowedViewRows(width, height, selected, count, rowHeight int, renderRow func(idx int) []string) string {
	perScreen := max(height/rowHeight, 1)
	offset := 0
	if selected >= perScreen {
		offset = selected - perScreen + 1
	}
	lines := make([]string, 0, height)
	for i := 0; i < perScreen && offset+i < count; i++ {
		lines = append(lines, renderRow(offset+i)...)
	}
	for len(lines) < height {
		lines = append(lines, strings.Repeat(" ", width))
	}
	return strings.Join(lines[:height], "\n")
}
