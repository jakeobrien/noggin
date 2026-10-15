package main

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"noggin/internal/tree"
)

// recencyFixture builds a real temp Root with two files (like peekFixture,
// but with a resolvable document-store path installed via withDocs, since
// M3.1's recording routes through withDocRecord like every other document-
// store write) — the starting position every recording test presses a key
// from.
func recencyFixture(t *testing.T) (m model, storePath string, aPath, bPath string) {
	t.Helper()
	dir := t.TempDir()
	root := mkdir(t, filepath.Join(dir, "root"))
	aPath = writeFileContent(t, filepath.Join(root, "a.md"), "# A\n")
	bPath = writeFileContent(t, filepath.Join(root, "b.md"), "# B\n")
	storePath = filepath.Join(dir, "documents")

	node, err := tree.Walk(root, tree.Options{Filter: tree.FilterMarkdown})
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	m = sized(t, newModel().withDocs(storePath, loadDocStore(storePath)).openFolder(node), 80, 24)
	if m.files.selectedNode().Name != "a.md" {
		t.Fatalf("setup: initial selection = %q, want a.md", m.files.selectedNode().Name)
	}
	return m, storePath, aPath, bPath
}

// TestActivateRecordsLastViewed is S4/M3.1's first case: Activate into DV1
// stamps the activated document's LastViewed, and the write survives a
// restart the same way TestWithDocRecordPersistsImmediately's own accessor-
// level check does.
func TestActivateRecordsLastViewed(t *testing.T) {
	m, storePath, aPath, _ := recencyFixture(t)
	before := time.Now()

	m = press(t, m, keys.Activate.Keys()[0])

	if m.message != "" {
		t.Fatalf("Activate: unexpected message %q", m.message)
	}
	got := m.docRecord(aPath).LastViewed
	if got.IsZero() || got.Before(before) {
		t.Errorf("LastViewed after Activate = %v, want a timestamp at/after %v", got, before)
	}

	restarted := newModel().withDocs(storePath, loadDocStore(storePath))
	if restarted.docRecord(aPath).LastViewed.IsZero() {
		t.Error("LastViewed did not survive a restart")
	}
}

// TestActivateDV2RecordsLastViewed is S4/M3.1's second viewer: the `ctrl+`
// targeting modifier stamps the document it loads into DV2 the same way the
// plain key does for DV1.
func TestActivateDV2RecordsLastViewed(t *testing.T) {
	m, _, aPath, _ := recencyFixture(t)
	before := time.Now()

	m = press(t, m, keys.ActivateDV2.Keys()[0])

	if m.dv2 == nil || m.dv2.doc.path != aPath {
		t.Fatalf("setup: ActivateDV2 did not load %q into DV2", aPath)
	}
	got := m.docRecord(aPath).LastViewed
	if got.IsZero() || got.Before(before) {
		t.Errorf("LastViewed after ActivateDV2 = %v, want a timestamp at/after %v", got, before)
	}
}

// TestFinishEditRecordsLastViewed is S4/M3.1's `$EDITOR` handoff case:
// resuming from an edit stamps the reopened document's LastViewed, the same
// as an Activate.
func TestFinishEditRecordsLastViewed(t *testing.T) {
	m, _, aPath, _ := recencyFixture(t)
	before := time.Now()

	m = m.finishEdit(editFinishedMsg{path: aPath})

	if m.message != "" {
		t.Fatalf("finishEdit: unexpected message %q", m.message)
	}
	got := m.docRecord(aPath).LastViewed
	if got.IsZero() || got.Before(before) {
		t.Errorf("LastViewed after finishEdit = %v, want a timestamp at/after %v", got, before)
	}
}

// storeWriteFailurePath is a docsPath that makes every saveDocStore call it
// (withDocRecord's own write) fail with a real write error distinct from
// M3.1's "no resolvable data directory" refusal (m.docsPath == ""): a
// directory sitting where the store file would be written, so
// os.WriteFile's own open fails (EISDIR) only once MkdirAll(filepath.Dir)
// has already succeeded — the write itself is what fails, not path
// resolution.
func storeWriteFailurePath(t *testing.T, dir string) string {
	t.Helper()
	return mkdir(t, filepath.Join(dir, "documents"))
}

