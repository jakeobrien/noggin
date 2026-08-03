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

Current task: M3.1b

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M2.1a done this session (D51): extracted resetPicker (cmd/noggin/picker.go), the shared
"bump seq, snapshot preview size, replace with newPicker(), restore seq, restore preview
size" sequence openPicker and dismissPicker used to duplicate verbatim. newPicker() is now
called only from inside resetPicker; both callers reduce to a bare m.resetPicker() call (plus
openPicker's open=true/Focus()). M2.1's and M3.2a's existing acceptance tests
(TestPickerOpensFromEveryFocusAndNavigatorHidden, TestPickerAlwaysOpensAsFileActiveRootEmptyPrompt,
TestPickerDismissRestoresThePreOpenFrameByteForByte, TestPickerReopenAfterDismissKeepsPreviewSize)
pass unchanged. `make check` green.

M3 (File Picker end to end) and all its original sub-tasks are done; M1.4a/b/c, M2.1a, M2.3a,
M3.1a/M3.2a/M3.2b/M3.3a/M3.3b (the prior review's fix-ups) are also done.

Next step: M3.1b (D52, effort 2) — pickerRootLabel's budget calc double-subtracts a gap,
making M3.1a/D50's disambiguation fix less effective than designed at realistic (80x24)
widths (T13, T15, T16, T17). Then M4.1: wire Grep mode to M1.4. M4.2 reuses M3.2's preview
mechanism (now behind resetPicker, so a grep-preview reset path added there goes through the
same carry-over automatically); before M4.3/M4.4, re-read D34 (rg line numbers vs. rendered
line space).

Also queued, non-blocking: M7.1a (T11, terminology housekeeping).

No blockers. TRIAGE.md empty.

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
| M1.4a| done   | 2026-08-03 | -         | D43 (TRIAGE T1) |
| M1.4b| done   | 2026-08-03 | -         | D45 (TRIAGE T3) |
| M1.4c| done   | -        | -         | D-none; TRIAGE T18, T19 |
| M2.1a| done   | -        | -         | D51; TRIAGE T14 |
| M2.3a| done   | 2026-08-03 | 2026-08-03 | D46 (TRIAGE T5); M2.3 re-validated |
| M3.1 | done   | 2026-08-03 | -         |      |
| M3.1a| done   | 2026-08-03 | -         | D50 (TRIAGE T10); re-reviewed, T13/T16 raised |
| M3.1b| todo   | -        | -         | D52; TRIAGE T13, T15, T16 |
| M3.2 | done   | 2026-08-03 | -         |      |
| M3.2a| done   | 2026-08-03 | -         | D48 (TRIAGE T6); fixed before M4.2 |
| M3.2b| done   | 2026-08-03 | -         | TRIAGE T12; re-reviewed, T20 raised |
| M3.3 | done   | 2026-08-03 | -         |      |
| M3.3a| done   | 2026-08-03 | -         | D49 (TRIAGE T7, T8) |
| M3.3b| done   | 2026-08-03 | -         | D49 (TRIAGE T9) |
| M4.1 | todo   | -        | -         |      |
| M4.2 | todo   | -        | -         |      |
| M4.3 | todo   | -        | -         |      |
| M4.4 | todo   | -        | -         |      |
| M5.1 | todo   | -        | -         |      |
| M5.2 | todo   | -        | -         |      |
| M5.3 | todo   | -        | -         |      |
| M6.1 | todo   | -        | -         | pre-agreed cut (D32) |
| M6.2 | todo   | -        | -         | pre-agreed cut (D32) |
| M7.1 | todo   | -        | -         |      |
| M7.1a| todo   | -        | -         | TRIAGE T11 |
| M7.2 | todo   | -        | -         |      |
