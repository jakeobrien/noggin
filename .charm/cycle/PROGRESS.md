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

Current task: **M10.2** — fix the S10 staleness indicator being unrenderable at realistic
terminal widths (see DECISIONS.md D53). The cycle is no longer ready to close.

## Handoff

replan (2026-08-10): user reported the staleness indicator (S10/M10.1) "does not seem to be
working," discovered *after* the previous session had marked the cycle ready to close.
Investigated via a researcher (static trace: scheduling/baseline/wiring in staleness.go all
sound) then a validator (live repro via the pty harness, real wall-clock ticks, multiple
terminal widths) — confirmed a real bug, not a false alarm: `appendStaleSuffix` (model.go)
requires its whole ~48-char suffix to fit in whatever's left after `hintText` greedily fills
the footer, so it never renders below ~200-244 columns in any pane, i.e. never at any width
this project's own test/pty fixtures use (80/100/120/132/200). The detection engine itself is
confirmed sound (fires correctly once width is widened past the threshold), and the separately
documented in-place-edit-to-a-non-Active-file gap reproduced exactly as designed — this is a
footer-rendering bug layered on top, not a flaw in D49's heuristic. Added PLAN.md M10.2
(reserve suffix room ahead of hintText's greedy fill; add a realistic-width `test/pty/checks.py`
check) and DECISIONS.md D53. Cleared M10.1's Validated stamp below to `-`: the 2026-08-08
validate pass only exercised staleness_test.go's existing 200/400-col render cases, never a
live run at an ordinary width, so it did not actually cover the "surfaces within one check
interval" clause end-to-end. TRIAGE.md stays empty (routed straight through rather than
parked as a row then immediately cleared — see its inbox comment).

Everything from the prior handoff (fresh-context validate of M7.3/M8.2/M11.1/M11.2/M12.1/
M7.1/M7.2; T20's peek_find_test.go fix; the charm-document README-sync pass for S3/S5/S8/S10)
stands unchanged — see git history (`3796e5c`, `2b0a1b3`, `e633738`) for that work; not
repeated here per the handoff's overwrite rule.

Next step: implement M10.2 (charm-build), then re-validate M10.1+M10.2 together at a
realistic width before the cycle is offered as ready to close again.

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
| M7.1  | done   | 2026-08-07 | 2026-08-07 | T13's gap (n/N ungated by peeking()) fixed by M7.3 and re-verified directly |
| M7.2  | done   | 2026-08-07 | 2026-08-07 | T13's scroll-restore-on-clear gap fixed by M7.3 and re-verified directly |
| M8.1  | done   | 2026-08-07 | 2026-08-08 |      |
| M9.1  | done   | 2026-08-07 | 2026-08-08 |      |
| M10.1 | done   | 2026-08-08 | -          | perf clause ("no perceptible cost on a large Root") verified architecturally, not benchmarked; indicator unrenderable below ~200-244 cols, see M10.2/D53 |
| M10.2 | todo   | -          | -          | fixes footer-width bug found post-close-readiness; see DECISIONS.md D53 |
| M7.3  | done   | 2026-08-07 | 2026-08-07 | fixes TRIAGE T13/T14/T16; see DECISIONS.md D51 |
| M8.2  | done   | 2026-08-07 | 2026-08-07 | fixes TRIAGE T18 |
| M11.1 | done   | 2026-08-07 | 2026-08-07 | README sync, TRIAGE T1/T6/T15; see DECISIONS.md D48 |
| M11.2 | done   | 2026-08-07 | 2026-08-07 | README Global table gap, TRIAGE T19 |
| M12.1 | done   | 2026-08-07 | 2026-08-07 | test-coverage backfill, TRIAGE T2/T4/T5/T7/T17; see DECISIONS.md D50 |
