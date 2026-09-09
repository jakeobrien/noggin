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

Current task: M8.3

## Handoff

M8.2 (Document Viewer location line, S9/S9a) is done: each Document Viewer's bottom row
shows, de-emphasized, `<Root name> · <relative path>` for a file under a Root, the full path
for an orphan, or S9a's fixed `CC BY-SA 4.0 · <URL> · Reformatted to markdown` for an article
(location.go, new: `documentLocation`, `viewer.locationDoc`, `model.viewerLocationLine`/
`composeViewerContent`); reflects a Peeked document while the Peek lasts, absent with no
document, and costs the content exactly one row — the Active Document's viewport and Peek's
own each independently reserve that row via a new `viewer.boxHeight` field (the raw per-pane
content height from layout) plus `docContentHeight`/`peekContentHeight`, recomputed at the end
of `render()`/`renderPeek()` once `v.err`/`v.peek.err` are known. `model.composeViewerContent`
(called from `frame()` for both DV1/DV2) is the one place the line's text joins the rendered
content, reading `m.roots.rootList()` fresh each render; `viewer.view()` itself is unchanged in
signature. `fill()`'s signature widened to take explicit width/height (not always `v.vp`'s)
since Peek's own dimensions can now genuinely differ from the Active Document's.

A fresh-context code-reviewer pass (mid-task) found, and this session then fixed, two Major
gaps: (1) starting an article fetch (Activate or Refresh) into a pane whose location line was
already showing left the viewport's reserved row stale during the fetch, and a *failed* fetch
never restored it either — fixed via two new viewer methods, `beginFetch`/`endFetch`, now the
sole way `pendingFetch` is ever set/cleared (`article_fetch.go`'s three call sites updated).
(2) `view()`'s fetching branch filled to the Active Document's own dimensions even when the
pending fetch targets Peek (dead code today, pending M5.2's real trigger) — fixed to branch on
`pendingFetch.peek`. Both fixes are pinned by regression tests, each confirmed to actually fail
against the pre-fix code (temporary revert, mutation-tested, then restored). A second reviewer
pass after both fixes found no further issues.

`location_test.go` (new) covers every acceptance clause: the four states (file/orphan/article/
no-document) with exact-text assertions, a renamed-Root variant, Peek reflection, DV1/DV2
independence, the content-height drop (direct + `TestViewerRewrapsOnResize`'s updated
expectation), an end-to-end `model.render()` frame check, and the two fetch-height regression
tests. Added 3 pty checks (`test/pty/checks.py`, before `done()`). `make check` and
`make check-pty` (~250 checks) both green. Not yet reviewed (tracked `Reviewed` gate) or
validated.

Cursor is now M8.3 (de-emphasized secondary text — Sources Nav path line, Picker File-row
paths; needs: M7.3, already done). M8.1 and M8.2 both done, neither reviewed/validated yet.
M1.11, M2.10 and M5.3 remain independent replan-added hygiene tasks, pull-forward-anytime.
Working tree: M8.2's code+test changes are committed as this task's own commit.

No blockers.

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
| M1.8  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T17) |
| M2.8  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T15) |
| M3.1  | done   | 2026-09-08 | 2026-09-08 |      |
| M3.2  | done   | 2026-09-08 | 2026-09-08 |      |
| M3.3  | done   | 2026-09-08 | 2026-09-08 |      |
| M3.4  | done   | 2026-09-08 | 2026-09-08 |      |
| M3.5  | done   | 2026-09-08 | 2026-09-08 |      |
| M4.1  | done   | 2026-09-08 | 2026-09-08 |      |
| M4.2  | done   | 2026-09-08 | 2026-09-08 |      |
| M4.3  | done   | 2026-09-08 | 2026-09-08 |      |
| M5.1  | done   | 2026-09-08 | 2026-09-08 |      |
| M5.2  | done   | 2026-09-08 | 2026-09-08 |      |
| M1.9  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T18) |
| M1.10 | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T2) |
| M2.9  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T1, T22) |
| M2.10 | todo   | -        | -         | added by replan 2026-09-08 (TRIAGE T2) |
| M3.6  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T20) |
| M3.7  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T21) |
| M4.4  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T23) |
| M1.11 | todo   | -        | -         | added by replan 2026-09-08 (TRIAGE T1, T3) |
| M5.3  | todo   | -        | -         | added by replan 2026-09-08 (TRIAGE T4) |
| M6.1  | done   | 2026-09-09 | 2026-09-08 | pre-agreed cut 1 |
| M7.1  | done   | 2026-09-08 | 2026-09-08 | pre-agreed cut 2 |
| M7.2  | done   | 2026-09-08 | 2026-09-08 | pre-agreed cut 2 |
| M7.3  | done   | 2026-09-08 | 2026-09-08 | pre-agreed cut 2 |
| M8.1  | done   | -        | -         |      |
| M8.2  | done   | -        | -         | scope widened by replan 2026-09-08: S9a attribution line |
| M8.3  | todo   | -        | -         |      |
| M9.1  | todo   | -        | -         | pre-agreed cut 3 |
| M10.1 | todo   | -        | -         |      |
