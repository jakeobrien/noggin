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

Current task: M4.2

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

**2026-09-03, later still yet again (charm-build) — M4.1 re-closed; cursor moves to M4.2.**
M4.1 was re-opened by charm-replan (D111) after M3.9 fixed the maintenance-template/`{{rp}}`
leak M4.1's own first pass had surfaced but left open (TRIAGE T1). Re-verified and updated
`cycle/findings/m4.1-conversion-record.md` throughout: header note and "Sample provenance"
now cite a fresh post-M3.9 regen (`/tmp/s5scratch`, same command/paths as before) instead of
the stale M3.8-era one; the mechanical re-check list now explicitly confirms no
maintenance-template marker or `{{rp}}` remnant anywhere; the fixed leak moved from "What
converts badly" into "What converts well" (a new "Citation-like superscripts" bullet,
keeping the Parsoid quirk detail — the maintenance-template family's visible `<sup>` carries
no `typeof` at all, unlike every other chrome category here — for a future cycle); the
selector-list entry, "Retune-exhaustion" narrative, and "Open items for replan" line were
all brought current (TRIAGE T1 marked closed, nothing left to route). Independently
re-verified rather than trusting M3.9's own report: regenerated all four S5 samples fresh,
grepped for every leak pattern (`citation needed`/`clarification needed`/`dubious`/
`disputed`, and a `[^0-9]: [0-9]+ ` rp-remnant pattern) with none found, checked balanced
`(`/`)`/`[`/`]` counts and no 3+ blank-line runs/empty items on all four, confirmed no
dangling `## External links`-style emptied heading on Major Arcana (M3.7's fix still holds),
and reran `make check` plus `WIKISPIKE_LIVE=1 go test ./cmd/wikispike -run Live -v` — all
clean. `findings/INDEX.md`'s summary line updated to match. Next step: charm-build picks up
M4.2 — log the go/no-go/retune-exhausted verdict in `DECISIONS.md`, citing M4.1's now-current
findings. TRIAGE.md is empty; no blockers.

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
| M3.3a | done    | 2026-09-03 | -         | re-reviewed post-M3.3b fix-up; portal-bar/sistersitebox selectors confirmed clean on fresh regen of all four samples; T1/T2/T3/T5/T6/T7 filed against the surrounding strip.go changes |
| M3.6  | done   | 2026-09-03 | 2026-09-03 | doc-comment, `transformHTML`, and "placeheld" typo fix all confirmed |
| M3.3b | done    | 2026-09-03 | -         | fix-up on M3.3a (TRIAGE T10/T12): text-pattern heuristic (`sisterProjectInlineLine`) over `<li>`/`<p>` whole-text catches the inline Commons-category leak; `TestStripChromeLiveSamples` now checks Major Arcana specifically; review found a new regression (T1, empty `## External links` heading) plus T2/T3/T5/T6/T7 — T1/T3 now M3.7, T2 decided (D110), T4/T5/T6/T7 now M3.8 |
| M3.5a | done    | 2026-09-03 | -         | fix-up on M3.5 (TRIAGE T11): `renderImages` strips `sup[typeof*="mw:Extension/ref"]` from an image node before reading its caption; fresh live regen of all four S5 titles shows no `[<digit>]` ref-marker pattern anywhere; review confirmed independently, filed T4 (minor doc-comment overstatement) — T4 now folded into M3.8 |
| M3.7 | done    | 2026-09-03 | 2026-09-03 | fix-up on M3.3b (TRIAGE T1/T3): `dropEmptyHeadings` drops any section heading emptied by chrome removal; fresh live regen confirms `major_arcana.md` no longer dangles `## External links`; sister-project mid-sentence-survival fixture added |
| M3.8 | done    | 2026-09-03 | 2026-09-03 | housekeeping pass 2 (TRIAGE T4/T5/T6/T7): doc-comment fix, `removeSisterProjectInlineLines` extracted, `elementText` reused in `imagePlaceholderText`, GLOSSARY.md Chrome entry updated |
| M3.9 | done    | -        | -         | fix-up on M2.2/M3.3a (TRIAGE T1, D111): `selectorCitationLikeSup` now also matches the maintenance-template family and the `{{rp}}` page-locator template; fresh regen of all four S5 samples clean |
| M4.1 | done    | -        | -         | re-closed post-M3.9: findings record updated + independently re-verified against a fresh regen; TRIAGE T1 fixed and confirmed gone |
| M4.2 | todo    | -        | -         |                                             |
