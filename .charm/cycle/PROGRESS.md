# Progress — C1 read-across-roots

## Cursor

Current task: M5.8 (T45, T46; D30) — not yet built, no blockers. Supersedes T44 (dropped, see
D30). Still outstanding once M5.8 lands: M5.6 and M5.7's first code-review `Reviewed` stamp
(every other key in that span already carries one) — likely what lets the cycle close.

## Handoff

M5.7 (T38-T41; D29) built and validated this session: `needsCwd`'s unused `cwd` param dropped,
`findMatches`' span-translation named as an `origSpan` helper, `resolveLaunch` gained
zero/many-arg `cwdErr` tests, and a fold-within-the-match regression test was added. T39/T41
proven meaningful by reverting and re-running; `make check` + `make check-pty` (59 checks)
green throughout, no behavior change.

The follow-on review/validate pass on M5.6 then surfaced T44 (validate, coverage question),
T45 (review, a factually wrong "window resized" message on the `toggleNavigator` path T44
asked about) and T46 (review, an untested `resolveLaunch` branch) — routed by this replan into
M5.8 (D30); TRIAGE.md is empty again.

Still relevant:
- Go with the grain (D1) now means Charm v2; D13's no-terminal-query-after-`Run()` rule
  stands (`markdownStyle` resolves dark/light before `Run()`).
- Dependency graph: 31 → 24 indirect deps on the v2 migration; M5.1 added
  `charmbracelet/x/ansi` direct and pulled in `atotto/clipboard` indirectly (unused) — noted
  for security-review time.
- Sandbox Go toolchain fully working; the Makefile's `unexport GOROOT` is defensive-only.

## Status

