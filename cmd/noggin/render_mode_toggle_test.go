package main

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"noggin/internal/tree"
)

// toggleDoc is a Markdown file with a "needle" on three separated lines, so a
// match is far enough from the others to be off-screen when another is current.
func toggleDoc() string {
	var b strings.Builder
	b.WriteString("# Title\n\nneedle one\n\n")
	for _, n := range []string{"two", "three"} {
		for i := 0; i < 40; i++ {
			fmt.Fprintf(&b, "filler %s %d\n\n", n, i)
		}
		fmt.Fprintf(&b, "needle %s\n\n", n)
	}
	return b.String()
}

// toggleModel is a 80x12 model with toggleDoc open in DV1, focus on it.
func toggleModel(t *testing.T, content string) model {
	t.Helper()
	dir := t.TempDir()
	path := writeFileContent(t, filepath.Join(dir, "doc.md"), content)
	node, err := tree.Walk(dir)
	if err != nil {
		t.Fatal(err)
	}
	m := sized(t, newModel().openFolder(node), 80, 12)
	return openPath(t, m, dv1Pane, path).setFocus(dv1Pane)
}

// currentRowText is the rendered text of the row the current match is on.
func currentRowText(t *testing.T, v viewer) string {
	t.Helper()
	f := v.findState
	if f.current < 0 || f.current >= len(f.matches) {
		t.Fatalf("no current match (%d of %d)", f.current, len(f.matches))
	}
	return ansi.Strip(strings.Split(v.vp.GetContent(), "\n")[f.matches[f.current].line])
}

// inView reports whether the current match's first row is on screen.
func inView(v viewer) bool {
	line := v.findState.matches[v.findState.current].line
	return line >= v.offset() && line < v.offset()+v.height()
}

func committedFind(t *testing.T, m model, query string, nexts int) model {
	t.Helper()
	m = press(t, m, keys.Find.Keys()[0])
	m = typeText(t, m, query)
	m = press(t, m, keys.Activate.Keys()[0])
	for ; nexts > 0; nexts-- {
		m = press(t, m, keys.FindNext.Keys()[0])
	}
	return m
}

func TestToggleKeepsCommittedFindAndMatchOrdinalMarkdownToText(t *testing.T) {
	m := committedFind(t, toggleModel(t, toggleDoc()), "needle", 1)
	if got := currentRowText(t, m.dv1); !strings.Contains(got, "needle two") {
		t.Fatalf("setup: current match row = %q", got)
	}

	m = press(t, m, keys.RenderMode.Keys()[0])
	v := m.dv1
	if v.text == nil || v.findState.input.Value() != "needle" || v.findState.editing {
		t.Fatalf("Text mode %v, query %q, editing %v: want Text, the kept committed query", v.text != nil, v.findState.input.Value(), v.findState.editing)
	}
	if len(v.findState.matches) != 3 || v.findState.current != 1 {
		t.Fatalf("matches %d, current %d, want 3 and 1", len(v.findState.matches), v.findState.current)
	}
	if got := currentRowText(t, v); !strings.HasSuffix(got, "needle two") {
		t.Errorf("current match row = %q, want the second needle", got)
	}
	if !inView(v) {
		t.Errorf("current match not in view (offset %d)", v.offset())
	}
	if rev := reversedPerRow(v); rev[v.findState.matches[1].line] != "needle" {
		t.Errorf("current match not highlighted in Text: %q", rev[v.findState.matches[1].line])
	}

	// n keeps working from the kept match.
	m = press(t, m, keys.FindNext.Keys()[0])
	if got := currentRowText(t, m.dv1); !strings.HasSuffix(got, "needle three") {
		t.Errorf("n after toggling landed on %q, want the third needle", got)
	}
}

