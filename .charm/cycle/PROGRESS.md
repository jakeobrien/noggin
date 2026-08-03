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

Current task: M3.1

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M2.5 done — M2 (the Picker modal shell) is complete. Two decisions, both logged as D42:
(1) with no rg the Picker still *opens* and says `Picker unavailable: <reason>` inside the
modal rather than refusing to open with a footer message — keeps S1's "reachable from every
Focus" literal and keeps M2.1–M2.4's tests meaningful without rg; (2) all status text renders
on one line inside the Prompt box, under the textinput, with the mode/Scope indicator as that
box's title (`promptOuterHeight` 4 → 5) — the Prompt is the only region spanning the modal's
full width, so the longest message fits unwrapped (in the Results column it would wrap at
every ordinary size).

`pickerStatus(p, avail)` (picker_view.go) picks exactly one message by priority: unavailable →
below threshold (`type at least 3 characters`) → `searching…` → `no results` → `first 200
results (truncated)` → "" (nothing to say). "searching" is *not* just `p.searching`: results
whose (query, mode, scope) labels don't match the Prompt's also read as searching, which is how
the gap between keystroke and matured timer stays honest. `pickerModeScope(p)` renders
`File · Active Root` / `Grep · All Roots` — the surface M5.1/M5.2's in-place switching will be
seen through. `overlayPicker`/`renderPicker` now take `ripgrepAvailability` (threaded from
`m.ripgrep` in `View`).

The threshold itself is `pickerQueryMinLength = 3` behind `model.searchable()` (rg available
AND query ≥ 3 runes), consulted by *both* `queuePickerSearch` and `dispatchPickerSearch` —
it replaced the old empty-Prompt refusal (empty is below threshold, so the rg-empty-pattern
hazard is still covered). Gotcha for later tasks: any fixture that must actually search needs
`withRipgrep(detectRipgrep())` — `openedFixture` does not set it, so `pickerRoot` (picker_test.go)
now does.

Evidence for the criterion: `TestPickerStatusStates` (picker_view_test.go, 8 subtests) asserts
the rendered text of all five states plus "searching" and the say-nothing case against the real
frame; `TestPickerIssuesNoSearchBelowTheThreshold` shows "r"/"re" dispatch 0 searches while
"rep" dispatches 1. `make check` and `make check-pty` both green.

M2 validated (fresh-context, `make check` + `make check-pty` green, tests independently judged
sound, plus a manual pty run of the real binary): M2.1, M2.2, M2.4, M2.5 stamped Validated
2026-08-03. M2.3 left `-` — its own acceptance criterion ("Footer and Help Modal cover [the
Picker's actions]") is only half met: Footer does, Help Modal never lists the in-Picker
bindings (PickerUp/Down/Mode/Scope) and is unreachable while the Picker is open by design.
This is exactly TRIAGE T5 (open, medium), independently corroborated, not a new finding — no
new TRIAGE row added. M2.3 can be re-validated once T5 is triaged/resolved.

Next step: M3.1 — wire the File Picker's results into the Results region, each row showing the
file's Root and its path relative to that Root, ordered by match strength. `pickerResults.files`
is already populated and labelled; the Results box currently renders empty content.

## Status

| Key  | Status | Reviewed | Validated | Note |
|------|--------|----------|-----------|------|
| M1.1 | done   | 2026-08-03 | 2026-08-03 |  |
| M1.2 | done   | 2026-08-03 | 2026-08-03 |  |
| M1.3 | done   | 2026-08-03 | 2026-08-03 |  |
| M1.4 | done   | 2026-08-03 | 2026-08-03 |  |
| M2.1 | done   | 2026-08-03 | 2026-08-03 |      |
| M2.2 | done   | 2026-08-03 | 2026-08-03 |      |
| M2.3 | done   | 2026-08-03 | -         | criterion partially unmet, see T5 |
| M2.4 | done   | 2026-08-03 | 2026-08-03 |      |
| M2.5 | done   | 2026-08-03 | 2026-08-03 |      |
| M3.1 | todo   | -        | -         |      |
| M3.2 | todo   | -        | -         |      |
| M3.3 | todo   | -        | -         |      |
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
| M7.2 | todo   | -        | -         |      |
