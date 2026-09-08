package main

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestSourcesNavShowsWikipediaRowWithZeroRoots is C6/M1.2's acceptance
// criterion (zero-Root half): the Wikipedia row renders (two lines: label,
// then domain) and the "No roots…" empty state is gone.
func TestSourcesNavShowsWikipediaRowWithZeroRoots(t *testing.T) {
	r := newRootsNav().setSize(30, 6)
	lines := strings.Split(r.view(), "\n")
	if !strings.Contains(lines[0], "Wikipedia") {
		t.Errorf("line 0 = %q, want the Wikipedia label", lines[0])
	}
	if !strings.Contains(lines[1], "en.wikipedia.org") {
		t.Errorf("line 1 = %q, want the Wikipedia domain", lines[1])
	}
	if strings.Contains(r.view(), "No roots") {
		t.Errorf("view = %q, want the old empty state gone", r.view())
	}
}

// TestSourcesNavShowsWikipediaRowLastWithSeveralRoots is C6/M1.2's
// acceptance criterion (several-Roots half): the Wikipedia row still
// renders, after every Root row.
func TestSourcesNavShowsWikipediaRowLastWithSeveralRoots(t *testing.T) {
	a := mkdir(t, filepath.Join(t.TempDir(), "a"))
	b := mkdir(t, filepath.Join(t.TempDir(), "b"))
	r := newRootsNav().setRoots([]root{{path: a}, {path: b}}, 0).setSize(30, 6)

	lines := strings.Split(r.view(), "\n")
	if len(lines) != 6 {
		t.Fatalf("view has %d lines, want 6 (three entries x rootRowHeight)", len(lines))
	}
	if !strings.Contains(lines[0], "a") || !strings.Contains(lines[2], "b") {
		t.Fatalf("root rows = %q / %q, want a's then b's own name lines", lines[0], lines[2])
	}
	if !strings.Contains(lines[4], "Wikipedia") {
		t.Errorf("last row name line = %q, want the Wikipedia label", lines[4])
	}
	if !strings.Contains(lines[5], "en.wikipedia.org") {
		t.Errorf("last row path line = %q, want the Wikipedia domain", lines[5])
	}
}

// TestSourcesNavWikipediaRowNeverMarkedBrokenOrStarred checks S1's "broken/
// starred markers never apply to it": neither the ★ nor the " (missing)"
// suffix ever appears on the Wikipedia row, regardless of selection or
// Active Source state.
func TestSourcesNavWikipediaRowNeverMarkedBrokenOrStarred(t *testing.T) {
	r := newRootsNav().setSize(30, 2)
	r.selected = 0 // the only row: Wikipedia
	view := r.view()
	if strings.Contains(view, "★") {
		t.Errorf("Wikipedia row view = %q, want no star marker", view)
	}
	if strings.Contains(view, "missing") {
		t.Errorf("Wikipedia row view = %q, want no broken marker", view)
	}
}

// TestSourcesNavActivateOnWikipediaRowMakesItTheActiveSource drives S1's
// "it can be selected and made the Active Source" through the model's own
// keypress path (Activate on the Sources Nav's Wikipedia row).
func TestSourcesNavActivateOnWikipediaRowMakesItTheActiveSource(t *testing.T) {
	dir := mkdir(t, filepath.Join(t.TempDir(), "root"))
	m := sized(t, newModel(), 80, 24)
	m.roots = m.roots.setRoots([]root{{path: dir}}, 0)
	m = m.activateRoot()
	m.focus = rootsPane
	m.roots = m.roots.move(1) // cursor onto the Wikipedia row (index 1)
	if !m.roots.selectedIsWikipedia() {
		t.Fatalf("setup: selection did not land on the Wikipedia row")
	}

	m = press(t, m, "enter")
	if !m.roots.activeIsWikipedia() {
		t.Fatal("Activate on the Wikipedia row did not make it the Active Source")
	}
	if m.roots.active != -1 {
		t.Errorf("roots.active = %d, want -1 (no Root active)", m.roots.active)
	}
}

// TestSourcesNavStarOnWikipediaRowIsRefusedWithFooterMessage is C6/M1.2's
// other acceptance criterion: `s` on the Wikipedia row is a no-op with a
// Footer message, not a silent no-op and not a toggled star.
func TestSourcesNavStarOnWikipediaRowIsRefusedWithFooterMessage(t *testing.T) {
	m := sized(t, newModel(), 80, 24)
	m.focus = rootsPane
	if !m.roots.selectedIsWikipedia() {
		t.Fatalf("setup: a Root-less session's only row should be Wikipedia")
	}

	m = press(t, m, "s")
	if m.message == "" {
		t.Error("Star on the Wikipedia row produced no Footer message")
	}
	if !strings.Contains(m.frame(), m.message) {
		t.Errorf("frame does not surface the refusal message %q", m.message)
	}
}
