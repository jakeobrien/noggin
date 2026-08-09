package main

import (
	"testing"
)

// C3/M1.5: the Picker's own home for S3's `ctrl+` targeting modifier
// (keys.ActivateDV2, ctrl+enter/ctrl+l per D55) — reusing the same binding
// the Files Nav's Activate already carries (M1.4/M1.4a), not a new one:
// both chords are raw control bytes, so handlePickerInput's always-typing
// Prompt (S2) never mistakes either for query text, closing TRIAGE T3
// without a Picker-specific carve-out.

// TestPickerActivateDV2LoadsFileResultIntoNewDV2AndMovesFocus is S3's load
// half for a fileMode Result: the modified chord creates DV2, loads the
// previewed Result into it, moves Focus there and dismisses the Picker,
// leaving DV1's own Active Document untouched. Driven once per bound key,
// since D55's `ctrl+l` fallback must do exactly what `ctrl+enter` does
// (mirrors target_test.go's TestActivateDV2LoadsSelectedFileIntoNewDV2AndMovesFocus
// for the Files Nav's own version of this modifier).
func TestPickerActivateDV2LoadsFileResultIntoNewDV2AndMovesFocus(t *testing.T) {
	dir := t.TempDir()
	writeFileContent(t, dir+"/existing.md", "existing\n")
	fileB := dir + "/report.md"
	writeFileContent(t, fileB, "a report line\n")

	for _, k := range keys.ActivateDV2.Keys() {
		t.Run(k, func(t *testing.T) {
			m := openedFixture(t, 80, 24).withRipgrep(ripgrepAvailability{available: true})
			m.roots = m.roots.setRoots([]root{{path: dir}}, 0)
			beforeDV1 := m.dv1.doc.path
			m, _ = m.openPicker()
			m.picker.input.SetValue("report")
			m.picker.results = pickerResults{
				query: "report",
				mode:  fileMode,
				scope: activeRootScope,
				files: []pickerFile{{root: root{path: dir}, path: fileB}},
			}
			m.picker.selected = 0
			if m.dv2 != nil {
				t.Fatal("setup: DV2 exists before the targeting modifier was pressed")
			}

			got := press(t, m, k)

			if got.picker.open {
				t.Errorf("%s: picker.open = true, want dismissed", k)
			}
			if got.dv2 == nil {
				t.Fatalf("%s did not create DV2", k)
			}
			if got.dv2.doc.path != fileB {
				t.Errorf("%s: DV2 Active Document = %q, want %q", k, got.dv2.doc.path, fileB)
			}
			if got.focus != dv2Pane {
				t.Errorf("%s: focus = %v, want dv2Pane", k, got.focus)
			}
			if got.dv1.doc.path != beforeDV1 {
				t.Errorf("%s: DV1 Active Document = %q, want unchanged %q", k, got.dv1.doc.path, beforeDV1)
			}
		})
	}
}

// TestPickerPlainActivateStillTargetsDV1WithActivateDV2Bound is the other
// half of M1.5's criterion: the unmodified Activate is unaffected by
// ActivateDV2 also being live in handlePickerInput's switch — it still
// loads into DV1 and never creates DV2. TestPickerActivateOpensResultAsActiveDocumentOnly
// and TestPickerActivateGrepResultOpensDocumentScrolledToMatchedLine already
// pin plain Activate's DV1 targeting in both Picker modes; this adds the
// "DV2 stays absent" half neither of those checks.
func TestPickerPlainActivateStillTargetsDV1WithActivateDV2Bound(t *testing.T) {
	fileB := t.TempDir() + "/report.md"
	writeFileContent(t, fileB, "a report line\n")

	m := openedFixture(t, 80, 24).withRipgrep(ripgrepAvailability{available: true})
	m, _ = m.openPicker()
	m.picker.input.SetValue("report")
	m.picker.results = pickerResults{
		query: "report",
		mode:  fileMode,
		scope: activeRootScope,
		files: []pickerFile{{path: fileB}},
	}
	m.picker.selected = 0

	got := press(t, m, keys.Activate.Keys()[0])

	if got.dv2 != nil {
		t.Error("plain Activate created DV2")
	}
	if got.dv1.doc.path != fileB {
		t.Errorf("DV1 Active Document = %q, want %q", got.dv1.doc.path, fileB)
	}
	if got.focus != dv1Pane {
		t.Errorf("focus = %v, want dv1Pane", got.focus)
	}
}

// TestPickerActivateDV2TypingLiteralLLStillTypes confirms the fallback
// chord's letter, typed without ctrl, is unaffected: `l`/`L` still reach
// the Prompt as query text (D55's whole point — a raw control byte can
// never be typed into an always-typing Prompt, so no carve-out was needed,
// but the plain letters sharing that key must still work exactly as
// before).
func TestPickerActivateDV2TypingLiteralLLStillTypes(t *testing.T) {
	m := openedFixture(t, 80, 24)
	m, _ = m.openPicker()

	m = typeText(t, m, "lL")

	if got := m.picker.input.Value(); got != "lL" {
		t.Errorf("prompt = %q, want %q", got, "lL")
	}
	if !m.picker.open {
		t.Error("picker.open = false after typing \"lL\", want true")
	}
	if m.dv2 != nil {
		t.Error("typing \"lL\" created DV2, want no-op")
	}
}

// TestPickerActivateDV2GrepResultLoadsIntoNewDV2WithFindState is S3's load
// half for a grepMode Result: activateGrepPickerInto (picker.go) must
// thread the DV2 target through commitQuery/applyCurrentMatch the same way
// activateGrepPicker's plain DV1 path already does — the Committed Find
// state (S2) ends up on DV2, not DV1, and DV1 is left untouched.
func TestPickerActivateDV2GrepResultLoadsIntoNewDV2WithFindState(t *testing.T) {
	requireRipgrep(t)
	dir := t.TempDir()
	const matchLine = 87
	path := numberedLinesFixture(t, dir, "long.md", 100, matchLine, "the target report line")

	m := openedFixture(t, 80, 24).withRipgrep(detectRipgrep())
	m.roots = m.roots.setRoots([]root{{path: dir}}, 0)
	beforeDV1 := m.dv1.doc.path
	m, _ = m.openPicker()
	m.picker.mode = grepMode

	m, seqs := typePickerQuery(t, m, "report")
	m, searches := deliverPickerTicks(t, m, seqs)
	if searches != 1 || len(m.picker.results.grep) != 1 {
		t.Fatalf("setup is broken: %d searches, %d grep results", searches, len(m.picker.results.grep))
	}

	got := press(t, m, keys.ActivateDV2.Keys()[0])

	if got.picker.open {
		t.Error("picker.open = true after ActivateDV2, want dismissed")
	}
	if got.dv2 == nil {
		t.Fatal("ActivateDV2 did not create DV2")
	}
	if got.dv2.doc.path != path {
		t.Errorf("DV2 Active Document = %q, want %q", got.dv2.doc.path, path)
	}
	if got.focus != dv2Pane {
		t.Errorf("focus = %v, want dv2Pane", got.focus)
	}
	if !got.dv2.hasFindMatches() {
		t.Fatal("ActivateDV2 did not leave DV2 with find matches for the grep query")
	}
	if got.dv2.offset() == 0 {
		t.Error("DV2 did not scroll away from the top to reach the matched line")
	}
	if got.dv1.doc.path != beforeDV1 {
		t.Errorf("DV1 Active Document = %q, want unchanged %q", got.dv1.doc.path, beforeDV1)
	}
	if got.dv1.hasFindMatches() {
		t.Error("DV1 has find matches after a grep Result was targeted at DV2")
	}
}
