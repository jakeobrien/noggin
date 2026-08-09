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
| T10 | build | med | test/pty/checks.py:1 | No pty coverage for `ctrl+x` (CloseDV2, S4/C3/M2.1) — M1.7 added pty coverage for the four C3 chords that existed at the time (ctrl+enter/ctrl+l/ctrl+p/ctrl+o) but CloseDV2 postdates it; unit coverage exists (close_dv2_test.go) but the real-binary pty path does not. | open |
| T11 | review | high | cmd/noggin/model.go:630 | M3.1's `refresh()`/`refreshViewer` (model.go:677) regression: the combining `switch` at model.go:651-658 unconditionally overwrites `m.message` with the "refreshed; … search cleared" drop text whenever `dv1Drop`/`dv2Drop` is non-empty, even when `refreshViewer` already set `m.message` to a read-failure ("cannot open …") — because `viewer.close()` also resets `findState` via `render()`, a doc whose read fails while it had a committed find always makes `dv1Drop`/`dv2Drop` non-empty too, so the read-failure message is silently discarded. This directly contradicts the function's own doc comment ("Either drop message still loses to a read-failure message above") and is a real regression from the pre-M3.1 single-viewer code, which guarded the drop message with `if m.message == ""` (that guard was dropped in the refactor). Verified by a scratch test (removed after review): deleting DV1's open, find-committed document then pressing Refresh reports `"refreshed; search cleared"` instead of naming the missing file. No existing test (refresh_test.go) covers a document that both fails to read and had a committed find. Fix: guard the message-combining switch in `refresh()` with `if m.message == ""` (or check it per-branch) the same way the pre-refactor code and `refreshViewer`'s own read-failure branch already do. | open |
| T12 | review | low | cmd/noggin/model.go:309 | `closeDV2` (S4/C3/M2.1) never calls `resetStaleBaseline`, so after closing DV2 the staleness baseline keeps tracking DV2's former Active Document's path until the next real read (openFolder/openDocumentInto/refresh) — an out-of-process edit to that now-invisible, no-longer-displayed file during that window still flips the combined `m.stale` signal true (verified with a scratch test, removed after review). Narrow and self-healing (any refresh clears it), but arguably at odds with S9's "reports … across those sources" where DV2 is no longer one of the sources once closed. Consider resetting/pruning the baseline's DV2 entry in `closeDV2` itself, or explicitly deciding (and documenting, e.g. via DECISIONS.md) that this transient false-positive window is acceptable. | open |
| T13 | review | low | cmd/noggin/model.go:651 | Duplicated message-combining logic: `refresh()`'s `switch { case dv1Drop != "" && dv2Drop != "": … }` (model.go:651-658) and `resize()`'s near-identical switch (model.go:1064-1071) differ only in the leading token (`"refreshed"` vs. `cause`) — same three-case shape, same `"dv1 %s, dv2 %s"`/`"%s"`/`"dv2 %s"` formats. Extract a small shared helper (e.g. `combineDrops(prefix, dv1Drop, dv2Drop string) string`) to keep the two call sites from drifting independently. | open |

<!-- T1 (review, .gitignore:2, low): fixed directly this session per the effort-scaling rule
     (one-line, low-risk) rather than routed through a plan task — /cmd/noggin/noggin added
     to .gitignore.
     T3 (build, high, shift+enter/L Picker fallback): triaged → DECISIONS.md D55 (ctrl+enter +
     ctrl+l, non-printable fallback closes it structurally) → PLAN.md M1.5 (updated).
     T2 (review, low, FocusDV2/resolveFocus edge case): triaged → DECISIONS.md D56 (left
     unpinned, same reasoning as D54).
     T4 (build, high, TestNoFooterHintNamesCtrlP collision): resolved directly in M1.4a's own
     build, per the plan's own "update ... hint tests to the new chords" done-when clause —
     the pre-C3 "no ctrl+p anywhere" invariant no longer holds now that D55 gives PeekDV2 that
     chord, so footer_test.go/help_test.go's TestNoFooterHintNamesCtrlP/TestNoHelpModalNamesCtrlP
     were replaced with TestCtrlPNamedOnlyByPeekDV2/TestCtrlPNamedOnlyInFilesPaneHelp, scoping
     the invariant to "ctrl+p only where PeekDV2 lives" instead of "nowhere".
     T5 (review, med, model.go:965): resize()'s find-dropped message is DV1-only, breaks S2
     parity → PLAN.md M1.6a (new task, effort 2).
     T6 (review, med, checks.py:1): no pty coverage for any C3 chord → PLAN.md M1.7 (new
     task, effort 2).
     T7 (review, low, checks.py:991): stale `viewerPane` identifier in a comment → fixed
     directly this session per the effort-scaling rule (one-line, low-risk).
     T8 (review, low, keymap.go:257): lowercase dv1/dv2 in help strings → DECISIONS.md D57
     (wontfix, kept as-is — consistent with the uniform lowercase hint style).
     T9 (validate, low, D55): degraded-terminal ctrl+enter fallback undocumented for users →
     DECISIONS.md D58 (wontfix — ctrl+l is already named alongside it everywhere). -->
