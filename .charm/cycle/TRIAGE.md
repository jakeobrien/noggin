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
     DECISIONS.md D58 (wontfix — ctrl+l is already named alongside it everywhere).
     T10 (build, med, checks.py:1): no pty coverage for ctrl+x/CloseDV2 → PLAN.md M2.1a (new
     task, effort 1).
     T11 (review, high, model.go:651): refresh()'s message-combining switch overwrites a
     read-failure message with a find-dropped message, no `if m.message == ""` guard —
     re-confirmed live by charm-validate (2026-08-09) → PLAN.md M3.1a (new task, effort 3;
     folded with T13's refactor since both touch the same switch) → DECISIONS.md D61.
     T12 (review, low, model.go:309): closeDV2 doesn't reset the stale baseline's DV2 entry →
     PLAN.md M2.1b (new task, effort 1).
     T13 (review, low, model.go:651): duplicated message-combining logic between refresh()
     and resize() → folded into PLAN.md M3.1a alongside T11's fix (see D61) rather than a
     standalone task, since both changes touch the same switch. -->