// TestActivateStoreWriteFailureSurfacesMessage is M3.4's own acceptance
// criterion (TRIAGE T13): recordViewed's write failure inside withDocRecord
// must not be silently overwritten by anything Activate does afterwards
// (openDocumentInto touches m.viewer/m.focus alone, never m.message) — the
// store failure is the last thing to touch m.message on this path, so it
// must be exactly what the footer shows once Activate returns, not an
// "untraceable" success/empty message papering over a real write failure.
// The document still opens (S1's "silence is not an error" is about
// refusing to run, not about masking a real error once something did run).
func TestActivateStoreWriteFailureSurfacesMessage(t *testing.T) {
	dir := t.TempDir()
	root := mkdir(t, filepath.Join(dir, "root"))
	aPath := writeFileContent(t, filepath.Join(root, "a.md"), "# A\n")
	badStore := storeWriteFailurePath(t, dir)

	node, err := tree.Walk(root, tree.Options{Filter: tree.FilterMarkdown})
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	m := sized(t, newModel().withDocs(badStore, loadDocStore(badStore)).openFolder(node), 80, 24)

	got := press(t, m, keys.Activate.Keys()[0])

	if got.message == "" {
		t.Fatal("message is empty after a store write failure, want the store's own error")
	}
	if !strings.Contains(got.message, "cannot save document store") {
		t.Errorf("message = %q, want it to reflect the store write failure verbatim, not an unrelated overwrite", got.message)
	}
	if got.dv1.doc.path != aPath {
		t.Errorf("Active Document = %q, want %q (a store failure does not refuse the open itself)", got.dv1.doc.path, aPath)
	}
}

// TestFinishEditStoreWriteFailureSurfacesMessage is M3.4's Edit-side case:
// finishEdit clears m.message to "" on a successful reload *before* calling
// recordViewed (model.go), so a subsequent store write failure must win —
// this pins that ordering rather than a future edit accidentally reversing
// it and letting the cleared "success" state silently swallow the failure.
func TestFinishEditStoreWriteFailureSurfacesMessage(t *testing.T) {
	dir := t.TempDir()
	docPath := writeFileContent(t, filepath.Join(dir, "a.md"), "# before\n")
	badStore := storeWriteFailurePath(t, dir)

	d, err := loadDocument(docPath)
	if err != nil {
		t.Fatalf("loading %s: %v", docPath, err)
	}
	m := sized(t, newModel().withDocs(badStore, loadDocStore(badStore)), 80, 24)
	m = m.openDocument(d)
	m.message = "stale message from before the edit"

	got := m.finishEdit(editFinishedMsg{path: docPath})

	if got.message == "" {
		t.Fatal("message is empty after a store write failure, want the store's own error")
	}
	if !strings.Contains(got.message, "cannot save document store") {
		t.Errorf("message = %q, want it to reflect the store write failure, not the stale pre-edit message or a silently cleared success", got.message)
	}
	if got.dv1.doc.path != docPath {
		t.Errorf("Active Document = %q, want %q (a store failure does not refuse the reload itself)", got.dv1.doc.path, docPath)
	}
}

// TestPeekDoesNotRecordLastViewed is S4/M3.1's stated exception: Peek, plain
// and targeted at DV2, leaves the peeked document's record untouched.
func TestPeekDoesNotRecordLastViewed(t *testing.T) {
	t.Run("Peek", func(t *testing.T) {
		m, _, aPath, _ := recencyFixture(t)

		m = press(t, m, keys.Peek.Keys()[0])

		if !m.dv1.peeking() || m.dv1.peek.doc.path != aPath {
			t.Fatalf("setup: Peek did not show %q", aPath)
		}
		if got := m.docRecord(aPath).LastViewed; !got.IsZero() {
			t.Errorf("LastViewed after Peek = %v, want zero (untouched)", got)
		}
	})
	t.Run("PeekDV2", func(t *testing.T) {
		m, _, aPath, _ := recencyFixture(t)

		m = press(t, m, keys.PeekDV2.Keys()[0])

		if m.dv2 == nil || !m.dv2.peeking() || m.dv2.peek.doc.path != aPath {
			t.Fatalf("setup: PeekDV2 did not show %q", aPath)
		}
		if got := m.docRecord(aPath).LastViewed; !got.IsZero() {
			t.Errorf("LastViewed after PeekDV2 = %v, want zero (untouched)", got)
		}
	})
}
