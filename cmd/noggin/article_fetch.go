package main

import (
	"fmt"

	tea "charm.land/bubbletea/v2"

	"noggin/internal/wikipedia"
)

// wikipediaFetcher is the seam a Wikipedia article fetch runs against (S3,
// C6/M4.2) — deliberately narrower than the whole *wikipedia.Client, the
// same reasoning pickerWikipediaClient's own doc comment already gives for
// keeping Search separate: a fake exercising Activate/Peek's fetch path
// never has to also implement Search, and a fake exercising a Picker search
// (M3.3's own tests) never has to also implement Fetch. model's own default
// (newModel) points both this and pickerWikipediaClient at the same
// *wikipedia.Client instance, so a real session's Picker search and article
// fetch share one session-scoped cache (M2.3) exactly as wikipediaClient's
// own doc comment already promises.
type wikipediaFetcher interface {
	Fetch(title string, refresh bool) (wikipedia.Article, error)
}

// withWikipediaFetcher replaces the session's article-fetch client (C6/
// M4.2), the Fetch-side counterpart to withWikipediaClient: tests inject a
// fake here instead, so Activate/Peek on a Wikipedia Result never reaches
// the network.
func (m model) withWikipediaFetcher(f wikipediaFetcher) model {
	m.wikipediaFetcher = f
	return m
}

// articleFetchMsg reports a Wikipedia article fetch's completion (S3, C6/
// M4.2): pane and token identify which viewer's pendingArticleFetch this
// answers, so applyArticleFetch can tell a completion that still matters
// from one whose target has moved on (PLAN M4.2's own phrase) — see
// pendingArticleFetch's own doc comment (viewer.go) for the full list of
// what invalidates a token. peek carries Activate/Peek semantics through to
// the handler, since msg alone — not the model — is what a tea.Cmd's
// goroutine can report. refresh marks a C6/M4.3 re-fetch (refreshArticleFetchInto,
// below): mutually exclusive with peek (a refresh always targets an already-
// open Active Document, never a Peek), and tells applyArticleFetch to
// re-render the article in place (viewer.refreshDoc) rather than navigate to
// it or start a Peek. title is the requested title, kept for a failure's
// Footer message; article is only meaningful when err is nil.
type articleFetchMsg struct {
	pane    pane
	token   int
	peek    bool
	refresh bool
	title   string
	article wikipedia.Article
	err     error
}

// fetchArticleInto starts an async Wikipedia article fetch targeting pane
// (S3, C6/M4.2): the Picker's Activate on a Wikipedia Result or a starred
// article surfaced by File mode's Starred scope (both peek == false,
// picker.go's activatePickerInto), or a future Peek trigger (peek == true —
// M5.2's own job to wire up a real keybinding to; this method's peek half is
// exercised directly by this task's own tests in the meantime, the same
// "test the mechanism ahead of its UI trigger" shape M4.1's articleDocument
// already set for this cycle, see its own doc comment in document.go).
//
// target is normalized to dv1Pane unless it is exactly dv2Pane, the same
// rule openDocumentInto/togglePeekInto already apply, and DV2 is created
// first if targeting it and it does not yet exist — the indicator
// (viewer.view/model.paneTitle) needs somewhere to show while the fetch is
// in flight, the same reason openDocumentInto's own ensureDV2 call must run
// before its write.
//
// Activate semantics (peek == false) move Focus to target immediately, the
// same as every other Activate (openDocumentInto's own contract) — the
// fetch itself is what is deferred, not Focus following it; a completion's
// eventual navigate (applyArticleFetch) then finds Focus already there,
// matching a plain Activate. Peek semantics (peek == true) never move Focus
// (S7) — mirrored here even though nothing yet calls this with peek == true
// from a real keypress.
//
// m.articleFetchSeq is a single, model-wide counter, not a per-viewer one:
// its only job is giving every dispatch here a value no other one before or
// after it shares, so two fetches — to the same viewer or different ones —
// can never be confused for each other by applyArticleFetch; nothing about
// it needs to be scoped narrower than that.
func (m model) fetchArticleInto(target pane, title string, peek bool) (model, tea.Cmd) {
	if target != dv2Pane {
		target = dv1Pane
	}
	if target == dv2Pane {
		m = m.ensureDV2()
	}
	if !peek {
		m = m.setFocus(m.resolveFocus(target))
	}
	m.articleFetchSeq++
	token := m.articleFetchSeq
	v := m.viewerFor(target)
	v.pendingFetch = &pendingArticleFetch{title: title, peek: peek, token: token}
	m = m.withViewerFor(target, v)
	fetcher := m.wikipediaFetcher
	return m, func() tea.Msg {
		article, err := fetcher.Fetch(title, false)
		return articleFetchMsg{pane: target, token: token, peek: peek, title: title, article: article, err: err}
	}
}

