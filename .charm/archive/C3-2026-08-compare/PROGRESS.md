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

Current task: **M2.1c** built and validated — the last open plan task, now closed out.
TRIAGE.md is empty; the cycle is ready for `/charm-close`.

## Handoff

Validated M2.1c this session (fresh-context check, not the building session): confirmed both
`staleBaseline` mutation-contract doc comments (staleness.go:21-28, model.go:126-128) name
`closeDV2`'s `withoutDoc` call (C3/M2.1b) as the sanctioned exception, and that this matches
the actual code at model.go:322 (`withoutDoc` gated by `isOpen()`). Cross-checked
`resetStaleBaseline`/`runStaleCheck`/`applyStaleResult`/`withoutDoc`'s own doc comments for
consistency — no contradictions found anywhere. Re-ran `make check` with a cleared test cache
(not the cached result) — build/vet/test all clean. No new findings; TRIAGE.md stays empty.

Next: no plan tasks remain. Proceed to `/charm-close`.

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
| M2.1c| done   | -          | 2026-08-09 | TRIAGE T14, D63       |
