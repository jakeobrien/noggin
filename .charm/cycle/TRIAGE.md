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
| T17 | review | medium | cmd/noggin/picker_view.go:279-286 vs cmd/noggin/files.go:368-373, cmd/noggin/rootsnav.go:196-199 | M2.11's new `starMarker()` helper doc comment claims the "★ "/"  " two-column field is "shared by every star-marking surface (files.go, rootsnav.go, model.go's Document Viewer title, and now Picker rows)", but files.go's and rootsnav.go's row renderers still hand-roll the identical literal independently — no actual consolidation happened, only Picker rows call `starMarker`. Either refactor files.go/rootsnav.go to call `starMarker` or reword the comment to stop implying they already do. | open |
| T18 | review | low-medium | cmd/noggin/recency_test.go (M3.4 tests) vs cmd/noggin/picker.go:604,716 | M3.4's regression tests (`TestActivateStoreWriteFailureSurfacesMessage`, `TestFinishEditStoreWriteFailureSurfacesMessage`) cover only `activateInto`/`finishEdit`, not the two `recordViewed` call sites (`activatePickerInto`, `activateGrepPickerInto`) M3.3 added immediately before M3.4 in the same replan batch. Traced by hand: no functional bug today — neither Picker path sets a success message and `dismissPicker`/`resetPicker` never touch `m.message`, so there is nothing for a `recordViewed` failure to clobber there yet — but GLOSSARY's Activate entry and D74's own reasoning treat Picker commit as "Activate" too, so this is narrower coverage than M3.4's stated goal and T13's spirit. Recommend a small follow-up test (e.g. `TestPickerActivateStoreWriteFailureSurfacesMessage`, file+grep) mirroring the existing two, as regression insurance against a future success-message addition on that path. Re-confirmed by validate (2026-08-10): still accurate, no functional bug. | open |
| T19 | validate | low | cmd/noggin/picker_star_test.go:121 vs cmd/noggin/picker_view.go:515,524 | M2.11's same-frame live-render toggle test (`TestPickerStarTogglingUpdatesTheRenderedMarkerLive`) exercises only File Picker mode; no equivalent test drives a real `m.View()` render through a `PickerStar` toggle in Grep mode (Grep's existing test, `TestPickerStarTogglesSelectedGrepResult`, covers store mutation but not the rendered marker). The underlying read (`docs[...].Starred`, computed fresh every render, mode-agnostic) makes a bug unlikely, but this falls short of M2.11's "both Picker modes" acceptance wording for the toggle clause specifically. Recommend a `Grep` variant of the same-frame toggle test for full parity. | open |
