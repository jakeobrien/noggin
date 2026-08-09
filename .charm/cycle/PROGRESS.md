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

Current task: none — all C3 plan tasks are `done`; TRIAGE.md is empty. Ready for a
final validate pass, then `/charm-close`.

## Handoff

Built M2.1b this session (TRIAGE T12): added `staleBaseline.withoutDoc(path)`
(staleness.go) and call it from `closeDV2` (model.go) right before nilling `m.dv2`, pruning
DV2's former document from the tracked-docs slice the instant it stops being visible —
rather than leaving it tracked until whatever refresh happens to run next. Updated
`TestResetStaleBaselineDropsDV2WhenItCloses` (its old form documented the pre-fix
behavior — asserted the drop only after an explicit `m.refresh()`; now asserts it
immediately, no refresh at all) and added `TestClosingDV2DoesNotFlipStaleForItsFormerDocument`
(M2.1b's own acceptance clause verbatim: edit the former document on disk right after
closing DV2, no intervening Refresh, confirm `m.stale` stays false). Both regression-proofed
by reverting the `withoutDoc` call and confirming they fail, then restoring and re-running
clean.

`make check` and `make check-pty` both pass in full (fresh `go clean -testcache` run, not
just cached). No spec change; SPEC.md and PLAN.md untouched. TRIAGE.md stays empty — all
three replan-added follow-up tasks (M3.1a, M2.1a, M2.1b) are now done.

Next: a final validate pass over M3.1a/M2.1a/M2.1b (M3.1a especially — it was validate that
surfaced T11 in the first place, so a fresh check that the fix actually holds is warranted),
then `/charm-close`.

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
| M2.1b| done   | -          | -          | TRIAGE T12 |
| M3.1a| done   | -          | -          | TRIAGE T11, T13, D61 |
