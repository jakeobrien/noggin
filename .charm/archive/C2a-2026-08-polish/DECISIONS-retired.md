# Retired decisions — C2a polish

Retired at close (2026-08-10) per `.charm/DECISIONS.md`'s retention sweep: each entry's
subject either no longer exists, was scoped to a now-complete task with no ongoing
constraint, or is fully self-contained in the shipped code/tests it explains. Moved here,
never deleted — see the `charm-close` report for the one-line reason each was retired.

## D45: C2a's plan orders tree-state memory and git markers ahead of refresh

- Status: decided
- Cycle: C2a
- Date: 2026-08-05
- Context: Planning C2a/PLAN.md. SPEC.md's S2 (manual refresh) explicitly reuses S3's
  per-Root preserve mechanism ("the same mechanism S2 needs") and recomputes S8's git markers
  "in the same pass". A researcher subagent's pass over the codebase confirmed neither
  mechanism exists yet: Files Nav state is a single global `expanded`/`selected` pair wiped on
  every Root switch (`files.go`), and no git tooling exists at all.
- Decision: milestone order is M3 (S3, per-Root tree-state memory) and M4 (S8, per-file git
  markers) ahead of M5 (S2, refresh), with M5 depending on both. M9 (S9, footer/Help Modal
  accuracy) is ordered last of the binding-affecting milestones (M2, M5, M6, M7) since it is a
  sweep over bindings those define, not new bindings itself.
- Why: building refresh first would mean redoing its preserve/recompute logic once S3 and S8
  land; low risk, purely a sequencing call, but worth recording since a later replan or a
  fresh session might otherwise wonder why S2 — the deliverable most of the cycle's pitch is
  about — is M5 and not M2.

## D48: M9.1's "no remaining reference to ctrl+p anywhere" is scoped to the Footer/Help Modal, not README

- Status: decided
- Cycle: C2a
- Date: 2026-08-07
- Context: PLAN.md M9.1 and SPEC.md S9's done-when both read literally as "no remaining
  reference to `ctrl+p` anywhere," with no explicit carve-out. README.md still documents
  `ctrl+p` opening the Picker in three places (bindings table, section heading, prose) — a
  gap TRIAGE T1 flagged back at M2.2 as outside that task's scope, on the grounds that "no
  plan task currently owns README the way M9.1 owns Footer/Help Modal." T6/T15 later
  extended the same reasoning to `alt+h` and Peek documentation gaps.
- Decision: M9.1's own scope is read as the Footer (`paneHints`/`pickerHints`, footer.go)
  and the Help Modal (`helpBindings`, help.go) only — the two artifacts S9 names by name
  ("correct in the Footer hints and the Help Modal for each pane"). README.md is left
  untouched; T1/T6/T15 remain open in TRIAGE.md for a future README-sync task to resolve
  together, rather than picked off piecemeal by whichever task happens to touch an adjacent
  binding.
- Why: T1 already recorded this reading at M2.2's own handoff, before M9.1 was ever worked;
  M9.1 implementing it consistently is following an existing call, not making a new one.
  Splitting README fixes across three different tasks (M2.2/M6.2/M9.1) each touching one
  binding would leave the doc updated in dribs with no task actually verifying its overall
  consistency — a dedicated pass, once T1/T6/T15 are triaged together, does that once.
  Recorded explicitly (not left implicit in TRIAGE rows alone) since a fresh session reading
  PLAN.md/SPEC.md's literal "anywhere" wording could otherwise reasonably read M9.1 as
  incomplete without this cross-reference.
- Retired: the gap it tracked (T1/T6/T15) is exactly what M11.1/M11.2 closed this cycle —
  subject resolved, no ongoing constraint.

## D50: M6.2's " · Hidden" Prompt-title suffix is a deliberate, minimal UI affordance

- Status: decided
- Cycle: C2a
- Date: 2026-08-09
- Context: M6.2/S4's `alt+h` dotfile-hiding toggle shipped with a " · Hidden" suffix on
  `pickerModeScope`'s title (`picker_view.go`) — not mandated by S4 or M6.2's own done-when,
  which only require the toggle to hide/show Results. A code-review pass (TRIAGE T12) flagged
  it as reasonable but undocumented, and separately (TRIAGE T7) found the rendering path had
  no `go test` coverage — only the more fragile `test/pty/checks.py` exercised the string.
- Decision: keep the suffix as shipped — it reuses the existing mode/Scope indicator
  mechanism, adds no new binding or state, and gives the toggle its only user-visible
  confirmation beyond the Results list itself changing. Coverage gap addressed by adding a
  `hideDotfiles: true` case to `TestPickerStatusStates` (picker_view_test.go), tracked as
  PLAN.md M12.1.
- Why: a minimal, reuse-based indicator is cheap and low-risk; recording it here (rather than
  leaving it discoverable only in the code comment) is what T12 actually asked for.
- Retired: scoped to closing out a review nit (T12) whose coverage gap M12.1 already
  backfilled; the rationale is now self-evident from the shipped code/tests.

## D51: Peek's Find-continuation gap is fixed, not accepted

