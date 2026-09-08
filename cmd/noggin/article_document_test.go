package main

import (
	"path/filepath"
	"strings"
	"testing"

	"noggin/internal/wikipedia"
)

// fixtureArticle is a directly-constructed article document (C6/M4.1) — the
// shape a fetch cmd (M4.2) will eventually produce from *wikipedia.Client's
// own Article, exercised here without any network involved since M4.1's own
// scope is the document/viewer/history plumbing, not the fetch itself.
func fixtureArticle() document {
	return articleDocument(wikipedia.Article{
		Title: "Leonhard Euler",
		URL:   "https://en.wikipedia.org/wiki/Leonhard_Euler",
		Markdown: "# Leonhard Euler\n\n" +
			"Leonhard Euler was a Swiss mathematician, physicist and engineer.\n\n" +
			strings.Repeat("Filler paragraph to push content past one screen.\n\n", 30) +
			"The final line of the article fixture.\n",
	})
}

// TestArticleDocumentIdentity is M4.1's own acceptance criterion for the
// document type itself: an article document is keyed by its canonical URL
// (S4's store key), labelled by its title rather than a URL-derived name,
// and told apart from a file document by isArticle() — the same "://"
// heuristic docStoreKey already used for the document store (D127).
func TestArticleDocumentIdentity(t *testing.T) {
	d := fixtureArticle()
	if !d.isOpen() {
		t.Fatal("articleDocument produced a closed document")
	}
	if !d.isArticle() {
		t.Error("isArticle() = false for an article document")
	}
	if d.path != "https://en.wikipedia.org/wiki/Leonhard_Euler" {
		t.Errorf("path = %q, want the canonical URL", d.path)
	}
	if got, want := d.label(), "Leonhard Euler"; got != want {
		t.Errorf("label() = %q, want %q (the article's title, not a URL-derived name)", got, want)
	}

	file := fixtureDoc(t)
	if file.isArticle() {
		t.Error("isArticle() = true for a file document")
	}
	if got, want := file.label(), filepath.Base(fixturePath); got != want {
		t.Errorf("file label() = %q, want %q (unchanged behaviour)", got, want)
	}
}

// TestViewerOpensAndLabelsAnArticle is the viewer half of the same
// criterion: opening an article document renders it and titles the pane by
// the article's title, both directly (openDocument) and under Peek.
func TestViewerOpensAndLabelsAnArticle(t *testing.T) {
	m := sized(t, newModel().openDocument(fixtureArticle()), 80, 24)
	if got, want := m.dv1.label(), "Leonhard Euler"; got != want {
		t.Errorf("DV1 label = %q, want %q", got, want)
	}
	view := m.View().Content
	if !strings.Contains(view, "Document: Leonhard Euler") {
		t.Errorf("pane title row does not show the article's title (viewer_star_render_test.go's own \"Document: <label>\" shape):\n%s", view)
	}
	if !strings.Contains(view, "Swiss mathematician") {
		t.Errorf("view does not show the article's rendered content:\n%s", view)
	}

	v := newViewer().setSize(80, 24).startPeek(fixtureArticle())
	if got, want := v.peekLabel(), "Leonhard Euler"; got != want {
		t.Errorf("peekLabel() = %q, want %q", got, want)
	}
}

// TestFindWithinAnArticleDocument is S3's "Find ... appl[ies] as for a
// file": a phrase near the end of an open article is found and brought on
// screen exactly as TestFindLocatesAPhraseNearTheEndOfTheDocument already
// proves for a file, since find.go never looks past the viewport's rendered
// content (no path/URL distinction to trip over).
func TestFindWithinAnArticleDocument(t *testing.T) {
	m := sized(t, newModel().openDocument(fixtureArticle()), 80, 24)

	m = press(t, m, "/")
	if !m.dv1.findEditing() {
		t.Fatal("/ did not start a find")
	}
	m = typeText(t, m, "final line of the article fixture")
	if got := m.dv1.offset(); got == 0 {
		t.Error("a query matching the article's tail did not scroll the viewport")
	}
	view := m.View().Content
	if !strings.Contains(view, "The final line of the article fixture.") {
		t.Errorf("view does not show the matched tail phrase after searching:\n%s", view)
	}

	m = press(t, m, "enter")
	if m.dv1.findEditing() {
		t.Error("enter did not commit the find (still editing)")
	}
	if !m.dv1.hasFindMatches() {
		t.Error("committing the find lost its matches")
	}
}

