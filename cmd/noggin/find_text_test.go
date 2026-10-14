package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// textViewer is a viewer of the given size showing src in Text mode.
func textViewer(t *testing.T, src string, width, height int) viewer {
	t.Helper()
	path := writeFileContent(t, filepath.Join(t.TempDir(), "doc.md"), src)
	v := newViewer().setSize(width, height)
	v.mode = modeText
	return v.open(mustLoad(t, path))
}

// reversedPerRow is, for each content row of v, the text drawn in reverse
// video (the indicated match span) with the gutter excluded by construction.
func reversedPerRow(v viewer) []string {
	var out []string
	for _, row := range strings.Split(v.vp.GetContent(), "\n") {
		runes := []rune(ansi.Strip(row))
		var b strings.Builder
		for i, c := range cellStyles(row) {
			if c.reverse && i < len(runes) {
				b.WriteRune(runes[i])
			}
		}
		out = append(out, b.String())
	}
	return out
}

func TestTextFindMatchSpanningASoftWrap(t *testing.T) {
	// Pane 20 wide, gutter "1 ": 18 cells per row. The query straddles the
	// break between "ddd" and "d eeee".
	v := textViewer(t, "aaaa bbbb cccc dddd eeee ffff\n", 20, 10)
	if got := viewerText(v); len(got) != 2 || got[0] != "1 aaaa bbbb cccc ddd" || got[1] != "  d eeee ffff" {
		t.Fatalf("setup: rows = %q", got)
	}

	v = v.commitQuery("dddd eeee")
	if len(v.findState.matches) != 1 {
		t.Fatalf("matches = %d, want 1 for a query spanning the wrap", len(v.findState.matches))
	}
	if m := v.findState.matches[0]; m.line != 0 || m.src != 0 || m.colStart != 15 || m.colEnd != 24 {
		t.Errorf("match = %+v, want starting on row 0, source line 0, bytes 15..24", m)
	}
	rev := reversedPerRow(v)
	if rev[0] != "ddd" || rev[1] != "d eeee" {
		t.Errorf("reversed per row = %q, want the span split across both rows", rev)
	}
}

func TestTextFindHighlightsOnlyTheMatchedSpan(t *testing.T) {
	v := textViewer(t, "the quick brown fox\nanother fox here\n", 40, 10)
	v = v.commitQuery("fox")
	if len(v.findState.matches) != 2 {
		t.Fatalf("matches = %d, want 2", len(v.findState.matches))
	}
	rev := reversedPerRow(v)
	if rev[0] != "fox" || rev[1] != "" {
		t.Errorf("reversed = %q, want only the current match's span %q", rev, "fox")
	}
	if v.vp.StyleLineFunc != nil {
		t.Error("Text find styles the whole line (StyleLineFunc set)")
	}
	// The line's own text is untouched.
	if got := viewerText(v)[0]; got != "1 the quick brown fox" {
		t.Errorf("row = %q", got)
	}

	v = v.findNext()
	if rev := reversedPerRow(v); rev[0] != "" || rev[1] != "fox" {
		t.Errorf("after n reversed = %q, want the span on the second line only", rev)
	}
	v = v.cancelFind()
	for i, r := range reversedPerRow(v) {
		if r != "" {
			t.Errorf("row %d still highlighted after cancel: %q", i, r)
		}
	}
}

func TestTextFindNextPreviousWrapAround(t *testing.T) {
	v := textViewer(t, "x one\ny\nx two\nz\nx three\n", 40, 10).commitQuery("x")
	if n := len(v.findState.matches); n != 3 {
		t.Fatalf("matches = %d, want 3", n)
	}
	if v.findState.current != 0 {
		t.Fatalf("current = %d, want 0", v.findState.current)
	}
	for _, want := range []int{1, 2, 0} {
		v = v.findNext()
		if v.findState.current != want {
			t.Fatalf("after n: current = %d, want %d", v.findState.current, want)
		}
	}
	v = v.findPrevious()
	if v.findState.current != 2 {
		t.Errorf("N from the first match: current = %d, want wrap to 2", v.findState.current)
	}
	if rev := reversedPerRow(v); rev[4] != "x" {
		t.Errorf("reversed = %q, want the third match indicated", rev)
	}
}

func TestTextFindIsSmartCase(t *testing.T) {
	v := textViewer(t, "Fox\nfox\nFOX\n", 40, 10)
	if n := len(v.commitQuery("fox").findState.matches); n != 3 {
		t.Errorf(`"fox" matches %d, want 3 (lower case is insensitive)`, n)
	}
	if n := len(v.commitQuery("Fox").findState.matches); n != 1 {
		t.Errorf(`"Fox" matches %d, want 1 (an upper-case letter is sensitive)`, n)
	}
}

