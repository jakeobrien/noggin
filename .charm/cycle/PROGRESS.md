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

No task in progress. M6 is complete (M6.1, M6.2 done, neither reviewed nor validated); M7.1,
M7.1a, M7.1b, M7.2 remain todo. Next recommended step: a fresh-context `code-review` +
`charm-validate` pass over the M6 rows (and M4.2a, whose Reviewed cell is still `-`) before
building M7.

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M6 (S5 match highlighting) is done — D32's pre-agreed cut was *not* taken. Shape, both rows:
match positions are computed as rune indexes by the search layer (`fileNameMatchPositions` for
fuzzy file-name matches, `grepMatchPositions` for literal smart-case line matches, both in
files_search.go, both converting byte offsets via the shared `runeIndexByByte`), and
`highlightMatches` (picker_view.go) bolds the given *columns* of an assembled row.
`pickerFileRow`/`pickerGrepRow` each gained a `query` parameter; existing layout tests pass `""`
and keep asserting plain text. `grepMatchPositions` reuses find.go's `foldLower`/`origSpan` so
S8's smart-case rule and its fold-length translation have one implementation, not two. D57
records why positions are recomputed rather than carried through `pickerResults`, and why the
highlight is bold derived from the row's own base style.

Gotchas for anyone touching row rendering: (1) lipgloss closes an inner style with a **full**
reset (`ESC[m`), so wrapping a partly-styled line in `pickerSelectedStyle` kills the selection's
reverse video from the first highlight on — `highlightMatches` therefore renders *every* run,
matched and unmatched, with its style fully specified. (2) Grep columns are offset by the
`rel:line  ` prefix and clipped (`clipColumns`) to the truncated row's visible cells, ellipsis
excluded. (3) `cellStyles`/`boldColumns` (picker_view_test.go) are the test-side SGR walker for
asserting *where* styling landed; `plainLines` throws that away.

All five new/changed tests were mutation-checked and each failed as intended: wrong suffix
offset, nested styling (selection bar), unclipped truncation columns, searching the whole row
instead of the line text, and smart case disabled. `make check` and `make check-pty` green.

Next step: M6's two rows have never been reviewed or validated (and M4.2a's Reviewed cell is
still `-`) — a fresh-context `code-review` plus `charm-validate` over those is the gate before
M7. Then M7.1 (pty checks for the Picker), with M7.1a/M7.1b as cheap opportunistic backfill.

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
| M4.2a| done   | -        | 2026-08-04 | D55 (TRIAGE T21, T22) |
| M4.3 | done   | 2026-08-03 | 2026-08-03 | M4.3 test coupling note, see handoff |
| M4.4 | done   | 2026-08-04 | 2026-08-03 |      |
| M5.1 | done   | 2026-08-04 | 2026-08-04 |      |
| M5.2 | done   | 2026-08-04 | 2026-08-04 |      |
| M5.3 | done   | 2026-08-04 | 2026-08-04 |      |
| M6.1 | done   | -        | -         | pre-agreed cut (D32); D57 |
| M6.2 | done   | -        | -         | pre-agreed cut (D32) not taken; D57 |
| M7.1 | todo   | -        | -         |      |
| M7.1a| todo   | -        | -         | TRIAGE T11, T24 |
| M7.1b| todo   | -        | -         | TRIAGE T1, T2, T3, T4, T23; D56 |
| M7.2 | todo   | -        | -         |      |
