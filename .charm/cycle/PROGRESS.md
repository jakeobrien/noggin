# Progress — C2a polish

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

Current task: M11.2 (README Global table gap, TRIAGE T19), a one-line follow-on to M11.1.
Once done, the closing `charm-validate` pass covers M7.3/M8.2/M11.1/M11.2/M12.1 — the only
work left before the cycle can close.

## Handoff

Replan (2026-08-09): M7.3/M8.2/M11.1/M12.1 are all done (see prior entries in git history /
DECISIONS.md D50/D51 for detail). TRIAGE.md's sole open item, T19 (README Global table omits
`r`/`f`/`R`), was routed to a new task M11.2 rather than deferred or dropped — a one-line
addition continuous with M11.1's already-established README-sync scope (D48), not a new
decision. TRIAGE.md is now empty.

Next step: do M11.2 (add `r`/`f`/`R` to README's Global keybinding table), then run the
closing `charm-validate` pass covering M7.3/M8.2/M11.1/M11.2/M12.1 together, from a fresh
context, per the skill's independence requirement. Nothing is blocked.

## Status

| Key   | Status | Reviewed | Validated | Note |
|-------|--------|----------|-----------|------|
| M1.1  | done   | 2026-08-06 | 2026-08-06 |      |
| M2.1  | done   | 2026-08-06 | 2026-08-06 |      |
| M2.2  | done   | 2026-08-06 | 2026-08-06 |      |
| M3.1  | done   | 2026-08-06 | 2026-08-06 |      |
| M4.1  | done   | 2026-08-06 | 2026-08-06 |      |
| M5.1  | done   | 2026-08-06 | 2026-08-06 |      |
| M6.1  | done   | 2026-08-06 | 2026-08-06 |      |
| M6.2  | done   | 2026-08-06 | 2026-08-06 |      |
| M7.1  | done   | 2026-08-07 | -         | see T13: broader S7 "cannot be searched with Find" gap (n/N ungated by peeking()) |
| M7.2  | done   | 2026-08-07 | -         | see T13: scroll-restore-on-clear violated in the same scenario |
| M8.1  | done   | 2026-08-07 | 2026-08-08 |      |
| M9.1  | done   | 2026-08-07 | 2026-08-08 |      |
| M10.1 | done   | 2026-08-08 | 2026-08-08 | perf clause ("no perceptible cost on a large Root") verified architecturally, not benchmarked |
| M7.3  | done   | -          | -          | fixes TRIAGE T13/T14/T16; see DECISIONS.md D51 |
| M8.2  | done   | -          | -          | fixes TRIAGE T18 |
| M11.1 | done   | -          | -          | README sync, TRIAGE T1/T6/T15; see DECISIONS.md D48 |
| M11.2 | todo   | -          | -          | README Global table gap, TRIAGE T19 |
| M12.1 | done   | -          | -          | test-coverage backfill, TRIAGE T2/T4/T5/T7/T17; see DECISIONS.md D50 |
