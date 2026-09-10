package main

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"

	"noggin/internal/tree"
	"noggin/internal/wikipedia"
)

// locationRootFixture builds a real temp Root containing one file (S9's
// "file under a Root" case) and opens it as DV1's Active Document, mirroring
// peekFixture's own real-Root setup (peek_test.go) rather than
// openedFixture's rootless one — S9's Root-relative form has nothing to
// compute without a covering Root.
func locationRootFixture(t *testing.T) (m model, rootPath, filePath string) {
	t.Helper()
	rootPath = mkdir(t, filepath.Join(t.TempDir(), "notes"))
	mkdir(t, filepath.Join(rootPath, "sub"))
	filePath = writeFileContent(t, filepath.Join(rootPath, "sub", "doc.md"), "# Doc\n")

	node, err := tree.Walk(rootPath)
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	// setRoots is what actually registers rootPath as a session Root
	// (rootForPath/documentLocation's own source, roots.go/location.go) --
	// openFolder alone only points the Files Nav's tree at it (S6), leaving
	// m.roots.rows empty, which would make every file under it read as an
	// orphan instead (S9's own "no session Root" case, the wrong one here).
	m = newModel()
	m.roots = m.roots.setRoots([]root{{path: rootPath}}, 0)
	m = sized(t, m.openFolder(node), 80, 24)
	d, err := loadDocument(filePath)
	if err != nil {
		t.Fatalf("loadDocument: %v", err)
	}
	m = m.openDocument(d)
	return m, rootPath, filePath
}

// TestLocationLineForFileUnderARoot is S9's own file case: the covering
// Root's own label (S8/M7.3's displayName, here the folder's default base
// name — no rename in play) paired with the file's path relative to that
// Root, exactly the "<Root name> · <relative path>" form PLAN M8.2 names.
func TestLocationLineForFileUnderARoot(t *testing.T) {
	m, rootPath, _ := locationRootFixture(t)
	want := filepath.Base(rootPath) + " · sub/doc.md"
	if got := m.viewerLocationLine(dv1Pane); got != want {
		t.Errorf("location line = %q, want %q", got, want)
	}
}

// TestLocationLineForFileUnderARenamedRoot pins S8's own promise that the
// location line reads a Root's persisted Name, not just its folder's base
// name (S9's own "used everywhere the Root's base name is shown today"
// naming this line explicitly), the same displayName() every other S8 site
// already reads from (M7.3).
func TestLocationLineForFileUnderARenamedRoot(t *testing.T) {
	m, _, _ := locationRootFixture(t)
	rows := m.roots.rows
	rows[0].root.name = "My Notes"
	m.roots.rows = rows
	want := "My Notes · sub/doc.md"
	if got := m.viewerLocationLine(dv1Pane); got != want {
		t.Errorf("location line = %q, want %q", got, want)
	}
}

// TestLocationLineForOrphanFile is S9's orphan case: with no session Root
// covering the open document (S1's "not attached to Roots"), the line is
// just the document's own full path, unpaired with any Root label —
// openedFixture's own document (testdata/fixture.md) is exactly this, since
// it opens with no Root at all.
func TestLocationLineForOrphanFile(t *testing.T) {
	m := openedFixture(t, 80, 24)
	if got, want := m.viewerLocationLine(dv1Pane), fixturePath; got != want {
		t.Errorf("location line = %q, want %q (the document's own full path)", got, want)
	}
}

// TestLocationLineForArticleIsTheExactThreeSegmentAttribution pins S9a's
// literal wording verbatim (D98/D112, PLAN M8.2's own "the article line's
// exact three-segment text"): the CC BY-SA 4.0 notice, the article's own
// canonical URL, and the modification disclosure, joined by " · " exactly
// like S9's Root/orphan forms above, replacing them entirely rather than
// combining with either.
func TestLocationLineForArticleIsTheExactThreeSegmentAttribution(t *testing.T) {
	m := sized(t, newModel().openDocument(fixtureArticle()), 80, 24)
	want := "CC BY-SA 4.0 · https://en.wikipedia.org/wiki/Leonhard_Euler · Reformatted to markdown"
	if got := m.viewerLocationLine(dv1Pane); got != want {
		t.Errorf("location line = %q, want %q", got, want)
	}
}

