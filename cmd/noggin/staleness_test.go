package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"noggin/internal/tree"
)

// --- staleBaseline / snapshotStaleness / stale() (S10, C2a/M10.1) ---------

func walkFor(t *testing.T, root string) *tree.Node {
	t.Helper()
	node, err := tree.Walk(root)
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	return node
}

// TestStaleBaselineNotStaleWhenNothingChanged is the negative control every
// positive case below needs: a baseline snapshotted and immediately
// compared against the same, untouched disk state reports not-stale.
func TestStaleBaselineNotStaleWhenNothingChanged(t *testing.T) {
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	sub := mkdir(t, filepath.Join(root, "sub"))
	writeFile(t, filepath.Join(sub, "keep.md"))
	docPath := writeFile(t, filepath.Join(root, "doc.md"))

	b := snapshotStaleness(walkFor(t, root), docPath)
	if b.stale() {
		t.Error("stale() = true immediately after snapshotting an untouched tree, want false")
	}
}

// TestStaleBaselineDetectsFileAddedInsideKnownNestedDirectory pins S10's
// core claim without a re-walk: adding a file inside an already-known
// subdirectory bumps that directory's own mtime, which a Stat against the
// path already recorded in dirModTimes picks up — no os.ReadDir involved.
func TestStaleBaselineDetectsFileAddedInsideKnownNestedDirectory(t *testing.T) {
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	sub := mkdir(t, filepath.Join(root, "sub"))
	writeFile(t, filepath.Join(sub, "keep.md"))

	b := snapshotStaleness(walkFor(t, root), "")
	if b.stale() {
		t.Fatalf("setup: baseline already stale before any change")
	}

	bumpMtimeForward(t, sub) // out-of-process create can land the same wall-clock second
	writeFile(t, filepath.Join(sub, "new.md"))

	if !b.stale() {
		t.Error("stale() = false after a new file appeared inside a known subdirectory, want true")
	}
}

// TestStaleBaselineDetectsDirectoryRemoved covers the other tree-shape
// change S10 must not miss: a whole known subdirectory vanishing (its own
// Stat now fails) rather than merely gaining an entry.
func TestStaleBaselineDetectsDirectoryRemoved(t *testing.T) {
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	sub := mkdir(t, filepath.Join(root, "sub"))
	writeFile(t, filepath.Join(sub, "keep.md"))

	b := snapshotStaleness(walkFor(t, root), "")
	if err := os.RemoveAll(sub); err != nil {
		t.Fatalf("RemoveAll: %v", err)
	}

	if !b.stale() {
		t.Error("stale() = false after a known subdirectory was removed, want true")
	}
}

