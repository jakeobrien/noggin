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

Current task: **M2.1c** (doc-comment-only, effort 1) — the last open plan task. TRIAGE.md is
now empty; once M2.1c is built and validated, the cycle is ready for `/charm-close`.

## Handoff

Replanned this session: triaged TRIAGE.md's two open items (T14, T15), both surfaced by the
prior cycle-end validate pass against M2.1b. T14 (med, `closeDV2`'s direct `staleBaseline`
mutation contradicts two standing "never mutated except by resetStaleBaseline/..." doc
comments — staleness.go:21-23, model.go:126) accepted as new task **M2.1c**: reword both
comments to name `closeDV2`'s `withoutDoc` call as the sanctioned exception. Logged as D63
(medium severity, accept-path). T15 (low, redundant `isOpen()` guard before `withoutDoc` in
`closeDV2`, model.go:317-319) decided wontfix — keep the guard, purely stylistic, mirrors
D57's precedent — logged as D64. TRIAGE.md inbox is empty again.

No spec change; SPEC.md untouched. PLAN.md gained M2.1c (suffixed under M2.1, needs: M2.1b,
effort 1). No other plan divergence found — all other `done` tasks' validation from the
prior session stands (M2.1a/M2.1b/M3.1a all independently confirmed PASS with break/restore
regression checks; `make check`/`make check-pty` both clean).

Next: build M2.1c (small doc-comment edit in staleness.go and model.go, no code-behavior
change, no new test needed since nothing testable changed — just `make check` to confirm
fmt/vet/build stay clean), then a final validate touch on M2.1c specifically, then
`/charm-close`.

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
| M2.1b| done   | 2026-08-09 | 2026-08-09 | TRIAGE T12; T14→M2.1c/D63, T15→D64 |
| M3.1a| done   | 2026-08-09 | 2026-08-09 | TRIAGE T11, T13, D61 |
| M2.1c| todo   | -          | -          | TRIAGE T14, D63       |
