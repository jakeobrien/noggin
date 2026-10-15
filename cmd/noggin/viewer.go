package main

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// emptyViewerText is what the Document Viewer says when nothing is open (S7).
const emptyViewerText = "No document open. Pick a file in the Navigator."

// viewer is the Document Viewer: the active document, its rendered form and the
// scroll position within it. Navigation is viewport scrolling only — no cursor,
// no selection (D4). findState is the S8 in-document search state, layered on top.
type viewer struct {
	vp  viewport.Model
	doc document

	// renderedWidth is the width doc was last rendered at, so a resize that
	// leaves the width alone does not pay to re-render.
	renderedWidth int

	// err is a render failure, shown in the pane instead of the document.
	err error

	findState find

	// mode is this viewer's Render Mode (D140, textview.go): sticky across
	// every document opened, Peeked or walked to by history in it, and
	// session-only. The zero value is Markdown, so a fresh viewer (a new DV2
	// included) starts there. modeFor gives the mode a particular document is
	// actually shown in, since a Text file ignores it.
	mode renderMode

	// docMode is the mode the open document is shown in (modeFor, resolved
	// once by render; Markdown with nothing open). Every mode-specific
	// behaviour (docMode.behaviour()) hangs off it.
	docMode renderMode

	// text is the Text layout of the open document, set by render exactly
	// when docMode is Text, nil otherwise. Find searches its source lines, and
	// textMatch is the span of the current match drawn reversed in it (nil
	// when none).
	text      *textLayout
	textMatch *textSpan

	// mutedContent/mutedPeekContent are the same rendered document and Peek
	// in the muted Glamour style (unfocused_style.go; identical line
	// structure, so the viewport's scroll offset and find highlight carry
	// over). muted is set on a per-frame copy by model.frame for an
	// unfocused viewer, and makes view() show them.
	mutedContent     string
	mutedPeekContent string
	muted            bool

	// peek is S7/M7.1's minimal preview state: a document shown in the pane
	// instead of doc, without becoming the Active Document. The zero value
	// (peekState{}, doc.isOpen() == false) means not Peeking, the same
	// sentinel document itself already uses — a pointer would work too, but
	// every other field on viewer is value-semantic (see withOffset's own
	// doc comment) and this keeps that true.
	peek peekState

	// history and historyPos are S9/C4-M6.1's per-viewer back/forward stack:
	// the document paths opened via model.navigate (Activate, the Picker's
	// Activate, and the launch-time document), in visit order, with
	// historyPos indexing the current entry. Zero value (nil, 0) means no
	// history yet, matching every viewer's fresh state — the pair is never
	// populated by open()/refreshDoc themselves (Peek's startPeek, the
	// $EDITOR handoff's re-open, and a Refresh re-render all leave it alone,
	// S9's own list of non-recording events); only pushHistory writes to it,
	// and only model.navigate calls that. A pointer-free, value-semantic pair
	// like every other viewer field (withOffset's own doc comment) — DV2's
	// history is simply whatever *m.dv2 carries, so it disappears the moment
	// closeDV2 drops the pointer, with nothing further to clean up.
	history    []string
	historyPos int

	// docOffsets is S9/C4-M6.2's per-document scroll memory: the last
	// scroll offset left behind for each path this viewer has ever had
	// open, keyed by path rather than by history entry — so the same
	// document recurring at two different history entries (pushHistory
	// allows non-adjacent repeats) still lands at one shared, most-recent
	// position, not two independent ones. Populated and consulted only by
	// open() (below): every path into a document — a plain Activate, the
	// Picker, launch, and history's own goBack/goForward re-open — already
	// funnels through it, so recording/restoring there covers all of them
	// with no extra call sites, the same reasoning M6.1's navigate/history
	// wrapper doc comments already make about reuse over special-casing.
	// refreshDoc and close() deliberately do not touch this map: neither
	// one leaves a document (refreshDoc re-renders the same one in place,
	// S2's own preserve-scroll contract; close's document has become
	// unreadable, nothing worth remembering), so open() is still the only
	// place a document is ever actually left. Session-only and dies with
	// the viewer, exactly like history itself.
	docOffsets map[string]int

	// articleCache holds every article document (C6/M4.1) this viewer has
	// ever had open, keyed by its canonical URL (docOffsets' own sibling,
	// populated and lazily allocated the same way). model.walkViewerHistory
	// (S9's Document History) is what consults it: a file's own history
	// entry is deliberately re-read from disk on Back/Forward ("picking up
	// latest content when navigating back", walkViewerHistory's own doc
	// comment), but an article has nothing on disk to re-read — D114's
	// pointer-only store rule means its content lives only here, in memory
	// for the session, exactly like the Wikipedia Client's own fetch cache
	// (M2.3). Every path that can ever land in history first passed through
	// open() below, so a history entry is never missing from this cache.
	// Session-only and dies with the viewer, exactly like history itself.
	articleCache map[string]document

	// boxHeight is the pane's own content box height, exactly as setSize was
	// last called with — the one thing docContentHeight/peekContentHeight
	// (location.go, S9/M8.2) need that v.vp.Height() itself can no longer
	// give directly, now that vp's operative height is boxHeight minus one
	// row whenever the S9/S9a location line will render beneath it. Kept
	// decoupled from vp.Height() precisely so the Active Document and Peek
	// (whose own location-line visibility can differ) each derive their own
	// viewport height from the same undivided box rather than from each
	// other's already-adjusted one (renderPeek's own former "copy v.vp's
	// height" shortcut would otherwise hand Peek the Active Document's row
	// reservation, not its own).
	boxHeight int

	// pendingFetch is C6/M4.2's async fetch-in-flight marker (S3: "the UI
	// stays responsive and shows a visible in-progress indicator"): non-nil
	// exactly while this viewer is waiting on a Wikipedia article fetch
	// (fetchArticleInto), targeting either this viewer's Active Document
	// (Activate, peek == false) or its Peek (peek == true, a future trigger,
	// M5.2's own job to wire up).
	//
	// open, refreshDoc, close and startPeek — every method below that
	// replaces what this viewer's Active Document or Peek actually shows —
	// clear it unconditionally: each is a real navigation, so it means "the
	// target moved on" (PLAN M4.2's own phrase) regardless of which kind of
	// fetch was pending. clearPeek only clears it when pendingFetch.peek is
	// true: Peek's own shared invalidation chokepoint (model.clearPeek/
	// clearPeekIn) fires from many unrelated events — any Focus change
	// (model.setFocus), a Files Nav selection move, a refresh, reopening the
	// Picker — none of which are a real navigation for an Activate-targeted
	// fetch (peek == false); dropping one of those on, say, a stray Focus
	// change would contradict S3's "on completion the targeted viewer
	// renders the canonical article" for a fetch still legitimately in
	// flight. A Peek-targeted fetch has no such conflict — Peek only ever
	// exists while Focus sits on the Files Nav in the first place, so every
	// one of clearPeek's own triggers already means its target peek is gone
	// too, PLAN M4.2's "Peek cleared" case exactly.
	//
	// Either way, applyArticleFetch (article_fetch.go) tells a completion
	// that still answers this viewer's current wait from a stale one with
	// nothing more than a token comparison. A pointer, not a value
	// (withOffset's own doc comment gives the same reasoning peekState's
	// zero value uses instead): "no fetch pending" needs to be exactly nil,
	// not a zero pendingArticleFetch that happens to compare equal to some
	// in-flight one's zero token.
	pendingFetch *pendingArticleFetch
}

