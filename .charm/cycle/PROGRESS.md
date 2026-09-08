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

Current task: M2.1

## Handoff

M1.5 done, completing M1 (Source model and Sources Nav): the document store's `docRecord`
(`docstore.go`) gained `Article bool`/`Title string` alongside the existing
`Starred`/`LastViewed`, zero-valued (false/"") for every file record exactly as before. A
file record's key stays its own absolute path; an article record is instead keyed by its own
canonical URL (M4.1's own future store-key contract), with Title carrying S1's "pointer —
canonical URL and title" second half. `loadDocStore` now accepts two line shapes: the
legacy, pre-M1.5 3-field `<starred>\t<lastViewed>\t<path>` (unchanged parsing, produces
byte-for-byte the same docRecord as before) and a new 5-field
`<starred>\t<lastViewed>\t<kind>\t<title>\t<key>` (`kind` is `file` or `article`; an article
line needs a non-empty key and title or is dropped, the same malformed-line tolerance as
any other field). `saveDocStore` always writes the new 5-field shape going forward (a file
record's `title` field simply empty) — only *loading* an old store is required to stay
unchanged, not the on-disk shape after any resave. `evictDocStore`/the count cap needed no
changes at all: both already operate generically on `Starred`/`LastViewed` across every
record in the map, so a mixed file+article store is bounded and starred-protected as one
pool automatically — covered by a new eviction test rather than requiring new production
code. New tests in `docstore_test.go`: `TestDocStoreMixedFileArticleRoundTrip`,
`TestLoadDocStoreArticleRequiresTitle`,
`TestLoadDocStoreLegacyFileFormatByteForByteEquivalent`,
`TestEvictDocStoreMixedFileArticleStore`. `go test ./...`/`make check` pass. Model-level
wiring (constructing/displaying article records) is intentionally out of scope here — M4.1
onward is where a `document`/viewer/list actually produces one.

Next: M1 is now fully complete (M1.1-M1.5 all done) — its own milestone boundary, validation
due per cycle-state before building past it (none of M1's tasks have a Validated date yet).
M2 (Wikipedia client, independent of M1) is the next unstarted milestone if validation is
deferred or run separately; read `findings/wikimedia-api.md` before M2.2/M2.3. D120 covers
the spike extraction (M2.1).

## Status

| Key   | Status | Reviewed | Validated | Note |
|-------|--------|----------|-----------|------|
| M1.1  | done   | 2026-09-07 | -       |      |
| M1.2  | done   | 2026-09-07 | -       |      |
| M1.3  | done   | -        | -         |      |
| M1.4  | done   | -        | -         |      |
| M1.5  | done   | -        | -         |      |
| M2.1  | todo   | -        | -         |      |
| M2.2  | todo   | -        | -         |      |
| M2.3  | todo   | -        | -         |      |
| M3.1  | todo   | -        | -         |      |
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
| M8.2  | todo   | -        | -         |      |
| M8.3  | todo   | -        | -         |      |
| M9.1  | todo   | -        | -         | pre-agreed cut 3 |
| M10.1 | todo   | -        | -         |      |
