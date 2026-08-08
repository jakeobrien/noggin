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

Current task: M1.2

## Handoff

M1.1 done. `viewer` is now a `dv1`/`dv2` pair (`dv2 *viewer`, nil = does not exist yet);
`pane` gained `dv1Pane`/`dv2Pane` replacing `viewerPane`; `paneOrder`/`shiftFocus`/`indexOfPane`
and the Tab/Shift+Tab bindings are gone entirely (S5). Key context for whoever picks up M1.2:

- The accessor shape landed in model.go: `viewerFor(pane)`/`withViewerFor` (deterministic
  pane→viewer mapping, creates dv2 on first write), `focusedViewer()`/`withFocusedViewer`
  (S6's "whichever viewer holds Focus, DV1 if Focus is elsewhere" — used by scroll, Find,
  FindNext/FindPrevious, Edit/finishEdit), and `focusedViewerPane()`. Peek, a plain
  Activate/openDocument and refresh() are hardwired to `m.dv1` directly instead — their
  target is never "whichever pane has Focus" (S3), so they don't use the focused-viewer
  accessor; a later shift+-modified action (M1.4/M1.5) should call `openDocumentInto`
  (already general) rather than reinvent targeting.
- `resolveFocus` now falls back through `focusOrder = []pane{rootsPane, filesPane, dv1Pane}`
  — dv2Pane is deliberately never a fallback target (see its doc comment).
- `editFinishedMsg` gained a `pane` field so `finishEdit` reopens into whichever viewer was
  edited; `openDocumentInto` normalizes any non-`dv2Pane` value (including pane's own
  zero-value `rootsPane`, from older test literals) to `dv1Pane`, so this is safe even where
  callers never set it.
- `frame()`/`renderPane`/`paneTitle` render DV1 alone in the single `l.viewer` region — M1.2's
  own job is splitting `layout.viewer` into a DV1/DV2 pair and rendering both; `paneTitle`/
  `renderPane`'s Peek-style check already generalizes via `viewerFor(p)`, so it should need no
  further change when dv2Pane starts rendering.
- `test/pty/checks.py` also referenced Tab/Shift+Tab pane-cycling in several places (M1.2's
  own pty checks, M3.3/M3.4/M4.2/M5.2 fixtures) — all rewritten to FocusRoots/FocusFiles or to
  assert the keys are now inert. `make check-pty` passes end to end again.
- Next step: start M1.2 (split `layout.go`/`model.frame()` 50/50 when DV2 exists).

## Status

| Key  | Status | Reviewed | Validated | Note |
|------|--------|----------|-----------|------|
| M1.1 | done   | -        | -         |      |
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
