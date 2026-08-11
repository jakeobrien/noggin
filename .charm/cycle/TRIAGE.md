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
| T11 | build | medium | test/pty/checks.py:582,584 | M2.3's star-marker column (files.go:367-369) inserted a two-char field between the git-status marker and the file name, so the C2a/M4.1 pty checks' single-space literals (`"M tracked.md"`, `"A new.md"`) no longer match the rendered row (now e.g. `"M   tracked.md"`); both fail under `make check-pty` while `make check` is unaffected. Surfaced by M1.1a's build, unrelated to its own config_env fix. | open |
| T12 | review | high | cmd/noggin/picker.go:579-596,699-718 | M3.1/S4: neither activatePickerInto nor activateGrepPickerInto calls recordViewed, so opening a document via the File or Grep Picker never records lastViewed even though GLOSSARY's Activate definition and S4 cover Picker commits (they land in DV1/DV2 via the same openDocumentInto); M3.1's own PLAN acceptance criterion is narrower than SPEC and doesn't catch this gap, and no DECISIONS/TRIAGE entry currently scopes the exclusion. | open |
| T13 | review | medium | cmd/noggin/model.go:462-469,990-995,250-268 | M3.1: recordViewed's store-write failure path sets m.message inside withDocRecord, which can silently overwrite/mask a just-succeeded Activate/Edit's own message with an unrelated store-save error; no test covers this interaction. | open |
| T14 | review | low | cmd/noggin/keymap.go:19-22 | M2.8: keys.Star's doc comment still describes it only as a Roots Nav action ("toggles the selected root's starred mark"), stale since M2.1/M2.2 gave Files Nav and the Document Viewer the same binding, and now more misleading given M2.8 explicitly extends its discoverability to those panes. | open |
| T15 | review | low | test/pty/checks.py:463 | M1.1a: comment says the test isolates "$HOME/$XDG_CONFIG_HOME" but config_env (fixed by this same commit, same file) now also resolves via $XDG_DATA_HOME — comment is stale. | open |
| T16 | review | low | cmd/noggin/model.go:270-282 | M2.7/M3.1: recordViewed recomputes and reassigns m.files.starred via filesStarredMarks() on every recency-only write even though only LastViewed changed (comment calls it "harmless"); bounded by docStoreCap today but couples an unrelated write path to Files Nav state — worth revisiting if this path gets hot. | open |
