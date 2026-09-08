package main

import (
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"

	"noggin/internal/wikipedia"
)

// fakeWikipediaFetcher is C6/M4.2's test double for wikipediaFetcher: a
// fixed, title-keyed table of articles or errors, so a model test can drive
// Activate/Peek's async fetch without ever reaching the network — the
// Fetch-side twin of fakeWikipediaSearcher (picker_wikipedia_search_test.go).
// calls records every title actually fetched, so a stale-completion test can
// assert what did or did not run, not merely what result landed.
type fakeWikipediaFetcher struct {
	articles map[string]wikipedia.Article
	errs     map[string]error
	calls    []string
}

func (f *fakeWikipediaFetcher) Fetch(title string, refresh bool) (wikipedia.Article, error) {
	f.calls = append(f.calls, title)
	if err, ok := f.errs[title]; ok {
		return wikipedia.Article{}, err
	}
	return f.articles[title], nil
}

// wikipediaModeWithResult opens the Picker already landed on wikipediaMode
// (pickerWithWikipediaActive, picker_wikipedia_mode_test.go) with a single,
// already-current Result selected and ready to Activate — every Activate
// test below starts from here, skipping the search/debounce machinery
// M3.3's own tests already cover.
func wikipediaModeWithResult(t *testing.T, fetcher *fakeWikipediaFetcher, result wikipedia.SearchResult) model {
	t.Helper()
	m := pickerWithWikipediaActive(t)
	m = m.withWikipediaFetcher(fetcher)
	m.picker.input.SetValue(result.Title)
	m.picker.results = pickerResults{
		query:     result.Title,
		mode:      wikipediaMode,
		wikipedia: []wikipedia.SearchResult{result},
	}
	m.picker.selected = 0
	return m
}

// runArticleFetch executes cmd (fetchArticleInto's own return) and delivers
// its resulting articleFetchMsg back through Update, the same
// dispatch-then-deliver shape deliverPickerTicks already gives the Picker's
// own async search (picker_wikipedia_search_test.go) — there is exactly one
// message a fetch's Cmd ever produces, so no generation loop is needed here.
func runArticleFetch(t *testing.T, m model, cmd tea.Cmd) model {
	t.Helper()
	if cmd == nil {
		t.Fatal("fetchArticleInto returned a nil Cmd")
	}
	msg := cmd()
	next, _ := m.Update(msg)
	return concrete(t, next)
}

var euler = wikipedia.SearchResult{Title: "Leonhard Euler", Description: "Swiss mathematician"}

var eulerArticle = wikipedia.Article{
	Title:    "Leonhard Euler",
	URL:      "https://en.wikipedia.org/wiki/Leonhard_Euler",
	Markdown: "# Leonhard Euler\n\nA mathematician.\n",
}

// TestActivateWikipediaResultDispatchesFetchWithIndicator is S3/C6/M4.2's
// "the UI stays responsive and shows a visible in-progress indicator" half,
// checked before the fetch ever completes: Activate on a Wikipedia Result
// dismisses the Picker and moves Focus to DV1 immediately (activate's own
// contract, unaffected by the fetch being deferred), but the Active
// Document itself is untouched until the fetch resolves, and DV1 reports
// itself as fetching in both the pane title and its content box.
func TestActivateWikipediaResultDispatchesFetchWithIndicator(t *testing.T) {
	fetcher := &fakeWikipediaFetcher{articles: map[string]wikipedia.Article{euler.Title: eulerArticle}}
	m := wikipediaModeWithResult(t, fetcher, euler)
	beforeDoc := m.dv1.doc.path

	got := press(t, m, keys.Activate.Keys()[0])

	if got.picker.open {
		t.Error("picker.open = true after Activate, want dismissed")
	}
	if got.focus != dv1Pane {
		t.Errorf("focus = %v after Activate, want dv1Pane", got.focus)
	}
	if got.dv1.doc.path != beforeDoc {
		t.Errorf("DV1 Active Document = %q, want unchanged %q until the fetch completes", got.dv1.doc.path, beforeDoc)
	}
	if !got.dv1.fetching() {
		t.Fatal("dv1.fetching() = false immediately after Activate, want a fetch in flight")
	}
	if got.dv1.pendingFetch.title != euler.Title {
		t.Errorf("pendingFetch.title = %q, want %q", got.dv1.pendingFetch.title, euler.Title)
	}
	if got.message != "" {
		t.Errorf("message = %q while a fetch is in flight, want none yet", got.message)
	}
	if got.paneTitle(dv1Pane) != "Fetching: "+euler.Title {
		t.Errorf("paneTitle(dv1Pane) = %q, want the fetching indicator", got.paneTitle(dv1Pane))
	}
	if got.dv1.view() == "" {
		t.Error("dv1.view() is empty while fetching, want a visible indicator")
	}
	if len(fetcher.calls) != 0 {
		t.Fatalf("Fetch calls = %v before the Cmd has run, want none yet — dispatch and execution are separate steps", fetcher.calls)
	}
}

