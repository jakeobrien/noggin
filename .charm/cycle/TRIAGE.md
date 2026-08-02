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
| T37 | review | medium | cmd/noggin/roots.go:98-116 (`needsCwd`), :132-136 (`resolveRoots`'s own stat) | `resolveLaunch` calls `needsCwd`, which stats the single launch argument, then falls through to `resolveRoots`, which stats the same path again independently — a TOCTOU swap between the two (e.g. a symlink retargeted from a directory to a file in that window) could pass `cwd=""` through to `resolveFile`, reintroducing T20's bare-`.` bug this whole fix exists to prevent. Narrow and unlikely, but the doc comments state the survives-the-failure invariant unconditionally. Recommend `needsCwd` return its stat result for `resolveRoots` to reuse (one stat, no race), or an explicit comment acknowledging the residual race. | open |
| T38 | review | low | cmd/noggin/roots.go:99-102 (`needsCwd` signature/body) | `needsCwd(args []string, cwd string) bool`'s `cwd` parameter is never read in the function body — every branch decides purely from `args[0]`'s shape (absolute/relative, dir/file), never from `cwd`'s value. Same class of nit as T19's unused `storePath` parameter (D24/M3.6). Recommend dropping the parameter or documenting why it's kept for signature symmetry with `resolveRoots`. | open |
| T39 | review | low | cmd/noggin/find.go:182-199 (`findMatches`) | `start, end := hStart, hEnd` declared then conditionally overwritten only when `smartCase` is true is an always-declare/sometimes-overwrite shape that reads awkwardly. Recommend a small `origSpan(toOrig []int, hStart, hEnd int) (int, int)` helper (identity when `toOrig == nil`) instead. Cosmetic only. | open |
| T40 | review | low | cmd/noggin/roots_test.go (resolveLaunch tests, ~line 280-341) | No test exercises `resolveLaunch` with `cwdErr != nil` at `len(args)==0` or `len(args)>1`, even though `needsCwd`'s own doc comment (roots.go:100-102) makes explicit claims about both cases ("zero arguments never touch cwd; more than one is refused... before cwd would ever be used"). Recommend one case each to lock in the documented behavior at the `resolveLaunch` level, not just `needsCwd`'s. | open |
| T41 | review | low | cmd/noggin/find_test.go:132-159 (`TestFindMatchesLineAttributionSurvivesCaseFoldLengthChange`) | T29's regression test only covers a fold-shrinking rune preceding a match on its own line, not a match that itself spans a fold-length-changing rune (e.g. query `"k"` against content containing the Kelvin sign `"K"`, U+212A) — the case that most directly exercises `ends[]`/`toOrig`'s translation of `colEnd`. Not currently broken (hand-traced correct), just unverified by a test. Recommend adding that case to the table. | open |
| T42 | review | low | .charm/cycle/PLAN.md (M5.4 bullet) | M5.4's task text ("...only a relative one still gets dropped") is stale against D28's actual, more precise fix: an absolute *file* argument still gets dropped (via `resolveFile` always anchoring on cwd), only an absolute *directory* argument survives. Code and D28 are correct; PLAN.md's own prose is the superseded framing a reader relying on PLAN.md alone would be misled by. Recommend amending the bullet to point at D28 rather than restate the coarser rule. | open |
| T43 | review | low | cmd/noggin/model.go:418-437 (`resize`) + cmd/noggin/viewer.go:71 (`render`) | T34's resize guard only fires for a find still mid-edit (`findEditing()`); a *committed* find (post-Enter, match indicated via `applyCurrentMatch`) is still silently discarded by any width-changing resize or Navigator toggle, since `viewer.render()` unconditionally resets `findState` on every re-render and the guard checks `findEditing()`, not `hasFindMatches()`. Pre-existing since M5.1, not a regression from M5.4, but the same "confusion is not allowed" (S1/S8) gap T34 fixed for the mid-edit case, left open for the committed case. Recommend a triage decision: new task, or documented wontfix. | open |

<!-- T28-T36 from the M3.5/M5.1-M5.3 backlog were routed to M5.4/M5.5 (PLAN.md), ROADMAP.md
     Unscheduled (T31), or dropped (T36) by replan C1/2026-08-02; see D27. T37-T43 above are
     fresh findings from the 2026-08-02 code review of M5.4/M5.5's own diff (git diff
     04ff677..HEAD -- cmd/noggin), covering the M3.1/M3.5/M4.2/M5.1/M5.2/M5.3 Reviewed-stamp
     gaps that diff cleared, plus M5.4/M5.5 themselves. -->
