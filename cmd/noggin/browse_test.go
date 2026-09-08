package main

import (
	"strings"
	"testing"
)

// TestBrowseModeCyclesTreeRecentStarred is M4.1's own acceptance wording at
// the filesNav level: keys.FilesMode cycles Tree -> Recent -> Starred -> Tree,
// wrapping indefinitely.
func TestBrowseModeCyclesTreeRecentStarred(t *testing.T) {
	f := newFilesNav()
	if f.mode != treeBrowse {
		t.Fatalf("initial mode = %v, want treeBrowse", f.mode)
	}
	f = f.cycleMode()
	if f.mode != recentBrowse {
		t.Fatalf("mode after one cycle = %v, want recentBrowse", f.mode)
	}
	f = f.cycleMode()
	if f.mode != starredBrowse {
		t.Fatalf("mode after two cycles = %v, want starredBrowse", f.mode)
	}
	f = f.cycleMode()
	if f.mode != treeBrowse {
		t.Fatalf("mode after three cycles = %v, want treeBrowse (wrapped)", f.mode)
	}
}

// TestBrowseScopeIsANoOpInTreeMode checks S5's "Tree ... the scope key is a
// no-op" directly: cycling scope while the mode is Tree changes nothing.
func TestBrowseScopeIsANoOpInTreeMode(t *testing.T) {
	f := newFilesNav()
	before := f.scope
	f = f.cycleScope()
	if f.scope != before || f.mode != treeBrowse {
		t.Errorf("cycling scope in Tree mode changed state: scope %v -> %v, mode %v", before, f.scope, f.mode)
	}
	if got := f.scopeLabel(); got != "-" {
		t.Errorf("Tree scope label = %q, want \"-\"", got)
	}
}

// TestBrowseScopeCyclesRecentsTwoValues checks Recent's own scope set (S5:
// "Active Source | All Sources"), wrapping after two values.
func TestBrowseScopeCyclesRecentsTwoValues(t *testing.T) {
	f := newFilesNav().cycleMode() // Tree -> Recent
	if got := f.scopeLabel(); got != "Active Source" {
		t.Fatalf("Recent's starting scope = %q, want Active Source", got)
	}
	f = f.cycleScope()
	if got := f.scopeLabel(); got != "All Sources" {
		t.Errorf("Recent's scope after one cycle = %q, want All Sources", got)
	}
	f = f.cycleScope()
	if got := f.scopeLabel(); got != "Active Source" {
		t.Errorf("Recent's scope after two cycles = %q, want Active Source (wrapped)", got)
	}
}

// TestBrowseScopeCyclesStarredsThreeValues checks Starred's own scope set
// (S5: "Active Source | All Starred | Orphans"), wrapping after three values.
func TestBrowseScopeCyclesStarredsThreeValues(t *testing.T) {
	f := newFilesNav().cycleMode().cycleMode() // Tree -> Recent -> Starred
	want := []string{"Active Source", "All Starred", "Orphans", "Active Source"}
	for i, w := range want {
		if got := f.scopeLabel(); got != w {
			t.Errorf("Starred's scope at step %d = %q, want %q", i, got, w)
		}
		f = f.cycleScope()
	}
}

// TestBrowseModeSwitchResetsScopeToTheNewModesFirstValue checks that a scope
// index carried from one mode is never read against a different mode's own
// value set: switching from Recent's second scope (All Sources) to Starred
// lands on Starred's own first value (Active Source), not on Starred's
// same-index (All Starred) or a stale label.
func TestBrowseModeSwitchResetsScopeToTheNewModesFirstValue(t *testing.T) {
	f := newFilesNav().cycleMode() // Tree -> Recent
	f = f.cycleScope()             // Active Source -> All Sources
	if got := f.scopeLabel(); got != "All Sources" {
		t.Fatalf("setup: Recent's scope = %q, want All Sources", got)
	}
	f = f.cycleMode() // Recent -> Starred
	if got := f.scopeLabel(); got != "Active Source" {
		t.Errorf("Starred's scope right after switching from Recent/All Sources = %q, want Active Source", got)
	}
}

// TestFilesModeScopeRowShowsTheCurrentPair drives filesNav.view() directly,
// checking the mode/scope row's own text for each mode.
func TestFilesModeScopeRowShowsTheCurrentPair(t *testing.T) {
	f := newFiles(fixtureTree()).setSize(30, 8)
	cases := []struct {
		mode  string
		scope string
	}{
		{"Tree", "-"},
		{"Recent", "Active Source"},
		{"Starred", "Active Source"},
	}
	for _, c := range cases {
		lines := strings.Split(f.view(), "\n")
		if !strings.Contains(lines[0], c.mode) || !strings.Contains(lines[0], c.scope) {
			t.Errorf("mode/scope row = %q, want it to show %q and %q", lines[0], c.mode, c.scope)
		}
		f = f.cycleMode()
	}
}

