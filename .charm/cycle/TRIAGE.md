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
| T14 | build | low | .charm/GLOSSARY.md:77 | Navigator glossary entry still says "Roots Nav above" despite the C6 rename to Sources Nav recorded at lines 114-120 (self-inconsistency, doc-only, found while doing M1.6) | open |
| T15 | review | low | internal/wikipedia/fetch.go:52-77 | `NewClient` and `ensureInit` duplicate zero-value initialization (`&http.Client{Timeout: fetchTimeout}`, `make(map[string]Article)`) in two places that can drift; recommend `NewClient` build a zero `Client{}` and call `ensureInit()` so there is one init path | open |
| T16 | review | low | cmd/noggin/focus_badge_test.go:13-46 | `TestPaneFocusBadgesShowAtNormalWidthRegardlessOfFocus` asserts a rendered line contains `pane.title()`'s own return value, so it is tautological and cannot pin the literal "Sources" string (would pass unchanged even if the rename were reverted); M1.6's PROGRESS.md handoff cites this test as satisfying the "pane label reads 'Sources'" acceptance criterion, but the real hardcoded-literal check is `TestRenderPaneDropsTheBadgeNotTheTitleWhenTooNarrow` (same file, lines 60/68) — criterion is met, but the handoff's citation is misleading for future maintainers; recommend correcting the citation or adding a literal "Sources" assertion to the first test | open |
| T17 | review | low | cmd/noggin/docstore.go:157-201 | M1.7 escapes only the article `<title>` field against embedded tab/newline (correct per its acceptance criterion); the same tab-separated-format corruption risk exists in principle for path-like `<key>` fields (e.g. a file path containing a raw tab/newline byte) which are not escaped — pre-existing, not introduced by this diff, likely low real-world risk, but worth a deliberate wontfix/defer decision rather than leaving implicit | open |
