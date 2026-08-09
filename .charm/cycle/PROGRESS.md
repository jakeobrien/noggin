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

Current task: M1.6a

## Handoff

Replanned (this session): triaged the five open TRIAGE.md items. Accepted T5 (resize's
find-dropped message is DV1-only, breaks S2 parity) as **M1.6a** and T6 (no pty coverage for
any C3 chord) as **M1.7**, both inserted into M1 right after M1.6, both `todo` — cursor moved
to M1.6a ahead of M2.1 since both close gaps in already-"done" M1 work before drift
compounds; M2.1 is unaffected and unblocked, just deferred a couple of tasks. T7 (stale
`viewerPane` identifier in a `checks.py` comment) fixed directly (low risk, effort-scaling
rule). T8 (lowercase `dv1`/`dv2` in help strings) and T9 (undocumented degraded-terminal
fallback) both closed wontfix — see DECISIONS.md D57/D58. TRIAGE.md is now empty.

M1.6 is done (prior session): verified DV1/DV2 Find/Edit/scroll routing already went through
`focusedViewer`/`withFocusedViewer`/`focusedViewerPane` end to end (Find's `startFind`/
`handleFindInput`/`FindNext`/`FindPrevious`, Edit's `startEdit` capturing
`focusedViewerPane()` into `editFinishedMsg.pane` for `finishEdit` to reopen into, and the
scroll case in `model.go`'s Update) — no code changes needed, this task was pure
verification plus the plan's own done-when test. Added `dv_independence_test.go`'s
`crossViewerFixture` (docA in DV1, docB in DV2, both scrollable, distinct markers) and four
tests: `TestFindIsIndependentPerViewer` (committed find driven at DV2 then DV1, each
leaves the other's query/matches alone, both directions), `TestFindEditingIsScopedToTheFocusedViewer`
(in-progress query typed at DV2 doesn't mark DV1 findEditing), `TestScrollIsIndependentPerViewer`
(scrolling one viewer's offset leaves the other's alone, both directions), and
`TestFinishEditReopensOnlyItsOwnPane` (finishEdit's `pane` field routes the reload to DV2
alone, DV1 untouched) — the write-side counterpart to `startEdit`'s already-read-only-covered
`focusedViewerPane()` capture (tea.ExecProcess's callback isn't inspectable without actually
running a subprocess, so this tests the routing directly rather than through a real edit
handoff). `make check` and `make check-pty` both pass.

Next: **M2.1** — bind `ctrl+x` globally (any pane's Focus, not while Picker or Help Modal owns
the keyboard) to close DV2: no-op if DV2 doesn't exist, Focus moves to DV1 if DV2 held it,
otherwise unchanged.

Still valid, unaffected by the M1.4/M1.4a rebind, M1.5 or M1.6 (from the original M1.4 build —
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

M1.3's own mechanics (FocusDV2's no-op on `dv2 == nil`, no Navigator reveal, hint
placement) are unaffected by M1.3a's rebind and stay pinned by the existing tests — no
further action needed there.

## Status

| Key  | Status | Reviewed | Validated | Note |
|------|--------|----------|-----------|------|
| M1.1 | done   | 2026-08-09 | 2026-08-09 |      |
| M1.2 | done   | 2026-08-09 | 2026-08-09 |      |
| M1.3 | done   | 2026-08-09 | 2026-08-09 |      |
| M1.3a| done   | 2026-08-09 | 2026-08-09 | rebind d/shift+d→o/ctrl+o, D55 |
| M1.4 | done   | 2026-08-09 | 2026-08-09 |      |
| M1.4a| done   | 2026-08-09 | 2026-08-09 | rebind shift+enter+L/shift+p→ctrl+enter+ctrl+l/ctrl+p, D55 |
| M1.5 | done   | 2026-08-09 | 2026-08-09 |      |
| M1.6 | done   | 2026-08-09 | 2026-08-09 |      |
| M1.6a| todo   | -        | -         | TRIAGE T5 |
| M1.7 | todo   | -        | -         | TRIAGE T6 |
| M2.1 | todo   | -        | -         |      |
| M3.1 | todo   | -        | -         |      |
| M4.1 | todo   | -        | -         |      |
| M5.1 | todo   | -        | -         |      |
| M5.2 | todo   | -        | -         |      |
