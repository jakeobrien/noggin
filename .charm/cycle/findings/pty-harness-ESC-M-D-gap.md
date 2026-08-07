# pty harness gap: bare ESC M (RI) / ESC D (IND) were unhandled

`test/pty/harness.py`'s `Screen.feed()` (the cell-diffing emulator `pty-verification.md`
already documents) had no case for the C1-equivalent two-byte escapes `ESC M` (RI, cursor up
one row, no column change) and `ESC D` (IND, cursor down one row) — only their CSI cousins
(`\x1b[A`/`\x1b[B`, i.e. `n`-row CUU/CUD) were handled. An unmatched `ESC M`/`ESC D` fell
through to the harness's single-byte-skip fallback, which then printed the following letter
(`M` or `D`) as a literal character at the cursor position — indistinguishable, at a glance,
from a real rendering bug (a stray letter corrupting a row, or a search substring
coincidentally landing inside otherwise-garbled text and making a check pass for the wrong
reason). Both symptoms were hit while building C2a/M8.1 (two-line Roots Nav rows): a stray
`M` corrupting a path row after a star-toggle redraw, and a pre-existing Grep-mode `alt+h`
check (`"needlecontent" not in hidden_grep`) that had silently always passed because the
Prompt's own query-echo text was scrambled by this same bug, not because the check's logic
was sound (it wasn't — see below).

`ultraviolet` (bubbletea v2's renderer) emits bare RI specifically to reposition the cursor up
by exactly one row when that is cheaper than a full CUP, a pattern that a two-line-per-entry
row list provokes on ordinary redraws (e.g. toggling a star on one line of a two-line entry)
in a way no earlier single-line-row pane in this codebase needed.

Fixed by adding both cases to `Screen.feed()`, matching the existing CUU/CUD clamp-not-scroll
behaviour (no scroll-region bookkeeping): `ESC M` → `self.y = max(self.y - 1, 0)`, `ESC D` →
`self.y = min(self.y + 1, self.rows - 1)`.

Separately, this also exposed that the Grep-mode `alt+h` check asserting `"needlecontent" not
in hidden_grep` could never have failed either way: `needlecontent` was the literal query
text, which the Prompt line always echoes back regardless of whether the search found
anything. Fixed to assert on `.secret.md` (the Result row's own filename) instead — present
only when an actual match is listed, unlike the query echo.

Cited by: `test/pty/harness.py`'s `Screen.feed()`, `test/pty/checks.py`'s `C2a/M8.1` and
`C2a/M6.2` (Grep alt+h) sections.
