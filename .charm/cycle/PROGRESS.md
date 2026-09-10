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

Current task: M8.2b done (build complete this session) — the cycle's plan work is now
fully built. All of M1-M10 plus M1.12/M9.3/M8.2b are `done`; validation is still
outstanding for 10 tasks (see below).

## Handoff

**Build 2026-09-10 (M8.2b):** added
`TestPeekLocationLineNeverClaimsARowAtAPathologicallyShortTerminal` (location_test.go),
the Peek-path counterpart to M8.2a's own
`TestLocationLineNeverClaimsARowAtAPathologicallyShortTerminal` — same shape (a
`sized(t, m, 80, 4)` terminal, Navigator hidden, frame-height and bottom-border
assertions), built on `peekFixture` (peek_test.go) instead of a bare `openDocument`,
since Peek needs a second file to select. Asserts `peekLocationVisible()` is false,
`peekContentHeight()` is 0, and `viewerLocationLine(dv1Pane)` is empty at `boxHeight`
0, then confirms the frame still closes with no overflow or misaligned border.
Sanity-checked the test actually catches the regression: temporarily reverting
`peekLocationVisible`'s `boxHeight >= 1` guard (location.go) made the new test fail
with the expected messages; restored, `make check` green (build, vet, fmt, test all
pass). No code behavior changed — a pure test-gap fill matching M8.2a's own precedent.
No new findings, no TRIAGE items, no DECISIONS entries.

Next step: charm-validate (fresh context) covering all 10 previously-unvalidated tasks
(M1.11, M2.11, M7.4, M8.2a, M8.4, M9.2, M8.5, M1.12, M9.3, M8.2b) and confirm the whole
cycle (M1-M10) remains green, then `charm-security-review`, then `/charm-close`. No
blockers.

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
| M2.11 | done   | 2026-09-10 | -         | added by replan 2026-09-09 (TRIAGE T31) |
| M3.6  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T20) |
| M3.7  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T21) |
| M4.4  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T23) |
| M1.11 | done   | 2026-09-10 | -         | added by replan 2026-09-08 (TRIAGE T1, T3); checks.py:164 fix closes TRIAGE T30 |
| M5.3  | done   | 2026-09-09 | 2026-09-09 | added by replan 2026-09-08 (TRIAGE T4) |
| M6.1  | done   | 2026-09-09 | 2026-09-08 | pre-agreed cut 1 |
| M7.1  | done   | 2026-09-08 | 2026-09-08 | pre-agreed cut 2 |
| M7.2  | done   | 2026-09-08 | 2026-09-08 | pre-agreed cut 2 |
| M7.3  | done   | 2026-09-08 | 2026-09-08 | pre-agreed cut 2 |
| M7.4  | done   | 2026-09-10 | -         | added by replan 2026-09-09 (TRIAGE T24, T25) |
| M8.1  | done   | 2026-09-09 | 2026-09-09 |      |
| M8.2  | done   | 2026-09-09 | 2026-09-09 | scope widened by replan 2026-09-08: S9a attribution line |
| M8.2a | done   | 2026-09-10 | -         | added by replan 2026-09-09 (TRIAGE T26, T28) |
| M8.2b | done   | -          | -         | added by replan 2026-09-10 (TRIAGE T36) |
| M8.3  | done   | 2026-09-09 | 2026-09-09 |      |
| M8.4  | done   | 2026-09-10 | -         | added by replan 2026-09-09 (TRIAGE T27) |
| M9.1  | done   | 2026-09-09 | 2026-09-09 | pre-agreed cut 3 |
| M9.2  | done   | 2026-09-10 | -         | added by replan 2026-09-09 (TRIAGE T29) |
| M10.1 | done   | 2026-09-09 | 2026-09-09 |      |
| M8.5  | done   | 2026-09-10 | -         | added by replan 2026-09-09 (TRIAGE T32) |
| M1.12 | done   | 2026-09-10 | -         | added by replan 2026-09-10 (TRIAGE T33) |
| M9.3  | done   | 2026-09-10 | -         | added by replan 2026-09-10 (TRIAGE T34) |