// TestActivateWikipediaResultOpensCanonicalArticleInDV1 is S3's full round
// trip for the unmodified key: once the fetch completes, DV1 renders the
// canonical article (D129/M4.1's articleDocument), the indicator is gone,
// and Activate's usual side effects (S4 recency, S9 history) apply exactly
// as they do for a file (TestActivateRecordsLastViewed's own shape).
func TestActivateWikipediaResultOpensCanonicalArticleInDV1(t *testing.T) {
	fetcher := &fakeWikipediaFetcher{articles: map[string]wikipedia.Article{euler.Title: eulerArticle}}
	m := wikipediaModeWithResult(t, fetcher, euler)
	m = m.withDocs(t.TempDir()+"/documents", map[string]docRecord{})
	before := m

	after, cmd := before.activatePickerInto(dv1Pane)
	got := runArticleFetch(t, after, cmd)

	if got.dv1.doc.path != eulerArticle.URL {
		t.Errorf("DV1 Active Document = %q, want the canonical URL %q", got.dv1.doc.path, eulerArticle.URL)
	}
	if got.dv1.doc.title != eulerArticle.Title {
		t.Errorf("DV1 Active Document title = %q, want %q", got.dv1.doc.title, eulerArticle.Title)
	}
	if got.dv1.fetching() {
		t.Error("dv1.fetching() = true after completion, want the indicator cleared")
	}
	if rec := got.docRecord(eulerArticle.URL); rec.LastViewed.IsZero() {
		t.Error("LastViewed is zero after Activate completed, want recency recorded (S4)")
	}
	if !got.dv1.canGoBack() && got.dv1.historyPos != 0 {
		t.Errorf("history position = %d after the first navigation, want 0", got.dv1.historyPos)
	}
	if len(fetcher.calls) != 1 || fetcher.calls[0] != euler.Title {
		t.Errorf("Fetch calls = %v, want exactly one call for %q", fetcher.calls, euler.Title)
	}
}

// TestActivateDV2WikipediaResultOpensCanonicalArticleInDV2 is S3's DV2
// half: the Picker's own `ctrl+` Targeting modifier (keys.ActivateDV2)
// routes the same fetch-and-open through openDocumentInto/navigate's DV2
// path, creating DV2 first, exactly as it already does for a file
// (activatePickerInto's own fileMode branch).
func TestActivateDV2WikipediaResultOpensCanonicalArticleInDV2(t *testing.T) {
	fetcher := &fakeWikipediaFetcher{articles: map[string]wikipedia.Article{euler.Title: eulerArticle}}
	m := wikipediaModeWithResult(t, fetcher, euler)
	if m.dv2 != nil {
		t.Fatal("setup: DV2 exists before the targeting modifier was pressed")
	}

	after, cmd := m.activatePickerInto(dv2Pane)
	got := runArticleFetch(t, after, cmd)

	if got.dv2 == nil {
		t.Fatal("ActivateDV2 did not create DV2")
	}
	if got.dv2.doc.path != eulerArticle.URL {
		t.Errorf("DV2 Active Document = %q, want the canonical URL %q", got.dv2.doc.path, eulerArticle.URL)
	}
	if got.focus != dv2Pane {
		t.Errorf("focus = %v, want dv2Pane", got.focus)
	}
	if got.dv1.doc.isOpen() {
		t.Errorf("DV1 Active Document = %q, want untouched (empty)", got.dv1.doc.path)
	}
}

