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

Current task: M5.1

## Handoff

M4.2 done (2026-09-08, `make check` clean, `go test -race ./cmd/noggin/...` clean). Async
fetch-and-open (S3): new `model.fetchArticleInto(target, title, peek) (model, tea.Cmd)` /
`applyArticleFetch(articleFetchMsg) model` pair (article_fetch.go), backed by a new
`viewer.pendingFetch *pendingArticleFetch{title, peek, token}` (viewer.go) that doubles as the
in-progress indicator (`viewer.fetching()`, checked ahead of Peek in `view()`/`paneTitle`/
`renderPane`) and the staleness guard applyArticleFetch compares its token against. `token` is
`model.articleFetchSeq`, a model-wide monotonic counter — uniqueness is its only job.

"Target moved on" invalidation: `open`/`refreshDoc`/`close`/`startPeek` (viewer.go) all clear
`pendingFetch` unconditionally (a real navigation always invalidates whichever kind of fetch
was pending); `clearPeek` clears it **only when `pendingFetch.peek == true`** — a fresh-context
code-reviewer pass caught that an unconditional clear there would drop an Activate-targeted
fetch on any of `clearPeek`'s many unrelated callers (every Focus change via `setFocus`,
`refresh()`, a Files Nav selection move, a mouse-wheel scroll) — none of PLAN M4.2's own named
triggers ("Peek cleared, Picker re-opened"). Fixed, with a regression test
(`TestActivateWikipediaFetchSurvivesAnUnrelatedFocusChange`) and the full reasoning in
`pendingFetch`'s own doc comment (viewer.go) — read it before touching `clearPeek` or adding
another `pendingFetch`-clearing call site. `closeDV2` drops a DV2-targeted `pendingFetch`
implicitly (the whole viewer goes with it); still correct only because `viewerFor`'s dv2→dv1
fallback plus `articleFetchSeq`'s never-reused tokens make a false match impossible — documented
at `closeDV2`, not a bug.

Wired to real keys: `activatePickerInto` (picker.go) gained a `wikipediaMode` branch and
replaced M3.5's `f.article` "isn't supported yet" refusal with a real `fetchArticleInto` call
(D128's own "M4.2's job" landmine) — its signature is now `(model, tea.Cmd)`, propagated through
`activatePicker`/`handlePickerInput`'s two call sites. `syncPreview`'s `f.article` placeholder
was rewritten in *text only* (D131): D119 ("every network call is user-initiated") does not
list a Results-list selection move, so the Preview still never fetches. New `wikipediaFetcher`
interface (Fetch only, article_fetch.go) is a separate model field from `wikipediaClient`
(Search only) — both point at the same `*wikipedia.Client` in `newModel()` so a real session
shares one cache, but a test fake for either never has to implement the other, per
`pickerWikipediaClient`'s own doc comment (D130/D131's full reasoning for both calls).

