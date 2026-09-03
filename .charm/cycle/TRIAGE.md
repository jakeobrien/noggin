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
| T1 | review | high | `cmd/wikispike/strip.go:103-107` | M3.3b's inline-sister-project removal deletes the sole `<li>` in Major Arcana's External-links section, leaving `## External links` as a dangling heading with zero content through EOF — verified by fresh live regen (`## External links` is the last line of the regenerated `major_arcana.md`, nothing beneath it). Violates S3a ("removal must not leave visible holes") and A2 (headings/sections legible and parsable); not caught by M2.4's mechanical hole checks (`cmd/wikispike/cleanup.go`'s `cleanupProse` only catches 3+ blank lines, empty list items, unmatched brackets — not an emptied section heading) or M3.3b's own regression test (`strip_test.go`'s `TestStripChromeLiveSamples` Major-Arcana block checks leaked text and table survival, not section structure). Recommend `cleanupMarkdown` gain a pass that drops a heading left with no content before the next same-or-higher-level heading/EOF, or have the chrome-removal step remove an emptied section's now-orphaned heading. | open |
| T2 | review | medium | `cmd/wikispike/strip.go:103-107` | M3.3b's `sisterProjectInlineLine` scan runs over every `<li>`/`<p>` in the whole document, not scoped to the External-links section as PLAN.md's M3.3b task text describes ("a text-pattern heuristic over External-links list items") — a silent scope widening versus the task wording. Likely harmless in practice (full-line anchor), but not called out as a deliberate broadening anywhere. Recommend either scoping the `Find` to the External-links/See-also section(s) or updating PLAN.md's task text to reflect the whole-document scope actually shipped. | open |
| T3 | review | medium | `cmd/wikispike/strip_test.go` (TestStripChromeRemovesNonArticleElements, ~line 11-38) | No test exercises `sisterProjectInlineLine`'s own stated safety property — that ordinary prose merely mentioning a sister project in passing (not a whole-line template match) survives untouched (claimed in `strip.go:17-19`). Only the removal path is tested. Recommend adding a fixture `<p>`/`<li>` with a sister-project phrase embedded mid-sentence and asserting it is *not* removed. | open |
| T4 | review | low | `cmd/wikispike/htmldoc.go:27-29` | Doc comment claims the shared `selectorCitationSup` lets `renderImages` "strip a leaked ref-marker digit out of a caption/**alt**" — but `alt` is a plain string attribute that cannot contain a nested `<sup>` element; only the figcaption case is real (correctly scoped in `images.go:35-44`'s own comment). Recommend dropping "/alt" from the htmldoc.go comment. | open |
| T5 | review | low | `cmd/wikispike/strip.go:61-62,103-107` | `stripChromeInPlace`'s doc comment ("holds stripChrome's actual selectors, factored out...") no longer fully describes the function after M3.3b added a non-selector, whole-text-regex removal step (`sisterProjectInlineLine` over `li, p`) alongside the CSS-selector removals. Consider updating the comment, or factoring the regex-based step into its own named helper (e.g. `stripInlineSisterProjectLinks`) for symmetry with the file's one-concern-per-block style. | open |
| T6 | review | low | `cmd/wikispike/images.go:63` vs `cmd/wikispike/strip.go:34-45` | `elementText` was added (M3.3b) specifically because goquery's `.Text()` doesn't skip `<style>`/`<script>` content that Parsoid can inline as a text-node sibling — but `imagePlaceholderText`'s caption extraction (`images.go:63`) still calls `s.Find("figcaption").First().Text()` directly, the same unguarded pattern. If a caption ever carries an inlined templatestyles `<style>` sibling (same failure class M3.3b just fixed elsewhere), its CSS would leak into the image placeholder. Recommend reusing `elementText` for the caption path too. | open |
| T7 | review | low | `.charm/GLOSSARY.md:23-27` | The "Chrome" glossary entry enumerates chrome types (navboxes, sidebars, infoboxes, edit-section links, citation superscripts, reference lists, metadata tables, images) but wasn't updated for the portal-bar / sister-project-box / inline-sister-project-link chrome categories M3.3a/M3.3b introduced as a standing category in PLAN.md and code comments. Recommend adding a phrase or forward reference so the glossary stays in sync. | open |

<!-- Emptied 2026-09-03 by charm-replan: T8/T9 -> DECISIONS.md D108 (wontfix-for-this-cycle);
     T10/T12 -> PLAN.md M3.3b; T11 -> PLAN.md M3.5a. See PROGRESS.md handoff and PLAN.md's
     2026-09-03 reconciliation note for detail. -->


