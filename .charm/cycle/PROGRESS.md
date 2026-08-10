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

Current task: M2.1

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M1.4 done — M1 (the document store) is now fully built. Wired the S1 store into model.go:
`docsPath`/`docs` fields, `withDocs(path, docs)` for one-shot startup loading (main.go now
calls `docStorePath()` + the new `loadDocuments()` wrapper alongside the existing Roots-store
wiring), and the single accessor pair everything else must use — `docRecord(path) docRecord`
(read, zero value for an unknown path, no Root attachment) and `withDocRecord(path, rec) model`
(write: clones the map, applies `evictDocStore` (M1.3), persists via `saveDocStore`, updates
`m.docs` only on success). `withDocRecord` declines the *whole* update — in-memory included —
with an `m.message`, both when `docsPath == ""` (unresolvable data dir) and on a save error,
mirroring `toggleStar`'s existing "must not appear to have taken effect" contract for the Roots
store. Nothing in files.go/viewer.go/picker.go calls this yet — that starts at M2.1; M1.4 only
had to expose the accessor pair, not consume it.
Evidence: `TestWithDocRecordPersistsImmediately` (restart test — M1.4's acceptance clause),
`TestDocRecordDefaultsToZeroValue`, `TestWithDocRecordDeclinesOnUnresolvableDataDir`,
`TestWithDocRecordEvicts` (confirms the write path bounds the store both in memory and on
disk). `make check` green; `TestEveryBindingIsCovered` untouched (M1.4 adds no bindings).
Next: M2.1, star/un-star the Files Nav's selected file. Model for the action:
`m = m.withDocRecord(path, rec)` where `rec := m.docRecord(path)` with `Starred` flipped —
mirrors the `toggleDocStar`-shape sketched in M1.4's own doc comments but not yet written
anywhere; M2.1 is the first real caller of the accessor pair.

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
| M2.1 | todo   | -        | -         |                                          |
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
