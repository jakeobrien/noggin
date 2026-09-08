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

Current task: M2.9

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

M5.1 done (2026-09-08, `make check` clean). Star/un-star an article from the focused Document
Viewer and from a Files Nav row: the viewer-title-row marker (`paneTitle`, M2.4) and the
Files-Nav-row toggle mechanism (`toggleDocStar`'s flat-mode branch, M2.1/M4.3) were already
generic over `docRecord`'s article-aware key (`docStoreKey`, M1.5) and needed no change — but a
probe test surfaced a real gap in `toggleActiveDocStar` (the viewer's own toggle, M2.2, which
predates M1.5's article records): it left a never-before-starred article's `Article`/`Title`
fields at `docRecord`'s zero value, so `saveDocStore` wrote it as an invalid `file`-kind line
(a URL fails `filepath.IsAbs`) that `loadDocStore` silently dropped on the next load — the star
looked like it worked for the rest of the session but did not survive a restart. Fixed by
stamping `Article`/`Title` unconditionally on every toggle, mirroring `togglePickerStar`'s own
Wikipedia branch (M3.4) — see D133 for the full reasoning, including why `toggleDocStar`'s flat
branch needs no equivalent fix yet (a flat row can only ever show an *already*-starred article
today; Recent mode omits every article until M6.1 gives it an article-aware existence check in
place of `pathMissing`'s plain `os.Stat`, which always fails on a URL).

New tests: `TestToggleActiveDocStarStarsAndPersistsAnArticle` (viewer_star_test.go, the
persistence round-trip the bug above broke), `TestDocumentViewerTitleShowsStarIndicatorFor
AnArticle` (viewer_star_render_test.go, the title-row `★` marker), and
`TestToggleDocStarInStarredModeDropsAStarredArticleRowImmediately` (browse_starred_test.go, an
already-starred article's Files Nav row, reached via Starred/All Starred scope since an
article's bucket key is always `""` — the same orphan bucket a Root-orphan file lands in).

M5.2 done (2026-09-08, `make check` clean, `go test -race ./cmd/noggin/...` clean). Files Nav
Starred lists now include articles for real, fixing the cosmetic gap M5.1's own handoff named:
a new `buildStarredBrowseRow` (browse_list.go) gives a starred article its own display shape
(name = the article's title, rootLabel = the fixed `wikipediaLabel`, never `buildBrowseRow`'s
file-shaped `filepath.Base(url)`/`pathMissing`-driven `(missing)`) instead of falling through
to the generic file path M4.3 shipped. `starredRows` buckets every starred article under its
own reserved key (`wikipediaSourceKey`, files.go's existing per-Source memory sentinel, reused
here for the same "impossible as a real Root path" property) rather than the "" orphan bucket
a Root-orphan file also keys on — the actual bug the old bucketing had: without this, Active
Source while Wikipedia is active and Orphans while a Root is active would both silently mix
articles in with unrelated Root-orphan files. `starredRows` gained a `wikipedia bool` parameter
(files.go passes `f.wikipedia`) so Active Source can tell "no active Root" apart from
"Wikipedia is the Active Source" — both read `activeRootPath == ""` otherwise. Orphans and All
Starred needed no scope-selection change: the new bucket key already keeps articles out of the
former and lets them fall naturally into the latter's every-bucket merge.

Activate and Peek on a Starred/Recent-flat-row article (S3) now fetch instead of failing
`loadDocument` on a URL (D129's own landmine, M4.3's Refresh precedent): `activateInto`/
`activate`/`togglePeek`/`togglePeekInto` all widen their return to `(model, tea.Cmd)` so their
new `row.article` branches (`activateFlatRowInto`, `togglePeekFlatRowInto`) can dispatch
`fetchArticleInto(target, row.name, peek)` — the same mechanism M4.2 built and M5.2's own
handoff flagged as "call this, don't reinvent it". `browseRow` gained an `article bool` field
(mirrors `pickerFile`'s own article/title pair, files_search.go) that both new branches read;
the already-Peeking/already-Active carve-out in `togglePeekFlatRowInto` needed no article-
specific case since `row.path` (an article's canonical URL) already equals `v.doc.path` once
open, the same identity a file's row already relies on.

New/updated tests: `TestStarredRowsWikipediaActiveSourceShowsArticlesByTitle`,
`TestStarredRowsAllStarredIncludesArticlesLabelledWikipediaWithARootActive`,
`TestStarredRowsOrphansExcludesArticles` (browse_starred_test.go, the bucketing fix at the
`starredRows` level); `TestActivateStarredArticleFlatRowFetchesAndOpens`,
`TestPeekStarredArticleFlatRowFetchesWithoutActivating` (same file, the Activate/Peek fetch
wiring, mirroring article_fetch_test.go's own Picker-Result shape via a shared
`starredArticleWikipediaModel` fixture). M5.1's own `TestToggleDocStarInStarredModeDropsA
StarredArticleRowImmediately` was updated in place (title/`Wikipedia` label assertions instead
of the old `Leonhard_Euler` URL-slug ones it had pinned as an accepted gap) — not a new test,
the same test now exercising the fixed rendering; widened to a 120-column fixture since 80
truncates `Wikipedia` (9 cols) below its own render budget, the same class of gotcha C2/D52/
TRIAGE T13 already named for the Picker's Root label. Every pre-existing `starredRows` call
site (tests and files.go) took the new `wikipedia bool` parameter; every pre-existing
`activateInto`/`togglePeekInto` test call site was updated to discard the new `tea.Cmd` return.
No TRIAGE items opened, no DECISIONS entry — the design (bucket-key reuse, return-signature
widening) follows precedent M4.2/M3.4/M3.5 already set closely enough to be low/medium risk,
not an open judgment call.

M1.9 done (2026-09-08, `make check` clean, `make check-pty` clean — 0 failures, was 11).
Fixed the 9 assertions in `test/pty/checks.py` D132 diagnosed as a real regression, not
flakiness: `"Roots"`/`focused() == "Roots"`/`footer.split(...) == "Roots"`/
`footer.startswith("Roots")` -> `"Sources"` (M1.6's pane-label rename) across M1.2 (both
sizes, one shared loop body so one fix covers the reported 80x24/100x30 duplicate),
M3.2, M3.3, M3.4, M5.2 and C2a/M2.1; `"File \u00b7 Active Root"` -> `"File \u00b7 Active
Source"` (M1.3's File-Picker scope-label rename; Grep's own "Active Root" wording is
unchanged per M1.3's acceptance criterion and was left alone) in M7.1 and C2a/M6.2. No
source changes — pty-harness-only, confirming D132's diagnosis that the renames themselves
were already correct and only the cross-cycle regression suite had drifted. Pre-existing
uncommitted `GLOSSARY.md`/`PROGRESS.md` changes in the tree (S9a Location-line wording,
a D52/TRIAGE-T13 citation fix) predate this task and were left untouched, out of scope.

M1.10 done (2026-09-08, `make check` clean, `go test -race ./cmd/noggin/...` clean). New
`TestDocStoreBackslashRoundTrip` (docstore_test.go) covers the one gap M1.7/M1.8's own
tab/newline round-trip tests left: a literal backslash in both a `<title>` and a `<key>`,
each also placed directly beside a literal `t`/`n` letter (not an actual tab/newline byte)
— the shape that would decode into a phantom tab/newline if `escapeDocField` ever stopped
escaping a raw backslash into `\\`, since `unescapeDocField`'s byte-at-a-time scan reads
any `\` immediately followed by `t`/`n` as an escape sequence regardless of intent.
Confirmed the test actually catches this class of regression, not just restates the
existing tab/newline coverage: temporarily stripped backslash from `escapeDocField`'s
`strings.NewReplacer` call and reran the test in isolation — it failed with exactly the
predicted phantom-tab corruption (`Title:Back\tSlash` decoded back as `Title:Back<TAB>
Slash`) and a dropped backslash-key record; reverted immediately (`docstore.go` diff empty
after revert, confirmed via `git diff --stat`). No production code changes — test-only,
per the task's own scope ("locking in escape-order correctness", not fixing a live bug).

Next: M2.9 (TRIAGE T1, T22: a wikipedia-package test gap, merged at replan — same
gap-shape, same package). Independent of M1.10's own work — next in cursor order.

**Replan note (2026-09-08):** triaged all seven open TRIAGE.md items into new PLAN tasks
(M1.9, M1.10, M2.9, M3.6, M3.7, M4.4 — see PLAN.md for each; none block the M5.1 cursor).
T18 was corrected in the process, not just accepted as filed: rerunning `make check-pty`
showed its 11 failures are a real, deterministic regression from M1.3's and M1.6's own
renames (`test/pty/checks.py` still asserts the pre-rename "Roots" pane label and the File
Picker's old "Active Root" scope label), not sandbox/pty-harness flakiness as filed — see
DECISIONS.md's correction entry. Also closed the open D122 judgment question (M2.3's
cache-key guarantee): now that M4.2/M3.4 have real callers, the codebase's own doc comments
(picker.go's `togglePickerStar`) already show the narrower guarantee is understood and
accepted by design — see DECISIONS.md.

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
| M4.1  | done   | 2026-09-08 | 2026-09-08 | review: code-reviewer subagent, same day; validate: fresh-context validator subagent, same day, PASS |
| M4.2  | done   | 2026-09-08 | 2026-09-08 | review: code-reviewer subagent, same day; found 1 high-sev functional bug (clearPeek scope) + 1 high-sev doc-comment placement, both fixed directly, plus 2 low fixed directly; validate: fresh-context validator subagent, same day, PASS (Peek mechanism correctly not yet wired to a key per D130, deferred to M5.2) |
| M4.3  | done   | 2026-09-08 | 2026-09-08 | review: code-reviewer subagent, same day; found 1 medium functional bug (m.message clobber across sync/async refresh) + 1 medium test gap (stale refresh completion) + 2 low (mirror-case test, GLOSSARY drift), all fixed/added directly; validate: fresh-context validator subagent, same day, PASS (TRIAGE T23: mirror DV-arrangement test gap for staleness/edit) |
| M5.1  | done   | -        | -         | fixed a real persistence bug found while building (D133: toggleActiveDocStar didn't stamp Article/Title for a never-before-starred article) |
| M5.2  | done   | -        | -         | fixed the M5.1-flagged cosmetic gap (article rows rendered file-shaped); Activate/Peek widened to (model, tea.Cmd) for the new fetch dispatch |
| M1.9  | done   | -        | -         | added by replan 2026-09-08 (TRIAGE T18, corrected: real regression from M1.3/M1.6 renames, not flakiness) |
| M1.10 | done   | -        | -         | added by replan 2026-09-08 (TRIAGE T2) |
| M2.9  | todo   | -        | -         | added by replan 2026-09-08 (TRIAGE T1, T22) |
| M3.6  | todo   | -        | -         | added by replan 2026-09-08 (TRIAGE T20) |
| M3.7  | todo   | -        | -         | added by replan 2026-09-08 (TRIAGE T21) |
| M4.4  | todo   | -        | -         | added by replan 2026-09-08 (TRIAGE T23) |
| M6.1  | todo   | -        | -         | pre-agreed cut 1 |
| M7.1  | todo   | -        | -         | pre-agreed cut 2 |
| M7.2  | todo   | -        | -         | pre-agreed cut 2 |
| M7.3  | todo   | -        | -         | pre-agreed cut 2 |
| M8.1  | todo   | -        | -         |      |
| M8.2  | todo   | -        | -         | scope widened by replan 2026-09-08: S9a attribution line |
| M8.3  | todo   | -        | -         |      |
| M9.1  | todo   | -        | -         | pre-agreed cut 3 |
| M10.1 | todo   | -        | -         |      |
