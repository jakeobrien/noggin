# Triage — C6 sources

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

<!-- empty as of 2026-09-08 replan (third pass): T1 -> PLAN M1.11 (checks.py:937 vacuous
     Roots/Sources assertion); T2 -> PLAN M2.10 (not-parallel-safe doc note on M2.9's
     var-mutating tests); T3 -> PLAN M1.11 (merged with T1, same file, same string-drift
     class); T4 -> PLAN M5.3 (rootLabel comment gap). All four were code-review findings
     against M5.1/M5.2/M1.9/M2.9, reconfirmed accurate at the preceding validate pass; none
     blocking, none material enough for a DECISIONS.md entry — same low/medium-risk hygiene
     shape as the M1.9/M1.10/M2.9/M3.6/M3.7/M4.4 tasks the prior replan pass already added.
     Earlier passes: T1(orig)/T22 -> M2.9; T2(orig) -> M1.10; T18 -> M1.9 (diagnosis
     corrected — see DECISIONS.md); T20 -> M3.6; T21 -> M3.7; T23 -> M4.4. -->