// TestLocationLineAbsentWithNoDocument is S9's own "absent when the viewer
// has no document": no line to compose, and — PLAN M8.2's own content
// height clause — nothing is reserved for one either, so a document-less
// viewer gets the content box's full, unreduced height back.
func TestLocationLineAbsentWithNoDocument(t *testing.T) {
	m := sized(t, newModel(), 80, 24)
	if got := m.viewerLocationLine(dv1Pane); got != "" {
		t.Errorf("location line = %q, want empty with no document open", got)
	}
	want := computeLayout(80, 24, false, false).viewer.content().height
	if got := m.dv1.height(); got != want {
		t.Errorf("viewport height = %d, want the full content height %d (no row reserved)", got, want)
	}
}

// TestLocationLineCostsOneRowOnlyWhenAShown is PLAN M8.2's content-height
// clause, isolated from any particular document state: opening a document
// costs the viewport exactly the location line's one row versus the same
// model before anything was ever opened, and closing it again (refresh.go's
// own "Active Document vanished" path, viewer.close()) gives that row back.
func TestLocationLineCostsOneRowOnlyWhenAShown(t *testing.T) {
	empty := sized(t, newModel(), 80, 24)
	emptyHeight := empty.dv1.height()

	opened := empty.openDocument(fixtureDoc(t))
	if got, want := opened.dv1.height(), emptyHeight-1; got != want {
		t.Errorf("viewport height with a document open = %d, want %d (one row less than empty's %d)", got, want, emptyHeight)
	}

	closed := opened.withFocusedViewer(opened.dv1.close())
	if got := closed.dv1.height(); got != emptyHeight {
		t.Errorf("viewport height after close() = %d, want the full %d back", got, emptyHeight)
	}
}

// TestLocationLineReflectsThePeekedDocumentNotTheActiveOne is S9's own
// "reflects a Peeked document while the Peek lasts": Peeking a second file
// shows *its* location, not the Active Document's (active.md) — peekFixture
// itself never registers a session Root (only openFolder, which points the
// Files Nav at one without touching m.roots.rows), so both read as orphans
// here (their own full paths); TestLocationLineForFileUnderARoot/
// ...ForARenamedRoot already cover the Root-labelled form on its own, so
// this test only needs the two paths to differ, which an orphan's own
// full-path form still does.
func TestLocationLineReflectsThePeekedDocumentNotTheActiveOne(t *testing.T) {
	m, _, _ := peekFixture(t)
	before := m.viewerLocationLine(dv1Pane)
	if !strings.HasSuffix(before, "active.md") {
		t.Fatalf("setup: location line before Peek = %q, want it to name active.md", before)
	}

	m = press(t, m, keys.Peek.Keys()[0])
	if !m.dv1.peeking() {
		t.Fatal("setup: viewer not Peeking after p")
	}

	got := m.viewerLocationLine(dv1Pane)
	if !strings.HasSuffix(got, "peeked.md") {
		t.Errorf("location line while Peeking = %q, want it to name peeked.md", got)
	}
	if got == before {
		t.Errorf("location line while Peeking = %q, unchanged from before Peek (%q) — reflects the Active Document instead", got, before)
	}
}

// TestLocationLineForDV2IsIndependentOfDV1 confirms composeViewerContent/
// viewerLocationLine are genuinely per-pane, not hardcoded to dv1Pane:
// opening a second file into DV2 (S1/C3's own targeting modifier) gives it
// its own location line naming its own file, independent of DV1's.
func TestLocationLineForDV2IsIndependentOfDV1(t *testing.T) {
	m, rootPath, _ := locationRootFixture(t)
	second := writeFileContent(t, filepath.Join(rootPath, "second.md"), "# Second\n")
	d, err := loadDocument(second)
	if err != nil {
		t.Fatalf("loadDocument: %v", err)
	}
	m = m.openDocumentInto(dv2Pane, d)

	wantDV1 := filepath.Base(rootPath) + " · sub/doc.md"
	wantDV2 := filepath.Base(rootPath) + " · second.md"
	if got := m.viewerLocationLine(dv1Pane); got != wantDV1 {
		t.Errorf("dv1Pane location line = %q, want %q", got, wantDV1)
	}
	if got := m.viewerLocationLine(dv2Pane); got != wantDV2 {
		t.Errorf("dv2Pane location line = %q, want %q", got, wantDV2)
	}
}

