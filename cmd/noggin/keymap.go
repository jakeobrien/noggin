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
	// missing (broken). It also commits an in-progress find query (M5.1),
	// the viewer's own use of the same physical key: "enter" always means
	// "accept what is selected or typed".
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

	// Find starts an incremental search over the open document (S8, M5.1):
	// global like Edit, and refused the same way with no document open.
	// FindNext and FindPrevious step through the matches it produces,
	// wrapping; Cancel dismisses a search in progress, leaving the view
	// where it is. While the query is being typed, every other key types
	// into it instead of acting on its usual binding — see
	// model.handleFindInput.
	Find         key.Binding
	FindNext     key.Binding
	FindPrevious key.Binding

	// Navigator hides and restores the whole sidebar (Roots Nav and Files
	// Nav), the Document Viewer re-wrapping to the width it gains (S9,
	// M5.2). Global, like Edit and Find, and it never resets either
	// Navigator pane's own selection — hiding and restoring leaves it
	// exactly as it was.
	Navigator key.Binding

	// Cancel dismisses a modal or in-progress input without acting on it.
	// M5.1's first use was dismissing a find query; a generic name and
	// binding on purpose, since S12's help modal (M5.3) uses the same
	// one-key dismissal.
	Cancel key.Binding

	// Help is the footer's help affordance (S11): the footer advertised it
	// under truncation (M4.2) before S12's modal existed to open on it —
	// M5.3 wired the modal onto this binding and its hint text, not a new
	// contract of its own.
	Help key.Binding

	// Picker opens the C2 Picker modal (S1): global, like Edit/Find/
	// Navigator, and reachable from every Focus including while the
	// Navigator is hidden. It always opens the same way — File Picker
	// mode, Active Root Scope, empty Prompt (model.openPicker) — and
	// Cancel is its one-key dismissal, the same generic binding Find and
	// the Help Modal already share.
	Picker key.Binding

	// PickerUp and PickerDown move the open Picker's Results selection
	// (S2, M2.3). Nav Keys' shared LineUp/LineDown (above) deliberately
	// are not reused here: their "k"/"j" alternates are printable, and
	// S2's "every printable keystroke goes to the Prompt" means the
	// Picker cannot treat them as movement the way every other pane
	// does — only the arrow keys survive that rule, so these bind to
	// them alone.
	PickerUp   key.Binding
	PickerDown key.Binding

	// PickerMode and PickerScope switch the open Picker's mode (File ↔
	// Grep, S7) and Scope (Active Root ↔ All Roots, S7) in place, without
	// touching the Prompt (M2.3 reserves them as non-typing; M5.1/M5.2
	// give them their actions). Tab/Shift+Tab rather than a letter, for
	// the same reason PickerUp/PickerDown avoid "j"/"k": both are
	// non-printable, so typing a query is never at odds with switching.
	PickerMode  key.Binding
	PickerScope key.Binding
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
	Find: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "find"),
	),
	FindNext: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "next match"),
	),
	FindPrevious: key.NewBinding(
		key.WithKeys("N"),
		key.WithHelp("N", "prev match"),
	),
	Navigator: key.NewBinding(
		key.WithKeys("ctrl+n"),
		key.WithHelp("ctrl+n", "sidebar"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Picker: key.NewBinding(
		key.WithKeys("ctrl+p"),
		key.WithHelp("ctrl+p", "picker"),
	),
	PickerUp: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "up"),
	),
	PickerDown: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "down"),
	),
	PickerMode: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "mode"),
	),
	PickerScope: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "scope"),
	),
}
