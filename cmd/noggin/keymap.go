package main

import (
	"charm.land/bubbles/v2/key"
)

// This is the one production file key literals may appear in (D2). Every
// binding in the program is resolved through `keys`; no other production file
// names a key — keymap_test.go carries its own test-only key-encoding table
// (see keyPress), exempt for the same reason its helpers are test-only.
// Retuning the keymap is therefore a one-file edit, and `rg` over the rest of
// the production tree stays free of key strings.

// keyMap holds every binding the program understands. Help text lives with the
// binding so the footer (S11) and help modal (S12) can be generated from it
// rather than duplicating the list.
type keyMap struct {
	NextPane key.Binding
	PrevPane key.Binding
	Quit     key.Binding

	// Roots Nav actions (S3): selection moves with the shared LineUp/LineDown
	// bindings; Star toggles the selected root's starred mark, persisted via
	// the S2 store. Activate (below) also opens the selected root.
	Star key.Binding

	// Files Nav folder control (S5): open and close a folder in place.
	// Selection moves with the shared LineUp/LineDown bindings.
	Expand   key.Binding
	Collapse key.Binding

	// Activate opens the selected entry: on the Files Nav, the selected file
	// becomes the active document and focus moves to the viewer (M2.3); a
	// folder is left for later. On the Roots Nav, the selected root becomes
	// the active root (S3), refused with a footer message if its path is
	// missing (broken).
	Activate key.Binding

	// Document Viewer scrolling (S7): line, half-page, page, top, bottom.
	LineUp       key.Binding
	LineDown     key.Binding
	HalfPageUp   key.Binding
	HalfPageDown key.Binding
	PageUp       key.Binding
	PageDown     key.Binding
	Top          key.Binding
	Bottom       key.Binding

	// Edit hands the active document to $EDITOR (S10, D3): noggin releases the
	// terminal, resumes on exit and re-renders it from disk. Global, not
	// pane-specific — with no document open it reports that instead of doing
	// nothing (M4.1).
	Edit key.Binding

	// Help is the footer's help affordance (S11): the footer advertises it
	// under truncation (M4.2) even before S12's modal exists to open on it —
	// the binding and its hint text are what M5.3 wires a modal onto, not a
	// new contract this milestone invents.
	Help key.Binding
}

// keys is the program's keymap.
var keys = keyMap{
	NextPane: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next pane"),
	),
	PrevPane: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev pane"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Star: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "star"),
	),
	Expand: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "expand"),
	),
	Collapse: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "collapse"),
	),
	Activate: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "open"),
	),
	LineUp: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	LineDown: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	HalfPageUp: key.NewBinding(
		key.WithKeys("ctrl+u"),
		key.WithHelp("ctrl+u", "½ page up"),
	),
	HalfPageDown: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "½ page down"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("pgup", "ctrl+b"),
		key.WithHelp("pgup", "page up"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("pgdown", "ctrl+f"),
		key.WithHelp("pgdn", "page down"),
	),
	Top: key.NewBinding(
		key.WithKeys("home", "g"),
		key.WithHelp("g", "top"),
	),
	Bottom: key.NewBinding(
		key.WithKeys("end", "G"),
		key.WithHelp("G", "bottom"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
}
