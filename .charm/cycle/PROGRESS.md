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

Current task: M2.2

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M2.1 done: new `cmd/noggin/picker.go` — `picker{open, mode, scope, input textinput.Model}`,
`newPicker()` (closed/File/ActiveRoot/empty), `model.openPicker()`/`dismissPicker()`
(replace-the-whole-struct, so S8 "no memory" holds by construction, not by clearing fields),
and `handlePickerInput` (Cancel-only for now, same shape as `handleHelpInput`). Wired into
`model.go`: new `picker picker` field, a global `keys.Picker` case (`ctrl+p`, D39) in the same
position as Edit/Find/Navigator/Help, and a `m.picker.open` gate ahead of the main switch so
an open Picker owns the keyboard the same exclusive way `findEditing()`/`helpOpen` already do.
No rendering/compositing yet (M2.2's job) — `render()` does not consult `m.picker` at all, so
the "frame restored byte-for-byte after dismissal" criterion holds trivially for now and
becomes a real test once M2.2 adds the overlay. Tests: `picker_test.go` — opens from every
Focus and with the Navigator hidden, always resets to File/ActiveRoot/empty-Prompt regardless
of what was perturbed before dismissal, dismissal restores Active Document/Root/scroll offset
and the exact rendered frame, and `q` while open does not quit. `allBindings()`/`keyMap`
field-count test updated for the new binding. `go test ./...`, `make check`, `make check-pty`
all green (no pty coverage added yet — M7.1's job).

Next step: M2.2 — modal layout and compositing (Prompt/Results/Preview inside one modal
composited over the live frame, clamped to the terminal, `TestHelpOverlayNeverExceedsTheTerminal`-shaped
test). `help.go`'s `overlayHelp` (clamp with MaxWidth/MaxHeight, then
`lipgloss.NewCompositor`) is the direct precedent to reuse/adapt — see
`findings/codebase-picker-integration.md`'s "Compositing and layout" section. Results/Preview
content itself can stay empty placeholders; M3.1/M3.2 wire real content. No new-ground gaps
for this task specifically — M2.4 (async `tea.Cmd`/debounce, no in-repo precedent) and
M4.3/M4.4 (rg raw-line vs. rendered-viewport-line addressing, D34) are later tasks' problems.
M6 (highlighting) is the pre-agreed cut per D32; drop it before trimming grep if the cycle
runs long.

## Status

| Key  | Status | Reviewed | Validated | Note |
|------|--------|----------|-----------|------|
| M1.1 | done   | 2026-08-03 | 2026-08-03 |  |
| M1.2 | done   | 2026-08-03 | 2026-08-03 |  |
| M1.3 | done   | 2026-08-03 | 2026-08-03 |  |
| M1.4 | done   | 2026-08-03 | 2026-08-03 |  |
| M2.1 | done   | -        | -         |      |
| M2.2 | todo   | -        | -         |      |
| M2.3 | todo   | -        | -         |      |
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
