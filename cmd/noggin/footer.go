package main

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
)

// paneHints lists the key bindings relevant to a focused pane, most important
// first (S11's "key hints relevant to the focused pane"). Order matters:
// hintText drops from the end of this list first when the footer must
// truncate, so put what a newcomer needs most up front.
//
// FocusRoots/FocusFiles/FocusDV1/FocusDV2 (S5/S6) are absent from every case
// below, but for two different reasons. FocusRoots/FocusFiles/FocusDV1 were
// removed here (S10, M7.2): each pane's own title row now carries its own
// focus-key badge instead (M7.1's right-aligned [r]/[f]/[o]/[^o]), a more
// direct, always-visible discoverability surface than a footer hint that
// would only ever compete with every other action for the same line — the
// Footer no longer needs to repeat what the pane itself already shows.
// FocusDV2 was never listed here in the first place (predates M7.2, C3/M1.3:
// it is usually a no-op with DV2 absent) — M7.2 is three removals, not four, for the
// same outcome ("no keys.Focus* binding in any Footer hint list"). help.go's
// helpBindings still lists all four unconditionally, unaffected by either
// reason: the badges replace the footer hint, not the Help Modal's own
// entry.
// Refresh (S2) (C2a/M9.1) is unaffected by the same change and stays
// global — reachable from every one of these three panes, the same way
// Navigator/Picker already are — so all three panes list it. CloseViewer
// is listed for the viewers only: from a viewer it always acts (DV1's
// own ctrl+w takes over DV2 or clears DV1), while from a nav pane it is a
// no-op whenever DV2 does not exist, the common case. Peek (S7) and
// C3/S3's two targeting modifiers, ActivateDV2/PeekDV2 (C3/M1.4), are
// Files-Nav-only by design, so they appear only in filesPane's list, not the
// other two — a hint for an inapplicable binding is worse than no hint, the
// same reasoning the viewer case below already gives for omitting
// FindNext/FindPrevious/Cancel.
//
// keys.Star (S2/M2.1/M2.2) now appears in every case below (M2.8, TRIAGE
// T5/T6): before this, only the Sources Nav's own star action had a footer
// hint, even though the Files Nav (toggleDocStar) and a focused Document
// Viewer (toggleActiveDocStar) can star just as well — a real action with no
// hint anywhere outside the Sources Nav. Placed right after each pane's own
// open/expand action, mirroring the Sources Nav's existing Activate-then-Star
// order.
//
// keys.FilesMode/keys.FilesScope (C4/M4.1) join filesPane's list here too
// (M4.8, TRIAGE T20), mirroring M2.8's keys.Star treatment above: they
// were a real, frequently-used action (S5's Browse mode/scope cycle) with no
// footer hint anywhere. Files-Nav-only by design, like Peek/ActivateDV2/
// PeekDV2 below, so they appear only in this case, placed right after
// Expand/Collapse — the pane's other own browsing actions — and before Star.
//
// keys.HistoryBack/keys.HistoryForward (C4/M6.1) join the viewer case below
// (M6.3, TRIAGE T41), mirroring M2.8's and M4.8's precedent above: S9's
// per-viewer back/forward walk had no footer hint anywhere. Viewer-only by
// design (DV1/DV2 each walk their own history; Sources Nav and Files Nav have
// none), so they appear only in the default case, placed right after Star —
// the pane's other own document action — and before Edit.
//
// keys.Rename (S8/M7.2) joins rootsPane's own list, right after Star — the
// same "pane's other own row action" placement FilesMode/FilesScope and
// HistoryBack/HistoryForward each got above. Sources-Nav-only by design
// (only a Root row can be renamed), so it appears only here, not in
// filesPane's or the viewer's own lists.
//
// Up/Down are omitted from the nav panes (and the Picker): moving a
// selection with the arrows needs no hint, and the room goes to actions that
// do — the Files Nav's view toggles (alt+h/alt+i, beside alt+m), the Sources
// Nav's reopen-picker and theme keys. The viewer keeps its movement keys but
// as the one navHint entry, which frees room for RenderMode and CloseViewer.
func paneHints(p pane) []key.Binding {
	switch p {
	case rootsPane:
		return []key.Binding{keys.Activate, keys.Star, keys.Rename, keys.Navigator, keys.Picker, keys.Refresh, keys.PickerRestore, keys.Theme}
	case filesPane:
		return []key.Binding{keys.Activate, keys.Expand, keys.Collapse, keys.Peek, keys.FilesMode, keys.FilesScope, keys.Star, keys.Navigator, keys.Picker, keys.Refresh, keys.FileFilter, keys.ToggleHidden, keys.ToggleIgnored, keys.ActivateDV2, keys.PeekDV2}
	default: // dv1Pane, dv2Pane
		// FindNext/FindPrevious and Cancel are omitted: they only do anything
		// once a search is active, and a hint for an inapplicable binding is
		// worse than no hint (S12/C1's help modal, M5.3, lists the full keymap
		// regardless — that is the discoverability backstop for these).
		return []key.Binding{navHint(), keys.Find, keys.Star, keys.HistoryBack, keys.HistoryForward, keys.Edit, keys.RenderMode, keys.CloseViewer, keys.Navigator, keys.Picker, keys.Refresh}
	}
}

