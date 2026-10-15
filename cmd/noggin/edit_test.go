package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestEditWithNoDocumentReportsFooterMessage covers S10's "with no document
// open, it reports that rather than doing nothing": the edit key must not
// silently no-op, and must not produce a command (there is nothing to hand
// $EDITOR).
func TestEditWithNoDocumentReportsFooterMessage(t *testing.T) {
	m := sized(t, newModel(), 80, 24)
	m.focus = dv1Pane
	next, cmd := m.Update(keyPress(keys.Edit.Keys()[0]))
	got := concrete(t, next)

	if cmd != nil {
		t.Errorf("cmd = %v, want nil", cmd)
	}
	if got.message == "" {
		t.Error("message is empty, want a report that no document is open")
	}
}

// TestEditWithOpenDocumentReleasesTheTerminal covers the handoff itself: with
// a document open, the edit key produces a command (tea.ExecProcess's, which
// pauses the Program and runs $EDITOR) rather than a message, and leaves the
// open document untouched until the editor actually exits. Viewer-only since
// D141, so Focus sits on DV1.
func TestEditWithOpenDocumentReleasesTheTerminal(t *testing.T) {
	d, err := loadDocument(fixturePath)
	if err != nil {
		t.Fatalf("loading %s: %v", fixturePath, err)
	}
	m := sized(t, newModel(), 80, 24)
	m = m.openDocument(d)
	m.focus = dv1Pane

	next, cmd := m.Update(keyPress(keys.Edit.Keys()[0]))
	got := concrete(t, next)

	if cmd == nil {
		t.Fatal("cmd = nil, want a command to release the terminal to $EDITOR")
	}
	if got.message != "" {
		t.Errorf("message = %q, want empty (a command was produced instead)", got.message)
	}
	if got.dv1.doc.path != fixturePath {
		t.Errorf("open document = %q, want unchanged (%q) until the editor exits",
			got.dv1.doc.path, fixturePath)
	}
}

// TestEditOnAWikipediaArticleIsRefusedWithFooterMessage is S3/C6/M4.3's own
// refusal clause, checked with the plan's own "article in DV1 beside a file
// in DV2" shape: an article has nothing on disk to hand $EDITOR, so `e` on
// one (Focus on DV1) reports a Footer message and produces no command — the
// same "refused, not silently ignored" shape
// TestEditWithNoDocumentReportsFooterMessage already gives the no-document
// case, distinct from Peek's own silent no-op (S7/M7.1's carve-out, which
// Edit does not share) — and DV2's own file is left completely untouched,
// since Edit acts only on the focused viewer's Active Document.
func TestEditOnAWikipediaArticleIsRefusedWithFooterMessage(t *testing.T) {
	filePath := writeFileContent(t, filepath.Join(t.TempDir(), "note.md"), "# A file\n\nUnaffected.\n")
	fileDoc, err := loadDocument(filePath)
	if err != nil {
		t.Fatalf("loadDocument: %v", err)
	}
	m := sized(t, newModel(), 80, 24)
	m = m.navigate(dv1Pane, fixtureArticle())
	m = m.openDocumentInto(dv2Pane, fileDoc)
	m.focus = dv1Pane

	next, cmd := m.Update(keyPress(keys.Edit.Keys()[0]))
	got := concrete(t, next)

	if cmd != nil {
		t.Errorf("cmd = %v, want nil (no $EDITOR handoff for an article)", cmd)
	}
	if got.message == "" {
		t.Error("message is empty, want a report that a Wikipedia article cannot be edited")
	}
	if got.dv1.doc.path != fixtureArticle().path {
		t.Errorf("DV1 Active Document = %q after a refused edit, want unchanged %q", got.dv1.doc.path, fixtureArticle().path)
	}
	if got.dv2.doc.path != filePath {
		t.Errorf("DV2 Active Document = %q after DV1's edit was refused, want unchanged %q", got.dv2.doc.path, filePath)
	}
}

// TestEditOnAWikipediaArticleInDV2IsRefusedWithFooterMessage is the mirror
// of TestEditOnAWikipediaArticleIsRefusedWithFooterMessage (M4.4/TRIAGE
// T23): startEdit acts on the focused viewer's Active Document regardless
// of DV1/DV2 (model.go's own focusedViewerPane, not a DV1-only path), but
// only the DV1-article/DV2-file arrangement was ever pinned. Same refusal,
// same untouched-DV2... except here it is DV1 (the file) that must stay
// untouched, and DV2 (the article) that must be refused.
func TestEditOnAWikipediaArticleInDV2IsRefusedWithFooterMessage(t *testing.T) {
	filePath := writeFileContent(t, filepath.Join(t.TempDir(), "note.md"), "# A file\n\nUnaffected.\n")
	fileDoc, err := loadDocument(filePath)
	if err != nil {
		t.Fatalf("loadDocument: %v", err)
	}
	m := sized(t, newModel(), 80, 24)
	m = m.navigate(dv1Pane, fileDoc)
	m = m.openDocumentInto(dv2Pane, fixtureArticle())
	m.focus = dv2Pane

	next, cmd := m.Update(keyPress(keys.Edit.Keys()[0]))
	got := concrete(t, next)

	if cmd != nil {
		t.Errorf("cmd = %v, want nil (no $EDITOR handoff for an article)", cmd)
	}
	if got.message == "" {
		t.Error("message is empty, want a report that a Wikipedia article cannot be edited")
	}
	if got.dv2.doc.path != fixtureArticle().path {
		t.Errorf("DV2 Active Document = %q after a refused edit, want unchanged %q", got.dv2.doc.path, fixtureArticle().path)
	}
	if got.dv1.doc.path != filePath {
		t.Errorf("DV1 Active Document = %q after DV2's edit was refused, want unchanged %q", got.dv1.doc.path, filePath)
	}
}

