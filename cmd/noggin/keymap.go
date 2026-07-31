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

	// Document Viewer scrolling (S7): line, half-page, page, top, bottom.
	LineUp       key.Binding
	LineDown     key.Binding
	HalfPageUp   key.Binding
	HalfPageDown key.Binding
	PageUp       key.Binding
	PageDown     key.Binding
	Top          key.Binding
	Bottom       key.Binding
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
}

// keyMsg builds the tea.KeyMsg that a binding's key name corresponds to. It
// lives here, beside the bindings, so that the names of non-rune keys stay
// inside this file — callers (tests included) pass a binding's key rather than
// spelling a key literal themselves.
func keyMsg(name string) tea.KeyMsg {
	if t, ok := namedKeys[name]; ok {
		return tea.KeyMsg{Type: t}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(name)}
}

// namedKeys are the non-rune keys the bindings above use, by the name Bubble
// Tea prints them under. Every key named in `keys` that is not a rune needs an
// entry here.
var namedKeys = map[string]tea.KeyType{
	"tab":       tea.KeyTab,
	"shift+tab": tea.KeyShiftTab,
	"ctrl+c":    tea.KeyCtrlC,
	"up":        tea.KeyUp,
	"down":      tea.KeyDown,
	"ctrl+u":    tea.KeyCtrlU,
	"ctrl+d":    tea.KeyCtrlD,
	"ctrl+b":    tea.KeyCtrlB,
	"ctrl+f":    tea.KeyCtrlF,
	"pgup":      tea.KeyPgUp,
	"pgdown":    tea.KeyPgDown,
	"home":      tea.KeyHome,
	"end":       tea.KeyEnd,
}
