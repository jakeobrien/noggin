package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"noggin/internal/tree"
	"noggin/internal/wikipedia"
)

// pane identifies a focusable region of the frame.
type pane int

const (
	rootsPane pane = iota
	filesPane
	dv1Pane
	dv2Pane
)

// focusOrder is the fallback order resolveFocus walks when the wanted pane
// is not visible: down the Navigator column, then out to DV1. DV2 is
// deliberately absent — falling back onto a pane that does not exist would
// be nonsensical, and DV1 (S1's always-present viewer) is the natural
// landing spot Tab/Shift+Tab's removal (S5) leaves as the one non-Navigator
// fallback. (Tab-cycling itself, and the paneOrder/shiftFocus it drove, are
// gone as of C3/M1.1 — S5 replaces cycling with four direct focus-jump
// keys.)
var focusOrder = []pane{rootsPane, filesPane, dv1Pane}

// title is the pane's heading, shown in its border and the footer. It is a
// short display label, not the GLOSSARY.md term ("Sources Nav", "Files Nav",
// "Document Viewer") — those name the pane in prose, but a title lives on a
// single content row shared with the Active Document's own label
// (paneTitle), so it stays terse on purpose rather than drifting to match the
// glossary.
func (p pane) title() string {
	switch p {
	case rootsPane:
		return "Sources"
	case filesPane:
		return "Files"
	default: // dv1Pane, dv2Pane
		return "Document"
	}
}

// focusKey is the key.Binding that jumps Focus straight to p (S5/S10,
// M7.1) — FocusRoots/FocusFiles/FocusDV1/FocusDV2, the same four bindings
// keymap.go already defines and the Footer used to hint directly (M7.2
// removes that hint list; the badge replaces it on the pane itself).
func (p pane) focusKey() key.Binding {
	switch p {
	case rootsPane:
		return keys.FocusRoots
	case filesPane:
		return keys.FocusFiles
	case dv2Pane:
		return keys.FocusDV2
	default: // dv1Pane
		return keys.FocusDV1
	}
}

// focusBadge is p's title-row badge (S10, M7.1): its focusKey's primary key,
// bracketed, with lipgloss's own "ctrl+" spelling compacted to the terminal
// convention "^" so `ctrl+o` reads as the four-character `[^o]` the
// acceptance criterion asks for, not the much wider `[ctrl+o]` — matching
// how every other badge (`[r]`, `[f]`, `[o]`) already fits comfortably
// beside a title. Derived from p.focusKey() rather than a literal so it
// tracks the binding if it is ever rebound.
func (p pane) focusBadge() string {
	k := strings.ReplaceAll(p.focusKey().Keys()[0], "ctrl+", "^")
	return "[" + k + "]"
}

// model is the root Bubble Tea model: the frame (Navigator column, Document
// Viewer, Footer), keyboard focus, and the terminal size the frame reflows to.
type model struct {
	width  int
	height int
	focus  pane
	files  filesNav

	// dv1 is the always-present Document Viewer (S1, C3). dv2 is the second,
	// optional one beside it (S1) — nil when it does not exist, which is
	// always, for now: M1.1 introduced the pair and the focused-viewer
	// accessors below (viewerFor/focusedViewer and their write halves); M1.3
	// added dv2's Focus-reachability (FocusDV2, a no-op while dv2 is nil, as
	// it always is until M1.4's Files Nav targeting modifier gives it a way
	// to actually create one). A pointer, unlike
	// every other value-semantic field here (viewer.go's withOffset doc
	// comment gives the same reasoning for viewer itself), is what lets
	// "does not exist" be represented directly rather than through a second
	// bool that could drift out of sync with it — safe under model's
	// copy-by-value style only because every write goes through
	// withViewerFor (a fresh *viewer, never an in-place *m.dv2 mutation).
	dv1 viewer
	dv2 *viewer

	// navigatorHidden is the S9 Navigator toggle: hides the whole sidebar
	// (Sources Nav and Files Nav) and gives the Document Viewer the full width
	// when set. It never touches either pane's own state (selection,
	// expansion) — only the layout computed from it (see layout()) — so
	// restoring it leaves that state exactly as it was (M5.2).
	navigatorHidden bool

	// hideDotfiles is S4/M6.2's session-scoped `alt+h` toggle: hides
	// dotfiles (and files under a dot-directory) from both Picker modes'
	// Results (files_search.go's listFiles/isDotfilePath). Lives on model,
	// not picker, deliberately: resetPicker (openPicker/dismissPicker)
	// replaces the whole picker struct with newPicker()'s zero value every
	// close/reopen, which is exactly what a field needing to survive that
	// within a session cannot be subject to — the same reason
	// navigatorHidden lives here rather than on any one pane. Only a fresh
	// model (relaunch) resets it, matching S4's "reset on relaunch".
	hideDotfiles bool

	// roots is the session's note sources (S1): the Sources Nav pane and its
	// selection, doubling as the record of which root is active. message
	// carries a transient footer notice — a launch-time rejection (S1) or a
	// refused action (e.g. activating a broken root, S3); M4.2 will add
	// clear-on-next-keypress and richer hints, but showing it at all cannot
	// wait for that milestone (see D22).
	roots   rootsNav
	message string

	// helpOpen is the S12/C1 Help Modal (M5.3): when set, it owns the keyboard
	// (handleHelpInput) and the frame gains its overlay (overlayHelp), the
	// same shape focusedViewer().findEditing() gives an in-progress search — an
	// exclusive input mode layered on top of the model's other state, not a
	// pane of its own, so it never touches m.focus.
	helpOpen bool

	// ripgrep is the S14 detection result (M1.1): whether the Picker's
	// search layer has a binary to run on, and why not when it doesn't.
	// Detected once at startup (main.go) and carried here so nothing
	// downstream re-shells-out to ask; it does not affect the launch path
	// itself — a session with no ripgrep still launches and reads normally.
	ripgrep ripgrepAvailability

	// picker is the S1 modal's state (M2.1): open/closed, mode, Scope and
	// Prompt. Reachable from every Focus, so — like helpOpen — it never
	// touches m.focus; it is a layer over the frame, not a pane.
	picker picker

	// wikipediaClient is the Wikipedia Source's own client (S2/S3, C6/M3.3):
	// one instance for the session, so a later article fetch (M4.2) shares
	// this same *wikipedia.Client's cache with whatever Picker searches ran
	// before it, rather than each call getting its own throwaway. Held as
	// the pickerWikipediaClient interface (Search only, for now) so a test
	// can inject a fake without a working Fetch; newModel defaults it to a
	// real *wikipedia.Client, whose construction makes no network call
	// itself (wikipedia.NewClient's own doc comment), so every test gets a
	// working default without reaching Wikimedia unless it actually
	// searches.
	wikipediaClient pickerWikipediaClient

	// wikipediaFetcher is Fetch's own seam (S3, C6/M4.2) — wikipediaClient's
	// counterpart, held separately so a Search-only test fake never has to
	// also implement Fetch and vice versa (wikipediaFetcher's own doc
	// comment, article_fetch.go). newModel points both fields at the same
	// *wikipedia.Client instance, so a real session's Picker search and
	// article fetch still share one session-scoped cache, exactly as
	// wikipediaClient's own doc comment above already promises.
	wikipediaFetcher wikipediaFetcher

	// articleFetchSeq is C6/M4.2's article-fetch generation counter: every
	// call to fetchArticleInto takes the next value as its dispatch's own
	// token (pendingArticleFetch, viewer.go), so no two fetches — to the
	// same viewer or different ones — can ever be confused for each other
	// by applyArticleFetch. Model-wide, not per-viewer, since uniqueness is
	// its only job (fetchArticleInto's own doc comment).
	articleFetchSeq int

	// staleBaseline/staleGen/stale are S10's periodic staleness check
	// (C2a/M10.1, staleness.go): the "known good" snapshot the disk is
	// compared against, the generation that invalidates a check dispatched
	// against a since-superseded baseline, and whether the last check found
	// a mismatch. Never mutated by anything but resetStaleBaseline,
	// runStaleCheck/applyStaleResult, and closeDV2's sanctioned withoutDoc
	// prune (C3/M2.1b) — see their own doc comments.
	staleBaseline staleBaseline
	staleGen      int
	stale         bool

	// docsPath and docs are the S1 document store (M1.1's dataDir(), M1.2's
	// path-keyed record, M1.3's eviction): docsPath is resolved once at
	// startup the same way roots.storePath is (withStorePath), and docs is
	// loaded once at the same time (withDocs, main.go's loadDocuments()) and
	// kept in memory from then on. docRecord/withDocRecord are the *only*
	// accessor pair anything — Files Nav, DV1/DV2, Picker — uses to read or
	// write a document's starred/lastViewed state; nothing outside this file
	// touches docs, docsPath or docstore.go's load/save/evict directly.
	docsPath string
	docs     map[string]docRecord
}

func newModel() model {
	// C6/M4.2: one *wikipedia.Client backs both wikipediaClient and
	// wikipediaFetcher, so a real session's Picker search and article fetch
	// share the client's one session-scoped cache (M2.3) — see
	// wikipediaFetcher's own doc comment.
	wikipediaClient := wikipedia.NewClient()
	return model{focus: filesPane, files: newFilesNav(), dv1: newViewer(), roots: newRootsNav(), picker: newPicker(), wikipediaClient: wikipediaClient, wikipediaFetcher: wikipediaClient}
}

// viewerFor is the viewer paired with a Document Viewer pane: dv2Pane maps
// to dv2 (DV1 if dv2 does not exist — the same "falls back to DV1" shape
// focusedViewer gives a non-viewer Focus), anything else to dv1. The read
// half of the pair with withViewerFor.
func (m model) viewerFor(p pane) viewer {
	if p == dv2Pane && m.dv2 != nil {
		return *m.dv2
	}
	return m.dv1
}

// withViewerFor writes v back to the viewer paired with pane p (see
// viewerFor) — creating dv2 if p is dv2Pane and it did not already exist,
// which is exactly the mechanism a later task's "create DV2 first if
// absent" targeting modifier (M1.4) needs; nothing in M1.1 calls it with
// dv2Pane yet, so dv2 stays nil throughout this task.
func (m model) withViewerFor(p pane, v viewer) model {
	if p == dv2Pane {
		m.dv2 = &v
	} else {
		m.dv1 = v
	}
	return m
}

// focusedViewer is the viewer whose pane currently holds Focus (S6): DV2
// when Focus is dv2Pane, DV1 otherwise — including when Focus is on a
// Navigator pane, matching today's single-viewer behavior exactly. This is
// what keeps every viewer-focus-scoped command (Edit, Find, FindNext/
// FindPrevious, scroll) behavior-preserving while DV2 stays unreachable
// (M1.1); it is not used for actions with an explicit, non-focus-dependent
// target — Peek and a plain Activate always target DV1 (S3), never
// "whichever pane has Focus" — so those read/write m.dv1 directly instead.
func (m model) focusedViewer() viewer { return m.viewerFor(m.focus) }

// withFocusedViewer is focusedViewer's write half.
func (m model) withFocusedViewer(v viewer) model { return m.withViewerFor(m.focus, v) }

// focusedViewerPane is the pane focusedViewer's value came from: dv2Pane
// when Focus is already there, dv1Pane otherwise — the target startEdit/
// startFind move Focus onto once they act, and the pane finishEdit reopens
// an edited document into.
func (m model) focusedViewerPane() pane {
	if m.focus == dv2Pane {
		return dv2Pane
	}
	return dv1Pane
}

// withStorePath sets the resolved starred-roots data path (M3.2's
// storePath(), M1.1's dataDir()) that a Sources Nav star toggle persists
// through. main.go resolves it once at startup so a keypress never
// re-derives dataDir() itself.
func (m model) withStorePath(path string) model {
	m.roots.storePath = path
	return m
}

// withDocs installs the S1 document store's resolved path and its startup
// contents (main.go's docStorePath() and loadDocuments()) — the one-shot-
// then-carry pattern withStorePath gives the Roots store, done once here
// rather than re-read on every access. Every subsequent read or write goes
// through docRecord/withDocRecord instead.
func (m model) withDocs(path string, docs map[string]docRecord) model {
	m.docsPath = path
	m.docs = docs
	return m.refreshFilesRecentSource()
}

