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

Current task: M2.4

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M2.3 done — star marker on Files Nav rows. Gave `filesNav` a `starred map[string]bool` field
(files.go) alongside the existing `status` (S8 git markers) field — same shape, same
"caller-recomputed snapshot, not a live lookup" contract — plus a `withStarred` setter and a
third 2-char row column (`"★ "`/`"  "`, files only, same glyph the Roots Nav's own renderRow
already uses). `model.filesStarredMarks()` derives the set fresh from `m.docs` (just the
`Starred` paths); called at `openFolder` and `refresh` (alongside `withStatus`) *and* inside
`toggleDocStar` itself — unlike git status, this store can change from a Files Nav keypress, so
the marker must update on the very next render, not just after a reload.
Gotcha: inserting the new column shifted three existing git-status render tests' expected
substrings (`gitstatus_test.go`, `refresh_test.go`) from e.g. `"M README.md"` to
`"M   README.md"` (status's own trailing space + the new unstarred `"  "` column) — fixed and
commented inline; this is the kind of row-layout ripple to expect from M2.4 (viewer title row)
and M7.1 (focus badges) too, both touching shared rendered rows/titles.
Evidence: `TestFilesNavRendersStarMarkers` (M2.3's acceptance clause — starred file marked,
unstarred file and a directory not, `Starred`-mode clause vacuous until M4.3 exists),
`TestToggleDocStarUpdatesTheRowMarkerImmediately` (both in new `files_star_render_test.go`).
`make check` green. `make check-pty`: confirmed pre-existing, unrelated M3.3/M3.4 failures
(TRIAGE T4 — the pty harness's `config_env` still targets the pre-M1.1 config-dir Roots-store
path); identical failure set on the commit before M2.1 too, nothing this session touched or
should fix.
Next: M2.4, a star indicator to the left of the document title in each Document Viewer's title
row (needs M1.4, done), legible with the Navigator hidden. Title-row rendering lives wherever
`paneTitle`/the viewer's title bar is built (not yet located this session) — check `layout.go`/
`model.go`'s `paneTitle` and how `renderPane` draws a pane's title line before assuming a single
edit site; likely reads `m.docRecord(v.doc.path).Starred` the same way `toggleActiveDocStar`
already reads `v.doc.path`.

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
