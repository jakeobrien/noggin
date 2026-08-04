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
| T1 | validate | low | picker_test.go:430, picker_test.go:517 | M4.3's headline acceptance test (`TestPickerActivateGrepResultOpensDocumentScrolledToMatchedLine`) alone doesn't isolate D34's ordinal-resolution bridge (its single-match fixture passes even if `resolveGrepLanding` is disabled); only paired with `TestPickerActivateGrepResultLandsOnTheSelectedOccurrenceNotJustTheFirst` is the criterion actually covered. Not a functional bug — a comment cross-referencing the two tests would make the coupling explicit. | open |
| T2 | validate | low | picker.go (activateGrepPicker), model.go:146 (openDocument) | No test directly asserts S10 ("does not move/expand/re-select the Files Nav tree") for the **grep** activation path specifically — only `TestPickerActivateOpensResultAsActiveDocumentOnly` (M3.3, file mode) covers it. Code path is shared (`openDocument` never touches `m.files`/`m.roots`), so the property structurally holds, but there's no direct assertion for grep. | open |
| T3 | validate | low | picker_view_test.go:280 (TestPickerFileResultsDisambiguateSameBasenameRoots) | M3.1a's disambiguation test compares whole rendered terminal lines, not the picker row in isolation — passes/fails on incidental surrounding frame content (nav-pane text, footer hints), not proven isolated to `pickerRootLabel`'s output. M3.1b's `TestPickerFileRowFillsItsFullBudget` is the test actually doing the isolating work; confirmed via mutation that reintroducing M3.1b's fixed bug still passes this test. Functionally fine (production code correct), test-quality gap. | open |
| T4 | validate | low | picker_test.go:376 (TestPickerActivateOpensResultAsActiveDocumentOnly, expansion assertion added by M3.3b) | The fixture's `m.files.expanded` stays an empty map throughout (no subdirectories to expand), so the `reflect.DeepEqual` assertion M3.3b added never actually exercises a same-cardinality key-substitution scenario — confirmed via mutation (forcing an expanded-map swap inside `activatePicker` still passed, since both sides are empty). The mechanism (DeepEqual over a genuinely non-empty map) is correct and sound in general; the fixture just doesn't reach it. Fix: give the fixture at least one expanded directory. | open |

<!-- Inbox empty: T6-T12 triaged 2026-08-03 (charm-replan). Routing:
     T6  -> PLAN M3.2a, D48 (accepted, material)
     T7  -> PLAN M3.3a, D49 (accepted, batch)
     T8  -> PLAN M3.3a, D49 (accepted, batch)
     T9  -> PLAN M3.3b, D49 (accepted, batch)
     T10 -> PLAN M3.1a, D50 (accepted, scope judgment)
     T11 -> PLAN M7.1a (accepted, low-cost housekeeping)
     T12 -> PLAN M3.2b (accepted, low-cost housekeeping)

     T13-T20 triaged 2026-08-03 (charm-replan). Routing:
     T13 -> PLAN M3.1b, D52 (accepted, material; bundled with T15, T16)
     T14 -> PLAN M2.1a, D51 (accepted, material)
     T15 -> PLAN M3.1b, D52 (accepted, bundled with T13)
     T16 -> PLAN M3.1b, D52 (accepted, bundled with T13)
     T17 -> PLAN M3.1b (accepted, low-cost, folded into T13's fix since same function)
     T18 -> PLAN M1.4c (accepted, low-cost; bundled with T19)
     T19 -> PLAN M1.4c (accepted, low-cost; bundled with T18)
     T20 -> dropped: reporter's own report called it a defensible deliberate choice (avoids
            OS-dependent err.Error() text brittleness) with no functional gap; low risk/value,
            no decision entry warranted per the escalation matrix -->
