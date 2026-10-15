package main

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// plainHints is a footer line as text: styling stripped and lower-cased, so a
// check names an action ("star") without restating its capitalisation.
func plainHints(s string) string { return strings.ToLower(ansi.Strip(s)) }

// focusHints are the Focus* bindings as they would read in a plainHints line,
// for the checks that they never appear there.
func focusHints() []string {
	var out []string
	for _, b := range []key.Binding{keys.FocusRoots, keys.FocusFiles, keys.FocusDV1, keys.FocusDV2} {
		out = append(out, plainHints(formatHint(b)))
	}
	return out
}

// TestHintTextChangesWithFocus checks S11's "key hints relevant to the
// focused pane": each pane's full-width hint text names an action that is
// specific to it and not the others (M4.2).
func TestHintTextChangesWithFocus(t *testing.T) {
	roots := plainHints(hintText(rootsPane, 200))
	files := plainHints(hintText(filesPane, 200))
	dv1 := plainHints(hintText(dv1Pane, 200))

	cases := []struct {
		name string
		text string
		want string
	}{
		{"roots names star", roots, "star"},
		{"files names expand", files, "expand"},
		{"viewer names pgup", dv1, "pgup"},
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
			text := plainHints(hintText(p, width))
			if !strings.Contains(text, "quit") {
				t.Errorf("plainHints(hintText(%v, %d)) = %q, want it to contain quit", p, width, text)
			}
		}
	}
}