// TestLocationLineHeightDuringAndAfterAnArticleRefresh is a code-review
// finding pinned as a regression test: refreshing (or re-Activating into) a
// pane that already has an article open, and so already has its location
// line reserving a row (docContentHeight), starts a fetch that draws no
// line at all while in flight (locationDoc's own "case v.fetching(): false"
// — view()'s own placeholder takes over the box, same as the empty/error
// cases). Without beginFetch/endFetch (viewer.go) restoring the full
// boxHeight for that placeholder and the correct height again once the
// fetch resolves — success *and* failure both, since a failure never
// touches the document at all — the box would either show a blank
// unstyled row during the fetch (one short of the placeholder's own full
// height) or overflow by one row afterward (a failure leaves the document,
// and so its location line, exactly as it was, but pendingFetch's own
// height reset would otherwise still be in effect).
func TestLocationLineHeightDuringAndAfterAnArticleRefresh(t *testing.T) {
	boxHeight := computeLayout(80, 24, false, false).viewer.content().height

	t.Run("successful refresh", func(t *testing.T) {
		updated := wikipedia.Article{Title: eulerArticle.Title, URL: eulerArticle.URL, Markdown: "# Leonhard Euler\n\nRe-fetched.\n"}
		fetcher := &fakeWikipediaFetcher{articles: map[string]wikipedia.Article{euler.Title: updated}}
		m := sized(t, newModel(), 80, 24).withWikipediaFetcher(fetcher)
		m = m.navigate(dv1Pane, fixtureArticle())
		if got := m.dv1.height(); got != boxHeight-1 {
			t.Fatalf("setup: height with the article's location line showing = %d, want %d", got, boxHeight-1)
		}

		after, cmd := m.refresh()
		if !after.dv1.fetching() {
			t.Fatal("setup: dv1 not fetching immediately after refresh")
		}
		if got := after.dv1.height(); got != boxHeight {
			t.Errorf("height while fetching = %d, want the full %d back (no line drawn meanwhile)", got, boxHeight)
		}

		got := runArticleFetch(t, after, cmd)
		if got.dv1.fetching() {
			t.Fatal("setup: dv1 still fetching after the Cmd ran")
		}
		if h := got.dv1.height(); h != boxHeight-1 {
			t.Errorf("height after a successful refresh = %d, want %d (the row reserved again)", h, boxHeight-1)
		}
	})

	t.Run("failed refresh", func(t *testing.T) {
		fetcher := &fakeWikipediaFetcher{errs: map[string]error{euler.Title: errors.New("offline")}}
		m := sized(t, newModel(), 80, 24).withWikipediaFetcher(fetcher)
		m = m.navigate(dv1Pane, fixtureArticle())

		after, cmd := m.refresh()
		if got := after.dv1.height(); got != boxHeight {
			t.Fatalf("height while fetching = %d, want the full %d", got, boxHeight)
		}

		got := runArticleFetch(t, after, cmd)
		if got.dv1.fetching() {
			t.Fatal("setup: dv1 still fetching after the Cmd ran")
		}
		if got.message == "" {
			t.Fatal("setup: no Footer message after a failed refresh")
		}
		// The article itself never changed, so its location line reads exactly
		// as before — the height must not have been left at the fetching
		// placeholder's full boxHeight (would overflow the box by one row once
		// composeViewerContent draws the line back in).
		if h := got.dv1.height(); h != boxHeight-1 {
			t.Errorf("height after a failed refresh = %d, want %d (the untouched article's row reserved again)", h, boxHeight-1)
		}
		if line := got.viewerLocationLine(dv1Pane); line == "" {
			t.Error("location line empty after a failed refresh, want the untouched article's own line")
		}
	})
}

