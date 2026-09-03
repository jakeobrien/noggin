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
| T2 | build | medium | `cmd/wikispike/strip.go` `stripChrome`; found via M3.5, `.charm/cycle/findings/images-experiment.md` | `stripChrome` (M2.2) doesn't match Wikipedia's portal-bar (`div.portal-bar`) or sister-project side-box (`div.sistersitebox`) chrome — both navigational, not article content. Pre-existing (predates M3.5); was silently masked because their icon images vanished without a trace, now visible as leaked list items/`[Image: icon]` placeholders (`why_anything.md`'s Philosophy/Physics/Astronomy/Religion/Science line, `e_math.md`/`major_arcana.md`'s Wikimedia Commons/Wikiquote lines). Not fixed as part of M3.5 (out of scope: chrome selection is M2.2's/M3.3's job). | open |
| T3 | review | medium | `tmp/wikispike-tests-1/list_languages.md:190` (`- Scala (into JVM bytecode))`, unbalanced paren, 386 `(` vs 388 `)` in file); noted but left unfixed in `.charm/cycle/findings/maths-experiment.md:67` and `.charm/cycle/findings/images-experiment.md:54` | M2.4's own done-when ("no unmatched `[`/`(` left by a removal, checked mechanically over all four") is not met — a genuine dangling `)` survives, a class `cleanupProse`'s literal-empty-pair regex (`cmd/wikispike/cleanup.go:59`, `reEmptyParens`) doesn't catch. Known to the builder (mentioned in two findings docs) but never routed here until now, and PROGRESS.md's M2.4 row is marked `done` unqualified. Recommend: either fix (`cleanupProse` needs a real bracket/paren-depth pass, not literal-empty-pair matching) or requalify M2.4's done note and cite this item. | open |
| T4 | review | low | `.charm/cycle/PROGRESS.md` handoff entries (M3.4/M3.5 fix-ups) claim "all four S5 `.md` samples regenerated" after each fix-up | Root-level sample copies (`why_anything.md`, `e_math-2.md`, `e_math-3.md`, `major_arcana.md`) are stale relative to the shipped pipeline — some predate the M3.4 display-math and M2.2 sidebar fix-ups, some predate the M3.5 backtick-wrap fix-up (`7154497`); only `tmp/wikispike-tests-1/*.md` (dated 2026-09-03 10:10) reflects every fix-up. Recommend: when M4.1 cites sample output, point at `tmp/wikispike-tests-1/` explicitly (or wherever the current off-sandbox scratch root ends up) and delete/replace the stale root copies so a future reader doesn't diff against them by accident. | open |
| T5 | review | low | `cmd/wikispike/links_test.go:33-36` | Doc comment above `TestStripLinksPreservesSurroundingProse` still reads `// TestArticleMarkdownHasNoLinkSyntax runs the full pipeline's HTML stages...` — stale from a rename; every other test file in the package keeps its leading comment identifier in sync with `func`. Recommend fixing the comment's identifier to match. | open |
| T6 | review | low | `cmd/wikispike/strip.go:18-20`,`84-86`; `links.go:17-19`,`33-35`; `math.go:44-46`,`65-67`; `images.go:35-37`,`52-54` | Identical goquery parse/re-serialize boilerplate (`NewDocumentFromReader` + error wrap, `doc.Html()` + error wrap) duplicated verbatim across all four DOM-transform files (~24 duplicated lines); a shared helper (e.g. `withGoqueryDoc(html, verb, fn)`) would centralize it and reduce drift risk as the spike is retuned. Related: the `mw:File`/`mw:Extension/math` `typeof` prefixes are separately hard-coded as string literals in both the strip selector and the corresponding render* selector (`strip.go:81` vs `images.go:40`; `strip.go:81` vs `math.go:49`) with no shared constant, despite comments describing them as "the same selector" — a future retune could silently desync them. Recommend a shared parse/serialize helper plus shared `const` for the two typeof prefixes. | open |
| T7 | review | low | `cmd/wikispike/images.go:54` | Error message `"rendering image-placeheld HTML: %w"` — "placeheld" is an awkward/typo coinage vs. sibling messages' `-ed`-noun pattern (`strip.go:86` "rendering stripped HTML", `links.go:35` "rendering delinked HTML"). Recommend "rendering image-placeholder HTML". | open |

<!-- T1 triaged 2026-08-14 (charm-replan): promoted to PLAN.md task M2.5 (needs M2.2-M2.4,
     effort 2), ahead of M3.3's retune iterations which would otherwise keep re-tripping the
     429. See DECISIONS.md D103. -->
