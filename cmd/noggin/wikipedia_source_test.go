package main

import (
	"path/filepath"
	"strings"
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
	m, _ = m.refresh()
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

// TestActivateWikipediaSourceSetsFilesNavToWikipediaStarred is C6/M1.4's
// model-level wiring check: activateWikipediaSource puts the Files Nav's
// tree aside (root == nil) and lands its Browse Mode on Starred, the same
// pair filesNav.setWikipedia gives directly (wikipedia_browse_test.go),
// visible in the rendered frame too.
func TestActivateWikipediaSourceSetsFilesNavToWikipediaStarred(t *testing.T) {
	m := sized(t, newModel(), 80, 24)
	m = m.activateWikipediaSource()
	if !m.files.wikipedia {
		t.Fatal("activateWikipediaSource did not mark the Files Nav as showing Wikipedia")
	}
	if m.files.root != nil {
		t.Error("activateWikipediaSource left the Files Nav's tree in place, want it set aside")
	}
	if m.files.mode != starredBrowse {
		t.Errorf("Files Nav mode after activateWikipediaSource = %v, want starredBrowse", m.files.mode)
	}
	if !strings.Contains(m.frame(), "Starred") {
		t.Errorf("frame does not show the Starred mode:\n%s", m.frame())
	}
}

// TestActivateRootAfterWikipediaRestoresTreeMode is the mirror direction:
// once a Root displaces Wikipedia as the Active Source (S1), the Files Nav
// is back on Tree -- a Root the pane has never shown starts fresh, exactly
// as it would from a brand new session (filesNav.setRoot's own contract).
func TestActivateRootAfterWikipediaRestoresTreeMode(t *testing.T) {
	dir := mkdir(t, filepath.Join(t.TempDir(), "root"))
	m := sized(t, newModel(), 80, 24)
	m.roots = m.roots.setRoots([]root{{path: dir}}, -1)

	m = m.activateWikipediaSource()
	m = m.activateRoot()

	if m.files.wikipedia {
		t.Error("Files Nav still marked as showing Wikipedia after activateRoot")
	}
	if m.files.mode != treeBrowse {
		t.Errorf("Files Nav mode after activateRoot = %v, want treeBrowse", m.files.mode)
	}
}
