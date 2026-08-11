package main

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestPickerActivateStoreWriteFailureSurfacesMessage is M3.5's Picker-path
// variant of M3.4's TestActivateStoreWriteFailureSurfacesMessage (TRIAGE
// T18/D77): activatePickerInto and activateGrepPickerInto each call
// recordViewed (M3.3) the same way activateInto does, so a store write
// failure there must surface the same way — the store's own error, not
// silently overwritten by anything Activate does afterwards, with the
// document still opened (a store failure does not refuse the open itself).
func TestPickerActivateStoreWriteFailureSurfacesMessage(t *testing.T) {
	t.Run("file Result", func(t *testing.T) {
		dir := t.TempDir()
		docPath := writeFileContent(t, filepath.Join(dir, "report.md"), "a report line\n")
		badStore := storeWriteFailurePath(t, dir)

		m := sized(t, newModel().withDocs(badStore, loadDocStore(badStore)), 80, 24)
		m, _ = m.openPicker()
		m.picker.input.SetValue("report")
		m.picker.results = pickerResults{
			query: "report",
			mode:  fileMode,
			scope: activeRootScope,
			files: []pickerFile{{path: docPath}},
		}
		m.picker.selected = 0

		got := press(t, m, keys.Activate.Keys()[0])

		if got.message == "" {
			t.Fatal("message is empty after a store write failure, want the store's own error")
		}
		if !strings.Contains(got.message, "cannot save document store") {
			t.Errorf("message = %q, want it to reflect the store write failure verbatim, not an unrelated overwrite", got.message)
		}
		if got.picker.open {
			t.Error("picker.open = true after Activate, want dismissed (a store failure does not refuse the open itself)")
		}
		if got.dv1.doc.path != docPath {
			t.Errorf("Active Document = %q, want %q", got.dv1.doc.path, docPath)
		}
	})

	t.Run("grep Result", func(t *testing.T) {
		dir := t.TempDir()
		docPath := writeFileContent(t, filepath.Join(dir, "report.md"), "a report line\n")
		badStore := storeWriteFailurePath(t, dir)

		m := sized(t, newModel().withDocs(badStore, loadDocStore(badStore)), 80, 24)
		m, _ = m.openPicker()
		m.picker.mode = grepMode
		m.picker.input.SetValue("report")
		m.picker.results = pickerResults{
			query: "report",
			mode:  grepMode,
			scope: activeRootScope,
			grep:  []grepResult{{file: pickerFile{path: docPath}, line: 1, text: "a report line"}},
		}
		m.picker.selected = 0

		got := press(t, m, keys.Activate.Keys()[0])

		if got.message == "" {
			t.Fatal("message is empty after a store write failure, want the store's own error")
		}
		if !strings.Contains(got.message, "cannot save document store") {
			t.Errorf("message = %q, want it to reflect the store write failure verbatim, not an unrelated overwrite", got.message)
		}
		if got.picker.open {
			t.Error("picker.open = true after Activate, want dismissed (a store failure does not refuse the open itself)")
		}
		if got.dv1.doc.path != docPath {
			t.Errorf("Active Document = %q, want %q", got.dv1.doc.path, docPath)
		}
	})
}
