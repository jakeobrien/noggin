package main

import "testing"

// TestPickerOpensFromEveryFocusAndNavigatorHidden is M2.1's "reachable from
// every Focus including while the Navigator is hidden" (S1): pressing the
// Picker binding opens it regardless of which pane currently has focus, and
// still opens it when the Navigator has been toggled away (so no Navigator
// pane could possibly have focus).
func TestPickerOpensFromEveryFocusAndNavigatorHidden(t *testing.T) {
	for _, p := range paneOrder {
		m := openedFixture(t, 80, 24)
		m.focus = p

		m = press(t, m, keys.Picker.Keys()[0])

		if !m.picker.open {
			t.Errorf("focus %v: picker.open = false after the Picker key, want true", p)
		}
	}

	m := openedFixture(t, 80, 24)
	m = press(t, m, keys.Navigator.Keys()[0])
	if !m.navigatorHidden {
		t.Fatal("Navigator toggle did not hide it; setup is broken")
	}

	m = press(t, m, keys.Picker.Keys()[0])
	if !m.picker.open {
		t.Error("navigator hidden: picker.open = false after the Picker key, want true")
	}
}

// TestPickerAlwaysOpensAsFileActiveRootEmptyPrompt is S1's fixed opening
// state, checked from a picker that (if memory existed) would say
// otherwise: Grep mode, All Roots Scope, a typed Prompt. Opening always
// resets to File Picker / Active Root / empty Prompt regardless.
func TestPickerAlwaysOpensAsFileActiveRootEmptyPrompt(t *testing.T) {
	m := openedFixture(t, 80, 24)
	m.picker = picker{open: false, mode: grepMode, scope: allRootsScope, input: newPicker().input}
	m.picker.input.SetValue("stale query")

	next, _ := m.openPicker()

	if next.picker.mode != fileMode {
		t.Errorf("mode = %v, want fileMode", next.picker.mode)
	}
	if next.picker.scope != activeRootScope {
		t.Errorf("scope = %v, want activeRootScope", next.picker.scope)
	}
	if next.picker.input.Value() != "" {
		t.Errorf("prompt = %q, want empty", next.picker.input.Value())
	}
	if !next.picker.open {
		t.Error("open = false, want true")
	}
}

// TestPickerDismissRestoresThePreOpenFrameByteForByte is M2.1's own
// acceptance criterion: opening the Picker, changing what it would show
// (mode, Scope, Prompt) and dismissing it leaves the session exactly as it
// was found — the same Active Document, Active Root and scroll position,
// and (since M2.1 predates any Picker rendering, M2.2 onward's job) the
// exact same rendered frame.
func TestPickerDismissRestoresThePreOpenFrameByteForByte(t *testing.T) {
	m := openedFixture(t, 80, 24)
	m = press(t, m, "j") // move the viewer's scroll position off zero
	beforeFrame := m.View().Content
	beforeDoc := m.viewer.doc.path
	beforeRoot := m.roots.active
	beforeOffset := m.viewer.offset()

	m, _ = m.openPicker()
	if !m.picker.open {
		t.Fatal("openPicker did not open the picker; setup is broken")
	}
	// Perturb the Picker's own state (mode/Scope/Prompt) so dismissal is
	// proven to discard it, not merely leave it alone because nothing
	// changed.
	m.picker.mode = grepMode
	m.picker.scope = allRootsScope
	m.picker.input.SetValue("something typed")

	m = press(t, m, keys.Cancel.Keys()[0])

	if m.picker.open {
		t.Fatal("picker.open = true after Cancel, want false")
	}
	if got := m.View().Content; got != beforeFrame {
		t.Errorf("frame after dismissal does not match the pre-open frame")
	}
	if m.viewer.doc.path != beforeDoc {
		t.Errorf("active document = %q, want %q", m.viewer.doc.path, beforeDoc)
	}
	if m.roots.active != beforeRoot {
		t.Errorf("active root = %d, want %d", m.roots.active, beforeRoot)
	}
	if m.viewer.offset() != beforeOffset {
		t.Errorf("scroll offset = %d, want %d", m.viewer.offset(), beforeOffset)
	}
	// A reopen must not resurrect the discarded mode/Scope/Prompt (S8).
	m, _ = m.openPicker()
	if m.picker.mode != fileMode || m.picker.scope != activeRootScope || m.picker.input.Value() != "" {
		t.Errorf("reopened picker = %+v, want File/ActiveRoot/empty", m.picker)
	}
}

// TestPickerOwnsKeyboardWhileOpen is the same exclusivity handleFindInput
// and handleHelpInput already give their own modals (model.go): while the
// Picker is open, a key bound elsewhere (here, Quit's "q") must not act on
// its usual binding.
func TestPickerOwnsKeyboardWhileOpen(t *testing.T) {
	m := openedFixture(t, 80, 24)
	m, _ = m.openPicker()

	next, _ := m.Update(keyPress("q"))
	got := concrete(t, next)
	if !got.picker.open {
		t.Error("picker.open = false after \"q\" while open, want true (q must not quit)")
	}
}
