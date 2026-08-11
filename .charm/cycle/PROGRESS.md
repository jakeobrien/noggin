# Progress — C4 recall

<!-- Per-cycle. Cursor and statuses ONLY — no work summaries (git history carries those).
     The Handoff section is the single exception: overwrite it each session, never append.
     Statuses: todo | in-progress | done | blocked | deferred.
     Blocked rows reference the DECISIONS.md / TRIAGE.md item they wait on — the set of
     blocked rows is the cycle's blocking-issue list.

     Reviewed / Validated columns: date (yyyy-mm-dd) the task was last covered by code
     review / validate, or `-` if it never has been. Set per task key, by the role that
     performed the check — or the parent that dispatched it as a subagent — for exactly the
     tasks that check actually covered, never assumed for a whole milestone. This is what
     makes reviewed/validated state accurate across replans that insert tasks earlier,
     partial mid-milestone checks, and skipped end-of-milestone gates — a single cycle-wide
     checkpoint can't represent any of those. Clear a cell back to `-` when the task's code
     changes after that date (fix-ups, reopened work) — it needs re-covering. -->

## Cursor

Current task: M5.2

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M4.15 built: fixed `filesNav.setRoot`'s selection-reset bug (TRIAGE T38) — its own
`clampSelected()` call ran before `model.openFolder`'s later `refreshFilesRecentSource`
re-points `f.activeRootPath` at the incoming Root, so a flat-mode (Recent/Starred) Active
Root restore re-clamped against the *outgoing* Root's still-stale bucket size, collapsing the
memory-restored row to 0. Fix: `setRoot` now sets `f.activeRootPath = root.Path` itself,
immediately after `f.root = root` and before the memory-restore/clamp block — its one caller
(`model.openFolder`) only ever passes a root that already is, or is about to become, the
session's Active Root, so this is exact, not a guess; `docs`/`sessionRoots` need no equivalent
fix since neither varies with which Root is active. Doc comment on `setRoot` updated to cite
the hand-off explicitly.

