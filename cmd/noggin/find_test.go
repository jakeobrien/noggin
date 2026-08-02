package main

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// TestFindLocatesAPhraseNearTheEndOfTheDocument is M5.1's headline
// acceptance check: a phrase in the last screen of the fixture is brought on
// screen and its line indicated, from a cold start at the top.
func TestFindLocatesAPhraseNearTheEndOfTheDocument(t *testing.T) {
	m := openedFixture(t, 80, 24)
	if m.viewer.offset() != 0 {
		t.Fatalf("fixture did not open at the top: offset = %d", m.viewer.offset())
	}

	m = press(t, m, "/")
	if !m.viewer.findEditing() {
		t.Fatal("/ did not start a find")
	}
	m = typeText(t, m, "final line of the fixture")

	if got := m.viewer.offset(); got == 0 {
		t.Error("typing a query matching the document's tail did not scroll the viewport")
	}
	view := m.View().Content
	if !strings.Contains(view, fixtureStructure["tail"]) {
		t.Errorf("view does not show the matched tail phrase %q after searching", fixtureStructure["tail"])
	}
	// The matched line is indicated by reverse video (find.go's
	// applyCurrentMatch); lipgloss emits "\x1b[7m" for Reverse(true)
	// regardless of TTY (verified directly against this lipgloss build).
	if !strings.Contains(view, "\x1b[7m") {
		t.Error("matched line is not indicated (no reverse-video escape in the view)")
	}

	m = press(t, m, "enter")
	if m.viewer.findEditing() {
		t.Error("enter did not commit the find (still editing)")
	}
	if !m.viewer.hasFindMatches() {
		t.Error("committing the find lost its matches")
	}
}

// TestFindTraversalWraps checks next/previous match traversal, including
// wrapping past either end.
func TestFindTraversalWraps(t *testing.T) {
	m := openedFixture(t, 80, 24)
	m = press(t, m, "/")
	m = typeText(t, m, "line")
	m = press(t, m, "enter")

	n := len(m.viewer.findState.matches)
	if n < 2 {
		t.Fatalf(`fixture has %d match(es) for "line", want at least 2 to test traversal`, n)
	}
	start := m.viewer.findState.current

	m = press(t, m, "n")
	if got := m.viewer.findState.current; got != (start+1)%n {
		t.Errorf("after one next, current = %d, want %d", got, (start+1)%n)
	}

	// Advance to the last match, then one more next must wrap to the first.
	for i := 0; i < n-1; i++ {
		m = press(t, m, "n")
	}
	if got := m.viewer.findState.current; got != start {
		t.Fatalf("after n-1 more nexts, current = %d, want back at %d", got, start)
	}
	m = press(t, m, "n")
	if got := m.viewer.findState.current; got != (start+1)%n {
		t.Errorf("next past the last match did not wrap to the first: current = %d, want %d", got, (start+1)%n)
	}

	// And previous from there wraps back the other way.
	m = press(t, m, "N")
	if got := m.viewer.findState.current; got != start {
		t.Errorf("previous did not wrap back: current = %d, want %d", got, start)
	}
	m = press(t, m, "N")
	if got := m.viewer.findState.current; got != n-1 {
		t.Errorf("previous past the first match did not wrap to the last: current = %d, want %d", got, n-1)
	}
}

// TestFindIsSmartCase checks S8's "case-insensitive unless the query
// contains uppercase."
func TestFindIsSmartCase(t *testing.T) {
	m := openedFixture(t, 80, 24)

	m = press(t, m, "/")
	m = typeText(t, m, "final")
	if n := len(m.viewer.findState.matches); n != 1 {
		t.Fatalf(`lowercase "final" matched %d times, want 1`, n)
	}

	m = press(t, m, "esc")
	m = press(t, m, "/")
	m = typeText(t, m, "Final")
	if n := len(m.viewer.findState.matches); n != 0 {
		t.Errorf(`uppercase "Final" matched %d times against lowercase text, want 0`, n)
	}
}

// TestFindDismissalLeavesTheViewWhereItIs checks S8's "dismissal that leaves
// the view where it is": cancelling a search does not un-scroll it.
func TestFindDismissalLeavesTheViewWhereItIs(t *testing.T) {
	m := openedFixture(t, 80, 24)
	m = press(t, m, "/")
	m = typeText(t, m, "final line of the fixture")
	scrolledTo := m.viewer.offset()
	if scrolledTo == 0 {
		t.Fatal("query did not scroll the viewport; nothing for this test to check")
	}

	m = press(t, m, "esc")
	if m.viewer.findEditing() {
		t.Error("escape did not dismiss the find")
	}
	if m.viewer.hasFindMatches() {
		t.Error("escape left matches behind; a dismissed find should have none")
	}
	if got := m.viewer.offset(); got != scrolledTo {
		t.Errorf("dismissing the find moved the viewport: offset = %d, want %d (where the search left it)", got, scrolledTo)
	}
}