// pendingArticleFetch is one viewer's in-flight Wikipedia article fetch
// (C6/M4.2, see viewer.pendingFetch's own doc comment for when it is set
// and cleared, including the asymmetry between peek's two values there).
// title is the requested title — Fetch's own canonical title/URL is only
// known once it reports back — shown by the in-progress indicator
// (viewer.view/model.paneTitle) and, on failure, in the Footer message.
// peek carries Activate/Peek semantics through to the completion handler
// (applyArticleFetch): peek fetches never become the Active Document, move
// Focus or record recency (S7). token is this fetch's unique identity
// (model.fetchArticleInto's own monotonic counter) — the one thing
// applyArticleFetch actually compares to decide whether an arriving
// articleFetchMsg still answers this pendingFetch or has been superseded.
type pendingArticleFetch struct {
	title string
	// peek == false covers both Activate and a C6/M4.3 Refresh re-fetch —
	// there is no separate refresh field here (unlike articleFetchMsg's own
	// refresh field, article_fetch.go) because both currently need the exact
	// same clearPeek treatment as an ordinary Activate-targeted fetch
	// (viewer.pendingFetch's own doc comment, above): neither is Peek, so
	// clearPeek's peek-only partial clear must leave either one alone. If a
	// future change ever needs to tell a Refresh-in-flight apart from an
	// Activate-in-flight at this layer, that is the moment to add a field
	// here rather than relying on this implicit equivalence.
	peek  bool
	token int
}

