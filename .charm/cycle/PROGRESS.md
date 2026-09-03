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

**2026-09-03, later same day (charm-build) — M3.5a done: renderImages now strips citation
superscripts from captions before reading them, cursor moves to M4.1.**
Root cause: `renderImages` (M3.5) read `figcaption.Text()` before `stripChrome`'s citation-
superscript removal ran (pipeline order: math → images → strip), so a ref marker sitting
inside a caption (e.g. Parmenides' "(c. 515 BC).[1][2]" on `why_anything.md`) leaked straight
into the placeholder text. Fixed by having `renderImages` remove `sup[typeof*="mw:Extension/
ref"]` from the image node itself (mutating the live node, not a clone — it is either about
to be replaced wholesale or left for stripChrome to remove outright, so the mutation is
harmless either way) before computing the placeholder description. Factored the selector
out to `htmldoc.go` (`selectorCitationSup`/`typeofRefExtension`) alongside the existing
file/math constants so `stripChrome`'s own removal reuses the same selector instead of a
parallel string literal. Added `TestRenderImagesStripsCitationSuperscriptFromCaption`
(unit, no live fetch) plus a fresh live regen of all four S5 titles confirming no `[Image:
...]` placeholder — or anywhere else in any sample — carries a leftover `[<digit>]`
ref-marker pattern (`grep -nE '\[[0-9]+\]' *.md` on a fresh /tmp scratch regen: no matches).
Verified: `make check` clean, full opt-in live suite (`WIKISPIKE_LIVE=1 go test ./cmd/
wikispike -run Live -v`) passes. Next step: charm-build picks up M4.1, now unblocked — all
of its `needs:` are done (M2.4a, M3.3, M3.3a, M3.3b, M3.4, M3.5, M3.5a, M3.6).

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
| M3.5a | done    | -        | -         | fix-up on M3.5 (TRIAGE T11): `renderImages` strips `sup[typeof*="mw:Extension/ref"]` from an image node before reading its caption; fresh live regen of all four S5 titles shows no `[<digit>]` ref-marker pattern anywhere |
| M4.1 | todo    | -        | -         | widened for M3.4/M3.5 report + TRIAGE T4, and now M3.3b/M3.5a; all `needs:` now done |
| M4.2 | todo    | -        | -         |                                             |
