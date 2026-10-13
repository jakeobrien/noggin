package main

import (
	uv "github.com/charmbracelet/ultraviolet"
)

// paintCanvas makes the frame's colours total: every cell that no style chose a
// foreground or background for gets the Theme's Fg / Bg, and every cell the
// frame left blank becomes an explicit Bg-coloured space. After it, no cell in
// the frame falls back to the terminal's own default colours, which is what
// makes the Theme opinionated rather than a suggestion layered on the user's.
//
// Doing this as a pass over the finished frame, rather than by giving every
// style a Background, is deliberate. A background set per style is lost at the
// first SGR reset, and lipgloss resets all over the place — nested styles,
// JoinHorizontal's gap padding, the viewport's line fill, Glamour's output,
// bubbles' text input. A cell-level pass is immune to all of that, and
// (unlike regexing SGR sequences) it understands wide cells and hyperlinks
// because it parses with the same engine Bubble Tea itself renders with.
//
// Reverse-video cells need no special care: they get the same Fg/Bg as any
// other, and the terminal swaps them, so a reversed row is Bg-on-Fg.
//
// width and height are the frame's cell size; a frame that is not positively
// sized is returned untouched. Content outside it would be clipped by the
// terminal anyway.
func paintCanvas(frame string, width, height int, t Theme) string {
	if width <= 0 || height <= 0 {
		return frame
	}
	// ScreenBuffer's default width method (WcWidth) is Bubble Tea's own default
	// until it learns the terminal does grapheme clustering (mode 2027); the two
	// differ only for exotic emoji sequences.
	buf := uv.NewScreenBuffer(width, height)
	uv.NewStyledString(frame).Draw(buf, buf.Bounds())
	for _, line := range buf.Lines {
		for x := range line {
			c := &line[x]
			if c.IsZero() {
				// The continuation column of a wide cell; it takes the wide
				// cell's style.
				continue
			}
			if c.Style.Fg == nil {
				c.Style.Fg = t.Fg
			}
			if c.Style.Bg == nil {
				c.Style.Bg = t.Bg
			}
		}
	}
	return buf.Render()
}
