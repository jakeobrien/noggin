package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"noggin/internal/tree"
)

// keepTheme restores the active Theme when the test ends: `theme` is
// process-wide, so a test that switches it must not leak the switch.
func keepTheme(t *testing.T) {
	t.Helper()
	was := theme
	t.Cleanup(func() { theme = was })
}

func TestThemesHaveDistinctNames(t *testing.T) {
	seen := map[string]bool{}
	for _, th := range themes {
		if th.Name == "" || seen[th.Name] {
			t.Errorf("theme name %q is empty or repeated", th.Name)
		}
		seen[th.Name] = true
	}
	if len(themes) < 2 {
		t.Errorf("%d themes; there is nothing to cycle", len(themes))
	}
}

func TestNextThemeCyclesEveryThemeAndWraps(t *testing.T) {
	name := themes[0].Name
	for i := 1; i <= len(themes); i++ {
		name = nextTheme(name).Name
		if want := themes[i%len(themes)].Name; name != want {
			t.Fatalf("step %d: next theme = %q, want %q", i, name, want)
		}
	}
	if got := nextTheme("no-such-theme").Name; got != themes[0].Name {
		t.Errorf("nextTheme of an unknown name = %q, want the first theme", got)
	}
}

// TestThemeKeyRepaintsTheWholeFrame: ctrl+t makes the next Theme active, names
// it in the Footer, and the very next frame is painted in it — no colour of the
// old Theme survives anywhere, which is what catches a style captured in a
// package var at startup.
func TestThemeKeyRepaintsTheWholeFrame(t *testing.T) {
	keepTheme(t)
	t.Setenv("NO_COLOR", "")
	old := themes[0]
	theme = old
	m := sized(t, newModel().openFolder(fixtureTree()), 80, 24)

	m = press(t, m, keys.Theme.Keys()[0])
	next := themes[1]
	if theme.Name != next.Name {
		t.Fatalf("active theme = %q, want %q", theme.Name, next.Name)
	}
	if want := "theme: " + next.Name; m.message != want {
		t.Errorf("footer message = %q, want %q", m.message, want)
	}
	v := m.View()
	if hexOf(v.BackgroundColor) != hexOf(next.Bg) || hexOf(v.ForegroundColor) != hexOf(next.Fg) {
		t.Errorf("terminal colours = %s on %s, want the new Theme's", hexOf(v.ForegroundColor), hexOf(v.BackgroundColor))
	}
	if !strings.Contains(v.Content, sgrColor(next.Accent)) {
		t.Error("frame has no cell in the new Theme's Accent (the focused border)")
	}
	for field, c := range colorFields(old) {
		if strings.Contains(v.Content, sgrColor(c)) {
			t.Errorf("frame still carries the old Theme's %s (%s)", field, hexOf(c))
		}
	}
}

func TestThemeKeyCyclesBackToTheFirst(t *testing.T) {
	keepTheme(t)
	theme = themes[0]
	m := sized(t, newModel().openFolder(fixtureTree()), 80, 24)
	for range themes {
		m = press(t, m, keys.Theme.Keys()[0])
	}
	if theme.Name != themes[0].Name {
		t.Errorf("after %d presses the theme is %q, want %q again", len(themes), theme.Name, themes[0].Name)
	}
}

// themeTextModel has a long text file with "needle" on three far-apart lines
// open in both viewers (a .txt file always renders as Text, whose gutter is
// Theme-coloured even under `go test`), Focus on DV1.
func themeTextModel(t *testing.T) model {
	t.Helper()
	var b strings.Builder
	for i := 0; i < 120; i++ {
		if i%40 == 5 {
			fmt.Fprintf(&b, "needle %d\n", i)
			continue
		}
		fmt.Fprintf(&b, "filler line %d\n", i)
	}
	dir := t.TempDir()
	path := writeFileContent(t, filepath.Join(dir, "notes.txt"), b.String())
	node, err := tree.Walk(dir, tree.Options{Filter: tree.FilterAll})
	if err != nil {
		t.Fatal(err)
	}
	m := sized(t, newModel().openFolder(node), 100, 16)
	m = openPath(t, m, dv1Pane, path)
	m = openPath(t, m, dv2Pane, path)
	return m.setFocus(dv1Pane)
}

