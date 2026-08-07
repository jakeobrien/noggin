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

Current task: none — all plan tasks done, all validated, TRIAGE.md is now empty. The cycle is
ready to close (`/charm-close`).

## Handoff

Fresh-context validate pass complete, deterministic + functional, for M7.3/M8.2/M11.1/M11.2/
M12.1 (all now stamped Validated 2026-08-07 below), plus M7.1/M7.2 (also stamped: M7.3's fix
and its regression test directly prove the exact clauses that left them unvalidated — see
their Note). `make check` and `make check-pty` both green, `-count=1` (not cache-trusted).
Verified independently: FindNext/FindPrevious no-op + exact-offset restore while Peeking
(peek_find_test.go), `startFind` routed through `setFocus`, the LineUp/LineDown broadening
comment (model.go:666); `windowedViewRows` renders zero entries at height=1 and pads (not
splits) at height=3 (rootsnav_test.go); README.md has zero `ctrl+p` references anywhere and
documents `alt+h`/Peek; the Global keybinding table lists `r`/`f`/`R`; all five M12.1 backfill
tests/checks exist, are correctly targeted, and pass. Also manually verified `make build &&
make install` and a custom `PREFIX` both place a working binary. No discrepancies found; no
new TRIAGE items raised by this pass.

replan (2026-08-09) closed out the one remaining loose end: TRIAGE.md's T20 (low severity,
a test-style nit — `peek_find_test.go` pressed raw `"n"`/`"N"` instead of
`keys.FindNext.Keys()[0]`/`keys.FindPrevious.Keys()[0]`) was low risk × value, mechanical, and
no-behavior-change, so it was fixed directly rather than routed to a plan task (mirroring
T8/T10's precedent) — see TRIAGE.md's inbox comment. `go test` (targeted) and `make check`
reverified green after the edit.

Next step: none required. `/charm-close` whenever the user is ready.

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
| M10.1 | done   | 2026-08-08 | 2026-08-08 | perf clause ("no perceptible cost on a large Root") verified architecturally, not benchmarked |
| M7.3  | done   | 2026-08-07 | 2026-08-07 | fixes TRIAGE T13/T14/T16; see DECISIONS.md D51 |
| M8.2  | done   | 2026-08-07 | 2026-08-07 | fixes TRIAGE T18 |
| M11.1 | done   | 2026-08-07 | 2026-08-07 | README sync, TRIAGE T1/T6/T15; see DECISIONS.md D48 |
| M11.2 | done   | 2026-08-07 | 2026-08-07 | README Global table gap, TRIAGE T19 |
| M12.1 | done   | 2026-08-07 | 2026-08-07 | test-coverage backfill, TRIAGE T2/T4/T5/T7/T17; see DECISIONS.md D50 |
