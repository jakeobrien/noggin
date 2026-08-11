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

Current task: M4.6

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

Build (2026-08-11): completed M4.5 (flat-list rows support Activate/Peek/Edit and their DV2
targeting; Expand/Collapse become no-ops). Re-read TRIAGE T21/D80 first as flagged: D80 (M4.3)
had already fixed Activate/Star's flat-row resolution, leaving Peek (and, transitively, Edit)
as the actual remaining gap — Edit itself needed no code change since it always acts on the
focused viewer's Active Document, never Files Nav selection, so it is already correct once
Activate/Peek load the right document into that viewer. `model.go`'s `togglePeekInto` now
checks `m.files.flatMode()` first and delegates to new `togglePeekFlatRowInto` (mirrors
`activateFlatRowInto`'s shape exactly): resolves against `selectedFlatRow()` instead of the
unrelated tree row `selectedNode()` would give, and refuses a broken Starred row with a footer
message ("%s no longer exists; not peeked") before attempting the load, the same "cannot be
activated" treatment S3 already gives Activate. `files.go`'s `expand()`/`collapse()` each gained
an `if f.flatMode() { return f }` guard at the top — previously they mutated `f.expanded`
against whatever tree node `selectedNode()` happened to resolve from a flat-list index, an
invisible-but-real state corruption TRIAGE T21 called out; now a true no-op, mode/selection/
expansion all provably unchanged. New `flat_actions_test.go`: Peek-into-DV1, Peek-into-DV2,
Activate-into-DV1-and-DV2 (all against a Recent row, per M4.5's own acceptance wording), a
Peek-refuses-broken-Starred-row test mirroring M4.3's Activate one, and an Expand/Collapse-
no-op test asserting `f.expanded`, `f.mode` and `f.selected` are all bit-for-bit unchanged
across both key presses in Recent mode. `make check` and `make check-pty` both green (full
runs). TRIAGE T21 updated (not cleared — replan's job) to record every action it named is now
resolved.

Not done / logged: TRIAGE T20 unchanged (build/medium — FilesMode/FilesScope discoverability).
TRIAGE T22 unchanged (build/low — Tree's own "no root"/"empty root" message conflation, from
M4.4, still out of scope here too).

Next: cursor is **M4.6** (session-only browse memory: mode remembered per Root, scope
remembered per (Root, mode); every session starts on Tree) — needs M4.1 (done). Nothing blocks
it.

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
| M4.4 | done   | -        | -         | render test walks all six mode/scope combos; TRIAGE T22 |
| M4.5 | done   | -        | -         | Peek fixed via togglePeekFlatRowInto; Expand/Collapse no-ops; TRIAGE T21 |
| M4.6 | todo   | -        | -         |                                          |
| M5.1 | todo   | -        | -         | Grep-side `Starred` is pre-agreed cut 4  |
| M6.1 | todo   | -        | -         |                                          |
| M6.2 | todo   | -        | -         | pre-agreed cut 2                         |
| M7.1 | todo   | -        | -         |                                          |
| M7.2 | todo   | -        | -         |                                          |
