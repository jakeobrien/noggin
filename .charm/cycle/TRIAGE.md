# Triage — C1 read-across-roots

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
| T9 | review | low | cmd/noggin/keymap.go:7-10 | M1.5 review: keymap.go's header comment ("no other file names a key") is now stale — keymap_test.go's `modifiers`/`namedKeys` maps (keymap_test.go:49-66) also name keys, by design (D17/M1.5 moved the test-only key-name inverse table there). keymap_test.go's own comment already scopes D2 correctly ("keymap.go remains the one production file that names a key"); update keymap.go's comment to add the same "production file" qualifier so the two files' claims agree. | open |
| T10 | review | low | cmd/noggin/viewer.go:87 | M1.5 review: `viewer.setOffset` is a pointer receiver, the only one on `viewer` — every other method (`setSize`, `open`, `render`, `offset`, `atBottom`, `lineCount`, `height`, `label`, `view`, `fill`, `scroll`) is a value receiver returning a new `viewer`, matching the type's functional/immutable-update convention. Recommend either a value-receiver `withOffset(n int) viewer` (update the one call site, viewer_test.go:194) or, if a mutator is intentional for test ergonomics, a one-line comment on why it's the deliberate exception. | open |

<!-- Empty as of this replan (2026-08-01). T1-T8 routed: T1, T2 dropped (already-handled,
     see PROGRESS.md history / D15); T3-T5 → PLAN.md M1.5 (D17); T6 → ROADMAP.md Unscheduled;
     T7 → PLAN.md M3.1 wording; T8 → PROGRESS.md Status table fixed directly. -->
