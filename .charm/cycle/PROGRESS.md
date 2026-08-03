# Progress — C2 find

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

Current task: M2.3a

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M1.4a and M1.4b done (this session). M1.4a: `grepFiles` (cmd/noggin/files_search.go) now
batches its file list across multiple `rg` invocations via `batchFilePaths`/`grepBatch`,
bounded by the injectable `grepArgMaxBytes` var (default 128KiB, well under real-world
ARG_MAX), merging results and applying `pickerResultCap`/truncation across the merged total
rather than per batch (`TestGrepFilesBatchesAcrossArgSizeLimit`). M1.4b: added the exact-cap/
not-truncated boundary case to `TestGrepFilesCapsAndReportsTruncation`, matching the fuzzy
test's existing coverage. `make check` green after both.

M3 (File Picker end to end) complete: M3.1–M3.3 done, not yet reviewed/validated (see Status
table). No in-flight changes. 2026-08-03 replan triaged review findings T6–T12 into new plan
tasks (see below) and left TRIAGE.md empty.

Gotcha carried forward: `picker.preview` (picker.go) is a real, persistently-sized
`viewport.Model`, unlike the Prompt's textinput. `openPicker` resets `m.picker` wholesale to
`newPicker()`'s zero state per S8 ("no memory"), so any field that holds a terminal-size
property rather than session state needs its Width/Height carried over across that reset
explicitly — `preview` and `seq` already do this (see their comments); check again when M4.x
adds fields to `picker`. Note: `preview`'s carry-over is currently one-sided — `openPicker`
does it but `dismissPicker` doesn't, which is exactly M3.2a/D48; fix that before trusting this
pattern for anything M4.x adds.

Next step: M4.1 — wire Grep mode to M1.4 (file, line number, matching line per row, truncated
to the pane). M4.2 (preview scrolled to the matched line) reuses M3.2's preview mechanism, so
pull M3.2a forward before M4.2 lands — its bug (dismiss-then-reopen loses preview size) would
otherwise silently follow into grep's preview path too. Before starting M4.3/M4.4, re-read D34:
rg's raw-file line numbers don't correspond to the Document Viewer's rendered line space, so
the hand-off re-derives the match by occurrence ordinal rather than trusting rg's line number,
with a documented degrade path.

Queued, non-blocking — pick up opportunistically or before M7 hardening: M2.3a (D46, Help
Modal's Picker-binding gap — M2.3 stays `Validated: -` until this lands and is re-validated;
cursor is here next), M3.1a (D50, Root label disambiguation for same-basename Roots), M3.2b
(T12, pin preview-error string), M3.3a/M3.3b (D49, activatePicker test-coverage gaps), M7.1a
(T11, terminology housekeeping).

Not queued, should land before M4.2: M3.2a (D48, dismissPicker preview-size bug — see gotcha
above).

No blockers.

## Status

| Key  | Status | Reviewed | Validated | Note |
|------|--------|----------|-----------|------|
| M1.1 | done   | 2026-08-03 | 2026-08-03 |  |
| M1.2 | done   | 2026-08-03 | 2026-08-03 |  |
| M1.3 | done   | 2026-08-03 | 2026-08-03 |  |
| M1.4 | done   | 2026-08-03 | 2026-08-03 |  |
| M2.1 | done   | 2026-08-03 | 2026-08-03 |      |
| M2.2 | done   | 2026-08-03 | 2026-08-03 |      |
| M2.3 | done   | 2026-08-03 | -         | criterion partially unmet, see M2.3a/D46 |
| M2.4 | done   | 2026-08-03 | 2026-08-03 |      |
| M2.5 | done   | 2026-08-03 | 2026-08-03 |      |
| M1.4a| done   | -        | -         | D43 (TRIAGE T1) |
| M1.4b| done   | -        | -         | D45 (TRIAGE T3) |
| M2.3a| todo   | -        | -         | D46 (TRIAGE T5); M2.3 re-validates once done |
| M3.1 | done   | 2026-08-03 | -         |      |
| M3.1a| todo   | -        | -         | D50 (TRIAGE T10) |
| M3.2 | done   | 2026-08-03 | -         |      |
| M3.2a| todo   | -        | -         | D48 (TRIAGE T6); fix before M4.2 |
| M3.2b| todo   | -        | -         | TRIAGE T12 |
| M3.3 | done   | 2026-08-03 | -         |      |
| M3.3a| todo   | -        | -         | D49 (TRIAGE T7, T8) |
| M3.3b| todo   | -        | -         | D49 (TRIAGE T9) |
| M4.1 | todo   | -        | -         |      |
| M4.2 | todo   | -        | -         |      |
| M4.3 | todo   | -        | -         |      |
| M4.4 | todo   | -        | -         |      |
| M5.1 | todo   | -        | -         |      |
| M5.2 | todo   | -        | -         |      |
| M5.3 | todo   | -        | -         |      |
| M6.1 | todo   | -        | -         | pre-agreed cut (D32) |
| M6.2 | todo   | -        | -         | pre-agreed cut (D32) |
| M7.1 | todo   | -        | -         |      |
| M7.1a| todo   | -        | -         | TRIAGE T11 |
| M7.2 | todo   | -        | -         |      |