// refreshFilesRecentSource pushes the S1 document store and the session's
// Roots into the Files Nav (filesNav.withRecentSource) so its Recent
// (M4.2) and Starred (M4.3) flat lists compute against current data on
// every render — the same "snapshot the caller refreshes" shape S8's
// gitStatus/withStatus and S3's filesStarredMarks/withStarred already give
// the pane's other per-file data (see withRecentSource's own doc comment).
// Called from every point either input can change: withDocs (startup),
// withDocRecord (every star/recency write), openFolder and applyResolution
// (the active root, and openFolder's own re-root), and refresh (belt and
// braces — refresh never itself changes docs or roots, but costs nothing to
// recompute alongside the tree/status/starred refresh it already does).
func (m model) refreshFilesRecentSource() model {
	m.files = m.files.withRecentSource(m.docs, m.roots.rootList(), m.roots.activePath())
	return m
}

// docRecord is the document store's read half (M1.4): the persisted state
// for path, or the zero value (unstarred, never viewed) if path has no
// record — which includes a path under no session Root at all (M1.2's "no
// attachment to Roots"), so every caller can look a path up unconditionally
// with no existence check of its own. key normalizes path/URL keys alike
// (docStoreKey, C6/M3.4) — the docs map holds both file paths and article
// canonical URLs (M1.5), and this is the one place both are looked up.
func (m model) docRecord(path string) docRecord {
	return m.docs[docStoreKey(path)]
}

// docStoreKey normalizes a document-store key for docRecord/withDocRecord:
// a file's absolute path is filepath.Clean'd (the pre-M1.5 behaviour, kept
// so a caller can pass an unclean path and still hit an existing record);
// an article's canonical URL (M1.5's other key shape, first written through
// this path by the Picker star, C6/M3.4) is returned untouched, per
// isArticleKey (document.go, C6/M4.1's shared home for this same "://"
// heuristic — filepath.Clean is a filesystem-path operation and would
// corrupt a URL's "//" scheme separator into "/", so it must never run on
// one).
func docStoreKey(path string) string {
	if isArticleKey(path) {
		return path
	}
	return filepath.Clean(path)
}

// withDocRecord is docRecord's write half: sets path's record to rec, bounds
// the store (evictDocStore, M1.3) and persists immediately — S2's "starring
// is persisted and survives restart", generalized to every field this store
// carries. A data directory that could not be resolved at startup
// (m.docsPath == "", the rare case loadDocuments already degrades on) or a
// write error declines the whole update, in-memory state included, with a
// message instead — the same "must not appear to have taken effect, or a
// restart would silently revert it" contract toggleStar already gives the
// Roots store, centralized here once so every future caller (star toggles,
// recency recording) inherits it for free rather than re-deriving it.
//
// It also refreshes the Files Nav's S3 star-marker set (filesStarredMarks)
// on every successful write, not only when the write originated from the
// Files Nav's own toggleDocStar — a star flipped from a focused Document
// Viewer (toggleActiveDocStar) or the Picker (togglePickerStar) changes the
// same m.docs this marker set is derived from, so a Files Nav row for that
// same path would otherwise show a stale marker until its next
// openFolder/refresh (M2.7/TRIAGE T7). recordViewed's lastViewed-only writes
// recompute the same unchanged marks — harmless, and simpler than carving
// out a starred-only variant of this write path.
func (m model) withDocRecord(path string, rec docRecord) model {
	if m.docsPath == "" {
		m.message = "cannot save document store: no resolvable data directory"
		return m
	}
	path = docStoreKey(path)
	docs := make(map[string]docRecord, len(m.docs)+1)
	for p, r := range m.docs {
		docs[p] = r
	}
	docs[path] = rec
	docs = evictDocStore(docs, docStoreCap)
	if err := saveDocStore(m.docsPath, docs); err != nil {
		m.message = fmt.Sprintf("cannot save document store: %v", err)
		return m
	}
	m.docs = docs
	m.files = m.files.withStarred(m.filesStarredMarks())
	return m.refreshFilesRecentSource()
}

// recordViewed stamps path's document record with the current time (S4,
// M3.1's "recency is recorded") and persists it via withDocRecord, the same
// generic write half toggleDocStar/toggleActiveDocStar already use for the
// Starred field — this is that pair's LastViewed analogue. It tolerates a
// path with no existing record the same way docRecord always has (S4's
// "regardless of whether its Root is starred, ad-hoc, or no longer in the
// session", M1.2's "no attachment to Roots"), and only ever touches
// LastViewed, leaving Starred exactly as docRecord already reads it.
func (m model) recordViewed(path string) model {
	rec := m.docRecord(path)
	rec.LastViewed = time.Now()
	return m.withDocRecord(path, rec)
}

// recordArticleViewed is recordViewed's article-aware sibling (C6/M6.1),
// used only by applyArticleFetch's own Activate completion (article_fetch.go):
// unlike a file path, an article's very first view can arrive for a record
// that has never been starred and so has never had toggleActiveDocStar's own
// Article/Title stamp applied — recordViewed's plain "copy the existing
// record, touch LastViewed" shape would otherwise leave such a record
// Article-false/Title-empty forever, which recentRows (C6/M6.1) and
// buildDocBrowseRow both read as a bare, missing-on-disk file rather than an
// article. title (S1's "pointer — canonical URL and title") is stamped
// unconditionally, the same as toggleActiveDocStar's own stamp — a
// re-Activate of an already-starred article just re-writes the same value.
func (m model) recordArticleViewed(path, title string) model {
	rec := m.docRecord(path)
	rec.LastViewed = time.Now()
	rec.Article = true
	rec.Title = title
	return m.withDocRecord(path, rec)
}

// withRipgrep installs the S14 ripgrep detection result (detectRipgrep())
// that main.go resolves once at startup, the same one-shot-then-carry
// pattern as withStorePath.
func (m model) withRipgrep(avail ripgrepAvailability) model {
	m.ripgrep = avail
	return m
}

// withWikipediaClient replaces the session's Wikipedia client (C6/M3.3),
// newModel's own *wikipedia.Client default included: tests use this to
// inject a fake satisfying pickerWikipediaClient instead, so a Wikipedia
// mode search in a test never reaches the network.
func (m model) withWikipediaClient(c pickerWikipediaClient) model {
	m.wikipediaClient = c
	return m
}

// applyResolution installs an S1 launch resolution: the session's roots, which
// one is active, an optional document, and any rejection message. Walking the
// active root or reading the document can fail; that becomes a message, never
// an exit — the session always starts (S1's "silence is not an error").
func (m model) applyResolution(res resolution) model {
	m.roots = m.roots.setRoots(res.roots, res.active)
	m.message = res.message
	// openFolder below already refreshes this when there is an active root
	// to walk; a bare launch with no roots at all never calls it, so the
	// Files Nav would otherwise start every Recent/Starred flat list against
	// a stale (zero-value) snapshot rather than the session's actual (empty)
	// roots — harmless in outcome (no root to browse anyway) but this keeps
	// the invariant "the snapshot always reflects m.roots" unconditional
	// rather than contingent on which branch below runs.
	m = m.refreshFilesRecentSource()

	if path := res.activeRootPath(); path != "" {
		if node, err := tree.Walk(path); err != nil {
			m.message = fmt.Sprintf("cannot read %s: %v", path, err)
		} else {
			m = m.openFolder(node)
		}
	}
	if res.docPath != "" {
		if doc, err := loadDocument(res.docPath); err != nil {
			m.message = fmt.Sprintf("cannot open %s: %v", res.docPath, err)
		} else {
			m = m.openDocument(doc)
		}
	}
	return m
}

// openFolder points the Files Nav at root and puts focus on it, which is what
// browsing a folder means. S8's git status markers are (re)computed here,
// fresh on every call — both the launch-time path (applyResolution) and
// every Root switch (activateRoot) count as a "load" for S8's purposes, so
// neither is worth special-casing separately from the other.
func (m model) openFolder(root *tree.Node) model {
	m.files = m.files.setRoot(root).withStatus(gitStatus(root.Path)).withStarred(m.filesStarredMarks())
	m = m.refreshFilesRecentSource()
	m = m.setFocus(m.resolveFocus(filesPane))
	// S10/C2a/M10.1: this is a fresh read of the Active Root's tree, so it
	// is also the new "known good" the periodic staleness check compares
	// against from here on.
	m = m.resetStaleBaseline()
	return m
}

// openDocument makes d DV1's Active Document and puts Focus on DV1 — the
// unmodified-key target every "plain" load uses (S3): Files Nav's Activate,
// the Picker's Activate, launch-time resolution (applyResolution) and
// finishEdit's own DV1 case. openDocumentInto is the general form a
// Targeting-modified load (a later task) or an edit resumed on DV2
// (finishEdit) needs instead.
func (m model) openDocument(d document) model {
	return m.navigate(dv1Pane, d)
}

// openDocumentInto makes d pane p's Active Document and puts Focus on p,
// which is what asking to read a document into a specific viewer means. p
// is normalized to dv1Pane unless it is exactly dv2Pane — the only other
// valid target — so a zero-value editFinishedMsg.pane (rootsPane, pane's
// own zero value) from an older test literal that predates the pane field
// still lands on DV1, matching every plain load's own target (S3).
func (m model) openDocumentInto(p pane, d document) model {
	if p != dv2Pane {
		p = dv1Pane
	}
	// S3's "creating DV2 first if it does not already exist": every load
	// targeting DV2 routes through here (Files Nav's ActivateDV2, M1.5's
	// Picker, an edit resumed on DV2), so creation lives here once rather
	// than at each call site. It must happen *before* the read below:
	// viewerFor(dv2Pane) falls back to DV1 while dv2 is nil, which would
	// open d into a copy of DV1 (its size, scroll and Find state included).
	if p == dv2Pane {
		m = m.ensureDV2()
	}
	m = m.withViewerFor(p, m.viewerFor(p).open(d))
	m = m.setFocus(m.resolveFocus(p))
	// S10/C2a/M10.1: this is a fresh read of the Active Document, so it is
	// also the new "known good" the periodic staleness check compares
	// against from here on.
	m = m.resetStaleBaseline()
	return m
}

// navigate is openDocumentInto's history-recording form (C4/M6.1): every
// "real" navigation to a new document — the Files Nav's Activate (plain and
// DV2-targeted), the Picker's Activate, and the launch-time document
// (applyResolution) — pushes onto the target viewer's own S9 back/forward
// stack in the same call that opens it (viewer.pushHistory). p is expected
// to already be dv1Pane or dv2Pane, exactly as openDocumentInto itself
// expects, since every caller here already resolved it (activateInto's
// target, the dv1Pane literal above, the Picker's target). openDocumentInto
// itself stays history-blind on purpose: the $EDITOR handoff (finishEdit)
// and history's own traversal (walkViewerHistory, below) both call it, or
// its underlying pieces, directly instead — neither counts as a
// navigation S9 wants recorded ("the $EDITOR handoff and Refresh do not
// create entries").
func (m model) navigate(p pane, d document) model {
	m = m.openDocumentInto(p, d)
	return m.withViewerFor(p, m.viewerFor(p).pushHistory(d.path))
}

// walkViewerHistory is historyBack/historyForward's shared step (C4/M6.1):
// asks the focused viewer to move one entry along its own S9 stack (dir is
// viewer.goBack or viewer.goForward), then reopens that entry directly into
// the same pane (viewer.open, which restores that document's own remembered
// scroll position per S9/C4-M6.2's docOffsets, or the top on a document this
// viewer has never had open before), never through navigate/openDocumentInto:
// the pane already holds Focus (no setFocus needed), DV2 already exists if
// that is where Focus is (visible() guarantees it), and this is history
// moving, not a new entry for pushHistory to record. A file's entry is
// re-read fresh from disk, picking up latest content when navigating back;
// an article's entry (C6/M4.1, isArticleKey) has nothing on disk to
// re-read, so it comes instead from the viewer's own articleCache — every
// article ever in history was opened through viewer.open first, which
// always populates that cache, so a lookup miss here would mean a broken
// invariant elsewhere, not a normal outcome, and is reported the same as
// any other failed reopen rather than assumed impossible. dir's own ok ==
// false (either end of history) and a Peeked viewer (moved.peeking(),
// matching toggleActiveDocStar's own guard — Peek only ever lives while
// Focus sits on the Files Nav, so this never actually fires today, but
// costs nothing to keep true) are both silent no-ops, the same "nothing to
// do" shape FindNext/FindPrevious give an empty match list, not a refusal.
// A read failure reports the same "cannot open" message activate()'s own
// read failure gives and leaves the model untouched — dir's moved viewer
// only ever exists in a local variable until a successful read is in hand,
// so a failed load costs nothing, the history position included.
func (m model) walkViewerHistory(dir func(viewer) (viewer, string, bool)) model {
	p := m.focusedViewerPane()
	moved, path, ok := dir(m.viewerFor(p))
	if !ok || moved.peeking() {
		return m
	}
	var d document
	if isArticleKey(path) {
		var found bool
		d, found = moved.cachedArticle(path)
		if !found {
			m.message = fmt.Sprintf("cannot open %s: article no longer available this session", path)
			return m
		}
	} else {
		var err error
		d, err = loadDocument(path)
		if err != nil {
			m.message = fmt.Sprintf("cannot open %s: %v", path, err)
			return m
		}
	}
	m = m.withViewerFor(p, moved.open(d))
	return m.resetStaleBaseline()
}

