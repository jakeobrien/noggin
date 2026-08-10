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

Current task: M2.2

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M2.1 done — Files Nav star/un-star, the accessor pair's first real caller. `model.toggleDocStar()`
(model.go) reads `m.files.selectedNode()`, no-ops on a nil/dir selection (mirrors `activate()`'s
own selection guard), else flips `docRecord(path).Starred` and writes it back through
`withDocRecord` — no new logic of its own, `withDocRecord` already owns eviction, persistence
and the decline-with-message contract. Wired on the *existing* `keys.Star` chord (`s`), matched
in `Update`'s `filesPane` case beside `Peek`/`ActivateDV2`/`PeekDV2` — reused rather than a new
binding since M2.1 didn't call for a distinct key and the Roots Nav's own `s` is already free
(Roots Nav and Files Nav are never focused at once). `TestEveryBindingIsCovered` needed no
update: `keyMap`'s field count didn't change.
Evidence: `TestToggleDocStarPersistsAndUpdatesDisplay` (restart round-trip + un-star, M2.1's
acceptance clause), `TestFilesNavStarKeyTogglesDocStar` (the actual keypress through `Update`,
not just the method), `TestToggleDocStarOnFolderIsNoOp`, `TestToggleDocStarWithNoDataDirDeclinesAndMessages`
(all in the new `files_star_test.go`). `make check` green.
Next: M2.2, star/un-star from a focused Document Viewer acting on *that* viewer's Active
Document (DV1 vs DV2 independently) — likely a `toggleDocStarForViewer(pane)` alongside
`toggleDocStar`, matched in `Update`'s `dv1Pane`/`dv2Pane` case rather than a global switch
entry, since DV1/DV2 have independent Active Documents.

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
| M2.2 | todo   | -        | -         |                                          |
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
