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
	if m.dv1.offset() != 0 {
		t.Fatalf("fixture did not open at the top: offset = %d", m.dv1.offset())
	}

	m = press(t, m, "/")
	if !m.dv1.findEditing() {
		t.Fatal("/ did not start a find")
	}
	m = typeText(t, m, "final line of the fixture")

	if got := m.dv1.offset(); got == 0 {
		t.Error("typing a query matching the document's tail did not scroll the viewport")
	}
	view := m.View().Content
	if !strings.Contains(view, fixtureStructure["tail"]) {
		t.Errorf("view does not show the matched tail phrase %q after searching", fixtureStructure["tail"])
	}
	// The matched line is indicated by reverse video (find.go's
	// applyCurrentMatch). hasReverse reads the cells, not a literal escape:
	// the frame's canvas pass (paintCanvas) folds reverse into a longer SGR.
	if !hasReverse(view) {
		t.Error("matched line is not indicated (no reverse-video escape in the view)")
	}

	m = press(t, m, "enter")
	if m.dv1.findEditing() {
		t.Error("enter did not commit the find (still editing)")
	}
	if !m.dv1.hasFindMatches() {
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

	n := len(m.dv1.findState.matches)
	if n < 2 {
		t.Fatalf(`fixture has %d match(es) for "line", want at least 2 to test traversal`, n)
	}
	start := m.dv1.findState.current

	m = press(t, m, "n")
	if got := m.dv1.findState.current; got != (start+1)%n {
		t.Errorf("after one next, current = %d, want %d", got, (start+1)%n)
	}

	// Advance to the last match, then one more next must wrap to the first.
	for i := 0; i < n-1; i++ {
		m = press(t, m, "n")
	}
	if got := m.dv1.findState.current; got != start {
		t.Fatalf("after n-1 more nexts, current = %d, want back at %d", got, start)
	}
	m = press(t, m, "n")
	if got := m.dv1.findState.current; got != (start+1)%n {
		t.Errorf("next past the last match did not wrap to the first: current = %d, want %d", got, (start+1)%n)
	}

	// And previous from there wraps back the other way.
	m = press(t, m, "N")
	if got := m.dv1.findState.current; got != start {
		t.Errorf("previous did not wrap back: current = %d, want %d", got, start)
	}
	m = press(t, m, "N")
	if got := m.dv1.findState.current; got != n-1 {
		t.Errorf("previous past the first match did not wrap to the last: current = %d, want %d", got, n-1)
	}
}

// TestFindIsSmartCase checks S8's "case-insensitive unless the query
// contains uppercase."
func TestFindIsSmartCase(t *testing.T) {
	m := openedFixture(t, 80, 24)

	m = press(t, m, "/")
	m = typeText(t, m, "final")
	if n := len(m.dv1.findState.matches); n != 1 {
		t.Fatalf(`lowercase "final" matched %d times, want 1`, n)
	}

	m = press(t, m, "esc")
	m = press(t, m, "/")
	m = typeText(t, m, "Final")
	if n := len(m.dv1.findState.matches); n != 0 {
		t.Errorf(`uppercase "Final" matched %d times against lowercase text, want 0`, n)
	}
}

