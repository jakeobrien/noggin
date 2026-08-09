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

Current task: M2.1a (TRIAGE T10, D62 — pty coverage for ctrl+x; M2.1b remains after it, any
order)

## Handoff

Built M3.1a this session: extracted the identical refresh()/resize() message-combining
switch into one shared helper, `combineDropMessages` (model.go), guarded with an
`m.message != ""` check so an existing Root-walk or per-viewer read-failure message can
never be overwritten by a find-drop notice — fixes TRIAGE T11 (refresh's switch previously
had no such guard, unlike refreshViewer's own read-failure branch) and T13 (the duplicated
switch) in the same change, per D61. Two regression tests added
(refresh_test.go: TestRefreshReadFailureMessageWinsOverFindDroppedMessage, single-viewer;
…WithDV2, dual-viewer racing a genuine DV2 find-drop against DV1's read failure) — both
confirmed to fail against the pre-fix (unguarded) switch and pass with the fix. `resize()`
now calls the same helper (behavior unchanged there — resize never set m.message before its
own switch ran, confirmed by its existing test suite still passing unchanged).

`make check` and `make check-pty` both pass in full, including the new tests and all
pre-existing C3 pty coverage (M1.7's four chords). No spec change; SPEC.md and PLAN.md are
untouched. TRIAGE.md stays empty (no new items surfaced).

Next: build M2.1a (pty coverage for ctrl+x), then M2.1b (closeDV2 staleness-baseline prune),
any order — both low effort, non-blocking. Once both are done and validated, TRIAGE.md is
already empty, so C3 should be ready for `/charm-close` pending a final validate pass over
just these two remaining new tasks (M3.1a's own fix should get a validate pass too, since
it was validate that surfaced T11 in the first place — a fresh check that the fix actually
holds is warranted before close).

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
| M2.1a| todo   | -          | -          | TRIAGE T10, D62 |
| M2.1b| todo   | -          | -          | TRIAGE T12 |
| M3.1a| done   | -          | -          | TRIAGE T11, T13, D61 |