// TestFetchingPlaceholderUsesPeeksOwnDimensionsWhenPeekTargeted is a
// second code-review finding pinned as a regression test: view()'s
// v.fetching() branch must fill *Peek's* box (v.peek.vp) when the pending
// fetch targets Peek, not the Active Document's (v.vp) — the two can now
// differ (boxHeight's own doc comment) whenever the Active Document already
// has its own location line showing (v.vp one row short of the full box)
// while a Peek-targeted fetch (no Peek open yet, so nothing has sized
// v.peek.vp before now) starts, wanting the full box for itself.
func TestFetchingPlaceholderUsesPeeksOwnDimensionsWhenPeekTargeted(t *testing.T) {
	fetcher := &fakeWikipediaFetcher{articles: map[string]wikipedia.Article{euler.Title: eulerArticle}}
	m := sized(t, newModel(), 80, 24).withWikipediaFetcher(fetcher)
	// An Active Document already open means v.vp reserves the location
	// line's row (boxHeight-1) — the state the Peek-targeted fetch below
	// must not be confused with.
	m = m.navigate(dv1Pane, fixtureDoc(t))
	boxHeight := computeLayout(80, 24, false, false).viewer.content().height
	if got := m.dv1.height(); got != boxHeight-1 {
		t.Fatalf("setup: Active Document viewport height = %d, want %d", got, boxHeight-1)
	}

	m, _ = m.fetchArticleInto(dv1Pane, euler.Title, true)
	if !m.dv1.fetching() {
		t.Fatal("setup: dv1 not fetching after fetchArticleInto(peek=true)")
	}

	view := m.dv1.view()
	if got, want := strings.Count(view, "\n")+1, boxHeight; got != want {
		t.Errorf("fetching placeholder height = %d lines, want the full box height %d (Peek's own, not the Active Document's %d)", got, want, boxHeight-1)
	}
	wantWidth := computeLayout(80, 24, false, false).viewer.content().width
	for i, line := range strings.Split(view, "\n") {
		if got := lipgloss.Width(line); got != wantWidth {
			t.Errorf("line %d width = %d, want %d", i, got, wantWidth)
		}
	}
}

// TestLocationLineComposedBeneathTheViewersOwnContent is an end-to-end
// check through model.render() (composeViewerContent's real caller,
// model.go's frame()): the line reaches the actual frame, as the DV1 box's
// own last content row — not merely a value some helper can compute, but
// what a user watching the real screen would see.
func TestLocationLineComposedBeneathTheViewersOwnContent(t *testing.T) {
	m, rootPath, _ := locationRootFixture(t)
	frame := m.render()
	want := filepath.Base(rootPath) + " · sub/doc.md"
	if !strings.Contains(frame, want) {
		t.Errorf("rendered frame does not contain the location line %q:\n%s", want, frame)
	}
}