- Status: decided
- Cycle: C2a
- Date: 2026-08-09
- Context: cycle-end validation (and TRIAGE T13, high severity) found that `FindNext`/
  `FindPrevious` (`n`/`N`, model.go) are matched in `Update`'s global switch ungated by
  `m.viewer.peeking()`, unlike `startEdit`/`startFind` a few lines above which both got
  explicit `peeking()` guards in the same M7.1/M7.2 diff. A pre-existing committed Find on the
  Active Document lets `n`/`N` silently drift the real viewport's YOffset while Peeking
  (invisible on screen, since the view shows `peek.vp`), and `clearPeek()` never touches that
  offset, so clearing Peek afterward restores the Active Document at the drifted position, not
  the pre-Peek one — violating S7's "cannot be searched with Find" and the cycle-level
  acceptance bullet on scroll restoration, in that specific reachable scenario. D44 named this
  exact class of risk generally ("collides with globals defined on the Active Document"),
  sharpest for Edit; the same reasoning covers Find continuation, which D44/M7.1/M7.2 did not
  actually test.
- Decision: gate `FindNext`/`FindPrevious` with `!m.viewer.peeking()`, mirroring
  `startEdit`/`startFind`'s existing guard shape, plus a regression test. Bundled in the same
  task (PLAN.md M7.3, reopening M7.1/M7.2's area rather than a fresh milestone): route
  `startFind`'s remaining direct `m.focus = ...` assignment through `model.setFocus` for
  consistency with every other call site (TRIAGE T14 — harmless today only because
  `startFind`'s own `peeking()` guard always returns first, but the one remaining site the
  `setFocus` chokepoint's own stated invariant doesn't structurally cover); and add a one-line
  comment documenting that `clearPeek()` firing on a boundary-clamped `LineUp`/`LineDown`
  (selection index unchanged) is an accepted broadening of "any selection change", not a bug
  (TRIAGE T16, functionally harmless — re-arriving at the same file just re-shows the Active
  Document).
- Why: this is a real, reachable spec violation (not a hypothetical), cheap to fix with the
  same guard shape already established two lines away, and the two low-severity nits sit in
  the exact code this fix already touches.
- Retired: scoped to the now-complete M7.3 fix; the behavior is enforced by its own regression
  test, no separate doc trail needed.

## D53: S10's staleness indicator is a real footer-width bug, not the documented detection gap

- Status: decided
- Cycle: C2a
- Date: 2026-08-10
- Context: user report, post-"ready to close" ("the background monitoring that should
  indicate when a file has changed does not seem to be working"). A researcher pass
  static-traced `staleness.go`'s scheduling/baseline/wiring and found it sound. A validator
  pass then live-repro'd via the pty harness at real wall-clock intervals and multiple
  terminal widths and found the actual defect one layer up, in footer composition:
  `appendStaleSuffix` (model.go:979-985) requires its entire ~48-char suffix to fit in
  whatever slack remains after `hintText` (footer.go) has already greedily filled the width
  with key hints, with no partial truncation of hints to make room. Measured thresholds:
  the suffix never fits below roughly 200-244 columns depending on pane — i.e. never, at
  every width this project's own `staleness_test.go` (200/400 cols) and `test/pty/checks.py`
  fixtures actually use. The underlying tick-reschedule/baseline-reset/generation-discard
  logic (D49) all fired correctly in the live repro once the width was widened past the
  threshold, and the documented in-place-edit-to-a-non-Active-tracked-file gap (staleness.go
  doc comment) reproduced exactly as designed — confirming the bug is footer rendering,
  not detection.
- Decision: accept as a real functional gap against M10.1/S10's own acceptance criterion
  ("surfaces the indicator within one check interval"), not a wontfix or a documented
  limitation — reopen M10 with a new task, PLAN.md M10.2, to reserve room for the suffix
  ahead of `hintText`'s greedy fill (or fall back to a minimal hint set while stale) and add
  a realistic-width live check to `test/pty/checks.py` so this class of gap is caught by
  `make check-pty` going forward. PROGRESS.md's M10.1 Validated stamp is cleared back to `-`:
  the 2026-08-08 validate pass exercised the render only at `staleness_test.go`'s existing
  200/400-column cases, never a live run at an ordinary terminal size, so it did not actually
  cover this clause of the acceptance criterion despite stamping it validated.
- Why: high value (a shipped, spec'd cycle deliverable is invisible in ordinary use — the
  cycle's own "ready to close" state was wrong) and low/medium risk (the fix is localized to
  footer composition; the detection engine is independently confirmed sound, so this doesn't
  reopen M10.1's harder design decisions, D49's heuristic or its indicator-surface choice) —
  accept and fix, don't defer. Logged because it reopens a milestone PROGRESS.md had marked
  both done and validated; a fresh session would otherwise trust that stamp without this
  trail explaining why it's now `-` again.
- Retired: scoped to the now-complete, re-validated M10.2; the guard against regressing it
  lives in the code itself (`renderFooter`'s reserved-width calculation), not as forward-
  looking knowledge.
