package main

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
)

// helpBindings lists every binding relevant to a pane's context for the S12/C1
// Help Modal (M5.3) — the modal's own promise ("lists the bindings that
// pane's keymap actually defines") is a wider contract than paneHints'
// footer subset: for the viewer, Cancel/FindNext/FindPrevious are
// deliberately left out of paneHints (see its comment, "a hint for an
// inapplicable binding is worse than no hint") but belong here, since this
// modal is the discoverability backstop that comment promises. The viewer's
// scroll bindings are pulled from scrollActions (viewer.go) rather than
// re-listed by hand, per that slice's own comment. keys.Picker is added to
// every case for the same reason (D39/M2.3): opening the Picker is global
// and reachable from any Focus, so its binding belongs in the discoverability
// backstop everywhere, the same way Edit/Navigator already are.
//
// keys.PickerUp/PickerDown/PickerMode/PickerScope are also added to every
// case, even though they only do anything once the Picker is already open
// (D46/M2.3a, TRIAGE T5): the Help Modal itself is unreachable while the
// Picker owns the keyboard (see handlePickerInput), so these three panes —
// before the Picker opens — are the *only* place Help is ever reachable
// from to document them. Leaving them out (the original M2.3 shape) left
// S2's "Footer and Help Modal cover the Picker's actions" only half met:
// they lived solely in the footer's pickerHintText, itself fragile (can
// truncate under a narrow terminal with no Help fallback). keys.ToggleHidden
// joins them for the identical reason (C2a/M9.1): it too only does anything
// once the Picker is open, and these three panes are the only place Help
// can ever show it.
//
// keys.PickerStar (S2/M2.5, keymap.go's own comment on it) is deliberately
// left off every case's list here, unlike the Picker-only bindings above:
// its Help() description ("star") collides with keys.Star's own. The
// Footer's own pickerHintText (footer.go) is PickerStar's one
// discoverability surface for now — real, but weaker than the other
// Picker-only actions' backstop here.
//
// keys.Star (S2/M2.1/M2.2) itself now joins filesPane's and the viewer's
// list here (M2.8, TRIAGE T5/T6), not only rootsPane's: the Files Nav
// (toggleDocStar) and a focused Document Viewer (toggleActiveDocStar) can
// star just as well as the Sources Nav, so the discoverability backstop this
// modal gives every other binding was missing for those two actions.
// TestHelpModalListsThePaneOwnBindings's filesPane/dv1Pane cases were
// reworded from pinning keys.Star's absence to pinning its presence.
//
// keys.Rename (S8/M7.2) joins rootsPane's own list only, right after Star,
// the identical Sources-Nav-only reasoning as keys.Peek/ActivateDV2/PeekDV2
// being filesPane-only below: only a Root row can ever be renamed.
//
// keys.FocusRoots/FocusFiles/FocusDV1 (S5/S6) and keys.Refresh (S2)
// (C2a/M9.1) are added to every case for the same reason keys.Picker
// already is: all are global, reachable from any Focus. keys.FocusDV2
// (S5, C3/M1.3) joins them here too, unlike in footer.go's paneHints: it is
// usually a no-op (DV2 absent, S1's common case) the same way
// FindNext/FindPrevious/Cancel usually are with no search active, but this
// modal is exactly the discoverability backstop those bindings' own comment
// (and this one) promises — so, like them, it is listed unconditionally
// even though the terser footer omits it. keys.CloseDV2 (S4, C3/M2.1)
// joins it for the identical reason: also usually a no-op with DV2 absent,
// also listed unconditionally here as its own discoverability backstop.
// keys.Peek (S7) and C3/S3's two
// targeting modifiers, keys.ActivateDV2/keys.PeekDV2 (C3/M1.4), are added
// only to filesPane, since all three are Files-Nav-only by design — listing
// them for rootsPane/dv1Pane/dv2Pane would document bindings those panes can
// never actually trigger.
//
// keys.FilesMode/keys.FilesScope (C4/M4.1) join filesPane's list too (M4.8,
// TRIAGE T20, D81), for the identical Files-Nav-only reason as Peek/
// ActivateDV2/PeekDV2 just above — S5's Browse mode/scope cycle has no other
// discoverability surface than this modal and paneHints' matching addition
// (footer.go). Their Help() key/desc ("tab"/"mode", "shift+tab"/"scope")
// happen to read identically to keys.PickerMode/keys.PickerScope's own,
// already listed here for the Picker-open case (see this func's comment
// above on those) — a real duplicate line, not a bug: the two pairs are
// never reachable at the same time (Picker closed vs. open), so the modal
// simply documents both contexts' "tab cycles mode" shape once each.
//
// keys.HistoryBack/keys.HistoryForward (C4/M6.1) join the viewer case below
// (M6.3, TRIAGE T41), for the identical viewer-only reason as the rest of
// this func's viewer-specific additions: S9's per-viewer back/forward walk
// has no discoverability surface but this modal and paneHints' matching
// addition (footer.go). Their Help() text ("["/"back", "]"/"forward")
// collides with nothing else already listed here, unlike keys.PickerStar
// (this func's own comment above) or keys.FilesMode/keys.FilesScope, so a
// plain string check is enough to pin their presence.
func helpBindings(p pane) []key.Binding {
	switch p {
	case rootsPane:
		return []key.Binding{
			keys.Activate, keys.Star, keys.Rename, keys.LineUp, keys.LineDown,
			keys.Edit, keys.RenderMode, keys.Navigator, keys.Picker, keys.PickerRestore,
			keys.PickerUp, keys.PickerDown, keys.PickerMode, keys.PickerScope, keys.ToggleHidden,
			keys.FocusRoots, keys.FocusFiles, keys.FocusDV1, keys.FocusDV2, keys.CloseDV2, keys.Refresh,
			keys.Help, keys.Quit,
		}
	case filesPane:
		return []key.Binding{
			keys.Activate, keys.Expand, keys.Collapse, keys.FilesMode, keys.FilesScope, keys.Star, keys.LineUp, keys.LineDown,
			keys.Edit, keys.RenderMode, keys.Navigator, keys.Picker, keys.PickerRestore,
			keys.PickerUp, keys.PickerDown, keys.PickerMode, keys.PickerScope, keys.ToggleHidden,
			keys.FocusRoots, keys.FocusFiles, keys.FocusDV1, keys.FocusDV2, keys.CloseDV2, keys.Refresh, keys.Peek,
			keys.ActivateDV2, keys.PeekDV2,
			keys.Help, keys.Quit,
		}
	default: // dv1Pane, dv2Pane
		bindings := make([]key.Binding, 0, len(scrollActions)+20)
		for _, a := range scrollActions {
			bindings = append(bindings, a.binding)
		}
		return append(bindings,
			keys.Find, keys.FindNext, keys.FindPrevious, keys.Cancel,
			keys.Star, keys.HistoryBack, keys.HistoryForward, keys.Edit, keys.RenderMode, keys.Navigator, keys.Picker, keys.PickerRestore,
			keys.PickerUp, keys.PickerDown, keys.PickerMode, keys.PickerScope, keys.ToggleHidden,
			keys.FocusRoots, keys.FocusFiles, keys.FocusDV1, keys.FocusDV2, keys.CloseDV2, keys.Refresh,
			keys.Help, keys.Quit,
		)
	}
}

