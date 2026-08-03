package main

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

// pickerSizes exercises the modal at a range of terminal sizes, the same
// spread frameSizes gives the main frame, plus a narrow terminal
// (TestHelpOverlayNeverExceedsTheTerminal's own size) where the modal's
// clamped footprint matters most.
var pickerSizes = []struct {
	name          string
	width, height int
}{
	{"standard 80x24", 80, 24},
	{"large 120x40", 120, 40},
	{"narrow 40x10", 40, 10},
}

// TestPickerModalRendersThreeRegionsAtSize is M2.2's own acceptance
// criterion's first half: opening the Picker renders all three regions
// (Prompt, Results, Preview) at several terminal sizes. Results/Preview
// carry no content yet (M3.1/M3.2), so they are identified by their titles;
// the Prompt is identified by its border occupying the modal's own width.
func TestPickerModalRendersThreeRegionsAtSize(t *testing.T) {
	for _, s := range pickerSizes {
		t.Run(s.name, func(t *testing.T) {
			m := openedFixture(t, s.width, s.height)
			m = press(t, m, keys.Picker.Keys()[0])
			if !m.picker.open {
				t.Fatal("picker did not open; setup is broken")
			}

			joined := strings.Join(plainLines(m.View().Content), "\n")
			if !strings.Contains(joined, "Results") {
				t.Error("rendered view is missing the Results region")
			}
			if !strings.Contains(joined, "Preview") {
				t.Error("rendered view is missing the Preview region")
			}

			l := computePickerLayout(s.width, s.height)
			if l.prompt.empty() {
				t.Fatal("computePickerLayout gave an empty Prompt region")
			}
			if l.results.empty() {
				t.Error("computePickerLayout gave an empty Results region")
			}
			if l.preview.empty() {
				t.Error("computePickerLayout gave an empty Preview region")
			}
		})
	}
}

// TestPickerModalNeverExceedsTheTerminal is M2.2's own acceptance
// criterion's second half, in the exact shape of
// TestHelpOverlayNeverExceedsTheTerminal: at a terminal too small for the
// modal's proportional footprint, the rendered frame is still exactly
// width x height, never larger.
func TestPickerModalNeverExceedsTheTerminal(t *testing.T) {
	for _, s := range pickerSizes {
		t.Run(s.name, func(t *testing.T) {
			m := openedFixture(t, s.width, s.height)
			m = press(t, m, keys.Picker.Keys()[0])

			view := m.View().Content
			lines := strings.Split(view, "\n")
			if len(lines) != s.height {
				t.Fatalf("rendered frame is %d rows, want exactly %d (the terminal's height)", len(lines), s.height)
			}
			for i, l := range lines {
				if w := lipgloss.Width(l); w > s.width {
					t.Errorf("row %d is %d cells wide, want at most %d (the terminal's width)", i, w, s.width)
				}
			}
		})
	}
}

// TestPickerModalLeavesTheFrameVisibleAroundIt is the compositing half of
// S1's "openable... composed of a Prompt, a Results list and a Preview
// pane": the exact TestHelpOverlayLeavesTheFrameVisibleAroundIt shape,
// checking that the pane underneath the modal is still showing outside its
// bounding box, proving real cell compositing rather than a full-screen
// replacement.
func TestPickerModalLeavesTheFrameVisibleAroundIt(t *testing.T) {
	m := sized(t, newModel().openFolder(fixtureTree()), 80, 24)
	beforeLines := plainLines(m.View().Content)

	m = press(t, m, keys.Picker.Keys()[0])
	afterLines := plainLines(m.View().Content)

	if len(afterLines) != len(beforeLines) {
		t.Fatalf("line count changed: %d vs %d", len(afterLines), len(beforeLines))
	}
	if afterLines[0] != beforeLines[0] {
		t.Errorf("top row changed under the modal overlay:\nbefore: %q\nafter:  %q", beforeLines[0], afterLines[0])
	}
	if strings.TrimRight(afterLines[len(afterLines)-1], " ") != strings.TrimRight(beforeLines[len(beforeLines)-1], " ") {
		t.Errorf("footer row changed under the modal overlay:\nbefore: %q\nafter:  %q",
			beforeLines[len(beforeLines)-1], afterLines[len(afterLines)-1])
	}
	changed := false
	for i := range afterLines {
		if strings.TrimRight(afterLines[i], " ") != strings.TrimRight(beforeLines[i], " ") {
			changed = true
			break
		}
	}
	if !changed {
		t.Error("no row changed; the modal apparently drew nothing")
	}
}

// TestComputePickerLayoutNeverExceedsTheTerminal is the layout-level
// counterpart to TestPickerModalNeverExceedsTheTerminal: at a range of
// sizes including degenerate tiny ones, every region's rect fits inside the
// terminal, and Results/Preview never overlap or exceed the modal's own
// width.
func TestComputePickerLayoutNeverExceedsTheTerminal(t *testing.T) {
	for _, size := range []struct{ width, height int }{
		{80, 24}, {120, 40}, {40, 10}, {10, 5}, {1, 1}, {0, 0},
	} {
		l := computePickerLayout(size.width, size.height)
		if l.prompt.width > size.width || l.results.width+l.preview.width > size.width {
			t.Errorf("size %v: widths = prompt %d, results+preview %d, want <= %d",
				size, l.prompt.width, l.results.width+l.preview.width, size.width)
		}
		if l.prompt.height+l.results.height > size.height || l.prompt.height+l.preview.height > size.height {
			t.Errorf("size %v: total height exceeds %d", size, size.height)
		}
		if l.results.width < 0 || l.preview.width < 0 || l.prompt.height < 0 {
			t.Errorf("size %v: negative region: %+v", size, l)
		}
	}
}
