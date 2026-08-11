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

// TestPickerFileResultsDisambiguateSameBasenameRoots is M3.1a's own
// acceptance criterion (D50/TRIAGE T10): two Roots sharing a base name but
// different parents must render distinguishable Root labels in File Picker
// rows, the same way rootsNav.renderRow already disambiguates two same-named
// roots by their left-truncated parent path
// (TestRootsNavDistinguishesSameNamedRootsByParent) — pickerRootLabel
// (picker_view.go) reuses that exact strategy instead of the bare
// filepath.Base(f.root.path) M3.1 shipped with.
//
// This test's own rowA/rowB comparison is over whole rendered terminal
// lines (plainLines(m.View().Content)), not pickerFileRow's output in
// isolation (M7.1b/TRIAGE T3) — it can only tell two Roots' rows apart,
// not prove pickerRootLabel's budget math correct in isolation from
// incidental surrounding frame content (nav pane, footer). That isolated
// proof is TestPickerFileRowFillsItsFullBudget below, which pins an
// independently hand-computed pickerFileRow string directly, confirmed by
// mutation to catch T13/D52's now-fixed bug (which this test's own
// wider-fixture history, T16/D52, shows it once missed).
func TestPickerFileResultsDisambiguateSameBasenameRoots(t *testing.T) {
	base := t.TempDir()
	rootA := mkdir(t, filepath.Join(base, "work", "admin", "project"))
	rootB := mkdir(t, filepath.Join(base, "other", "project"))
	fileA := filepath.Join(rootA, "report.md")
	fileB := filepath.Join(rootB, "report.md")

	// The project's standard fixture size (TRIAGE T16/D52): this test used
	// to run at a wider 140x30 that masked pickerRootLabel's double-counted
	// gap budget (T13/D52) by having budget to spare either way. At 80x24
	// the Results list's content width (23, computePickerLayout) is tight
	// enough that the bug actually mattered — both rows' parent segments
	// collapsed to a bare, identical "…" — so this size is what the fix
	// must be verified against, not the wider one the bug hid behind.
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

	lines := plainLines(m.View().Content)
	var rowA, rowB string
	for _, l := range lines {
		if !strings.Contains(l, "report.md") {
			continue
		}
		if rowA == "" {
			rowA = l
		} else if rowB == "" {
			rowB = l
		}
	}
	if rowA == "" || rowB == "" {
		t.Fatalf("expected two rows containing report.md, got:\n%s", strings.Join(lines, "\n"))
	}
	// At this width, the parent's disambiguating budget (3, see
	// TestPickerFileRowFillsItsFullBudget) is only enough for a short tail
	// like "…n/"/"…r/", not the full "admin"/"other" segment a wider
	// terminal (140x30, this test's own size before D52) would show — the
	// row-inequality check above is what this criterion (S3's
	// disambiguation, not full readability) actually requires.
	if rowA == rowB {
		t.Errorf("two Roots sharing a base name (%q) rendered identical rows, want their parent paths to distinguish them:\n%s", filepath.Base(rootA), strings.Join(lines, "\n"))
	}
}

