package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"

	"noggin/internal/wikipedia"
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
	// Results list (resultsPercent of it) "type at least 3 characters"
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
// included: styledTitle (already styled by the caller) on the first content
// row, then content. This is the frame's own renderPane's box-drawing
// shape, pulled out so the Picker's regions (which have no pane/focus of
// their own) can reuse it instead of a second copy — renderPane itself now
// just supplies its per-pane style and title and delegates here.
//
// The title arrives pre-styled, not as a bare string renderBox styles
// itself, since S7/M7.1 needs one title (the Document Viewer's, while
// Peeking) styled differently (inverted, peekTitleStyle) from every other
// pane's titleStyle — a choice only the caller can make.
//
// lipgloss v2 counts the border inside Width/Height (PR #451), so r's outer
// size is passed straight through rather than pre-deducting the border, the
// same "pass it straight through" renderPane's own comment already notes.
func renderBox(style lipgloss.Style, r rect, styledTitle, content string) string {
	if r.empty() {
		return ""
	}
	// Height() alone is a floor, not a ceiling (lipgloss v2 pads content
	// shorter than it, but passes taller content through unchanged); MaxHeight
	// is the actual clamp. This is a structural backstop, not the fix for any
	// specific pane's overflow — each pane's own renderer is responsible for
	// never handing renderBox more than r.height lines in the first place —
	// but it keeps a future regression from growing the whole frame and
	// pushing panes above it off screen the way an unclamped Files Nav did.
	box := style.Width(r.width).Height(r.height).MaxHeight(r.height)
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
func pickerModeScope(p picker, hideDotfiles bool) string {
	mode := "File"
	// Grep's own Scope keeps reading "Active Root" (C6/M1.3): Grep never
	// reaches Wikipedia (S6), so its activeRootScope value always names a
	// literal Root. File Picker's identical zero value instead reads
	// "Active Source" — today that can be a Root or (S6) Wikipedia itself.
	// Wikipedia mode (S2, C6/M3.2) has no Scope at all: it always reads
	// "-", ignoring p.scope entirely (switchPickerScope never changes it
	// in this mode either, so there is nothing stale to worry about).
	scope := "Active Source"
	switch p.mode {
	case grepMode:
		mode = "Grep"
		scope = "Active Root"
	case wikipediaMode:
		mode = "Wikipedia"
		scope = "-"
	}
	if p.mode != wikipediaMode {
		switch p.scope {
		case allRootsScope:
			scope = "All Roots"
		case starredScope:
			scope = "Starred"
		}
	}
	label := mode + " · " + scope
	if hideDotfiles {
		// S4/M6.2's alt+h toggle has no other visible confirmation — the
		// title is the one place mode/Scope already surface "what would
		// Enter search", and a dotfile-hiding candidate set is exactly that
		// same kind of fact about the search, not a new indicator of its own.
		label += " · Hidden"
	}
	return label
}

// Picker status lines (S12, S14). Exact strings, asserted by
// picker_view_test.go — they are the modal's only way of distinguishing
// "nothing to show yet" from "nothing to show", so each is worded to say
// which.
const (
	// pickerBelowThresholdStatus's count tracks pickerQueryMinLength (S13,
	// C6/M3.1: 2, revised down from S12/C2's 3).
	pickerBelowThresholdStatus = "type at least 2 characters"
	pickerSearchingStatus      = "searching…"
	pickerNoResultsStatus      = "no results"
)

// pickerWikipediaPreviewPlaceholder is the short placeholder S2 asks the
// Preview pane to show in Wikipedia mode instead of article content
// ("There is no Preview"): it never depends on the selection or on any
// fetch, unlike the Preview's raw-bytes File/Grep content, so it renders
// unconditionally whenever p.mode is wikipediaMode (renderPicker).
const pickerWikipediaPreviewPlaceholder = "No preview — Activate to open the article"

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
func pickerStatus(p picker, avail ripgrepAvailability, hideDotfiles bool) string {
	// C6/M3.3: Wikipedia mode needs no ripgrep (searchable's own doc
	// comment) — S14's unavailability message is a File/Grep-only concern.
	if !avail.available && p.mode != wikipediaMode {
		return "Picker unavailable: " + avail.reason
	}
	query := p.input.Value()
	if len([]rune(query)) < pickerQueryMinLength {
		return pickerBelowThresholdStatus
	}
	// C6/M3.3: a Wikipedia search failure takes priority over the generic
	// searching/no-results captions below — it is more specific information
	// about the exact same "nothing new to show" state, and applyPickerResults
	// has already relabeled p.results so resultsCurrent still holds while
	// this is set (wikipediaErr's own doc comment).
	if p.mode == wikipediaMode && p.wikipediaErr != "" {
		return p.wikipediaErr
	}
	if !p.resultsCurrent(hideDotfiles) {
		return pickerSearchingStatus
	}
	if p.searching {
		return pickerSearchingStatus
	}
	if len(p.results.files) == 0 && len(p.results.grep) == 0 && len(p.results.wikipedia) == 0 {
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

// pickerRootLabel renders a Root's label for a File Picker row (S3, M3.1a/
// D50): its base name, then a gap, then its left-truncated parent path
// (`leftTruncate`, rootsnav.go) — the exact same "prominent base name plus
// disambiguating parent" layout rootsNav.renderRow gives every Root row,
// reused here rather than the bare filepath.Base(f.root.path) M3.1 shipped
// with, which read two Roots sharing a base name (different parents)
// identically. budget is what's left of width after the name, the one gap
// ahead of the parent segment, and rest (everything the row still renders
// after the Root label, including that segment's own connecting gap — e.g.
// "  " plus the file's relative path) — the same single-gap "reserve
// everything else first" shape rootsNav.renderRow's own budget calculation
// uses for its row's suffix. Subtracting gap's width a second time here
// (TRIAGE T13/D52) double-counted a gap that only appears once in the
// rendered row, under-computing budget by one gap width (~2 cols) on every
// row — caught only after M3.1a's own disambiguation test happened to run
// at a width wide enough (140x30, not the project's standard 80x24) to not
// need the lost budget.
func pickerRootLabel(f pickerFile, width int, rest string) string {
	name := filepath.Base(f.root.path)
	parent := filepath.Dir(f.root.path) + string(filepath.Separator)
	const gap = "  "
	budget := max(width-lipgloss.Width(name)-lipgloss.Width(gap)-lipgloss.Width(rest), 0)
	return name + gap + leftTruncate(parent, budget)
}

// pickerLocationLabel is pickerFileRow's "where is this file" segment
// (S3/S8, C4/M5.1): a covering Root's own name+parent (pickerRootLabel)
// for an ordinary candidate, or — for a Starred-scope orphan
// (pickerFile.orphan, starredPickerFiles) — the document's own
// left-truncated parent directory alone, since there is no Root name to
// pair it with. This is Picker rows' own counterpart to
// browse_list.go's buildBrowseRow/files.go's renderFlatRow orphan fallback
// (orphanDir): the same "no covering Root" fact, rendered in whichever
// layout its own caller already uses — name+parent there, Root-label+rel
// here.
func pickerLocationLabel(f pickerFile, width int, rest string) string {
	if f.orphan {
		budget := max(width-lipgloss.Width(rest), 0)
		return leftTruncate(filepath.Dir(f.path), budget)
	}
	return pickerRootLabel(f, width, rest)
}

// pickerFileRow renders one File Picker Result (S3): the file's Root and its
// path relative to that Root, so two same-named files in different Roots
// read apart (M3.1's own acceptance criterion), including when two Roots
// also share a base name (M3.1a/D50, pickerRootLabel). f.path is always
// inside f.root.path (both pickerFile's own fields and listFiles'
// construction guarantee it) for every non-orphan candidate, so
// filepath.Rel here cannot fail in practice for one; the fallback to the
// absolute path exists only so a row still renders *something* if that
// invariant were ever broken, rather than silently dropping the row. A
// Starred-scope orphan (S8, C4/M5.1) has no f.root at all —
// starredPickerFiles leaves it the zero value — so rel is just the file's
// own base name there, paired with pickerLocationLabel's parent-directory
// fallback rather than a Root's own name+parent. suffix ("  "+rel) is
// computed once and reused both as pickerLocationLabel's rest argument and
// as the literal text appended after it (TRIAGE T17/D52) — the two must
// always agree since they describe the same rendered characters, so a
// single local var makes that structural rather than something two call
// sites have to keep in sync by hand.
// query is what the row was matched against, needed only to highlight the
// matched characters of the file name (S5/M6.1) — an empty query renders
// exactly the unhighlighted row M3.1 shipped.
// starred is S3/D72's Picker-side star marker: a starred Result renders the
// same "\u2605 "/"  " two-column field Files Nav, Sources Nav and the Document
// Viewer already use, computed live from the caller's docs lookup rather
// than cached on picker itself (unlike M2.7's m.files.starred, there is
// nothing here that needs syncing — every frame re-renders straight from
// the S1 store). The marker is prepended to line, not to the already-
// rendered row, and columns is shifted to match: line's own text stays
// exactly what suffixColumns computed positions against, and
// highlightMatches' self-contained-per-run styling (its own doc comment)
// then covers the marker for free — a separately pre-styled marker would
// introduce a mid-string reset that cancels a selected row's reverse video
// from there on, the same nesting hazard that comment warns about.
func pickerFileRow(f pickerFile, query string, width int, selected bool, starred bool) string {
	marker := starMarker(starred)
	contentWidth := max(width-lipgloss.Width(marker), 0)
	// An orphan (S8) has no covering Root to be relative to —
	// pickerLocationLabel already names its directory, so the suffix here
	// is just the file's own name, the same "no further path to add" shape
	// buildBrowseRow's orphan row gives the Files Nav's own flat list.
	rel := filepath.Base(f.path)
	if !f.orphan {
		if r, err := filepath.Rel(f.root.path, f.path); err == nil {
			rel = r
		} else {
			rel = f.path
		}
	}
	suffix := "  " + rel
	line := pickerLocationLabel(f, contentWidth, suffix) + suffix
	name := filepath.Base(f.path)
	columns := suffixColumns(line, name, fileNameMatchPositions(name, query))
	full, shifted := prependMarker(marker, line, columns)
	return styleRow(highlightMatches(full, shifted, selected), width, selected)
}

// starMarker returns the two-column starred-state field used by Picker rows
// (S3/D72): "\u2605 " when starred, two blank columns otherwise, so every row
// in a column stays aligned regardless of mixed starred state. files.go,
// rootsnav.go and model.go's Document Viewer title render the identical
// "\u2605 "/"  " literal independently rather than calling this helper
// (TRIAGE T17/D76) — keep the two in sync by hand if either changes.
func starMarker(starred bool) string {
	if starred {
		return "\u2605 "
	}
	return "  "
}

// prependMarker puts marker in front of line and shifts columns (rune
// indexes into line, as suffixColumns/clipColumns already produce) by
// marker's own rune width, so they keep addressing the same characters of
// the combined string — shared by pickerFileRow and pickerGrepRow, both of
// which compute columns against the unprefixed line and then need the
// marker folded in before highlightMatches sees it.
func prependMarker(marker, line string, columns []int) (string, []int) {
	offset := len([]rune(marker))
	shifted := make([]int, len(columns))
	for i, c := range columns {
		shifted[i] = c + offset
	}
	return marker + line, shifted
}

// suffixColumns translates positions within name — rune indexes into the
// name itself (fileNameMatchPositions) — into columns of line, given that
// name is line's suffix: every File Picker row ends with the file name
// (pickerFileRow's rel always does). If it is not the suffix, nil is
// returned and the row simply renders unhighlighted, rather than emphasis
// landing on whatever characters happen to sit at those columns: a row
// missing its emphasis is a cosmetic loss, emphasis on the wrong characters
// is a lie about what matched.
func suffixColumns(line, name string, positions []int) []int {
	if len(positions) == 0 || !strings.HasSuffix(line, name) {
		return nil
	}
	offset := len([]rune(line)) - len([]rune(name))
	columns := make([]int, 0, len(positions))
	for _, p := range positions {
		columns = append(columns, offset+p)
	}
	return columns
}

// clipColumns keeps the columns that are still visible in a row truncated to
// visible cells (pickerGrepRow's rightTruncate), dropping the rest: a long
// matched line can have its match cut off entirely, or cut through the
// middle, and only the part still on screen may be emphasized.
func clipColumns(columns []int, visible int) []int {
	var kept []int
	for _, c := range columns {
		if c >= 0 && c < visible {
			kept = append(kept, c)
		}
	}
	return kept
}

// highlightMatches bolds line's runes at the given columns (rune indexes
// into line itself), leaving every other character alone — S5's match
// emphasis, shared by both modes' rows (M6.1's fuzzy-matched file-name
// characters via suffixColumns, M6.2's literal query occurrences via
// grepMatchPositions).
//
// A matched character is picked out with bold, and bold only. The emphasis
// has to read both on a plain row and on the selected row's reverse-video
// bar (pickerSelectedStyle), which rules out a foreground colour on its own:
// reverse video swaps foreground and background, so a colour chosen to stand
// out against the pane's background lands on the selected row's *text*
// colour and can vanish into it. Bold is an attribute, so it composes with
// reverse instead of competing with it — which is also why the highlight is
// derived from the row's own base style (base.Bold(true)) rather than being
// a standalone style that would drop the selection's reverse.
//
// Every run is rendered with its style fully specified, including the
// unmatched runs and including the selection's own reverse video, rather
// than styling only the matched runs and letting styleRow's outer
// pickerSelectedStyle wrap the result. Nesting is what forces that: lipgloss
// closes an inner style with a full reset (`ESC[m`, verified against v2.0.5),
// which also cancels the enclosing reverse video, so an outer-wrapped
// selected row would lose its bar from the first highlight to the end of the
// text. Self-contained runs have nothing to cancel. styleRow still applies
// the same base style outside for the row's padding and width clamp.
func highlightMatches(line string, columns []int, selected bool) string {
	if len(columns) == 0 {
		return line
	}
	runes := []rune(line)
	matched := make(map[int]bool, len(columns))
	for _, c := range columns {
		matched[c] = true
	}
	base := lipgloss.NewStyle()
	if selected {
		base = pickerSelectedStyle
	}
	highlight := base.Bold(true)

	var b strings.Builder
	for i := 0; i < len(runes); {
		j := i
		for j < len(runes) && matched[j] == matched[i] {
			j++
		}
		style := base
		if matched[i] {
			style = highlight
		}
		b.WriteString(style.Render(string(runes[i:j])))
		i = j
	}
	return b.String()
}

// styleRow applies the Results list's shared row styling — reverse video
// when selected (pickerSelectedStyle), plain otherwise — and clamps to
// exactly width, the four-line `.Width(width).MaxWidth(width).Render(line)`
// tail pickerFileRow and pickerGrepRow both ended with verbatim (TRIAGE
// T21/D55) before this extraction, and M6's own row renderer is an
// anticipated third caller. Each caller is responsible for shaping/
// truncating line itself first (pickerFileRow via pickerRootLabel's budget
// math, pickerGrepRow via rightTruncate) — styleRow only applies the
// selection style and the width clamp common to both, the same split
// pickerGrepRow's own doc comment already draws between shaping and
// styling.
func styleRow(line string, width int, selected bool) string {
	style := lipgloss.NewStyle()
	if selected {
		style = pickerSelectedStyle
	}
	return style.Width(width).MaxWidth(width).Render(line)
}

// rightTruncate keeps s's prefix, replacing the tail with an ellipsis when
// it does not fit in width — the mirror of leftTruncate (rootsnav.go, which
// keeps a Root's suffix instead). Grep Picker rows (pickerGrepRow, M4.1) use
// this rather than leftTruncate: the file/line-number prefix identifies the
// row, and the match itself is more likely to sit near the start of a long
// line than its end, so the prefix is what's worth keeping visible.
func rightTruncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	if width == 1 {
		return "…"
	}
	runes := []rune(s)
	keep := width - 1
	if keep > len(runes) {
		keep = len(runes)
	}
	return string(runes[:keep]) + "…"
}

// pickerGrepRow renders one Grep Picker Result (S4, M4.1): the file's path
// relative to its Root, the matching line's 1-based line number (rg's own
// count — D34 records why the Document Viewer hand-off cannot use it
// directly, but the Preview/Results side of the Picker always addresses the
// raw file, so rg's number is exactly right here) and the line's own text.
// Unlike pickerFileRow's Root label (M3.1a/D50), this does not disambiguate
// same-named Roots — S4 only asks for "the file, the line number and the
// line", not Root disambiguation, and the line text itself is almost always
// enough to tell two matches apart. A Starred-scope orphan (S8, C4/M5.1)
// falls out of the same filepath.Rel call for free: g.file.root is the zero
// value (root.path == ""), so filepath.Rel("", g.file.path) errors and rel
// falls back to g.file.path — the absolute path, which is exactly what an
// orphan (no Root to be relative to) should show here; no orphan-specific
// branch is needed the way pickerFileRow's own rel computation needs one,
// since this row never tries a name+parent Root label in the first place.
//
// The assembled line is right-truncated (rightTruncate) to width *before*
// styling, unlike pickerFileRow's budget-computed-to-fit construction:
// pickerFileRow's own text is bounded by construction (pickerRootLabel's
// budget math already leaves room for it), but a matched line's text is
// arbitrary length, so it must be capped explicitly. This matters because
// lipgloss's Style.Width word-wraps first and only then applies MaxWidth
// (style.go) — relying on Width+MaxWidth alone the way pickerFileRow does
// would silently turn one long match into a wrapped multi-line row, growing
// the Results list past "one row per match" (M4.1's own acceptance
// criterion: truncate to the pane, don't wrap the layout). Truncating first
// means the string handed to Width is already <= width, so its word-wrap is
// a no-op and only padding/selection styling happens there.
// query is what rg matched to produce this row, needed only to highlight the
// matched text within the line (S5/M6.2) — an empty query renders exactly
// the unhighlighted row M4.1 shipped. The match is located within g.text
// only, offset by the prefix's width, never searched for in the row as a
// whole: a query like "md" or "3" occurs in the file/line-number prefix of
// plenty of rows without that being what matched.
func pickerGrepRow(g grepResult, query string, width int, selected bool, starred bool) string {
	marker := starMarker(starred)
	contentWidth := max(width-lipgloss.Width(marker), 0)
	rel, err := filepath.Rel(g.file.root.path, g.file.path)
	if err != nil {
		rel = g.file.path
	}
	prefix := fmt.Sprintf("%s:%d  ", rel, g.line)
	full := prefix + g.text
	line := rightTruncate(full, contentWidth)

	// A truncated row's last cell is rightTruncate's own ellipsis, not the
	// line's text, so it is never part of a match — hence visible stops one
	// short of the rendered row there.
	visible := len([]rune(line))
	if line != full {
		visible = max(visible-1, 0)
	}
	offset := len([]rune(prefix))
	var columns []int
	for _, p := range grepMatchPositions(g.text, query) {
		columns = append(columns, offset+p)
	}
	clipped := clipColumns(columns, visible)
	marked, shifted := prependMarker(marker, line, clipped)
	return styleRow(highlightMatches(marked, shifted, selected), width, selected)
}

// pickerWikipediaRow renders one Wikipedia Picker Result (S2, C6/M3.3/
// M3.4): the article's title, plus its short description when Wikipedia
// supplies one, with a star marker (S2's "the Picker's star key works on a
// Wikipedia Result exactly as on a file Result") — its own row shape,
// unlike pickerFileRow/pickerGrepRow: there is no Root or path to show (an
// article has neither) and no match highlighting (S2 does not ask for one
// — Wikipedia's own title-search ranking is trusted as-is, unlike the
// fuzzy/literal matching this codebase does itself for File/Grep). With no
// per-character highlighting to keep separately styled (unlike
// pickerFileRow/pickerGrepRow's prependMarker+highlightMatches split), the
// marker is simply prepended to the plain line before the single
// whole-row style styleRow applies — nothing here risks the partial-style
// reverse-video hazard prependMarker's own doc comment warns about.
func pickerWikipediaRow(r wikipedia.SearchResult, width int, selected bool, starred bool) string {
	marker := starMarker(starred)
	line := r.Title
	if r.Description != "" {
		line += "  " + r.Description
	}
	contentWidth := max(width-lipgloss.Width(marker), 0)
	line = rightTruncate(line, contentWidth)
	return styleRow(marker+line, width, selected)
}

// pickerResultsContent renders the Results list's content (M3.1, M4.1),
// sized to exactly width x height — the same "expected to be exactly
// r.content() in size" contract renderPane's own doc comment gives every
// frame pane. Each mode renders its own row shape (pickerFileRow/
// pickerGrepRow); a result set that does not currently answer the
// Prompt/mode/Scope (p.resultsCurrent) is treated the same as none at all
// for either mode — pickerStatus is already saying "searching" for that
// same mismatch, and showing the previous mode's stale rows underneath that
// caption would contradict it.
func pickerResultsContent(p picker, width, height int, hideDotfiles bool, docs map[string]docRecord) string {
	if width <= 0 || height <= 0 || !p.resultsCurrent(hideDotfiles) {
		return ""
	}
	if p.mode == grepMode {
		grep := p.results.grep
		if len(grep) == 0 {
			return ""
		}
		return windowedView(width, height, p.selected, len(grep), func(idx int) string {
			g := grep[idx]
			return pickerGrepRow(g, p.results.query, width, idx == p.selected, docs[filepath.Clean(g.file.path)].Starred)
		})
	}
	if p.mode == wikipediaMode {
		results := p.results.wikipedia
		if len(results) == 0 {
			return ""
		}
		return windowedView(width, height, p.selected, len(results), func(idx int) string {
			r := results[idx]
			return pickerWikipediaRow(r, width, idx == p.selected, docs[wikipedia.ArticleURL(r.Title)].Starred)
		})
	}
	files := p.results.files
	if len(files) == 0 {
		return ""
	}
	return windowedView(width, height, p.selected, len(files), func(idx int) string {
		f := files[idx]
		return pickerFileRow(f, p.results.query, width, idx == p.selected, docs[filepath.Clean(f.path)].Starred)
	})
}

// pickerPreviewContent is the Preview pane's own content (M3.2, S2/C6/M3.3):
// the Preview viewport's rendered view for File/Grep mode, or Wikipedia
// mode's short placeholder — S2's "There is no Preview: the Preview pane
// shows a short placeholder instead of article content" — regardless of the
// selection or of whatever the viewport happens to still hold from a
// previous File/Grep mode search (syncPreview never targets wikipediaMode,
// see pickerPreviewTarget's own doc comment).
func pickerPreviewContent(p picker) string {
	if p.mode == wikipediaMode {
		return pickerWikipediaPreviewPlaceholder
	}
	return p.preview.View()
}

// renderPicker composes the Picker modal's three regions (S1): a bordered
// Prompt spanning the modal's width — titled with the mode/Scope indicator
// and captioned with the status line (M2.5) — with Results and Preview side
// by side beneath it. Prompt uses focusedBorderStyle — it always owns the
// keyboard while the Picker is open, the same accent renderPane gives
// whichever frame pane currently has focus — Results/Preview use the plain
// borderStyle, same as an unfocused frame pane. Results now carries the File
// Picker's rows (M3.1); Preview still carries none (M3.2 fills it).
func renderPicker(p picker, l pickerLayout, avail ripgrepAvailability, hideDotfiles bool, docs map[string]docRecord) string {
	promptContent := p.promptView(l.prompt.content().width)
	if status := pickerStatus(p, avail, hideDotfiles); status != "" {
		promptContent += "\n" + status
	}
	prompt := renderBox(focusedBorderStyle, l.prompt, titleStyle.Render(pickerModeScope(p, hideDotfiles)), promptContent)
	resultsContent := pickerResultsContent(p, l.results.content().width, l.results.content().height, hideDotfiles, docs)
	results := renderBox(borderStyle, l.results, titleStyle.Render("Results"), resultsContent)
	preview := renderBox(borderStyle, l.preview, titleStyle.Render("Preview"), pickerPreviewContent(p))
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
func overlayPicker(frame string, width, height int, p picker, avail ripgrepAvailability, hideDotfiles bool, docs map[string]docRecord) string {
	l := computePickerLayout(width, height)
	box := lipgloss.NewStyle().MaxWidth(width).MaxHeight(height).Render(renderPicker(p, l, avail, hideDotfiles, docs))
	bw, bh := lipgloss.Width(box), lipgloss.Height(box)
	x := max((width-bw)/2, 0)
	y := max((height-bh)/2, 0)

	base := lipgloss.NewLayer(frame).X(0).Y(0).Z(0)
	modal := lipgloss.NewLayer(box).X(x).Y(y).Z(1)
	return lipgloss.NewCompositor(base, modal).Render()
}
