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
| T1 | review | medium | test/pty/checks.py:937 | "M5.2 toggling hides the whole sidebar" asserts `"Roots" not in hidden_lines` — vacuous since M1.6 renamed the pane to "Sources"; the literal string "Roots" never appears in the UI now regardless of Navigator visibility, so this check passes unconditionally and no longer verifies that hiding the Navigator removes the Sources/Files labels. M1.9 fixed the other 9 stale-string assertions (confirmed via grep + a clean `make check-pty` run) but missed this one. Fix: assert `"Sources" not in hidden_lines` instead. | open |
| T2 | review | low | internal/wikipedia/fetch_test.go (TestClientFetchViaNewClientDelegatesToFetchAt), internal/wikipedia/search_test.go (TestClientSearchDelegatesToSearchTitles) | M2.9's new tests mutate package-level vars (`fetchBase`, `wikiBase`, `searchBase`) with a save/restore-on-defer pattern, the same hazard class the codebase already documents for `grepArgMaxBytes` (cmd/noggin/files_search.go:437-449: "no test in this package may call t.Parallel() while any test still mutates ... this way"). Neither new test's doc comment carries the equivalent not-parallel-safe warning. No active bug (no test in internal/wikipedia currently calls t.Parallel()), but the established convention for this exact pattern wasn't followed. Fix: add a short doc note at the var declarations or the two tests, mirroring files_search.go's. | open |
| T3 | review | low | test/pty/checks.py:398,490,551,1428 | check() description strings still read "...moves focus to Roots"/"...reaches Roots" even though M1.9 changed the assertions themselves to check for "Sources". Cosmetic only (doesn't affect pass/fail) but glossary-inconsistent with GLOSSARY.md's Sources Nav rename; could confuse whoever reads a future failure. Fix: reword the description strings to "Sources" to match the renamed pane and the assertions already checking for it. | open |
| T4 | review | low | cmd/noggin/browse_list.go:23-25 | `browseRow.rootLabel`'s own doc comment ("the covering Root's base name when orphan is false; meaningless (and left "") otherwise") doesn't mention the third case M5.2 introduced: for an article row (`article == true`), rootLabel holds the fixed `wikipediaLabel`, not a Root's base name. Not incorrect, just incomplete — the `article` field's own comment below explains it, but `rootLabel`'s comment doesn't cross-reference it. Fix: add one clause to rootLabel's comment noting the article case. | open |

<!-- empty as of 2026-09-08 replan (second pass): T1 -> PLAN M2.9; T2 -> PLAN M1.10; T18 ->
     PLAN M1.9 (diagnosis corrected: a real M1.3/M1.6 rename regression, not flakiness — see
     DECISIONS.md); T20 -> PLAN M3.6; T21 -> PLAN M3.7; T22 -> PLAN M2.9 (merged with T1, same
     gap-shape, same package); T23 -> PLAN M4.4. -->
