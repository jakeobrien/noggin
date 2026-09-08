package main

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"noggin/internal/tree"
)

// starredFixtureFile creates a real file at dir/name, mirroring
// recentFixtureFile — starredRows stats the disk directly (buildBrowseRow's
// broken field), so a fixture path must exist unless the test wants a
// broken row on purpose.
func starredFixtureFile(t *testing.T, dir, name string) string {
	t.Helper()
	return writeFileContent(t, filepath.Join(dir, name), "content\n")
}

// TestStarredRowsActiveRootScopeOnlyShowsThatRoot is M4.3's own acceptance
// wording, first clause: Active Source reads only the Starred documents under
// the active Root, never another Root's or an orphan's.
func TestStarredRowsActiveRootScopeOnlyShowsThatRoot(t *testing.T) {
	dir := t.TempDir()
	rootA := mkdir(t, filepath.Join(dir, "root-a"))
	rootB := mkdir(t, filepath.Join(dir, "root-b"))
	aDoc := starredFixtureFile(t, rootA, "a.md")
	bDoc := starredFixtureFile(t, rootB, "b.md")

	docs := map[string]docRecord{
		aDoc: {Starred: true},
		bDoc: {Starred: true},
	}
	roots := []root{{path: rootA}, {path: rootB}}

	rows := starredRows(docs, roots, rootA, activeRootBrowseScope)
	if len(rows) != 1 || rows[0].path != aDoc {
		t.Errorf("Active Source scope rows = %+v, want exactly [%q]", rows, aDoc)
	}
}

// TestStarredRowsAllStarredIncludesOrphansAndOtherRoots is M4.3's own
// acceptance wording: a starred document under no Root appears in All
// Starred (and, per the next test, in Orphans), alongside every Root's own
// starred documents.
func TestStarredRowsAllStarredIncludesOrphansAndOtherRoots(t *testing.T) {
	dir := t.TempDir()
	rootA := mkdir(t, filepath.Join(dir, "root-a"))
	orphanDir := mkdir(t, filepath.Join(dir, "elsewhere"))
	aDoc := starredFixtureFile(t, rootA, "a.md")
	orphanDoc := starredFixtureFile(t, orphanDir, "orphan.md")

	docs := map[string]docRecord{
		aDoc:      {Starred: true},
		orphanDoc: {Starred: true},
	}
	roots := []root{{path: rootA}}

	rows := starredRows(docs, roots, rootA, allStarredBrowseScope)
	got := map[string]browseRow{}
	for _, r := range rows {
		got[r.path] = r
	}
	if len(got) != 2 {
		t.Fatalf("All Starred rows = %+v, want exactly 2", rows)
	}
	if r, ok := got[aDoc]; !ok || r.orphan || r.rootLabel != "root-a" {
		t.Errorf("a.md row = %+v, want rootLabel %q, orphan false", got[aDoc], "root-a")
	}
	if r, ok := got[orphanDoc]; !ok || !r.orphan || r.orphanDir != orphanDir {
		t.Errorf("orphan.md row = %+v, want orphan true, orphanDir %q", got[orphanDoc], orphanDir)
	}
}

// TestStarredRowsOrphansScopeOnlyShowsOrphans confirms Orphans' own
// membership (S5, PLAN.md's pre-agreed cut 3, still carried by the model
// regardless — browse.go's modeScopes): only the no-Root bucket, never a
// Root-covered starred document.
func TestStarredRowsOrphansScopeOnlyShowsOrphans(t *testing.T) {
	dir := t.TempDir()
	rootA := mkdir(t, filepath.Join(dir, "root-a"))
	orphanDir := mkdir(t, filepath.Join(dir, "elsewhere"))
	aDoc := starredFixtureFile(t, rootA, "a.md")
	orphanDoc := starredFixtureFile(t, orphanDir, "orphan.md")

	docs := map[string]docRecord{
		aDoc:      {Starred: true},
		orphanDoc: {Starred: true},
	}
	roots := []root{{path: rootA}}

	rows := starredRows(docs, roots, rootA, orphansBrowseScope)
	if len(rows) != 1 || rows[0].path != orphanDoc {
		t.Errorf("Orphans scope rows = %+v, want exactly [%q]", rows, orphanDoc)
	}
}

