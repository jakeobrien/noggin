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

Current task: M3.4

## Handoff

M3.3 (S2) is done: Wikipedia mode's search is real now. `model.wikipediaClient`
(`pickerWikipediaClient` interface, `Search(query, limit)` only) defaults to a real
`*wikipedia.Client` in `newModel`; tests inject a fake via `withWikipediaClient`.
`pickerSearchCmd`'s `wikipediaMode` branch calls it and reports back via `pickerResultsMsg`'s
new `err` field. Rows render via `pickerWikipediaRow` (title + description, no match
highlighting, no star marker — that's M3.4). `searchable()`/`pickerStatus` no longer require
ripgrep for `wikipediaMode` (only File/Grep do).

Failure handling (S2 "leaves the last Results in place"): `picker.wikipediaErr` holds the
last search's error text; `applyPickerResults`'s `err` branch relabels only
`results.query/mode/scope/hideDotfiles` (so `resultsCurrent` still holds) and leaves the
actual rows untouched rather than replacing them with the failed attempt's empty ones.
`pickerStatus` shows `wikipediaErr` ahead of the generic searching/no-results captions;
`queuePickerSearch` clears it on every new attempt (keystroke or mode/Scope switch) — same
"goes stale the instant something changes" rule the Results labels already have.

Preview: new `pickerPreviewContent(p picker)` helper (picker_view.go) returns
`pickerWikipediaPreviewPlaceholder` for `wikipediaMode`, `p.preview.View()` otherwise;
`renderPicker` now calls it instead of reading `p.preview.View()` directly.

D124 (Search/Fetch config asymmetry) is now `decided`, not open: `*wikipedia.Client.Search`
is a one-line delegator to the existing free `Search`/`searchHTTPClient` — the Picker's
injectable-client need only required *a* `Search` method, not a shared `http.Client`/config
with Fetch. Revisit only if M4.2 surfaces a concrete shared-config need.

New `cmd/noggin/picker_wikipedia_search_test.go` (`fakeWikipediaSearcher`) covers results
rendering (title+description), stale-generation discard, failure-keeps-rows, the error
clearing on the next attempt, the Preview placeholder, and no-ripgrep-needed. `make check`
clean; `make check-pty` unchanged at the same pre-existing 11 failures (TRIAGE T18).

Next: M3.4 (Picker star key toggles a Wikipedia Result's Document Store record, shows the
marker — needs M1.5, M3.3). Gotcha to check first: `togglePickerStar`/`pickerSelectedPath`
today only handle `fileMode`/`grepMode` (a Wikipedia Result silently no-ops, which is why
M3.3 needed no change there) — M3.4 needs its own candidate lookup, and a title-search
`wikipedia.SearchResult` carries no canonical URL (only `Fetch`, M4.2, resolves that), so
M3.4/S4's "the store's key is the canonical URL" (M4.1) needs a concrete answer for starring
a Result that has never been fetched — read S4/S6 and M1.5's article-record shape before
assuming title is an adequate key. No milestone-boundary gate before M3.4.

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
| M3.2  | done   | -        | -         |      |
| M3.3  | done   | -        | -         |      |
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
