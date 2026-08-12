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

Current task: M7.2 (done; cycle's last task)

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

replan (2026-08-12): all 46 plan tasks reconciled against PROGRESS's `done` rows, no drift.
TRIAGE's sole open item, T44 (low, stale "22"→"23" explicit-binding count in
`test/pty/checks.py:1028`'s M6.3 comment), fixed directly per effort-scaling (too small to
warrant a plan task); `make check-pty` reconfirmed green. TRIAGE.md now empty.

validate (2026-08-13, whole-cycle, fresh-context via three parallel validator subagents split
M1-2/M3-4/M5-7+cycle-level): every one of the 46 tasks' acceptance criteria independently
re-verified against its test(s) *and*, for the higher-risk ones (M1.1 clean-break, M1.3
eviction, M2.6 un-star-leaves-docs-untouched, M3.2 provenance, M4.2/M4.6/M4.7/M4.15 flat-row
bugs, M5.1 Grep-Starred, M6.1/M6.2 history/scroll semantics, M7.1/M7.2 badges/footer), the
implementation itself — no overclaims found. `make check` and `make check-pty` both green.
All seven of SPEC.md's cycle-level Acceptance bullets independently assessed: all satisfied.
One spec-hygiene note (not a defect): SPEC.md's Pre-agreed-cuts list is stale on 3 of 4 items
(cuts #1/#3/#4 all shipped anyway per D89/D93/M2.5, only #2 was a real cut) — a reader of
SPEC.md alone would misjudge what got cut without cross-referencing DECISIONS.md; flagged for
a future spec-hygiene pass, no cycle action taken since D89/D93 already chose not to rewrite
after the fact. No TRIAGE items raised, no DECISIONS.md entries needed.

Cycle C4 is content-complete and validated. Next step: `/charm-close`.

## Status

| Key  | Status | Reviewed | Validated | Note                                     |
|------|--------|----------|-----------|------------------------------------------|
| M1.1 | done   | 2026-08-10 | 2026-08-13 |                                       |
| M1.2 | done   | 2026-08-10 | 2026-08-13 |                                       |
| M1.3 | done   | 2026-08-10 | 2026-08-13 |                                       |
| M1.4 | done   | 2026-08-10 | 2026-08-13 |                                       |
| M2.1 | done   | 2026-08-10 | 2026-08-13 |                                      |
| M2.2 | done   | 2026-08-10 | 2026-08-13 |                                      |
| M2.3 | done   | 2026-08-10 | 2026-08-13 |                                      |
| M2.4 | done   | 2026-08-10 | 2026-08-13 |                                      |
| M2.5 | done   | 2026-08-10 | 2026-08-13 | pre-agreed cut 1, not actually cut   |
| M2.6 | done   | 2026-08-10 | 2026-08-13 |                                      |
| M1.1a | done  | 2026-08-11 | 2026-08-13 | replan: pty harness fix, TRIAGE T4 (D71) |
| M2.7 | done   | 2026-08-11 | 2026-08-13 | replan: stale marker fix, TRIAGE T7 (D70) |
| M2.8 | done   | 2026-08-11 | 2026-08-13 | replan: star-binding discoverability, TRIAGE T5/T6 |
| M2.9 | done   | 2026-08-11 | 2026-08-13 | replan: naming cleanup, TRIAGE T1/T2/T9  |
| M3.1 | done   | 2026-08-11 | 2026-08-13 | SPEC/GLOSSARY-level gap re: Picker Activate not recording lastViewed — tracked as TRIAGE T12, not a M3.1 overclaim |
| M3.2 | done   | 2026-08-10 | 2026-08-13 |                                          |
| M3.3 | done   | 2026-08-10 | 2026-08-13 | replan: Picker never records lastViewed, TRIAGE T12 (D74) |
| M3.4 | done   | 2026-08-10 | 2026-08-13 | replan: recordViewed failure message clobber, TRIAGE T13; no bug found, test coverage added; review found coverage gap re: Picker call sites, resolved as M3.5 (D77) |
| M3.5 | done   | 2026-08-10 | 2026-08-13 | replan: Picker-path store-write-failure regression tests, TRIAGE T18 (D77) |
| M2.10 | done  | 2026-08-10 | 2026-08-13 | replan: pty star-marker column checks, TRIAGE T11        |
| M2.11 | done  | 2026-08-10 | 2026-08-13 | replan: Picker star marker (D72, TRIAGE T8), built same session; review found stale doc comment, resolved as M2.12 (D76); validate found Grep-mode live-toggle coverage gap, resolved as M2.13 (D78) |
| M2.12 | done  | 2026-08-10 | 2026-08-13 | replan: correct starMarker() doc comment, TRIAGE T17 (D76) |
| M2.13 | done  | 2026-08-10 | 2026-08-13 | replan: Grep same-frame star-toggle render test, TRIAGE T19 (D78) |
| M4.1 | done   | 2026-08-12 | 2026-08-13 | discoverability gap logged as TRIAGE T20; review found selection-index gap, TRIAGE T23 |
| M4.2 | done   | 2026-08-12 | 2026-08-13 | orphan-bucket cap + live-recompute judgment calls, D79; TRIAGE T21; review: T26, T29 |
| M4.3 | done   | 2026-08-12 | 2026-08-13 | flat-row Activate/Star wiring added beyond listing, D80; narrows TRIAGE T21; review: T23; pre-agreed cut 3, not actually cut (D93, cf. M2.5/M5.1) |
| M4.4 | done   | 2026-08-12 | 2026-08-13 | render test walks all six mode/scope combos; TRIAGE T22 |
| M4.5 | done   | 2026-08-12 | 2026-08-13 | Peek fixed via togglePeekFlatRowInto; Expand/Collapse no-ops; TRIAGE T21; review: T23, T24, T25 |
| M4.6 | done   | 2026-08-12 | 2026-08-13 | review found selection-index restore gap, TRIAGE T23                |
| M4.7 | done   | 2026-08-12 | 2026-08-13 | replan: flat-row selection re-clamp bug, TRIAGE T23/T24 (D82); drops T21 as already-handled; review: T37 |
| M4.8 | done   | 2026-08-12 | 2026-08-13 | replan: Browse cycle keys' footer/help discoverability, TRIAGE T20 (D81); review: T32/T33 |
| M4.9 | done   | 2026-08-12 | 2026-08-13 | replan: Tree's own empty-vs-no-Root message, TRIAGE T22; review: T34  |
| M4.10 | done  | 2026-08-12 | 2026-08-13 | replan: bundled test-coverage gaps, TRIAGE T25/T26/T29/T31 (D83); review: T36 |
| M4.11 | done  | 2026-08-12 | 2026-08-13 | replan: comment-literal and citation fixes, TRIAGE T27/T28; review: T35 |
| M4.12 | done  | 2026-08-11 | 2026-08-13 | replan: M4.8 test-coverage gaps, TRIAGE T32/T33 (D86)    |
| M4.13 | done  | 2026-08-11 | 2026-08-13 | replan: flatEmptyText default branch + tie-break comment, TRIAGE T34/T36 (D87); review: T39 |
| M4.14 | done  | 2026-08-11 | 2026-08-13 | replan: M4.7 acceptance-wording literal test gap, TRIAGE T37 (D88) |
| M5.1 | done   | 2026-08-11 | 2026-08-13         | delivered both File- and Grep-side Starred (D89, not actually cut, cf. M2.5); review found grepFiles dead code, TRIAGE T42 |
| M6.1 | done   | 2026-08-11 | 2026-08-13         |                                          |
| M6.2 | done   | 2026-08-11 | 2026-08-13         | pre-agreed cut 2, not actually cut (D91); review found untested re-Activate scroll-preserve behavior, TRIAGE T43 |
| M4.15 | done  | 2026-08-11 | 2026-08-13         | replan: openFolder flat-mode selection-reset bug, TRIAGE T38 |
| M5.2 | done   | 2026-08-11 | 2026-08-13         | replan: delete dead grepFiles, TRIAGE T42 (D94) |
| M6.3 | done   | 2026-08-11 | 2026-08-13         | replan: HistoryBack/Forward Footer/Help discoverability, TRIAGE T41; check-pty M5.3 viewer-block height widened 36→38; review: T44 |
| M6.4 | done   | 2026-08-11 | 2026-08-13         | replan: D91 re-Activate scroll-preserve regression test, TRIAGE T43; confirmed no bug, coverage gap only |
| M7.1 | done   | 2026-08-11 | 2026-08-13         |                                        |
| M7.2 | done   | 2026-08-11 | 2026-08-13         |                                        |
