package main

import "charm.land/lipgloss/v2"

// dimColor is the muted foreground every pane's empty-state message shares
// with the footer's own resting text (Files Nav, Roots Nav, Document Viewer,
// footer) — one declaration so retuning it is a one-file edit (T25), rather
// than the same lipgloss.Color("245") literal independently repeated in four
// files. A var, not a const: lipgloss.Color returns the color.Color
// interface, which cannot be a constant expression.
var dimColor = lipgloss.Color("245")

// staleRowStyle is S9/C3/M5.1's dedicated staleness row: reverse video
// (Reverse swaps the specified Foreground into the row's background) in the
// same highlight color the Help Modal's border already uses (214,
// modalStyle's own "a color no pane border uses" pick) — reused here rather
// than adding a second, arbitrary "pay attention" color.
var staleRowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Reverse(true)
