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

Current task: M3.4

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

**2026-09-02 replan (this session):** SPEC.md was amended between sessions (D104,
`8a1bd64`) to widen the retune budget — S3→S3a, S6→S6a, S7→S7a — adding two costed
experiments (Unicode maths approximation, image caption/alt/filename placeholders) that
PLAN.md hadn't caught up to. Reconciled: M3.4 rewritten to try Unicode first then fall back
to raw-text preservation (effort 2→3); new M3.5 added for the image placeholder experiment
(needs M3.3, effort 3); M4.1 widened to require per-experiment reporting for both M3.4 and
M3.5 (effort 2→3, needs now include M3.4/M3.5); M2/M3/M4 Delivers lines updated to the `a`-
suffixed spec keys. No TRIAGE items were open to route (inbox is empty) and D104 already
carries the material-change rationale, so no new DECISIONS.md entry was needed for the plan
mechanics themselves.

**Same session, user-reported (off-sandbox):** M3.1 done — user ran `wikispike` for all
four S5 titles and wrote the four `.md` files to a scratch root outside the repo. M3.2's
substantive work is also done — user read all four in noggin's Document Viewer at real pane
widths and judged them: acceptable on most counts, with the sole exception being exactly the
math (`E (mathematical constant)`) and image-placeholder gaps already surfaced and re-spec'ed
last session as D104/S3a/S6a. **Gap:** M3.2's own acceptance criterion asks for that verdict
"recorded in the findings file" — no findings file exists yet (`findings/INDEX.md` is still
the bare template), and the report above is a summary, not the per-article A1–A4 breakdown
(plus A5 known-limitations) the criterion names. Left `in-progress` rather than `done` for
that reason; writing it up is the immediate next step, and needs the user's per-article
detail (which article(s) hit A1/A2/A3/A4, and which of those are A5 known-limitations vs.
real shortfalls) to be accurate rather than inferred.

Pipeline (`main.go`): fetch → `articleMarkdownFromHTML` = `stripChrome` (M2.2) →
`stripLinks` (M2.3) → `htmlToMarkdown` (M2.1) → `cleanupMarkdown` (M2.4). Note for M3.5:
this pipeline currently strips images without a trace (old S3) — the image-placeholder
experiment means no longer deleting the image node in `stripChrome`.

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
| M3.2 | done    | -        | -         | judged (user); findings file not yet written  |
| M3.3 | todo    | -        | -         | repeatable                                  |
| M3.4 | todo    | -        | -         | S6a maths exception (needs M3.3)            |
| M3.5 | todo    | -        | -         | S6a image exception (needs M3.3)            |
| M4.1 | todo    | -        | -         | widened for M3.4/M3.5 per-experiment report |
| M4.2 | todo    | -        | -         |                                             |
