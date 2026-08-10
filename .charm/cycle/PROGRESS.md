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

Current task: M3.2

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M3.1 done — `lastViewed` is now recorded at exactly the two call sites S4 names:
`model.activateInto` (Files Nav Activate into DV1 or, via the `ctrl+` modifier, DV2) and
`model.finishEdit` (the `$EDITOR` handoff), both via a new `model.recordViewed` helper that
wraps `docRecord`/`withDocRecord` the same way `toggleDocStar`/`toggleActiveDocStar` already do
for `Starred`. `model.activateRoot` and `togglePeekInto` were left untouched, so Root activation
and Peek (plain and `ctrl+` into DV2) never record, per S4's own trigger list.
D73 logs a scope call: launch-time document opens (`applyResolution`, which shares
`openDocumentInto` with the two recording call sites) do **not** record — S4/M3.1's wording
names Activate and the edit handoff only, not launch.
Since `withDocRecord` declines with a footer message when `docsPath == ""` (M1.4's contract,
now also inherited by recording), one pre-existing test
(`TestFinishEditReloadsDocumentFromDisk`) needed a resolvable `docsPath` via `withDocs` added
to keep its own "message cleared" assertion honest — the only pre-existing test this change
touched; `go test ./...` found no other collateral failures.
Evidence: new `recency_test.go` — `TestActivateRecordsLastViewed` and
`TestActivateDV2RecordsLastViewed` confirm a timestamp at/after a captured `before` for each
viewer (the DV1 case also round-trips through a simulated restart), `TestFinishEditRecordsLast-
Viewed` covers the edit handoff, and `TestPeekDoesNotRecordLastViewed` (subtests `Peek`/
`PeekDV2`) confirm the record stays zero. `make check` green.
Next: **M3.2** — record regardless of Root provenance (ad-hoc, un-starred, or absent Roots):
since `recordViewed`/`docRecord`/`withDocRecord` already have no Root-awareness at all (M1.2's
"no attachment to Roots", inherited structurally the same way M2.6 found for the star field),
this is likely another verification-only task — write a test that Activates a document from an
ad-hoc Root and one whose Root is not in the session at all, and confirm both records persist,
before assuming any production change is needed.

Replan (2026-08-10): triaged TRIAGE.md's nine open items — added four new tasks (M1.1a pty
harness fix, M2.7 stale Files-Nav-marker fix, M2.8 star-binding discoverability, M2.9 naming
cleanup) logged as D70/D71 for the material ones; deferred T3 (shared store save/load helper)
to ROADMAP.md Unscheduled; opened D72 for T8's Picker-star-marker wording question — needs a
user call on whether to add a visible marker to Picker rows or amend S2/M2.5's wording, since
S3 never lists the Picker as a marker surface. TRIAGE.md is now empty. None of these four new
tasks block M3.1; cursor stays there, but they're cheap (effort 1-2) and touch files M3+ will
also touch, so consider slotting them in opportunistically rather than saving them for a
dedicated pass at the end.

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
| M1.1a | todo  | -        | -         | replan: pty harness fix, TRIAGE T4 (D71) |
| M2.7 | todo   | -        | -         | replan: stale marker fix, TRIAGE T7 (D70) |
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
