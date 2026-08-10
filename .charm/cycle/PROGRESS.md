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

Current task: M1.4

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M1.3 done: `evictDocStore(records, cap)` in `cmd/noggin/docstore.go`, `docStoreCap = 1000`
(D68 — no cap number existed anywhere before this task; picked and logged it). Walks records
oldest-`LastViewed`-first once over cap: deletes unstarred candidates, clears (never removes)
a starred candidate's timestamp, stops as soon as back within cap — so a starred record
reached only after the cap is already satisfied is left fully untouched, and if unstarred
records run out before reaching cap the store is deliberately left over cap rather than ever
removing a star. Pure function, not yet wired into load/save — M1.4 owns wiring the store into
the model lifecycle (single accessor pair calling load -> mutate -> evict -> save).
`TestEvictDocStoreBoundsAndProtectsStars` covers all three M1.3 done-when clauses in one
scenario; `make check` green.
Next: M1.4 — wire the doc store into model.go the way store.go's Roots store already is
(main.go loads once via a new loadDocStore-based entry point, threads a path through
withStorePath-style, model exposes a read/write accessor pair; every future consumer —
Files Nav M2.1, DV1/DV2 M2.2, Picker M2.5, recency M3.1/M3.2 — goes through that pair, never
touching docstore.go's load/save/evict directly). Watch for the cap: M1.4's write path should
run eviction before every save (docs won't self-evict on their own).

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
| M1.3 | done   | -        | -         |                                          |
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
