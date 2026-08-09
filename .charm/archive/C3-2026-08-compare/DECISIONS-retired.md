# Retired decisions — C3 compare

Retired at close (2026-08-09) per `.charm/DECISIONS.md`'s retention sweep: each entry's
subject either no longer exists, was scoped to a now-complete task with no ongoing
constraint, or is fully self-contained in the shipped code/tests it explains. Moved here,
never deleted — see the `charm-close` report for the one-line reason each was retired.

## D14: Adopt Charm v2 now, at the skeleton, not later

- Status: decided
- Cycle: C1
- Date: 2026-07-31
- Context: C1 was planned against Bubble Tea v1.3.10 because v2 was pre-release at the time
  (`PROGRESS.md` handoff). It is not any more: `bubbletea/v2` v2.0.8, `bubbles/v2` v2.1.1,
  `lipgloss/v2` v2.0.5, `glamour/v2` v2.0.1 are all stable, under a new module path
  (`charm.land/…`, not `github.com/charmbracelet/…`). bubbles v2 and lipgloss v2 hard-require
  bubbletea v2, so this is one all-or-nothing swap.
- Decision: Migrate all four to v2 as a new task M1.4, ahead of M2.1, before any further
  surface is built. Re-scoped M5.1 down (3 → 2) because `viewport` v2 ships
  `SetHighlights`/`HighlightNext`/`HighlightPrevious`/`EnsureVisible`. D13's rule ("no terminal
  query after `Run()`") stands unchanged; only its cause is removed.
- Why: Cost scales with surface, and surface is at its minimum — three tasks, ~600 lines of
  non-test code, and the migration is mechanical (`View() tea.View`, `tea.KeyPressMsg` with
  `Code`/`Text`, viewport getters, `lipgloss.Color` as a function). Paying it at M2.1 is
  cheapest; paying it at M5 means redoing the tree, the roots panes and the footer. Staying on
  v1 is the expensive option, not the safe one: v1 is now the dead branch, the D13 deadlock we
  worked around is designed out of v2, `viewport`'s new highlight API is most of S8, and
  `tea.WithWindowSize`/`WithColorProfile` relieve the pty-harness pain
  (`findings/pty-verification.md`). Accepted risks: a young major with visible churn (8 patch
  releases), a pre-1.0 `ultraviolet` dependency, and the pty recipe needing re-verification
  against v2's different startup queries. Full migration surface, verified against the real
  modules, in `findings/bubbletea-v2-migration.md`. This is a dependency decision, not a spec
  change — no deliverable, criterion or key changes meaning.
- Retired: completed one-time dependency migration; the code has simply used Charm v2 ever
  since, which is now self-evident from the imports rather than something needing a standing
  decision entry to explain.

## D52: Three low-severity review nits accepted as-is

- Status: decided
- Cycle: C2a
- Date: 2026-08-09
- Context: three low-severity, non-blocking findings from code review, triaged together at
  cycle-end replan.
- Decision (wontfix, TRIAGE T3): `filesNav.expanded`'s unbounded growth across repeated
  refreshes (never pruned for paths that no longer exist) stays as-is — already documented as
  deliberate in `refreshRoot`'s own doc comment, harmless dead weight, not worth added
  bookkeeping this cycle.
- Decision (wontfix, TRIAGE T9): `ToggleHidden` keeps its name rather than being renamed to
  match its `Picker*`-prefixed siblings (`PickerUp`/`PickerDown`/`PickerMode`/`PickerScope`) —
  churn-only with no behavior change; revisit only if the Picker-only naming convention is
  otherwise being touched.
