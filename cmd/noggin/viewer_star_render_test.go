package main

import (
	"path/filepath"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

// TestDocumentViewerTitleShowsStarIndicatorWithNavigatorHidden is M2.4's
// acceptance criterion: with the Navigator hidden (S3's own "legible even
// with the Navigator hidden" clause), DV1's and DV2's title rows carry the
// "★ " indicator immediately to the left of the document label for a
// starred Active Document only — an unstarred one gets no indicator, and
// starring in one viewer never marks the other.
func TestDocumentViewerTitleShowsStarIndicatorWithNavigatorHidden(t *testing.T) {
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

	m := sized(t, newModel().withDocs(storePath, loadDocStore(storePath)), 80, 24)
	m = m.openDocument(docA)              // DV1, focus -> dv1Pane
	m = m.openDocumentInto(dv2Pane, docB) // DV2, focus -> dv2Pane
	m = m.toggleNavigator()               // hide the Navigator
	if !m.navigatorHidden {
		t.Fatalf("setup: Navigator not hidden")
	}

	view := m.View().Content
	if !strings.Contains(view, "Document: a.md") || !strings.Contains(view, "Document: b.md") {
		t.Fatalf("setup: expected titles not found before starring; view = %q", view)
	}
	if strings.Contains(view, "★") {
		t.Fatalf("setup: unexpected star before either document is starred; view = %q", view)
	}

	// Star a.md from DV1.
	m = press(t, m, keys.FocusDV1.Keys()[0])
	m = press(t, m, keys.Star.Keys()[0])
	view = m.View().Content
	if !strings.Contains(view, "Document: ★ a.md") {
		t.Errorf("DV1 title after starring a.md = %q, want a star immediately left of the label", view)
	}
	if strings.Contains(view, "★ b.md") {
		t.Errorf("DV2 title affected by starring a.md; view = %q", view)
	}

	// Star b.md from DV2 too: both indicators now present, independently.
	m = press(t, m, keys.FocusDV2.Keys()[0])
	m = press(t, m, keys.Star.Keys()[0])
	view = m.View().Content
	if !strings.Contains(view, "Document: ★ a.md") || !strings.Contains(view, "Document: ★ b.md") {
		t.Errorf("titles after starring both = %q, want both star indicators present", view)
	}

	// Un-star a.md from DV1: its indicator clears, b.md's stays.
	m = press(t, m, keys.FocusDV1.Keys()[0])
	m = press(t, m, keys.Star.Keys()[0])
	view = m.View().Content
	if strings.Contains(view, "★ a.md") {
		t.Errorf("DV1 title after un-starring a.md = %q, want no star indicator", view)
	}
	if !strings.Contains(view, "★ b.md") {
		t.Errorf("DV2 title after un-starring a.md = %q, want its own star indicator unaffected", view)
	}
}

// TestDocumentViewerTitleShowsStarIndicatorForAnArticle is C6/M5.1's own
// acceptance criterion for the viewer title row: the "★ " indicator
// M2.4 already gives a starred file's title applies identically to a
// starred article's title (its own label, the article's title — M4.1 — not
// a URL), and clears the same way on un-star.
func TestDocumentViewerTitleShowsStarIndicatorForAnArticle(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "documents")
	article := fixtureArticle()

	m := sized(t, newModel().withDocs(storePath, loadDocStore(storePath)).openDocument(article), 80, 24)

	view := m.View().Content
	if !strings.Contains(view, "Document: Leonhard Euler") {
		t.Fatalf("setup: expected title not found before starring; view = %q", view)
	}
	if strings.Contains(view, "★") {
		t.Fatalf("setup: unexpected star before the article is starred; view = %q", view)
	}

	m = press(t, m, keys.Star.Keys()[0])
	view = m.View().Content
	if !strings.Contains(view, "Document: ★ Leonhard Euler") {
		t.Errorf("DV1 title after starring the article = %q, want a star immediately left of the title", view)
	}

	m = press(t, m, keys.Star.Keys()[0])
	view = m.View().Content
	if strings.Contains(view, "★") {
		t.Errorf("DV1 title after un-starring the article = %q, want no star indicator", view)
	}
}

// TestDocumentViewerTitleStarIndicatorDegradesAtNarrowWidth confirms the
// title row still renders within its pane's width once the "★ " indicator
// is added — no line exceeding the box width, and the label still visible
// rather than the box silently vanishing or panicking at a narrow size.
func TestDocumentViewerTitleStarIndicatorDegradesAtNarrowWidth(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "documents")
	path := writeFileContent(t, filepath.Join(dir, "a.md"), "alpha")

	doc, err := loadDocument(path)
	if err != nil {
		t.Fatalf("loadDocument: %v", err)
	}

	m := sized(t, newModel().withDocs(storePath, loadDocStore(storePath)), 80, 24)
	m = m.openDocument(doc)
	m = m.toggleActiveDocStar()
	m = m.toggleNavigator() // narrowest realistic case: no Navigator competing for width
	m = m.resize(24, 24, resizeCauseWindow)

	view := m.View().Content
	for i, line := range strings.Split(view, "\n") {
		if w := lipgloss.Width(line); w > 24 {
			t.Errorf("line %d width = %d, want <= 24 (terminal width); line = %q", i, w, line)
		}
	}
	if !strings.Contains(view, "a.md") {
		t.Errorf("view at narrow width = %q, want the document label still present", view)
	}
}
