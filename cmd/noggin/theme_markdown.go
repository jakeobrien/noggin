package main

import (
	"encoding/json"
	"fmt"
	"image/color"

	"charm.land/glamour/v2/ansi"
	"charm.land/lipgloss/v2"
)

// markdownStyleConfig is the Glamour stylesheet for Theme t: markdown's
// structure (margins, bullets, heading prefixes, ...) fixed here, every colour
// taken from t.Markdown and t.Syntax. It replaces the hand-edited
// styles/dark.json and light.json, so a colour is changed in one place — the
// Theme — and a new Theme restyles the document with no stylesheet to write.
func markdownStyleConfig(t Theme) ansi.StyleConfig {
	md, sx := t.Markdown, t.Syntax

	// fg is a foreground-only primitive; Glamour wants hex strings.
	fg := func(c color.Color) ansi.StylePrimitive {
		return ansi.StylePrimitive{Color: ptr(hexOf(c))}
	}
	bold := func(p ansi.StylePrimitive) ansi.StylePrimitive { p.Bold = ptr(true); return p }
	underline := func(p ansi.StylePrimitive) ansi.StylePrimitive { p.Underline = ptr(true); return p }
	prefixed := func(p ansi.StyleBlock, prefix string) ansi.StyleBlock { p.Prefix = prefix; return p }
	block := func(p ansi.StylePrimitive) ansi.StyleBlock { return ansi.StyleBlock{StylePrimitive: p} }

	h1 := ansi.StylePrimitive{
		Prefix:          " ",
		Suffix:          " ",
		Color:           ptr(hexOf(md.H1Fg)),
		BackgroundColor: ptr(hexOf(md.H1Bg)),
		Bold:            ptr(true),
	}
	h6 := fg(md.H6)
	h6.Bold = ptr(false)

	cfg := ansi.StyleConfig{
		Document: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{BlockPrefix: "\n", BlockSuffix: "\n", Color: ptr(hexOf(t.Fg))},
			Margin:         ptr(uint(2)),
		},
		BlockQuote: ansi.StyleBlock{Indent: ptr(uint(1)), IndentToken: ptr("│ ")},
		List:       ansi.StyleList{LevelIndent: 2},

		Heading: block(bold(ansi.StylePrimitive{BlockSuffix: "\n", Color: ptr(hexOf(md.Heading))})),
		H1:      block(h1),
		H2:      block(underline(fg(md.H2))),
		H3:      prefixed(block(fg(md.H3)), "### "),
		H4:      prefixed(block(fg(md.H4)), "#### "),
		H5:      prefixed(block(fg(md.H5)), "##### "),
		H6:      prefixed(block(h6), "###### "),

		Strikethrough: ansi.StylePrimitive{CrossedOut: ptr(true)},
		Emph:          func() ansi.StylePrimitive { p := fg(md.Emph); p.Italic = ptr(true); return p }(),
		Strong:        bold(fg(md.Strong)),
		HorizontalRule: func() ansi.StylePrimitive {
			p := fg(md.Rule)
			p.Format = "\n--------\n"
			return p
		}(),

		Item:        ansi.StylePrimitive{BlockPrefix: "• "},
		Enumeration: ansi.StylePrimitive{BlockPrefix: ". "},
		Task:        ansi.StyleTask{Ticked: "[✓] ", Unticked: "[ ] "},

		Link:     underline(fg(md.Link)),
		LinkText: bold(fg(md.LinkText)),

		Image: underline(fg(md.Image)),
		ImageText: func() ansi.StylePrimitive {
			p := fg(md.ImageText)
			p.Format = "Image: {{.text}} →"
			return p
		}(),

		// Inline code. The non-breaking spaces pad the chip without being
		// trimmed as whitespace.
		Code: block(ansi.StylePrimitive{
			Prefix:          "\u00a0",
			Suffix:          "\u00a0",
			Color:           ptr(hexOf(md.Code)),
			BackgroundColor: ptr(hexOf(md.CodeBg)),
		}),

		DefinitionDescription: ansi.StylePrimitive{BlockPrefix: "\n🠶 "},
	}

	cfg.CodeBlock = ansi.StyleCodeBlock{
		StyleBlock: ansi.StyleBlock{
			StylePrimitive: fg(sx.Text),
			Margin:         ptr(uint(2)),
		},
		Chroma: chromaStyle(t),
	}
	return cfg
}

