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
	dv1 := hintText(dv1Pane, 200)

	cases := []struct {
		name string
		text string
		want string
	}{
		{"roots names star", roots, "star"},
		{"files names expand", files, "expand"},
		{"viewer names page", dv1, "page"},
	}
	for _, c := range cases {
		if !strings.Contains(c.text, c.want) {
			t.Errorf("%s: hint text = %q, want it to contain %q", c.name, c.text, c.want)
		}
	}
	if roots == files || files == dv1 || roots == dv1 {
		t.Errorf("hint text does not differ by focus: roots=%q files=%q viewer=%q", roots, files, dv1)
	}
}

// TestHintTextAlwaysIncludesQuit checks S11's "always including quit", at
// both a generous and a narrow width.
func TestHintTextAlwaysIncludesQuit(t *testing.T) {
	for _, width := range []int{200, 30, 10} {
		for _, p := range []pane{rootsPane, filesPane, dv1Pane, dv2Pane} {
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
	for _, p := range []pane{rootsPane, filesPane, dv1Pane, dv2Pane} {
		text := hintText(p, 200)
		if !strings.Contains(text, "picker") {
			t.Errorf("pane %v: hint text = %q, want it to contain %q", p, text, "picker")
		}
	}
}

// TestPickerHintTextNamesPickerActionsNotThePane is S2's "the Footer shows
// the Picker's actions": pickerHintText names the Picker's own non-typing
// bindings (here, cancel and up/down selection movement, and M2.5's own
// star toggle) rather than the pane-specific hints hintText would give for,
// say, the Files pane ("expand"/"collapse", absent here). "star" moved from
// the not-wanted list to the wanted one in M2.5: it is no longer a leaked
// pane hint once the Picker has its own star action (keys.PickerStar,
// distinct from the Roots Nav/Files Nav/viewer's own keys.Star — see
// keymap.go's own comment on PickerStar for why).
func TestPickerHintTextNamesPickerActionsNotThePane(t *testing.T) {
	text := pickerHintText(200)
	for _, want := range []string{"Picker", "cancel", "up", "down", "star"} {
		if !strings.Contains(text, want) {
			t.Errorf("pickerHintText = %q, want it to contain %q", text, want)
		}
	}
	for _, notWant := range []string{"expand", "collapse"} {
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

// TestPaneHintsNameFocusAndRefreshFromEveryPane is C2a/M9.1's own sweep of
// S6's global bindings: FocusRoots ("r"), FocusFiles ("f") and Refresh ("R")
// are reachable from every pane (S6), so every pane's footer hints must name
// all three, the same always-there treatment Edit/Navigator/Picker already
// get. FocusDV1 ("dv1", C3/M1.3) joins them here — it is just as globally
// applicable as FocusRoots/FocusFiles, unlike FocusDV2, which paneHints
// deliberately omits everywhere (its own doc comment) since it is usually a
// no-op with DV2 absent; TestPaneHintsOmitFocusDV2Everywhere checks that
// half separately.
func TestPaneHintsNameFocusAndRefreshFromEveryPane(t *testing.T) {
	for _, p := range []pane{rootsPane, filesPane, dv1Pane, dv2Pane} {
		// 220, not 200 (M4.8): filesPane's own hint list grew by
		// FilesMode/FilesScope, which now pushes "refresh" past 200 cols'
		// truncation point (footer.go's paneHints — it drops from the end).
		text := hintText(p, 220)
		for _, want := range []string{"roots", "files", "dv1", "refresh"} {
			if !strings.Contains(text, want) {
				t.Errorf("pane %v: hint text = %q, want it to contain %q", p, text, want)
			}
		}
	}
}

// TestPaneHintsOmitFocusDV2Everywhere is the other half of C3/M1.3's own
// addition: FocusDV2 ("ctrl+o") never appears in the footer's terse hints, from
// any pane — paneHints' doc comment gives the same "inapplicable binding is
// worse than no hint" reasoning the viewer case already applies to
// FindNext/FindPrevious/Cancel. S12's help modal is its discoverability
// backstop instead (TestHelpBindingsNameFocusDV1AndDV2FromEveryPane).
func TestPaneHintsOmitFocusDV2Everywhere(t *testing.T) {
	for _, p := range []pane{rootsPane, filesPane, dv1Pane, dv2Pane} {
		if text := hintText(p, 200); strings.Contains(text, "dv2") {
			t.Errorf("pane %v: hint text = %q, want it to omit FocusDV2", p, text)
		}
	}
}

// TestPaneHintsOmitCloseDV2Everywhere is CloseDV2's (S4, C3/M2.1) own half
// of the same carve-out TestPaneHintsOmitFocusDV2Everywhere pins for
// FocusDV2: also usually a no-op with DV2 absent, so paneHints omits it
// from every pane's terse footer hints too — the Help Modal is its
// discoverability backstop instead
// (TestHelpBindingsNameFocusRefreshAndToggleHiddenFromEveryPane).
func TestPaneHintsOmitCloseDV2Everywhere(t *testing.T) {
	for _, p := range []pane{rootsPane, filesPane, dv1Pane, dv2Pane} {
		if text := hintText(p, 200); strings.Contains(text, "close dv2") {
			t.Errorf("pane %v: hint text = %q, want it to omit CloseDV2", p, text)
		}
	}
}

// TestPaneHintsNamePeekOnlyInFilesPane checks S7's "Peek, in the Files Nav
// only" against the footer: filesPane's hints name it, but rootsPane's and
// dv1Pane/dv2Pane's — which can never trigger it — do not, the same
// inapplicable-binding-is-worse-than-no-hint reasoning paneHints already
// gives for the viewer's own omissions.
func TestPaneHintsNamePeekOnlyInFilesPane(t *testing.T) {
	// 220, not 200 (M4.8): see TestPaneHintsNameFocusAndRefreshFromEveryPane's
	// comment — filesPane's grown hint list now truncates "peek" away at 200.
	if !strings.Contains(hintText(filesPane, 220), "peek") {
		t.Errorf("filesPane hint text = %q, want it to contain %q", hintText(filesPane, 220), "peek")
	}
	for _, p := range []pane{rootsPane, dv1Pane, dv2Pane} {
		text := hintText(p, 220)
		if strings.Contains(text, "peek") {
			t.Errorf("pane %v: hint text = %q, unexpectedly names Peek (Files-Nav-only)", p, text)
		}
	}
}

// TestPaneHintsNameFilesModeAndScopeOnlyInFilesPane checks S5's Browse
// mode/scope cycle keys (C4/M4.1) against the footer, mirroring
// TestPaneHintsNamePeekOnlyInFilesPane's shape directly above: filesPane's
// hints name them, but rootsPane's and dv1Pane/dv2Pane's — which can never
// trigger a Files-Nav-only cycle — do not.
func TestPaneHintsNameFilesModeAndScopeOnlyInFilesPane(t *testing.T) {
	// 220, matching TestPaneHintsNamePeekOnlyInFilesPane: FilesMode/FilesScope
	// sort earlier than Peek in paneHints' filesPane list, so any width wide
	// enough to keep "peek" untruncated keeps these two too.
	text := hintText(filesPane, 220)
	if !strings.Contains(text, "mode") {
		t.Errorf("filesPane hint text = %q, want it to contain %q", text, "mode")
	}
	if !strings.Contains(text, "scope") {
		t.Errorf("filesPane hint text = %q, want it to contain %q", text, "scope")
	}
	for _, p := range []pane{rootsPane, dv1Pane, dv2Pane} {
		text := hintText(p, 220)
		if strings.Contains(text, "mode") {
			t.Errorf("pane %v: hint text = %q, unexpectedly names mode (Files-Nav-only)", p, text)
		}
		if strings.Contains(text, "scope") {
			t.Errorf("pane %v: hint text = %q, unexpectedly names scope (Files-Nav-only)", p, text)
		}
	}
}

// TestPaneHintsNameHistoryBackAndForwardOnlyInViewerPanes checks S9's
// per-viewer back/forward walk keys (C4/M6.1) against the footer, mirroring
// TestPaneHintsNamePeekOnlyInFilesPane's shape: dv1Pane's and dv2Pane's
// hints name them, but rootsPane's and filesPane's — which have no history
// of their own to walk — do not.
func TestPaneHintsNameHistoryBackAndForwardOnlyInViewerPanes(t *testing.T) {
	for _, p := range []pane{dv1Pane, dv2Pane} {
		text := hintText(p, 200)
		if !strings.Contains(text, "back") {
			t.Errorf("pane %v hint text = %q, want it to contain %q", p, text, "back")
		}
		if !strings.Contains(text, "forward") {
			t.Errorf("pane %v hint text = %q, want it to contain %q", p, text, "forward")
		}
	}
	for _, p := range []pane{rootsPane, filesPane} {
		text := hintText(p, 200)
		if strings.Contains(text, "back") {
			t.Errorf("pane %v: hint text = %q, unexpectedly names back (viewer-only)", p, text)
		}
		if strings.Contains(text, "forward") {
			t.Errorf("pane %v: hint text = %q, unexpectedly names forward (viewer-only)", p, text)
		}
	}
}

// TestPickerHintTextNamesToggleHidden checks S4/M6.2's dotfile toggle is
// named among the open Picker's footer hints (pickerHints), the same
// always-does-something-while-open reasoning PickerMode/PickerScope already
// get.
func TestPickerHintTextNamesToggleHidden(t *testing.T) {
	text := pickerHintText(200)
	if !strings.Contains(text, "hidden files") {
		t.Errorf("pickerHintText = %q, want it to contain %q", text, "hidden files")
	}
}

// TestCtrlPNamedOnlyByPeekDV2 supersedes the pre-C3 invariant "no footer hint
// may ever name ctrl+p" (M2.2 retired it as the Picker's own open binding in
// favor of "space"): C3/M1.4a's D55 rebind gives PeekDV2 the same chord, so
// it is a legitimate hint again — but only in filesPane's list, the one
// paneHints entry PeekDV2 actually belongs to. Every other pane's hints and
// the Picker's own hint line must still never mention it.
func TestCtrlPNamedOnlyByPeekDV2(t *testing.T) {
	for _, p := range []pane{rootsPane, dv1Pane, dv2Pane} {
		if text := hintText(p, 200); strings.Contains(text, "ctrl+p") {
			t.Errorf("pane %v: hint text = %q, unexpectedly names ctrl+p", p, text)
		}
	}
	if text := pickerHintText(200); strings.Contains(text, "ctrl+p") {
		t.Errorf("pickerHintText = %q, unexpectedly names ctrl+p", text)
	}
	// filesPane's full hint list (paneHints) is wide enough that 200 cols
	// already truncates ActivateDV2/PeekDV2 away (they sort last, least
	// essential) regardless of this rebind, so this half needs a generous
	// width to actually observe the untruncated line. 270, not 260 (M4.8):
	// FilesMode/FilesScope grew the full line past 260 too.
	if text := hintText(filesPane, 270); !strings.Contains(text, "ctrl+p") {
		t.Errorf("filesPane hint text = %q, want it to name ctrl+p (PeekDV2, D55)", text)
	}
}
