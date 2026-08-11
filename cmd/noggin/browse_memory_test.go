package main

import "testing"

// TestBrowseMemoryFreshSessionStartsOnTree is S7's own "session-only ...
// every session starts on Tree" clause: a brand new pane, never having set
// a Root at all, is already on Tree with no scope memory to speak of.
func TestBrowseMemoryFreshSessionStartsOnTree(t *testing.T) {
	f := newFilesNav()
	if f.mode != treeBrowse {
		t.Fatalf("fresh filesNav mode = %v, want treeBrowse", f.mode)
	}
	if got := f.scopeLabel(); got != "-" {
		t.Errorf("fresh filesNav scope label = %q, want \"-\"", got)
	}
}

// TestBrowseMemoryModePerRootScopePerRootMode is M4.6's own acceptance
// wording: distinct mode/scope pairs set on two Roots survive switching away
// and back to each, independently. Root A is left on (Recent, All Roots),
// Root B on (Starred, Orphans) — different modes and different scope values,
// so a bug that remembered only one axis, or shared scope across modes,
// would show up as either pair coming back wrong.
func TestBrowseMemoryModePerRootScopePerRootMode(t *testing.T) {
	rootA, rootB := fixtureTree(), fixtureTreeB()

	f := newFilesNav().setRoot(rootA)
	f = f.cycleMode()  // Tree -> Recent
	f = f.cycleScope() // Active Root -> All Roots
	if f.mode != recentBrowse || f.scope != allRootsBrowseScope {
		t.Fatalf("setup: Root A's pair = (%v, %v), want (Recent, All Roots)", f.mode, f.scope)
	}

	// Root B has never been shown: starts fresh on Tree, not carrying A's
	// pair over by accident (S7's per-session-start rule applies per Root's
	// first visit too).
	f = f.setRoot(rootB)
	if f.mode != treeBrowse {
		t.Fatalf("Root B's first visit mode = %v, want treeBrowse", f.mode)
	}
	f = f.cycleMode()  // Tree -> Recent
	f = f.cycleMode()  // Recent -> Starred
	f = f.cycleScope() // Active Root -> All Starred
	f = f.cycleScope() // All Starred -> Orphans
	if f.mode != starredBrowse || f.scope != orphansBrowseScope {
		t.Fatalf("setup: Root B's pair = (%v, %v), want (Starred, Orphans)", f.mode, f.scope)
	}

	// Back to Root A: its own pair, not B's, and not Recent's own first
	// scope value either — the actual value A was left on.
	f = f.setRoot(rootA)
	if f.mode != recentBrowse {
		t.Errorf("Root A's mode after A\u2192B\u2192A = %v, want recentBrowse", f.mode)
	}
	if f.scope != allRootsBrowseScope {
		t.Errorf("Root A's scope after A\u2192B\u2192A = %v, want allRootsBrowseScope", f.scope)
	}

	// And Root B again: its own pair, unaffected by A's round trip.
	f = f.setRoot(rootB)
	if f.mode != starredBrowse {
		t.Errorf("Root B's mode after A\u2192B\u2192A\u2192B = %v, want starredBrowse", f.mode)
	}
	if f.scope != orphansBrowseScope {
		t.Errorf("Root B's scope after A\u2192B\u2192A\u2192B = %v, want orphansBrowseScope", f.scope)
	}
}

// TestBrowseMemoryScopePerModeSurvivesAThirdModeVisit checks the "per
// (Root, mode)" half specifically: visiting a second mode on the same Root
// and changing its scope must not disturb the first mode's own remembered
// scope on that same Root — scopeByMode is keyed by mode, not overwritten
// wholesale on every cycleMode.
func TestBrowseMemoryScopePerModeSurvivesAThirdModeVisit(t *testing.T) {
	rootA, rootB := fixtureTree(), fixtureTreeB()

	f := newFilesNav().setRoot(rootA)
	f = f.cycleMode()  // Tree -> Recent
	f = f.cycleScope() // Active Root -> All Roots (Recent's own scope on A)
	f = f.cycleMode()  // Recent -> Starred
	f = f.cycleScope() // Active Root -> All Starred (Starred's own scope on A)

	// Leave and return to Root A without ever switching mode again here —
	// setRoot alone is what must round-trip both modes' scopes.
	f = f.setRoot(rootB)
	f = f.setRoot(rootA)

	if f.mode != starredBrowse || f.scope != allStarredBrowseScope {
		t.Fatalf("Root A's current pair after the round trip = (%v, %v), want (Starred, All Starred)", f.mode, f.scope)
	}
	f = f.cycleMode() // Starred -> Tree
	f = f.cycleMode() // Tree -> Recent
	if f.scope != allRootsBrowseScope {
		t.Errorf("Root A's Recent scope after visiting Starred = %v, want allRootsBrowseScope (unchanged)", f.scope)
	}
}
