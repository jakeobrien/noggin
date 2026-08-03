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

Current task: M4.1

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M3.2a done (this session, D48/TRIAGE T6): `dismissPicker` (picker.go) now carries the Preview
viewport's Width/Height over its reset the same way `openPicker` already does, instead of
letting `newPicker()`'s fresh 0x0 zero value stick until the next `tea.WindowSizeMsg`. New
`TestPickerReopenAfterDismissKeepsPreviewSize` (picker_test.go) drives open → dismiss → reopen
with no resize in between and asserts the Preview's Width/Height are unchanged; verified it
fails without the fix (0x0) and passes with it. `make check` green. No in-flight changes.
M3 (File Picker end to end) is now fully done including its sub-tasks.

Next step: M4.1 — wire Grep mode to M1.4 (file, line number, matching line per row, truncated
to the pane). The preview-size carry-over pattern (openPicker/dismissPicker/setPreviewSize) is
now consistent on both sides, so M4.2 (preview scrolled to the matched line) can build on it
without inheriting the bug. Before M4.3/M4.4, re-read D34: rg's raw-file line numbers don't
correspond to the Document Viewer's rendered line space, so the hand-off re-derives the match
by occurrence ordinal rather than trusting rg's line number, with a documented degrade path.

Also queued, non-blocking — pick up opportunistically or before M7 hardening: M3.3a/M3.3b
(D49, activatePicker test-coverage gaps), M7.1a (T11, terminology housekeeping).

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
| M1.4a| done   | -        | -         | D43 (TRIAGE T1) |
| M1.4b| done   | -        | -         | D45 (TRIAGE T3) |
| M2.3a| done   | -        | 2026-08-03 | D46 (TRIAGE T5); M2.3 re-validated |
| M3.1 | done   | 2026-08-03 | -         |      |
| M3.1a| done   | -        | -         | D50 (TRIAGE T10) |
| M3.2 | done   | 2026-08-03 | -         |      |
| M3.2a| done   | -        | -         | D48 (TRIAGE T6); fixed before M4.2 |
| M3.2b| done   | -        | -         | TRIAGE T12 |
| M3.3 | done   | 2026-08-03 | -         |      |
| M3.3a| todo   | -        | -         | D49 (TRIAGE T7, T8) |
| M3.3b| todo   | -        | -         | D49 (TRIAGE T9) |
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
