# Triage — C4a wikipedia-spike

<!-- Per-cycle triage INBOX for actionable items surfaced out-of-band: code-review and
     validate findings, bugs / in-scope issues discovered during build that were not
     anticipated in the plan, work deliberately skipped per the escalation matrix (high,
     non-blocking, work-shaped), and the consequences of plan-invalidating discoveries.
     NOT a backlog and NOT a decision log (judgment questions go to DECISIONS.md).

     Only charm-replan parses items in full; every other role reads this inbox at summary
     level (count, severities, items referenced by blocked tasks).

     Lifecycle: consume-and-clear. Items are added `open`, then triaged by charm-replan per
     the escalation matrix — promoted to a PLAN.md task (with an acceptance criterion),
     deferred beyond this cycle into ROADMAP.md Unscheduled, recorded as a DECISIONS.md entry
     (wontfix), or dropped with a reason. Once triaged, the item is removed here (the outcome
     lives in PLAN / ROADMAP / DECISIONS). This doc stays small by construction; if it is
     growing, that means triage is overdue. It must be empty before a cycle can close.

     Keep one line per item. Ids (T1, T2…) are transient scratch labels, not durable keys —
     they are only for referring to an item until it is triaged, and may be reused after
     the inbox is cleared. Cite like findings: filepath:line. Severity is the reporter's
     estimate, not ground truth — triage filters false positives and wontfix.

     Sources: review | validate | build | docs | user. Status: open | triaged. -->

| Id | Source | Severity | Citation | Item | Status |
|----|--------|----------|----------|------|--------|
| T8 | review | medium | cmd/wikispike/cleanup.go:79-92,141-159 | `dropUnmatchedClosers` (M2.4a) tracks paren/bracket depth across the whole flushed prose chunk (only split at fence boundaries in `cleanupMarkdown`), not per removal/paragraph — a single unmatched opener anywhere earlier in the chunk permanently raises the depth baseline for everything after it in that chunk, silently masking any later genuinely-dangling closer instead of dropping it. Verified: `dropUnmatchedClosers("An unmatched (opener earlier in the doc.\n\nLater a stray closer) appears alone.\n", '(', ')')` leaves the downstream stray `)` untouched. D107 documents only the half-open-interval limitation (`[0, 1)`), not this materially more likely cross-paragraph masking case; the mechanical acceptance check (aggregate open/close count balance over the whole document) can't detect this failure mode either, since a masked pair can still net out balanced. No existing test (`cleanup_test.go`) exercises two independent, non-adjacent parenthetical groups in one chunk, so the gap shipped uncaught. Recommend scoping the depth counter per blank-line-delimited paragraph (or otherwise narrowing chunk scope) rather than the whole flushed prose block, and add a regression test with an unrelated unmatched opener preceding a genuinely dangling closer; if the wider scope is accepted anyway, document it explicitly in `dropUnmatchedClosers`'s doc comment and in D107 as a second known limitation. | open |
| T9 | review | low | .charm/DECISIONS.md:1345-1354 (D107) | D107's "known limitation" writeup covers only the half-open-interval (`[0, 1)`) case; T8's cross-paragraph masking case is a distinct, more likely limitation of the same `dropUnmatchedClosers` design and isn't mentioned there. If T8 is triaged as wontfix/accepted-as-is, D107 should be amended to record it as a second known limitation so M4.1's findings don't overstate the fix's reliability. | open |
| T10 | validate | medium | cmd/wikispike/strip.go:48-52; strip_test.go:98-112 | M3.3a's done-when ("none of the four current converted samples shows a leaked portal-bar or sister-project line") FAILS on independent re-verification: a fresh live regen of `major_arcana.md` still shows `- \`[Image: Wikimedia Commons logo]\` Media related to Major Arcana at Wikimedia Commons` under `## External links` — the exact sample/line PLAN.md cited as the motivating case. Root cause: Wikipedia renders this as an inline `Commons category` template inside a plain `<li>` in the External-links list, not as `div.portal-bar`/`div.sistersitebox` (confirmed absent from the live Parsoid HTML via direct curl) — `strip.go:48-52`'s selectors only catch the boxed-template variant. Shipped unnoticed because `strip_test.go`'s M3.3a regression guard (`strip_test.go:98-112`) never checks Major Arcana at all — only "Why is there anything at all?" (portal-bar) and a fixture-only Commons string on "E (mathematical constant)" that can't match real live text. Needs either a broader selector/heuristic (e.g. matching the External-links list-item text pattern "Media related to ... at Wikimedia Commons"/"... at Wikiquote") or a different removal strategy, plus a regression test that actually exercises Major Arcana. | open |
| T11 | validate | low | cmd/wikispike/images.go:57; sample why_anything.md:3 | Image placeholders can leak citation-superscript reference markers (e.g. `[Image: ... Parmenides (c. 515 BC).[1][2]]`) because `renderImages` reads `figcaption.Text()` before `stripChrome` removes `sup[typeof*="mw:Extension/ref"]`. Doesn't trip any mechanical check (brackets stay balanced) and doesn't fail M3.5's literal acceptance criterion, but conflicts with S3a's "citation superscripts... removed" intent and is untested. Fix: run ref-superscript removal (or a scoped strip) before/within `renderImages`'s caption extraction, or filter `[\d]` ref-marker patterns out of the caption text. | open |
| T12 | validate | low | cmd/wikispike/strip_test.go:98-112 | Test-coverage gap adjacent to T10: the M3.3a live regression guard checks only 2 of the 4 S5 titles and never Major Arcana, the one PLAN.md explicitly cites as showing the sister-project leak — so a real regression in exactly the cited case shipped with a passing test suite. Should be folded into T10's fix (add Major Arcana to the guard) rather than tracked separately if T10 is accepted. | open |


