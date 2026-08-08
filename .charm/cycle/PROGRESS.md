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

Current task: M1.3

## Handoff

M1.2 done. `layout.go` gained `layout.splitViewer(dv2Exists bool) (dv1, dv2 rect)`: a pure
50/50 split of `l.viewer`'s width (DV1 gets the extra column on an odd width, D54's "no
minimum-width guard"), or DV1 = the full `l.viewer` / DV2 = the zero rect when `dv2Exists` is
false. Both halves share `l.viewer`'s height unchanged. Key context for whoever picks up
M1.3:

- `model.resize()`, `model.visible(dv2Pane)`/`visible(dv1Pane)` and `model.frame()` all call
  `l.splitViewer(m.dv2 != nil)` now instead of reading `l.viewer` directly for the Document
  Viewer region(s) — `l.viewer` itself is unchanged (still the combined, undivided region;
  frame_test.go's existing assertions against it still hold) and is the thing `splitViewer`
  divides, not a field that moved.
- `frame()` renders DV1 at its half, then joins DV2's render (`renderPane(dv2Pane, ...)`) to
  its right only when `m.dv2 != nil` — DV2 is still never reachable through any real key
  binding (M1.3 adds the first one, the `shift+d` focus-jump), so this only exercises through
  tests that call `m.withViewerFor(dv2Pane, newViewer())` directly.
- `resize()` now sizes `*m.dv2` (when non-nil) from its own half's `.content()`, mirroring
  DV1's `setSize` call — no find/message handling for DV2 yet (that's M1.6's per-viewer Find
  independence; DV2 cannot hold an in-progress find before M1.4 gives it a way to open one).
- New tests: `TestSplitViewerEqualHalvesWhenDV2Exists` / `TestSplitViewerDV1FullWidthWithoutDV2`
  (pure `splitViewer` unit tests, both frameSizes) and `TestFrameRendersBothViewersWhenDV2Exists`
  (model-level, creates dv2 via `withViewerFor`, checks the title row draws "Document" twice
  and `visible(dv2Pane)` is true) in frame_test.go. Verified the last one is not vacuous by
  temporarily reverting `frame()`'s split-when-exists behavior and confirming it fails.
- `make check` and `make check-pty` both pass unchanged — DV2 still being unreachable in real
  usage means no pty fixture needed updating.
- Next step: start M1.3 (direct focus-jump keys `r`/`f`/`d`/`shift+d`; `shift+d` a no-op with
  DV2 absent).

## Status

| Key  | Status | Reviewed | Validated | Note |
|------|--------|----------|-----------|------|
| M1.1 | done   | -        | -         |      |
| M1.2 | done   | -        | -         |      |
| M1.3 | todo   | -        | -         |      |
| M1.4 | todo   | -        | -         |      |
| M1.5 | todo   | -        | -         |      |
| M1.6 | todo   | -        | -         |      |
| M2.1 | todo   | -        | -         |      |
| M3.1 | todo   | -        | -         |      |
| M4.1 | todo   | -        | -         |      |
| M5.1 | todo   | -        | -         |      |
| M5.2 | todo   | -        | -         |      |
