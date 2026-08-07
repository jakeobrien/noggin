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

Current task: M7.3 — fixes the Find-during-Peek gap that left M7.1/M7.2 unvalidated. Then
M8.2, M11.1, M12.1 in any order (no dependencies between them).

## Handoff

Replan (2026-08-09) triaged all 18 open TRIAGE.md items so the inbox is now empty:

- **T13/T14/T16 → M7.3** (new task, needs M7.1/M7.2): the validator's blocking finding —
  `FindNext`/`FindPrevious` ungated by `peeking()` — gets a real fix (guard + regression
  test), bundled with two cheap same-area nits (`startFind`'s direct focus assignment;
  a clarifying comment on the boundary-clamp clear condition). See DECISIONS.md D51.
- **T18 → M8.2** (new task, needs M8.1): `windowedViewRows` can render a split two-line
  entry when Roots Nav height < 2; fix floors `perScreen` to 0 in that case, plus a
  height=1/height=3 regression test.
- **T1/T6/T15 → M11.1** (new milestone): the README-sync task D48 already anticipated —
  drops `ctrl+p`, documents `alt+h` and Peek.
- **T2/T4/T5/T7/T17 → M12.1** (new milestone): five review-flagged test-coverage gaps,
  no functional change, bundled into one backfill task. T7 (medium severity) also got a
  DECISIONS.md entry (D50) since it's the one item at medium+.
- **T8**: fixed directly — one-line addendum to D46 cross-referencing D41 (why git's Stat
  calls are sync/gated while `rg`'s are async/debounced).
- **T10**: fixed directly — dropped the incorrect "D58" citation from the one
  `test/pty/checks.py` comment that misattributed the async-ordering-hazard fix to it; the
  *other* D58 citation in the same file was actually correct (a different fix) and left
  alone.
- **T12 → D50, T3/T9/T11 → D52**: wontfix/documented-as-is, each with its own recorded
  rationale (D50, D52).

No spec change; SPEC.md/PLAN.md stay mutually consistent with the four new tasks appended
(never renumbering M7.1/M7.2/M8.1's existing keys). Next step: work M7.3 first (it's what
turned M7.1/M7.2's Validated cell to `-`), then M8.2/M11.1/M12.1 in any order, then a final
`charm-validate` pass covering all four before the cycle can close.

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
| M7.3  | todo   | -          | -          | fixes TRIAGE T13/T14/T16; see DECISIONS.md D51 |
| M8.2  | todo   | -          | -          | fixes TRIAGE T18 |
| M11.1 | todo   | -          | -          | README sync, TRIAGE T1/T6/T15; see DECISIONS.md D48 |
| M12.1 | todo   | -          | -          | test-coverage backfill, TRIAGE T2/T4/T5/T7/T17; see DECISIONS.md D50 |
