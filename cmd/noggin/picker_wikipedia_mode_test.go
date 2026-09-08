package main

import "testing"

// This file is C6/M3.2's own coverage: Wikipedia joining the Picker's mode
// cycle, the landing mode an open picks depending on the Active Source, and
// Wikipedia mode's own lack of a Scope (S2). Wikipedia mode's actual search
// (results, Preview placeholder) is M3.3's, not covered here.

// pickerWithRootActive opens the Picker with dir as the sole, active Root —
// the "a Root is the Active Source" half of S2's "reachable from any Active
// Source" and its unrevised "lands on File mode" default.
func pickerWithRootActive(t *testing.T) model {
	t.Helper()
	dir := mkdir(t, t.TempDir()+"/root")
	m := sized(t, newModel(), 80, 24)
	m.roots = m.roots.setRoots([]root{{path: dir}}, 0)
	m, _ = m.openPicker()
	return m
}

// pickerWithWikipediaActive opens the Picker with the Wikipedia Query
// Source, rather than any Root, as the Active Source.
func pickerWithWikipediaActive(t *testing.T) model {
	t.Helper()
	m := sized(t, newModel(), 80, 24)
	m = m.activateWikipediaSource()
	m, _ = m.openPicker()
	return m
}

// TestOpenPickerLandsOnFileModeWithARootActive is S2's unrevised half of its
// own landing rule: opening the Picker with a Root as the Active Source
// still lands on File mode, exactly as before Wikipedia mode existed.
func TestOpenPickerLandsOnFileModeWithARootActive(t *testing.T) {
	m := pickerWithRootActive(t)
	if m.picker.mode != fileMode {
		t.Fatalf("mode on open with a Root active = %v, want fileMode", m.picker.mode)
	}
}

// TestOpenPickerLandsOnWikipediaModeWithWikipediaActive is S2's own new
// landing rule, verbatim: "Opening the Picker while Wikipedia is the Active
// Source lands on Wikipedia mode." Scope is left at its zero value —
// Wikipedia mode ignores it (switchPickerScope no-ops, pickerModeScope
// always reads "-") so there is nothing meaningful to assert about it here
// beyond "unchanged from open".
func TestOpenPickerLandsOnWikipediaModeWithWikipediaActive(t *testing.T) {
	m := pickerWithWikipediaActive(t)
	if m.picker.mode != wikipediaMode {
		t.Fatalf("mode on open with Wikipedia active = %v, want wikipediaMode", m.picker.mode)
	}
	if m.picker.scope != activeRootScope {
		t.Errorf("scope on open = %v, want the zero value (Wikipedia mode ignores Scope)", m.picker.scope)
	}
}

// pickerModeCycle drives n presses of the PickerMode key and returns the
// mode after each one, in order — shared by both Active-Source cases below
// so the fixed File → Grep → Wikipedia → File order (S2, C6/M3.2) is
// asserted identically regardless of which kind of Source, or which mode,
// started it.
func pickerModeCycle(t *testing.T, m model, n int) []pickerMode {
	t.Helper()
	modes := make([]pickerMode, n)
	for i := range modes {
		next, _ := m.Update(keyPress(keys.PickerMode.Keys()[0]))
		m = concrete(t, next)
		modes[i] = m.picker.mode
	}
	return modes
}

// TestPickerModeCycleFromARootActiveSource is S2/C6/M3.2's three-way cycle,
// verbatim ("File → Grep → Wikipedia"), driven from a Root Active Source
// (which lands the Picker on File mode) through two full laps so the
// wraparound back to File is covered too, not just the first pass.
func TestPickerModeCycleFromARootActiveSource(t *testing.T) {
	m := pickerWithRootActive(t)
	if m.picker.mode != fileMode {
		t.Fatalf("setup: mode = %v, want fileMode", m.picker.mode)
	}
	got := pickerModeCycle(t, m, 6)
	want := []pickerMode{grepMode, wikipediaMode, fileMode, grepMode, wikipediaMode, fileMode}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("mode after press %d = %v, want %v", i+1, got[i], w)
		}
	}
}

// TestPickerModeCycleFromWikipediaActiveSource is the same cycle's other
// half: S2's "reachable from any Active Source" means the mode key is not
// restricted while Wikipedia is the Active Source either, and the cycle
// order is identical no matter which mode it started on.
func TestPickerModeCycleFromWikipediaActiveSource(t *testing.T) {
	m := pickerWithWikipediaActive(t)
	if m.picker.mode != wikipediaMode {
		t.Fatalf("setup: mode = %v, want wikipediaMode", m.picker.mode)
	}
	got := pickerModeCycle(t, m, 4)
	want := []pickerMode{fileMode, grepMode, wikipediaMode, fileMode}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("mode after press %d = %v, want %v", i+1, got[i], w)
		}
	}
}

// TestPickerScopeKeyIsANoOpInWikipediaMode is S2's own wording, verbatim:
// "Wikipedia mode has no Scope (the scope key is a no-op and reads -)".
// Checked against the model here — the rendered "-" label is
// picker_view_test.go's TestPickerStatusStates-style coverage — pressing
// PickerScope in Wikipedia mode must change nothing and dispatch no search,
// a true no-op rather than a cycle that happens to have one value.
func TestPickerScopeKeyIsANoOpInWikipediaMode(t *testing.T) {
	m := pickerWithWikipediaActive(t)
	before := m.picker.scope
	beforeSeq := m.picker.seq

	next, cmd := m.Update(keyPress(keys.PickerScope.Keys()[0]))
	m = concrete(t, next)

	if m.picker.scope != before {
		t.Errorf("scope after PickerScope in Wikipedia mode = %v, want unchanged %v", m.picker.scope, before)
	}
	if m.picker.seq != beforeSeq {
		t.Errorf("seq after PickerScope in Wikipedia mode = %d, want unchanged %d (no search generation opened)", m.picker.seq, beforeSeq)
	}
	if cmd != nil {
		t.Error("PickerScope in Wikipedia mode scheduled a Cmd, want none (a true no-op)")
	}
}
