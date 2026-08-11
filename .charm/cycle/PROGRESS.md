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

Current task: M5.1

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

Build (2026-08-11): completed M4.6 (session-only browse memory: mode remembered per Root,
scope remembered per (Root, mode); every session starts on Tree). `files.go`'s `filesNav` gained
a live `scopeByMode map[browseMode]browseScope` field (the current Root's own per-mode scope
history) and a `withScopeRemembered` helper that snapshots the outgoing mode's live `f.scope`
into it — called from both `cycleMode` (leaving a mode without leaving the Root) and `remember`
(leaving the Root itself), so neither path can lose a mode's scope. `scopeForCurrentMode` reads
it back (falling through to `modeScopes(mode)[0]`, or the zero value for Tree, for a mode this
Root has never shown). `filesNavMemory` gained `mode`/`scopeByMode` fields alongside its existing
`expanded`/`selectedPath`, so `setRoot` now restores all four together on a Root it has shown
before, and resets all four to Tree/no-scope-memory on one it hasn't — the same shape M3.1's
expansion/selection round-trip already had, just extended to the two new axes. `cycleMode`'s own
behavior changed: switching mode now restores that mode's last scope on the *current* Root
(not just resets to the new mode's first value) — this was flagged as an M4.6 TODO in `cycleMode`'s
own pre-existing doc comment, not a design decision made fresh here. Existing test
`TestBrowseModeSwitchResetsScopeToTheNewModesFirstValue` still passes unmodified: it switches to
a mode never yet visited this Root/session, so the fallback-to-first-value path is still what
fires; behavior for an *already-visited* mode is the new coverage in `browse_memory_test.go`
(`TestBrowseMemoryFreshSessionStartsOnTree`, `TestBrowseMemoryModePerRootScopePerRootMode`,
`TestBrowseMemoryScopePerModeSurvivesAThirdModeVisit`). `make check` and `make check-pty` both
green (full runs, no pty behavior touched by this task).

Not done / logged: TRIAGE T20 unchanged (build/medium — FilesMode/FilesScope discoverability).
TRIAGE T21 still open only for replan to confirm/clear (already fully resolved as of M4.5).
TRIAGE T22 unchanged (build/low — Tree's own "no root"/"empty root" message conflation).

Next: cursor is **M5.1** (Starred as a third Picker scope value, both File and Grep Picker,
resolving to the whole store rather than a set of Roots) — needs M1.4, M2.1 (both done).
Nothing blocks it. M4 is now fully done (M4.1–M4.6); M5 is next up.

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
| M4.1 | done   | 2026-08-12 | -       | discoverability gap logged as TRIAGE T20; review found selection-index gap, TRIAGE T23 |
| M4.2 | done   | 2026-08-12 | -       | orphan-bucket cap + live-recompute judgment calls, D79; TRIAGE T21; review: T26, T29 |
| M4.3 | done   | 2026-08-12 | -       | flat-row Activate/Star wiring added beyond listing, D80; narrows TRIAGE T21; review: T23 |
| M4.4 | done   | 2026-08-12 | -       | render test walks all six mode/scope combos; TRIAGE T22 |
| M4.5 | done   | 2026-08-12 | -       | Peek fixed via togglePeekFlatRowInto; Expand/Collapse no-ops; TRIAGE T21; review: T23, T24, T25 |
| M4.6 | done   | 2026-08-12 | -       | review found selection-index restore gap, TRIAGE T23                |
| M5.1 | todo   | -        | -         | Grep-side `Starred` is pre-agreed cut 4  |
| M6.1 | todo   | -        | -         |                                          |
| M6.2 | todo   | -        | -         | pre-agreed cut 2                         |
| M7.1 | todo   | -        | -         |                                          |
| M7.2 | todo   | -        | -         |                                          |