// TestThemeKeyRestylesViewersInPlace: a viewer's pre-rendered content is
// re-rendered in the new Theme, focused and muted alike and in DV2 too, while
// its text, scroll position and committed find (query, matches, current match
// and its indication) are exactly as they were — even scrolled away from the
// current match, which must not be jumped back to.
func TestThemeKeyRestylesViewersInPlace(t *testing.T) {
	keepTheme(t)
	theme = themes[0]
	m := committedFind(t, themeTextModel(t), "needle", 1)
	m = press(t, m, keys.HalfPageDown.Keys()[0])
	m = press(t, m, keys.HalfPageDown.Keys()[0])
	before := m.dv1
	if before.findState.current != 1 || len(before.findState.matches) != 3 || before.textMatch == nil {
		t.Fatalf("setup: current %d of %d matches, indicated %v", before.findState.current, len(before.findState.matches), before.textMatch != nil)
	}
	if inView(before) {
		t.Fatalf("setup: the current match should be scrolled out of view")
	}

	m = press(t, m, keys.Theme.Keys()[0])
	next := themes[1]
	after := m.dv1

	if after.offset() != before.offset() {
		t.Errorf("offset = %d, want %d kept", after.offset(), before.offset())
	}
	if after.findState.input.Value() != "needle" || after.findState.current != 1 || len(after.findState.matches) != 3 {
		t.Errorf("find = %q, current %d of %d; want the committed find kept",
			after.findState.input.Value(), after.findState.current, len(after.findState.matches))
	}
	if after.textMatch == nil || *after.textMatch != *before.textMatch {
		t.Errorf("match indication = %v, want %v kept", after.textMatch, before.textMatch)
	}
	if got, want := ansi.Strip(after.vp.GetContent()), ansi.Strip(before.vp.GetContent()); got != want {
		t.Error("restyle changed the viewer's text")
	}
	for what, got := range map[string]string{
		"DV1 content": after.vp.GetContent(),
		"DV2 content": m.dv2.vp.GetContent(),
	} {
		if !strings.Contains(got, sgrColor(next.Subtle)) || strings.Contains(got, sgrColor(themes[0].Subtle)) {
			t.Errorf("%s gutter is not in the new Theme's Subtle", what)
		}
	}
	if !strings.Contains(after.mutedContent, sgrColor(next.UnfocusedSubtle)) {
		t.Error("muted content gutter is not in the new Theme's UnfocusedSubtle")
	}
}

func TestThemeChoiceRoundTrips(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "theme")
	want := themes[len(themes)-1]
	if err := saveThemeChoice(path, want.Name); err != nil {
		t.Fatal(err)
	}
	got, ok := loadThemeChoice(path)
	if !ok || got.Name != want.Name {
		t.Fatalf("loadThemeChoice = %q, %v; want %q", got.Name, ok, want.Name)
	}
}

func TestLoadThemeChoiceIgnoresMissingEmptyAndUnknown(t *testing.T) {
	dir := t.TempDir()
	if _, ok := loadThemeChoice(filepath.Join(dir, "absent")); ok {
		t.Error("a missing file yielded a theme")
	}
	if _, ok := loadThemeChoice(""); ok {
		t.Error("an empty path yielded a theme")
	}
	for name, body := range map[string]string{"empty": "", "unknown": "no-such-theme\n"} {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, ok := loadThemeChoice(p); ok {
			t.Errorf("%s file yielded a theme", name)
		}
	}
}

func TestThemeKeyPersistsTheChoice(t *testing.T) {
	keepTheme(t)
	theme = themes[0]
	path := filepath.Join(t.TempDir(), "theme")
	m := sized(t, newModel().withThemePath(path).openFolder(fixtureTree()), 80, 24)
	press(t, m, keys.Theme.Keys()[0])
	got, ok := loadThemeChoice(path)
	if !ok || got.Name != themes[1].Name {
		t.Errorf("persisted theme = %q, %v; want %q", got.Name, ok, themes[1].Name)
	}
}
