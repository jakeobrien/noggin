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
| T11 | review | low | internal/tree/tree.go:34-39 | M2.1 review: `Walk`'s defensive "root is a file, not a dir" branch (sets `IsDir=false`, returns without walking) contradicts its own doc comment ("root must be a directory; a file argument is M3.1's concern") and is never exercised — `main.go` only calls `Walk` after confirming `info.IsDir()`, and no test covers this branch. Recommend dropping it and letting `os.ReadDir` fail naturally on a non-directory root, or, if the behaviour is meant to be part of the contract, add a test and drop the "not this function's concern" wording. | open |
| T12 | review | low | internal/tree/policy.go:11 | M2.1 review: policy.go's doc comment calls the allowlist "what noggin considers a readable document (S6)" but GLOSSARY.md's term is "Readable file" ("a file noggin will show and render") — align the comment's wording with the glossary term. | open |
| T13 | review | low | internal/tree/policy.go:37-42 | M2.1 review: `isReadable` matches extensions case-insensitively (`strings.ToLower`) but the extensionless conventional names (README/LICENSE/CHANGELOG) match by exact case only, so a real-world `Readme` or `readme` (no extension) file would silently not be shown. S6 doesn't specify either way; needs a decision on whether names should also normalize case, then a test locking in the choice. | open |
| T14 | review | low | internal/tree/gitignore.go:53-56 | M2.1 review: the escaped-leading-`\#`/`\!` handling in `compilePattern` has no test case — `TestIgnoreMatch`'s "comment is inert" case doesn't exercise the escape path. Add a case asserting `\!important.md` matches the literal name `!important.md` rather than being treated as a negation line. | open |
| T15 | review | low | cmd/noggin/model.go:76-90 | M2.3 review: `activate()`'s read-error path returns the model unchanged with no message at all — silent per SPEC.md's cross-cutting principle ("Silence is not an error, but confusion is… they never fail invisibly"). Deliberate per the M2.3 handoff in PROGRESS.md (deferred to M4.2's footer-message work) but not yet tracked as a formal open item anywhere; logging here so it isn't lost before M4.2 lands. | open |

<!-- Empty as of this replan (2026-08-01). T1-T8 routed: T1, T2 dropped (already-handled,
     see PROGRESS.md history / D15); T3-T5 → PLAN.md M1.5 (D17); T6 → ROADMAP.md Unscheduled;
     T7 → PLAN.md M3.1 wording; T8 → PROGRESS.md Status table fixed directly. -->
