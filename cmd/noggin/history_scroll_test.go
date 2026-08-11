package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"noggin/internal/tree"
)

// longDocFixture is threeDocFixture's own shape (a single Root, opened with
// Focus on the Files Nav and the top row selected) but with a.md long enough
// to have real scroll room, plus two short files (b.md, c.md) to navigate
// away to — S9/C4-M6.2's own test needs a document that can actually be
// scrolled, which threeDocFixture's one-line fixtures cannot.
func longDocFixture(t *testing.T) (m model, docA, docB, docC string) {
	t.Helper()
	dir := t.TempDir()
	var lines []string
	for i := 1; i <= 200; i++ {
		lines = append(lines, fmt.Sprintf("line %d", i))
	}
	docA = writeFileContent(t, filepath.Join(dir, "a.md"), strings.Join(lines, "\n")+"\n")
	docB = writeFileContent(t, filepath.Join(dir, "b.md"), "doc b\n")
	docC = writeFileContent(t, filepath.Join(dir, "c.md"), "doc c\n")

	node, err := tree.Walk(dir)
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	m = sized(t, newModel(), 80, 24)
	m = m.openFolder(node)
	if got := m.files.selectedNode(); got == nil || got.Path != docA {
		t.Fatalf("setup: selected node = %+v, want %q", got, docA)
	}
	return m, docA, docB, docC
}

// TestReActivatingTheAlreadyActiveDocumentPreservesScrollPosition regression-
// tests D91's flagged-but-untested side effect (TRIAGE T43): re-Activating a
// document that is already the Active Document, with no navigation in
// between, preserves its scroll position rather than resetting to top.
// D91 put M6.2's docOffsets save-before-leaving/restore-on-arrival logic
// inside viewer.open() itself rather than only in the history walk, so
// every Activate funnels through it, including a no-motion re-Activate of
// the document already open — open() banks the outgoing (= incoming)
// document's offset via rememberOffset(), then immediately restores that
// same value from docOffsets, a no-op by construction rather than a special
// case, per D91's own reasoning. This was flagged as untested by review
// (D91) rather than demonstrated broken; this test closes that gap.
func TestReActivatingTheAlreadyActiveDocumentPreservesScrollPosition(t *testing.T) {
	m, docA, _, _ := longDocFixture(t)

	m = activateSelected(t, m) // a.md, top
	if total := m.dv1.lineCount(); total <= m.dv1.height() {
		t.Fatalf("fixture renders %d lines into a %d-line viewport; nothing to scroll", total, m.dv1.height())
	}
	m.dv1 = m.dv1.withOffset(52)
	wantOffset := m.dv1.offset()

	m = activateSelected(t, m) // a.md again — same document, no navigation away
	if m.dv1.doc.path != docA {
		t.Fatalf("setup: DV1 = %q, want %q", m.dv1.doc.path, docA)
	}
	if got := m.dv1.offset(); got != wantOffset {
		t.Errorf("offset after re-Activating the already-active document = %d, want unchanged %d", got, wantOffset)
	}
}

// TestHistoryBackRestoresRememberedScrollPosition is M6.2's own acceptance
// criterion, verbatim: "a test scrolls a document, navigates away and back,
// and lands at the remembered offset".
func TestHistoryBackRestoresRememberedScrollPosition(t *testing.T) {
	m, docA, docB, _ := longDocFixture(t)

	m = activateSelected(t, m) // a.md, top
	if total := m.dv1.lineCount(); total <= m.dv1.height() {
		t.Fatalf("fixture renders %d lines into a %d-line viewport; nothing to scroll", total, m.dv1.height())
	}
	m.dv1 = m.dv1.withOffset(37)
	wantOffset := m.dv1.offset()

	m = press(t, m, keys.FocusFiles.Keys()[0])
	m.files = m.files.move(1)
	m = activateSelected(t, m) // b.md — leaves a.md, banking its offset
	if m.dv1.doc.path != docB {
		t.Fatalf("setup: DV1 = %q, want %q", m.dv1.doc.path, docB)
	}
	if got := m.dv1.offset(); got != 0 {
		t.Fatalf("setup: b.md offset = %d, want 0 (never scrolled)", got)
	}

	m = press(t, m, keys.FocusDV1.Keys()[0])
	m = press(t, m, keys.HistoryBack.Keys()[0]) // b -> a
	if m.dv1.doc.path != docA {
		t.Fatalf("Back = %q, want %q", m.dv1.doc.path, docA)
	}
	if got := m.dv1.offset(); got != wantOffset {
		t.Errorf("offset after Back onto a re-visited document = %d, want remembered %d", got, wantOffset)
	}
}

