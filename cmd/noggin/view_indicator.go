package main

import (
	"strings"

	"charm.land/lipgloss/v2"

	"noggin/internal/tree"
)

// The view-toggle indicator is a short, right-aligned run of plain
// text on the Files Nav status row and the Picker's title, one slot per view
// toggle: `h` when hidden files are shown (else a space), `i` when ignored
// files are shown (else a space), then ` md` or `txt` for the File Filter —
// padded to a fixed width so the slots never shift as values change. Slots
// are separated by single spaces; a context where a toggle has no effect
// omits its slot entirely (pickerIndicator).

func hiddenSlot(hideDotfiles bool) string {
	if hideDotfiles {
		return " "
	}
	return "h"
}

func ignoredSlot(showIgnored bool) string {
	if showIgnored {
		return "i"
	}
	return " "
}

func filterSlot(f tree.Filter) string {
	if f == tree.FilterAll {
		return "txt"
	}
	return " md"
}

// viewIndicator joins the given slots into the indicator.
func viewIndicator(slots ...string) string {
	return strings.Join(slots, " ")
}

// statusLine lays left and right out on one line exactly width cells wide,
// right flush against the right edge. Under width pressure right is dropped
// first, then left is truncated.
func statusLine(left, right string, width int) string {
	if width <= 0 {
		return ""
	}
	lw, rw := lipgloss.Width(left), lipgloss.Width(right)
	if right != "" && lw+1+rw <= width {
		return left + strings.Repeat(" ", width-lw-rw) + right
	}
	left = rightTruncate(left, width)
	return left + strings.Repeat(" ", width-lipgloss.Width(left))
}
