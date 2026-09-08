package main

import (
	"path/filepath"
	"testing"
)

// TestToggleDocStarPersistsAndUpdatesDisplay is M2.1's acceptance criterion:
// star/un-star the Files Nav's selected file, persisted through the S1
// document store (M1.4's accessor pair) rather than the Roots store
// toggleStar already covers. Rebuilding the model against the same store
// path stands in for a restart, the same shape
// TestWithDocRecordPersistsImmediately already gives the accessor pair
// directly.
func TestToggleDocStarPersistsAndUpdatesDisplay(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "documents")

	m := sized(t, newModel().withDocs(storePath, loadDocStore(storePath)), 80, 24)
	m = m.openFolder(fixtureTree())
	// fixtureTree opens with "docs" selected and collapsed; move to the
	// top-level file "README.md".
	for m.files.selectedNode().Name != "README.md" {
		m.files = m.files.move(1)
	}
	target := m.files.selectedNode().Path

	m = m.toggleDocStar()
	if m.message != "" {
		t.Fatalf("toggleDocStar: unexpected message %q", m.message)
	}
	if got := m.docRecord(target); !got.Starred {
		t.Fatalf("docRecord(%s) after toggleDocStar = %+v, want Starred", target, got)
	}

	// "Restart": a fresh model loading the same on-disk path sees it too.
	restarted := newModel().withDocs(storePath, loadDocStore(storePath))
	if got := restarted.docRecord(target); !got.Starred {
		t.Errorf("docRecord after restart = %+v, want Starred", got)
	}

	// Toggling again un-stars and clears it, on the same model instance
	// (selection unchanged) and confirmed against a fresh reload.
	m = m.toggleDocStar()
	if got := m.docRecord(target); got.Starred {
		t.Error("docRecord after toggling twice still Starred, want false")
	}
	restarted = newModel().withDocs(storePath, loadDocStore(storePath))
	if got := restarted.docRecord(target); got.Starred {
		t.Error("docRecord after restart still Starred, want false")
	}
}

// TestFilesNavStarKeyTogglesDocStar confirms the wiring, not just the
// method: with Focus on the Files Nav, pressing the shared keys.Star chord
// (the same physical key the Sources Nav's own case binds) reaches
// toggleDocStar through model.Update's filesPane case.
func TestFilesNavStarKeyTogglesDocStar(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "documents")

	m := sized(t, newModel().withDocs(storePath, loadDocStore(storePath)), 80, 24)
	m = m.openFolder(fixtureTree())
	for m.files.selectedNode().Name != "README.md" {
		m.files = m.files.move(1)
	}
	m.focus = filesPane
	target := m.files.selectedNode().Path

	next, _ := m.Update(keyPress(keys.Star.Keys()[0]))
	m = next.(model)
	if got := m.docRecord(target); !got.Starred {
		t.Errorf("docRecord(%s) after pressing Star from filesPane = %+v, want Starred", target, got)
	}
}

// TestToggleDocStarOnFolderIsNoOp confirms a folder selection (which carries
// no document record) is left alone, the same "no-op on a directory" shape
// activate() already gives its own selection.
func TestToggleDocStarOnFolderIsNoOp(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "documents")

	m := sized(t, newModel().withDocs(storePath, loadDocStore(storePath)), 80, 24)
	m = m.openFolder(fixtureTree())
	if !m.files.selectedNode().IsDir {
		t.Fatalf("selection = %q, want a directory to exercise the no-op", m.files.selectedNode().Name)
	}
	target := m.files.selectedNode().Path

	before := m.docs
	m = m.toggleDocStar()
	if got := m.docRecord(target); got.Starred {
		t.Errorf("docRecord(%s) after toggleDocStar on a folder = %+v, want unchanged", target, got)
	}
	if len(m.docs) != len(before) {
		t.Errorf("toggleDocStar on a folder changed the store size: got %d records, want %d", len(m.docs), len(before))
	}
}

// TestToggleDocStarWithNoDataDirDeclinesAndMessages mirrors
// TestToggleStarWithNoStorePathDeclinesAndMessages and
// TestWithDocRecordDeclinesOnUnresolvableDataDir: with no resolvable data
// directory (docsPath == "", withDocs never called), the toggle must not
// appear to have flipped, and a footer message explains why.
func TestToggleDocStarWithNoDataDirDeclinesAndMessages(t *testing.T) {
	m := sized(t, newModel(), 80, 24) // docsPath defaults to ""
	m = m.openFolder(fixtureTree())
	for m.files.selectedNode().Name != "README.md" {
		m.files = m.files.move(1)
	}
	target := m.files.selectedNode().Path

	m = m.toggleDocStar()
	if got := m.docRecord(target); got.Starred {
		t.Errorf("docRecord after toggleDocStar with no data dir = %+v, want unchanged (zero value)", got)
	}
	if m.message == "" {
		t.Error("want a footer message explaining the decline, got none")
	}
}
