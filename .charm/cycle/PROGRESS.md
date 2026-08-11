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

Current task: M4.3

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

Build (2026-08-11): completed M4.2 (`Recent` mode's actual flat-list rows). New
`browse_list.go`: `browseRow` (path/name/rootLabel-or-orphanDir/orphan) and `recentRows(docs,
roots, activeRootPath, scope)` — buckets every non-zero-`LastViewed`, still-on-disk
(`pathMissing`) record by its covering session Root (`roots.go`'s new `rootForPath`, first
match, roots never nest) or into a `""` orphan bucket, sorts each bucket newest-first
(`sortByRecency`, tie-broken by path) and caps it at `recentPerRootCap` (100, S6) *before*
scope is applied — Active Root reads just that Root's own bucket (`activeRootPath == ""`
short-circuits to none, never `buckets[""]` by accident), All Roots merges every
already-capped bucket (orphans included — D79: this is where Recent's own two right-hand
label forms both appear, since Recent's scope set has no dedicated Orphans value the way
Starred's does) sorted again with no further cap. `filesNav` gained a `docs`/`sessionRoots`/
`activeRootPath` snapshot (`withRecentSource`) that `model.refreshFilesRecentSource` pushes
in at every point that could change it (`withDocs`, `withDocRecord`, `openFolder`,
`applyResolution`, `refresh`) — rows themselves are never cached, `flatRows()`/`recentRows`
recompute fresh on every `rowsView` call instead (D79, same "no cache to go stale" shape
M2.11 chose over M2.7's marker cache). `rowsView` branches on `mode == recentBrowse` to a new
`flatRowsView`/`renderFlatRow` (star marker per S3, name left, label right-aligned via
`leftTruncate`, `recentEmptyText` placeholder pending M4.4's per-combination wording); `move`/
`rowCount` now count whichever list (`flatRows()` or `visibleRows()`) the current mode
actually shows. `browse_list_test.go` covers ordering, the per-Root cap, the merged All
Roots ordering + both label forms, Active-Root-only isolation, omission of a deleted file,
zero-`LastViewed` exclusion, the no-active-root edge case, and two model-level render checks
(Recent shows the flat rows and its own empty message). `make check` and `make check-pty`
both green (full runs).

Not done / logged: TRIAGE T20 (unchanged, build/medium — FilesMode/FilesScope still lack a
footer hint or Help Modal listing). New TRIAGE T21 (build, medium): while browsing Recent
mode, Star/Activate/Peek/Edit still act on `filesNav.selectedNode()`'s tree row (same index,
unrelated list) rather than the flat row on screen — M4.5's own acceptance wording names
Activate/Peek/Edit-and-DV2-targeting and "Expand/Collapse become no-ops" but never mentions
Star, so replan should confirm M4.5 covers it too or add a task that does. D79 (DECISIONS.md)
records the orphan-bucket-cap and live-recompute judgment calls this task made where SPEC.md
was silent.

Next: cursor is **M4.3** (`Starred` mode's own flat-list rows across Active Root/All
Starred/Orphans) — needs M4.1 (done), M2.1 (done). Nothing blocks it. TRIAGE has two open
items (T20, T21; both non-blocking) — a future replan should triage both, and M4.4 (empty-
state messaging) still needs M4.3 done first.

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
| M4.3 | todo   | -        | -         | `Orphans` scope is pre-agreed cut 3      |
| M4.4 | todo   | -        | -         |                                          |
| M4.5 | todo   | -        | -         |                                          |
| M4.6 | todo   | -        | -         |                                          |
| M5.1 | todo   | -        | -         | Grep-side `Starred` is pre-agreed cut 4  |
| M6.1 | todo   | -        | -         |                                          |
| M6.2 | todo   | -        | -         | pre-agreed cut 2                         |
| M7.1 | todo   | -        | -         |                                          |
| M7.2 | todo   | -        | -         |                                          |
