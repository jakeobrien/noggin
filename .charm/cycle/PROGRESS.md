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

Current task: M2.2

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M1.1, M1.2 and M2.1 done. `cmd/wikispike/` is the spike's own `main`
(`wikispike <title> <output path>`), header-commented experimental; grep from `cmd/noggin/`
and `internal/` is clean. `fetch.go` fetches Parsoid HTML from
`https://en.wikipedia.org/api/rest_v1/page/html/<escaped title>` with a descriptive
User-Agent (S8). All four S5 titles fetch live (158–570 KB each).

`convert.go`'s `htmlToMarkdown` (M2.1) wraps `JohannesKaufmann/html-to-markdown/v2` with the
base, commonmark and table plugins (defaults, no options tuned yet) — `articleMarkdown` in
`main.go` now fetches then converts. Verified against live Major Arcana: headings,
paragraphs and wikitable content tables all render (`## Section`, prose, `| --- |` rows).
`goquery` was added by `go get` while scoping M2.1 but never used, so `go mod tidy` dropped
it — M2.2 is the task that will actually need it for element removal; re-add there, not now.
`pipeline` is a package var so `run()`'s tests stay off the network. Live API test is opt-in
(D101): `WIKISPIKE_LIVE=1 go test ./cmd/wikispike -run Live -v`; keep `make check` hermetic —
`convert_test.go` follows the same opt-in shape for its live conversion check. User-Agent
contact is env-driven (D102, `WIKISPIKE_CONTACT`).

Next: M2.2 — strip non-article chrome (navboxes, infoboxes, edit-section links, ref
superscripts/lists, images) before conversion. Current live output still carries all of
these (see Major Arcana's raw ref-link markup in the table cells) — expected, M2.1 was
conversion-only. Fetched samples for offline iteration are in `/tmp/wikispike-m12/*.html`
(regenerable, not committed).

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
| M1.1 | done    | -        | -         | `cmd/wikispike/`                            |
| M1.2 | done    | -        | -         | live fetch verified, all four S5 titles     |
| M2.1 | done    | -        | -         |                                             |
| M2.2 | todo    | -        | -         |                                             |
| M2.3 | todo    | -        | -         |                                             |
| M2.4 | todo    | -        | -         |                                             |
| M3.1 | todo    | -        | -         |                                             |
| M3.2 | todo    | -        | -         | user-judged                                 |
| M3.3 | todo    | -        | -         | repeatable                                  |
| M3.4 | todo    | -        | -         | optional (S6 exception)                     |
| M4.1 | todo    | -        | -         |                                             |
| M4.2 | todo    | -        | -         |                                             |
