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

Current task: M1.1

## Handoff

Plan just written (PLAN.md, M1–M5, 15 tasks covering S1–S9). Nothing built yet — this is a
fresh cycle start. Key context for whoever picks up M1.1:

- D53 (shift+enter vs. plain enter) is resolved: bind the load-into-DV2 actions to both
  `shift+enter` and a fallback key `L`; no keyboard-enhancement enable call needed. See
  DECISIONS.md D53's Resolution.
- D54 records a deliberate no-guard decision for narrow-terminal 50/50 splits — don't
  reinvent a collapse threshold for it.
- Architecture note from planning research: `viewer` (viewer.go) is already a clean,
  self-contained value type (own doc/Peek/Find/scroll) — duplicating it to `dv1`/`dv2` is
  the easy part of M1.1. The real work is routing every `m.viewer`-touching call site
  (model.go's Update, startEdit/togglePeek/startFind, handleFindInput, refresh()) through a
  focused-viewer accessor instead. `paneOrder`/`shiftFocus`/Tab bindings are being deleted,
  not extended, in the same task.
- Test files that assume a single viewer and will need rewriting rather than extending:
  `frame_test.go` (three-region geometry, Tab-cycling), `staleness_test.go`'s Footer-suffix
  cluster (superseded by M5), peek/find test fixtures (single-viewer fixtures — good
  templates for DV2 variants, not directly reusable).
- Next step: start M1.1.

## Status

| Key  | Status | Reviewed | Validated | Note |
|------|--------|----------|-----------|------|
| M1.1 | todo   | -        | -         |      |
| M1.2 | todo   | -        | -         |      |
| M1.3 | todo   | -        | -         |      |
| M1.4 | todo   | -        | -         |      |
| M1.5 | todo   | -        | -         |      |
| M1.6 | todo   | -        | -         |      |
| M2.1 | todo   | -        | -         |      |
| M3.1 | todo   | -        | -         |      |
| M4.1 | todo   | -        | -         |      |
| M5.1 | todo   | -        | -         |      |
| M5.2 | todo   | -        | -         |      |
