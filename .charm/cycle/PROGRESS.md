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

Current task: M3.1a (fixes TRIAGE T11's cycle-close-blocking regression; highest priority of
the three replan-added follow-ups)

## Handoff

charm-replan ran this session, consuming all four TRIAGE.md items the prior charm-validate
pass left open (inbox is now empty). Routed: T11 (high, refresh() message-priority
regression) + T13 (low, duplicated combining logic) folded into one new task **M3.1a**
(D61) since both touch the same switch; T10 (med, no pty coverage for ctrl+x) → new task
**M2.1a** (D62); T12 (low, closeDV2 doesn't reset the stale baseline) → new task **M2.1b**
(no log needed, low/non-material). All three are `todo`, not yet built. M3.1's Status stays
`done` (content/scroll half is genuinely correct) but its Note now points at M3.1a instead
of the resolved-and-removed triage ids.

No spec change involved — SPEC.md, PLAN.md's other tasks and PROGRESS.md's other statuses
are untouched. DECISIONS.md open items (D9/D10/D33) predate C3 and are out of scope.

Next: build M3.1a first (it's what's blocking full cycle-level acceptance), then M2.1a/M2.1b
(low effort, non-blocking, any order). Once all three are done and validated, TRIAGE.md is
already empty, so C3 should be ready for `/charm-close` pending a final validate pass over
just the three new tasks.

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
| M2.1 | done   | 2026-08-09 | 2026-08-09 | see M2.1a (T10), M2.1b (T12) |
| M3.1 | done   | 2026-08-09 | 2026-08-09 | PARTIAL — see M3.1a (fixes T11/T13), D61 |
| M4.1 | done   | 2026-08-09 | 2026-08-09 |      |
| M5.1 | done   | 2026-08-09 | 2026-08-09 |      |
| M5.2 | done   | 2026-08-09 | 2026-08-09 |      |
| M2.1a| todo   | -          | -          | TRIAGE T10, D62 |
| M2.1b| todo   | -          | -          | TRIAGE T12 |
| M3.1a| todo   | -          | -          | TRIAGE T11, T13, D61 |
