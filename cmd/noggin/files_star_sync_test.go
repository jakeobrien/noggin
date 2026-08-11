package main

import (
	"path/filepath"
	"strings"
	"testing"

	"noggin/internal/tree"
)

// TestFilesNavMarkerSyncsWhenStarredFromDocumentViewer is M2.7's acceptance
// criterion (TRIAGE T7 / D70): before this fix, only toggleDocStar (the
// Files Nav's own Star action) refreshed m.files.starred, so a star flipped
// from a focused Document Viewer's Active Document left that same file's
// Files Nav row showing a stale marker until the pane's next
// openFolder/refresh. Starring from DV1 must update the row immediately,
// with no manual refresh in between.
func TestFilesNavMarkerSyncsWhenStarredFromDocumentViewer(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "documents")

	m := sized(t, newModel().withDocs(storePath, loadDocStore(storePath)), 80, 24)
	m = m.openFolder(fixtureTree())

	// The README.md fixture row and the Active Document opened into DV1
	// share the same path, so starring one is starring the other.
	doc := document{path: "/root/README.md", src: []byte("# readme")}
	m = m.openDocument(doc)
	if m.focus != dv1Pane {
		t.Fatalf("setup: focus = %v, want dv1Pane", m.focus)
	}

	before := lineContaining(strings.Split(m.files.view(), "\n"), "README.md")
	if strings.Contains(before, "★") {
		t.Fatalf("setup: README.md already marked before starring")
	}

	m = press(t, m, keys.Star.Keys()[0])
	if got := m.docRecord(doc.path); !got.Starred {
		t.Fatalf("docRecord(README.md) after DV1 star = %+v, want Starred", got)
	}

	after := lineContaining(strings.Split(m.files.view(), "\n"), "README.md")
	if !strings.Contains(after, "★ README.md") {
		t.Errorf("Files Nav README.md row after starring from DV1 = %q, want a star marker with no manual refresh", after)
	}

	// Un-starring from DV1 clears the same row's marker straight back.
	m = press(t, m, keys.Star.Keys()[0])
	cleared := lineContaining(strings.Split(m.files.view(), "\n"), "README.md")
	if strings.Contains(cleared, "★") {
		t.Errorf("Files Nav README.md row after un-starring from DV1 = %q, want no star marker", cleared)
	}
}

// TestFilesNavMarkerSyncsWhenStarredFromPicker is M2.7's Picker half: the
// same stale-marker gap existed for toggleStarPicker, since it also wrote
// through withDocRecord without refreshing m.files.starred.
func TestFilesNavMarkerSyncsWhenStarredFromPicker(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "documents")
	path := writeFileContent(t, filepath.Join(dir, "report.md"), "report")

	treeRoot := &tree.Node{
		Name: filepath.Base(dir), Path: dir, IsDir: true,
		Children: []*tree.Node{
			{Name: "report.md", Path: path},
		},
	}

	m := sized(t, newModel().withDocs(storePath, loadDocStore(storePath)), 80, 24).
		withRipgrep(ripgrepAvailability{available: true})
	m = m.openFolder(treeRoot)

	before := lineContaining(strings.Split(m.files.view(), "\n"), "report.md")
	if strings.Contains(before, "★") {
		t.Fatalf("setup: report.md already marked before starring")
	}

	m, _ = m.openPicker()
	m.picker.input.SetValue("report")
	m.picker.results = pickerResults{
		query: "report",
		mode:  fileMode,
		scope: activeRootScope,
		files: []pickerFile{{root: root{path: dir}, path: path}},
	}
	m.picker.selected = 0

	m = press(t, m, keys.PickerStar.Keys()[0])
	if got := m.docRecord(path); !got.Starred {
		t.Fatalf("docRecord(report.md) after Picker star = %+v, want Starred", got)
	}

	after := lineContaining(strings.Split(m.files.view(), "\n"), "report.md")
	if !strings.Contains(after, "★ report.md") {
		t.Errorf("Files Nav report.md row after starring from Picker = %q, want a star marker with no manual refresh", after)
	}
}
