package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"math"
	"reflect"
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"

	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
)

// sgrColor is c as the "2;r;g;b" tail of a truecolor SGR sequence (38;… for
// foreground, 48;… for background).
func sgrColor(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("2;%d;%d;%d", r>>8, g>>8, b>>8)
}

// luminance is WCAG 2's relative luminance.
func luminance(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	lin := func(v uint32) float64 {
		x := float64(v>>8) / 255
		if x <= 0.03928 {
			return x / 12.92
		}
		return math.Pow((x+0.055)/1.055, 2.4)
	}
	return 0.2126*lin(r) + 0.7152*lin(g) + 0.0722*lin(b)
}

// contrast is WCAG 2's contrast ratio between two colours (1 to 21).
func contrast(a, b color.Color) float64 {
	la, lb := luminance(a), luminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

// colorFields returns every color.Color field of struct v, by name.
func colorFields(v any) map[string]color.Color {
	out := map[string]color.Color{}
	rv := reflect.ValueOf(v)
	for i := 0; i < rv.NumField(); i++ {
		if c, ok := rv.Field(i).Interface().(color.Color); ok {
			out[rv.Type().Field(i).Name] = c
		}
	}
	return out
}

// TestThemesMeetContrast holds every shipped Theme to a readable-by-
// construction bar, since noggin no longer borrows the terminal's colours to
// get one: text tokens reach WCAG AA (4.5:1) on the canvas, syntax tokens on
// the code background, the unfocused muted text a looser 3:1 (it is meant to
// recede), and borders 3:1 as non-text UI.
func TestThemesMeetContrast(t *testing.T) {
	for name, th := range allThemes {
		check := func(group string, fields map[string]color.Color, bg color.Color, min float64, skip ...string) {
		fields:
			for field, c := range fields {
				for _, s := range skip {
					if field == s {
						continue fields
					}
				}
				if got := contrast(c, bg); got < min {
					t.Errorf("%s: %s.%s contrast %.2f:1, want >= %.1f:1", name, group, field, got, min)
				}
			}
		}
		check("Theme", map[string]color.Color{
			"Fg": th.Fg, "Subtle": th.Subtle, "Accent": th.Accent, "Highlight": th.Highlight,
		}, th.Bg, 4.5)
		check("Theme", map[string]color.Color{"UnfocusedText": th.UnfocusedText, "UnfocusedSubtle": th.UnfocusedSubtle}, th.Bg, 3)
		check("Theme", map[string]color.Color{"Border": th.Border}, th.Bg, 3)
		check("Markdown", colorFields(th.Markdown), th.Bg, 4.5, "H1Fg", "H1Bg", "CodeBg", "Rule")
		check("Markdown", map[string]color.Color{"Rule": th.Markdown.Rule}, th.Bg, 3)
		check("Syntax", colorFields(th.Syntax), th.Markdown.CodeBg, 4.5, "ErrorFg", "ErrorBg")

		// Text drawn on a coloured bar: the title bar (Bg on Accent), the H1
		// chip, the error chip, and unfocused rows on their bar.
		for what, pair := range map[string][2]color.Color{
			"title bar":          {th.Bg, th.Accent},
			"H1 chip":            {th.Markdown.H1Fg, th.Markdown.H1Bg},
			"syntax error chip":  {th.Syntax.ErrorFg, th.Syntax.ErrorBg},
			"unfocused row":      {th.UnfocusedText, th.UnfocusedBar},
			"unfocused selected": {th.Fg, th.UnfocusedBar},
		} {
			if got := contrast(pair[0], pair[1]); got < 4.5 {
				t.Errorf("%s: %s contrast %.2f:1, want >= 4.5:1", name, what, got)
			}
		}
	}
}

// TestMarkdownStyleConfigStaysInThePalette: every colour in the generated
// stylesheet is one the Theme names, so no literal can sneak in beside it.
func TestMarkdownStyleConfigStaysInThePalette(t *testing.T) {
	th := nightTheme
	palette := map[string]bool{hexOf(th.Fg): true}
	for _, fields := range []map[string]color.Color{colorFields(th.Markdown), colorFields(th.Syntax)} {
		for _, c := range fields {
			palette[hexOf(c)] = true
		}
	}
	src, err := json.Marshal(markdownStyleConfig(th))
	if err != nil {
		t.Fatal(err)
	}
	var doc any
	if err := json.Unmarshal(src, &doc); err != nil {
		t.Fatal(err)
	}
	var walk func(v any)
	seen := 0
	walk = func(v any) {
		switch x := v.(type) {
		case map[string]any:
			for k, val := range x {
				if s, ok := val.(string); ok && (k == "color" || k == "background_color") {
					seen++
					if !palette[s] {
						t.Errorf("stylesheet colour %s (%s) is not a Theme token", s, k)
					}
					continue
				}
				walk(val)
			}
		case []any:
			for _, val := range x {
				walk(val)
			}
		}
	}
	walk(doc)
	if seen < 20 {
		t.Errorf("walked only %d stylesheet colours; the stylesheet looks empty", seen)
	}
}

// TestMutedMarkdownStyleConfigDiffersFromFocused: muting recolours the
// stylesheet and keeps its structure.
func TestMutedMarkdownStyleConfigDiffersFromFocused(t *testing.T) {
	focused := markdownStyleConfig(nightTheme)
	muted, err := mutedMarkdownStyleConfig(nightTheme)
	if err != nil {
		t.Fatal(err)
	}
	if *muted.H3.Color == *focused.H3.Color {
		t.Errorf("muted H3 colour = focused (%s)", *muted.H3.Color)
	}
	if muted.H3.Prefix != focused.H3.Prefix || *muted.Document.Margin != *focused.Document.Margin {
		t.Errorf("muting changed non-colour fields")
	}
}

// paintedCells draws a frame and returns its cells, line by line.
func paintedCells(frame string, w, h int) []uv.Line {
	scr := uv.NewScreenBuffer(w, h)
	uv.NewStyledString(frame).Draw(scr, scr.Bounds())
	return scr.Lines
}

// TestPaintCanvasColoursEveryCell: unstyled text, styled text (including a
// style that sets only one of fg/bg), blank cells and a wide cell all come out
// with an explicit Fg and Bg, the styles that chose a colour keep it, and the
// visible text is unchanged.
func TestPaintCanvasColoursEveryCell(t *testing.T) {
	th := nightTheme
	red := lipgloss.Color("#ff0000")
	blue := lipgloss.Color("#0000ff")
	frame := strings.Join([]string{
		"plain " + lipgloss.NewStyle().Foreground(red).Render("red") + " tail",
		lipgloss.NewStyle().Background(blue).Render("bg") + "  漢字 x",
		"",
	}, "\n")
	out := paintCanvas(frame, 20, 3, th)

	trim := func(s string) []string {
		lines := strings.Split(ansi.Strip(s), "\n")
		for i := range lines {
			lines[i] = strings.TrimRight(lines[i], " ")
		}
		return lines
	}
	if got, want := trim(out), trim(frame); !reflect.DeepEqual(got, want) {
		t.Errorf("visible text changed:\n got %q\nwant %q", got, want)
	}

	same := func(a, b color.Color) bool { return hexOf(a) == hexOf(b) }
	lines := paintedCells(out, 20, 3)
	for y, line := range lines {
		for x := range line {
			c := line[x]
			if c.IsZero() { // wide-cell continuation
				continue
			}
			if c.Style.Fg == nil || c.Style.Bg == nil {
				t.Fatalf("cell (%d,%d) %q has no explicit colours: %+v", x, y, c.Content, c.Style)
			}
		}
	}
	// "plain " keeps the Theme canvas; "red" keeps red on the canvas Bg.
	if c := lines[0][0]; !same(c.Style.Fg, th.Fg) || !same(c.Style.Bg, th.Bg) {
		t.Errorf("unstyled cell = fg %s bg %s, want the Theme's", hexOf(c.Style.Fg), hexOf(c.Style.Bg))
	}
	if c := lines[0][6]; !same(c.Style.Fg, red) || !same(c.Style.Bg, th.Bg) {
		t.Errorf("red cell = fg %s bg %s, want red on the canvas", hexOf(c.Style.Fg), hexOf(c.Style.Bg))
	}
	// A style that chose only a background keeps it and takes the Theme's Fg.
	if c := lines[1][0]; !same(c.Style.Bg, blue) || !same(c.Style.Fg, th.Fg) {
		t.Errorf("bg-only cell = fg %s bg %s, want Theme fg on blue", hexOf(c.Style.Fg), hexOf(c.Style.Bg))
	}
	// A frame cell nothing wrote is a Bg-coloured space.
	if c := lines[2][5]; c.Content != " " || !same(c.Style.Bg, th.Bg) {
		t.Errorf("blank cell = %q bg %s, want a Bg space", c.Content, hexOf(c.Style.Bg))
	}
	// The wide cell is still one two-column cell followed by its placeholder.
	var wide int
	for x, c := range lines[1] {
		if c.Content == "漢" {
			wide = x
		}
	}
	if wide == 0 || lines[1][wide].Width != 2 || !lines[1][wide+1].IsZero() {
		t.Errorf("wide cell damaged: %+v", lines[1][wide:wide+2])
	}
}

// TestPaintCanvasKeepsReverseVideoReadable: a reverse-video cell gets the
// Theme's Fg and Bg like any other (the terminal swaps them), so it shows
// Bg-on-Fg rather than the user's colours.
func TestPaintCanvasKeepsReverseVideoReadable(t *testing.T) {
	out := paintCanvas(lipgloss.NewStyle().Reverse(true).Render("sel"), 6, 1, nightTheme)
	c := paintedCells(out, 6, 1)[0][0]
	if c.Style.Attrs&uv.AttrReverse == 0 {
		t.Fatalf("reverse attribute lost: %+v", c.Style)
	}
	if hexOf(c.Style.Fg) != hexOf(nightTheme.Fg) || hexOf(c.Style.Bg) != hexOf(nightTheme.Bg) {
		t.Errorf("reverse cell colours = %s on %s, want the Theme's", hexOf(c.Style.Fg), hexOf(c.Style.Bg))
	}
}

func TestPaintCanvasLeavesUnsizedFramesAlone(t *testing.T) {
	if got := paintCanvas("x", 0, 0, nightTheme); got != "x" {
		t.Errorf("paintCanvas on an unsized frame = %q, want it untouched", got)
	}
}

// TestViewIsFullyPainted: the model's real View paints every cell and asks the
// terminal for the Theme's default colours too.
func TestViewIsFullyPainted(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	m := sized(t, newModel().openFolder(fixtureTree()), 80, 24)
	v := m.View()
	if v.BackgroundColor == nil || hexOf(v.BackgroundColor) != hexOf(theme.Bg) {
		t.Errorf("View().BackgroundColor = %v, want Theme.Bg", v.BackgroundColor)
	}
	if v.ForegroundColor == nil || hexOf(v.ForegroundColor) != hexOf(theme.Fg) {
		t.Errorf("View().ForegroundColor = %v, want Theme.Fg", v.ForegroundColor)
	}
	lines := paintedCells(v.Content, 80, 24)
	if len(lines) != 24 {
		t.Fatalf("painted %d lines, want 24", len(lines))
	}
	for y, line := range lines {
		for x, c := range line {
			if c.IsZero() {
				continue
			}
			if c.Style.Fg == nil || c.Style.Bg == nil {
				t.Fatalf("View cell (%d,%d) %q is not fully painted: %+v", x, y, c.Content, c.Style)
			}
		}
	}
}

// hasReverse reports whether any visible cell of s is reverse video. It reads
// the cells (cellStyles, picker_view_test.go) rather than looking for a literal
// "\x1b[7m", because paintCanvas folds reverse into a longer SGR sequence.
func hasReverse(s string) bool {
	for _, c := range cellStyles(s) {
		if c.reverse {
			return true
		}
	}
	return false
}

// styledLineWith returns the first raw line of view whose visible text
// contains text.
func styledLineWith(view, text string) (string, bool) {
	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(ansi.Strip(line), text) {
			return line, true
		}
	}
	return "", false
}

