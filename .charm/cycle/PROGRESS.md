# Progress — C4 recall

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

Current task: M2.3

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M2.2 done — star/un-star from a focused Document Viewer, acting on that viewer's own Active
Document. `model.toggleActiveDocStar()` (model.go) reads `m.focusedViewer()` — already DV1 vs
DV2 by Focus, so no extra targeting was needed for independence — and flips
`docRecord(v.doc.path).Starred` through `withDocRecord`, same as M2.1's `toggleDocStar`. Two
refusals mirror `startEdit`/`startFind`'s existing shape exactly: silent no-op while Peeking
(S7/M7.1 — the Peeked document is never the Active Document, so acting on "the Active Document"
while something else is on screen would be the same disturb-what's-not-displayed surprise those
two already guard against), and a footer message with no document open. Wired on the same
shared `keys.Star` chord, matched in `Update`'s `case dv1Pane, dv2Pane:` after the scroll check
fails (scroll owns no `s` binding, so no collision). `TestEveryBindingIsCovered` again needed no
update — still the same existing binding, now matched in a third pane case.
Evidence (new `viewer_star_test.go`): `TestToggleActiveDocStarIsIndependentPerViewer` (M2.2's
acceptance clause verbatim — stars from DV2, then DV1, then un-stars from DV2, checking the
other viewer's record is untouched at every step, plus a restart round-trip),
`TestToggleActiveDocStarNoOpWithNoDocumentOpen`, `TestToggleActiveDocStarNoOpWhilePeeking`.
`make check` green.
Next: M2.3, a star marker on Files Nav rows in every browse mode except `Starred` (needs M2.1,
done; the marker-in-Starred clause itself needs M4.3, not yet built — M2.3 only has to cover
Tree for now, per its own plan wording "absent from every row of a Starred list" being trivially
true while `Starred` mode doesn't exist yet). Likely touches `filesNav.renderRow` (files.go) to
consult `model.docRecord` — note renderRow currently takes no model/docs argument, so it may
need one threaded through, or the marker decided before rendering and passed in per row.

Gotchas surfaced while planning:
- `store.go` today is Roots-specific (`loadStore`/`saveStore`/`toggleStore`, plain one-path-
  per-line format). M1.2's document store needs a two-field record — give it its own file and
  names rather than overloading these.
- `frame_test.go`'s `TestEveryBindingIsCovered` will fail on any new binding (M4.1's two cycle
  keys, M6.1's two history keys) until the coverage list is updated.
- M5.1 changes the meaning of Picker scope: `liveRootsInScope` (`scope.go:26`) is Root-set-
  based, but `Starred` is a file set. Expect to refactor rather than add a branch.
- Gate is `make check`; `make check-pty` is the separate pty layer.

## Status

| Key  | Status | Reviewed | Validated | Note                                     |
|------|--------|----------|-----------|------------------------------------------|
| M1.1 | done   | 2026-08-10 | 2026-08-10 |                                       |
| M1.2 | done   | 2026-08-10 | 2026-08-10 |                                       |
| M1.3 | done   | 2026-08-10 | 2026-08-10 |                                       |
| M1.4 | done   | 2026-08-10 | 2026-08-10 |                                       |
| M2.1 | done   | -        | -         |                                          |
| M2.2 | done   | -        | -         |                                          |
| M2.3 | todo   | -        | -         |                                          |
| M2.4 | todo   | -        | -         |                                          |
| M2.5 | todo   | -        | -         | pre-agreed cut 1                         |
| M2.6 | todo   | -        | -         |                                          |
| M3.1 | todo   | -        | -         |                                          |
| M3.2 | todo   | -        | -         |                                          |
| M4.1 | todo   | -        | -         |                                          |
| M4.2 | todo   | -        | -         |                                          |
| M4.3 | todo   | -        | -         | `Orphans` scope is pre-agreed cut 3      |
| M4.4 | todo   | -        | -         |                                          |
| M4.5 | todo   | -        | -         |                                          |
| M4.6 | todo   | -        | -         |                                          |
| M5.1 | todo   | -        | -         | Grep-side `Starred` is pre-agreed cut 4  |
| M6.1 | todo   | -        | -         |                                          |
| M6.2 | todo   | -        | -         | pre-agreed cut 2                         |
| M7.1 | todo   | -        | -         |                                          |
| M7.2 | todo   | -        | -         |                                          |