// TestHistoryRevisitingADifferentEntryForTheSameDocumentSharesItsOffset is
// M6.2's other acceptance clause: "revisiting via a *different* history
// entry for the same document lands at the same offset" — the memory is
// keyed by document, not by history entry index. History here is
// a(entry0) -> b(1) -> a(2) -> c(3): a is scrolled once after entry0 and
// again, further, after entry2; walking all the way back from c must show
// entry2's scroll amount at *both* a-entries, not entry0's own original one.
func TestHistoryRevisitingADifferentEntryForTheSameDocumentSharesItsOffset(t *testing.T) {
	m, docA, docB, docC := longDocFixture(t)

	m = activateSelected(t, m) // a, entry 0
	m.dv1 = m.dv1.withOffset(10)
	firstAOffset := m.dv1.offset()

	m = press(t, m, keys.FocusFiles.Keys()[0])
	m.files = m.files.move(1)
	m = activateSelected(t, m) // b, entry 1 — banks a's offset (10)
	if m.dv1.doc.path != docB {
		t.Fatalf("setup: DV1 = %q, want %q", m.dv1.doc.path, docB)
	}

	m = press(t, m, keys.FocusFiles.Keys()[0])
	m.files = m.files.move(-1)
	m = activateSelected(t, m) // a again, entry 2 (not adjacent to entry 0's b)
	if m.dv1.doc.path != docA {
		t.Fatalf("setup: DV1 = %q, want %q", m.dv1.doc.path, docA)
	}
	if got := m.dv1.offset(); got != firstAOffset {
		t.Fatalf("setup: re-activating a landed at %d, want its remembered %d", got, firstAOffset)
	}
	m.dv1 = m.dv1.withOffset(60)
	secondAOffset := m.dv1.offset()
	if secondAOffset == firstAOffset {
		t.Fatalf("setup is broken: second scroll (%d) must differ from the first (%d)", secondAOffset, firstAOffset)
	}

	m = press(t, m, keys.FocusFiles.Keys()[0])
	m.files = m.files.move(2)  // from a.md (row 0) to c.md (row 2)
	m = activateSelected(t, m) // c, entry 3 — banks a's second offset (60)
	if m.dv1.doc.path != docC {
		t.Fatalf("setup: DV1 = %q, want %q", m.dv1.doc.path, docC)
	}

	m = press(t, m, keys.FocusDV1.Keys()[0])
	m = press(t, m, keys.HistoryBack.Keys()[0]) // c -> a (entry 2)
	if m.dv1.doc.path != docA {
		t.Fatalf("Back once = %q, want %q (entry 2)", m.dv1.doc.path, docA)
	}
	if got := m.dv1.offset(); got != secondAOffset {
		t.Errorf("offset at entry 2 (a) = %d, want its own remembered %d", got, secondAOffset)
	}

	m = press(t, m, keys.HistoryBack.Keys()[0]) // a -> b (entry 1)
	if m.dv1.doc.path != docB {
		t.Fatalf("Back twice = %q, want %q (entry 1)", m.dv1.doc.path, docB)
	}

	m = press(t, m, keys.HistoryBack.Keys()[0]) // b -> a (entry 0, a different entry, same document)
	if m.dv1.doc.path != docA {
		t.Fatalf("Back three times = %q, want %q (entry 0)", m.dv1.doc.path, docA)
	}
	if got := m.dv1.offset(); got != secondAOffset {
		t.Errorf("offset at entry 0 (a, a different entry for the same document) = %d, want the shared, most recently remembered %d (not entry 0's own original %d)", got, secondAOffset, firstAOffset)
	}
}
