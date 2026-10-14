package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"noggin/internal/tree"
)

// recentFixtureFile creates a real file at dir/name so pathMissing (S6's
// "a record whose file no longer exists is silently omitted") reads it as
// present — recentRows must never be exercised against fixture paths that
// merely look plausible, unlike fixtureTree's Tree-mode fixtures, since it
// stats the disk directly.
func recentFixtureFile(t *testing.T, dir, name string) string {
	t.Helper()
	return writeFileContent(t, filepath.Join(dir, name), "content\n")
}

// TestRecentRowsOrdersByTimestamp is M4.2's own acceptance wording, first
// clause: within one Root, rows come back newest-LastViewed-first.
func TestRecentRowsOrdersByTimestamp(t *testing.T) {
	dir := t.TempDir()
	rootDir := mkdir(t, filepath.Join(dir, "root"))
	oldest := recentFixtureFile(t, rootDir, "oldest.md")
	middle := recentFixtureFile(t, rootDir, "middle.md")
	newest := recentFixtureFile(t, rootDir, "newest.md")

	now := time.Now()
	docs := map[string]docRecord{
		oldest: {LastViewed: now.Add(-2 * time.Hour)},
		middle: {LastViewed: now.Add(-1 * time.Hour)},
		newest: {LastViewed: now},
	}
	roots := []root{{path: rootDir}}

	rows := recentRows(docs, roots, rootDir, false, activeRootBrowseScope)
	want := []string{newest, middle, oldest}
	if len(rows) != len(want) {
		t.Fatalf("recentRows returned %d rows, want %d", len(rows), len(want))
	}
	for i, w := range want {
		if rows[i].path != w {
			t.Errorf("row %d = %q, want %q (newest-first order)", i, rows[i].path, w)
		}
	}
}

// TestRecentRowsCapsAtOneHundredPerRoot is M4.2's per-Root cap: with more
// than recentPerRootCap documents under one Root, only the recentPerRootCap
// most recently viewed survive, oldest dropped first.
func TestRecentRowsCapsAtOneHundredPerRoot(t *testing.T) {
	dir := t.TempDir()
	rootDir := mkdir(t, filepath.Join(dir, "root"))

	const total = recentPerRootCap + 10
	docs := make(map[string]docRecord, total)
	now := time.Now()
	var newestPaths []string
	for i := 0; i < total; i++ {
		p := recentFixtureFile(t, rootDir, fmt.Sprintf("doc-%03d.md", i))
		// i=0 is the oldest (viewed longest ago), i=total-1 the newest.
		docs[p] = docRecord{LastViewed: now.Add(time.Duration(i) * time.Minute)}
		if i >= total-recentPerRootCap {
			newestPaths = append(newestPaths, p)
		}
	}
	roots := []root{{path: rootDir}}

	rows := recentRows(docs, roots, rootDir, false, activeRootBrowseScope)
	if len(rows) != recentPerRootCap {
		t.Fatalf("recentRows returned %d rows, want the cap %d", len(rows), recentPerRootCap)
	}
	got := map[string]bool{}
	for _, r := range rows {
		got[r.path] = true
	}
	for _, p := range newestPaths {
		if !got[p] {
			t.Errorf("newest document %q missing from the capped list; want only the oldest dropped", p)
		}
	}
}

