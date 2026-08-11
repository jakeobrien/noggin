# Progress — C4 recall

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

Current task: M3.4

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M3.3 built: `activatePickerInto` (fileMode) and `activateGrepPickerInto` (grepMode), both in
picker.go, called `openDocumentInto` directly without ever routing through `activateInto` —
TRIAGE T12's bug. Fix: each now calls `m.recordViewed(path)` immediately before
`openDocumentInto`, the same ordering `activateInto` (model.go) already uses, right before
the `dismissPicker()` return. New test file `recency_picker_test.go` (`TestPickerActivateRecordsLastViewed`,
4 subtests: file/grep Result x DV1/DV2 target) stages `m.picker.results` directly the same
way picker_test.go's existing Activate tests do (no live ripgrep search needed) and confirms
a timestamp lands + survives a reload via `loadDocStore`. Verified red-then-green: reverted
picker.go alone and confirmed all 4 subtests fail with a zero `LastViewed`, then confirmed
they pass with the fix restored. `make check` green; `gofmt -l` clean on both changed files.

Next: **M3.4** (recordViewed's failure-message clobber), then M2.10 (independent, any order)
before M4.1; consider charm-validate/code-review for the M3 milestone once M3.4 lands (M2.11
could use the same pass), since M3 was about to close that gate before this replan reopened
it.

## Status

| Key  | Status | Reviewed | Validated | Note                                     |
|------|--------|----------|-----------|------------------------------------------|
| M1.1 | done   | 2026-08-10 | 2026-08-10 |                                       |
| M1.2 | done   | 2026-08-10 | 2026-08-10 |                                       |
| M1.3 | done   | 2026-08-10 | 2026-08-10 |                                       |
| M1.4 | done   | 2026-08-10 | 2026-08-10 |                                       |
| M2.1 | done   | 2026-08-10 | 2026-08-11 |                                      |
| M2.2 | done   | 2026-08-10 | 2026-08-11 |                                      |
| M2.3 | done   | 2026-08-10 | 2026-08-11 |                                      |
| M2.4 | done   | 2026-08-10 | 2026-08-11 |                                      |
| M2.5 | done   | 2026-08-10 | 2026-08-11 | pre-agreed cut 1, not actually cut   |
| M2.6 | done   | 2026-08-10 | 2026-08-11 |                                      |
| M1.1a | done  | 2026-08-11 | -       | replan: pty harness fix, TRIAGE T4 (D71) |
| M2.7 | done   | 2026-08-11 | 2026-08-11 | replan: stale marker fix, TRIAGE T7 (D70) |
| M2.8 | done   | 2026-08-11 | 2026-08-11 | replan: star-binding discoverability, TRIAGE T5/T6 |
| M2.9 | done   | 2026-08-11 | 2026-08-11 | replan: naming cleanup, TRIAGE T1/T2/T9  |
| M3.1 | done   | 2026-08-11 | 2026-08-11 | SPEC/GLOSSARY-level gap re: Picker Activate not recording lastViewed — tracked as TRIAGE T12, not a M3.1 overclaim |
| M3.2 | done   | -        | -         |                                          |
| M3.3 | done   | -        | -         | replan: Picker never records lastViewed, TRIAGE T12 (D74) |
| M3.4 | todo   | -        | -         | replan: recordViewed failure message clobber, TRIAGE T13 |
| M2.10 | todo  | -        | -         | replan: pty star-marker column checks, TRIAGE T11        |
| M2.11 | done  | -        | -         | replan: Picker star marker (D72, TRIAGE T8), built same session |
| M4.1 | todo   | -        | -         |                                          |
| M4.2 | todo   | -        | -         |                                          |
| M4.3 | todo   | -        | -         | `Orphans` scope is pre-agreed cut 3      |
| M4.4 | todo   | -        | -         |                                          |
| M4.5 | todo   | -        | -         |                                          |
| M4.6 | todo   | -        | -         |                                          |
| M5.1 | todo   | -        | -         | Grep-side `Starred` is pre-agreed cut 4  |
| M6.1 | todo   | -        | -         |                                          |
| M6.2 | todo   | -        | -         | pre-agreed cut 2                         |
| M7.1 | todo   | -        | -         |                                          |
| M7.2 | todo   | -        | -         |                                          |