// TestPeekWikipediaArticleStartsPeekWithoutActivating exercises Peek
// semantics directly (fetchArticleInto's peek == true branch): no real
// keybinding reaches this yet (M5.2's own job — see fetchArticleInto's doc
// comment), but the mechanism itself is exactly what M5.2 will call, so
// this tests it the same way M4.1's own tests exercised articleDocument
// ahead of any caller. S7's contract holds: the fetched article shows in
// Peek, the Active Document and Focus are both untouched, and recency is
// not recorded (S4: "Peek does not count").
func TestPeekWikipediaArticleStartsPeekWithoutActivating(t *testing.T) {
	fetcher := &fakeWikipediaFetcher{articles: map[string]wikipedia.Article{euler.Title: eulerArticle}}
	m := sized(t, newModel(), 80, 24).withWikipediaFetcher(fetcher)
	m = m.withDocs(t.TempDir()+"/documents", map[string]docRecord{})

	after, cmd := m.fetchArticleInto(dv1Pane, euler.Title, true)
	if after.focus != m.focus {
		t.Errorf("focus changed to %v on Peek dispatch, want unchanged %v (S7)", after.focus, m.focus)
	}
	got := runArticleFetch(t, after, cmd)

	if !got.dv1.peeking() {
		t.Fatal("dv1.peeking() = false after a Peek fetch completed")
	}
	if got.dv1.peekLabel() != eulerArticle.Title {
		t.Errorf("peekLabel = %q, want %q", got.dv1.peekLabel(), eulerArticle.Title)
	}
	if got.dv1.doc.isOpen() {
		t.Errorf("Active Document = %q after a Peek, want untouched (empty)", got.dv1.doc.path)
	}
	if got.focus != m.focus {
		t.Errorf("focus = %v after a Peek completed, want unchanged %v (S7)", got.focus, m.focus)
	}
	if rec := got.docRecord(eulerArticle.URL); !rec.LastViewed.IsZero() {
		t.Error("LastViewed is set after a Peek, want recency not recorded (S4)")
	}
}

// TestWikipediaFetchFailureLeavesDocumentUntouchedAndReportsMessage is S3's
// failure case: "a failed fetch ... surfaces as a Footer message and leaves
// the viewer's current document untouched" — driven for Activate, whose
// target already has an Active Document in place to prove untouched
// against (a Peek failure has nothing analogous to leave behind, since
// startPeek is never reached).
func TestWikipediaFetchFailureLeavesDocumentUntouchedAndReportsMessage(t *testing.T) {
	fetcher := &fakeWikipediaFetcher{errs: map[string]error{euler.Title: errors.New("offline")}}
	m := wikipediaModeWithResult(t, fetcher, euler)
	// Give DV1 a real document first, so "untouched" is a meaningful claim.
	m.dv1 = m.dv1.open(document{path: "/tmp/existing.md", src: []byte("# Existing\n")})
	beforeDoc := m.dv1.doc

	after, cmd := m.activatePickerInto(dv1Pane)
	got := runArticleFetch(t, after, cmd)

	if got.message == "" {
		t.Error("message is empty after a failed fetch, want a Footer message")
	}
	if got.dv1.doc.path != beforeDoc.path {
		t.Errorf("DV1 Active Document = %q after a failed fetch, want untouched %q", got.dv1.doc.path, beforeDoc.path)
	}
	if got.dv1.fetching() {
		t.Error("dv1.fetching() = true after a failed fetch, want the indicator cleared")
	}
}

// TestPeekWikipediaFetchFailureLeavesPeekUntouched is the Peek-targeted twin
// of TestWikipediaFetchFailureLeavesDocumentUntouchedAndReportsMessage: a
// failed Peek fetch must not start Peek at all (there is nothing to leave
// "untouched" the way an Activate-targeted failure leaves the Active
// Document, but startPeek itself must simply never run) and still reports
// the same kind of Footer message.
func TestPeekWikipediaFetchFailureLeavesPeekUntouched(t *testing.T) {
	fetcher := &fakeWikipediaFetcher{errs: map[string]error{euler.Title: errors.New("offline")}}
	m := sized(t, newModel(), 80, 24).withWikipediaFetcher(fetcher)

	after, cmd := m.fetchArticleInto(dv1Pane, euler.Title, true)
	got := runArticleFetch(t, after, cmd)

	if got.message == "" {
		t.Error("message is empty after a failed Peek fetch, want a Footer message")
	}
	if got.dv1.peeking() {
		t.Error("dv1.peeking() = true after a failed Peek fetch, want Peek never started")
	}
	if got.dv1.fetching() {
		t.Error("dv1.fetching() = true after a failed Peek fetch, want the indicator cleared")
	}
}