// TestLocationLineNeverClaimsARowAtAPathologicallyShortTerminal is TRIAGE
// T26's regression test (PLAN M8.2a): docContentHeight already floored the
// viewport itself at 0 for a pathologically short terminal (boxHeight 0,
// e.g. 80x4 once the border/title rows are subtracted — computeLayout's own
// content() math), but composeViewerContent still appended the location
// line unconditionally whenever a document was open, growing the rendered
// pane by one row it never budgeted for and corrupting the frame
// (renderBox's own MaxHeight backstop then clips the pane's *bottom border*
// to make room, rather than the line that overflowed it — DV1 never
// closes). docLocationVisible/peekLocationVisible now also require
// boxHeight >= 1, following computeLayout's own staleRow precedent (never
// claim a row the frame didn't budget for) — this pins both the unit-level
// gate and the end-to-end symptom it fixes.
func TestLocationLineNeverClaimsARowAtAPathologicallyShortTerminal(t *testing.T) {
	m := sized(t, newModel(), 80, 4)
	m = m.openDocument(fixtureDoc(t))

	if got := m.dv1.boxHeight; got != 0 {
		t.Fatalf("setup: dv1.boxHeight = %d, want 0 (the pathological case this test targets)", got)
	}
	if m.dv1.docLocationVisible() {
		t.Error("docLocationVisible = true at boxHeight 0, want false — nothing to reserve the row from")
	}
	if got := m.dv1.docContentHeight(); got != 0 {
		t.Errorf("docContentHeight = %d, want 0 (no row subtracted for a line that cannot show)", got)
	}
	if got := m.viewerLocationLine(dv1Pane); got != "" {
		t.Errorf("viewerLocationLine = %q, want empty — nothing to draw with no row to spare", got)
	}

	// End-to-end: with the Navigator hidden (isolating DV1's own pane from
	// the Sources/Files Nav column's unrelated squeeze at this size), DV1's
	// box must still close with its own bottom border and the frame must
	// stay at exactly the requested height — the observable form of the bug
	// (a swallowed border, not a taller frame, since renderBox's MaxHeight
	// backstop clips instead of growing).
	m = press(t, m, keys.Navigator.Keys()[0])
	view := m.View().Content
	if got := lipgloss.Height(view); got != 4 {
		t.Fatalf("frame height = %d, want 4", got)
	}
	lines := strings.Split(view, "\n")
	if len(lines) != 4 {
		t.Fatalf("frame has %d lines, want 4", len(lines))
	}
	if !strings.Contains(lines[2], "╰") {
		t.Errorf("DV1's own bottom border row is missing (line 2 = %q) — the pane never closes", lines[2])
	}
}

// TestPeekLocationLineNeverClaimsARowAtAPathologicallyShortTerminal is
// M8.2b's own test (TRIAGE T36): the sibling regression test above
// (M8.2a) only ever exercises the Active-Document path
// (docLocationVisible/docContentHeight); the Peek path
// (peekLocationVisible/peekContentHeight) is gated by the very same
// v.boxHeight >= 1 check and the very same shared contentHeight helper,
// but M8.2a's own PLAN task named it explicitly and had no test of its
// own covering it. Mirrors the sibling test's shape exactly, substituting
// a Peek (peekFixture's own real-Root setup, since Peek needs a second
// file to select) for the Active Document open.
func TestPeekLocationLineNeverClaimsARowAtAPathologicallyShortTerminal(t *testing.T) {
	m, _, _ := peekFixture(t)
	m = sized(t, m, 80, 4)

	if got := m.dv1.boxHeight; got != 0 {
		t.Fatalf("setup: dv1.boxHeight = %d, want 0 (the pathological case this test targets)", got)
	}

	m = press(t, m, keys.Peek.Keys()[0])
	if !m.dv1.peeking() {
		t.Fatal("setup: viewer not Peeking after p")
	}

	if m.dv1.peekLocationVisible() {
		t.Error("peekLocationVisible = true at boxHeight 0, want false — nothing to reserve the row from")
	}
	if got := m.dv1.peekContentHeight(); got != 0 {
		t.Errorf("peekContentHeight = %d, want 0 (no row subtracted for a line that cannot show)", got)
	}
	if got := m.viewerLocationLine(dv1Pane); got != "" {
		t.Errorf("viewerLocationLine = %q, want empty — nothing to draw with no row to spare", got)
	}

	// End-to-end: with the Navigator hidden (isolating DV1's own pane from
	// the Sources/Files Nav column's unrelated squeeze at this size), DV1's
	// box must still close with its own bottom border and the frame must
	// stay at exactly the requested height, matching the Active-Document
	// sibling test's own end-to-end assertion.
	m = press(t, m, keys.Navigator.Keys()[0])
	view := m.View().Content
	if got := lipgloss.Height(view); got != 4 {
		t.Fatalf("frame height = %d, want 4", got)
	}
	lines := strings.Split(view, "\n")
	if len(lines) != 4 {
		t.Fatalf("frame has %d lines, want 4", len(lines))
	}
	if !strings.Contains(lines[2], "╰") {
		t.Errorf("DV1's own bottom border row is missing (line 2 = %q) — the pane never closes", lines[2])
	}
}
