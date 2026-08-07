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

Current task: M8.2. Then M11.1, M12.1 in any order (no dependencies between them), then a
final `charm-validate` pass covering M7.3/M8.2/M11.1/M12.1 before the cycle can close.

## Handoff

M7.3 is done (D51/TRIAGE T13/T14/T16, all three): `FindNext`/`FindPrevious` (`n`/`N`,
model.go's global switch) are now gated `!m.viewer.peeking()`, the same shape
`startEdit`/`startFind` already had — verified as a real fix, not a no-op change, by
reverting it locally and confirming the new regression test
(`TestFindNextPreviousAreNoOpsWhilePeeking`, cmd/noggin/peek_find_test.go) fails without
the guard (drifted offset/match-index, exactly D51's described scenario) and passes with
it. `startFind`'s own `m.focus = m.resolveFocus(viewerPane)` now routes through
`m.setFocus(...)` like every other Focus assignment (T14). A one-line comment documents
the boundary-clamped `LineUp`/`LineDown` clearPeek broadening as accepted, not a bug
(T16). `make check` and `make check-pty` both green after the change — no pty coverage
needed since this is pure model-level state, already exercised by the existing
`TestPeekCannotBeFound` and the new unit test together.

Next step: M8.2 (`windowedViewRows` split-row fix, TRIAGE T18) — read PLAN.md's M8.2 entry
and nav.go directly; no dependency on M7.3's area. M11.1 and M12.1 remain independent of
both. TRIAGE.md is empty; nothing else outstanding from the 2026-08-09 replan.

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
| M8.2  | todo   | -          | -          | fixes TRIAGE T18 |
| M11.1 | todo   | -          | -          | README sync, TRIAGE T1/T6/T15; see DECISIONS.md D48 |
| M12.1 | todo   | -          | -          | test-coverage backfill, TRIAGE T2/T4/T5/T7/T17; see DECISIONS.md D50 |
