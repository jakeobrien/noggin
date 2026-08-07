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

Current task: none — M10.2 implemented this session. The cycle is not yet re-offered as
ready to close: M10.1+M10.2 need a fresh-context validate pass together at a realistic width
before that happens again (M10.1's Validated stamp is still `-` from the prior session).

## Handoff

build (2026-08-10): implemented M10.2. Root cause (D53) was `hintText`/`pickerHintText`
(footer.go) filling the footer width greedily before `appendStaleSuffix` (model.go) ever got
a turn, so the suffix had no slack left at any ordinary width. Fix: `renderFooter` (model.go)
now computes `hintWidth = r.width - len(" · "+staleMessage)` whenever the suffix will actually
be appended (`m.stale && m.message == ""`) and builds the hint set against that reduced width
instead of the full one — `hintText`/`pickerHintText`'s own existing truncation (drop
least-essential hints first, per `paneHints`' order) then makes room for the suffix ahead of
time, and `appendStaleSuffix`'s own fits-or-drops check against the real width is unchanged.
Verified manually at 80/100/120/132/200 cols (all now show the suffix; previously none did
below ~200). Added `TestFooterShowsStaleIndicatorAtARealisticWidth` (staleness_test.go,
80/100/120/132) alongside the existing 200/400-col cases, and a new live
`test/pty/checks.py` block (C2a/M10.2, end of file): real out-of-process append to a tracked
file, a real `staleCheckInterval`-length wait (7s), at 80 cols, asserting the footer shows the
indicator — plus a before-check that it is absent pre-edit. `make check` and `make check-pty`
both green (full check-pty run, all checks including the three new M10.2 ones passed).

Everything from the prior handoff (fresh-context validate of M7.3/M8.2/M11.1/M11.2/M12.1/
M7.1/M7.2; T20's peek_find_test.go fix; the charm-document README-sync pass for S3/S5/S8/S10;
D53's investigation of the original report) stands unchanged — see git history for that work;
not repeated here per the handoff's overwrite rule.

Next step: fresh-context validate of M10.1+M10.2 together at a realistic width (the gap the
2026-08-08 validate pass missed, per D53) before offering the cycle as ready to close again.

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
| M10.2 | done   | -          | -          | fixes footer-width bug found post-close-readiness; see DECISIONS.md D53 |
| M7.3  | done   | 2026-08-07 | 2026-08-07 | fixes TRIAGE T13/T14/T16; see DECISIONS.md D51 |
| M8.2  | done   | 2026-08-07 | 2026-08-07 | fixes TRIAGE T18 |
| M11.1 | done   | 2026-08-07 | 2026-08-07 | README sync, TRIAGE T1/T6/T15; see DECISIONS.md D48 |
| M11.2 | done   | 2026-08-07 | 2026-08-07 | README Global table gap, TRIAGE T19 |
| M12.1 | done   | 2026-08-07 | 2026-08-07 | test-coverage backfill, TRIAGE T2/T4/T5/T7/T17; see DECISIONS.md D50 |
