package main

import (
	"path/filepath"

	"charm.land/lipgloss/v2"
)

// wikipediaAttribution is S9a's own fixed wording for a Wikipedia article's
// location line — the three obligations ADR-0007 identifies (attribution via
// the article's own canonical URL, a licensing notice naming the license, and
// an indication the content was modified from the original) as one string,
// the URL spliced in between the other two segments at call time
// (documentLocation, below). Never shown for a file — S9's own file/orphan
// forms replace it entirely, they never combine.
const (
	wikipediaAttributionPrefix = "CC BY-SA 4.0 · "
	wikipediaAttributionSuffix = " · Reformatted to markdown"
)

// documentLocation renders S9/S9a's location line text for doc: an article
// (doc.isArticle(), C6/M4.1) gets S9a's fixed three-segment CC BY-SA
// attribution around its own canonical URL (doc.path); a file
// under a covering session Root (rootForPath, roots.go) gets that Root's
// own label (root.displayName, S8/M7.3) and its path relative to the Root
// — the same filepath.Rel-with-fallback pairing pickerFileRow
// (picker_view.go) already establishes for an identical "Root + rel path"
// row; a file under no Root (an orphan, S1) gets its own full path, the
// one case with no Root label to pair anything against.
//
// Takes roots as a parameter rather than reading them off any stored
// field, the same shape buildBrowseRow (browse_list.go) and pickerRootLabel
// (picker_view.go) already give their own identical "which Root covers
// this path" computation — S9's location line is a fourth reader of
// rootForPath/displayName, not a reason to cache a copy anywhere new.
//
// The render mode the line ends with is not part of this: see
// viewerLocation.
func documentLocation(doc document, roots []root) string {
	if doc.isArticle() {
		return wikipediaAttributionPrefix + doc.path + wikipediaAttributionSuffix
	}
	if r, ok := rootForPath(roots, doc.path); ok {
		rel, err := filepath.Rel(r.path, doc.path)
		if err != nil {
			// Cannot happen in practice — rootForPath already establishes
			// doc.path sits strictly inside r.path — but pickerFileRow's own
			// identical fallback (picker_view.go) is precedent for not
			// letting a broken invariant silently drop the row instead of
			// degrading to the absolute path.
			rel = doc.path
		}
		return r.displayName() + " · " + rel
	}
	return doc.path
}

// docLocationVisible reports whether the Active Document currently has a
// location line to show: view()'s own "real content" case (err == nil and a
// document is open) — nothing to locate on a render error or with nothing
// open, matching view()'s own placeholder branches for both — AND
// v.boxHeight >= 1 (TRIAGE T26/PLAN M8.2a): a pathologically short terminal
// can leave the pane no row to spare for one at all, and the location line
// must never claim a row it cannot be given, the same "never overflow the
// frame it's given" rule computeLayout's own staleRow already follows.
// Also what contentHeight (below) reads to decide whether vp reserves the
// line's one row — the same boolean drives both, so the row is never
// reserved without a line drawn in it or vice versa.
func (v viewer) docLocationVisible() bool {
	return v.boxHeight >= 1 && v.err == nil && v.doc.isOpen()
}

// peekLocationVisible is docLocationVisible's Peek counterpart (S9's "while
// the Peek lasts"): independent of the Active Document's own state, since a
// Peek can be open (or erroring) while the Active Document underneath it is
// in any state of its own — see boxHeight's own doc comment (viewer.go) for
// why the two viewports must derive their height from this independently
// rather than one copying the other's. Carries the same boxHeight >= 1
// floor as docLocationVisible, for the same reason.
func (v viewer) peekLocationVisible() bool {
	return v.boxHeight >= 1 && v.peek.err == nil && v.peek.isOpen()
}

