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

<!-- Empty as of this replan (2026-08-01). T9-T27 routed: T12, T13, T14 → PLAN.md M2.4 (T13
     also D23); T11 → PLAN.md M2.4 (drop-or-test the dead branch); T17, T18, T20 → PLAN.md
     M3.5; T9, T10, T16, T19, T21, T22, T23, T24, T25, T27 → PLAN.md M3.6; T15, T26 → folded
     into PLAN.md M4.2's acceptance criteria. Rationale for the batching in D24. -->
