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
| T1 | build | low | `cmd/wikispike/strip.go` (`selectorCitationSup`) | `stripChrome` only matches `typeof*="mw:Extension/ref"` citation superscripts; Wikipedia's maintenance-template family rendered as a `typeof="mw:Transclusion"` `<sup>` (`{{citation needed}}`, `{{clarification needed}}`, `{{dubious}}`, `{{disputed}}`, `{{rp\|<page>}}`) is a distinct template mechanism and leaks as literal text — found live in 3/4 S5 samples during M4.1's fresh regen: `major_arcana.md` (`[*clarification needed*]`, `[*dubious – discuss*]`), `list_languages.md` (`[*citation needed*]`), `why_anything.md` (a bare `: 17 ` from `{{rp\|17}}`, no brackets at all). Doesn't fail any mechanical check (brackets stay balanced, since these templates carry their own literal `[`/`]` text or none). See `.charm/cycle/findings/m4.1-conversion-record.md` for detail. | open |

<!-- Emptied 2026-09-03 by charm-replan: T8/T9 -> DECISIONS.md D108 (wontfix-for-this-cycle);
     T10/T12 -> PLAN.md M3.3b; T11 -> PLAN.md M3.5a. See PROGRESS.md handoff and PLAN.md's
     2026-09-03 reconciliation note for detail. -->

<!-- Emptied 2026-09-03, later still, by charm-replan: T1/T3 -> PLAN.md M3.7 (DECISIONS.md
     D109); T2 -> DECISIONS.md D110 (decided, not a bug); T4/T5/T6/T7 -> PLAN.md M3.8 (no
     decision entry needed, same precedent as M3.6's own T5/T6/T7 batch). See PLAN.md's
     2026-09-03, later still reconciliation note for detail. -->