// TestFindDismissalLeavesTheViewWhereItIs checks S8's "dismissal that leaves
// the view where it is": cancelling a search does not un-scroll it.
func TestFindDismissalLeavesTheViewWhereItIs(t *testing.T) {
	m := openedFixture(t, 80, 24)
	m = press(t, m, "/")
	m = typeText(t, m, "final line of the fixture")
	scrolledTo := m.dv1.offset()
	if scrolledTo == 0 {
		t.Fatal("query did not scroll the viewport; nothing for this test to check")
	}

	m = press(t, m, "esc")
	if m.dv1.findEditing() {
		t.Error("escape did not dismiss the find")
	}
	if m.dv1.hasFindMatches() {
		t.Error("escape left matches behind; a dismissed find should have none")
	}
	if got := m.dv1.offset(); got != scrolledTo {
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

// TestFindMatchesLineAttributionSurvivesFoldWithinTheMatch is T41: the case
// above covers a fold-length-changing rune preceding a match on the same
// line; this covers one the match itself spans, so the fold happens between
// hStart and hEnd rather than before either of them, exercising origSpan's
// translation of both ends of the same span rather than just offsets ahead
// of it.
func TestFindMatchesLineAttributionSurvivesFoldWithinTheMatch(t *testing.T) {
	const kelvin = "\u212A" // "K", folds to ASCII 'k' — 3 bytes down to 1
	content := "line one\n" + kelvin + "elvin"

	matches := findMatches(content, "kelvin")

	if len(matches) != 1 {
		t.Fatalf(`findMatches(%q, "kelvin") = %d matches, want 1`, content, len(matches))
	}
	m := matches[0]
	if m.line != 1 {
		t.Errorf("line = %d, want 1 (the match is on the second line)", m.line)
	}
	// stripped's own byte space: the Kelvin sign is 3 bytes, so colEnd must
	// land at 8 (3 + len("elvin")), not 6 (len("kelvin") in the folded,
	// byte-length-1 space) — the exact drift origSpan exists to prevent.
	if m.colStart != 0 || m.colEnd != 8 {
		t.Errorf("colStart, colEnd = %d, %d, want 0, 8 (in stripped's byte space, not the folded haystack's)", m.colStart, m.colEnd)
	}
}

// TestFindingWithNoDocumentOpenReportsIt mirrors startEdit's shape (M4.1):
// a global action with nothing to act on says so instead of doing nothing.
func TestFindingWithNoDocumentOpenReportsIt(t *testing.T) {
	m := sized(t, newModel(), 80, 24)
	m = press(t, m, "/")
	if m.dv1.findEditing() {
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
	if !m.dv1.findEditing() {
		t.Fatal("test setup: find is not editing")
	}

	next, _ := m.Update(tea.WindowSizeMsg{Width: 60, Height: 20})
	m = concrete(t, next)

	if m.dv1.findEditing() {
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
	if m.dv1.findEditing() {
		t.Fatal("test setup: find should not be editing")
	}

	next, _ := m.Update(tea.WindowSizeMsg{Width: 60, Height: 20})
	m = concrete(t, next)

	if m.message != "" {
		t.Errorf("message = %q, want no message from a resize with no find in progress", m.message)
	}
}

// TestWindowResizeAfterCommittedFindSurfacesMessage is T43: unlike the
// mid-edit case T34 covers, committing a find (Enter) clears findEditing
// but keeps its matches and indication (commitFind's doc comment) for
// FindNext/FindPrevious to keep using. A resize's re-render drops those
// matches exactly as it would an in-progress query, but T34's guard alone
// misses it since editing was already false — this must be caught too,
// not silently dropped.
func TestWindowResizeAfterCommittedFindSurfacesMessage(t *testing.T) {
	m := openedFixture(t, 80, 24)
	m = press(t, m, "/")
	m = typeText(t, m, "final")
	m = press(t, m, "enter")
	if m.dv1.findEditing() {
		t.Fatal("test setup: enter did not commit the find")
	}
	if !m.dv1.hasFindMatches() {
		t.Fatal("test setup: committed find has no matches to lose")
	}

	next, _ := m.Update(tea.WindowSizeMsg{Width: 60, Height: 20})
	m = concrete(t, next)

	if m.dv1.hasFindMatches() {
		t.Fatal("test setup: resize did not actually invalidate the committed find (width unchanged?)")
	}
	if !strings.Contains(m.message, "resize") && !strings.Contains(m.message, "search") {
		t.Errorf("message = %q, want it to explain the cleared search", m.message)
	}
}

// TestNavigatorToggleAfterCommittedFindNamesTheToggleNotAResize is T45: T43's
// resize guard above catches a committed find lost to toggleNavigator (it
// drives the very same resize path), but until this fix it did so by always
// saying "window resized" even when the real trigger was ctrl+n — a
// factually wrong footer message that violates SPEC.md's "confusion is not
// allowed" worse than silence would. This drives the committed find through
// the Navigator toggle itself, not a tea.WindowSizeMsg, and asserts the
// message names the sidebar toggle rather than a resize.
func TestNavigatorToggleAfterCommittedFindNamesTheToggleNotAResize(t *testing.T) {
	m := openedFixture(t, 80, 24)
	m = press(t, m, "/")
	m = typeText(t, m, "final")
	m = press(t, m, "enter")
	if m.dv1.findEditing() {
		t.Fatal("test setup: enter did not commit the find")
	}
	if !m.dv1.hasFindMatches() {
		t.Fatal("test setup: committed find has no matches to lose")
	}

	m = press(t, m, keys.Navigator.Keys()[0])

	if m.dv1.hasFindMatches() {
		t.Fatal("test setup: toggling the Navigator did not actually invalidate the committed find (width unchanged?)")
	}
	if strings.Contains(m.message, "resize") {
		t.Errorf("message = %q, claims a resize happened, but the trigger was ctrl+n", m.message)
	}
	if !strings.Contains(m.message, "sidebar") && !strings.Contains(m.message, "toggle") {
		t.Errorf("message = %q, want it to name the sidebar toggle as the actual cause", m.message)
	}
}

// TestWindowResizeDropsDV2FindOnlyNamesDV2 is C3/M1.6a: resize's
// find-dropped message must extend to DV2 symmetrically (S2 parity). With
// DV1 unaffected (no find active there), a drop in DV2 alone must name dv2
// explicitly rather than reusing DV1's unprefixed wording, which would
// otherwise mislead about which viewer actually lost its search.
func TestWindowResizeDropsDV2FindOnlyNamesDV2(t *testing.T) {
	m, _, _ := crossViewerFixture(t) // focus on dv2Pane
	m = press(t, m, "/")
	m = typeText(t, m, "bravo-marker")
	m = press(t, m, "enter")
	if !m.dv2.hasFindMatches() {
		t.Fatal("test setup: DV2's find did not commit")
	}
	if m.dv1.findEditing() || m.dv1.hasFindMatches() {
		t.Fatal("test setup: DV1 unexpectedly has find state")
	}

	next, _ := m.Update(tea.WindowSizeMsg{Width: 60, Height: 10})
	m = concrete(t, next)

	if m.dv2.hasFindMatches() {
		t.Fatal("test setup: resize did not actually invalidate DV2's find (width unchanged?)")
	}
	if !strings.Contains(m.message, "dv2") {
		t.Errorf("message = %q, want it to name dv2 as the viewer that lost its search", m.message)
	}
	if strings.Contains(m.message, "dv1") {
		t.Errorf("message = %q, wrongly names dv1 (its find was never active)", m.message)
	}
}

// TestWindowResizeDroppingBothViewersFindsNamesBoth is C3/M1.6a's headline
// case: distinct committed finds in both DV1 and DV2, both lost to the same
// resize, must combine into the single m.message slot naming both drops
// rather than reporting only one or one overwriting the other.
func TestWindowResizeDroppingBothViewersFindsNamesBoth(t *testing.T) {
	m, _, _ := crossViewerFixture(t) // focus on dv2Pane
	m = press(t, m, "/")
	m = typeText(t, m, "bravo-marker")
	m = press(t, m, "enter")
	if !m.dv2.hasFindMatches() {
		t.Fatal("test setup: DV2's find did not commit")
	}

	m = press(t, m, keys.FocusDV1.Keys()[0])
	m = press(t, m, "/")
	m = typeText(t, m, "alpha-marker")
	m = press(t, m, "enter")
	if !m.dv1.hasFindMatches() {
		t.Fatal("test setup: DV1's find did not commit")
	}

	next, _ := m.Update(tea.WindowSizeMsg{Width: 60, Height: 10})
	m = concrete(t, next)

	if m.dv1.hasFindMatches() || m.dv2.hasFindMatches() {
		t.Fatal("test setup: resize did not actually invalidate both finds (width unchanged?)")
	}
	if !strings.Contains(m.message, "dv1") || !strings.Contains(m.message, "dv2") {
		t.Errorf("message = %q, want it to name both dv1 and dv2 as having lost their search", m.message)
	}
}

// TestFindQueryCapturesEveryKeyWhileEditing checks that a query character
// which is also a binding elsewhere (q, the quit key) types into the query
// instead of acting on its binding.
func TestFindQueryCapturesEveryKeyWhileEditing(t *testing.T) {
	m := openedFixture(t, 80, 24)
	m = press(t, m, "/")
	m = typeText(t, m, "q")
	if !m.dv1.findEditing() {
		t.Fatal("typing q while editing a find query exited find mode")
	}
	if got := m.dv1.findState.input.Value(); got != "q" {
		t.Errorf("query = %q, want %q", got, "q")
	}
}
