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

// TestPaneHintsNameThePickerFromEveryPane is D39's "M2.3 makes the [ctrl+p]
// binding visible/discoverable": every pane's footer hints now name the
// Picker, the same way they already name Edit/Navigator regardless of
// focus.
func TestPaneHintsNameThePickerFromEveryPane(t *testing.T) {
	for _, p := range paneOrder {
		text := hintText(p, 200)
		if !strings.Contains(text, "picker") {
			t.Errorf("pane %v: hint text = %q, want it to contain %q", p, text, "picker")
		}
	}
}

// TestPickerHintTextNamesPickerActionsNotThePane is S2's "the Footer shows
// the Picker's actions": pickerHintText names the Picker's own non-typing
// bindings (here, cancel and up/down selection movement) rather than the
// pane-specific hints hintText would give for, say, the Files pane
// ("expand"/"collapse", absent here).
func TestPickerHintTextNamesPickerActionsNotThePane(t *testing.T) {
	text := pickerHintText(200)
	for _, want := range []string{"Picker", "cancel", "up", "down"} {
		if !strings.Contains(text, want) {
			t.Errorf("pickerHintText = %q, want it to contain %q", text, want)
		}
	}
	for _, notWant := range []string{"expand", "collapse", "star"} {
		if strings.Contains(text, notWant) {
			t.Errorf("pickerHintText = %q, unexpectedly contains pane-specific hint %q", text, notWant)
		}
	}
}

// TestFooterShowsPickerHintsWhileOpen is the model-level wiring for the
// same requirement: opening the Picker while a Nav pane is focused swaps
// the rendered footer from that pane's hints to the Picker's own, and
// closing it restores the original.
func TestFooterShowsPickerHintsWhileOpen(t *testing.T) {
	m := openedFixture(t, 80, 24)
	m.focus = filesPane
	before := m.renderFooter(rect{width: 80, height: 1})

	m, _ = m.openPicker()
	openFooter := m.renderFooter(rect{width: 80, height: 1})
	if strings.Contains(openFooter, "expand") {
		t.Errorf("footer while picker open = %q, unexpectedly still names Files pane's \"expand\"", openFooter)
	}
	if !strings.Contains(openFooter, "cancel") {
		t.Errorf("footer while picker open = %q, want it to name the Picker's cancel", openFooter)
	}

	m = m.dismissPicker()
	after := m.renderFooter(rect{width: 80, height: 1})
	if after != before {
		t.Errorf("footer after dismissal = %q, want it restored to %q", after, before)
	}
}