// historyBack and historyForward are keys.HistoryBack/keys.HistoryForward's
// actions (C4/M6.1): each focused Document Viewer walks its own S9
// back/forward stack independently, never the Active Source (walkViewerHistory
// never calls openFolder/setRoot, so m.roots is untouched either way).
func (m model) historyBack() model    { return m.walkViewerHistory(viewer.goBack) }
func (m model) historyForward() model { return m.walkViewerHistory(viewer.goForward) }

// ensureDV2 creates DV2 if it does not exist yet (S1: on demand, never by an
// explicit "open a second viewer" action) and re-lays out the frame, which is
// what splits the viewer region 50/50 (M1.2) and sizes both halves — a viewer
// created without that step would carry newViewer()'s zero size, and DV1 would
// keep drawing at the full width it no longer owns. A no-op once DV2 exists,
// so callers can ask unconditionally. The new viewer starts empty: whatever
// asked for it (a load, a peek) fills it immediately afterward.
func (m model) ensureDV2() model {
	if m.dv2 != nil {
		return m
	}
	v := newViewer()
	m.dv2 = &v
	return m.resize(m.width, m.height, resizeCauseSecondViewer)
}

// closeDV2 is S4/C3/M2.1's `ctrl+x` binding: closes DV2, a no-op if it does
// not exist (mirroring FocusDV2's own "a bare request has no document to
// create/close" shape). Focus is set to dv1Pane directly (not through
// resolveFocus's generic focusOrder fallback) when it was on dv2Pane: DV2 no
// longer being visible would otherwise make resolveFocus(dv2Pane) fall back
// through rootsPane/filesPane/dv1Pane, landing wherever the Navigator
// happens to be visible rather than S4's unconditional "Focus moves to DV1".
// Any other Focus is left untouched, matching S4 exactly. Routing the actual
// re-layout through resize() (the same path ensureDV2 drives, going the
// other way) is what gives DV1 back the full viewer width and re-renders it
// at that width — dropping (and reporting) any find in progress the same
// way any other width change already does, rather than a special case here.
//
// C3/M2.1b (TRIAGE T12): closing DV2 is not itself a fresh read of anything
// (unlike resetStaleBaseline's callers), but it does make DV2's former
// document invisible, so staleBaseline.withoutDoc prunes that one path
// right here rather than leaving it tracked until whatever refresh happens
// next — otherwise a change to that now-invisible document between now and
// then could still flip the combined m.stale signal for something no
// longer on screen anywhere.
//
// C6/M4.2: dropping m.dv2 wholesale also drops any DV2-targeted
// pendingArticleFetch it was carrying, without an explicit clear the way
// open/refreshDoc/close/startPeek/clearPeek each give viewer.pendingFetch
// (pendingFetch's own doc comment, viewer.go) — there is no viewer left
// here to clear it on. Its late completion (applyArticleFetch) still drops
// correctly, but only because viewerFor(dv2Pane) falls back to dv1 once
// m.dv2 is nil, and DV1's own token (if any) can never coincidentally match
// a stale DV2 fetch's (model.fetchArticleInto's counter is model-wide and
// monotonic, never reused) — not because anything here says so directly.
func (m model) closeDV2() model {
	if m.dv2 == nil {
		return m
	}
	if m.dv2.doc.isOpen() {
		m.staleBaseline = m.staleBaseline.withoutDoc(m.dv2.doc.path)
	}
	m.dv2 = nil
	if m.focus == dv2Pane {
		m = m.setFocus(dv1Pane)
	}
	return m.resize(m.width, m.height, resizeCauseViewerClosed)
}

// setFocus moves Focus to p, clearing Peek (S7, M7.1) whenever Focus
// actually changes. Peek only ever lives while Focus sits on the Files Nav
// (see togglePeek, the only place that starts it), so every one of the
// many call sites that can move Focus — focusPane, openDocumentInto,
// openFolder, openDocument, and a resize that resolves Focus off a pane that just
// became invisible — needs the same "leaving filesPane ends Peek" rule.
// Routing every Focus assignment through here is what keeps that a
// structural guarantee instead of a per-call-site rule someone can forget
// to add at the next one (flagged before M7 started, see PROGRESS.md's
// foot-gun list).
func (m model) setFocus(p pane) model {
	if p != m.focus {
		m = m.clearPeek()
	}
	m.focus = p
	return m
}

// activate opens the Files Nav's selected file as the active document,
// moving focus to the viewer (M2.3). A folder or an empty pane is a no-op —
// activating a folder is left for later, per the Files Nav's own
// expand/collapse bindings. A read error is a footer message rather than a
// silent no-op (T15), the same "silence is not an error" shape as
// activateRoot's own refusal below and S1's launch-time version of the rule.
//
// C6/M5.2 widens the return to (model, tea.Cmd): a starred article row
// (activateFlatRowInto's own new branch) opens asynchronously like every
// other Activate-a-Wikipedia-thing path (fetchArticleInto), unlike a file,
// which still opens synchronously with a nil Cmd.
func (m model) activate() (model, tea.Cmd) {
	return m.activateInto(dv1Pane)
}

// activateInto is activate's targeted form (C3/S3, M1.4): the Files Nav's
// selected file loaded into target rather than always DV1 — dv2Pane when the
// ActivateDV2 modifier (ctrl+enter / ctrl+l, D55) was used, dv1Pane for the
// plain key.
// Every refusal above (folder, empty pane, read error) applies identically
// whichever viewer is targeted, and is checked before openDocumentInto runs,
// so a refused load never leaves a newly created, empty DV2 behind.
//
// While the Files Nav is in a flat-list mode (Recent/Starred, M4.2/M4.3)
// this resolves against the flat row on screen instead (selectedFlatRow),
// not the unrelated tree row underneath (TRIAGE T21) — delegating to
// activateFlatRowInto, which also gives Starred's own "broken starred
// document ... cannot be activated" (S3) the same refused-with-a-message
// shape activateRoot gives a broken Root.
//
// C6/M5.2 widens the return to (model, tea.Cmd) for activateFlatRowInto's
// own new article branch; the tree branch below always returns a nil Cmd,
// since Recent/Starred (C6/M6.1 widens Recent to a starred-or-not article,
// not only a starred one) are the only modes an article can appear in at
// all — Tree never shows one.
func (m model) activateInto(target pane) (model, tea.Cmd) {
	if m.files.flatMode() {
		row, ok := m.files.selectedFlatRow()
		if !ok {
			return m, nil
		}
		return m.activateFlatRowInto(row, target)
	}
	n := m.files.selectedNode()
	if n == nil || n.IsDir {
		return m, nil
	}
	d, err := loadDocument(n.Path)
	if err != nil {
		m.message = fmt.Sprintf("cannot open %s: %v", n.Path, err)
		return m, nil
	}
	// S4/M3.1: Activate records recency regardless of which viewer it
	// targets — unlike activateRoot, which never records (S4 names
	// Activate-in-a-viewer, not Root activation).
	m = m.recordViewed(n.Path)
	return m.navigate(target, d), nil
}

// activateFlatRowInto is activateInto's flat-list branch (M4.3): row.broken
// (Starred's own S3 clause; always false for a Recent row, since recentRows
// never produces one for a missing file, and always false for an article
// row too, browseRow.article's own doc comment) is refused with a footer
// message mirroring activateRoot's own broken-Root refusal below, rather
// than the generic "cannot open" read-error wording — the two are
// different refusals (the file was never even attempted to load) even
// though both end in a message.
//
// C6/M5.2: row.article routes through fetchArticleInto (S3) instead of
// loadDocument — row.path is the article's canonical URL here, not a real
// file, the same D129 landmine M4.3's Refresh already sidestepped; row.name
// already carries the article's title (buildDocBrowseRow), exactly
// what Fetch takes, mirroring activatePickerInto's own f.article branch
// (picker.go). fetchArticleInto handles Focus, DV2 creation and recency
// itself on completion (applyArticleFetch), so this returns its Cmd
// directly rather than doing any of that here. Every other row behaves
// exactly like activateInto's tree branch, synchronously, with a nil Cmd.
func (m model) activateFlatRowInto(row browseRow, target pane) (model, tea.Cmd) {
	if row.broken {
		m.message = fmt.Sprintf("%s no longer exists; not activated", row.path)
		return m, nil
	}
	if row.article {
		return m.fetchArticleInto(target, row.name, false)
	}
	d, err := loadDocument(row.path)
	if err != nil {
		m.message = fmt.Sprintf("cannot open %s: %v", row.path, err)
		return m, nil
	}
	m = m.recordViewed(row.path)
	return m.navigate(target, d), nil
}

// activateRoot makes the Sources Nav's selected entry the active root,
// re-rooting the Files Nav (S4; M3.4 verifies the open document survives
// this) while leaving focus free to land in the Files Nav, same shape as
// activate() above. A broken root (missing path) is refused with a footer
// message rather than failing silently (S3), the same shape activate() now
// gives its own read-error path (T15).
func (m model) activateRoot() model {
	row, ok := m.roots.selectedRow()
	if !ok {
		return m
	}
	if row.broken {
		m.message = fmt.Sprintf("%s no longer exists; not activated", row.root.path)
		return m
	}
	node, err := tree.Walk(row.root.path)
	if err != nil {
		m.message = fmt.Sprintf("cannot read %s: %v", row.root.path, err)
		return m
	}
	m.roots = m.roots.activate(m.roots.selected)
	m.message = ""
	return m.openFolder(node)
}

// activateWikipediaSource makes the Wikipedia Query Source the model's
// Active Source (S1, C6/M1.1) — activateRoot's mirror for the one Source
// that is never a Sources Nav row index. It refreshes exactly what a Root
// switch already refreshes (the Files Nav's Recent/Starred snapshot, the
// staleness baseline) so both kinds of Active Source keep those invariants
// unconditionally, not just for Roots. filesNav.setWikipedia (C6/M1.4) puts
// the Files Nav's tree aside for good (m.files.root becomes nil) and
// restricts its Browse Modes to Starred/Recent; every other Root-only
// subsystem — tree walk (nothing to walk without an active Root path), grep
// (scope.go's liveRootsInScope, gated on active < 0), git markers and
// staleness (both gated on m.files.root/the Active Root's path), and
// Refresh (model.refresh's own m.files.root != nil guard) — already
// tolerates "no active Root" exactly the same way an empty session does,
// which is what rootsNav.activateWikipedia's cleared active achieves.
func (m model) activateWikipediaSource() model {
	m.roots = m.roots.activateWikipedia()
	m.files = m.files.setWikipedia()
	m.message = ""
	m = m.refreshFilesRecentSource()
	return m.resetStaleBaseline()
}

// toggleStar flips the Sources Nav selection's starred mark and persists it via
// the S2 store (M3.2's toggleStore), using the storePath resolved once at
// startup (withStorePath). A data directory that could not be resolved
// (m.roots.storePath == "", the same rare case loadStarredRoots degrades on)
// declines the toggle with a message instead (T18), the same as the
// explicit-write-error branch below — a star that cannot be persisted must
// not appear to have flipped, or a restart would silently revert it with no
// explanation.
func (m model) toggleStar() model {
	row, ok := m.roots.selectedRow()
	if !ok {
		return m
	}
	if m.roots.storePath == "" {
		m.message = "cannot save starred roots: no resolvable data directory"
		return m
	}
	if _, err := toggleStore(m.roots.storePath, row.root.path); err != nil {
		m.message = fmt.Sprintf("cannot save starred roots: %v", err)
		return m
	}
	m.roots = m.roots.toggleStarred(m.roots.selected)
	return m
}

// startRename begins editing the Sources Nav's selected Root's Name inline
// (S8/M7.2). The Wikipedia row is refused earlier, in Update's rootsPane
// case (mirroring toggleStar's own split there for Star); this only has to
// guard the other refusal S8 asks for — an unstarred ad-hoc Root cannot be
// renamed until it is starred.
func (m model) startRename() (model, tea.Cmd) {
	row, ok := m.roots.selectedRow()
	if !ok {
		return m, nil
	}
	if !row.root.starred {
		m.message = "star this root before renaming it"
		return m, nil
	}
	next, cmd := m.roots.startRename()
	m.roots = next
	return m, cmd
}

