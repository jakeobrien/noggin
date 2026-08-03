package main

import (
	"path/filepath"
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
// replacement. The footer row is deliberately not asserted unchanged here
// (unlike the Help Modal's own version of this test): M2.3 makes the
// footer itself show the Picker's own hints while it is open (S2), so that
// row is expected to change as part of opening the Picker, not proof the
// overlay leaked past its bounds — TestFooterShowsPickerHintsWhileOpen
// covers that change directly; this test keeps only the top row (never
// touched by either the footer swap or the centered modal at 80x24).
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

// TestPickerFileResultsShowRootAndRelativePath is M3.1's own acceptance
// criterion, verbatim: a fixture with same-named files in two Roots renders
// two distinguishable rows (S3's "Root and its path relative to that Root"),
// in the order results.files already gives them (fuzzyFiles/M1.3 owns match-
// strength ordering; this only checks rendering does not reshuffle it).
func TestPickerFileResultsShowRootAndRelativePath(t *testing.T) {
	rootA := t.TempDir()
	rootB := t.TempDir()
	fileA := filepath.Join(rootA, "report.md")
	fileB := filepath.Join(rootB, "report.md")

	m := openedFixture(t, 80, 24).withRipgrep(ripgrepAvailability{available: true})
	m, _ = m.openPicker()
	m.picker.input.SetValue("report")
	m.picker.results = pickerResults{
		query: "report",
		mode:  fileMode,
		scope: activeRootScope,
		files: []pickerFile{
			{root: root{path: rootA}, path: fileA},
			{root: root{path: rootB}, path: fileB},
		},
	}

	joined := strings.Join(plainLines(m.View().Content), "\n")
	wantA := filepath.Base(rootA) + "  report.md"
	wantB := filepath.Base(rootB) + "  report.md"
	if wantA == wantB {
		t.Fatal("test fixture produced identical rows; setup is broken")
	}
	if !strings.Contains(joined, wantA) {
		t.Errorf("rendered modal is missing %q\n%s", wantA, joined)
	}
	if !strings.Contains(joined, wantB) {
		t.Errorf("rendered modal is missing %q\n%s", wantB, joined)
	}
	if ia, ib := strings.Index(joined, wantA), strings.Index(joined, wantB); ia < 0 || ib < 0 || ia > ib {
		t.Errorf("rows are out of results.files order: %q before %q wanted, got indices %d, %d", wantA, wantB, ia, ib)
	}
}

// TestPickerStatusStates is M2.5's own acceptance criterion: the rendered
// text of each of the five states (S12's mode/Scope indicator, the
// sub-threshold message, "no results" and the truncation notice; S14's
// unavailability report). Each case builds the Picker state that state
// describes and asserts against the real rendered frame, so the assertion
// covers the whole path — status selection, the Prompt box's title and
// caption rows, and the modal's own width being enough for the text.
func TestPickerStatusStates(t *testing.T) {
	// results labels a result set as answering query in File/Active Root
	// mode, which is what a Picker that has just searched holds.
	results := func(query string, files []pickerFile, truncated bool) pickerResults {
		return pickerResults{query: query, mode: fileMode, scope: activeRootScope, files: files, truncated: truncated}
	}
	found := []pickerFile{{root: root{path: "/tmp/notes"}, path: "/tmp/notes/report.md"}}

	for _, tc := range []struct {
		name  string
		avail ripgrepAvailability
		setup func(p *picker)
		want  []string
		deny  []string
	}{
		{
			name:  "mode and Scope indicator",
			avail: ripgrepAvailability{available: true},
			setup: func(p *picker) {},
			want:  []string{"File · Active Root"},
		},
		{
			name:  "Grep and All Roots indicator",
			avail: ripgrepAvailability{available: true},
			setup: func(p *picker) {
				p.mode, p.scope = grepMode, allRootsScope
			},
			want: []string{"Grep · All Roots"},
		},
		{
			name:  "below the query threshold",
			avail: ripgrepAvailability{available: true},
			setup: func(p *picker) { p.input.SetValue("re") },
			want:  []string{"type at least 3 characters"},
			deny:  []string{"no results"},
		},
		{
			name:  "no results",
			avail: ripgrepAvailability{available: true},
			setup: func(p *picker) {
				p.input.SetValue("zzz")
				p.results = results("zzz", nil, false)
			},
			want: []string{"no results"},
			deny: []string{"type at least 3 characters"},
		},
		{
			name:  "truncated at the cap",
			avail: ripgrepAvailability{available: true},
			setup: func(p *picker) {
				p.input.SetValue("report")
				p.results = results("report", found, true)
			},
			want: []string{"first 200 results (truncated)"},
			deny: []string{"no results"},
		},
		{
			name:  "ripgrep unavailable",
			avail: ripgrepAvailability{reason: "ripgrep (rg) not found on PATH"},
			setup: func(p *picker) { p.input.SetValue("report") },
			want:  []string{"Picker unavailable: ripgrep (rg) not found on PATH"},
			deny:  []string{"type at least 3 characters", "no results"},
		},
		{
			name:  "still searching",
			avail: ripgrepAvailability{available: true},
			setup: func(p *picker) {
				p.input.SetValue("report")
				p.searching = true
			},
			want: []string{"searching"},
			deny: []string{"no results"},
		},
		{
			name:  "results that answer the current query say nothing extra",
			avail: ripgrepAvailability{available: true},
			setup: func(p *picker) {
				p.input.SetValue("report")
				p.results = results("report", found, false)
			},
			deny: []string{"no results", "searching", "type at least 3 characters", "truncated"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := openedFixture(t, 80, 24).withRipgrep(tc.avail)
			m, _ = m.openPicker()
			tc.setup(&m.picker)

			joined := strings.Join(plainLines(m.View().Content), "\n")
			for _, want := range tc.want {
				if !strings.Contains(joined, want) {
					t.Errorf("rendered modal is missing %q\n%s", want, joined)
				}
			}
			for _, deny := range tc.deny {
				if strings.Contains(joined, deny) {
					t.Errorf("rendered modal unexpectedly contains %q\n%s", deny, joined)
				}
			}
		})
	}
}
