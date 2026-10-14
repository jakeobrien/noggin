package main

import (
	"math"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"noggin/internal/tree"
)

// viewerText is a viewer's rendered content with ANSI stripped, one string per
// content row.
func viewerText(v viewer) []string {
	return strings.Split(ansi.Strip(v.vp.GetContent()), "\n")
}

func TestRenderTextGutterWrapAndTabs(t *testing.T) {
	src := []byte("one\n\ttabbed\n" + strings.Repeat("x", 30) + "\n")
	lines := strings.Split(ansi.Strip(renderText(src, 12, false)), "\n")
	// Three source lines, so a one-digit gutter plus a space: 3 text cells fewer
	// than the pane... wrapped width is 12-2 = 10.
	want := []string{
		"1 one",
		"2     tabbed", // tab expanded to column 4; 4+6 = 10 cells fit exactly
		"3 " + strings.Repeat("x", 10),
		"  " + strings.Repeat("x", 10), // continuation rows have a blank gutter
		"  " + strings.Repeat("x", 10),
	}
	if len(lines) != len(want) {
		t.Fatalf("rows = %q, want %q", lines, want)
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Errorf("row %d = %q, want %q", i, lines[i], want[i])
		}
		if w := ansi.StringWidth(lines[i]); w > 12 {
			t.Errorf("row %d is %d cells wide, pane is 12", i, w)
		}
	}
}

func TestRenderTextGutterWidthFollowsLineCount(t *testing.T) {
	src := []byte(strings.Repeat("a\n", 12))
	lines := strings.Split(ansi.Strip(renderText(src, 40, false)), "\n")
	if len(lines) != 12 {
		t.Fatalf("rows = %d, want 12 (a trailing newline is not another line)", len(lines))
	}
	if lines[0] != " 1 a" || lines[11] != "12 a" {
		t.Errorf("gutter not right-aligned to two digits: %q ... %q", lines[0], lines[11])
	}
}

func TestRenderTextIsRawAndUncoloured(t *testing.T) {
	src := []byte("---\ntitle: x\n---\n# Heading **bold** [l](http://u)\r\nbell\x1b[31m\n")
	out := renderText(src, 80, false)
	plain := strings.Split(ansi.Strip(out), "\n")
	for i, want := range []string{
		"1 ---", "2 title: x", "3 ---", "4 # Heading **bold** [l](http://u)",
	} {
		if plain[i] != want {
			t.Errorf("row %d = %q, want %q (raw source, frontmatter and CR handled)", i, plain[i], want)
		}
	}
	if strings.Contains(out, "\x1b[31m") || strings.ContainsRune(ansi.Strip(out), '\x1b') {
		t.Errorf("a control character in the source reached the terminal: %q", out)
	}
	// Body text carries no style of its own: the escapes are the gutter's.
	body := strings.SplitN(out, "\n", 2)[0]
	if strings.Contains(strings.TrimPrefix(body, body[:strings.Index(body, "---")]), "\x1b") {
		t.Errorf("body text is styled: %q", body)
	}
}

func TestRenderTextMutedDiffersFromFocused(t *testing.T) {
	src := []byte("hello\n")
	if renderText(src, 20, true) == renderText(src, 20, false) {
		t.Error("muted Text render is identical to the focused one")
	}
	if ansi.Strip(renderText(src, 20, true)) != ansi.Strip(renderText(src, 20, false)) {
		t.Error("muted and focused Text renders differ in layout")
	}
}

func TestRenderTextEmpty(t *testing.T) {
	if got := renderText(nil, 20, false); got != "" {
		t.Errorf("empty source = %q, want empty", got)
	}
}

// renderModeFixture opens a markdown file and a Text file in a Root, focus on
// DV1.
func renderModeFixture(t *testing.T) (m model, md, other, txt string) {
	t.Helper()
	dir := t.TempDir()
	md = writeFileContent(t, filepath.Join(dir, "a.md"), "# Title\n\nsome **bold** text\n")
	other = writeFileContent(t, filepath.Join(dir, "b.md"), "# Other\n\nsecond\n")
	txt = writeFileContent(t, filepath.Join(dir, "notes.txt"), "plain\nnotes\n")
	node, err := tree.Walk(dir, tree.FilterMarkdown)
	if err != nil {
		t.Fatal(err)
	}
	m = sized(t, newModel().openFolder(node), 80, 24)
	d, err := loadDocument(md)
	if err != nil {
		t.Fatal(err)
	}
	return m.openDocument(d), md, other, txt
}

func openPath(t *testing.T, m model, p pane, path string) model {
	t.Helper()
	d, err := loadDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	return m.openDocumentInto(p, d)
}

