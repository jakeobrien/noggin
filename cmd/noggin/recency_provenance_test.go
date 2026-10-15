package main

import (
	"path/filepath"
	"testing"
	"time"

	"noggin/internal/tree"
)

// TestActivateRecordsLastViewedRegardlessOfRootProvenance is M3.2's own
// acceptance criterion, verbatim: S4's "regardless of whether its Root is
// starred, ad-hoc, or no longer in the session" — M3.1 already built
// recordViewed/docRecord as a plain path-keyed write with no Root
// attachment (ADR-0005), so this test exists to pin that guarantee at the
// Activate call site itself rather than only at the accessor level
// (TestWithDocRecordPersistsImmediately), for the two provenance cases S4
// names beyond the plain-Root one recency_test.go already covers.
func TestActivateRecordsLastViewedRegardlessOfRootProvenance(t *testing.T) {
	t.Run("ad-hoc Root", func(t *testing.T) {
		// An ad-hoc Root (roots.go's own doc comment: "a launch argument
		// produces an unstarred ad-hoc root") is still a session Root — it
		// appears in m.roots, just unstarred — so this exercises S4's
		// "ad-hoc" clause distinctly from the "no longer in the session"
		// one below.
		dir := t.TempDir()
		project := mkdir(t, filepath.Join(dir, "project"))
		docPath := writeFileContent(t, filepath.Join(project, "doc.md"), "# Doc\n")
		docsPath := filepath.Join(dir, "documents")

		node, err := tree.Walk(project, tree.Options{Filter: tree.FilterMarkdown})
		if err != nil {
			t.Fatalf("tree.Walk: %v", err)
		}
		m := sized(t, newModel().withDocs(docsPath, loadDocStore(docsPath)).openFolder(node), 80, 24)
		m.roots = m.roots.setRoots([]root{{path: project, starred: false}}, 0)
		if m.files.selectedNode() == nil || m.files.selectedNode().Path != docPath {
			t.Fatalf("setup: selected node = %+v, want %q", m.files.selectedNode(), docPath)
		}

		before := time.Now()
		m = press(t, m, keys.Activate.Keys()[0])

		if m.message != "" {
			t.Fatalf("Activate: unexpected message %q", m.message)
		}
		got := m.docRecord(docPath).LastViewed
		if got.IsZero() || got.Before(before) {
			t.Errorf("LastViewed after Activate from an ad-hoc Root = %v, want a timestamp at/after %v", got, before)
		}
		if reloaded := loadDocStore(docsPath); reloaded[docPath].LastViewed.IsZero() {
			t.Error("LastViewed did not survive a restart")
		}
	})

	t.Run("Root not in the session at all", func(t *testing.T) {
		// The Files Nav's Active Root (m.files.root, set here directly via
		// openFolder, the same pattern root_lifecycle_docs_test.go's own
		// M2.6 test uses) need not be a member of m.roots at all — the two
		// are independent (ADR-0005). m.roots is
		// left holding an entirely unrelated Root, standing in for "this
		// document's Root left the session".
		dir := t.TempDir()
		project := mkdir(t, filepath.Join(dir, "project"))
		docPath := writeFileContent(t, filepath.Join(project, "doc.md"), "# Doc\n")
		elsewhere := mkdir(t, filepath.Join(dir, "elsewhere"))
		docsPath := filepath.Join(dir, "documents")

		node, err := tree.Walk(project, tree.Options{Filter: tree.FilterMarkdown})
		if err != nil {
			t.Fatalf("tree.Walk: %v", err)
		}
		m := sized(t, newModel().withDocs(docsPath, loadDocStore(docsPath)).openFolder(node), 80, 24)
		m.roots = m.roots.setRoots([]root{{path: elsewhere, starred: true}}, 0)
		if m.files.selectedNode() == nil || m.files.selectedNode().Path != docPath {
			t.Fatalf("setup: selected node = %+v, want %q", m.files.selectedNode(), docPath)
		}

		before := time.Now()
		m = press(t, m, keys.Activate.Keys()[0])

		if m.message != "" {
			t.Fatalf("Activate: unexpected message %q", m.message)
		}
		got := m.docRecord(docPath).LastViewed
		if got.IsZero() || got.Before(before) {
			t.Errorf("LastViewed after Activate from a Root absent from the session = %v, want a timestamp at/after %v", got, before)
		}
		if reloaded := loadDocStore(docsPath); reloaded[docPath].LastViewed.IsZero() {
			t.Error("LastViewed did not survive a restart")
		}
	})
}