// refreshArticleFetchInto starts an async re-fetch of target's already-open
// article, bypassing the session cache (S3/C6/M4.3: "R re-fetches any open
// Wikipedia article"). Unlike fetchArticleInto (Activate/Peek), it never
// moves Focus and never creates DV2: refresh only ever acts on a viewer that
// already has an article open (model.refreshViewer's own isArticle() guard
// is what routes here instead of loadDocument), so target already exists
// and already holds whatever Focus it held before R was pressed. Sharing
// fetchArticleInto's pendingFetch/token machinery gives a re-fetch the same
// in-progress indicator and the same "target moved on" staleness guard
// (applyArticleFetch's token check) a plain Activate/Peek fetch already has,
// for the same reasons — a refresh is exactly as much a real network fetch
// as either of those (D119).
func (m model) refreshArticleFetchInto(target pane) (model, tea.Cmd) {
	v := m.viewerFor(target)
	title := v.doc.title
	m.articleFetchSeq++
	token := m.articleFetchSeq
	v.pendingFetch = &pendingArticleFetch{title: title, token: token}
	m = m.withViewerFor(target, v)
	fetcher := m.wikipediaFetcher
	return m, func() tea.Msg {
		article, err := fetcher.Fetch(title, true) // bypass the cache (S3's own Refresh rule)
		return articleFetchMsg{pane: target, token: token, refresh: true, title: title, article: article, err: err}
	}
}

// applyArticleFetch is fetchArticleInto's/refreshArticleFetchInto's shared
// completion handler (S3, C6/M4.2, refresh half C6/M4.3): a stale token —
// msg.pane's viewer either never had this fetch pending (already resolved,
// or dropped by open/refreshDoc/close/startPeek/clearPeek in the meantime,
// see pendingArticleFetch's own doc comment in viewer.go) or has since
// started a newer one of its own — is silently dropped, exactly as a stale
// Picker search generation already is (applyPickerResults' own precedent).
// A real failure surfaces as a Footer message (S3's own wording: "offline,
// article gone, timeout") and leaves the viewer exactly as pendingFetch
// found it — neither open()'s document, startPeek's peek, nor a refresh's
// own previously-open article was ever touched by the dispatching method
// itself, so there is nothing further to undo. A success builds the
// canonical articleDocument (D129, M4.1) and one of three things: a refresh
// (msg.refresh, C6/M4.3) re-renders it in place via viewer.refreshDoc — the
// same scroll-preserving, find-dropping re-render a file's own refresh
// already gives, with no recency/history write since nothing navigated —
// or, for Activate/Peek (mutually exclusive with refresh), starts Peek
// (peek == true, S7's own "never becomes the Active Document, never moves
// Focus, never records recency") or records recency and navigates (peek ==
// false, S4/S9 — recordArticleViewed-then-navigate, the article-aware form
// (C6/M6.1) of the same recordViewed-then-navigate pair activatePickerInto's
// own fileMode branch already gives a file).
func (m model) applyArticleFetch(msg articleFetchMsg) model {
	v := m.viewerFor(msg.pane)
	if v.pendingFetch == nil || v.pendingFetch.token != msg.token {
		return m
	}
	v.pendingFetch = nil
	m = m.withViewerFor(msg.pane, v)
	if msg.err != nil {
		// D61's own priority rule (model.combineDropMessages, model.go): never
		// clobber a message already set. Matters most for msg.refresh (C6/
		// M4.3), whose completion arrives after refresh()'s own synchronous
		// half (the other viewer's read failure or dropped-find message) has
		// already returned — without this guard, a failed article re-fetch
		// would silently erase whatever refresh() already reported about the
		// other viewer in the very same R press.
		if m.message == "" {
			m.message = fmt.Sprintf("cannot fetch %s: %v", msg.title, msg.err)
		}
		return m
	}
	doc := articleDocument(msg.article)
	if msg.refresh {
		return m.withViewerFor(msg.pane, m.viewerFor(msg.pane).refreshDoc(doc))
	}
	if msg.peek {
		return m.withViewerFor(msg.pane, m.viewerFor(msg.pane).startPeek(doc))
	}
	// C6/M6.1: recordArticleViewed, not recordViewed — a first-time Activate
	// on an article never starred before has no existing Article/Title stamp
	// for recordViewed's plain LastViewed-only write to preserve (see its own
	// doc comment).
	m = m.recordArticleViewed(doc.path, doc.title)
	return m.navigate(msg.pane, doc)
}
