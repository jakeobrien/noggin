package main

import (
	"encoding/json"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

// TestUnfocusedSelectedRowIsADimBar: a focused nav pane's selected row is a
// reverse-video bar; an unfocused one is a grey background with no reverse.
func TestUnfocusedSelectedRowIsADimBar(t *testing.T) {
	m := sized(t, newModel().openFolder(fixtureTree()), 80, 24)
	selLine := func(f filesNav) string { return strings.Split(f.view(), "\n")[1] }

	focused := selLine(m.files)
	if !strings.Contains(focused, "7m") {
		t.Errorf("focused selected row is not reverse-video: %q", focused)
	}
	f := m.files
	f.unfocused = true
	dim := selLine(f)
	if !strings.Contains(dim, "48;"+sgrColor(theme.UnfocusedBar)) || hasReverse(dim) {
		t.Errorf("unfocused selected row is not a plain dim bar: %q", dim)
	}
}

func TestMuteColorsRewritesColorsOnly(t *testing.T) {
	var doc any
	src := `{"document":{"color":"223","margin":2},"code_block":{"chroma":{"keyword":{"color":"#00AAFF"},"background":{"background_color":"#373737"}}}}`
	if err := json.Unmarshal([]byte(src), &doc); err != nil {
		t.Fatal(err)
	}
	muteColors(doc, theme.Bg, theme.MutedFade)
	out, _ := json.Marshal(doc)
	got := string(out)
	for _, orig := range []string{`"223"`, `#00AAFF`, `#373737`} {
		if strings.Contains(got, orig) {
			t.Errorf("muted style still has original colour %s: %s", orig, got)
		}
	}
	if !strings.Contains(got, `"margin":2`) {
		t.Errorf("muteColors touched a non-colour field: %s", got)
	}
}

func TestMutedColorMovesTowardTheBackground(t *testing.T) {
	dark := mutedColor("#ffffff", lipgloss.Color("#000000"), 0.5)
	light := mutedColor("#000000", lipgloss.Color("#ffffff"), 0.5)
	if dark >= "#ffffff" || dark == "#000000" {
		t.Errorf("dark-theme white muted to %s, want a mid grey", dark)
	}
	if light <= "#000000" || light == "#ffffff" {
		t.Errorf("light-theme black muted to %s, want a mid grey", light)
	}
}

// TestMutedViewerShowsMutedContentAtTheSameOffset: a muted viewer swaps in
// the muted render without disturbing scroll position, and an unmuted one is
// unaffected.
func TestMutedViewerShowsMutedContentAtTheSameOffset(t *testing.T) {
	var lines, mutedLines []string
	for i := 0; i < 50; i++ {
		lines = append(lines, "line")
		mutedLines = append(mutedLines, "MUTED")
	}
	v := newViewer()
	v.vp.SetWidth(20)
	v.vp.SetHeight(5)
	v.doc = document{path: "/tmp/x.md", src: []byte("x")}
	v.vp.SetContent(strings.Join(lines, "\n"))
	v.mutedContent = strings.Join(mutedLines, "\n")
	v = v.withOffset(10)

	if got := v.view(); strings.Contains(got, "MUTED") {
		t.Fatalf("unmuted viewer showed muted content: %q", got)
	}
	v.muted = true
	got := v.view()
	if !strings.Contains(got, "MUTED") {
		t.Fatalf("muted viewer did not show muted content: %q", got)
	}
	if v.offset() != 10 {
		t.Errorf("view() disturbed the viewer's own offset: %d", v.offset())
	}
}

// TestUnfocusedNavPanesMuteUnselectedRowsOnly: an unfocused nav pane's
// non-selected text is recoloured, while its selected row is left to
// selectedRowStyle.
func TestUnfocusedNavPanesMuteUnselectedRowsOnly(t *testing.T) {
	m := sized(t, newModel().openFolder(fixtureTree()), 80, 24)

	f := m.files
	f.unfocused = true
	muted := f.view()
	if m.files.view() == muted {
		t.Fatalf("unfocusing the Files Nav changed nothing: %q", muted)
	}
	if !strings.Contains(muted, "38;"+sgrColor(theme.UnfocusedText)) && !strings.Contains(muted, "38;"+sgrColor(theme.UnfocusedSubtle)) {
		t.Errorf("unfocused Files Nav has no muted foreground: %q", muted)
	}
	if sel := strings.Split(muted, "\n")[1]; strings.Contains(sel, "38;"+sgrColor(theme.UnfocusedText)) {
		t.Errorf("selected row was muted as well: %q", sel)
	}

	r := newRootsNav().setRoots([]root{{path: t.TempDir()}, {path: t.TempDir()}}, -1).setSize(40, 8)
	plainR := r.view()
	r.unfocused = true
	if plainR == r.view() {
		t.Errorf("unfocusing the Sources Nav changed nothing: %q", plainR)
	}
}