// fetching reports whether this viewer has a Wikipedia article fetch in
// flight (C6/M4.2) — the in-progress indicator's own condition, checked by
// view() and model.paneTitle/renderPane ahead of Peek and the Active
// Document, the same "one thing at a time" precedence Peek's own view()
// branch already gives everything beneath it.
func (v viewer) fetching() bool { return v.pendingFetch != nil }

// beginFetch marks pending an async article fetch targeting either the
// Active Document (peek == false) or Peek (peek == true), and resets
// whichever viewport it targets back to its full boxHeight (S9/M8.2):
// view()'s fetching branch takes over the content box unconditionally
// (its own doc comment) and locationDoc's identical "case v.fetching():
// false" draws no location line beneath it, so nothing should stay
// reserved for one meanwhile — without this, a pane whose location line was
// already showing before the fetch started would leave the fetching
// placeholder one row short of the box, an unstyled blank row where the
// placeholder's own last row belongs (renderBox's Height floor pads
// instead of drawing anything there). article_fetch.go's
// fetchArticleInto/refreshArticleFetchInto are the two callers;
// applyArticleFetch's own completion handler is what restores the correct
// height again once the fetch resolves (success routes through
// open/refreshDoc/startPeek, which already recompute it; a failure
// recomputes it directly, since neither the document nor its own error
// state changed).
func (v viewer) beginFetch(title string, peek bool, token int) viewer {
	v.pendingFetch = &pendingArticleFetch{title: title, peek: peek, token: token}
	if peek {
		// A Peek-targeted fetch can be the very first thing that ever touches
		// v.peek (no Peek open yet to have sized it) — v.peek.vp's own width
		// would still be its zero value without this, unlike v.vp's (always
		// kept current by setSize). renderPeek/startPeek already do the same
		// "copy v.vp's width" once Peek actually starts; this just gives the
		// fetching placeholder the same width a moment earlier.
		v.peek.vp.SetWidth(v.vp.Width())
		v.peek.vp.SetHeight(v.boxHeight)
	} else {
		v.vp.SetHeight(v.boxHeight)
	}
	return v
}

// endFetch clears a resolved fetch (success or failure) and restores
// whichever viewport it targeted to its own current, correct height
// (docContentHeight/peekContentHeight) — the beginFetch's own doc comment
// explains why a failure needs this explicitly: open/refreshDoc/startPeek
// never run on that path, so nothing else would undo beginFetch's own
// unconditional full-boxHeight reset. A success's own follow-up call
// (open/refreshDoc/startPeek) harmlessly recomputes the same height again
// right after, since v.doc/v.err (or v.peek's) have changed by then.
func (v viewer) endFetch(peek bool) viewer {
	v.pendingFetch = nil
	if peek {
		v.peek.vp.SetHeight(v.peekContentHeight())
	} else {
		v.vp.SetHeight(v.docContentHeight())
	}
	return v
}

