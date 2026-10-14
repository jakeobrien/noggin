package main

import (
	"fmt"
	"image/color"

	"charm.land/lipgloss/v2"
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
// To add a look, write another Theme value and point `theme` at it. Nothing
// else names a colour.
type Theme struct {
	// Bg and Fg are the canvas: the colours of every cell nothing else styles.
	Bg, Fg color.Color

	// Subtle is secondary text: empty-state messages, paths, the footer.
	Subtle color.Color
	// Border frames an unfocused pane or box.
	Border color.Color
	// Accent marks focus: the focused pane's border and title bar.
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

// theme is the active Theme. A package var so a test can swap it; production
// code never reassigns it.
var theme = gruvboxMaterialTheme

// gruvboxMaterialTheme is the only shipped Theme: a warm, low-glare dark
// palette after Gruvbox Material (dark, medium contrast). Foreground tokens
// keep at least WCAG AA contrast (4.5:1) against Bg, and syntax tokens against
// Markdown.CodeBg — theme_test.go enforces both for any Theme in allThemes.
var gruvboxMaterialTheme = Theme{
	Bg:        lipgloss.Color("#282828"),
	Fg:        lipgloss.Color("#d4be98"),
	Subtle:    lipgloss.Color("#a89984"),
	Border:    lipgloss.Color("#7c6f64"),
	Accent:    lipgloss.Color("#a9b665"),
	Highlight: lipgloss.Color("#d8a657"),

	UnfocusedBar:    lipgloss.Color("#45403d"),
	UnfocusedText:   lipgloss.Color("#c5b18a"),
	UnfocusedSubtle: lipgloss.Color("#928374"),

	Markdown: MarkdownTheme{
		H1Fg:      lipgloss.Color("#282828"),
		H1Bg:      lipgloss.Color("#d8a657"),
		Heading:   lipgloss.Color("#ea6962"),
		H2:        lipgloss.Color("#ea6962"),
		H3:        lipgloss.Color("#e78a4e"),
		H4:        lipgloss.Color("#a9b665"),
		H5:        lipgloss.Color("#7daea3"),
		H6:        lipgloss.Color("#d3869b"),
		Emph:      lipgloss.Color("#d3869b"),
		Strong:    lipgloss.Color("#e78a4e"),
		Link:      lipgloss.Color("#7daea3"),
		LinkText:  lipgloss.Color("#89b482"),
		Image:     lipgloss.Color("#d3869b"),
		ImageText: lipgloss.Color("#a89984"),
		Code:      lipgloss.Color("#89b482"),
		CodeBg:    lipgloss.Color("#32302f"),
		Rule:      lipgloss.Color("#7c6f64"),
	},
	Syntax: SyntaxTheme{
		Text:        lipgloss.Color("#d4be98"),
		Comment:     lipgloss.Color("#a89984"),
		Preproc:     lipgloss.Color("#d3869b"),
		Keyword:     lipgloss.Color("#ef7770"),
		Type:        lipgloss.Color("#d8a657"),
		Function:    lipgloss.Color("#a9b665"),
		Builtin:     lipgloss.Color("#89b482"),
		String:      lipgloss.Color("#a9b665"),
		Escape:      lipgloss.Color("#89b482"),
		Number:      lipgloss.Color("#d3869b"),
		Operator:    lipgloss.Color("#e78a4e"),
		Punctuation: lipgloss.Color("#a89984"),
		Tag:         lipgloss.Color("#ef7770"),
		Attribute:   lipgloss.Color("#d8a657"),
		Inserted:    lipgloss.Color("#a9b665"),
		Deleted:     lipgloss.Color("#ef7770"),
		ErrorFg:     lipgloss.Color("#282828"),
		ErrorBg:     lipgloss.Color("#ea6962"),
	},
}

// allThemes is every shipped Theme, for the tests that hold each to the same
// contrast bar.
var allThemes = map[string]Theme{"gruvbox-material": gruvboxMaterialTheme}

// hexOf is c as "#rrggbb", the form Glamour's stylesheets take.
func hexOf(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", uint8(r>>8), uint8(g>>8), uint8(b>>8))
}
