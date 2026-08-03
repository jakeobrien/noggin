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

Current task: M3.2

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

Replanned 2026-08-03: TRIAGE cleared (T1-T6), all routed. T1 (grepFiles' unbounded rg arg list,
ARG_MAX risk) → M1.4a. T3 (grep cap-test boundary asymmetry) → M1.4b. T5 (Help Modal never
lists the Picker's own bindings) → M2.3a — M2.3 stays `Validated: -` until M2.3a lands and is
re-validated. T2 (listFiles' undocumented Root-absorption extension) and T6 (`picker.input` vs
glossary "Prompt") logged as decisions (D44, D47), no code change. T4 (M1.1's launch-property
acceptance criterion untested but holding by construction) dropped as already-handled — the
guarantee is structural (`main.go` calls `detectRipgrep()` unconditionally, no launch-gating
branch exists), not worth a dedicated test. See D43-D47.

M3.1 done — the Results region now renders the File Picker's rows (picker_view.go's
`pickerResultsContent`/`pickerFileRow`): `<root's filepath.Base>  <path relative to that root>`,
in `results.files`' own order (fuzzyFiles/M1.3 owns match-strength ranking; this only avoids
reshuffling it). Rows are only shown when `picker.resultsCurrent()` — a new helper, factored out
of `pickerStatus`'s existing three-way (query, mode, scope) label comparison — is true, so a
stale mode's rows can never flash under a Results box whose caption is already saying
"searching" for a newer generation.

"Wiring" turned out to need a Results selection cursor too (`picker.selected`), since a Result
list nobody can move over isn't yet usable and M3.2 (Preview "following selection changes") and
M3.3 ("activate a Result") both need one to exist — added now rather than deferred, on the
judgment that splitting it into its own task would be busywork ahead of tasks that assume it.
`picker.move(delta)` clamps via the existing `moveSelection` (T16) against `picker.resultCount()`
(files count in fileMode, grep count in grepMode — the grepMode arm is dead code until M4.1 wires
Grep results, but costs nothing to have right now). `PickerUp`/`PickerDown`, reserved as no-op
non-typing keys by M2.3, now call it. `selected` resets to 0 whenever a fresh result set lands
(`applyPickerResults`) or the Prompt drops below threshold (`queuePickerSearch`) — an index into a
candidate list that no longer exists has nothing sensible to preserve.

Evidence for the criterion: `TestPickerFileResultsShowRootAndRelativePath` (picker_view_test.go)
puts same-named `report.md` files in two `t.TempDir()` Roots and asserts both `<root> report.md`
rows render, distinguishable, in `results.files` order. `TestPickerUpDownMoveTheResultsSelection`
and `TestPickerSelectionResetsOnFreshResults` (picker_test.go) cover the selection cursor's
move/clamp and its reset on a fresh generation. `make check` and `make check-pty` both green.

Next step: M3.2 — the Preview pane renders the selected Result's file as raw, unrendered text
from the top, scrollable, following selection changes (`picker.selected` is now there for it to
read). M1.4a/M1.4b/M2.3a are queued but not blocking — pick them up opportunistically or before
M7 hardening.

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
| M1.4a| todo   | -        | -         | D43 (TRIAGE T1) |
| M1.4b| todo   | -        | -         | D45 (TRIAGE T3) |
| M2.3a| todo   | -        | -         | D46 (TRIAGE T5); M2.3 re-validates once done |
| M3.1 | done   | -        | -         |      |
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
