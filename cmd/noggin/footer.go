package main

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
)

// paneHints lists the key bindings relevant to a focused pane, most important
// first (S11's "key hints relevant to the focused pane"). Order matters:
// hintText drops from the end of this list first when the footer must
// truncate, so put what a newcomer needs most up front.
func paneHints(p pane) []key.Binding {
	switch p {
	case rootsPane:
		return []key.Binding{keys.Activate, keys.Star, keys.LineUp, keys.LineDown, keys.NextPane, keys.Edit}
	case filesPane:
		return []key.Binding{keys.Activate, keys.Expand, keys.Collapse, keys.LineUp, keys.LineDown, keys.NextPane, keys.Edit}
	default: // viewerPane
		return []key.Binding{keys.LineUp, keys.LineDown, keys.PageUp, keys.PageDown, keys.Top, keys.Bottom, keys.Edit, keys.NextPane}
	}
}

// alwaysHints appear last on every footer, in this order, and are never
// dropped by hintText's truncation: S11 requires quit always shown, and the
// help affordance advertised even when nothing else fits.
var alwaysHints = []key.Binding{keys.Help, keys.Quit}

// hintText builds the footer's key-hint line for a focused pane at a given
// width: the pane's title, then as many of its hints as fit, then
// alwaysHints. It tries the full hint set first and drops pane-specific
// hints one at a time from the end (least essential, per paneHints' order)
// until the line fits or none are left — "truncate gracefully" (S11).
// alwaysHints are appended at every step, so quit and the help affordance
// survive even the narrowest footer that still shows anything at all. If
// even the title plus alwaysHints alone still overflow, the title itself is
// the last thing dropped, so those two never fail to show for want of space.
func hintText(p pane, width int) string {
	hints := paneHints(p)
	for n := len(hints); n >= 0; n-- {
		text := joinHints(p.title(), hints[:n])
		if lipgloss.Width(text) <= width {
			return text
		}
	}
	return joinHints("", nil)
}

// joinHints renders title and hints (plus alwaysHints) as one " · "-joined
// line, the shape every footer and the future help modal (S12) can share. An
// empty title is omitted rather than leaving a stray leading separator.
func joinHints(title string, hints []key.Binding) string {
	parts := make([]string, 0, len(hints)+len(alwaysHints)+1)
	if title != "" {
		parts = append(parts, title)
	}
	for _, h := range hints {
		parts = append(parts, formatHint(h))
	}
	for _, h := range alwaysHints {
		parts = append(parts, formatHint(h))
	}
	return strings.Join(parts, " · ")
}

// formatHint renders one binding as "<key> <description>", the same shape
// the pre-M4.2 footer used for its lone quit hint.
func formatHint(k key.Binding) string {
	h := k.Help()
	return h.Key + " " + h.Desc
}
