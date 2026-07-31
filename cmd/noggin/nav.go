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
// rootsNav.view and filesNav.view (T16), which differed only in how a row at
// a given index was rendered — that is the one thing renderRow supplies.
func windowedView(width, height, selected, count int, renderRow func(idx int) string) string {
	offset := 0
	if selected >= height {
		offset = selected - height + 1
	}
	lines := make([]string, height)
	for i := range lines {
		idx := offset + i
		if idx < count {
			lines[i] = renderRow(idx)
		} else {
			lines[i] = strings.Repeat(" ", width)
		}
	}
	return strings.Join(lines, "\n")
}