// navHint folds the Document Viewer's scroll keys into one "Nav" hint, so the
// footer spends one entry, not six, on movement everyone expects. A
// display-only binding: it matches no key, it only carries Help text built
// from the real bindings' own.
func navHint() key.Binding {
	var ks []string
	for _, b := range []key.Binding{keys.LineUp, keys.LineDown, keys.PageUp, keys.PageDown, keys.Top, keys.Bottom} {
		ks = append(ks, b.Help().Key)
	}
	return key.NewBinding(key.WithHelp(strings.Join(ks, "/"), "nav"))
}

// pickerHints lists the key bindings relevant to the open Picker (S2's "the
// Footer shows the Picker's actions", M2.3), most important first — the
// same truncation-order contract paneHints' own comment gives, since both
// feed the same truncatedHintText. Cancel is included here (unlike
// paneHints' viewer case, which omits it as "inapplicable most of the
// time"): while the Picker is open Cancel always does something —
// dismisses it — so it is never the inapplicable hint that comment warns
// against. ToggleHidden (S4/M6.2, C2a/M9.1) is included for the same
// reason PickerMode/PickerScope are: it always does something while the
// Picker is open, regardless of mode or Scope. Star (S2/M2.5, the same
// ctrl+s every pane stars with) joins them for the identical reason —
// star/un-star always acts on whatever Result is currently selected.
// ActivateDV2 closes the list: the Picker opens its selection in DV2 as the
// Files Nav does, a real action with no other footer hint. PickerUp/PickerDown
// are left off, like the nav panes' Up/Down.
func pickerHints() []key.Binding {
	return []key.Binding{
		keys.Activate, keys.Cancel,
		keys.PageUp, keys.PageDown,
		keys.PickerMode, keys.PickerScope,
		keys.ToggleHidden, keys.FileFilter, keys.Star, keys.ToggleIgnored,
		keys.ActivateDV2,
	}
}

// renameFooterText is the Footer's line while the Sources Nav's rename
// input (S8/M7.2) has the keyboard. Unlike find's own query, which
// model.renderFooter renders directly (findView), the rename input lives
// inline on the selected row instead (rootsnav.go's renderRow — S8's own
// "edits ... inline on its row"), so the footer has nothing of its own to
// draw there; it only needs to say what Enter/Esc do right now, since
// rootsPane's ordinary hints (Activate "open", Star "star" …) no longer
// apply while every other key types into the row's input instead
// (model.handleRenameInput). Display-only bindings rather than
// keys.Activate/keys.Cancel: their own Help() text would misdescribe them
// ("open" is not what Enter does mid-rename). Help/Quit are left off (no
// alwaysHints): both are unreachable while renaming, the same exclusivity
// that keeps them off the footer during an in-progress find query.
func renameFooterText() string {
	return renderHints("Renaming", []key.Binding{
		key.NewBinding(key.WithHelp("enter", "confirm")),
		key.NewBinding(key.WithHelp("esc", "cancel")),
	})
}

