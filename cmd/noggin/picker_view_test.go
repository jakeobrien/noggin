package main

import (
	"path/filepath"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
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
	resultsWidth := computePickerLayout(80, 24).results.content().width
	// contentWidth mirrors pickerFileRow's own budget math: 2 columns are
	// reserved for the (here, unstarred) star marker (D72) before the Root
	// label's own truncation budget is computed.
	contentWidth := resultsWidth - 2
	wantA := "  " + pickerRootLabel(pickerFile{root: root{path: rootA}, path: fileA}, contentWidth, "  report.md") + "  report.md"
	wantB := "  " + pickerRootLabel(pickerFile{root: root{path: rootB}, path: fileB}, contentWidth, "  report.md") + "  report.md"
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

// C1/D22's Picker-row disambiguation tests (TestPickerFileResultsDisambig
// uateSameBasenameRoots, M3.1a/D50/TRIAGE T10, and TestPickerFileRowFillsIts
// FullBudget, which pinned the same parent-fragment's budget math against
// TRIAGE T13/T15/D52) are retired here, not merely updated: D117/M8.3
// dropped the Root's parent-path fragment from Picker File rows entirely
// (pickerRootLabel), so there is no longer a disambiguating budget for
// either test to pin — two Roots sharing a base name are now, by design,
// indistinguishable in a File Picker row (S8's own per-Root name is meant to
// cover that case instead). TestRootsNavDistinguishesSameNamedRootsByParent
// (rootsnav_test.go) is unaffected — D117 only ever touched Picker rows, not
// the Sources Nav's own path line. TestPickerRootLabelTruncatesALongName
// below is this file's replacement budget-pinning test, scoped to what
// pickerRootLabel actually does post-D117.

// TestPickerRootLabelTruncatesALongName is pickerRootLabel's own remaining
// budget-math case now that D117 dropped the parent fragment: a Root name
// too long for what's left of width after rest is right-truncated
// (rightTruncate, the same "keep the prefix, ellipsis the tail" rule every
// other name-like segment in this file uses), rather than overflowing the
// row for styleRow's Width/MaxWidth to silently word-wrap
// (pickerGrepRow's own doc comment describes exactly that hazard for an
// unbounded segment).
func TestPickerRootLabelTruncatesALongName(t *testing.T) {
	f := pickerFile{root: root{path: "/tmp/x", name: "A Very Long Persisted Root Name"}}
	const width = 20
	const rest = "  report.md" // 11 cells, leaving a 9-cell budget for the name

	// want is a hand-computed literal, not derived by calling rightTruncate
	// (the same independence TestPickerFileRowFillsItsFullBudget's own now-
	// retired doc comment describes): budget is 20-11=9, so leftTruncate
	// keeps the name's first 8 runes ("A Very L") plus a trailing ellipsis.
	const want = "A Very L\u2026"
	if got := pickerRootLabel(f, width, rest); got != want {
		t.Errorf("pickerRootLabel() = %q, want %q (right-truncated to the budget rest leaves)", got, want)
	}
	if w := lipgloss.Width(want + rest); w > width {
		t.Fatalf("test's own want+rest = %d cells, want at most %d (width) — fixture is broken", w, width)
	}

	// A name that already fits is left untouched — no truncation ellipsis
	// where there's room enough not to need one.
	short := pickerFile{root: root{path: "/tmp/x", name: "Work"}}
	if got := pickerRootLabel(short, width, rest); got != "Work" {
		t.Errorf("pickerRootLabel() = %q, want the untruncated name %q", got, "Work")
	}
}

// TestPickerFileRowRendersAnOrphanCandidateByItsOwnParentDirectory is S8/
// M5.1's own rendering counterpart to browse_list.go's orphan fallback
// (buildBrowseRow/renderFlatRow): a Starred-scope orphan candidate
// (pickerFile.orphan, starredPickerFiles) has no Root to disambiguate it
// with, so pickerLocationLabel shows the document's own left-truncated
// parent directory instead of a Root's own name label, and the trailing
// path segment is just the file's own base name (no meaningful
// "relative to a Root" path exists for it). Checks run against
// ansi.Strip'd text: S10/M8.3 dims the trailing suffix (dimColor), so the
// raw string carries SGR codes around it that a literal HasSuffix/Contains
// would otherwise trip on without changing what the row actually shows.
func TestPickerFileRowRendersAnOrphanCandidateByItsOwnParentDirectory(t *testing.T) {
	f := pickerFile{path: "/home/user/elsewhere/orphan.md", orphan: true}
	const width = 20

	got := ansi.Strip(pickerFileRow(f, "", width, false, false))
	wantSuffix := "  orphan.md"
	if !strings.HasSuffix(strings.TrimRight(got, " "), wantSuffix) {
		t.Fatalf("pickerFileRow(orphan) = %q, want it to end (ignoring trailing pad) with %q (the file's own base name, no Root-relative path)", got, wantSuffix)
	}
	if strings.Contains(got, "/home/user/elsewhere") {
		t.Errorf("pickerFileRow(orphan) = %q, expected the parent directory left-truncated away at this width, not shown in full", got)
	}
	if w := lipgloss.Width(got); w != width {
		t.Errorf("pickerFileRow(orphan) width = %d, want %d", w, width)
	}

	// Wide enough that leftTruncate has no need to truncate: the orphan's
	// full parent directory shows plainly, mirroring what a Root's own name
	// label would show for a non-orphan candidate at the same width.
	const wideWidth = 80
	wide := ansi.Strip(pickerFileRow(f, "", wideWidth, false, false))
	if !strings.Contains(wide, "/home/user/elsewhere") {
		t.Errorf("pickerFileRow(orphan) at width %d = %q, want the full parent directory visible", wideWidth, wide)
	}
}

// TestPickerFileRowShowsStarMarkerForAStarredResult closes the S3/D72 gap
// TRIAGE T8 found: S2/M2.5's own wording already talked about a Picker
// Result's "star marker" being cleared, but no Picker row actually rendered
// one. A starred Result's row now carries the same "\u2605 "/"  " two-column
// field Files Nav and Sources Nav rows already use, prepended ahead of the
// Root label rather than appended — the marker is styling-adjacent state,
// not part of the Root label text pickerRootLabel computes (D117/M8.3: no
// longer disambiguating, just the Root's own name).
func TestPickerFileRowShowsStarMarkerForAStarredResult(t *testing.T) {
	f := pickerFile{root: root{path: "/home/user/notes"}, path: "/home/user/notes/report.md"}
	const width = 40

	starred := pickerFileRow(f, "", width, false, true)
	unstarred := pickerFileRow(f, "", width, false, false)

	if !strings.HasPrefix(starred, "\u2605 ") {
		t.Errorf("starred pickerFileRow() = %q, want it to start with the \"\u2605 \" marker", starred)
	}
	if strings.Contains(unstarred, "\u2605") {
		t.Errorf("unstarred pickerFileRow() = %q, want no star glyph", unstarred)
	}
	if !strings.HasPrefix(unstarred, "  ") {
		t.Errorf("unstarred pickerFileRow() = %q, want it to start with two blank columns in the marker's place", unstarred)
	}
	// The marker occupies its own 2-column field — the rest of the row (Root
	// label, path, file name) must be byte-identical whether starred or not,
	// and both rows must still fill exactly width.
	if got, want := strings.TrimPrefix(starred, "\u2605 "), strings.TrimPrefix(unstarred, "  "); got != want {
		t.Errorf("starred/unstarred rows differ beyond the marker field:\n starred (marker stripped) = %q\nunstarred (marker stripped) = %q", got, want)
	}
	if w := lipgloss.Width(starred); w != width {
		t.Errorf("starred pickerFileRow() width = %d, want %d", w, width)
	}
}

// TestPickerGrepRowShowsStarMarkerForAStarredResult is
// TestPickerFileRowShowsStarMarkerForAStarredResult's Grep Picker mirror
// (S2 names "the selected Result" without restricting it to File mode, the
// same reasoning TestPickerStarTogglesSelectedGrepResult already gives for
// the toggle itself).
func TestPickerGrepRowShowsStarMarkerForAStarredResult(t *testing.T) {
	f := pickerFile{root: root{path: "/tmp/notes"}, path: "/tmp/notes/report.md"}
	g := grepResult{file: f, line: 5, text: "the quarterly report is late"}
	const width = 40

	starred := pickerGrepRow(g, "", width, false, true)
	unstarred := pickerGrepRow(g, "", width, false, false)

	if !strings.HasPrefix(starred, "\u2605 ") {
		t.Errorf("starred pickerGrepRow() = %q, want it to start with the \"\u2605 \" marker", starred)
	}
	if strings.Contains(unstarred, "\u2605") {
		t.Errorf("unstarred pickerGrepRow() = %q, want no star glyph", unstarred)
	}
	if got, want := strings.TrimPrefix(starred, "\u2605 "), strings.TrimPrefix(unstarred, "  "); got != want {
		t.Errorf("starred/unstarred rows differ beyond the marker field:\n starred (marker stripped) = %q\nunstarred (marker stripped) = %q", got, want)
	}
	if w := lipgloss.Width(starred); w != width {
		t.Errorf("starred pickerGrepRow() width = %d, want %d", w, width)
	}
}

// TestPickerGrepRowRendersAnOrphanCandidateByItsFullPath is S8/M5.1's Grep
// Picker counterpart to TestPickerFileRowRendersAnOrphanCandidateByItsOwn-
// ParentDirectory: a Starred-scope orphan Grep Result (g.file.orphan, zero-
// value g.file.root) has no Root to be relative to, so the row's own
// filepath.Rel(g.file.root.path, g.file.path) call errors (an empty base)
// and falls back to the file's full absolute path — pickerGrepRow's own
// pre-existing fallback, needing no orphan-specific branch (its doc comment
// records why: this row never disambiguates by Root in the first place).
func TestPickerGrepRowRendersAnOrphanCandidateByItsFullPath(t *testing.T) {
	f := pickerFile{path: "/home/user/elsewhere/orphan.md", orphan: true}
	g := grepResult{file: f, line: 3, text: "a needle in an orphan"}
	const width = 80

	got := pickerGrepRow(g, "", width, false, false)
	if !strings.Contains(got, "/home/user/elsewhere/orphan.md:3") {
		t.Errorf("pickerGrepRow(orphan) = %q, want the file's full absolute path and line number", got)
	}
}

// TestPickerGrepResultsShowFileLineNumberAndText is M4.1's own acceptance
// criterion's first half (S4): a Grep Result set renders one Results row per
// matching line, each showing the file (relative to its Root, the same
// filepath.Rel base pickerFileRow's own "path relative to that Root" uses),
// the correct 1-based line number and the line's own text — the same
// "set results directly, assert the rendered rows" shape
// TestPickerFileResultsShowRootAndRelativePath already gives the File
// Picker's own row (M3.1), exercised through grepMode's rendering path
// (pickerResultsContent/pickerGrepRow) instead of fileMode's.
//
// wantA/wantB (report.md/notes.md) are still derived by calling
// pickerGrepRow itself, same as before — useful as a sanity check that the
// fixture produces the right file:line prefixes, but unable to catch a bug
// in pickerGrepRow's own text handling (TRIAGE T22/D55: a tautological
// assertion, the same shape T15/D52 already fixed for the file row test).
// wantC (short.md) closes that gap: its expected row is a literal computed
// by hand from styleRow's left-align-and-pad behaviour (confirmed against
// the real renderer, not assumed) rather than by calling pickerGrepRow, and
// short.md's own "hi" text is short enough that rightTruncate never cuts it
// — so the full literal (including the match's line text) has to appear
// verbatim in the rendered modal for this to pass, which a dropped- or
// mangled-text bug in pickerGrepRow would fail.
func TestPickerGrepResultsShowFileLineNumberAndText(t *testing.T) {
	rootDir := t.TempDir()
	fileA := filepath.Join(rootDir, "report.md")
	fileB := filepath.Join(rootDir, "notes.md")
	fileC := filepath.Join(rootDir, "short.md")
	r := root{path: rootDir}

	m := openedFixture(t, 80, 24).withRipgrep(ripgrepAvailability{available: true})
	m, _ = m.openPicker()
	m.picker.mode = grepMode
	m.picker.input.SetValue("report")
	m.picker.results = pickerResults{
		query: "report",
		mode:  grepMode,
		scope: activeRootScope,
		grep: []grepResult{
			{file: pickerFile{root: r, path: fileA}, line: 3, text: "the quarterly report"},
			{file: pickerFile{root: r, path: fileB}, line: 12, text: "see report above"},
			{file: pickerFile{root: r, path: fileC}, line: 1, text: "hi"},
		},
	}

	joined := strings.Join(plainLines(m.View().Content), "\n")
	resultsWidth := computePickerLayout(80, 24).results.content().width
	// Rendered with an empty query, so the expected strings are plain text:
	// joined is ANSI-stripped, and match highlighting (M6.2) is styling only
	// — it never changes a row's characters, which is itself asserted by
	// TestPickerGrepRowHighlightsTheQueryOccurrence.
	wantA := pickerGrepRow(m.picker.results.grep[0], "", resultsWidth, false, false)
	wantB := pickerGrepRow(m.picker.results.grep[1], "", resultsWidth, false, false)
	if !strings.Contains(wantA, "report.md:3") || !strings.Contains(wantB, "notes.md:12") {
		t.Fatalf("test fixture did not produce the expected file:line prefixes: %q, %q", wantA, wantB)
	}
	if !strings.Contains(joined, wantA) {
		t.Errorf("rendered modal is missing %q\n%s", wantA, joined)
	}
	if !strings.Contains(joined, wantB) {
		t.Errorf("rendered modal is missing %q\n%s", wantB, joined)
	}

	// wantC is hand-computed, not derived from pickerGrepRow: "short.md:1  hi"
	// (14 runes), prefixed by the (unstarred) 2-column star marker (D72) and
	// padded with spaces to resultsWidth (23 at this fixture's size) by
	// styleRow's Width(resultsWidth), since 14 < 21 (resultsWidth minus the
	// marker) means rightTruncate is a no-op and nothing needs truncating.
	if resultsWidth != 23 {
		t.Fatalf("computePickerLayout(80, 24).results.content().width = %d, want 23 — wantC below is hand-computed for exactly that width", resultsWidth)
	}
	wantC := "  " + "short.md:1  hi" + strings.Repeat(" ", resultsWidth-2-len("short.md:1  hi"))
	if !strings.Contains(joined, wantC) {
		t.Errorf("rendered modal is missing %q (short.md's match line text \"hi\", hand-computed independently of pickerGrepRow)\n%s", wantC, joined)
	}
}

// TestPickerGrepRowTruncatesLongLinesWithoutWrapping is M4.1's own
// acceptance criterion's second half: a matching line far longer than the
// Results list's width is truncated to fit the pane rather than wrapped
// across multiple rows — pickerGrepRow must still return exactly one line
// at exactly the given width, unlike a bare Style.Width(...).Render() of an
// over-length string (lipgloss word-wraps before MaxWidth ever applies, see
// pickerGrepRow's own doc comment), which would grow the row into several.
func TestPickerGrepRowTruncatesLongLinesWithoutWrapping(t *testing.T) {
	f := pickerFile{root: root{path: "/tmp/notes"}, path: "/tmp/notes/report.md"}
	longText := strings.Repeat("the matched line keeps going ", 10)
	g := grepResult{file: f, line: 5, text: longText}
	const width = 23 // computePickerLayout(80, 24).results.content().width

	// Empty query: this pins the row's truncation layout, with no match
	// highlighting in play (M6.2 covers a truncated row's highlight in
	// TestPickerGrepRowHighlightsTheQueryOccurrence).
	got := pickerGrepRow(g, "", width, false, false)

	if lines := strings.Split(got, "\n"); len(lines) != 1 {
		t.Fatalf("pickerGrepRow() produced %d lines, want exactly 1 (truncated, not wrapped):\n%q", len(lines), got)
	}
	if w := lipgloss.Width(got); w != width {
		t.Errorf("pickerGrepRow() rendered width = %d, want exactly %d", w, width)
	}
	if strings.Contains(got, longText) {
		t.Errorf("pickerGrepRow() rendered the full untruncated line: %q", got)
	}
	if !strings.Contains(got, "report.md:5") {
		t.Errorf("pickerGrepRow() dropped the file:line prefix under truncation: %q", got)
	}
}

// TestPickerStatusStates is M2.5's own acceptance criterion: the rendered
// text of each of the five states (S12/C2's mode/Scope indicator, the
// sub-threshold message, "no results" and the truncation notice; S14's
// unavailability report). Each case builds the Picker state that state
// describes and asserts against the real rendered frame, so the assertion
// covers the whole path — status selection, the Prompt box's title and
// caption rows, and the modal's own width being enough for the text.
func TestPickerStatusStates(t *testing.T) {
	// results labels a result set as answering query in File/Active Source
	// mode, which is what a Picker that has just searched holds.
	results := func(query string, files []pickerFile, truncated bool) pickerResults {
		return pickerResults{query: query, mode: fileMode, scope: activeRootScope, files: files, truncated: truncated}
	}
	found := []pickerFile{{root: root{path: "/tmp/notes"}, path: "/tmp/notes/report.md"}}

	for _, tc := range []struct {
		name         string
		avail        ripgrepAvailability
		hideDotfiles bool
		setup        func(p *picker)
		want         []string
		deny         []string
	}{
		{
			name:  "mode and Scope indicator",
			avail: ripgrepAvailability{available: true},
			setup: func(p *picker) {},
			want:  []string{"File · Active Source"},
		},
		{
			// TRIAGE T7 / DECISIONS.md D50: the alt+h dotfile-hiding toggle's
			// only visible confirmation is this " · Hidden" suffix on the
			// same title row the mode/Scope indicator above already covers —
			// previously only test/pty/checks.py exercised the string.
			name:         "mode and Scope indicator with dotfiles hidden",
			avail:        ripgrepAvailability{available: true},
			hideDotfiles: true,
			setup:        func(p *picker) {},
			want:         []string{"File · Active Source · Hidden"},
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
			// S2/C6/M3.2: Wikipedia mode has no Scope, reading "-" regardless
			// of p.scope's own value (left at allRootsScope here on purpose,
			// to prove it is genuinely ignored rather than incidentally zero).
			name:  "Wikipedia mode indicator reads a dash Scope",
			avail: ripgrepAvailability{available: true},
			setup: func(p *picker) {
				p.mode, p.scope = wikipediaMode, allRootsScope
			},
			want: []string{"Wikipedia · -"},
			deny: []string{"All Roots"},
		},
		{
			name:  "below the query threshold",
			avail: ripgrepAvailability{available: true},
			setup: func(p *picker) { p.input.SetValue("r") },
			want:  []string{"type at least 2 characters"},
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
			deny: []string{"type at least 2 characters"},
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
			deny:  []string{"type at least 2 characters", "no results"},
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
			deny: []string{"no results", "searching", "type at least 2 characters", "truncated"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := openedFixture(t, 80, 24).withRipgrep(tc.avail)
			m.hideDotfiles = tc.hideDotfiles
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

// cellStyle is the styling in force at one visible cell of a rendered row,
// as far as the Picker's row highlighting cares: bold (the match emphasis,
// M6.1) and reverse video (the selection bar, pickerSelectedStyle).
type cellStyle struct {
	bold, reverse, dim bool
}

// cellStyles walks a rendered string's SGR escape sequences and reports the
// style in force at each visible cell — the only way to check *where* a
// highlight landed, which ansi.Strip (plainLines) deliberately discards and
// a string comparison against a hand-built expected string could only check
// by pinning lipgloss's exact escape serialization. Handles the subset
// lipgloss emits here: bold on/off (1/21/22), reverse on/off (7/27), the
// ANSI256 foreground sequence (38;5;<n>) S10/M8.3's dim styling
// (dimColor/rootsPathStyle/locationLineStyle, all lipgloss.Color("245"))
// serializes as — tracked as dim only when n is exactly dimColorCode, not
// any other foreground colour — and a full reset (0, or the empty
// parameter list lipgloss writes as `ESC[m`).
func cellStyles(s string) []cellStyle {
	var (
		cells   []cellStyle
		current cellStyle
		runes   = []rune(s)
	)
	for i := 0; i < len(runes); i++ {
		if runes[i] == '\x1b' && i+1 < len(runes) && runes[i+1] == '[' {
			j := i + 2
			for j < len(runes) && runes[j] != 'm' {
				j++
			}
			if j < len(runes) {
				params := string(runes[i+2 : j])
				if params == "" {
					current = cellStyle{}
				}
				tokens := strings.Split(params, ";")
				for k := 0; k < len(tokens); k++ {
					switch tokens[k] {
					case "", "0":
						current = cellStyle{}
					case "1":
						current.bold = true
					case "21", "22":
						current.bold = false
					case "7":
						current.reverse = true
					case "27":
						current.reverse = false
					case "38":
						// ANSI256 foreground is "38;5;<n>"; anything else
						// (e.g. truecolor's "38;2;r;g;b") is a foreground this
						// codebase never emits and is left un-classified as dim.
						if k+2 < len(tokens) && tokens[k+1] == "5" {
							current.dim = tokens[k+2] == dimColorCode
							k += 2
						}
					case "39":
						current.dim = false
					}
				}
			}
			i = j
			continue
		}
		cells = append(cells, current)
	}
	return cells
}

// dimColorCode is dimColor's (style.go) own ANSI256 code, as a string —
// what cellStyles matches "38;5;<n>" against to tell S10/M8.3's
// de-emphasis apart from some other foreground colour.
const dimColorCode = "245"

// boldColumns is which cells of a rendered row carry the match emphasis.
func boldColumns(s string) []int {
	var cols []int
	for i, c := range cellStyles(s) {
		if c.bold {
			cols = append(cols, i)
		}
	}
	return cols
}

// dimColumns is boldColumns' own S10/M8.3 counterpart: which cells of a
// rendered row carry the de-emphasized (dimColor) styling.
func dimColumns(s string) []int {
	var cols []int
	for i, c := range cellStyles(s) {
		if c.dim {
			cols = append(cols, i)
		}
	}
	return cols
}

// TestPickerFileRowHighlightsMatchedNameCharacters is M6.1's own acceptance
// criterion (S5): the styled runs in a rendered File Picker row fall exactly
// on the fuzzy-matched positions of the file name, and nowhere else. The
// expected columns are hand-computed rather than derived from the highlighter
// — the fuzzy match of "rpt" against "report.md" can only be r(0), p(2),
// t(5) (those are the name's only r/p/t after the first, in order), offset by
// where the name starts in the row — so a highlighter that drifted onto the
// path, the Root label or the wrong characters of the name fails here.
// Synthetic (non-tempdir) paths keep every column knowable.
func TestPickerFileRowHighlightsMatchedNameCharacters(t *testing.T) {
	f := pickerFile{
		root: root{path: "/home/user/notes"},
		path: "/home/user/notes/sub/report.md",
	}
	const width = 40
	const name = "report.md"

	got := pickerFileRow(f, "rpt", width, false, false)

	// The highlight is styling only: the row's displayed characters must be
	// exactly the unhighlighted row's. Both sides are ansi.Strip'd — unlike
	// before S10/M8.3, the "unhighlighted" call still carries the path
	// segment's unconditional dim styling (dimColor), so it is not itself
	// plain text to compare against directly.
	plain := ansi.Strip(got)
	if want := ansi.Strip(pickerFileRow(f, "", width, false, false)); plain != want {
		t.Fatalf("highlighted row's plain text = %q, want %q (highlighting must not change the row's content or width)", plain, want)
	}

	nameAt := strings.LastIndex(plain, name)
	if nameAt < 0 {
		t.Fatalf("row %q does not contain %q; fixture is broken", plain, name)
	}
	// "report.md": r=0, e=1, p=2, o=3, r=4, t=5 — the match for "rpt".
	want := []int{nameAt + 0, nameAt + 2, nameAt + 5}
	if gotCols := boldColumns(got); !equalInts(gotCols, want) {
		t.Errorf("bold columns = %v, want %v\nrow: %q\nplain: %q", gotCols, want, got, plain)
	}
	// The Root label and the path components are matched against nothing
	// (D31: the File Picker matches the file name only), so nothing before
	// the name may be emphasized — implied by the exact comparison above,
	// asserted separately so a failure says which half broke.
	for _, c := range boldColumns(got) {
		if c < nameAt {
			t.Errorf("bold column %d falls before the file name (starts at %d) — highlighting escaped into the Root label or path", c, nameAt)
		}
	}
}

// TestPickerFileRowHighlightKeepsSelectionBar guards the nesting hazard
// highlightMatches' own comment describes: lipgloss closes an inner style
// with a full reset, so highlighting a selected row by wrapping it in
// pickerSelectedStyle after the fact would cancel the reverse-video bar from
// the first matched character onward. Every cell of a selected row — padding
// included — must stay reversed, while the matched cells are additionally
// bold.
func TestPickerFileRowHighlightKeepsSelectionBar(t *testing.T) {
	f := pickerFile{
		root: root{path: "/home/user/notes"},
		path: "/home/user/notes/sub/report.md",
	}
	const width = 40

	got := pickerFileRow(f, "rpt", width, true, false)
	cells := cellStyles(got)
	if len(cells) != width {
		t.Fatalf("selected row rendered %d cells, want %d", len(cells), width)
	}
	for i, c := range cells {
		if !c.reverse {
			t.Fatalf("cell %d of the selected row is not reversed — the selection bar is broken from there on\nrow: %q", i, got)
		}
	}
	if bold := boldColumns(got); len(bold) != 3 {
		t.Errorf("selected row bold columns = %v, want 3 (the match's own characters)", bold)
	}
}

// TestPickerFileRowDropsTheParentPathFragment is S10/D117/M8.3's own
// acceptance criterion, the fragment-absence half: a row is exactly
// "<Root name>" + path relative to the Root, with no left-truncated parent
// segment (leftTruncate's own "…" ellipsis, or any part of the Root's own
// parent directory) in between — unlike the pre-D117 layout
// TestPickerFileResultsDisambiguateSameBasenameRoots/TestPickerFileRowFills
// ItsFullBudget (both retired, see the note above
// TestPickerRootLabelTruncatesALongName) used to pin.
func TestPickerFileRowDropsTheParentPathFragment(t *testing.T) {
	f := pickerFile{
		root: root{path: "/home/user/work/admin/project"},
		path: "/home/user/work/admin/project/report.md",
	}
	const width = 40

	got := ansi.Strip(pickerFileRow(f, "", width, false, false))
	want := "  project  report.md" + strings.Repeat(" ", width-len("  project  report.md"))
	if got != want {
		t.Errorf("pickerFileRow() = %q, want %q (Root name + path relative to the Root, no parent fragment)", got, want)
	}
	if strings.Contains(got, "…") || strings.Contains(got, "admin") || strings.Contains(got, "work") {
		t.Errorf("pickerFileRow() = %q, want no trace of the Root's parent path (D117 dropped the fragment)", got)
	}
}

// TestPickerFileRowDimsThePathButNotTheRootLabel is S10/M8.3's own
// acceptance criterion, the styling half: the Root label (pickerLocation
// Label) renders at normal weight, and everything from the path's own gap
// onward (S10's "the file's path") renders de-emphasized (dimColor) — with
// no query, so no matched-character emphasis is in play (that's
// TestPickerFileRowMatchedPathCharactersAreNeverDimmed below).
func TestPickerFileRowDimsThePathButNotTheRootLabel(t *testing.T) {
	f := pickerFile{root: root{path: "/home/user/notes"}, path: "/home/user/notes/sub/report.md"}
	const width = 40

	got := pickerFileRow(f, "", width, false, false)
	plain := ansi.Strip(got)
	labelAt := strings.Index(plain, "notes")
	pathAt := strings.Index(plain, "sub/report.md")
	if labelAt < 0 || pathAt < 0 {
		t.Fatalf("row %q missing the Root label or path; fixture is broken", plain)
	}

	dim := dimColumns(got)
	for i := labelAt; i < labelAt+len("notes"); i++ {
		for _, d := range dim {
			if d == i {
				t.Errorf("column %d (Root label \"notes\") is dimmed, want normal weight\nrow: %q", i, got)
			}
		}
	}
	// Only the row's own content is checked, not styleRow's trailing pad
	// past it (an unselected row's pad renders with no properties at all,
	// dim included — the same behaviour an unselected Bold row's own pad
	// already has, TestRootsNavRendersTwoLinesPerEntry's own comment notes
	// it for reverse+bold): S10 only asks that the path *text* read
	// de-emphasized, not that trailing blank cells carry any styling.
	contentEnd := strings.TrimRight(plain, " ")
	for i := pathAt; i < len(contentEnd); i++ {
		found := false
		for _, d := range dim {
			if d == i {
				found = true
			}
		}
		if !found {
			t.Errorf("column %d (path %q) is not dimmed\nrow: %q", i, plain[i:], got)
			break
		}
	}
}

// TestPickerFileRowMatchedPathCharactersAreNeverDimmed is S10's own
// "matched characters of the path emphasized against it": a matched file-
// name character that falls inside the de-emphasized path segment (every
// File Picker row's match is inside that segment — D31, the Root label is
// never matched against) is bold and never dim, even though every other
// character alongside it is dimmed.
func TestPickerFileRowMatchedPathCharactersAreNeverDimmed(t *testing.T) {
	f := pickerFile{root: root{path: "/home/user/notes"}, path: "/home/user/notes/sub/report.md"}
	const width = 40

	got := pickerFileRow(f, "rpt", width, false, false)
	bold := boldColumns(got)
	if len(bold) == 0 {
		t.Fatalf("row %q has no bold (matched) columns; fixture/query is broken", got)
	}
	dim := make(map[int]bool)
	for _, d := range dimColumns(got) {
		dim[d] = true
	}
	for _, b := range bold {
		if dim[b] {
			t.Errorf("matched column %d is both bold and dim, want emphasis to win (S10's \"emphasized against it\")\nrow: %q", b, got)
		}
	}
	// The dim segment must still have plenty of non-matched, actually-dimmed
	// columns — otherwise this test would trivially pass by having nothing to
	// dim in the first place.
	if len(dimColumns(got)) == 0 {
		t.Fatal("row has no dimmed columns at all; fixture is broken")
	}
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// columnRange is n consecutive columns starting at start — the shape a
// literal query's occurrence takes in a rendered row.
func columnRange(start, n int) []int {
	cols := make([]int, 0, n)
	for i := 0; i < n; i++ {
		cols = append(cols, start+i)
	}
	return cols
}

// TestPickerGrepRowHighlightsTheQueryOccurrence is M6.2's own acceptance
// criterion (S5): the styled run covers the query's occurrence within the
// matched line, including in a row whose line was truncated. Expected
// columns are hand-computed from the row's own construction — the prefix
// "report.md:5  " is 13 cells (9 + ":" + "5" + two spaces), so a match at
// index i of the line's text lands at column 13+i — rather than being read
// back out of the highlighter.
//
// The fixture file is deliberately named report.md while the query is
// "report": the prefix contains the query too, and highlighting it would be
// a lie about what rg matched (it matched the line, not the path), so every
// case here also pins that the prefix is never emphasized.
func TestPickerGrepRowHighlightsTheQueryOccurrence(t *testing.T) {
	f := pickerFile{root: root{path: "/tmp/notes"}, path: "/tmp/notes/report.md"}
	const prefix = 15 // marker(2, D72) + "report.md:5  "(13)

	cases := []struct {
		name     string
		text     string
		query    string
		width    int
		want     []int
		wantPlan string
	}{
		{
			name:     "occurrence within the line",
			text:     "the quarterly report is late",
			query:    "report",
			width:    60,
			want:     columnRange(prefix+14, 6), // "the quarterly " is 14 cells
			wantPlan: "the line's own occurrence, not the report.md prefix",
		},
		{
			name:     "smart case: lowercase query matches mixed case",
			text:     "the quarterly Report is late",
			query:    "report",
			width:    60,
			want:     columnRange(prefix+14, 6),
			wantPlan: "rg's --smart-case matched it, so the highlight must too",
		},
		{
			name:     "smart case: uppercase query does not match lowercase",
			text:     "the quarterly report is late",
			query:    "Report",
			width:    60,
			want:     nil,
			wantPlan: "an uppercase query is case-sensitive (S8), so nothing in this line matched",
		},
		{
			name:     "every occurrence in the line",
			text:     "report and report",
			query:    "report",
			width:    60,
			want:     append(columnRange(prefix, 6), columnRange(prefix+11, 6)...), // "report and " is 11 cells
			wantPlan: "rg reports the line once however often the query occurs in it",
		},
		{
			// width 25 leaves a content width of 23 (25 minus the 2-column
			// star marker, D72) — the same content budget this case pinned
			// before the marker existed — so the row keeps 22 cells of text
			// plus rightTruncate's ellipsis. The match starts at content
			// column 13+7=20 and runs to 25, so only its first two
			// characters survive the cut; shifted by the marker's own 2
			// columns, that lands at absolute columns 22 and 23.
			name:     "truncated row: only the visible part of the match",
			text:     "xxxxxxxreport and more text past the cut",
			query:    "report",
			width:    25,
			want:     []int{22, 23},
			wantPlan: "the rest of the match was truncated away",
		},
		{
			name:     "truncated row: match entirely past the cut",
			text:     strings.Repeat("x", 30) + "report",
			query:    "report",
			width:    23,
			want:     nil,
			wantPlan: "the whole match was truncated away, so there is nothing to emphasize",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			g := grepResult{file: f, line: 5, text: c.text}
			got := pickerGrepRow(g, c.query, c.width, false, false)

			// Highlighting is styling only: the row's characters and width
			// must be exactly the unhighlighted row's.
			plain := ansi.Strip(got)
			if want := pickerGrepRow(g, "", c.width, false, false); plain != want {
				t.Fatalf("highlighted row's plain text = %q, want %q", plain, want)
			}
			if w := lipgloss.Width(got); w != c.width {
				t.Errorf("rendered width = %d, want %d", w, c.width)
			}
			if cols := boldColumns(got); !equalInts(cols, c.want) {
				t.Errorf("bold columns = %v, want %v (%s)\nrow: %q\nplain: %q", cols, c.want, c.wantPlan, got, plain)
			}
		})
	}

	// A selected grep row keeps its reverse-video bar across every cell,
	// highlight included — the same nesting hazard
	// TestPickerFileRowHighlightKeepsSelectionBar pins for file rows, now
	// that both share highlightMatches.
	g := grepResult{file: f, line: 5, text: "the quarterly report is late"}
	selected := pickerGrepRow(g, "report", 60, true, false)
	for i, c := range cellStyles(selected) {
		if !c.reverse {
			t.Fatalf("cell %d of the selected grep row is not reversed\nrow: %q", i, selected)
		}
	}
}