// TestFindMatchesLineAttributionSurvivesCaseFoldLengthChange is T29:
// findMatches' smart-case path lower-cases the haystack to search it, and
// case folding is not always byte-length-preserving — the Kelvin sign
// U+212A (three bytes) folds to the single-byte 'k'. A naive byte offset
// found in the folded string, treated as if it were the original's own
// offset, drifts once such a rune precedes a match on the same line, and
// can misattribute the match to the wrong line entirely (not just the wrong
// column) once the drift crosses a newline boundary — exactly what this
// fixture forces: the Kelvin sign sits alone on line 0, so the drift pulls
// the reported position back across the "\n" into line 0 unless findMatches
// maps offsets back to the original string correctly.
func TestFindMatchesLineAttributionSurvivesCaseFoldLengthChange(t *testing.T) {
	const kelvin = "\u212A" // "K", folds to ASCII 'k' — 3 bytes down to 1
	content := kelvin + "\nmatch"

	matches := findMatches(content, "match")

	if len(matches) != 1 {
		t.Fatalf(`findMatches(%q, "match") = %d matches, want 1`, content, len(matches))
	}
	m := matches[0]
	if m.line != 1 {
		t.Errorf("line = %d, want 1 (the match is on the line after the Kelvin sign)", m.line)
	}
	if m.colStart != 0 || m.colEnd != 5 {
		t.Errorf("colStart, colEnd = %d, %d, want 0, 5", m.colStart, m.colEnd)
	}
}

// TestFindingWithNoDocumentOpenReportsIt mirrors startEdit's shape (M4.1):
// a global action with nothing to act on says so instead of doing nothing.
func TestFindingWithNoDocumentOpenReportsIt(t *testing.T) {
	m := sized(t, newModel(), 80, 24)
	m = press(t, m, "/")
	if m.viewer.findEditing() {
		t.Error("find started with no document open")
	}
	if !strings.Contains(m.message, "no document open") {
		t.Errorf("message = %q, want it to report no document open", m.message)
	}
}

// TestWindowResizeMidFindEditSurfacesMessage is T34: a real terminal resize
// (tea.WindowSizeMsg) that changes the viewer's width re-renders it, which
// re-wraps the document and so invalidates any find in progress (viewer.go's
// render doc comment) — including one still being typed. Silently dropping
// it back then would be the only lose-state path in the program with no
// footer message; model.resize must report it instead.
func TestWindowResizeMidFindEditSurfacesMessage(t *testing.T) {
	m := openedFixture(t, 80, 24)
	m = press(t, m, "/")
	m = typeText(t, m, "final")
	if !m.viewer.findEditing() {
		t.Fatal("test setup: find is not editing")
	}

	next, _ := m.Update(tea.WindowSizeMsg{Width: 60, Height: 20})
	m = concrete(t, next)

	if m.viewer.findEditing() {
		t.Fatal("test setup: resize did not actually invalidate the find (width unchanged?)")
	}
	if !strings.Contains(m.message, "resize") && !strings.Contains(m.message, "search") {
		t.Errorf("message = %q, want it to explain the dismissed search", m.message)
	}
}

// TestWindowResizeWithNoFindInProgressLeavesMessageAlone checks the guard on
// the fix above does not fire when there was nothing to lose: a resize with
// no find in progress must not manufacture a message out of nothing.
func TestWindowResizeWithNoFindInProgressLeavesMessageAlone(t *testing.T) {
	m := openedFixture(t, 80, 24)
	if m.viewer.findEditing() {
		t.Fatal("test setup: find should not be editing")
	}

	next, _ := m.Update(tea.WindowSizeMsg{Width: 60, Height: 20})
	m = concrete(t, next)

	if m.message != "" {
		t.Errorf("message = %q, want no message from a resize with no find in progress", m.message)
	}
}

// TestFindQueryCapturesEveryKeyWhileEditing checks that a query character
// which is also a binding elsewhere (q, the quit key) types into the query
// instead of acting on its binding.
func TestFindQueryCapturesEveryKeyWhileEditing(t *testing.T) {
	m := openedFixture(t, 80, 24)
	m = press(t, m, "/")
	m = typeText(t, m, "q")
	if !m.viewer.findEditing() {
		t.Fatal("typing q while editing a find query exited find mode")
	}
	if got := m.viewer.findState.input.Value(); got != "q" {
		t.Errorf("query = %q, want %q", got, "q")
	}
}
