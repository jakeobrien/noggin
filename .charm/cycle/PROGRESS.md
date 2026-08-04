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

Validate (2026-08-05): fresh-context validation pass over M7.1/M7.1a/M7.1b/M7.1c/M7.2 —
all five PASS, no discrepancies requiring TRIAGE or a PLAN change. Every task in the cycle is
now `done` and `Validated`. Next recommended step: `charm-close`.

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

Validate pass (fresh subagent context) over the five unvalidated M7.x tasks, deterministic
first: `make check` and `make check-pty` both clean (build/vet/fmt/test + full pty suite,
including M7.1's Picker checks and M7.1c's fix). Then per-task verification against code/tests
(not PROGRESS's word): M7.1 (all required flows present incl. before/after-dismissal frame
identity, test/pty/checks.py:815-926), M7.1a (zero comments use "Results region"/"Results
column"; two `t.Error` message strings still do, at picker_view_test.go:41,52 — criterion says
"comment", so this is not a failure, just worth knowing), M7.1b (all five sub-items a-e located
and substantive), M7.1c (independently reproduced: stripped Preview's Reverse StyleLineFunc,
`make check-pty` failed exactly the one Preview-highlight check, reverted, suite clean again;
`out or s.out` confirmed gone), M7.2 (README documents ripgrep requirement + ignore-rule
consequence + Picker keys, spot-checked 3 items against code, all accurate; TRIAGE.md empty;
spot-checked 2 of SPEC.md's cycle-level acceptance bullets independently true). Working tree
left clean (git status/diff empty) after M7.1c's negative-control edit was reverted.

No blockers, no open TRIAGE, no DECISIONS entries needed — nothing actionable surfaced. Next
step: `charm-close`.

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
| M1.4c| done   | 2026-08-04 | 2026-08-03 | D-none; TRIAGE T18, T19; TRIAGE T23 |
| M2.1a| done   | 2026-08-04 | 2026-08-03 | D51; TRIAGE T14 |
| M2.3a| done   | 2026-08-03 | 2026-08-03 | D46 (TRIAGE T5); M2.3 re-validated |
| M3.1 | done   | 2026-08-03 | 2026-08-03 |      |
| M3.1a| done   | 2026-08-03 | 2026-08-03 | D50 (TRIAGE T10); re-reviewed, T13/T16 raised; TRIAGE T3 (disambiguation test compares full frame lines, not isolated row) |
| M3.1b| done   | 2026-08-04 | 2026-08-03 | D52; TRIAGE T13, T15, T16, T17 |
| M3.2 | done   | 2026-08-03 | 2026-08-03 |      |
| M3.2a| done   | 2026-08-03 | 2026-08-03 | D48 (TRIAGE T6); fixed before M4.2 |
| M3.2b| done   | 2026-08-03 | 2026-08-03 | TRIAGE T12; re-reviewed, T20 raised |
| M3.3 | done   | 2026-08-03 | 2026-08-03 |      |
| M3.3a| done   | 2026-08-03 | 2026-08-03 | D49 (TRIAGE T7, T8); grep-mode no-op test superseded/removed by M4.3 (897ef6a) once grep Activate became real — not a regression, criterion-as-written no longer checkable |
| M3.3b| done   | 2026-08-03 | 2026-08-03 | D49 (TRIAGE T9); TRIAGE T4 (fixture's expanded map stays empty, doesn't exercise the substitution scenario) |
| M4.1 | done   | 2026-08-04 | 2026-08-03 | D53; TRIAGE T21, T22, T24 |
| M4.2 | done   | 2026-08-04 | 2026-08-03 | D54; TRIAGE T24 |
| M4.2a| done   | 2026-08-04 | 2026-08-04 | D55 (TRIAGE T21, T22) |
| M4.3 | done   | 2026-08-03 | 2026-08-03 | M4.3 test coupling note, see handoff |
| M4.4 | done   | 2026-08-04 | 2026-08-03 |      |
| M5.1 | done   | 2026-08-04 | 2026-08-04 |      |
| M5.2 | done   | 2026-08-04 | 2026-08-04 |      |
| M5.3 | done   | 2026-08-04 | 2026-08-04 |      |
| M6.1 | done   | 2026-08-04 | 2026-08-04 | pre-agreed cut (D32) not taken; D57 |
| M6.2 | done   | 2026-08-04 | 2026-08-04 | pre-agreed cut (D32) not taken; D57 |
| M7.1 | done   | 2026-08-04 | 2026-08-05 | pty checks for the Picker; findings/picker-pty-async-ordering.md; D58 (TRIAGE T26 -> M7.1c) |
| M7.1a| done   | 2026-08-04 | 2026-08-05 | TRIAGE T11, T24; two t.Error strings still say "Results region" (picker_view_test.go:41,52) — criterion is comment-scoped, not a failure |
| M7.1b| done   | 2026-08-04 | 2026-08-05 | TRIAGE T1, T2, T3, T4, T23; D56 |
| M7.2 | done   | 2026-08-04 | 2026-08-05 |      |
| M7.1c| done   | -        | 2026-08-05 | TRIAGE T26; D58 |