// TestRecencyAndStarredNameTiesBreakByPath is TRIAGE T26's own gap (D83):
// no test exercised either sortByRecency's or starredRows' own path
// tie-break for two records sharing an exact timestamp/name —
// sortByRecency's doc comment already claims the tie-break (mirroring
// evictDocStore's oldest-first analogue) and starredRows' own name-sort
// does too, but nothing confirmed the ordering when the primary key
// actually ties.
func TestRecencyAndStarredNameTiesBreakByPath(t *testing.T) {
	t.Run("sortByRecency", func(t *testing.T) {
		ts := time.Now()
		entries := []docTimestamp{
			{path: "/docs/zebra.md", ts: ts},
			{path: "/docs/apple.md", ts: ts},
			{path: "/docs/mango.md", ts: ts},
		}
		sortByRecency(entries)
		want := []string{"/docs/apple.md", "/docs/mango.md", "/docs/zebra.md"}
		for i, w := range want {
			if entries[i].path != w {
				t.Errorf("entry %d = %q, want %q (tied timestamps break by path)", i, entries[i].path, w)
			}
		}
	})

	t.Run("starredRows", func(t *testing.T) {
		dir := t.TempDir()
		dirA := mkdir(t, filepath.Join(dir, "a"))
		dirB := mkdir(t, filepath.Join(dir, "b"))
		// Same base name in two different Roots — starredRows' own name sort
		// (filepath.Base) ties, so the tie-break must fall through to the
		// full path, which sorts dirA's copy first ("a" < "b"). That ordering
		// is guaranteed by the literal subdirectory names chosen below, not
		// by t.TempDir()'s own naming: dirA and dirB share the same parent
		// (dir), so the comparison reduces to "a" vs "b" regardless of what
		// t.TempDir() names that shared parent.
		docA := starredFixtureFile(t, dirA, "same.md")
		docB := starredFixtureFile(t, dirB, "same.md")
		docs := map[string]docRecord{
			docA: {Starred: true},
			docB: {Starred: true},
		}
		roots := []root{{path: dirA}, {path: dirB}}

		rows := starredRows(docs, roots, dirA, false, allStarredBrowseScope)
		want := []string{docA, docB}
		if len(rows) != len(want) {
			t.Fatalf("starredRows returned %d rows, want %d", len(rows), len(want))
		}
		for i, w := range want {
			if rows[i].path != w {
				t.Errorf("row %d = %q, want %q (tied names break by path)", i, rows[i].path, w)
			}
		}
	})
}

// TestRecentRowsCapsOrphanBucketAtOneHundred is TRIAGE T29's own gap
// (D83): recentRows buckets an orphan document (no session Root covers it)
// under the "" key exactly like any Root's own bucket, and the per-bucket
// cap loop (recentRows' own comment: "Each bucket is ordered newest-first
// and capped at recentPerRootCap ... independently") applies to every key
// including "" — but nothing confirmed that past the ordinary Root-keyed
// case TestRecentRowsCapsAtOneHundredPerRoot already covers.
func TestRecentRowsCapsOrphanBucketAtOneHundred(t *testing.T) {
	dir := t.TempDir()

	const total = recentPerRootCap + 10
	docs := make(map[string]docRecord, total)
	now := time.Now()
	var newestPaths []string
	for i := 0; i < total; i++ {
		p := recentFixtureFile(t, dir, fmt.Sprintf("orphan-%03d.md", i))
		docs[p] = docRecord{LastViewed: now.Add(time.Duration(i) * time.Minute)}
		if i >= total-recentPerRootCap {
			newestPaths = append(newestPaths, p)
		}
	}
	// No roots at all: every document above lands in the "" (orphan) bucket.
	rows := recentRows(docs, nil, "", false, allRootsBrowseScope)
	if len(rows) != recentPerRootCap {
		t.Fatalf("recentRows returned %d orphan rows, want the cap %d", len(rows), recentPerRootCap)
	}
	got := map[string]bool{}
	for _, r := range rows {
		if !r.orphan {
			t.Errorf("row %+v is not marked orphan", r)
		}
		got[r.path] = true
	}
	for _, p := range newestPaths {
		if !got[p] {
			t.Errorf("newest orphan document %q missing from the capped list; want only the oldest dropped", p)
		}
	}
}

