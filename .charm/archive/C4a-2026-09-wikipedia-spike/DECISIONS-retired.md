# Retired decisions — C4a wikipedia-spike

Retired at close (2026-09-03) per `.charm/DECISIONS.md`'s retention sweep: each entry's
subject is scoped to a now-complete plan task with no ongoing constraint, its substance is
captured in the promoted `knowledge/wikipedia-*.md` files or in the code/tests it explains,
or (D104) it is the historical rationale for a mid-cycle scope widening already realized in
the shipped experiments and D112's verdict. Moved here, never deleted.

## D99: Wikimedia domains needed allow-listing in the sandbox

- Status: decided
- Cycle: C4a
- Date: 2026-08-14
- Context: Found while planning. `https://en.wikipedia.org/api/rest_v1/...` and
  `https://api.wikimedia.org/...` both return `Blocked by network policy: no matching allow
  rule`. The Go module proxy is reachable, so dependencies are not affected. Every task from
  M1.2 onward needs a live article; nothing after M1.1 can be verified without it.
- Decision: user allow-listed `en.wikipedia.org` on 2026-08-14. Verified: REST HTML returns
  200 for all four S5 sample titles (including the `?` and parenthesised ones), and Go's
  `net/http` reaches it through the sandbox proxy with no TLS workaround. M1.2 unblocked.
- Why not routed around while blocked: hand-saved HTML fixtures would have made M1.2's
  acceptance criterion a lie and would not have exercised the real API's output; per the
  sandbox policy, ask rather than work around.

## D100: The spike is a separate `main`, not a mode or flag of the noggin binary

- Status: decided
- Cycle: C4a
- Date: 2026-08-14
- Context: S1 requires the spike be unreachable from the TUI and to not change noggin's
  behaviour in any way. A hidden flag on `cmd/noggin` would put experimental code on the
  shipped binary's path and make "does not change behaviour" an argument rather than a grep.
- Decision: its own `main` package alongside `cmd/noggin`, with M1.1's acceptance criterion
  being a mechanical no-reference check from `cmd/noggin/` and `internal/`.
- Why: makes S1's isolation requirement machine-checkable, and makes OQ1's "drop the wrapper,
  keep the conversion" a directory deletion rather than a disentangling job.

## D101: The live-API test is opt-in; `make check` stays hermetic

