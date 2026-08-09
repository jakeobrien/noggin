# Progress — C3 compare

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

Current task: M1.6

## Handoff

M1.5 is done (this session): the Picker's own home for S3's `ctrl+` targeting modifier.
`picker.go` generalizes `activatePicker()`/`activateGrepPicker()` into targeted forms
`activatePickerInto(target)`/`activateGrepPickerInto(target)` (the untargeted names now thin
wrappers calling `...Into(dv1Pane)`), threading `target` through `openDocumentInto` (fileMode)
and the pane-generic `viewerFor`/`withViewerFor` accessors in place of the old hardcoded
`m.dv1` (grepMode's `commitQuery`/`applyCurrentMatch` wiring) — no behavior change for
`dv1Pane`, DV2 support falls out for both Picker modes symmetrically. `handlePickerInput`
matches `keys.ActivateDV2` (reused as-is, no new binding) ahead of `Activate` and routes to
`activatePickerInto(dv2Pane)`; both its chords are raw control bytes so no Prompt carve-out
was needed (TRIAGE T3 stayed closed, per the plan's own prediction). New coverage in
`picker_target_test.go`: fileMode load-into-DV2 for both `ActivateDV2` keys, plain Activate
still targets DV1, literal `l`/`L` still types, and a grepMode load-into-DV2 exercising the
new `activateGrepPickerInto` find-state path (query committed and current match landed on
DV2, not DV1) — this last one goes slightly beyond the plan's literal wording (which only
names "a Picker Result" generically) but was cheap and closes a codepath the generalization
itself introduced; also added `keys.ActivateDV2` to the existing `TestPickerNonTypingKeysDoNotType`
non-typing set. `make check` and `make check-pty` both pass.

Next: **M1.6** — verify/finish DV1/DV2 Find/Edit/scroll routing so each acts only on its own
viewer's Active Document once it holds Focus (S2, S6). Much of the plumbing already exists
(`focusedViewer`/`withFocusedViewer` accessors from M1.1; the dv1Pane/dv2Pane scroll case in
`model.go`'s Update already routes through `focusedViewer()`) — this task is chiefly about
confirming Find and Edit do too, and adding the cross-viewer-independence test the plan's
done-when calls for (distinct Find state in each, drive one, assert the other untouched, both
directions).

Still valid, unaffected by the M1.4/M1.4a rebind or M1.5 (from the original M1.4 build —
`openDocumentInto`/`ensureDV2`/`togglePeekInto`, full detail in git history):

- DV2 creation happens inside `openDocumentInto`, via `model.ensureDV2()`, *before*
  `viewerFor(dv2Pane)` is read (it falls back to DV1 while `dv2 == nil`, so a later creation
  would clone DV1's size/state).
- `ensureDV2` re-lays out via `resize(..., resizeCauseSecondViewer)`; without that the new
  viewer keeps `newViewer()`'s zero size and DV1 keeps width it no longer owns. Creating DV2
  can emit resize's existing "…; search cleared" message when DV1 had a committed find —
  deliberate (T43/T45 shape), not a bug.
- Peek is per-target: `togglePeekInto(target)` (`togglePeek()` = the dv1Pane case),
  `clearPeekIn(p)` for one viewer, `clearPeek()` clears **both**. Refusals (folder / empty
  pane / read error) are checked *before* `ensureDV2`, so a no-op never leaves an empty DV2
  behind — pinned by test.
- Not done here, still M1.6's: DV1/DV2 Find/Edit/scroll independence.

M1.3's own mechanics (FocusDV2's no-op on `dv2 == nil`, no Navigator reveal, hint
placement) are unaffected by M1.3a's rebind and stay pinned by the existing tests — no
further action needed there.

## Status

| Key  | Status | Reviewed | Validated | Note |
|------|--------|----------|-----------|------|
| M1.1 | done   | -        | -         |      |
| M1.2 | done   | -        | -         |      |
| M1.3 | done   | 2026-08-08 | -       |      |
| M1.3a| done   | -        | -         | rebind d/shift+d→o/ctrl+o, D55 |
| M1.4 | done   | -        | -         |      |
| M1.4a| done   | -        | -         | rebind shift+enter+L/shift+p→ctrl+enter+ctrl+l/ctrl+p, D55 |
| M1.5 | done   | -        | -         |      |
| M1.6 | todo   | -        | -         |      |
| M2.1 | todo   | -        | -         |      |
| M3.1 | todo   | -        | -         |      |
| M4.1 | todo   | -        | -         |      |
| M5.1 | todo   | -        | -         |      |
| M5.2 | todo   | -        | -         |      |
