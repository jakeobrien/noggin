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

Current task: M3.5

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

**2026-09-02, this session:** Closed out the M3.2 gap flagged by the prior session — got
the user's per-article A1–A4 breakdown and wrote it up as
`cycle/findings/sample-judging.md`: all four pass A1/A2/A4; only `E (mathematical
constant)` fails A3 (stripped maths leaves visible gaps), which is exactly A5's named
known-limitation case (go with a known limitation, not a no-go). The earlier handoff's
mention of an "image-placeholder gap" doesn't reproduce in this structured pass — noted in
the findings file as a discrepancy, doesn't change M3.5's scope since D104 commits to that
experiment regardless. Fixed the Status-table/Handoff contradiction the prior session left
(table said M3.2 `done`, handoff said it should be `in-progress`) — now genuinely `done`.
M3.3 (generic retune) had nothing left to retune beyond the one shortfall already carved
into its own task, so recorded `done` too, citing the findings file.

**M3.4 (maths) done.** New file `cmd/wikispike/math.go` (+ `math_test.go`); new pipeline
stage `renderMath` runs before `stripChrome` in `articleMarkdownFromHTML`
(`cmd/wikispike/main.go`). Per formula: try `github.com/doug/termtex` (new dependency,
permitted by D104) for a Unicode approximation, keep it only if single-line, else fall back
to the raw TeX source Parsoid already exposes in `data-mw`'s `body.extsrc` — both paths
guarantee no mid-sentence gap. Verified against a live fetch of `E (mathematical constant)`:
every previously-orphaned sentence now reads complete; 63/104 (61%) formulas got Unicode,
41/104 raw-text fallback; `make check` and a mechanical re-check (blank-line runs, empty
list items, bracket balance) clean across all four S5 samples. Findings:
`cycle/findings/maths-experiment.md`. `strip.go`'s math-removal comment updated to reflect
it now only catches the rare span `renderMath` couldn't recover a source for.

Five `.md` conversion outputs sit untracked at the repo root (`e_math.md`,
`list_languages.md`, `major_arcana.md`, `multi_paradigm.md`, `why_anything.md`) — left in
place at the user's request, for their reference/comparison (predate M3.4, so don't reflect
the maths fix — regenerate via `go run ./cmd/wikispike "<title>" <path>` for an up-to-date
comparison). Note `multi_paradigm.md` is a conversion of a *different* article ("Comparison
of multi-paradigm programming languages"), not the S5 title ("List of programming languages
by type", which is `list_languages.md`) — harmless, just don't mistake it for a fifth
sample.

Pipeline (`main.go`) is now: fetch → `renderMath` (M3.4) → `stripChrome` (M2.2) →
`stripLinks` (M2.3) → `htmlToMarkdown` (M2.1) → `cleanupMarkdown` (M2.4). Note for M3.5:
this pipeline still strips images without a trace (old S3) — the image-placeholder
experiment means no longer deleting the image node in `stripChrome`, likely following the
same "new pipeline stage before `stripChrome`" shape `renderMath` just established.

**Next step:** implement M3.5 — image caption/alt/filename placeholder experiment.

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
| M3.1 | done    | -        | -         | scratch root, 4 .md files (user, off-sandbox) |
| M3.2 | done    | -        | -         | `findings/sample-judging.md`; only E fails A3 |
| M3.3 | done    | -        | -         | nothing to retune beyond M3.4's carve-out   |
| M3.4 | done    | -        | -         | `math.go`, termtex dep; findings recorded   |
| M3.5 | todo    | -        | -         | S6a image exception (needs M3.3)            |
| M4.1 | todo    | -        | -         | widened for M3.4/M3.5 per-experiment report |
| M4.2 | todo    | -        | -         |                                             |
