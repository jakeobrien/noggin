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

Current task: M2.1b (TRIAGE T12 — closeDV2 doesn't reset the stale baseline)

## Handoff

Built M2.1a this session: added `test/pty/checks.py` KEY table entry `CloseDV2` (`\x18`,
ctrl+x, plain C0 byte) and a new check block reusing M1.7's c3_root alpha/bravo fixtures —
opens DV2 via ctrl+enter (focus lands on DV2), presses ctrl+x, and confirms DV2's content is
gone, DV1's own document is untouched, and Focus actually lands on DV1 (G scrolls DV1, not a
stale DV2). Regression-proofed by temporarily wiring the KEY entry to the wrong byte (ctrl+w
instead of ctrl+x) and confirming the two content/focus checks fail as expected, then
restoring the correct byte and re-running clean.

`make check` and `make check-pty` both pass in full (all pre-existing C3 pty coverage plus
the new CloseDV2 block). No spec change; SPEC.md and PLAN.md untouched. TRIAGE.md stays
empty.

Next: build M2.1b (closeDV2 should prune DV2's entry from the staleness baseline
immediately, not just on the next refresh — PLAN.md's own acceptance clause). Once done and
validated, TRIAGE.md is already empty, so C3 should be ready for `/charm-close` pending a
final validate pass over all three new tasks (M3.1a/M2.1a/M2.1b) — M3.1a especially, since
it was validate that surfaced T11 in the first place.

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
| M2.1a| done   | -          | -          | TRIAGE T10, D62 |
| M2.1b| todo   | -          | -          | TRIAGE T12 |
| M3.1a| done   | -          | -          | TRIAGE T11, T13, D61 |