// handleRenameInput is where a keypress goes while the Sources Nav's rename
// input has the keyboard (m.roots.renameEditing()): Cancel discards the
// edit, Activate (enter) commits it (commitRename) — and everything else,
// including keys bound elsewhere, types into the input instead of acting on
// its usual binding. The identical exclusivity handleFindInput already
// gives an in-progress find query.
func (m model) handleRenameInput(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Cancel):
		m.roots = m.roots.cancelRename()
		return m, nil
	case key.Matches(msg, keys.Activate):
		return m.commitRename(), nil
	}
	next, cmd := m.roots.updateRenameQuery(msg)
	m.roots = next
	return m, cmd
}

// commitRename applies the Sources Nav's in-progress rename (S8): the typed
// Name, trimmed, or "" (the default) when it is empty or all whitespace. A
// data directory that could not be resolved (m.roots.storePath == "")
// declines with a Footer message and leaves the edit open rather than
// discarding it silently — the same "must not appear applied if it cannot
// persist" discipline toggleStar already gives a star write, with the
// caveat that here nothing was ever displayed as applied in the first place
// for a Cancel-shaped undo to restore. A write failure gets the identical
// treatment.
func (m model) commitRename() model {
	row, ok := m.roots.selectedRow()
	if !ok {
		m.roots = m.roots.cancelRename()
		return m
	}
	name := strings.TrimSpace(m.roots.renameInput.Value())
	if m.roots.storePath == "" {
		m.message = "cannot save the rename: no resolvable data directory"
		return m
	}
	if _, err := renameStore(m.roots.storePath, row.root.path, name); err != nil {
		m.message = fmt.Sprintf("cannot save the rename: %v", err)
		return m
	}
	m.roots = m.roots.applyRename(name)
	return m
}

// toggleDocStar flips the Files Nav selection's starred mark in the S1
// document store and persists it immediately via withDocRecord, the same
// generic accessor pair M1.4 exposed for every future caller of this store
// — toggleStar (above) is its Roots-store analogue, but this one is not a
// wrapper around it: the two stores are deliberately separate (M1.2's "no
// attachment to Roots"), so a document can be starred whether or not its
// Root is even part of the session. A folder or an empty pane is a no-op,
// the same shape activate() already gives those selections — folders carry
// no document record to flip. withDocRecord itself already declines (with
// a footer message) when the data directory can't be resolved or the write
// fails, so no separate guard is needed here.
func (m model) toggleDocStar() model {
	// While the Files Nav is in a flat-list mode (Recent/Starred, M4.2/M4.3)
	// this resolves against the flat row on screen instead (selectedFlatRow),
	// not the unrelated tree row underneath (TRIAGE T21) — this is what
	// makes S2's "un-starring a document while it is displayed in a Files
	// Nav Starred list removes that row immediately" true for the key
	// itself, not only for a star flip that happened to arrive from
	// elsewhere (a Document Viewer, the Picker): withDocRecord's refresh
	// (below) already made that part true on its own.
	if m.files.flatMode() {
		row, ok := m.files.selectedFlatRow()
		if !ok {
			return m
		}
		rec := m.docRecord(row.path)
		rec.Starred = !rec.Starred
		// C6/M5.1, re-confirmed still correct at C6/M6.1: unlike
		// toggleActiveDocStar, this does not stamp Article/Title itself when
		// row.path isArticleKey — a flat row only ever shows a document
		// already in m.docs (recentRows/starredRows both build rows from
		// existing records), and every article record reaching m.docs at all
		// is already stamped correctly by the time it does: a star
		// (toggleActiveDocStar/togglePickerStar) stamps it directly, and an
		// Activate (recordArticleViewed, C6/M6.1 — the only other way an
		// article record is ever created, ahead of M6.1 the sole reason an
		// unstarred article could reach a flat row at all) stamps it too — so
		// row.path here always already carries a correctly-stamped record
		// regardless of which of those wrote it first.
		// Re-deriving a title from row.name would be actively wrong: it is
		// filepath.Base(row.path), which for a URL yields the slug
		// ("Leonhard_Euler"), not the article's real title. This is why that
		// re-derivation is never needed, even now that Recent's own rows
		// (M6.1) let an unstarred article reach this flat-row branch for the
		// first time — recordArticleViewed already got there first.
		return m.withDocRecord(row.path, rec)
	}
	n := m.files.selectedNode()
	if n == nil || n.IsDir {
		return m
	}
	rec := m.docRecord(n.Path)
	rec.Starred = !rec.Starred
	// S3: the marker updates immediately, not only after the pane's next
	// openFolder/refresh — withDocRecord itself refreshes the Files Nav's
	// star-marker set on every successful write (M2.7), so no separate
	// recompute is needed here. A declined write (m.docsPath == "" or a
	// save error, both m.message-only refusals withDocRecord already
	// makes) leaves m.docs and the marker set untouched either way.
	return m.withDocRecord(n.Path, rec)
}

// filesStarredMarks derives the Files Nav's S3 star-marker set from the S1
// document store: just the Starred paths, membership only (renderRow reads
// this map with a plain `f.starred[path]` lookup, so an absent key already
// reads as "not starred" — no need to store the false ones). Recomputed by
// every caller that could have changed either the store or the pane's
// current root — openFolder, refresh, and (via M2.7's withDocRecord) every
// star toggle regardless of which pane it originated from — rather than
// kept live against m.docs, the same "snapshot the caller refreshes" shape
// S8's gitStatus/withStatus already gives the pane's other per-file marker.
func (m model) filesStarredMarks() map[string]bool {
	marks := make(map[string]bool, len(m.docs))
	for path, rec := range m.docs {
		if rec.Starred {
			marks[path] = true
		}
	}
	return marks
}

// toggleActiveDocStar flips the S1 document store's star mark for the
// focused Document Viewer's Active Document (S2: "acting on that viewer's
// Active Document, so DV1 and DV2 star different documents") — unlike
// toggleDocStar's Files Nav selection, this one is inherently per-viewer:
// focusedViewer already resolves to DV1 or DV2 by Focus (viewerFor), so
// starring from each acts on that viewer's own v.doc.path independently,
// with no extra targeting needed. Silent no-op while Peeking, the same
// shape startEdit/startFind already give (S7/M7.1): the Peeked document is
// never the Active Document, so acting on "the Active Document" here while
// something else is on screen would be exactly the same disturb-what's-not-
// displayed surprise those two guards exist to prevent. With no document
// open, a footer message explains it, the same shape startEdit/startFind
// give their own "nothing to act on" case.
//
// C6/M5.1: for an article (v.doc.isArticle()), Article/Title are stamped on
// the record every time, not just when it is first created — the same
// unconditional shape togglePickerStar's own Wikipedia branch already uses,
// needed because docRecord's zero value (Article: false) is what a
// never-before-starred article's record starts as. Without this, an
// article starred for the first time from a viewer would round-trip through
// saveDocStore as an invalid "file" record (its URL key fails
// filepath.IsAbs) and loadDocStore would silently drop the line on the next
// load — the star would appear to work for the rest of the session but not
// survive a restart. Cheap and harmless to re-stamp on every toggle
// (un-starring included): the title is already known from the open
// document, and a file's Article/Title are never touched here at all.
func (m model) toggleActiveDocStar() model {
	v := m.focusedViewer()
	if v.peeking() {
		return m
	}
	if !v.doc.isOpen() {
		m.message = "no document open to star"
		return m
	}
	rec := m.docRecord(v.doc.path)
	rec.Starred = !rec.Starred
	if v.doc.isArticle() {
		rec.Article = true
		rec.Title = v.doc.title
	}
	return m.withDocRecord(v.doc.path, rec)
}

// editFinishedMsg reports that $EDITOR has exited (S10, D3). path is the
// document that was handed off, carried on the message rather than re-read
// from the viewer at delivery time, since focus and the open document could in
// principle have changed while the terminal was away. err is any failure
// starting or running the editor.
type editFinishedMsg struct {
	path string
	// pane is which Document Viewer was handed to $EDITOR (startEdit's own
	// focusedViewerPane at the time), so finishEdit reopens the re-read
	// document into the same one — DV1 by default, or DV2 once it is
	// reachable and holding Focus (M1.6).
	pane pane
	err  error
}

// startEdit hands the active document to $EDITOR (falling back to vi),
// releasing the terminal for the duration and resuming when it exits (D3).
// It is global, not pane-specific: S10 names "the active document", not "the
// focused pane". With no document open it reports that in the footer rather
// than doing nothing.
func (m model) startEdit() (model, tea.Cmd) {
	v := m.focusedViewer()
	// S7/M7.1: Edit is a silent no-op while Peeking — unlike every other
	// refusal below, not even a footer message, since Peek's whole point is
	// a look that leaves nothing else disturbed.
	if v.peeking() {
		return m, nil
	}
	if !v.doc.isOpen() {
		m.message = "no document open to edit"
		return m, nil
	}
	// S3/C6/M4.3: an article has nothing on disk for $EDITOR to open, and
	// the out-of-scope list rules out building a handoff for one — refused
	// with a Footer message, the same shape as "no document open" above,
	// rather than a silent no-op (Edit is not one of Peek's carve-outs).
	if v.doc.isArticle() {
		m.message = "cannot edit a Wikipedia article"
		return m, nil
	}
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	path := v.doc.path
	p := m.focusedViewerPane()
	return m, tea.ExecProcess(exec.Command(editor, path), func(err error) tea.Msg {
		return editFinishedMsg{path: path, pane: p, err: err}
	})
}

// toggleNavigator flips the S9 Navigator toggle and re-lays-out the frame at
// the current terminal size. Hiding moves focus off a Navigator pane it can
// no longer land in, the same way a resize below the narrow-terminal
// threshold does (resolveFocus); showing it again leaves focus exactly where
// it was, since the Document Viewer stays visible throughout and the toggle
// never touches either Navigator pane's own selection state.
//
// It drives the same resize path a real tea.WindowSizeMsg does (both re-lay
// out the frame and can invalidate a find, see resize's doc comment), but is
// not itself a resize — resizeCauseNavigatorToggle keeps resize's own
// dismissed/cleared messages honest about which one actually happened (T45).
func (m model) toggleNavigator() model {
	m.navigatorHidden = !m.navigatorHidden
	return m.resize(m.width, m.height, resizeCauseNavigatorToggle)
}

// focusPane moves Focus straight to target (S5/S6, M2.1/C3-M1.3: FocusRoots/
// FocusFiles/FocusDV1/FocusDV2), revealing the Navigator first via the same
// resize path toggleNavigator uses if the S9 toggle currently has it hidden
// and target lives inside that column (rootsPane/filesPane) — but only ever
// un-hiding, never the reverse, which is what keeps this one-directional
// where toggleNavigator is a flip. A viewer target (dv1Pane/dv2Pane) never
// triggers this: the Document Viewer is never part of the Navigator column,
// so jumping to it revealing an unrelated, unrequested sidebar would be a
// surprising side effect FocusRoots/FocusFiles's own reveal has no analogue
// for. resolveFocus (not a bare assignment) is what makes this a no-op
// rather than a focus jump to an invisible pane on a terminal too narrow to
// carry the Navigator at all (computeLayout): target only wins here when it
// is actually visible.
func (m model) focusPane(target pane) model {
	if m.navigatorHidden && (target == rootsPane || target == filesPane) {
		m.navigatorHidden = false
		m = m.resize(m.width, m.height, resizeCauseNavigatorToggle)
	}
	m = m.setFocus(m.resolveFocus(target))
	return m
}

// togglePeek is S7/M7.1's `p` binding: shows the Files Nav's selected
// document in the Document Viewer without making it the Active Document
// and without moving Focus (Update's filesPane case is what keeps this
// Files-Nav-only — the binding is matched there, not in the global
// switch). It is also Peek's own "end" path: pressing p again while
// already Peeking clears it (S7's "a second p" clearing rule) rather than
// re-starting on the same node — every other clearing trigger moves Focus,
// changes the selection, or refreshes, none of which a bare second p does,
// so it needs its own toggle-off here. A folder, an empty pane, or the
// Active Document itself (S7's own carve-out, "p on the Active Document is
// a no-op") leave Peek exactly as found; a read failure is a footer
// message, the same "silence is not an error" shape activate()'s own read
// failure already gives.
//
// C6/M5.2 widens the return to (model, tea.Cmd): a starred article row
// (togglePeekFlatRowInto's own new branch) starts Peek asynchronously like
// every other Wikipedia fetch (fetchArticleInto), unlike a file, which
// still starts Peek synchronously with a nil Cmd.
func (m model) togglePeek() (model, tea.Cmd) {
	// S3: the unmodified key always targets DV1 (ctrl+p targets DV2, D55,
	// see togglePeekInto) — an explicit target, not the focused-viewer
	// accessor, since Peek's target is never "whichever pane has Focus" (Peek is
	// Files-Nav-only by design, so Focus is never on a viewer pane here
	// anyway).
	return m.togglePeekInto(dv1Pane)
}