// chromaStyle maps Chroma's many token types onto the Theme's few SyntaxTheme
// slots.
func chromaStyle(t Theme) *ansi.Chroma {
	sx := t.Syntax
	fg := func(c color.Color) ansi.StylePrimitive {
		return ansi.StylePrimitive{Color: ptr(hexOf(c))}
	}
	return &ansi.Chroma{
		Text: fg(sx.Text),
		Error: ansi.StylePrimitive{
			Color:           ptr(hexOf(sx.ErrorFg)),
			BackgroundColor: ptr(hexOf(sx.ErrorBg)),
		},
		Comment:             fg(sx.Comment),
		CommentPreproc:      fg(sx.Preproc),
		Keyword:             fg(sx.Keyword),
		KeywordReserved:     fg(sx.Keyword),
		KeywordNamespace:    fg(sx.Keyword),
		KeywordType:         fg(sx.Type),
		Operator:            fg(sx.Operator),
		Punctuation:         fg(sx.Punctuation),
		Name:                fg(sx.Text),
		NameBuiltin:         fg(sx.Builtin),
		NameTag:             fg(sx.Tag),
		NameAttribute:       fg(sx.Attribute),
		NameClass:           func() ansi.StylePrimitive { p := fg(sx.Type); p.Bold = ptr(true); return p }(),
		NameConstant:        fg(sx.Number),
		NameDecorator:       fg(sx.Attribute),
		NameFunction:        fg(sx.Function),
		LiteralNumber:       fg(sx.Number),
		LiteralString:       fg(sx.String),
		LiteralStringEscape: fg(sx.Escape),
		GenericDeleted:      fg(sx.Deleted),
		GenericEmph:         ansi.StylePrimitive{Italic: ptr(true)},
		GenericInserted:     fg(sx.Inserted),
		GenericStrong:       ansi.StylePrimitive{Bold: ptr(true)},
		GenericSubheading:   fg(sx.Comment),
		Background:          ansi.StylePrimitive{BackgroundColor: ptr(hexOf(t.Markdown.CodeBg))},
	}
}

func ptr[T any](v T) *T { return &v }

// mutedMarkdownStyleConfig is markdownStyleConfig(t) with every colour
// desaturated and pulled toward t.Bg: what an unfocused Document Viewer shows
// (unfocused_style.go). Derived from the finished stylesheet, not stored
// beside it, so it can never drift from the focused render.
func mutedMarkdownStyleConfig(t Theme) (ansi.StyleConfig, error) {
	var doc any
	src, err := json.Marshal(markdownStyleConfig(t))
	if err != nil {
		return ansi.StyleConfig{}, fmt.Errorf("muted style: %w", err)
	}
	if err := json.Unmarshal(src, &doc); err != nil {
		return ansi.StyleConfig{}, fmt.Errorf("muted style: %w", err)
	}
	muteColors(doc, t.Bg, t.MutedFade)
	src, err = json.Marshal(doc)
	if err != nil {
		return ansi.StyleConfig{}, fmt.Errorf("muted style: %w", err)
	}
	var cfg ansi.StyleConfig
	if err := json.Unmarshal(src, &cfg); err != nil {
		return ansi.StyleConfig{}, fmt.Errorf("muted style: %w", err)
	}
	return cfg, nil
}

// muteColors rewrites every "color"/"background_color" string in a decoded
// stylesheet with mutedColor, leaving every other field alone.
func muteColors(v any, bg color.Color, fade float64) {
	switch t := v.(type) {
	case map[string]any:
		for k, val := range t {
			if s, ok := val.(string); ok && (k == "color" || k == "background_color") && s != "" {
				t[k] = mutedColor(s, bg, fade)
				continue
			}
			muteColors(val, bg, fade)
		}
	case []any:
		for _, val := range t {
			muteColors(val, bg, fade)
		}
	}
}

// mutedColor desaturates c halfway toward its own luma, then blends it the
// fraction fade (Theme.MutedFade) toward bg — darker on a dark theme, lighter
// on a light one, with no light/dark branch because bg already says which way
// "back" is.
func mutedColor(c string, bg color.Color, fade float64) string {
	r16, g16, b16, _ := lipgloss.Color(c).RGBA()
	br16, bg16, bb16, _ := bg.RGBA()
	r, g, b := float64(r16>>8), float64(g16>>8), float64(b16>>8)
	luma := 0.299*r + 0.587*g + 0.114*b
	const desat = 0.5
	r, g, b = r+(luma-r)*desat, g+(luma-g)*desat, b+(luma-b)*desat
	mix := func(x float64, back uint32) uint8 {
		return uint8(x + (float64(back>>8)-x)*fade)
	}
	return fmt.Sprintf("#%02x%02x%02x", mix(r, br16), mix(g, bg16), mix(b, bb16))
}
