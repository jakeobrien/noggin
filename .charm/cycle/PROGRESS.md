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

Current task: M2.9

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M2.8 done — keys.Star and keys.PickerStar are now discoverable outside the Roots Nav
(TRIAGE T5/T6). keys.PickerStar already lived in the Picker's footer hints (pickerHints,
footer.go) since M2.5's own build; the real gap was keys.Star missing from the Files Nav's
and each Document Viewer's (dv1Pane/dv2Pane, shared `default` switch case) paneHints
(footer.go) and Help Modal listing (help.go's helpBindings) — the binding already worked
from both (model.go's Update switch), it just had no hint anywhere but the Roots Nav.
keys.PickerStar stays deliberately off every Help Modal listing (its "star" Help()
description collides with keys.Star's own, now present in all three panes' lists, not just
Roots' — help.go's own comment covers the reasoning), so its footer hint remains its only
discoverability surface, unchanged from before.
Reworded the pinned TestHelpModalListsThePaneOwnBindings cases (filesPane/dv1Pane now want
"star", no longer wantNot) and bumped two go-test/pty-check terminal heights (30→32,
filesPane's Help Modal box grew from 28→29 rows) to keep the centered modal off the top
border — confirmed via TestHelpOverlayLeavesTheFrameVisibleAroundIt and the M5.3 pty block.
Also reworded the pty M5.3 viewer-context check (asserted "star" absent; now asserts it's
present) and its stale row-count comment. `make check` and `make check-pty` both green apart
from the pre-existing, unrelated T11.
Next: **M2.9** — naming cleanup (TestLoadStarredRootsNoConfigDir→...NoDataDir, evictDocStore's
cap→limit, toggleStarPicker→togglePickerStar); its dependencies (M1.1, M1.3, M2.5) are all
done. M3.2 (recency regardless of Root provenance) is also independently available.

## Status

| Key  | Status | Reviewed | Validated | Note                                     |
|------|--------|----------|-----------|------------------------------------------|
| M1.1 | done   | 2026-08-10 | 2026-08-10 |                                       |
| M1.2 | done   | 2026-08-10 | 2026-08-10 |                                       |
| M1.3 | done   | 2026-08-10 | 2026-08-10 |                                       |
| M1.4 | done   | 2026-08-10 | 2026-08-10 |                                       |
| M2.1 | done   | 2026-08-10 | -         |                                          |
| M2.2 | done   | 2026-08-10 | -         |                                          |
| M2.3 | done   | 2026-08-10 | -         |                                          |
| M2.4 | done   | 2026-08-10 | -         |                                          |
| M2.5 | done   | 2026-08-10 | -         | pre-agreed cut 1                         |
| M2.6 | done   | 2026-08-10 | -         |                                          |
| M1.1a | done  | -        | -         | replan: pty harness fix, TRIAGE T4 (D71) |
| M2.7 | done   | -        | -         | replan: stale marker fix, TRIAGE T7 (D70) |
| M2.8 | done   | -        | -         | replan: star-binding discoverability, TRIAGE T5/T6 |
| M2.9 | todo   | -        | -         | replan: naming cleanup, TRIAGE T1/T2/T9  |
| M3.1 | done   | -        | -         |                                          |
| M3.2 | todo   | -        | -         |                                          |
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
