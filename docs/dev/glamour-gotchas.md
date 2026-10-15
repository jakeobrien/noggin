# Glamour v2 gotchas

Third-party behaviour that has shaped how noggin renders markdown. Checked against glamour v2.0.1.

## No frontmatter support

`NewTermRenderer` builds goldmark with only GFM, DefinitionList and `WithAutoHeadingID`. Unhandled frontmatter renders *wrong*, not just ugly: the opening `---` is a thematic break, the YAML becomes a paragraph (with inline markdown interpreted), and the closing `---` is a setext underline, so the block renders as an H2 above the real title. GitHub's renderer does the same. Upstream issue: https://github.com/charmbracelet/glamour/issues/205.

noggin splits frontmatter off before Glamour sees it (`splitFrontmatter`, `document.go`) and draws a Frontmatter Card, falling back to rewriting the fences as a code block (`fenceFrontmatter`).

## No hook for goldmark extensions

`TermRenderer.md` is unexported and no `TermRendererOption` reaches it. `WithEmoji` works only because it's in-package. Anything needing a goldmark extension (frontmatter nodes, wiki-links, custom syntax) must bypass `TermRenderer`: build `goldmark.New(...)` yourself and install the public `ansi.NewRenderer(ansi.Options{...})` as its node renderer, about 15 lines. noggin only uses `WithStyles`, `WithChromaFormatter` and `WithWordWrap`, each of which sets one `ansi.Options` field. Also copy Glamour's default word-wrap of 80, which applies when width ≤ 0.

If you do this:

- **An unknown AST node prints a warning to stdout** with `fmt.Println` (`ansi/elements.go`). Inside Bubble Tea that corrupts the frame. Any extension producing a node kind Glamour doesn't know triggers it.
- `goldmark-meta`'s `WithTable()` synthesises `ast.String` cells, which crash Glamour (nil deref in `ansi/table.go`).

## The Chroma style is registered once per process

Glamour registers its code-block Chroma style under a fixed name and skips registration if the name exists (`ansi/codeblock.go`). A second render with different code colours keeps the first-registered colours for code blocks. This affects the muted unfocused render (a known, accepted quirk), and probably a Theme switch too: noggin has no workaround, so code blocks likely keep the launch Theme's colours after `ctrl+t`. The fix would be a distinct Chroma style name per Theme and variant, which `ansi.StyleConfig` doesn't expose, so it means patching or bypassing Glamour's code-block element.

## `WithAutoStyle` hangs inside Bubble Tea

`glamour.WithAutoStyle()` queries the terminal background on every render. Inside a running Bubble Tea program the reply is consumed by Bubble Tea's input reader, so the first render blocks forever. Never query the terminal after the program starts. noggin now paints from an explicit Theme and doesn't query at all.

## If tags come from frontmatter

Parse frontmatter by splitting and decoding YAML (`frontmatterFields` in `frontmatter_card.go` is the parse to share), reading only the file's head. Don't parse whole documents through goldmark to get it: a single-run benchmark on a 121 KB document measured goldmark with a frontmatter extension at ~4 ms vs ~7 µs for split + `yaml.Unmarshal`.