// alwaysHints appear last on every footer, in this order, and are never
// dropped by hintText's truncation: S11 requires quit always shown, and the
// help affordance advertised even when nothing else fits.
var alwaysHints = []key.Binding{keys.Help, keys.Quit}

// hintText builds the footer's key-hint line for a focused pane at a given
// width: the pane's title, then as many of its hints as fit, then
// alwaysHints. It tries the full hint set first and drops pane-specific
// hints one at a time from the end (least essential, per paneHints' order)
// until the line fits or none are left — "truncate gracefully" (S11).
// alwaysHints are appended at every step, so quit and the help affordance
// survive even the narrowest footer that still shows anything at all. If
// even the title plus alwaysHints alone still overflow, the title itself is
// the last thing dropped, so those two never fail to show for want of space.
func hintText(p pane, width int) string {
	return truncatedHintText(p.title(), paneHints(p), width)
}

// pickerHintText is hintText's Picker-shaped twin (M2.3): the open Picker
// has no pane/title of its own the way a frame region does (it is a layer
// over the frame, not a pane — picker.go's own doc comments), so this names
// it "Picker" directly rather than routing through pane.title(), sharing
// truncatedHintText's truncation with hintText rather than a second copy of
// it.
func pickerHintText(width int) string {
	return truncatedHintText("Picker", pickerHints(), width)
}

// truncatedHintText builds one footer-shaped hint line for a title and hint
// set at a given width: as many hints as fit (dropped from the end, least
// essential per the caller's own ordering), then alwaysHints, truncating
// gracefully per S11. hintText and pickerHintText are both this with a
// different (title, hints) pair — the truncation algorithm itself does not
// care which.
func truncatedHintText(title string, hints []key.Binding, width int) string {
	for n := len(hints); n >= 0; n-- {
		text := joinHints(title, hints[:n])
		if lipgloss.Width(text) <= width {
			return text
		}
	}
	return joinHints("", nil)
}

// hintSeparator divides the footer's entries.
const hintSeparator = " | "

// joinHints renders title and hints, plus alwaysHints, as one footer line, the
// shape every hint footer shares. An empty title is omitted rather than
// leaving a stray leading separator.
func joinHints(title string, hints []key.Binding) string {
	return renderHints(title, append(append([]key.Binding{}, hints...), alwaysHints...))
}

// renderHints renders title then each hint, joined by hintSeparator in the
// Border colour so the entries read apart without competing with them.
func renderHints(title string, hints []key.Binding) string {
	parts := make([]string, 0, len(hints)+1)
	if title != "" {
		parts = append(parts, footerStyle().Render(title))
	}
	for _, h := range hints {
		parts = append(parts, formatHint(h))
	}
	return strings.Join(parts, lipgloss.NewStyle().Foreground(theme.Border).Render(hintSeparator))
}

// formatHint renders one binding as "<key> <Action>": the key in the Theme's
// Accent, so the keys are what the eye finds along the line, and the action
// capitalised in the footer's Subtle colour. Every piece is styled here,
// rather than inheriting footerStyle from renderFooter's wrapping Render,
// because an inner style's reset would end that inherited colour early.
func formatHint(k key.Binding) string {
	h := k.Help()
	return lipgloss.NewStyle().Foreground(theme.Accent).Render(h.Key) + " " +
		footerStyle().Render(capitalize(h.Desc))
}

// capitalize upper-cases s's first letter, leaving the rest as written.
func capitalize(s string) string {
	r, n := utf8.DecodeRuneInString(s)
	if n == 0 {
		return s
	}
	return string(unicode.ToUpper(r)) + s[n:]
}
