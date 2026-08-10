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
| T1 | review | low | cmd/noggin/store_test.go:133 | `TestLoadStarredRootsNoConfigDir`'s name still says "NoConfigDir" though M1.1's clean break renamed the concept to data dir everywhere else (its own doc comment and body already say "data directory"/`XDG_DATA_HOME`); rename to `TestLoadStarredRootsNoDataDir` for consistency. | open |
| T2 | review | low | cmd/noggin/docstore.go:123 | `evictDocStore(records map[string]docRecord, cap int)` names its parameter `cap`, shadowing Go's predeclared `cap()` builtin inside the function body; rename to e.g. `limit` or `maxRecords`. | open |
| T3 | review | low | cmd/noggin/docstore.go:166 (cf. cmd/noggin/store.go:110) | `saveDocStore` duplicates `saveStore`'s `os.MkdirAll`-then-`os.WriteFile` shape near-verbatim; low priority now with only two stores, but worth a shared helper if a third store format is ever added. | open |
| T4 | build | medium | test/pty/checks.py:81-91,460-483 | `make check-pty`'s M3.3/M3.4 checks (Roots pane listing/marking/persistence) fail: `config_env` still builds the starred-roots store path under `$XDG_CONFIG_HOME/noggin/roots`, the location M1.1's clean break moved off of this cycle (now `dataDir()`/`storePath()`); the harness never got the matching update, so it starts every M3.3/M3.4 process with an empty store and then can't find the persisted file it expects. Confirmed pre-existing (same failures on the commit before M2.1/M2.2/M2.3, not caused by this cycle's work); `make check` (the actual gate) is unaffected and green. | open |