| Key  | Status | Reviewed   | Validated  | Note |
|------|--------|------------|------------|-------|
| M1.1 | done   | 2026-08-01 | 2026-07-31 | re-verified on v2 by M1.4 |
| M1.2 | done   | 2026-08-01 | 2026-07-31 | re-verified on v2 by M1.4 |
| M1.3 | done   | 2026-08-01 | 2026-07-31 | re-verified on v2 by M1.4 |
| M1.4 | done   | 2026-08-01 | 2026-07-31 | M1 complete and validated |
| M1.5 | done   | 2026-07-31 | 2026-07-31 | two-pass review (T9, T10 logged); validated (fresh-context subagent) |
| M2.1 | done   | 2026-07-31 | 2026-07-31 | internal/tree: policy + walk + hand-rolled gitignore (D18); review found T11-T14; validated |
| M2.2 | done   | 2026-07-31 | 2026-07-31 | Files Nav pane: filesNav over immutable tree (D19); 8 unit tests + 8 pty checks; validated |
| M2.3 | done   | 2026-07-31 | 2026-07-31 | Activate binding (enter): filesNav → loadDocument → openDocument; 3 unit tests + 5 pty checks; review found T15; validated |
| M2.4 | done   | 2026-08-01 | 2026-08-01 | internal/tree hardening (T11-T14, D23): dead branch dropped, doc comments aligned, name matching case-insensitive, escape test added |
| M3.1 | done   | 2026-08-02 | 2026-08-02 | S1 launch contract; roots.go resolveRoots (D20, T7 fixed); 9 subtests + pty spot-check; re-covered by the 2026-08-02 review of the M5.4 diff to roots.go (found T37, T38, T40) |
| M3.2 | done   | 2026-07-31 | 2026-07-31 | starred-roots store (D21); store.go load/save/toggle + loadStarredRoots wired into main.go; 9 tests |
| M3.3 | done   | 2026-07-31 | 2026-07-31 | Roots Nav pane (D22); rootsnav.go + model.{activateRoot,toggleStar}; 11 unit tests + 26 pty checks |
| M3.4 | done   | 2026-07-31 | 2026-07-31 | test-only: proved via `TestActivateRootLeavesOpenDocumentUntouched`, no prod code change |
| M3.5 | done   | 2026-08-02 | 2026-08-02 | replan C1/2026-08-01: roots/messaging correctness fixes (T17, T18, T20); review found T28 (fixed by M5.4); re-covered by the 2026-08-02 review of the M5.4 diff to roots.go (found T37, T38, T40) |
| M3.6 | done   | 2026-08-01 | 2026-08-01 | replan C1/2026-08-01: M2/M3 code-review cleanup batch (T9, T10, T16, T19, T21-T25, T27) |
| M4.1 | done   | 2026-08-01 | 2026-08-01 | edit handoff (S10/D3): tea.ExecProcess release/resume + finishEdit re-render from disk |
| M4.2 | done   | 2026-08-02 | 2026-08-02 | footer key hints (S11); T15 (activate read-error message) and T26 (empty-Roots-Nav guidance) folded in; re-covered by the 2026-08-02 review of the M5.4 diff to model.go (found T43, adjacent not blocking) |
| M5.1 | done   | 2026-08-02 | 2026-08-02 | find in document (S8); find.go + keymap/model wiring; 8 unit tests + 5 pty checks; found and fixed a real bubbles/v2 SetHighlights bug along the way (see findings); fresh-context two-pass review logged T29-T36; re-covered by the 2026-08-02 review of the M5.4/M5.5 diff to find.go (found T39, T41, T43) |
| M5.2 | done   | 2026-08-02 | 2026-08-02 | navigator toggle (S9, D25); computeLayout gained hideNav; model.toggleNavigator/resize/layout; ctrl+n; 4 unit tests + 8 pty checks; fresh-context two-pass review logged T29-T36; re-covered by the 2026-08-02 review of the M5.4 diff to model.go's resize() (found T43, adjacent not blocking) |
| M5.3 | done   | 2026-08-02 | 2026-08-02 | help modal (S12, D26); help.go (helpBindings/helpBox/overlayHelp, lipgloss v2 Layer/Compositor); model.go helpOpen/openHelp/handleHelpInput; 10 unit tests + 11 pty checks; self fresh-context review found and fixed an unclamped-overlay overflow bug; independent fresh-context two-pass review found no fidelity issues, logged T29-T36 (mostly M5.1-scoped); re-covered by the 2026-08-02 review of the M5.4 diff to model.go (no help.go-specific findings; help.go itself not re-diffed since it was untouched by M5.4/M5.5) |
| M5.4 | done   | 2026-08-02 | 2026-08-02 | M5-review correctness/UX fixes (T28, T29, T34; D28): resolveLaunch/needsCwd in roots.go (T28), findMatches/foldLower in find.go (T29), model.resize's find-editing check (T34); 8 new/changed unit tests, each proven to fail against the pre-fix code by reverting; full check + check-pty (59) green; cleared M3.1/M3.5/M4.2/M5.1/M5.2/M5.3's Reviewed stamps (roots.go/find.go/model.go all changed since); two-pass fresh-context review (2026-08-02) found T37-T41 (standards) and T42 (spec fidelity, PLAN.md text stale against D28) — no blocking fidelity gaps, T28/T29/T34 confirmed delivered as D27/D28 describe |
| M5.5 | done   | 2026-08-02 | 2026-08-02 | replan C1/2026-08-02 (D27): M5-review standards cleanup (T30, T32, T33, T35); em-dash re-encode in find.go/navigator_test.go, keymap.go Cancel/Help tense fix, viewer.find renamed findState, FindNext/FindPrevious global-dispatch comment; cosmetic only, no behavior change; two-pass fresh-context review (2026-08-02) confirmed T30/T32/T33/T35 complete with no leftovers, found T43 (adjacent, pre-existing, not a regression) |
| M5.6 | done   | 2026-08-02 | 2026-08-02 | replan C1/2026-08-03 (D29): roots/find correctness fixes (T37, T43) — needsCwd/resolveLaunch now share one stat via resolvePathArg/baseResolution; model.resize's hasFindMatches check surfaces "search cleared" for a committed find lost to a re-render; new test proven to fail pre-fix; validate (2026-08-02) found T43's acceptance text names resize-or-Navigator-toggle but only resize is directly tested (T44 logged); fresh-context two-pass review (2026-08-02) confirmed T37/T43 delivered as specified but found T45 (the toggleNavigator path T44 flagged as untested does fire, and produces a factually wrong "window resized" message) and T46 (T37's refactor added an untested not-found-absolute-arg code path) |
| M5.7 | done   | 2026-08-02 | 2026-08-02 | replan C1/2026-08-03 (D29): standards/test-hardening cleanup (T38, T39, T40, T41); needsCwd's cwd param dropped, findMatches' span shape named as origSpan, resolveLaunch gained zero/many-arg cwdErr tests, fold-within-match regression test added; fresh-context two-pass review (2026-08-02) confirmed T38-T41 complete with no leftovers, no fidelity issues |