// peekState is one Peeked document: its own read-only viewport, kept
// entirely separate from vp so Peek can never move the Active Document's
// scroll position — clearing Peek must restore vp exactly as it was found.
// It never scrolls (S7), so GotoTop on every (re-)render is deliberate, not
// a no-op that happens not to matter.
type peekState struct {
	doc document
	vp  viewport.Model
	err error
}

func (p peekState) isOpen() bool { return p.doc.isOpen() }

func newViewer() viewer {
	vp := viewport.New(viewport.WithWidth(0), viewport.WithHeight(0))
	// The viewport's own KeyMap carries key literals; leaving it populated
	// would put bindings outside keymap.go (D2) and duplicate the ones below.
	vp.KeyMap = viewport.KeyMap{}
	// v2 stops padding short content out to the viewport's height unless asked;
	// the pane contract is a content box of exactly its size.
	vp.FillHeight = true
	return viewer{vp: vp, findState: newFind()}
}

// setSize sets the content box the viewer draws into, re-rendering (and so
// re-wrapping) only when the width actually changed. Peek's own viewport
// (S7, M7.1) always tracks the same box dimensions here, even on a
// height-only change render() would skip — render()/renderPeek only re-wrap
// the content on a width change, but a stale height would clip Peek wrong
// on a height-only resize if this didn't run unconditionally.
func (v viewer) setSize(width, height int) viewer {
	v.boxHeight = max(height, 0)
	v.vp.SetWidth(max(width, 0))
	v.vp.SetHeight(v.docContentHeight())
	if v.peeking() {
		v.peek.vp.SetWidth(max(width, 0))
		v.peek.vp.SetHeight(v.peekContentHeight())
	}
	if v.vp.Width() != v.renderedWidth {
		v = v.render()
	}
	return v
}

// open makes d the active document, landing at that document's own
// remembered scroll position (S9/C4-M6.2's docOffsets, above) if this
// viewer has ever had it open before, or the top otherwise — the same
// "top on a document never seen in this viewer" behavior open() always had,
// now shared with the doc-revisit case rather than unconditional. The
// currently-open document's own offset (if any) is banked into docOffsets
// first, before it is replaced, since v.offset() only ever reads the
// outgoing document's position.
func (v viewer) open(d document) viewer {
	// C6/M4.2: a real load answers (or bypasses) whatever fetch this viewer
	// might have had pending — "the target moved on", per pendingFetch's own
	// doc comment.
	v.pendingFetch = nil
	if v.doc.isOpen() {
		v = v.rememberOffset()
	}
	v.doc = d
	if d.isArticle() {
		if v.articleCache == nil {
			v.articleCache = map[string]document{}
		}
		v.articleCache[d.path] = d
	}
	v = v.render()
	if off, ok := v.docOffsets[d.path]; ok {
		v.vp.SetYOffset(off)
	} else {
		v.vp.GotoTop()
	}
	return v
}

// rememberOffset banks the currently-open document's scroll position into
// docOffsets, keyed by its path — open()'s own "leaving a document" step.
// Lazily allocates the map on first use rather than in newViewer(), keeping
// every other viewer field's zero-value-means-empty contract (a viewer that
// has never left a document costs nothing beyond a nil map).
func (v viewer) rememberOffset() viewer {
	if v.docOffsets == nil {
		v.docOffsets = map[string]int{}
	}
	v.docOffsets[v.doc.path] = v.offset()
	return v
}

