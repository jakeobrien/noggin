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
	// Nav's selected file, or a focused Document Viewer's Active Document
	// (C4/M2.1, M2.2, M2.8). Sources Nav selection moves with the shared
	// LineUp/LineDown bindings; Activate (below) also opens the selected root.
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

	// Edit hands the active document to $EDITOR (S10, D3): noggin releases the
	// terminal, resumes on exit and re-renders it from disk. Global, not
	// pane-specific — with no document open it reports that instead of doing
	// nothing (M4.1). Refused with a Footer message on a Wikipedia article
	// (S3, C6/M4.3): there is nothing on disk to hand $EDITOR.
	Edit key.Binding

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

	// FocusDV1 and FocusDV2 (S5, C3/M1.3, rebound C3/M1.3a per D55) round out
	// the four direct focus-jump keys FocusRoots/FocusFiles started: `o`
	// moves Focus straight to DV1, always applicable since DV1 always exists
	// (S1). `ctrl+o` moves Focus to DV2, but is a genuine no-op if DV2
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

	// CloseDV2 (S4, C3/M2.1) closes DV2: `ctrl+x`, global like FocusDV1/
	// FocusDV2 — reachable from any pane's Focus, including DV1/DV2
	// themselves — but, unlike them, never matched while a modal (Picker,
	// Help Modal) owns the keyboard (S4's own carve-out; enforced the same
	// way Navigator/Edit/Find already are, by living in model.Update's
	// global switch, which those gates short-circuit before reaching). A
	// genuine no-op if DV2 does not exist, the same shape FocusDV2 already
	// has, for the same reason: nothing to close. When it does exist and
	// closes, Focus moves to DV1 if DV2 held it, otherwise Focus is left
	// exactly where it was (S4) — model.closeDV2's own doc comment.
	CloseDV2 key.Binding

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
	// (S2, M2.3). Nav Keys' shared LineUp/LineDown (above) deliberately
	// are not reused here: their "k"/"j" alternates are printable, and
	// S2's "every printable keystroke goes to the Prompt" means the
	// Picker cannot treat them as movement the way every other pane
	// does — only the arrow keys survive that rule, so these bind to
	// them alone.
	PickerUp   key.Binding
	PickerDown key.Binding

	// PickerMode and PickerScope switch the open Picker's mode (File →
	// Grep → Wikipedia, S7; a three-way cycle since S2/C6/M3.2, was File ↔
	// Grep) and Scope (its first value labelled Active Root for Grep,
	// Active Source for File Picker, C6/M1.3; then All Roots, then Starred,
	// S7/S8; Wikipedia mode has none — the key is a no-op there, S2) in
	// place, without touching the Prompt (M2.3 reserves them as non-typing;
	// M5.1/M5.2 give them their actions, C4/M5.1 adds Scope's third value).
	// Tab/Shift+Tab rather than a letter, for the same reason
	// PickerUp/PickerDown avoid "j"/"k": both are non-printable, so typing
	// a query is never at odds with switching.
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

	// PickerStar is S2/M2.5's Picker-only star toggle: acts on the Picker's
	// currently selected Result (either mode), the same conceptual action
	// keys.Star already gives the Sources Nav, Files Nav and a focused Document
	// Viewer, but under its own binding rather than reusing keys.Star's plain
	// "s" — for the identical reason PickerUp/PickerDown avoid "j"/"k" and
	// ToggleHidden avoids a bare letter: S2's "every printable keystroke goes
	// to the Prompt" means the Picker cannot treat "s" as anything but query
	// text. alt+s, matching ToggleHidden's own alt+ prefix for a Picker-only
	// toggle.
	PickerStar key.Binding

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
	// PeekDV2 needs no fallback of its own: like `ctrl+o` (FocusDV2), a
	// plain Ctrl+letter is a raw C0 control byte every terminal delivers
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
	// than the global switch, and do nothing from any other Focus. Tab and
	// Shift+Tab, reused from keys.PickerMode/keys.PickerScope's own physical
	// chords rather than new ones: the two contexts never overlap (the
	// Picker owns the keyboard exclusively while open, handlePickerInput's
	// own gate, so Update's filesPane case is only ever reached with the
	// Picker closed) the same way keys.Activate's "enter" already means one
	// thing on the Files Nav and another on the Sources Nav/Picker without
	// collision. FilesMode cycles Tree -> Recent -> Starred -> Tree ...
	// (filesNav.cycleMode); FilesScope cycles the current mode's own Browse
	// Scope values (filesNav.cycleScope), a no-op while the mode is Tree,
	// which has none (S5).
	FilesMode  key.Binding
	FilesScope key.Binding

	// HistoryBack and HistoryForward walk a focused Document Viewer's own S9
	// back/forward stack (C4/M6.1): matched in Update's dv1Pane/dv2Pane case
	// beside keys.Star, so — like it — they act on whichever viewer holds
	// Focus (DV1 and DV2 each walk their own history independently) and do
	// nothing from the Sources Nav or Files Nav. `[`/`]`, unused elsewhere in
	// the keymap and free of S2's Picker-Prompt "always typing" constraint
	// PickerUp/PickerDown, ToggleHidden and PickerStar each had to route
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
	// (S8), matched in Update's rootsPane case beside Activate/Star. "r" and
	// "n"/"N" — the obvious mnemonics for "rename"/"name" — are already
	// FocusRoots and FindNext/FindPrevious, both global, so this binds "c"
	// instead (free, and — like every binding in this file — a plain letter
	// rather than a function key: the codebase has no F-key precedent, for
	// the same portability reason D55's ActivateDV2 comment gives for
	// avoiding raw escape sequences a terminal might not forward).
	Rename key.Binding
}

// keys is the program's keymap.
var keys = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Star: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "star"),
	),
	Expand: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "expand"),
	),
	Collapse: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "collapse"),
	),
	Activate: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "open"),
	),
	LineUp: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	LineDown: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
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
		key.WithKeys("ctrl+n"),
		key.WithHelp("ctrl+n", "sidebar"),
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
		key.WithKeys("o"),
		key.WithHelp("o", "dv1"),
	),
	FocusDV2: key.NewBinding(
		key.WithKeys("ctrl+o"),
		key.WithHelp("ctrl+o", "dv2"),
	),
	CloseDV2: key.NewBinding(
		key.WithKeys("ctrl+x"),
		key.WithHelp("ctrl+x", "close dv2"),
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
		key.WithKeys("tab"),
		key.WithHelp("tab", "mode"),
	),
	PickerScope: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "scope"),
	),
	ToggleHidden: key.NewBinding(
		key.WithKeys("alt+h"),
		key.WithHelp("alt+h", "hidden files"),
	),
	PickerStar: key.NewBinding(
		key.WithKeys("alt+s"),
		key.WithHelp("alt+s", "star"),
	),
	Peek: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "peek"),
	),
	FilesMode: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "mode"),
	),
	FilesScope: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "scope"),
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
		key.WithKeys("c"),
		key.WithHelp("c", "rename"),
	),
}
