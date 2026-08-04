# Progress — C2 find

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

M4.4 done this session: confirmed the pre-build assumption that M4.3's `activateGrepPicker` ->
`commitQuery` hand-off already leaves the Document Viewer in a real committed-find state
(`commitQuery`'s doc comment says as much) — no production code changed. Added
`TestPickerActivateGrepResultCommitsFindWithNextPreviousTraversal` (picker_test.go): a
3-occurrence fixture, activates the first Grep Result, asserts `findState.input.Value()`
equals the grep query and `len(findState.matches) == 3`, then drives `keys.FindNext`/
`keys.FindPrevious` via `press()` (matching `TestFindTraversalWraps`'s typed-search shape) and
asserts `current` advances/returns correctly. Sanity-sabotaged `findNext` (made it a no-op) to
confirm the test actually fails rather than passing vacuously — it did, on both the next and
the previous assertion — then reverted (find.go diff is clean; verified with `git diff --stat`).
`make check` and `make check-pty` both green.

Backfill validation this session, per charm-validate, of every `done` task that had never been
Validated: M4 (previous pass) plus M1.4a/b/c, M2.1a, and the whole of M3 (M3.1, M3.1a, M3.1b,
M3.2, M3.2a, M3.2b, M3.3, M3.3a, M3.3b) — two fresh-context validator subagents, run
sequentially (not parallel) since both do mutation-and-revert probes against overlapping
files (picker.go, picker_test.go). `make check` green both times. All 13 tasks: verdict PASS
(functionally meet their PLAN.md acceptance criteria and SPEC deliverables); Validated column
stamped 2026-08-03 for all of them. Every done task in the cycle now has a Validated date
except M6.x (not started) and M5.x/M7.x (not yet built).

Four non-blocking findings, all confirmed via mutation-and-revert (broke the behavior, watched
a test fail or fail-to-fail, reverted — git tree confirmed clean throughout both sessions):
- T1: M4.3's headline test alone doesn't isolate D34's ordinal-resolution bridge, only covered
  paired with a second named test.
- T2: no direct test asserts Files Nav is untouched on the grep activation path specifically
  (only file-mode); code path is shared so the property holds structurally.
- T3: M3.1a's disambiguation test compares whole rendered frame lines, not the picker row in
  isolation — M3.1b's row-budget test is what actually isolates the regression it's meant to
  catch.
- T4: M3.3b's expansion-equality fixture never produces a non-empty `expanded` map, so the
  DeepEqual assertion it added can't demonstrate catching a same-cardinality substitution as
  intended (mechanism is sound, fixture doesn't reach it).
All four are test-quality/coverage gaps, not functional bugs; none required flipping a status.
Also noted directly in PROGRESS's M3.3a row (no TRIAGE item needed, just documentation): its
"Grep-mode no-op" test was correctly superseded/removed by M4.3 (897ef6a) once grep Activate
became real — criterion-as-written is no longer checkable, not a regression.

Next step: M5.1 — switch mode (File ↔ Grep) inside the open Picker, carrying the typed Prompt
over and re-running the search in the new mode. Likely needs a mode-switch keybinding wired
into picker.go's key handling (check keymap.go for what, if anything, already reserves one) plus
a query-rerun call reusing the existing debounced-search machinery (`typePickerQuery`/
`deliverPickerTicks`'s production counterpart) rather than typing a fresh keystroke.

Also queued, non-blocking: M7.1a (TRIAGE T11 already triaged/routed there, removed from inbox).

No blockers. TRIAGE.md has 4 open items (T1-T4), all low-severity test-coverage observations
from this session's validation passes, awaiting charm-replan triage.

## Status

| Key  | Status | Reviewed | Validated | Note |
|------|--------|----------|-----------|------|
| M1.1 | done   | 2026-08-03 | 2026-08-03 |  |
| M1.2 | done   | 2026-08-03 | 2026-08-03 |  |
| M1.3 | done   | 2026-08-03 | 2026-08-03 |  |
| M1.4 | done   | 2026-08-03 | 2026-08-03 |  |
| M2.1 | done   | 2026-08-03 | 2026-08-03 |      |
| M2.2 | done   | 2026-08-03 | 2026-08-03 |      |
| M2.3 | done   | 2026-08-03 | 2026-08-03 | M2.3a landed; re-validated |
| M2.4 | done   | 2026-08-03 | 2026-08-03 |      |
| M2.5 | done   | 2026-08-03 | 2026-08-03 |      |
| M1.4a| done   | 2026-08-03 | 2026-08-03 | D43 (TRIAGE T1) |
| M1.4b| done   | 2026-08-03 | 2026-08-03 | D45 (TRIAGE T3) |
| M1.4c| done   | -        | 2026-08-03 | D-none; TRIAGE T18, T19 |
| M2.1a| done   | -        | 2026-08-03 | D51; TRIAGE T14 |
| M2.3a| done   | 2026-08-03 | 2026-08-03 | D46 (TRIAGE T5); M2.3 re-validated |
| M3.1 | done   | 2026-08-03 | 2026-08-03 |      |
| M3.1a| done   | 2026-08-03 | 2026-08-03 | D50 (TRIAGE T10); re-reviewed, T13/T16 raised; TRIAGE T3 (disambiguation test compares full frame lines, not isolated row) |
| M3.1b| done   | -        | 2026-08-03 | D52; TRIAGE T13, T15, T16, T17 |
| M3.2 | done   | 2026-08-03 | 2026-08-03 |      |
| M3.2a| done   | 2026-08-03 | 2026-08-03 | D48 (TRIAGE T6); fixed before M4.2 |
| M3.2b| done   | 2026-08-03 | 2026-08-03 | TRIAGE T12; re-reviewed, T20 raised |
| M3.3 | done   | 2026-08-03 | 2026-08-03 |      |
| M3.3a| done   | 2026-08-03 | 2026-08-03 | D49 (TRIAGE T7, T8); grep-mode no-op test superseded/removed by M4.3 (897ef6a) once grep Activate became real — not a regression, criterion-as-written no longer checkable |
| M3.3b| done   | 2026-08-03 | 2026-08-03 | D49 (TRIAGE T9); TRIAGE T4 (fixture's expanded map stays empty, doesn't exercise the substitution scenario) |
| M4.1 | done   | -        | 2026-08-03 | D53  |
| M4.2 | done   | -        | 2026-08-03 | D54  |
| M4.3 | done   | 2026-08-03 | 2026-08-03 | M4.3 test coupling note, see handoff |
| M4.4 | done   | -        | 2026-08-03 |      |
| M5.1 | todo   | -        | -         |      |
| M5.2 | todo   | -        | -         |      |
| M5.3 | todo   | -        | -         |      |
| M6.1 | todo   | -        | -         | pre-agreed cut (D32) |
| M6.2 | todo   | -        | -         | pre-agreed cut (D32) |
| M7.1 | todo   | -        | -         |      |
| M7.1a| todo   | -        | -         | TRIAGE T11 |
| M7.2 | todo   | -        | -         |      |