// Text shows what Glamour drops: link destinations and emphasis markers.
func TestTextFindSeesLinkURLsAndMarkdownSyntax(t *testing.T) {
	src := "see [the docs](https://example.com/guide) and **bold** words\n"
	text := textViewer(t, src, 80, 10)
	for _, q := range []string{"example.com/guide", "**bold**"} {
		if n := len(text.commitQuery(q).findState.matches); n != 1 {
			t.Errorf("Text find %q: %d matches, want 1", q, n)
		}
	}
	rev := reversedPerRow(text.commitQuery("example.com/guide"))
	if rev[0] != "example.com/guide" {
		t.Errorf("reversed = %q, want just the URL part", rev)
	}

	// Markdown mode is unchanged: found in the rendered content, the whole line
	// indicated.
	md := text
	md.mode = modeMarkdown
	md = md.render().commitQuery("bold")
	if md.text != nil || len(md.findState.matches) != 1 || md.findState.matches[0].inText || md.vp.StyleLineFunc == nil {
		t.Errorf("Markdown find changed: layout %v, matches %+v", md.text != nil, md.findState.matches)
	}
}

func TestTextFindScrollsToTheRowOfAWrappedMatch(t *testing.T) {
	var b strings.Builder
	for i := 0; i < 40; i++ {
		b.WriteString("filler\n")
	}
	b.WriteString(strings.Repeat("word ", 10) + "needle\n") // wraps in a 20-wide pane
	v := textViewer(t, b.String(), 20, 6).commitQuery("needle")
	m := v.findState.matches[0]
	if m.src != 40 || m.line <= 40 {
		t.Fatalf("match = %+v, want source line 40 on a continuation row after row 40", m)
	}
	if off := v.offset(); m.line < off || m.line >= off+v.height() {
		t.Errorf("match row %d not in view [%d,%d)", m.line, off, off+v.height())
	}
	rev := reversedPerRow(v)
	if got := strings.Join(rev[m.line:], ""); got != "needle" || rev[m.line] == "needle" {
		t.Errorf("reversed from row %d = %q, want %q split across the wrap", m.line, rev[m.line:], "needle")
	}
}

func TestTextFindHighlightFollowsTheMutedContent(t *testing.T) {
	v := textViewer(t, "alpha beta\n", 40, 5).commitQuery("beta")
	v.muted = true
	view := v.view()
	var reversed strings.Builder
	runes := []rune(ansi.Strip(view))
	for i, c := range cellStyles(view) {
		if c.reverse && i < len(runes) {
			reversed.WriteRune(runes[i])
		}
	}
	if !strings.Contains(reversed.String(), "beta") || strings.Contains(reversed.String(), "alpha") {
		t.Errorf("muted view reverses %q, want just the span", reversed.String())
	}
}

func TestTextGrepLanding(t *testing.T) {
	// match is a Text find match on source line src, starting at display byte col.
	match := func(src, col int) findMatch { return findMatch{src: src, colStart: col, inText: true} }
	for _, tc := range []struct {
		name    string
		src     string
		matches []findMatch
		line    int // rg's 1-based line
		col     int // rg's byte column in the raw line
		want    int
		ok      bool
	}{
		// "\tx y x" displays as "    x y x": x at display bytes 4 and 8.
		{"tab expands to a stop", "a\n\tx y x\n", []findMatch{match(0, 0), match(1, 4), match(1, 8)}, 2, 1, 1, true},
		{"second occurrence is exact, not first-on-line", "a\n\tx y x\n", []findMatch{match(0, 0), match(1, 4), match(1, 8)}, 2, 5, 2, true},
		{"no match at the column falls back to the line's first", "\tx y x\n", []findMatch{match(0, 4), match(0, 8)}, 1, 3, 0, true},
		{"no match on the line", "a\nb\nc\n", []findMatch{match(0, 0)}, 3, 0, 0, false},
		{"BOM, tab and carriage return", "\ufeff\tx\r\n", []findMatch{match(0, 1)}, 1, 4, 0, true},
		{"BOM counts toward the tab stop", "\ufeff\tx\ufeff\r\n", []findMatch{match(0, 0), match(0, 1)}, 1, 4, 1, true},
		{"a control character is a 3-byte replacement", "\x01x\n", []findMatch{match(0, 0), match(0, 3)}, 1, 1, 1, true},
		{"column past the end", "ab\n", []findMatch{match(0, 2)}, 1, 99, 0, true},
	} {
		got, ok := textGrepLanding([]byte(tc.src), tc.matches, tc.line, tc.col)
		if got != tc.want || ok != tc.ok {
			t.Errorf("%s: textGrepLanding = (%d, %v), want (%d, %v)", tc.name, got, ok, tc.want, tc.ok)
		}
	}
}