// contentHeight is docContentHeight's and peekContentHeight's shared
// implementation (TRIAGE T28/PLAN M8.2a folded the two near-identical
// functions into this one): boxHeight less the S9/S9a location line's one
// row whenever locationVisible holds, boxHeight unchanged otherwise (S9's
// "absent with no document", and the same treatment for a render error,
// extends to the row it would have cost). The max floor guards the
// arithmetic itself; locationVisible's own boxHeight >= 1 gate is what
// actually keeps this from ever being asked to reserve a row that doesn't
// exist in the first place.
func (v viewer) contentHeight(locationVisible bool) int {
	if locationVisible {
		return max(v.boxHeight-1, 0)
	}
	return max(v.boxHeight, 0)
}

// docContentHeight is vp's own operative height (S9/M8.2). Read at the end
// of render() (once v.err has actually settled) and inside setSize (a
// resize, which cannot itself change v.err/v.doc, so it always agrees with
// whatever render last decided).
func (v viewer) docContentHeight() int {
	return v.contentHeight(v.docLocationVisible())
}

// peekContentHeight is docContentHeight's Peek counterpart, driven by
// peekLocationVisible instead.
func (v viewer) peekContentHeight() int {
	return v.contentHeight(v.peekLocationVisible())
}

// locationDoc is the document a Document Viewer pane's own S9/S9a location
// line describes right now, mirroring view()'s own precedence exactly (S9's
// wording is literally "reflects a Peeked document while the Peek lasts"):
// nothing while a fetch is in flight (view()'s own placeholder, nothing
// real to locate yet), the Peeked document while Peek is open and not
// erroring, otherwise the Active Document when a real render is showing.
// The second return is false exactly when view() itself would show a
// placeholder instead of real content — model.viewerLocationLine's own
// signal to draw no line at all.
func (v viewer) locationDoc() (document, bool) {
	switch {
	case v.fetching():
		return document{}, false
	case v.peeking():
		if !v.peekLocationVisible() {
			return document{}, false
		}
		return v.peek.doc, true
	default:
		if !v.docLocationVisible() {
			return document{}, false
		}
		return v.doc, true
	}
}

// viewerLocationLine composes pane p's own S9/S9a location line text, or ""
// when locationDoc says there is none to show right now (model.
// composeViewerContent's own cue to add no row at all) — the one place
// documentLocation is ever called with the session's live Root list
// (m.roots.rootList()), computed fresh at render time rather than cached on
// viewer itself (documentLocation's own doc comment gives the same
// reasoning).
func (m model) viewerLocationLine(p pane) string {
	place, suffix, _ := m.viewerLocation(p)
	return place + suffix
}

// viewerLocation is viewerLocationLine in its two parts: the document's place
// (documentLocation) and the closing render mode segment, `· Markdown` or
// `· Text`, shown in the mode the document is actually displayed in.
// ok is false when there is no line to draw.
func (m model) viewerLocation(p pane) (place, suffix string, ok bool) {
	v := m.viewerFor(p)
	doc, ok := v.locationDoc()
	if !ok {
		return "", "", false
	}
	return documentLocation(doc, m.roots.rootList()), " · " + v.modeFor(doc).String(), true
}

// composeViewerContent appends pane p's own S9/S9a location line beneath
// its already-rendered content (viewer.view(), unchanged by this — it stays
// exactly the width x boxHeight-or-less box M8.1 and earlier tests already
// exercise), styled de-emphasized (locationLineStyle, style.go) and
// right-truncated to width so a long path never wraps the row. content is
// returned unchanged when there is no line to add (viewerLocationLine's own
// "" case) — the row viewer.view()'s own docContentHeight/peekContentHeight
// already left unreserved in that case, so nothing needs adding here
// either; the two decisions (reserve the row, draw the row) are driven by
// the same docLocationVisible/peekLocationVisible booleans, so they can
// never disagree.
func (m model) composeViewerContent(p pane, width int, content string) string {
	place, suffix, ok := m.viewerLocation(p)
	if !ok {
		return content
	}
	// Truncate the place, not the render mode, which the line always ends
	// with.
	text := rightTruncate(place+suffix, width)
	if room := width - lipgloss.Width(suffix); room >= 2 {
		text = rightTruncate(place, room) + suffix
	}
	row := locationLineStyle().Width(width).MaxWidth(width).Render(text)
	return content + "\n" + row
}
