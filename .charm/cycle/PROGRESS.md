# Progress — C6 sources

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

Current task: M7.4

## Handoff

**Build 2026-09-09 (M2.11):** M2.10's not-parallel-safe mutation-contract doc comment was
duplicated near-verbatim across `internal/wikipedia/fetch_test.go`
(`TestClientFetchViaNewClientDelegatesToFetchAt`) and `search_test.go`
(`TestClientSearchDelegatesToSearchTitles`), TRIAGE T31. Made fetch_test.go's copy canonical
(kept the full contract statement, now says so explicitly and names search_test.go's test as
pointing back to it); replaced search_test.go's copy with a short pointer to the canonical
statement instead of restating it, so there's exactly one place that describes the rule in
full — no wording drift possible since only one copy has the wording. `make check`
(build/vet/fmt/test) green, `internal/wikipedia` package tests pass (0.059s, not cached, ran
fresh).

Next step: cursor moved to M7.4 (next todo task in key order; independent,
pull-forward-anytime hygiene task per PLAN.md — stale/mis-cited store.go doc comments). M8.2a,
M8.4, M9.2 remain independent and can be pulled in any order; M10.1 (README, last task) should
come after those since it documents the behaviour/citations they touch. No blockers.

## Status

| Key   | Status | Reviewed | Validated | Note |
|-------|--------|----------|-----------|------|
| M1.1  | done   | 2026-09-08 | 2026-09-08 |      |
| M1.2  | done   | 2026-09-08 | 2026-09-08 |      |
| M1.3  | done   | 2026-09-08 | 2026-09-08 |      |
| M1.4  | done   | 2026-09-08 | 2026-09-08 |      |
| M1.5  | done   | 2026-09-08 | 2026-09-08 |      |
| M2.1  | done   | 2026-09-08 | 2026-09-08 |      |
| M2.2  | done   | 2026-09-08 | 2026-09-08 |      |
| M2.3  | done   | 2026-09-08 | 2026-09-08 |      |
| M1.6  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T1, T2) |
| M1.7  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T3) |
| M2.4  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T4, T5) |
| M2.5  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T6, T7, T8) |
| M2.6  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T10) |
| M2.7  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T11, T12) |
| M1.8  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T17) |
| M2.8  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T15) |
| M3.1  | done   | 2026-09-08 | 2026-09-08 |      |
| M3.2  | done   | 2026-09-08 | 2026-09-08 |      |
| M3.3  | done   | 2026-09-08 | 2026-09-08 |      |
| M3.4  | done   | 2026-09-08 | 2026-09-08 |      |
| M3.5  | done   | 2026-09-08 | 2026-09-08 |      |
| M4.1  | done   | 2026-09-08 | 2026-09-08 |      |
| M4.2  | done   | 2026-09-08 | 2026-09-08 |      |
| M4.3  | done   | 2026-09-08 | 2026-09-08 |      |
| M5.1  | done   | 2026-09-08 | 2026-09-08 |      |
| M5.2  | done   | 2026-09-08 | 2026-09-08 |      |
| M1.9  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T18) |
| M1.10 | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T2) |
| M2.9  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T1, T22) |
| M2.10 | done   | 2026-09-09 | 2026-09-09 | added by replan 2026-09-08 (TRIAGE T2) |
| M2.11 | done   | -        | -         | added by replan 2026-09-09 (TRIAGE T31) |
| M3.6  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T20) |
| M3.7  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T21) |
| M4.4  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T23) |
| M1.11 | done   | -        | -         | added by replan 2026-09-08 (TRIAGE T1, T3); checks.py:164 fix closes TRIAGE T30 |
| M5.3  | done   | 2026-09-09 | 2026-09-09 | added by replan 2026-09-08 (TRIAGE T4) |
| M6.1  | done   | 2026-09-09 | 2026-09-08 | pre-agreed cut 1 |
| M7.1  | done   | 2026-09-08 | 2026-09-08 | pre-agreed cut 2 |
| M7.2  | done   | 2026-09-08 | 2026-09-08 | pre-agreed cut 2 |
| M7.3  | done   | 2026-09-08 | 2026-09-08 | pre-agreed cut 2 |
| M7.4  | todo   | -        | -         | added by replan 2026-09-09 (TRIAGE T24, T25) |
| M8.1  | done   | 2026-09-09 | 2026-09-09 |      |
| M8.2  | done   | 2026-09-09 | 2026-09-09 | scope widened by replan 2026-09-08: S9a attribution line |
| M8.2a | todo   | -        | -         | added by replan 2026-09-09 (TRIAGE T26, T28) |
| M8.3  | done   | 2026-09-09 | 2026-09-09 |      |
| M8.4  | todo   | -        | -         | added by replan 2026-09-09 (TRIAGE T27) |
| M9.1  | done   | 2026-09-09 | 2026-09-09 | pre-agreed cut 3 |
| M9.2  | todo   | -        | -         | added by replan 2026-09-09 (TRIAGE T29) |
| M10.1 | todo   | -        | -         |      |
