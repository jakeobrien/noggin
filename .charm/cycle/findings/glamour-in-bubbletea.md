# Driving Glamour from inside a Bubble Tea program

> **Re-verified against Charm v2 in M1.4 (2026-07-31).** The deadlock below is genuinely gone:
> glamour v2 has no `WithAutoStyle` and queries the terminal at no point. But **`expandTabs` is
> still required** — that defect changed mechanism, not existence:
>
> - Glamour v2 still emits **6 literal tabs** into the fixture's rendered code fence.
> - lipgloss v2 *does* convert tabs itself (`Style.TabWidth`, default 4) — but only inside
>   `Render`. `lipgloss.Width("ab\tcd")` still returns **4**, not 8. The measurement is still
>   wrong; only who is misled by it changed.
> - New consequence: the viewport slices a tab-bearing line believing it three cells shorter
>   than it draws, then the pane box expands the tab and **wraps the overflow onto an extra
>   row** — the frame ends up one row taller than the terminal. (Under v1 the same undercount
>   drew *through* the pane's right border.) On a real terminal this scrolls the top border off.
> - Verified by removing the `expandTabs` call: `TestFrameStaysWithinTheTerminalWhileScrolling`
>   fails with `view height = 25, want 24`, and `make check-pty` fails "frame border intact
>   while scrolling" at both sizes. That removal is the standing negative control for the guard.
> - `tabStop = 4` now coincides with lipgloss v2's default `TabWidth`, so the two agree on where
>   a tab lands. Anything changing either must change both.

Two things bite when Glamour renders inside a running Bubble Tea program, both found while
building M1.3. Both are fixed in `cmd/noggin/document.go`; this is the why.

## `WithAutoStyle` deadlocks the program

`glamour.WithAutoStyle()` is `WithStandardStyle("auto")`, and `getDefaultStyle` resolves
`"auto"` *at render time* by calling `termenv.HasDarkBackground()`, which writes an OSC 11
query to the terminal and blocks reading stdin for the reply
(`glamour@v1.0.0/glamour.go:306-315`).

Inside a running Bubble Tea program stdin belongs to Bubble Tea's input reader, which swallows
the reply, so the query never completes and the first render never returns. Observed
symptom: the program emits its startup queries and then draws a single near-empty line
(56 bytes total) and hangs forever — no frame, no crash, no error.

Fix: make the same decision `WithAutoStyle` would, once, before Bubble Tea takes stdin, and
name the style explicitly (`glamour.WithStandardStyle`). `lipgloss.HasDarkBackground()` caches
behind a `sync.Once` (`lipgloss/renderer.go:138-151`), so after the first call no render
touches the terminal. Consequence: a terminal that changes its background mid-session keeps
the style it started with — acceptable, and the alternative is a hang.

Generalises: **any** library call that queries the terminal is unsafe once Bubble Tea is
running. Resolve it at startup. Relevant to M4.1, where the editor handoff hands the terminal
away and takes it back.

## Literal tabs survive into rendered output

Glamour passes code-fence contents through verbatim, tabs included — a tab-indented Go
snippet renders with real `\t` bytes in the output. A tab's width is a property of the
terminal, not the string, and lipgloss measures it as one cell, so Glamour wraps such a line
short and the terminal then draws it straight through the pane's right border.

Fix: expand tabs to 4-column tab stops before rendering, in the same pre-processing pass as
frontmatter stripping. Four is also CommonMark's own figure for how far a leading tab advances
when determining block structure, so this does not change how the markdown parses.

## Glamour's no-TTY style is what tests see

`WithAutoStyle`'s no-TTY branch (and so ours) selects `styles.NoTTYStyleConfig`, which emits
plain text: headings keep their `#` markers, emphasis keeps its `*`, tables become
pipe-and-dash. Structure is still assertable under `go test` (table cells, `•` bullets, `[x]`
tasks, blockquote prefixes); colour and box-drawing are not. Same split as
`pty-verification.md`: geometry and text in tests, appearance in a pty.

## Dependency note

`glamour@v1.0.0` pulls `lipgloss` up from `v1.1.0` to the pseudo-version
`v1.1.1-0.20250404203927-76690c660834`, so the direct requirement moves to a pre-release
commit. Also adds chroma, goldmark, bluemonday, `golang.org/x/net` and friends — Glamour is
not a small dependency. Worth a look at security-review time.