// TestRecentRowsAllRootsMergesByTimestampAcrossRootsAndOrphans is M4.2's
// merged-ordering clause: All Roots interleaves two Roots' own per-Root
// lists, plus an orphan document under no session Root, into one
// timestamp-ordered list — also the case that exercises Recent's two
// right-hand label forms (S6), since Recent's scope set (S5) has no
// dedicated Orphans value the way Starred's does.
func TestRecentRowsAllRootsMergesByTimestampAcrossRootsAndOrphans(t *testing.T) {
	dir := t.TempDir()
	rootA := mkdir(t, filepath.Join(dir, "root-a"))
	rootB := mkdir(t, filepath.Join(dir, "root-b"))
	orphanDir := mkdir(t, filepath.Join(dir, "elsewhere"))

	aDoc := recentFixtureFile(t, rootA, "a.md")
	bDoc := recentFixtureFile(t, rootB, "b.md")
	orphanDoc := recentFixtureFile(t, orphanDir, "orphan.md")

	now := time.Now()
	docs := map[string]docRecord{
		aDoc:      {LastViewed: now.Add(-2 * time.Hour)}, // oldest
		orphanDoc: {LastViewed: now.Add(-1 * time.Hour)}, // middle
		bDoc:      {LastViewed: now},                     // newest
	}
	roots := []root{{path: rootA}, {path: rootB}}

	rows := recentRows(docs, roots, rootA, false, allRootsBrowseScope)
	wantOrder := []string{bDoc, orphanDoc, aDoc}
	if len(rows) != len(wantOrder) {
		t.Fatalf("recentRows returned %d rows, want %d", len(rows), len(wantOrder))
	}
	for i, w := range wantOrder {
		if rows[i].path != w {
			t.Errorf("row %d = %q, want %q", i, rows[i].path, w)
		}
	}

	// Right-hand label forms: a Root-owned row carries the Root's own base
	// name, never the orphan/left-truncated form, and vice versa.
	for _, r := range rows {
		switch r.path {
		case aDoc:
			if r.orphan || r.rootLabel != "root-a" {
				t.Errorf("a.md row = %+v, want rootLabel \"root-a\", orphan false", r)
			}
		case bDoc:
			if r.orphan || r.rootLabel != "root-b" {
				t.Errorf("b.md row = %+v, want rootLabel \"root-b\", orphan false", r)
			}
		case orphanDoc:
			if !r.orphan || r.orphanDir != orphanDir {
				t.Errorf("orphan.md row = %+v, want orphan true, orphanDir %q", r, orphanDir)
			}
		}
	}
}

// TestRecentRowsActiveRootScopeOnlyShowsThatRoot checks Active Source scope in
// isolation, against the same two-Root fixture shape as the All Roots test:
// only rootA's own document appears, never rootB's or the orphan's.
func TestRecentRowsActiveRootScopeOnlyShowsThatRoot(t *testing.T) {
	dir := t.TempDir()
	rootA := mkdir(t, filepath.Join(dir, "root-a"))
	rootB := mkdir(t, filepath.Join(dir, "root-b"))
	aDoc := recentFixtureFile(t, rootA, "a.md")
	bDoc := recentFixtureFile(t, rootB, "b.md")

	docs := map[string]docRecord{
		aDoc: {LastViewed: time.Now().Add(-time.Hour)},
		bDoc: {LastViewed: time.Now()},
	}
	roots := []root{{path: rootA}, {path: rootB}}

	rows := recentRows(docs, roots, rootA, false, activeRootBrowseScope)
	if len(rows) != 1 || rows[0].path != aDoc {
		t.Errorf("Active Source scope rows = %+v, want exactly [%q]", rows, aDoc)
	}
}

// TestRecentRowsOmitsDeletedFile is M4.2's own acceptance wording, last
// clause: a record whose file no longer exists on disk is silently dropped,
// never left in the Recent list as e.g. a broken entry (S6, unlike a broken
// star — M4.3's own concern).
func TestRecentRowsOmitsDeletedFile(t *testing.T) {
	dir := t.TempDir()
	rootDir := mkdir(t, filepath.Join(dir, "root"))
	present := recentFixtureFile(t, rootDir, "present.md")
	gone := filepath.Join(rootDir, "gone.md") // never created on disk

	docs := map[string]docRecord{
		present: {LastViewed: time.Now()},
		gone:    {LastViewed: time.Now().Add(time.Minute)}, // newer, but must still be dropped
	}
	roots := []root{{path: rootDir}}

	rows := recentRows(docs, roots, rootDir, false, activeRootBrowseScope)
	if len(rows) != 1 || rows[0].path != present {
		t.Errorf("recentRows = %+v, want exactly the present file, gone.md silently omitted", rows)
	}
}