// togglePeekInto is togglePeek's targeted form (C3/S3, M1.4): Peek routed to
// target — dv2Pane under the PeekDV2 modifier (ctrl+p, D55), creating DV2 first
// if absent, dv1Pane for the plain key. Every rule togglePeek's own comment
// lists is per-target here: the second-press toggle-off applies to the peek
// living in *that* viewer (each viewer has its own Peek state, S2, so a peek
// in DV1 and one in DV2 can coexist and are cleared independently), and the
// Active-Document carve-out compares against that viewer's own document. DV2
// is created only after every refusal has been cleared and the document has
// actually loaded — a no-op must not leave an empty DV2 behind — and Focus
// never moves either way (S3), so a freshly created DV2 shows its peek while
// Focus stays in the Files Nav.
//
// While the Files Nav is in a flat-list mode (Recent/Starred, M4.2/M4.3)
// this resolves against the flat row on screen instead (selectedFlatRow),
// not the unrelated tree row underneath (TRIAGE T21, M4.5's own fix for
// Peek — activateInto/toggleDocStar already got this in M4.3/D80) —
// delegating to togglePeekFlatRowInto, which also refuses a broken Starred
// row (S3) with the same "no longer exists" wording activateFlatRowInto
// already gives Activate.
func (m model) togglePeekInto(target pane) (model, tea.Cmd) {
	if target != dv2Pane {
		target = dv1Pane
	}
	// viewerFor falls back to DV1 while dv2 is nil, so every read of the
	// target viewer below is gated on it actually existing.
	exists := target == dv1Pane || m.dv2 != nil
	if exists && m.viewerFor(target).peeking() {
		return m.clearPeekIn(target), nil
	}
	if m.files.flatMode() {
		row, ok := m.files.selectedFlatRow()
		if !ok {
			return m, nil
		}
		return m.togglePeekFlatRowInto(row, target, exists)
	}
	n := m.files.selectedNode()
	if n == nil || n.IsDir {
		return m, nil
	}
	if exists {
		if v := m.viewerFor(target); v.doc.isOpen() && n.Path == v.doc.path {
			return m, nil
		}
	}
	d, err := loadDocument(n.Path)
	if err != nil {
		m.message = fmt.Sprintf("cannot open %s: %v", n.Path, err)
		return m, nil
	}
	if target == dv2Pane {
		m = m.ensureDV2()
	}
	return m.withViewerFor(target, m.viewerFor(target).startPeek(d)), nil
}

// togglePeekFlatRowInto is togglePeekInto's flat-list branch (M4.5), the
// same shape activateFlatRowInto (M4.3) already gives Activate: row.broken
// is refused with a footer message before any load is attempted, otherwise
// this behaves exactly like togglePeekInto's tree branch — the
// already-peeking case is handled by the caller before either branch runs,
// since it depends only on the target viewer, never on which row is
// selected.
//
// C6/M5.2: row.article routes through fetchArticleInto (S3, peek == true)
// instead of loadDocument, the same D129-landmine dodge activateFlatRowInto
// makes for Activate — the already-Peeking/already-Active carve-out just
// above already covers an article correctly (row.path is its canonical URL,
// exactly what v.doc.path holds once open, article_fetch.go's
// articleDocument), so no separate check is needed for it here.
func (m model) togglePeekFlatRowInto(row browseRow, target pane, exists bool) (model, tea.Cmd) {
	if row.broken {
		m.message = fmt.Sprintf("%s no longer exists; not peeked", row.path)
		return m, nil
	}
	if exists {
		if v := m.viewerFor(target); v.doc.isOpen() && row.path == v.doc.path {
			return m, nil
		}
	}
	if row.article {
		return m.fetchArticleInto(target, row.name, true)
	}
	d, err := loadDocument(row.path)
	if err != nil {
		m.message = fmt.Sprintf("cannot open %s: %v", row.path, err)
		return m, nil
	}
	if target == dv2Pane {
		m = m.ensureDV2()
	}
	return m.withViewerFor(target, m.viewerFor(target).startPeek(d)), nil
}

// clearPeek ends Peek (S7, M7.1) in both viewers if either is active; a
// no-op otherwise. The one place every clearing trigger but the second-press
// toggle (togglePeekInto, which clears only the viewer it was aimed at)
// routes through: the Focus-change rule (setFocus) and refresh(), per S7's
// own list. Both are viewer-agnostic events — Peek only ever lives while
// Focus sits on the Files Nav, so a Focus change or a refresh ends every
// peek on screen, DV2's included (S2 keeps the two peeks independent, it
// does not exempt one from a global clearing trigger).
func (m model) clearPeek() model {
	m = m.clearPeekIn(dv1Pane)
	return m.clearPeekIn(dv2Pane)
}

// clearPeekIn ends Peek in one viewer only; a no-op if that viewer is not
// peeking or (for dv2Pane) does not exist.
func (m model) clearPeekIn(p pane) model {
	if p == dv2Pane {
		if m.dv2 != nil {
			v := m.dv2.clearPeek()
			m.dv2 = &v
		}
		return m
	}
	m.dv1 = m.dv1.clearPeek()
	return m
}

// refresh re-reads the Active Root's file tree and the Active Document from
// disk on demand (S2, C2a/M5.1). It is the one place both halves of a
// refresh happen together, so S8's git markers are always recomputed in the
// same pass as the tree they describe, never left stale behind it:
//   - The Files Nav's tree is re-walked from m.files.root's own Path (the
//     Active Root; a nil root, e.g. no launch argument and no roots at all,
//     leaves the pane exactly as it was — nothing to refresh). filesNav.
//     refreshRoot preserves expansion and moves selection to the same node
//     or its nearest surviving ancestor (S2), never touching M3.1's
//     per-Root memory, which is setRoot's own concern, not a live refresh's.
//   - The Document Viewer's document, if any is open, is re-read from its
//     own path. A successful re-read re-renders in place via viewer.
//     refreshDoc, which preserves scroll (clamped to the end if the
//     document shrank, S2) the same way any re-render already does; a
//     failed one empties the viewer via viewer.close (S2's "no longer
//     exists or is unreadable").
//
// Either read failing is a footer message, the same "silence is not an
// error" shape activate()/activateRoot()/finishEdit() already give their own
// read failures — but a Root-walk failure wins over both viewers' read
// failures if more than one happens at once (rare: the Root and one or both
// Active Documents vanishing in the same refresh), since a whole Root going
// missing is the more consequential event. Between DV1 and DV2, DV1 wins
// (C3/M3.1) — the same "DV1 first" precedence resize's own dv1Drop/dv2Drop
// combination (below) already gives the two viewers, and simpler than
// concatenating two unrelated read-failure sentences into one line.
//
// A successful document re-render can still silently drop a committed find
// (T43's case in resize's own doc comment — hasFindMatches was true, now
// isn't): refreshDoc/close both go through viewer.render(), the same
// re-render resize already treats as confusing to lose without saying so.
// An in-progress query (findEditing) can't reach this method at all —
// handleFindInput owns every key but Cancel/Activate while typing, and
// Refresh is no exception — so unlike resize's own two-case check, only the
// committed-match case is live here per viewer. C3/M3.1 extends this to DV2
// (whichever exists) using the same "dv1 …"/"dv2 …" combination resize's own
// dv1Drop/dv2Drop already establish for two simultaneous drops — DV1's own
// wording stays unprefixed when DV2 drops nothing, so every pre-M3.1 test
// reading "search cleared"/"search dismissed" verbatim still passes. Either
// drop message still loses to a read-failure message above, the same
// priority resize gives cause over its own dismissed/cleared text.
func (m model) refresh() (model, tea.Cmd) {
	// S7/M7.1: a refresh is one of Peek's clearing triggers.
	m = m.clearPeek()
	if m.files.root != nil {
		if node, err := tree.Walk(m.files.root.Path); err != nil {
			m.message = fmt.Sprintf("cannot read %s: %v", m.files.root.Path, err)
		} else {
			m.files = m.files.refreshRoot(node).withStatus(gitStatus(node.Path)).withStarred(m.filesStarredMarks())
			m = m.refreshFilesRecentSource()
		}
	}
	// C3/M3.1: not focus-scoped — a refresh acts on both viewers whichever
	// exist, "regardless of which pane holds Focus" (S7), unlike Edit/Find/
	// scroll's "whichever viewer has Focus" routing. dv1Drop/dv2Drop mirror
	// resize's own pair (this doc comment's third paragraph) so the two
	// call sites combine two simultaneous find-drops into m.message the
	// same way. C6/M4.3: an open article's own "drop" step is deferred into
	// its own async fetch cmd (below) instead — see refreshViewer's own doc
	// comment for why there is nothing to combine into dv1Drop/dv2Drop for
	// that case.
	var dv1Drop, dv2Drop string
	var dv1Cmd, dv2Cmd tea.Cmd
	m, dv1Drop, dv1Cmd = m.refreshViewer(dv1Pane)
	if m.dv2 != nil {
		m, dv2Drop, dv2Cmd = m.refreshViewer(dv2Pane)
	}
	m = m.combineDropMessages("refreshed", dv1Drop, dv2Drop)
	// S10/C2a/M10.1: a refresh is the user acting on the indicator ("press
	// R"), so it both clears it and re-baselines against whatever the model
	// ends up showing above — including a read failure's fallback (an
	// emptied viewer, an untouched tree) — since that is the state now
	// actually on screen, known-good by construction. An article's own
	// refetch (dv1Cmd/dv2Cmd, still in flight) contributes no doc path here
	// either way — snapshotStaleness already omits a URL from tracking
	// (staleness.go's own "omit, don't half-track" note, C6/M4.1), which is
	// exactly S3's "the staleness indicator never fires for one".
	m = m.resetStaleBaseline()
	return m, tea.Batch(dv1Cmd, dv2Cmd)
}

// refreshViewer is refresh's per-viewer step (C3/M3.1), applied to whichever
// of dv1Pane/dv2Pane is passed: a no-op if that viewer has no document open,
// otherwise a re-read from its own path — refreshDoc in place on a
// successful read (preserving scroll, S2), close on failure. A read failure
// becomes m.message unless one is already set (the Root-walk failure in
// refresh above wins, per this method's own doc comment); it returns
// "search cleared"/"search dismissed" when the re-render silently dropped
// that viewer's committed find, "" otherwise, for refresh's caller to
// combine with the other viewer's own verdict.
//
// C6/M4.3: an open article routes to refreshArticleFetchInto instead —
// loadDocument would try os.ReadFile on a URL and fail loudly, the exact
// landmine D129's own revisit note named. The returned Cmd is what refresh's
// caller batches to actually run the re-fetch; nothing here can report a
// dropped find synchronously the way the file branch does, since the article
// content (and whatever it does to hasFindMatches) does not change until
// that Cmd's own completion (applyArticleFetch) runs — that asymmetry is
// accepted for this task's scope (a dropped find loses its Footer message
// for an article's own refresh only; the find itself is still dropped, same
// as any other re-render, T34/T43's rule), rather than threading refresh's
// whole message-combining machinery through an async completion.
func (m model) refreshViewer(p pane) (model, string, tea.Cmd) {
	v := m.viewerFor(p)
	if !v.doc.isOpen() {
		return m, "", nil
	}
	if v.doc.isArticle() {
		next, cmd := m.refreshArticleFetchInto(p)
		return next, "", cmd
	}
	path := v.doc.path
	findHadMatches := v.hasFindMatches()
	if d, err := loadDocument(path); err != nil {
		if m.message == "" {
			m.message = fmt.Sprintf("cannot open %s: %v", path, err)
		}
		v = v.close()
	} else {
		v = v.refreshDoc(d)
	}
	m = m.withViewerFor(p, v)
	if findHadMatches && !v.hasFindMatches() {
		return m, "search cleared", nil
	}
	return m, "", nil
}

