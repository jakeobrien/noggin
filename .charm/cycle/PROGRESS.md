# Progress — C6 sources

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

Current task: M1.11

## Handoff

M8.3 (de-emphasized secondary text, S10/D117) is done, completing milestone M8: the Sources
Nav's path line (line 2, both a Root row's and the Wikipedia row's) renders de-emphasized via
`rootsPathStyle = Foreground(dimColor)` (rootsnav.go); a File Picker row's path segment (from
the location label's own trailing gap onward) renders the same way, with a fuzzy-matched
character never dimmed even inside that segment (`highlightMatches`' new `dimFrom int`
parameter and per-column `class` priority — matched beats dim beats base — picker_view.go);
`pickerGrepRow` passes the new `noDim` sentinel, unaffected. D117's own change — dropping the
Root's parent-path fragment from Picker File rows entirely — landed in the same task:
`pickerRootLabel` is now just `rightTruncate(f.root.displayName(), budget)`, no parent segment;
two now-obsolete disambiguation/budget tests tied to that fragment
(`TestPickerFileResultsDisambiguateSameBasenameRoots`, `TestPickerFileRowFillsItsFullBudget`)
are retired with an explanatory note in picker_view_test.go, replaced by
`TestPickerRootLabelTruncatesALongName` (the label's own remaining truncation case) and three
new S10-specific tests (`TestPickerFileRowDropsTheParentPathFragment`,
`TestPickerFileRowDimsThePathButNotTheRootLabel`,
`TestPickerFileRowMatchedPathCharactersAreNeverDimmed`). `TestRootsNavPathLineIsDeEmphasized`
covers the Sources Nav half. `cellStyles` (picker_view_test.go) gained a `dim bool` field and
ANSI256-foreground-sequence parsing (`38;5;<n>` vs `dimColorCode`), reused by a new
`dimColumns` helper and by rootsnav_test.go's new `lineFullyReversed` (a fresh-context review
found the two rewritten reverse-video assertions in `TestRootsNavHighlightsActiveRoot`/
`TestRootsNavSelectionHighlightsBothLines` were only passing on a lucky padding-run coincidence
once a combined-attribute row filled its width exactly — confirmed by tracing lipgloss's own
SGR serialization directly). 5 new `make check-pty` checks added (S10 dim-styling-reaches-
the-wire + fragment absence). A second reviewer pass (stale-comment sweep across
picker_view.go/picker_view_test.go/rootlabel_test.go/files_search.go, all fixed) found no
further issues. `make check` and `make check-pty` (~255 checks) both green.

Milestone M8 (M8.1/M8.2/M8.3) is now all `done` but none of the three rows carry a `Validated`
date and only M8.3 (this task) has a `Reviewed` date — per cycle-state.md's state table, the
next build session's own orient gate check will hit "current milestone complete, validation
due" and should surface it rather than pushing further into M9 first (see Status below).

Cursor is now M1.11 (the `make check-pty` string-drift fix, pulled forward ahead of M9.1 per
the milestone-order note at the top of PLAN.md — M9.1's own acceptance criterion leans on that
same check-pty gate). M2.10 and M5.3 remain independent replan-added hygiene tasks,
pull-forward-anytime. Working tree: M8.3's code+test changes are committed as this task's own
commit.

No blockers.

## Status

| Key   | Status | Reviewed | Validated | Note |
|-------|--------|----------|-----------|------|
| M1.1  | done   | 2026-09-08 | 2026-09-08 |      |
| M1.2  | done   | 2026-09-08 | 2026-09-08 |      |
| M1.3  | done   | 2026-09-08 | 2026-09-08 |      |
| M1.4  | done   | 2026-09-08 | 2026-09-08 |      |
| M1.5  | done   | 2026-09-08 | 2026-09-08 |      |
| M2.1  | done   | 2026-09-08 | 2026-09-08 |      |
| M2.2  | done   | 2026-09-08 | 2026-09-08 |      |
| M2.3  | done   | 2026-09-08 | 2026-09-08 |      |
| M1.6  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T1, T2) |
| M1.7  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T3) |
| M2.4  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T4, T5) |
| M2.5  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T6, T7, T8) |
| M2.6  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T10) |
| M2.7  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T11, T12) |
| M1.8  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T17) |
| M2.8  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T15) |
| M3.1  | done   | 2026-09-08 | 2026-09-08 |      |
| M3.2  | done   | 2026-09-08 | 2026-09-08 |      |
| M3.3  | done   | 2026-09-08 | 2026-09-08 |      |
| M3.4  | done   | 2026-09-08 | 2026-09-08 |      |
| M3.5  | done   | 2026-09-08 | 2026-09-08 |      |
| M4.1  | done   | 2026-09-08 | 2026-09-08 |      |
| M4.2  | done   | 2026-09-08 | 2026-09-08 |      |
| M4.3  | done   | 2026-09-08 | 2026-09-08 |      |
| M5.1  | done   | 2026-09-08 | 2026-09-08 |      |
| M5.2  | done   | 2026-09-08 | 2026-09-08 |      |
| M1.9  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T18) |
| M1.10 | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T2) |
| M2.9  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T1, T22) |
| M2.10 | todo   | -        | -         | added by replan 2026-09-08 (TRIAGE T2) |
| M3.6  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T20) |
| M3.7  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T21) |
| M4.4  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T23) |
| M1.11 | todo   | -        | -         | added by replan 2026-09-08 (TRIAGE T1, T3) |
| M5.3  | todo   | -        | -         | added by replan 2026-09-08 (TRIAGE T4) |
| M6.1  | done   | 2026-09-09 | 2026-09-08 | pre-agreed cut 1 |
| M7.1  | done   | 2026-09-08 | 2026-09-08 | pre-agreed cut 2 |
| M7.2  | done   | 2026-09-08 | 2026-09-08 | pre-agreed cut 2 |
| M7.3  | done   | 2026-09-08 | 2026-09-08 | pre-agreed cut 2 |
| M8.1  | done   | -        | -         |      |
| M8.2  | done   | -        | -         | scope widened by replan 2026-09-08: S9a attribution line |
| M8.3  | done   | 2026-09-09 | -       |      |
| M9.1  | todo   | -        | -         | pre-agreed cut 3 |
| M10.1 | todo   | -        | -         |      |