// TestThemedMarkdownIsTruecolour: rendering markdown through the Theme's
// stylesheet emits only truecolor foregrounds — including inside fenced code,
// where Glamour's default Chroma formatter would quantize to the 256 palette.
// (renderMarkdown itself is colourless under `go test`, hence the direct
// Glamour call; it mirrors renderMarkdownStyled's options.)
func TestThemedMarkdownIsTruecolour(t *testing.T) {
	r, err := glamour.NewTermRenderer(
		glamour.WithStyles(markdownStyleConfig(nightTheme)),
		glamour.WithChromaFormatter("terminal16m"),
		glamour.WithWordWrap(60),
	)
	if err != nil {
		t.Fatal(err)
	}
	out, err := r.Render("# Title\n\nSome *emph* and `code`.\n\n```go\nfunc main() { return \"x\" }\n```\n")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "38;2;") || !strings.Contains(out, "48;2;") {
		t.Errorf("themed markdown has no truecolor sequences: %q", out)
	}
	if strings.Contains(out, "38;5;") || strings.Contains(out, "48;5;") {
		t.Errorf("themed markdown still contains 256-colour sequences: %q", out)
	}
}

// TestNoColorLeavesTheTerminalsDefaultsAlone: under NO_COLOR the View does not
// recolour the terminal's own default fg/bg.
func TestNoColorLeavesTheTerminalsDefaultsAlone(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	v := sized(t, newModel().openFolder(fixtureTree()), 80, 24).View()
	if v.BackgroundColor != nil || v.ForegroundColor != nil {
		t.Errorf("View sets terminal colours under NO_COLOR: bg=%v fg=%v", v.BackgroundColor, v.ForegroundColor)
	}
}
