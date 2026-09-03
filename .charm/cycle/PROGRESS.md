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

Current task: M3.3a

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

**2026-09-03, this session (charm-build) — M2.4a done, cursor moved to M3.3a.**
`cleanupProse` gained `dropUnmatchedClosers` (`cleanup.go`), a depth-based pass that drops
a dangling `)`/`]` with no matching opener; parens and brackets tracked independently,
closers-only (asymmetric — an unmatched opener is left alone), see DECISIONS D107.
Re-investigated D106's premise while building: the `list_languages.md` dangling `)` isn't
from a removal at all — both instances (`Scala (into JVM bytecode))`, `RUNCOM (...) CTSS)
programs)`) are already unbalanced in the live article's own wikitext, pre-dating this
pipeline. D106's other option (a real depth pass) covers it regardless of root cause.
Verified mechanically: regenerated all four current `tmp/wikispike-tests-1/*.md` samples
with `bin/wikispike` — all four now have equal `(`/`)` and `[`/`]` counts (previously
list_languages.md was 386/388). Added a live-test check for the same (`cleanup_test.go`'s
`TestPipelineLiveSamplesHaveNoStructuralHoles`, extended) plus two unit tests
(`TestCleanupMarkdownDropsDanglingClosers`,
`TestDropUnmatchedClosersLeavesUnmatchedOpenersAlone`). `make check` and the full opt-in
live suite (`WIKISPIKE_LIVE=1 go test ./cmd/wikispike -run Live -v`) both pass. Root-level
stray sample copies (`why_anything.md`, `major_arcana.md`, `e_math-2.md`, `e_math-3.md`) at
repo root were left untouched — that's M4.1's TRIAGE-T4-derived done-when, not this task.
Next step: work M3.3a (portal-bar/sistersitebox chrome), then M3.6, then resume M4.1/M4.2.

## Status

| Key  | Status  | Reviewed | Validated | Note                                        |
|------|---------|----------|-----------|---------------------------------------------|
| M1.1 | done    | 2026-09-03 | 2026-08-14 | `cmd/wikispike/`                            |
| M1.2 | done    | 2026-09-03 | 2026-08-14 | live fetch verified, all four S5 titles     |
| M2.1 | done    | 2026-09-03 | -         | re-verify after M2.5's main.go extraction   |
| M2.2 | done    | 2026-09-03 | -         | reopened post-M2.5 for table.sidebar (user-directed fix-up); re-verify |
| M2.3 | done    | 2026-09-03 | -         | re-verify after M2.5's main.go extraction   |
| M2.4 | done    | 2026-09-03 | -         | re-verify after M2.5's main.go extraction; review found T3 (dangling paren, done-when not fully met) — fix tracked as M2.4a, D106 |
| M2.5 | done    | 2026-09-03 | -         | consolidate live tests (was TRIAGE T1)      |
| M3.1 | done    | 2026-09-03 | -         | scratch root, 4 .md files (user, off-sandbox) |
| M3.2 | done    | 2026-09-03 | -         | `findings/sample-judging.md`; only E fails A3 |
| M3.3 | done    | 2026-09-03 | -         | nothing to retune beyond M3.4's carve-out   |
| M3.4 | done    | 2026-09-03 | -         | `math.go`, termtex dep; findings recorded   |
| M3.5 | done    | 2026-09-03 | -         | `images.go`; no new dep; findings recorded  |
| M2.4a | done   | -        | -         | fix-up (TRIAGE T3): dangling-paren gap; see D107 |
| M3.3a | todo   | -        | -         | fix-up (TRIAGE T2): portal-bar/sistersitebox |
| M3.6  | todo   | -        | -         | housekeeping (TRIAGE T5/T6/T7)              |
| M4.1 | todo    | -        | -         | widened for M3.4/M3.5 report + TRIAGE T4    |
| M4.2 | todo    | -        | -         |                                             |
