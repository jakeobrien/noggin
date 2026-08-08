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
| T1 | review | low | `.gitignore:2` | Only root-level `/noggin` is ignored, not `cmd/noggin/noggin` — a stray `go build .` inside `cmd/noggin/` leaves an untracked 15MB binary that can be committed by accident (hit once during C3/M1.3). One-line fix, deliberately not taken mid-task. | open |
| T3 | build | high | `.charm/DECISIONS.md:722` (D53), `cmd/noggin/picker.go:746` | D53's `L` fallback for `shift+enter` cannot be applied in the Picker as D53 says it should: the Picker's Prompt is always-typing (S2/C2), so `handlePickerInput` matching a printable `L` would swallow it instead of typing it. M1.5 needs either shift+enter-only in the Picker (accepting no fallback there) or a non-printable Picker fallback (alt+enter, ctrl+l). Files Nav side (M1.4) is unaffected and shipped with both keys. | open |
| T2 | review | low | `cmd/noggin/model.go:718`, `cmd/noggin/layout.go:52` | Untested edge case, not a known defect: `splitViewer` has no minimum-width guard (D54), so at an extreme terminal width DV2's rect can be empty while `m.dv2 != nil`. FocusDV2 then falls through `resolveFocus`'s focusOrder (correct per S5, which only mandates a no-op when DV2 *does not exist*, not when it is merely invisible), but nothing pins that reading. Decide whether to test it or record the reading as a decision. | open |
