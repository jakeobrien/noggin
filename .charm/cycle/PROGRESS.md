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

Current task: M3.5a

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

**2026-09-03, later same day (charm-build) — M3.3b done: broadened chrome-strip heuristic
fixes the Major Arcana leak, cursor moves to M3.5a.**
Root cause: Major Arcana's leak is a Commons category-inline template rendered as a bare
`<li>` in External links (sibling `<span>`/`<a>` sharing one Parsoid `about` id, no wrapping
div/table for a CSS selector) — confirmed via a live HTML dump. Fixed with a text-pattern
heuristic (`sisterProjectInlineLine`, `strip.go`) matching the whole trimmed text of any
`<li>`/`<p>` against known sister-project phrasings, so passing mentions survive. Gotcha:
Parsoid inlines the template's `<style data-mw-deduplicate=...>` CSS as a text-node sibling
in the same `<li>`; goquery's `.Text()` doesn't skip style/script like the pipeline's own
html-to-markdown step does, so a naive call broke the whole-string anchor match — fixed with
`elementText` (clone, strip style/script, then read text); unit fixture extended to cover
this trap without a live fetch. T12 folded in: `TestStripChromeLiveSamples` now asserts on
Major Arcana specifically. Verified: `make check`/`go vet` clean; full opt-in live suite
passes including the new assertion; fresh `go run` regen shows `## External links` ending
cleanly, content table intact. Next step: charm-build picks up M3.5a (T11's citation-marker
leak in image placeholders), then M4.1 can proceed — its `needs:` list is now satisfiable.

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
| M3.3a | done    | -        | -         | unblocked: M3.3b's broadened heuristic fixes the cited leak (fresh live regen confirmed); Reviewed/Validated cleared, `strip.go` changed since the 2026-09-03 validation — needs re-covering |
| M3.6  | done   | 2026-09-03 | 2026-09-03 | doc-comment, `transformHTML`, and "placeheld" typo fix all confirmed |
| M3.3b | done    | -        | -         | fix-up on M3.3a (TRIAGE T10/T12): text-pattern heuristic (`sisterProjectInlineLine`) over `<li>`/`<p>` whole-text catches the inline Commons-category leak; `TestStripChromeLiveSamples` now checks Major Arcana specifically |
| M3.5a | todo    | -        | -         | fix-up on M3.5 (TRIAGE T11): strip ref-superscripts before `renderImages` reads captions |
| M4.1 | todo    | -        | -         | widened for M3.4/M3.5 report + TRIAGE T4, and now M3.3b/M3.5a |
| M4.2 | todo    | -        | -         |                                             |