- Decision (wontfix, TRIAGE T11): `parseGitStatus`'s assumption that `git rev-parse
  --show-toplevel`'s output and a Root's own path resolve identically (e.g. through the same
  symlinks) stays untested — a divergence degrades silently to "no markers", still S8-
  compliant, not a crash; a symlinked-Root edge case sits outside S8's explicit out-of-scope
  list but isn't worth a dedicated test absent an actual bug report.
- Why: all three are genuinely low risk and low value relative to the cost of acting on them
  now; recorded so a fresh session doesn't re-raise them as new findings.
- Retired: three task-scoped, low-severity wontfix outcomes with no forward design
  implication; overdue for retirement since C2a's own close.

## D57: `dv1`/`dv2` in help/hint strings stay lowercase (TRIAGE T8)

- Status: decided
- Cycle: C3
- Date: 2026-08-09
- Context: review flagged that `FocusDV1`/`FocusDV2`/`ActivateDV2`/`PeekDV2`'s help strings
  (`keymap.go`) use lowercase `dv1`/`dv2` ("open in dv2", "peek in dv2") — the only place these
  terms surface user-facing — while GLOSSARY.md capitalizes them (`DV1`/`DV2`) as proper
  nouns, and the panes are titled "Document" in the UI, so neither casing is directly
  attested elsewhere.
- Decision: keep lowercase, as-is. Footer hints and the Help Modal already use a uniform
  lowercase terse style for every other binding ("roots", "files", "refresh", "edit",
  "find") regardless of how a term is cased in prose (GLOSSARY, SPEC, PLAN); capitalizing
  just `dv1`/`dv2` would make them the only capitalized hint in that list, a worse
  inconsistency than diverging from GLOSSARY's prose casing.
- Why: low risk and low value either way (cosmetic, four strings); the hint-style
  consistency argument favors as-is over a change, and this is now recorded so review
  doesn't re-raise it as a fresh finding.
- Retired: cosmetic wontfix (help-string capitalization), no forward constraint beyond the
  already-established lowercase hint convention documented elsewhere.

## D58: Degraded-terminal `ctrl+enter` fallback stays undocumented in README/Help (TRIAGE T9)

- Status: decided
- Cycle: C3
- Date: 2026-08-09
- Context: validate flagged that a terminal sending `ctrl+enter` byte-identical to plain
  `enter` (D53/D55's known degradation on terminals without Kitty-protocol/CSI-u/
  modifyOtherKeys support) silently targets DV1 instead of DV2, with no README or Help
  Modal text explaining why or naming the `ctrl+l` fallback as the workaround.
- Decision: not worth surfacing this cycle. `ctrl+l` is already named alongside
  `ctrl+enter` everywhere the binding appears (footer hint, Help Modal, D55), so a user who
  tries the compact hint's alternate chord is not left without a path — the gap is
  explaining *why* `ctrl+enter` alone didn't work, which is a narrower, lower-value need
  than discoverability of the fallback itself.
- Why: low risk (no functional gap, both chords already work identically) and low value
  relative to the cost of writing and maintaining terminal-capability caveats in
  user-facing docs for an edge case; revisit if a real user reports confusion.
- Retired: wontfix on a documentation gap for an edge case; no ongoing constraint on future
  code.

## D53: shift+enter's terminal disambiguation is unverified — open, deferred to build

- Status: decided
- Cycle: C3
- Date: 2026-08-08
- Context: C3/S3 uses `shift+enter` (Files Nav and Picker) to target-and-load into DV2,
  distinct from plain `enter` targeting DV1. Whether a terminal reports `shift+enter`
  differently from plain `enter` at all depends on its keyboard protocol (Kitty keyboard
  protocol / CSI-u support); many common terminals send identical bytes for both. The
  identical risk exists for `ctrl+enter` — switching modifiers doesn't change it, only
  keeps the scheme consistent with `shift+p`/`shift+d` elsewhere in C3.
- Decision: not resolved here — spec agreement does not depend on it. To be verified during
  planning/build (e.g. against the pty harness's terminal emulation and/or real terminals in
  use). If disambiguation doesn't hold up, a fallback binding for "load into DV2" is needed;
  none chosen yet.
- Why: this is a feasibility question, not a product one — resolving it doesn't change what
  S3 promises, only how reliably one of its bindings can be implemented.
- Resolution (at planning, C3): verified against `bubbletea/v2`'s source (`cursed_renderer.go`,
  `ultraviolet`'s decoder) — basic key disambiguation (the flag that makes `shift+enter`
  arrive as `KeyEnter, Mod: ModShift` rather than plain `KeyEnter`) is sent unconditionally by
  the renderer with zero app-side enabling code needed, so `shift+enter` *is* delivered as a
  distinct event on terminals implementing Kitty-protocol `CSI u` or xterm's
  `modifyOtherKeys` (kitty, WezTerm, Ghostty, foot, xterm itself). On terminals/multiplexers
  that don't (Alacritty, iTerm2, GNOME Terminal/VTE apps, Windows Terminal, Apple Terminal,
  default tmux/screen configs), the terminal's own driver has no distinct byte sequence to
  send — `shift+enter` arrives byte-identical to plain `enter`, not missing and not an error,
  so noggin has no runtime signal to detect the degradation and drive a fallback UI. Given
  that split, `shift+enter` cannot be the sole binding for load-into-DV2 (rules out accepting
  the degradation as-is). Chosen fix: bind the action to both `shift+enter` and a plain,
  unmodified fallback key, `L` (unused in `keymap.go` today; mnemonic "Load into DV2"),
  applied identically wherever `shift+enter` targets DV2 (Files Nav's load, the Picker's
  Activate) — `shift+p`'s Peek-into-DV2 and `shift+d`'s focus-jump need no such fallback,
  since shift+letter is just the capital letter and every terminal sends it as a distinct
  byte regardless of keyboard-protocol support; only shift+`enter` lacks a legacy
  disambiguation path. No enabling code (`tea.View.KeyboardEnhancements`) is needed either
  way — the base disambiguation is automatic. This is a binding choice, not a spec change
  (D2: deliverables are their actions, not their bindings) — S3's contract ("a `shift+`
  targeting modifier") is satisfied by `shift+enter` remaining the primary chord; `L` is a
  same-action alias for terminals that can't carry it.
- Retired: superseded by D55, which already carries forward everything still needed (the
  terminal-disambiguation research); the shift+enter chord itself is dead.

## D56: TRIAGE T2 (FocusDV2's `resolveFocus` fallback when DV2 is invisible) left unpinned

- Status: decided
- Cycle: C3
- Date: 2026-08-08
- Context: TRIAGE T2 (review) noted `splitViewer` has no minimum-width guard (D54), so at an
  extreme terminal width DV2's rect can be empty while `m.dv2 != nil`; FocusDV2 then falls
  through `resolveFocus`'s focusOrder rather than staying inert — correct per S5, which only
  mandates a no-op when DV2 does not exist, not when it is merely invisible — but nothing
  pins that reading with a test.
- Decision: leave it unpinned; no dedicated test added.
- Why: same reasoning as D54 — the fixed 50/50 split deliberately has no collapse behavior,
  so an unusably narrow DV2 half is already an accepted, graceful-degradation edge case, not
  a defect. The reading T2 describes is already correct per S5's own wording; a test would
  pin a reading of an edge case D54 already decided not to special-case, for a scenario
  (terminal too narrow for two useful viewer halves) nobody has reported hitting. Revisit only
  if a real report surfaces.
- Retired: a footnote fully subsumed by D54's still-live "no collapse behavior" stance; no
  independent forward constraint.

## D61: TRIAGE T11 (refresh() message-priority regression) and T13 (duplicated combining logic) folded into one fix task, M3.1a

- Status: decided
- Cycle: C3
- Date: 2026-08-09
- Context: charm-validate's cycle-end pass reproduced T11 live (already open, source
  `review`): `refresh()`'s message-combining switch (model.go:651-658) has no
  `if m.message == ""` guard, unlike `refreshViewer`'s own guarded branch — a document that
  fails to re-read while it had a committed Find loses its read-failure message to the
  find-dropped message. This contradicts `refresh()`'s own doc comment and M3.1's stated
  acceptance clause ("preserving … the message-priority behavior the single-viewer case
  already has"), and undermines SPEC.md's cycle-level acceptance bullet 5. T13 (low, review)
  separately flagged that `refresh()`'s and `resize()`'s combining switches are structurally
  identical and drift-prone.
- Decision: accept both — a new task, **M3.1a**, guards the switch (fixing T11) and extracts
  the shared combining logic into one helper used by both `refresh()` and `resize()` (fixing
  T13 as a side effect of touching the same code), rather than two separate tasks or leaving
  either for a later cycle. M3.1 itself stays `done` in PROGRESS.md (its content/scroll
  half is genuinely correct and tested) but is annotated PARTIAL pending M3.1a.
- Why: T11 is high severity and material — it reopens a milestone already marked `done`
  (M3.1) and blocks the cycle-level acceptance criterion, so it is logged rather than silently
  folded in, per the same reasoning D55 used for reopening `done` milestones. Fixing it in the
  same task as T13's refactor is the cheaper path: both changes touch the identical switch,
  and extracting the helper while adding the guard is one diff instead of two touching the
  same lines twice. Not escalated to the user: this is an implementation-correctness fix
  against SPEC.md as already agreed, not a change to what SPEC.md promises.
- Retired: task-scoped fix, fully captured by the shipped code/tests it produced (M3.1a).

## D62: TRIAGE T10 (no pty coverage for ctrl+x/CloseDV2) routed to a new task, M2.1a

- Status: decided
- Cycle: C3
- Date: 2026-08-09
- Context: M1.7 added real-binary pty coverage for the four targeting/focus chords that
  existed at the time it was built; CloseDV2's `ctrl+x` (M2.1) postdates it and was never
  added to `test/pty/checks.py`, even though unit coverage (`close_dv2_test.go`) exists.
- Decision: accept — new task **M2.1a** (PLAN.md): extend `checks.py`/`harness.py`'s KEY
  table with `ctrl+x` and a pty check driving CloseDV2 against the real binary.
- Why: medium severity per the reporter, and the gap is real (pty is noggin's only check on
  what actually reaches a terminal, per `.charm/knowledge/pty-verification.md`) — logged
  per the accept-path's medium+ rule. Low effort (1), no design question involved.
- Retired: task-scoped fix, fully captured by the shipped pty coverage it produced (M2.1a).

## D63: TRIAGE T14 (staleBaseline mutation-contract doc comments contradict closeDV2) routed to a new task, M2.1c

- Status: decided
- Cycle: C3
- Date: 2026-08-09
- Context: charm-validate's cycle-end review pass flagged that M2.1b's `closeDV2`
  (model.go:317-319) mutates `m.staleBaseline` directly via `withoutDoc`, but two standing
  doc comments it didn't update still say the baseline is only ever replaced wholesale by
  `resetStaleBaseline` or touched by `runStaleCheck`/`applyStaleResult`: the `staleBaseline`
  type's own comment ("established fresh every time … never touched in between",
  staleness.go:21-23) and `model.staleBaseline`'s field comment ("Never mutated by anything
  but resetStaleBaseline and runStaleCheck/applyStaleResult", model.go:126). Both are now
  literally false statements about the code.
- Decision: accept — new task **M2.1c** (PLAN.md): update both comments to name `closeDV2`'s
  prune-in-place call as the sanctioned exception, so the documented contract matches actual
  behavior again.
- Why: medium severity per the reporter, and it's material in the sense that a doc comment
  asserting an invariant the code no longer honors is exactly the kind of thing that misleads
  the next person to touch this file — logged per the accept-path's medium+ rule. Effort is
  trivial (1); this is a doc-only fix, not a design question, so no alternative was
  considered besides which of the two comments states the exception (decided: both, since
  both currently make the same false claim independently).
- Retired: task-scoped doc-comment fix; the reconciled comments in staleness.go/model.go now
  state this exception directly, so the decision record is redundant with the code itself.

## D64: TRIAGE T15 (redundant guard before `withoutDoc` in closeDV2) — wontfix, kept as belt-and-suspenders

- Status: decided
- Cycle: C3
- Date: 2026-08-09
- Context: `closeDV2`'s `if m.dv2.doc.isOpen()` guard before calling
  `staleBaseline.withoutDoc(m.dv2.doc.path)` (model.go:317-319) is redundant — `withoutDoc`
  already no-ops on an empty path per its own doc comment (staleness.go:103-104). Reporter
  rated it low severity, suggesting either dropping the guard or explicitly noting it's
  belt-and-suspenders.
- Decision: wontfix — leave the guard in place, as-is. Not worth a task or a code change.
- Why: low risk×value, purely stylistic, no functional difference either way (mirrors D57's
  precedent of declining a cosmetic-only change that has an equally valid alternative
  reading). The explicit guard also reads slightly clearer at the call site than relying on
  a callee's no-op contract, which is a legitimate reason to keep it beyond pure inertia.
- Retired: task-scoped wontfix, purely stylistic, no functional difference either way.
