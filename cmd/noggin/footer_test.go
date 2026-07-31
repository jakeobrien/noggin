package main

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
)

// TestHintTextChangesWithFocus checks S11's "key hints relevant to the
// focused pane": each pane's full-width hint text names an action that is
// specific to it and not the others (M4.2).
func TestHintTextChangesWithFocus(t *testing.T) {
	roots := hintText(rootsPane, 200)
	files := hintText(filesPane, 200)
	viewer := hintText(viewerPane, 200)

	cases := []struct {
		name string
		text string
		want string
	}{
		{"roots names star", roots, "star"},
		{"files names expand", files, "expand"},
		{"viewer names page", viewer, "page"},
	}
	for _, c := range cases {
		if !strings.Contains(c.text, c.want) {
			t.Errorf("%s: hint text = %q, want it to contain %q", c.name, c.text, c.want)
		}
	}
	if roots == files || files == viewer || roots == viewer {
		t.Errorf("hint text does not differ by focus: roots=%q files=%q viewer=%q", roots, files, viewer)
	}
}

// TestHintTextAlwaysIncludesQuit checks S11's "always including quit", at
// both a generous and a narrow width.
func TestHintTextAlwaysIncludesQuit(t *testing.T) {
	for _, width := range []int{200, 30, 10} {
		for _, p := range paneOrder {
			text := hintText(p, width)
			if !strings.Contains(text, "quit") {
				t.Errorf("hintText(%v, %d) = %q, want it to contain quit", p, width, text)
			}
		}
	}
}

// TestHintTextTruncatesButKeepsHelpAffordance is S11's "truncate gracefully
// ... advertising the help affordance": a narrow footer drops pane-specific
// hints but still names help.
func TestHintTextTruncatesButKeepsHelpAffordance(t *testing.T) {
	wide := hintText(filesPane, 200)
	narrow := hintText(filesPane, 20)

	if len(narrow) >= len(wide) {
		t.Errorf("narrow hint text (%q) is not shorter than wide (%q)", narrow, wide)
	}
	if !strings.Contains(narrow, "help") {
		t.Errorf("narrow hint text = %q, want it to still name help", narrow)
	}
	if lipgloss.Width(narrow) > 20 {
		t.Errorf("narrow hint text %q is %d cells wide, want <= 20", narrow, lipgloss.Width(narrow))
	}
}

// TestJoinHintsSeparatesWithMiddleDot checks the render shape the footer and
// (later) the help modal share: title and hints joined by " · ", the same
// separator the pre-M4.2 footer used.
func TestJoinHintsSeparatesWithMiddleDot(t *testing.T) {
	text := joinHints("Files", []key.Binding{keys.Activate})
	want := "Files · " + formatHint(keys.Activate) + " · " + formatHint(keys.Help) + " · " + formatHint(keys.Quit)
	if text != want {
		t.Errorf("joinHints = %q, want %q", text, want)
	}
}
