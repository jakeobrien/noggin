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

Current task: M2.4

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M1.1, M1.2, M2.1, M2.2 and M2.3 done. `cmd/wikispike/` is the spike's own `main`
(`wikispike <title> <output path>`), header-commented experimental; grep from `cmd/noggin/`
and `internal/` is clean. `fetch.go` fetches Parsoid HTML from
`https://en.wikipedia.org/api/rest_v1/page/html/<escaped title>` with a descriptive
User-Agent (S8). All four S5 titles fetch live (158–570 KB each).

Pipeline order in `main.go`'s `articleMarkdown`: fetch → `stripChrome` (M2.2) →
`stripLinks` (M2.3) → `htmlToMarkdown` (M2.1). `convert.go`'s `htmlToMarkdown` wraps
`JohannesKaufmann/html-to-markdown/v2` with the base, commonmark and table plugins
(defaults, no options tuned yet). `strip.go`'s `stripChrome` (goquery) removes:
`div.navbox`, `table.infobox`/`.infobox`, `.mw-editsection` (defensive, never present in
Parsoid), `sup[typeof*="mw:Extension/ref"]` (citation superscripts), the whole enclosing
`<section>` of any `.mw-references-wrap`/`[typeof*="mw:Extension/references"]`
(reference-list sections — the references typeof lands on the wrapping div for some
articles and on the inner `<ol>` for others, so match the class too), `table.metadata`/
`table.ambox`, and every image — `[typeof^="mw:File"]`, `[typeof="mw:Extension/math"]`
(rendered formulae are also served as an `<img>`) and a final bare-`<img>` sweep.
`links.go`'s `stripLinks` (goquery) unwraps every remaining `<a>` — `ReplaceWithHtml(inner)`
replaces the anchor with its own inner HTML, keeping child markup (bold/italic) and text in
place with no href. Verified live against all four S5 titles: no `[edit]`, no
`## References`/`## Notes` heading, no `![`, no navbox/infobox leak, no `](` or bare
`http(s)://`, Major Arcana's wikitable content tables intact with card names surviving as
plain text (previously wikilinks in nearly every cell). `goquery` needed re-adding after
M2.1 (`go mod tidy` had dropped it since it went unused then). `pipeline` is a package var
so `run()`'s tests stay off the network. Live API test is opt-in (D101):
`WIKISPIKE_LIVE=1 go test ./cmd/wikispike -run Live -v`; keep `make check` hermetic —
`convert_test.go`, `strip_test.go` and `links_test.go` follow the same opt-in shape.

T1 (open, medium): the growing set of per-task live tests now trips the REST API's 429 when
run together via `-run Live`; individual `-run <TestName>` passes. Space out live runs or
run one test at a time until T1 is triaged — don't read this as a code regression.
User-Agent contact is env-driven (D102, `WIKISPIKE_CONTACT`).

Next: M2.4 — clean up the holes M2.2/M2.3 leave: orphaned captions, empty table
rows/list items, dangling brackets/parens, runs of 3+ blank lines. Likely culprits already
visible: emptied `<dd>`/`<li>` wrappers where a math-formula `<img>` was removed (M2.2), and
any table cell that was purely a link with no other content. Fetched samples for offline
iteration are in `/tmp/wikispike-m12/*.html` (regenerable, not committed).

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
| M2.3 | done    | -        | -         |                                             |
| M2.4 | todo    | -        | -         |                                             |
| M3.1 | todo    | -        | -         |                                             |
| M3.2 | todo    | -        | -         | user-judged                                 |
| M3.3 | todo    | -        | -         | repeatable                                  |
| M3.4 | todo    | -        | -         | optional (S6 exception)                     |
| M4.1 | todo    | -        | -         |                                             |
| M4.2 | todo    | -        | -         |                                             |
