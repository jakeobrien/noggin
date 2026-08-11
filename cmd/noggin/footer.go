package main

import (
	"strings"

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
// it is usually a no-op with DV2 absent, the same "inapplicable binding"
// treatment CloseDV2 gets below) — M7.2 is three removals, not four, for the
// same outcome ("no keys.Focus* binding in any Footer hint list"). help.go's
// helpBindings still lists all four unconditionally, unaffected by either
// reason: the badges replace the footer hint, not the Help Modal's own
// entry.
// Refresh (S2) (C2a/M9.1) is unaffected by the same change and stays
// global — reachable from every one of these three panes, the same way
// Edit/Navigator/Picker already are — so all three panes list it. CloseDV2
// (S4, C3/M2.1) is a no-op whenever DV2 does not exist, which is every
// pane's common case (DV2 is absent at launch, S1, and stays that way until
// something creates it), so it gets the same "inapplicable binding"
// treatment as the viewer case's FindNext/FindPrevious/Cancel below —
// S12's help modal is its discoverability backstop instead. Peek (S7) and
// C3/S3's two targeting modifiers, ActivateDV2/PeekDV2 (C3/M1.4), are
// Files-Nav-only by design, so they appear only in filesPane's list, not the
// other two — a hint for an inapplicable binding is worse than no hint, the
// same reasoning the viewer case below already gives for omitting
// FindNext/FindPrevious/Cancel.
//
// keys.Star (S2/M2.1/M2.2) now appears in every case below (M2.8, TRIAGE
// T5/T6): before this, only the Roots Nav's own star action had a footer
// hint, even though the Files Nav (toggleDocStar) and a focused Document
// Viewer (toggleActiveDocStar) can star just as well — a real action with no
// hint anywhere outside the Roots Nav. Placed right after each pane's own
// open/expand action, mirroring the Roots Nav's existing Activate-then-Star
// order.
//
// keys.FilesMode/keys.FilesScope (C4/M4.1) join filesPane's list here too
// (M4.8, TRIAGE T20, D81), mirroring M2.8's keys.Star treatment above: they
// were a real, frequently-used action (S5's Browse mode/scope cycle) with no
// footer hint anywhere. Files-Nav-only by design, like Peek/ActivateDV2/
// PeekDV2 below, so they appear only in this case, placed right after
// Expand/Collapse — the pane's other own browsing actions — and before Star.
//
// keys.HistoryBack/keys.HistoryForward (C4/M6.1) join the viewer case below
// (M6.3, TRIAGE T41), mirroring M2.8's and M4.8's precedent above: S9's
// per-viewer back/forward walk had no footer hint anywhere. Viewer-only by
// design (DV1/DV2 each walk their own history; Roots Nav and Files Nav have
// none), so they appear only in the default case, placed right after Star —
// the pane's other own document action — and before Edit.
func paneHints(p pane) []key.Binding {
	switch p {
	case rootsPane:
		return []key.Binding{keys.Activate, keys.Star, keys.LineUp, keys.LineDown, keys.Edit, keys.Navigator, keys.Picker, keys.Refresh}
	case filesPane:
		return []key.Binding{keys.Activate, keys.Expand, keys.Collapse, keys.FilesMode, keys.FilesScope, keys.Star, keys.LineUp, keys.LineDown, keys.Edit, keys.Navigator, keys.Picker, keys.Refresh, keys.Peek, keys.ActivateDV2, keys.PeekDV2}
	default: // dv1Pane, dv2Pane
		// FindNext/FindPrevious and Cancel are omitted: they only do anything
		// once a search is active, and a hint for an inapplicable binding is
		// worse than no hint (S12's help modal, M5.3, lists the full keymap
		// regardless — that is the discoverability backstop for these).
		return []key.Binding{keys.LineUp, keys.LineDown, keys.PageUp, keys.PageDown, keys.Top, keys.Bottom, keys.Find, keys.Star, keys.HistoryBack, keys.HistoryForward, keys.Edit, keys.Navigator, keys.Picker, keys.Refresh}
	}
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
// Picker is open, regardless of mode or Scope. PickerStar (S2/M2.5) joins
// them for the identical reason — star/un-star always acts on whatever
// Result is currently selected.
func pickerHints() []key.Binding {
	return []key.Binding{
		keys.Activate, keys.Cancel,
		keys.PickerUp, keys.PickerDown,
		keys.PageUp, keys.PageDown,
		keys.PickerMode, keys.PickerScope,
		keys.ToggleHidden, keys.PickerStar,
	}
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

// joinHints renders title and hints (plus alwaysHints) as one " · "-joined
// line, the shape every footer and the future help modal (S12) can share. An
// empty title is omitted rather than leaving a stray leading separator.
func joinHints(title string, hints []key.Binding) string {
	parts := make([]string, 0, len(hints)+len(alwaysHints)+1)
	if title != "" {
		parts = append(parts, title)
	}
	for _, h := range hints {
		parts = append(parts, formatHint(h))
	}
	for _, h := range alwaysHints {
		parts = append(parts, formatHint(h))
	}
	return strings.Join(parts, " · ")
}

// formatHint renders one binding as "<key> <description>", the same shape
// the pre-M4.2 footer used for its lone quit hint.
func formatHint(k key.Binding) string {
	h := k.Help()
	return h.Key + " " + h.Desc
}
