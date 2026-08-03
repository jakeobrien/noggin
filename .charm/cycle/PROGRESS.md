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

Current task: M2.3

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M2.2 done: new `cmd/noggin/picker_view.go` — `pickerLayout{prompt, results, preview rect}` and
`computePickerLayout(width, height)` size the modal itself (80%/80% of the terminal, clamped
to `minPickerWidth`/`minPickerHeight` and never above the terminal's own size), splitting the
remaining body between Results/Preview via `splitPickerWidth` (layout.go's `splitHeight`
fallback shape, even split if a proportional share would starve either side). `renderPicker`
composes a focused-style bordered Prompt (textinput via `promptView`, the same
"SetWidth-then-View on an addressable local copy" shape `find.go`'s `findView` already uses)
over Results/Preview boxes side by side — both empty/untitled-content placeholders for now
(M3.1/M3.2 wire real content; M2.5 wires the status/empty-state text). `overlayPicker` is
`overlayHelp`'s clamp-then-`lipgloss.NewCompositor` shape reused verbatim for the Picker.
Extracted the shared box-drawing (`renderBox`) out of `model.renderPane` so both call the same
code rather than duplicating the border+title+content shape a third time. Wired into
`model.render()`: a three-way switch (plain frame / `overlayHelp` / `overlayPicker`) replacing
the old two-way one — Help and Picker are mutually exclusive by Update's existing gating, so
at most one overlay ever applies. Visually spot-checked with an ad-hoc debug print (removed
before commit): the modal composites correctly over the live frame, centered, frame borders
visible around its edges, Prompt cursor visible.

Tests: `picker_view_test.go` — renders all three regions (Results/Preview by title text,
Prompt/Results/Preview rects non-empty from `computePickerLayout`) at 80x24/120x40/40x10;
never exceeds the terminal at those same sizes plus degenerate ones down to 1x1 and 0x0
(`TestComputePickerLayoutNeverExceedsTheTerminal`); frame stays visible around the modal's
edges (`TestHelpOverlayLeavesTheFrameVisibleAroundIt`'s exact shape). `go test ./...`,
`make check`, `make check-pty` all green — pty suite unaffected (no pty coverage for the
Picker yet, M7.1's job); the `renderPane` refactor caused no visible/behavioral change per
the unchanged pty output.

Next step: M2.3 — always-typing input dispatch: every printable key types into the Prompt;
selection movement, Preview scrolling, Activate and Cancel are on non-typing keys; Footer and
Help Modal cover the Picker's actions. `handlePickerInput` (picker.go) currently only handles
Cancel and swallows everything else — this task routes the rest into `p.input.Update(msg)`
(find.go's `updateFindQuery` is the precedent: run the query through textinput.Update, then
recompute) and adds the movement/scroll/mode/Scope-switch bindings this milestone's later
tasks (M5.1/M5.2) will act on. `footer.go`'s `paneHints`/`alwaysHints` and `help.go`'s
`helpBindings` are the precedent for the Footer/Help coverage half — note the Picker has no
`pane`, so whatever dispatches those will need a Picker-shaped variant, not a `pane` case.

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
| M2.2 | done   | -        | -         |      |
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
