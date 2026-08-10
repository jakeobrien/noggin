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

Current task: M1.3

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M1.2 done: new `cmd/noggin/docstore.go` — `docRecord{Starred, LastViewed}` keyed by absolute
path, its own file (`documents`, `docStoreName`) alongside the Roots store under `dataDir()`
(M1.1), own tab-separated two-field format (D67) — no attachment to Roots, no interaction with
store.go's plain-list Roots format at all. `loadDocStore`/`saveDocStore` tolerate
missing/malformed the way `loadStore` does (bad line dropped, load continues). Tests cover
round-trip (starred+viewed, starred+unviewed, unstarred+viewed), absent-file, garbage-line,
and a record for a path under no session Root loading exactly like any other path
(`TestDocStoreNoAttachmentToRoots`). Nothing wired into main.go/model.go yet — that's M1.4.
`make check` green. Next: M1.3, bound the store (evict oldest `lastViewed` first, never evict a
starred record, eviction clears the timestamp but leaves the star) — needs an eviction
function operating on the `map[string]docRecord` this task introduced; no cap number is fixed
yet in SPEC/DECISIONS, so pick one and log it (D-series) unless PLAN.md's M1.3 entry already
pins one — recheck it first.

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
| M1.2 | done   | -        | -         |                                          |
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
