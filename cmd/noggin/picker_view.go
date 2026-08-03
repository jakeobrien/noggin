package main

import (
	"charm.land/lipgloss/v2"
)

// Picker modal proportions (M2.2): the modal's own footprint at a terminal
// size, and the split of that footprint between its three regions (S1's
// Prompt, Results list, Preview pane). layout.go's own computeLayout
// "knows nothing about modals" (its doc comment) on purpose — this is the
// first multi-pane layout *inside* a modal in the codebase
// (codebase-picker-integration.md), so it gets its own home here rather
// than stretching layout.go's contract.
const (
	// pickerWidthPercent/pickerHeightPercent are the modal's share of the
	// terminal, the same shape navigatorPercent gives the Navigator column
	// — clamped below by min*, and implicitly above by the terminal's own
	// size (clamp's hi), so the modal never claims more room than exists
	// (M2.2's "never exceeds the terminal") and always leaves a margin
	// around its edges on anything but the smallest terminals (M2.2's "the
	// frame stays visible around its edges").
	pickerWidthPercent  = 80
	pickerHeightPercent = 80
	minPickerWidth      = 40
	minPickerHeight     = 10

	// promptOuterHeight is the Prompt's own bordered box: border top and
	// bottom, a title row (empty for now — M2.5 puts the mode/Scope
	// indicator there), and one content row for the textinput itself.
	promptOuterHeight = 4

	// resultsPercent is the Results list's share of the modal's width, the
	// remainder going to the Preview pane — the same shape rootsPercent
	// gives the Navigator column's own vertical split, with the same
	// even-split fallback (splitPickerWidth) layout.go's splitHeight
	// already gives that split when a proportional share would starve
	// either side.
	resultsPercent  = 40
	minResultsWidth = 20
)

// pickerLayout is the Picker modal's own regions at a terminal size, in the
// same outer-box rect units (layout.go) the frame's own panes use.
type pickerLayout struct {
	prompt  rect
	results rect
	preview rect
}

// computePickerLayout sizes the modal at a terminal of width x height: a
// bordered Prompt spanning the modal's width on top, Results and Preview
// splitting the remaining height side by side beneath it. Clamped first —
// the same clamp-before-render discipline overlayHelp's own comment
// explains (a Compositor sizes its canvas to the union of every layer's
// bounds, not the base layer's, so an unclamped modal would grow the
// rendered frame past the terminal rather than fitting inside it).
func computePickerLayout(width, height int) pickerLayout {
	if width <= 0 || height <= 0 {
		return pickerLayout{}
	}
	w := clamp(width*pickerWidthPercent/100, minPickerWidth, width)
	h := clamp(height*pickerHeightPercent/100, minPickerHeight, height)

	promptHeight := min(promptOuterHeight, h)
	bodyHeight := h - promptHeight
	resultsWidth, previewWidth := splitPickerWidth(w)

	return pickerLayout{
		prompt:  rect{width: w, height: promptHeight},
		results: rect{width: resultsWidth, height: bodyHeight},
		preview: rect{width: previewWidth, height: bodyHeight},
	}
}

// splitPickerWidth divides the modal's width between Results and Preview,
// falling back to an even split when the proportional share would starve
// either side — the same fallback shape layout.go's splitHeight already
// gives the Navigator column's vertical split.
func splitPickerWidth(total int) (results, preview int) {
	if total <= 0 {
		return 0, 0
	}
	results = total * resultsPercent / 100
	if results < minResultsWidth || total-results < minResultsWidth {
		results = total / 2
	}
	return results, total - results
}

// renderBox draws one bordered region at exactly r's outer size, border
// included: a styled title on the first content row, then content. This is
// the frame's own renderPane's box-drawing shape, pulled out so the
// Picker's regions (which have no pane/focus of their own) can reuse it
// instead of a second copy — renderPane itself now just supplies its
// per-pane style and title and delegates here.
//
// lipgloss v2 counts the border inside Width/Height (PR #451), so r's outer
// size is passed straight through rather than pre-deducting the border, the
// same "pass it straight through" renderPane's own comment already notes.
func renderBox(style lipgloss.Style, r rect, title, content string) string {
	if r.empty() {
		return ""
	}
	box := style.Width(r.width).Height(r.height)
	styledTitle := titleStyle.Render(title)
	if content == "" {
		return box.Render(styledTitle)
	}
	return box.Render(styledTitle + "\n" + content)
}

// promptView renders the Prompt's textinput at width — the same
// "SetWidth then View" shape find.go's own findView already gives its
// textinput. SetWidth is a pointer-receiver method and p is a parameter
// here, so p.input is addressable and this never mutates the caller's
// picker.
func (p picker) promptView(width int) string {
	p.input.SetWidth(width)
	return p.input.View()
}

// renderPicker composes the Picker modal's three regions (S1): a bordered
// Prompt spanning the modal's width, Results and Preview side by side
// beneath it. Prompt uses focusedBorderStyle — it always owns the keyboard
// while the Picker is open, the same accent renderPane gives whichever
// frame pane currently has focus — Results/Preview use the plain
// borderStyle, same as an unfocused frame pane. M2.2 wires only this shell:
// Results/Preview carry no content yet (M3.1/M3.2 fill them) and no
// status/empty-state text (M2.5's job) — just the three regions, sized and
// bordered.
func renderPicker(p picker, l pickerLayout) string {
	prompt := renderBox(focusedBorderStyle, l.prompt, "", p.promptView(l.prompt.content().width))
	results := renderBox(borderStyle, l.results, "Results", "")
	preview := renderBox(borderStyle, l.preview, "Preview", "")
	body := lipgloss.JoinHorizontal(lipgloss.Top, results, preview)
	return lipgloss.JoinVertical(lipgloss.Left, prompt, body)
}

// overlayPicker composes the Picker modal on top of an already-rendered
// frame, centered, via lipgloss v2's cell-based layer compositor — the same
// true compositing overlayHelp already gives the Help Modal, so the pane
// underneath stays visible around the modal's edges (M2.2's "the frame
// stays visible around its edges"). Clamped to (width, height) first, for
// the reason computePickerLayout's own comment gives; MaxWidth/MaxHeight
// here is the suspenders to computePickerLayout's belt (layout-correct
// sizing does not by itself guarantee lipgloss never pads a row wider, the
// same reasoning overlayHelp's comment already spells out).
func overlayPicker(frame string, width, height int, p picker) string {
	l := computePickerLayout(width, height)
	box := lipgloss.NewStyle().MaxWidth(width).MaxHeight(height).Render(renderPicker(p, l))
	bw, bh := lipgloss.Width(box), lipgloss.Height(box)
	x := max((width-bw)/2, 0)
	y := max((height-bh)/2, 0)

	base := lipgloss.NewLayer(frame).X(0).Y(0).Z(0)
	modal := lipgloss.NewLayer(box).X(x).Y(y).Z(1)
	return lipgloss.NewCompositor(base, modal).Render()
}
