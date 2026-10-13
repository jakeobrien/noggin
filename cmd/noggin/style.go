package main

import (
	"charm.land/lipgloss/v2"
)

// The styles below are the ones several files share, derived from the active
// Theme (theme.go). Package-level style vars elsewhere read theme.Subtle and
// friends directly; this file holds the ones with a story.
//
// theme is a package var initialised before these (Go orders package-level
// initialisation by dependency), so reading it here is safe.

// staleRowStyle is S9/C3/M5.1's dedicated staleness row: reverse video
// (Reverse swaps the specified Foreground into the row's background) in the
// Theme's Highlight colour, which the Help Modal's border also uses — "a colour
// no pane border uses" — reused here rather than adding a second, arbitrary
// "pay attention" colour.
var staleRowStyle = lipgloss.NewStyle().Foreground(theme.Highlight).Reverse(true)

// locationLineStyle is S9/S9a's own de-emphasized bottom row (M8.2): the
// same Subtle colour every pane's empty-state/footer text already uses, plain
// (no Reverse) unlike staleRowStyle — S9 never calls it out for attention
// the way S10's staleness row does, only asks that it read as secondary.
var locationLineStyle = lipgloss.NewStyle().Foreground(theme.Subtle)
