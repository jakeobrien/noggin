package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"charm.land/glamour/v2"
	"charm.land/glamour/v2/styles"
	"golang.org/x/term"

	"noggin/internal/wikipedia"
)

// document is the active document: where it came from and its raw bytes. The
// bytes are kept so the viewer can re-render at a new width without touching
// the disk; re-reading from disk is an explicit act (M4.1's edit handoff).
//
// C6/M4.1 widens path beyond a file's absolute path to also carry an
// article's canonical URL (S3, S4's store key) — the same generalization
// docStoreKey (model.go) already made for the document store, reused here so
// the viewer, S9 Document History and Find (which never look past a
// document's rendered src) work on either kind unchanged. title is an
// article's own display title (S1's "pointer — canonical URL and title");
// always empty for a file, whose display name is instead derived from path
// (label(), below).
type document struct {
	path  string
	title string
	src   []byte
}

// isOpen reports whether there is a document to render. The zero document is
// "nothing open", which the viewer shows as its empty state.
func (d document) isOpen() bool { return d.path != "" }

// isArticleKey reports whether key names a Wikipedia article (its canonical
// URL) rather than a file (an absolute path) — the same "://" heuristic
// docStoreKey (model.go) relies on to tell the document store's two key
// shapes apart, shared here (D127's own precedent: this is a reliable way to
// tell them apart in practice, not a filesystem-level guarantee) so a
// document's identity can be classified without threading a redundant bool
// field through every constructor and test literal.
func isArticleKey(key string) bool { return strings.Contains(key, "://") }

// isArticle reports whether d is an article rather than a file, per
// isArticleKey.
func (d document) isArticle() bool { return isArticleKey(d.path) }

// label names d for a pane title or list row, or is empty for the zero
// document (mirrors isOpen's "nothing open" contract) — an article's own
// title (S1's "pointer ... and title") for an article, the base name of path
// for a file, exactly as every label computation here did before M4.1
// generalized path to also hold a URL.
func (d document) label() string {
	if !d.isOpen() {
		return ""
	}
	if d.isArticle() {
		return d.title
	}
	return filepath.Base(d.path)
}

// loadDocument reads path into a document.
func loadDocument(path string) (document, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return document{}, err
	}
	return document{path: path, src: src}, nil
}

// articleDocument builds a document for a fetched Wikipedia article (S3):
// path is the article's canonical URL (S4's store key), title its own
// display title, and src the already markdown-converted content
// (D112/M2.1's pipeline) — the same shape loadDocument produces for a file,
// so viewer.open/render (document.go/viewer.go) and everything downstream
// (Document History, Find) treat the two identically from here on. M4.2's
// fetch-completion handler is the eventual real caller; M4.1 exercises it
// directly in tests since M4.2 has not yet landed the fetch cmd that will
// call it in production.
func articleDocument(a wikipedia.Article) document {
	return document{path: a.URL, title: a.Title, src: []byte(a.Markdown)}
}

// frontmatterFence matches a line that is exactly a `---` fence. Both ends of a
// YAML frontmatter block are such a line.
var frontmatterFence = regexp.MustCompile(`(?m)^---[ \t]*\r?\n`)

// stripFrontmatter removes a leading YAML frontmatter block, fences included.
//
// Glamour has no frontmatter support and CommonMark reads the closing `---` as
// a setext heading underline, so an unstripped block renders as a spurious H2
// above the document's real title (D5, .charm/knowledge/frontmatter-rendering.md). Glow
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

// markdownColoured reports whether markdown is rendered in the Theme's colours.
// It is false only when stdout is not a terminal, where Glamour's own colourless
// NoTTYStyle is used instead: noggin is a TUI and never runs that way, but
// `go test`'s output is not a terminal, and viewer_test.go's fixtureStructure
// assertions rely on it being plain.
//
// Decided once. Colour no longer depends on the terminal's background, so
// there is no terminal query to run before Bubble Tea takes stdin.
var markdownColoured = sync.OnceValue(func() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
})

// tabStop is the column interval tabs are expanded to. Four is CommonMark's
// own figure for how far a tab advances when it determines block structure, and
// also lipgloss v2's default Style.TabWidth, so the two agree on where a tab
// lands.
const tabStop = 4

// expandTabs replaces tabs with spaces out to the next tab stop.
//
// Glamour still hands the contents of a code fence through verbatim, tabs
// included, and a tab's width is a property of the terminal, not the string —
// lipgloss measures a raw tab as one cell. Under v2 the consequence changed
// shape but not severity: lipgloss expands tabs to four spaces inside Render,
// so the viewport slices the line believing it is three cells shorter than it
// renders, and the pane box then wraps the overflow onto an extra row — the
// frame grows past the terminal's height. (Under v1 the same undercount drew
// through the pane's right border instead.) Expanding up front makes the
// measured width the real one either way. Tab stops of four are also exactly
// how CommonMark reads a leading tab, so this does not change how the markdown
// parses.
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
	return renderMarkdownStyled(src, width, false)
}

// renderMarkdownStyled is renderMarkdown with a muted option
// (unfocused_style.go): muted renders from the derived muted stylesheet, and falls
// back to the normal render where there is none (the no-TTY style).
func renderMarkdownStyled(src []byte, width int, muted bool) (string, error) {
	styleOption := glamour.WithStandardStyle(styles.NoTTYStyle)
	if markdownColoured() {
		cfg := markdownStyleConfig(theme)
		if muted {
			var err error
			if cfg, err = mutedMarkdownStyleConfig(theme); err != nil {
				return "", fmt.Errorf("markdown renderer: %w", err)
			}
		}
		styleOption = glamour.WithStyles(cfg)
	}
	// terminal16m: Glamour defaults to terminal256, which would quantize the
	// Theme's truecolor code colours to the 256 palette before Bubble Tea sees them.
	options := []glamour.TermRendererOption{styleOption, glamour.WithChromaFormatter("terminal16m")}
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
