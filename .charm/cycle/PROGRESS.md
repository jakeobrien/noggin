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

Current task: the closing `charm-validate` pass, covering M7.3/M8.2/M11.1/M12.1. That is
the only work left before the cycle can close.

## Handoff

M12.1 is done (TRIAGE T2/T4/T5/T7/T17, all five; DECISIONS.md D50 named T7 specifically).
No production code changed — test-only backfill, as the task's own done-when required:

- T4/T5 (`cmd/noggin/refresh_test.go`): `TestRefreshWorksWithFocusOnRootsPane` drives
  `keys.Refresh` with `m.focus = rootsPane` (the one pane the existing suite skipped, unlike
  `focus_test.go`'s every-pane table for FocusRoots/FocusFiles); `TestRefreshDoesNotTouchOtherRoots`
  configures two roots, deletes the *inactive* one's directory and adds a file under the
  Active Root, then refreshes — pins that the Active Root's tree picks up the new file while
  the inactive root's row (including its now-stale `broken` snapshot) is untouched, the
  negative clause T5 flagged as true-by-construction but unguarded.
- T7 (`cmd/noggin/picker_view_test.go`): added a `hideDotfiles: true` case to
  `TestPickerStatusStates` asserting `"File · Active Root · Hidden"`, closing the gap D50
  named (only `test/pty/checks.py` exercised the ` · Hidden` suffix before this).
- T2 (`test/pty/checks.py`): added a `FocusRoots` (`r`) KEY entry (previously absent
  entirely) and a new C2a/M2.1 pty section asserting `r`/`f` actually redraw the
  focused/unfocused border colour (`38;5;205`/`38;5;240`) in the byte delta produced by each
  keypress (the `mark = len(s.out)` idiom the Peek reverse-video check already uses) — not
  just `m.focus`'s internal state, which go test already covered.
- T17 (`test/pty/checks.py`): widened the M5.3 viewer-context session from 80x24 to 80x36
  (helpBindings(viewerPane)'s list renders a measured 31-row box, taller than both the
  24-row session and the Files-pane block's own 25-row box/30-row session just above it) and
  added the same top-border/footer-unchanged assertions that Files-pane block already had,
  so a future binding pushing an assertion past `overlayHelp`'s clamp would now fail loudly.

Every new/changed test passes; `make check` and `make check-pty` are both green (full
output checked, no other check regressed).

Next step: all four remaining cursor tasks (M7.3, M8.2, M11.1, M12.1) are done. The only
work left is the closing `charm-validate` pass this cycle's cursor has named throughout —
covering all four together, from a fresh context, per the skill's independence requirement.
TRIAGE.md carries one open, non-blocking item (T19, logged during M11.1: README's Global
table also omits `r`/`f`/`R`) for replan to pick up; it does not block validation or close.

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
| M11.1 | done   | -          | -          | README sync, TRIAGE T1/T6/T15; see DECISIONS.md D48 |
| M12.1 | done   | -          | -          | test-coverage backfill, TRIAGE T2/T4/T5/T7/T17; see DECISIONS.md D50 |
