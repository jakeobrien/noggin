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

Current task: M11.1 and M12.1, in any order (no dependencies between them), then a final
`charm-validate` pass covering M7.3/M8.2/M11.1/M12.1 before the cycle can close.

## Handoff

M8.2 is done (TRIAGE T18): `windowedViewRows` (nav.go) computed `perScreen :=
max(height/rowHeight, 1)`, which forced at least one whole entry to render even when the
pane's height was smaller than `rowHeight` — the entry's lines were then sliced down to
`height`, producing a truncated first line rather than the intended zero entries. Dropped
the forced minimum (`perScreen := height/rowHeight`); when `height < rowHeight`,
`perScreen` is 0, the render loop is a no-op, and the existing pad loop fills the box with
blank rows instead. Verified as a real fix by reverting it locally and confirming the new
regression test (`TestRootsNavPaneShorterThanRowHeightRendersZeroEntries`,
cmd/noggin/rootsnav_test.go) fails without the guard (renders alpha's reverse-video name
line truncated to one row) and passes with it. A companion test
(`TestRootsNavOddHeightPadsRatherThanSplitting`, height=3 with rowHeight=2) was added for
the acceptance criterion's other edge case but passed even before the fix — that path
(`height/rowHeight` already >= 1) was never broken, only the below-rowHeight case was.
`make check` and `make check-pty` both green — no new pty case added since this is pure
model-level windowing logic already exercised by unit tests, and the existing M8.1 pty
checks (two-line row rendering) exercise the same view path at a normal height without
regressing.

Next step: M11.1 (README sync, TRIAGE T1/T6/T15, see DECISIONS.md D48) and M12.1
(test-coverage backfill, TRIAGE T2/T4/T5/T7/T17, see DECISIONS.md D50) remain, in either
order — no dependencies between them or on M8.2's area. TRIAGE.md is empty; nothing else
outstanding from the 2026-08-09 replan.

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
| M11.1 | todo   | -          | -          | README sync, TRIAGE T1/T6/T15; see DECISIONS.md D48 |
| M12.1 | todo   | -          | -          | test-coverage backfill, TRIAGE T2/T4/T5/T7/T17; see DECISIONS.md D50 |
