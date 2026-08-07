package main

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// TestHelpOpensOverAnyPaneWithoutMovingFocus is S12's "opens over any pane":
// pressing Help opens the modal from whichever pane has focus, and does not
// move focus (the Help Modal is a layer over the frame, not a pane swap —
// unlike activate()/activateRoot(), which do move focus).
func TestHelpOpensOverAnyPaneWithoutMovingFocus(t *testing.T) {
	for _, p := range []pane{rootsPane, filesPane, viewerPane} {
		m := openedFixture(t, 80, 24)
		m.focus = p
		m = press(t, m, "?")
		if !m.helpOpen {
			t.Fatalf("focus %v: Help did not open the modal", p)
		}
		if m.focus != p {
			t.Errorf("focus %v: opening Help changed focus to %v", p, m.focus)
		}
	}
}

// TestHelpModalListsThePaneOwnBindings is the acceptance criterion's own
// wording: the modal "lists the bindings that pane's keymap actually
// defines" — checked directly against helpBox's own content (isolating it
// from whatever the pane underneath happens to be showing, which the modal's
// true compositing leaves visible around its edges — see
// TestHelpOverlayLeavesTheFrameVisibleAroundIt for that).
func TestHelpModalListsThePaneOwnBindings(t *testing.T) {
	cases := []struct {
		pane    pane
		want    []string
		wantNot []string
	}{
		{rootsPane, []string{"star", "up", "down"}, []string{"expand", "collapse", "next match"}},
		{filesPane, []string{"expand", "collapse"}, []string{"star", "next match"}},
		{viewerPane, []string{"page up", "next match", "prev match", "cancel"}, []string{"star", "expand"}},
	}
	for _, c := range cases {
		box := helpBox(c.pane)
		if !strings.Contains(box, "Help: "+c.pane.title()) {
			t.Errorf("pane %v: modal does not show its own title, got:\n%s", c.pane, box)
		}
		for _, want := range c.want {
			if !strings.Contains(box, want) {
				t.Errorf("pane %v: modal missing hint %q, got:\n%s", c.pane, want, box)
			}
		}
		for _, notWant := range c.wantNot {
			if strings.Contains(box, notWant) {
				t.Errorf("pane %v: modal unexpectedly contains %q (not this pane's binding)", c.pane, notWant)
			}
		}
	}
}

