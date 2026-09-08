package main

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"noggin/internal/tree"
	"noggin/internal/wikipedia"
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

	rows := starredRows(docs, roots, rootA, false, activeRootBrowseScope)
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

	rows := starredRows(docs, roots, rootA, false, allStarredBrowseScope)
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

	rows := starredRows(docs, roots, rootA, false, orphansBrowseScope)
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

	if rows := starredRows(docs, roots, rootDir, false, activeRootBrowseScope); len(rows) != 0 {
		t.Errorf("Active Source scope rows = %+v, want none (never starred)", rows)
	}
	if rows := starredRows(docs, roots, rootDir, false, allStarredBrowseScope); len(rows) != 0 {
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

	rows := starredRows(docs, roots, rootDir, false, activeRootBrowseScope)
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

	rows := starredRows(docs, roots, rootDir, false, activeRootBrowseScope)
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

	rows := starredRows(docs, nil, "", false, activeRootBrowseScope)
	if len(rows) != 0 {
		t.Errorf("starredRows with no active root = %+v, want none", rows)
	}
}

// TestStarredRowsWikipediaActiveSourceShowsArticlesByTitle is C6/M5.2's own
// S5 clause: with Wikipedia the Active Source (wikipedia == true), Active
// Source reads the article bucket, not activeRootPath's (always "" while
// Wikipedia is active, filesNav.setWikipedia) — a starred Root-orphan file,
// which shares that same "" key, must never leak in (the cosmetic gap this
// task's own handoff named), and the row shows the article's title, not a
// URL-derived name.
func TestStarredRowsWikipediaActiveSourceShowsArticlesByTitle(t *testing.T) {
	dir := t.TempDir()
	orphanDoc := starredFixtureFile(t, dir, "orphan.md")
	article := fixtureArticle()
	docs := map[string]docRecord{
		orphanDoc:    {Starred: true},
		article.path: {Starred: true, Article: true, Title: article.title},
	}

	rows := starredRows(docs, nil, "", true, activeRootBrowseScope)
	if len(rows) != 1 || rows[0].path != article.path {
		t.Fatalf("Wikipedia Active Source rows = %+v, want exactly the article", rows)
	}
	if rows[0].name != article.title {
		t.Errorf("row.name = %q, want the article's title %q", rows[0].name, article.title)
	}
	if !rows[0].article {
		t.Errorf("row.article = false, want true")
	}
	if rows[0].rootLabel != wikipediaLabel {
		t.Errorf("row.rootLabel = %q, want %q", rows[0].rootLabel, wikipediaLabel)
	}
}

// TestStarredRowsAllStarredIncludesArticlesLabelledWikipediaWithARootActive
// is S6's own clause: with a Root active (not Wikipedia), All Starred still
// includes a starred article alongside the Root's own starred files,
// labelled Wikipedia where a file is labelled by its covering Root's name.
func TestStarredRowsAllStarredIncludesArticlesLabelledWikipediaWithARootActive(t *testing.T) {
	dir := t.TempDir()
	rootA := mkdir(t, filepath.Join(dir, "root-a"))
	aDoc := starredFixtureFile(t, rootA, "a.md")
	article := fixtureArticle()
	docs := map[string]docRecord{
		aDoc:         {Starred: true},
		article.path: {Starred: true, Article: true, Title: article.title},
	}
	roots := []root{{path: rootA}}

	rows := starredRows(docs, roots, rootA, false, allStarredBrowseScope)
	got := map[string]browseRow{}
	for _, r := range rows {
		got[r.path] = r
	}
	if len(got) != 2 {
		t.Fatalf("All Starred rows = %+v, want exactly 2", rows)
	}
	if ar, ok := got[article.path]; !ok || !ar.article || ar.name != article.title || ar.rootLabel != wikipediaLabel {
		t.Errorf("article row = %+v, want name %q labelled %q", ar, article.title, wikipediaLabel)
	}
	if fr, ok := got[aDoc]; !ok || fr.article || fr.rootLabel != "root-a" {
		t.Errorf("a.md row = %+v, want rootLabel %q, article false", fr, "root-a")
	}
}

