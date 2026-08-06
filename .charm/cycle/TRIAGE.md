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
| T3 | review | low | `cmd/noggin/files.go:150-163` (`refreshRoot`) | Refresh (M5.1) never prunes `filesNav.expanded` entries for paths that no longer exist on disk — deliberate (harmless dead weight, per its own doc comment), but across a long session with heavy churn under a large Root, repeated refreshes can only ever grow the map, never shrink it. No observable misbehavior; flagging as an unbounded-growth code smell in case a future session cares about long-running memory. | open |
| T4 | review | low | `cmd/noggin/refresh_test.go` | M5.1's `keys.Refresh` tests only drive the binding with focus on `filesPane`/`viewerPane`; no test exercises it with focus on `rootsPane` specifically, unlike M2.1's own every-pane table (`focus_test.go`'s `TestFocusBindingsMoveFocusFromEveryPane`). Low risk since `Update`'s dispatch of `keys.Refresh` is unconditional on `m.focus` (`model.go`), but a spec-literal ("from any pane") gap the project's own precedent would normally cover explicitly. | open |
| T5 | review | low | `cmd/noggin/refresh_test.go` | No test configures more than one Root and then refreshes, to pin S2's negative clause ("nothing else is refreshed": other Roots and their broken marks untouched) the way a positive assertion would. True by construction today (`model.refresh()` only ever touches `m.files`/`m.viewer`), but not regression-proofed against a future refactor that widens `refresh()`'s reach. | open |
| T6 | build | low | `README.md:195-249` ("Finding across files (the Picker)" section) | M6.2 added the `alt+h` dotfile-hiding toggle (S4); README's Picker section still only says dotfiles/dot-directories are included by default, with no mention that this can be toggled. M9.1 already owns the Footer/Help Modal half of S4's binding sweep, but no plan task owns README the way it owns those — same gap T1 already flags for `ctrl+p`, now with a second instance. | open |
