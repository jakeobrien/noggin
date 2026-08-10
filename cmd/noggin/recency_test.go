package main

import (
	"path/filepath"
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

	node, err := tree.Walk(root)
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
