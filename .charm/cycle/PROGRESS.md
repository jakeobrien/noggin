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

M7.1b done (2026-08-04): all five validation-pass test-hygiene sub-items (TRIAGE T1–T4, T23 /
D56) addressed — see Handoff for detail. Only M7.2 remains todo in this milestone; it needs
M7.1, already satisfied. Next recommended step: M7.2 (docs + cycle-level acceptance
walkthrough), the last task in the cycle.

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M7.1b (validation-pass test hygiene backfill) is done, all five sub-items in
cmd/noggin/picker_test.go, picker_view_test.go and files_search_test.go:
(a) `TestPickerActivateGrepResultOpensDocumentScrolledToMatchedLine` and
`TestPickerActivateGrepResultLandsOnTheSelectedOccurrenceNotJustTheFirst` (T1) now cross-
reference each other's doc comments as a pair covering D34's ordinal-resolution bridge; (b)
new `TestPickerActivateGrepResultDoesNotChangeFilesNav` (T2) covers S10 for the grep Activate
path, mirroring M3.3's file-mode `TestPickerActivateOpensResultAsActiveDocumentOnly`; (c)
`TestPickerFileResultsDisambiguateSameBasenameRoots`'s doc comment (T3) now explains its
rowA/rowB check is over whole frame lines, not isolated pickerFileRow output, and
cross-references `TestPickerFileRowFillsItsFullBudget` as the test that actually isolates it;
(d) that same M3.3 fixture (T4) now expands a real subdirectory so its
`reflect.DeepEqual(m.files.expanded, ...)` assertion exercises a non-empty map (mirrored into
the new (b) test too); (e) `TestGrepFilesBatchedResultsExactlyAtCapAreNotTruncated` (T23/D56)
got the one-line forward-looking-guard comment D56 called for. `make check` (build/vet/fmt/
test) is green; all touched/added tests individually verified passing.

Next step: M7.2 (docs + cycle-level acceptance walkthrough) — the cycle's last planned task.
TRIAGE still carries one open, low-severity item (T25, a Status-table wording nit) — not
blocking, but due for `charm-replan` to drain once the cycle reaches a milestone/close
boundary.

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
| M7.2 | todo   | -        | -         |      |