// TestRecentRowsIgnoresNeverViewedRecords checks S4's "a zero LastViewed
// never appears" directly: a starred-but-never-viewed record (LastViewed
// zero) is absent from every scope, matching S1's independent Starred/
// LastViewed fields.
func TestRecentRowsIgnoresNeverViewedRecords(t *testing.T) {
	dir := t.TempDir()
	rootDir := mkdir(t, filepath.Join(dir, "root"))
	neverViewed := recentFixtureFile(t, rootDir, "never.md")

	docs := map[string]docRecord{neverViewed: {Starred: true}}
	roots := []root{{path: rootDir}}

	if rows := recentRows(docs, roots, rootDir, false, activeRootBrowseScope); len(rows) != 0 {
		t.Errorf("Active Source scope rows = %+v, want none (zero LastViewed)", rows)
	}
	if rows := recentRows(docs, roots, rootDir, false, allRootsBrowseScope); len(rows) != 0 {
		t.Errorf("All Roots scope rows = %+v, want none (zero LastViewed)", rows)
	}
}

// TestRecentRowsActiveRootScopeWithNoActiveRootIsEmpty covers the "" active-
// root-path edge case explicitly: with no root active, Active Source scope has
// no bucket to read, not the orphan bucket by accident (both key on "").
func TestRecentRowsActiveRootScopeWithNoActiveRootIsEmpty(t *testing.T) {
	dir := t.TempDir()
	orphanDoc := recentFixtureFile(t, dir, "orphan.md")
	docs := map[string]docRecord{orphanDoc: {LastViewed: time.Now()}}

	rows := recentRows(docs, nil, "", false, activeRootBrowseScope)
	if len(rows) != 0 {
		t.Errorf("recentRows with no active root = %+v, want none", rows)
	}
}

// TestFilesNavRecentModeRendersFlatRows is M4.2's own model-level wiring
// check: switching the Files Nav to Recent mode (keys.FilesMode) actually
// shows the flat list — file names and each row's own right-hand label —
// rather than continuing to show the Tree's rows underneath the mode/scope
// row M4.1 already reserved.
func TestFilesNavRecentModeRendersFlatRows(t *testing.T) {
	dir := t.TempDir()
	rootDir := mkdir(t, filepath.Join(dir, "project"))
	viewed := recentFixtureFile(t, rootDir, "viewed.md")
	recentFixtureFile(t, rootDir, "unviewed.md")

	node, err := tree.Walk(rootDir, tree.FilterMarkdown)
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	docs := map[string]docRecord{viewed: {LastViewed: time.Now()}}
	m := sized(t, newModel().withDocs("", docs), 80, 24)
	m.roots = m.roots.setRoots([]root{{path: rootDir}}, 0)
	m = m.openFolder(node)

	m = press(t, m, keys.FilesMode.Keys()[0]) // Tree -> Recent
	if m.files.mode != recentBrowse {
		t.Fatalf("setup: mode = %v, want recentBrowse", m.files.mode)
	}

	view := m.files.view()
	if !strings.Contains(view, "viewed.md") {
		t.Errorf("Recent view does not show the viewed document:\n%s", view)
	}
	if strings.Contains(view, "unviewed.md") {
		t.Errorf("Recent view shows a never-viewed document:\n%s", view)
	}
	if !strings.Contains(view, "project") {
		t.Errorf("Recent view does not show the Root's base name as the row's label:\n%s", view)
	}
}

// TestFilesNavRecentModeEmptyShowsAMessage checks Recent mode's own resting
// state: with nothing recently viewed, the pane shows Active Source scope's
// own M4.4 message rather than a blank list.
func TestFilesNavRecentModeEmptyShowsAMessage(t *testing.T) {
	dir := t.TempDir()
	rootDir := mkdir(t, filepath.Join(dir, "project"))
	recentFixtureFile(t, rootDir, "unviewed.md")

	node, err := tree.Walk(rootDir, tree.FilterMarkdown)
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	m := sized(t, newModel().withDocs("", map[string]docRecord{}), 80, 24)
	m.roots = m.roots.setRoots([]root{{path: rootDir}}, 0)
	m = m.openFolder(node)

	m = press(t, m, keys.FilesMode.Keys()[0]) // Tree -> Recent

	view := normalizedRows(m.files.view())
	want := flatEmptyText(recentBrowse, activeRootBrowseScope)
	if !strings.Contains(view, want) {
		t.Errorf("empty Recent view = %q, want it to contain %q", view, want)
	}
}

