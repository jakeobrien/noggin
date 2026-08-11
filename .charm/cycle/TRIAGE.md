# Triage — C4 recall

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
| T38 | build | low | cmd/noggin/files.go:159 (setRoot) | Switching the active Root back onto one left in a flat mode can land selection at 0 instead of the nearest still-valid row: openFolder calls filesNav.setRoot (its own clampSelected runs against the *outgoing* root's still-stale activeRootPath, computing 0 rows for the just-restored mode and zeroing selection) before model.refreshFilesRecentSource re-points activeRootPath and re-clamps against the correct list — by then the original Tree-restored index is already lost to 0, not re-derived. Never out-of-range (still a valid selection), so no crash/no-op; found while building M4.14 (TestSettingRootRestoringAFlatModeReClampsSelection, flat_selection_clamp_test.go, deliberately isolates activeRootPath to avoid this and test setRoot's own restore logic cleanly instead). | open |
| T39 | review | low | cmd/noggin/browse_list_test.go:117-121 (TestRecencyAndStarredNameTiesBreakByPath) | M4.13's acceptance wording (PLAN.md M4.13, D87) literally asked for "a one-line comment" here; the comment landed as 4 lines. Functionally fine (accurate, matches the project's own norm of thorough narrative comments) and the code/behavior are correct — flagging only for the same literal-acceptance-wording standard this cycle applies to itself elsewhere (e.g. M4.14/T37). Recommend: either amend M4.13's PLAN wording after the fact to match what was actually delivered, or leave as a wontfix DECISIONS.md note — no code change needed. | open |
| T40 | validate | low | cmd/noggin/browse.go:69-78, browse_list.go:159-186, browse_starred_test.go:76-96 | `Orphans` scope (SPEC.md pre-agreed cut #3; PLAN.md M4.3 note `*Orphans is pre-agreed cut 3.*`) is in fact fully implemented, reachable via `keys.FilesScope`, and covered by dedicated tests (`TestStarredRowsOrphansScopeOnlyShowsOrphans`, the six-combo empty-state test, browse-memory round-trip) — it was never actually cut. SPEC.md/PLAN.md wording still reads as if dropped, with no flag comparable to M2.5's "pre-agreed cut 1, not actually cut" PROGRESS note. Doc-only; no code change needed. Recommend a PLAN.md/SPEC.md wording fix or a PROGRESS.md note at next replan. | open |

