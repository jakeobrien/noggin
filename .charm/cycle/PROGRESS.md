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

Replan (2026-08-04): fresh-context code review covered M7.1/M7.1a/M7.1b/M7.2 (all now
Reviewed) and raised TRIAGE T25–T28. Triaged this session: T25 actioned directly (Status-table
wording fix, below); T26 promoted to new task **M7.1c** (D58); T27 moot (the Handoff paragraph
it cited is being replaced by this one); T28 dropped (matches an existing local convention,
low risk/value, no decision warranted — see TRIAGE.md's routing note). TRIAGE inbox is empty
again. Next recommended step: M7.1c (small, opportunistic, no dependency), then a
milestone/cycle-close pass (validate the M7.x tasks — Validated is still `-` on all of them —
then `charm-close`).

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

This session was a replan, not a build: no code changed. A fresh-context code review (maker/
checker split) covered the diff since M6 for M7.1, M7.1a, M7.1b and M7.2 — all four are
Reviewed 2026-08-04, Validated still `-`. Findings (TRIAGE T25–T28) have been triaged and the
inbox is empty; the substantive one is **T26/D58**: M7.1's pty check "the Preview highlights
the matched line (reverse video on the wire)" (test/pty/checks.py:907-911) is tautological —
it only checks that `\x1b[7m` appears anywhere in the frame, which the selected Results row's
own reverse-video styling already guarantees regardless of whether the Preview highlights
anything. M7.1 stays `done` (its literal acceptance criterion doesn't require this specific
check to be non-tautological, and the real behavior is independently covered by a unit test —
see D58), but the fix is now **M7.1c**, a small opportunistic follow-up.

Also fixed directly (T25, too small to warrant a plan task): M6.1's Status-table note read
"pre-agreed cut (D32); D57" with no qualifier, contradicting the Handoff's own (correct)
statement that D32's cut was not taken for M6 — now reads "pre-agreed cut (D32) not taken;
D57", matching M6.2's row.

Next step: M7.1c (fix the tautological pty check, effort 1, no dependency) — the only `todo`
task left in the plan. After that, every M7.x task needs a validate pass (Validated is `-` on
all of M7.1/M7.1a/M7.1b/M7.2/M7.1c) before `charm-close`.

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
| M7.1c| todo   | -        | -         | TRIAGE T26; D58 |