// TestStarredRowsExcludesUnstarredRecords confirms a record present in the
// store but not Starred never appears in any scope, regardless of its
// LastViewed.
func TestStarredRowsExcludesUnstarredRecords(t *testing.T) {
	dir := t.TempDir()
	rootDir := mkdir(t, filepath.Join(dir, "root"))
	unstarred := starredFixtureFile(t, rootDir, "unstarred.md")

	docs := map[string]docRecord{unstarred: {LastViewed: time.Now()}}
	roots := []root{{path: rootDir}}

	if rows := starredRows(docs, roots, rootDir, activeRootBrowseScope); len(rows) != 0 {
		t.Errorf("Active Source scope rows = %+v, want none (never starred)", rows)
	}
	if rows := starredRows(docs, roots, rootDir, allStarredBrowseScope); len(rows) != 0 {
		t.Errorf("All Starred scope rows = %+v, want none (never starred)", rows)
	}
}

// TestStarredRowsIncludesNeverViewedRecords is S1's "Starred and LastViewed
// are independent fields" applied to Starred's own list: a starred document
// with a zero LastViewed still appears — unlike Recent, which requires a
// non-zero LastViewed (recentRows' own TestRecentRowsIgnoresNeverViewedRecords).
func TestStarredRowsIncludesNeverViewedRecords(t *testing.T) {
	dir := t.TempDir()
	rootDir := mkdir(t, filepath.Join(dir, "root"))
	neverViewed := starredFixtureFile(t, rootDir, "never.md")

	docs := map[string]docRecord{neverViewed: {Starred: true}}
	roots := []root{{path: rootDir}}

	rows := starredRows(docs, roots, rootDir, activeRootBrowseScope)
	if len(rows) != 1 || rows[0].path != neverViewed {
		t.Errorf("Active Source scope rows = %+v, want exactly [%q]", rows, neverViewed)
	}
}

// TestStarredRowsMarksBrokenFileRatherThanOmitting is S3's own clause: a
// starred document whose file no longer exists on disk is still listed and
// marked broken, the opposite of Recent's silent omission
// (TestRecentRowsOmitsDeletedFile).
func TestStarredRowsMarksBrokenFileRatherThanOmitting(t *testing.T) {
	dir := t.TempDir()
	rootDir := mkdir(t, filepath.Join(dir, "root"))
	present := starredFixtureFile(t, rootDir, "present.md")
	gone := filepath.Join(rootDir, "gone.md") // never created on disk

	docs := map[string]docRecord{
		present: {Starred: true},
		gone:    {Starred: true},
	}
	roots := []root{{path: rootDir}}

	rows := starredRows(docs, roots, rootDir, activeRootBrowseScope)
	if len(rows) != 2 {
		t.Fatalf("starredRows = %+v, want both rows (broken document still listed)", rows)
	}
	byPath := map[string]browseRow{}
	for _, r := range rows {
		byPath[r.path] = r
	}
	if byPath[present].broken {
		t.Errorf("present.md row = %+v, want broken false", byPath[present])
	}
	if !byPath[gone].broken {
		t.Errorf("gone.md row = %+v, want broken true", byPath[gone])
	}
}

// TestStarredRowsActiveRootScopeWithNoActiveRootIsEmpty mirrors Recent's own
// edge case: with no root active, Active Source scope has no bucket to read,
// not the orphan bucket by accident (both key on "").
func TestStarredRowsActiveRootScopeWithNoActiveRootIsEmpty(t *testing.T) {
	dir := t.TempDir()
	orphanDoc := starredFixtureFile(t, dir, "orphan.md")
	docs := map[string]docRecord{orphanDoc: {Starred: true}}

	rows := starredRows(docs, nil, "", activeRootBrowseScope)
	if len(rows) != 0 {
		t.Errorf("starredRows with no active root = %+v, want none", rows)
	}
}

// TestStarredRowsOrdersByName confirms a deterministic order (S5/S6 name no
// timestamp axis for Starred, unlike Recent): rows come back name-sorted.
func TestStarredRowsOrdersByName(t *testing.T) {
	dir := t.TempDir()
	rootDir := mkdir(t, filepath.Join(dir, "root"))
	zebra := starredFixtureFile(t, rootDir, "zebra.md")
	apple := starredFixtureFile(t, rootDir, "apple.md")
	mango := starredFixtureFile(t, rootDir, "mango.md")

	docs := map[string]docRecord{
		zebra: {Starred: true},
		apple: {Starred: true},
		mango: {Starred: true},
	}
	roots := []root{{path: rootDir}}

	rows := starredRows(docs, roots, rootDir, activeRootBrowseScope)
	want := []string{apple, mango, zebra}
	if len(rows) != len(want) {
		t.Fatalf("starredRows returned %d rows, want %d", len(rows), len(want))
	}
	for i, w := range want {
		if rows[i].path != w {
			t.Errorf("row %d = %q, want %q (name order)", i, rows[i].path, w)
		}
	}
}

