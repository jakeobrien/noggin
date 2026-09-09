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
| T24 | review | low | cmd/noggin/store.go:82 | `storedRoot`'s doc comment says "toggleStore is the only writer besides a future rename (M7.2)" — M7.2 landed in this same diff and `renameStore` (store.go) is now a second writer; "future" is stale. Update to state both writers plainly (no tense implying M7.2 is still pending). | open |
| T25 | review | low | cmd/noggin/store.go:231 (also .charm/cycle/PROGRESS.md M7.1 handoff paragraph) | `escapeStoreField`/`unescapeStoreField`'s doc comment cites D61 ("the two stores are deliberately separate formats (D61)") to justify not sharing parsing code with docstore.go, but D61 is "Document memory is one global, path-keyed store — shadow Roots rejected" — unrelated to keeping the two stores' escaping code separate. The actual format decisions are D21 (Roots store: line-based text file) and D67 (Document store: tab-separated fields); neither states the two shouldn't share code. Either correct the citation to D21/D67 or drop the D-number and state the reasoning as new (D61 does not cover it). | open |
| T26 | review | high | cmd/noggin/location.go:79-93 (docContentHeight/peekContentHeight), cmd/noggin/location.go:145-155 (composeViewerContent) | M8.2's location line is drawn whenever a document is open regardless of whether boxHeight actually has a spare row for it: docContentHeight/peekContentHeight floor at `max(boxHeight-1, 0)` instead of suppressing the line, so at a pathologically short terminal (verified via `sized(t, newModel(), 80, 4)` then `openDocument`, giving boxHeight 0) composeViewerContent still appends `"\n"+row`, overflowing the pane's height budget by 1-2 rows and corrupting the whole frame (DV1's own bottom border disappears, panes misalign) — reproducible, not covered by any existing test. computeLayout's own staleRow already shows the project's precedent for graceful degradation (`staleH := min(staleRowHeight, remaining)`, never overflowing); the location line should follow the same pattern — e.g. gate docLocationVisible/peekLocationVisible (or composeViewerContent) on boxHeight >= 1 rather than only flooring the reduced height at 0. | open |
| T27 | review | low | cmd/noggin/viewer.go:601 (also footer.go:168, keymap.go:15, files_search.go:18,283) | scrollActions' doc comment cites "the help modal (S12)" — carried through M8.1's own edit to this exact comment — but this cycle's SPEC.md now assigns S12 to M8.1's own "Half-page paging" item, the very spec this diff implements, so the bare, undisambiguated "(S12)" now collides with a different current-cycle meaning right next to the code that delivers it. The same bare "(S12)" (apparently meaning Help Modal, from an older cycle) recurs at footer.go:168, keymap.go:15 and files_search.go:18/283 — pre-existing, not touched by this diff, but now equally ambiguous. picker_view.go:192 already shows this project's own disambiguation convention ("S12/C2"); the Help Modal references should be re-cited the same way (e.g. "S12/C\<n\>") or have the stale S-number dropped where it no longer resolves unambiguously against the current SPEC.md. | open |
| T28 | review | low | cmd/noggin/location.go:79-93 | docContentHeight and peekContentHeight are identical apart from which visibility method they call (`docLocationVisible` vs `peekLocationVisible`); now that both variants exist side by side, consider one shared helper (e.g. `contentHeight(visible bool) int`) called by each one-line wrapper, to remove the duplication. | open |

<!-- empty as of 2026-09-08 replan (third pass): T1 -> PLAN M1.11 (checks.py:937 vacuous
     Roots/Sources assertion); T2 -> PLAN M2.10 (not-parallel-safe doc note on M2.9's
     var-mutating tests); T3 -> PLAN M1.11 (merged with T1, same file, same string-drift
     class); T4 -> PLAN M5.3 (rootLabel comment gap). All four were code-review findings
     against M5.1/M5.2/M1.9/M2.9, reconfirmed accurate at the preceding validate pass; none
     blocking, none material enough for a DECISIONS.md entry — same low/medium-risk hygiene
     shape as the M1.9/M1.10/M2.9/M3.6/M3.7/M4.4 tasks the prior replan pass already added.
     Earlier passes: T1(orig)/T22 -> M2.9; T2(orig) -> M1.10; T18 -> M1.9 (diagnosis
     corrected — see DECISIONS.md); T20 -> M3.6; T21 -> M3.7; T23 -> M4.4. -->
