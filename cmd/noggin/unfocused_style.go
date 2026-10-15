package main

import (
	"charm.land/lipgloss/v2"
)

// Unfocused-pane styling (S11). The focused pane is marked by a thick
// border and a title bar (model.renderPane); an unfocused pane additionally
// recedes:
//
//   - a nav pane's selected row is a dim grey bar instead of a bright
//     reverse-video one, and its other text takes a muted foreground;
//   - a Document Viewer shows a desaturated Glamour render of its document,
//     from a stylesheet derived from the focused one (mutedMarkdownStyleConfig).
//
// None of it post-processes ANSI: every choice is made where the style is
// built (or, for documents, by rendering a second time from a derived
// stylesheet), so nothing fights lipgloss's reset/reflow behaviour
// (lipgloss-reflow-sgr-continuity.md). The one frame-wide pass, paintCanvas
// (canvas.go), only fills in colours nothing chose.

// selectedRowStyle styles a nav pane's selected row: the usual bright
// reverse-video bar when the pane has focus, a plain grey background when it
// does not. An explicit background rather than Reverse+Faint: Faint's effect
// on a reverse-video row is terminal-specific (some dim the text, some the
// bar), whereas a grey background dims the bar itself and leaves the text at
// full intensity. No Reverse is involved, so there is nothing to swap.
func selectedRowStyle(st lipgloss.Style, unfocused bool) lipgloss.Style {
	if unfocused {
		return st.Background(theme.UnfocusedBar)
	}
	return st.Reverse(true)
}

// mutedText is the nav-pane counterpart of the muted Glamour style: it
// gives a non-selected row's style a muted foreground while its pane is
// unfocused. dim marks text that is already de-emphasised (paths, mode/scope
// line, empty-state text), which gets an extra step down so the hierarchy
// survives. Selected rows are never passed through here, so the selection
// treatment stays independent of it.
func mutedText(st lipgloss.Style, on, dim bool) lipgloss.Style {
	if !on {
		return st
	}
	if dim {
		return st.Foreground(theme.UnfocusedSubtle)
	}
	return st.Foreground(theme.UnfocusedText)
}
