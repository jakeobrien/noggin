# A dual-render-mode approach to a git-status gutter

Exploratory (not yet a plan task). Answers: could a "wrapped / unwrapped" render toggle make a
per-line git-status gutter (adds/modifies/deletes) tractable in the Document Viewer, without
forking Glamour's renderer internals?

## The core obstacle

Git diff hunks are addressed by source-file line number. Glamour's rendered output has no
stable relationship to source line numbers today: `WithWordWrap` reflows whole paragraphs
(N source lines -> M rendered lines, M generally != N), and `document.go`'s own
`stripFrontmatter`/`expandTabs` pre-processing already changes byte offsets/line counts before
Glamour ever sees the text. A naive "rendered line K == source line K" gutter is wrong from the
first wrapped paragraph on.

## The proposed fix: a second, line-preserving render mode

Disabling `WithWordWrap` alone does **not** restore 1:1 correspondence. In
`charm.land/glamour/v2@v2.0.1/ansi/paragraph.go:44-48`, every paragraph's block buffer has its
internal newlines replaced with spaces *before* wrapping is even considered, unless
`WithPreservedNewLines()` is also set:

```go
blk := bs.Current().Block.String()
if !ctx.options.PreserveNewLines {
    blk = strings.ReplaceAll(blk, "\n", " ")   // paragraph.go:46 — source line breaks die here
}
flow := lipgloss.Wrap(blk, int(bs.Width(ctx)), "")   // paragraph.go:48
```

Combine **both** existing public options:

```go
glamour.WithWordWrap(0),          // any value < 1
glamour.WithPreservedNewLines(),
```

- `WithPreservedNewLines()` skips the `ReplaceAll`, so each source line inside a paragraph stays
  on its own line in `blk`.
- A wrap width `< 1` makes `bs.Width(ctx)` (`blockstack.go:61-66`) evaluate to 0, and
  `github.com/charmbracelet/x/ansi@v0.11.6/wrap.go:294-296`'s `wrap()` short-circuits
  (`if limit < 1 { return s }`) — the string is returned completely untouched.

Together: prose gets exact 1:1 source-line-to-rendered-line correspondence, using two options
that already exist (the second is also what Glow's `-n/--preserve-new-lines` flag maps to) —
no fork of Glamour internals required.

This is uniform across block types, not paragraph-only: `heading.go:68` and
`blockelement.go:36` (covers blockquotes, lists, definition lists) funnel through the same
`bs.Width(ctx)` -> `lipgloss.Wrap` path. Code blocks are unaffected either way in both
modes — `CodeBlockElement.Render` never calls `lipgloss.Wrap`; source lines pass through
chroma verbatim already (`ansi/codeblock.go`), so they already have perfect line
correspondence today, in both wrapped and unwrapped rendering.

## What still has to be built (smaller than a renderer fork, but real)

Even in the unwrapped+preserved-newlines mode, rendered-line-index != source-line-number
exactly: Glamour still inserts structural lines with no source counterpart — document
top/bottom margin, a blank separator before non-first headings/paragraphs
(`paragraph.go:29`, `heading.go:43`), blockquote's leading `\n`
(`elements.go` blockquote case: `Entering: "\n"`), etc.

So the remaining work is a **line-alignment pass in application code**, not a renderer fork:
parse the source once with `goldmark` (already an indirect dependency,
`github.com/yuin/goldmark v1.7.13`) purely to get each top-level block's source line-span
(`ast.Node.Lines()`), then walk the rendered output in the same document order, skipping the
blank lines Glamour predictably inserts between blocks, to build a source-line -> rendered-line
lookup table. Bounded, but real — comparable in shape to `find.go`'s existing "compute matches
once against the final content, feed line numbers to `EnsureVisible`/`StyleLineFunc`" pattern.

## Display-side cost: unwrapped prose needs horizontal scroll noggin doesn't have

Checked `charm.land/bubbles/v2@v2.1.1/viewport/viewport.go:358-363`: with `SoftWrap` off, an
over-width line is **clipped**, not wrapped — `ansi.Cut(lines[i], m.xOffset, m.xOffset+maxWidth)`
— using an `xOffset` field the viewport already supports for horizontal scrolling. But `grep`
across the codebase found no `xOffset`/`SetXOffset` usage anywhere in `cmd/noggin` — nothing
wires it up. Shipping an unwrapped mode as useful (not just as gutter plumbing) means adding
horizontal-scroll key bindings, which is its own small scope decision against
`.charm/DECISIONS.md` D4's "viewport scrolling only, no cursor" framing.

Tables have a separate wrap/truncate path (`WithTableWrap`, keyed off column content rather
than `WordWrap`) — likely unaffected by this toggle, not independently verified here.

## Find-in-document highlighting is not actually coupled to wrap mode

The original framing ("wrapped mode gives up find highlighting") doesn't hold up:
`cmd/noggin/find.go:186-190`'s `findMatches` already computes matches directly against
whatever the final rendered content is (`ansi.Strip(content)`, split by line), independent of
wrap setting — see `.charm/knowledge/viewport-search-ansi-offsets.md`. Find-in-document already
works correctly today in wrapped mode; it operates in rendered-line space regardless of how
that content was produced. The gutter is the only feature that actually needs the unwrapped
mode's source-line semantics — find does not need Mode 2 for correctness.

## Effort read

- Two Glamour options (`WithWordWrap(0)`, `WithPreservedNewLines()`) behind a mode toggle:
  trivial.
- Source-line <-> rendered-line lookup table via a second `goldmark` parse + ordered block-line
  bookkeeping: ~2-3/5 — meaningfully cheaper than forking `ansi/renderer.go`/`margin.go` for
  block-position-aware rendering (the alternative "tier B" approach), because it stays entirely
  in application code against Glamour's public output.
- Horizontal scroll wiring for the unwrapped mode to be readable at all: ~1-2/5, separate scope
  decision.
- Gutter drawing itself: reuse the existing `viewport.Model.StyleLineFunc` hook (same primitive
  `find.go` already uses for match highlighting) — e.g. a single-sided `lipgloss.Border` per
  line, colored by status. Cheap once the line lookup exists.
- Git data acquisition (`git diff -U0`/`git status --porcelain`, repo-root discovery — not
  currently in `internal/tree`, which only handles `.gitignore`): ~1-2/5.

Net: turns what looked like a 4-5/5 "fork the renderer" problem into something closer to 3/5
overall, at the cost of shipping two visibly different reading experiences (wrapped prose vs.
raw line-preserved text) rather than one, and taking on the horizontal-scroll gap.

## Open questions if this becomes a real task

- Is the gutter meant only for the currently-open file, or also a file-level indicator in the
  Navigator tree (a much cheaper, mapping-free alternative — see effort discussion in the
  parent conversation, not filed here)?
- Does the mode toggle live per-document, globally, or auto-select (e.g. unwrapped whenever a
  gutter is requested)?
- Freshness: recompute git status only on open/explicit reload (consistent with M4.1's
  deliberate-reload philosophy), not on a background watch.
