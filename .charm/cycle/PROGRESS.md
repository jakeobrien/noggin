# Progress — C4a wikipedia-spike

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

Current task: M3.1

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M1.1–M2.5 done. M2 (fetch-worthy markdown) is complete; M2.5 (this session) consolidated
the five per-task opt-in live tests (TRIAGE T1) onto one shared fetch cache: new
`live_test.go`'s `fetchLiveSample(t, title)` fetches each S5 title's HTML once (`sync.Once`
+ map) regardless of how many live tests need it; `liveMarkdown(t, title)` runs the full
pipeline on top via a new `articleMarkdownFromHTML(html)`, extracted from `main.go`'s
`articleMarkdown` (pure split, no logic change). All 5 live tests
(`fetch/strip/convert/links/cleanup_test.go`) now use these plus a shared `skipUnlessLive`
guard. Verified: `WIKISPIKE_LIVE=1 go test ./cmd/wikispike -run Live -v` — 20/20 subtests
pass, 0.77s, no 429. `make check` clean.

Validated/Reviewed: M1.1/M1.2 stand at 2026-08-14 (fetch.go/isolation untouched this
session). M2.1–M2.4's Validated cleared to `-` per the doc-map rule — their acceptance
runs through the pipeline function this session's extraction touched (behavior unchanged,
confirmed by the live run above, but a fresh-context validate, not the builder's own tests,
is what the process counts). M2.5 itself has never been reviewed/validated. TRIAGE.md is
empty.

Pipeline (`main.go`): fetch → `articleMarkdownFromHTML` = `stripChrome` (M2.2) →
`stripLinks` (M2.3) → `htmlToMarkdown` (M2.1) → `cleanupMarkdown` (M2.4).

Next: M3.1 — convert the four S5 titles into a scratch root *outside* the repo (not
`/tmp/wikispike-m12`, which is HTML fixtures, not the deliverable) for M3.2's noggin
read-through. Gotcha carried forward: `wikipedia-to-markdown.md` at the repo root sketches
*rendered* HTML, not Parsoid's — keep preferring Parsoid's `typeof`/`rel` attributes and
stable class names over its selector list.

## Status

| Key  | Status  | Reviewed | Validated | Note                                        |
|------|---------|----------|-----------|---------------------------------------------|
| M1.1 | done    | -        | 2026-08-14 | `cmd/wikispike/`                            |
| M1.2 | done    | -        | 2026-08-14 | live fetch verified, all four S5 titles     |
| M2.1 | done    | -        | -         | re-verify after M2.5's main.go extraction   |
| M2.2 | done    | -        | -         | re-verify after M2.5's main.go extraction   |
| M2.3 | done    | -        | -         | re-verify after M2.5's main.go extraction   |
| M2.4 | done    | -        | -         | re-verify after M2.5's main.go extraction   |
| M2.5 | done    | -        | -         | consolidate live tests (was TRIAGE T1)      |
| M3.1 | todo    | -        | -         |                                             |
| M3.2 | todo    | -        | -         | user-judged                                 |
| M3.3 | todo    | -        | -         | repeatable                                  |
| M3.4 | todo    | -        | -         | optional (S6 exception)                     |
| M4.1 | todo    | -        | -         |                                             |
| M4.2 | todo    | -        | -         |                                             |
