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

Current task: M4.11

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M4.10 (2026-08-12): closed the four review-flagged test-coverage gaps D83 bundled
(TRIAGE T25/T26/T29/T31) — no bug in any of the four, pure test additions.
`TestRecencyAndStarredNameTiesBreakByPath` (browse_list_test.go, two subtests) covers
`sortByRecency`'s and `starredRows`' own path tie-break for entries sharing an exact
timestamp/name (T26). `TestRecentRowsCapsOrphanBucketAtOneHundred` confirms `recentRows`'
per-bucket cap applies to the "" (orphan) bucket the same as any Root-keyed one, past 100
orphan documents (T29). `TestFilesNavViewHeightOneOmitsRowList` (browse_test.go) covers
`filesNav.view()`'s own height==1 short-circuit, the one branch every sibling already had a
test for (T25). `TestFilesNavRecentModeShowsStarMarker` asserts the star marker itself on a
starred Recent-mode row (present) and an unstarred one (absent), mirroring Starred mode's own
absence assertion (T31). All four exist and pass; `make check` green. No pty coverage needed —
plain-text assertions, no new rendering ambiguity.

Next: M4.11 — fix broken `\u2014` comment literals and the wrong "scope.go" citation (should
read browse.go) across `browse_list.go` and its test (TRIAGE T27/T28). Small, mechanical.

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
| M4.1 | done   | 2026-08-12 | -       | discoverability gap logged as TRIAGE T20; review found selection-index gap, TRIAGE T23 |
| M4.2 | done   | 2026-08-12 | -       | orphan-bucket cap + live-recompute judgment calls, D79; TRIAGE T21; review: T26, T29 |
| M4.3 | done   | 2026-08-12 | -       | flat-row Activate/Star wiring added beyond listing, D80; narrows TRIAGE T21; review: T23 |
| M4.4 | done   | 2026-08-12 | -       | render test walks all six mode/scope combos; TRIAGE T22 |
| M4.5 | done   | 2026-08-12 | -       | Peek fixed via togglePeekFlatRowInto; Expand/Collapse no-ops; TRIAGE T21; review: T23, T24, T25 |
| M4.6 | done   | 2026-08-12 | -       | review found selection-index restore gap, TRIAGE T23                |
| M4.7 | done   | -        | -         | replan: flat-row selection re-clamp bug, TRIAGE T23/T24 (D82); drops T21 as already-handled |
| M4.8 | done   | -        | -         | replan: Browse cycle keys' footer/help discoverability, TRIAGE T20 (D81) |
| M4.9 | done   | -        | -         | replan: Tree's own empty-vs-no-Root message, TRIAGE T22  |
| M4.10 | done  | -        | -         | replan: bundled test-coverage gaps, TRIAGE T25/T26/T29/T31 (D83) |
| M4.11 | todo  | -        | -         | replan: comment-literal and citation fixes, TRIAGE T27/T28 |
| M5.1 | todo   | -        | -         | Grep-side `Starred` is pre-agreed cut 4  |
| M6.1 | todo   | -        | -         |                                          |
| M6.2 | todo   | -        | -         | pre-agreed cut 2                         |
| M7.1 | todo   | -        | -         |                                          |
| M7.2 | todo   | -        | -         |                                          |
