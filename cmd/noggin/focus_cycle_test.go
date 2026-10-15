package main

import (
	"testing"
)

// D141: Tab cycles Focus forward Sources -> Files -> DV1 -> DV2 -> Sources
// and Shift+Tab walks the same ring backward, skipping every pane that is not
// visible. With a single focusable pane it is a no-op, and it never reveals a
// hidden Navigator.

// cycleFixture is a Root-opened model at a normal width, with DV2 added when
// withDV2 is set.
func cycleFixture(t *testing.T, withDV2 bool) model {
	t.Helper()
	m := sized(t, newModel().openFolder(fixtureTree()), 120, 30)
	if withDV2 {
		m = m.withViewerFor(dv2Pane, newViewer())
		m = sized(t, m, 120, 30)
	}
	return m
}

// walk presses key len(want) times from start, checking Focus after each.
func walk(t *testing.T, m model, start pane, keyName string, want []pane) model {
	t.Helper()
	m.focus = start
	for i, w := range want {
		m = press(t, m, keyName)
		if m.focus != w {
			t.Fatalf("after %s #%d from %v: focus = %v, want %v", keyName, i+1, start, m.focus, w)
		}
	}
	return m
}

func TestTabCyclesEveryVisiblePaneForward(t *testing.T) {
	m := cycleFixture(t, true)
	walk(t, m, rootsPane, keys.FocusNext.Keys()[0], []pane{filesPane, dv1Pane, dv2Pane, rootsPane})
}

func TestShiftTabCyclesBackward(t *testing.T) {
	m := cycleFixture(t, true)
	walk(t, m, rootsPane, keys.FocusPrevious.Keys()[0], []pane{dv2Pane, dv1Pane, filesPane, rootsPane})
}

func TestTabSkipsMissingDV2(t *testing.T) {
	m := cycleFixture(t, false)
	walk(t, m, dv1Pane, keys.FocusNext.Keys()[0], []pane{rootsPane, filesPane, dv1Pane})
}

func TestTabSkipsHiddenNavigatorAndNeverRevealsIt(t *testing.T) {
	m := cycleFixture(t, true)
	m = press(t, m, keys.Navigator.Keys()[0])
	if !m.navigatorHidden {
		t.Fatal("setup: Navigator toggle did not hide the sidebar")
	}
	m = walk(t, m, dv1Pane, keys.FocusNext.Keys()[0], []pane{dv2Pane, dv1Pane})
	if !m.navigatorHidden {
		t.Error("Tab revealed the hidden Navigator")
	}
}

func TestTabIsNoOpWithOnlyDV1Focusable(t *testing.T) {
	m := cycleFixture(t, false)
	m = press(t, m, keys.Navigator.Keys()[0])
	walk(t, m, dv1Pane, keys.FocusNext.Keys()[0], []pane{dv1Pane})
	walk(t, m, dv1Pane, keys.FocusPrevious.Keys()[0], []pane{dv1Pane})
}

func TestTabSkipsNavigatorOnATooNarrowTerminal(t *testing.T) {
	narrow := minNavigatorWidth + minViewerWidth - 1
	m := sized(t, newModel().openFolder(fixtureTree()), narrow, 24)
	m = m.withViewerFor(dv2Pane, newViewer())
	m = sized(t, m, narrow, 24)
	if m.visible(rootsPane) || m.visible(filesPane) {
		t.Fatal("setup: Navigator still visible on a too-narrow terminal")
	}
	walk(t, m, dv1Pane, keys.FocusNext.Keys()[0], []pane{dv2Pane, dv1Pane})
}

// TestTabIsInertWhileThePickerIsOpen: the Picker owns the keyboard, and Tab
// is not one of its keys, so it neither moves Focus nor types into the Prompt.
func TestTabIsInertWhileThePickerIsOpen(t *testing.T) {
	m := cycleFixture(t, false)
	m.focus = filesPane
	m = press(t, m, keys.Picker.Keys()[0])
	if !m.picker.open {
		t.Fatal("setup: Picker did not open")
	}
	m = press(t, m, keys.FocusNext.Keys()[0])
	if m.focus != filesPane {
		t.Errorf("focus = %v after Tab in the Picker, want unchanged filesPane", m.focus)
	}
	if got := m.picker.input.Value(); got != "" {
		t.Errorf("Prompt = %q after Tab, want empty", got)
	}
}