// A match that wraps is brought fully into view when it fits, not only its
// first row.
func TestTextFindShowsTheWholeWrappedMatch(t *testing.T) {
	var b strings.Builder
	for i := 0; i < 30; i++ {
		b.WriteString("filler\n")
	}
	b.WriteString(strings.Repeat("a", 15) + "needle" + strings.Repeat("b", 10) + "\n")
	v := textViewer(t, b.String(), 21, 5) // gutter "31 ": 18 cells a row; the match is on rows 30-31
	// Park the match's first row on the last visible row.
	v = v.withOffset(30 - (v.height() - 1))
	v = v.commitQuery("needle")
	m := v.findState.matches[0]
	if m.endLine <= m.line {
		t.Fatalf("setup: match %+v does not wrap", m)
	}
	if off := v.offset(); m.line < off || m.endLine >= off+v.height() {
		t.Errorf("match rows %d-%d not both in view [%d,%d)", m.line, m.endLine, off, off+v.height())
	}
}

// ripgrep reports the byte column of a line's first match.
func TestGrepResultsCarryRipgrepsColumn(t *testing.T) {
	requireRipgrep(t)
	path := writeFileContent(t, filepath.Join(t.TempDir(), "a.md"), "nothing\n\tx needle\n")
	got, _ := grepPickerFiles([]pickerFile{{path: path}}, "needle")
	if len(got) != 1 || got[0].line != 2 || got[0].col != 3 {
		t.Errorf("results = %+v, want one on line 2 at byte column 3", got)
	}
}

// Activating a Grep Result into a Text-mode viewer lands on ripgrep's own
// line and column, with that match the current one of a committed Find.
func TestPickerActivateGrepResultInTextModeLandsOnExactMatch(t *testing.T) {
	requireRipgrep(t)
	dir := t.TempDir()
	var b strings.Builder
	for i := 1; i <= 60; i++ {
		switch i {
		case 20:
			b.WriteString("a report line\n")
		case 45:
			// Tab, then the query twice; wraps in the 40-wide pane.
			b.WriteString("\tlead report and then more words, a second report\n")
		default:
			b.WriteString("filler\n")
		}
	}
	path := writeFileContent(t, filepath.Join(dir, "long.md"), b.String())

	m := openedFixture(t, 80, 24).withRipgrep(detectRipgrep())
	m = press(t, m, keys.RenderMode.Keys()[0])
	m.roots = m.roots.setRoots([]root{{path: dir}}, 0)
	m, _ = m.openPicker()
	m.picker.mode = grepMode
	m, seqs := typePickerQuery(t, m, "report")
	m, _ = deliverPickerTicks(t, m, seqs)
	if len(m.picker.results.grep) != 2 {
		t.Fatalf("setup: %d grep results, want 2", len(m.picker.results.grep))
	}
	for i, g := range m.picker.results.grep {
		if g.line == 45 {
			m.picker.selected = i
		}
	}

	got := press(t, m, keys.Activate.Keys()[0])
	if got.dv1.doc.path != path || got.dv1.mode != modeText || got.dv1.text == nil {
		t.Fatalf("setup: document %q in mode %v", got.dv1.doc.path, got.dv1.mode)
	}
	f := got.dv1.findState
	if len(f.matches) != 3 || f.current < 0 {
		t.Fatalf("committed find: %d matches, current %d, want 3 matches", len(f.matches), f.current)
	}
	cur := f.matches[f.current]
	// Raw column 1 (after the tab) is display column 4 on source line 44.
	if cur.src != 44 || cur.colStart != 4+len("lead ") {
		t.Errorf("landed on %+v, want source line 44 at its first occurrence (display byte %d)", cur, 4+len("lead "))
	}
	if off := got.dv1.offset(); cur.line < off || cur.line >= off+got.dv1.height() {
		t.Errorf("landing row %d not in view [%d,%d)", cur.line, off, off+got.dv1.height())
	}
	if rev := reversedPerRow(got.dv1); rev[cur.line] != "report" {
		t.Errorf("reversed on the landing row = %q, want %q", rev[cur.line], "report")
	}

	// n walks on to the second occurrence on the same line.
	next := press(t, got, keys.FindNext.Keys()[0])
	if c := next.dv1.findState.matches[next.dv1.findState.current]; c.src != 44 || c.colStart <= cur.colStart {
		t.Errorf("n landed on %+v, want the second occurrence on line 44", c)
	}
}

// Find maps a byte offset in a source line onto its wrapped rows, which only
// works if the rows concatenate back to exactly the line.
func TestTextLayoutRowsReassembleTheLine(t *testing.T) {
	src := "plain words   with  runs   of spaces at the wrap point \n" +
		"日本語のテキストは幅が二つ分ありますので折り返しに注意\n" +
		"emoji 👩‍👩‍👧 and e\u0301 combining marks go here, wrapped\n" +
		"          leading spaces and a very long unbroken token: " + strings.Repeat("z", 50) + "\n"
	for width := 4; width <= 40; width += 3 {
		l := layoutText([]byte(src), width)
		for i, line := range textLines([]byte(src)) {
			if got := strings.Join(l.rows[i], ""); got != line {
				t.Fatalf("width %d, line %d: rows %q reassemble to %q, want %q", width, i, l.rows[i], got, line)
			}
		}
	}
}
