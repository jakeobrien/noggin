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

Current task: (none — all M1–M5 tasks done; cycle close/validation due)

## Handoff

M5.2 is done (this session), and with it every task in PLAN.md (M1–M5): extended the
staleness baseline (S10/C2a) to track both DV1's and DV2's Active Documents, whichever
exist, per S7. `staleness.go`: `staleBaseline.docPath/docModTime/docSize` (one document)
became `docs []docSnapshot` (a new small struct: path/modTime/size); `snapshotStaleness`
took a `docPath string` and is now variadic (`docPaths ...string`, backward-compatible with
every existing single-path call site); `stale()` loops the slice instead of one inline
check. `resetStaleBaseline` now gathers DV1's path unconditionally and DV2's only when
`m.dv2 != nil` (an absent DV2 contributes no path at all, not a placeholder empty one,
matching `snapshotStaleness`'s own "whichever exist" contract) — the M1.1-era comment
flagging this as DV1-only pending M5.2 is gone. Added `trackedDocPaths()` (test-facing read
of the tracked paths) since the fields are no longer directly comparable one-for-one.
No change was needed anywhere else: `openFolder`/`openDocumentInto`/`refresh` already called
`resetStaleBaseline` on every fresh read for both viewers (M1.1/M3.1's own plumbing), and
M5.1's row/layout already render off `m.stale` alone, so the new detection reaches S9's
row for free. Added the plan's own done-when verbatim
(`TestStaleCheckDetectsOnlyDV2DocumentEdited`/`...DV1DocumentEdited` — edit one viewer's
document only, tree and the other viewer untouched, confirm both `m.stale` and the rendered
row) plus two wiring tests (`TestResetStaleBaselineTracksBothViewersActiveDocuments`,
`TestResetStaleBaselineDropsDV2WhenItCloses`) and fixed two pre-existing tests that read the
now-gone `docPath` field directly. Verified the four new tests actually bite by regression
injection (temporarily reverted `resetStaleBaseline` to DV1-only — three of the four failed
with the expected messages, the fourth (DV1-edited) correctly still passed since it never
exercised the DV2 path; reverted clean). `make check` and `make check-pty` both pass.

Next: no task remains in PLAN.md — M1–M5 are all done. This milestone boundary (in fact the
whole cycle) has validation due (see cycle-state) before C3 can close; none of C3's work has
been through charm-validate yet (the Reviewed/Validated columns below are `-` for M1.6a
onward). Recommend a full charm-validate pass against SPEC.md's cycle-level acceptance next,
not further building.

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
| M4.1 | done   | -        | -         |      |
| M5.1 | done   | -        | -         |      |
| M5.2 | done   | -        | -         |      |