// refreshDoc replaces the open document with a freshly re-read d (S2,
// C2a/M5.1) and re-renders it in place, deliberately not following open()'s
// GotoTop(): render's SetContent call already clamps the offset down to the
// new end if the document shrank past it (viewport.SetContentLines), which
// is exactly S2's "clamped to the end" contract, so no extra clamping is
// needed here — leaving the offset alone is what preserves scroll position
// otherwise. Find state is dropped the same way any other re-render already
// drops it (T34/T43, render's own doc comment) — S2 allows this ("Find
// state may be lost").
//
// C6/M4.3's own caller (applyArticleFetch's refresh branch) hands this an
// article re-fetched with the cache bypassed: d replaces articleCache's own
// entry for its URL the same way open() populates one, so a later Back/
// Forward through history (walkViewerHistory) reopens the freshly-refreshed
// content rather than the stale copy cached before R was pressed — refresh
// is the one caller other than open() that ever leaves a genuinely new
// version of an article's content behind for the same path.
func (v viewer) refreshDoc(d document) viewer {
	v.pendingFetch = nil // C6/M4.2: see pendingFetch's own doc comment.
	v.doc = d
	if d.isArticle() {
		if v.articleCache == nil {
			v.articleCache = map[string]document{}
		}
		v.articleCache[d.path] = d
	}
	return v.render()
}

// close empties the viewer: refresh's path (S2, C2a/M5.1) for when the
// Active Document has vanished or become unreadable underneath it — the
// empty state open() never produces, since it is only ever called with a
// document that just loaded successfully.
func (v viewer) close() viewer {
	v.pendingFetch = nil // C6/M4.2: see pendingFetch's own doc comment.
	v.doc = document{}
	return v.render()
}

// pushHistory appends path as the S9 history's new current entry (C4/
// M6.1), discarding any forward entries left by an earlier goBack —
// browser semantics: "activating a new document after going back discards
// the forward entries". A no-op when path is already the current entry,
// which is what keeps the same document from ever landing in two adjacent
// entries (S9's own rule) — every other insertion only ever appends after
// truncating to historyPos, so this one check at push time is enough to
// hold that invariant across the whole stack, not just the latest push.
// The empty-history case is handled separately since slicing history[:1]
// against a nil/empty slice would panic.
func (v viewer) pushHistory(path string) viewer {
	if len(v.history) == 0 {
		v.history = []string{path}
		v.historyPos = 0
		return v
	}
	if v.history[v.historyPos] == path {
		return v
	}
	kept := append([]string{}, v.history[:v.historyPos+1]...)
	v.history = append(kept, path)
	v.historyPos = len(v.history) - 1
	return v
}

// canGoBack reports whether the S9 history has an earlier entry to move to.
func (v viewer) canGoBack() bool { return v.historyPos > 0 }

// canGoForward reports whether the S9 history has a later entry to move to
// — true only after goBack has moved off the most recent entry and nothing
// has pushed a fresh one since (pushHistory's own truncation).
func (v viewer) canGoForward() bool { return v.historyPos < len(v.history)-1 }

// goBack and goForward move the S9 history position one entry in their own
// direction, returning the path to open there — model.walkViewerHistory is
// what actually loads and opens it, since a viewer method has no access to
// disk. ok is false, v unchanged, at either end of history: there is
// nothing to move to, not a refusal, the same "nothing to do" shape
// FindNext/FindPrevious already give an empty match list.
func (v viewer) goBack() (viewer, string, bool) {
	if !v.canGoBack() {
		return v, "", false
	}
	v.historyPos--
	return v, v.history[v.historyPos], true
}

func (v viewer) goForward() (viewer, string, bool) {
	if !v.canGoForward() {
		return v, "", false
	}
	v.historyPos++
	return v, v.history[v.historyPos], true
}

// peeking reports whether a document is currently Peeked (S7, M7.1).
func (v viewer) peeking() bool { return v.peek.isOpen() }

// peekLabel names the Peeked document for the pane title, or is empty when
// not Peeking — the Peek equivalent of label().
func (v viewer) peekLabel() string {
	if !v.peeking() {
		return ""
	}
	return v.peek.doc.label()
}