// TestStarredRowsOrphansExcludesArticles is S5's own "no Orphans; that is a
// Root concept" applied to the value itself, not just its reachability
// (already pinned by TestWikipediaStarredScopeExcludesOrphans): a starred
// article never appears in Orphans even while a Root is active and Orphans
// is reachable — an article's own bucket (wikipediaSourceKey) is distinct
// from the "" orphan bucket a Root-orphan file lands in.
func TestStarredRowsOrphansExcludesArticles(t *testing.T) {
	dir := t.TempDir()
	rootA := mkdir(t, filepath.Join(dir, "root-a"))
	orphanDir := mkdir(t, filepath.Join(dir, "elsewhere"))
	orphanDoc := starredFixtureFile(t, orphanDir, "orphan.md")
	article := fixtureArticle()
	docs := map[string]docRecord{
		orphanDoc:    {Starred: true},
		article.path: {Starred: true, Article: true, Title: article.title},
	}
	roots := []root{{path: rootA}}

	rows := starredRows(docs, roots, rootA, false, orphansBrowseScope)
	if len(rows) != 1 || rows[0].path != orphanDoc {
		t.Errorf("Orphans scope rows = %+v, want exactly [%q] (the article excluded)", rows, orphanDoc)
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

	rows := starredRows(docs, roots, rootDir, false, activeRootBrowseScope)
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

	m, _ = m.activateInto(dv1Pane)
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

	m, _ = m.activateInto(dv1Pane)
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

// TestToggleDocStarInStarredModeDropsAStarredArticleRowImmediately is
// C6/M5.1's own acceptance criterion for the Files Nav half: an already-
// starred article (M1.5's own record shape, Article/Title set the way
// togglePickerStar/toggleActiveDocStar both stamp it) appears in a Starred
// flat list like any starred file, and un-starring it from that row drops
// it immediately, the same shape TestToggleDocStarInStarredModeDropsThe
// RowImmediately already pins for a file. An article can only ever reach a
// Files Nav row already-starred today (Recent's own pathMissing check omits
// every article until M6.1 gives it an article-aware existence test, and
// Starred's own row set is Starred==true by definition) — All Starred scope
// is used here, not Active Source, since a Root is never activated in this
// test (a real title/`Wikipedia`-label row either way, C6/M5.2's own fix to
// the cosmetic gap this test used to pin: buildStarredBrowseRow now renders
// the article's own title, not a URL-derived name).
func TestToggleDocStarInStarredModeDropsAStarredArticleRowImmediately(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "documents")
	article := fixtureArticle()
	docs := map[string]docRecord{article.path: {Starred: true, Article: true, Title: article.title}}

	// 120, not the project's standard 80: at 80 the Files Nav column's own
	// right-hand budget truncates "Wikipedia" (9 cols) to "...ipedia", the
	// same class of width-dependent gotcha D52/TRIAGE T13 already named for
	// the Picker's own Root label.
	m := sized(t, newModel().withDocs(storePath, docs), 120, 24)
	m = press(t, m, keys.FilesMode.Keys()[0])  // Tree -> Recent
	m = press(t, m, keys.FilesMode.Keys()[0])  // Recent -> Starred
	m = press(t, m, keys.FilesScope.Keys()[0]) // Active Source -> All Starred
	if !strings.Contains(m.files.view(), article.title) || !strings.Contains(m.files.view(), "Wikipedia") {
		t.Fatalf("setup: Starred/All Starred view does not show the article by title/Wikipedia label:\n%s", m.files.view())
	}

	m = m.toggleDocStar()
	if m.message != "" {
		t.Fatalf("toggleDocStar on the article row: unexpected message %q", m.message)
	}
	got := m.docRecord(article.path)
	if got.Starred {
		t.Fatalf("docRecord(article) after toggleDocStar = %+v, want Starred false", got)
	}
	if !got.Article || got.Title != article.title {
		t.Errorf("docRecord(article) after un-starring = %+v, want Article/Title preserved", got)
	}
	if strings.Contains(m.files.view(), article.title) {
		t.Errorf("Starred view after un-starring the article still shows it:\n%s", m.files.view())
	}

	restarted := newModel().withDocs(storePath, loadDocStore(storePath))
	if got := restarted.docRecord(article.path); got.Starred {
		t.Errorf("docRecord(article) after restart = %+v, want Starred false", got)
	}
}

// starredArticleWikipediaModel is C6/M5.2's own fixture for the Activate/
// Peek-from-a-flat-row tests below: Wikipedia active (its first visit lands
// on Starred, filesNav.setWikipedia) with exactly one starred article on
// screen, ready to Activate/Peek — the flat-row mirror of article_fetch_
// test.go's own wikipediaModeWithResult (the Picker's Result-row fixture).
func starredArticleWikipediaModel(t *testing.T, fetcher *fakeWikipediaFetcher) model {
	t.Helper()
	storePath := filepath.Join(t.TempDir(), "documents")
	docs := map[string]docRecord{eulerArticle.URL: {Starred: true, Article: true, Title: eulerArticle.Title}}
	m := sized(t, newModel().withDocs(storePath, docs), 80, 24).withWikipediaFetcher(fetcher)
	m = m.activateWikipediaSource()
	if m.files.mode != starredBrowse {
		t.Fatalf("setup: mode = %v, want starredBrowse", m.files.mode)
	}
	rows := m.files.flatRows()
	if len(rows) != 1 || !rows[0].article {
		t.Fatalf("setup: flatRows = %+v, want exactly the starred article", rows)
	}
	return m
}

// TestActivateStarredArticleFlatRowFetchesAndOpens is C6/M5.2's own
// acceptance clause "Activate ... on an article row fetch per M4.2": with
// Wikipedia active and a starred article the sole Starred/Active Source
// row, Activate dispatches fetchArticleInto (row.article, activateFlatRow
// Into) rather than failing loadDocument on the article's canonical URL
// (D129's own landmine) — the same round trip article_fetch_test.go's own
// TestActivateWikipediaResultOpensCanonicalArticleInDV1 already pins for a
// Picker Result, now reached from a Files Nav Starred row instead.
func TestActivateStarredArticleFlatRowFetchesAndOpens(t *testing.T) {
	fetcher := &fakeWikipediaFetcher{articles: map[string]wikipedia.Article{eulerArticle.Title: eulerArticle}}
	m := starredArticleWikipediaModel(t, fetcher)

	after, cmd := m.activateInto(dv1Pane)
	got := runArticleFetch(t, after, cmd)

	if got.dv1.doc.path != eulerArticle.URL {
		t.Errorf("DV1 Active Document = %q, want the canonical URL %q", got.dv1.doc.path, eulerArticle.URL)
	}
	if got.dv1.doc.title != eulerArticle.Title {
		t.Errorf("DV1 Active Document title = %q, want %q", got.dv1.doc.title, eulerArticle.Title)
	}
	if got.focus != dv1Pane {
		t.Errorf("focus = %v after Activate, want dv1Pane", got.focus)
	}
	if rec := got.docRecord(eulerArticle.URL); rec.LastViewed.IsZero() {
		t.Error("LastViewed is zero after Activate completed, want recency recorded (S4)")
	}
	if len(fetcher.calls) != 1 || fetcher.calls[0] != eulerArticle.Title {
		t.Errorf("Fetch calls = %v, want exactly one call for %q", fetcher.calls, eulerArticle.Title)
	}
}

// TestPeekStarredArticleFlatRowFetchesWithoutActivating is C6/M5.2's own
// acceptance clause's Peek half: togglePeekFlatRowInto's new row.article
// branch fetches (peek == true) and starts Peek on completion, without ever
// moving Focus or recording recency (S7's "Peek does not count", the same
// contract article_fetch_test.go's TestPeekWikipediaArticleStartsPeek
// WithoutActivating already pins for a Picker Result).
func TestPeekStarredArticleFlatRowFetchesWithoutActivating(t *testing.T) {
	fetcher := &fakeWikipediaFetcher{articles: map[string]wikipedia.Article{eulerArticle.Title: eulerArticle}}
	m := starredArticleWikipediaModel(t, fetcher)
	beforeFocus := m.focus

	after, cmd := m.togglePeekInto(dv1Pane)
	got := runArticleFetch(t, after, cmd)

	if !got.dv1.peeking() || got.dv1.peek.doc.path != eulerArticle.URL {
		t.Fatalf("togglePeekInto did not peek the article in DV1: peek = %+v", got.dv1.peek.doc)
	}
	if got.dv1.doc.isOpen() {
		t.Errorf("Peek changed DV1's Active Document, want it untouched")
	}
	if got.focus != beforeFocus {
		t.Errorf("focus = %v after Peek, want unchanged %v (Peek never moves Focus)", got.focus, beforeFocus)
	}
	if rec := got.docRecord(eulerArticle.URL).LastViewed; !rec.IsZero() {
		t.Error("LastViewed recorded after Peek, want zero (S4: Peek does not count)")
	}
}
