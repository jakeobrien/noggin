package main

import (
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

// TestDocRecordDefaultsToZeroValue: a path with no record reads as the zero
// value (unstarred, never viewed) rather than requiring an existence check —
// including a path under no session Root at all, since model.docs has no
// notion of Roots either (M1.2's "no attachment to Roots", threaded through
// the model accessor pair).
func TestDocRecordDefaultsToZeroValue(t *testing.T) {
	m := newModel()
	got := m.docRecord(filepath.Join(t.TempDir(), "no-root-covers-this", "orphan.md"))
	if got.Starred || !got.LastViewed.IsZero() {
		t.Errorf("docRecord(unknown) = %+v, want zero value", got)
	}
}

// TestWithDocRecordPersistsImmediately is M1.4's acceptance criterion: a
// write through withDocRecord survives a fresh load of the same store path,
// standing in for a process restart the way launchFixture does for the
// Roots store.
func TestWithDocRecordPersistsImmediately(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "documents")
	doc := filepath.Join(dir, "note.md")

	m := newModel().withDocs(path, loadDocStore(path))
	m = m.withDocRecord(doc, docRecord{Starred: true})
	if m.message != "" {
		t.Fatalf("withDocRecord: unexpected message %q", m.message)
	}

	// Read back through the same model (in-memory accessor).
	if got := m.docRecord(doc); !got.Starred {
		t.Errorf("docRecord after write = %+v, want Starred", got)
	}

	// "Restart": a fresh model loading the same on-disk path sees it too.
	restarted := newModel().withDocs(path, loadDocStore(path))
	if got := restarted.docRecord(doc); !got.Starred {
		t.Errorf("docRecord after restart = %+v, want Starred", got)
	}
}

// TestWithDocRecordDeclinesOnUnresolvableDataDir: matching toggleStar's
// contract for the Roots store, a write attempted with no resolvable data
// directory (docsPath == "") declines the whole update — including the
// in-memory copy — rather than appearing to succeed only to silently revert
// on the next real restart, and leaves a message.
func TestWithDocRecordDeclinesOnUnresolvableDataDir(t *testing.T) {
	m := newModel() // withDocs never called: docsPath == ""
	doc := filepath.Join(t.TempDir(), "note.md")

	m = m.withDocRecord(doc, docRecord{Starred: true})
	if m.message == "" {
		t.Errorf("withDocRecord with no data dir: expected a message, got none")
	}
	if got := m.docRecord(doc); got.Starred {
		t.Errorf("docRecord after declined write = %+v, want unchanged (zero value)", got)
	}
}

// TestWithDocRecordEvicts confirms withDocRecord runs the store through
// evictDocStore (M1.3) before persisting: writing past docStoreCap with
// only unstarred records present drops the oldest one.
func TestWithDocRecordEvicts(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "documents")
	base := time.Now().Add(-time.Hour)

	seed := map[string]docRecord{}
	for i := 0; i < docStoreCap; i++ {
		seed[filepath.Join(dir, "old", strconv.Itoa(i))] = docRecord{
			LastViewed: base.Add(time.Duration(i) * time.Second),
		}
	}
	oldest := filepath.Join(dir, "old", strconv.Itoa(0))

	m := newModel().withDocs(path, seed)
	newDoc := filepath.Join(dir, "new.md")
	m = m.withDocRecord(newDoc, docRecord{LastViewed: time.Now()})
	if m.message != "" {
		t.Fatalf("withDocRecord: unexpected message %q", m.message)
	}

	if len(m.docs) != docStoreCap {
		t.Fatalf("withDocRecord left %d records, want cap %d", len(m.docs), docStoreCap)
	}
	if _, ok := m.docs[oldest]; ok {
		t.Errorf("oldest unstarred record survived a write that should have evicted it")
	}
	if _, ok := m.docs[newDoc]; !ok {
		t.Errorf("newly written record is missing")
	}

	// The eviction is persisted too, not just applied in memory.
	onDisk := loadDocStore(path)
	if len(onDisk) != docStoreCap {
		t.Errorf("on-disk store has %d records, want cap %d", len(onDisk), docStoreCap)
	}
}