**Peek is not wired to any real key this task** — D130: PLAN M4.2's "Peeking it" is a future
Files-Nav article row's Peek (M5.2's own job, Files Nav Recent/Starred rows do not exist yet),
not a new Picker-internal Peek (S2 already forecloses a live Preview; GLOSSARY's Peek is
Files-Nav-only). `fetchArticleInto`'s `peek == true` half is fully built and tested directly
(mirroring M4.1's own `articleDocument`-ahead-of-its-caller precedent) — **M5.2 should call
`fetchArticleInto(target, title, true)` from its Peek keypress, not reinvent this.**

New tests: `cmd/noggin/article_fetch_test.go` (Activate into DV1/DV2 + completion, Peek start,
a failure for each of Activate/Peek, the in-flight indicator, two stale-completion shapes —
re-navigate-away and Picker-reopen-clears-Peek — plus the Focus-change regression above);
`picker_scope_wikipedia_test.go`'s M3.5-era refusal test rewritten to assert the new dispatch
behaviour instead. Fresh-context code-reviewer pass: 1 high-severity functional bug (the
clearPeek scope above, fixed + regression-tested), 1 high-severity doc-comment misplacement
(fixed), 2 low (a `%q`/`%s` message-format inconsistency, fixed; `closeDV2`'s implicit-fallback
reliance, now documented) — no TRIAGE items opened, everything fixed directly (all low/medium
risk once found). D130/D131 record the two scope judgment calls.

M4.3 done (2026-09-08, `make check` clean, `go test -race ./cmd/noggin/...` clean). `R`/e on an
article (S3's remaining three clauses): `model.refresh()` now returns `(model, tea.Cmd)` —
`refreshViewer` (model.go) routes an open article to a new `refreshArticleFetchInto` (article_
fetch.go) instead of `loadDocument` (which would `os.ReadFile` a URL and fail loudly, D129's
named landmine), reusing fetchArticleInto's `pendingFetch`/token machinery unchanged — same
in-progress indicator, same "target moved on" staleness guard — but never moving Focus or
creating DV2 (a refresh only ever acts on a viewer that already has the article open).
`articleFetchMsg` gained a `refresh bool` (mutually exclusive with `peek`); `applyArticleFetch`'s
new branch calls `viewer.refreshDoc` in place (preserving scroll, same as a file) rather than
`navigate`/`startPeek` — no recency/history write, since nothing navigated. `refreshDoc` itself
now also refreshes the viewer's `articleCache` entry for the doc's URL, so a later Back/Forward
through history reopens the freshly-refetched content instead of the stale pre-refresh copy.
`fetcher.Fetch(title, true)` is the cache-bypass call (S3's own "bypasses the cache"); pinned by
a `fakeWikipediaFetcher.refreshFlags` addition. Staleness needed **no new code**: `staleness.go`'s
pre-existing "omit, don't half-track" `os.Stat`-failure rule already keeps a URL untracked (its
own C6/M4.1 code comment already named this) — M4.3's job there was pinning it with a test, not
building anything. `e` (startEdit, model.go) gains one guard: `v.doc.isArticle()` refuses with a
Footer message before the `$EDITOR` handoff, the same shape as "no document open" just above it
(not a silent no-op — Edit does not share Peek's carve-out).

Fresh-context code-reviewer pass (same day): 1 medium functional bug — `applyArticleFetch`'s
failure branch unconditionally overwrote `m.message`, silently clobbering whatever refresh()'s
own synchronous half (the *other* viewer's read failure or dropped-find message) had already set
for the same `R` press, breaking D61's "never clobber a message already set" rule right at the
sync/async boundary a mixed article+file refresh introduces — fixed with an `if m.message == ""`
guard, regression-tested
(`TestRefreshFailedArticleFetchDoesNotClobberDV2sOwnRefreshMessage`). 1 medium test gap (no
stale/superseded **refresh**-completion test existed, only Activate/Peek's) — added
(`TestStaleArticleRefreshDroppedWhenTargetNavigatesAway`). 2 low: every new test only covered
"article in DV1, file in DV2" per PLAN's literal wording — added the DV2-article/DV1-file mirror
(`TestRefreshUpdatesAFileInDV1BesideAnArticleInDV2`); GLOSSARY's "Refresh" entry updated for the
article-refetches-over-the-network case. No TRIAGE items opened — everything fixed directly.

Next: M5.1 (star/un-star an article from the focused viewer and a Files Nav row). `document.
isArticle()`/`label()` (M4.1) and the store's article-record shape (M1.5) are already in place;
M5.1's own job is wiring the Star key and rendering the marker on both surfaces — see M3.4's
Picker-star precedent for the store-toggle shape. **Peek is still not wired to any real key**
(D130 — that is M5.2's job, not M5.1's): M5.1 only needs `fetchArticleInto`'s Activate half, which
M4.2 already built and M4.3 leaves untouched.

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
| M4.1  | done   | 2026-09-08 | -         | review: code-reviewer subagent, same day |
| M4.2  | done   | 2026-09-08 | -         | review: code-reviewer subagent, same day; found 1 high-sev functional bug (clearPeek scope) + 1 high-sev doc-comment placement, both fixed directly, plus 2 low fixed directly |
| M4.3  | done   | 2026-09-08 | -       | review: code-reviewer subagent, same day; found 1 medium functional bug (m.message clobber across sync/async refresh) + 1 medium test gap (stale refresh completion) + 2 low (mirror-case test, GLOSSARY drift), all fixed/added directly |
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
