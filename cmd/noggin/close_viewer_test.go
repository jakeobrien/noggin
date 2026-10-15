package main

import "testing"

// D141's CloseViewer (ctrl+w) acts on the focused viewer: from DV2 it closes
// DV2 (Focus -> DV1); from DV1 with DV2 open, DV1 takes over DV2's whole
// state and DV2 closes; from DV1 alone it clears DV1; from a Navigator pane
// it does nothing.

func TestCloseViewerFromDV2ClosesItAndFocusesDV1(t *testing.T) {
	m, pathA, _ := crossViewerFixture(t) // focus on dv2Pane
	got := press(t, m, keys.CloseViewer.Keys()[0])
	if got.dv2 != nil {
		t.Error("dv2 still exists after CloseViewer from DV2")
	}
	if got.focus != dv1Pane {
		t.Errorf("focus = %v, want dv1Pane", got.focus)
	}
	if got.dv1.doc.path != pathA {
		t.Errorf("DV1 document = %q, want unchanged %q", got.dv1.doc.path, pathA)
	}
}

func TestCloseViewerFromDV1MovesDV2IntoDV1(t *testing.T) {
	m, pathA, pathB := crossViewerFixture(t)
	m = m.withViewerFor(dv2Pane, m.viewerFor(dv2Pane).pushHistory(pathA).pushHistory(pathB))
	m = press(t, m, keys.FocusDV1.Keys()[0])
	wantHistory := historyPaths(*m.dv2)

	got := press(t, m, keys.CloseViewer.Keys()[0])
	if got.dv2 != nil {
		t.Fatal("dv2 still exists after CloseViewer from DV1")
	}
	if got.focus != dv1Pane {
		t.Errorf("focus = %v, want dv1Pane", got.focus)
	}
	if got.dv1.doc.path != pathB {
		t.Errorf("DV1 document = %q, want DV2's %q", got.dv1.doc.path, pathB)
	}
	if h := historyPaths(got.dv1); len(h) != len(wantHistory) || h[len(h)-1] != pathB {
		t.Errorf("DV1 history = %v, want DV2's %v", h, wantHistory)
	}
	if paths := got.staleBaseline.trackedDocPaths(); len(paths) != 1 || paths[0] != pathB {
		t.Errorf("tracked doc paths = %v, want only [%q] (DV1's old document is gone)", paths, pathB)
	}
	full := got.layout().viewer.content().width
	if got.dv1.renderedWidth != full {
		t.Errorf("DV1 rendered width = %d, want the full viewer width %d", got.dv1.renderedWidth, full)
	}
}

func TestCloseViewerFromLoneDV1ClearsIt(t *testing.T) {
	m, pathA, _ := crossViewerFixture(t)
	m = press(t, m, keys.FocusDV1.Keys()[0])
	m = m.closeDV2()
	if m.dv1.doc.path != pathA {
		t.Fatalf("setup: DV1 = %q, want %q", m.dv1.doc.path, pathA)
	}

	got := press(t, m, keys.CloseViewer.Keys()[0])
	if got.dv1.doc.isOpen() {
		t.Errorf("DV1 still has %q open after CloseViewer", got.dv1.doc.path)
	}
	if got.dv1.canGoBack() || len(got.dv1.history) != 0 {
		t.Errorf("DV1 history = %v, want empty", historyPaths(got.dv1))
	}
	if got.focus != dv1Pane {
		t.Errorf("focus = %v, want dv1Pane", got.focus)
	}
	if got.message != "" {
		t.Errorf("message = %q, want none", got.message)
	}
	if paths := got.staleBaseline.trackedDocPaths(); len(paths) != 0 {
		t.Errorf("tracked doc paths = %v, want none", paths)
	}
	if !got.dv1.vp.FillHeight || got.dv1.height() == 0 {
		t.Error("cleared DV1 lost its size")
	}
}

func TestCloseViewerFromNavigatorIsNoOp(t *testing.T) {
	for _, start := range []pane{rootsPane, filesPane} {
		t.Run(start.title(), func(t *testing.T) {
			m := sized(t, newModel().openFolder(fixtureTree()), 80, 24)
			m = m.withViewerFor(dv2Pane, newViewer())
			m = sized(t, m, 80, 24)
			m.focus = start

			got := press(t, m, keys.CloseViewer.Keys()[0])
			if got.dv2 == nil {
				t.Error("CloseViewer closed DV2 from a Navigator pane")
			}
			if got.focus != start {
				t.Errorf("focus = %v, want unchanged %v", got.focus, start)
			}
		})
	}
}

// TestCloseViewerInertWhileModalOwnsKeyboard checks S4's other carve-out:
// CloseViewer does nothing while the Picker or the Help Modal owns the
// keyboard, the same exclusivity every other global binding already has
// there (model.Update's three gates run before the switch CloseViewer is
// matched in).
func TestCloseViewerInertWhileModalOwnsKeyboard(t *testing.T) {
	t.Run("picker", func(t *testing.T) {
		m := sized(t, newModel(), 80, 24)
		m = m.withViewerFor(dv2Pane, newViewer())
		m = sized(t, m, 80, 24)
		m = press(t, m, keys.Picker.Keys()[0])
		if !m.picker.open {
			t.Fatal("test setup: picker did not open")
		}

		got := press(t, m, keys.CloseViewer.Keys()[0])

		if got.dv2 == nil {
			t.Error("CloseViewer closed dv2 while the picker owned the keyboard")
		}
		if !got.picker.open {
			t.Error("CloseViewer closed the picker")
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

		got := press(t, m, keys.CloseViewer.Keys()[0])

		if got.dv2 == nil {
			t.Error("CloseViewer closed dv2 while the help modal owned the keyboard")
		}
		if !got.helpOpen {
			t.Error("CloseViewer closed the help modal")
		}
	})
}

// TestCloseViewerMergeRetargetsDV2sInFlightFetch: DV1 inherits DV2's pending
// article fetch, so its late completion (still addressed to dv2Pane) lands in
// DV1 rather than resurrecting DV2.
func TestCloseViewerMergeRetargetsDV2sInFlightFetch(t *testing.T) {
	m, _, _ := crossViewerFixture(t)
	const token = 9999
	m = m.withViewerFor(dv2Pane, m.viewerFor(dv2Pane).beginFetch(eulerArticle.Title, false, token))
	m = press(t, m, keys.FocusDV1.Keys()[0])
	m = press(t, m, keys.CloseViewer.Keys()[0])
	if m.dv2 != nil || !m.dv1.fetching() {
		t.Fatalf("setup: dv2=%v dv1.fetching=%v, want DV2 closed and DV1 carrying the fetch", m.dv2 != nil, m.dv1.fetching())
	}

	next, _ := m.Update(articleFetchMsg{pane: dv2Pane, token: token, title: eulerArticle.Title, article: eulerArticle})
	got := concrete(t, next)
	if got.dv2 != nil {
		t.Error("the late fetch resurrected DV2")
	}
	if got.dv1.doc.path != eulerArticle.URL {
		t.Errorf("DV1 document = %q, want the fetched article %q", got.dv1.doc.path, eulerArticle.URL)
	}
}
