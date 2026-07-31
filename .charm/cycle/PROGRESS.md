# Progress — C1 read-across-roots

## Cursor

Current task: M3.1 (next: M3.2)

Last reviewed: M1.4 (db8a2e6), 2026-08-01 — two-pass standards + spec-fidelity review of the
whole M1 milestone (no prior formal checkpoint existed); findings logged as T3-T8 in
`TRIAGE.md`. M1.5, M2.1, M2.2, M2.3 have since been validated (2026-07-31, see Handoff).

## Handoff

M2 validated 2026-07-31 (fresh-context subagent, per charm-validate): `make check` and
`make check-pty` (72/72) both clean; M1.5, M2.1, M2.2 and M2.3's acceptance criteria each
independently confirmed against source and tests (not just "tests pass") — `keyPress`/
`modifiers`/`namedKeys` test-only (keymap_test.go), `viewer_test.go` off `bubbles/viewport`
internals, `pane.title()` divergence commented; M2.1's fixture exercises exactly the
gitignored-.md/.git/.charm/main.go/README cases; M2.2's expand/collapse/selection/empty-folder
cases all present; M2.3's activate-to-viewer-with-focus path covered by both unit and pty
checks. No discrepancies found beyond the already-open low-severity TRIAGE items (T9-T15,
all cosmetic/hardening, none touch these four criteria) — Status table's Validated column
stamped for M1.5/M2.1/M2.2/M2.3; nothing added to TRIAGE, nothing flipped in Status.

Next: **M3.1 — resolve the session's roots from the launch argument (S1).** This replaces,
rather than supplements, `main.go`'s current temporary dir/file-argument scaffolding (dir →
`openFolder`, file → `openDocument`, both `os.Exit(1)` on error) — per T7 and the plan's own
note, a bad path must become a rejection message plus a session that still starts, not a
process exit. The table-driven cases the plan lists (bare launch, dir argument, file argument,
already-starred dir, non-existent path, subdirectory of a starred root, parent of a starred
root, more than one argument) imply a starred-roots concept that M3.2 (the store) has not been
built yet — M3.1's acceptance criterion talks about "already a starred root" cases, so either
M3.1 needs a minimal in-memory roots concept ahead of M3.2's persistent store, or those specific
cases wait for M3.2 to land; worth a fresh read of S1's spec text before starting, since PLAN.md
alone under-specifies the roots data shape at this boundary. No footer-message mechanism exists
yet for rejections either (M4.2) — M3.1 will need to decide where a rejection message lives in
the meantime (a temporary field on `model`, most likely) or explicitly punt display to M4.2 and
just ensure the session still starts.

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
| M3.1 | todo   | -        | -         |      |
| M3.2 | todo   | -        | -         |      |
| M3.3 | todo   | -        | -         |      |
| M3.4 | todo   | -        | -         |      |
| M4.1 | todo   | -        | -         |      |
| M4.2 | todo   | -        | -         |      |
| M5.1 | todo   | -        | -         |      |
| M5.2 | todo   | -        | -         |      |
| M5.3 | todo   | -        | -         |      |
