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

Current task: M2.3

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M1.1, M1.2, M2.1 and M2.2 done. `cmd/wikispike/` is the spike's own `main`
(`wikispike <title> <output path>`), header-commented experimental; grep from `cmd/noggin/`
and `internal/` is clean. `fetch.go` fetches Parsoid HTML from
`https://en.wikipedia.org/api/rest_v1/page/html/<escaped title>` with a descriptive
User-Agent (S8). All four S5 titles fetch live (158–570 KB each).

`convert.go`'s `htmlToMarkdown` (M2.1) wraps `JohannesKaufmann/html-to-markdown/v2` with the
base, commonmark and table plugins (defaults, no options tuned yet). `strip.go`'s
`stripChrome` (M2.2, goquery) runs first in `articleMarkdown` and removes: `div.navbox`,
`table.infobox`/`.infobox`, `.mw-editsection` (defensive, never present in Parsoid),
`sup[typeof*="mw:Extension/ref"]` (citation superscripts), the whole enclosing `<section>`
of any `.mw-references-wrap`/`[typeof*="mw:Extension/references"]` (reference-list
sections, so no empty "References"/"Notes" heading survives either), `table.metadata`/
`table.ambox`, and every image — `[typeof^="mw:File"]`, `[typeof="mw:Extension/math"]`
(rendered formulae are also served as an `<img>`) and a final bare-`<img>` sweep. Gotcha:
the references typeof lands on the wrapping div for some articles and on the inner `<ol>`
for others — match on the class name too, not just the attribute. Verified live against all
four S5 titles: no `[edit]`, no `## References`/`## Notes` heading, no `![`, no
navbox/infobox leak, Major Arcana's wikitable content tables intact (24 `|` rows).
`goquery` needed re-adding (`go mod tidy` had dropped it after M2.1 since it went unused).
`pipeline` is a package var so `run()`'s tests stay off the network. Live API test is opt-in
(D101): `WIKISPIKE_LIVE=1 go test ./cmd/wikispike -run Live -v`; keep `make check` hermetic —
`convert_test.go` and `strip_test.go` follow the same opt-in shape. Note: the live REST API
rate-limited (429) under repeated back-to-back fetches during this session — space out live
test runs or reuse one fetch across assertions rather than re-fetching per check. User-Agent
contact is env-driven (D102, `WIKISPIKE_CONTACT`).

Next: M2.3 — strip link markup, keeping the link's text as ordinary prose. Current live
output still carries full markdown link syntax throughout (e.g. Major Arcana's content table
cells: `[The Fool](./The_Fool_%28tarot_card%29 "The Fool (tarot card)")`) — expected, M2.2
didn't touch links. Fetched samples for offline iteration are in `/tmp/wikispike-m12/*.html`
(regenerable, not committed).

Gotcha carried from M2.2: `wikipedia-to-markdown.md` at the repo root is a sketch, not a
spec, and its selector list is written for *rendered* article HTML, not Parsoid's. Prefer
Parsoid's `typeof`/`rel` attributes and stable class names over the sketch's class list.

## Status

| Key  | Status  | Reviewed | Validated | Note                                        |
|------|---------|----------|-----------|---------------------------------------------|
| M1.1 | done    | -        | -         | `cmd/wikispike/`                            |
| M1.2 | done    | -        | -         | live fetch verified, all four S5 titles     |
| M2.1 | done    | -        | -         |                                             |
| M2.2 | done    | -        | -         |                                             |
| M2.3 | todo    | -        | -         |                                             |
| M2.4 | todo    | -        | -         |                                             |
| M3.1 | todo    | -        | -         |                                             |
| M3.2 | todo    | -        | -         | user-judged                                 |
| M3.3 | todo    | -        | -         | repeatable                                  |
| M3.4 | todo    | -        | -         | optional (S6 exception)                     |
| M4.1 | todo    | -        | -         |                                             |
| M4.2 | todo    | -        | -         |                                             |
