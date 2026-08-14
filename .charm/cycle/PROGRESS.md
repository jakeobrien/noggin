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

Current task: M1.1

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

Cycle just planned; no code written yet. Start at M1.1 (scaffold the spike `main`).

Network is fine: `en.wikipedia.org` was allow-listed and verified end-to-end (D99) — REST
HTML 200 for all four S5 titles, and Go's `net/http` reaches it through the sandbox proxy
with no TLS special-casing. The Go module proxy is reachable too, so pulling `goquery` /
`html-to-markdown` will work.

Gotcha for M2.2: `wikipedia-to-markdown.md` at the repo root is a sketch, not a spec, and
its selector list is written for *rendered* article HTML. The REST v1 endpoint serves
**Parsoid** HTML, which differs — a Major Arcana probe found zero `.mw-editsection` and
zero `.infobox`, refs as `sup.mw-ref[typeof^="mw:Extension/ref"]`, the reference list as
`typeof="mw:Extension/references"`, images as `typeof^="mw:File"`, and navboxes as
`div.navbox` wrapping `table.navbox-inner`. Content tables are `table.wikitable`. Prefer
Parsoid's `typeof`/`rel` attributes over the sketch's class list.

## Status

| Key  | Status  | Reviewed | Validated | Note                                        |
|------|---------|----------|-----------|---------------------------------------------|
| M1.1 | todo    | -        | -         |                                             |
| M1.2 | todo    | -        | -         |                                             |
| M2.1 | todo    | -        | -         |                                             |
| M2.2 | todo    | -        | -         |                                             |
| M2.3 | todo    | -        | -         |                                             |
| M2.4 | todo    | -        | -         |                                             |
| M3.1 | todo    | -        | -         |                                             |
| M3.2 | todo    | -        | -         | user-judged                                 |
| M3.3 | todo    | -        | -         | repeatable                                  |
| M3.4 | todo    | -        | -         | optional (S6 exception)                     |
| M4.1 | todo    | -        | -         |                                             |
| M4.2 | todo    | -        | -         |                                             |