New: openfolder_flat_restore_test.go's `TestOpenFolderRestoresFlatModeSelectionAgainstThe-
CorrectRoot` drives `model.openFolder` itself (not `filesNav.setRoot` in isolation, which
`TestSettingRootRestoringAFlatModeReClampsSelection` already covers for a different shrink and
explicitly deferred this exact ordering hand-off in its own doc comment) — two Roots with
differing Starred/Active-Root bucket sizes, switching Root A (Starred, row 2 selected) away to
Root B and back, asserting selection lands back on row 2 (a3.md), not 0. Confirmed failing
pre-fix (asserted `selected == 0`, `selectedFlatRow` returned a1.md) and passing post-fix.
`make check` and `make check-pty` both green; `go test ./... -v` shows 422 passing, 0 failing.

Cursor moves to M5.2 (delete dead `grepFiles`, D94), next in the four-item cleanup queue
(M5.2, M6.3, M6.4) this session's replan sequenced ahead of M7.

## Status

| Key  | Status | Reviewed | Validated | Note                                     |
|------|--------|----------|-----------|------------------------------------------|
| M1.1 | done   | 2026-08-10 | 2026-08-10 |                                       |
| M1.2 | done   | 2026-08-10 | 2026-08-10 |                                       |
| M1.3 | done   | 2026-08-10 | 2026-08-10 |                                       |
| M1.4 | done   | 2026-08-10 | 2026-08-10 |                                       |
| M2.1 | done   | 2026-08-10 | 2026-08-11 |                                      |
| M2.2 | done   | 2026-08-10 | 2026-08-11 |                                      |
| M2.3 | done   | 2026-08-10 | 2026-08-11 |                                      |
| M2.4 | done   | 2026-08-10 | 2026-08-11 |                                      |
| M2.5 | done   | 2026-08-10 | 2026-08-11 | pre-agreed cut 1, not actually cut   |
| M2.6 | done   | 2026-08-10 | 2026-08-11 |                                      |
| M1.1a | done  | 2026-08-11 | 2026-08-11 | replan: pty harness fix, TRIAGE T4 (D71) |
| M2.7 | done   | 2026-08-11 | 2026-08-11 | replan: stale marker fix, TRIAGE T7 (D70) |
| M2.8 | done   | 2026-08-11 | 2026-08-11 | replan: star-binding discoverability, TRIAGE T5/T6 |
| M2.9 | done   | 2026-08-11 | 2026-08-11 | replan: naming cleanup, TRIAGE T1/T2/T9  |
| M3.1 | done   | 2026-08-11 | 2026-08-10 | SPEC/GLOSSARY-level gap re: Picker Activate not recording lastViewed — tracked as TRIAGE T12, not a M3.1 overclaim |
| M3.2 | done   | 2026-08-10 | 2026-08-10 |                                          |
| M3.3 | done   | 2026-08-10 | 2026-08-10 | replan: Picker never records lastViewed, TRIAGE T12 (D74) |
| M3.4 | done   | 2026-08-10 | 2026-08-10 | replan: recordViewed failure message clobber, TRIAGE T13; no bug found, test coverage added; review found coverage gap re: Picker call sites, resolved as M3.5 (D77) |
| M3.5 | done   | 2026-08-10 | 2026-08-11 | replan: Picker-path store-write-failure regression tests, TRIAGE T18 (D77) |
| M2.10 | done  | 2026-08-10 | 2026-08-11 | replan: pty star-marker column checks, TRIAGE T11        |
| M2.11 | done  | 2026-08-10 | 2026-08-10 | replan: Picker star marker (D72, TRIAGE T8), built same session; review found stale doc comment, resolved as M2.12 (D76); validate found Grep-mode live-toggle coverage gap, resolved as M2.13 (D78) |
| M2.12 | done  | 2026-08-10 | 2026-08-11 | replan: correct starMarker() doc comment, TRIAGE T17 (D76) |
| M2.13 | done  | 2026-08-10 | 2026-08-11 | replan: Grep same-frame star-toggle render test, TRIAGE T19 (D78) |
| M4.1 | done   | 2026-08-12 | 2026-08-11 | discoverability gap logged as TRIAGE T20; review found selection-index gap, TRIAGE T23 |
| M4.2 | done   | 2026-08-12 | 2026-08-11 | orphan-bucket cap + live-recompute judgment calls, D79; TRIAGE T21; review: T26, T29 |
| M4.3 | done   | 2026-08-12 | 2026-08-11 | flat-row Activate/Star wiring added beyond listing, D80; narrows TRIAGE T21; review: T23; pre-agreed cut 3, not actually cut (D93, cf. M2.5/M5.1) |
| M4.4 | done   | 2026-08-12 | 2026-08-11 | render test walks all six mode/scope combos; TRIAGE T22 |
| M4.5 | done   | 2026-08-12 | 2026-08-11 | Peek fixed via togglePeekFlatRowInto; Expand/Collapse no-ops; TRIAGE T21; review: T23, T24, T25 |
| M4.6 | done   | 2026-08-12 | 2026-08-11 | review found selection-index restore gap, TRIAGE T23                |
| M4.7 | done   | 2026-08-12 | 2026-08-11 | replan: flat-row selection re-clamp bug, TRIAGE T23/T24 (D82); drops T21 as already-handled; review: T37 |
| M4.8 | done   | 2026-08-12 | 2026-08-11 | replan: Browse cycle keys' footer/help discoverability, TRIAGE T20 (D81); review: T32/T33 |
| M4.9 | done   | 2026-08-12 | 2026-08-11 | replan: Tree's own empty-vs-no-Root message, TRIAGE T22; review: T34  |
| M4.10 | done  | 2026-08-12 | 2026-08-11 | replan: bundled test-coverage gaps, TRIAGE T25/T26/T29/T31 (D83); review: T36 |
| M4.11 | done  | 2026-08-12 | 2026-08-11 | replan: comment-literal and citation fixes, TRIAGE T27/T28; review: T35 |
| M4.12 | done  | 2026-08-11 | 2026-08-11 | replan: M4.8 test-coverage gaps, TRIAGE T32/T33 (D86)    |
| M4.13 | done  | 2026-08-11 | 2026-08-11 | replan: flatEmptyText default branch + tie-break comment, TRIAGE T34/T36 (D87); review: T39 |
| M4.14 | done  | 2026-08-11 | 2026-08-11 | replan: M4.7 acceptance-wording literal test gap, TRIAGE T37 (D88) |
| M5.1 | done   | 2026-08-11 | -         | delivered both File- and Grep-side Starred (D89, not actually cut, cf. M2.5); review found grepFiles dead code, TRIAGE T42 |
| M6.1 | done   | 2026-08-11 | -         |                                          |
| M6.2 | done   | 2026-08-11 | -         | pre-agreed cut 2, not actually cut (D91); review found untested re-Activate scroll-preserve behavior, TRIAGE T43 |
| M4.15 | done  | -        | -         | replan: openFolder flat-mode selection-reset bug, TRIAGE T38 |
| M5.2 | todo   | -        | -         | replan: delete dead grepFiles, TRIAGE T42 (D94) |
| M6.3 | todo   | -        | -         | replan: HistoryBack/Forward Footer/Help discoverability, TRIAGE T41 |
| M6.4 | todo   | -        | -         | replan: D91 re-Activate scroll-preserve regression test, TRIAGE T43 |
| M7.1 | todo   | -        | -         |                                          |
| M7.2 | todo   | -        | -         |                                          |
