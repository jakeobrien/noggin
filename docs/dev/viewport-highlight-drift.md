# bubbles v2 `viewport.SetHighlights` drifts under ANSI styling

Don't use `viewport.Model.SetHighlights` / `HighlightNext` / `HighlightPrevious` on ANSI-styled content such as Glamour output. This is the lower-level reason for ADR-0002.

## The bug

`SetHighlights(matches [][]int)` takes byte offsets and converts them to (line, column) in `parseMatches` (`charm.land/bubbles/v2@v2.1.1/viewport/highlight.go`). It walks graphemes of `ansi.Strip(content)` and accumulates a byte position, then indexes the **unstripped** `content` with it to detect line breaks. Once any escape precedes the position, the two strings have diverged by the length of every escape removed so far, so the reported **line** is wrong, and the error compounds down the document. It's not just a column-precision problem: scrolling and line marking land on the wrong line.

`go test` can't catch it: under no TTY Glamour emits no escapes, so every offset is correct by construction. It only shows in a pty (`make check-pty` has a dedicated check).

## What noggin does instead

`EnsureVisible(line, colstart, colend)` and the `StyleLineFunc` field take line numbers directly and do no offset walking, so they're safe. `cmd/noggin/find.go` finds matches in `ansi.Strip(content)`, counts newlines in that same string, and drives `EnsureVisible`/`StyleLineFunc` with the result. Searching the stripped text also avoids matching inside SGR sequences (every one ends in `m`).

This relies on the viewport's own `SoftWrap` staying off: the viewer wraps content to the pane before `SetContent`, so a line in the stripped content is the viewport's line index. Text Render Mode has no ANSI at all, which is why it can highlight exact spans.

If bubbles fixes `parseMatches` to index one string consistently, re-verify in a pty before switching to the library path.