// combineDropMessages builds refresh's and resize's shared m.message shape
// (TRIAGE T13 — the two call sites' combining switches were identical and
// drift-prone) from each viewer's own find-drop verdict (dv1Drop/dv2Drop,
// "" if that viewer lost nothing), prefixed by cause ("refreshed", or a
// resizeCause's own wording): both drops name each viewer explicitly
// ("dv1 …, dv2 …"), a single drop leaves DV1's own wording unprefixed and
// names DV2's explicitly, matching every pre-C3/M1.6a test's wording.
//
// It never overwrites an existing m.message (TRIAGE T11, D61): a Root-walk
// or per-viewer read failure set earlier in the same refresh/resize call is
// strictly more consequential than a find silently dropping, the same
// priority refreshViewer's own read-failure branch already gives itself —
// before this guard, refresh's switch had none, so a document that failed
// to re-read while it had a committed Find lost its read-failure message to
// the find-dropped one, contradicting refresh's own doc comment.
func (m model) combineDropMessages(cause string, dv1Drop, dv2Drop string) model {
	if m.message != "" || (dv1Drop == "" && dv2Drop == "") {
		return m
	}
	switch {
	case dv1Drop != "" && dv2Drop != "":
		m.message = fmt.Sprintf("%s; dv1 %s, dv2 %s", cause, dv1Drop, dv2Drop)
	case dv1Drop != "":
		m.message = fmt.Sprintf("%s; %s", cause, dv1Drop)
	case dv2Drop != "":
		m.message = fmt.Sprintf("%s; dv2 %s", cause, dv2Drop)
	}
	return m
}

// startFind begins an S8 search over the active document, moving focus to
// the viewer the same way activate() and activateRoot() do — the keys a
// search needs next (typing, Cancel, Activate to commit) all target it. With
// no document open it reports that instead, the same shape as startEdit.
func (m model) startFind() (model, tea.Cmd) {
	v := m.focusedViewer()
	// S7/M7.1: Find is a silent no-op while Peeking, the same shape as
	// startEdit's own guard — the Peeked document "cannot be searched with
	// Find" (S7), and Find only ever targets the Active Document underneath
	// it, never the Peeked one, so the safe answer is to do nothing at all
	// rather than search a document that is not even on screen right now.
	if v.peeking() {
		return m, nil
	}
	if !v.doc.isOpen() {
		m.message = "no document open to find in"
		return m, nil
	}
	next, cmd := v.startFind()
	m = m.withFocusedViewer(next)
	// C2a/M7.3 (D51/TRIAGE T14): routed through setFocus like every other
	// Focus assignment, not a bare m.focus = ..., for consistency with the
	// chokepoint's own stated invariant — harmless in practice today only
	// because the peeking() guard above always fires first, but this way
	// nothing else has to reason about that ordering to trust setFocus.
	m = m.setFocus(m.resolveFocus(m.focusedViewerPane()))
	return m, cmd
}

// handleFindInput is where a keypress goes while a find query has the
// keyboard (focusedViewer().findEditing()): Cancel dismisses it, Activate (enter)
// commits it, and everything else — including keys bound elsewhere, like q
// or the scroll keys — types into the query instead of acting on its usual
// binding. That exclusivity is deliberate: a query for the word "question"
// must not quit the program on its first letter.
func (m model) handleFindInput(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Cancel):
		m = m.withFocusedViewer(m.focusedViewer().cancelFind())
		return m, nil
	case key.Matches(msg, keys.Activate):
		m = m.withFocusedViewer(m.focusedViewer().commitFind())
		return m, nil
	}
	next, cmd := m.focusedViewer().updateFindQuery(msg)
	m = m.withFocusedViewer(next)
	return m, cmd
}

// openHelp opens the S12/C1 Help Modal for the focused pane's context. Unlike
// toggleNavigator, this is one-directional — handleHelpInput is the only
// path that closes it, so the main switch below only ever calls this when
// m.helpOpen is already false (the helpOpen branch there returns early
// otherwise). It never touches m.focus: the modal is a layer over the
// frame, not a pane swap, so closing it always returns focus exactly where
// it was without needing to save or restore it.
func (m model) openHelp() model {
	m.helpOpen = true
	return m
}

// handleHelpInput is where a keypress goes while the Help Modal is open
// (m.helpOpen): Help and Cancel both close it — Help because pressing the
// affordance again is the natural undo, Cancel because it is the generic
// one-key dismissal every modal in this program shares (see keymap.go's
// Cancel doc). Everything else is a no-op: the modal owns the keyboard while
// it is up, the same exclusivity handleFindInput gives an in-progress
// search, so a keypress meant for the pane underneath cannot leak through
// the modal by accident.
func (m model) handleHelpInput(msg tea.KeyPressMsg) model {
	if key.Matches(msg, keys.Help) || key.Matches(msg, keys.Cancel) {
		m.helpOpen = false
	}
	return m
}

// finishEdit resumes after $EDITOR exits, re-rendering the edited document
// from disk (D3's "no file watching" — this is the one explicit re-read) so
// changes are immediately visible. Reusing loadDocument and openDocument keeps
// this the same read-and-open path as M2.3's activate and S1's launch, rather
// than a third copy of it. A failure to launch or a since-vanished file is a
// footer message, not a crash — the same "silence is not an error" shape as
// activate() and applyResolution().
func (m model) finishEdit(msg editFinishedMsg) model {
	if msg.err != nil {
		m.message = fmt.Sprintf("edit failed: %v", msg.err)
		return m
	}
	d, err := loadDocument(msg.path)
	if err != nil {
		m.message = fmt.Sprintf("cannot open %s: %v", msg.path, err)
		return m
	}
	m.message = ""
	// S4/M3.1: the $EDITOR handoff records recency the same as an Activate.
	m = m.recordViewed(msg.path)
	return m.openDocumentInto(msg.pane, d)
}

func (m model) Init() tea.Cmd {
	// S10/C2a/M10.1: arms the first periodic staleness tick against
	// whichever baseline applyResolution already established before the
	// program started (main.go calls it before tea.NewProgram(m).Run()) —
	// the check runs for the life of the session, not once.
	return m.scheduleStaleCheck()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case editFinishedMsg:
		return m.finishEdit(msg), nil
	// The Picker's two async messages (M2.4). Neither is a keypress, so
	// neither clears m.message (S11 ties that to the user's next key, not to
	// a timer maturing or a background search reporting back), and both are
	// gated on the Picker still being open inside their handlers rather than
	// here — a message can outlive the modal that asked for it.
	case pickerSearchMsg:
		return m.dispatchPickerSearch(msg)
	case pickerResultsMsg:
		return m.applyPickerResults(msg), nil
	// C6/M4.2: a Wikipedia article fetch's completion (fetchArticleInto),
	// the same "neither is a keypress" shape as the Picker's own pair above
	// — applyArticleFetch's own stale-token check (article_fetch.go) is what
	// stands in for a Picker-open-style gate here, since there is no single
	// modal this message is scoped to the way pickerResultsMsg is scoped to
	// m.picker.open.
	case articleFetchMsg:
		return m.applyArticleFetch(msg), nil
	// S10/C2a/M10.1's own periodic pair, same shape as the Picker's above:
	// a matured timer dispatches the actual (cheap) disk check, which
	// reports back its own message so the check itself never blocks Update.
	case staleCheckMsg:
		return m.runStaleCheck(msg)
	case staleResultMsg:
		return m.applyStaleResult(msg)
	case tea.WindowSizeMsg:
		return m.resize(msg.Width, msg.Height, resizeCauseWindow), nil
	// tea.MouseWheelMsg (bug fix, paired with MouseModeCellMotion in View):
	// wheel scroll is aimed at whichever pane the cursor sits over — not
	// necessarily the focused one — matching ordinary terminal-mouse
	// expectations rather than requiring a click-to-focus first.
	case tea.MouseWheelMsg:
		return m.handleMouseWheel(msg), nil
	case tea.KeyPressMsg:
		// A transient message (D22) is displaced by the action it reports on;
		// any keypress after that clears it (S11), whether or not this one
		// sets a fresh message of its own.
		m.message = ""
		// A find query in progress owns every key but Cancel and Activate
		// (see handleFindInput); this must run before Quit and the rest below
		// get a look at the keypress.
		if m.focusedViewer().findEditing() {
			return m.handleFindInput(msg)
		}
		// The Sources Nav's rename input (S8/M7.2) owns the keyboard the same
		// exclusive way a find-in-progress does — see handleRenameInput.
		if m.roots.renameEditing() {
			return m.handleRenameInput(msg)
		}
		if m.helpOpen {
			return m.handleHelpInput(msg), nil
		}
		// The Picker (S1) is reachable from every Focus and, once open, owns
		// the keyboard the same exclusive way a find-in-progress or the Help
		// Modal does — see handlePickerInput.
		if m.picker.open {
			return m.handlePickerInput(msg)
		}
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Navigator):
			return m.toggleNavigator(), nil
		case key.Matches(msg, keys.FocusRoots):
			return m.focusPane(rootsPane), nil
		case key.Matches(msg, keys.FocusFiles):
			return m.focusPane(filesPane), nil
		case key.Matches(msg, keys.FocusDV1):
			return m.focusPane(dv1Pane), nil
		// FocusDV2 (C3/S5) is a genuine no-op with DV2 absent, not a fall back
		// through focusOrder the way an ordinary invisible target would be
		// (resolveFocus's fallback is for a target hidden by terminal width,
		// not one that flatly does not exist) — short-circuited here, before
		// focusPane/resolveFocus ever run, so Focus provably never moves.
		case key.Matches(msg, keys.FocusDV2):
			if m.dv2 == nil {
				return m, nil
			}
			return m.focusPane(dv2Pane), nil
		// CloseDV2 (S4, C3/M2.1) lives in this global switch, not a pane-
		// specific case, so it is reachable from every one of Sources Nav/Files
		// Nav/DV1/DV2 — and, like every other binding in this switch, already
		// unreachable while a find is editing, the Help Modal is open or the
		// Picker is open, per the three gates above that return before this
		// switch ever runs.
		case key.Matches(msg, keys.CloseDV2):
			return m.closeDV2(), nil
		case key.Matches(msg, keys.Refresh):
			return m.refresh()
		case key.Matches(msg, keys.Help):
			return m.openHelp(), nil
		case key.Matches(msg, keys.Picker):
			return m.openPicker()
		case key.Matches(msg, keys.Edit):
			return m.startEdit()
		case key.Matches(msg, keys.Find):
			return m.startFind()
		// FindNext and FindPrevious are global like Find itself, not scoped to
		// viewer focus: stepping through matches is meaningless anywhere else,
		// but the query lives in the Document Viewer regardless of which pane
		// currently has focus.
		//
		// C2a/M7.3 (D51/TRIAGE T13): both are silent no-ops while Peeking, the
		// same shape startEdit/startFind already had — without this guard a
		// pre-existing committed Find on the Active Document lets n/N drift
		// the real viewport's YOffset while the screen shows peek.vp instead,
		// invisibly violating S7's "cannot be searched with Find" and leaving
		// clearPeek() to restore the drifted position rather than the exact
		// pre-Peek one.
		case key.Matches(msg, keys.FindNext):
			if !m.focusedViewer().peeking() && m.focusedViewer().hasFindMatches() {
				m = m.withFocusedViewer(m.focusedViewer().findNext())
			}
			return m, nil
		case key.Matches(msg, keys.FindPrevious):
			if !m.focusedViewer().peeking() && m.focusedViewer().hasFindMatches() {
				m = m.withFocusedViewer(m.focusedViewer().findPrevious())
			}
			return m, nil
		}
		// Movement keys belong to whichever pane has focus: a Document Viewer
		// scrolls (S6), the Files Nav moves its selection and opens/closes
		// folders.
		switch m.focus {
		case dv1Pane, dv2Pane:
			if next, scrolled := m.focusedViewer().scroll(msg); scrolled {
				m = m.withFocusedViewer(next)
				return m, nil
			}
			// M2.2: star/un-star acts on this viewer's own Active Document, so
			// DV1 and DV2 star independently — the same shared keys.Star chord
			// the Sources Nav and Files Nav cases each bind to their own action.
			if key.Matches(msg, keys.Star) {
				return m.toggleActiveDocStar(), nil
			}
			// HistoryBack/HistoryForward (S9, C4/M6.1) walk this viewer's own
			// back/forward stack, the same per-pane targeting keys.Star just
			// got above — DV1 and DV2 each own their own history.
			if key.Matches(msg, keys.HistoryBack) {
				return m.historyBack(), nil
			}
			if key.Matches(msg, keys.HistoryForward) {
				return m.historyForward(), nil
			}
		case filesPane:
			if key.Matches(msg, keys.Activate) {
				return m.activate()
			}
			// Peek (S7, M7.1) is Files-Nav-only, unlike every other binding
			// matched above in the global switch — that is what keeps it
			// from firing on the Sources Nav or Document Viewer.
			if key.Matches(msg, keys.Peek) {
				return m.togglePeek()
			}
			// C3/S3's targeting modifiers (M1.4), matched here beside the
			// unmodified keys they modify: same actions, aimed at DV2 and
			// creating it if absent. Neither can be confused with its plain
			// twin — key.Matches compares the whole keystroke string, so
			// "ctrl+enter"/"ctrl+l" never match Activate's "enter" and
			// "ctrl+p" never matches Peek's "p" (D55).
			if key.Matches(msg, keys.ActivateDV2) {
				return m.activateInto(dv2Pane)
			}
			if key.Matches(msg, keys.PeekDV2) {
				return m.togglePeekInto(dv2Pane)
			}
			// M2.1: star/un-star the Files Nav's selected file, the same
			// physical key (keys.Star) the Sources Nav's own case below already
			// binds — the two are matched independently per pane, not shared
			// logic, since they act on two entirely separate stores.
			if key.Matches(msg, keys.Star) {
				return m.toggleDocStar(), nil
			}
			// A selection change is one of Peek's clearing triggers (S7);
			// Expand/Collapse leave the selected node itself unchanged
			// (only the rows around it), so only LineUp/LineDown count.
			//
			// C2a/M7.3 (D51/TRIAGE T16): this also fires on a boundary-clamped
			// LineUp/LineDown that leaves the selected index unchanged (already
			// at the top/bottom) — an accepted broadening of "any selection
			// change", not a bug, since re-arriving at the same file just
			// re-shows the Active Document.
			if key.Matches(msg, keys.LineUp) || key.Matches(msg, keys.LineDown) {
				m = m.clearPeek()
			}
			if next, handled := m.files.handleKey(msg); handled {
				m.files = next
				return m, nil
			}
		case rootsPane:
			// S1/C6/M1.2: the Wikipedia row is a Source like any Root for
			// Activate, but `s` is refused with a Footer message instead of
			// toggleStar's silent no-op ("nothing can add, remove or hide
			// it") — both branches below must be checked before falling
			// into activateRoot/toggleStar, which only know how to act on a
			// Root row (selectedRow's own "false" for this selection).
			if m.roots.selectedIsWikipedia() {
				if key.Matches(msg, keys.Activate) {
					return m.activateWikipediaSource(), nil
				}
				if key.Matches(msg, keys.Star) {
					m.message = "the Wikipedia Source cannot be starred"
					return m, nil
				}
				if key.Matches(msg, keys.Rename) {
					m.message = "the Wikipedia Source cannot be renamed"
					return m, nil
				}
			}
			if key.Matches(msg, keys.Activate) {
				return m.activateRoot(), nil
			}
			if key.Matches(msg, keys.Star) {
				return m.toggleStar(), nil
			}
			if key.Matches(msg, keys.Rename) {
				return m.startRename()
			}
			if next, handled := m.roots.handleKey(msg); handled {
				m.roots = next
				return m, nil
			}
		}
	}
	return m, nil
}

