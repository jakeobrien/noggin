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

Current task: M3.1

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M2.5 done — star/un-star the Picker's selected Result (pre-agreed cut 1, not needed after
all). `picker.go`'s new `toggleStarPicker`/`pickerSelectedPath` toggle the S1 document store
(`docRecord`/`withDocRecord`, the same pair every other star surface uses) for whichever mode is
open — `pickerPreviewCandidate` (fileMode) or `pickerGrepCandidate` (grepMode), both already
gated on `resultsCurrent` so a stale search generation can't be starred any more than it can be
previewed or activated. Deliberately touches nothing on `m.picker` itself, which is what makes
S2's "leaves the Result in place with its marker cleared" true for free — nothing here could
move the row or the cursor.
Gotcha, real design conflict: `keys.Star` is a bare `"s"`, unusable in the Picker's
always-typing Prompt (`handlePickerInput`) — every `s` typed into a query would have toggled the
star instead of typing. Added a new binding, `PickerStar` (`alt+s`, mirroring `ToggleHidden`'s
own `alt+` prefix), rather than reusing `keys.Star` — logged as **D69**. Wired into
`pickerHints()` (Footer) but *not* `helpBindings()`'s per-pane lists — its Help description
("star") collides with `TestHelpModalListsThePaneOwnBindings`, which already pins plain
`keys.Star` absent from the Files Nav's/viewer's own Help Modal; reworking that pinned wording
for one binding wasn't worth it. Filed as **TRIAGE T6** (plus **T5**, a pre-existing M2.1/M2.2
gap noticed along the way: the Files Nav's and Document Viewer's own `keys.Star` are likewise
missing from their panes' Footer/Help hints). Also updated
`TestPickerHintTextNamesPickerActionsNotThePane` (footer_test.go): "star" moved from its
not-wanted list to its wanted one now that the Picker has a real star action.
Evidence: `TestPickerStarTogglesSelectedFileResult` (M2.5's acceptance clause verbatim — star,
then un-star, `picker.results`/`picker.selected` byte-identical throughout, restart
round-trips), `TestPickerStarTogglesSelectedGrepResult` (S2 doesn't restrict the action to File
mode, only M5.1's Starred *scope* is cut), `TestPickerStarNoOpWithNothingSelected` — all in new
`picker_star_test.go`. `make check` green.
Next: M2.6 (needs: M2.1 done, M3.1 not yet) is skipped for now — its acceptance criterion needs
`lastViewed` records to exist before it can check they survive a Root un-starring, and M3.1
doesn't exist yet. Cursor moves to **M3.1** instead: record `lastViewed` when a document is
Activated in DV1/DV2 or opened via the `$EDITOR` handoff, never on a Peek. `toggleActiveDocStar`/
`toggleDocStar`'s `docRecord`/`withDocRecord` pattern applies again; the new call sites are
`model.activate`/`activateInto` (Files Nav), `model.activateRoot`'s document-opening path if any,
and `finishEdit` (the `$EDITOR` handoff) — not yet located precisely this session. Return to M2.6
once M3.1 lands.

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
| M2.3 | done   | -        | -         |                                          |
| M2.4 | done   | -        | -         |                                          |
| M2.5 | done   | -        | -         | pre-agreed cut 1                         |
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
