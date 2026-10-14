package main

import "testing"

func restoreKey() string { return keys.PickerRestore.Keys()[0] }

// dismissWith opens the Picker, sets its state, and dismisses it with esc.
func dismissWith(t *testing.T, m model, mode pickerMode, scope pickerScope, query string) model {
	t.Helper()
	m, _ = m.openPicker()
	m.picker.mode, m.picker.scope = mode, scope
	m.picker.input.SetValue(query)
	return press(t, m, keys.Cancel.Keys()[0])
}

func TestPickerRestoreReopensToLastQueryModeAndScope(t *testing.T) {
	m := openedFixture(t, 80, 24)
	m = dismissWith(t, m, grepMode, allRootsScope, "needle")

	m = press(t, m, restoreKey())

	if !m.picker.open {
		t.Fatal("picker not open")
	}
	if got := m.picker.input.Value(); got != "needle" {
		t.Errorf("query = %q, want needle", got)
	}
	if m.picker.mode != grepMode || m.picker.scope != allRootsScope {
		t.Errorf("mode/scope = %v/%v, want grepMode/allRootsScope", m.picker.mode, m.picker.scope)
	}
	if m.picker.input.Position() != len("needle") {
		t.Errorf("cursor = %d, want at end", m.picker.input.Position())
	}
}

func TestPickerRestoreSearchesImmediately(t *testing.T) {
	m := openedFixture(t, 80, 24)
	m = dismissWith(t, m, fileMode, activeRootScope, "needle")
	seq := m.picker.seq

	m, cmd := m.restorePicker()
	if cmd == nil || m.picker.seq <= seq {
		t.Error("restore did not queue a search")
	}
}

func TestPickerRestoreWithNothingSavedOpensNormally(t *testing.T) {
	m := openedFixture(t, 80, 24)
	m = press(t, m, restoreKey())
	if !m.picker.open || m.picker.input.Value() != "" || m.picker.mode != fileMode || m.picker.scope != activeRootScope {
		t.Errorf("want a normal fresh open, got open=%v query=%q mode=%v scope=%v",
			m.picker.open, m.picker.input.Value(), m.picker.mode, m.picker.scope)
	}
}

func TestEmptyDismissDoesNotOverwriteLastPicker(t *testing.T) {
	m := openedFixture(t, 80, 24)
	m = dismissWith(t, m, grepMode, allRootsScope, "needle")
	m = dismissWith(t, m, fileMode, activeRootScope, "")

	m = press(t, m, restoreKey())
	if got := m.picker.input.Value(); got != "needle" || m.picker.mode != grepMode {
		t.Errorf("restored %q/%v, want the earlier non-empty session", got, m.picker.mode)
	}
}

func TestNonEmptyDismissOverwritesLastPicker(t *testing.T) {
	m := openedFixture(t, 80, 24)
	m = dismissWith(t, m, grepMode, allRootsScope, "old")
	m = dismissWith(t, m, fileMode, activeRootScope, "x") // sub-threshold still counts

	m = press(t, m, restoreKey())
	if got := m.picker.input.Value(); got != "x" || m.picker.mode != fileMode {
		t.Errorf("restored %q/%v, want x/fileMode", got, m.picker.mode)
	}
}

func TestPlainPickerKeyStillOpensFreshAndKeepsLast(t *testing.T) {
	m := openedFixture(t, 80, 24)
	m = dismissWith(t, m, grepMode, allRootsScope, "needle")

	m = press(t, m, keys.Picker.Keys()[0])
	if m.picker.input.Value() != "" || m.picker.mode != fileMode {
		t.Errorf("plain open not fresh: %q/%v", m.picker.input.Value(), m.picker.mode)
	}
	m = press(t, m, keys.Cancel.Keys()[0]) // empty dismissal
	m = press(t, m, restoreKey())
	if m.picker.input.Value() != "needle" {
		t.Errorf("last query lost: %q", m.picker.input.Value())
	}
}

func TestPickerRestoreKeyWhileOpenIsNoOp(t *testing.T) {
	m := openedFixture(t, 80, 24)
	m, _ = m.openPicker()
	m.picker.input.SetValue("abc")
	for _, k := range keys.PickerRestore.Keys() {
		m = press(t, m, k)
	}
	if got := m.picker.input.Value(); got != "abc" || !m.picker.open {
		t.Errorf("query=%q open=%v, want unchanged", got, m.picker.open)
	}
}

func TestActivateSavesLastPicker(t *testing.T) {
	m := openedFixture(t, 80, 24)
	m, _ = m.openPicker()
	m.picker.input.SetValue("needle")
	m = m.dismissPicker() // every exit path (esc, activate) goes through this
	if m.lastPicker == nil || m.lastPicker.query != "needle" {
		t.Errorf("lastPicker = %+v", m.lastPicker)
	}
}