// TestFilesNavStarredModeRendersFlatRowsWithoutStarMarker is M4.3's own
// model-level wiring check, mirroring M4.2's TestFilesNavRecentModeRendersFlatRows:
// switching to Starred mode shows the flat list, and per S3 never shows the
// star marker (membership already says it).
func TestFilesNavStarredModeRendersFlatRowsWithoutStarMarker(t *testing.T) {
	dir := t.TempDir()
	rootDir := mkdir(t, filepath.Join(dir, "project"))
	starred := starredFixtureFile(t, rootDir, "starred.md")
	starredFixtureFile(t, rootDir, "unstarred.md")

	node, err := tree.Walk(rootDir)
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	docs := map[string]docRecord{starred: {Starred: true}}
	m := sized(t, newModel().withDocs("", docs), 80, 24)
	m.roots = m.roots.setRoots([]root{{path: rootDir}}, 0)
	m = m.openFolder(node)

	m = press(t, m, keys.FilesMode.Keys()[0]) // Tree -> Recent
	m = press(t, m, keys.FilesMode.Keys()[0]) // Recent -> Starred
	if m.files.mode != starredBrowse {
		t.Fatalf("setup: mode = %v, want starredBrowse", m.files.mode)
	}

	view := m.files.view()
	if !strings.Contains(view, "starred.md") {
		t.Errorf("Starred view does not show the starred document:\n%s", view)
	}
	if strings.Contains(view, "unstarred.md") {
		t.Errorf("Starred view shows an unstarred document:\n%s", view)
	}
	if !strings.Contains(view, "project") {
		t.Errorf("Starred view does not show the Root's base name as the row's label:\n%s", view)
	}
	if strings.Contains(view, "\u2605") {
		t.Errorf("Starred view shows a star marker, want none (membership already says it):\n%s", view)
	}
}

// TestFilesNavStarredModeEmptyShowsAMessage checks Starred mode's own
// resting state: Active Source scope's own M4.4 message.
func TestFilesNavStarredModeEmptyShowsAMessage(t *testing.T) {
	dir := t.TempDir()
	rootDir := mkdir(t, filepath.Join(dir, "project"))
	starredFixtureFile(t, rootDir, "unstarred.md")

	node, err := tree.Walk(rootDir)
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	m := sized(t, newModel().withDocs("", map[string]docRecord{}), 80, 24)
	m.roots = m.roots.setRoots([]root{{path: rootDir}}, 0)
	m = m.openFolder(node)

	m = press(t, m, keys.FilesMode.Keys()[0])
	m = press(t, m, keys.FilesMode.Keys()[0])

	view := normalizedRows(m.files.view())
	want := flatEmptyText(starredBrowse, activeRootBrowseScope)
	if !strings.Contains(view, want) {
		t.Errorf("empty Starred view = %q, want it to contain %q", view, want)
	}
}

// TestFilesNavStarredModeMarksBrokenRow confirms the broken-row rendering
// (S3): a starred document whose file no longer exists shows dimmed with a
// "(missing)" suffix, the same treatment rootsnav.go gives a broken Root.
func TestFilesNavStarredModeMarksBrokenRow(t *testing.T) {
	dir := t.TempDir()
	rootDir := mkdir(t, filepath.Join(dir, "project"))
	present := starredFixtureFile(t, rootDir, "present.md")
	gonePath := filepath.Join(rootDir, "gone.md")

	node, err := tree.Walk(rootDir)
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	docs := map[string]docRecord{
		present:  {Starred: true},
		gonePath: {Starred: true},
	}
	m := sized(t, newModel().withDocs("", docs), 80, 24)
	m.roots = m.roots.setRoots([]root{{path: rootDir}}, 0)
	m = m.openFolder(node)

	m = press(t, m, keys.FilesMode.Keys()[0])
	m = press(t, m, keys.FilesMode.Keys()[0])

	view := m.files.view()
	goneLine := lineContaining(strings.Split(view, "\n"), "gone.md")
	if !strings.Contains(goneLine, "(missing)") {
		t.Errorf("gone.md row = %q, want a \"(missing)\" suffix", goneLine)
	}
}