// TestFinishEditReloadsDocumentFromDisk covers D3's "re-renders from disk":
// finishEdit is what M4.1's handoff resumes into once $EDITOR exits, and it
// must reuse loadDocument (per the PROGRESS.md handoff) rather than duplicate
// its read, so a change made while the terminal was away becomes visible.
func TestFinishEditReloadsDocumentFromDisk(t *testing.T) {
	path := filepath.Join(t.TempDir(), "note.md")
	if err := os.WriteFile(path, []byte("# before\n"), 0o644); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}
	d, err := loadDocument(path)
	if err != nil {
		t.Fatalf("loading %s: %v", path, err)
	}
	// M3.1: finishEdit now records recency, which routes through
	// withDocRecord like every other document-store write — a resolvable
	// docsPath is required or that write declines with a message, which
	// would otherwise leak into this test's own "cleared" assertion below.
	storePath := filepath.Join(t.TempDir(), "documents")
	m := sized(t, newModel().withDocs(storePath, loadDocStore(storePath)), 80, 24)
	m = m.openDocument(d)
	m.message = "stale message"

	if err := os.WriteFile(path, []byte("# before\nappended line\n"), 0o644); err != nil {
		t.Fatalf("appending to fixture: %v", err)
	}

	got := m.finishEdit(editFinishedMsg{path: path})

	if got.message != "" {
		t.Errorf("message = %q, want cleared on a successful reload", got.message)
	}
	if !strings.Contains(got.dv1.view(), "appended line") {
		t.Errorf("viewer content does not contain the edit made on disk")
	}
}

// TestFinishEditWithEditorErrorReportsFooterMessage covers the editor itself
// failing (nonzero exit, or failing to start) — the document is left as it
// was rather than an attempted, possibly-wrong reload.
func TestFinishEditWithEditorErrorReportsFooterMessage(t *testing.T) {
	d, err := loadDocument(fixturePath)
	if err != nil {
		t.Fatalf("loading %s: %v", fixturePath, err)
	}
	m := sized(t, newModel(), 80, 24)
	m = m.openDocument(d)

	got := m.finishEdit(editFinishedMsg{path: fixturePath, err: os.ErrPermission})

	if got.message == "" {
		t.Error("message is empty, want a report that the edit failed")
	}
	if got.dv1.doc.path != fixturePath {
		t.Errorf("open document = %q, want unchanged after an editor error", got.dv1.doc.path)
	}
}

// TestFinishEditWithVanishedFileReportsFooterMessage covers the document
// disappearing while $EDITOR had it open (e.g. :saveas'd away, or deleted):
// the reload fails, and that is a footer message, same "silence is not an
// error" shape as activate() and applyResolution().
func TestFinishEditWithVanishedFileReportsFooterMessage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "gone.md")
	if err := os.WriteFile(path, []byte("# once here\n"), 0o644); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}
	d, err := loadDocument(path)
	if err != nil {
		t.Fatalf("loading %s: %v", path, err)
	}
	m := sized(t, newModel(), 80, 24)
	m = m.openDocument(d)

	if err := os.Remove(path); err != nil {
		t.Fatalf("removing fixture: %v", err)
	}

	got := m.finishEdit(editFinishedMsg{path: path})

	if got.message == "" {
		t.Error("message is empty, want a report that the document could not be reopened")
	}
}

// TestEditIsViewerOnly pins D141: `e` edits only from a focused Document
// Viewer. On the Files Nav it does nothing at all — no command, no message —
// even with a document open.
func TestEditIsViewerOnly(t *testing.T) {
	d, err := loadDocument(fixturePath)
	if err != nil {
		t.Fatalf("loading %s: %v", fixturePath, err)
	}
	m := sized(t, newModel(), 80, 24)
	m = m.openDocument(d)
	m.focus = filesPane

	next, cmd := m.Update(keyPress(keys.Edit.Keys()[0]))
	got := concrete(t, next)
	if cmd != nil {
		t.Errorf("cmd = %v, want nil: Edit is viewer-only", cmd)
	}
	if got.message != "" {
		t.Errorf("message = %q, want empty", got.message)
	}
}
