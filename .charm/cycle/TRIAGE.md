# Triage — C2a polish

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
| T19 | docs | low | `README.md:87-98` (Global keybinding table) | While fixing M11.1's `ctrl+p`/`alt+h`/Peek gaps, found the Global table also never lists `r`/`f` (FocusRoots/FocusFiles, S6) or `R` (Refresh, S2) — both shipped in earlier milestones (M2.1, M5.1) with no README task ever owning them, the same "no plan task owns README" shape as T1/T6/T15. Left out of M11.1 since its acceptance criterion named only `ctrl+p`/`alt+h`/`p`. | open |

<!-- All items triaged 2026-08-09. Routed to: PLAN.md M7.3 (T13/T14/T16), M8.2 (T18),
     M11.1 (T1/T6/T15), M12.1 (T2/T4/T5/T7/T17); fixed directly (T8: D46 addendum; T10:
     stray D58 citation dropped); DECISIONS.md wontfix/documented (D50: T12, plus T7's
     coverage note; D52: T3/T9/T11). See PROGRESS.md Handoff for the full mapping. -->
