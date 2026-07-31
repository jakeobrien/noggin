# Progress — C1 read-across-roots

## Cursor

Current task: M3.4 (next: M4.1)

Last reviewed: M1.4 (db8a2e6), 2026-08-01 — two-pass standards + spec-fidelity review of the
whole M1 milestone (no prior formal checkpoint existed); findings logged as T3-T8 in
`TRIAGE.md`. M1.5, M2.1, M2.2, M2.3 have since been validated (2026-07-31, see Handoff).

## Handoff

M3.3 done (Roots Nav pane, S3, D22). New `cmd/noggin/rootsnav.go`: `rootsNav` is the pane —
a selection cursor (`selected`) plus which row is active (`active`) over a snapshot of `[]root`
(`rows`, each paired with a once-computed `broken` flag from `os.Stat`, mirroring the Files
Nav's static-tree-plus-live-selection split). `model.{roots, activeRoot}` (D20's placeholder
fields) are gone — `model.roots` is now the `rootsNav` itself; `model.withStorePath` threads
`storePath()` in once at startup (`main.go`) so a star toggle never re-derives
`os.UserConfigDir()`. Two new model methods do the cross-cutting work `rootsNav.handleKey`
can't: `activateRoot` (walks the selected root's tree, refuses a broken one with a footer
message, otherwise reuses `openFolder`) and `toggleStar` (calls M3.2's `toggleStore`, then
flips the pane's own display). `Activate` (enter) is reused from the Files Nav; a new `Star`
binding (`s`) is added. `renderFooter` now shows `m.message` when set (D22) — a deliberate,
minimal pull-forward of M4.2's footer work, scoped to exactly what S3's acceptance criterion
needs; hints/truncation/clear-on-keypress are still M4.2's.

Evidence: `make check` clean. 11 new unit tests (`rootsnav_test.go`): order preserved from S1,
cursor starts on the active root, two same-named roots distinguished by parent path (leaves
different rendered lines), missing path marked broken (in-model and on screen), active marker
tracks `active` not the cursor, starred marker toggles, selection clamps at both ends, the
empty-roots state, `activateRoot` re-roots the Files Nav and moves focus on success, refuses a
broken root (message set, active root and Files Nav both unchanged), and `toggleStar` persists
through M3.2's store (plus round-trips with no store path resolvable). `make check-pty`: 85/85
pass (26 new) — a bare launch with three pre-starred roots (two same-named "project" dirs under
different parents, one never-created path) renders all three, the same-named pair distinguished
by `admin`/`other` in their truncated parent paths, the missing one marked `(missing)`; shift+tab
reaches the Roots pane; the active/starred markers (`●`/`★`) appear and move correctly;
un-starring a root updates the screen and the on-disk store file; activating the broken root is
refused with a footer message naming it, and the active marker stays put. Harness change:
`Session` now takes an optional `env` dict (applied in the child only) so a pty check can point
`$HOME`/`$XDG_CONFIG_HOME` at an isolated config dir — needed to give the Roots Nav check a real
starred-roots store without touching the sandbox's actual one.

Next: **M3.4 — Activating a root re-roots the Files Nav without disturbing the open document
(S4, needs M2.2+M3.3).** Much of this likely already holds as a byproduct of M3.3's
`activateRoot` reusing `openFolder` (which only touches `m.files`/`m.focus`, never `m.viewer`) —
the task is to write the test that proves it: with a document open from root A, activate root B
and assert B's Files tree is collapsed at top level while the viewer still shows A's document at
its previous scroll offset. Check whether `filesNav.setRoot`'s reset-to-top-level-collapsed
behavior already satisfies "resetting expansion and selection" or needs anything extra.

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
| M3.3 | done   | -        | -         | Roots Nav pane (D22); rootsnav.go + model.{activateRoot,toggleStar}; 11 unit tests + 26 pty checks |
| M3.4 | todo   | -        | -         |      |
| M4.1 | todo   | -        | -         |      |
| M4.2 | todo   | -        | -         |      |
| M5.1 | todo   | -        | -         |      |
| M5.2 | todo   | -        | -         |      |
| M5.3 | todo   | -        | -         |      |
