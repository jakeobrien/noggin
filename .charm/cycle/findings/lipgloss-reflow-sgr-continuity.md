# lipgloss v2's word-wrap re-asserts SGR state across a "\n" it didn't create

A third gotcha in the same family as `.charm/knowledge/lipgloss-v2-row-rendering-gotchas.md`
(read that first) — found building M9.1's S11 background tint (`cmd/noggin/style.go`'s
`tintPaneBackground`).

## The gotcha

`renderPane` hands its (already fully-rendered, ANSI-styled) title+content string to a
*second* `Style.Render` call — `renderBox`'s outer `box.Render(...)`, the one that draws the
border. That outer call's own word-wrap step (`lipgloss.Style.Render`'s "Word wrap" section,
`charm.land/lipgloss/v2`'s `style.go`) treats every `"\n"` in the input as a line it might
need to reflow, and — to keep a genuinely soft-wrapped, ANSI-styled paragraph looking
continuous across a real wrap point — re-asserts whatever SGR state was active *right before*
that `"\n"` at the start of the following line.

This means: leaving a custom SGR attribute (here, an injected background tint) merely
*active* across an internal `"\n"` — even without writing anything new at the boundary itself
— is enough for this second, unrelated `Style.Render` call to re-inject it at the start of the
next line. Confirmed live via `make check-pty` against the real binary: an early version of
`tintPaneBackground` deferred injection past each `"\n"` (to avoid deciding based on
incomplete lookahead) but left the tint's background *value* ambient going into the newline;
the outer box render then re-emitted it right before the next row's own reverse-video escape,
combining the tint with `SGR 7` and corrupting the swap (the selected row's text rendered in
the tint colour instead of the terminal's default foreground). `tintPaneBackground`'s own
isolated output never showed the leak — it only appeared after the second `Style.Render` pass,
which is why testing the function alone (however carefully) wasn't enough; the regression test
that actually needed to exist runs through `renderPane` end to end
(`TestUnfocusedFilesNavSelectedRowStaysCleanThroughRenderPane`, `focus_background_test.go`).

## The fix

Never leave a custom attribute merely "active" across a `"\n"` a function's own output
contains. Explicitly clear it (a targeted SGR — `\x1b[49m`, default background, not a full
reset, so nothing else has to be guessed at) immediately before every such `"\n"`, then
re-decide whether to reassert it once the actual next character/escape is known. This is the
opposite of the "defer past the newline" instinct that avoids premature decisions — deferring
*without also actively clearing* is exactly what lets an outer reflow step carry the value
across on its own.

## When this matters again

Any future renderer that (a) builds multi-line content out of several independently-styled
runs joined by `"\n"`, and (b) hands that content to *another* `Style.Render` call (a
`Width()`/`Height()`/border wrapper, `renderBox`'s own shape) is exposed to this. Reread before
adding another cross-cutting per-line SGR injection (a hover/dim/highlight layer, a future
style-pass "zebra striping", etc.) anywhere content flows through a second `Style.Render`.
