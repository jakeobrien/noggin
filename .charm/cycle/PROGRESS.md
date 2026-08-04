# Progress — C2 find

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

No task in progress. M1-M5 are done; M4.2a and M7.1b (inserted by the 2026-08-04 replan) plus
M6/M7 remain todo. The M5 validation gate (below) is the recommended next step.

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M1-M5 are built; M4.4 and M5.1-M5.3 landed most recently (see git history for detail). M5 has
no charm-validate pass yet (Reviewed/Validated both `-` on M5.1-M5.3) — a completed milestone
with validation due is a process gate; address before continuing into M6/M7.

2026-08-04 replan: drained TRIAGE.md (T1-T4, T21-T24; inbox now empty), inserting M4.2a (D55:
shared `styleRow` helper + grep-row test fix, before M6's third row renderer) and M7.1b
(bundled low-cost test-hygiene backfill, D56). No spec/scope change. DECISIONS.md gained D55, D56.

Next step: charm-validate M5.1-M5.3, then M4.2a before M6 (M6 is D32's pre-agreed cut — skip to
M7 if time is short). No blockers, no in-flight changes.

## Status

| Key  | Status | Reviewed | Validated | Note |
|------|--------|----------|-----------|------|
| M1.1 | done   | 2026-08-03 | 2026-08-03 |  |
| M1.2 | done   | 2026-08-03 | 2026-08-03 |  |
| M1.3 | done   | 2026-08-03 | 2026-08-03 |  |
| M1.4 | done   | 2026-08-03 | 2026-08-03 |  |
| M2.1 | done   | 2026-08-03 | 2026-08-03 |      |
| M2.2 | done   | 2026-08-03 | 2026-08-03 |      |
| M2.3 | done   | 2026-08-03 | 2026-08-03 | M2.3a landed; re-validated |
| M2.4 | done   | 2026-08-03 | 2026-08-03 |      |
| M2.5 | done   | 2026-08-03 | 2026-08-03 |      |
| M1.4a| done   | 2026-08-03 | 2026-08-03 | D43 (TRIAGE T1) |
| M1.4b| done   | 2026-08-03 | 2026-08-03 | D45 (TRIAGE T3) |
| M1.4c| done   | 2026-08-04 | 2026-08-03 | D-none; TRIAGE T18, T19; TRIAGE T23 |
| M2.1a| done   | 2026-08-04 | 2026-08-03 | D51; TRIAGE T14 |
| M2.3a| done   | 2026-08-03 | 2026-08-03 | D46 (TRIAGE T5); M2.3 re-validated |
| M3.1 | done   | 2026-08-03 | 2026-08-03 |      |
| M3.1a| done   | 2026-08-03 | 2026-08-03 | D50 (TRIAGE T10); re-reviewed, T13/T16 raised; TRIAGE T3 (disambiguation test compares full frame lines, not isolated row) |
| M3.1b| done   | 2026-08-04 | 2026-08-03 | D52; TRIAGE T13, T15, T16, T17 |
| M3.2 | done   | 2026-08-03 | 2026-08-03 |      |
| M3.2a| done   | 2026-08-03 | 2026-08-03 | D48 (TRIAGE T6); fixed before M4.2 |
| M3.2b| done   | 2026-08-03 | 2026-08-03 | TRIAGE T12; re-reviewed, T20 raised |
| M3.3 | done   | 2026-08-03 | 2026-08-03 |      |
| M3.3a| done   | 2026-08-03 | 2026-08-03 | D49 (TRIAGE T7, T8); grep-mode no-op test superseded/removed by M4.3 (897ef6a) once grep Activate became real — not a regression, criterion-as-written no longer checkable |
| M3.3b| done   | 2026-08-03 | 2026-08-03 | D49 (TRIAGE T9); TRIAGE T4 (fixture's expanded map stays empty, doesn't exercise the substitution scenario) |
| M4.1 | done   | 2026-08-04 | 2026-08-03 | D53; TRIAGE T21, T22, T24 |
| M4.2 | done   | 2026-08-04 | 2026-08-03 | D54; TRIAGE T24 |
| M4.2a| todo   | -        | -         | D55 (TRIAGE T21, T22); address before M6 |
| M4.3 | done   | 2026-08-03 | 2026-08-03 | M4.3 test coupling note, see handoff |
| M4.4 | done   | 2026-08-04 | 2026-08-03 |      |
| M5.1 | done   | 2026-08-04 | -         |      |
| M5.2 | done   | 2026-08-04 | -         |      |
| M5.3 | done   | 2026-08-04 | -         |      |
| M6.1 | todo   | -        | -         | pre-agreed cut (D32) |
| M6.2 | todo   | -        | -         | pre-agreed cut (D32) |
| M7.1 | todo   | -        | -         |      |
| M7.1a| todo   | -        | -         | TRIAGE T11, T24 |
| M7.1b| todo   | -        | -         | TRIAGE T1, T2, T3, T4, T23; D56 |
| M7.2 | todo   | -        | -         |      |
