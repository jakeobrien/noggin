# Triage — C1 read-across-roots

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

<!-- T37-T43 (the 2026-08-02 code review of M5.4/M5.5's own diff) were routed by
     replan C1/2026-08-03 — T37 and T43 to a new PLAN.md task M5.6 (correctness), T38-T41 to
     a new PLAN.md task M5.7 (standards/test-hardening), T42 fixed directly in PLAN.md's own
     M5.4 bullet text (a documentation accuracy fix, not code); see D29. T28-T36 from the
     M3.5/M5.1-M5.3 backlog were routed to M5.4/M5.5, ROADMAP.md Unscheduled (T31), or dropped
     (T36) by replan C1/2026-08-02; see D27. T44 (validate) and T45/T46 (review, from M5.6's
     own diff) were routed by this replan (2026-08-03) — T44 dropped as superseded by T45's
     investigation, T45 and T46 bundled into a new PLAN.md task M5.8; see D30. -->
