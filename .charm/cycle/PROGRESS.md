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

M7.2 done (2026-08-04): README documents the `ripgrep` requirement, its ignore-rule
consequence and the Picker's keys; every cycle-level acceptance bullet in SPEC.md was
confirmed by hand against the real binary or an existing deterministic test, none raised to
TRIAGE — see Handoff for detail. This was the cycle's last planned task. Next recommended
step: a milestone/cycle-close pass (review + validate the M7.x tasks, drain TRIAGE T25, then
`charm-close`).

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M7.2 is done: README.md gained a "Finding across files (the Picker)" section (modes, Scope,
status/empty-state strings, Activate/dismiss behaviour) plus a "Requires `ripgrep`" subsection
(the requirement, the exact unavailability message, and the ignore-rule consequence: `rg`'s
hidden-file defaults mean the Picker won't search/show dotfiles the Files Nav does show), a
"While the Picker is open" keybinding table, and a `ctrl+p` row in the Global table; the stale
"Fuzzy find/grep across files ... not part of this release" line was removed from `## Scope`.
Every command/example (including pre-existing `make build`/`make run`) was checked against
`bin/noggin` directly (built fresh) or `make check-pty`, both green.

Each SPEC.md cycle-level acceptance bullet was confirmed, none raised to TRIAGE: bullets 1–6
via an ad hoc pty walkthrough (two-Roots-by-basename fixture, File find→preview→activate,
Grep switch→search→activate→land+highlight→n/N traversal, Scope switch finding a
non-active-Root file, no-rg launch + unavailability message, hidden-file allowlist
enforcement); bullet on dismiss-leaves-everything-unchanged via the existing
`TestPickerDismissRestoresThePreOpenFrameByteForByte` (stronger than pty replay: perturbs
mode/Scope/Prompt then compares `View().Content` byte-for-byte, no reliance on the terminal
emulator); bullet on Prompt responsiveness via `TestPickerSearchIsDebouncedAndShowsOnlyTheFinalQuery`
/`TestPickerDiscardsStaleResults`. Gotcha: chaining Picker actions (switch→switch→dismiss) in
an ad hoc pty script tripped the same class of `Screen`-emulator replay artifact
picker-pty-async-ordering.md documents for switch-after-search — not exercised by the
carefully-tuned committed `checks.py`, and superseded here by the stronger unit test above, so
no new finding or TRIAGE item; a future pty check chaining >2 Picker actions before an assert
should budget extra pump time or split the sequence, per that finding.

This was the cycle's last planned task. TRIAGE still carries one open, low-severity item (T25,
a Status-table wording nit) — not blocking, but due for `charm-replan` at the close boundary.
Next step: a milestone/cycle-close pass — review + validate the M7.x tasks (none Reviewed/
Validated yet), drain T25, then `charm-close`.

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
| M6.1 | done   | 2026-08-04 | 2026-08-04 | pre-agreed cut (D32); D57; TRIAGE T25 |
| M6.2 | done   | 2026-08-04 | 2026-08-04 | pre-agreed cut (D32) not taken; D57 |
| M7.1 | done   | -        | -         | pty checks for the Picker; findings/picker-pty-async-ordering.md |
| M7.1a| done   | -        | -         | TRIAGE T11, T24 |
| M7.1b| done   | -        | -         | TRIAGE T1, T2, T3, T4, T23; D56 |
| M7.2 | done   | -        | -         |      |
