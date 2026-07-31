package main

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// This file is the single place key literals may appear (D2: the keymap is
// provisional and build-time-only). Every binding in the program is resolved
// through `keys`; no other file names a key. Retuning the keymap is therefore a
// one-file edit, and `rg` over the rest of the tree stays free of key strings.

// keyMap holds every binding the program understands. Help text lives with the
// binding so the footer (S11) and help modal (S12) can be generated from it
// rather than duplicating the list.
type keyMap struct {
	NextPane key.Binding
	PrevPane key.Binding
	Quit     key.Binding
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
}

// keyMsg builds the tea.KeyMsg that a binding's key name corresponds to. It
// lives here, beside the bindings, so that the names of non-rune keys stay
// inside this file — callers (tests included) pass a binding's key rather than
// spelling a key literal themselves.
func keyMsg(name string) tea.KeyMsg {
	switch name {
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "shift+tab":
		return tea.KeyMsg{Type: tea.KeyShiftTab}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(name)}
}
