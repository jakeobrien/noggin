# Triage — C2 find

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
| T1 | review | medium | cmd/noggin/files_search.go:146-166 | `grepFiles` passes every file in the resolved Scope as a literal CLI argument to `rg` with no batching/chunking; an All Roots search (or a large Active Root) can push the argument list past the OS's ARG_MAX, and `cmd.Output()`'s error in that case is silently folded into "no results" (same path as a genuine no-match), giving no signal that the search didn't actually run. Recommend batching the file list across multiple `rg` invocations (merging results) or otherwise bounding argument size. | open |
| T2 | review | low | cmd/noggin/files_search.go:40-46 | `listFiles`'s doc comment documents a design extension beyond SPEC S6's literal text — absorbing a live-but-`rg`-unreadable Root (e.g. permission denied) the same way a Broken Root is silently skipped — but unlike this milestone's other implementation calls (D36, D37, D38) it has no DECISIONS.md entry. Recommend logging it (e.g. folded into D36) for consistency with the cycle's decision-logging practice. | open |
| T3 | review | low | cmd/noggin/files_search_test.go:197-227,294-321 | `TestFuzzyFilesCapsAndReportsTruncation` covers the exact-cap/not-truncated boundary in addition to the over-cap case; `TestGrepFilesCapsAndReportsTruncation` only covers the over-cap case, an asymmetry in test coverage between the two Pickers' shared cap/truncation contract (S12/D37). Recommend adding the same boundary case to the grep test. | open |
| T4 | review | low | cmd/noggin/ripgrep_test.go:11-32; cmd/noggin/main.go:30 | M1.4/PLAN.md M1.1's done-when ("a run with `rg` absent from PATH still launches and opens a document") has no test exercising that combined property — only `detectRipgrep`'s own flag/message is unit-tested. The property currently holds by construction (`main.go` calls `detectRipgrep()` unconditionally with no branching that could gate launch), but the plan's literal acceptance criterion isn't exercised by any test. Recommend a small launch/model-level test, or an explicit note that the structural guarantee is considered sufficient. | open |
| T5 | review | medium | cmd/noggin/help.go:14-40; cmd/noggin/picker_test.go (TestHelpBindingsNameThePickerFromEveryPane) | S2 ("The Footer shows the Picker's actions, and the Help Modal covers them") and PLAN.md M2.3's done-when ("Footer/Help show the Picker's own actions while it is open") are only half satisfied: `helpBindings` adds `keys.Picker` (the key that *opens* the modal) to every pane's Help listing, but never lists the Picker's own in-Picker actions (`PickerUp`/`PickerDown`/`PickerMode`/`PickerScope`) anywhere, and the Help Modal is by design unreachable while the Picker is open (`TestHelpUnreachableWhilePickerOpen`), so those actions are never documented in Help at all -- only in the footer's `pickerHintText`, which can itself truncate away under a narrow terminal with no Help fallback. Recommend either listing the Picker's own bindings in `helpBindings` for the panes Help remains reachable from, or logging a DECISIONS.md entry explaining why Help coverage is intentionally limited to the opening key. | open |
| T6 | review | low | cmd/noggin/picker.go:37-49 (picker struct); cmd/noggin/picker_view.go:130-146 (promptView) | The GLOSSARY.md term "Prompt" is used throughout doc comments and status text (`pickerModeScope`, `pickerStatus`, `promptView`) but the actual field holding it is named `input` (`picker.input`), not `prompt` -- the one place in this milestone's code where the glossary term and the Go identifier diverge (contrast `mode`/`scope`, which match `pickerMode`/`pickerScope` exactly). Recommend renaming `picker.input` to `picker.prompt` for consistency, or noting the divergence is intentional (e.g. because it wraps `textinput.Model` generically). | open |
