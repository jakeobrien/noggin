# YAML frontmatter and Glamour rendering

Question: what does a note with YAML frontmatter look like rendered through Glamour, and
what does it cost to hide it? (Asked while specifying C1/S7.)

## Findings

- **Glamour does not handle frontmatter.** `NewTermRenderer` builds goldmark with only
  `extension.GFM`, `extension.DefinitionList` and `parser.WithAutoHeadingID()` — no
  frontmatter/meta extension. (`glamour/glamour.go`, `NewTermRenderer`; open upstream issue
  https://github.com/charmbracelet/glamour/issues/205)
- **Unstripped frontmatter renders wrong, not merely ugly.** Under CommonMark rules the
  opening `---` is a thematic break, the YAML lines form a paragraph, and the *closing* `---`
  is read as a **setext heading underline** — so `title: foo / tags: [a, b]` renders as an H2
  above the document's real H1. (Derived from goldmark's unmodified CommonMark parsing plus
  the confirmed absence of any intercepting extension; not executed.)
- **Glow pre-strips the bytes.** `main.go` calls `utils.RemoveFrontmatter(b)` on raw file
  bytes before rendering; `utils/utils.go` matches `(?m)^---\r?\n(\s*\r?\n)?` twice and slices
  off everything through the second match. No YAML parsing at all. Glow does not extract a
  title from frontmatter either. (`glow/main.go` `executeCLI`; `glow/utils/utils.go`)
- **Glamour exposes no hook for goldmark extensions.** `TermRenderer.md` is unexported and no
  `TermRendererOption` reaches it; `WithEmoji` only works because it is in-package. A consumer
  wanting real frontmatter-aware parsing must bypass `TermRenderer` and wire its own goldmark
  instance onto the public `ansi.NewRenderer`, re-implementing the style/width/base-URL option
  plumbing. Meaningfully more expensive than pre-stripping.

## Consequence

Pre-strip before rendering, mirroring Glow. Zero new dependencies.

Relevant later: when tags arrive (C5), tag parsing reads the frontmatter block separately from
rendering — stripping for display does not preclude parsing for metadata. If frontmatter
*values* are ever needed at render time (e.g. a title), revisit
`github.com/abhinav/goldmark-frontmatter` + `ansi.NewRenderer`.
