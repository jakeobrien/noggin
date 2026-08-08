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

Current task: M1.3a

## Handoff

Replanned 2026-08-08 (TRIAGE T3 + DECISIONS D53 prompted it): the `shift+` modifier scheme
for S3 (targeting) and S5 (focus-jump) is reverted to `ctrl+`, per new **D55**. This reopens
M1.3 and M1.4 (both previously `done`) via suffixed rework tasks that now come *before* M1.5
in build order; M1.5 itself is updated in place (it hadn't been built yet, so no suffix
needed). Full rationale, and the rejected `ctrl+.` alternative, are in D55; D56 separately
disposes of T2.

- **M1.3a** (`keymap.go` + `focus_test.go`/`frame_test.go`/hint tests): FocusDV1 `d`→`o`,
  FocusDV2 `shift+d`/`D`→`ctrl+o`. Pure rebind — M1.3's behavior (no-op when DV2 absent, no
  Navigator reveal, etc., described below) is unchanged; only the literal keys and every
  place naming them (doc comments, `allBindings()`, hint/help `want` lists) move.
- **M1.4a** (`keymap.go` + `target_test.go`/hint tests): ActivateDV2 `shift+enter`/`L`→
  `ctrl+enter`/`ctrl+l`; PeekDV2 `shift+p`/`P`→`ctrl+p`. Also a pure rebind — the
  `openDocumentInto`/`ensureDV2`/`togglePeekInto` mechanics from the original M1.4 build
  (below) are untouched; only the chords change.
- **M1.5**, once M1.3a/M1.4a land: bind `ctrl+enter` (+ `ctrl+l` fallback) into
  `handlePickerInput`'s Activate-into-DV2 path, reusing `openDocumentInto(dv2Pane, d)` exactly
  as M1.4 does (see below). **TRIAGE T3 is closed by construction now**: both `ctrl+enter`
  and `ctrl+l` are non-printable control chords, so neither risks being typed into the Prompt
  by `m.picker.input.Update` — no Picker-specific carve-out needed, just one
  `case key.Matches(msg, keys.ActivateDV2):` alongside the existing `Activate` case in
  `handlePickerInput` (picker.go).
- TRIAGE cleared this session: **T1** (`.gitignore` gap) fixed directly —
  `/cmd/noggin/noggin` added. **T2** (FocusDV2/`resolveFocus` edge case) left unpinned by
  test per D56 (same reasoning as D54). **T3** closed per D55/M1.5 above.

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

Still valid, unaffected by the rebind (from the original M1.3 build):

- `model.Update` short-circuits FocusDV2 on `m.dv2 == nil` *before* calling `focusPane`, so
  it is a genuine no-op (S5) rather than a `resolveFocus`/focusOrder fallback to some other
  pane. It never creates DV2 — creation is S3's job (M1.4/M1.5's targeting modifiers).
- `focusPane`'s reveal-a-hidden-Navigator step is gated to `target == rootsPane ||
  target == filesPane` — DV1/DV2 jumps never reveal a hidden Navigator as a side effect.
- Hints: FocusDV1 is in every pane's `paneHints`/`helpBindings`; FocusDV2 is in
  `helpBindings` only (usually a no-op, same footer treatment as FindNext/FindPrevious/
  Cancel). Pinned by `TestPaneHintsOmitFocusDV2Everywhere` and related `want`-list tests —
  M1.3a must update the literal keys these tests assert, not their shape.

## Status

| Key  | Status | Reviewed | Validated | Note |
|------|--------|----------|-----------|------|
| M1.1 | done   | -        | -         |      |
| M1.2 | done   | -        | -         |      |
| M1.3 | done   | 2026-08-08 | -       |      |
| M1.3a| todo   | -        | -         | rebind d/shift+d→o/ctrl+o, D55 |
| M1.4 | done   | -        | -         |      |
| M1.4a| todo   | -        | -         | rebind shift+enter+L/shift+p→ctrl+enter+ctrl+l/ctrl+p, D55 |
| M1.5 | todo   | -        | -         | needs M1.4a first |
| M1.6 | todo   | -        | -         |      |
| M2.1 | todo   | -        | -         |      |
| M3.1 | todo   | -        | -         |      |
| M4.1 | todo   | -        | -         |      |
| M5.1 | todo   | -        | -         |      |
| M5.2 | todo   | -        | -         |      |