// TestFilesNavRecentModeShowsStarMarker is TRIAGE T31's own gap (D83):
// renderFlatRow's own doc comment says Recent rows "get the same star
// marker prefix Tree rows carry (Recent is not Starred, so membership alone
// does not already say it)", but no render test asserted the marker itself
// on a starred Recent row, unlike Starred mode's own
// TestFilesNavStarredModeRendersFlatRowsWithoutStarMarker (which asserts the
// marker's absence).
func TestFilesNavRecentModeShowsStarMarker(t *testing.T) {
	dir := t.TempDir()
	rootDir := mkdir(t, filepath.Join(dir, "project"))
	starred := recentFixtureFile(t, rootDir, "starred.md")
	unstarred := recentFixtureFile(t, rootDir, "unstarred.md")

	node, err := tree.Walk(rootDir, tree.FilterMarkdown)
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	now := time.Now()
	docs := map[string]docRecord{
		starred:   {LastViewed: now, Starred: true},
		unstarred: {LastViewed: now.Add(-time.Minute)},
	}
	m := sized(t, newModel().withDocs("", docs), 80, 24)
	m.roots = m.roots.setRoots([]root{{path: rootDir}}, 0)
	m = m.openFolder(node)

	m = press(t, m, keys.FilesMode.Keys()[0]) // Tree -> Recent

	lines := strings.Split(m.files.view(), "\n")
	starredLine := lineContaining(lines, "starred.md")
	unstarredLine := lineContaining(lines, "unstarred.md")
	if !strings.Contains(starredLine, "\u2605") {
		t.Errorf("starred.md row = %q, want the star marker", starredLine)
	}
	if strings.Contains(unstarredLine, "\u2605") {
		t.Errorf("unstarred.md row = %q, want no star marker", unstarredLine)
	}
}

// TestRecentRowsNeverOmitsAnArticleAsMissing is C6/M6.1's own fix pinned
// directly: an article's path is a URL, which pathMissing (os.Stat) always
// reports as missing, so recentRows must skip that check for an article
// record entirely rather than silently omitting every article the way it
// would a genuinely deleted file (TestRecentRowsOmitsDeletedFile's own
// clause) — S7's "same omission of entries that can no longer be
// resolved" has nothing for a fetchable-on-demand article to ever fail.
func TestRecentRowsNeverOmitsAnArticleAsMissing(t *testing.T) {
	article := fixtureArticle()
	docs := map[string]docRecord{
		article.path: {LastViewed: time.Now(), Article: true, Title: article.title},
	}

	rows := recentRows(docs, nil, "", true, activeRootBrowseScope)
	if len(rows) != 1 || rows[0].path != article.path {
		t.Fatalf("recentRows = %+v, want exactly the article, never omitted as missing", rows)
	}
}

// TestRecentRowsWikipediaActiveSourceShowsArticlesByRecency is C6/M6.1's
// own S5 clause for Recent (mirroring starredRows'
// TestStarredRowsWikipediaActiveSourceShowsArticlesByTitle): with Wikipedia
// the Active Source (wikipedia == true), Active Source scope reads the
// article bucket, not activeRootPath's (always "" while Wikipedia is
// active) — ordered newest-first like any other Recent list, rows named by
// title and labelled Wikipedia, never leaking a Root-orphan file that
// happens to share the "" key.
func TestRecentRowsWikipediaActiveSourceShowsArticlesByRecency(t *testing.T) {
	dir := t.TempDir()
	orphanDoc := recentFixtureFile(t, dir, "orphan.md")
	older := fixtureArticle()
	newer := document{path: "https://en.wikipedia.org/wiki/Carl_Friedrich_Gauss", title: "Carl Friedrich Gauss"}

	now := time.Now()
	docs := map[string]docRecord{
		orphanDoc:  {LastViewed: now},
		older.path: {LastViewed: now.Add(-time.Hour), Article: true, Title: older.title},
		newer.path: {LastViewed: now, Article: true, Title: newer.title},
	}

	rows := recentRows(docs, nil, "", true, activeRootBrowseScope)
	want := []string{newer.path, older.path}
	if len(rows) != len(want) {
		t.Fatalf("Wikipedia Active Source rows = %+v, want exactly the two articles (orphan file excluded)", rows)
	}
	for i, w := range want {
		if rows[i].path != w {
			t.Errorf("row %d path = %q, want %q (newest-first)", i, rows[i].path, w)
		}
		if !rows[i].article {
			t.Errorf("row %d article = false, want true", i)
		}
		if rows[i].rootLabel != wikipediaLabel {
			t.Errorf("row %d rootLabel = %q, want %q", i, rows[i].rootLabel, wikipediaLabel)
		}
	}
	if rows[0].name != newer.title || rows[1].name != older.title {
		t.Errorf("row names = [%q, %q], want [%q, %q]", rows[0].name, rows[1].name, newer.title, older.title)
	}
}

