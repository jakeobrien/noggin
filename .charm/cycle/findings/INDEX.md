# Findings index

- `frontmatter-rendering.md` — Glamour ignores YAML frontmatter and CommonMark mis-renders it
  as an H2; Glow pre-strips the bytes. Settles C1/S7's frontmatter behavior and flags the
  cost of ever needing frontmatter *values* at render time.
- `bubbletea-v2-migration.md` — the verified Charm v2 surface: the `charm.land` module-path
  rename, the all-or-nothing coupling, every API break that touches our ~600 lines, and what v2
  buys (viewport search for S8, D13's hazard designed out, TTY-free test options). Read before
  M1.4; settles D14.
- `pty-verification.md` — driving the real binary from a pty needs the OSC 11 + cursor-position
  startup queries answered or it hangs; `go test` never sees colour. The recipe for every later
  end-to-end check — **derived against Bubble Tea v1; re-verify under v2 in M1.4** (v2 sends a
  different startup query set and offers `WithWindowSize`/`WithColorProfile` instead).
- `glamour-in-bubbletea.md` — `WithAutoStyle` queries the terminal at render time and deadlocks
  against Bubble Tea's input reader (the program hangs with no frame); Glamour also leaks
  literal tabs into rendered code fences, which lipgloss mis-measures. Both fixed in M1.3;
  the general rule — resolve terminal queries before Bubble Tea starts — applies to M4.1.
  **Diagnosis is v1-era**: glamour v2 removed `WithAutoStyle` and queries nothing, and lipgloss
  v2 converts tabs itself, so both root causes are gone under v2 (D14) — the rule survives, the
  workarounds may not.
- `go-toolchain-env.md` — **resolved 2026-07-31**: the sandbox's `GOROOT` was misset,
  breaking `go`; re-validated fresh and the container now has a working go1.26.5 toolchain,
  gcc/cgo, and gopls/staticcheck/goimports, no workaround needed. Kept for history and Charm
  library versions at plan time.
