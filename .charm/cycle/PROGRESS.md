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

Current task: M3.1

## Handoff

M2.1 is done (this session): added `keys.CloseDV2` (`ctrl+x`) and `model.closeDV2` —
global, matched in Update's already-gated switch (unreachable while a find is editing, the
Help Modal is open, or the Picker is open, the same three gates every other global binding
shares), a no-op when `m.dv2 == nil`, otherwise clears `m.dv2`, moves Focus straight to
`dv1Pane` (bypassing `resolveFocus`'s generic focusOrder fallback — that would land on
whichever Navigator pane happens to be visible instead of S4's unconditional "Focus moves
to DV1") only when Focus was on `dv2Pane`, then re-lays-out through `resize()` with a new
`resizeCauseViewerClosed` cause, the mirror of `ensureDV2`'s `resizeCauseSecondViewer` —
giving DV1 back the full viewer width and reusing resize's existing find-drop messaging for
free. Followed FocusDV2's established precedent for the two hint surfaces: listed
unconditionally in the Help Modal's `helpBindings` (discoverability backstop) but omitted
from the footer's terse `paneHints` (usually a no-op with DV2 absent) — both sides pinned
with dedicated tests. New `close_dv2_test.go` covers the full pane sweep (Focus-landing
rule), the DV2-absent no-op, and inertness under both modals; verified the sweep test
actually bites by regression injection (temporarily dropping the `keys.CloseDV2` case from
Update failed exactly that test, reverted clean). `make check` and `make check-pty` both
pass (no pty entries added for `ctrl+x` — logged as TRIAGE T10, since M1.7's own pty task
predates this binding).

Next: **M3.1** — extend `refresh()` to re-read both DV1's and DV2's Active Documents
(whichever exist), not just one, in the same single action.

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
| M3.1 | todo   | -        | -         |      |
| M4.1 | todo   | -        | -         |      |
| M5.1 | todo   | -        | -         |      |
| M5.2 | todo   | -        | -         |      |
