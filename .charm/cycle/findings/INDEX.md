# Findings index

- `frontmatter-rendering.md` — Glamour ignores YAML frontmatter and CommonMark mis-renders it
  as an H2; Glow pre-strips the bytes. Settles C1/S7's frontmatter behavior and flags the
  cost of ever needing frontmatter *values* at render time.
- `bubbletea-v2-migration.md` — the verified Charm v2 surface: the `charm.land` module-path
  rename, the all-or-nothing coupling, every API break that touches our ~600 lines, and what v2
  buys (viewport search for S8, D13's hazard designed out, TTY-free test options). Settled D14;
  **migration completed in M1.4** — the closing section records what the note got right, what it
  got wrong (`expandTabs` is *not* redundant), and the one thing it missed (v2 ships no public
  parser for its own key names, which D2 forces us to hand-roll).
- `pty-verification.md` — driving the real binary from a pty: the startup queries must be
  answered or it hangs, and `go test` never sees colour. **Re-derived against v2 in M1.4** —
  different queries (BEL-terminated OSC 11, Primary DA, two DECRQM; no cursor-position report)
  and, more importantly, v2's cell-diffing renderer means a harness must emulate a screen
  (ECH/CUF/DCH/tabs) rather than strip escapes. The harness is now checked in at `test/pty/`,
  run by `make check-pty` — 59 checks. The recipe for every later end-to-end check.
- `glamour-in-bubbletea.md` — `WithAutoStyle` queries the terminal at render time and deadlocks
  against Bubble Tea's input reader (the program hangs with no frame); Glamour also leaks
  literal tabs into rendered code fences, which lipgloss mis-measures. Both fixed in M1.3;
  the general rule — resolve terminal queries before Bubble Tea starts — applies to M4.1.
  **Re-verified under v2 in M1.4**: the deadlock is designed out (glamour v2 queries nothing),
  but `expandTabs` is still required — lipgloss v2 expands tabs in `Render` yet still measures a
  raw tab as one cell, so the frame now grows a row instead of overflowing the border.
- `gitignore-matching.md` — why M2.1's walk hand-rolls `.gitignore` (dependency weight vs
  shell-out's wrong-for-arbitrary-roots vs a testable subset), the exact subset covered, and
  what is deliberately left out (global excludes, re-include under a pruned dir). Settled the
  M2.1 handoff's open dependency question.
- `go-toolchain-env.md` — **resolved 2026-07-31**: the sandbox's `GOROOT` was misset,
  breaking `go`; re-validated fresh and the container now has a working go1.26.5 toolchain,
  gcc/cgo, and gopls/staticcheck/goimports, no workaround needed. Kept for history and Charm
  library versions at plan time.
