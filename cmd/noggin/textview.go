package main

import (
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"noggin/internal/tree"
)

// renderMode is how a Document Viewer shows a document (D140): Markdown
// (rendered by Glamour) or Text (raw source, line-numbered). One per viewer;
// the zero value is Markdown, which is what every viewer starts on.
type renderMode int

const (
	modeMarkdown renderMode = iota
	modeText
)

func (r renderMode) String() string {
	if r == modeText {
		return "Text"
	}
	return "Markdown"
}

// toggled is the other mode.
func (r renderMode) toggled() renderMode {
	if r == modeText {
		return modeMarkdown
	}
	return modeText
}

// isTextFile reports whether d is a Text file (D140): a file that is not a
// Markdown file. It is always shown as text, whatever the viewer's mode. An
// article is never one: its converted markdown can be shown either way.
func isTextFile(d document) bool {
	return d.isOpen() && !d.isArticle() && !tree.IsMarkdown(filepath.Base(d.path))
}

// modeFor is the mode d is actually shown in by this viewer: its own mode,
// except that a Text file is always shown as Text.
func (v viewer) modeFor(d document) renderMode {
	if isTextFile(d) {
		return modeText
	}
	return v.mode
}

// shownDoc is the document the pane shows: the Peeked one while Peeking, the
// Active Document otherwise.
func (v viewer) shownDoc() document {
	if v.peeking() {
		return v.peek.doc
	}
	return v.doc
}

// toggleMode flips the viewer's render mode and re-renders. The find state is
// reset by the re-render (the match positions live in the other mode's line
// space); keeping it is a separate piece of work.
func (v viewer) toggleMode() viewer {
	v.mode = v.mode.toggled()
	return v.render()
}

// renderDocument renders d for width in the given mode. muted selects the
// unfocused-pane variant (D136).
func renderDocument(d document, mode renderMode, width int, muted bool) (string, error) {
	if mode == modeText {
		return renderText(d.src, width, muted), nil
	}
	return renderMarkdownStyled(d.src, width, muted)
}

// toggleRenderMode is keys.RenderMode's action: flip the render mode of the
// viewer `/` and `e` would target (DV2 when it holds Focus, DV1 otherwise).
// A Text file is shown as text whatever the mode, so there is nothing to
// toggle there: refused with a Footer message.
func (m model) toggleRenderMode() model {
	p := m.focusedViewerPane()
	v := m.viewerFor(p)
	if shown := v.shownDoc(); isTextFile(shown) {
		m.message = "cannot switch render mode: " + shown.label() + " is a text file"
		return m
	}
	return m.withViewerFor(p, v.toggleMode())
}

var (
	gutterStyle      = lipgloss.NewStyle().Foreground(theme.Subtle)
	mutedGutterStyle = lipgloss.NewStyle().Foreground(theme.UnfocusedSubtle)
	mutedTextStyle   = lipgloss.NewStyle().Foreground(theme.UnfocusedText)
)

// renderText renders src as raw text for a pane width cells wide: every
// source line numbered in a dim gutter, tabs expanded, long lines soft-wrapped
// with a blank gutter on the continuation rows. There is no syntax colour; the
// only styling is the gutter's, and the muted variant of both for an
// unfocused pane (D136). A non-positive width does not wrap.
//
// Control characters other than tab would be interpreted by the terminal
// rather than shown, so they are replaced; a trailing carriage return (CRLF
// files) is dropped.
func renderText(src []byte, width int, muted bool) string {
	text := strings.TrimPrefix(string(expandTabs(src)), "\ufeff")
	if text == "" {
		return ""
	}
	lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
	digits := len(strconv.Itoa(len(lines)))
	gutter := digits + 1
	// A pane too narrow for a gutter and any text beside it gets the text
	// alone, so the rows stay within the pane.
	if width > 0 && width <= gutter+1 {
		gutter = 0
	}
	wrapAt := 0
	if width > 0 {
		wrapAt = width - gutter
	}

	numStyle, bodyStyle := gutterStyle, lipgloss.NewStyle()
	if muted {
		numStyle, bodyStyle = mutedGutterStyle, mutedTextStyle
	}
	blank := strings.Repeat(" ", gutter)

	var out strings.Builder
	for i, line := range lines {
		line = sanitizeControl(strings.TrimSuffix(line, "\r"))
		rows := []string{line}
		if wrapAt > 0 {
			rows = strings.Split(ansi.Hardwrap(line, wrapAt, true), "\n")
		}
		for j, row := range rows {
			if i+j > 0 {
				out.WriteByte('\n')
			}
			if gutter > 0 {
				if j == 0 {
					out.WriteString(numStyle.Render(leftPad(strconv.Itoa(i+1), digits) + " "))
				} else {
					out.WriteString(blank)
				}
			}
			if row != "" && muted {
				row = bodyStyle.Render(row)
			}
			out.WriteString(row)
		}
	}
	return out.String()
}

func leftPad(s string, width int) string {
	if n := width - len(s); n > 0 {
		return strings.Repeat(" ", n) + s
	}
	return s
}

// sanitizeControl replaces control characters (ESC and friends) with U+FFFD so
// a file's contents can never drive the terminal. Tabs are already expanded.
func sanitizeControl(s string) string {
	if !strings.ContainsFunc(s, unicode.IsControl) {
		return s
	}
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return '\ufffd'
		}
		return r
	}, s)
}