// cachedArticle looks up an article previously opened in this viewer by its
// canonical URL (C6/M4.1) — model.walkViewerHistory's own way to reopen an
// article history entry without a network fetch (articleCache's own doc
// comment, above).
func (v viewer) cachedArticle(url string) (document, bool) {
	d, ok := v.articleCache[url]
	return d, ok
}

// startPeek shows d in the pane without disturbing doc or vp (S7, M7.1): a
// fresh viewport sized to match the Active Document's, rendered once at the
// current width. model.togglePeek is the one caller, and is what enforces
// the surrounding rules (not the Active Document, not already Peeking) —
// this method only ever renders whatever it is given.
func (v viewer) startPeek(d document) viewer {
	v.pendingFetch = nil // C6/M4.2: see pendingFetch's own doc comment.
	vp := viewport.New(viewport.WithWidth(v.vp.Width()), viewport.WithHeight(v.vp.Height()))
	vp.FillHeight = true
	v.peek = peekState{doc: d, vp: vp}
	return v.renderPeek()
}

// clearPeek ends Peek, restoring the Active Document's own view — nothing
// to restore on vp itself, since Peek never touched it (S7, M7.1).
func (v viewer) clearPeek() viewer {
	// C6/M4.2: only a Peek-targeted pending fetch is invalidated here — see
	// pendingFetch's own doc comment for why an Activate-targeted one must
	// survive this method's much broader set of callers.
	if v.pendingFetch != nil && v.pendingFetch.peek {
		v.pendingFetch = nil
	}
	v.peek = peekState{}
	return v
}

// renderPeek (re-)renders the Peeked document into its own viewport, sized
// to the pane's current width/height — called from startPeek and from
// render() so a width change (setSize) reflows Peek exactly like the Active
// Document, and a no-op when not Peeking. GotoTop on every call is
// deliberate: Peek never scrolls (S7), so it is always shown from the top
// regardless of what a previous render left YOffset at. toggleMode is the one
// caller that puts the Peek's place back afterwards (D140).
func (v viewer) renderPeek() viewer {
	if !v.peeking() {
		return v
	}
	v.peek.vp.SetWidth(v.vp.Width())
	out, err := v.renderDoc(v.peek.doc, v.peek.vp.Width(), false)
	if err != nil {
		v.peek.err = err
		// S9/M8.2: Peek's own location-line visibility (peekContentHeight)
		// is independent of the Active Document's (v.vp.Height() may now
		// differ from v.peek.vp.Height() — boxHeight's own doc comment) --
		// no longer simply "copy v.vp's height", now that the two can
		// reserve their row independently.
		v.peek.vp.SetHeight(v.peekContentHeight())
		v.peek.vp.SetContent("")
		return v
	}
	v.peek.err = nil
	v.peek.vp.SetHeight(v.peekContentHeight())
	v.peek.vp.SetContent(strings.TrimRight(out, "\n"))
	v.peek.vp.GotoTop()
	v.mutedPeekContent = ""
	if m, err := v.renderDoc(v.peek.doc, v.peek.vp.Width(), true); err == nil {
		v.mutedPeekContent = strings.TrimRight(m, "\n")
	}
	return v
}

