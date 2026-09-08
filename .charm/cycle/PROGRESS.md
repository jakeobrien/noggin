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

Current task: M4.1

## Handoff

M3 (Wikipedia Picker) is validated (2026-09-08, fresh-context `validator` subagent, per
charm-validate) — the milestone-boundary gate is cleared, M4 may start.

**Validation result**: `make check` clean; `make check-pty` reproduces TRIAGE T18's
pre-existing 11-failure baseline (same test names, not a new regression). All five
acceptance criteria (M3.1–M3.5) and the SPEC deliverables they deliver (S2, S6, S13)
independently verified against code + tests, not just PROGRESS.md's claim — all PASS, high
confidence. No new discrepancies beyond the already-open TRIAGE T21/T22 (both re-confirmed
accurately characterized, still low-severity, not gate-blocking). Manual/interactive
confirmation of the live-TUI Wikipedia flow is still recommended (see the validator's report
for exact steps) since all evidence so far is test-level, not pty-exercised — no pty suite
covers Wikipedia mode yet.

S6's four rules, each with its own test in new `cmd/noggin/picker_scope_wikipedia_test.go`:

1. With Wikipedia the Active Source, File/Grep never land on/cycle to `activeRootScope` (no
   Active Root exists to search) — `switchPickerMode` bumps a stale `activeRootScope` to
   `allRootsScope` when landing on File/Grep from `wikipediaMode`; `switchPickerScope` cycles
   only `allRootsScope ↔ starredScope` while `m.roots.activeIsWikipedia()`. Unconditional on
   Active Source (S6's other three rules apply regardless of which Source is active):
2. File mode's Starred scope includes a starred article, matched on **title** —
   `starredPickerFiles` (files_search.go) now also emits `pickerFile{article: true, title:
   rec.Title}` pseudo-candidates for `docRecord.Article` records, skipping every file-specific
   filter (pathMissing/Readable/hideDotfiles — none describe a URL); `fileNameSource.String`
   and `pickerFileRow`/`pickerLocationLabel` (picker_view.go) branch on `f.article` to rank/
   render/highlight against `f.title` instead of a path-derived name, and to label the row
   `wikipediaLabel` ("Wikipedia") instead of a Root name.
3. File's All Roots scope is files only — already true unchanged (`listFiles` never reads
   `m.docs`); a test locks it in now that Starred's own candidate set can contain articles.
4. Grep excludes articles in every scope — new `filesOnly` (files_search.go) filters
   `starredPickerFiles`' article pseudo-candidates out before `pickerSearchCmd`'s grepMode
   branch ever calls `grepPickerFiles`, rather than relying on rg's own handling of a bogus
   URL argument.

Found and fixed on M3.5's own critical path, surfaced by fresh-context code review
(code-reviewer subagent) as a high-severity gap: M3.5 is what first makes an article
pseudo-candidate *selectable* in File mode (Starred scope), but `activatePickerInto` and
`syncPreview` (picker.go) were not taught about `f.article` — Activate raised a raw
`"cannot open https://...: no such file or directory"` Footer message and selecting the row
showed the same raw OS error in the live Preview. Both now guard on `f.article` explicitly:
Activate refuses with "`<title>`: opening a Wikipedia article from here isn't supported yet"
and dismisses (the same "message + dismiss" shape a genuine read-error Activate already has);
Preview shows a fixed placeholder instead of attempting the read. Real fetch-and-render from
this path is still M4.2's job (D128) — **M4.2 needs to confirm its acceptance criteria cover
Activate from File mode's Starred scope, not only `wikipediaMode`'s own Results list**, and
replace both guards with the real behaviour then, not before.

`docs[filepath.Clean(f.path)]` in `pickerResultsContent`'s fileMode star-marker lookup
(picker_view.go) had the same URL-corruption risk M3.4/D127 already fixed in model.go — fixed
here too, now that a fileMode Result can carry an article's URL as `f.path`. Un-starring an
article from File mode's Starred scope already works correctly as a side effect (routes
through the same `pickerSelectedPath`/`docRecord`/`withDocRecord` path M3.4 fixed) — not one
of S6's four named rules, so not separately tested, but noted since it satisfies part of S4's
"every place a file can" ahead of M5.1 building it deliberately.

Also added (low-severity review findings, fixed directly rather than logged): direct unit
tests for `starredPickerFiles`'/`fileNameSource`'s article branches
(`files_search_test.go`), and an end-to-end rendered-row assertion for the "labelled
Wikipedia" claim (previously only checked via a direct `pickerLocationLabel` call).
`make check` clean. TRIAGE T20 (GLOSSARY's `Result` entry) is unchanged, still open — File
Picker Results can now also be article-shaped, which only strengthens that item's case.

Next: M4.1 (`document` carries source identity — file path or article URL+title — for the
viewer title row/Document History/Find; needs M1.5).

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
| M3.1  | done   | 2026-09-08 | 2026-09-08 |      |
| M3.2  | done   | 2026-09-08 | 2026-09-08 |      |
| M3.3  | done   | 2026-09-08 | 2026-09-08 |      |
| M3.4  | done   | 2026-09-08 | 2026-09-08 |      |
| M3.5  | done   | 2026-09-08 | 2026-09-08 | review found 2 low-sev findings (TRIAGE T21, T22), independently reconfirmed at validate, still open/low-sev |
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
