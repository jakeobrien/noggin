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

M2.6 done — confirmed the Root lifecycle never touches document state, without needing M3.1
first. `model.toggleStar` (the Roots star toggle) only ever touches `m.roots` and the Roots
store (`toggleStore`) — by construction, it has no code path that reaches `m.docs`/
`withDocRecord` at all (D61's "not attached to Roots" already holds structurally, not just by
convention), so M2.6 turned out to be a verification-only task rather than needing any
production change.
Un-blocked ahead of M3.1: rather than waiting for M3.1's automatic `lastViewed` recording, the
test seeds a "recently-viewed" record directly via `saveDocStore`/`withDocs` (the same fixture
shape `TestWithDocRecordEvicts` already uses for a timestamp-bearing record) — M2.6's claim is
about the Root lifecycle, not about how a timestamp got there, so a hand-seeded one proves the
same thing M3.1's real recording would. Comparison is `reflect.DeepEqual` against the store
re-read through `loadDocStore` (not the raw seed struct) to sidestep `time.Time`'s monotonic-
reading footgun, plus a raw on-disk byte comparison for the literal "byte-identical" wording.
Evidence: `TestUnstarringRootLeavesDocumentStoreUntouched` (new `root_lifecycle_docs_test.go`)
stars then un-stars a Root holding one starred and one recently-viewed (unstarred) document,
confirms `model.docs` and the on-disk store file are both unchanged throughout. `make check`
green.
Next: **M3.1** — record `lastViewed` when a document is Activated in DV1 or DV2, or opened via
the `$EDITOR` handoff; Peek never records. `toggleActiveDocStar`/`toggleDocStar`'s
`docRecord`/`withDocRecord` pattern applies again, called alongside (not instead of) each
action's existing effect. Likely call sites, not yet confirmed this session: `model.activate`/
`model.activateInto` (Files Nav Activate into DV1/DV2), and `model.finishEdit` (the `$EDITOR`
handoff) — `model.activateRoot` itself should NOT record (S4 names Activate-in-a-viewer and the
edit handoff only, not "Root activation"). Tests need to confirm a Peek (including a `ctrl+`
peek into DV2, `togglePeekInto`) leaves the record untouched — Peek's own read path
(`loadDocument` in `togglePeekInto`) must stay separate from whatever wraps Activate's.

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
| M2.1 | done   | 2026-08-10 | -         |                                          |
| M2.2 | done   | 2026-08-10 | -         |                                          |
| M2.3 | done   | 2026-08-10 | -         |                                          |
| M2.4 | done   | 2026-08-10 | -         |                                          |
| M2.5 | done   | 2026-08-10 | -         | pre-agreed cut 1                         |
| M2.6 | done   | 2026-08-10 | -         |                                          |
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
