# Verifying the TUI end-to-end in a pty

How to drive the real binary from a pty harness. **Re-derived against Charm v2 in M1.4**
(2026-07-31); the v1 recipe is kept at the bottom because the shape of the problem survived
even though every specific answer changed.

The harness is now checked in — `test/pty/harness.py` (the emulator and pty plumbing) and
`test/pty/checks.py` (the assertions), run by `make check-pty`. It was rebuilt from scratch
twice because it only ever lived in `/tmp`; it lives in the repo now so the third time does
not happen.

## The startup handshake must be answered — v2 asks different questions

Under `TERM=xterm-256color` the v2 program emits these before its first frame and blocks
until they are answered. Captured off the wire, not from docs:

```
\x1b]11;?\x07      OSC 11 background colour   → \x1b]11;rgb:0000/0000/0000\x07
\x1b[c             Primary DA                 → \x1b[?62;22c
\x1b[?2026$p       DECRQM, synchronised output → \x1b[?2026;0$y   (0 = unrecognised)
\x1b[?2027$p       DECRQM, grapheme clustering → \x1b[?2027;0$y
```

Differences from v1 that will bite a harness ported over unchanged:

- **OSC 11 is BEL-terminated (`\x07`) now, not ST (`\x1b\\`).** A harness matching only the ST
  form answers nothing and hangs. Answer both forms.
- **The cursor-position report (`\x1b[6n`) is gone**, and two DECRQM mode queries and Primary DA
  are new. Answering the v1 set alone is not enough.
- OSC 11 still arrives twice — once from `lipgloss.HasDarkBackground` before `Run()` (D13) and
  once from Bubble Tea itself. **Answer every occurrence in every chunk**, not once per read;
  this is the one v1 rule that carried over verbatim.

## Stripping escapes no longer reads the screen — you need a cell grid

This is the big one. v2 renders through ultraviolet, which **diffs cells and edits rows in
place**. It does not redraw rows as literal characters. Observed in noggin's own frames:

- blank runs come out as `ECH` (`\x1b[54X`) plus `CUF` (`\x1b[54C`) instead of spaces — strip
  the escapes and every gap in the row vanishes, so rows measure far short of the terminal
  width;
- horizontal **tabs** are used as cursor movement (`\x1b[23A\t\t` = up 23 rows, then to column
  16 on 8-column stops);
- shrinking text is fixed up with **DCH** (`\x1b[3P`) — the renderer overwrote `Document · q
  quit` with `Roots ` and then deleted three characters to close the gap. A harness that
  ignores DCH reads the footer as `Roots nt · q quit` and concludes focus is in a pane that
  does not exist. This cost real time; it looked exactly like a product bug.

So `test/pty/harness.py` carries a ~90-line `Screen` emulator: CUU/CUD/CUF/CUB, CUP, ECH, DCH,
ICH, IL, DL, EL, ED, tabs at 8-column stops, `\r`/`\n` with scroll. Read geometry and text off
`Screen.raw_lines()`. SGR is parsed and dropped — assert **colour against the raw byte stream**
(`b"38;5;205"` for the focused border), where it is unambiguous, and assert **text and geometry
against the grid**.

A stripped-escapes harness is not merely less accurate under v2; it is wrong in both
directions, dropping cells that are on screen and keeping cells that are not.

## Height overflow is invisible in-grid — assert the border

A frame taller than the terminal does not produce over-long rows; the screen scrolls, so the
grid still measures exactly `cols × rows`. The visible symptom is that **the top border row
scrolls off**. `checks.py` therefore asserts row 0 is `╭…╮` and row −2 is `╰…╯` at every scroll
step. Without that assertion the whole suite passed with a one-row-too-tall frame (the
`expandTabs` regression in `glamour-in-bubbletea.md`) — a harness this elaborate needs a
negative control, so removing `expandTabs` is the standing check that these assertions bite.

## Colours are still absent under `go test`

Unchanged from v1: with the test binary's stdout a pipe there is no TTY, glamour falls back to
its no-TTY style and lipgloss emits no escapes. Unit tests can assert geometry, text and
structure from `View().Content`; colour needs the pty. `tea.WithWindowSize` /
`tea.WithColorProfile` were **not** adopted — see TRIAGE; unit tests already reach `View()`
directly without a program, so the options would only help if a check needed a running
program *and* deterministic colour, which none currently does.

## Current coverage

`make check-pty`: 59 checks, all passing at 80x24 and 100x30 — startup handshake, first frame
before input, exit 0 and terminal restored (`\x1b[?25h`, `\x1b[?2004l`) on both `q` and
`ctrl+c`, exact frame geometry, all three regions, focused/unfocused border colour on the wire,
the full focus cycle read off the footer in both directions (which also exercises Bubble Tea's
real decoding of `tab` and `shift+tab`, something the synthetic `keyPress` helper cannot
vouch for), resize reflow, the fixture rendered with frontmatter absent and no literal tab on
screen, half-page scrolling to the code fence and `G` to the last line with the border intact
throughout, `ctrl+b` paging back, and the empty state on bare launch.

Key bytes in `checks.py` are a deliberate, documented D2 carve-out: the harness is on the far
side of a language boundary from `keymap.go`, so they live in one `KEY` table naming the
binding each stands for.

## Historical: the v1 recipe (superseded)

v1 emitted `\x1b]11;?\x1b\\` (ST-terminated, from termenv via lipgloss' default renderer) and
`\x1b[6n` (Bubble Tea's startup), and answering only one of the two still hung. v1's renderer
wrote rows as literal characters, so stripping CSI and OSC sequences and splitting on `\r?\n`
was enough to read the screen — the reason M1.1's harness worked at all was that its `View()`
returned a plain string, so no lipgloss renderer was instantiated and no query was emitted.
`TERM=dumb` sidestepped the handshake. None of that generalises to v2.
