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
| T20 | build | medium | cmd/noggin/footer.go:41 (paneHints), cmd/noggin/help.go:60 (helpBindings) | keys.FilesMode/keys.FilesScope (M4.1, S5's two Browse cycle keys) are not yet in filesPane's footer hints or Help Modal listing, unlike every other Files-Nav-only binding (Peek/ActivateDV2/PeekDV2) added alongside its own task, and unlike M2.8's precedent (Star's own dedicated discoverability task). Mirrors T5/T6's shape. | open |
| T21 | build | medium | cmd/noggin/model.go (togglePeekInto, now fixed via selectedFlatRow/togglePeekFlatRowInto) | Originally: Star/Activate/Peek/Edit all resolved the Files Nav's "selected document" via the tree-based selectedNode() regardless of Browse Mode, silently acting on the wrong document in Recent/Starred. D80 (M4.3) fixed Star/Activate; M4.5 fixed Peek the same way (togglePeekFlatRowInto, with a broken-row refusal mirroring Activate's) and confirmed Edit needs no separate fix — it always acts on the focused viewer's Active Document, never Files Nav selection, so it is already correct once Activate/Peek load the right document. Expand/Collapse are now explicit no-ops in flat mode (files.go). Every action T21 named is now resolved; kept open only for replan to confirm and clear. | open |
| T22 | build | low | cmd/noggin/files.go:17 (emptyFilesText), files.go rowsView's Tree branch | Tree mode's empty-state message ("No folder open.") is shared between two different reasons: no Root open at all, and a Root open whose visible rows happen to be zero (an empty folder, or every top-level entry gitignored). The wording is only accurate for the former; M4.4 gave Recent/Starred their own per-scope wording but left this pre-existing Tree conflation as-is (out of that task's scope). | open |