// TestRecentRowsAllSourcesIncludesArticlesLabelledWikipediaWithARootActive
// is S6's own clause applied to Recent (mirroring starredRows'
// TestStarredRowsAllStarredIncludesArticlesLabelledWikipediaWithARootActive):
// with a Root active (not Wikipedia), All Sources still merges in a
// recently-viewed article alongside the Root's own recent files, labelled
// Wikipedia where a file is labelled by its covering Root's name, ordered
// by recency across both.
func TestRecentRowsAllSourcesIncludesArticlesLabelledWikipediaWithARootActive(t *testing.T) {
	dir := t.TempDir()
	rootA := mkdir(t, filepath.Join(dir, "root-a"))
	aDoc := recentFixtureFile(t, rootA, "a.md")
	article := fixtureArticle()
	roots := []root{{path: rootA}}

	now := time.Now()
	docs := map[string]docRecord{
		aDoc:         {LastViewed: now.Add(-time.Hour)},
		article.path: {LastViewed: now, Article: true, Title: article.title},
	}

	rows := recentRows(docs, roots, rootA, false, allRootsBrowseScope)
	want := []string{article.path, aDoc}
	if len(rows) != len(want) {
		t.Fatalf("All Sources rows = %+v, want exactly %v", rows, want)
	}
	for i, w := range want {
		if rows[i].path != w {
			t.Errorf("row %d path = %q, want %q (newest-first)", i, rows[i].path, w)
		}
	}
	if !rows[0].article || rows[0].name != article.title || rows[0].rootLabel != wikipediaLabel {
		t.Errorf("article row = %+v, want name %q labelled %q", rows[0], article.title, wikipediaLabel)
	}
	if rows[1].article || rows[1].rootLabel != "root-a" {
		t.Errorf("a.md row = %+v, want rootLabel %q, article false", rows[1], "root-a")
	}
}

// TestRecentRowsCapsWikipediaArticleBucketAtOneHundred is S7's own "same
// per-Source cap as files" clause: an article bucket (keyed by
// wikipediaSourceKey) is capped at recentPerRootCap exactly like any Root's
// own bucket (TestRecentRowsCapsAtOneHundredPerRoot's own shape), oldest
// articles dropped first.
func TestRecentRowsCapsWikipediaArticleBucketAtOneHundred(t *testing.T) {
	const total = recentPerRootCap + 10
	docs := make(map[string]docRecord, total)
	now := time.Now()
	var newestPaths []string
	for i := 0; i < total; i++ {
		p := fmt.Sprintf("https://en.wikipedia.org/wiki/Article_%03d", i)
		docs[p] = docRecord{LastViewed: now.Add(time.Duration(i) * time.Minute), Article: true, Title: fmt.Sprintf("Article %03d", i)}
		if i >= total-recentPerRootCap {
			newestPaths = append(newestPaths, p)
		}
	}

	rows := recentRows(docs, nil, "", true, activeRootBrowseScope)
	if len(rows) != recentPerRootCap {
		t.Fatalf("recentRows returned %d article rows, want the cap %d", len(rows), recentPerRootCap)
	}
	got := map[string]bool{}
	for _, r := range rows {
		got[r.path] = true
	}
	for _, p := range newestPaths {
		if !got[p] {
			t.Errorf("newest article %q missing from the capped list; want only the oldest dropped", p)
		}
	}
}
