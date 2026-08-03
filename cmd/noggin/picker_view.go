package main

import (
	"fmt"
	"path/filepath"

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
	// bottom, a title row (the mode/Scope indicator, M2.5), one content row
	// for the textinput itself, and one more for the status line
	// (pickerStatus). The status line lives in the Prompt's box rather than
	// the Results box because it is the widest text the modal renders and the
	// Prompt is the only region spanning the modal's full width — in the
	// Results column (resultsPercent of it) "type at least 3 characters"
	// would wrap at every ordinary terminal size.
	promptOuterHeight = 5

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

// pickerModeScope is the modal's mode-and-Scope indicator (S12's "the modal
// shows its mode and Scope", M2.5), rendered as the Prompt box's title. It
// is the Picker's answer to "what would Enter search?", and the surface
// M5.1/M5.2's in-place switching will be seen through — the reason it reads
// off p.mode/p.scope rather than being passed in as text.
func pickerModeScope(p picker) string {
	mode := "File"
	if p.mode == grepMode {
		mode = "Grep"
	}
	scope := "Active Root"
	if p.scope == allRootsScope {
		scope = "All Roots"
	}
	return mode + " · " + scope
}

// Picker status lines (S12, S14). Exact strings, asserted by
// picker_view_test.go — they are the modal's only way of distinguishing
// "nothing to show yet" from "nothing to show", so each is worded to say
// which.
const (
	pickerBelowThresholdStatus = "type at least 3 characters"
	pickerSearchingStatus      = "searching…"
	pickerNoResultsStatus      = "no results"
)

// pickerStatus is the one line under the Prompt saying what state the search
// is in (S12's empty states, S14's unavailability). Order is priority order,
// most fundamental first: with no ripgrep nothing can ever be searched, so
// that is said before anything about the query; below the threshold no search
// was issued, so "no results" would be a lie; and until results labelled with
// the current query arrive a search is either scheduled or in flight, which is
// "searching", not "nothing found".
//
// Because results carry the query, mode and Scope they answer for
// (pickerResults), "do these results describe what the Prompt now says" is a
// comparison rather than an assumption — the only honest way to caption an
// asynchronously-filled list (M2.4).
func pickerStatus(p picker, avail ripgrepAvailability) string {
	if !avail.available {
		return "Picker unavailable: " + avail.reason
	}
	query := p.input.Value()
	if len([]rune(query)) < pickerQueryMinLength {
		return pickerBelowThresholdStatus
	}
	if !p.resultsCurrent() {
		return pickerSearchingStatus
	}
	if p.searching {
		return pickerSearchingStatus
	}
	if len(p.results.files) == 0 && len(p.results.grep) == 0 {
		return pickerNoResultsStatus
	}
	if p.results.truncated {
		return fmt.Sprintf("first %d results (truncated)", pickerResultCap)
	}
	return ""
}

// pickerSelectedStyle draws the Results row under the selection cursor in
// reverse video — the same bar-of-the-whole-line convention filesNav's
// filesSelectedStyle and rootsNav's reversed row already give their own
// lists, so a Picker Result reads the same way a Nav selection does.
var pickerSelectedStyle = lipgloss.NewStyle().Reverse(true)

// pickerFileRow renders one File Picker Result (S3): the file's Root and its
// path relative to that Root, so two same-named files in different Roots
// read apart (M3.1's own acceptance criterion) — filepath.Base(f.root.path)
// is the Root's label (the same prominent-name half rootsNav.renderRow gives
// a root, though without that row's own left-truncated-parent disambiguation:
// two Roots sharing a base name is not a case this cycle's fixtures or S3
// itself asks this row to solve). f.path is always inside f.root.path (both
// pickerFile's own fields and listFiles' construction guarantee it), so
// filepath.Rel here cannot fail in practice; the fallback to the absolute
// path exists only so a row still renders *something* if that invariant were
// ever broken, rather than silently dropping the row.
func pickerFileRow(f pickerFile, width int, selected bool) string {
	rel, err := filepath.Rel(f.root.path, f.path)
	if err != nil {
		rel = f.path
	}
	line := filepath.Base(f.root.path) + "  " + rel
	style := lipgloss.NewStyle()
	if selected {
		style = pickerSelectedStyle
	}
	return style.Width(width).MaxWidth(width).Render(line)
}

// pickerResultsContent renders the Results region's content (M3.1), sized to
// exactly width x height — the same "expected to be exactly r.content() in
// size" contract renderPane's own doc comment gives every frame pane. Only
// fileMode has rows to show yet (M4.1 gives grepMode its own); a result set
// that does not currently answer the Prompt/mode/Scope (p.resultsCurrent) is
// treated the same as none at all — pickerStatus is already saying
// "searching" for that same mismatch, and showing the previous mode's stale
// rows underneath that caption would contradict it.
func pickerResultsContent(p picker, width, height int) string {
	if width <= 0 || height <= 0 || p.mode != fileMode || !p.resultsCurrent() {
		return ""
	}
	files := p.results.files
	if len(files) == 0 {
		return ""
	}
	return windowedView(width, height, p.selected, len(files), func(idx int) string {
		return pickerFileRow(files[idx], width, idx == p.selected)
	})
}

// renderPicker composes the Picker modal's three regions (S1): a bordered
// Prompt spanning the modal's width — titled with the mode/Scope indicator
// and captioned with the status line (M2.5) — with Results and Preview side
// by side beneath it. Prompt uses focusedBorderStyle — it always owns the
// keyboard while the Picker is open, the same accent renderPane gives
// whichever frame pane currently has focus — Results/Preview use the plain
// borderStyle, same as an unfocused frame pane. Results now carries the File
// Picker's rows (M3.1); Preview still carries none (M3.2 fills it).
func renderPicker(p picker, l pickerLayout, avail ripgrepAvailability) string {
	promptContent := p.promptView(l.prompt.content().width)
	if status := pickerStatus(p, avail); status != "" {
		promptContent += "\n" + status
	}
	prompt := renderBox(focusedBorderStyle, l.prompt, pickerModeScope(p), promptContent)
	resultsContent := pickerResultsContent(p, l.results.content().width, l.results.content().height)
	results := renderBox(borderStyle, l.results, "Results", resultsContent)
	preview := renderBox(borderStyle, l.preview, "Preview", p.preview.View())
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
func overlayPicker(frame string, width, height int, p picker, avail ripgrepAvailability) string {
	l := computePickerLayout(width, height)
	box := lipgloss.NewStyle().MaxWidth(width).MaxHeight(height).Render(renderPicker(p, l, avail))
	bw, bh := lipgloss.Width(box), lipgloss.Height(box)
	x := max((width-bw)/2, 0)
	y := max((height-bh)/2, 0)

	base := lipgloss.NewLayer(frame).X(0).Y(0).Z(0)
	modal := lipgloss.NewLayer(box).X(x).Y(y).Z(1)
	return lipgloss.NewCompositor(base, modal).Render()
}
