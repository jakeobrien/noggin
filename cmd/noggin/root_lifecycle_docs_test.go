package main

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"noggin/internal/tree"
)

// TestUnstarringRootLeavesDocumentStoreUntouched is M2.6's own acceptance
// criterion, verbatim: un-starring a Root holding starred and recently-
// viewed documents leaves every document record byte-identical afterwards
// (S2's "nothing about the Root lifecycle destroys document stars", D61).
//
// This does not depend on M3.1 (recording `lastViewed` on Activate): the
// "recently-viewed" record here is seeded directly into the S1 store via
// saveDocStore/withDocs, the same fixture shape TestWithDocRecordEvicts
// already uses for a timestamp-bearing record — M2.6's own claim is about
// the Root lifecycle never touching this store, which model.toggleStar's
// implementation (it writes only to the Roots store, never m.docs) makes
// true regardless of how a record's timestamp got there.
func TestUnstarringRootLeavesDocumentStoreUntouched(t *testing.T) {
	dir := t.TempDir()
	target := mkdir(t, filepath.Join(dir, "project"))
	docsPath := filepath.Join(dir, "documents")
	rootsStorePath := filepath.Join(dir, "roots")

	starredDoc := filepath.Join(target, "starred.md")
	recentDoc := filepath.Join(target, "recent.md")
	seed := map[string]docRecord{
		starredDoc: {Starred: true, LastViewed: time.Now().Add(-time.Hour).Truncate(time.Second)},
		recentDoc:  {Starred: false, LastViewed: time.Now().Add(-time.Minute).Truncate(time.Second)},
	}
	if err := saveDocStore(docsPath, seed); err != nil {
		t.Fatalf("saveDocStore: %v", err)
	}
	// Route the "expected" value through the same disk round-trip the model
	// itself will load (RFC3339Nano parse, dropping any monotonic reading
	// seed's time.Now() carries) so the DeepEqual below compares like with
	// like rather than tripping over time.Time's own internal representation.
	initial := loadDocStore(docsPath)
	before, err := os.ReadFile(docsPath)
	if err != nil {
		t.Fatalf("reading seeded store: %v", err)
	}

	node, err := tree.Walk(target, tree.Options{Filter: tree.FilterMarkdown})
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	m := sized(t, newModel().withDocs(docsPath, initial).withStorePath(rootsStorePath), 80, 24)
	m = m.openFolder(node)
	m.roots = m.roots.setRoots([]root{{path: target}}, -1)
	m.focus = rootsPane

	m = m.toggleStar() // star the Root
	if row, ok := m.roots.selectedRow(); !ok || !row.root.starred {
		t.Fatalf("setup: Root not starred after toggleStar (ok=%v)", ok)
	}
	m = m.toggleStar() // un-star it again
	if row, ok := m.roots.selectedRow(); !ok || row.root.starred {
		t.Fatalf("setup: Root still starred after a second toggleStar (ok=%v)", ok)
	}

	if !reflect.DeepEqual(m.docs, initial) {
		t.Errorf("model.docs after starring then un-starring its Root = %+v, want unchanged %+v", m.docs, initial)
	}
	after, err := os.ReadFile(docsPath)
	if err != nil {
		t.Fatalf("reading store after un-star: %v", err)
	}
	if !bytes.Equal(before, after) {
		t.Errorf("document store file changed after starring/un-starring its Root:\nbefore: %q\nafter:  %q", before, after)
	}
}
