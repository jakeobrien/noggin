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

Current task: M1.4

## Handoff

M1.3 done. All four S5 focus-jump keys now exist: `r`/`f` (unchanged, C2a/M2.1) plus new
`keys.FocusDV1` and `keys.FocusDV2`. Key context for whoever picks up M1.4:

- **Binding literals**: FocusDV1 is `"d"`; FocusDV2 is bound as the literal `"D"`, not
  `"shift+d"`. Bubble Tea v2's `Key.String()` returns the produced `Text` for a printable key
  before falling back to `Keystroke()`, so a real Shift+D reports `"D"` and `"shift+d"` would
  never match — same convention `Bottom`/`"G"` and `FindPrevious`/`"N"` already use. **M1.4/
  M1.5's `shift+enter` is the opposite case**: `enter` is non-printable (no Text), so there
  the modifier spelling `"shift+enter"` is the right literal — and D53's `L` fallback exists
  because many terminals don't report it at all.
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
- Verification: `make check` + `make check-pty` pass; both the DV2 no-op guard and the
  Navigator-reveal gate were mutation-tested (reverted, confirmed the tests fail, restored).
  A dispatched code review found no correctness bugs; its actionable leftovers are TRIAGE
  T1/T2 (both low).
- Next step: start M1.4 (Files Nav `shift+enter`/`L` and `shift+p` targeting DV2, creating it
  first if absent). Use the already-general `openDocumentInto(dv2Pane, d)` rather than
  reinventing targeting; note `togglePeek` is currently hardwired to `m.dv1` and will need a
  target parameter for `shift+p`.

## Status

| Key  | Status | Reviewed | Validated | Note |
|------|--------|----------|-----------|------|
| M1.1 | done   | -        | -         |      |
| M1.2 | done   | -        | -         |      |
| M1.3 | done   | 2026-08-08 | -       |      |
| M1.4 | todo   | -        | -         |      |
| M1.5 | todo   | -        | -         |      |
| M1.6 | todo   | -        | -         |      |
| M2.1 | todo   | -        | -         |      |
| M3.1 | todo   | -        | -         |      |
| M4.1 | todo   | -        | -         |      |
| M5.1 | todo   | -        | -         |      |
| M5.2 | todo   | -        | -         |      |
