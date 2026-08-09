package main

import (
	"fmt"
	"os"
	"os/exec"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"noggin/internal/tree"
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
// short display label, not the GLOSSARY.md term ("Roots Nav", "Files Nav",
// "Document Viewer") — those name the pane in prose, but a title lives on a
// single content row shared with the Active Document's own label
// (paneTitle), so it stays terse on purpose rather than drifting to match the
// glossary.
func (p pane) title() string {
	switch p {
	case rootsPane:
		return "Roots"
	case filesPane:
		return "Files"
	default: // dv1Pane, dv2Pane
		return "Document"
	}
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
	// (Roots Nav and Files Nav) and gives the Document Viewer the full width
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

	// roots is the session's note sources (S1): the Roots Nav pane and its
	// selection, doubling as the record of which root is active. message
	// carries a transient footer notice — a launch-time rejection (S1) or a
	// refused action (e.g. activating a broken root, S3); M4.2 will add
	// clear-on-next-keypress and richer hints, but showing it at all cannot
	// wait for that milestone (see D22).
	roots   rootsNav
	message string

	// helpOpen is the S12 Help Modal (M5.3): when set, it owns the keyboard
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

	// staleBaseline/staleGen/stale are S10's periodic staleness check
	// (C2a/M10.1, staleness.go): the "known good" snapshot the disk is
	// compared against, the generation that invalidates a check dispatched
	// against a since-superseded baseline, and whether the last check found
	// a mismatch. Never mutated by anything but resetStaleBaseline and
	// runStaleCheck/applyStaleResult — see their own doc comments.
	staleBaseline staleBaseline
	staleGen      int
	stale         bool
}

func newModel() model {
	return model{focus: filesPane, files: newFilesNav(), dv1: newViewer(), roots: newRootsNav(), picker: newPicker()}
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

// withStorePath sets the resolved starred-roots config path (M3.2's
// storePath()) that a Roots Nav star toggle persists through. main.go
// resolves it once at startup so a keypress never re-derives
// os.UserConfigDir() itself.
func (m model) withStorePath(path string) model {
	m.roots.storePath = path
	return m
}

// withRipgrep installs the S14 ripgrep detection result (detectRipgrep())
// that main.go resolves once at startup, the same one-shot-then-carry
// pattern as withStorePath.
func (m model) withRipgrep(avail ripgrepAvailability) model {
	m.ripgrep = avail
	return m
}

// applyResolution installs an S1 launch resolution: the session's roots, which
// one is active, an optional document, and any rejection message. Walking the
// active root or reading the document can fail; that becomes a message, never
// an exit — the session always starts (S1's "silence is not an error").
func (m model) applyResolution(res resolution) model {
	m.roots = m.roots.setRoots(res.roots, res.active)
	m.message = res.message

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
	m.files = m.files.setRoot(root).withStatus(gitStatus(root.Path))
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
	return m.openDocumentInto(dv1Pane, d)
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
func (m model) closeDV2() model {
	if m.dv2 == nil {
		return m
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
func (m model) activate() model {
	return m.activateInto(dv1Pane)
}

// activateInto is activate's targeted form (C3/S3, M1.4): the Files Nav's
// selected file loaded into target rather than always DV1 — dv2Pane when the
// ActivateDV2 modifier (ctrl+enter / ctrl+l, D55) was used, dv1Pane for the
// plain key.
// Every refusal above (folder, empty pane, read error) applies identically
// whichever viewer is targeted, and is checked before openDocumentInto runs,
// so a refused load never leaves a newly created, empty DV2 behind.
func (m model) activateInto(target pane) model {
	n := m.files.selectedNode()
	if n == nil || n.IsDir {
		return m
	}
	d, err := loadDocument(n.Path)
	if err != nil {
		m.message = fmt.Sprintf("cannot open %s: %v", n.Path, err)
		return m
	}
	return m.openDocumentInto(target, d)
}

// activateRoot makes the Roots Nav's selected entry the active root,
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

// toggleStar flips the Roots Nav selection's starred mark and persists it via
// the S2 store (M3.2's toggleStore), using the storePath resolved once at
// startup (withStorePath). A config directory that could not be resolved
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
		m.message = "cannot save starred roots: no resolvable config directory"
		return m
	}
	if _, err := toggleStore(m.roots.storePath, row.root.path); err != nil {
		m.message = fmt.Sprintf("cannot save starred roots: %v", err)
		return m
	}
	m.roots = m.roots.toggleStarred(m.roots.selected)
	return m
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
func (m model) togglePeek() model {
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
func (m model) togglePeekInto(target pane) model {
	if target != dv2Pane {
		target = dv1Pane
	}
	// viewerFor falls back to DV1 while dv2 is nil, so every read of the
	// target viewer below is gated on it actually existing.
	exists := target == dv1Pane || m.dv2 != nil
	if exists && m.viewerFor(target).peeking() {
		return m.clearPeekIn(target)
	}
	n := m.files.selectedNode()
	if n == nil || n.IsDir {
		return m
	}
	if exists {
		if v := m.viewerFor(target); v.doc.isOpen() && n.Path == v.doc.path {
			return m
		}
	}
	d, err := loadDocument(n.Path)
	if err != nil {
		m.message = fmt.Sprintf("cannot open %s: %v", n.Path, err)
		return m
	}
	if target == dv2Pane {
		m = m.ensureDV2()
	}
	return m.withViewerFor(target, m.viewerFor(target).startPeek(d))
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
func (m model) refresh() model {
	// S7/M7.1: a refresh is one of Peek's clearing triggers.
	m = m.clearPeek()
	if m.files.root != nil {
		if node, err := tree.Walk(m.files.root.Path); err != nil {
			m.message = fmt.Sprintf("cannot read %s: %v", m.files.root.Path, err)
		} else {
			m.files = m.files.refreshRoot(node).withStatus(gitStatus(node.Path))
		}
	}
	// C3/M3.1: not focus-scoped — a refresh acts on both viewers whichever
	// exist, "regardless of which pane holds Focus" (S7), unlike Edit/Find/
	// scroll's "whichever viewer has Focus" routing. dv1Drop/dv2Drop mirror
	// resize's own pair (this doc comment's third paragraph) so the two
	// call sites combine two simultaneous find-drops into m.message the
	// same way.
	var dv1Drop, dv2Drop string
	m, dv1Drop = m.refreshViewer(dv1Pane)
	if m.dv2 != nil {
		m, dv2Drop = m.refreshViewer(dv2Pane)
	}
	switch {
	case dv1Drop != "" && dv2Drop != "":
		m.message = fmt.Sprintf("refreshed; dv1 %s, dv2 %s", dv1Drop, dv2Drop)
	case dv1Drop != "":
		m.message = fmt.Sprintf("refreshed; %s", dv1Drop)
	case dv2Drop != "":
		m.message = fmt.Sprintf("refreshed; dv2 %s", dv2Drop)
	}
	// S10/C2a/M10.1: a refresh is the user acting on the indicator ("press
	// R"), so it both clears it and re-baselines against whatever the model
	// ends up showing above — including a read failure's fallback (an
	// emptied viewer, an untouched tree) — since that is the state now
	// actually on screen, known-good by construction.
	m = m.resetStaleBaseline()
	return m
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
func (m model) refreshViewer(p pane) (model, string) {
	v := m.viewerFor(p)
	if !v.doc.isOpen() {
		return m, ""
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
		return m, "search cleared"
	}
	return m, ""
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

// openHelp opens the S12 Help Modal for the focused pane's context. Unlike
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
	// S10/C2a/M10.1's own periodic pair, same shape as the Picker's above:
	// a matured timer dispatches the actual (cheap) disk check, which
	// reports back its own message so the check itself never blocks Update.
	case staleCheckMsg:
		return m.runStaleCheck(msg)
	case staleResultMsg:
		return m.applyStaleResult(msg)
	case tea.WindowSizeMsg:
		return m.resize(msg.Width, msg.Height, resizeCauseWindow), nil
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
		// specific case, so it is reachable from every one of Roots Nav/Files
		// Nav/DV1/DV2 — and, like every other binding in this switch, already
		// unreachable while a find is editing, the Help Modal is open or the
		// Picker is open, per the three gates above that return before this
		// switch ever runs.
		case key.Matches(msg, keys.CloseDV2):
			return m.closeDV2(), nil
		case key.Matches(msg, keys.Refresh):
			return m.refresh(), nil
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
		case filesPane:
			if key.Matches(msg, keys.Activate) {
				return m.activate(), nil
			}
			// Peek (S7, M7.1) is Files-Nav-only, unlike every other binding
			// matched above in the global switch — that is what keeps it
			// from firing on the Roots Nav or Document Viewer.
			if key.Matches(msg, keys.Peek) {
				return m.togglePeek(), nil
			}
			// C3/S3's targeting modifiers (M1.4), matched here beside the
			// unmodified keys they modify: same actions, aimed at DV2 and
			// creating it if absent. Neither can be confused with its plain
			// twin — key.Matches compares the whole keystroke string, so
			// "ctrl+enter"/"ctrl+l" never match Activate's "enter" and
			// "ctrl+p" never matches Peek's "p" (D55).
			if key.Matches(msg, keys.ActivateDV2) {
				return m.activateInto(dv2Pane), nil
			}
			if key.Matches(msg, keys.PeekDV2) {
				return m.togglePeekInto(dv2Pane), nil
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
			if key.Matches(msg, keys.Activate) {
				return m.activateRoot(), nil
			}
			if key.Matches(msg, keys.Star) {
				return m.toggleStar(), nil
			}
			if next, handled := m.roots.handleKey(msg); handled {
				m.roots = next
				return m, nil
			}
		}
	}
	return m, nil
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
	switch {
	case dv1Drop != "" && dv2Drop != "":
		m.message = fmt.Sprintf("%s; dv1 %s, dv2 %s", cause, dv1Drop, dv2Drop)
	case dv1Drop != "":
		m.message = fmt.Sprintf("%s; %s", cause, dv1Drop)
	case dv2Drop != "":
		m.message = fmt.Sprintf("%s; dv2 %s", cause, dv2Drop)
	}
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
// through focusOrder (Roots Nav, Files Nav, DV1) when want is not visible.
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
			BorderForeground(lipgloss.Color("240"))
	focusedBorderStyle = borderStyle.
				BorderForeground(lipgloss.Color("205"))
	titleStyle  = lipgloss.NewStyle().Bold(true)
	footerStyle = lipgloss.NewStyle().
			Foreground(dimColor)
	// peekTitleStyle is S7/M7.1's inverted Peek header: the same title row
	// every pane already draws, in reverse video instead of bold, so a
	// Peeked document's heading cannot be mistaken for the Active
	// Document's own (which stays titleStyle).
	peekTitleStyle = titleStyle.Reverse(true)
)

// View returns the frame. Bubble Tea v2 derives terminal state (alt screen,
// cursor, mouse mode, window title) from the returned View on every render, so
// every field left at its zero value is a deliberate "off": noggin draws inline
// with no cursor, exactly as it did under v1.
func (m model) View() tea.View {
	return tea.NewView(m.render())
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
		return overlayPicker(frame, m.width, m.height, m.picker, m.ripgrep, m.hideDotfiles)
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

	dv1Rect, dv2Rect := l.splitViewer(m.dv2 != nil)
	body := m.renderPane(dv1Pane, dv1Rect, m.dv1.view())
	if m.dv2 != nil {
		body = lipgloss.JoinHorizontal(lipgloss.Top, body, m.renderPane(dv2Pane, dv2Rect, m.dv2.view()))
	}
	if l.navigatorWidth() > 0 {
		navigator := lipgloss.JoinVertical(lipgloss.Left,
			m.renderPane(rootsPane, l.roots, m.roots.view()),
			m.renderPane(filesPane, l.files, m.files.view()),
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
	if m.focus == p {
		style = focusedBorderStyle
	}
	// S7/M7.1: a Peeked document gets the inverted header instead of the
	// title row every other pane (and the Active Document) draws with
	// titleStyle — the one visible way S7 requires Peek to be identifiable
	// as not the Active Document.
	titleRenderStyle := titleStyle
	if (p == dv1Pane || p == dv2Pane) && m.viewerFor(p).peeking() {
		titleRenderStyle = peekTitleStyle
	}
	return renderBox(style, r, titleRenderStyle.Render(m.paneTitle(p)), content)
}

// paneTitle is a pane's heading: its name, plus what it is currently showing
// where that is worth saying. While Peeking (S7, M7.1), the Document
// Viewer's title names the Peeked document instead of the Active one —
// renderPane is what styles that differently (inverted, not bold) so the
// two are never visually confusable.
func (m model) paneTitle(p pane) string {
	if p == dv1Pane || p == dv2Pane {
		v := m.viewerFor(p)
		if v.peeking() {
			return "Peek: " + v.peekLabel()
		}
		if label := v.label(); label != "" {
			return p.title() + ": " + label
		}
	}
	return p.title()
}

// renderFooter draws the footer line: an in-progress find query (M5.1)
// first, since it needs the keyboard and so must be shown; otherwise a
// transient message (D22) when one is set; otherwise the open Picker's own
// hints (pickerHintText, M2.3) when it is open, or the focused pane's key
// hints (hintText) otherwise, truncated to fit (S11). Update clears
// m.message on every keypress, so a message shown here never survives past
// the next key.
//
// The find prompt is not run through footerStyle's dim foreground: it is
// already a complete rendering (textinput.Model.View(), cursor included) and
// muting its own styling would fight it rather than match the other two
// cases, which are plain, unstyled text until footerStyle colours them.
//
// S10's indicator (m.stale) no longer touches this line at all (C3/M5.1
// moved it to its own dedicated row, renderStaleRow, per S9) — the Footer's
// own priority order (find, then message, then hints) is unaffected by
// staleness either way.
func (m model) renderFooter(r rect) string {
	if m.focusedViewer().findEditing() {
		return lipgloss.NewStyle().Width(r.width).MaxWidth(r.width).MaxHeight(r.height).
			Render(m.focusedViewer().findView(r.width))
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
