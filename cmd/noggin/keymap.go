package main

import (
	"charm.land/bubbles/v2/key"
)

// This is the one production file key literals may appear in (D2). Every
// binding in the program is resolved through `keys`; no other production file
// names a key — keymap_test.go carries its own test-only key-encoding table
// (see keyPress), exempt for the same reason its helpers are test-only.
// Retuning the keymap is therefore a one-file edit, and `rg` over the rest of
// the production tree stays free of key strings.

// keyMap holds every binding the program understands. Help text lives with the
// binding so the footer (S11) and help modal (S12/C1) can be generated from
// it rather than duplicating the list.
type keyMap struct {
	Quit key.Binding

	// Star toggles the selected item's starred mark, persisted via the S1/S2
	// document (or Roots) store: the Sources Nav's selected root, the Files
	// Nav's selected file, a focused Document Viewer's Active Document
	// (C4/M2.1, M2.2, M2.8), or the open Picker's selected Result. One chord
	// in every context (D141): ctrl+s is a control chord, so the Picker's
	// always-typing Prompt never sees it as query text and needs no
	// Picker-only star binding of its own.
	Star key.Binding

	// Files Nav folder control (S5): open and close a folder in place.
	// Selection moves with the shared LineUp/LineDown bindings.
	Expand   key.Binding
	Collapse key.Binding

	// Activate opens the selected entry: on the Files Nav, the selected file
	// becomes the active document and focus moves to the viewer (M2.3); a
	// folder is left for later. On the Sources Nav, the selected root becomes
	// the active root (S3), refused with a footer message if its path is
	// missing (broken). It also commits an in-progress find query (M5.1),
	// the viewer's own use of the same physical key: "enter" always means
	// "accept what is selected or typed".
	Activate key.Binding

	// Document Viewer scrolling (S7): line, half-page, page, top, bottom.
	// LineUp/LineDown (and Expand/Collapse above) are the arrow keys alone:
	// D141 retired the h/j/k/l aliases.
	// PageUp/PageDown land on the same offset as HalfPageUp/HalfPageDown in
	// a Document Viewer (M8.1, viewer.go's scrollActions) — kept as separate
	// bindings (distinct chords) rather than merged because the Picker's own
	// Preview reuses these same two bindings for a genuinely full page
	// (picker.go's handlePickerInput), unchanged by M8.1.
	LineUp       key.Binding
	LineDown     key.Binding
	HalfPageUp   key.Binding
	HalfPageDown key.Binding
	PageUp       key.Binding
	PageDown     key.Binding
	Top          key.Binding
	Bottom       key.Binding

	// Edit hands the focused Document Viewer's document to $EDITOR (S10, D3):
	// noggin releases the terminal, resumes on exit and re-renders it from
	// disk. Viewer-only since D141 — matched in Update's dv1Pane/dv2Pane
	// case, which frees the same physical key for Rename on the Sources Nav;
	// on the Files Nav it does nothing. With no document open it reports
	// that instead of doing nothing (M4.1). Refused with a Footer message on
	// a Wikipedia article (S3, C6/M4.3): there is nothing on disk to hand
	// $EDITOR.
	Edit key.Binding

	// RenderMode toggles the targeted viewer between Markdown and Text
	// rendering (D140, textview.go). Global like Edit and Find, and aimed at
	// the same viewer: DV1, or DV2 when DV2 holds Focus. Refused with a Footer
	// message on a Text file, which is always shown as text.
	RenderMode key.Binding

	// Find starts an incremental search over the open document (S8, M5.1):
	// global like Edit, and refused the same way with no document open.
	// FindNext and FindPrevious step through the matches it produces,
	// wrapping; Cancel dismisses a search in progress, leaving the view
	// where it is. While the query is being typed, every other key types
	// into it instead of acting on its usual binding — see
	// model.handleFindInput.
	Find         key.Binding
	FindNext     key.Binding
	FindPrevious key.Binding

	// Navigator hides and restores the whole sidebar (Sources Nav and Files
	// Nav), the Document Viewer re-wrapping to the width it gains (S9,
	// M5.2). Global, like Edit and Find, and it never resets either
	// Navigator pane's own selection — hiding and restoring leaves it
	// exactly as it was.
	Navigator key.Binding

	// FocusRoots and FocusFiles (S6, M2.1) move Focus straight to the Sources
	// Nav or Files Nav from any of the Sources Nav, Files Nav or Document
	// Viewer, revealing the Navigator first if the S9 toggle has it hidden
	// — never hiding it. Global like Navigator/Edit/Find, so both are inert
	// while the Picker owns the keyboard (model.Update's picker.open gate
	// runs before the switch these are matched in, the same way it already
	// shields Navigator/Edit/Find).
	FocusRoots key.Binding
	FocusFiles key.Binding

	// FocusDV1 and FocusDV2 (S5, C3/M1.3, rebound per D55 and again per
	// D141) round out the four direct focus-jump keys FocusRoots/FocusFiles
	// started: `d` moves Focus straight to DV1, always applicable since DV1
	// always exists (S1). `D` moves Focus to DV2, but is a genuine no-op if DV2
	// does not exist (C3/S5: "a bare focus request has no document to
	// create") — unlike FocusRoots/FocusFiles falling back through
	// focusOrder when their target is merely hidden by width, model.Update
	// short-circuits this one before calling focusPane at all, so Focus
	// never moves elsewhere as a side effect. Neither reveals a hidden
	// Navigator the way FocusRoots/FocusFiles do: that behavior exists
	// because those two panes live inside the Navigator column and would
	// otherwise be unreachable while it's hidden, but DV1/DV2 are never in
	// that column, so unhiding the sidebar on a viewer jump would be an
	// unrelated, unrequested side effect (model.focusPane's own doc
	// comment).
	FocusDV1 key.Binding
	FocusDV2 key.Binding

	// FocusNext and FocusPrevious (D141) cycle Focus through the visible
	// panes in the fixed ring Sources Nav -> Files Nav -> DV1 -> DV2,
	// forward and backward (model.cycleFocus). They complement the four
	// direct jumps above rather than replacing them: a pane that is not
	// visible (a hidden or width-dropped Navigator, a DV2 that does not
	// exist) is skipped, so neither ever reveals the Navigator, and with a
	// single visible pane both are no-ops. Global, so they are inert while
	// Find, Rename, the Help Modal or the Picker owns the keyboard.
	FocusNext     key.Binding
	FocusPrevious key.Binding

	// CloseViewer (D141, replacing S4's ctrl+x CloseDV2) closes the focused
	// Document Viewer: from DV2 it closes DV2 and Focus moves to DV1; from
	// DV1 with DV2 open, DV1 takes over DV2's whole state and DV2 closes;
	// from DV1 alone it clears DV1 back to empty. A no-op from the Sources
	// Nav or Files Nav (model.closeViewer). Matched in model.Update's global
	// switch, so it is never reached while Find, Rename, the Help Modal or
	// the Picker owns the keyboard.
	CloseViewer key.Binding

	// Refresh re-reads the Active Root's file tree and the Active Document
	// from disk on demand (S2, C2a/M5.1): global, like Edit/Find/Navigator,
	// reachable from every pane. model.refresh preserves Files Nav
	// expansion/selection (nearest surviving ancestor if the selected node
	// is gone, filesNav.refreshRoot) and the viewer's scroll position
	// (clamped to the end if the document shrank, viewer.refreshDoc), drops
	// a committed find's matches the same way any other re-render already
	// does (an in-progress query can't reach this binding at all —
	// handleFindInput owns every key but Cancel/Activate while typing), and
	// recomputes S8's git markers in the same pass. An open Wikipedia
	// article (S3, C6/M4.3) re-fetches instead of re-reading from disk,
	// bypassing the session cache — the one case where Refresh reaches the
	// network (D119) — and never trips the staleness indicator (nothing
	// about an article is ever tracked by it, staleness.go's own note).
	Refresh key.Binding

	// Cancel dismisses a modal or in-progress input without acting on it.
	// M5.1's first use was dismissing a find query; a generic name and
	// binding on purpose, since S12/C1's help modal (M5.3) uses the same
	// one-key dismissal.
	Cancel key.Binding

	// Help is the footer's help affordance (S11): the footer advertised it
	// under truncation (M4.2) before S12/C1's modal existed to open on it —
	// M5.3 wired the modal onto this binding and its hint text, not a new
	// contract of its own.
	Help key.Binding

	// Picker opens the C2 Picker modal (S1): global, like Edit/Find/
	// Navigator, and reachable from every Focus including while the
	// Navigator is hidden. It always opens the same way — File Picker
	// mode, Active Source Scope, empty Prompt (model.openPicker) — and
	// Cancel is its one-key dismissal, the same generic binding Find and
	// the Help Modal already share.
	Picker key.Binding

	// PickerRestore (ctrl+space) opens the Picker to the last non-empty
	// session's query, mode and Scope (model.restorePicker), or just opens
	// it normally when there is none. Global like Picker. Both names are
	// bound because the terminal sends the same NUL byte for ctrl+space and
	// ctrl+@, and the decoder may report either.
	PickerRestore key.Binding

	// PickerUp and PickerDown move the open Picker's Results selection
	// (S2, M2.3). They bind the same arrow keys as LineUp/LineDown but stay
	// separate bindings: S2's "every printable keystroke goes to the Prompt"
	// means the Picker may only ever move on non-printable keys, so keeping
	// its movement apart means a printable alias added to LineUp/LineDown
	// can never leak into the Picker (the h/j/k/l aliases D141 retired were
	// exactly that case).
	PickerUp   key.Binding
	PickerDown key.Binding

	// PickerMode and PickerScope switch the open Picker's mode (File →
	// Grep → Wikipedia, S7; a three-way cycle since S2/C6/M3.2, was File ↔
	// Grep) and Scope (its first value labelled Active Root for Grep,
	// Active Source for File Picker, C6/M1.3; then All Roots, then Starred,
	// S7/S8; Wikipedia mode has none — the key is a no-op there, S2) in
	// place, without touching the Prompt (M2.3 reserves them as non-typing;
	// M5.1/M5.2 give them their actions, C4/M5.1 adds Scope's third value).
	// alt+t/alt+a (D141, were Tab/Shift+Tab) rather than a bare letter,
	// for the same reason PickerUp/PickerDown avoid printable keys: an alt
	// chord carries no text, so typing a query is never at odds with
	// switching. The letters mirror the Files Nav's own FilesMode/FilesScope
	// (t/a), which can be bare because the Files Nav is never a typing
	// context.
	PickerMode  key.Binding
	PickerScope key.Binding

	// ToggleHidden is S4/M6.2's session-scoped dotfile toggle: hides and
	// shows dotfiles (and files under a dot-directory) in both Picker
	// modes' Results, reset only on relaunch (model.hideDotfiles, not
	// picker-owned — see its own doc comment for why). alt+h rather than
	// a bare letter for the same non-printable reason PickerMode/
	// PickerScope avoid one: S2's "always-typing" Prompt (handlePickerInput)
	// must not treat this as query text.
	ToggleHidden key.Binding

	// ActivateDV2 and PeekDV2 are C3/S3's `ctrl+` targeting modifier on
	// the Files Nav (C3/M1.4, rebound C3/M1.4a per D55): the same load/peek
	// the unmodified Activate/Peek perform, routed to DV2 instead of DV1
	// and creating DV2 first if it does not exist. Both are Files-Nav-only
	// for now, matched in Update's filesPane case beside the keys they
	// modify (M1.5 gives ActivateDV2 its second home, the Picker's
	// Activate). Focus behavior mirrors each unmodified key exactly: a
	// load moves Focus to the viewer it loaded into, a peek never moves
	// Focus.
	//
	// ActivateDV2 carries two keys, not one (D53, rebound by D55):
	// `ctrl+enter` is the primary chord, but terminals without
	// Kitty-protocol `CSI u` or xterm's modifyOtherKeys send it
	// byte-identical to plain `enter`, with no runtime signal noggin could
	// detect and degrade on, so `ctrl+l` is bound as a fallback — D55
	// chose a second control chord rather than D53's original printable
	// `L`, since a raw control byte can never be typed into the Picker's
	// always-typing Prompt as query text (TRIAGE T3, closed structurally by
	// this without a Picker-specific carve-out — M1.5's own comment).
	// PeekDV2 needs no fallback of its own: a plain Ctrl+letter is a raw C0 control byte every terminal delivers
	// identically regardless of Kitty/CSI-u/modifyOtherKeys support, so
	// there is nothing to degrade (D55).
	ActivateDV2 key.Binding
	PeekDV2     key.Binding

	// Peek (S7, M7.1) shows the Files Nav's selected document in the
	// Document Viewer without making it the Active Document and without
	// moving Focus off the Files Nav — Files-Nav-only, unlike every other
	// binding above it, so it is matched in Update's filesPane case, not
	// the global switch. model.togglePeek is both "start" and "end": a
	// second p while already Peeking clears it (S7's own "a second p"
	// clearing rule), and it is a no-op on the Active Document (S7's own
	// carve-out) or a directory. Every other clearing trigger S7 lists
	// (selection change, Focus change, opening the Picker, a refresh) is
	// enforced at its own call site — see model.setFocus and model.refresh.
	Peek key.Binding

	// FilesMode and FilesScope are S5's two Browse cycle keys (C4/M4.1):
	// Files-Nav-only, like Peek above, so they are matched in Update's
	// filesPane case (via filesNav.handleKey, beside Expand/Collapse) rather
	// than the global switch, and do nothing from any other Focus. `t` and
	// `a` (D141, were Tab/Shift+Tab): bare letters are safe here because the
	// Files Nav is never a typing context; the Picker's own PickerMode/
	// PickerScope are the alt+ versions of the same letters. FilesMode cycles Tree -> Recent -> Starred -> Tree ...
	// (filesNav.cycleMode); FilesScope cycles the current mode's own Browse
	// Scope values (filesNav.cycleScope), a no-op while the mode is Tree,
	// which has none (S5).
	FilesMode  key.Binding
	FilesScope key.Binding

	// FileFilter toggles the Active Root's File Filter between Markdown and
	// All (D140, tree.Filter), from the Files Nav (matched in Update's
	// filesPane case) and from the open Picker (handlePickerInput) alike:
	// one per-Root state, one chord. alt+m rather than a bare letter so the
	// Picker's always-typing Prompt never sees it as text. In the Files Nav
	// it bounds the Tree only; with the Wikipedia Source active it does
	// nothing but say so in the Footer.
	FileFilter key.Binding

	// HistoryBack and HistoryForward walk a focused Document Viewer's own S9
	// back/forward stack (C4/M6.1): matched in Update's dv1Pane/dv2Pane case
	// beside keys.Star, so — like it — they act on whichever viewer holds
	// Focus (DV1 and DV2 each walk their own history independently) and do
	// nothing from the Sources Nav or Files Nav. `[`/`]`, unused elsewhere in
	// the keymap and free of S2's Picker-Prompt "always typing" constraint
	// PickerUp/PickerDown and ToggleHidden each had to route
	// around, since neither the Files Nav nor a focused viewer is ever a
	// typing context the way the Picker's Prompt is.
	HistoryBack    key.Binding
	HistoryForward key.Binding

	// Rename (S8/M7.2) edits the Sources Nav's selected Root's Name inline on
	// its row: Enter confirms, Cancel (esc) discards, an empty name reverts
	// to the default (the folder's base name) — model.handleRenameInput owns
	// every other key while it is open, the identical exclusivity
	// handleFindInput already gives an in-progress find query. Refused with a
	// Footer message for the Wikipedia row and for an unstarred ad-hoc Root
	// (S8), matched in Update's rootsPane case beside Activate/Star. "e"
	// (D141, was "c"): "edit this Root's name", the same physical key that
	// edits a document in a focused Document Viewer — Edit is viewer-only, so
	// the two never meet.
	Rename key.Binding
}

