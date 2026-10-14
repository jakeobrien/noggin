package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"noggin/internal/tree"
)

// D140: binary files and files over tree.MaxOpenSize are refused with a
// Footer message on every open route. These tests drive each route against a
// Markdown-named file whose content is binary (it passes the Tree/Picker
// listing, so only the open-time check can stop it) or oversize.

func writeBinary(t *testing.T, path string) string {
	t.Helper()
	if err := os.WriteFile(path, []byte("MZ\x00\x01\x02binary"), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeOversize(t *testing.T, path string) string {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := f.Truncate(tree.MaxOpenSize + 1); err != nil {
		t.Fatal(err)
	}
	return path
}

// refusalFixture is a Root holding good.md (open and fine) and a binary
// bad.md, with ripgrep marked available for Picker routes.
func refusalFixture(t *testing.T) (m model, good, bad string) {
	t.Helper()
	dir := mkdir(t, filepath.Join(t.TempDir(), "root"))
	good = writeFileContent(t, filepath.Join(dir, "good.md"), "# good\n")
	bad = writeBinary(t, filepath.Join(dir, "bad.md"))
	node, err := tree.Walk(dir, tree.FilterMarkdown)
	if err != nil {
		t.Fatal(err)
	}
	m = sized(t, newModel().openFolder(node), 80, 24).withRipgrep(ripgrepAvailability{available: true})
	return m, good, bad
}

func wantRefused(t *testing.T, route string, m model, name, reason string) {
	t.Helper()
	want := "cannot open " + name + ": " + reason
	if m.message != want {
		t.Errorf("%s: message = %q, want %q", route, m.message, want)
	}
}

func selectNode(t *testing.T, m model, name string) model {
	t.Helper()
	for i := 0; i < 20 && (m.files.selectedNode() == nil || m.files.selectedNode().Name != name); i++ {
		m = press(t, m, "down")
	}
	if m.files.selectedNode() == nil || m.files.selectedNode().Name != name {
		t.Fatalf("could not select %s", name)
	}
	return m
}

func TestRefuseBinaryAndOversizeLoadDocument(t *testing.T) {
	dir := t.TempDir()
	if _, err := loadDocument(writeBinary(t, filepath.Join(dir, "b.md"))); err != tree.ErrBinary {
		t.Errorf("binary: err = %v, want ErrBinary", err)
	}
	if _, err := loadDocument(writeOversize(t, filepath.Join(dir, "big.md"))); err != tree.ErrTooLarge {
		t.Errorf("oversize: err = %v, want ErrTooLarge", err)
	}
	if got := cannotOpen("/x/y/z.md", tree.ErrTooLarge); got != "cannot open z.md: too large" {
		t.Errorf("cannotOpen = %q", got)
	}
}

func TestRefuseFilesNavActivateAndPeek(t *testing.T) {
	for _, target := range []pane{dv1Pane, dv2Pane} {
		m, _, _ := refusalFixture(t)
		m = selectNode(t, m, "bad.md")
		m, _ = m.activateInto(target)
		wantRefused(t, "activate", m, "bad.md", "binary file")
		if m.viewerFor(target).doc.isOpen() {
			t.Errorf("activate: a document opened despite the refusal")
		}

		m.message = ""
		m, _ = m.togglePeekInto(target)
		wantRefused(t, "peek", m, "bad.md", "binary file")
	}
}

func TestRefuseOversizeFilesNavActivate(t *testing.T) {
	m, _, bad := refusalFixture(t)
	writeOversize(t, bad)
	m = selectNode(t, m, "bad.md")
	m, _ = m.activateInto(dv1Pane)
	wantRefused(t, "activate", m, "bad.md", "too large")
}

// Recent and Starred rows are flat rows that need not be Markdown.
func TestRefuseFlatRowActivateAndPeek(t *testing.T) {
	m, _, _ := refusalFixture(t)
	bin := writeBinary(t, filepath.Join(t.TempDir(), "blob.dat"))
	row := browseRow{path: bin, name: "blob.dat"}

	m, _ = m.activateFlatRowInto(row, dv1Pane)
	wantRefused(t, "recent/starred activate", m, "blob.dat", "binary file")

	m.message = ""
	m, _ = m.togglePeekFlatRowInto(row, dv1Pane, false)
	wantRefused(t, "recent/starred peek", m, "blob.dat", "binary file")
}

func TestRefuseLaunchWithBinaryFile(t *testing.T) {
	bin := writeBinary(t, filepath.Join(t.TempDir(), "blob.dat"))
	m := newModel().applyResolution(resolveFile(bin, filepath.Dir(bin), nil))
	wantRefused(t, "noggin <file>", m, "blob.dat", "binary file")
	if m.dv1.doc.isOpen() {
		t.Error("launch opened a binary file")
	}
}

func TestRefuseHistoryWalkOntoNowBinaryFile(t *testing.T) {
	m, good, bad := refusalFixture(t)
	writeFileContent(t, bad, "# fine for now\n")
	for _, p := range []string{bad, good} {
		d, err := loadDocument(p)
		if err != nil {
			t.Fatal(err)
		}
		m = m.navigate(dv1Pane, d)
	}
	writeBinary(t, bad)
	m = m.historyBack()
	wantRefused(t, "history", m, "bad.md", "binary file")
	if m.dv1.doc.path != good {
		t.Errorf("history walked onto a refused file: DV1 = %q", m.dv1.doc.path)
	}
}

func TestRefuseRefreshWhenOpenFileBecameBinary(t *testing.T) {
	m, good, _ := refusalFixture(t)
	d, err := loadDocument(good)
	if err != nil {
		t.Fatal(err)
	}
	m = m.openDocument(d)
	writeBinary(t, good)
	m = press(t, m, keys.Refresh.Keys()[0])
	wantRefused(t, "refresh", m, "good.md", "binary file")
	if m.dv1.doc.isOpen() {
		t.Error("refresh kept a now-binary document open")
	}
}

func TestRefuseFinishEditOntoBinaryFile(t *testing.T) {
	m, _, bad := refusalFixture(t)
	m = m.finishEdit(editFinishedMsg{pane: dv1Pane, path: bad})
	wantRefused(t, "edit", m, "bad.md", "binary file")
}

func TestRefusePickerFileActivate(t *testing.T) {
	m, _, bad := refusalFixture(t)
	dir := filepath.Dir(bad)
	m, _ = m.openPicker()
	m.picker.input.SetValue("bad")
	m.picker.results = pickerResults{
		query: "bad",
		mode:  fileMode,
		scope: activeRootScope,
		files: []pickerFile{{root: root{path: dir}, path: bad}},
	}
	m.picker.selected = 0
	m, _ = m.activatePickerInto(dv1Pane)
	wantRefused(t, "picker", m, "bad.md", "binary file")
	if m.dv1.doc.isOpen() {
		t.Error("picker opened a binary file")
	}
}

func TestRefusePickerGrepActivate(t *testing.T) {
	m, _, bad := refusalFixture(t)
	dir := filepath.Dir(bad)
	m, _ = m.openPicker()
	m.picker.mode = grepMode
	m.picker.input.SetValue("binary")
	m.picker.results = pickerResults{
		query: "binary",
		mode:  grepMode,
		scope: activeRootScope,
		grep:  []grepResult{{file: pickerFile{root: root{path: dir}, path: bad}, line: 1}},
	}
	m.picker.selected = 0
	m = m.activateGrepPickerInto(dv1Pane)
	wantRefused(t, "grep picker", m, "bad.md", "binary file")
	if m.dv1.doc.isOpen() {
		t.Error("grep picker opened a binary file")
	}
}

func TestPickerPreviewExplainsRefusal(t *testing.T) {
	m, _, bad := refusalFixture(t)
	dir := filepath.Dir(bad)
	m, _ = m.openPicker()
	m.picker.input.SetValue("bad")
	m.picker.results = pickerResults{
		query: "bad", mode: fileMode, scope: activeRootScope,
		files: []pickerFile{{root: root{path: dir}, path: bad}},
	}
	m.picker.selected = 0
	p := m.picker.syncPreview(m.hideDotfiles)
	if got := p.preview.GetContent(); !strings.Contains(got, "binary file") {
		t.Errorf("preview = %q, want it to say binary file", got)
	}
}
