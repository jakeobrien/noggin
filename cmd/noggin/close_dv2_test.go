package main

import "testing"

// TestCloseDV2FromEveryPane is S4/C3/M2.1's full sweep: `ctrl+x` closes DV2
// from every non-modal Focus (Roots Nav, Files Nav, DV1, DV2 itself), and
// Focus lands per S4's own rule — DV1 if DV2 held Focus, unchanged
// otherwise.
func TestCloseDV2FromEveryPane(t *testing.T) {
	for _, start := range []pane{rootsPane, filesPane, dv1Pane, dv2Pane} {
		t.Run(start.title(), func(t *testing.T) {
			m := sized(t, newModel().openFolder(fixtureTree()), 80, 24)
			m = m.withViewerFor(dv2Pane, newViewer())
			m = sized(t, m, 80, 24) // re-lay-out now that dv2 exists
			m.focus = start

			got := press(t, m, keys.CloseDV2.Keys()[0])

			if got.dv2 != nil {
				t.Error("dv2 still exists after CloseDV2")
			}
			want := start
			if start == dv2Pane {
				want = dv1Pane
			}
			if got.focus != want {
				t.Errorf("focus after CloseDV2 from %v = %v, want %v", start, got.focus, want)
			}
		})
	}
}

// TestCloseDV2NoOpWithoutDV2 pins S4's carve-out: `ctrl+x` does nothing at
// all — no dv2 created, Focus unchanged — when DV2 does not already exist,
// the same shape FocusDV2 already has (focus_test.go's
// TestFocusDV2InertWithoutDV2).
func TestCloseDV2NoOpWithoutDV2(t *testing.T) {
	for _, start := range []pane{rootsPane, filesPane, dv1Pane} {
		t.Run(start.title(), func(t *testing.T) {
			m := sized(t, newModel(), 80, 24)
			m.focus = start

			got := press(t, m, keys.CloseDV2.Keys()[0])

			if got.dv2 != nil {
				t.Error("CloseDV2 created dv2 as a side effect; it must be a no-op with DV2 absent")
			}
			if got.focus != start {
				t.Errorf("focus after CloseDV2 with no dv2, from %v = %v, want unchanged %v", start, got.focus, start)
			}
		})
	}
}

// TestCloseDV2InertWhileModalOwnsKeyboard checks S4's other carve-out:
// `ctrl+x` does nothing while the Picker or the Help Modal owns the
// keyboard, the same exclusivity every other global binding already has
// there (model.Update's three gates run before the switch CloseDV2 is
// matched in).
func TestCloseDV2InertWhileModalOwnsKeyboard(t *testing.T) {
	t.Run("picker", func(t *testing.T) {
		m := sized(t, newModel(), 80, 24)
		m = m.withViewerFor(dv2Pane, newViewer())
		m = sized(t, m, 80, 24)
		m = press(t, m, keys.Picker.Keys()[0])
		if !m.picker.open {
			t.Fatal("test setup: picker did not open")
		}

		got := press(t, m, keys.CloseDV2.Keys()[0])

		if got.dv2 == nil {
			t.Error("CloseDV2 closed dv2 while the picker owned the keyboard")
		}
		if !got.picker.open {
			t.Error("CloseDV2 closed the picker")
		}
	})

	t.Run("help modal", func(t *testing.T) {
		m := sized(t, newModel(), 80, 24)
		m = m.withViewerFor(dv2Pane, newViewer())
		m = sized(t, m, 80, 24)
		m = press(t, m, keys.Help.Keys()[0])
		if !m.helpOpen {
			t.Fatal("test setup: help modal did not open")
		}

		got := press(t, m, keys.CloseDV2.Keys()[0])

		if got.dv2 == nil {
			t.Error("CloseDV2 closed dv2 while the help modal owned the keyboard")
		}
		if !got.helpOpen {
			t.Error("CloseDV2 closed the help modal")
		}
	})
}
