# Progress — C1 read-across-roots

## Cursor

Current task: M3.6 (next: M4.2 — reordered by replan, see below)

Last reviewed: M3.1-M3.4 (1bc7c04), 2026-07-31 — standards + spec-fidelity review of the M3
milestone; findings logged as T16-T26 in `TRIAGE.md`. M3.1-M3.4 have since been validated
(2026-07-31, see Handoff).

## Handoff

M2.4 done (2026-08-01): internal/tree hardening (T11-T14, D23). `tree.go`'s `Walk` dropped the
dead non-directory branch entirely (T11) — it no longer stats `root` itself; a non-directory
argument now fails naturally when `walkDir`'s `os.ReadDir` returns ENOTDIR, so the contract
("root must be a directory") is enforced by the filesystem call that already needs to happen,
not a redundant pre-check. `policy.go`'s doc comments now say "Readable file" (T12, matching
GLOSSARY.md) instead of "readable document". `isReadable` matches the conventional
extensionless names case-insensitively via `strings.ToUpper` (T13/D23), same shape as the
already-case-insensitive extension path; locked by a new `policy_test.go`'s table-driven
`TestIsReadable` (18 cases spanning both the extension and name paths, case variants, and
negatives). `gitignore_test.go` gained the T14 case: `\!important.md` matches the literal name
`!important.md` (the escape-handling code already did this correctly — the gap was test
coverage, not behavior). Evidence: `make check` clean (`go test ./...` incl. the new
`TestIsReadable` and the new `TestIgnoreMatch` subcase); `make check-pty` clean (all 59+
checks, confirming M2/M3/M4 pty coverage is untouched by the internal/tree change). Committed.

M3.5 done (2026-08-01): roots/messaging correctness fixes (T17, T18, T20; D24 routing).
`roots.go`'s too-many-arguments message no longer implies `args[0]` was honored — it now
reports the count and says none were used (T17). `model.go`'s `toggleStar` treats an
unresolvable `storePath` (`""`) the same as the explicit `toggleStore` write-error branch: it
declines the flip and sets `m.message`, rather than silently updating the display with nothing
to persist it (T18) — `TestToggleStarWithNoStorePathDeclinesAndMessages` (renamed from
`...StillUpdatesDisplay`) locks the new behavior. `roots.go` gained `resolveCwd` (tries
`os.Getwd`, falls back to an absolute `$PWD` env value, else returns the error) and
`resolveLaunch` (main.go's wrapper: drops a relative launch argument rather than resolving it
against `.` when neither works, surfacing a message) — `main.go` now calls `resolveLaunch`
instead of building `cwd` inline (T20). New tests: `TestTooManyArgumentsMessageDoesNotImplyFirstWasHonored`,
`TestResolveCwdFallsBackToPWD` (reproduces the Getwd failure for real via chdir+rmdir, confirmed
non-skipped on this platform), `TestResolveLaunchDropsArgOnUnresolvableCwd`. Evidence: `make
check` clean; `make check-pty` clean (all 59+ checks). Committed.

Next: **M3.6 — M2/M3 code-review cleanup batch** (T9, T10, T16, T19, T21-T25, T27; see D24 for
the routing rationale). Ten small, mostly independent nits across `internal/tree`, `keymap.go`,
`viewer.go`, `rootsnav.go`/`filesnav.go` (shared windowing helper), `roots.go`/`rootsnav.go`
(the `firstIndex`/`storePath` nits), a shared color constant, and turning M3's two ad-hoc
validation scripts into permanent `test/pty` checks (T27) — work through the PLAN.md M3.6 entry
item by item; none of them depend on each other except sharing the same touched files. M4.2
follows after.

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
| M2.4 | done   | -        | -         | internal/tree hardening (T11-T14, D23): dead branch dropped, doc comments aligned, name matching case-insensitive, escape test added |
| M3.1 | done   | 2026-07-31 | 2026-07-31 | S1 launch contract; roots.go resolveRoots (D20, T7 fixed); 9 subtests + pty spot-check |
| M3.2 | done   | 2026-07-31 | 2026-07-31 | starred-roots store (D21); store.go load/save/toggle + loadStarredRoots wired into main.go; 9 tests |
| M3.3 | done   | 2026-07-31 | 2026-07-31 | Roots Nav pane (D22); rootsnav.go + model.{activateRoot,toggleStar}; 11 unit tests + 26 pty checks |
| M3.4 | done   | 2026-07-31 | 2026-07-31 | test-only: proved via `TestActivateRootLeavesOpenDocumentUntouched`, no prod code change |
| M3.5 | done   | -        | -         | replan C1/2026-08-01: roots/messaging correctness fixes (T17, T18, T20) |
| M3.6 | todo   | -        | -         | replan C1/2026-08-01: M2/M3 code-review cleanup batch (T9, T10, T16, T19, T21-T25, T27) |
| M4.1 | done   | -        | -         | edit handoff (S10/D3): tea.ExecProcess release/resume + finishEdit re-render from disk |
| M4.2 | todo   | -        | -         | replan C1/2026-08-01: scope grew (T15, T26 folded in), effort 2→3 |
| M5.1 | todo   | -        | -         |      |
| M5.2 | todo   | -        | -         |      |
| M5.3 | todo   | -        | -         |      |
