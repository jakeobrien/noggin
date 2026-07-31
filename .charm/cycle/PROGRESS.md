# Progress — C1 read-across-roots

## Cursor

Current task: M3.2 (next: M3.3)

Last reviewed: M1.4 (db8a2e6), 2026-08-01 — two-pass standards + spec-fidelity review of the
whole M1 milestone (no prior formal checkpoint existed); findings logged as T3-T8 in
`TRIAGE.md`. M1.5, M2.1, M2.2, M2.3 have since been validated (2026-07-31, see Handoff).

## Handoff

M3.1 done (S1 launch contract). New `cmd/noggin/roots.go`: `resolveRoots(args, starred []string,
cwd) resolution` — pure, stats only the single argument, returns the active root as a *path*.
The starred roots are **injected** (D20), so M3.1 needs no store; `main.go` passes `nil` for now.
`model.applyResolution` does the side-effecting walk (`tree.Walk`) and document read, downgrading
any failure to `model.message` — no `os.Exit` on a bad path anymore (T7 resolved). Sub/parent
detection is pure `filepath.Rel` on cleaned absolute paths, so a starred root need not exist on
disk (matches S2). New minimal scaffolding for later: `root{path, starred}` type and
`model.{roots, activeRoot, message}` fields, all for M3.2/M3.3 to build on.

Evidence: `make check` clean; `make check-pty` (all pass); `TestResolveRoots` covers all eight
listed cases + a first-run bonus (9 subtests pass); pty spot-check confirmed `noggin /no/such/path`
renders a session and quits exit 0 rather than exiting early. See D20 for the design rationale.

**Deferred to M4.2, not lost:** `model.message` is stored but not yet displayed — the footer that
surfaces launch rejections is M4.2's. Until then a rejected launch starts silently (the message
is on the model, ready to render). Also: the file-argument case dedupes cwd against starred by
exact path only; it deliberately does **not** apply the sub/parent rejection logic to cwd,
because S1 says a file opens "whether or not it falls inside any root."

Next: **M3.2 — starred-roots store (S2, D7).** One plain hand-editable file under the user's
config directory, tolerant load (missing/garbage → empty list, no error), star toggle written
through, missing stored paths kept not pruned. This is what supplies the `starred []string` that
M3.1's `resolveRoots` and `main.go` currently stub as `nil` — wiring the loaded list into
`main.go` is part of landing M3.2. `os.UserConfigDir()` is the D7 config-dir source.

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
| M3.1 | done   | -        | -         | S1 launch contract; roots.go resolveRoots (D20, T7 fixed); 9 subtests + pty spot-check |
| M3.2 | todo   | -        | -         |      |
| M3.3 | todo   | -        | -         |      |
| M3.4 | todo   | -        | -         |      |
| M4.1 | todo   | -        | -         |      |
| M4.2 | todo   | -        | -         |      |
| M5.1 | todo   | -        | -         |      |
| M5.2 | todo   | -        | -         |      |
| M5.3 | todo   | -        | -         |      |
