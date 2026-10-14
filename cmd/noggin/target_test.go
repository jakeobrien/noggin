package main

import (
	"path/filepath"
	"testing"

	"noggin/internal/tree"
)

// C3/M1.4: the Files Nav's `ctrl+` targeting modifiers (S3, rebound from
// `shift+` to `ctrl+` in C3/M1.4a per D55 — mechanics unchanged). The
// fixture is peekFixture's — active.md open in DV1, peeked.md selected in
// the Files Nav, Focus on the Files Nav — which is exactly the position both
// modifiers are pressed from.

// TestActivateDV2LoadsSelectedFileIntoNewDV2AndMovesFocus is S3's load half:
// the modified chord creates DV2, loads the selection into it and moves Focus
// there, leaving DV1's own Active Document untouched. Driven once per bound
// key, since D55's `ctrl+l` fallback must do exactly what `ctrl+enter` does.
func TestActivateDV2LoadsSelectedFileIntoNewDV2AndMovesFocus(t *testing.T) {
	for _, k := range keys.ActivateDV2.Keys() {
		t.Run(k, func(t *testing.T) {
			m, activePath, peekedPath := peekFixture(t)
			if m.dv2 != nil {
				t.Fatal("setup: DV2 exists before the targeting modifier was pressed")
			}

			m = press(t, m, k)

			if m.dv2 == nil {
				t.Fatalf("%s did not create DV2", k)
			}
			if m.dv2.doc.path != peekedPath {
				t.Errorf("DV2 Active Document = %q, want %q", m.dv2.doc.path, peekedPath)
			}
			if m.focus != dv2Pane {
				t.Errorf("focus after %s = %v, want dv2Pane", k, m.focus)
			}
			if m.dv1.doc.path != activePath {
				t.Errorf("DV1 Active Document = %q, want unchanged %q", m.dv1.doc.path, activePath)
			}
			// Creating DV2 must re-lay out the frame (M1.2's 50/50 split):
			// the new viewer sized to its own half rather than left at
			// newViewer()'s zero size or cloned from DV1, and DV1 shrunk
			// out of the full width it no longer owns.
			dv1Rect, dv2Rect := m.layout().splitViewer(true)
			if want := dv1Rect.content().width; m.dv1.vp.Width() != want {
				t.Errorf("DV1 viewport width = %d, want its half of the split (%d)", m.dv1.vp.Width(), want)
			}
			if want := dv2Rect.content().width; m.dv2.vp.Width() != want {
				t.Errorf("DV2 viewport width = %d, want its half of the split (%d)", m.dv2.vp.Width(), want)
			}
		})
	}
}

// TestPeekDV2PeeksSelectedFileIntoNewDV2WithoutMovingFocus is S3's peek half:
// the same creation, but the peek lands in DV2's Peek state (not its Active
// Document) and Focus stays in the Files Nav, mirroring plain `p` exactly.
func TestPeekDV2PeeksSelectedFileIntoNewDV2WithoutMovingFocus(t *testing.T) {
	m, activePath, peekedPath := peekFixture(t)

	m = press(t, m, keys.PeekDV2.Keys()[0])

	if m.dv2 == nil {
		t.Fatal("PeekDV2 did not create DV2")
	}
	if !m.dv2.peeking() {
		t.Fatal("DV2 is not Peeking after PeekDV2")
	}
	if m.dv2.peek.doc.path != peekedPath {
		t.Errorf("DV2 Peeked document = %q, want %q", m.dv2.peek.doc.path, peekedPath)
	}
	if m.dv2.doc.isOpen() {
		t.Errorf("DV2 Active Document = %q, want none (a peek is not a load)", m.dv2.doc.path)
	}
	if m.focus != filesPane {
		t.Errorf("focus after PeekDV2 = %v, want filesPane (unchanged)", m.focus)
	}
	if m.dv1.peeking() {
		t.Error("DV1 is Peeking after a peek targeted at DV2")
	}
	if m.dv1.doc.path != activePath {
		t.Errorf("DV1 Active Document = %q, want unchanged %q", m.dv1.doc.path, activePath)
	}
}

// TestPlainActivateAndPeekStillTargetDV1 is the other half of M1.4's
// criterion: the unmodified keys are unaffected by the modifiers existing —
// they still load/peek into DV1 and never create DV2 (S3).
func TestPlainActivateAndPeekStillTargetDV1(t *testing.T) {
	t.Run("Activate", func(t *testing.T) {
		m, _, peekedPath := peekFixture(t)

		m = press(t, m, keys.Activate.Keys()[0])

		if m.dv2 != nil {
			t.Error("plain Activate created DV2")
		}
		if m.dv1.doc.path != peekedPath {
			t.Errorf("DV1 Active Document = %q, want %q", m.dv1.doc.path, peekedPath)
		}
		if m.focus != dv1Pane {
			t.Errorf("focus after plain Activate = %v, want dv1Pane", m.focus)
		}
	})
	t.Run("Peek", func(t *testing.T) {
		m, activePath, peekedPath := peekFixture(t)

		m = press(t, m, keys.Peek.Keys()[0])

		if m.dv2 != nil {
			t.Error("plain Peek created DV2")
		}
		if !m.dv1.peeking() || m.dv1.peek.doc.path != peekedPath {
			t.Errorf("DV1 Peeked document = %q (peeking=%v), want %q", m.dv1.peek.doc.path, m.dv1.peeking(), peekedPath)
		}
		if m.dv1.doc.path != activePath {
			t.Errorf("DV1 Active Document = %q, want unchanged %q", m.dv1.doc.path, activePath)
		}
		if m.focus != filesPane {
			t.Errorf("focus after plain Peek = %v, want filesPane (unchanged)", m.focus)
		}
	})
}

