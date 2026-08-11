package main

import (
	"path/filepath"
	"testing"
	"time"
)

// pickerRecencyFixture is recencyFixture's Picker-side counterpart: a
// resolvable document store (withDocs, same as recencyFixture) plus one
// file and an already-open Picker, ready for a manually-staged Result the
// same way picker_test.go's own Activate tests stage pickerResults
// directly rather than running a live search.
func pickerRecencyFixture(t *testing.T) (m model, storePath, docPath string) {
	t.Helper()
	dir := t.TempDir()
	docPath = writeFileContent(t, filepath.Join(dir, "report.md"), "a report line\n")
	storePath = filepath.Join(dir, "documents")
	m = sized(t, newModel().withDocs(storePath, loadDocStore(storePath)), 80, 24)
	m, _ = m.openPicker()
	return m, storePath, docPath
}

// TestPickerActivateRecordsLastViewed is M3.3's own acceptance criterion:
// activating a File Picker Result and a Grep Picker Result, into both DV1
// and DV2, each stamps the activated document's LastViewed the same way
// the Files Nav's own Activate already does (M3.1's recordViewed) —
// TRIAGE T12 found activatePickerInto/activateGrepPickerInto bypassing
// recordViewed entirely by calling openDocumentInto directly.
func TestPickerActivateRecordsLastViewed(t *testing.T) {
	targets := []struct {
		name   string
		target pane
		key    string
	}{
		{"DV1", dv1Pane, keys.Activate.Keys()[0]},
		{"DV2", dv2Pane, keys.ActivateDV2.Keys()[0]},
	}

	t.Run("file Result", func(t *testing.T) {
		for _, tc := range targets {
			t.Run(tc.name, func(t *testing.T) {
				m, storePath, docPath := pickerRecencyFixture(t)
				m.picker.input.SetValue("report")
				m.picker.results = pickerResults{
					query: "report",
					mode:  fileMode,
					scope: activeRootScope,
					files: []pickerFile{{path: docPath}},
				}
				m.picker.selected = 0
				before := time.Now()

				got := press(t, m, tc.key)

				if got.picker.open {
					t.Fatal("picker.open = true after Activate, want dismissed")
				}
				if v := got.viewerFor(tc.target); v.doc.path != docPath {
					t.Fatalf("setup: Active Document at %v = %q, want %q", tc.target, v.doc.path, docPath)
				}
				ts := got.docRecord(docPath).LastViewed
				if ts.IsZero() || ts.Before(before) {
					t.Errorf("LastViewed after Picker Activate into %v = %v, want a timestamp at/after %v", tc.target, ts, before)
				}
				if reloaded := loadDocStore(storePath); reloaded[docPath].LastViewed.IsZero() {
					t.Error("LastViewed did not survive a restart")
				}
			})
		}
	})

	t.Run("grep Result", func(t *testing.T) {
		for _, tc := range targets {
			t.Run(tc.name, func(t *testing.T) {
				m, storePath, docPath := pickerRecencyFixture(t)
				m.picker.mode = grepMode
				m.picker.input.SetValue("report")
				m.picker.results = pickerResults{
					query: "report",
					mode:  grepMode,
					scope: activeRootScope,
					grep:  []grepResult{{file: pickerFile{path: docPath}, line: 1, text: "a report line"}},
				}
				m.picker.selected = 0
				before := time.Now()

				got := press(t, m, tc.key)

				if got.picker.open {
					t.Fatal("picker.open = true after Activate, want dismissed")
				}
				if v := got.viewerFor(tc.target); v.doc.path != docPath {
					t.Fatalf("setup: Active Document at %v = %q, want %q", tc.target, v.doc.path, docPath)
				}
				ts := got.docRecord(docPath).LastViewed
				if ts.IsZero() || ts.Before(before) {
					t.Errorf("LastViewed after Picker Activate (grep) into %v = %v, want a timestamp at/after %v", tc.target, ts, before)
				}
				if reloaded := loadDocStore(storePath); reloaded[docPath].LastViewed.IsZero() {
					t.Error("LastViewed did not survive a restart")
				}
			})
		}
	})
}
