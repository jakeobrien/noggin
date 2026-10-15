package main

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// renameFixture is a model with one starred Root, focused on the Sources
// Nav, ready to drive a rename through — the same shape
// TestToggleStarPersistsAndUpdatesDisplay (rootsnav_test.go) builds for its
// own S2 store round trip.
func renameFixture(t *testing.T, target string) (model, string) {
	t.Helper()
	storeFile := filepath.Join(t.TempDir(), "config", "roots")
	// A starred Root is, by construction, already in the store (S2/M3.2's
	// toggleStar writes through immediately) — renameStore's own "target not
	// present is a no-op" doc comment leans on that invariant, so the
	// fixture must actually hold it, not merely flip root.starred in memory.
	if err := saveStore(storeFile, []storedRoot{{Path: target}}); err != nil {
		t.Fatal(err)
	}
	m := sized(t, newModel().withStorePath(storeFile), 80, 24)
	m.roots = m.roots.setRoots([]root{{path: target, starred: true}}, 0)
	m.focus = rootsPane
	return m, storeFile
}

// TestRenameEditsSelectedRootNameInline is M7.2's headline acceptance
// check: the rename key opens an inline edit on the selected row, typing
// changes it, Enter commits — the row shows the new name and the S1/S2
// store round-trips it (storedRoot.Name), the same store-persistence
// contract toggleStar's own test pins for the starred mark.
func TestRenameEditsSelectedRootNameInline(t *testing.T) {
	target := mkdir(t, filepath.Join(t.TempDir(), "project"))
	m, storeFile := renameFixture(t, target)

	m = press(t, m, keys.Rename.Keys()[0])
	if !m.roots.renameEditing() {
		t.Fatal("c did not start a rename")
	}
	// startRename seeds the input with the current effective name (S8's
	// default, the folder base name) rather than blank.
	if got := m.roots.renameInput.Value(); got != "project" {
		t.Errorf("seeded rename input = %q, want the current default name %q", got, "project")
	}

	// Replace the seeded value directly (the same SetValue-on-the-textinput
	// convention picker_test.go and friends already use for their own Prompt)
	// rather than simulating a clear: backspace has no keymap.go binding to
	// drive it through (D2), and what is under test here is commit/seed
	// behavior, not keystroke-level editing.
	m.roots.renameInput.SetValue("Work")
	m = press(t, m, "enter")

	if m.roots.renameEditing() {
		t.Error("enter did not commit the rename (still editing)")
	}
	row, ok := m.roots.selectedRow()
	if !ok || row.root.name != "Work" {
		t.Fatalf("selected row after rename: ok=%v name=%q, want %q", ok, row.root.name, "Work")
	}
	if got := row.root.displayName(); got != "Work" {
		t.Errorf("displayName() = %q, want %q", got, "Work")
	}

	got := loadStore(storeFile)
	want := []storedRoot{{Path: target, Name: "Work"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("loadStore after rename = %+v, want %+v", got, want)
	}
}

// TestRenameCancelDiscardsEdit checks S8's "Esc cancels": the Root's Name
// (and the on-disk store) is untouched, and the pane goes back to its
// ordinary display.
func TestRenameCancelDiscardsEdit(t *testing.T) {
	target := mkdir(t, filepath.Join(t.TempDir(), "project"))
	m, storeFile := renameFixture(t, target)

	m = press(t, m, keys.Rename.Keys()[0])
	m = typeText(t, m, "should not stick")
	m = press(t, m, "esc")

	if m.roots.renameEditing() {
		t.Error("esc did not cancel the rename (still editing)")
	}
	row, ok := m.roots.selectedRow()
	if !ok || row.root.name != "" {
		t.Fatalf("selected row after cancel: ok=%v name=%q, want unnamed", ok, row.root.name)
	}
	want := []storedRoot{{Path: target}}
	if got := loadStore(storeFile); !reflect.DeepEqual(got, want) {
		t.Errorf("loadStore after cancel = %+v, want %+v (unchanged)", got, want)
	}
}

// TestRenameEmptyRevertsToDefault is S8's "an empty name reverts to the
// default": clearing a previously-set custom Name and committing brings
// displayName back to the folder's base name, and persists that reversion
// (an empty Name, not the old one) to the store.
func TestRenameEmptyRevertsToDefault(t *testing.T) {
	target := mkdir(t, filepath.Join(t.TempDir(), "project"))
	m, storeFile := renameFixture(t, target)
	m.roots = m.roots.setRoots([]root{{path: target, starred: true, name: "Custom"}}, 0)
	if err := saveStore(storeFile, []storedRoot{{Path: target, Name: "Custom"}}); err != nil {
		t.Fatal(err)
	}

	m = press(t, m, keys.Rename.Keys()[0])
	if got := m.roots.renameInput.Value(); got != "Custom" {
		t.Fatalf("seeded rename input = %q, want the existing custom name %q", got, "Custom")
	}
	m.roots.renameInput.SetValue("")
	m = press(t, m, "enter")

	row, ok := m.roots.selectedRow()
	if !ok || row.root.name != "" {
		t.Fatalf("selected row after empty rename: ok=%v name=%q, want \"\" (default)", ok, row.root.name)
	}
	if got := row.root.displayName(); got != "project" {
		t.Errorf("displayName() = %q, want the folder base name %q", got, "project")
	}
	got := loadStore(storeFile)
	want := []storedRoot{{Path: target}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("loadStore after reverting = %+v, want %+v (Name cleared, not left as %q)", got, want, "Custom")
	}
}

// TestRenameWhitespaceOnlyRevertsToDefault checks the "or all whitespace"
// half of commitRename's own doc comment: a name of only spaces is trimmed
// to empty, the same revert-to-default outcome as a literally empty input.
func TestRenameWhitespaceOnlyRevertsToDefault(t *testing.T) {
	target := mkdir(t, filepath.Join(t.TempDir(), "project"))
	m, _ := renameFixture(t, target)

	m = press(t, m, keys.Rename.Keys()[0])
	m.roots.renameInput.SetValue("   ")
	m = press(t, m, "enter")

	row, ok := m.roots.selectedRow()
	if !ok || row.root.name != "" {
		t.Fatalf("selected row after whitespace-only rename: ok=%v name=%q, want \"\" (default)", ok, row.root.name)
	}
}

// TestRenameRefusedForUnstarredAdhocRoot is S8's other refusal: an ad-hoc
// (unstarred) Root cannot be renamed until it is starred, and the attempt
// is a Footer message, not a silent no-op or a crash.
func TestRenameRefusedForUnstarredAdhocRoot(t *testing.T) {
	target := mkdir(t, filepath.Join(t.TempDir(), "adhoc"))
	m, _ := renameFixture(t, target)
	m.roots = m.roots.setRoots([]root{{path: target}}, 0) // unstarred

	m = press(t, m, keys.Rename.Keys()[0])
	if m.roots.renameEditing() {
		t.Error("rename started on an unstarred ad-hoc root, want refused")
	}
	if m.message == "" {
		t.Error("message is empty after refusing to rename an unstarred root, want a footer message")
	}
}

// TestRenameRefusedForWikipediaSource checks S8's third refusal: the
// Wikipedia row cannot be renamed at all, matching the Footer-message
// treatment its own Star refusal already gets (M1.2).
func TestRenameRefusedForWikipediaSource(t *testing.T) {
	m, _ := renameFixture(t, mkdir(t, filepath.Join(t.TempDir(), "unused")))
	m.roots = m.roots.activateWikipedia()
	m.roots.selected = len(m.roots.rows) // the Wikipedia row

	m = press(t, m, keys.Rename.Keys()[0])
	if m.roots.renameEditing() {
		t.Error("rename started on the Wikipedia row, want refused")
	}
	if !strings.Contains(m.message, "Wikipedia") {
		t.Errorf("message = %q, want it to name the Wikipedia Source", m.message)
	}
}

// TestRenameAllowedAfterStarringGivesDefaultName is S8's own follow-on to
// the ad-hoc refusal: once an ad-hoc Root is starred, it can be renamed, and
// the rename input seeds from "the default name" (the folder base name) —
// not some leftover value — since starring never invents a Name of its own
// (toggleStore always adds a fresh entry unnamed).
func TestRenameAllowedAfterStarringGivesDefaultName(t *testing.T) {
	target := mkdir(t, filepath.Join(t.TempDir(), "adhoc"))
	storeFile := filepath.Join(t.TempDir(), "config", "roots")
	m := sized(t, newModel().withStorePath(storeFile), 80, 24)
	m.roots = m.roots.setRoots([]root{{path: target}}, 0) // unstarred
	m.focus = rootsPane

	m = press(t, m, keys.Rename.Keys()[0])
	if m.roots.renameEditing() {
		t.Fatal("rename started on an unstarred ad-hoc root before it was even starred")
	}

	m = press(t, m, keys.Star.Keys()[0]) // star it (toggleStar)
	row, ok := m.roots.selectedRow()
	if !ok || !row.root.starred {
		t.Fatalf("selected row after star: ok=%v starred=%v, want true", ok, row.root.starred)
	}

	m = press(t, m, keys.Rename.Keys()[0])
	if !m.roots.renameEditing() {
		t.Fatal("rename still refused after starring the ad-hoc root")
	}
	if got := m.roots.renameInput.Value(); got != "adhoc" {
		t.Errorf("seeded rename input after starring = %q, want the default name %q", got, "adhoc")
	}
}

// TestRenameOwnsKeyboardWhileEditing checks the exclusivity M7.2 borrows
// from find (handleFindInput): every key but Cancel/Activate types into the
// rename input instead of acting on its usual binding — here, `s` (Star)
// and `q` (Quit) both go into the text instead of starring or quitting.
func TestRenameOwnsKeyboardWhileEditing(t *testing.T) {
	target := mkdir(t, filepath.Join(t.TempDir(), "project"))
	m, _ := renameFixture(t, target)

	m = press(t, m, keys.Rename.Keys()[0])
	m.roots.renameInput.SetValue("")
	m = press(t, m, "s")
	m = press(t, m, "q")

	if got, want := m.roots.renameInput.Value(), "sq"; got != want {
		t.Errorf("rename input = %q, want %q (s/q typed in, not acted on)", got, want)
	}
	row, ok := m.roots.selectedRow()
	if !ok || row.root.starred != true {
		t.Fatalf("star flipped while renaming: ok=%v starred=%v, want unchanged true", ok, row.root.starred)
	}
}

// TestRenamePersistsAcrossReload is the acceptance criterion in prose: the
// Name persists across sessions (S8). renameFixture's storeFile stands in
// for a restart, the same "fresh loadStore call" shape
// TestToggleStorePersistsAcrossReload (store_test.go) uses for the starred
// mark.
func TestRenamePersistsAcrossReload(t *testing.T) {
	target := mkdir(t, filepath.Join(t.TempDir(), "project"))
	m, storeFile := renameFixture(t, target)

	m = press(t, m, keys.Rename.Keys()[0])
	m.roots.renameInput.SetValue("Work")
	m = press(t, m, "enter")

	// A fresh launch, as a new process would perform (main.go's own
	// sequence): loadStore (standing in for loadStarredRoots, dataDir not
	// being under test control here — dv2_persistence_test.go's own
	// launchFixture makes the identical substitution) -> resolveLaunch ->
	// starredRoots.
	stored := loadStore(storeFile)
	res := resolveLaunch(nil, stored, t.TempDir(), nil)
	if len(res.roots) != 1 || res.roots[0].displayName() != "Work" {
		t.Fatalf("relaunch roots = %+v, want one Root named Work", res.roots)
	}
}

// TestPaneHintsNameRenameOnlyForRootsPane checks the Footer half of M7.2's
// discoverability requirement: rootsPane's hints name it, filesPane's and
// the viewer's do not (only a Root row can ever be renamed).
func TestPaneHintsNameRenameOnlyForRootsPane(t *testing.T) {
	if !strings.Contains(plainHints(hintText(rootsPane, 200)), "rename") {
		t.Errorf("rootsPane hint text = %q, want it to name rename", plainHints(hintText(rootsPane, 200)))
	}
	for _, p := range []pane{filesPane, dv1Pane} {
		if strings.Contains(plainHints(hintText(p, 200)), "rename") {
			t.Errorf("pane %v hint text unexpectedly names rename", p)
		}
	}
}

// TestHelpModalNamesRenameOnlyForRootsPane is help.go's half of the same
// requirement.
func TestHelpModalNamesRenameOnlyForRootsPane(t *testing.T) {
	if !strings.Contains(helpBox(rootsPane), "rename") {
		t.Errorf("rootsPane help modal = %q, want it to name rename", helpBox(rootsPane))
	}
	for _, p := range []pane{filesPane, dv1Pane} {
		if strings.Contains(helpBox(p), "rename") {
			t.Errorf("pane %v help modal unexpectedly names rename", p)
		}
	}
}

// TestRenameFooterReplacesOrdinaryHintsWhileEditing checks the Footer's own
// special case for an in-progress rename (model.renderFooter): the input
// lives inline on the row (not the footer, unlike find), so the footer
// shows a plain confirm/cancel line instead of rootsPane's ordinary hints,
// the same "editing takes over the footer" shape find gives its own query.
func TestRenameFooterReplacesOrdinaryHintsWhileEditing(t *testing.T) {
	target := mkdir(t, filepath.Join(t.TempDir(), "project"))
	m, _ := renameFixture(t, target)

	before := plainHints(m.renderFooter(rect{width: 80, height: 1}))
	if !strings.Contains(before, "star") {
		t.Fatalf("footer before renaming = %q, want the ordinary rootsPane hints", before)
	}

	m = press(t, m, keys.Rename.Keys()[0])
	during := plainHints(m.renderFooter(rect{width: 80, height: 1}))
	if strings.Contains(during, "star") || strings.Contains(during, "open") {
		t.Errorf("footer while renaming = %q, want rootsPane's ordinary hints replaced", during)
	}
	if !strings.Contains(during, "confirm") || !strings.Contains(during, "cancel") {
		t.Errorf("footer while renaming = %q, want it to explain confirm/cancel", during)
	}
}

// TestRenameRendersInlineOnTheRowNotTheFooter is S8's own "inline on its
// row" wording, checked directly: the typed text appears in the Sources Nav
// pane's own rendered view, unlike find's query (which lives in the footer
// only, never in the Document Viewer's own content).
func TestRenameRendersInlineOnTheRowNotTheFooter(t *testing.T) {
	target := mkdir(t, filepath.Join(t.TempDir(), "project"))
	m, _ := renameFixture(t, target)

	m = press(t, m, keys.Rename.Keys()[0])
	m.roots.renameInput.SetValue("Distinctive")

	// ansi.Strip first: the textinput's own cursor rendering (a reverse-video
	// single cell) interleaves ANSI escapes with the typed text (the same
	// "complete rendering, cursor included" find.go's findView comment
	// describes), which would otherwise split "Distinctive" across escape
	// codes and defeat a raw substring check (help_test.go's plainLines does
	// the identical strip for the same reason).
	if view := ansi.Strip(m.roots.view()); !strings.Contains(view, "Distinctive") {
		t.Errorf("Sources Nav view does not show the in-progress rename text inline on the row, got:\n%s", view)
	}
}
