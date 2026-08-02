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
| T44 | validate | low | find_test.go (TestWindowResizeAfterCommittedFindSurfacesMessage); model.go:237 toggleNavigator | M5.6/T43's acceptance text names "a resize **or** Navigator toggle" arriving after a committed find, but only the raw-resize path has a dedicated test; `toggleNavigator` calls the same guarded `model.resize`, so the behavior is very likely correct, but the criterion isn't 1:1 covered. Decide: add a toggleNavigator-driven test, or accept given the shared code path. | open |
| T45 | review | medium | cmd/noggin/model.go:439 (resize's `findHadMatches && !hasFindMatches()` case); model.go:237-240 toggleNavigator | Deepens T44: verified by a temporary test (`press(t, m, "/")` → type → `enter` → `press(t, m, "ctrl+n")`, removed after) that toggling the Navigator after committing a find sets `m.message = "window resized; search cleared"` even though no resize occurred — unlike T34's guard, T43's `hasFindMatches()` check is not gated against `toggleNavigator` (a committed, non-editing find leaves the Navigator binding free to reach `resize`), so the message is factually wrong on a path PLAN.md's own M5.6 text names as expected ("a resize **or** Navigator toggle"). Recommend reworking the message to not claim a window resize when the trigger was `toggleNavigator` (e.g. drop "window resized;" or thread the actual cause through). | open |
| T46 | review | low | cmd/noggin/roots.go:71-121 (`resolveLaunch`/`needsCwd`); cmd/noggin/roots_test.go (no matching case near :306-341) | T37's stat-reuse refactor added a code path in `resolveLaunch` for an absolute, *not-found* single argument combined with an unresolvable cwd (`needsCwd` returns `stated=true, statErr!=nil`, routed through `resolvePathArg`'s `err != nil` branch) — existing tests cover the dir-exists (`TestResolveLaunchKeepsAbsoluteDirArgOnUnresolvableCwd`) and file-exists (`TestResolveLaunchStillDropsAbsoluteFileArgOnUnresolvableCwd`) cases but not this one. Traced by hand to behave correctly (same "cannot open %s: no such path" message as before), but the new code path itself is unexercised by any test. Recommend a `TestResolveLaunchAbsoluteNotFoundArgOnUnresolvableCwd`-shaped case. | open |

<!-- T37-T43 (the 2026-08-02 code review of M5.4/M5.5's own diff) were routed by
     replan C1/2026-08-03 — T37 and T43 to a new PLAN.md task M5.6 (correctness), T38-T41 to
     a new PLAN.md task M5.7 (standards/test-hardening), T42 fixed directly in PLAN.md's own
     M5.4 bullet text (a documentation accuracy fix, not code); see D29. T28-T36 from the
     M3.5/M5.1-M5.3 backlog were routed to M5.4/M5.5, ROADMAP.md Unscheduled (T31), or dropped
     (T36) by replan C1/2026-08-02; see D27. -->
