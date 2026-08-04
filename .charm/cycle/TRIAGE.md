# Triage — C2 find

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
| T25 | review | low | .charm/cycle/PROGRESS.md:92-93 | M6.1's Status-table note reads "pre-agreed cut (D32); D57" with no qualifier, while M6.2's reads "pre-agreed cut (D32) not taken; D57" and the Handoff correctly states D32 was not taken for M6 as a whole (both tasks were built). Read alone, M6.1's row misleadingly implies the cut was applied to it. Recommendation: reword M6.1's note to match M6.2's "not taken" phrasing (or state it once at the M6 milestone level) so the Status table doesn't contradict the Handoff. | open |

<!-- Inbox empty: T1-T4, T21-T24 triaged 2026-08-04 (charm-replan). Routing:
     T1  -> PLAN M7.1b(a) (accepted, low-cost housekeeping, bundled)
     T2  -> PLAN M7.1b(b) (accepted, low-cost housekeeping, bundled)
     T3  -> PLAN M7.1b(c) (accepted, low-cost housekeeping, bundled)
     T4  -> PLAN M7.1b(d) (accepted, low-cost housekeeping, bundled)
     T21 -> PLAN M4.2a, D55 (accepted, material; bundled with T22)
     T22 -> PLAN M4.2a, D55 (accepted, material; bundled with T21)
     T23 -> D56 (decided: keep test, add rationale comment via PLAN M7.1b(e))
     T24 -> dropped: already covered by queued PLAN M7.1a, whose acceptance criterion ("no
            comment in the Picker's code uses either variant") is codebase-wide, not a
            point-in-time count — no separate action needed; noted on M7.1a for traceability

     Inbox empty: T6-T12 triaged 2026-08-03 (charm-replan). Routing:
     T6  -> PLAN M3.2a, D48 (accepted, material)
     T7  -> PLAN M3.3a, D49 (accepted, batch)
     T8  -> PLAN M3.3a, D49 (accepted, batch)
     T9  -> PLAN M3.3b, D49 (accepted, batch)
     T10 -> PLAN M3.1a, D50 (accepted, scope judgment)
     T11 -> PLAN M7.1a (accepted, low-cost housekeeping)
     T12 -> PLAN M3.2b (accepted, low-cost housekeeping)

     T13-T20 triaged 2026-08-03 (charm-replan). Routing:
     T13 -> PLAN M3.1b, D52 (accepted, material; bundled with T15, T16)
     T14 -> PLAN M2.1a, D51 (accepted, material)
     T15 -> PLAN M3.1b, D52 (accepted, bundled with T13)
     T16 -> PLAN M3.1b, D52 (accepted, bundled with T13)
     T17 -> PLAN M3.1b (accepted, low-cost, folded into T13's fix since same function)
     T18 -> PLAN M1.4c (accepted, low-cost; bundled with T19)
     T19 -> PLAN M1.4c (accepted, low-cost; bundled with T18)
     T20 -> dropped: reporter's own report called it a defensible deliberate choice (avoids
            OS-dependent err.Error() text brittleness) with no functional gap; low risk/value,
            no decision entry warranted per the escalation matrix -->
