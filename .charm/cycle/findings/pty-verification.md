# Verifying the TUI end-to-end in a pty

How to drive the real binary from a pty harness now that lipgloss styles are in play, and
why the M1.1-era harness stopped working.

## The startup handshake must be answered

Under `TERM=xterm-256color` the program emits two queries before its first frame and blocks
until both are answered:

- `\x1b]11;?\x1b\\` — OSC 11 background-colour query (termenv, via lipgloss' default renderer)
- `\x1b[6n` — cursor position report (Bubble Tea's startup)

A harness that answers neither sees only those bytes and hangs; answering only OSC 11 still
hangs. Answer both, echoing back on the pty master:

- `\x1b]11;rgb:0000/0000/0000\x1b\\`
- `\x1b[<rows>;1R`

This is harness fidelity, not a product defect — a real terminal answers both. M1.1's harness
worked because its `View()` returned a plain string, so no lipgloss renderer was ever
instantiated and no query was emitted. `TERM=dumb` also sidesteps the handshake (and still
emits SGR colour codes), which is the cheap fallback when only layout matters.

Working harness (Python `pty.fork` + `select`, TIOCSWINSZ for the size, respond to queries as
they appear in each chunk): verified 2026-07-31 against `bin/noggin` at 80x24 and 100x30 —
exit code 0 on `q`, `\x1b[?25h` / `\x1b[?2004l` present (terminal restored), focused border
`38;5;205` and unfocused `38;5;240` both present.

## Colours are absent under `go test`

lipgloss detects no TTY when the test binary's stdout is a pipe, so styled output renders
plain. Tests can assert geometry, text and structure from `View()` (`lipgloss.Width` /
`Height` handle the ANSI-free case fine) but never colour — colour needs the pty harness.

Cited: `cmd/noggin/model.go` (styles), `cmd/noggin/frame_test.go` (geometry assertions).