- Status: decided
- Cycle: C4a
- Date: 2026-08-14
- Context: M1.2's acceptance criterion is a real fetch of all four S5 titles, but `make check`
  is the per-change loop and must not depend on the network (or on D99's allow-list surviving).
- Decision: `TestFetchLiveSampleTitles` skips unless `WIKISPIKE_LIVE=1`; the hermetic tests
  cover the User-Agent (S8) and title escaping (S2) against an `httptest` server. Acceptance
  is demonstrated by running the live test explicitly:
  `WIKISPIKE_LIVE=1 go test ./cmd/wikispike -run Live -v`.
- Why: keeps the fast loop deterministic while still exercising the real API on demand; a
  network-dependent `make check` would fail for reasons unrelated to the change under test.

## D102: The spike's User-Agent contact is operator-supplied, not hard-coded

- Status: decided
- Cycle: C4a
- Date: 2026-08-14
- Context: S8 requires a descriptive User-Agent per Wikimedia policy, which asks for a way to
  contact whoever is running the tool. The repo carries no canonical contact or public URL,
  and inventing one would be a false claim in a header sent to a third party.
- Decision: `noggin-wikispike/0.1 (<contact>)`, where `<contact>` comes from
  `WIKISPIKE_CONTACT` and defaults to "local experimental spike, not distributed".
- Why: honest and policy-shaped either way, and it keeps the choice out of code for a tool
  that S1 says is throwaway. Revisit if any Wikipedia capability is ever distributed (OQ2).

## D103: TRIAGE T1 (live-test 429s) promoted to a plan task rather than left ad hoc

- Status: decided
- Cycle: C4a
- Date: 2026-08-14
- Context: charm-replan triage of TRIAGE T1 — each of M1.2/M2.2/M2.3/M2.4 added its own
  opt-in live test that independently re-fetches all four S5 titles; running the whole
  `-run Live` suite together trips the REST API's 429 (individual `-run <TestName>` runs
  pass, confirmed again during this cycle's M1/M2 validation pass). `make check` is
  unaffected (D101 — live tests are opt-in and hermetic-excluded).
- Decision: accepted as PLAN.md task M2.5 (consolidate to one shared per-title live
  fixture), rather than left as an ad hoc fix-when-convenient or deferred to ROADMAP. Not
  wontfix: M3.3 is explicitly a repeatable retune-and-re-judge task, and repeated live-test
  runs during retuning are exactly what will keep re-tripping this.
- Why: medium severity but directly reachable in the very work still ahead in this cycle
  (M3.3's repetition), not a one-off inconvenience already behind us — worth a scoped task
  now rather than risking it recur silently through every retune round. Not deferred beyond
  the cycle: the fix is cheap (effort 2) and the underlying test code may not survive past
  this cycle at all (OQ1/D97), so parking it on the roadmap for a future cycle to pick up
  would likely never be acted on.

## D104: The C4a retune budget grows to include two costed experiments — maths Unicode, image placeholders

- Status: decided
- Cycle: C4a
- Date: 2026-09-02
- Context: reading the four sample articles surfaced a "swiss-cheese" effect from missing
  maths and images — exactly the risk A5 anticipates as a known limitation, not an automatic
  no-go, but real enough that the user wants it addressed before recording the verdict.
  Wikimedia's Parsoid HTML already exposes an image's caption/alt/filename and a formula's raw
  TeX source for free (no OCR, no new fetch); and glamour, noggin's existing render pipeline,
  already turns `![alt](src)` into a styled text placeholder — so both experiments turned out
  to be far cheaper than the original spec's blanket exclusions assumed.
- Decision: two experiments added to the retune budget (S6a, superseding S6), scoped as
  research/costing, not committed feature work — a later cycle with its own agreed spec still
  decides what ships:
  - **Maths**: try a Unicode textual approximation first (new dependency permitted), falling
    back to the already-permitted raw-TeX text preservation if Unicode conversion is not a
    small-ish lift.
  - **Images**: replace the stripped image with a plain-text placeholder derived from its
    caption, falling back to alt-text, falling back to filename; no clickable link (S3a,
    superseding S3). Try the near-zero-cost path (stop deleting the image node, let the
    existing render pipeline carry it) before writing new extraction logic.
  Findings (S7a, superseding S7) must record what was tried, effort spent, any dependency
  added, resulting quality, and a carry-forward recommendation for each. Out of scope's
  "rendering images"/"rendering maths" — not wanted lines are superseded accordingly; full
  visual/graphical rendering of either remains excluded.
- Why: this is genuinely new capability being added mid-cycle, not a relocated one — flagged
  explicitly rather than absorbed quietly as "just a retune", per the amend process. The user
  chose to keep it in this cycle (not defer to the roadmap) because it converts the cycle's
  deliverable from a bare verdict into "a verdict plus a costed feature set", material input
  for whichever future cycle actually builds Wikipedia support. Both experiments stay
  bounded by the existing retune budget's cheapest-path-first order and builder's-judgement
  stop rule (S6/OQ4), now explicitly extended to cover them.

## D105: TRIAGE T2 (stripChrome misses portal-bar/sistersitebox) promoted to task M3.3a

- Status: decided
- Cycle: C4a
- Date: 2026-09-02
- Context: charm-replan triage of TRIAGE T2 — M3.5's live regeneration surfaced a
  pre-existing M2.2 gap: `stripChrome` doesn't match Wikipedia's portal-bar
  (`div.portal-bar`) or sister-project side-box (`div.sistersitebox`) chrome. Both are
  navigational, not article content, and were already leaking as bare text before M3.5;
  M3.5's own image-placeholder work made the leak visible (`[Image: icon]` lines) instead
  of silent, but the root cause predates and is unrelated to M3.5.
- Decision: accepted as PLAN.md task M3.3a (needs M2.2, M3.5; effort 2) rather than left ad
  hoc or deferred to ROADMAP — the same category of fix as the sidebar gap already
  fixed directly this session (M2.2's `table.sidebar`), but routed through TRIAGE/replan
  since it was already filed rather than caught in the moment.
- Why: medium severity and directly relevant to A3 ("no large gaps or artifacts") — the
  leaked lines are visible defects in the very samples M4.1's findings will cite. Cheap
  (effort 2, same selector-list pattern as the existing sidebar fix) and the underlying
  code may not survive the cycle (OQ1/D97) — worth fixing now rather than parking it where
  a future cycle would need to rediscover it. Not wontfix: it directly contradicts S3a's
  "removal must not leave visible holes" contract.

## D106: TRIAGE T3 (M2.4's dangling-paren gap) promoted to task M2.4a

- Status: decided
- Cycle: C4a
- Date: 2026-09-02
- Context: charm-replan triage of TRIAGE T3 — a code-review pass found a genuine dangling
  `)` in `tmp/wikispike-tests-1/list_languages.md` (386 `(` vs 388 `)`) that M2.4's own
  done-when ("no unmatched `[`/`(` left by a removal, checked mechanically over all four")
  claims does not occur. Root cause: `cleanupProse`'s `reEmptyParens` only catches literal
  empty pairs, not a real orphaned bracket left after a removal. Known to the builder
  (mentioned in two findings docs) but never routed through triage until this session, and
  PROGRESS.md's M2.4 row is marked `done` unqualified.
- Decision: accepted as PLAN.md task M2.4a (needs M2.4; effort 2) — give `cleanupProse` a
  real paren/bracket-depth pass (or a targeted fix for the specific removal that orphans
  it) rather than requalify M2.4's done note and move on. M2.4's PROGRESS.md row stays
  `done` (the milestone shipped) but this fix-up must land before M4.1's findings are
  written, since findings should describe the pipeline's actual current behavior.
- Why: medium severity, a real defect in already-`done` work that contradicts that task's
  own machine-checkable acceptance criterion — not a matter of judgment, a bug. Cheap
  (effort 2) and directly reachable by M4.1's own mechanical re-check; fixing now avoids
  M4.1 either citing a known-broken sample or having to caveat around it.

## D107: M2.4a — depth-based dangling-closer pass, asymmetric (closers only)

- Status: decided
- Cycle: C4a
- Date: 2026-09-03
- Context: building M2.4a, re-investigated D106's root cause. Neither `list_languages.md`
  dangling `)` traces to a removal — both are already present in the live article's own
  wikitext (`Scala (into JVM bytecode))`, `RUNCOM (... CTSS) programs)`), pre-dating this
  pipeline touching the article at all. D106's "targeted fix for the specific removal that
  orphans it" option therefore doesn't apply; its other option ("a real paren/bracket-
  depth pass") does and is what shipped, as `dropUnmatchedClosers` (`cleanup.go`).
- Decision: the pass tracks parens and brackets as two independent depth counters and
  drops a closer when its counter is already at zero, but deliberately never removes an
  opener left unmatched at end-of-chunk — asymmetric by design. Accepted trade-off: a
  legitimate mixed-bracket idiom like a half-open interval "[0, 1)" would read as an
  unmatched `[` and an unmatched `)` and lose the `)`; no sample in this cycle's S5 set
  contains that idiom (verified). Documented in `dropUnmatchedClosers`'s doc comment, not
  filed to TRIAGE — hypothetical, not observed, and the low-risk asymmetric choice is
  preferable to the alternative (also removing unmatched openers, which would silently
  delete more real text on the one shape actually observed: a stray opener is a visible,
  honest artifact rather than a deleted character).
- Why: meets M2.4a's mechanical done-when (all four current S5 samples now have balanced
  `(`/`)` and `[`/`]` counts — verified via a live regenerate + count, and via a new live
  test) with the smaller, more conservative of the two viable fixes; keeps M4.1's findings
  accurate about root cause (an upstream source quirk, not solely a pipeline artifact).

## D108: TRIAGE T8/T9 (dropUnmatchedClosers cross-paragraph masking) — wontfix for this cycle

- Status: decided
- Cycle: C4a
- Date: 2026-09-03
- Context: code review (TRIAGE T8) found that `dropUnmatchedClosers` (M2.4a/D107) tracks
  paren/bracket depth across the whole flushed prose chunk, not per paragraph — an unmatched
  opener anywhere earlier in a chunk permanently raises the depth baseline, silently masking
  a later genuinely-dangling closer instead of dropping it, demonstrated with a synthetic
  two-paragraph example. T9 separately noted D107's known-limitation writeup covers only the
  half-open-interval case (`[0, 1)`), not this one. Neither is observed in any of the four
  live S5 samples — only reproduced synthetically.
- Decision: not fixed this cycle. `dropUnmatchedClosers` keeps its current whole-chunk depth
  scope, documented as a second known limitation alongside D107's half-open-interval case:
  an earlier unmatched opener in the same flushed chunk can mask a later genuinely-dangling
  closer, and the mechanical balanced-count acceptance check cannot detect it since a masked
  pair still nets out balanced. T8/T9 are dropped from TRIAGE without a PLAN.md task.
- Why: medium severity per the reporter, but risk × value is low in practice — not observed
  in any sample this cycle's S5 set actually produces, only in a hand-constructed synthetic
  case, and the fix (narrowing chunk scope to per-paragraph) is more invasive than M2.4a's
  own conservative asymmetric design (D107) chose to be, for a spike whose own survival past
  this cycle is still an open question (OQ1). Revisit if a future cycle keeps the conversion
  logic and a real sample surfaces this pattern; until then, the documented limitation is a
  cheaper and more honest outcome than either a rushed fix or leaving it undocumented.
## D109: TRIAGE T1/T3 (M3.3b's own regression: emptied External-links heading, missing safety-property test) promoted to task M3.7

- Status: decided
- Cycle: C4a
- Date: 2026-09-03
- Context: code review of M3.3b found (T1, high) that its own inline-sister-project removal
  deletes the sole `<li>` in Major Arcana's External-links section, leaving `## External
  links` as a dangling, content-free heading through EOF — verified by fresh live regen.
  This contradicts M3.3b's own "done when" wording ("no leaked... line", implicitly not "a
  new hole instead") and directly violates S3a ("removal must not leave visible holes") and
  A2 (headings/sections legible and parsable); neither M2.4's mechanical hole checks nor
  M3.3b's own regression test look for an emptied section heading. Separately (T3, medium),
  no test exercises `sisterProjectInlineLine`'s own stated safety property — that an
  ordinary sentence merely mentioning a sister project in passing survives untouched — only
  the removal path is tested.
- Decision: accepted as PLAN.md task M3.7 (needs M3.3b; effort 2) rather than requalifying
  M3.3b's done note and moving on — drop a section heading left with no content before the
  next same-or-higher-level heading/EOF (scoped to whatever `stripChrome` empties, not just
  External links), and add the missing safety-property fixture.
- Why: T1 is a real defect in already-`done` work that contradicts the cycle's own
  acceptance bar (S3a, A2) — the same category and remedy as D106's dangling-paren gap, not
  a matter of judgment. Cheap (effort 2) and directly reachable by M4.1's own findings,
  which should not have to cite or caveat around a known-broken sample. T3 is bundled in
  since it is the same code area and a natural companion assertion, not because it alone
  would warrant a task.

## D110: TRIAGE T2 (sisterProjectInlineLine's whole-document scope) — decided, not a bug

- Status: decided
- Cycle: C4a
- Date: 2026-09-03
- Context: code review found M3.3b's `sisterProjectInlineLine` scan runs over every
  `<li>`/`<p>` in the whole document, not scoped to the External-links section the way
  PLAN.md's M3.3b task text describes ("a text-pattern heuristic over External-links list
  items") — a silent scope widening versus the task wording, though likely harmless in
  practice since the match is a full-line anchor.
- Decision: keep the whole-document scope as shipped; it is not a defect. PLAN.md's task
  history is left unedited (never rewrite plan history) — this entry is the record of what
  actually shipped and why it's accepted.
- Why: low risk in practice (a full-line anchor rarely false-matches ordinary prose, and T3
  above adds a regression test for exactly that safety property) and actually the more
  robust choice — sister-project mentions are not guaranteed to sit only in an
  External-links section across arbitrary articles, and scoping the `Find` to that section
  would need its own heading-boundary detection logic for no clear benefit. Narrower than
  medium value to fix, so decided rather than turned into a task.

## D111: TRIAGE T1 (mw:Transclusion maintenance-template citation leak) promoted to task M3.9

- Status: decided
- Cycle: C4a
- Date: 2026-09-03
- Context: M4.1's own fresh regen (a findings task, not a fix-up) surfaced a new leak:
  `selectorCitationSup` (`sup[typeof*="mw:Extension/ref"]`) matches reference-citation
  superscripts but not Wikipedia's separate maintenance-template family — `{{citation
  needed}}`, `{{clarification needed}}`, `{{dubious}}`, `{{disputed}}`, and the page-locator
  `{{rp|<page>}}` — which renders as a distinct `typeof="mw:Transclusion"` `<sup>`. Leaks as
  literal bracketed text (or, for `{{rp}}`, an unbracketed page number) in 3 of the 4 S5
  judging samples (`major_arcana.md`, `list_languages.md`, `why_anything.md`). Doesn't fail
  any mechanical check (brackets stay balanced), which is why it survived every prior
  mechanical re-check and was only caught by eyeball during M4.1.
- Decision: accepted as PLAN.md task M3.9 (needs M3.3a; effort 2) rather than a roadmap
  deferral or a documented known limitation — extend the citation-superscript
  selector/removal to also match the maintenance-template family, add a regression fixture,
  and update `m4.1-conversion-record.md`'s "known gap" language once fixed. M4.1 is
  re-dependencied on M3.9 and reopened (`done` → `todo`) since its findings text currently
  describes the leak as unfixed.
- Why: this is the same category and remedy as D105/D106/D109's chrome-selector
  widenings — extending coverage of a citation-superscript removal S3a already commits to,
  not new capability, so it stays inside S6a's retune budget. Unlike D108's cross-paragraph
  masking (synthetic-only, medium severity, more invasive fix), this is observed live in 3
  of the 4 samples the verdict (M4.2) is about to judge, and the fix is cheap (extend an
  existing selector, same shape as M3.3a). Fixing it before the verdict is recorded keeps
  M4.1's findings honest about the sample it cites rather than caveating around a known,
  fixable leak in the very evidence the go/no-go call rests on.

