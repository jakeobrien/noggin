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

M1.4 done. The Files Nav's S3 targeting modifiers exist and DV2 is now creatable at runtime
for the first time. Key context for whoever picks up M1.5 (the Picker's `shift+enter`):

- **Blocking-ish first**: TRIAGE **T3 (high)** — D53's `L` fallback cannot be reused in the
  Picker. `handlePickerInput` (picker.go:715) matches bindings *before* falling through to
  the always-typing Prompt, so binding a printable `L` there would swallow a literal "L"
  from every query (S2/C2). M1.5 must decide: shift+enter-only in the Picker, or a
  non-printable Picker-side fallback (`alt+enter`, `ctrl+l`). That is a decision, not an
  implementation detail — take it before writing M1.5.
- New bindings (`keymap.go`): `keys.ActivateDV2` = `"shift+enter"` + `"L"`, `keys.PeekDV2` =
  `"P"`. Note the two literal conventions: `shift+enter` uses the modifier spelling (enter is
  non-printable, no Text), `P` uses the produced character (same as `D`/`G`/`N`). Both are
  matched in `Update`'s `filesPane` case beside the plain keys they modify — no global-switch
  collisions (checked: nothing global binds L/P/shift+enter).
- **The reusable half for M1.5**: `openDocumentInto(dv2Pane, d)` now does the DV2 creation
  itself, via the new `model.ensureDV2()`. Creation *must* happen before `viewerFor(dv2Pane)`
  is read (it falls back to DV1 while `dv2 == nil`, so the load would clone DV1's size/state).
  So M1.5's Picker path is just `openDocumentInto(dv2Pane, d)` — nothing else to wire.
- `ensureDV2` re-lays out via `resize(..., resizeCauseSecondViewer)` (new cause, "second
  viewer opened"): without that the new viewer keeps `newViewer()`'s zero size and DV1 keeps
  the full width it no longer owns. It also means creating DV2 can emit resize's existing
  "…; search cleared" message when DV1 had a committed find — deliberate and honest (T43/T45
  shape), not a bug.
- Peek went per-target: `togglePeekInto(target)` (`togglePeek()` = the dv1Pane case),
  `clearPeekIn(p)` for one viewer, and `clearPeek()` now clears **both** (setFocus/refresh are
  viewer-agnostic triggers). Two peeks can coexist (DV1 + DV2) and toggle off independently.
  Refusals (folder / empty pane / read error) are all checked *before* `ensureDV2`, so a
  no-op never leaves an empty DV2 behind — pinned by test.
- Hints: both new bindings added to `paneHints(filesPane)` (footer) and
  `helpBindings(filesPane)` (help) only — Files-Nav-only, same treatment `keys.Peek` gets.
  Also added to `allBindings()` in frame_test.go (field-count test).
- Tests: new `cmd/noggin/target_test.go` (7 tests, reusing `peekFixture`). Verification:
  `make check` + `make check-pty` pass; four mutations tried and all caught (drop ensureDV2 in
  openDocumentInto → width assertions fail; clearPeek clearing only DV1; moving ensureDV2
  before the refusal checks; activateInto ignoring its target).
- Not done here, still M1.6's: DV1/DV2 Find/Edit/scroll independence.

---

Earlier (M1.3) context, still current: all four S5 focus-jump keys exist — `r`/`f` (C2a/M2.1)
plus `keys.FocusDV1` and `keys.FocusDV2`.

- `model.Update` short-circuits FocusDV2 on `m.dv2 == nil` *before* calling `focusPane`, so
  it is a genuine no-op (S5) rather than a `resolveFocus`/focusOrder fallback to some other
  pane. It never creates DV2 — creation is S3's job, i.e. M1.4/M1.5's targeting modifiers.
- `focusPane` changed shape: its reveal-a-hidden-Navigator step is now gated to
  `target == rootsPane || target == filesPane`. Previously unconditional, which would have
  made `d`/`D` reveal the sidebar as an unrequested side effect (DV1/DV2 are never inside
  the Navigator column). Existing FocusRoots/FocusFiles reveal behavior is unchanged.
- Hints: `keys.FocusDV1` was added to every pane's `paneHints` (footer.go) and
  `helpBindings` (help.go); `keys.FocusDV2` was added to `helpBindings` **only** — it is
  usually a no-op (DV2 absent), so it gets the footer's existing "inapplicable binding is
  worse than no hint" treatment (same as FindNext/FindPrevious/Cancel) with the Help Modal as
  its discoverability backstop. Both are pinned by tests
  (`TestPaneHintsOmitFocusDV2Everywhere`, and the extended `want` lists in
  `TestPaneHintsNameFocusAndRefreshFromEveryPane` /
  `TestHelpBindingsNameFocusRefreshAndToggleHiddenFromEveryPane`).
- New/changed tests in focus_test.go: `TestFocusJumpKeysMoveFocusFromEveryPane` (4x4 sweep,
  DV2 pre-created — this absorbed and replaced the old, now-subset
  `TestFocusBindingsMoveFocusFromEveryPane`), `TestFocusDV2InertWithoutDV2`,
  `TestFocusDV1AndDV2DoNotRevealHiddenNavigator`, `TestFocusDV1AndDV2InertWhilePickerIsOpen`.
  New bindings also added to `allBindings()` in frame_test.go (`TestEveryBindingIsCovered`
  counts keyMap fields, so any new binding must go there).
- Its actionable leftovers are TRIAGE T1/T2 (both low).

## Status

| Key  | Status | Reviewed | Validated | Note |
|------|--------|----------|-----------|------|
| M1.1 | done   | -        | -         |      |
| M1.2 | done   | -        | -         |      |
| M1.3 | done   | 2026-08-08 | -       |      |
| M1.4 | done   | -        | -         |      |
| M1.5 | todo   | -        | -         |      |
| M1.6 | todo   | -        | -         |      |
| M2.1 | todo   | -        | -         |      |
| M3.1 | todo   | -        | -         |      |
| M4.1 | todo   | -        | -         |      |
| M5.1 | todo   | -        | -         |      |
| M5.2 | todo   | -        | -         |      |
