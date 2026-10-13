package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"sync"

	"charm.land/glamour/v2/styles"
	"charm.land/lipgloss/v2"
)

// Unfocused-pane styling (S11, D136). The focused pane is marked by a thick
// border and a title bar (model.renderPane); an unfocused pane additionally
// recedes:
//
//   - a nav pane's selected row is a dim grey bar instead of a bright
//     reverse-video one, and its other text takes a muted foreground;
//   - a Document Viewer shows a desaturated, darkened Glamour render of its
//     document, from a stylesheet derived from the embedded one.
//
// None of it post-processes ANSI: every choice is made where the style is
// built (or, for documents, by rendering a second time from a derived
// stylesheet), so nothing fights lipgloss's reset/reflow behaviour
// (lipgloss-reflow-sgr-continuity.md).

// selectedRowStyle styles a nav pane's selected row: the usual bright
// reverse-video bar when the pane has focus, a plain grey background when it
// does not. An explicit background rather than Reverse+Faint: Faint's effect
// on a reverse-video row is terminal-specific (some dim the text, some the
// bar), whereas a grey background dims the bar itself and leaves the text at
// full intensity. No Reverse is involved, so there is nothing to swap.
func selectedRowStyle(st lipgloss.Style, unfocused bool) lipgloss.Style {
	if unfocused {
		return st.Background(dimBarColor())
	}
	return st.Reverse(true)
}

// dimBarColor is the unfocused selection bar: a mid grey that sits between
// the terminal background and a full-intensity reverse bar. The light
// stylesheet (a light terminal) gets a light grey instead of a dark one.
func dimBarColor() color.Color {
	if markdownStyle() == styles.LightStyle {
		return lipgloss.Color("252")
	}
	return lipgloss.Color("238")
}

// mutedText is the nav-pane counterpart of the muted Glamour style: it
// gives a non-selected row's style a muted foreground while its pane is
// unfocused. dim marks text that is already de-emphasised (paths, mode/scope
// line, empty-state text), which gets an extra step down so the hierarchy
// survives. Selected rows are never passed through here, so the selection
// treatment stays independent of it.
func mutedText(st lipgloss.Style, on, dim bool) lipgloss.Style {
	if !on {
		return st
	}
	light := markdownStyle() == styles.LightStyle
	switch {
	case light && dim:
		return st.Foreground(lipgloss.Color("247"))
	case light:
		return st.Foreground(lipgloss.Color("240"))
	case dim:
		return st.Foreground(lipgloss.Color("241"))
	default:
		return st.Foreground(lipgloss.Color("246"))
	}
}

// mutedStyleJSON derives a muted copy of the embedded dark/light stylesheet
// for markdownStyle(): every "color"/"background_color" (including Chroma's
// code-block palette) is desaturated and pushed toward the background. ok is
// false when the style has no embedded sheet (the no-TTY style), in which
// case the caller renders normally — it has no colour to mute.
var mutedStyleJSON = sync.OnceValues(func() (b []byte, err error) {
	style := markdownStyle()
	path, ok := customStylePaths[style]
	if !ok {
		return nil, nil
	}
	src, err := customStyleFiles.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc any
	if err := json.Unmarshal(src, &doc); err != nil {
		return nil, fmt.Errorf("muted style: %w", err)
	}
	muteColors(doc, style == styles.DarkStyle)
	return json.Marshal(doc)
})

func muteColors(v any, dark bool) {
	switch t := v.(type) {
	case map[string]any:
		for k, val := range t {
			if s, ok := val.(string); ok && (k == "color" || k == "background_color") && s != "" {
				t[k] = mutedColor(s, dark)
				continue
			}
			muteColors(val, dark)
		}
	case []any:
		for _, val := range t {
			muteColors(val, dark)
		}
	}
}

// mutedColor desaturates c 60% toward its own luma, then pushes it toward
// the (assumed) background: darker on a dark theme, lighter on a light one.
func mutedColor(c string, dark bool) string {
	r16, g16, b16, _ := lipgloss.Color(c).RGBA()
	r, g, b := float64(r16>>8), float64(g16>>8), float64(b16>>8)
	luma := 0.299*r + 0.587*g + 0.114*b
	const desat = 0.6
	r, g, b = r+(luma-r)*desat, g+(luma-g)*desat, b+(luma-b)*desat
	fade := func(x float64) float64 {
		if dark {
			return x * 0.55
		}
		return 255 - (255-x)*0.5
	}
	return fmt.Sprintf("#%02x%02x%02x", uint8(fade(r)), uint8(fade(g)), uint8(fade(b)))
}
