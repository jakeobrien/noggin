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

Current task: M4.1

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

**2026-09-02, this session (fix-up on M2.2, done, ad hoc, cursor stayed on M4.1) —
sidebars now stripped.** User flagged `e_math.md`'s History section opening with a
disjointed run of paragraphs/lists (Wikipedia's "Part of a series of articles on e" box)
ahead of the section's real prose. Root cause: Wikipedia's `<table class="sidebar">`
template (a second, distinct navigation-box class from `div.navbox`, already stripped)
wasn't matched by `stripChrome` (M2.2) — same category of gap as TRIAGE T2's
portal-bar/sister-box discovery, but a different selector, so fixed directly rather than
added to T2. `strip.go` now also removes `table.sidebar`; `strip_test.go` gained a nested
unit fixture matching the live structural placement plus a live-sample regression check
(`WIKISPIKE_LIVE=1`, verified passing against the real article). GLOSSARY's Chrome entry
updated. `make check`/`go test ./...` clean; all four S5 `.md` samples regenerated (only
`e_math.md` and `list_languages.md` changed size — both had a sidebar template).

**2026-09-02, earlier this session (fix-up on M3.4, done, ad hoc, cursor stayed on M4.1) —
display formulas now get their own line.** User flagged raw TeX still visible in `e_math.md`;
turned out Parsoid marks a formula display/block-style (`data-mw.attrs.display=="block"`,
`mwe-math-element-block` class) and Wikipedia gives it its own line visually even though
the span stays nested in the same `<p>`. `renderMath` now wraps display formulas in
`<pre>` (auto-splits the paragraph when htmlToMarkdown re-parses; keeps full multi-line
termtex Unicode instead of falling back to raw TeX). Found and fixed in the same session,
via code-reviewer: `cleanupMarkdown`'s empty-parens/brackets/blank-run/whitespace passes
ran over the whole string, corrupting fenced ASCII-art math — fixed by carving fenced
regions out before any pass runs (`cleanupProse` helper), with tests reproducing the
corruption first. `make check`/`go test ./...` clean; all four S5 `.md` samples
regenerated and re-checked. Detail + why this reopens M3.4's "deliberately did not build"
call: `cycle/findings/maths-experiment.md`'s Addendum. Not run through charm-replan (small,
session-scoped, user-directed) — flagging here so M4.1's findings write-up accounts for it.

**2026-09-02, earlier this session — M3.5 (images) done.** New pipeline stage `renderImages`
(`cmd/wikispike/images.go` + `images_test.go`), same shape as M3.4's `renderMath`: a
goquery pass over `[typeof^="mw:File"]` nodes before `stripChrome`, replacing each with a
plain-text `[Image: <description>]` placeholder — caption (flattened via `.Text()`) → img
`alt` → filename from `resource`/`href` (File: prefix/extension stripped, underscores back
to spaces). No new dependency. `stripChrome`'s image selector is now the fallback for
whatever `renderImages` can't recover a description for (same relationship it already had
with `renderMath`). Pipeline: fetch → `renderMath` (M3.4) → `renderImages` (M3.5) →
`stripChrome` → `stripLinks` → `htmlToMarkdown` → `cleanupMarkdown`.

Verified against live fetches of all four S5 titles: every content figure that used to
vanish now shows its caption in place (none needed the alt/filename fallback live; that
path is covered by unit tests instead). Mechanical re-check (blank lines, empty list
items, bracket balance) and `make check`/`go test ./...` clean; new opt-in live test
`TestArticleMarkdownLiveSamplesShowImagePlaceholders`. Findings:
`cycle/findings/images-experiment.md`. Regenerated the four S5 root `.md` reference files
(`why_anything.md`, `e_math.md`, `list_languages.md`, `major_arcana.md` — still
untracked/off-sandbox per the user's request; `multi_paradigm.md` untouched, not an S5
title).

**Discovery routed to TRIAGE, not fixed:** regenerating surfaced a pre-existing
chrome-stripping gap (M2.2, not this experiment) — `stripChrome` doesn't match Wikipedia's
portal-bar or sister-project side-box chrome, so those boxes' list items were already
leaking as bare text; now visible as `[Image: icon]` etc. instead of silent. Logged as
TRIAGE T2.

**Next step:** M4.1 — write the cycle's consolidated findings record (what converts
well/badly, retune outcome, both S6a experiments' summaries — `maths-experiment.md` and
`images-experiment.md` already exist and can be cited/condensed). TRIAGE has one open item
(T2, medium, non-blocking) — doesn't block M4.1/M4.2, but drain it via charm-replan before
the cycle closes.

## Status

| Key  | Status  | Reviewed | Validated | Note                                        |
|------|---------|----------|-----------|---------------------------------------------|
| M1.1 | done    | -        | 2026-08-14 | `cmd/wikispike/`                            |
| M1.2 | done    | -        | 2026-08-14 | live fetch verified, all four S5 titles     |
| M2.1 | done    | -        | -         | re-verify after M2.5's main.go extraction   |
| M2.2 | done    | -        | -         | reopened post-M2.5 for table.sidebar (user-directed fix-up); re-verify |
| M2.3 | done    | -        | -         | re-verify after M2.5's main.go extraction   |
| M2.4 | done    | -        | -         | re-verify after M2.5's main.go extraction   |
| M2.5 | done    | -        | -         | consolidate live tests (was TRIAGE T1)      |
| M3.1 | done    | -        | -         | scratch root, 4 .md files (user, off-sandbox) |
| M3.2 | done    | -        | -         | `findings/sample-judging.md`; only E fails A3 |
| M3.3 | done    | -        | -         | nothing to retune beyond M3.4's carve-out   |
| M3.4 | done    | -        | -         | `math.go`, termtex dep; findings recorded   |
| M3.5 | done    | -        | -         | `images.go`; no new dep; findings recorded  |
| M4.1 | todo    | -        | -         | widened for M3.4/M3.5 per-experiment report |
| M4.2 | todo    | -        | -         |                                             |
