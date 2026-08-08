# Triage — C3 compare

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
| T4 | build | high | footer_test.go:202 (`TestNoFooterHintNamesCtrlP`) | M1.4a's D55 rebind (PeekDV2 → `ctrl+p`) will make filesPane's footer hints contain the literal string "ctrl+p" (`paneHints` already lists `keys.PeekDV2` there), which this pre-existing C2a/M2.2 invariant test ("no footer hint may ever name ctrl+p again", a retired unrelated binding) forbids outright — M1.4a's plan doesn't mention this test; it will fail unless scoped/adjusted. | open |

<!-- T1 (review, .gitignore:2, low): fixed directly this session per the effort-scaling rule
     (one-line, low-risk) rather than routed through a plan task — /cmd/noggin/noggin added
     to .gitignore.
     T3 (build, high, shift+enter/L Picker fallback): triaged → DECISIONS.md D55 (ctrl+enter +
     ctrl+l, non-printable fallback closes it structurally) → PLAN.md M1.5 (updated).
     T2 (review, low, FocusDV2/resolveFocus edge case): triaged → DECISIONS.md D56 (left
     unpinned, same reasoning as D54). -->
