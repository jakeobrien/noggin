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
var theme = nightTheme

// nightTheme is the only shipped Theme: a cool, low-glare dark palette (after
// Tokyo Night). Foreground tokens keep at least WCAG AA contrast (4.5:1)
// against Bg, and syntax tokens against Markdown.CodeBg — theme_test.go
// enforces both for any Theme in allThemes.
var nightTheme = Theme{
	Bg:        lipgloss.Color("#1a1b26"),
	Fg:        lipgloss.Color("#c0caf5"),
	Subtle:    lipgloss.Color("#8a93bd"),
	Border:    lipgloss.Color("#6a7399"),
	Accent:    lipgloss.Color("#bb9af7"),
	Highlight: lipgloss.Color("#e0af68"),

	UnfocusedBar:    lipgloss.Color("#2f334d"),
	UnfocusedText:   lipgloss.Color("#9aa5ce"),
	UnfocusedSubtle: lipgloss.Color("#7680a8"),

	Markdown: MarkdownTheme{
		H1Fg:      lipgloss.Color("#1a1b26"),
		H1Bg:      lipgloss.Color("#7aa2f7"),
		Heading:   lipgloss.Color("#7aa2f7"),
		H2:        lipgloss.Color("#ff9e64"),
		H3:        lipgloss.Color("#9ece6a"),
		H4:        lipgloss.Color("#7dcfff"),
		H5:        lipgloss.Color("#bb9af7"),
		H6:        lipgloss.Color("#e0af68"),
		Emph:      lipgloss.Color("#73daca"),
		Strong:    lipgloss.Color("#73daca"),
		Link:      lipgloss.Color("#7dcfff"),
		LinkText:  lipgloss.Color("#9ece6a"),
		Image:     lipgloss.Color("#f7768e"),
		ImageText: lipgloss.Color("#8a93bd"),
		Code:      lipgloss.Color("#ff9e64"),
		CodeBg:    lipgloss.Color("#292e42"),
		Rule:      lipgloss.Color("#6a7399"),
	},
	Syntax: SyntaxTheme{
		Text:        lipgloss.Color("#c0caf5"),
		Comment:     lipgloss.Color("#9099c0"),
		Preproc:     lipgloss.Color("#ff9e64"),
		Keyword:     lipgloss.Color("#bb9af7"),
		Type:        lipgloss.Color("#2ac3de"),
		Function:    lipgloss.Color("#7aa2f7"),
		Builtin:     lipgloss.Color("#7dcfff"),
		String:      lipgloss.Color("#9ece6a"),
		Escape:      lipgloss.Color("#89ddff"),
		Number:      lipgloss.Color("#ff9e64"),
		Operator:    lipgloss.Color("#89ddff"),
		Punctuation: lipgloss.Color("#a9b1d6"),
		Tag:         lipgloss.Color("#f7768e"),
		Attribute:   lipgloss.Color("#e0af68"),
		Inserted:    lipgloss.Color("#9ece6a"),
		Deleted:     lipgloss.Color("#f7768e"),
		ErrorFg:     lipgloss.Color("#1a1b26"),
		ErrorBg:     lipgloss.Color("#f7768e"),
	},
}

// allThemes is every shipped Theme, for the tests that hold each to the same
// contrast bar.
var allThemes = map[string]Theme{"night": nightTheme}

// hexOf is c as "#rrggbb", the form Glamour's stylesheets take.
func hexOf(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", uint8(r>>8), uint8(g>>8), uint8(b>>8))
}
