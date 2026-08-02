# Verifying the TUI end-to-end in a pty

As of: C1

How to drive the real binary from a pty harness, against Charm v2 (bubbletea/bubbles/lipgloss
v2). The harness is checked in — `test/pty/harness.py` (the emulator and pty plumbing) and
`test/pty/checks.py` (the assertions), run by `make check-pty`. It lives in the repo, not
`/tmp`, specifically so it never has to be rebuilt from scratch again.

## The startup handshake must be answered

Under `TERM=xterm-256color` the v2 program emits these before its first frame and blocks
until they are answered. Captured off the wire, not from docs:

```
\x1b]11;?\x07      OSC 11 background colour   → \x1b]11;rgb:0000/0000/0000\x07
\x1b[c             Primary DA                 → \x1b[?62;22c
\x1b[?2026$p       DECRQM, synchronised output → \x1b[?2026;0$y   (0 = unrecognised)
\x1b[?2027$p       DECRQM, grapheme clustering → \x1b[?2027;0$y
```

Gotchas for a harness:

- **OSC 11 is BEL-terminated (`\x07`), not ST (`\x1b\\`).** A harness matching only the ST
  form answers nothing and hangs. Answer both forms.
- OSC 11 arrives **twice** — once from `lipgloss.HasDarkBackground` before `Run()` (the
  before-`Run()` terminal-query rule, see `.charm/DECISIONS.md` D13) and once from Bubble Tea
  itself. Answer every occurrence in every chunk, not once per read.

## Stripping escapes does not read the screen — you need a cell grid

v2 renders through `ultraviolet`, which **diffs cells and edits rows in place** rather than
redrawing rows as literal characters. Observed in noggin's own frames:

- blank runs come out as `ECH` (`\x1b[54X`) plus `CUF` (`\x1b[54C`) instead of spaces — strip
  the escapes and every gap in the row vanishes, so rows measure far short of the terminal
  width;
- horizontal **tabs** are used as cursor movement (`\x1b[23A\t\t` = up 23 rows, then to column
  16 on 8-column stops);
- shrinking text is fixed up with **DCH** (`\x1b[3P`) — a harness that ignores DCH will misread
  shrunk footer text as belonging to the wrong pane. This looked exactly like a product bug
  and cost real time to track down.

So `test/pty/harness.py` carries a ~90-line `Screen` emulator: CUU/CUD/CUF/CUB, CUP, ECH, DCH,
ICH, IL, DL, EL, ED, tabs at 8-column stops, `\r`/`\n` with scroll. Read geometry and text off
`Screen.raw_lines()`. SGR is parsed and dropped — assert **colour against the raw byte stream**
(e.g. `b"38;5;205"` for the focused border) and assert **text and geometry against the grid**.

A stripped-escapes harness is wrong in both directions under v2: it drops cells that are on
screen and keeps cells that are not.

## Height overflow is invisible in-grid — assert the border

A frame taller than the terminal does not produce over-long rows; the screen scrolls, so the
grid still measures exactly `cols × rows`. The visible symptom is that the **top border row
scrolls off**. `checks.py` therefore asserts row 0 is `╭…╮` and row −2 is `╰…╯` at every scroll
step — this is what caught the `expandTabs` regression (see D16). A harness this elaborate
needs a negative control; removing `expandTabs` is the standing check that these assertions
actually bite.

## Colours are still absent under `go test`

With the test binary's stdout a pipe there is no TTY, so glamour falls back to its no-TTY
style and lipgloss emits no escapes. Unit tests assert geometry, text and structure from
`View().Content`; colour needs the pty.

## Current coverage

144+ checks as of M5.8 (2026-08-02) across startup handshake, exit/restore behaviour, exact
frame geometry, focus cycling, resize reflow, Roots/Files Nav, find, help modal, navigator
toggle, and more — grown milestone by milestone. Key bytes in `checks.py` are a deliberate
carve-out from the D2 (provisional keymap) rule: the harness is on the far side of a language
boundary from `keymap.go`, so they live in one `KEY` table naming the binding each stands for.