// render rebuilds the viewport's content from the document at the current width.
//
// A re-render invalidates any find in progress (M5.1): a resize re-wraps the
// document, so every match's line number would be wrong, and a re-render also
// means a different document is open (or the same one just re-read from disk,
// M4.1) — either way, stale matches over content that no longer exists.
// toggleMode (textview.go) is the one caller that carries a find across a
// re-render: it re-runs the query in the new mode itself.
func (v viewer) render() viewer {
	v.renderedWidth = v.vp.Width()
	v.err = nil
	v.findState = newFind()
	v.vp.StyleLineFunc = nil
	v.mutedContent = ""
	v.text, v.textMatch = nil, nil
	v.docMode = modeMarkdown
	if v.doc.isOpen() {
		v.docMode = v.modeFor(v.doc)
	}
	var out string
	var err error
	if v.docMode == modeText {
		// One layout serves both the focused and muted content, and Find.
		v.text = layoutText(v.doc.src, v.vp.Width())
		out = v.text.render(false)
	} else if v.doc.isOpen() {
		out, err = v.renderDoc(v.doc, v.vp.Width(), false)
	}
	if !v.doc.isOpen() {
		v.vp.SetHeight(v.docContentHeight())
		v.vp.SetContent("")
	} else if err != nil {
		v.err = err
		v.vp.SetHeight(v.docContentHeight())
		v.vp.SetContent("")
	} else {
		// docContentHeight is set only now, not up front: it reads v.err,
		// which this branch (and the err one above) only just settled
		// (S9/M8.2 — the location line, and so the row it costs, depends
		// on whether rendering actually succeeded, not just whether a
		// document is open).
		v.vp.SetHeight(v.docContentHeight())
		// Glamour pads the document with trailing blank lines; they would
		// be scrollable emptiness past the end.
		if v.docMode == modeText {
			// Not trimmed: the layout's row count is what Find indexes by.
			v.vp.SetContent(out)
			v.mutedContent = v.text.render(true)
		} else {
			v.vp.SetContent(strings.TrimRight(out, "\n"))
			if m, err := v.renderDoc(v.doc, v.vp.Width(), true); err == nil {
				v.mutedContent = strings.TrimRight(m, "\n")
			}
		}
	}
	// A width change reflows Peek (S7, M7.1) exactly like the Active
	// Document above — render() is the one path both go through on a
	// setSize; a doc-only re-render (open/refreshDoc/close) re-renders Peek
	// too, harmlessly, since renderPeek is a no-op when not Peeking.
	return v.renderPeek()
}

// offset is the viewport's current vertical scroll position, in rendered lines.
func (v viewer) offset() int { return v.vp.YOffset() }

// withOffset moves the viewport to a given scroll position directly, for
// tests that need to start from a known offset rather than scroll there one
// action at a time. A value receiver, like every other viewer method (T10) —
// the pointer receiver this used to carry bought nothing here, since
// SetYOffset already mutates the shared vp.Model by reference.
func (v viewer) withOffset(n int) viewer { v.vp.SetYOffset(n); return v }

// atBottom reports whether the viewport is scrolled to the end of the document.
func (v viewer) atBottom() bool { return v.vp.AtBottom() }

// lineCount is the number of rendered lines in the open document.
func (v viewer) lineCount() int { return v.vp.TotalLineCount() }

// height is the viewport's content height, in rows.
func (v viewer) height() int { return v.vp.Height() }

// label names the open document for the pane title, or is empty when none is.
func (v viewer) label() string { return v.doc.label() }

func emptyViewerStyle() lipgloss.Style { return lipgloss.NewStyle().Foreground(theme.Subtle) }