// paneAt returns the pane whose region contains the terminal cell (x, y) —
// 0-based, (0,0) at the top-left, the same coordinate space tea.Mouse uses —
// or false when the cell lands outside every pane (the stale row, the
// footer, a collapsed Navigator's now-absent column, or off the frame
// entirely). layout() only hands back each region's size, not its on-screen
// offset, so the offsets are rebuilt here from the same pieces frame() joins
// in the same order: the Navigator column (Sources Nav above Files Nav) at
// x=0..navWidth, the Document Viewer (DV1, then DV2 if present) beside it —
// both columns sharing the same body height, rows 0..l.viewer.height.
func (m model) paneAt(x, y int) (pane, bool) {
	if x < 0 || y < 0 || x >= m.width || y >= m.height {
		return 0, false
	}
	l := m.layout()
	navWidth := l.navigatorWidth()
	if navWidth > 0 && x < navWidth {
		if y < l.roots.height {
			return rootsPane, true
		}
		if y < l.roots.height+l.files.height {
			return filesPane, true
		}
		return 0, false
	}
	if y >= l.viewer.height {
		return 0, false
	}
	dv1Rect, _ := l.splitViewer(m.dv2 != nil)
	if m.dv2 != nil && x-navWidth >= dv1Rect.width {
		return dv2Pane, true
	}
	return dv1Pane, true
}

// handleMouseWheel scrolls whichever pane the wheel event's cell lands over
// (paneAt) by one row/line per tick — the same magnitude keys.LineUp/
// LineDown already move, just aimed by cursor position instead of Focus. A
// wheel event over the stale row, the footer, or outside the frame is a
// no-op, the same as any other unbound input. Files Nav selection moving is
// one of Peek's clearing triggers (S7), the same as LineUp/LineDown's own
// handling in Update above, so a wheel scroll over the Files Nav clears Peek
// too rather than leaving it stale against a selection the wheel just moved
// past.
func (m model) handleMouseWheel(msg tea.MouseWheelMsg) model {
	mouse := msg.Mouse()
	p, ok := m.paneAt(mouse.X, mouse.Y)
	if !ok {
		return m
	}
	var delta int
	switch mouse.Button {
	case tea.MouseWheelUp:
		delta = -1
	case tea.MouseWheelDown:
		delta = 1
	default:
		return m
	}
	switch p {
	case dv1Pane, dv2Pane:
		v := m.viewerFor(p)
		if delta < 0 {
			v.vp.ScrollUp(1)
		} else {
			v.vp.ScrollDown(1)
		}
		return m.withViewerFor(p, v)
	case filesPane:
		m = m.clearPeek()
		m.files = m.files.move(delta)
	case rootsPane:
		m.roots = m.roots.move(delta)
	}
	return m
}

// resizeCause names what drove a call to resize, so its dismissed/cleared
// find messages (below) can say what actually happened instead of
// hardcoding "window resized" even when the real trigger was a Navigator
// toggle (T45) — both drive the frame through this one path (see resize's
// doc comment), but they are not the same event, and the footer must not
// claim one for the other.
type resizeCause string

const (
	resizeCauseWindow          resizeCause = "window resized"
	resizeCauseNavigatorToggle resizeCause = "sidebar toggled"
	// resizeCauseSecondViewer is ensureDV2's (C3/M1.4): creating DV2 halves
	// the region DV1 was drawing in, which re-renders it exactly as a real
	// resize would — including silently dropping a committed find, the
	// case resize's messages exist to explain. Naming it separately keeps
	// that explanation honest about what actually happened (T45).
	resizeCauseSecondViewer resizeCause = "second viewer opened"
	// resizeCauseViewerClosed is closeDV2's (S4, C3/M2.1): the mirror image
	// of resizeCauseSecondViewer — closing DV2 gives DV1 back the full
	// viewer width, the same kind of width change that can silently drop a
	// find in progress, so it is named separately for the same reason.
	resizeCauseViewerClosed resizeCause = "second viewer closed"
)

// resize sets the terminal size and re-lays-out the frame: resolving focus
// onto a still-visible pane, then sizing each pane's content box from the
// resulting layout (which also reflects the S9 Navigator toggle, see
// layout()). Both a real terminal resize and toggleNavigator drive the frame
// through this one path, so they reflow identically; cause records which of
// the two actually happened, for the messages below.
func (m model) resize(width, height int, cause resizeCause) model {
	m.width, m.height = width, height
	m = m.setFocus(m.resolveFocus(m.focus))
	l := m.layout()
	dv1Rect, dv2Rect := l.splitViewer(m.dv2 != nil)
	viewerContent := dv1Rect.content()
	// A width change re-renders the viewer (setSize), and a re-render
	// silently drops any find in progress (viewer.render's doc comment) --
	// fine when nothing was actively being typed, but a query mid-edit
	// vanishing with no explanation violates S1/S8's "confusion is not
	// allowed" (T34). Catch it here, the one path both a real terminal
	// resize and toggleNavigator drive through, by comparing editing state
	// before and after: toggleNavigator can never trigger this in practice
	// (handleFindInput owns every key but Cancel/Activate while editing, so
	// the Navigator binding never reaches it), but a tea.WindowSizeMsg is
	// not gated the same way.
	//
	// The same re-render just as silently drops a *committed* find (T43):
	// editing is already false post-Enter, so the check above alone misses
	// it, but applyCurrentMatch's indication and FindNext/FindPrevious's
	// traversal are exactly as lost as an in-progress query would be.
	// hasFindMatches() catches that case the same way, checked after the
	// editing case so a find that was both editing and matched (the normal
	// case) reports the more specific "search dismissed" message once, not
	// both. Either message leads with cause (T45) rather than assuming a
	// resize: toggleNavigator can in fact reach the committed-find branch
	// (unlike the mid-edit one above, nothing gates it), so the message must
	// name whichever of the two actually fired.
	//
	// C3/M1.6a: DV2 (once it exists, M1.2+) is exactly as vulnerable to a
	// re-render dropping its find — S2's independence guarantee has to hold
	// at this messaging layer too, not just the routing M1.6 already pinned.
	// dv1Drop/dv2Drop capture each viewer's own dismissed/cleared verdict
	// ("" if that viewer lost nothing) before they combine into the single
	// m.message slot below; DV1's own wording is left exactly as it was
	// (unprefixed) when DV2 drops nothing, so every pre-M1.6a test reading
	// "search cleared"/"search dismissed" verbatim still passes — DV2's is
	// named explicitly ("dv2 search …") since an unprefixed message would be
	// ambiguous once two viewers exist, matching the lowercase dv1/dv2 style
	// keymap.go's help text already uses (D57).
	findWasEditing := m.dv1.findEditing()
	findHadMatches := m.dv1.hasFindMatches()
	m.dv1 = m.dv1.setSize(viewerContent.width, viewerContent.height)
	var dv1Drop, dv2Drop string
	switch {
	case findWasEditing && !m.dv1.findEditing():
		dv1Drop = "search dismissed"
	case findHadMatches && !m.dv1.hasFindMatches():
		dv1Drop = "search cleared"
	}
	if m.dv2 != nil {
		dv2Content := dv2Rect.content()
		findWasEditing2 := m.dv2.findEditing()
		findHadMatches2 := m.dv2.hasFindMatches()
		v := (*m.dv2).setSize(dv2Content.width, dv2Content.height)
		m.dv2 = &v
		switch {
		case findWasEditing2 && !m.dv2.findEditing():
			dv2Drop = "search dismissed"
		case findHadMatches2 && !m.dv2.hasFindMatches():
			dv2Drop = "search cleared"
		}
	}
	m = m.combineDropMessages(string(cause), dv1Drop, dv2Drop)
	filesContent := l.files.content()
	m.files = m.files.setSize(filesContent.width, filesContent.height)
	rootsContent := l.roots.content()
	m.roots = m.roots.setSize(rootsContent.width, rootsContent.height)
	// The Picker modal is sized here too, not only from render (M3.2): its
	// Preview pane holds a real viewport.Model whose scroll position must
	// survive across renders (unlike promptView's textinput, sized fresh
	// each View call), so PageUp/PageDown/HalfPageUp/HalfPageDown
	// (handlePickerInput) need a persistently-sized viewport to page
	// against regardless of whether the Picker is currently open.
	pickerPreviewContent := computePickerLayout(width, height).preview.content()
	m.picker = m.picker.setPreviewSize(pickerPreviewContent.width, pickerPreviewContent.height)
	return m
}

// layout is the frame's regions at the model's current terminal size and
// Navigator visibility (S9).
func (m model) layout() layout {
	return computeLayout(m.width, m.height, m.navigatorHidden, m.stale)
}

// visible reports whether a pane has a region to draw in at the current size.
// A terminal too narrow for the Navigator column, or the Navigator toggle
// (S9) hiding it, both hide its panes (see computeLayout), and focus must
// not land in one.
func (m model) visible(p pane) bool {
	l := m.layout()
	switch p {
	case rootsPane:
		return !l.roots.empty()
	case filesPane:
		return !l.files.empty()
	case dv2Pane:
		// DV2 is visible once it exists (m.dv2 non-nil — still never true
		// before a later task creates it) and its half of the split
		// (splitViewer) is non-empty.
		if m.dv2 == nil {
			return false
		}
		_, dv2Rect := l.splitViewer(true)
		return !dv2Rect.empty()
	default: // dv1Pane
		dv1Rect, _ := l.splitViewer(m.dv2 != nil)
		return !dv1Rect.empty()
	}
}