// TestStaleWikipediaFetchDroppedWhenTargetActivatesElsewhere is PLAN M4.2's
// "a completion whose target moved on ... is dropped" for Activate: DV1
// navigates to something else (a plain file) while a Wikipedia fetch is
// still in flight for it; the fetch's own token is invalidated the moment
// open() runs (viewer.pendingFetch's own doc comment), so its late
// completion must not overwrite the file that has since been activated.
func TestStaleWikipediaFetchDroppedWhenTargetActivatesElsewhere(t *testing.T) {
	fetcher := &fakeWikipediaFetcher{articles: map[string]wikipedia.Article{euler.Title: eulerArticle}}
	m := wikipediaModeWithResult(t, fetcher, euler)

	after, cmd := m.activatePickerInto(dv1Pane)
	if !after.dv1.fetching() {
		t.Fatal("setup: dv1 is not fetching after Activate")
	}
	// The target "moves on" before the fetch resolves: a real file lands in
	// DV1 via the ordinary open() path (navigate's own underlying call).
	moved := after.navigate(dv1Pane, document{path: "/tmp/other.md", src: []byte("# Other\n")})
	if moved.dv1.fetching() {
		t.Fatal("setup: dv1 still reports fetching after a real navigation, want it cleared")
	}

	got := runArticleFetch(t, moved, cmd)

	if got.dv1.doc.path != "/tmp/other.md" {
		t.Errorf("DV1 Active Document = %q after a stale completion arrived, want the file it moved on to (%q)", got.dv1.doc.path, "/tmp/other.md")
	}
	if got.message != "" {
		t.Errorf("message = %q after a dropped stale completion, want none", got.message)
	}
}

// TestActivateWikipediaFetchSurvivesAnUnrelatedFocusChange guards against a
// review-caught regression: viewer.clearPeek is model.clearPeek/clearPeekIn's
// shared invalidation chokepoint, called from far more places than a literal
// Peek clearing or Picker reopen — every Focus change (model.setFocus), a
// refresh, a Files Nav selection move, a mouse-wheel scroll over it. None of
// those are "the target moved on" for an Activate-targeted fetch (peek ==
// false): the user merely glanced elsewhere while a fetch that will still
// land in DV1 is in flight, and S3/M4.2 name only Peek clearing and Picker
// reopening as invalidating triggers, not an unrelated Focus change. Only a
// Peek-targeted pendingFetch (peek == true) may be cleared by clearPeek — see
// viewer.pendingFetch's own doc comment.
func TestActivateWikipediaFetchSurvivesAnUnrelatedFocusChange(t *testing.T) {
	fetcher := &fakeWikipediaFetcher{articles: map[string]wikipedia.Article{euler.Title: eulerArticle}}
	m := wikipediaModeWithResult(t, fetcher, euler)

	after, cmd := m.activatePickerInto(dv1Pane)
	if !after.dv1.fetching() {
		t.Fatal("setup: dv1 is not fetching after Activate")
	}

	// An unrelated Focus change: to the Files Nav and back, neither of
	// which is Peek clearing or a Picker reopen.
	moved := after.setFocus(filesPane)
	moved = moved.setFocus(dv1Pane)
	if !moved.dv1.fetching() {
		t.Fatal("dv1.fetching() = false after an unrelated Focus change, want the Activate-targeted fetch to survive it")
	}

	got := runArticleFetch(t, moved, cmd)

	if got.dv1.doc.path != eulerArticle.URL {
		t.Errorf("DV1 Active Document = %q after an unrelated Focus change, want the article to still land: %q", got.dv1.doc.path, eulerArticle.URL)
	}
}

// TestStaleWikipediaFetchDroppedWhenPeekCleared is PLAN M4.2's other named
// case: reopening the Picker clears Peek (openPicker's own clearPeek call)
// — the same event that would clear a real, already-showing Peek — which
// must invalidate a Peek-targeted fetch already in flight so its late
// completion cannot start Peek back up after the fact.
func TestStaleWikipediaFetchDroppedWhenPeekCleared(t *testing.T) {
	fetcher := &fakeWikipediaFetcher{articles: map[string]wikipedia.Article{euler.Title: eulerArticle}}
	m := sized(t, newModel(), 80, 24).withWikipediaFetcher(fetcher)

	after, cmd := m.fetchArticleInto(dv1Pane, euler.Title, true)
	if !after.dv1.fetching() {
		t.Fatal("setup: dv1 is not fetching after the Peek dispatch")
	}

	// Reopening the Picker clears Peek (model.openPicker's own clearPeek
	// call) — one of PLAN M4.2's two named "target moved on" triggers.
	cleared, _ := after.openPicker()
	if cleared.dv1.fetching() {
		t.Fatal("setup: dv1 still reports fetching after openPicker, want it cleared")
	}

	got := runArticleFetch(t, cleared, cmd)

	if got.dv1.peeking() {
		t.Error("dv1.peeking() = true after a dropped stale Peek completion, want it to stay cleared")
	}
	if got.message != "" {
		t.Errorf("message = %q after a dropped stale completion, want none", got.message)
	}
}
