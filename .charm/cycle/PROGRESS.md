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

Current task: M12.1. Then a final `charm-validate` pass covering M7.3/M8.2/M11.1/M12.1
before the cycle can close.

## Handoff

M11.1 is done (TRIAGE T1/T6/T15, all three; DECISIONS.md D48 scoped this as the dedicated
README-sync task): every remaining `ctrl+p` reference in README.md is gone (Global
keybinding table row, the "While the Picker is open" heading, and the "Finding across
files" opening prose all now say `space`, matching `keys.Picker`). `alt+h` (T6) is now
documented in both the Picker's own keybinding table and a new sentence in the "Requires
ripgrep" prose describing the ` · Hidden` title suffix and its session-only scope — wording
cross-checked against `pickerModeScope` (picker_view.go) verbatim. Peek (T15) got a new
row in the Files Nav keybinding table (`p`, what it does, how it ends). Verified with
`grep -n ctrl+p README.md` (no hits) and a manual re-read of every touched section against
the keymap/S4/S7 spec text. No code changed, so only `make check` was run (green); no
`check-pty` needed.

While fixing this, noticed the Global table also never mentions `r`/`f`/`R`
(FocusRoots/FocusFiles/Refresh) — out of M11.1's named scope (its acceptance criterion
named only `ctrl+p`/`alt+h`/`p`), so logged as TRIAGE T19 rather than folded in here.

Next step: M12.1 (test-coverage backfill, TRIAGE T2/T4/T5/T7/T17, see DECISIONS.md D50) is
the last cursor task before the closing `charm-validate` pass. TRIAGE.md carries one open
item (T19, low severity, doc-only) from this session — not blocking, but replan should
triage it before close.

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
| M12.1 | todo   | -          | -          | test-coverage backfill, TRIAGE T2/T4/T5/T7/T17; see DECISIONS.md D50 |
