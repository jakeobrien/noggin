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

Current task: M2.4

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M2.3 done: always-typing input dispatch (S2). `handlePickerInput` (picker.go) now mirrors
`updateFindQuery`'s shape — a switch of non-typing keys (`Cancel` dismisses; `Activate`, two
new arrow-only `PickerUp`/`PickerDown` bindings, the Document Viewer's existing
`PageUp`/`PageDown`/`HalfPageUp`/`HalfPageDown` reused as-is for Preview scrolling, and two new
`PickerMode` (`tab`)/`PickerScope` (`shift+tab`) bindings are all recognized and swallowed, with
no action yet — M3.1/M3.2/M5.1/M5.2 wire those) — then falls through to
`m.picker.input.Update(msg)` for everything else, typing it into the Prompt regardless of what
it binds to elsewhere. D40 records why the new bindings are arrow/Tab/Shift+Tab rather than
reusing `LineUp`/`Top` etc. as-is: those double up a letter (`j`/`k`/`g`/`G`), which would break
S2 the moment a query contained one.

Footer/Help coverage (the task's other half): `paneHints`/`helpBindings` now list `keys.Picker`
in every pane (D39's binding is now discoverable); `footer.go` gained `pickerHints`/
`pickerHintText`, sharing a new `truncatedHintText` helper with `hintText` rather than a second
truncation loop; `model.renderFooter` swaps to `pickerHintText` while `m.picker.open`. Help
Modal itself stays unreachable while the Picker is open — same precedent as Find (`?` types
into the Prompt instead), verified by a new test mirroring `TestHelpUnreachableWhileFindEditing`.

Updated `TestPickerModalLeavesTheFrameVisibleAroundIt` (M2.2's test): its "footer unchanged"
assertion no longer holds now that the footer legitimately changes with the Picker (intentional,
not overlay leakage) — narrowed to the top-row/any-row-changed checks, footer content now
covered by the new `TestFooterShowsPickerHintsWhileOpen`.

Tests: `picker_test.go` (typing a string containing `j`/`k`/`g`/`/` lands intact in the Prompt
with no side effects on the doc/viewer/help; every non-typing key stays open with an empty
Prompt; Cancel still dismisses), `footer_test.go`/`help_test.go` (Picker named from every pane's
hints/help; `pickerHintText` names Picker actions not the focused pane's; footer swap on
open/dismiss). `go test ./...`, `make check`, `make check-pty` all green — pty suite unaffected
(no dedicated Picker pty coverage yet, M7.1's job).

Next step: M2.4 — search dispatch as async work (debounced, stale-query results discarded).
First `tea.Cmd`/goroutine work in the codebase (no in-repo precedent, per
codebase-picker-integration.md) — read that finding's "no async precedent" section before
starting; `main.go`'s single `tea.ExecProcess` $EDITOR handoff is blocking and not a usable
shape here.

## Status

| Key  | Status | Reviewed | Validated | Note |
|------|--------|----------|-----------|------|
| M1.1 | done   | 2026-08-03 | 2026-08-03 |  |
| M1.2 | done   | 2026-08-03 | 2026-08-03 |  |
| M1.3 | done   | 2026-08-03 | 2026-08-03 |  |
| M1.4 | done   | 2026-08-03 | 2026-08-03 |  |
| M2.1 | done   | -        | -         |      |
| M2.2 | done   | -        | -         |      |
| M2.3 | done   | -        | -         |      |
| M2.4 | todo   | -        | -         |      |
| M2.5 | todo   | -        | -         |      |
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
