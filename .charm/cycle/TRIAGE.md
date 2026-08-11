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
| T32 | review | medium | cmd/noggin/help_test.go:44 (vs. cmd/noggin/help.go:89) | `TestHelpModalListsThePaneOwnBindings`'s `filesPane` case only asserts `"expand"`/`"collapse"`/`"star"` present; it cannot catch a regression that drops M4.8's `keys.FilesMode`/`keys.FilesScope` from `filesPane`'s `helpBindings`, since `keys.PickerMode`/`keys.PickerScope` (already listed unconditionally for `filesPane`) render the identical `"mode"`/`"scope"` text — add an identity-based assertion (matching `b.Help().Key`/`.Desc`, as `TestHelpBindingsListPickersOwnBindingsFromEveryPane` does) rather than a substring check. | open |
| T33 | review | medium | cmd/noggin/footer_test.go:150-165,197-213 (vs. cmd/noggin/footer.go:53) | No footer test asserts `"mode"`/`"scope"` actually appear in `filesPane`'s hint text; M4.8's touches to `TestPaneHintsNameFocusAndRefreshFromEveryPane`/`TestPaneHintsNamePeekOnlyInFilesPane` only bumped the truncation width so unrelated strings keep fitting. Unlike help.go, `paneHints`'s `pickerHints` (Picker-open overlay) never leaks `PickerMode`/`PickerScope` into `filesPane`'s footer hints, so this is a true untested addition — add a `TestPaneHintsNameFilesModeAndScopeOnlyInFilesPane` mirroring `TestPaneHintsNamePeekOnlyInFilesPane`'s shape. | open |
| T34 | review | low | cmd/noggin/browse.go:88-93 (default case, vs. M4.9's new `emptyTreeText` split) | `flatEmptyText`'s doc comment now says its `default` branch is an unreachable defensive fallback, but that fallback still returns `emptyFilesText` (the "no Root open" message), not `emptyTreeText` (the "open but empty" message M4.9 introduced) that a Tree-mode caller would actually want if the branch were ever hit — stale/misleading given the new split; correct the returned value or the comment's framing. | open |
| T35 | review | low | cmd/noggin/files.go:159-178 (`setRoot`), 305-335 (`withRecentSource`/`cycleMode`), 369-386 (`cycleScope`) | `clampSelected` (files.go:513) recomputes `rowCount()`/`flatRows()` on every mode/scope/root transition, and the subsequent render recomputes it again — compounds the existing "always recompute, never cache" design with an extra full doc-map scan per transition. Bounded by the 100/Root cap so low risk; consider a comment acknowledging the double-scan or reusing the already-computed result where call sites allow. | open |
| T36 | review | low | cmd/noggin/browse_list_test.go:112-140 (`TestRecencyAndStarredNameTiesBreakByPath`) | The `starredRows` tie-break assertion (docA before docB) is only correct because subdirectory literals `"a"`/`"b"` sort lexicographically, not because of any documented guarantee — add a one-line comment noting the ordering comes from the literal subdirectory names, not `t.TempDir()`, to remove ambiguity for a future reader. | open |
| T37 | review | low | cmd/noggin/flat_selection_clamp_test.go (whole file, vs. PLAN.md M4.7's acceptance wording) | M4.7's acceptance criterion names four re-clamp call sites needing coverage (`cycleMode`, `cycleScope`, `setRoot`'s flat-mode memory restore, un-starring the selected row out of Starred) but only two have dedicated regression tests (`TestCyclingToAShorterFlatListReClampsSelection` covers `cycleMode`; `TestUnstarringLastVisibleStarredRowReClampsSelection` covers the un-star path via `withRecentSource`). `cycleScope` shrinking the row set and `setRoot`'s memory-restore branch (files.go:141-147's own doc-comment calls this out) have no direct assertion on `f.selected` after the shrink — low risk since all four funnel through the shared `clampSelected()`, but a literal gap against the task's "done when" enumeration. Add two more cases exercising `cycleScope` and `setRoot`'s restore path. | open |
