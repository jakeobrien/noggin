package main

import (
	"path/filepath"
	"testing"
)

// TestActivateWikipediaSourceWithZeroRootsTolerated is C6/M1.1's acceptance
// criterion: activating the Wikipedia Source with zero Roots renders every
// pane and runs Refresh without panic or error, and every Root-only
// subsystem (tree walk, grep, git markers, tree staleness, Refresh) treats
// it exactly like the "no active Root" state those subsystems already
// tolerated before Wikipedia existed.
func TestActivateWikipediaSourceWithZeroRootsTolerated(t *testing.T) {
	m := newModel()
	if got := m.roots.activeIsWikipedia(); got {
		t.Fatalf("fresh model already reports Wikipedia active, want false")
	}

	m = m.activateWikipediaSource()
	if !m.roots.activeIsWikipedia() {
		t.Fatal("activateWikipediaSource did not mark Wikipedia as the Active Source")
	}
	if m.roots.active != -1 {
		t.Errorf("roots.active = %d, want -1 (no Root active while Wikipedia is)", m.roots.active)
	}
	if got := m.roots.activePath(); got != "" {
		t.Errorf("activePath() = %q, want \"\" while Wikipedia is active", got)
	}

	// Render every pane at a realistic size without panicking.
	m = sized(t, m, 80, 24)
	view := m.View().Content
	if view == "" {
		t.Fatal("View() rendered nothing after activating the Wikipedia Source")
	}

	// Refresh must not panic or set an error message: there is no active
	// Root's tree to walk, so the tree-walk branch is a no-op, exactly as it
	// already is for a Root-less session.
	m = m.refresh()
	if m.message != "" {
		t.Errorf("refresh() message = %q, want \"\" (nothing to refresh)", m.message)
	}
}

// TestActivateRootDisplacesWikipediaSource checks the mirror direction: once
// a Root is activated, it -- not Wikipedia -- is the Active Source again,
// since the two are mutually exclusive.
func TestActivateRootDisplacesWikipediaSource(t *testing.T) {
	dir := mkdir(t, filepath.Join(t.TempDir(), "root"))
	m := newModel()
	m.roots = m.roots.setRoots([]root{{path: dir}}, -1)
	m = m.activateWikipediaSource()
	if !m.roots.activeIsWikipedia() {
		t.Fatal("activateWikipediaSource did not take effect")
	}

	m = m.activateRoot()
	if m.roots.activeIsWikipedia() {
		t.Error("activateRoot left Wikipedia marked as the Active Source")
	}
	if m.roots.active != m.roots.selected {
		t.Errorf("roots.active = %d, want %d (the activated Root)", m.roots.active, m.roots.selected)
	}
}
