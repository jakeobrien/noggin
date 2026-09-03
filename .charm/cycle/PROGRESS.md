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

Current task: M3.9

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

**2026-09-03, later still yet (charm-replan) — triaged TRIAGE T1; cursor moves to M3.9.**
Routed the sole open TRIAGE item (T1, the `mw:Transclusion` maintenance-template citation
leak M4.1's own fresh regen surfaced) to a new PLAN.md fix-up task, M3.9 (needs M3.3a,
effort 2), rather than a roadmap deferral or wontfix — it's the same category of fix as
M3.3a/M3.3b's chrome-selector widenings (extending `selectorCitationSup`'s coverage, not
new capability), affects 3/4 of the judging sample, and is cheap. Logged as DECISIONS.md
D111. Since M4.1's own findings text currently records the leak as an unfixed known gap,
M4.1 is re-dependencied on M3.9 and reopened (`done` → `todo`) — its Reviewed/Validated
cells were already `-`, so nothing to clear. M4.2 is untouched (still gated on M4.1).
Next step: charm-build picks up M3.9 — extend the citation-superscript
selector/removal to also match the maintenance-template family, add a regression fixture,
fresh-regen all four S5 samples to confirm no leak remains, then update
`cycle/findings/m4.1-conversion-record.md`'s "known gap" language before M4.1 can be
re-closed and M4.2 picked up after it.

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
| M3.9 | todo    | -        | -         | fix-up on M2.2/M3.3a (TRIAGE T1, D111): maintenance-template citation leak |
| M4.1 | todo    | -        | -         | re-opened (was done): findings record needs updating once M3.9 lands (was: consolidated findings record + fresh S5 regen + TRIAGE T4 stale-copy cleanup; surfaced new TRIAGE T1) |
| M4.2 | todo    | -        | -         |                                             |
