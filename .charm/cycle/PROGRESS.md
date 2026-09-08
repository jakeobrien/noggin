# Progress — C6 sources

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

Current task: M3.2

## Handoff

M1 and M2 (Source model + Wikipedia client, M1.1–M1.8 and M2.1–M2.8, including the eight
replan-added cleanup tasks from TRIAGE T1–T8/T10–T12/T15/T17) are all `done` as of
2026-09-08; M1.8/M2.8 are not yet Reviewed/Validated (see the Status table). No reopen
needed for M1.1–M2.7 (D125).

M1.8 (TRIAGE T17): `escapeDocTitle`/`unescapeDocTitle` renamed to `escapeDocField`/
`unescapeDocField` and now also applied to the C6/M1.5 five-field shape's `<key>` field
(both `file` and `article` records) on save/load; the legacy three-field shape's key stays
raw/unescaped so existing on-disk stores load unchanged. New test
`TestDocStoreKeyTabAndNewlineRoundTrip` in `cmd/noggin/docstore_test.go` covers a file key
and an article key each carrying an embedded tab or newline.

M2.8 (TRIAGE T15): `internal/wikipedia.NewClient` now builds `&Client{}` then calls
`ensureInit()` instead of constructing `httpClient`/`cache` directly — one init path shared
with the lazy zero-value case. M2.4's zero-value and `-race` concurrency tests
(`TestFetchAtZeroValueClientDoesNotPanic`, `TestFetchAtConcurrentUseIsRaceFree`) still pass.

DECISIONS D122 (fetch cache's narrower-than-documented guarantee — revisit at M4.2), D123
(equivalence-claim verifiability) and D124 (Search/Fetch config asymmetry — revisit at M3.3)
are open by design, not blocking.

SPEC.md amended 2026-09-08 (`/brainstorm amend`, user-confirmed): new S9a closes D98 (CC
BY-SA attribution, open since C4a) — a Wikipedia article's location line now reads `CC BY-SA
4.0 · <URL> · Reformatted to markdown` instead of a plain URL. M8.2 (still `todo`, not yet
built) is updated to build the three-segment line; no other task's scope or status changes.
D98 is now `decided`.

M3.1 (S13): `pickerQueryMinLength` dropped 3→2 (`pickerBelowThresholdStatus` now "type at
least 2 characters"); `queuePickerSearch`'s debounce is now per-mode via `pickerDebounceFor`
— File/Grep keep `pickerSearchDebounce` (120ms), a new `wikipediaMode` pickerMode value (not
yet reachable from `switchPickerMode`; that wiring is M3.2's) gets its own, longer
`pickerWikipediaSearchDebounce` (300ms). New/updated tests:
`TestPickerWikipediaModeDebounceExceedsFileModeDebounce` (debounce ordering, direct on
`pickerDebounceFor`), `TestPickerIssuesNoSearchBelowTheThreshold` (1 char below, 2 chars
searches) and `TestPickerStatusStates`/pty `checks.py` (threshold string). TRIAGE T18 logs 11
pre-existing `make check-pty` failures (verified identical with these changes stashed out, so
not caused by M3.1); `make check` is clean. TRIAGE otherwise empty; both T1/T2 were
independent of M3 and pulled forward per the plan's own note.

Next: M3.2 (Wikipedia mode joins the File→Grep→Wikipedia cycle, needs M1.1) — read `SPEC.md`
S2 and `PLAN.md`'s M3 section first. No further milestone-boundary gate needed before M3.2.

## Status

| Key   | Status | Reviewed | Validated | Note |
|-------|--------|----------|-----------|------|
| M1.1  | done   | 2026-09-08 | 2026-09-08 |      |
| M1.2  | done   | 2026-09-08 | 2026-09-08 |      |
| M1.3  | done   | 2026-09-08 | 2026-09-08 |      |
| M1.4  | done   | 2026-09-08 | 2026-09-08 |      |
| M1.5  | done   | 2026-09-08 | 2026-09-08 |      |
| M2.1  | done   | 2026-09-08 | 2026-09-08 |      |
| M2.2  | done   | 2026-09-08 | 2026-09-08 |      |
| M2.3  | done   | 2026-09-08 | 2026-09-08 |      |
| M1.6  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T1, T2) |
| M1.7  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T3) |
| M2.4  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T4, T5) |
| M2.5  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T6, T7, T8) |
| M2.6  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T10) |
| M2.7  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T11, T12) |
| M1.8  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T17); review found 1 low-sev finding (TRIAGE T2), independently reproduced and PASSed at validate |
| M2.8  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T15); review found 1 low-sev finding (TRIAGE T1), independently reproduced and PASSed at validate |
| M3.1  | done   | -        | -         |      |
| M3.2  | todo   | -        | -         |      |
| M3.3  | todo   | -        | -         |      |
| M3.4  | todo   | -        | -         |      |
| M3.5  | todo   | -        | -         |      |
| M4.1  | todo   | -        | -         |      |
| M4.2  | todo   | -        | -         |      |
| M4.3  | todo   | -        | -         |      |
| M5.1  | todo   | -        | -         |      |
| M5.2  | todo   | -        | -         |      |
| M6.1  | todo   | -        | -         | pre-agreed cut 1 |
| M7.1  | todo   | -        | -         | pre-agreed cut 2 |
| M7.2  | todo   | -        | -         | pre-agreed cut 2 |
| M7.3  | todo   | -        | -         | pre-agreed cut 2 |
| M8.1  | todo   | -        | -         |      |
| M8.2  | todo   | -        | -         | scope widened by replan 2026-09-08: S9a attribution line |
| M8.3  | todo   | -        | -         |      |
| M9.1  | todo   | -        | -         | pre-agreed cut 3 |
| M10.1 | todo   | -        | -         |      |
