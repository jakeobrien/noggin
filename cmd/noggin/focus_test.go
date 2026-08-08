package main

import "testing"

// TestFocusJumpKeysMoveFocusFromEveryPane is S5's full sweep of all four
// direct focus-jump keys (FocusRoots/FocusFiles, carried over from C2a/M2.1,
// plus FocusDV1/FocusDV2 this task, C3/M1.3, adds): each lands Focus on its
// own pane from every one of the other three, DV2 included once it exists.
func TestFocusJumpKeysMoveFocusFromEveryPane(t *testing.T) {
	base := sized(t, newModel().openFolder(fixtureTree()), 80, 24)
	base = base.withViewerFor(dv2Pane, newViewer())
	base = sized(t, base, 80, 24) // re-lay-out now that dv2 exists

	starts := []struct {
		name string
		pane pane
	}{
		{"rootsPane", rootsPane},
		{"filesPane", filesPane},
		{"dv1Pane", dv1Pane},
		{"dv2Pane", dv2Pane},
	}
	jumps := []struct {
		name string
		key  string
		want pane
	}{
		{"FocusRoots", keys.FocusRoots.Keys()[0], rootsPane},
		{"FocusFiles", keys.FocusFiles.Keys()[0], filesPane},
		{"FocusDV1", keys.FocusDV1.Keys()[0], dv1Pane},
		{"FocusDV2", keys.FocusDV2.Keys()[0], dv2Pane},
	}

	for _, start := range starts {
		t.Run(start.name, func(t *testing.T) {
			for _, j := range jumps {
				m := base
				m.focus = start.pane
				got := press(t, m, j.key)
				if got.focus != j.want {
					t.Errorf("%s from %v = %v, want %v", j.name, start.pane, got.focus, j.want)
				}
			}
		})
	}
}

// TestFocusDV2InertWithoutDV2 pins S5's carve-out: FocusDV2 (ctrl+o) does
// nothing at all — Focus stays exactly where it was — when DV2 does not
// exist, rather than falling back through focusOrder the way an ordinary
// hidden-by-width target would (resolveFocus). "A bare focus request has
// no document to create" (S5), so it is not focusPane's job to land
// somewhere else instead.
func TestFocusDV2InertWithoutDV2(t *testing.T) {
	for _, start := range []pane{rootsPane, filesPane, dv1Pane} {
		t.Run(start.title(), func(t *testing.T) {
			m := sized(t, newModel(), 80, 24)
			m.focus = start

			got := press(t, m, keys.FocusDV2.Keys()[0])
			if got.focus != start {
				t.Errorf("focus after FocusDV2 with no dv2, from %v = %v, want unchanged %v", start, got.focus, start)
			}
			if got.dv2 != nil {
				t.Error("FocusDV2 created dv2 as a side effect; it must only focus, never create (S3 owns creation)")
			}
		})
	}
}

// TestFocusDV1AndDV2DoNotRevealHiddenNavigator: unlike FocusRoots/FocusFiles
// (whose target lives inside the Navigator column, so hiding it makes the
// jump target itself invisible), FocusDV1/FocusDV2's target is the Document
// Viewer, which is never part of that column — jumping to it must not have
// the side effect of revealing a hidden sidebar the user did not ask for
// (model.focusPane's own doc comment).
func TestFocusDV1AndDV2DoNotRevealHiddenNavigator(t *testing.T) {
	for _, tc := range []struct {
		name string
		key  string
		want pane
		dv2  bool
	}{
		{"FocusDV1", keys.FocusDV1.Keys()[0], dv1Pane, false},
		{"FocusDV2", keys.FocusDV2.Keys()[0], dv2Pane, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := sized(t, newModel(), 80, 24)
			if tc.dv2 {
				m = m.withViewerFor(dv2Pane, newViewer())
				m = sized(t, m, 80, 24)
			}
			m = press(t, m, keys.Navigator.Keys()[0]) // hide
			if !m.navigatorHidden {
				t.Fatal("test setup: navigator did not hide")
			}

			m = press(t, m, tc.key)
			if !m.navigatorHidden {
				t.Error("navigator revealed by a viewer focus-jump, want it to stay hidden")
			}
			if m.focus != tc.want {
				t.Errorf("focus = %v, want %v", m.focus, tc.want)
			}
		})
	}
}

