# lipgloss v2 gotchas

Traps for any renderer that builds rows or panes from several styled runs.

## `Width(...).MaxWidth(...)` wraps before it truncates

`Style.Width` word-wraps content to fit before `MaxWidth` gets a chance to truncate. Content of arbitrary length (a grep match's line text, a long path) must be truncated to the target width **before** it reaches a styled `Width(...).MaxWidth(...)` render, or an over-long line silently becomes several rows. See `rightTruncate` / `leftTruncate` and the truncate-then-style order in `picker_view.go`.

## A partial styled run inside a reverse-video row must restate the row's style

Styling part of a row (e.g. bolding fuzzy-matched characters) with a style that doesn't also carry the row's reverse video makes the inner render end with a full reset (`ESC[m`), which cancels the selection bar for the rest of the line. Render every run, matched and unmatched, with the row's base style plus the difference. See `highlightMatches` in `picker_view.go`; guarded by `TestPickerFileRowHighlightKeepsSelectionBar`.

The same mechanism defeats wrapping a whole pane in one outer `Style.Background(...)`: every nested `Render` closes with a bare reset, cancelling the outer background after the first run.

## An outer `Render` re-asserts SGR state across a `"\n"` it didn't create

When already-styled multi-line content is passed to a second `Style.Render` (e.g. `renderBox` drawing a border around a pane), that render's word-wrap step treats each `"\n"` as a possible wrap point and re-asserts whatever SGR state was active just before it at the start of the next line. So an attribute your output leaves merely *active* across a newline, without writing anything at the boundary, is re-injected on the next line by the outer render. A test of your function alone won't see it; only a test through the outer render (or a pty check) will.

Fix: explicitly clear any custom attribute (a targeted SGR such as `\x1b[49m`, not a full reset) immediately before every `"\n"` in your output, and reassert it after if needed.

This bit a since-removed per-line background tint for unfocused panes. Reread before adding any per-line SGR layer (highlight, dimming, striping) whose output flows through another `Style.Render`. Prefer choosing styles where content is built over rewriting ANSI after the fact.
