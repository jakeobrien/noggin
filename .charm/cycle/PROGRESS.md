# Progress — C1 read-across-roots

## Cursor

Current task: M4.2 (next milestone: M4)

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

M3.6 done (2026-08-01): M2/M3 code-review cleanup batch (T9, T10, T16, T19, T21-T25, T27; D24
routing) — ten independent nits, worked item by item. `keymap.go`'s header now says "the one
production file key literals may appear in" and cross-references `keymap_test.go`'s own
test-only key-encoding table, so the file's scope reads unambiguous next to D2 (T9). `viewer.go`'s
`setOffset` is now `withOffset`, a value receiver like every other `viewer` method, since its
pointer receiver bought nothing `SetYOffset`'s own reference semantics didn't already give it
(T10) — both call sites (`rootsnav_test.go`, `viewer_test.go`) updated to the `m.viewer =
m.viewer.withOffset(n)` shape. `rootsNav.move`/`filesNav.move` and their near-identical
`view` windowing (scroll-to-keep-selection-visible, pad short lists to height) now share
`moveSelection` and `windowedView` in a new `nav.go` (T16). `newRootsNav` dropped its always-""
`storePath` parameter — every call site passed `""`, letting `model.withStorePath` do the actual
work after construction — and its struct doc comment's stray "(T: rare — ...)" annotation became
plain prose while it was being touched anyway (T19, T23). `roots.go`'s `firstIndex` (a
one-line "0 if any roots else -1" helper, called from a single site) is inlined with a comment
in its place (T24). The `lipgloss.Color("245")` literal duplicated across `files.go`,
`rootsnav.go`, `viewer.go` and `model.go`'s `footerStyle` is now one `dimColor` var in a new
`style.go` (T25). New unit tests: `TestActivateRootClearsStaleMessageOnSuccess` (T21, a stale
`m.message` does not survive a successful `activateRoot`) and a `TestResolveRoots` table case
for `resolveFile`'s own cwd-is-already-starred dedup path, distinct from `resolveDir`'s (T22).
Finally, T27: the two ad-hoc validation scripts from the M3 validate pass (starred-roots
restart persistence across two real OS processes sharing an isolated `$HOME`/`$XDG_CONFIG_HOME`;
cross-root activation leaving the open document's viewer pane byte-identical, driven through the
real binary via pty) are now permanent checks in `test/pty/checks.py` — the restart check spawns
two separate `Session`s against one store file, the cross-root check byte-diffs only the viewer's
columns (right of the Navigator's box edge) so Files Nav/footer churn that IS expected to change
can't mask a regression. Evidence: `make check` clean (`go test ./...` incl. the two new tests);
`make check-pty` clean, 105 checks total (was 59+; the two new T27 checks plus their setup/exit
assertions account for the growth). Committed.

Next: **M4.2 — Footer: key hints for the focused pane** (T15, T26 folded in; needs M1.2). Done
when: hints change with focus between Roots Nav, Files Nav and viewer; a narrow terminal shows a
truncated hint set still naming the help affordance; a rejected action's message clears on the
next keypress; `activate()`'s file-read error surfaces a footer message instead of failing
silently (T15); and an empty Roots Nav (no starred, no ad-hoc root) shows guidance for a first
run with neither (T26). M3 is now fully done (all six tasks); M4.1 is done, M4.2 is the only
task left before M5's cut-candidate milestone.

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
| M3.6 | done   | -        | -         | replan C1/2026-08-01: M2/M3 code-review cleanup batch (T9, T10, T16, T19, T21-T25, T27) |
| M4.1 | done   | -        | -         | edit handoff (S10/D3): tea.ExecProcess release/resume + finishEdit re-render from disk |
| M4.2 | todo   | -        | -         | replan C1/2026-08-01: scope grew (T15, T26 folded in), effort 2→3 |
| M5.1 | todo   | -        | -         |      |
| M5.2 | todo   | -        | -         |      |
| M5.3 | todo   | -        | -         |      |
