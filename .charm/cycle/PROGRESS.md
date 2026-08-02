# Progress — C1 read-across-roots

## Cursor

Current task: none — a fresh-context two-pass code review closed the `Reviewed` gap this
session. Next up: the remaining `Validated` gaps across M5.1-M5.5, then charm-replan to triage
T37-T43, before the cycle can close.

## Handoff

Fresh-context two-pass review (standards + spec fidelity) done this session, scoped to the
diff since the last review commit (`git diff 04ff677..HEAD -- cmd/noggin`, i.e. M5.4 + M5.5's
changes to `find.go`, `roots.go`, `model.go` and their tests, plus M5.5's cosmetic-only touches
to `keymap.go`/`viewer.go`/`navigator_test.go`/`help_test.go`). This re-covers the
Reviewed-stamp gap M5.4 opened for M3.1, M3.5, M4.2, M5.1, M5.2, M5.3, and covers M5.4/M5.5
themselves for the first time — all eight stamped `2026-08-02` above. No fixes made (review is
read-only toward code); findings logged to TRIAGE.md as T37-T43 (one medium, six low) for
charm-replan: T37 (roots.go double-stat TOCTOU, medium), T38 (needsCwd's unused `cwd` param),
T39 (find.go stylistic), T40/T41 (test-coverage gaps), T42 (PLAN.md's M5.4 bullet stale against
D28), T43 (a committed find, not just mid-edit, is still silently dropped on resize — adjacent
to T34, pre-existing since M5.1, not a regression). `make check` reconfirmed green. Did not
re-review help.go, footer.go or files not touched since the last review pass — out of this
session's diff scope.

M5.5 built and done in the prior session (T30, T32, T33, T35 — standards cleanup, all
cosmetic, no behavior change):

- T30: the 13 `\u2014` occurrences in `find.go`/`navigator_test.go` were literal six-character
  escape text — re-encoded to the actual em dash.
- T32: `keymap.go`'s `Cancel`/`Help` doc comments described S12's help modal (M5.3) in future
  tense; reworded to past/present tense now that the modal exists.
- T33: `viewer`'s `find` field renamed to `findState` (struct type `find`/`newFind()`
  unchanged) — mechanical rename across `viewer.go`/`find.go`/`find_test.go`/`help_test.go`.
- T35: added a comment at `model.go`'s `FindNext`/`FindPrevious` dispatch confirming they are
  intentionally global like `Find`, not scoped to viewer focus.

Evidence: `make check` and `make check-pty` (59 checks) both green; grep confirms zero leftover
`\u2014` escape text and zero stray `.find`/`find:` references. No behavior changed, so no
`Reviewed` stamps needed clearing beyond what M5.4 already cleared.

Next: the accumulated `Reviewed`/`Validated` gaps across M3.1, M3.5, M4.2, M5.1-M5.5 — likely
a code-review + validate pass before the cycle can close. No blockers.

Still relevant:
- Go with the grain (D1) now means Charm v2; D13's no-terminal-query-after-`Run()` rule
  stands (`markdownStyle` resolves dark/light before `Run()`).
- Dependency graph: 31 → 24 indirect deps on the v2 migration; M5.1 added
  `charmbracelet/x/ansi` direct and pulled in `atotto/clipboard` indirectly (unused) — noted
  for security-review time.
