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

Current task: M1.2

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M1.1 done: added `dataDir()` in `cmd/noggin/store.go` (hand-rolled, mirrors
`os.UserCacheDir()`'s own Unix/Darwin/Windows/Plan9 split — D66) and rewired `storePath()`
through it instead of `os.UserConfigDir()`. Starred-Roots store is now a clean break: old
config-dir `roots` file is never read/migrated (`TestStoreIgnoresConfigDirCleanBreak`).
Updated doc comments in rootsnav.go/model.go/main.go/dv2_persistence_test.go that referenced
the old `os.UserConfigDir()` mechanism; the "no resolvable config directory" user message is
now "no resolvable data directory" (message_test.go only checks substring, unaffected).
`make check` green. Next: M1.2, the path-keyed document store — per the planning gotcha, give
it its own file/names rather than overloading store.go's Roots-specific plain format.

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
| M1.1 | done   | -        | -         |                                          |
| M1.2 | todo   | -        | -         |                                          |
| M1.3 | todo   | -        | -         |                                          |
| M1.4 | todo   | -        | -         |                                          |
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
