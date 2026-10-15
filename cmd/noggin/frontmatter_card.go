package main

import (
	"image/color"
	"math"
	"strings"

	"charm.land/lipgloss/v2"
	yaml "go.yaml.in/yaml/v3"
)

// The Frontmatter Card (D138): a document's YAML frontmatter drawn as a
// bordered key/value box above its rendered body, instead of as raw YAML.
// Frontmatter the card cannot show faithfully (nested values, TOML, invalid
// YAML, too narrow a pane) falls back to the verbatim code block
// (fenceFrontmatter), so nothing in it is ever hidden.

// frontmatterField is one top-level frontmatter entry as the card shows it.
type frontmatterField struct{ key, value string }

// frontmatterFields parses YAML frontmatter content into the card's fields, in
// source order (hence yaml.Node rather than a map). ok is false for anything
// the card cannot show faithfully: invalid YAML, a document that is not a
// mapping, an empty one, or any value that is not a scalar or a list of
// scalars.
func frontmatterFields(content []byte) (fields []frontmatterField, ok bool) {
	var doc yaml.Node
	if err := yaml.Unmarshal(content, &doc); err != nil {
		return nil, false
	}
	if doc.Kind != yaml.DocumentNode || len(doc.Content) != 1 || doc.Content[0].Kind != yaml.MappingNode {
		return nil, false
	}
	m := doc.Content[0]
	for i := 0; i+1 < len(m.Content); i += 2 {
		k, v := m.Content[i], m.Content[i+1]
		if k.Kind != yaml.ScalarNode {
			return nil, false
		}
		value, ok := fieldValue(v)
		if !ok {
			return nil, false
		}
		fields = append(fields, frontmatterField{k.Value, value})
	}
	return fields, len(fields) > 0
}

// fieldValue is v as card text: a scalar's value (blank for null; a block
// scalar keeps its lines), or a list of scalars comma-joined.
func fieldValue(v *yaml.Node) (string, bool) {
	switch v.Kind {
	case yaml.ScalarNode:
		if v.Tag == "!!null" {
			return "", true
		}
		return strings.TrimRight(v.Value, "\n"), true
	case yaml.SequenceNode:
		items := make([]string, 0, len(v.Content))
		for _, item := range v.Content {
			if item.Kind != yaml.ScalarNode {
				return "", false
			}
			items = append(items, item.Value)
		}
		return strings.Join(items, ", "), true
	}
	return "", false
}

const (
	// cardMargin is the card's left margin, matching Glamour's Document margin
	// (markdownStyleConfig) so the card lines up with the body text. The same
	// amount is left free on the right.
	cardMargin = 2
	// cardMinWidth is the narrowest card drawn, border included. Narrower,
	// the code block reads better.
	cardMinWidth = 16
	// cardChrome is the border plus one cell of padding, each side.
	cardChrome = 4
	// cardGap separates the key column from the value column.
	cardGap = 2
)

// frontmatterCard draws fields as the card, shrunk to its content but no wider
// than width allows (a non-positive width is unbounded). Keys sit in a column
// as wide as the longest (at most half the card); values wrap within
// theirs with a hanging indent. ok is false when width is too narrow for a
// card. muted draws it in the unfocused pane's receded colours (D136).
func frontmatterCard(fields []frontmatterField, width int, muted bool) (string, bool) {
	outer := math.MaxInt32
	if width > 0 {
		outer = width - 2*cardMargin
		if outer < cardMinWidth {
			return "", false
		}
	}
	inner := outer - cardChrome

	keyWidth, valueWidth := 0, 1
	for _, f := range fields {
		keyWidth = max(keyWidth, lipgloss.Width(f.key))
		valueWidth = max(valueWidth, lipgloss.Width(string(expandTabs([]byte(f.value)))))
	}
	keyWidth = min(keyWidth, inner/2)
	valueWidth = min(valueWidth, inner-keyWidth-cardGap)

	keyStyle := cardStyle(theme.Subtle, muted).Width(keyWidth)
	valueStyle := cardStyle(theme.Fg, muted).Width(valueWidth)
	rows := make([]string, len(fields))
	for i, f := range fields {
		rows[i] = lipgloss.JoinHorizontal(lipgloss.Top,
			keyStyle.Render(f.key),
			strings.Repeat(" ", cardGap),
			valueStyle.Render(string(expandTabs([]byte(f.value)))),
		)
	}
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		MarginLeft(cardMargin)
	if markdownColoured() {
		box = box.BorderForeground(cardColor(theme.Border, muted))
	}
	return box.Render(strings.Join(rows, "\n")), true
}

// cardStyle is a style in colour c, or uncoloured where markdown is (the no-TTY
// style, markdownColoured).
func cardStyle(c color.Color, muted bool) lipgloss.Style {
	if !markdownColoured() {
		return lipgloss.NewStyle()
	}
	return lipgloss.NewStyle().Foreground(cardColor(c, muted))
}

// cardColor is c, receded the way the muted Glamour stylesheet recedes its
// colours when muted.
func cardColor(c color.Color, muted bool) color.Color {
	if !muted {
		return c
	}
	return lipgloss.Color(mutedColor(hexOf(c), theme.Bg, theme.MutedFade))
}

// frontmatterHeader splits src into the card drawn above the document and the
// source Glamour renders below it. Frontmatter the card cannot show stays in
// the source as a verbatim code block (fenceFrontmatter), and header is empty.
// header ends in a newline; Glamour's own output begins with a blank line, so
// the card gets a blank line above and below.
func frontmatterHeader(src []byte, width int, muted bool) (header string, body []byte) {
	fm, ok := splitFrontmatter(src)
	if !ok || fm.lang != "yaml" {
		return "", fenceFrontmatter(src)
	}
	fields, ok := frontmatterFields(fm.content)
	if !ok {
		return "", fenceFrontmatter(src)
	}
	card, ok := frontmatterCard(fields, width, muted)
	if !ok {
		return "", fenceFrontmatter(src)
	}
	return "\n" + card + "\n", fm.rest
}

// frontmatterSpan is where src's frontmatter sits on either side of rendering
// at width: the source lines its block spans, and the rendered lines its card
// (with the blank line above) takes at the top of renderMarkdown's output. Both
// are zero when there is no card: the code-block fallback keeps every source
// line where it was.
func frontmatterSpan(src []byte, width int) (sourceLines, renderedLines int) {
	header, _ := frontmatterHeader(src, width, false)
	if header == "" {
		return 0, 0
	}
	fm, _ := splitFrontmatter(src)
	return fm.lines(), strings.Count(header, "\n")
}
