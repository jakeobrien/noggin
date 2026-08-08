package main

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
)

// helpBindings lists every binding relevant to a pane's context for the S12
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
// keys.FocusRoots/FocusFiles (S6) and keys.Refresh (S2) (C2a/M9.1) are
// added to every case for the same reason keys.Picker already is: all
// three are global, reachable from any Focus. keys.Peek (S7) is added only
// to filesPane, since it is Files-Nav-only by design — listing it for
// rootsPane/dv1Pane/dv2Pane would document a binding that pane can never
// actually trigger.
//
// M1.1: Tab/Shift+Tab pane-cycling (keys.NextPane/PrevPane) is gone (S5) — a
// later task (M1.3) adds the direct dv1Pane/dv2Pane focus-jump keys.
func helpBindings(p pane) []key.Binding {
	switch p {
	case rootsPane:
		return []key.Binding{
			keys.Activate, keys.Star, keys.LineUp, keys.LineDown,
			keys.Edit, keys.Navigator, keys.Picker,
			keys.PickerUp, keys.PickerDown, keys.PickerMode, keys.PickerScope, keys.ToggleHidden,
			keys.FocusRoots, keys.FocusFiles, keys.Refresh,
			keys.Help, keys.Quit,
		}
	case filesPane:
		return []key.Binding{
			keys.Activate, keys.Expand, keys.Collapse, keys.LineUp, keys.LineDown,
			keys.Edit, keys.Navigator, keys.Picker,
			keys.PickerUp, keys.PickerDown, keys.PickerMode, keys.PickerScope, keys.ToggleHidden,
			keys.FocusRoots, keys.FocusFiles, keys.Refresh, keys.Peek,
			keys.Help, keys.Quit,
		}
	default: // dv1Pane, dv2Pane
		bindings := make([]key.Binding, 0, len(scrollActions)+17)
		for _, a := range scrollActions {
			bindings = append(bindings, a.binding)
		}
		return append(bindings,
			keys.Find, keys.FindNext, keys.FindPrevious, keys.Cancel,
			keys.Edit, keys.Navigator, keys.Picker,
			keys.PickerUp, keys.PickerDown, keys.PickerMode, keys.PickerScope, keys.ToggleHidden,
			keys.FocusRoots, keys.FocusFiles, keys.Refresh,
			keys.Help, keys.Quit,
		)
	}
}

// modalStyle borders the Help Modal in a color no pane border uses
// (borderStyle's 240, focusedBorderStyle's 205), so it reads as a layer above
// the frame rather than another (possibly focused-looking) pane within it.
var modalStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("214")).
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
