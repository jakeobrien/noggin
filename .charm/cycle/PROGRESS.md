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

Current task: M3.3

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

Replan session: triaged all six open TRIAGE items (T11-T16). T12 (Picker Activate never
records lastViewed) and T13 (recordViewed's failure message can clobber a success message)
accepted as new tasks **M3.3**, **M3.4** (D74) — M3 is reopened, no longer fully `done`; M4.2
now also needs M3.3 so `Recent` mode isn't built against a known-incomplete data source. T11
(pty check literals broken by M2.3's star column) accepted as **M2.10**. T14 and T15 (stale
doc comments) were trivial and fixed directly during replan, not turned into tasks —
keys.Star's doc comment (keymap.go) and a stale `$XDG_CONFIG_HOME` comment (test/pty/checks.py)
both updated. T16 (recordViewed's harmless-looking m.files.starred recompute) logged wontfix
(D75), no task. TRIAGE.md is now empty.

The user then resolved the previously-open **D72** directly (Picker rows should show a
visible star marker) — SPEC.md's S3 amended in place, and **M2.11** built the same session:
`pickerFileRow`/`pickerGrepRow` (picker_view.go) gained a `starred bool` param and prepend the
shared "\u2605 "/"  " marker, computed live from `m.docs` via `pickerResultsContent`'s new `docs`
param (threaded through `renderPicker`/`overlayPicker`); the marker is folded into
`highlightMatches`' existing self-contained-per-run styling via a new `prependMarker` helper
rather than pre-styled separately, avoiding the reset-cancels-reverse hazard that comment
warns about. Nine existing pinned-string/pinned-column tests in picker_view_test.go needed
updating for the new 2-column budget (content width is now `width-2`); three new tests added
(two unit, one integration covering live toggle-without-refresh). `make check` green.

Next: **M3.3**, then M3.4 and M2.10 (independent, any order) before M4.1; consider
charm-validate/code-review for the M3 milestone once M3.3/M3.4 land (M2.11 could use the same
pass), since M3 was about to close that gate when this replan reopened it.

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
| M3.3 | todo   | -        | -         | replan: Picker never records lastViewed, TRIAGE T12 (D74) |
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
