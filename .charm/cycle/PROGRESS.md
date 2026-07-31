# Progress — C1 read-across-roots

## Cursor

Current task: M3.3 (next: M3.4)

Last reviewed: M1.4 (db8a2e6), 2026-08-01 — two-pass standards + spec-fidelity review of the
whole M1 milestone (no prior formal checkpoint existed); findings logged as T3-T8 in
`TRIAGE.md`. M1.5, M2.1, M2.2, M2.3 have since been validated (2026-07-31, see Handoff).

## Handoff

M3.2 done (starred-roots store, S2, D21). New `cmd/noggin/store.go`: `loadStore`/`saveStore`/
`toggleStore` take an explicit path (testable without touching a real config dir); `storePath()`
isolates `os.UserConfigDir()` (D7); `loadStarredRoots()` is the `main.go` entry point, folding a
non-resolvable config dir into "no starred roots". Format is one absolute path per line, `#`
comments and blanks ignored (same shape as the `.gitignore` files noggin already parses); a
"garbage" line is anything that isn't a bare absolute path, dropped silently rather than erring
— see D21 for why structural filtering was chosen over a stricter parseable format.
`toggleStore` writes the whole list through on every call (load, flip, save) — no batching.
`main.go` now calls `loadStarredRoots()` instead of passing `nil` into `resolveRoots`, so a real
`~/.config/noggin/roots` (or platform equivalent) is live starting this task.

Evidence: `make check` clean; `make check-pty` (all 59 pass, unaffected — no `noggin` config dir
exists in the sandbox, so bare-launch pty checks still see zero starred roots). 9 new store
tests: missing file, garbage file, comments/blanks ignored, missing-on-disk path kept unpruned,
save creates the config dir, toggle-then-reload proves restart survival (the acceptance
criterion in prose), toggle removes an already-starred path, and a cleared-`$HOME`/
`$XDG_CONFIG_HOME` case proves `loadStarredRoots` degrades to empty rather than panicking.

Next: **M3.3 — Roots Nav pane (S3, needs M3.1+M3.2).** Ad-hoc root first then starred roots,
each entry showing base name plus left-truncated parent path, active highlighted, starred
marked, missing paths marked broken and not activatable; move / activate / toggle-star actions.
Toggle-star is the first real caller of `toggleStore` — it will need the resolved `storePath()`
threaded into the model (or a small wrapper) so the UI action can persist without re-deriving
the path each time. Rendering follows M2.2's Files Nav pattern (a pane-local struct holding
selection over an immutable list) but roots.go's `root` type already exists — M3.3 is display +
interaction, not new data modeling.

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
| M3.2 | done   | -        | -         | starred-roots store (D21); store.go load/save/toggle + loadStarredRoots wired into main.go; 9 tests |
| M3.3 | todo   | -        | -         |      |
| M3.4 | todo   | -        | -         |      |
| M4.1 | todo   | -        | -         |      |
| M4.2 | todo   | -        | -         |      |
| M5.1 | todo   | -        | -         |      |
| M5.2 | todo   | -        | -         |      |
| M5.3 | todo   | -        | -         |      |
