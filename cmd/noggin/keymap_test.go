package main

import (
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
)

// keyPress builds the key press message that a binding's key name corresponds
// to. It lives here, in tests only, so that the names of modifiers and
// non-rune keys stay test-only scaffolding — callers pass a binding's key
// rather than spelling a key literal themselves (keymap.go remains the one
// production file that names a key, per D2).
//
// Bubble Tea v2 has no public parser for its own key names (`Key.Keystroke` is
// one-way and ultraviolet's name table is internal), so this is the inverse of
// that formatting, covering the modifiers and named keys `keys` (keymap.go)
// uses. TestKeyPressRoundTrip asserts every key in every binding survives the
// round trip, which is what stops this table drifting from the bindings.
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

// modifiers are the modifier prefixes the bindings in keymap.go use.
var modifiers = map[string]tea.KeyMod{
	"ctrl":  tea.ModCtrl,
	"alt":   tea.ModAlt,
	"shift": tea.ModShift,
}

// namedKeys are the non-printable keys the bindings in keymap.go use, by the
// name Bubble Tea prints them under. Every key named in `keys` that is not a
// printable character needs an entry here.
var namedKeys = map[string]rune{
	"tab":    tea.KeyTab,
	"up":     tea.KeyUp,
	"down":   tea.KeyDown,
	"left":   tea.KeyLeft,
	"right":  tea.KeyRight,
	"pgup":   tea.KeyPgUp,
	"pgdown": tea.KeyPgDown,
	"home":   tea.KeyHome,
	"end":    tea.KeyEnd,
}
