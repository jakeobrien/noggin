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

Current task: M1.4c

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

Replan (this session, fresh context): reviewed the M1.4a/M1.4b/M2.3a/M3.1a/M3.2a/M3.2b/M3.3a/
M3.3b diff (each task's Reviewed stamp was `-`) and triaged the resulting TRIAGE T13-T20.
Three small tasks inserted ahead of M4.1, cursor moved to the first:
- **M1.4c** (low, no decision) — grepArgMaxBytes parallel-safety hygiene + a missing
  combined cap/batch-boundary test (T18, T19).
- **M2.1a** (D51, medium) — de-duplicate openPicker/dismissPicker's reset-carry-over
  sequence before M4.x can add a third reset path and independently regress D48 again (T14).
- **M3.1b** (D52, medium) — pickerRootLabel's budget calc double-subtracts a gap, making
  M3.1a/D50's disambiguation fix less effective than designed at realistic (80x24) widths;
  also fixes the two test-quality gaps that masked it, folding in T17 (T13, T15, T16, T17).
T20 dropped (not promoted): reviewer's own report called it a defensible deliberate choice,
no functional gap.

M3 (File Picker end to end) and all its original sub-tasks are done; M1.4a/b, M2.3a,
M3.1a/M3.2a/M3.2b/M3.3a/M3.3b (the prior review's fix-ups) are also done and now reviewed.

Next step: M1.4c, then M2.1a, then M3.1b (each effort 2, no code written yet this session) —
then M4.1: wire Grep mode to M1.4. M4.2 reuses M3.2's preview mechanism (fixed by M3.2a, but
see M2.1a); before M4.3/M4.4, re-read D34 (rg line numbers vs. rendered line space).

Also queued, non-blocking: M7.1a (T11, terminology housekeeping).

No blockers. TRIAGE.md empty.

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
| M1.4a| done   | 2026-08-03 | -         | D43 (TRIAGE T1) |
| M1.4b| done   | 2026-08-03 | -         | D45 (TRIAGE T3) |
| M1.4c| todo   | -        | -         | D-none; TRIAGE T18, T19 |
| M2.1a| todo   | -        | -         | D51; TRIAGE T14 |
| M2.3a| done   | 2026-08-03 | 2026-08-03 | D46 (TRIAGE T5); M2.3 re-validated |
| M3.1 | done   | 2026-08-03 | -         |      |
| M3.1a| done   | 2026-08-03 | -         | D50 (TRIAGE T10); re-reviewed, T13/T16 raised |
| M3.1b| todo   | -        | -         | D52; TRIAGE T13, T15, T16 |
| M3.2 | done   | 2026-08-03 | -         |      |
| M3.2a| done   | 2026-08-03 | -         | D48 (TRIAGE T6); fixed before M4.2 |
| M3.2b| done   | 2026-08-03 | -         | TRIAGE T12; re-reviewed, T20 raised |
| M3.3 | done   | 2026-08-03 | -         |      |
| M3.3a| done   | 2026-08-03 | -         | D49 (TRIAGE T7, T8) |
| M3.3b| done   | 2026-08-03 | -         | D49 (TRIAGE T9) |
| M4.1 | todo   | -        | -         |      |
| M4.2 | todo   | -        | -         |      |
| M4.3 | todo   | -        | -         |      |
| M4.4 | todo   | -        | -         |      |
| M5.1 | todo   | -        | -         |      |
| M5.2 | todo   | -        | -         |      |
| M5.3 | todo   | -        | -         |      |
| M6.1 | todo   | -        | -         | pre-agreed cut (D32) |
| M6.2 | todo   | -        | -         | pre-agreed cut (D32) |
| M7.1 | todo   | -        | -         |      |
| M7.1a| todo   | -        | -         | TRIAGE T11 |
| M7.2 | todo   | -        | -         |      |