// TestFilesModeScopeRowCostsOneRowOfHeight is S5's own "costing the file
// list one row of height in every mode" clause: the view is still exactly
// f.height lines (the box's own contract), but the row list beneath the
// mode/scope row only gets height-1 of them, so a box just tall enough for
// N rows plus the header shows N-1 rows once populated with N rows worth of
// files, not N.
func TestFilesModeScopeRowCostsOneRowOfHeight(t *testing.T) {
	f := newFiles(fixtureTree()).setSize(30, 5) // 4 top-level rows, height 5
	lines := strings.Split(f.view(), "\n")
	if len(lines) != 5 {
		t.Fatalf("view has %d lines, want 5 (the box height)", len(lines))
	}
	// lines[0] is the mode/scope row; only 4 lines remain for up to 4 files,
	// which happens to fit exactly here, so grow the box down to prove the
	// row really is reserved rather than just coincidentally unused.
	fSmaller := newFiles(fixtureTree()).setSize(30, 4)
	linesSmaller := strings.Split(fSmaller.view(), "\n")
	if len(linesSmaller) != 4 {
		t.Fatalf("smaller view has %d lines, want 4", len(linesSmaller))
	}
	// With only 3 rows left for 4 files, the last (notes.txt) must be
	// scrolled off rather than shown — proof the row list only got
	// height-1, not height.
	if strings.Contains(linesSmaller[3], "notes.txt") {
		t.Error("all 4 files visible in a 3-row list; want the mode/scope row to have actually cost a row")
	}
}

// TestFilesNavViewHeightOneOmitsRowList is TRIAGE T25's own gap (D83):
// view()'s own doc comment describes a height-1 short-circuit ("No room
// left for the list past the mode/scope row itself") that every sibling
// branch of rowsView/flatRowsView already has a dedicated test for, but
// this one branch of view() itself did not — a height of exactly 1 must
// render only the mode/scope row, with no trailing newline appended for an
// empty row list underneath it.
func TestFilesNavViewHeightOneOmitsRowList(t *testing.T) {
	f := newFiles(fixtureTree()).setSize(30, 1)
	view := f.view()
	if strings.Contains(view, "\n") {
		t.Errorf("view at height 1 = %q, want a single line with no newline", view)
	}
	if view != f.modeScopeRow() {
		t.Errorf("view at height 1 = %q, want exactly the mode/scope row %q", view, f.modeScopeRow())
	}
}

// TestFilesModeScopeKeysActOnlyFromFilesNavFocus is the model-level half of
// M4.1's acceptance criterion: the two cycle keys do nothing from any Focus
// other than the Files Nav.
func TestFilesModeScopeKeysActOnlyFromFilesNavFocus(t *testing.T) {
	m := sized(t, newModel().openFolder(fixtureTree()), 80, 24)
	m = press(t, m, "r") // focus -> Roots Nav
	if m.focus != rootsPane {
		t.Fatalf("setup: focus = %v, want rootsPane", m.focus)
	}
	m = press(t, m, keys.FilesMode.Keys()[0])
	m = press(t, m, keys.FilesScope.Keys()[0])
	if m.files.mode != treeBrowse || m.files.scope != activeRootBrowseScope {
		t.Errorf("Files Nav browse state changed while Roots Nav had focus: mode %v, scope %v", m.files.mode, m.files.scope)
	}
}

// TestFilesModeScopeKeysCycleFromFilesNavFocus is the mirror check: from
// Files Nav Focus, the same two keys do act, and the rendered frame reflects
// the new pair immediately.
func TestFilesModeScopeKeysCycleFromFilesNavFocus(t *testing.T) {
	m := sized(t, newModel().openFolder(fixtureTree()), 80, 24)
	if m.focus != filesPane {
		t.Fatalf("setup: focus = %v, want filesPane", m.focus)
	}
	m = press(t, m, keys.FilesMode.Keys()[0])
	if m.files.mode != recentBrowse {
		t.Fatalf("mode after FilesMode = %v, want recentBrowse", m.files.mode)
	}
	if !strings.Contains(m.frame(), "Recent") {
		t.Errorf("frame does not show the new mode:\n%s", m.frame())
	}
	m = press(t, m, keys.FilesScope.Keys()[0])
	if m.files.scope != allRootsBrowseScope {
		t.Fatalf("scope after FilesScope = %v, want allRootsBrowseScope", m.files.scope)
	}
	if !strings.Contains(m.frame(), "All Sources") {
		t.Errorf("frame does not show the new scope:\n%s", m.frame())
	}
}
