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

Current task: none — all C3 plan tasks are `done` and independently validated. TRIAGE.md is
**not** empty (T14, T15 open, from the M2.1b review) — a cycle cannot close with an open
inbox. Next step is `/charm-replan` to triage T14/T15, not `/charm-close` yet.

## Handoff

Final cycle-end validate pass (fresh-context subagent, deterministic + independent
re-verification, not just re-reading tests): `make check` and `make check-pty` both pass
clean (fresh `go clean -testcache`; check-pty 148/148). M2.1a, M2.1b and M3.1a — the three
tasks Reviewed but not yet Validated — each independently confirmed PASS against their
PLAN.md done-when clauses, including break/restore regression checks on M2.1b's
`staleBaseline.withoutDoc` call and M3.1a's `combineDropMessages` guard (both neutralized,
confirmed the relevant tests fail, then restored exactly — working tree left clean). All six
SPEC.md cycle-level acceptance bullets also spot-checked PASS against existing tests/code. No
discrepancies found between PROGRESS.md's prior claims and reality. Validated column stamped
below for M2.1a/M2.1b/M3.1a.

Discrepancy found: the previous session's handoff claimed "TRIAGE.md stays empty" but it
actually carries two open review findings (T14, T15) from that same M2.1b work — added but
never mentioned in the prose handoff, only in the Status table's M2.1b note. Per the doc-map,
a cycle cannot close with a non-empty TRIAGE.md.

Next: run `/charm-replan` to triage T14/T15 (doc-comment contract for `closeDV2`'s baseline
mutation; redundant guard nit) into DECISIONS.md/PLAN.md/dropped as appropriate — only then
is `/charm-close` warranted.

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
| M2.1a| done   | 2026-08-09 | 2026-08-09 | TRIAGE T10, D62 |
| M2.1b| done   | 2026-08-09 | 2026-08-09 | TRIAGE T12; T14/T15 open, blocks close |
| M3.1a| done   | 2026-08-09 | 2026-08-09 | TRIAGE T11, T13, D61 |
