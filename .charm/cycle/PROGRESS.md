# Progress — C2 find

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

Current task: M4.3

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M3.1b done this session (D52): fixed pickerRootLabel's budget formula (cmd/noggin/
picker_view.go), which subtracted the name/parent gap twice — once explicitly, once already
embedded in rest — under-computing the parent-path truncation budget by one gap width (~2
cols) on every row (T13). Folded in T17 while touching the function: pickerFileRow now builds
"  "+rel once into a local `suffix` var, reused as both pickerRootLabel's rest argument and
the literal text appended after it. Added TestPickerFileRowFillsItsFullBudget
(picker_view_test.go): pins an independently-hand-computed exact label/row string (not
derived by calling pickerRootLabel, unlike TestPickerFileResultsShowRootAndRelativePath —
T15) using synthetic non-tempdir paths, and asserts the rendered row's width equals the
layout's content width (23 at 80x24). Moved TestPickerFileResultsDisambiguateSameBasenameRoots
from 140x30 to the standard 80x24 fixture (T16); confirmed it passes there now (it failed
before the fix — both rows collapsed to a bare "…"), adjusting its assertions to check row
inequality rather than full-word substrings the tighter real budget no longer fits. `make
check` green; full Picker suite passes.

M4.2 done this session (D54): Preview pane now opens scrolled to a Grep Result's matched line
with that line highlighted (S9). Reused find.go's `EnsureVisible`/`StyleLineFunc` line-level
approach rather than `viewport.Model`'s own `SetHighlights` (consistent with
`.charm/knowledge/viewport-search-ansi-offsets.md`'s established practice, even though the
Preview's raw unstyled text isn't actually exposed to that bug) — concretely, `syncPreview`
(picker.go) now calls `SetYOffset(line-1)` plus a whole-line reverse-video `StyleLineFunc` for
a Grep target, `GotoTop()`/`nil` for a File one. Added a new `pickerPreviewTarget` (mode-aware,
covers both fileMode and grepMode) alongside the unchanged, file-mode-only
`pickerPreviewCandidate` (still what `activatePicker` alone uses — Activate stays a no-op in
grepMode until M4.3's hand-off, D34). Added a `previewLine` field paired with `previewPath` as
`syncPreview`'s re-targeting key: a plain path comparison would miss the Results cursor moving
between two matches in the *same* file, which must still re-scroll/re-highlight — caught by
writing `TestPickerGrepPreviewRetargetsWithinTheSameFile` before assuming path-only was enough.
Two new tests in picker_test.go (`TestPickerGrepPreviewOpensScrolledToMatchedLineWithHighlight`,
`TestPickerGrepPreviewRetargetsWithinTheSameFile`); all prior Picker tests still pass unchanged.
`make check` green.

Next step: M4.3 — Activate a Grep Result: file becomes the Active Document, Document Viewer
lands on the matched line. This is where D34 actually matters (re-read it first): rg's raw-file
line number must be re-resolved by occurrence ordinal against the *rendered* document via
`findMatches`/`commitFind`/`applyCurrentMatch` (find.go), not reused directly the way M4.2's
Preview (raw text, same line space as rg) could. Degrade path: query spans stripped syntax/
frontmatter → open with the query as a Committed Find on whatever matches exist (or plain if
none) — never fail the activation.

Also queued, non-blocking: M7.1a (T11, terminology housekeeping).

No blockers. TRIAGE.md empty.

## Status

| Key  | Status | Reviewed | Validated | Note |
|------|--------|----------|-----------|------|
| M1.1 | done   | 2026-08-03 | 2026-08-03 |  |
| M1.2 | done   | 2026-08-03 | 2026-08-03 |  |
| M1.3 | done   | 2026-08-03 | 2026-08-03 |  |
| M1.4 | done   | 2026-08-03 | 2026-08-03 |  |
| M2.1 | done   | 2026-08-03 | 2026-08-03 |      |
| M2.2 | done   | 2026-08-03 | 2026-08-03 |      |
| M2.3 | done   | 2026-08-03 | 2026-08-03 | M2.3a landed; re-validated |
| M2.4 | done   | 2026-08-03 | 2026-08-03 |      |
| M2.5 | done   | 2026-08-03 | 2026-08-03 |      |
| M1.4a| done   | 2026-08-03 | -         | D43 (TRIAGE T1) |
| M1.4b| done   | 2026-08-03 | -         | D45 (TRIAGE T3) |
| M1.4c| done   | -        | -         | D-none; TRIAGE T18, T19 |
| M2.1a| done   | -        | -         | D51; TRIAGE T14 |
| M2.3a| done   | 2026-08-03 | 2026-08-03 | D46 (TRIAGE T5); M2.3 re-validated |
| M3.1 | done   | 2026-08-03 | -         |      |
| M3.1a| done   | 2026-08-03 | -         | D50 (TRIAGE T10); re-reviewed, T13/T16 raised |
| M3.1b| done   | -        | -         | D52; TRIAGE T13, T15, T16, T17 |
| M3.2 | done   | 2026-08-03 | -         |      |
| M3.2a| done   | 2026-08-03 | -         | D48 (TRIAGE T6); fixed before M4.2 |
| M3.2b| done   | 2026-08-03 | -         | TRIAGE T12; re-reviewed, T20 raised |
| M3.3 | done   | 2026-08-03 | -         |      |
| M3.3a| done   | 2026-08-03 | -         | D49 (TRIAGE T7, T8) |
| M3.3b| done   | 2026-08-03 | -         | D49 (TRIAGE T9) |
| M4.1 | done   | -        | -         | D53  |
| M4.2 | done   | -        | -         | D54  |
| M4.3 | todo   | -        | -         |      |
| M4.4 | todo   | -        | -         |      |
| M5.1 | todo   | -        | -         |      |
| M5.2 | todo   | -        | -         |      |
| M5.3 | todo   | -        | -         |      |
| M6.1 | todo   | -        | -         | pre-agreed cut (D32) |
| M6.2 | todo   | -        | -         | pre-agreed cut (D32) |
| M7.1 | todo   | -        | -         |      |
| M7.1a| todo   | -        | -         | TRIAGE T11 |
| M7.2 | todo   | -        | -         |      |