func TestToggleKeepsCommittedFindAndMatchOrdinalTextToMarkdown(t *testing.T) {
	m := toggleModel(t, toggleDoc())
	m = press(t, m, keys.RenderMode.Keys()[0]) // to Text
	m = committedFind(t, m, "needle", 2)       // third needle
	if got := currentRowText(t, m.dv1); !strings.HasSuffix(got, "needle three") {
		t.Fatalf("setup: current match row = %q", got)
	}

	m = press(t, m, keys.RenderMode.Keys()[0])
	v := m.dv1
	if v.text != nil || v.findState.input.Value() != "needle" {
		t.Fatalf("Text layout %v, query %q: want Markdown and the kept query", v.text != nil, v.findState.input.Value())
	}
	if len(v.findState.matches) != 3 || v.findState.current != 2 {
		t.Fatalf("matches %d, current %d, want 3 and 2", len(v.findState.matches), v.findState.current)
	}
	if got := currentRowText(t, v); !strings.Contains(got, "needle three") {
		t.Errorf("current match row = %q, want the third needle", got)
	}
	if !inView(v) {
		t.Errorf("current match not in view (offset %d)", v.offset())
	}
}

// Occurrences Glamour drops (here an HTML comment) shift the ordinal mapping, as in
// D34: the Nth occurrence in one mode is the Nth in the other, so a match that
// has no counterpart degrades to the first match rather than failing.
func TestToggleWithAnUnmappableMatchDegradesToTheFirstMatch(t *testing.T) {
	m := toggleModel(t, "<!-- needle -->\n\nbody needle\n")
	m = press(t, m, keys.RenderMode.Keys()[0]) // to Text: two matches
	m = committedFind(t, m, "needle", 1)       // the second
	if v := m.dv1; len(v.findState.matches) != 2 || v.findState.current != 1 {
		t.Fatalf("setup: matches %d, current %d", len(v.findState.matches), v.findState.current)
	}
	m = press(t, m, keys.RenderMode.Keys()[0]) // Markdown shows one
	v := m.dv1
	if len(v.findState.matches) != 1 || v.findState.current != 0 {
		t.Errorf("matches %d, current %d, want 1 and 0", len(v.findState.matches), v.findState.current)
	}
}

