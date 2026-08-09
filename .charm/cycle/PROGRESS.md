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

Current task: M1.5

## Handoff

M1.4a is done (this session): `keymap.go`'s `ActivateDV2`/`PeekDV2` now bind
`ctrl+enter`/`ctrl+l`/`ctrl+p` per D55 (was `shift+enter`/`L`/`shift+p`/`P`), plus their doc
comments in `keymap.go`/`model.go`/`target_test.go`. No test hard-codes the old chords
(`target_test.go`/`frame_test.go` go through `keys.ActivateDV2.Keys()`/`keys.PeekDV2.Keys()`),
so only comment strings and the two footer/help invariant tests needed touching. Both
`ctrl+enter` and `ctrl+l` round-trip through `keyPress` with no new `namedKeys` entry
(`enter`/`l` were already covered).

**TRIAGE T4 resolved as part of this build** (per the plan's own done-when clause, not a
separate replan decision): the pre-C3 invariant "no footer/help-modal hint may ever name
ctrl+p" (`TestNoFooterHintNamesCtrlP`, `TestNoHelpModalNamesCtrlP`) is false by design now
that D55 gives PeekDV2 that chord. Replaced both with `TestCtrlPNamedOnlyByPeekDV2`
(footer_test.go) and `TestCtrlPNamedOnlyInFilesPaneHelp` (help_test.go): ctrl+p must still
never appear for rootsPane/dv1Pane/dv2Pane or the Picker's own hints, but must appear for
filesPane (where PeekDV2 lives). One gotcha: filesPane's full `paneHints` line is 227 cells
wide, so the footer version of that assertion needs a generous width (260) to observe it
untruncated — at the project's usual 200-col test width it was already being dropped as
least-essential, unrelated to this rebind. `make check` and `make check-pty` both pass.

Next: **M1.5** — bind `ctrl+enter`/`ctrl+l` into `handlePickerInput`'s Activate-into-DV2
path via `openDocumentInto(dv2Pane, d)`, reusing `keys.ActivateDV2` (already both chords, per
D55) rather than a new binding; both chords are non-printable so TRIAGE T3 (closed) needs no
Picker-specific carve-out. Confirm typing a literal `l`/`L` into the Prompt still works
(plain `l`/`L` are untouched — only the `ctrl+` forms are ActivateDV2).

Still valid, unaffected by the M1.4/M1.4a rebind (from the original M1.4 build —
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
| M1.5 | todo   | -        | -         | needs M1.4a first |
| M1.6 | todo   | -        | -         |      |
| M2.1 | todo   | -        | -         |      |
| M3.1 | todo   | -        | -         |      |
| M4.1 | todo   | -        | -         |      |
| M5.1 | todo   | -        | -         |      |
| M5.2 | todo   | -        | -         |      |
