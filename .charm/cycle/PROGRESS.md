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

**2026-09-03, later still yet again (charm-build) — M3.9 done; cursor moves to M4.1.**
Fixed TRIAGE T1 (D111): `cmd/wikispike/htmldoc.go` gained `selectorCitationLikeSup`,
unioning the pre-existing `selectorCitationSup` (`sup[typeof*="mw:Extension/ref"]`) with two
new selectors discovered by fetching the live Parsoid HTML directly and inspecting the
actual nodes: `sup.Inline-Template` for the maintenance-template family (`{{citation
needed}}`, `{{clarification needed}}`, `{{dubious}}`, `{{disputed}}` — these carry *no*
`typeof` on the visible sup at all; Parsoid's `typeof="mw:Transclusion"` marker lands on a
preceding invisible `<link rel="mw:PageProp/Category">` sharing the same `about` id instead)
and `sup[typeof="mw:Transclusion"]` exact-match for the `{{rp|<page>}}` page-locator
template (whose own sup *does* carry that typeof directly; a real citation ref's typeof is
always the compound `"mw:Transclusion mw:Extension/ref"` so this exact match never
double-fires). `strip.go`'s main removal and `images.go`'s M3.5a caption-cleanup call both
switched to the combined selector. Added a unit fixture (`strip_test.go`) plus live
regression guards in `TestStripChromeLiveSamples` for the three affected titles (Major
Arcana, List of programming languages by type, Why is there anything at all?). Fresh live
regen of all four S5 samples to `/tmp/s5scratch` confirms: no `citation needed`/
`clarification needed`/`dubious`/`disputed` text anywhere, no bare `{{rp}}` remnant (checked
with a `[^0-9]: [0-9]+ ` pattern), balanced `(`/`)` and `[`/`]` counts on all four, no 3+
blank-line runs, no empty list items, and the surrounding prose reads cleanly across each
removal point (spot-checked, e.g. Major Arcana's "...firmly established in the minds of
occult practitioners."). `WIKISPIKE_LIVE=1 go test ./cmd/wikispike -run Live -v` and
`make check` both clean. Updated `cycle/findings/m4.1-conversion-record.md`'s "known gap"
language (the leak entry, the selector-list line, and the "open items for replan" line) and
`findings/INDEX.md`'s summary line to reflect the fix, per M3.9's own acceptance criterion.
Next step: charm-build picks up M4.1 (re-opened by charm-replan, D111) — it needs a fresh
regen citation reflecting this fix (the current text still narrates the M3.8-era regen that
predates M3.9) before it can close again and hand off to M4.2.

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
| M4.1 | todo    | -        | -         | re-opened (was done): needs a fresh regen citation reflecting M3.9's fix (was: consolidated findings record + fresh S5 regen + TRIAGE T4 stale-copy cleanup; surfaced TRIAGE T1, now fixed) |
| M4.2 | todo    | -        | -         |                                             |
