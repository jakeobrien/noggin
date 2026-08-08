package main

import "testing"

// TestNavigatorToggleHidesAndRestoresTheSidebar exercises S9 end to end: one
// key press hides the whole sidebar (both Navigator panes lose their
// region), a second restores it to exactly what it was before.
func TestNavigatorToggleHidesAndRestoresTheSidebar(t *testing.T) {
	m := sized(t, newModel(), 80, 24)
	before := m.layout()
	if before.navigatorWidth() == 0 {
		t.Fatal("test setup: navigator should be visible at 80x24")
	}

	m = press(t, m, keys.Navigator.Keys()[0])
	if !m.navigatorHidden {
		t.Fatal("Navigator toggle did not set navigatorHidden")
	}
	hidden := m.layout()
	if hidden.navigatorWidth() != 0 {
		t.Errorf("navigator width after hiding = %d, want 0", hidden.navigatorWidth())
	}
	if hidden.viewer.width != 80 {
		t.Errorf("viewer width after hiding = %d, want %d (full terminal width)", hidden.viewer.width, 80)
	}
	wantWidth := hidden.viewer.content().width
	if m.dv1.renderedWidth != wantWidth {
		t.Errorf("viewer did not re-wrap to the width it gained: rendered at %d, want %d",
			m.dv1.renderedWidth, wantWidth)
	}

	m = press(t, m, keys.Navigator.Keys()[0])
	if m.navigatorHidden {
		t.Fatal("second Navigator toggle did not clear navigatorHidden")
	}
	restored := m.layout()
	if restored != before {
		t.Errorf("layout after restoring = %+v, want %+v", restored, before)
	}
}

// TestNavigatorToggleRewrapsTheOpenDocument is the acceptance criterion's own
// wording: hiding the sidebar must actually re-wrap the document text to the
// width it gained, not just resize an empty viewport (the test above uses no
// open document, so it only proves the width arithmetic). Gaining width means
// fewer wrapped lines, the same signal TestViewerRewrapsOnResize uses for an
// ordinary resize.
func TestNavigatorToggleRewrapsTheOpenDocument(t *testing.T) {
	m := openedFixture(t, 80, 24)
	withSidebar := m.dv1.lineCount()

	m = press(t, m, keys.Navigator.Keys()[0])
	fullWidth := m.dv1.lineCount()
	if fullWidth >= withSidebar {
		t.Errorf("line count with the sidebar hidden = %d, want fewer than %d (wider viewport wraps less)",
			fullWidth, withSidebar)
	}

	m = press(t, m, keys.Navigator.Keys()[0])
	if got := m.dv1.lineCount(); got != withSidebar {
		t.Errorf("line count after restoring the sidebar = %d, want %d", got, withSidebar)
	}
}

// TestNavigatorToggleKeepsTheSidebarsSelectionIntact moves the Files Nav
// selection, hides the sidebar, restores it, and checks the selection never
// moved — hiding must not reset either Navigator pane's own state, only its
// visibility.
func TestNavigatorToggleKeepsTheSidebarsSelectionIntact(t *testing.T) {
	m := sized(t, newModel().openFolder(fixtureTree()), 80, 24)
	m = press(t, m, keys.LineDown.Keys()[0])
	m = press(t, m, keys.LineDown.Keys()[0])
	wantSelected := m.files.selected
	if wantSelected == 0 {
		t.Fatal("test setup: selection did not move")
	}

	m = press(t, m, keys.Navigator.Keys()[0]) // hide
	if got := m.files.selected; got != wantSelected {
		t.Errorf("selection while hidden = %d, want %d", got, wantSelected)
	}
	m = press(t, m, keys.Navigator.Keys()[0]) // restore
	if got := m.files.selected; got != wantSelected {
		t.Errorf("selection after restoring = %d, want %d", got, wantSelected)
	}
}

// TestNavigatorToggleMovesFocusOffAHiddenPane mirrors
// TestNarrowTerminalDropsNavigatorAndKeepsFocusVisible (frame_test.go): a
// hidden Navigator is exactly as unfocusable as a narrow-terminal-dropped one,
// so focus already on a Navigator pane must move, not vanish.
func TestNavigatorToggleMovesFocusOffAHiddenPane(t *testing.T) {
	m := sized(t, newModel(), 80, 24)
	if m.focus != filesPane {
		t.Fatalf("test setup: focus = %v, want filesPane", m.focus)
	}

	m = press(t, m, keys.Navigator.Keys()[0])
	if m.focus != dv1Pane {
		t.Errorf("focus after hiding the navigator = %v, want %v", m.focus, dv1Pane)
	}
	if m.visible(filesPane) || m.visible(rootsPane) {
		t.Error("a Navigator pane reports visible while the sidebar is hidden")
	}
}
