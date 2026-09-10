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

<!-- empty as of 2026-09-10 replan (eighth pass): T36 (review finding — M8.2a's own
     regression test only exercises the Active-Document path, docLocationVisible/
     docContentHeight; the Peek path, peekLocationVisible/peekContentHeight, covered by the
     same boxHeight>=1 gate and shared contentHeight helper and named explicitly in M8.2a's
     own task text, has no boxHeight-0 regression test of its own) -> PLAN M8.2b, same
     test-gap convention this cycle already used for T31/T34. Accepted rather than deferred
     (too small/mechanical to survive the next prune as its own ROADMAP line) or wontfixed
     (not a declined idea, just coverage left out of M8.2a's own narrower AC); no
     DECISIONS.md entry — low severity, not material, and risk is already low since
     peekLocationVisible reads the same v.boxHeight field as the doc path (viewer.go:88-99).

     earlier passes: seventh (2026-09-10): T35 (M1.12/M9.3 build commit d1ced0f
     falsely stamped Reviewed=2026-09-10 on 7 unrelated already-done tasks — M1.11, M2.11,
     M7.4, M8.2a, M8.4, M9.2, M8.5 — with no review of them ever performed) -> confirmed
     accurate against git history; fixed directly in PROGRESS.md (reset the 7 cells to `-`,
     not a PLAN task since this is tracking-doc correctness, not code work) and logged in
     DECISIONS.md D137 (medium — process integrity, not blocking). Not deferred (it is a
     one-time correction, nothing ongoing to schedule) and not wontfixed (the finding was
     right, not a declined idea).

     earlier passes: sixth (2026-09-10): T33 (stale "...Roots differs from Files
     and Document" wording in a checks.py:784 description string, missed by both M1.9's and
     M1.11's rename sweeps since the assertion itself isn't vacuous) -> PLAN M1.12, same
     rename-sweep convention as M1.9/M1.11. T34 (M9.2's regression test doesn't cover the
     component-value-0 case its own doc comment claims protection for) -> PLAN M9.3, same
     test-gap convention as M2.9-M2.11. Both accepted rather than deferred (too small/
     mechanical to survive the next prune as their own ROADMAP line) or wontfixed (not
     declined ideas, just coverage/wording gaps left by the tasks that introduced them); no
     DECISIONS.md entry for either — low severity, not material.

     earlier passes: fifth (2026-09-09): T32 (M8.4 fixed only the 5 bare `(S12)`
     sites its own AC named; the same now-ambiguous bare `S12` recurs at ~20 more sites, a
     mix of C1's help-modal sense and C2's query-threshold/result-cap sense) -> PLAN M8.5,
     same disambiguation convention as M8.4. Accepted rather than deferred to ROADMAP (too
     small/mechanical to survive the next prune as its own line) or wontfixed (not a
     declined idea, just leftover scope from M8.4's own narrowed AC); no DECISIONS.md entry
     — low severity, not material.

     earlier passes: fourth (2026-09-09): T24+T25 -> PLAN M7.4 (stale/mis-cited store.go doc
     comments — "future (M7.2)" writer wording, wrong D-number citation). T26 (real bug:
     location line overflows a pathologically short terminal) merged with T28 (dedup nit,
     same functions) -> PLAN M8.2a. T27 (ambiguous bare "(S12)" citation, now collides with
     M8.1's own S12) -> PLAN M8.4. T29 (real bug: tintPaneBackground's SGR-7 detector
     misreads a colour-index/truecolor component of 7 or 27 as reverse video) -> PLAN M9.2.
     T30 (M1.11's checks.py "Roots"/"Sources" fix missed one occurrence at line 164) dropped,
     not promoted — already covered verbatim by M1.11's own still-open acceptance criterion.
     T31 (duplicated not-parallel-safe doc comment across two wikipedia test files) -> PLAN
     M2.11.

     see PLAN.md task notes and git history for the T1-T23 mapping (third replan pass,
     2026-09-08). -->
