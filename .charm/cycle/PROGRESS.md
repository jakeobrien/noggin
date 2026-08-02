# Progress — C1 read-across-roots

## Cursor

Current task: none — M5.4 built this session. Next up: M5.5, then the accumulated
`Reviewed`/`Validated` gaps across M3.1, M3.5, M4.2, M5.1-M5.4, before the cycle can close.

## Handoff

M5.4 built and done this session (T28, T29, T34; D27, and D28 for a T28 scope refinement):

- T28 (`roots.go`): `resolveLaunch` no longer drops an absolute launch argument on a
  `resolveCwd` failure it doesn't need. The real fault line is directory-vs-file, not
  absolute-vs-relative (D28) — new `needsCwd(args, cwd)` stats the argument to tell them
  apart: an absolute directory never touches `cwd` again once resolved (`resolveDir`), but a
  file argument's ad-hoc root is always `cwd` (`resolveFile`), so it is still correctly
  dropped even when absolute.
- T29 (`find.go`): `findMatches`' smart-case path lower-cased the whole document and searched
  that copy directly; case folding is not byte-length-preserving (Kelvin sign U+212A, 3
  bytes, folds to 1-byte 'k'), so a found offset there could misattribute a match to the
  wrong line. New `foldLower` returns a byte-offset mapping back to the original string
  alongside the folded one; line/col are now always computed in the original's space.
- T34 (`model.go`): `resize()` compares `m.viewer.findEditing()` before/after `setSize` and
  sets a footer message when the resize's re-render (which already silently invalidates any
  find, per `viewer.render`) dropped a query being actively typed. `toggleNavigator` shares
  this path but can never trigger it (`handleFindInput` owns the keyboard while editing).

Evidence: `make check` and `make check-pty` (59 checks) both green; each new test
independently confirmed to fail against the pre-fix code via `git stash` on the one changed
file. Cleared `Reviewed` back to `-` for M3.1, M3.5, M4.2, M5.1, M5.2, M5.3 (their reviewed
code in `roots.go`/`find.go`/`model.go` changed this session); M3.6 left alone — its own
`roots.go` diff doesn't overlap what changed here.

Still relevant:
- Go with the grain (D1) now means Charm v2; D13's no-terminal-query-after-`Run()` rule
  stands (`markdownStyle` resolves dark/light before `Run()`).
- Dependency graph: 31 → 24 indirect deps on the v2 migration; M5.1 added
  `charmbracelet/x/ansi` direct and pulled in `atotto/clipboard` indirectly (unused) — noted
  for security-review time.
- Sandbox Go toolchain fully working; the Makefile's `unexport GOROOT` is defensive-only.

Next: M5.5 (T30, T32, T33, T35 — standards cleanup), then the accumulated `Reviewed`/
`Validated` gaps across M3.1, M3.5, M4.2, M5.1-M5.4 before the cycle can close.

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
| M2.4 | done   | 2026-08-01 | 2026-08-01 | internal/tree hardening (T11-T14, D23): dead branch dropped, doc comments aligned, name matching case-insensitive, escape test added |
| M3.1 | done   | -      | 2026-07-31 | S1 launch contract; roots.go resolveRoots (D20, T7 fixed); 9 subtests + pty spot-check; roots.go changed by M5.4 (T28) — Reviewed cleared, needs re-covering |
| M3.2 | done   | 2026-07-31 | 2026-07-31 | starred-roots store (D21); store.go load/save/toggle + loadStarredRoots wired into main.go; 9 tests |
| M3.3 | done   | 2026-07-31 | 2026-07-31 | Roots Nav pane (D22); rootsnav.go + model.{activateRoot,toggleStar}; 11 unit tests + 26 pty checks |
| M3.4 | done   | 2026-07-31 | 2026-07-31 | test-only: proved via `TestActivateRootLeavesOpenDocumentUntouched`, no prod code change |
| M3.5 | done   | -      | 2026-08-01 | replan C1/2026-08-01: roots/messaging correctness fixes (T17, T18, T20); review found T28 (fixed by M5.4); roots.go changed by M5.4 — Reviewed cleared, needs re-covering |
| M3.6 | done   | 2026-08-01 | 2026-08-01 | replan C1/2026-08-01: M2/M3 code-review cleanup batch (T9, T10, T16, T19, T21-T25, T27) |
| M4.1 | done   | 2026-08-01 | 2026-08-01 | edit handoff (S10/D3): tea.ExecProcess release/resume + finishEdit re-render from disk |
| M4.2 | done   | -      | 2026-08-01 | footer key hints (S11); T15 (activate read-error message) and T26 (empty-Roots-Nav guidance) folded in; model.go changed by M5.4 (T34) — Reviewed cleared, needs re-covering |
| M5.1 | done   | -      | -         | find in document (S8); find.go + keymap/model wiring; 8 unit tests + 5 pty checks; found and fixed a real bubbles/v2 SetHighlights bug along the way (see findings); fresh-context two-pass review logged T29-T36; find.go changed by M5.4 (T29) — Reviewed cleared, needs re-covering |
| M5.2 | done   | -      | -         | navigator toggle (S9, D25); computeLayout gained hideNav; model.toggleNavigator/resize/layout; ctrl+n; 4 unit tests + 8 pty checks; fresh-context two-pass review logged T29-T36; model.go's resize() changed by M5.4 (T34) — Reviewed cleared, needs re-covering |
| M5.3 | done   | -      | -         | help modal (S12, D26); help.go (helpBindings/helpBox/overlayHelp, lipgloss v2 Layer/Compositor); model.go helpOpen/openHelp/handleHelpInput; 10 unit tests + 11 pty checks; self fresh-context review found and fixed an unclamped-overlay overflow bug; independent fresh-context two-pass review found no fidelity issues, logged T29-T36 (mostly M5.1-scoped); model.go changed by M5.4 (T34) — Reviewed cleared, needs re-covering |
| M5.4 | done   | -      | -         | M5-review correctness/UX fixes (T28, T29, T34; D28): resolveLaunch/needsCwd in roots.go (T28), findMatches/foldLower in find.go (T29), model.resize's find-editing check (T34); 8 new/changed unit tests, each proven to fail against the pre-fix code by reverting; full check + check-pty (59) green; cleared M3.1/M3.5/M4.2/M5.1/M5.2/M5.3's Reviewed stamps (roots.go/find.go/model.go all changed since) |
| M5.5 | planned | -      | -         | replan C1/2026-08-02 (D27): M5-review standards cleanup (T30, T32, T33, T35) |
