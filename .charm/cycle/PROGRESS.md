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

**2026-09-03, later same day (charm-validate, fresh-context subagent) — M3.3a FAILS re-verification; M4.1 blocked.**
Ran a deep functional validation (deterministic checks + live re-exercise of the spike
against all four S5 titles, fresh output, not stale samples) over every task claimed
`done`: M1.1, M1.2, M2.1–M2.5, M2.4a, M3.1–M3.6, M3.3a. `make check`, `go vet`, `go test ./...`,
and the full opt-in live suite all pass; M1.1/M1.2/M2.1–M2.5/M2.4a/M3.4/M3.5/M3.6 all PASS
on fresh live evidence (see Status table for per-task citations). M3.1–M3.3 are corroborated
(the human eyeball-judgment step can't be independently replayed by an agent).
**M3.3a FAILS**: its own cited case still leaks. A fresh live regen of `major_arcana.md`
still shows a sister-project line (`Media related to Major Arcana at Wikimedia Commons`)
under `## External links` — Wikipedia renders this instance as an inline `Commons category`
template inside a plain `<li>`, not `div.portal-bar`/`div.sistersitebox`, so `strip.go`'s
M3.3a selectors don't catch it. `strip_test.go`'s M3.3a regression guard never actually
checks Major Arcana (only the other two titles), so this shipped with a green test suite.
Filed as TRIAGE T10 (the leak itself, medium) + T12 (the test-coverage gap that let it
ship, low) — needs a new fix-up task via replan before M4.1 proceeds, since M4.1 depends
on M3.3a. Also found and filed T11 (low): image placeholders can leak citation-marker
digits (`[1][2]`) from figcaptions because `renderImages` runs before ref-superscript
stripping — doesn't fail any mechanical check but conflicts with S3a's intent.
Status table: M3.3a flipped from `done` to `blocked` (on T10); Validated column stamped
2026-09-03 for every task this pass covered (all except M4.1/M4.2, not yet built).
Next step: charm-replan drains T10/T11/T12 (and pre-existing T8/T9) into a fix-up task
(e.g. M3.3b) before M4.1 can proceed — M4.1's `needs:` list still cites M3.3a as done.

## Status

| Key  | Status  | Reviewed | Validated | Note                                        |
|------|---------|----------|-----------|---------------------------------------------|
| M1.1 | done    | 2026-09-03 | 2026-09-03 | `cmd/wikispike/`; re-verified post-refactor, no leakage into noggin |
| M1.2 | done    | 2026-09-03 | 2026-09-03 | live fetch verified, all four S5 titles, fresh re-run |
| M2.1 | done    | 2026-09-03 | 2026-09-03 | fresh regen: headings/prose/table present   |
| M2.2 | done    | 2026-09-03 | 2026-09-03 | fresh regen clean except the leak now tracked under M3.3a/T10 |
| M2.3 | done    | 2026-09-03 | 2026-09-03 | fresh regen: no link syntax/URLs on any of the four |
| M2.4 | done    | 2026-09-03 | 2026-09-03 | fresh regen: no 3+ blank-line runs, no empty list items; done-when's dangling-paren gap closed by M2.4a |
| M2.5 | done    | 2026-09-03 | 2026-09-03 | shared live fixture confirmed, whole-suite `-run Live` passes without 429 |
| M3.1 | done    | 2026-09-03 | 2026-09-03 | original scratch root off-sandbox, user-attested; independently reproduced equivalent at /tmp/s5scratch |
| M3.2 | done    | 2026-09-03 | 2026-09-03 | human eyeball-judgment step corroborated, not independently re-playable |
| M3.3 | done    | 2026-09-03 | 2026-09-03 | corroborated: nothing to retune beyond M3.4's carve-out |
| M3.4 | done    | 2026-09-03 | 2026-09-03 | fresh live regen: Unicode + raw-TeX fallback both present, no mid-sentence gaps |
| M3.5 | done    | 2026-09-03 | 2026-09-03 | fresh regen: placeholders at every image position, balanced brackets; T11 filed (citation-marker leak, doesn't fail this criterion) |
| M2.4a | done   | 2026-09-03 | 2026-09-03 | fresh regen: list_languages.md now 386/386 `(`/`)`, 3/3 `[`/`]`; T8/T9 (cross-paragraph masking) remain open, distinct known limitation |
| M3.3a | blocked | 2026-09-03 | 2026-09-03 | FAILS re-verification: major_arcana.md still leaks a sister-project line via an inline-template variant the fix's selectors miss — see TRIAGE T10/T12 |
| M3.6  | done   | 2026-09-03 | 2026-09-03 | doc-comment, `transformHTML`, and "placeheld" typo fix all confirmed |
| M4.1 | todo    | -        | -         | widened for M3.4/M3.5 report + TRIAGE T4    |
| M4.2 | todo    | -        | -         |                                             |
