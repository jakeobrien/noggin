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

Current task: M2.4

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

M2.2 done: `internal/wikipedia/search.go` adds `Search(query, limit)` against Wikimedia's
title-search endpoint (`searchBase`), returning `[]SearchResult{Title, Description}` in
Wikipedia's own order; `Description` is empty when the endpoint sends `null`. Errors are two
distinguishable exported types per the plan's done-when: `*ServiceError` (non-2xx response —
StatusCode/Status) and `*OfflineError` (the request never reached Wikimedia — wraps the
underlying net error), so a caller can tell offline from a service failure without string-
matching. `userAgent()` is a new package-level helper (mirrors `cmd/wikispike/fetch.go`'s but
reads `NOGGIN_WIKIPEDIA_CONTACT`, not `WIKISPIKE_CONTACT` — noggin's own env var, not the
spike's) satisfying S8. `searchTitles(base, query, limit)` takes the base URL as a parameter
so `search_test.go` points it at `httptest` servers instead of the network; `Search` is the
thin production wrapper pinning `searchBase`. Five httptest-backed tests cover: results with
and without a description, an empty result list, a 500 service error (asserted to *not* also
match `*OfflineError`), a connection failure via a closed-then-hit `httptest` server
(asserted to *not* also match `*ServiceError`), and the `User-Agent` header value. Verified:
`make check` passes; `go test ./internal/wikipedia/... -run TestSearch -v` shows all five
new tests passing individually.

M2.3 done: `internal/wikipedia/fetch.go` adds `Client` (`NewClient()`, `Fetch(title, refresh)
(Article, error)`), completing M2's Wikipedia client. `Fetch` GETs `fetchBase` + escaped
title; Go's default `http.Client` follows the endpoint's 307 automatically, and
`canonicalFromURL` reads the canonical title/URL back out of the final `resp.Request.URL`
(handles both the redirect-landing path `.../page/<Title>/html` and an already-canonical
request's own `.../page/<Title>` path). `Article{Title, URL, Markdown}` is the result after
running `ArticleMarkdownFromHTML`. Errors reuse M2.2's `*ServiceError`/`*OfflineError` split
rather than adding new types: a 404 ("article gone") and a client-timeout both surface
naturally through the same two categories `http.Client.Do` already distinguishes for search.
Caching design (a judgment call, logged here rather than DECISIONS.md since it is
implementation shape, not a product question): the cache is keyed by the canonical URL
computed *before* the request, from the requested title — so it only pays off when the
requested title is already canonical, which S3 guarantees for every title noggin itself
records (Starred/Recent/Refresh all reopen by canonical title). A title that turns out to be
a redirect always costs one request; there is no way to know its canonical URL without
asking. `refresh=true` bypasses the lookup and overwrites the cache entry. `Client` holds its
cache behind a mutex (safe for concurrent use, matching the async-fetch shape M4.x will need).
Five httptest-backed tests (`fetch_test.go`, exercising the unexported `fetchAt` with
overridable base URLs) cover: a redirect resolving to the canonical title/URL, a 404, a
timeout (via a client with a 5ms timeout against a handler that sleeps 50ms), a cache hit
(second fetch of the same title makes no second request), and a refresh bypass (second
request made, cache repopulated, a following non-refresh fetch returns the new value).
Verified: `make check` and `go test ./... -race` both pass; `go test ./internal/wikipedia/...
-run TestFetchAt -v` shows all five new tests passing individually.

M2 (Wikipedia client) is now fully done — M2.1/M2.2/M2.3 all `done`, and (contrary to this
handoff's stale prior wording) already `Reviewed`/`Validated` 2026-09-08 per the Status
table; that review pass is what surfaced TRIAGE T4–T13.

2026-09-08 replan: triaged all 13 open TRIAGE items (T1–T13). Accepted as new PLAN tasks:
M1.6/M1.7 (T1–T3, "Roots Nav"→"Sources Nav" rename propagation and docstore title
tab/newline guard) and M2.4–M2.7 (T4–T8, T10–T12, Wikipedia-client hardening and cleanup) —
all six are queued cleanup work, low-effort and independent of the M3 critical path, pullable
forward like M3.1/M8.1 whenever a short slice is wanted. T9 (Search/Fetch API-shape
asymmetry) became DECISIONS.md D124 (open, to resolve by/at M3.3). T13 was dropped as a
duplicate of the already-open D122. TRIAGE.md is now empty. See D124/D125.

M1.6 done: mechanically propagated GLOSSARY's C6 rename ("Roots Nav" → "Sources Nav")
across every comment and string in `cmd/noggin` — 24 files touched, including three
line-wrapped occurrences (`keymap.go`, `nav.go`, `main.go`) a plain substring sed/perl pass
missed because the phrase split across a line break in the source comment. The rendered pane
title itself (`model.go`'s `pane.title()` for `rootsPane`) changed from the literal `"Roots"`
to `"Sources"`; `focus_badge_test.go`'s two width-arithmetic tests were re-derived for the
new title's length (7 chars vs. 5: interior-column and rect-width constants shifted from
9/11/10 to 11/13/12) rather than just renaming the string, since the pane border only fits a
title+badge for panes wide enough. `TestPaneFocusBadgesShowAtNormalWidthRegardlessOfFocus`
(uses `pane.title()` dynamically) is the "rootsnav render test asserts the pane label reads
'Sources'" the acceptance criterion calls for. Root/Roots-the-data-concept references (`All
Roots` scope label, `rootsPane`/`rootsnav.go` identifiers, `FocusRoots`, `loadStarredRoots`,
etc.) were deliberately left alone — GLOSSARY still uses "Root"/"Roots" for that concept;
only the Roots-*Nav* pane name moved. A pre-existing GLOSSARY self-inconsistency found along
the way (Navigator entry at GLOSSARY.md:77 still says "Roots Nav above" despite its own
rename note at lines 114-120) is out of this task's `cmd/noggin` scope — logged as TRIAGE T14
for replan rather than fixed here. Verified: `grep -rn "Roots Nav" cmd/noggin` and `grep -n
'"Roots"' cmd/noggin/*.go` both empty; `make check` and `go test ./... -count=1` pass; a
fresh-context code-reviewer subagent independently re-derived the width arithmetic and
confirmed no leftover renames or grammar breakage — PASS on both review passes.

M1.7 done: `docstore.go` guards the article `<title>` field against the tab and newline bytes
that would otherwise corrupt the store's tab-separated, one-record-per-line format — a raw
tab would shift `loadDocStore`'s fixed `SplitN(line, "\t", 5)` field positions (bleeding the
title's tail into the `<key>` field), and a raw newline would split one record's line into
two, producing an unparsable orphan. Fix: two new helpers, `escapeDocTitle` (backslash, tab,
newline → `\\`, `\t`, `\n`, backslash escaped first so the encoding is unambiguous to
reverse) called from `saveDocStore`'s title write, and `unescapeDocTitle` (its inverse,
tolerant of a malformed trailing lone backslash rather than panicking) called from
`loadDocStore`'s 5-field-shape title read; the legacy 3-field shape has no title field and is
untouched. `TestDocStoreArticleTitleTabAndNewlineRoundTrip` covers a tab title, a newline
title and an adjacent plain-title record all in the same store file, asserting both the
load(save(x))==x round-trip *and* the on-disk bytes directly (exact line count, exactly 4 tab
boundaries per line, the literal escaped substrings present) so a subtly-broken line/field
count that happened to still parse could not slip past unnoticed — the on-disk assertion was
added after a fresh-context code-reviewer subagent flagged the round-trip-only version as a
minor test-thoroughness gap. Verified: `make check` and `go test ./cmd/noggin/... -run
TestDocStore -v` pass; the reviewer's two passes (standards, spec fidelity) both PASS with no
remaining findings.

Next: M2.4–M2.7 (Wikipedia-client hardening/cleanup, TRIAGE T4–T8, T10–T12) — queued triage
cleanup, pulled forward ahead of M3.1 at the user's request. After those, M3.1 (the first
Wikipedia Picker task): read `SPEC.md` S2 and the M3 section of `PLAN.md` before starting. M2
already cleared its review/validate gate, so no further milestone-boundary check is needed
before M3.

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
| M1.6  | done   | -        | -         | added by replan 2026-09-08 (TRIAGE T1, T2) |
| M1.7  | done   | -        | -         | added by replan 2026-09-08 (TRIAGE T3) |
| M2.4  | todo   | -        | -         | added by replan 2026-09-08 (TRIAGE T4, T5) |
| M2.5  | todo   | -        | -         | added by replan 2026-09-08 (TRIAGE T6, T7, T8) |
| M2.6  | todo   | -        | -         | added by replan 2026-09-08 (TRIAGE T10) |
| M2.7  | todo   | -        | -         | added by replan 2026-09-08 (TRIAGE T11, T12) |
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