// resolveFocus keeps focus on a visible pane across a resize, falling back
// through focusOrder (Sources Nav, Files Nav, DV1) when want is not visible.
// DV2 is deliberately never a fallback target — see focusOrder's own doc
// comment.
func (m model) resolveFocus(want pane) pane {
	if m.visible(want) {
		return want
	}
	for _, p := range focusOrder {
		if m.visible(p) {
			return p
		}
	}
	return want
}

var (
	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Border)
	focusedBorderStyle = borderStyle.
				BorderForeground(theme.Accent)
	// focusedPaneBorderStyle is the focused pane's border (S11): thick, in
	// the same accent colour as focusedBorderStyle (which the Picker's own
	// prompt box keeps using, unchanged). Every border shape is one cell
	// wide, so thick vs rounded never moves layout.
	focusedPaneBorderStyle = focusedBorderStyle.Border(lipgloss.ThickBorder())
	// focusedTitleStyle is the focused pane's title bar (S11): bold canvas-
	// coloured text on the border's own accent colour, so the two read as one
	// cue.
	focusedTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Bg).
				Background(theme.Accent)
	titleStyle  = lipgloss.NewStyle().Bold(true)
	footerStyle = lipgloss.NewStyle().
			Foreground(theme.Subtle)
	// peekTitleStyle is S7/M7.1's inverted Peek header: the same title row
	// every pane already draws, in reverse video instead of bold, so a
	// Peeked document's heading cannot be mistaken for the Active
	// Document's own (which stays titleStyle).
	peekTitleStyle = titleStyle.Reverse(true)
)

// View returns the frame. Bubble Tea v2 derives terminal state (alt screen,
// cursor, mouse mode, window title) from the returned View on every render.
// AltScreen and MouseMode are set explicitly below (bug fix: noggin used to
// leave both at their zero value, which meant it never took over the
// terminal — nothing was left in the alt screen buffer to restore on exit,
// and mouse wheel events were never captured, so they fell through to the
// terminal's own native scrollback instead of scrolling a pane). Every other
// field (cursor, window title, ...) stays at its zero value, unchanged from
// before.
func (m model) View() tea.View {
	v := tea.NewView(paintCanvas(m.render(), m.width, m.height, theme))
	v.AltScreen = true
	// Also set the terminal's own default colours (OSC 10/11; Bubble Tea
	// restores them on exit). paintCanvas has already coloured every cell, so
	// this only matters where the terminal paints without our cells: the margin
	// around the cell grid, and the instant after a resize. Terminals that
	// ignore the request lose nothing. Skipped under NO_COLOR: Bubble Tea
	// already strips every cell colour then, and recolouring the terminal's
	// defaults behind plain text would defeat the request.
	if os.Getenv("NO_COLOR") == "" {
		v.BackgroundColor = theme.Bg
		v.ForegroundColor = theme.Fg
	}
	// MouseModeCellMotion is enough for click/release/wheel events (which is
	// all handleMouseWheel below uses) without also reporting motion when no
	// button is pressed the way MouseModeAllMotion would.
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

// render is the frame (frame()) plus, when the Help Modal or the Picker is
// open, its overlay on top (overlayHelp / overlayPicker) — the one place
// that composition happens, so View() itself stays a one-liner. The two
// modals are mutually exclusive by construction (Update's handleHelpInput/
// handlePickerInput gating, model.go's Update), so at most one overlay ever
// applies.
func (m model) render() string {
	frame := m.frame()
	switch {
	case m.width <= 0 || m.height <= 0:
		return frame
	case m.helpOpen:
		return overlayHelp(frame, m.width, m.height, m.focus)
	case m.picker.open:
		return overlayPicker(frame, m.width, m.height, m.picker, m.ripgrep, m.hideDotfiles, m.docs)
	default:
		return frame
	}
}

// frame composes the three regions into the string View wraps.
func (m model) frame() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	l := m.layout()

	// Per-frame copies carry the unfocused styling (unfocused_style.go)
	// into the view methods without touching model state.
	roots, files, dv1 := m.roots, m.files, m.dv1
	roots.unfocused = m.focus != rootsPane
	files.unfocused = m.focus != filesPane
	dv1.muted = m.focus != dv1Pane

	dv1Rect, dv2Rect := l.splitViewer(m.dv2 != nil)
	dv1Content := m.composeViewerContent(dv1Pane, dv1Rect.content().width, dv1.view())
	body := m.renderPane(dv1Pane, dv1Rect, dv1Content)
	if m.dv2 != nil {
		dv2 := *m.dv2
		dv2.muted = m.focus != dv2Pane
		dv2Content := m.composeViewerContent(dv2Pane, dv2Rect.content().width, dv2.view())
		body = lipgloss.JoinHorizontal(lipgloss.Top, body, m.renderPane(dv2Pane, dv2Rect, dv2Content))
	}
	if l.navigatorWidth() > 0 {
		navigator := lipgloss.JoinVertical(lipgloss.Left,
			m.renderPane(rootsPane, l.roots, roots.view()),
			m.renderPane(filesPane, l.files, files.view()),
		)
		body = lipgloss.JoinHorizontal(lipgloss.Top, navigator, body)
	}

	if !l.staleRow.empty() {
		body = lipgloss.JoinVertical(lipgloss.Left, body, m.renderStaleRow(l.staleRow))
	}

	if l.footer.empty() {
		return body
	}
	return lipgloss.JoinVertical(lipgloss.Left, body, m.renderFooter(l.footer))
}

// renderPane draws one region at exactly r's outer size, border included: the
// pane's title on the first content row, then content, which is expected to be
// exactly r.content() in size.
//
// lipgloss v2 counts the border inside Width/Height (PR #451), so the region's
// outer size is passed straight through rather than pre-deducting the border as
// v1 required.
func (m model) renderPane(p pane, r rect, content string) string {
	style := borderStyle
	focused := m.focus == p
	if focused {
		style = focusedPaneBorderStyle
	}
	// S7/M7.1: a Peeked document gets the inverted header instead of the
	// title row every other pane (and the Active Document) draws with
	// titleStyle — the one visible way S7 requires Peek to be identifiable
	// as not the Active Document.
	titleRenderStyle := titleStyle
	// C6/M4.2: an in-flight fetch gets the same inverted header Peek does
	// (below) — neither is the Active Document's own title, so neither
	// should read like one.
	inverted := (p == dv1Pane || p == dv2Pane) && (m.viewerFor(p).peeking() || m.viewerFor(p).fetching())
	if inverted {
		titleRenderStyle = peekTitleStyle
	}
	row := titleRow(m.paneTitle(p), p.focusBadge(), r.content().width)
	// S11: the focused pane's title is a full-width accent bar, alongside
	// its thick border. An inverted (Peek/fetch) header keeps its own look
	// so it stays distinguishable from the Active Document's title.
	if focused && !inverted {
		if pad := r.content().width - lipgloss.Width(row); pad > 0 {
			row += strings.Repeat(" ", pad)
		}
		titleRenderStyle = focusedTitleStyle
	}
	title := titleRenderStyle.Render(row)
	return renderBox(style, r, title, content)
}

// titleRow lays out a pane's title row: the title on the left, its focus-key
// badge (S10, M7.1) right-aligned against it, both within the pane's
// interior width — the same content width r.content() already computes for
// the row beneath. The badge is dropped, not the title, when width is too
// tight for both plus the one column of separation between them; a
// too-narrow title is left for the surrounding box's own Width-driven
// truncation to handle, as it always has.
func titleRow(title, badge string, width int) string {
	titleWidth := lipgloss.Width(title)
	badgeWidth := lipgloss.Width(badge)
	if width <= 0 || titleWidth+1+badgeWidth > width {
		return title
	}
	return title + strings.Repeat(" ", width-titleWidth-badgeWidth) + badge
}

// paneTitle is a pane's heading: its name, plus what it is currently showing
// where that is worth saying. While Peeking (S7, M7.1), the Document
// Viewer's title names the Peeked document instead of the Active one —
// renderPane is what styles that differently (inverted, not bold) so the
// two are never visually confusable.
//
// S3/M2.4: a starred Active Document gets the same "★ " glyph the Files Nav
// and Sources Nav rows already use, immediately to the left of the label, so
// a star reads the same way in every pane that shows one. It only applies
// to the Active Document's own title — Peek names a different, non-Active
// document (S7's own concept), which S2/S3 never speak to, so the Peek
// branch above returns before this ever runs. docRecord tolerates any path,
// including one under no session Root, so this needs no isOpen-style guard
// beyond the label check already gating it.
func (m model) paneTitle(p pane) string {
	if p == dv1Pane || p == dv2Pane {
		v := m.viewerFor(p)
		// C6/M4.2: named ahead of Peek since a fetch targeting Peek (peek ==
		// true, not yet reachable from any keybinding — M5.2's own job) has
		// no peekLabel yet to show; a fetch targeting the Active Document
		// (peek == false) hasn't replaced v.label() yet either, for the same
		// reason.
		if v.fetching() {
			return "Fetching: " + v.pendingFetch.title
		}
		if v.peeking() {
			return "Peek: " + v.peekLabel()
		}
		if label := v.label(); label != "" {
			star := ""
			if m.docRecord(v.doc.path).Starred {
				star = "★ "
			}
			return p.title() + ": " + star + label
		}
	}
	return p.title()
}

// renderFooter draws the footer line: an in-progress find query (M5.1)
// first, since it needs the keyboard and so must be shown; the Sources
// Nav's in-progress rename (S8/M7.2) next, for the identical reason —
// unlike find, the rename input itself renders inline on its row
// (rootsnav.go's renderRow), not here, so this only names what Enter/Esc do
// right now, since rootsPane's ordinary hints (Activate "open", Star
// "star" …) no longer apply while every other key types into the row's
// input instead (model.handleRenameInput); otherwise a transient message
// (D22) when one is set; otherwise the open Picker's own hints
// (pickerHintText, M2.3) when it is open, or the focused pane's key hints
// (hintText) otherwise, truncated to fit (S11). Update clears m.message on
// every keypress, so a message shown here never survives past the next key.
//
// The find prompt is not run through footerStyle's dim foreground: it is
// already a complete rendering (textinput.Model.View(), cursor included) and
// muting its own styling would fight it rather than match the other two
// cases, which are plain, unstyled text until footerStyle colours them. The
// rename line is plain, unstyled text like the message/hints cases, so it
// does go through footerStyle, unlike the find prompt.
//
// S10's indicator (m.stale) no longer touches this line at all (C3/M5.1
// moved it to its own dedicated row, renderStaleRow, per S9) — the Footer's
// own priority order (find, then rename, then message, then hints) is
// unaffected by staleness either way.
func (m model) renderFooter(r rect) string {
	if m.focusedViewer().findEditing() {
		return lipgloss.NewStyle().Width(r.width).MaxWidth(r.width).MaxHeight(r.height).
			Render(m.focusedViewer().findView(r.width))
	}
	if m.roots.renameEditing() {
		return footerStyle.Width(r.width).MaxWidth(r.width).MaxHeight(r.height).
			Render(renameFooterText)
	}
	text := hintText(m.focus, r.width)
	// The open Picker (S2, M2.3) owns the keyboard the same exclusive way a
	// find-in-progress does, but unlike Find it does not take over the
	// footer row itself (its own Prompt already renders inside the modal) --
	// it only swaps which hints the footer shows, from the pane underneath's
	// to the Picker's own.
	if m.picker.open {
		text = pickerHintText(r.width)
	}
	if m.message != "" {
		text = m.message
	}
	return footerStyle.Width(r.width).MaxWidth(r.width).MaxHeight(r.height).Render(text)
}

// renderStaleRow draws S9's dedicated staleness row (C3/M5.1): frame() only
// calls this when computeLayout has already reserved a non-empty l.staleRow
// (i.e. m.stale), so this always has something to draw once reached — it
// does not re-check m.stale itself. Styled in reverse video against
// staleRowStyle's highlight color (the Help Modal border's own 214, reused
// rather than adding a second "pay attention" color) rather than plain text,
// per S9. staleMessage already carries the "press R to refresh" resolution
// (staleness.go) — this row only changes where that wording is drawn, not
// the wording itself. rightTruncate (picker_view.go) pre-shrinks the text
// before Width/MaxWidth ever sees it, since Width alone would word-wrap an
// over-length string instead of truncating it (see
// lipgloss-v2-row-rendering-gotchas.md).
func (m model) renderStaleRow(r rect) string {
	text := rightTruncate(staleMessage, r.width)
	return staleRowStyle.Width(r.width).MaxWidth(r.width).MaxHeight(r.height).Render(text)
}
