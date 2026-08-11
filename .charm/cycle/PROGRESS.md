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

Current task: M4.2

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

Build (2026-08-11): completed M4.1. Added `browse.go`: `browseMode`
(treeBrowse/recentBrowse/starredBrowse) and `browseScope`
(activeRootBrowseScope/allRootsBrowseScope/allStarredBrowseScope/orphansBrowseScope), plus
`modeScopes(mode)` giving each mode its own ordered scope set (none for Tree, two for
Recent, three for Starred — Orphans is in the model now even though M4.3's own *listing* of
it is pre-agreed cut 3, per that task's note; costs nothing to carry). `filesNav` gained
`mode`/`scope` fields, `cycleMode`/`cycleScope` (a mode switch resets scope to the new
mode's own first value — no M4.6 per-Root/per-mode memory yet, that's its own task) and
`scopeLabel` ("-" for Tree regardless of the scope field). `view()` now always reserves its
first line for the mode/scope row (`modeScopeRow`, " · "-joined like the footer's own hints)
and renders the row list into the remaining height-1 (`rowsView`), with a height==1 guard so
a pane with room for only the header doesn't overshoot the exactly-height-lines contract.
keys.FilesMode/keys.FilesScope (tab/shift+tab, reusing the Picker's own physical chords —
the two contexts are mutually exclusive by construction, see keymap.go's own comment) are
matched in `filesNav.handleKey` beside Expand/Collapse, so they only ever fire with Files
Nav Focus, same as every other Files-Nav-only binding. Updated the two pre-existing
files_test.go view tests for the one-row shift; added `browse_test.go` covering the mode
cycle/wrap, each mode's own scope set and wrap, Tree's scope no-op, the reset-on-mode-switch
rule, the row's own text per mode, the one-row-height cost (a box that fits N rows loses one
to a box N-1 tall the same way), and the model-level "acts only from Files Nav Focus" /
"acts and renders immediately from it" pair. `make check` and `make check-pty` both green
(full runs).

Not done, logged as TRIAGE T20 (build, medium): keys.FilesMode/FilesScope have no footer
hint or Help Modal listing yet — every other Files-Nav-only binding got one alongside its
own task (Peek/ActivateDV2/PeekDV2), and M2.8 set precedent for treating this as its own
task rather than folding it into the acceptance criterion that introduced the binding.

Next: cursor is **M4.2** (`Recent` mode's actual flat-list rows) — needs M4.1 (done), M3.1
(done), M3.3 (done). Nothing blocks it. TRIAGE has one open item (T20, non-blocking); a
future replan should decide whether it becomes its own task or is folded into whichever of
M4.2/M4.3/M4.5 next touches filesPane's hint lists.

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
