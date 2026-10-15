# The pty harness

`make check-pty` drives the real binary from a pseudo-terminal. `test/pty/harness.py` holds the pty plumbing and a small screen emulator; `test/pty/checks.py` holds the assertions. It is the only thing that verifies what actually reaches a terminal: colour, cell geometry, the startup handshake and frame height are all invisible to `go test`, which has no TTY.

## Colour is absent under `go test`

The test binary's stdout is a pipe, so markdown falls back to Glamour's colourless no-TTY style (`markdownColoured`, `document.go`) and lipgloss emits no escapes. Unit tests assert geometry, text and structure from `View().Content`; anything about colour or SGR attributes (reverse video, Peek's inverted header, focus borders) needs a pty check.

## The startup handshake must be answered

Under `TERM=xterm-256color`, Bubble Tea v2 emits these before its first frame and blocks until they are answered:

```
\x1b]11;?\x07      OSC 11 background colour    → \x1b]11;rgb:0000/0000/0000\x07
\x1b[c             Primary DA                  → \x1b[?62;22c
\x1b[?2026$p       DECRQM, synchronised output → \x1b[?2026;0$y   (0 = unrecognised)
\x1b[?2027$p       DECRQM, grapheme clustering → \x1b[?2027;0$y
```

- OSC 11 is BEL-terminated (`\x07`), not ST (`\x1b\\`). A harness matching only ST answers nothing and hangs. Answer both forms.
- Answer every occurrence in every chunk, not once per read: a query can arrive more than once.

## Stripping escapes doesn't read the screen; you need a cell grid

v2 renders through `ultraviolet`, which diffs cells and edits rows in place:

- blank runs are `ECH` (`\x1b[54X`) plus `CUF` (`\x1b[54C`), not spaces. Strip escapes and every gap vanishes;
- horizontal tabs are used as cursor movement (`\x1b[23A\t\t`);
- shrinking text is fixed up with `DCH` (`\x1b[3P`);
- moving the cursor by exactly one row uses bare `ESC M` (RI) / `ESC D` (IND), not CSI. Two-line rows provoke this on ordinary redraws.

So `Screen` in `harness.py` emulates CUU/CUD/CUF/CUB, CUP, ECH, DCH, ICH, IL, DL, EL, ED, RI, IND, 8-column tabs, and `\r`/`\n` with scroll. Assert text and geometry against `Screen.raw_lines()`; assert colour against the raw byte stream (e.g. a theme colour's SGR bytes). SGR is parsed and dropped by the grid.

**A harness gap looks exactly like a product bug.** An unhandled escape falls through to a single-byte skip and prints its final byte (`M`, `D`, `P`) as a literal character in some row; a missing DCH made footer text appear to belong to the wrong pane. Suspect `Screen.feed()` before the app when a row has a stray letter.

## Height overflow is invisible in the grid; assert the border

A frame taller than the terminal doesn't produce over-long rows: the screen scrolls, and the grid still measures exactly `cols × rows`. The symptom is that the top border scrolls off. `checks.py` asserts row 0 is `╭…╮` and the bottom pane border is intact at every scroll step. Removing `expandTabs` is the standing negative control that these assertions bite.

## Async redraws race the next keystroke

An async command's redraw (e.g. a Picker search landing) races the next keypress. The bytes that follow can be identical across runs while the emulator's cursor state differs, because an earlier redraw moved it, and the cursor-relative moves (`\x1b[NA`, `\x1b[ND`) then land on the wrong row. Seen as a ~1-in-4 flake when switching Picker mode or Scope right after a search had landed.

Sequence pty actions so nothing changes the frame while an async result may be in flight:

- do mode/Scope switches while the Prompt is empty (an empty Prompt dispatches no search);
- then type the query, so exactly one search settles;
- keep any follow-up action (activate) in the mode/Scope the search ran in.

## Writing assertions

- **Assert on something absent in the failing case.** The Prompt always echoes the query, so asserting the query text is "not in" the screen can never fail; assert on the Result row's filename instead.
- **Key bytes live in one `KEY` table** naming the binding each byte stands for. The harness is across a language boundary from `keymap.go`, so it must put real bytes on the wire; keep that knowledge in one place.
