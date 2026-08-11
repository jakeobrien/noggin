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

Current task: M4.4

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

Build (2026-08-11): completed M4.3 (`Starred` mode's own flat-list rows across Active
Root/All Starred/Orphans). `browse_list.go` gained `starredRows(docs, roots, activeRootPath,
scope)`, `recentRows`'s own shape with the three differences S3/S6 call for a Starred list:
every `Starred` record appears regardless of `LastViewed` (S1's independent fields), a record
whose file no longer exists is *marked broken* rather than omitted (`browseRow` gained a
`broken` field, always computed by `buildBrowseRow` via `pathMissing`), and rows sort by name
(no timestamp axis to order by) rather than recency, with no per-Root cap (`recentPerRootCap`
is Recent's own S6 bound). Buckets reuse `recentRows`' own `rootForPath`-keyed shape (a `""`
orphan bucket, never capped); Active Root reads only its own bucket, All Starred merges every
bucket, Orphans reads just `""`. `files.go`: `flatMode()` (`mode == recentBrowse ||
starredBrowse`) is now the one predicate `flatRows`/`rowCount`/`rowsView` branch on;
`flatRows()` dispatches to `recentRows`/`starredRows` by mode; `renderFlatRow` branches on
mode too — Recent keeps its star-marker-prefix left segment, Starred instead omits the marker
(S3: "membership already says it") and marks a broken row dimmed with a `" (missing)"` suffix
(mirrors `rootsnav.go`'s broken-Root treatment); `starredEmptyText` is the new mode's own
placeholder pending M4.4. New `selectedFlatRow()` accessor (the flat row at `f.selected`
while `flatMode()`) — added because M4.3's own acceptance ( "activating a broken star is
refused the way a Broken Root is", "un-starring drops the row") cannot be tested at all
without *some* path to attempt those two actions on a flat row, and none existed (TRIAGE
T21's exact citation). `activateInto`/`toggleDocStar` (model.go) now route through
`selectedFlatRow()` first while `flatMode()`, falling back to the tree path only for Tree;
`activateFlatRowInto` refuses a broken row with `activateRoot`'s own wording
(`"%s no longer exists; not activated"`). Scoped to exactly these two actions — Peek/Edit/
DV2-targeting still resolve via `selectedNode()` (still wrong while Recent/Starred is showing)
until M4.5; Expand/Collapse untouched. D80 (DECISIONS.md) records this judgment call and its
scoping, and narrows (not closes) TRIAGE T21 to Peek/Edit only. New
`browse_starred_test.go`: scope-membership tests (Active Root/All Starred/Orphans, including
an orphan appearing in the latter two only), unstarred-exclusion, never-viewed-inclusion,
broken-marked-not-omitted, no-active-root edge case, name ordering, two model-level render
checks (flat rows without a star marker, empty message, broken-row "(missing)" marking), and
the two new model-level action tests (`activateInto` refuses/opens a Starred row,
`toggleDocStar` drops a displayed Starred row immediately). `make check` and `make check-pty`
both green (full runs).

Not done / logged: TRIAGE T20 unchanged (build/medium — FilesMode/FilesScope discoverability).
TRIAGE T21 narrowed by D80 to Peek/Edit only (Activate/Star now correctly target the flat row
in Recent/Starred) — a future replan should re-read T21 against D80 before retriaging it, and
confirm M4.5's own acceptance wording still needs updating for Peek/Edit-only rather than the
original Activate/Peek/Edit/Star list.

Next: cursor is **M4.4** (every mode/scope combination with nothing to show gets a
considered message; also the point to reconcile `recentEmptyText`/`starredEmptyText`'s
current placeholder wording) — needs M4.2 (done), M4.3 (done). Nothing blocks it.

## Status

| Key  | Status | Reviewed | Validated | Note                                     |
|------|--------|----------|-----------|------------------------------------------|
| M1.1 | done   | 2026-08-10 | 2026-08-10 |                                       |
| M1.2 | done   | 2026-08-10 | 2026-08-10 |                                       |
| M1.3 | done   | 2026-08-10 | 2026-08-10 |                                       |
| M1.4 | done   | 2026-08-10 | 2026-08-10 |                                       |
| M2.1 | done   | 2026-08-10 | 2026-08-11 |                                      |
| M2.2 | done   | 2026-08-10 | 2026-08-11 |                                      |
| M2.3 | done   | 2026-08-10 | 2026-08-11 |                                      |
| M2.4 | done   | 2026-08-10 | 2026-08-11 |                                      |
| M2.5 | done   | 2026-08-10 | 2026-08-11 | pre-agreed cut 1, not actually cut   |
| M2.6 | done   | 2026-08-10 | 2026-08-11 |                                      |
| M1.1a | done  | 2026-08-11 | 2026-08-11 | replan: pty harness fix, TRIAGE T4 (D71) |
| M2.7 | done   | 2026-08-11 | 2026-08-11 | replan: stale marker fix, TRIAGE T7 (D70) |
| M2.8 | done   | 2026-08-11 | 2026-08-11 | replan: star-binding discoverability, TRIAGE T5/T6 |
| M2.9 | done   | 2026-08-11 | 2026-08-11 | replan: naming cleanup, TRIAGE T1/T2/T9  |
| M3.1 | done   | 2026-08-11 | 2026-08-10 | SPEC/GLOSSARY-level gap re: Picker Activate not recording lastViewed — tracked as TRIAGE T12, not a M3.1 overclaim |
| M3.2 | done   | 2026-08-10 | 2026-08-10 |                                          |
| M3.3 | done   | 2026-08-10 | 2026-08-10 | replan: Picker never records lastViewed, TRIAGE T12 (D74) |
| M3.4 | done   | 2026-08-10 | 2026-08-10 | replan: recordViewed failure message clobber, TRIAGE T13; no bug found, test coverage added; review found coverage gap re: Picker call sites, resolved as M3.5 (D77) |
| M3.5 | done   | 2026-08-10 | 2026-08-11 | replan: Picker-path store-write-failure regression tests, TRIAGE T18 (D77) |
| M2.10 | done  | 2026-08-10 | 2026-08-11 | replan: pty star-marker column checks, TRIAGE T11        |
| M2.11 | done  | 2026-08-10 | 2026-08-10 | replan: Picker star marker (D72, TRIAGE T8), built same session; review found stale doc comment, resolved as M2.12 (D76); validate found Grep-mode live-toggle coverage gap, resolved as M2.13 (D78) |
| M2.12 | done  | 2026-08-10 | 2026-08-11 | replan: correct starMarker() doc comment, TRIAGE T17 (D76) |
| M2.13 | done  | 2026-08-10 | 2026-08-11 | replan: Grep same-frame star-toggle render test, TRIAGE T19 (D78) |
| M4.1 | done   | -        | -         | discoverability gap logged as TRIAGE T20   |
| M4.2 | done   | -        | -         | orphan-bucket cap + live-recompute judgment calls, D79; TRIAGE T21 |
| M4.3 | done   | -        | -         | flat-row Activate/Star wiring added beyond listing, D80; narrows TRIAGE T21 |
| M4.4 | todo   | -        | -         |                                          |
| M4.5 | todo   | -        | -         |                                          |
| M4.6 | todo   | -        | -         |                                          |
| M5.1 | todo   | -        | -         | Grep-side `Starred` is pre-agreed cut 4  |
| M6.1 | todo   | -        | -         |                                          |
| M6.2 | todo   | -        | -         | pre-agreed cut 2                         |
| M7.1 | todo   | -        | -         |                                          |
| M7.2 | todo   | -        | -         |                                          |