// TestHelpOverlayLeavesTheFrameVisibleAroundIt checks the "opens over" half
// of S12 directly: the composited view still shows the pane underneath
// outside the modal's own box (here, the top row's border and the footer,
// which the centered modal at 80x24 does not reach), proving real cell
// compositing rather than a full-screen replacement. Comparisons run on
// ansi.Strip'd text: the compositor's canvas round-trip re-serializes SGR
// runs (e.g. trims trailing padding, merges adjacent same-style borders)
// without changing what a terminal actually displays, so a byte-exact
// comparison would fail on formatting, not content.
func TestHelpOverlayLeavesTheFrameVisibleAroundIt(t *testing.T) {
	// Height 30, not 24 (C2a/M9.1 grew filesPane's helpBindings list past
	// what a 24-row terminal fits without clamping — see
	// TestHelpOverlayNeverExceedsTheTerminal for that clamped case, which
	// this test is not exercising).
	m := sized(t, newModel().openFolder(fixtureTree()), 80, 30)
	beforeLines := plainLines(m.View().Content)

	m = press(t, m, "?")
	afterLines := plainLines(m.View().Content)

	if len(afterLines) != len(beforeLines) {
		t.Fatalf("line count changed: %d vs %d", len(afterLines), len(beforeLines))
	}
	if afterLines[0] != beforeLines[0] {
		t.Errorf("top row changed under the modal overlay:\nbefore: %q\nafter:  %q", beforeLines[0], afterLines[0])
	}
	// The canvas round-trip also trims trailing padding spaces on
	// otherwise-unaffected rows (invisible on a real terminal either way),
	// so trailing space is ignored in this comparison too.
	if strings.TrimRight(afterLines[len(afterLines)-1], " ") != strings.TrimRight(beforeLines[len(beforeLines)-1], " ") {
		t.Errorf("footer row changed under the modal overlay:\nbefore: %q\nafter:  %q",
			beforeLines[len(beforeLines)-1], afterLines[len(afterLines)-1])
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

// plainLines splits a rendered frame into its rows with ANSI escapes
// stripped, for comparisons that care about displayed content, not styling
// serialization.
func plainLines(s string) []string {
	return strings.Split(ansi.Strip(s), "\n")
}

// TestHelpBindingsAlwaysIncludeHelpAndQuit checks every pane's list ends with
// the alwaysHints pair, the same guarantee the footer gives (S11) so the
// modal's own affordance and the way out are never the thing missing.
func TestHelpBindingsAlwaysIncludeHelpAndQuit(t *testing.T) {
	for _, p := range []pane{rootsPane, filesPane, viewerPane} {
		bindings := helpBindings(p)
		if len(bindings) < 2 ||
			bindings[len(bindings)-2].Help().Key != keys.Help.Help().Key ||
			bindings[len(bindings)-1].Help().Key != keys.Quit.Help().Key {
			t.Errorf("pane %v: bindings do not end with Help, Quit", p)
		}
	}
}

// TestHelpDismissesWithCancelReturningFocusWhereItWas covers "one key closes
// it returning focus where it was" directly: the frame after Cancel is
// byte-identical to the frame before Help was pressed, the same
// byte-identical-round-trip proof M5.2's navigator toggle test uses.
func TestHelpDismissesWithCancelReturningFocusWhereItWas(t *testing.T) {
	m := openedFixture(t, 80, 24)
	before := m.View().Content

	m = press(t, m, "?")
	if !m.helpOpen {
		t.Fatal("Help did not open the modal")
	}
	m = press(t, m, "esc")
	if m.helpOpen {
		t.Error("esc did not close the modal")
	}
	if got := m.View().Content; got != before {
		t.Errorf("frame after dismissing help != frame before opening it")
	}
}

// TestHelpDismissesWithHelpKeyToo checks the modal's own affordance also
// closes it — pressing "?" again is the natural undo, per
// handleHelpInput's doc.
func TestHelpDismissesWithHelpKeyToo(t *testing.T) {
	m := openedFixture(t, 80, 24)
	m = press(t, m, "?")
	m = press(t, m, "?")
	if m.helpOpen {
		t.Error("pressing Help twice did not close the modal")
	}
}

// TestHelpModalOwnsTheKeyboardWhileOpen checks the modal's exclusivity: a
// keypress meant for the pane underneath (here, the viewer's scroll) must
// not leak through while the modal is up, the same shape
// handleFindInput gives an in-progress search.
func TestHelpModalOwnsTheKeyboardWhileOpen(t *testing.T) {
	m := openedFixture(t, 80, 24)
	before := m.viewer.offset()

	m = press(t, m, "?")
	m = press(t, m, "G") // Bottom: would scroll if it reached the viewer
	if got := m.viewer.offset(); got != before {
		t.Errorf("a scroll key reached the viewer while Help was open: offset = %d, want %d", got, before)
	}
	if !m.helpOpen {
		t.Error("an unrelated key closed the modal; only Help/Cancel should")
	}
}

// TestHelpModalSwallowsQuitToo checks D26's specific claim directly: Quit is
// among the keys the modal traps, not just movement — pressing q or ctrl+c
// while Help is open must not quit the program.
func TestHelpModalSwallowsQuitToo(t *testing.T) {
	m := openedFixture(t, 80, 24)
	m = press(t, m, "?")
	for _, k := range keys.Quit.Keys() {
		next, cmd := m.Update(keyPress(k))
		m = concrete(t, next)
		if cmd != nil {
			if _, ok := cmd().(tea.QuitMsg); ok {
				t.Errorf("key %q quit while Help was open", k)
			}
		}
		if !m.helpOpen {
			t.Errorf("key %q closed the modal; only Help/Cancel should", k)
		}
	}
}

// TestHelpOverlayNeverExceedsTheTerminal is the edge case behind D26's
// clamping: the viewer's full binding list (scrollActions plus ten more, see
// helpBindings) renders taller than a short terminal. Without clamping, the
// compositor would size its canvas to the modal's own height rather than the
// terminal's (verified directly against this lipgloss build), growing the
// frame past what View() can draw inline. The clamp must keep the rendered
// frame at exactly width x height regardless.
func TestHelpOverlayNeverExceedsTheTerminal(t *testing.T) {
	m := openedFixture(t, 40, 10)
	m.focus = viewerPane
	m = press(t, m, "?")

	view := m.View().Content
	lines := strings.Split(view, "\n")
	if len(lines) != 10 {
		t.Fatalf("rendered frame is %d rows, want exactly 10 (the terminal's height)", len(lines))
	}
	// A row narrower than the terminal is fine — the canvas round-trip
	// trims trailing blank cells, same as any shorter line elsewhere in the
	// frame (bubbletea v2's cell-diffing renderer erases the rest, see
	// pty-verification.md) — but a row must never be *wider* than the
	// terminal: that is the overflow this test guards against.
	for i, l := range lines {
		if w := lipgloss.Width(l); w > 40 {
			t.Errorf("row %d is %d cells wide, want at most 40 (the terminal's width)", i, w)
		}
	}
}

// TestHelpUnreachableWhileFindEditing checks the precedence between the two
// exclusive input modes: an in-progress find query owns the keyboard first
// (model.Update checks findEditing before helpOpen), so "?" types into the
// query rather than opening Help.
func TestHelpUnreachableWhileFindEditing(t *testing.T) {
	m := openedFixture(t, 80, 24)
	m = press(t, m, "/")
	m = typeText(t, m, "?")
	if m.helpOpen {
		t.Error("Help opened while a find query had the keyboard")
	}
	if got := m.viewer.findState.input.Value(); got != "?" {
		t.Errorf("query = %q, want %q (the ? typed into it)", got, "?")
	}
}

// TestHelpUnreachableWhilePickerOpen is the Picker's own version of the
// check above (S2's "always typing" applies to it exactly the same way):
// "?" types into the Prompt rather than opening Help, since handlePickerInput
// runs before the Help binding is ever consulted (model.Update).
func TestHelpUnreachableWhilePickerOpen(t *testing.T) {
	m := openedFixture(t, 80, 24)
	m, _ = m.openPicker()
	m = typeText(t, m, "?")
	if m.helpOpen {
		t.Error("Help opened while the Picker had the keyboard")
	}
	if got := m.picker.input.Value(); got != "?" {
		t.Errorf("prompt = %q, want %q (the ? typed into it)", got, "?")
	}
}

// TestHelpBindingsListPickersOwnBindingsFromEveryPane is M2.3a's own
// acceptance criterion (D46/TRIAGE T5): the Help Modal is by design
// unreachable while the Picker owns the keyboard (handlePickerInput), so
// the three panes Help *is* reachable from, before the Picker opens, are
// where its own in-Picker actions (PickerUp/PickerDown/PickerMode/
// PickerScope) must be discoverable — not just keys.Picker, the key that
// opens it (already covered by TestHelpBindingsNameThePickerFromEveryPane).
// Closes the gap the footer's pickerHintText alone left open (fragile under
// a narrow terminal, with no Help fallback).
func TestHelpBindingsListPickersOwnBindingsFromEveryPane(t *testing.T) {
	own := []key.Binding{keys.PickerUp, keys.PickerDown, keys.PickerMode, keys.PickerScope}
	for _, p := range []pane{rootsPane, filesPane, viewerPane} {
		bindings := helpBindings(p)
		for _, want := range own {
			found := false
			for _, b := range bindings {
				if b.Help().Key == want.Help().Key && b.Help().Desc == want.Help().Desc {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("pane %v: helpBindings missing Picker's own binding %q (%q)", p, want.Help().Key, want.Help().Desc)
			}
		}
	}
}

// TestHelpBindingsNameThePickerFromEveryPane is helpBindings' half of D39's
// "M2.3 makes the [ctrl+p] binding... discoverable": the Help Modal's
// full binding list for every pane now names the Picker, the same
// always-available-regardless-of-focus treatment Edit/Navigator already
// get.
func TestHelpBindingsNameThePickerFromEveryPane(t *testing.T) {
	for _, p := range []pane{rootsPane, filesPane, viewerPane} {
		box := helpBox(p)
		if !strings.Contains(box, "picker") {
			t.Errorf("pane %v: help modal = %q, want it to name the picker", p, box)
		}
	}
}

// TestHelpBindingsNameFocusRefreshAndToggleHiddenFromEveryPane is C2a/M9.1's
// help-modal half of the same S6/S4 sweep TestPaneHintsNameFocusAndRefreshFromEveryPane
// checks for the footer: FocusRoots/FocusFiles/Refresh are global (S6), so
// every pane's Help Modal must list them; ToggleHidden only does anything
// once the Picker is open, but — like PickerUp/PickerDown/PickerMode/
// PickerScope before it — these three panes are the only place Help is ever
// reachable from to document it (see helpBindings' own comment).
func TestHelpBindingsNameFocusRefreshAndToggleHiddenFromEveryPane(t *testing.T) {
	want := []key.Binding{keys.FocusRoots, keys.FocusFiles, keys.Refresh, keys.ToggleHidden}
	for _, p := range []pane{rootsPane, filesPane, viewerPane} {
		bindings := helpBindings(p)
		for _, w := range want {
			found := false
			for _, b := range bindings {
				if b.Help().Key == w.Help().Key && b.Help().Desc == w.Help().Desc {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("pane %v: helpBindings missing %q (%q)", p, w.Help().Key, w.Help().Desc)
			}
		}
	}
}

// TestHelpBindingsListPeekOnlyForFilesPane checks S7's "Peek, in the Files
// Nav only" against the Help Modal: filesPane's binding list names it, but
// rootsPane's and viewerPane's do not, since Peek can never be triggered
// from either.
func TestHelpBindingsListPeekOnlyForFilesPane(t *testing.T) {
	filesBox := helpBox(filesPane)
	if !strings.Contains(filesBox, "peek") {
		t.Errorf("filesPane help modal = %q, want it to name peek", filesBox)
	}
	for _, p := range []pane{rootsPane, viewerPane} {
		box := helpBox(p)
		if strings.Contains(box, "peek") {
			t.Errorf("pane %v: help modal = %q, unexpectedly names Peek (Files-Nav-only)", p, box)
		}
	}
}

// TestNoHelpModalNamesCtrlP is the Help Modal's half of S9's literal "no
// remaining reference to ctrl+p anywhere".
func TestNoHelpModalNamesCtrlP(t *testing.T) {
	for _, p := range []pane{rootsPane, filesPane, viewerPane} {
		if box := helpBox(p); strings.Contains(box, "ctrl+p") {
			t.Errorf("pane %v: help modal = %q, unexpectedly still names ctrl+p", p, box)
		}
	}
}
