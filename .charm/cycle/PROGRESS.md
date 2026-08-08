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

Current task: M1.4a

## Handoff

M1.3a is done (this session): `keymap.go`'s `FocusDV1`/`FocusDV2` now bind `o`/`ctrl+o` per
D55, including the stray `FocusDV2/"D"` doc-comment example that lived in the
`ActivateDV2`/`PeekDV2` block. No test hard-codes the old chords (`focus_test.go`/
`frame_test.go` go through `keys.FocusDV1.Keys()[0]` etc.), so only two comment strings
needed touching. `ctrl+o` round-trips through `keyPress`/`allBindings` with no new
`namedKeys` entry. `make check` and `make check-pty` both pass.

Next: **M1.4a** — rebind ActivateDV2 `shift+enter`/`L`→`ctrl+enter`/`ctrl+l` and PeekDV2
`shift+p`/`P`→`ctrl+p` (D55); pure rebind, `openDocumentInto`/`ensureDV2`/`togglePeekInto`
mechanics (M1.4 build, still valid, git history) are untouched.

**New discovery, TRIAGE T4**: `footer_test.go`'s `TestNoFooterHintNamesCtrlP` (a C2a/M2.2
invariant, unrelated feature) will break once PeekDV2 becomes `ctrl+p` — filesPane's
`paneHints` already lists `keys.PeekDV2`. M1.4a's build must resolve this (adjust/scope the
test), not silently ignore a new failure.

After M1.4a: **M1.5** binds `ctrl+enter`/`ctrl+l` into `handlePickerInput`'s
Activate-into-DV2 path via `openDocumentInto(dv2Pane, d)`; both chords are non-printable so
TRIAGE T3 (closed) needs no Picker-specific carve-out.

Still valid, unaffected by the rebind (from the original M1.4 build — `openDocumentInto`/
`ensureDV2`/`togglePeekInto`, full detail in git history):

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
| M1.4a| todo   | -        | -         | rebind shift+enter+L/shift+p→ctrl+enter+ctrl+l/ctrl+p, D55 |
| M1.5 | todo   | -        | -         | needs M1.4a first |
| M1.6 | todo   | -        | -         |      |
| M2.1 | todo   | -        | -         |      |
| M3.1 | todo   | -        | -         |      |
| M4.1 | todo   | -        | -         |      |
| M5.1 | todo   | -        | -         |      |
| M5.2 | todo   | -        | -         |      |