// TestActivateDV2ReusesExistingDV2 pins "creating it first if absent" as
// exactly that: a second targeted load replaces DV2's document in place
// rather than building a second viewer or disturbing DV1.
func TestActivateDV2ReusesExistingDV2(t *testing.T) {
	m, activePath, peekedPath := peekFixture(t)

	m = press(t, m, keys.ActivateDV2.Keys()[0])
	first := m.dv2
	if first == nil || first.doc.path != peekedPath {
		t.Fatalf("setup: DV2 not loaded with %q", peekedPath)
	}

	m = press(t, m, keys.FocusFiles.Keys()[0])
	m.files = m.files.move(-1) // back to active.md
	if m.files.selectedNode().Name != "active.md" {
		t.Fatalf("setup: selection = %q, want active.md", m.files.selectedNode().Name)
	}
	m = press(t, m, keys.ActivateDV2.Keys()[0])

	if m.dv2 == nil || m.dv2.doc.path != activePath {
		t.Fatalf("DV2 Active Document = %v, want %q", m.dv2, activePath)
	}
	if m.focus != dv2Pane {
		t.Errorf("focus = %v, want dv2Pane", m.focus)
	}
	if m.dv1.doc.path != activePath {
		t.Errorf("DV1 Active Document = %q, want unchanged %q", m.dv1.doc.path, activePath)
	}
}

// TestPeekDV2SecondPressClearsOnlyItsOwnPeek: each viewer owns its Peek state
// (S2), so the second-press toggle-off applies to the targeted viewer alone —
// a peek in DV1 survives DV2's being cleared.
func TestPeekDV2SecondPressClearsOnlyItsOwnPeek(t *testing.T) {
	m, _, peekedPath := peekFixture(t)

	m = press(t, m, keys.Peek.Keys()[0])    // peek into DV1
	m = press(t, m, keys.PeekDV2.Keys()[0]) // and into DV2
	if !m.dv1.peeking() || m.dv2 == nil || !m.dv2.peeking() {
		t.Fatalf("setup: want both viewers Peeking %q", peekedPath)
	}

	m = press(t, m, keys.PeekDV2.Keys()[0])

	if m.dv2.peeking() {
		t.Error("DV2 still Peeking after a second PeekDV2")
	}
	if !m.dv1.peeking() {
		t.Error("DV1's own Peek was cleared by DV2's toggle-off")
	}
	if m.focus != filesPane {
		t.Errorf("focus = %v, want filesPane (unchanged)", m.focus)
	}
}

// TestFocusChangeClearsBothViewersPeeks: S7's clearing triggers are
// viewer-agnostic — Peek only lives while Focus is on the Files Nav, so
// leaving it ends every peek on screen, DV2's included.
func TestFocusChangeClearsBothViewersPeeks(t *testing.T) {
	m, _, _ := peekFixture(t)

	m = press(t, m, keys.Peek.Keys()[0])
	m = press(t, m, keys.PeekDV2.Keys()[0])
	m = press(t, m, keys.FocusRoots.Keys()[0])

	if m.dv1.peeking() {
		t.Error("DV1 still Peeking after Focus left the Files Nav")
	}
	if m.dv2 == nil || m.dv2.peeking() {
		t.Error("DV2 still Peeking after Focus left the Files Nav")
	}
}

// TestTargetingModifiersRefuseADirectoryWithoutCreatingDV2: a refused action
// must not leave an empty DV2 behind — the folder no-op both unmodified keys
// already have applies before creation, not after it.
func TestTargetingModifiersRefuseADirectoryWithoutCreatingDV2(t *testing.T) {
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	mkdir(t, filepath.Join(root, "sub"))
	writeFileContent(t, filepath.Join(root, "a.md"), "# A\n")
	node, err := tree.Walk(root, tree.FilterMarkdown)
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	base := sized(t, newModel().openFolder(node), 80, 24)
	base = base.focusPane(filesPane)
	if !base.files.selectedNode().IsDir {
		t.Fatalf("setup: selection = %q, want the directory (dirs sort first)", base.files.selectedNode().Name)
	}

	for _, k := range append(keys.ActivateDV2.Keys(), keys.PeekDV2.Keys()...) {
		t.Run(k, func(t *testing.T) {
			m := press(t, base, k)
			if m.dv2 != nil {
				t.Errorf("%s on a directory created DV2", k)
			}
			if m.focus != filesPane {
				t.Errorf("focus after %s on a directory = %v, want filesPane", k, m.focus)
			}
		})
	}
}