// keys is the program's keymap.
var keys = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Star: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "star"),
	),
	Expand: key.NewBinding(
		key.WithKeys("right"),
		key.WithHelp("→", "expand"),
	),
	Collapse: key.NewBinding(
		key.WithKeys("left"),
		key.WithHelp("←", "collapse"),
	),
	Activate: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "open"),
	),
	LineUp: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "up"),
	),
	LineDown: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "down"),
	),
	HalfPageUp: key.NewBinding(
		key.WithKeys("ctrl+u"),
		key.WithHelp("ctrl+u", "½ page up"),
	),
	HalfPageDown: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "½ page down"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("pgup", "ctrl+b"),
		key.WithHelp("pgup", "page up"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("pgdown", "ctrl+f"),
		key.WithHelp("pgdn", "page down"),
	),
	Top: key.NewBinding(
		key.WithKeys("home", "g"),
		key.WithHelp("g", "top"),
	),
	Bottom: key.NewBinding(
		key.WithKeys("end", "G"),
		key.WithHelp("G", "bottom"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit"),
	),
	RenderMode: key.NewBinding(
		key.WithKeys("m"),
		key.WithHelp("m", "markdown/text"),
	),
	Find: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "find"),
	),
	FindNext: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "next match"),
	),
	FindPrevious: key.NewBinding(
		key.WithKeys("N"),
		key.WithHelp("N", "prev match"),
	),
	Navigator: key.NewBinding(
		key.WithKeys("F"),
		key.WithHelp("F", "sidebar"),
	),
	FocusRoots: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "roots"),
	),
	FocusFiles: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "files"),
	),
	FocusDV1: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "dv1"),
	),
	FocusDV2: key.NewBinding(
		key.WithKeys("D"),
		key.WithHelp("D", "dv2"),
	),
	FocusNext: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next pane"),
	),
	FocusPrevious: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev pane"),
	),
	CloseViewer: key.NewBinding(
		key.WithKeys("ctrl+w"),
		key.WithHelp("ctrl+w", "close viewer"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("R"),
		key.WithHelp("R", "refresh"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Picker: key.NewBinding(
		key.WithKeys("space"),
		key.WithHelp("space", "picker"),
	),
	PickerRestore: key.NewBinding(
		key.WithKeys("ctrl+space", "ctrl+@"),
		key.WithHelp("ctrl+space", "reopen picker"),
	),
	PickerUp: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "up"),
	),
	PickerDown: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "down"),
	),
	PickerMode: key.NewBinding(
		key.WithKeys("alt+t"),
		key.WithHelp("alt+t", "mode"),
	),
	PickerScope: key.NewBinding(
		key.WithKeys("alt+a"),
		key.WithHelp("alt+a", "scope"),
	),
	ToggleHidden: key.NewBinding(
		key.WithKeys("alt+h"),
		key.WithHelp("alt+h", "hidden files"),
	),
	Peek: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "peek"),
	),
	FilesMode: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "mode"),
	),
	FilesScope: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "scope"),
	),
	FileFilter: key.NewBinding(
		key.WithKeys("alt+m"),
		key.WithHelp("alt+m", "file filter"),
	),
	ActivateDV2: key.NewBinding(
		key.WithKeys("ctrl+enter", "ctrl+l"),
		key.WithHelp("ctrl+enter/ctrl+l", "open in dv2"),
	),
	PeekDV2: key.NewBinding(
		key.WithKeys("ctrl+p"),
		key.WithHelp("ctrl+p", "peek in dv2"),
	),
	HistoryBack: key.NewBinding(
		key.WithKeys("["),
		key.WithHelp("[", "back"),
	),
	HistoryForward: key.NewBinding(
		key.WithKeys("]"),
		key.WithHelp("]", "forward"),
	),
	Rename: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "rename"),
	),
}