// view renders the viewer's content box at exactly its width x height.
func (v viewer) view() string {
	if v.vp.Width() <= 0 || v.vp.Height() <= 0 {
		return ""
	}
	// C6/M4.2: an in-flight fetch takes over the content box ahead of
	// everything below — there is nothing yet to show for either an Active
	// Document about to be replaced (peek == false) or a Peek about to
	// start (peek == true), so this is unconditional on which, the same way
	// Peek's own branch just beneath it is unconditional on the Active
	// Document's own err/empty/open state.
	if v.fetching() {
		// S9/M8.2: a Peek-targeted fetch (pendingFetch.peek) fills Peek's own
		// box, not the Active Document's — beginFetch (above) already resizes
		// whichever one it targets to match, mirroring the identical peek.err
		// branch just below.
		width, height := v.vp.Width(), v.vp.Height()
		if v.pendingFetch.peek {
			width, height = v.peek.vp.Width(), v.peek.vp.Height()
		}
		return v.fill(emptyViewerStyle(), width, height, fmt.Sprintf("Fetching %q\u2026", v.pendingFetch.title))
	}
	// A Peeked document (S7, M7.1) takes over the content box entirely,
	// ahead of the Active Document's own err/empty/open cases below — those
	// describe vp, which Peek never touches or replaces.
	if v.peeking() {
		if v.peek.err != nil {
			// S9/M8.2: Peek's own dimensions, not v.vp's — the two viewports
			// can now differ in height (boxHeight's own doc comment), so
			// filling to the wrong one would over- or under-fill Peek's box.
			return v.fill(emptyViewerStyle(), v.peek.vp.Width(), v.peek.vp.Height(), v.peek.err.Error())
		}
		if v.muted && v.mutedPeekContent != "" {
			vp := v.peek.vp
			vp.SetContent(v.mutedPeekContent)
			return vp.View()
		}
		return v.peek.vp.View()
	}
	switch {
	case v.err != nil:
		return v.fill(emptyViewerStyle(), v.vp.Width(), v.vp.Height(), v.err.Error())
	case !v.doc.isOpen():
		return v.fill(emptyViewerStyle(), v.vp.Width(), v.vp.Height(), emptyViewerText)
	default:
		if v.muted && v.mutedContent != "" {
			vp := v.vp
			vp.SetContent(v.mutedContent)
			return vp.View()
		}
		return v.vp.View()
	}
}

// fill lays text out over a width x height content box, padding and
// truncating so the result is exactly that size — the same contract
// viewport.View() keeps. Takes width/height explicitly (S9/M8.2), rather
// than always reading v.vp's own, since a caller filling Peek's box needs
// Peek's own dimensions, which can now differ from the Active Document's.
func (v viewer) fill(style lipgloss.Style, width, height int, text string) string {
	return style.
		Width(width).Height(height).
		MaxWidth(width).MaxHeight(height).
		Render(text)
}

// scrollActions pairs each scroll binding with the viewport move it performs.
// The four moves S7/S9 allow are line, half-page (bound twice — ctrl+u/
// ctrl+d and, per M8.1, pgup/pgdn/ctrl+b/ctrl+f land on the same offset,
// since a Document Viewer has no separate full-page distance any more), top
// and bottom. Update walks this list, and so can the help modal (S12/C1),
// so a new scroll action is one entry here plus its binding in keymap.go.
//
// keys.PageUp/keys.PageDown still call vp.PageUp()/vp.PageDown() in the
// Picker's own Preview scroll (picker.go's handlePickerInput) — that pane is
// deliberately unchanged by M8.1 (its own doc comment says "other panes
// unchanged"), and is a wholly separate call site from this table, so the two
// panes can move by different distances on the identical key.
var scrollActions = []struct {
	binding key.Binding
	move    func(*viewport.Model)
}{
	{keys.LineUp, func(vp *viewport.Model) { vp.ScrollUp(1) }},
	{keys.LineDown, func(vp *viewport.Model) { vp.ScrollDown(1) }},
	{keys.HalfPageUp, func(vp *viewport.Model) { vp.HalfPageUp() }},
	{keys.HalfPageDown, func(vp *viewport.Model) { vp.HalfPageDown() }},
	{keys.PageUp, func(vp *viewport.Model) { vp.HalfPageUp() }},
	{keys.PageDown, func(vp *viewport.Model) { vp.HalfPageDown() }},
	{keys.Top, func(vp *viewport.Model) { vp.GotoTop() }},
	{keys.Bottom, func(vp *viewport.Model) { vp.GotoBottom() }},
}

// scroll applies the scroll action msg is bound to, reporting whether it was
// one.
func (v viewer) scroll(msg tea.KeyPressMsg) (viewer, bool) {
	for _, action := range scrollActions {
		if key.Matches(msg, action.binding) {
			action.move(&v.vp)
			return v, true
		}
	}
	return v, false
}