// TestFocusBindingsRevealAHiddenNavigator checks the other half of M2.1's
// acceptance criterion: pressing either binding while the S9 toggle has the
// Navigator hidden reveals it, the same way TestNavigatorToggleHidesAndRestoresTheSidebar
// (navigator_test.go) checks the toggle itself.
func TestFocusBindingsRevealAHiddenNavigator(t *testing.T) {
	for _, tc := range []struct {
		name string
		key  string
		want pane
	}{
		{"FocusRoots", keys.FocusRoots.Keys()[0], rootsPane},
		{"FocusFiles", keys.FocusFiles.Keys()[0], filesPane},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := sized(t, newModel(), 80, 24)
			m = press(t, m, keys.Navigator.Keys()[0]) // hide
			if !m.navigatorHidden {
				t.Fatal("test setup: navigator did not hide")
			}

			m = press(t, m, tc.key)
			if m.navigatorHidden {
				t.Error("navigator still hidden after a focus binding")
			}
			if m.layout().navigatorWidth() == 0 {
				t.Error("navigator width is 0 after a focus binding should have revealed it")
			}
			if m.focus != tc.want {
				t.Errorf("focus = %v, want %v", m.focus, tc.want)
			}
		})
	}
}

// TestFocusBindingsNeverHideTheNavigator asserts the "never hiding it" half
// of M2.1: pressing FocusRoots/FocusFiles while the Navigator is already
// visible must not toggle it off, unlike keys.Navigator itself.
func TestFocusBindingsNeverHideTheNavigator(t *testing.T) {
	m := sized(t, newModel(), 80, 24)
	if m.navigatorHidden {
		t.Fatal("test setup: navigator should start visible")
	}

	m = press(t, m, keys.FocusRoots.Keys()[0])
	m = press(t, m, keys.FocusFiles.Keys()[0])
	m = press(t, m, keys.FocusRoots.Keys()[0])

	if m.navigatorHidden {
		t.Error("navigator hidden after focus bindings alone; they must never hide it")
	}
}

// TestFocusBindingsInertWhilePickerIsOpen checks M2.1's Picker-gating clause:
// FocusRoots/FocusFiles do nothing while the Picker owns the keyboard, the
// same exclusivity Edit/Find/Navigator already have there.
func TestFocusBindingsInertWhilePickerIsOpen(t *testing.T) {
	m := sized(t, newModel(), 80, 24)
	m = press(t, m, keys.Picker.Keys()[0])
	if !m.picker.open {
		t.Fatal("test setup: picker did not open")
	}
	focusBefore := m.focus

	m = press(t, m, keys.FocusRoots.Keys()[0])
	if !m.picker.open {
		t.Error("FocusRoots closed the picker")
	}
	if m.focus != focusBefore {
		t.Errorf("focus changed to %v while picker open, want unchanged %v", m.focus, focusBefore)
	}

	m = press(t, m, keys.FocusFiles.Keys()[0])
	if !m.picker.open {
		t.Error("FocusFiles closed the picker")
	}
	if m.focus != focusBefore {
		t.Errorf("focus changed to %v while picker open, want unchanged %v", m.focus, focusBefore)
	}
}

// TestFocusDV1AndDV2InertWhilePickerIsOpen extends the Picker-gating clause
// above to C3/M1.3's own two focus-jump keys, DV2 pre-created so FocusDV2
// has something to (not) jump to.
func TestFocusDV1AndDV2InertWhilePickerIsOpen(t *testing.T) {
	m := sized(t, newModel(), 80, 24)
	m = m.withViewerFor(dv2Pane, newViewer())
	m = sized(t, m, 80, 24)
	m = press(t, m, keys.Picker.Keys()[0])
	if !m.picker.open {
		t.Fatal("test setup: picker did not open")
	}
	focusBefore := m.focus

	m = press(t, m, keys.FocusDV1.Keys()[0])
	if !m.picker.open {
		t.Error("FocusDV1 closed the picker")
	}
	if m.focus != focusBefore {
		t.Errorf("focus changed to %v while picker open, want unchanged %v", m.focus, focusBefore)
	}

	m = press(t, m, keys.FocusDV2.Keys()[0])
	if !m.picker.open {
		t.Error("FocusDV2 closed the picker")
	}
	if m.focus != focusBefore {
		t.Errorf("focus changed to %v while picker open, want unchanged %v", m.focus, focusBefore)
	}
}