// TestHistoryWalksBetweenAFileAndAnArticle is S3's "Document History ...
// appl[ies] as for a file", the cross-class case: a file is activated, then
// an article is navigated to directly (standing in for M4.2's not-yet-built
// fetch-completion handler, which will call model.navigate the same way),
// and Back/Forward walks between the two without ever touching the
// network — the article's reopen comes from the viewer's own articleCache
// (viewer.go), not a fresh fetch, which a stray attempt at
// loadDocument(url) would surface as a "cannot open" message instead of the
// article's content.
func TestHistoryWalksBetweenAFileAndAnArticle(t *testing.T) {
	dir := t.TempDir()
	filePath := writeFileContent(t, filepath.Join(dir, "note.md"), "# A file\n\nJust a file.\n")

	m := sized(t, newModel(), 80, 24)
	fileDoc, err := loadDocument(filePath)
	if err != nil {
		t.Fatalf("loadDocument: %v", err)
	}
	m = m.navigate(dv1Pane, fileDoc)
	m = m.navigate(dv1Pane, fixtureArticle())

	if got, want := m.dv1.doc.path, "https://en.wikipedia.org/wiki/Leonhard_Euler"; got != want {
		t.Fatalf("setup: DV1 = %q, want %q", got, want)
	}
	if m.message != "" {
		t.Fatalf("setup: unexpected message %q", m.message)
	}

	m = press(t, m, keys.FocusDV1.Keys()[0])
	m = press(t, m, keys.HistoryBack.Keys()[0])
	if m.dv1.doc.path != filePath {
		t.Errorf("Back from the article = %q, want the file %q (message: %q)", m.dv1.doc.path, filePath, m.message)
	}
	if m.message != "" {
		t.Errorf("Back from the article produced a message: %q", m.message)
	}

	m = press(t, m, keys.HistoryForward.Keys()[0])
	if got, want := m.dv1.doc.path, "https://en.wikipedia.org/wiki/Leonhard_Euler"; got != want {
		t.Errorf("Forward back to the article = %q, want %q (message: %q)", got, want, m.message)
	}
	if m.message != "" {
		t.Errorf("Forward back to the article produced a message (a real fetch was attempted?): %q", m.message)
	}
	if got, want := m.dv1.label(), "Leonhard Euler"; got != want {
		t.Errorf("label() after walking Forward back to the article = %q, want %q", got, want)
	}
	view := m.View().Content
	if !strings.Contains(view, "Swiss mathematician") {
		t.Errorf("view after walking Forward back to the article does not show its content:\n%s", view)
	}
}

// TestHistoryBackWithAnUncachedArticleReportsAMessageNotACrash locks in
// walkViewerHistory's own defensive branch: if a URL somehow reaches history
// without a matching articleCache entry (an invariant break, not a normal
// outcome per walkViewerHistory's own doc comment), the walk fails with a
// Footer message and leaves the model's current document untouched, the
// same shape any other failed history reopen already gives, rather than
// panicking or silently closing the viewer.
func TestHistoryBackWithAnUncachedArticleReportsAMessageNotACrash(t *testing.T) {
	dir := t.TempDir()
	filePathA := writeFileContent(t, filepath.Join(dir, "a.md"), "# A file\n\nJust a file.\n")
	filePathB := writeFileContent(t, filepath.Join(dir, "b.md"), "# Another file\n\nAlso just a file.\n")

	m := sized(t, newModel(), 80, 24)
	docA, err := loadDocument(filePathA)
	if err != nil {
		t.Fatalf("loadDocument(a): %v", err)
	}
	docB, err := loadDocument(filePathB)
	if err != nil {
		t.Fatalf("loadDocument(b): %v", err)
	}
	// a -> article -> b, so Back from b lands on the article.
	m = m.navigate(dv1Pane, docA)
	m = m.navigate(dv1Pane, fixtureArticle())
	m = m.navigate(dv1Pane, docB)

	// Break the invariant deliberately: drop the article from articleCache
	// without touching history, standing in for whatever bug would have to
	// exist elsewhere to produce this state for real.
	m.dv1.articleCache = map[string]document{}

	m = press(t, m, keys.FocusDV1.Keys()[0])
	m = press(t, m, keys.HistoryBack.Keys()[0])

	if got, want := m.dv1.doc.path, filePathB; got != want {
		t.Errorf("DV1 after a failed history reopen = %q, want unchanged %q", got, want)
	}
	if !strings.Contains(m.message, "cannot open") {
		t.Errorf("message = %q, want a \"cannot open ...\" Footer message", m.message)
	}
}
