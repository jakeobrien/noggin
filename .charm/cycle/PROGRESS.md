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

Current task: M4.1

## Handoff

M3.1 is done (this session): `model.refresh()` now re-reads both DV1's and DV2's Active
Documents (whichever exist) in the same single action, not just DV1 — factored the
per-viewer read/refreshDoc/close/find-drop logic out of the old DV1-only inline block into a
new `refreshViewer(p pane) (model, string)` helper (reused for both panes via `viewerFor`/
`withViewerFor`, the same accessor pair M1.1 established), called once for `dv1Pane`
unconditionally and once more for `dv2Pane` when `m.dv2 != nil`. The two read-failure/
find-drop outcomes combine into `m.message` the same way `resize()`'s own `dv1Drop`/
`dv2Drop` pair already do (mirrored the wording/priority verbatim: DV1 unprefixed when DV2
drops nothing, `"dv1 …, dv2 …"` when both drop, `"dv2 …"` when only DV2 drops) — so S6's
independence guarantee holds at refresh's messaging layer too, not just its routing. Root-
walk failure still wins over both viewers' read failures; between the two viewers, DV1 wins
(new call-order rule, undocumented before since DV2 didn't exist) — both pinned by new
tests. Added three tests to `refresh_test.go`: the plan's own acceptance verbatim (both
viewers open on distinct documents, both changed on disk, one Refresh grows both and
preserves each viewer's own scroll offset), the two-viewer find-drop combination, and the
DV1-wins-DV2 read-failure priority. Verified all three actually bite by regression injection
(temporarily short-circuiting the DV2 branch in `refresh()` failed exactly those three,
reverted clean). `make check` and `make check-pty` both pass (no new keys, so no pty
additions needed).

Next: **M4.1** — confirm DV2's existence, Active Document and Peek/Find state are never
persisted or restored at launch (effort 1, needs M1.1).

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
| M1.6a| done   | -        | -         | TRIAGE T5, D59 |
| M1.7 | done   | -        | -         | TRIAGE T6 |
| M2.1 | done   | -        | -         |      |
| M3.1 | done   | -        | -         |      |
| M4.1 | todo   | -        | -         |      |
| M5.1 | todo   | -        | -         |      |
| M5.2 | todo   | -        | -         |      |
