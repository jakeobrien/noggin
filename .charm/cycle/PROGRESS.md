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

Current task: M6.1

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M5.1 built: `pickerScope` gained a third value (`starredScope`), a shared axis both File
and Grep Picker modes read (`switchPickerScope` now cycles Active Root -> All Roots ->
Starred -> Active Root; mode switches never touch it). Candidate resolution is new
(`starredPickerFiles`, files_search.go): the whole S1 store's starred documents, orphans
included, dropping missing files (unlike Browse's own broken-star display) and still
honouring the Readable allowlist and hideDotfiles. `grepFiles` was split into a thin
roots-resolving wrapper plus `grepPickerFiles` (batching/cap logic, takes a file list
directly) so Starred scope's Grep search reuses it without going through
`liveRootsInScope`'s roots-shaped contract. `pickerFile` gained an `orphan bool` field;
row rendering (`pickerLocationLabel`, picker_view.go) falls back to the document's own
parent directory for an orphan, mirroring browse_list.go's existing orphan treatment
(pickerGrepRow needed no change — its existing filepath.Rel error-fallback already does
the right thing for an empty root path).

Built both File- *and* Grep-side Starred scope despite PLAN.md/SPEC.md naming Grep-side a
pre-agreed cut (#4): the shared-scope/shared-candidate-set design made Grep search under
Starred work for free, and M5.1's own acceptance wording already asked for both literally.
Logged as D89 (same "not actually cut" class as M2.5/TRIAGE T40 — SPEC.md/PLAN.md wording
now stale in the same way, left for a future doc-wording pass rather than edited here).

All new: files_search_test.go (starredPickerFiles unit tests), picker_view_test.go (orphan
row rendering, both modes), picker_scope_starred_test.go (scope-cycle, mode-switch-
preserves-scope, and full File/Grep search-under-Starred acceptance tests). Updated a
stale pty check (test/pty/checks.py) that assumed Scope was still a two-way toggle. `make
check` and `make check-pty` both green. Next: M6.1 (per-viewer document history) is
unblocked to start next session — no gate issues (TRIAGE still holds only the same three
low-severity open items from before this session, none blocking).

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
| M4.1 | done   | 2026-08-12 | 2026-08-11 | discoverability gap logged as TRIAGE T20; review found selection-index gap, TRIAGE T23 |
| M4.2 | done   | 2026-08-12 | 2026-08-11 | orphan-bucket cap + live-recompute judgment calls, D79; TRIAGE T21; review: T26, T29 |
| M4.3 | done   | 2026-08-12 | 2026-08-11 | flat-row Activate/Star wiring added beyond listing, D80; narrows TRIAGE T21; review: T23; validate found stale pre-agreed-cut-3 note, TRIAGE T40 |
| M4.4 | done   | 2026-08-12 | 2026-08-11 | render test walks all six mode/scope combos; TRIAGE T22 |
| M4.5 | done   | 2026-08-12 | 2026-08-11 | Peek fixed via togglePeekFlatRowInto; Expand/Collapse no-ops; TRIAGE T21; review: T23, T24, T25 |
| M4.6 | done   | 2026-08-12 | 2026-08-11 | review found selection-index restore gap, TRIAGE T23                |
| M4.7 | done   | 2026-08-12 | 2026-08-11 | replan: flat-row selection re-clamp bug, TRIAGE T23/T24 (D82); drops T21 as already-handled; review: T37 |
| M4.8 | done   | 2026-08-12 | 2026-08-11 | replan: Browse cycle keys' footer/help discoverability, TRIAGE T20 (D81); review: T32/T33 |
| M4.9 | done   | 2026-08-12 | 2026-08-11 | replan: Tree's own empty-vs-no-Root message, TRIAGE T22; review: T34  |
| M4.10 | done  | 2026-08-12 | 2026-08-11 | replan: bundled test-coverage gaps, TRIAGE T25/T26/T29/T31 (D83); review: T36 |
| M4.11 | done  | 2026-08-12 | 2026-08-11 | replan: comment-literal and citation fixes, TRIAGE T27/T28; review: T35 |
| M4.12 | done  | 2026-08-11 | 2026-08-11 | replan: M4.8 test-coverage gaps, TRIAGE T32/T33 (D86)    |
| M4.13 | done  | 2026-08-11 | 2026-08-11 | replan: flatEmptyText default branch + tie-break comment, TRIAGE T34/T36 (D87); review: T39 |
| M4.14 | done  | 2026-08-11 | 2026-08-11 | replan: M4.7 acceptance-wording literal test gap, TRIAGE T37 (D88) |
| M5.1 | done   | -        | -         | delivered both File- and Grep-side Starred (D89, not actually cut, cf. M2.5) |
| M6.1 | todo   | -        | -         |                                          |
| M6.2 | todo   | -        | -         | pre-agreed cut 2                         |
| M7.1 | todo   | -        | -         |                                          |
| M7.2 | todo   | -        | -         |                                          |
