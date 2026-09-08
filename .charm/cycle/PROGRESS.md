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

Current task: M1.5

## Handoff

M1.4 done: with the Wikipedia Query Source active, the Files Nav's Browse Mode cycle drops
Tree entirely and toggles Recent<->Starred (`filesNav.cycleMode`), landing on Starred the
first time a session shows Wikipedia (`filesNav.setWikipedia`, `f.root` set aside to nil);
Starred's own scope set drops `Orphans` while Wikipedia is active (`modeScopes(mode,
wikipedia)`); Recent's second scope value is relabelled `All Sources` (was `All Roots`)
unconditionally, for every Active Source, not just Wikipedia (`browseScope.label()`,
`flatEmptyText`'s matching wording). Per-Source browse memory now covers Wikipedia via a new
reserved key (`wikipediaSourceKey`) in the same `filesNav.memory` map Roots use, sharing the
remember/restore halves (`rememberOutgoing`/`restoreSource`) `setRoot` and the new
`setWikipedia` both call — verified round-tripping a mode/scope pair across a
Wikipedia->Root->Wikipedia switch, and that a Root's own pair is untouched by a Wikipedia
visit in between. `model.activateWikipediaSource` now calls `filesNav.setWikipedia` (the
piece M1.1's own comment deferred to this task). New tests: `wikipedia_browse_test.go`
(filesNav-level: mode cycle, both scope cycles, mode/scope row labels, the memory
round-trip) plus two model-level tests appended to `wikipedia_source_test.go`. Existing
"All Roots"-label assertions in `browse_test.go` updated in place (3 spots); Picker's own
separate `pickerScope.allRootsScope` label (`scope.go`/`picker_view.go`) is untouched —
different type, different task (M3.5). `go test ./...`/`make check` pass.

Next: M1.5 (Document Store admits article records) is queued next — the last M1 task, needs
nothing further from M1.4 beyond what is already in place. M2 (Wikipedia client) remains
independent of M1 and may be interleaved instead. Read `findings/wikimedia-api.md` before
M2.2/M2.3. D120 covers the spike extraction (M2.1).

## Status

| Key   | Status | Reviewed | Validated | Note |
|-------|--------|----------|-----------|------|
| M1.1  | done   | 2026-09-07 | -       |      |
| M1.2  | done   | 2026-09-07 | -       |      |
| M1.3  | done   | -        | -         |      |
| M1.4  | done   | -        | -         |      |
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