// TestPickerFileRowFillsItsFullBudget pins pickerRootLabel/pickerFileRow's
// exact output at the project's standard 80x24 fixture (TRIAGE T13/T15,
// D52), independently of pickerRootLabel itself: the expected string below
// is hand-computed from computePickerLayout's own results content width
// (23) and leftTruncate's documented suffix-keeping behaviour, not derived
// by calling the function under test — unlike
// TestPickerFileResultsShowRootAndRelativePath, which builds its expected
// strings by calling pickerRootLabel and so cannot catch a bug in
// pickerRootLabel's own budget math (T15). Synthetic (non-tempdir) paths
// are used so every character of the expected string is knowable ahead of
// time: pickerFileRow never touches the filesystem, only filepath
// arithmetic. Before the T13 fix, pickerRootLabel's budget formula
// double-subtracted the name/parent gap, under-computing the parent's
// truncation budget by one gap width (3 vs. the correct 5) and rendering a
// bare "…" for both a root's label here instead of the disambiguating
// "…n/" this pins.
func TestPickerFileRowFillsItsFullBudget(t *testing.T) {
	f := pickerFile{
		root: root{path: "/home/user/work/admin/project"},
		path: "/home/user/work/admin/project/report.md",
	}
	const width = 23        // computePickerLayout(80, 24).results.content().width
	const contentWidth = 21 // width minus the 2-column star marker field (D72)

	wantLabel := "project  \u2026" // name(7) + gap(2) + leftTruncate(".../work/admin/", 1)
	if got := pickerRootLabel(f, contentWidth, "  report.md"); got != wantLabel {
		t.Fatalf("pickerRootLabel() = %q, want %q", got, wantLabel)
	}

	wantRow := "  " + wantLabel + "  report.md" // "  " is the (unstarred) marker
	// Empty query: this test pins the row's layout arithmetic, so it renders
	// the row with no match highlighting (M6.1) in play — the exact-string
	// comparison below is over plain text, and highlighting is pinned
	// separately by TestPickerFileRowHighlightsMatchedNameCharacters.
	got := pickerFileRow(f, "", width, false, false)
	if got != wantRow {
		t.Errorf("pickerFileRow() = %q, want %q", got, wantRow)
	}
	if w := lipgloss.Width(got); w != width {
		t.Errorf("pickerFileRow() rendered width = %d, want exactly %d (the Results content width) — the budget formula is leaving the parent segment short of what's actually available", w, width)
	}
}

// TestPickerFileRowShowsStarMarkerForAStarredResult closes the S3/D72 gap
// TRIAGE T8 found: S2/M2.5's own wording already talked about a Picker
// Result's "star marker" being cleared, but no Picker row actually rendered
// one. A starred Result's row now carries the same "\u2605 "/"  " two-column
// field Files Nav and Roots Nav rows already use, prepended ahead of the
// Root label rather than appended — the marker is styling-adjacent state,
// not part of the disambiguating path text pickerRootLabel computes.
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
			want:  []string{"File · Active Root"},
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
			want:         []string{"File · Active Root · Hidden"},
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
	bold, reverse bool
}

// cellStyles walks a rendered string's SGR escape sequences and reports the
// style in force at each visible cell — the only way to check *where* a
// highlight landed, which ansi.Strip (plainLines) deliberately discards and
// a string comparison against a hand-built expected string could only check
// by pinning lipgloss's exact escape serialization. Handles the subset
// lipgloss emits here: bold on/off (1/21/22), reverse on/off (7/27) and a
// full reset (0, or the empty parameter list lipgloss writes as `ESC[m`).
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
				for _, p := range strings.Split(params, ";") {
					switch p {
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

// TestPickerFileRowHighlightsMatchedNameCharacters is M6.1's own acceptance
// criterion (S5): the styled runs in a rendered File Picker row fall exactly
// on the fuzzy-matched positions of the file name, and nowhere else. The
// expected columns are hand-computed rather than derived from the highlighter
// — the fuzzy match of "rpt" against "report.md" can only be r(0), p(2),
// t(5) (those are the name's only r/p/t after the first, in order), offset by
// where the name starts in the row — so a highlighter that drifted onto the
// path, the Root label or the wrong characters of the name fails here.
// Synthetic (non-tempdir) paths keep every column knowable, the same reason
// TestPickerFileRowFillsItsFullBudget uses them.
func TestPickerFileRowHighlightsMatchedNameCharacters(t *testing.T) {
	f := pickerFile{
		root: root{path: "/home/user/notes"},
		path: "/home/user/notes/sub/report.md",
	}
	const width = 40
	const name = "report.md"

	got := pickerFileRow(f, "rpt", width, false, false)

	// The highlight is styling only: the row's displayed characters must be
	// exactly the unhighlighted row's.
	plain := ansi.Strip(got)
	if want := pickerFileRow(f, "", width, false, false); plain != want {
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
