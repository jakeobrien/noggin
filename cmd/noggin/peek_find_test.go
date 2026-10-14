package main

import (
	"path/filepath"
	"testing"

	"noggin/internal/tree"
)

// TestFindNextPreviousAreNoOpsWhilePeeking checks C2a/M7.3's fix (D51,
// TRIAGE T13/T14/T16): a pre-existing committed Find on the Active Document
// must not let n/N drift its viewport while Peeking a different file — the
// gap D51 describes, where the real viewport's YOffset could silently move
// even though the screen shows peek.vp instead — and clearing Peek
// afterward must restore the exact pre-Peek scroll offset and match index,
// not a drifted one.
func TestFindNextPreviousAreNoOpsWhilePeeking(t *testing.T) {
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	activePath := writeFileContent(t, filepath.Join(root, "active.md"), listContent(60))
	writeFileContent(t, filepath.Join(root, "peeked.md"), "# Peeked\n")

	node, err := tree.Walk(root, tree.FilterMarkdown)
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	m := sized(t, newModel().openFolder(node), 80, 10)

	d, err := loadDocument(activePath)
	if err != nil {
		t.Fatalf("loadDocument(active): %v", err)
	}
	m = m.openDocument(d)

	m = press(t, m, "/")
	m = typeText(t, m, "item 5")
	m = press(t, m, "enter")
	if !m.dv1.hasFindMatches() {
		t.Fatalf("setup: committed find has no matches")
	}
	n := len(m.dv1.findState.matches)
	if n < 2 {
		t.Fatalf(`setup: only %d match(es) for "item 5", want at least 2 to test traversal`, n)
	}
	wantOffset := m.dv1.offset()
	wantCurrent := m.dv1.findState.current

	m = m.focusPane(filesPane)
	m.files = m.files.move(1) // active.md -> peeked.md
	if m.files.selectedNode().Name != "peeked.md" {
		t.Fatalf("setup: selection = %q, want peeked.md", m.files.selectedNode().Name)
	}
	m = press(t, m, keys.Peek.Keys()[0])
	if !m.dv1.peeking() {
		t.Fatalf("setup: Peek did not start")
	}

	// Three plain nexts, not an equal mix of n and N: alternating the two
	// would cancel back to the starting index/offset even with the guard
	// missing entirely, which would make this assertion pass for the wrong
	// reason. A run of nexts only returns to start if each one is actually
	// gated away.
	m = press(t, m, keys.FindNext.Keys()[0])
	m = press(t, m, keys.FindNext.Keys()[0])
	m = press(t, m, keys.FindNext.Keys()[0])
	if got := m.dv1.findState.current; got != wantCurrent {
		t.Errorf("find match index after n while Peeking = %d, want unchanged %d", got, wantCurrent)
	}
	if got := m.dv1.offset(); got != wantOffset {
		t.Errorf("viewer offset after n while Peeking = %d, want unchanged %d", got, wantOffset)
	}

	m = press(t, m, keys.FindPrevious.Keys()[0])
	m = press(t, m, keys.FindPrevious.Keys()[0])
	if got := m.dv1.findState.current; got != wantCurrent {
		t.Errorf("find match index after N while Peeking = %d, want unchanged %d", got, wantCurrent)
	}
	if got := m.dv1.offset(); got != wantOffset {
		t.Errorf("viewer offset after N while Peeking = %d, want unchanged %d", got, wantOffset)
	}

	m = press(t, m, keys.Peek.Keys()[0]) // second p: clear
	if m.dv1.peeking() {
		t.Error("still Peeking after the second p")
	}
	if got := m.dv1.offset(); got != wantOffset {
		t.Errorf("Active Document offset after clearing Peek = %d, want exact pre-Peek offset %d (no drift)", got, wantOffset)
	}
	if got := m.dv1.findState.current; got != wantCurrent {
		t.Errorf("find match index after clearing Peek = %d, want unchanged %d", got, wantCurrent)
	}
}
