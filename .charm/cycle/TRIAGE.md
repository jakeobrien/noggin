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
| T1 | build | low | `README.md:96,134,204` | M2.2 retired `ctrl+p`/bound `space` to `keys.Picker`; README still documents `ctrl+p` opening the Picker in three places (bindings table, section heading, prose) — needs a doc-sync pass (no plan task currently owns README, unlike M9.1's Footer/Help Modal scope). | open |
| T2 | review | low | `test/pty/checks.py`, `cmd/noggin/focus_test.go` | M2.1's `r`/`f` global focus bindings (FocusRoots/FocusFiles) have thorough `go test` coverage but no `test/pty` check — unlike the Navigator toggle (`ctrl+n`), which has both. Per Makefile's own comment, pty checks are "the only way to verify what actually reaches a terminal" (border/focus highlight actually moving); the go tests only assert `m.focus`/internal state. Recommend adding `r`/`f` entries to `test/pty/checks.py`'s `KEY` table plus a check that the focused pane's border highlight moves, mirroring the existing Navigator-toggle checks. | open |
