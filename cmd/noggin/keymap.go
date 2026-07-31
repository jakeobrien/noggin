package main

import (
	"strings"
	"unicode/utf8"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
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

// keyPress builds the key press message that a binding's key name corresponds
// to. It lives here, beside the bindings, so that the names of modifiers and
// non-rune keys stay inside this file — callers (tests included) pass a
// binding's key rather than spelling a key literal themselves.
//
// Bubble Tea v2 has no public parser for its own key names (`Key.Keystroke` is
// one-way and ultraviolet's name table is internal), so this is the inverse of
// that formatting, covering the modifiers and named keys `keys` above uses.
// TestKeyPressRoundTrip asserts every key in every binding survives the round
// trip, which is what stops this table drifting from the bindings.
func keyPress(name string) tea.KeyPressMsg {
	var k tea.Key
	parts := strings.Split(name, "+")
	for i, part := range parts {
		// A trailing "+" part is the plus key itself, not a separator.
		if mod, ok := modifiers[part]; ok && i < len(parts)-1 {
			k.Mod |= mod
			continue
		}
		if code, ok := namedKeys[part]; ok {
			k.Code = code
			continue
		}
		// Anything else must be a printable character, which carries Text as
		// well as a code — that is what makes Key.String() report it verbatim.
		code, size := utf8.DecodeRuneInString(part)
		if size != len(part) {
			panic("keymap: no key named " + part + "; add it to namedKeys")
		}
		k.Code, k.Text = code, part
	}
	// A modified printable key has no text: ctrl+c is a keystroke, not input.
	if k.Mod&^tea.ModShift != 0 {
		k.Text = ""
	}
	return tea.KeyPressMsg(k)
}

// modifiers are the modifier prefixes the bindings above use.
var modifiers = map[string]tea.KeyMod{
	"ctrl":  tea.ModCtrl,
	"alt":   tea.ModAlt,
	"shift": tea.ModShift,
}

// namedKeys are the non-printable keys the bindings above use, by the name
// Bubble Tea prints them under. Every key named in `keys` that is not a
// printable character needs an entry here.
var namedKeys = map[string]rune{
	"tab":    tea.KeyTab,
	"up":     tea.KeyUp,
	"down":   tea.KeyDown,
	"pgup":   tea.KeyPgUp,
	"pgdown": tea.KeyPgDown,
	"home":   tea.KeyHome,
	"end":    tea.KeyEnd,
}
