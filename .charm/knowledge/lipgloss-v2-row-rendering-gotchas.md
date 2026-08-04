# lipgloss v2 row-rendering gotchas

As of: C2

Salvaged from C2/D53 and C2/D57 (retired as task-scoped decisions once M4.1/M6.1/M6.2
shipped and were tested) — the underlying lipgloss v2 behaviour is a recurring trap for any
future row renderer, not specific to the Picker.

## `Style.Width(...).MaxWidth(...)` word-wraps before it truncates

`lipgloss.Style`'s `Width` reflows (word-wraps) content to fit before `MaxWidth` ever gets a
chance to truncate it (`charm.land/lipgloss/v2` `style.go`). Content of unbounded/arbitrary
length (e.g. a grep match's line text) must be pre-truncated to the target width *before* it
is handed to a styled `Width(...).MaxWidth(...)` render — otherwise an over-length line
silently wraps into multiple rows instead of truncating to one.

Cited by: `cmd/noggin/picker_view.go`'s `rightTruncate` (mirrors `rootsnav.go`'s
`leftTruncate`) and `pickerGrepRow`'s truncate-then-style order.

## A partial styled run inside a reverse-video row must fully specify its style

Styling only *part* of a row's text (e.g. bolding the fuzzy-matched substring within a file
name) with a style that doesn't also restate the surrounding reverse-video attribute will have
lipgloss's inner style emit a full SGR reset (`ESC[m`) that cancels the outer row's
reverse-video bar for the rest of the line. Every run — matched and unmatched — must render
with its style fully specified (the row's base style plus whatever is different), not just the
run being highlighted.

Cited by: `cmd/noggin/picker_view.go`'s `highlightMatches`; the regression this guards against
is pinned by `TestPickerFileRowHighlightKeepsSelectionBar`.
