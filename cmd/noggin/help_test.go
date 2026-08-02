package main

import (
	"strings"
	"testing"

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
	m := sized(t, newModel().openFolder(fixtureTree()), 80, 24)
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
