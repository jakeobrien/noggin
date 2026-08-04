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

Build (2026-08-05): M7.1c done — fixed the tautological pty Preview-highlight check (T26/D58).
Every M7.x task is now `done`; none has been Validated yet. Next recommended step: a
milestone/cycle-close validation pass over M7.1/M7.1a/M7.1b/M7.2/M7.1c, then `charm-close`.

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M7.1c (T26/D58): fixed test/pty/checks.py's tautological "Preview highlights the matched
line" check. Root cause was worse than D58's write-up assumed: the literal `\x1b[7m` search
didn't just coincidentally pass off the selected Results row's own reverse video
(pickerSelectedStyle) — the genuine Preview highlight never even produced that exact substring
on the wire, because lipgloss combines the reverse attribute with the adjacent fg-color reset
into one SGR (`\x1b[39;7m`), which doesn't contain `\x1b[7m` as a substring. The check now uses
a regex matching SGR param 7 anywhere in a semicolon list, requiring it to immediately precede
the Preview's untruncated rendering of the matched line ("# Guide") — the Results row only ever
shows that line truncated with an ellipsis ("# Gui…"), so only a genuine Preview highlight can
satisfy it. Also dropped the redundant `out or s.out` (same object, per D58).

Verified with a negative control: temporarily stripped the Preview's StyleLineFunc reverse
styling in picker.go, rebuilt, confirmed this specific check (only this one) failed under
`make check-pty`, then reverted and confirmed all checks pass clean again. Final diff is
test/pty/checks.py only — no product code changed.

Every M7.x task (M7.1, M7.1a, M7.1b, M7.2, M7.1c) is now `done`; none has been Validated yet.
Next step: a validate pass over all five, then `charm-close`. No blockers, no open TRIAGE.

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
| M7.1 | done   | 2026-08-04 | -         | pty checks for the Picker; findings/picker-pty-async-ordering.md; D58 (TRIAGE T26 -> M7.1c) |
| M7.1a| done   | 2026-08-04 | -         | TRIAGE T11, T24 |
| M7.1b| done   | 2026-08-04 | -         | TRIAGE T1, T2, T3, T4, T23; D56 |
| M7.2 | done   | 2026-08-04 | -         |      |
| M7.1c| done   | -        | -         | TRIAGE T26; D58 |
