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
| T41 | build | low | cmd/noggin/keymap.go (HistoryBack/HistoryForward), footer.go, help.go | M6.1's new `keys.HistoryBack`/`keys.HistoryForward` (`[`/`]`) are not yet listed in the Document Viewer's Footer hints or Help Modal binding list — the same discoverability gap M2.8/M4.8 later closed for `keys.Star` and `keys.FilesMode`/`keys.FilesScope` after their own introducing milestones. Not part of M6.1's acceptance criterion; functionality is complete and tested (history_test.go) without it. | open |
| T42 | review | medium | cmd/noggin/files_search.go:439 (grepFiles), cmd/noggin/picker.go:545-550 (pickerSearchCmd) | M5.1's split of `grepFiles` into `grepPickerFiles` left `grepFiles` itself dead in production: `pickerSearchCmd` now resolves `files` itself (`starredPickerFiles` or `listFiles`) and calls `grepPickerFiles(files, query)` directly, so `grepFiles` has no remaining production caller — it is kept alive only by `files_search_test.go`'s pre-existing tests, which were not updated to call `grepPickerFiles(listFiles(...), query)` directly instead. `grepFiles`'s own doc comment (files_search.go:443-449, "grepFiles (above) is one caller, resolving files the roots-based way (listFiles) for Active Root/All Roots") is now factually wrong — that caller no longer exists. Recommend either deleting `grepFiles` and repointing its tests at `grepPickerFiles(listFiles(...), query)`, or, if a thin roots-based wrapper is still wanted for some future caller, correcting the doc comment and wiring it back into `pickerSearchCmd`'s non-Starred branch. | open |
| T43 | review | low | cmd/noggin/history_test.go:219-240 (TestSameDocumentNeverAdjacentInHistory), .charm/DECISIONS.md D91 | D91 explicitly flags a visible M6.2 behavior change — re-Activating the document already open in a viewer (no history motion) now preserves its scroll position instead of `open()`'s previous unconditional `GotoTop()` — as "untested today". `TestSameDocumentNeverAdjacentInHistory` re-activates the same document twice but never scrolls it first, so it exercises the history-stack side of same-document re-Activate without covering the scroll-preservation side D91 calls out. Recommend a small regression test (mirroring history_scroll_test.go's style): scroll a document, re-Activate the same one with no navigation in between, and assert the offset survives rather than resetting to top. | open |

