# bubbles/v2 viewport's SetHighlights drifts off-line under real ANSI styling

Discovered building M5.1 (S8, find-in-document). Corrects part of D14 and
`bubbletea-v2-migration.md`'s "what it buys us" #1: viewport gaining
`SetHighlights`/`HighlightNext`/`HighlightPrevious`/`EnsureVisible` does not
mean S8 is "library behaviour instead of hand-rolled offset mapping" — the
first of those four is unsafe to use against Glamour-rendered content, and
noggin does not use it.

## The bug

`viewport.Model.SetHighlights(matches [][]int)` takes byte-offset ranges
"measured against the given content" and hands them to `parseMatches`
(`charm.land/bubbles/v2@v2.1.1/viewport/highlight.go:20`), which walks
`uniseg.NewGraphemes(ansi.Strip(content))` to convert each offset into a
(line, column). The walk accumulates a `bytePos` counter from the **stripped**
string's grapheme byte-lengths, then indexes `content[bytePos]` — the
**original, unstripped** string — to detect line breaks
(`highlight.go:45,58`). Once any ANSI escape sequence has appeared before
`bytePos`, that index no longer points at the same character in `content` it
points at in the stripped string: the two strings have diverged by the byte
length of every escape removed so far. The reported line number is wrong from
that point on, and the error compounds with every further escape.

Confirmed empirically (not just by reading the code): rendering
`cmd/noggin/testdata/fixture.md` with Glamour's real `styles.DarkStyle` (what
a real terminal gets — `markdownStyle()` only picks the colourless
`NoTTYStyle` when stdout is not a terminal, which is why `go test` never
observes this) and locating a match near the end of the document gives line 8
from this walk, against a true line of 78. A synthetic case with one long
combined SGR sequence (`\x1b[1;4;38;5;200m…`) shows the same failure with a
single escape.

It is not merely a "character-exact highlighting" shortfall (S8 already
allows for that) — the **line** SetHighlights reports, and so where
`EnsureVisible`/`HighlightNext` scroll to and which line `StyleLineFunc`-style
highlighting would mark, is wrong. `go test` cannot catch this: Glamour's
no-TTY style contains no escapes, so every offset in a unit test is already
correct by construction. It only shows up under a real pty (`make
check-pty`), which is why that is where the regression check for this lives
(`test/pty/checks.py`, M5.1 section).

## What still works, and what noggin uses instead

`viewport.Model.EnsureVisible(line, colstart, colend)` and the public
`StyleLineFunc func(int) lipgloss.Style` field are **not** implicated — both
take a line number directly and do no byte-offset walking of their own.
`cmd/noggin/find.go` computes matches and their line numbers itself, once,
against `ansi.Strip(viewport.Model.GetContent())` (consistent indexing: the
newline count up to an offset is computed in the same string the offset was
found in), then drives `EnsureVisible` and `StyleLineFunc` with that. This
also sidesteps a second, smaller risk of searching the styled content
directly: an SGR sequence's own bytes can coincidentally contain the query
(every one ends in the literal byte `m`).

This works because noggin's viewport never turns on `SoftWrap` — `viewer.go`
already wraps to the pane width via Glamour before `SetContent`, so a line in
the stripped content is the same line index the viewport itself uses; no
further reflow happens inside the viewport to account for.

## Consequence for future viewport search/highlight work

Do not reach for `SetHighlights`/`HighlightNext`/`HighlightPrevious` against
content that has been through Glamour (or any ANSI styling) without
re-verifying this in a pty first. If bubbles fixes the walk to track position
consistently in one string, this note should be re-checked against the fixed
version before reverting to the library path.

## Sources

- `charm.land/bubbles/v2@v2.1.1/viewport/highlight.go` (`parseMatches`,
  `coords`), `viewport.go` (`SetHighlights`, `EnsureVisible`, `styleLines`,
  `StyleLineFunc` field).
- Local verification: a throwaway probe rendering the fixture with
  `styles.DarkStyle` and replaying `parseMatches`' own walk against it (line 8
  vs true line 78), and a synthetic multi-case check varying escape length
  relative to plain text.
