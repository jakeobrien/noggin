package main

import (
	"fmt"
	"image/color"
)

// Theme is noggin's whole colour scheme: one struct of semantic tokens that
// every style in the app is derived from. noggin is deliberately opinionated
// about its look and does not inherit the terminal's palette: every
// cell is painted from this struct (paintCanvas, canvas.go), so contrast is a
// property of the Theme and is checked by theme_test.go, not a gamble on the
// user's terminal colours. There is consequently no light/dark detection and no
// terminal query.
//
// Colours are truecolor. Bubble Tea downsamples them to the terminal's own
// colour profile at write time, so a 256-colour terminal gets the nearest
// approximation without any code here.
//
// Themes are hot-swappable: keys.Theme cycles `theme` through `themes`
// (model.cycleTheme; why a global, ADR-0009). Every style is therefore built from `theme` at the moment
// it is used, never captured in a package-level var at startup, and the one
// thing rendered ahead of time — a Document Viewer's content — is re-rendered on
// a switch (viewer.restyle).
//
// To add a look, write another Theme value and append it to `themes`. Nothing
// else names a colour.
type Theme struct {
	// Name identifies the Theme in the Footer when it is switched to.
	Name string

	// Bg and Fg are the canvas: the colours of every cell nothing else styles.
	Bg, Fg color.Color

	// Subtle is secondary text: empty-state messages, paths, the footer.
	Subtle color.Color
	// Border frames an unfocused pane or box.
	Border color.Color
	// Accent marks focus: the focused pane's border and title bar. It
	// also colours the keys in the Footer's hints.
	Accent color.Color
	// Highlight asks for attention: the staleness row and the Help Modal's
	// border. A colour no pane border uses, so it cannot be mistaken for focus.
	Highlight color.Color

	// UnfocusedBar is the selected row of an unfocused nav pane. It is a
	// background, so the row recedes instead of competing with the focused
	// pane's reverse-video bar. UnfocusedText and UnfocusedSubtle are the
	// foregrounds of that pane's other rows: its normal text and its
	// already-subtle text.
	UnfocusedBar, UnfocusedText, UnfocusedSubtle color.Color
	// MutedFade is how far an unfocused Document Viewer's colours are pulled
	// toward Bg (mutedColor): 0 leaves them as focused, 1 erases them. Set per
	// Theme because the same fraction dims a low-contrast palette more than a
	// high-contrast one; each is tuned so muted body text keeps comfortable
	// contrast (theme_test.go holds it to WCAG AA) yet still visibly recedes.
	MutedFade float64

	Markdown MarkdownTheme
	Syntax   SyntaxTheme
}

// MarkdownTheme colours rendered markdown (markdownStyleConfig).
type MarkdownTheme struct {
	// H1 is a chip: H1Fg text on an H1Bg bar. H2–H6 are plain foregrounds.
	H1Fg, H1Bg         color.Color
	H2, H3, H4, H5, H6 color.Color
	Heading            color.Color // the shared base of all headings
	Emph, Strong       color.Color
	Link, LinkText     color.Color
	Image, ImageText   color.Color
	Code               color.Color // inline code text
	CodeBg             color.Color // inline code and fenced code blocks
	Rule               color.Color // horizontal rule
}

// SyntaxTheme colours fenced code (Chroma). Several Chroma token types share a
// slot; markdownStyleConfig has the mapping.
type SyntaxTheme struct {
	Text, Comment, Preproc, Keyword, Type, Function, Builtin color.Color
	String, Escape, Number, Operator, Punctuation            color.Color
	Tag, Attribute, Inserted, Deleted                        color.Color
	ErrorFg, ErrorBg                                         color.Color
}

// themes is every shipped Theme, in the order keys.Theme cycles them. The first
// is the startup default. Foreground tokens keep at least WCAG AA contrast
// (4.5:1) against Bg, and syntax tokens against Markdown.CodeBg —
// theme_test.go enforces both for every Theme here.
var themes = []Theme{gruvboxMaterialTheme, tokyoNightTheme, everforestTheme, charmTheme}

// theme is the active Theme: process-wide, like the terminal it paints.
// model.cycleTheme is the one production writer; a test that switches it
// restores it (keepTheme, theme_switch_test.go).
var theme = themes[0]

// nextTheme is the Theme after the one named name in the cycle, wrapping
// around; the first Theme if name is not one of them.
func nextTheme(name string) Theme {
	for i, t := range themes {
		if t.Name == name {
			return themes[(i+1)%len(themes)]
		}
	}
	return themes[0]
}

// hexOf is c as "#rrggbb", the form Glamour's stylesheets take.
func hexOf(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", uint8(r>>8), uint8(g>>8), uint8(b>>8))
}
