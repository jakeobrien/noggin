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

<!-- empty as of 2026-09-09 replan (fourth pass): T24+T25 -> PLAN M7.4 (stale/mis-cited
     store.go doc comments — "future (M7.2)" writer wording, wrong D-number citation). T26
     (real bug: location line overflows a pathologically short terminal) merged with T28
     (dedup nit, same functions) -> PLAN M8.2a. T27 (ambiguous bare "(S12)" citation, now
     collides with M8.1's own S12) -> PLAN M8.4. T29 (real bug: tintPaneBackground's SGR-7
     detector misreads a colour-index/truecolor component of 7 or 27 as reverse video) ->
     PLAN M9.2. T30 (M1.11's checks.py "Roots"/"Sources" fix missed one occurrence at line
     164) dropped, not promoted — already covered verbatim by M1.11's own still-open
     acceptance criterion (M1.11 flipped back to in-progress by the prior validate pass; no
     new task needed for a clause the existing task already owns). T31 (duplicated
     not-parallel-safe doc comment across two wikipedia test files) -> PLAN M2.11. All eight
     items were review findings against M2.10/M7.1-M7.2/M8.2/M9.1, none blocking, none
     material enough for a DECISIONS.md entry — same low/medium hygiene-or-narrow-bugfix
     shape as every prior batch this cycle.

     earlier passes: see PLAN.md task notes and git history for the T1-T23 mapping (third
     replan pass, 2026-09-08). -->
