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

Current task: M2.7

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

M2.4 done: `internal/wikipedia/fetch.go`'s `Client` is now usable from its zero value
(`Client{}`, not just via `NewClient()`) — a new `ensureInit()` method, guarded by the
existing `c.mu` mutex, lazily allocates `httpClient` and `cache` on first use if either is
nil, called as `fetchAt`'s first statement (before either field is touched, on every call
path including the public `Fetch` wrapper). Without it, a bare `Client{}`'s first fetch
panicked twice over: a nil `*http.Client` field dereferenced inside `(*http.Client).Do`, and
(had that not panicked first) a nil-map write populating the cache after a successful fetch.
The whole check-and-init runs inside one held lock, so concurrent first-use callers can't
both observe nil and both allocate — no double-checked-locking pattern needed.
`TestFetchAtZeroValueClientDoesNotPanic` covers the zero-value path end-to-end (traced to
confirm it would genuinely have panicked pre-fix, not vacuously passed), and
`TestFetchAtConcurrentUseIsRaceFree` drives 20 goroutines' first-ever calls into a shared
zero-value `*Client` concurrently under `go test -race`, backing the `Client` doc comment's
pre-existing "safe for concurrent use" claim with an actual test (M2.4's other acceptance
clause). Verified: `make check` passes; `go test ./internal/wikipedia/... -race -run
TestFetchAt -v` and `-race -count=5` (5 repeats) both clean, no races, no flakes; a
fresh-context code-reviewer subagent traced the lock ordering, the happens-before chain and
the pre-fix panic paths independently — PASS on both review passes, no findings.

M2.5 done: three independent `internal/wikipedia` hygiene fixes. TRIAGE T6:
`fetchAt`/`fetchCanonicalHTML`/`canonicalFromURL`'s `fetchBase`/`wikiBase` parameters are now
`fetchBaseURL`/`wikiBaseURL`, no longer shadowing the package-level constants of the same
name. TRIAGE T7: `canonicalFromURL` used to decide "was this a redirect" by checking whether
the final URL's last path segment was literally "html" — chose to guard rather than just
document, since a real fix was cheap: it now compares the final response URL's path against
the *originally requested* URL's path (threaded through as a new parameter), an exact
structural signal instead of a text guess. Investigating this closely (with a fresh-context
code-reviewer subagent) surfaced that the old heuristic could never actually produce a wrong
answer for any real title, given `fetchBase`'s own fixed `.../page/html/` shape (verified by
running the new edge-case test against the pre-fix code, where it also passed) — the fix's
real value is making correctness structural rather than a coincidence of `fetchBase`'s
current shape, and the doc comment/test comment were reworded after that finding to say so
honestly rather than claim a reproducible bug. A second new test
(`TestFetchAtRedirectToSingleSegmentPathIsRejected`) covers the new `len(segments) < 2`
guard branch (a malformed/unexpected redirect shape errors cleanly rather than panicking).
TRIAGE T8: `ServiceError`/`OfflineError`'s doc comments in `search.go` now note both
`search.go` and `fetch.go` construct them, not just search. Verified: `make check` and `go
test ./internal/wikipedia/... -race -count=3` all pass, no flakes; two fresh-context
code-reviewer passes (the second specifically re-checking the first pass's framing finding)
both end PASS with no remaining findings.

M2.6 done: `internal/wikipedia/live_test.go`'s `fetchLiveArticleHTML` no longer keeps its own
copy of the exact-title-fetch HTTP logic (request-building, status-check, body-read) — it
now calls the package's own unexported `fetchCanonicalHTML` (fetch.go, M2.3's real fetch),
discarding the canonical title/URL and keeping only the raw HTML. Deliberately not
`Client.Fetch`/`fetchAt`: both convert to markdown before returning, but `convert_test.go`,
`images_test.go` and `strip_test.go` all need *raw* HTML from `fetchLiveSample` to test
individual pipeline stages, not the converted end product — confirmed each would break
(wrong input type, or a substring check that would silently always fail) if fed markdown
instead. `fetchLiveSample`'s caller-facing shape and caching needed no changes.
`liveRestBase` and the manually-set User-Agent string are gone (fetchCanonicalHTML sets its
own, from the production `userAgent()`); the stale "M2.3 adds the real canonical fetch..."
comment (written when M2.3 was still future work) is reworded to past tense now that it has
landed. Verified: `make check` passes; both opt-in live suites pass with
`WIKISPIKE_LIVE=1` — `internal/wikipedia`'s five live test functions (one pre-existing,
unrelated skip) and `cmd/wikispike`'s own; a fresh-context code-reviewer subagent confirmed
no duplicated logic remains, no stale/unused imports, and the raw-vs-markdown call-site
choice is correct by tracing all three raw-HTML consumers.

Next: M2.7 (Wikipedia-client nits, TRIAGE T11, T12) — the last queued triage cleanup, pulled
forward ahead of M3.1 at the user's request. After it, M3.1 (the first Wikipedia Picker
task): read `SPEC.md` S2 and the M3 section of `PLAN.md` before starting. M2 already cleared
its review/validate gate, so no further milestone-boundary check is needed before M3.

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
| M2.4  | done   | -        | -         | added by replan 2026-09-08 (TRIAGE T4, T5) |
| M2.5  | done   | -        | -         | added by replan 2026-09-08 (TRIAGE T6, T7, T8) |
| M2.6  | done   | -        | -         | added by replan 2026-09-08 (TRIAGE T10) |
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