// modalStyle borders the Help Modal in a colour no pane border uses
// (Theme.Border, Theme.Accent), so it reads as a layer above
// the frame rather than another (possibly focused-looking) pane within it.
var modalStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(theme.Highlight).
	Padding(0, 1)

// helpBox renders the Help Modal's content for a pane's context: the pane's
// title, then every binding helpBindings names for it, through bubbles/help's
// own column renderer (going with the grain, D1, rather than hand-rolling
// key/description alignment). ShowAll picks the "full help" (column) form
// over the single-line "short help" form the footer already uses for its own
// truncating hints; leaving the model's width at zero (its zero value) means
// FullHelpView never truncates — the modal is the backstop, so it always
// shows everything.
func helpBox(p pane) string {
	hm := help.New()
	hm.ShowAll = true
	body := hm.FullHelpView([][]key.Binding{helpBindings(p)})
	title := titleStyle.Render("Help: " + p.title())
	return modalStyle.Render(title + "\n\n" + body)
}

// overlayHelp composes the Help Modal on top of an already-rendered frame,
// centered, via lipgloss v2's cell-based layer compositor — true compositing
// (only the modal's own cells change) rather than lipgloss.Place's
// whitespace-only placement, so the pane underneath stays visible around the
// modal's edges, which is what the Help Modal "opening over" a pane means.
//
// The box is clamped to (width, height) first: a Compositor sizes its output
// canvas to the union of every layer's bounds, not the base layer's own size
// (verified against this lipgloss build), so an unclamped modal taller or
// wider than the terminal — e.g. the viewer's full binding list on a short
// terminal — would grow the rendered frame past the terminal's actual size.
// Since View() draws inline with no alt screen, that means unwanted
// scrolling rather than a modal "over" the pane; clamping instead truncates
// the binding list gracefully via lipgloss's own MaxWidth/MaxHeight, the same
// tool the footer already uses for its own space budget.
func overlayHelp(frame string, width, height int, focused pane) string {
	box := lipgloss.NewStyle().MaxWidth(width).MaxHeight(height).Render(helpBox(focused))
	bw, bh := lipgloss.Width(box), lipgloss.Height(box)
	x := max((width-bw)/2, 0)
	y := max((height-bh)/2, 0)

	base := lipgloss.NewLayer(frame).X(0).Y(0).Z(0)
	modal := lipgloss.NewLayer(box).X(x).Y(y).Z(1)
	return lipgloss.NewCompositor(base, modal).Render()
}
