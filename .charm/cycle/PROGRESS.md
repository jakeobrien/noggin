# Progress — C3 compare

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

Current task: (none — all M1–M5 tasks done; cycle close/validation due)

## Handoff

Fresh-context charm-validate pass run this session (cycle-end scope) via the validator
subagent. `make check` and `make check-pty` both pass clean. M1.6a, M1.7, M2.1, M4.1, M5.1,
M5.2 each independently verified against their PLAN.md done-when and stamped Validated below
(2026-08-09). M1.1–M1.6/M1.3a/M1.4/M1.4a/M1.5 were already stamped Validated earlier today and
not re-run (no code change since; relied on indirectly via tests that build on them).

**M3.1 is only PARTIAL**, not a clean pass: TRIAGE T11 (already open, source `review`) was
reproduced live — `refresh()`'s message-combining switch (cmd/noggin/model.go:651-658) has no
`if m.message == ""` guard (unlike `refreshViewer`'s own guarded branch at model.go:685), so a
read-failure message is silently overwritten by a find-dropped message whenever both occur
together. Violates M3.1's own "message-priority behavior" clause and SPEC.md's cycle-level
acceptance bullet 5. T11 already fully describes the fix (guard the switch like the
pre-refactor code and `refreshViewer` do) — no new TRIAGE entry added — but its severity
should read as cycle-close-blocking, not deferrable. Recommend charm-replan promote it to a
fix task before C3 closes.

Next: no task remains in PLAN.md. Cycle-level validation is otherwise clean (make check/
check-pty pass, all other M-keys independently verified) — C3 can close once T11 is fixed
(or explicitly wontfix'd, which would need a DECISIONS.md entry since it contradicts a stated
SPEC acceptance bullet). TRIAGE.md still has T10/T12/T13 open too (all low/med, non-blocking
by prior triage) — must be empty before close regardless.

## Status

| Key  | Status | Reviewed | Validated | Note |
|------|--------|----------|-----------|------|
| M1.1 | done   | 2026-08-09 | 2026-08-09 |      |
| M1.2 | done   | 2026-08-09 | 2026-08-09 |      |
| M1.3 | done   | 2026-08-09 | 2026-08-09 |      |
| M1.3a| done   | 2026-08-09 | 2026-08-09 | rebind d/shift+d→o/ctrl+o, D55 |
| M1.4 | done   | 2026-08-09 | 2026-08-09 |      |
| M1.4a| done   | 2026-08-09 | 2026-08-09 | rebind shift+enter+L/shift+p→ctrl+enter+ctrl+l/ctrl+p, D55 |
| M1.5 | done   | 2026-08-09 | 2026-08-09 |      |
| M1.6 | done   | 2026-08-09 | 2026-08-09 |      |
| M1.6a| done   | 2026-08-09 | 2026-08-09 | TRIAGE T5, D59 |
| M1.7 | done   | 2026-08-09 | 2026-08-09 | TRIAGE T6 |
| M2.1 | done   | 2026-08-09 | 2026-08-09 | TRIAGE T10, T12 |
| M3.1 | done   | 2026-08-09 | 2026-08-09 | PARTIAL — TRIAGE T11 confirmed (message-priority regression), T13 |
| M4.1 | done   | 2026-08-09 | 2026-08-09 |      |
| M5.1 | done   | 2026-08-09 | 2026-08-09 |      |
| M5.2 | done   | 2026-08-09 | 2026-08-09 |      |
