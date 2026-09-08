# Progress — C6 sources

<!-- Per-cycle. Cursor and statuses ONLY — no work summaries (git history carries those).
     The Handoff section is the single exception: overwrite it each session, never append.
     Statuses: todo | in-progress | done | blocked | deferred.
     Blocked rows reference the DECISIONS.md / TRIAGE.md item they wait on — the set of
     blocked rows is the cycle's blocking-issue list.

     Reviewed / Validated columns: date (yyyy-mm-dd) the task was last covered by code
     review / validate, or `-` if it never has been. Set per task key, by the role that
     performed the check — or the parent that dispatched it as a subagent — for exactly the
     tasks that check actually covered, never assumed for a whole milestone. This is what
     makes reviewed/validated state accurate across replans that insert tasks earlier,
     partial mid-milestone checks, and skipped end-of-milestone gates — a single cycle-wide
     checkpoint can't represent any of those. Clear a cell back to `-` when the task's code
     changes after that date (fix-ups, reopened work) — it needs re-covering. -->

## Cursor

Current task: M1.4

## Handoff

M1.3 done: relabelled every non-Grep, non-Root-mechanism "Active Root" to "Active Source" —
the two actually-rendered scope labels (`browseScope.label()`'s Recent/Starred default in
`browse.go`, and `pickerModeScope`'s File-mode default in `picker_view.go`, now mode-aware:
File reads "Active Source", Grep still reads "Active Root" since Grep never reaches
Wikipedia, S6) plus every doc comment/test comment describing those same two scopes or
general Active-Source session state (`rootsnav.go`, `files.go`, `browse_list.go`,
`picker.go`, `keymap.go`, `model.go`'s `historyBack`/`historyForward`, and the matching test
files). Left untouched as Root-only mechanism or Grep-scope: `scope.go`, `staleness.go`,
`files_search.go`'s roots-based resolution comments, Refresh/staleness/git-marker comments in
`model.go`/`files.go`/`keymap.go`, and `picker.go`'s `activatePicker`/history-walk Root-to-Root
tests (`history_test.go`, `picker_test.go`'s Grep-mode/Root-switch cases) — `grep -rn
"Active Root" cmd/noggin --include=*.go` now matches only those two buckets (verified by
reading every remaining hit). `go test ./...`/`make check` pass; existing tests were updated
in place, none deleted.

Next: M1.4 (Files Nav with Wikipedia active: Starred/Recent-only modes, per-Source browse
memory) is queued next — needs M1.1, now done. M2 (Wikipedia client) remains independent of M1
and may be interleaved instead. Read `findings/wikimedia-api.md` before M2.2/M2.3. D120 covers
the spike extraction (M2.1).

## Status

| Key   | Status | Reviewed | Validated | Note |
|-------|--------|----------|-----------|------|
| M1.1  | done   | 2026-09-07 | -       |      |
| M1.2  | done   | 2026-09-07 | -       |      |
| M1.3  | done   | -        | -         |      |
| M1.4  | todo   | -        | -         |      |
| M1.5  | todo   | -        | -         |      |
| M2.1  | todo   | -        | -         |      |
| M2.2  | todo   | -        | -         |      |
| M2.3  | todo   | -        | -         |      |
| M3.1  | todo   | -        | -         |      |
| M3.2  | todo   | -        | -         |      |
| M3.3  | todo   | -        | -         |      |
| M3.4  | todo   | -        | -         |      |
| M3.5  | todo   | -        | -         |      |
| M4.1  | todo   | -        | -         |      |
| M4.2  | todo   | -        | -         |      |
| M4.3  | todo   | -        | -         |      |
| M5.1  | todo   | -        | -         |      |
| M5.2  | todo   | -        | -         |      |
| M6.1  | todo   | -        | -         | pre-agreed cut 1 |
| M7.1  | todo   | -        | -         | pre-agreed cut 2 |
| M7.2  | todo   | -        | -         | pre-agreed cut 2 |
| M7.3  | todo   | -        | -         | pre-agreed cut 2 |
| M8.1  | todo   | -        | -         |      |
| M8.2  | todo   | -        | -         |      |
| M8.3  | todo   | -        | -         |      |
| M9.1  | todo   | -        | -         | pre-agreed cut 3 |
| M10.1 | todo   | -        | -         |      |
