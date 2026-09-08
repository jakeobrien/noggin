package main

import (
	"path/filepath"
	"testing"
)

// TestToggleActiveDocStarIsIndependentPerViewer is M2.2's acceptance
// criterion: starring from a focused Document Viewer acts on *that*
// viewer's own Active Document (S2), so DV1 and DV2 star two different
// documents with neither action disturbing the other.
func TestToggleActiveDocStarIsIndependentPerViewer(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "documents")
	pathA := writeFileContent(t, filepath.Join(dir, "a.md"), "alpha")
	pathB := writeFileContent(t, filepath.Join(dir, "b.md"), "bravo")

	docA, err := loadDocument(pathA)
	if err != nil {
		t.Fatalf("loadDocument(a): %v", err)
	}
	docB, err := loadDocument(pathB)
	if err != nil {
		t.Fatalf("loadDocument(b): %v", err)
	}

	m := sized(t, newModel().withDocs(storePath, loadDocStore(storePath)), 80, 10)
	m = m.openDocument(docA)              // DV1, focus -> dv1Pane
	m = m.openDocumentInto(dv2Pane, docB) // DV2, focus -> dv2Pane
	if m.focus != dv2Pane {
		t.Fatalf("setup: focus = %v, want dv2Pane", m.focus)
	}

	// Star from DV2 first (it already holds Focus).
	m = press(t, m, keys.Star.Keys()[0])
	if m.message != "" {
		t.Fatalf("toggleActiveDocStar from DV2: unexpected message %q", m.message)
	}
	if got := m.docRecord(pathB); !got.Starred {
		t.Errorf("docRecord(b) after starring from DV2 = %+v, want Starred", got)
	}
	if got := m.docRecord(pathA); got.Starred {
		t.Errorf("docRecord(a) after starring from DV2 = %+v, want unaffected", got)
	}

	// Now star from DV1.
	m = press(t, m, keys.FocusDV1.Keys()[0])
	m = press(t, m, keys.Star.Keys()[0])
	if got := m.docRecord(pathA); !got.Starred {
		t.Errorf("docRecord(a) after starring from DV1 = %+v, want Starred", got)
	}
	if got := m.docRecord(pathB); !got.Starred {
		t.Errorf("docRecord(b) after starring from DV1 = %+v, want still Starred (unaffected)", got)
	}

	// Un-star from DV2: only b's record clears.
	m = press(t, m, keys.FocusDV2.Keys()[0])
	m = press(t, m, keys.Star.Keys()[0])
	if got := m.docRecord(pathB); got.Starred {
		t.Errorf("docRecord(b) after un-starring from DV2 = %+v, want cleared", got)
	}
	if got := m.docRecord(pathA); !got.Starred {
		t.Errorf("docRecord(a) after un-starring DV2 = %+v, want still Starred (unaffected)", got)
	}

	// Persists: a fresh model loading the same store path sees both records.
	restarted := newModel().withDocs(storePath, loadDocStore(storePath))
	if got := restarted.docRecord(pathA); !got.Starred {
		t.Errorf("docRecord(a) after restart = %+v, want Starred", got)
	}
	if got := restarted.docRecord(pathB); got.Starred {
		t.Errorf("docRecord(b) after restart = %+v, want unstarred", got)
	}
}

// TestToggleActiveDocStarStarsAndPersistsAnArticle is C6/M5.1's own
// acceptance criterion for the Document Viewer half: starring an article
// (keyed by its canonical URL, M4.1) from the focused viewer must stamp
// Article/Title on the record the same way togglePickerStar's Wikipedia
// branch already does (M3.4) — without that, the record round-trips
// through saveDocStore as an invalid "file" line (its URL key fails
// filepath.IsAbs) and loadDocStore silently drops it, so the star would
// look like it worked for the rest of the session but vanish on restart.
// This exercises the real gap: toggleActiveDocStar previously left
// Article/Title at docRecord's zero value for a never-before-starred
// article.
func TestToggleActiveDocStarStarsAndPersistsAnArticle(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "documents")
	article := fixtureArticle()

	m := sized(t, newModel().withDocs(storePath, loadDocStore(storePath)).openDocument(article), 80, 24)

	m = press(t, m, keys.Star.Keys()[0])
	if m.message != "" {
		t.Fatalf("toggleActiveDocStar on an article: unexpected message %q", m.message)
	}
	got := m.docRecord(article.path)
	if !got.Starred || !got.Article || got.Title != article.title {
		t.Fatalf("docRecord(article) after starring = %+v, want Starred, Article, Title %q", got, article.title)
	}

	// Persists: a fresh model loading the same on-disk store sees the same
	// record — the round-trip this test exists to pin.
	restarted := newModel().withDocs(storePath, loadDocStore(storePath))
	if got := restarted.docRecord(article.path); !got.Starred || !got.Article || got.Title != article.title {
		t.Errorf("docRecord(article) after restart = %+v, want Starred, Article, Title %q", got, article.title)
	}

	// Un-starring clears Starred but leaves Article/Title known (the same
	// "a record persists with Starred=false" shape a file's own record
	// already gives, M2.1).
	m = press(t, m, keys.Star.Keys()[0])
	got = m.docRecord(article.path)
	if got.Starred {
		t.Error("docRecord(article) after un-starring still Starred, want false")
	}
	if !got.Article || got.Title != article.title {
		t.Errorf("docRecord(article) after un-starring = %+v, want Article/Title preserved", got)
	}
	restarted = newModel().withDocs(storePath, loadDocStore(storePath))
	if got := restarted.docRecord(article.path); got.Starred {
		t.Errorf("docRecord(article) after restart still Starred, want false")
	}
}

// TestToggleActiveDocStarNoOpWithNoDocumentOpen confirms the empty-pane
// refusal (a footer message, the same shape startEdit/startFind give their
// own "nothing to act on" case) rather than a silent no-op or a panic on a
// zero-value document.
func TestToggleActiveDocStarNoOpWithNoDocumentOpen(t *testing.T) {
	m := sized(t, newModel(), 80, 24) // no document opened
	m.focus = dv1Pane

	m = m.toggleActiveDocStar()
	if m.message == "" {
		t.Error("want a footer message with no document open, got none")
	}
}

// TestToggleActiveDocStarNoOpWhilePeeking mirrors startEdit/startFind's own
// S7/M7.1 guard: Peek shows a document without making it Active, so a star
// press while Peeking must leave both the Active Document's and the
// Peeked document's records untouched, and leave no message either — the
// same "do nothing at all" shape those two already give.
func TestToggleActiveDocStarNoOpWhilePeeking(t *testing.T) {
	m, activePath, peekedPath := peekFixture(t)

	m = press(t, m, keys.Peek.Keys()[0])
	if !m.dv1.peeking() {
		t.Fatalf("setup: expected DV1 to be Peeking")
	}
	m.focus = dv1Pane

	m = m.toggleActiveDocStar()
	if got := m.docRecord(activePath); got.Starred {
		t.Errorf("docRecord(active) after starring while Peeking = %+v, want unaffected", got)
	}
	if got := m.docRecord(peekedPath); got.Starred {
		t.Errorf("docRecord(peeked) after starring while Peeking = %+v, want unaffected", got)
	}
	if m.message != "" {
		t.Errorf("toggleActiveDocStar while Peeking: want silent no-op, got message %q", m.message)
	}
}
