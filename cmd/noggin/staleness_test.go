package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"

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

// TestStaleRowAbsentWhenNotStale is S9/C3/M5.1's own done-when, negative
// half: the row occupies zero height, and the frame contains no trace of
// staleMessage at all, while nothing is stale.
func TestStaleRowAbsentWhenNotStale(t *testing.T) {
	l := computeLayout(80, 24, false, false)
	if !l.staleRow.empty() {
		t.Errorf("staleRow = %+v, want empty (zero height) when not stale", l.staleRow)
	}

	m := sized(t, newModel(), 80, 24)
	if strings.Contains(m.frame(), staleMessage) {
		t.Error("frame contains the stale message while nothing is stale")
	}
}

// TestStaleRowPresentWhenStale is S9/C3/M5.1's own done-when, positive half:
// the row claims its one line and shows staleMessage — driven at 80 cols, a
// width the old Footer-suffix design (D49, appendStaleSuffix) never
// rendered the indicator at all: it only ever fit above ~200-244 cols
// (D53), since hintText's own greedy fill left it no room. A dedicated row
// has no such competition.
func TestStaleRowPresentWhenStale(t *testing.T) {
	l := computeLayout(80, 24, false, true)
	if l.staleRow.height != staleRowHeight {
		t.Errorf("staleRow height = %d, want %d", l.staleRow.height, staleRowHeight)
	}
	if l.staleRow.width != 80 {
		t.Errorf("staleRow width = %d, want 80 (full terminal width)", l.staleRow.width)
	}

	m := sized(t, newModel(), 80, 24)
	m.stale = true
	if !strings.Contains(m.frame(), staleMessage) {
		t.Errorf("frame at 80 cols does not contain %q while stale", staleMessage)
	}
}

// TestStaleRowClaimsSpaceFromTheBody pins the "occupies no space otherwise"
// contract's other side: becoming stale shrinks the body region (viewer
// height) by exactly staleRowHeight, the row's own height, at a fixed
// terminal size — the row is not drawn over the body or the Footer, it is
// budgeted alongside them.
func TestStaleRowClaimsSpaceFromTheBody(t *testing.T) {
	fresh := computeLayout(80, 24, false, false)
	stale := computeLayout(80, 24, false, true)

	if stale.viewer.height != fresh.viewer.height-staleRowHeight {
		t.Errorf("viewer height while stale = %d, want %d (fresh %d minus the row's %d)",
			stale.viewer.height, fresh.viewer.height-staleRowHeight, fresh.viewer.height, staleRowHeight)
	}
	if stale.footer != fresh.footer {
		t.Errorf("footer rect changed by staleness: %+v vs %+v, want unchanged", stale.footer, fresh.footer)
	}
}

// TestStaleRowDoesNotAffectFooterContent is the M5.1 corollary of moving
// the indicator off the Footer entirely: renderFooter's own output for a
// given (focus, message, width) is now identical whether or not m.stale is
// set — there is no more width reservation or suffix append for it to do.
func TestStaleRowDoesNotAffectFooterContent(t *testing.T) {
	for _, width := range []int{80, 100, 120, 132, 400} {
		fresh := sized(t, newModel(), width, 24)
		stale := fresh
		stale.stale = true

		freshFooter := fresh.renderFooter(rect{width: width, height: 1})
		staleFooter := stale.renderFooter(rect{width: width, height: 1})
		if freshFooter != staleFooter {
			t.Errorf("width %d: footer changed by staleness:\nfresh: %q\nstale: %q", width, freshFooter, staleFooter)
		}
	}
}

// TestStaleRowCoexistsWithATransientMessage: unlike the old Footer-suffix
// design (where D22's transient message and S10's standing indicator shared
// one line and the message had to win), the row is its own line now — a
// transient message on the Footer and the stale row both show at once.
func TestStaleRowCoexistsWithATransientMessage(t *testing.T) {
	m := sized(t, newModel(), 80, 24)
	m.stale = true
	m.message = "cannot open x: boom"

	frame := m.frame()
	if !strings.Contains(frame, staleMessage) {
		t.Error("frame does not contain the stale row's message while a transient message is also set")
	}
	if !strings.Contains(frame, m.message) {
		t.Errorf("frame does not contain the transient message %q", m.message)
	}
}

// TestRenderStaleRowTruncatesRatherThanWraps: at a width narrower than
// staleMessage itself, the row truncates to one line (rightTruncate) rather
// than word-wrapping into two — the lipgloss v2 Width-wraps-before-MaxWidth-
// truncates gotcha (lipgloss-v2-row-rendering-gotchas.md) applied here.
func TestRenderStaleRowTruncatesRatherThanWraps(t *testing.T) {
	width := 20
	if width >= lipgloss.Width(staleMessage) {
		t.Fatalf("test width %d must be shorter than staleMessage (%d)", width, lipgloss.Width(staleMessage))
	}
	m := newModel()
	row := m.renderStaleRow(rect{width: width, height: 1})
	if got := lipgloss.Height(row); got != 1 {
		t.Errorf("renderStaleRow height = %d, want 1 (no wrap)", got)
	}
	if got := lipgloss.Width(row); got != width {
		t.Errorf("renderStaleRow width = %d, want %d", got, width)
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

// TestRenderStaleRowFillsWidthEvenWhenTextIsShorter: the row is a highlight
// bar (S9: "highlight color, inverse text"), so it must fill r.width even
// when staleMessage itself is shorter — not just render the bare text and
// leave the rest of the row unstyled.
func TestRenderStaleRowFillsWidthEvenWhenTextIsShorter(t *testing.T) {
	m := newModel()
	width := lipgloss.Width(staleMessage) + 20
	row := m.renderStaleRow(rect{width: width, height: 1})
	if got := lipgloss.Width(row); got != width {
		t.Errorf("renderStaleRow width = %d, want %d (filled, not just the bare text)", got, width)
	}
}
