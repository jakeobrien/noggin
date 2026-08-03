package main

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

// pickerMode selects which kind of search a Picker runs (S1, GLOSSARY.md
// "File Picker"/"Grep Picker"): fuzzy file-name matching or literal grep.
// The Picker always opens in fileMode (S1's fixed default); only an
// in-Picker switch (M5.1) ever changes it.
type pickerMode int

const (
	fileMode pickerMode = iota
	grepMode
)

// picker is the S1 modal's state: whether it is open, which mode and Scope
// it is searching, and its Prompt (the typed query, S2). It carries no
// results yet — M2.2 onward wire the layout, dispatch and search itself;
// this is the shell M2.1 opens and dismisses.
//
// S8's "no memory" means every field here is discarded on dismiss:
// dismissPicker (below) replaces the whole struct with newPicker()'s zero
// state rather than clearing fields piecemeal, so nothing can survive by
// accident.
type picker struct {
	open  bool
	mode  pickerMode
	scope pickerScope
	input textinput.Model
}

// newPicker is the closed, no-search state: nothing open, nothing typed,
// File Picker mode and Active Root Scope — the zero values of pickerMode
// and pickerScope already are S1's fixed opening state, so there is nothing
// to set beyond a fresh Prompt.
func newPicker() picker {
	return picker{input: textinput.New()}
}

// openPicker opens the Picker (S1): always File Picker mode, Active Root
// Scope, empty Prompt, regardless of what a previous open (in this or an
// earlier session) left behind — S1's fixed opening state and S8's "no
// memory" are the same requirement seen from two sides, so replacing the
// whole state with newPicker() satisfies both at once.
//
// Reachable from every Focus (S1), so unlike startFind/startEdit/activate,
// this never touches m.focus — opening the Picker is not "go to a pane",
// it's a modal layered on top of whichever pane already has focus, the same
// non-pane shape m.helpOpen has (model.go's openHelp).
func (m model) openPicker() (model, tea.Cmd) {
	m.picker = newPicker()
	m.picker.open = true
	cmd := m.picker.input.Focus()
	return m, cmd
}

// dismissPicker closes the Picker, discarding its state entirely (S8): the
// pre-open frame — Active Document, Active Root, scroll positions — was
// never touched by an open Picker in the first place (nothing here reaches
// m.viewer, m.roots or m.files), so restoring it is just this: replace
// picker with its closed zero value.
func (m model) dismissPicker() model {
	m.picker = newPicker()
	return m
}

// handlePickerInput is where a keypress goes while the Picker is open
// (m.picker.open): S2's "always-typing" shape, the same one
// updateFindQuery already gives Find (find.go) — every key but a fixed set
// of non-typing ones falls through to m.picker.input.Update, typing the
// keystroke into the Prompt regardless of what it is bound to everywhere
// else (a query for "jog" must not move a selection or switch mode on its
// second letter).
//
// The non-typing set is exactly S2's list: Cancel dismisses; Activate,
// PickerUp/PickerDown (Results selection), PageUp/PageDown/HalfPageUp/
// HalfPageDown (Preview scrolling, reused as-is from the Document Viewer's
// own scroll bindings — they are already non-printable, so nothing about
// "always typing" requires a second copy) and PickerMode/PickerScope (mode/
// Scope switch) are all recognised and swallowed here. Selection, preview
// and mode/Scope switching have no state to act on yet — M3.1/M3.2 give
// Results/Preview content and M5.1/M5.2 give mode/Scope their actions — so
// today these keys do nothing beyond not typing, which is this task's whole
// job (M2.3): reserving them as non-typing before there is anything to move.
func (m model) handlePickerInput(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Cancel):
		return m.dismissPicker(), nil
	case key.Matches(msg, keys.Activate),
		key.Matches(msg, keys.PickerUp),
		key.Matches(msg, keys.PickerDown),
		key.Matches(msg, keys.PageUp),
		key.Matches(msg, keys.PageDown),
		key.Matches(msg, keys.HalfPageUp),
		key.Matches(msg, keys.HalfPageDown),
		key.Matches(msg, keys.PickerMode),
		key.Matches(msg, keys.PickerScope):
		return m, nil
	}
	next, cmd := m.picker.input.Update(msg)
	m.picker.input = next
	return m, cmd
}
