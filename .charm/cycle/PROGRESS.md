# Progress — C1 read-across-roots

## Cursor

Current task: M4.1 (next: M4.2)

Last reviewed: M3.1-M3.4 (1bc7c04), 2026-07-31 — standards + spec-fidelity review of the M3
milestone; findings logged as T16-T26 in `TRIAGE.md`. M3.1-M3.4 have since been validated
(2026-07-31, see Handoff).

## Handoff

M3 milestone (M3.1-M3.4) validated fresh-context (2026-07-31, source: charm-validate): all four
PASS, no discrepancies vs. their PROGRESS.md claims. `make check` and `make check-pty` clean.
Strongest checks: M3.2's "survives a restart" ran as two separate real OS processes sharing an
isolated `$HOME`/`$XDG_CONFIG_HOME`; M3.3/M3.4 were re-exercised end-to-end via pty against the
real binary (cross-root byte-diff of the viewer pane), not just unit-level model assertions. One
new low-severity finding, T27 (TRIAGE.md): those two validation checks were ad hoc scripts, not
added as permanent regression tests — consider hardening into `test/pty`.

Next: **M4.1 — Edit handoff: release the terminal to `$EDITOR` (fallback `vi`), resume on exit and
re-render from disk (D3)** (needs M1.3). Acceptance: with `EDITOR` set to a script that appends a
line, the edit key returns to noggin showing the appended line, and with no document open the
footer reports that instead. Recall the standing constraint from D13/`glamour-in-bubbletea.md`:
`markdownStyle` resolves dark/light once via `lipgloss.HasDarkBackground` before `Run()`; returning
from `$EDITOR` must not re-query the terminal. Likely needs a `tea.ExecProcess`-style release/resume
(check bubbletea v2's current API for this) plus a re-render-from-disk path that reuses `loadDocument`
rather than duplicating it, and a footer message for the no-document case (`renderFooter` already
supports `m.message`, per M3.3/D22).

Still relevant:
- Go with the grain (D1) — now meaning Charm v2.
- D13's rule (no terminal query after `Run()`) stands, but v2 removed its cause. `markdownStyle`
  still resolves dark/light before `Run()` via `lipgloss.HasDarkBackground(os.Stdin, os.Stdout)`;
  `tea.RequestBackgroundColor` is the in-program alternative but is unreliable under tmux/WSL2.
  This still constrains M4.1: returning from `$EDITOR` must not re-query.
- Dependency graph got smaller, not bigger: termenv, muesli/{ansi,reflow,termenv}, go-osc52,
  go-isatty, go-localereader, coninput and x/cellbuf all dropped; ultraviolet (pre-1.0
  pseudo-version), x/termios and x/windows came in. 31 indirect deps down to 24. Noted for
  security-review time.
- Sandbox Go toolchain fully working; the Makefile's `unexport GOROOT` is defensive-only.

No blockers.

## Status

| Key  | Status | Reviewed | Validated | Note |
|------|--------|----------|-----------|------|
| M1.1 | done   | 2026-08-01 | 2026-07-31 | re-verified on v2 by M1.4 |
| M1.2 | done   | 2026-08-01 | 2026-07-31 | re-verified on v2 by M1.4 |
| M1.3 | done   | 2026-08-01 | 2026-07-31 | re-verified on v2 by M1.4 |
| M1.4 | done   | 2026-08-01 | 2026-07-31 | M1 complete and validated |
| M1.5 | done   | 2026-07-31 | 2026-07-31 | two-pass review (T9, T10 logged); validated (fresh-context subagent) |
| M2.1 | done   | 2026-07-31 | 2026-07-31 | internal/tree: policy + walk + hand-rolled gitignore (D18); review found T11-T14; validated |
| M2.2 | done   | 2026-07-31 | 2026-07-31 | Files Nav pane: filesNav over immutable tree (D19); 8 unit tests + 8 pty checks; validated |
| M2.3 | done   | 2026-07-31 | 2026-07-31 | Activate binding (enter): filesNav → loadDocument → openDocument; 3 unit tests + 5 pty checks; review found T15; validated |
| M3.1 | done   | 2026-07-31 | 2026-07-31 | S1 launch contract; roots.go resolveRoots (D20, T7 fixed); 9 subtests + pty spot-check |
| M3.2 | done   | 2026-07-31 | 2026-07-31 | starred-roots store (D21); store.go load/save/toggle + loadStarredRoots wired into main.go; 9 tests |
| M3.3 | done   | 2026-07-31 | 2026-07-31 | Roots Nav pane (D22); rootsnav.go + model.{activateRoot,toggleStar}; 11 unit tests + 26 pty checks |
| M3.4 | done   | 2026-07-31 | 2026-07-31 | test-only: proved via `TestActivateRootLeavesOpenDocumentUntouched`, no prod code change |
| M4.1 | todo   | -        | -         |      |
| M4.2 | todo   | -        | -         |      |
| M5.1 | todo   | -        | -         |      |
| M5.2 | todo   | -        | -         |      |
| M5.3 | todo   | -        | -         |      |
