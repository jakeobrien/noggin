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

M1.1, M1.2, M2.1–M2.4 done — M2 (fetch-worthy markdown) is complete. `cmd/wikispike/` is
the spike's own `main` (`wikispike <title> <output path>`), header-commented experimental;
grep from `cmd/noggin/` and `internal/` is clean.

Pipeline order in `main.go`'s `articleMarkdown`: fetch → `stripChrome` (M2.2) →
`stripLinks` (M2.3) → `htmlToMarkdown` (M2.1) → `cleanupMarkdown` (M2.4, new this session).
`cleanup.go`'s `cleanupMarkdown` runs last because the holes it closes are markdown-shaped,
not DOM-shaped: by the time an element is gone, what's left is plain text on either side
with no marker of what used to sit between them. It closes three things, confirmed against
adversarial fixtures (a formula and a lone citation each alone inside parens; a vanished
block leaving 3+ blank lines; an empty list-item marker) and re-verified live against all
four S5 titles (`cleanup_test.go`'s `TestPipelineLiveSamplesHaveNoStructuralHoles`, opt-in):
empty/whitespace-only parens and brackets (`()`/`[]`, e.g. "denoted () after" when the whole
parenthetical was a removed formula or lone citation) with the resulting double-space/
space-before-punctuation also collapsed, 3+ consecutive blank lines collapsed to one, and
empty list-item lines dropped. It deliberately does *not* try to repair prose that merely
reads awkwardly because a formula was dropped mid-sentence (e.g. "One has where is the...")
— that's S6's problem, scoped to the optional M3.4, not this task.

Investigated first with an ad hoc debug harness against the M1.2-era cached samples in
`/tmp/wikispike-m12/*.html` (regenerable, not committed): the *live* samples as fetched
already passed all three mechanical checks with no cleanup code at all — the real holes
only showed up once I hand-built adversarial fixtures (formula/citation alone in parens).
Table cells that are empty after their sole citation was stripped (e.g. a wikitable "Note"
column) are left as-is — a legitimately blank cell, not a hole; M2.4's acceptance criterion
doesn't cover it and forcing every cell non-empty isn't the same problem.

`convert.go`'s `htmlToMarkdown` wraps `JohannesKaufmann/html-to-markdown/v2` with base,
commonmark and table plugins (defaults). `strip.go`'s `stripChrome` (goquery) removes
navboxes, infoboxes, edit-section links (defensive), citation superscripts, the whole
enclosing reference-list `<section>`, metadata tables, and every image (including math
formulae, served as `<img>`/`typeof="mw:Extension/math"`). `links.go`'s `stripLinks`
unwraps every remaining `<a>`, keeping its inner content in place with no href.

T1 (open, medium, unchanged): the growing set of per-task live tests trips the REST API's
429 when run together via `-run Live`; individual `-run <TestName>` passes — this session's
new `TestPipelineLiveSamplesHaveNoStructuralHoles` was verified solo. Space out live runs
until T1 is triaged. User-Agent contact is env-driven (D102, `WIKISPIKE_CONTACT`).

Next: M3.1 — convert the four S5 titles into a scratch root *outside* the repo (not
`/tmp/wikispike-m12`, which is HTML fixtures, not the deliverable) for M3.2's noggin
read-through. Gotcha carried forward: `wikipedia-to-markdown.md` at the repo root is a
sketch for *rendered* HTML, not Parsoid's — keep preferring Parsoid's `typeof`/`rel`
attributes and stable class names over its selector list.

## Status

| Key  | Status  | Reviewed | Validated | Note                                        |
|------|---------|----------|-----------|---------------------------------------------|
| M1.1 | done    | -        | 2026-08-14 | `cmd/wikispike/`                            |
| M1.2 | done    | -        | 2026-08-14 | live fetch verified, all four S5 titles     |
| M2.1 | done    | -        | 2026-08-14 |                                             |
| M2.2 | done    | -        | 2026-08-14 |                                             |
| M2.3 | done    | -        | 2026-08-14 |                                             |
| M2.4 | done    | -        | 2026-08-14 |                                             |
| M3.1 | todo    | -        | -         |                                             |
| M3.2 | todo    | -        | -         | user-judged                                 |
| M3.3 | todo    | -        | -         | repeatable                                  |
| M3.4 | todo    | -        | -         | optional (S6 exception)                     |
| M4.1 | todo    | -        | -         |                                             |
| M4.2 | todo    | -        | -         |                                             |