- Sandbox Go toolchain fully working; the Makefile's `unexport GOROOT` is defensive-only.

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
| M3.1 | done   | 2026-08-02 | 2026-07-31 | S1 launch contract; roots.go resolveRoots (D20, T7 fixed); 9 subtests + pty spot-check; re-covered by the 2026-08-02 review of the M5.4 diff to roots.go (found T37, T38, T40) |
| M3.2 | done   | 2026-07-31 | 2026-07-31 | starred-roots store (D21); store.go load/save/toggle + loadStarredRoots wired into main.go; 9 tests |
| M3.3 | done   | 2026-07-31 | 2026-07-31 | Roots Nav pane (D22); rootsnav.go + model.{activateRoot,toggleStar}; 11 unit tests + 26 pty checks |
| M3.4 | done   | 2026-07-31 | 2026-07-31 | test-only: proved via `TestActivateRootLeavesOpenDocumentUntouched`, no prod code change |
| M3.5 | done   | 2026-08-02 | 2026-08-01 | replan C1/2026-08-01: roots/messaging correctness fixes (T17, T18, T20); review found T28 (fixed by M5.4); re-covered by the 2026-08-02 review of the M5.4 diff to roots.go (found T37, T38, T40) |
| M3.6 | done   | 2026-08-01 | 2026-08-01 | replan C1/2026-08-01: M2/M3 code-review cleanup batch (T9, T10, T16, T19, T21-T25, T27) |
| M4.1 | done   | 2026-08-01 | 2026-08-01 | edit handoff (S10/D3): tea.ExecProcess release/resume + finishEdit re-render from disk |
| M4.2 | done   | 2026-08-02 | 2026-08-01 | footer key hints (S11); T15 (activate read-error message) and T26 (empty-Roots-Nav guidance) folded in; re-covered by the 2026-08-02 review of the M5.4 diff to model.go (found T43, adjacent not blocking) |
| M5.1 | done   | 2026-08-02 | -         | find in document (S8); find.go + keymap/model wiring; 8 unit tests + 5 pty checks; found and fixed a real bubbles/v2 SetHighlights bug along the way (see findings); fresh-context two-pass review logged T29-T36; re-covered by the 2026-08-02 review of the M5.4/M5.5 diff to find.go (found T39, T41, T43) |
| M5.2 | done   | 2026-08-02 | -         | navigator toggle (S9, D25); computeLayout gained hideNav; model.toggleNavigator/resize/layout; ctrl+n; 4 unit tests + 8 pty checks; fresh-context two-pass review logged T29-T36; re-covered by the 2026-08-02 review of the M5.4 diff to model.go's resize() (found T43, adjacent not blocking) |
| M5.3 | done   | 2026-08-02 | -         | help modal (S12, D26); help.go (helpBindings/helpBox/overlayHelp, lipgloss v2 Layer/Compositor); model.go helpOpen/openHelp/handleHelpInput; 10 unit tests + 11 pty checks; self fresh-context review found and fixed an unclamped-overlay overflow bug; independent fresh-context two-pass review found no fidelity issues, logged T29-T36 (mostly M5.1-scoped); re-covered by the 2026-08-02 review of the M5.4 diff to model.go (no help.go-specific findings; help.go itself not re-diffed since it was untouched by M5.4/M5.5) |
| M5.4 | done   | 2026-08-02 | -         | M5-review correctness/UX fixes (T28, T29, T34; D28): resolveLaunch/needsCwd in roots.go (T28), findMatches/foldLower in find.go (T29), model.resize's find-editing check (T34); 8 new/changed unit tests, each proven to fail against the pre-fix code by reverting; full check + check-pty (59) green; cleared M3.1/M3.5/M4.2/M5.1/M5.2/M5.3's Reviewed stamps (roots.go/find.go/model.go all changed since); two-pass fresh-context review (2026-08-02) found T37-T41 (standards) and T42 (spec fidelity, PLAN.md text stale against D28) — no blocking fidelity gaps, T28/T29/T34 confirmed delivered as D27/D28 describe |
| M5.5 | done    | 2026-08-02 | -         | replan C1/2026-08-02 (D27): M5-review standards cleanup (T30, T32, T33, T35); em-dash re-encode in find.go/navigator_test.go, keymap.go Cancel/Help tense fix, viewer.find renamed findState, FindNext/FindPrevious global-dispatch comment; cosmetic only, no behavior change; two-pass fresh-context review (2026-08-02) confirmed T30/T32/T33/T35 complete with no leftovers, found T43 (adjacent, pre-existing, not a regression) |
