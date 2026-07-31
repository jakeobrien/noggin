package main

// Frame proportions, in one place so the layout can be retuned without reading
// the render code. The frame is: a Navigator column (Roots Nav stacked above
// Files Nav) beside the Document Viewer, with a Footer line beneath both.
const (
	footerHeight = 1

	// A pane is a border box with a title on its first content row; both eat
	// into the space its contents get.
	paneBorderWidth = 1
	paneTitleHeight = 1

	// navigatorPercent is the Navigator column's share of the terminal width,
	// clamped to the min/max below; rootsPercent is the Roots Nav pane's share
	// of the Navigator column's height.
	navigatorPercent = 30
	rootsPercent     = 35

	minNavigatorWidth = 18
	maxNavigatorWidth = 40
	minViewerWidth    = 20
	minPaneHeight     = 3
)

// rect is a region's outer size in cells (borders included).
type rect struct {
	width  int
	height int
}

func (r rect) empty() bool { return r.width <= 0 || r.height <= 0 }

// layout is the frame's regions at a given terminal size.
type layout struct {
	roots  rect
	files  rect
	viewer rect
	footer rect
}

// navigatorWidth is the width of the Navigator column, 0 when the terminal is
// too narrow to carry one.
func (l layout) navigatorWidth() int { return l.roots.width }

// content is the box a pane's contents get: its region less border and title.
func (r rect) content() rect {
	return rect{
		width:  max(r.width-2*paneBorderWidth, 0),
		height: max(r.height-2*paneBorderWidth-paneTitleHeight, 0),
	}
}

// computeLayout divides a terminal of width x height into the frame's regions.
// A terminal too narrow for both columns drops the Navigator rather than
// rendering two unusable slivers.
func computeLayout(width, height int) layout {
	if width <= 0 || height <= 0 {
		return layout{}
	}

	body := height - footerHeight
	if body < 0 {
		body = 0
	}

	navWidth := clamp(width*navigatorPercent/100, minNavigatorWidth, maxNavigatorWidth)
	if width-navWidth < minViewerWidth {
		navWidth = 0
	}

	rootsHeight, filesHeight := splitHeight(body)
	if navWidth == 0 {
		rootsHeight, filesHeight = 0, 0
	}

	return layout{
		roots:  rect{width: navWidth, height: rootsHeight},
		files:  rect{width: navWidth, height: filesHeight},
		viewer: rect{width: width - navWidth, height: body},
		footer: rect{width: width, height: min(footerHeight, height)},
	}
}

// splitHeight divides the Navigator column between its two panes, falling back
// to an even split when the proportional share would starve either pane.
func splitHeight(total int) (top, bottom int) {
	if total <= 0 {
		return 0, 0
	}
	top = total * rootsPercent / 100
	if top < minPaneHeight || total-top < minPaneHeight {
		top = total / 2
	}
	return top, total - top
}

func clamp(v, lo, hi int) int {
	return min(max(v, lo), hi)
}