func TestToggleKeepsAFindInProgress(t *testing.T) {
	v := newViewer().setSize(80, 12)
	v = v.open(mustLoad(t, writeFileContent(t, filepath.Join(t.TempDir(), "d.md"), toggleDoc())))
	v, _ = v.startFind()
	for _, r := range "needle" {
		v, _ = v.updateFindQuery(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	v = v.toggleMode()
	f := v.findState
	if !f.editing || f.input.Value() != "needle" || len(f.matches) != 3 || f.current != 0 {
		t.Errorf("editing %v, query %q, %d matches, current %d: want the query still being edited over 3 matches", f.editing, f.input.Value(), len(f.matches), f.current)
	}
	if v.text == nil {
		t.Error("not in Text mode")
	}
}

// A Grep landing that fell back to the first match in Markdown lands on
// ripgrep's exact line and column when switched to Text.
func TestToggleAfterAFallenBackGrepLandingUsesRipgrepsLocation(t *testing.T) {
	requireRipgrep(t)
	dir := t.TempDir()
	var b strings.Builder
	b.WriteString("<!-- needle -->\n\n") // dropped from the rendered text
	for i := 0; i < 30; i++ {
		b.WriteString("filler\n\n")
	}
	b.WriteString("first needle\n\n")
	for i := 0; i < 30; i++ {
		b.WriteString("filler\n\n")
	}
	b.WriteString("\tthird needle\n") // raw ordinal 2: past Markdown's 2 matches
	writeFileContent(t, filepath.Join(dir, "g.md"), b.String())

	m := openedFixture(t, 80, 12).withRipgrep(detectRipgrep())
	m.roots = m.roots.setRoots([]root{{path: dir}}, 0)
	m, _ = m.openPicker()
	m.picker.mode = grepMode
	m, seqs := typePickerQuery(t, m, "needle")
	m, _ = deliverPickerTicks(t, m, seqs)
	if len(m.picker.results.grep) != 3 {
		t.Fatalf("setup: %d grep results, want 3", len(m.picker.results.grep))
	}
	for i, g := range m.picker.results.grep {
		if strings.HasSuffix(g.text, "third needle") {
			m.picker.selected = i
		}
	}
	m = press(t, m, keys.Activate.Keys()[0])
	v := m.dv1
	if v.text != nil || len(v.findState.matches) != 2 || v.findState.current != 0 {
		t.Fatalf("setup: Markdown landing should have fallen back to the first of 2 matches, got current %d of %d", v.findState.current, len(v.findState.matches))
	}

	m = press(t, m, keys.RenderMode.Keys()[0])
	v = m.dv1
	f := v.findState
	if v.text == nil || len(f.matches) != 3 || f.current != 2 {
		t.Fatalf("Text mode %v, current %d of %d, want the third of 3 (ripgrep's line)", v.text != nil, f.current, len(f.matches))
	}
	// Raw column 1 after the tab is display column 4, then "third ".
	if c := f.matches[f.current]; c.colStart != 4+len("third ") || !inView(v) {
		t.Errorf("landed at %+v (in view %v), want display byte %d", c, inView(v), 4+len("third "))
	}
	if got := currentRowText(t, v); !strings.HasSuffix(got, "third needle") {
		t.Errorf("current row = %q", got)
	}

	// Moving on leaves the landing behind: back in Markdown the ordinal rules.
	m = press(t, m, keys.FindPrevious.Keys()[0])
	m = press(t, m, keys.RenderMode.Keys()[0])
	if f := m.dv1.findState; f.current != 1 {
		t.Errorf("after N and back to Markdown, current = %d, want 1", f.current)
	}
}

// With no find, the toggle keeps the same proportional scroll position.
func TestToggleWithoutFindKeepsProportionalScroll(t *testing.T) {
	var b strings.Builder
	for i := 0; i < 100; i++ {
		fmt.Fprintf(&b, "paragraph %d\n\n\n\n", i) // far fewer rows rendered than in source
	}
	m := toggleModel(t, b.String())
	frac := func(v viewer) float64 {
		return float64(v.offset()) / float64(v.lineCount()-v.height())
	}
	for _, mode := range []string{"Markdown to Text", "Text to Markdown"} {
		for _, at := range []float64{0, 0.5, 0.9, 1} {
			v := m.dv1
			v = v.withOffset(int(math.Round(at * float64(v.lineCount()-v.height()))))
			m.dv1 = v
			before := frac(v)
			got := press(t, m, keys.RenderMode.Keys()[0])
			after := frac(got.dv1)
			if tol := 1 / float64(got.dv1.lineCount()-got.dv1.height()); math.Abs(after-before) > tol {
				t.Errorf("%s at %.2f: scroll fraction %.3f -> %.3f", mode, at, before, after)
			}
			if got.dv1.findState.input.Value() != "" {
				t.Errorf("%s: a find appeared", mode)
			}
		}
		m = press(t, m, keys.RenderMode.Keys()[0])
	}
}

// The Frontmatter Card drops the comment's occurrence, so the body is counted
// apart from the frontmatter (D138) and the body's matches still line up.
func TestToggleCountsFrontmatterApartFromTheBody(t *testing.T) {
	m := toggleModel(t, "---\ntitle: 'needle' # needle\n---\n\nbody needle one\n\nbody needle two\n")
	m = committedFind(t, m, "needle", 2) // Markdown: card, one, two
	if v := m.dv1; len(v.findState.matches) != 3 || v.findState.current != 2 {
		t.Fatalf("setup: matches %d, current %d", len(v.findState.matches), v.findState.current)
	}
	m = press(t, m, keys.RenderMode.Keys()[0])
	if got := currentRowText(t, m.dv1); !strings.HasSuffix(got, "body needle two") || m.dv1.findState.current != 3 {
		t.Errorf("Text: current %d on %q, want 3 on the second body needle", m.dv1.findState.current, got)
	}
	m = press(t, m, keys.RenderMode.Keys()[0])
	if got := currentRowText(t, m.dv1); !strings.Contains(got, "body needle two") || m.dv1.findState.current != 2 {
		t.Errorf("Markdown: current %d on %q, want 2 on the second body needle", m.dv1.findState.current, got)
	}
}