func TestRenderModeDefaultsToMarkdownAndToggles(t *testing.T) {
	m, _, _, _ := renderModeFixture(t)
	if m.dv1.mode != modeMarkdown {
		t.Fatalf("default mode = %v, want Markdown", m.dv1.mode)
	}
	if got := viewerText(m.dv1); strings.HasPrefix(strings.TrimSpace(got[0]), "1 ") || strings.Contains(strings.Join(got, "\n"), "1 # Title") {
		t.Fatalf("Markdown mode shows the Text rendering: %q", got)
	}

	m = press(t, m, keys.RenderMode.Keys()[0])
	if m.dv1.mode != modeText {
		t.Fatalf("mode after m = %v, want Text", m.dv1.mode)
	}
	if got := viewerText(m.dv1); got[0] != "1 # Title" || got[2] != "3 some **bold** text" {
		t.Errorf("Text mode rows = %q, want raw source with a gutter", got)
	}

	m = press(t, m, keys.RenderMode.Keys()[0])
	if m.dv1.mode != modeMarkdown {
		t.Errorf("mode after second m = %v, want Markdown", m.dv1.mode)
	}
	if got := strings.Join(viewerText(m.dv1), "\n"); strings.Contains(got, "1 # Title") {
		t.Errorf("toggling back left the Text rendering: %q", got)
	}
}

func TestRenderModeToggleFromEveryNonModalPaneTargetsDV1(t *testing.T) {
	for _, p := range []pane{rootsPane, filesPane, dv1Pane} {
		m, _, _, _ := renderModeFixture(t)
		m = m.setFocus(p)
		m = press(t, m, keys.RenderMode.Keys()[0])
		if m.dv1.mode != modeText {
			t.Errorf("focus %v: DV1 mode = %v, want Text", p, m.dv1.mode)
		}
	}
}

func TestRenderModeIsPerViewer(t *testing.T) {
	m, md, other, _ := renderModeFixture(t)
	m = openPath(t, m, dv2Pane, other)
	if m.dv2.mode != modeMarkdown {
		t.Fatalf("a new DV2 starts on %v, want Markdown", m.dv2.mode)
	}
	m = m.setFocus(dv2Pane)
	m = press(t, m, keys.RenderMode.Keys()[0])
	if m.dv2.mode != modeText || m.dv1.mode != modeMarkdown {
		t.Errorf("m with DV2 focused: dv1=%v dv2=%v, want Markdown/Text", m.dv1.mode, m.dv2.mode)
	}
	// DV1 stays on its own document, in its own mode.
	if m.dv1.doc.path != md {
		t.Errorf("DV1 document changed to %q", m.dv1.doc.path)
	}
	m = m.setFocus(dv1Pane)
	m = press(t, m, keys.RenderMode.Keys()[0])
	if m.dv1.mode != modeText || m.dv2.mode != modeText {
		t.Errorf("after m on DV1: dv1=%v dv2=%v", m.dv1.mode, m.dv2.mode)
	}

	// A DV2 that is closed and recreated starts on Markdown again.
	m = m.closeDV2()
	m = openPath(t, m, dv2Pane, other)
	if m.dv2.mode != modeMarkdown {
		t.Errorf("recreated DV2 mode = %v, want Markdown", m.dv2.mode)
	}
}

func TestRenderModeIsStickyAcrossOpenPeekAndHistory(t *testing.T) {
	m, md, other, _ := renderModeFixture(t)
	m = press(t, m, keys.RenderMode.Keys()[0])

	m = m.navigate(dv1Pane, mustLoad(t, other))
	if m.dv1.mode != modeText || viewerText(m.dv1)[0] != "1 # Other" {
		t.Errorf("a newly opened document lost Text mode: %q", viewerText(m.dv1))
	}

	m = m.historyBack()
	if m.dv1.doc.path != md || m.dv1.mode != modeText || viewerText(m.dv1)[0] != "1 # Title" {
		t.Errorf("history back lost Text mode: %q", viewerText(m.dv1))
	}
	m = m.historyForward()
	if m.dv1.doc.path != other || m.dv1.mode != modeText {
		t.Errorf("history forward lost Text mode")
	}

	m = m.startPeekFor(t, md)
	if got := ansi.Strip(m.dv1.peek.vp.GetContent()); !strings.HasPrefix(got, "1 # Title") {
		t.Errorf("Peek in Text mode = %q, want raw source", got)
	}
}

