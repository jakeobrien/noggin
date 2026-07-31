package main

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"sync"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// document is the active document: where it came from and its raw bytes. The
// bytes are kept so the viewer can re-render at a new width without touching
// the disk; re-reading from disk is an explicit act (M4.1's edit handoff).
type document struct {
	path string
	src  []byte
}

// isOpen reports whether there is a document to render. The zero document is
// "nothing open", which the viewer shows as its empty state.
func (d document) isOpen() bool { return d.path != "" }

// loadDocument reads path into a document.
func loadDocument(path string) (document, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return document{}, err
	}
	return document{path: path, src: src}, nil
}

// frontmatterFence matches a line that is exactly a `---` fence. Both ends of a
// YAML frontmatter block are such a line.
var frontmatterFence = regexp.MustCompile(`(?m)^---[ \t]*\r?\n`)

// stripFrontmatter removes a leading YAML frontmatter block, fences included.
//
// Glamour has no frontmatter support and CommonMark reads the closing `---` as
// a setext heading underline, so an unstripped block renders as a spurious H2
// above the document's real title (D5, findings/frontmatter-rendering.md). Glow
// solves this by slicing the bytes off before rendering and so do we — no YAML
// parsing, no extra dependency.
//
// Unlike Glow, we require the opening fence at offset 0. Glow takes the first
// two fences found anywhere, which silently eats the head of a document whose
// first two thematic breaks are `---`.
func stripFrontmatter(src []byte) []byte {
	fences := frontmatterFence.FindAllIndex(src, 2)
	if len(fences) < 2 || fences[0][0] != 0 {
		return src
	}
	return src[fences[1][1]:]
}

// markdownStyle is the Glamour style to render with, decided once.
//
// Glamour's WithAutoStyle re-decides on every render, and deciding means asking
// the terminal for its background colour (termenv OSC 11). Inside a running
// Bubble Tea program that reply is swallowed by Bubble Tea's own input reader,
// so the query never completes and the render blocks forever — the program
// draws nothing. Making the same choice WithAutoStyle would, once and up front,
// avoids the deadlock: lipgloss caches its own background lookup, and after that
// no render touches the terminal.
//
// It must be resolved before Bubble Tea takes over stdin; main does that.
var markdownStyle = sync.OnceValue(func() string {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return styles.NoTTYStyle
	}
	if lipgloss.HasDarkBackground() {
		return styles.DarkStyle
	}
	return styles.LightStyle
})

// tabStop is the column interval tabs are expanded to. Four is CommonMark's
// own figure for how far a tab advances when it determines block structure.
const tabStop = 4

// expandTabs replaces tabs with spaces out to the next tab stop.
//
// Glamour hands the contents of a code fence through verbatim, tabs included,
// and a tab's rendered width is a property of the terminal, not the string —
// lipgloss measures it as one cell. So a tab-indented code fence wraps short and
// then draws straight through the pane's right border. Expanding first makes the
// measured width the real one. Tab stops of four are also exactly how CommonMark
// reads a leading tab, so this does not change how the markdown parses.
func expandTabs(src []byte) []byte {
	if !bytes.ContainsRune(src, '\t') {
		return src
	}
	out := make([]byte, 0, len(src)+len(src)/8)
	column := 0
	for _, b := range src {
		switch b {
		case '\t':
			pad := tabStop - column%tabStop
			out = append(out, bytes.Repeat([]byte{' '}, pad)...)
			column += pad
		case '\n':
			out = append(out, b)
			column = 0
		default:
			out = append(out, b)
			// Counting bytes overcounts multi-byte runes, which only ever
			// moves a following tab to an earlier stop. Close enough for
			// alignment, and never wider than the terminal.
			column++
		}
	}
	return out
}

// renderMarkdown renders markdown to styled, width-wrapped terminal output.
// Frontmatter is stripped first (D5). Width is the number of cells the result
// may occupy; a non-positive width renders unwrapped.
func renderMarkdown(src []byte, width int) (string, error) {
	options := []glamour.TermRendererOption{glamour.WithStandardStyle(markdownStyle())}
	if width > 0 {
		options = append(options, glamour.WithWordWrap(width))
	}
	renderer, err := glamour.NewTermRenderer(options...)
	if err != nil {
		return "", fmt.Errorf("markdown renderer: %w", err)
	}
	out, err := renderer.RenderBytes(expandTabs(stripFrontmatter(src)))
	if err != nil {
		return "", fmt.Errorf("render markdown: %w", err)
	}
	return string(out), nil
}