// TestStaleBaselineDetectsActiveDocumentEditedInPlace is S10's Active
// Document half: an in-place rewrite (no rename, so no directory mtime
// bump at all) is still caught because the document's own path is statted
// directly on every check, independent of the directory-mtime heuristic.
func TestStaleBaselineDetectsActiveDocumentEditedInPlace(t *testing.T) {
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	docPath := writeFileContent(t, filepath.Join(root, "doc.md"), "one\n")

	b := snapshotStaleness(walkFor(t, root), docPath)
	if b.stale() {
		t.Fatalf("setup: baseline already stale before any change")
	}

	bumpMtimeForward(t, docPath)
	if err := os.WriteFile(docPath, []byte("one\ntwo\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if !b.stale() {
		t.Error("stale() = false after the Active Document was edited in place, want true")
	}
}

// TestStaleBaselineDetectsActiveDocumentDeleted matches refresh's own
// "no longer exists" case (refresh_test.go) for the indicator's own check.
func TestStaleBaselineDetectsActiveDocumentDeleted(t *testing.T) {
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	docPath := writeFile(t, filepath.Join(root, "doc.md"))

	b := snapshotStaleness(walkFor(t, root), docPath)
	if err := os.Remove(docPath); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	if !b.stale() {
		t.Error("stale() = false after the Active Document was deleted, want true")
	}
}

// TestStaleBaselineIgnoresChangesOutsideTheKnownTree is the flip side of the
// nested-addition case: a change to a file that lives outside the snapshotted
// tree and isn't the Active Document must never register — the baseline
// only ever looks at paths it was actually given.
func TestStaleBaselineIgnoresChangesOutsideTheKnownTree(t *testing.T) {
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	writeFile(t, filepath.Join(root, "doc.md"))
	elsewhere := mkdir(t, filepath.Join(t.TempDir(), "elsewhere"))

	b := snapshotStaleness(walkFor(t, root), "")
	writeFile(t, filepath.Join(elsewhere, "unrelated.md"))

	if b.stale() {
		t.Error("stale() = true after a change entirely outside the snapshotted tree, want false")
	}
}

// bumpMtimeForward nudges path's mtime visibly forward before the caller
// makes its own out-of-process change, so the assertion cannot pass merely
// by coincidence of two operations landing in the same filesystem-mtime
// tick (some filesystems have 1s resolution).
func bumpMtimeForward(t *testing.T, path string) {
	t.Helper()
	past := time.Now().Add(-2 * time.Second)
	if err := os.Chtimes(path, past, past); err != nil {
		t.Fatalf("Chtimes: %v", err)
	}
}

// --- model integration: resetStaleBaseline / runStaleCheck / applyStaleResult ---

// TestOpenFolderAndOpenDocumentEstablishAndAdvanceTheBaseline is the wiring
// claim: every fresh read of the Active Root's tree or the Active Document
// (re)baselines and opens a new generation, so a check dispatched against
// an old one is later discarded (see the generation test below).
func TestOpenFolderAndOpenDocumentEstablishAndAdvanceTheBaseline(t *testing.T) {
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	docPath := writeFile(t, filepath.Join(root, "doc.md"))

	m := newModel()
	if m.staleGen != 0 {
		t.Fatalf("setup: staleGen = %d before any load, want 0", m.staleGen)
	}

	m = m.openFolder(walkFor(t, root))
	genAfterFolder := m.staleGen
	if genAfterFolder == 0 {
		t.Error("staleGen unchanged after openFolder, want it advanced")
	}
	if m.stale {
		t.Error("stale = true immediately after openFolder, want false")
	}

	d, err := loadDocument(docPath)
	if err != nil {
		t.Fatalf("loadDocument: %v", err)
	}
	m = m.openDocument(d)
	if m.staleGen == genAfterFolder {
		t.Error("staleGen unchanged after openDocument, want it advanced again")
	}
	if m.staleBaseline.docPath != docPath {
		t.Errorf("baseline docPath = %q, want %q", m.staleBaseline.docPath, docPath)
	}
}

// TestRunStaleCheckDiscardsASupersededGenerationWithoutStatting is the
// generation-independence claim (staleness.go's own runStaleCheck doc
// comment): a timer that matures after resetStaleBaseline has already run
// (here, by a refresh) still dispatches a real check — against the fresh
// baseline now current, not one captured back when the timer was
// scheduled — rather than discarding and waiting a further full interval
// the way picker.go's debounce discard would. Detected here via the
// generation the dispatched check's own result carries: it matches the
// model's *post-refresh* staleGen, proving the check ran against current
// state.
func TestRunStaleCheckAlwaysDispatchesAgainstTheCurrentBaseline(t *testing.T) {
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	m := sized(t, newModel().openFolder(walkFor(t, root)), 80, 24)

	m = m.refresh() // advances staleGen without any staleness
	genAfterRefresh := m.staleGen

	// A timer matured before this refresh (an empty staleCheckMsg carries
	// no generation of its own — see the type's doc comment) still
	// dispatches, and against the model's current state.
	_, cmd := m.runStaleCheck(staleCheckMsg{})
	if cmd == nil {
		t.Fatal("runStaleCheck returned a nil Cmd, want the dispatched check")
	}
	msg, ok := cmd().(staleResultMsg)
	if !ok {
		t.Fatalf("runStaleCheck's Cmd produced %T, want staleResultMsg", cmd())
	}
	if msg.gen != genAfterRefresh {
		t.Errorf("dispatched check's gen = %d, want %d (the post-refresh generation, proving it checked current state)", msg.gen, genAfterRefresh)
	}
	if msg.stale {
		t.Error("dispatched check reported stale=true against an unchanged, freshly-reset baseline, want false")
	}
}

// TestRunStaleCheckAndApplyStaleResultRoundTripADetectedChange drives the
// real two-message pair synchronously (async-cmd-testing.md's own recipe):
// dispatch against the current generation, invoke the returned (non-Tick)
// Cmd directly to get the result message, then apply it.
func TestRunStaleCheckAndApplyStaleResultRoundTripADetectedChange(t *testing.T) {
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	sub := mkdir(t, filepath.Join(root, "sub"))
	writeFile(t, filepath.Join(sub, "keep.md"))
	m := sized(t, newModel().openFolder(walkFor(t, root)), 80, 24)
	rootNodeBefore := m.files.root

	bumpMtimeForward(t, sub)
	writeFile(t, filepath.Join(sub, "new.md")) // out-of-process change

	next, cmd := m.runStaleCheck(staleCheckMsg{})
	if cmd == nil {
		t.Fatal("runStaleCheck returned a nil Cmd, want the dispatched check")
	}
	msg, ok := cmd().(staleResultMsg)
	if !ok {
		t.Fatalf("dispatched check produced %T, want staleResultMsg", cmd())
	}
	if !msg.stale {
		t.Error("dispatched check reported stale=false after a nested file was added, want true")
	}

	final, reschedule := next.applyStaleResult(msg)
	if !final.stale {
		t.Error("model.stale = false after applying a stale=true result, want true")
	}
	if reschedule == nil {
		t.Error("applyStaleResult returned a nil Cmd, want the next tick scheduled")
	}
	// No re-walk: the Files Nav's tree node is the exact same object
	// throughout the whole check — nothing here ever called tree.Walk.
	if final.files.root != rootNodeBefore {
		t.Error("m.files.root pointer changed across a staleness check, want the tree left untouched")
	}
}

// TestApplyStaleResultDropsASupersededResult is applyStaleResult's own half
// of the generation contract: a result computed against a baseline that
// has since been reset must not flip the indicator on.
func TestApplyStaleResultDropsASupersededResult(t *testing.T) {
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	m := sized(t, newModel().openFolder(walkFor(t, root)), 80, 24)
	staleGenAtDispatchTime := m.staleGen

	m = m.refresh() // advances staleGen

	final, _ := m.applyStaleResult(staleResultMsg{gen: staleGenAtDispatchTime, stale: true})
	if final.stale {
		t.Error("applyStaleResult applied a result from a superseded generation, want it dropped")
	}
}

// TestRefreshClearsTheIndicatorAndRebaselines is S10's "never acts, but
// pressing R does" — the acceptance criterion's own escape hatch: once
// stale, the one way back to not-stale is the user's own refresh.
func TestRefreshClearsTheIndicatorAndRebaselines(t *testing.T) {
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	sub := mkdir(t, filepath.Join(root, "sub"))
	writeFile(t, filepath.Join(sub, "keep.md"))
	m := sized(t, newModel().openFolder(walkFor(t, root)), 80, 24)

	bumpMtimeForward(t, sub)
	writeFile(t, filepath.Join(sub, "new.md"))
	m = m.runStaleCheckSync(t)
	if !m.stale {
		t.Fatalf("setup: stale = false after a detectable change, want true")
	}

	m = press(t, m, keys.Refresh.Keys()[0])
	if m.stale {
		t.Error("stale = true after a refresh, want it cleared")
	}
}

// runStaleCheckSync is the test helper every staleness test above inlines
// by hand for a single case; here it is named once for callers (like
// TestRefreshClearsTheIndicatorAndRebaselines) that just need "run one
// check to completion" without re-asserting the round trip's own steps.
func (m model) runStaleCheckSync(t *testing.T) model {
	t.Helper()
	next, cmd := m.runStaleCheck(staleCheckMsg{})
	if cmd == nil {
		t.Fatal("runStaleCheck returned a nil Cmd")
	}
	msg, ok := cmd().(staleResultMsg)
	if !ok {
		t.Fatalf("runStaleCheck's Cmd produced %T, want staleResultMsg", cmd())
	}
	final, _ := next.applyStaleResult(msg)
	return final
}

// --- Init / footer surfacing -----------------------------------------------

// TestInitSchedulesTheFirstStaleCheck is the "runs for the life of the
// session, not once" claim's starting point.
func TestInitSchedulesTheFirstStaleCheck(t *testing.T) {
	if newModel().Init() == nil {
		t.Error("Init() returned a nil Cmd, want the first staleness tick scheduled")
	}
}

// TestUpdateDispatchesStaleCheckAndResultMessages exercises the same round
// trip as TestRunStaleCheckAndApplyStaleResultRoundTripADetectedChange but
// through model.Update's own message switch, so the wiring in Update itself
// (not just the two methods it calls) is covered.
func TestUpdateDispatchesStaleCheckAndResultMessages(t *testing.T) {
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	sub := mkdir(t, filepath.Join(root, "sub"))
	writeFile(t, filepath.Join(sub, "keep.md"))
	m := sized(t, newModel().openFolder(walkFor(t, root)), 80, 24)

	bumpMtimeForward(t, sub)
	writeFile(t, filepath.Join(sub, "new.md"))

	next, cmd := m.Update(staleCheckMsg{})
	m = concrete(t, next)
	if cmd == nil {
		t.Fatal("Update(staleCheckMsg) returned a nil Cmd")
	}
	msg, ok := cmd().(staleResultMsg)
	if !ok {
		t.Fatalf("Update(staleCheckMsg)'s Cmd produced %T, want staleResultMsg", cmd())
	}

	next, cmd = m.Update(msg)
	m = concrete(t, next)
	if cmd == nil {
		t.Error("Update(staleResultMsg) returned a nil Cmd, want the next tick scheduled")
	}
	if !m.stale {
		t.Error("m.stale = false after Update round-tripped a stale=true result, want true")
	}
}

// TestFooterShowsStaleIndicatorAppendedToPaneHints is S10's UI surface (this
// task's own design choice, see DECISIONS.md): the indicator's fixed
// wording appears alongside the focused pane's normal hints, not in place
// of them.
func TestFooterShowsStaleIndicatorAppendedToPaneHints(t *testing.T) {
	m := sized(t, newModel(), 400, 24)
	m.stale = true

	footer := m.renderFooter(rect{width: 400, height: 1})
	if !strings.Contains(footer, staleMessage) {
		t.Errorf("footer = %q, want it to contain %q", footer, staleMessage)
	}
	if !strings.Contains(footer, formatHint(keys.Refresh)) {
		t.Errorf("footer = %q, want the pane's own hints (e.g. Refresh) still present", footer)
	}
}

// TestFooterMessageTakesPriorityOverStaleIndicator is D22's transient
// message winning over S10's standing one — the indicator would otherwise
// bury a just-produced, more specific notice.
func TestFooterMessageTakesPriorityOverStaleIndicator(t *testing.T) {
	m := sized(t, newModel(), 200, 24)
	m.stale = true
	m.message = "cannot open x: boom"

	footer := m.renderFooter(rect{width: 200, height: 1})
	if strings.Contains(footer, staleMessage) {
		t.Errorf("footer = %q, want the transient message alone, no stale suffix", footer)
	}
	if !strings.Contains(footer, m.message) {
		t.Errorf("footer = %q, want the transient message %q", footer, m.message)
	}
}

// TestFooterShowsStaleIndicatorAtARealisticWidth is M10.2's regression case
// (DECISIONS.md D53): TestFooterShowsStaleIndicatorAppendedToPaneHints and
// TestFooterShowsStaleIndicatorAlongsidePickerHints only ever exercised
// 200/400 cols, wide enough that appendStaleSuffix always had slack left
// over from hintText's greedy fill. Every width this project's own
// test/pty fixtures actually use (80/100/120/132) is narrower than that,
// and the indicator never rendered at any of them before renderFooter
// started reserving room for the suffix ahead of hintText's fill.
func TestFooterShowsStaleIndicatorAtARealisticWidth(t *testing.T) {
	for _, width := range []int{80, 100, 120, 132} {
		m := sized(t, newModel(), width, 24)
		m.stale = true

		footer := m.renderFooter(rect{width: width, height: 1})
		if !strings.Contains(footer, staleMessage) {
			t.Errorf("width %d: footer = %q, want it to contain %q", width, footer, staleMessage)
		}
		if !strings.Contains(footer, formatHint(keys.Quit)) {
			t.Errorf("width %d: footer = %q, want alwaysHints (e.g. quit) still present", width, footer)
		}
	}
}

// TestFooterShowsStaleIndicatorAlongsidePickerHints covers the other hint
// source renderFooter can append the suffix to (D49: "pane hints or the
// open Picker's") — TestFooterShowsStaleIndicatorAppendedToPaneHints only
// exercises the pane-hints branch.
func TestFooterShowsStaleIndicatorAlongsidePickerHints(t *testing.T) {
	m := sized(t, newModel(), 400, 24)
	m.picker.open = true
	m.stale = true

	footer := m.renderFooter(rect{width: 400, height: 1})
	if !strings.Contains(footer, staleMessage) {
		t.Errorf("footer = %q, want it to contain %q", footer, staleMessage)
	}
	if !strings.Contains(footer, formatHint(keys.PickerMode)) {
		t.Errorf("footer = %q, want the Picker's own hints (e.g. PickerMode) still present", footer)
	}
}

// TestSnapshotStalenessUntracksADocumentWhoseStatFails is D49/the fix for
// a narrow TOCTOU window: a docPath that cannot be statted at snapshot
// time must be left untracked (docPath == ""), not half-recorded with a
// zero-value ModTime/Size that could never compare equal to the real
// file's on a later check and would pin stale() permanently true.
func TestSnapshotStalenessUntracksADocumentWhoseStatFails(t *testing.T) {
	root := mkdir(t, filepath.Join(t.TempDir(), "root"))
	missing := filepath.Join(root, "gone.md") // never created

	b := snapshotStaleness(walkFor(t, root), missing)
	if b.docPath != "" {
		t.Errorf("docPath = %q after snapshotting an unstattable document, want \"\" (untracked)", b.docPath)
	}
	if b.stale() {
		t.Error("stale() = true for a baseline with no trackable document and an untouched tree, want false")
	}
}

// TestAppendStaleSuffixOmitsWhenItWouldOverflow is renderFooter's own
// truncate-gracefully note: a terminal too narrow for hints plus the
// indicator shows the hints alone rather than truncating either.
func TestAppendStaleSuffixOmitsWhenItWouldOverflow(t *testing.T) {
	text := "Files · enter open"
	got := appendStaleSuffix(text, len(text)) // exactly fits text, no room for the suffix
	if got != text {
		t.Errorf("appendStaleSuffix at exact width = %q, want unchanged %q", got, text)
	}

	wide := appendStaleSuffix(text, len(text)+len(" · "+staleMessage))
	if wide == text {
		t.Errorf("appendStaleSuffix with room to spare left text unchanged, want the suffix appended")
	}
}
