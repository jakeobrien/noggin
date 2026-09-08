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

Current task: M2.2

## Handoff

M1 (M1.1-M1.5) validated 2026-09-08 via a fresh-context subagent — all five tasks PASS, no
discrepancies; see the Status table. Pre-existing low-severity TRIAGE items T1-T3 ("Roots
Nav" naming drift, an unescaped-tab edge case in docstore's title field) noted but not
blocking; still open, not yet triaged by replan.

M2.1 done: the spike's fetch-agnostic conversion pipeline (math, images, chrome strip,
links, convert, cleanup — `cmd/wikispike/{math,images,strip,links,convert,cleanup,
htmldoc}.go` and their tests) moved verbatim (only `package main` → `package wikipedia`)
into new `internal/wikipedia/`, importable by `cmd/noggin` per D120. The pipeline's
combinator (formerly `main.articleMarkdownFromHTML`) is now the package's one exported
entry point, `wikipedia.ArticleMarkdownFromHTML(html) (string, error)`. `cmd/wikispike`
keeps its own exact-title fetch (`fetch.go`) — that stays spike-only; M2.3 adds the real
canonical fetch cmd/noggin will use, in the same internal package. `cmd/wikispike/main.go`
is now a thin wrapper: fetch by exact title, then call `wikipedia.ArticleMarkdownFromHTML`.
The moved live tests (opt-in, `WIKISPIKE_LIVE=1`) needed their own tiny fetch fixture
(`internal/wikipedia/live_test.go`) since a `main` package can't be imported by a test in
another package — duplicates ~40 lines of exact-title-fetch shape from `fetch.go`,
deliberately, as temporary test-only plumbing pending M2.3's real fetch landing in the same
package. `sampleTitles` is likewise duplicated (one small var) rather than shared across
packages. Verified: `go build/vet/test ./...` and `make check` all pass; the opt-in live
suite passes in both `cmd/wikispike` and `internal/wikipedia`; regenerating "Major Arcana"
through `wikispike` before and after the move produced byte-identical markdown (16003
bytes, `diff` empty).

Next: M2.2 (title search) — read `findings/wikimedia-api.md` first (already read this
session); build the title-search endpoint client in `internal/wikipedia`, httptest-backed,
no network in `go test ./...`. M2.3 (canonical fetch) can follow in either order per the
plan's ordering note but M2.2 is the cursor.

## Status

| Key   | Status | Reviewed | Validated | Note |
|-------|--------|----------|-----------|------|
| M1.1  | done   | 2026-09-08 | 2026-09-08 |      |
| M1.2  | done   | 2026-09-08 | 2026-09-08 |      |
| M1.3  | done   | 2026-09-08 | 2026-09-08 |      |
| M1.4  | done   | 2026-09-08 | 2026-09-08 |      |
| M1.5  | done   | 2026-09-08 | 2026-09-08 |      |
| M2.1  | done   | -        | -         |      |
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
