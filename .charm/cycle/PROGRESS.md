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

Current task: M2.8

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M2.7 done — the Files Nav's S3 star-marker set now stays in sync regardless of which
surface stars a document. Root cause: `toggleActiveDocStar` (Document Viewer) and
`toggleStarPicker` (Picker) both wrote through `withDocRecord` but only `toggleDocStar`
(the Files Nav's own action) refreshed `m.files.starred` afterwards, so a star flipped
from DV1/DV2 or the Picker left that same file's Files Nav row stale until the pane's
next openFolder/refresh. Fix: centralized the `m.files.withStarred(m.filesStarredMarks())`
refresh inside `withDocRecord` itself (model.go:239-268) — the one write half all three
star toggles and `recordViewed` already share — and removed the now-redundant explicit
call from `toggleDocStar`. New tests `TestFilesNavMarkerSyncsWhenStarredFromDocumentViewer`
/ `...FromPicker` (files_star_sync_test.go) reproduce the stale marker on the pre-fix code
(confirmed via `git stash`) and pass with the fix. `make check` and `make check-pty` both
green apart from the pre-existing T11 (unrelated, M2.3's marker-column spacing vs. two
C2a/M4.1 pty checks — still open, not touched here).
Next: **M2.8** — Footer/Help Modal discoverability for `keys.Star`/`keys.PickerStar`
(TRIAGE T5/T6); its dependencies (M2.1, M2.2, M2.5) are all done. M2.9 (naming cleanup)
is independently available too if a different slice is wanted.

## Status

| Key  | Status | Reviewed | Validated | Note                                     |
|------|--------|----------|-----------|------------------------------------------|
| M1.1 | done   | 2026-08-10 | 2026-08-10 |                                       |
| M1.2 | done   | 2026-08-10 | 2026-08-10 |                                       |
| M1.3 | done   | 2026-08-10 | 2026-08-10 |                                       |
| M1.4 | done   | 2026-08-10 | 2026-08-10 |                                       |
| M2.1 | done   | 2026-08-10 | -         |                                          |
| M2.2 | done   | 2026-08-10 | -         |                                          |
| M2.3 | done   | 2026-08-10 | -         |                                          |
| M2.4 | done   | 2026-08-10 | -         |                                          |
| M2.5 | done   | 2026-08-10 | -         | pre-agreed cut 1                         |
| M2.6 | done   | 2026-08-10 | -         |                                          |
| M1.1a | done  | -        | -         | replan: pty harness fix, TRIAGE T4 (D71) |
| M2.7 | done   | -        | -         | replan: stale marker fix, TRIAGE T7 (D70) |
| M2.8 | todo   | -        | -         | replan: star-binding discoverability, TRIAGE T5/T6 |
| M2.9 | todo   | -        | -         | replan: naming cleanup, TRIAGE T1/T2/T9  |
| M3.1 | done   | -        | -         |                                          |
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