// TestActivateRefusesBrokenStarredRow is M4.3's own acceptance wording:
// activating a broken star is refused the way a Broken Root is (a footer
// message, no document opened) rather than attempting the load.
func TestActivateRefusesBrokenStarredRow(t *testing.T) {
	dir := t.TempDir()
	rootDir := mkdir(t, filepath.Join(dir, "project"))
	gonePath := filepath.Join(rootDir, "gone.md")

	node, err := tree.Walk(rootDir)
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	docs := map[string]docRecord{gonePath: {Starred: true}}
	m := sized(t, newModel().withDocs("", docs), 80, 24)
	m.roots = m.roots.setRoots([]root{{path: rootDir}}, 0)
	m = m.openFolder(node)

	m = press(t, m, keys.FilesMode.Keys()[0])
	m = press(t, m, keys.FilesMode.Keys()[0])
	if len(m.files.flatRows()) != 1 {
		t.Fatalf("setup: flatRows = %+v, want exactly the broken row", m.files.flatRows())
	}

	m = m.activateInto(dv1Pane)
	if m.dv1.doc.isOpen() {
		t.Errorf("activateInto opened the broken document, want refused")
	}
	if !strings.Contains(m.message, gonePath) || !strings.Contains(m.message, "not activated") {
		t.Errorf("message = %q, want a refusal naming %q", m.message, gonePath)
	}
}

// TestActivateOpensNonBrokenStarredRow confirms activateFlatRowInto's happy
// path: a Starred row whose file exists opens normally, recording recency
// the same way a Tree Activate does (S4).
func TestActivateOpensNonBrokenStarredRow(t *testing.T) {
	dir := t.TempDir()
	rootDir := mkdir(t, filepath.Join(dir, "project"))
	present := starredFixtureFile(t, rootDir, "present.md")
	storePath := filepath.Join(dir, "documents")

	node, err := tree.Walk(rootDir)
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	docs := map[string]docRecord{present: {Starred: true}}
	m := sized(t, newModel().withDocs(storePath, docs), 80, 24)
	m.roots = m.roots.setRoots([]root{{path: rootDir}}, 0)
	m = m.openFolder(node)

	m = press(t, m, keys.FilesMode.Keys()[0])
	m = press(t, m, keys.FilesMode.Keys()[0])

	m = m.activateInto(dv1Pane)
	if !m.dv1.doc.isOpen() || m.dv1.doc.path != present {
		t.Fatalf("activateInto did not open %q into DV1: doc = %+v", present, m.dv1.doc)
	}
	if got := m.docRecord(present).LastViewed; got.IsZero() {
		t.Errorf("docRecord(%s).LastViewed still zero after Activate, want recorded", present)
	}
}

// TestToggleDocStarInStarredModeDropsTheRowImmediately is S2/M4.3's own
// acceptance wording: un-starring a document while it is displayed in a
// Files Nav Starred list removes that row immediately, in the very next
// render — not just when the star flip originates elsewhere.
func TestToggleDocStarInStarredModeDropsTheRowImmediately(t *testing.T) {
	dir := t.TempDir()
	rootDir := mkdir(t, filepath.Join(dir, "project"))
	target := starredFixtureFile(t, rootDir, "target.md")
	storePath := filepath.Join(dir, "documents")

	node, err := tree.Walk(rootDir)
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	docs := map[string]docRecord{target: {Starred: true}}
	m := sized(t, newModel().withDocs(storePath, docs), 80, 24)
	m.roots = m.roots.setRoots([]root{{path: rootDir}}, 0)
	m = m.openFolder(node)

	m = press(t, m, keys.FilesMode.Keys()[0])
	m = press(t, m, keys.FilesMode.Keys()[0])
	if !strings.Contains(m.files.view(), "target.md") {
		t.Fatalf("setup: Starred view does not show target.md:\n%s", m.files.view())
	}

	m = m.toggleDocStar()
	if got := m.docRecord(target); got.Starred {
		t.Fatalf("docRecord after toggleDocStar = %+v, want Starred false", got)
	}
	if strings.Contains(m.files.view(), "target.md") {
		t.Errorf("Starred view after un-starring still shows target.md:\n%s", m.files.view())
	}
}
