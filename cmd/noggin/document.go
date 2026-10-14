package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
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

// utf8BOM is the byte order mark some editors write at the start of a file.
// A frontmatter fence after one is still the document's first line.
var utf8BOM = []byte("\ufeff")

// frontmatterClosers maps an opening frontmatter fence to the code block
// language its contents are highlighted as and the fences that close it:
// `---` YAML (closed by `---`, or YAML's own `...` document-end marker) and
// `+++` TOML (Hugo).
var frontmatterClosers = map[string]struct {
	lang    string
	closers []string
}{
	"---": {"yaml", []string{"---", "..."}},
	"+++": {"toml", []string{"+++"}},
}

// frontmatter is a leading frontmatter block, split out of a document's source.
type frontmatter struct {
	lang    string // "yaml" or "toml": what the block is written in
	opener  []byte // the opening fence line, line ending included
	content []byte // the lines between the fences
	closer  []byte // the closing fence line; no line ending at end of file
	rest    []byte // the document after the closing fence
}

// lines is the number of source lines the block spans, fences included.
func (f frontmatter) lines() int {
	return bytes.Count(f.opener, []byte("\n")) + bytes.Count(f.content, []byte("\n")) + 1
}

// splitFrontmatter finds a leading frontmatter block in src.
//
// The opening fence must be the first line (after an optional byte order mark)
// and both fences exactly three characters at column 0, trailing whitespace
// allowed. Glow instead takes the first two `---` lines found anywhere, which
// eats the head of a document whose first two thematic breaks are `---`; and
// `----` is a thematic break, not a fence. Without a closing fence there is no
// block.
func splitFrontmatter(src []byte) (frontmatter, bool) {
	opener, rest := cutLine(bytes.TrimPrefix(src, utf8BOM))
	fence, ok := frontmatterClosers[fenceText(opener)]
	if !ok || !bytes.HasSuffix(opener, []byte("\n")) {
		return frontmatter{}, false
	}
	for block := rest; len(block) > 0; {
		line, after := cutLine(block)
		if slices.Contains(fence.closers, fenceText(line)) {
			return frontmatter{
				lang:    fence.lang,
				opener:  opener,
				content: rest[:len(rest)-len(block)],
				closer:  line,
				rest:    after,
			}, true
		}
		block = after
	}
	return frontmatter{}, false
}

// fenceFrontmatter rewrites a leading frontmatter block as a fenced code block,
// so Glamour shows it verbatim and Chroma-highlighted ahead of the document
// (D138). It is the fallback for frontmatter the Frontmatter Card cannot
// show. Left as-is, CommonMark reads the opening `---` as a thematic break
// and the closing one as a setext heading underline, rendering the block as a
// spurious H2 (.charm/knowledge/frontmatter-rendering.md).
//
// Only the two fence lines are replaced, each by one line, so the result has
// exactly the source's lines: resolveGrepLanding relies on ripgrep's raw-file
// line numbers addressing it unchanged. An empty block becomes blank lines
// rather than an empty code box. Source without frontmatter is returned
// untouched.
func fenceFrontmatter(src []byte) []byte {
	fm, ok := splitFrontmatter(src)
	if !ok {
		return src
	}
	var out bytes.Buffer
	if len(bytes.TrimSpace(fm.content)) == 0 {
		out.Write(lineEnding(fm.opener))
		out.Write(fm.content)
		out.Write(lineEnding(fm.closer))
	} else {
		ticks := strings.Repeat("`", max(3, longestBacktickRun(fm.content)+1))
		out.WriteString(ticks + fm.lang)
		out.Write(lineEnding(fm.opener))
		out.Write(fm.content)
		out.WriteString(ticks)
		out.Write(lineEnding(fm.closer))
	}
	out.Write(fm.rest)
	return out.Bytes()
}

// cutLine splits b after its first line ending; the last line may have none.
func cutLine(b []byte) (line, rest []byte) {
	if i := bytes.IndexByte(b, '\n'); i >= 0 {
		return b[:i+1], b[i+1:]
	}
	return b, nil
}

// lineEnding is line's line ending: "\n", "\r\n", or none at end of file.
func lineEnding(line []byte) []byte {
	return line[len(bytes.TrimRight(line, "\r\n")):]
}

// fenceText is line without its line ending and trailing whitespace.
func fenceText(line []byte) string {
	return string(bytes.TrimRight(line, " \t\r\n"))
}

// longestBacktickRun is the length of the longest run of backticks in b, so a
// code fence one longer cannot be closed early by anything inside it.
func longestBacktickRun(b []byte) int {
	longest, run := 0, 0
	for _, c := range b {
		if c == '`' {
			run++
			longest = max(longest, run)
		} else {
			run = 0
		}
	}
	return longest
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
// Frontmatter is shown as a Frontmatter Card above the document, or as a code
// block where the card cannot show it (D138). Width is the number of cells the
// result may occupy; a non-positive width renders unwrapped.
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
	header, body := frontmatterHeader(src, width, muted)
	out, err := renderer.RenderBytes(expandTabs(body))
	if err != nil {
		return "", fmt.Errorf("render markdown: %w", err)
	}
	return header + string(out), nil
}