func mustLoad(t *testing.T, path string) document {
	t.Helper()
	d, err := loadDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// startPeekFor Peeks path in DV1 directly.
func (m model) startPeekFor(t *testing.T, path string) model {
	t.Helper()
	m.dv1 = m.dv1.startPeek(mustLoad(t, path))
	return m
}

func TestRenderModeToggleReRendersPeek(t *testing.T) {
	m, _, other, _ := renderModeFixture(t)
	m = m.startPeekFor(t, other)
	m = press(t, m, keys.RenderMode.Keys()[0])
	if got := ansi.Strip(m.dv1.peek.vp.GetContent()); !strings.HasPrefix(got, "1 # Other") {
		t.Errorf("Peek after m = %q, want Text rendering", got)
	}
}

func TestTextFileAlwaysRendersAsTextAndRefusesToggle(t *testing.T) {
	m, _, _, txt := renderModeFixture(t)
	m = m.navigate(dv1Pane, mustLoad(t, txt))
	if m.dv1.mode != modeMarkdown {
		t.Fatalf("opening a Text file changed the viewer mode to %v", m.dv1.mode)
	}
	if got := viewerText(m.dv1); got[0] != "1 plain" {
		t.Errorf("Text file rows = %q, want raw text with a gutter", got)
	}

	m = press(t, m, keys.RenderMode.Keys()[0])
	if m.dv1.mode != modeMarkdown {
		t.Errorf("m on a Text file changed the mode to %v", m.dv1.mode)
	}
	if !strings.Contains(m.message, "notes.txt") {
		t.Errorf("message = %q, want a refusal naming notes.txt", m.message)
	}
}

func TestStickyMarkdownModeSurvivesAVisitToATextFile(t *testing.T) {
	m, md, _, txt := renderModeFixture(t)
	m = m.navigate(dv1Pane, mustLoad(t, txt))
	m = m.navigate(dv1Pane, mustLoad(t, md))
	if got := strings.Join(viewerText(m.dv1), "\n"); strings.Contains(got, "1 # Title") {
		t.Errorf("Markdown file after a Text file shows the Text rendering: %q", got)
	}
}

func TestRenderModeOnAnArticleShowsConvertedMarkdown(t *testing.T) {
	m := sized(t, newModel().openDocument(fixtureArticle()), 80, 24)
	m = press(t, m, keys.RenderMode.Keys()[0])
	if m.dv1.mode != modeText {
		t.Fatalf("article refused the toggle: mode = %v, message %q", m.dv1.mode, m.message)
	}
	if got := viewerText(m.dv1)[0]; got != " 1 # Leonhard Euler" {
		t.Errorf("article first row = %q, want its converted markdown", got)
	}
}

func TestLocationLineEndsWithTheRenderMode(t *testing.T) {
	m, _, _, txt := renderModeFixture(t)
	if got := m.viewerLocationLine(dv1Pane); !strings.HasSuffix(got, " · Markdown") {
		t.Errorf("location line = %q, want it to end · Markdown", got)
	}
	m = press(t, m, keys.RenderMode.Keys()[0])
	if got := m.viewerLocationLine(dv1Pane); !strings.HasSuffix(got, " · Text") {
		t.Errorf("location line = %q, want it to end · Text", got)
	}
	// A Text file reads Text whatever the (Markdown-mode) viewer is set to.
	m = press(t, m, keys.RenderMode.Keys()[0])
	m = m.navigate(dv1Pane, mustLoad(t, txt))
	if got := m.viewerLocationLine(dv1Pane); !strings.HasSuffix(got, "notes.txt · Text") {
		t.Errorf("location line for a Text file = %q", got)
	}
	// An article keeps its attribution, then the mode.
	a := sized(t, newModel().openDocument(fixtureArticle()), 80, 24)
	if got := a.viewerLocationLine(dv1Pane); !strings.HasSuffix(got, wikipediaAttributionSuffix+" · Markdown") {
		t.Errorf("article location line = %q", got)
	}
	// The mode survives truncation on a narrow pane: the line is drawn with
	// its tail intact.
	if row := ansi.Strip(a.composeViewerContent(dv1Pane, 30, "x")); !strings.HasSuffix(row, "· Markdown") {
		t.Errorf("narrow location row = %q, want the mode kept", row)
	}
}

func TestMutedViewerShowsMutedText(t *testing.T) {
	m, _, _, _ := renderModeFixture(t)
	m = press(t, m, keys.RenderMode.Keys()[0])
	v := m.dv1
	focused := v.view()
	v.muted = true
	muted := v.view()
	if focused == muted {
		t.Error("an unfocused viewer in Text mode is not muted")
	}
	if ansi.Strip(focused) != ansi.Strip(muted) {
		t.Error("muting changed the Text layout")
	}
}

func TestRenderModeBindingIsDocumented(t *testing.T) {
	for _, p := range []pane{rootsPane, filesPane, dv1Pane, dv2Pane} {
		found := false
		for _, b := range helpBindings(p) {
			if b.Help().Key == keys.RenderMode.Help().Key {
				found = true
			}
		}
		if !found {
			t.Errorf("Help modal for pane %v does not list the render mode key", p)
		}
	}
}

// peekToggleFixture is a DV1 and a DV2, each holding a long Active Document
// with a committed find ("item 5") and a long Peek over it, scrolled
// part-way; focus is on p.
func peekToggleFixture(t *testing.T, p pane) model {
	t.Helper()
	dir := t.TempDir()
	active := writeFileContent(t, filepath.Join(dir, "active.md"), listContent(60))
	peeked := writeFileContent(t, filepath.Join(dir, "peeked.md"), listContent(90))
	m := sized(t, newModel(), 80, 14)
	m = openPath(t, m, dv1Pane, active)
	m = openPath(t, m, dv2Pane, active)
	m = m.setFocus(p)
	m = press(t, m, "/")
	m = typeText(t, m, "item 5")
	m = press(t, m, "enter")
	m = press(t, m, keys.FindNext.Keys()[0])
	if len(m.viewerFor(p).findState.matches) < 3 {
		t.Fatalf("setup: too few matches")
	}
	v := m.viewerFor(p).startPeek(mustLoad(t, peeked))
	v.peek.vp.SetYOffset(v.peek.vp.TotalLineCount() / 3)
	if v.peek.vp.YOffset() == 0 {
		t.Fatal("setup: Peek did not scroll")
	}
	return m.withViewerFor(p, v)
}

func TestRenderModeToggleDuringPeekKeepsPeekScrollProportion(t *testing.T) {
	for _, p := range []pane{dv1Pane, dv2Pane} {
		m := peekToggleFixture(t, p)
		before := scrollFractionOf(m.viewerFor(p).peek.vp)
		m = press(t, m, keys.RenderMode.Keys()[0])
		v := m.viewerFor(p)
		if v.mode != modeText || !v.peeking() {
			t.Fatalf("pane %v: mode = %v peeking = %v, want Text and still Peeking", p, v.mode, v.peeking())
		}
		if got := ansi.Strip(v.peek.vp.GetContent()); !strings.HasPrefix(got, " 1 - item 1") {
			t.Errorf("pane %v: Peek not re-rendered as Text: %.30q", p, got)
		}
		got := scrollFractionOf(v.peek.vp)
		if got == 0 || math.Abs(got-before) > 0.05 {
			t.Errorf("pane %v: Peek scroll fraction = %.2f after toggle, want about %.2f", p, got, before)
		}
	}
}

func TestRenderModeToggleDuringPeekCarriesActiveDocumentUnderneath(t *testing.T) {
	for _, p := range []pane{dv1Pane, dv2Pane} {
		m := peekToggleFixture(t, p)
		before := m.viewerFor(p)
		m = press(t, m, keys.RenderMode.Keys()[0])
		after := m.viewerFor(p)
		if after.findState.matches[after.findState.current].line < 0 {
			t.Fatalf("pane %v: no current match line", p)
		}
		if got, want := after.vp.YOffset(), after.findState.matches[after.findState.current].line; got > want || want >= got+after.vp.Height() {
			t.Errorf("pane %v: Active Document offset %d does not show the current match (row %d)", p, got, want)
		}
		if got := after.findState.input.Value(); got != "item 5" {
			t.Errorf("pane %v: Active Document find query = %q, want kept", p, got)
		}
		if after.findState.current != before.findState.current {
			t.Errorf("pane %v: current match = %d, want %d", p, after.findState.current, before.findState.current)
		}
		m = m.withViewerFor(p, m.viewerFor(p).clearPeek())
		v := m.viewerFor(p)
		if v.docMode != modeText || v.findState.input.Value() != "item 5" {
			t.Errorf("pane %v: after ending the Peek mode = %v query = %q", p, v.docMode, v.findState.input.Value())
		}
	}
}

func TestRenderModeToggleDuringPeekKeepsActiveScrollProportion(t *testing.T) {
	for _, p := range []pane{dv1Pane, dv2Pane} {
		m := peekToggleFixture(t, p)
		v := m.viewerFor(p)
		v.findState = newFind() // no find: scroll is carried as a proportion
		v = v.withScrollFraction(0.5)
		m = m.withViewerFor(p, v)
		m = press(t, m, keys.RenderMode.Keys()[0])
		if got := scrollFractionOf(m.viewerFor(p).vp); math.Abs(got-0.5) > 0.05 {
			t.Errorf("pane %v: Active Document scroll fraction = %.2f, want about 0.5", p, got)
		}
	}
}

func TestRenderModeRefusedOnPeekedTextFileOverMarkdown(t *testing.T) {
	m, _, _, txt := renderModeFixture(t)
	m = m.startPeekFor(t, txt)
	m = press(t, m, keys.RenderMode.Keys()[0])
	if m.dv1.mode != modeMarkdown || !strings.Contains(m.message, "notes.txt") {
		t.Errorf("mode = %v message = %q, want a refusal naming notes.txt", m.dv1.mode, m.message)
	}
}