// TestHintTextTruncatesButKeepsHelpAffordance is S11's "truncate gracefully
// ... advertising the help affordance": a narrow footer drops pane-specific
// hints but still names help.
func TestHintTextTruncatesButKeepsHelpAffordance(t *testing.T) {
	wide := plainHints(hintText(filesPane, 200))
	narrow := plainHints(hintText(filesPane, 20))

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

// TestJoinHintsSeparatesWithPipes checks the footer's entry shape: title and
// "<key> <Action>" hints joined by " | ", the action capitalised.
func TestJoinHintsSeparatesWithPipes(t *testing.T) {
	text := ansi.Strip(joinHints("Files", []key.Binding{keys.Activate}))
	want := "Files | enter Open | ? Help | q Quit"
	if text != want {
		t.Errorf("joinHints = %q, want %q", text, want)
	}
}

// TestHintKeysTakeTheAccentColour: each hint's key is drawn in the Theme's
// Accent so the keys stand out along the line; the action is not.
func TestHintKeysTakeTheAccentColour(t *testing.T) {
	accent := "38;" + sgrColor(theme.Accent) + "m"
	got := formatHint(keys.Activate)
	if !strings.Contains(got, accent+"enter") {
		t.Errorf("formatHint = %q, want the key in Accent (%s)", got, accent)
	}
	if strings.Contains(got, accent+"Open") {
		t.Errorf("formatHint = %q, want the action not in Accent", got)
	}
}

// TestNavPanesOmitUpDownAndTheViewerFoldsItsMovement: the nav panes and the
// Picker leave the arrow keys unhinted; the viewer keeps its movement keys as
// one "Nav" entry.
func TestNavPanesOmitUpDownAndTheViewerFoldsItsMovement(t *testing.T) {
	for _, text := range []string{plainHints(hintText(rootsPane, 400)), plainHints(hintText(filesPane, 400)), plainHints(pickerHintText(400))} {
		if strings.Contains(text, "↑") || strings.Contains(text, "↓") {
			t.Errorf("hint text = %q, want no up/down hint", text)
		}
	}
	dv := ansi.Strip(hintText(dv1Pane, 400))
	if !strings.Contains(dv, "Document | ↑/↓/pgup/pgdn/g/G Nav | ") {
		t.Errorf("viewer hint text = %q, want one leading Nav entry", dv)
	}
	if strings.Contains(dv, "Page up") || strings.Contains(dv, "Top") {
		t.Errorf("viewer hint text = %q, want movement only inside the Nav entry", dv)
	}
}

// TestPaneHintsNameThePickerFromEveryPane is D39's "M2.3 makes the [ctrl+p]
// binding visible/discoverable": every pane's footer hints now name the
// Picker, the same way they already name Edit/Navigator regardless of
// focus.
func TestPaneHintsNameThePickerFromEveryPane(t *testing.T) {
	for _, p := range []pane{rootsPane, filesPane, dv1Pane, dv2Pane} {
		text := plainHints(hintText(p, 200))
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
// pane hint once the Picker has its own star action (keys.Star,
// distinct from the Sources Nav/Files Nav/viewer's own keys.Star — see
// keymap.go's own comment on PickerStar for why).
func TestPickerHintTextNamesPickerActionsNotThePane(t *testing.T) {
	text := plainHints(pickerHintText(200))
	for _, want := range []string{"picker", "cancel", "page up", "page down", "star"} {
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
	before := plainHints(m.renderFooter(rect{width: 80, height: 1}))

	m, _ = m.openPicker()
	openFooter := plainHints(m.renderFooter(rect{width: 80, height: 1}))
	if strings.Contains(openFooter, "expand") {
		t.Errorf("footer while picker open = %q, unexpectedly still names Files pane's \"expand\"", openFooter)
	}
	if !strings.Contains(openFooter, "cancel") {
		t.Errorf("footer while picker open = %q, want it to name the Picker's cancel", openFooter)
	}

	m = m.dismissPicker()
	after := plainHints(m.renderFooter(rect{width: 80, height: 1}))
	if after != before {
		t.Errorf("footer after dismissal = %q, want it restored to %q", after, before)
	}
}

// TestPaneHintsNameRefreshFromEveryPane is C2a/M9.1's own check: Refresh
// ("R") is reachable from every pane (S6), so every pane's footer hints
// must name it, the same always-there treatment Edit/Navigator/Picker
// already get.
func TestPaneHintsNameRefreshFromEveryPane(t *testing.T) {
	for _, p := range []pane{rootsPane, filesPane, dv1Pane, dv2Pane} {
		text := plainHints(hintText(p, 200))
		if !strings.Contains(text, "refresh") {
			t.Errorf("pane %v: hint text = %q, want it to contain %q", p, text, "refresh")
		}
	}
}

// TestPaneHintsOmitEveryFocusBindingEverywhere is M7.2's own acceptance
// check (S10): none of FocusRoots/FocusFiles/FocusDV1/FocusDV2 appear in any
// pane's footer hints any more, from any pane — M7.1's per-pane title-row
// badge ([r]/[f]/[o]/[^o]) replaced the footer as their discoverability
// surface (footer.go's paneHints doc comment), so repeating them here would
// just be redundant with something the pane itself already shows. FocusDV2
// was never listed in the first place (it predates this task, C3/M1.3, for
// the "usually a no-op with DV2 absent" reason paneHints' doc comment
// gives); this folds its own former check
// (TestPaneHintsOmitFocusDV2Everywhere) in alongside the other three so one
// test covers the whole Focus* family. S12/C1's help modal remains every
// binding's discoverability backstop regardless
// (TestHelpBindingsNameFocusDV1AndDV2FromEveryPane).
func TestPaneHintsOmitEveryFocusBindingEverywhere(t *testing.T) {
	for _, p := range []pane{rootsPane, filesPane, dv1Pane, dv2Pane} {
		text := plainHints(hintText(p, 200))
		for _, notWant := range focusHints() {
			if strings.Contains(text, notWant) {
				t.Errorf("pane %v: hint text = %q, unexpectedly names a Focus binding (%q)", p, text, notWant)
			}
		}
	}
}

// TestPaneHintsNameCloseViewerOnlyInViewerPanes: ctrl+w always acts from a
// viewer (D141), so the viewers hint it; from a nav pane it is usually a
// no-op (DV2 absent), so they do not — the Help Modal is its backstop there.
func TestPaneHintsNameCloseViewerOnlyInViewerPanes(t *testing.T) {
	for _, p := range []pane{dv1Pane, dv2Pane} {
		if text := plainHints(hintText(p, 200)); !strings.Contains(text, "close viewer") {
			t.Errorf("pane %v: hint text = %q, want it to name CloseViewer", p, text)
		}
	}
	for _, p := range []pane{rootsPane, filesPane} {
		if text := plainHints(hintText(p, 400)); strings.Contains(text, "close viewer") {
			t.Errorf("pane %v: hint text = %q, want it to omit CloseViewer", p, text)
		}
	}
}

// TestPaneHintsNamePeekOnlyInFilesPane checks S7's "Peek, in the Files Nav
// only" against the footer: filesPane's hints name it, but rootsPane's and
// dv1Pane/dv2Pane's — which can never trigger it — do not, the same
// inapplicable-binding-is-worse-than-no-hint reasoning paneHints already
// gives for the viewer's own omissions.
func TestPaneHintsNamePeekOnlyInFilesPane(t *testing.T) {
	if !strings.Contains(plainHints(hintText(filesPane, 200)), "peek") {
		t.Errorf("filesPane hint text = %q, want it to contain %q", plainHints(hintText(filesPane, 200)), "peek")
	}
	for _, p := range []pane{rootsPane, dv1Pane, dv2Pane} {
		text := plainHints(hintText(p, 200))
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
	// Matching TestPaneHintsNamePeekOnlyInFilesPane: FilesMode/FilesScope
	// sort earlier than Peek in paneHints' filesPane list, so any width wide
	// enough to keep "peek" untruncated keeps these two too.
	text := plainHints(hintText(filesPane, 200))
	if !strings.Contains(text, "mode") {
		t.Errorf("filesPane hint text = %q, want it to contain %q", text, "mode")
	}
	if !strings.Contains(text, "scope") {
		t.Errorf("filesPane hint text = %q, want it to contain %q", text, "scope")
	}
	for _, p := range []pane{rootsPane, dv1Pane, dv2Pane} {
		text := plainHints(hintText(p, 200))
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
		text := plainHints(hintText(p, 200))
		if !strings.Contains(text, "back") {
			t.Errorf("pane %v hint text = %q, want it to contain %q", p, text, "back")
		}
		if !strings.Contains(text, "forward") {
			t.Errorf("pane %v hint text = %q, want it to contain %q", p, text, "forward")
		}
	}
	for _, p := range []pane{rootsPane, filesPane} {
		text := plainHints(hintText(p, 200))
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
	text := plainHints(pickerHintText(200))
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
		if text := plainHints(hintText(p, 200)); strings.Contains(text, "ctrl+p") {
			t.Errorf("pane %v: hint text = %q, unexpectedly names ctrl+p", p, text)
		}
	}
	if text := plainHints(pickerHintText(200)); strings.Contains(text, "ctrl+p") {
		t.Errorf("pickerHintText = %q, unexpectedly names ctrl+p", text)
	}
	// filesPane's full hint list (paneHints) is wide enough that 200 cols
	// already truncates ActivateDV2/PeekDV2 away (they sort last, least
	// essential) regardless of this rebind, so this half needs a generous
	// width to actually observe the untruncated line. 270, not 260 (M4.8):
	// FilesMode/FilesScope grew the full line past 260 too.
	if text := plainHints(hintText(filesPane, 270)); !strings.Contains(text, "ctrl+p") {
		t.Errorf("filesPane hint text = %q, want it to name ctrl+p (PeekDV2, D55)", text)
	}
}

// TestFooterAtWideWidthOmitsFocusHintsButKeepsEverythingElse is M7.2's own
// render-level acceptance check, going through model.renderFooter (the
// frame's actual entry point) rather than the bare hintText helper the
// tests above already cover: at a wide-enough width every pane's rendered
// Footer still names Refresh, Help and Quit plus its own pane-local hints
// (paneHints' doc comment), unchanged by M7.2's removal, and never a
// Focus* binding — the M7.1 title-row badge is that discoverability
// surface now instead. 200, not a wider width, deliberately: filesPane's
// full hint list includes ActivateDV2's "open in dv2" (unrelated to
// FocusDV2, S4/C3/M1.4), which would collide with a bare "dv2" substring
// check at a width generous enough to keep it untruncated — 200 already
// truncates it away (TestCtrlPNamedOnlyByPeekDV2's own comment), the same
// width every other Focus-omission check in this file already uses.
func TestFooterAtWideWidthOmitsFocusHintsButKeepsEverythingElse(t *testing.T) {
	m := sized(t, newModel().openFolder(fixtureTree()), 80, 24)
	m = m.withViewerFor(dv1Pane, openedFixture(t, 80, 24).dv1)
	m = sized(t, m, 80, 24)

	for _, want := range []struct {
		pane     pane
		haveWant []string
	}{
		{rootsPane, []string{"star", "refresh", "help", "quit"}},
		{filesPane, []string{"expand", "collapse", "star", "refresh", "peek", "help", "quit"}},
		{dv1Pane, []string{"pgup", "star", "refresh", "help", "quit"}},
	} {
		m.focus = want.pane
		text := plainHints(m.renderFooter(rect{width: 200, height: 1}))
		for _, h := range want.haveWant {
			if !strings.Contains(text, h) {
				t.Errorf("pane %v: footer = %q, want it to contain %q", want.pane, text, h)
			}
		}
		for _, notWant := range focusHints() {
			if strings.Contains(text, notWant) {
				t.Errorf("pane %v: footer = %q, unexpectedly names a Focus binding (%q)", want.pane, text, notWant)
			}
		}
	}
}
