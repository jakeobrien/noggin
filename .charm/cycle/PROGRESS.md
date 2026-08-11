# Progress — C4 recall

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

Current task: M7.2

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M6.3 built: added `keys.HistoryBack`/`keys.HistoryForward` (S9's `[`/`]` back/forward walk,
C4/M6.1) to the viewer case in both discoverability surfaces, mirroring M2.8's/M4.8's
precedent. `footer.go`'s `paneHints` default case (dv1Pane/dv2Pane) now lists them right after
`keys.Star`; `help.go`'s `helpBindings` default case lists them right after `keys.Star` too,
ahead of Edit/Navigator/Picker. Their Help() text (`back`/`forward`) collides with nothing
already listed, unlike PickerStar or FilesMode/FilesScope, so no identity-based test was
needed (a plain string check suffices). Added `TestPaneHintsNameHistoryBackAndForwardOnly-
InViewerPanes` (footer_test.go, mirrors `TestPaneHintsNamePeekOnlyInFilesPane`'s shape) and
extended `TestHelpModalListsThePaneOwnBindings`'s dv1Pane case with "back"/"forward" in `want`
and added them to rootsPane's/filesPane's `wantNot` (help_test.go).

Growing the viewer's Help Modal binding list by two pushed `test/pty/checks.py`'s M5.3
viewer-context block (previously widened to height 36 for TRIAGE T17) past its margin: the
modal's box grew from 33 to 35 rows, leaving no room for the centered modal to clear row 0,
failing "the top border is unchanged outside the modal". Widened that Session to height 38
and updated its comment's row-count math accordingly — same fix shape as T17's own history.

`make check` and `make check-pty` both green.

M7.1 built: each pane's title row (`model.go`'s `renderPane`) now carries its own
right-aligned focus-key badge — `pane.focusKey()` maps rootsPane/filesPane/dv1Pane/dv2Pane to
keys.FocusRoots/FocusFiles/FocusDV1/FocusDV2, `pane.focusBadge()` brackets the primary key and
compacts "ctrl+" to "^" (so FocusDV2 reads `[^o]`, not the much wider `[ctrl+o]`), and the new
`titleRow(title, badge, width)` helper right-aligns the badge against the title within the
pane's interior width, dropping the badge (never the title) when they don't both fit.
renderPane now builds its styled title via `titleRenderStyle.Render(titleRow(...))` instead of
`titleRenderStyle.Render(m.paneTitle(p))` directly — Peek's inverted header (S7) still applies
to the whole row unchanged, since titleRow's output is one string passed through the same
Render call. DV2's badge only ever renders when DV2 exists (renderPane is only called on
dv2Pane inside the existing `m.dv2 != nil` branch) — no extra guard needed.

Updated `TestPeekHeaderIsInvertedFromTheActiveDocuments` (peek_test.go), which had built its
expected plain/inverted title strings from the bare title text — now stale since the title row
carries a badge — to build them through `titleRow` instead, preserving the original style-
inversion assertion rather than weakening it. Added `focus_badge_test.go`: an integration test
rendering the full frame at 80x24 with DV2 open, Focus parked on filesPane, checking all four
badges appear on the right rows including three unfocused panes; a `renderPane`-level test
pinning the exact-fit-vs-one-column-short boundary (rootsPane: title 5 + separator 1 + badge 3
= 9 interior columns) confirming the badge (not the title) drops; and a direct `titleRow` unit
test pinning the same boundary. `make check` and `make check-pty` both green (check-pty's
existing screens already show the new badges verbatim in PASS output, confirming no unrelated
screen regressed). code-reviewer pass found only a doc-comment nit (fixed).

Cursor moves to M7.2 (remove the Footer's focus-key hints, now redundant with the badges).

## Status

| Key  | Status | Reviewed | Validated | Note                                     |
|------|--------|----------|-----------|------------------------------------------|
| M1.1 | done   | 2026-08-10 | 2026-08-10 |                                       |
| M1.2 | done   | 2026-08-10 | 2026-08-10 |                                       |
| M1.3 | done   | 2026-08-10 | 2026-08-10 |                                       |
| M1.4 | done   | 2026-08-10 | 2026-08-10 |                                       |
| M2.1 | done   | 2026-08-10 | 2026-08-11 |                                      |
| M2.2 | done   | 2026-08-10 | 2026-08-11 |                                      |
| M2.3 | done   | 2026-08-10 | 2026-08-11 |                                      |
| M2.4 | done   | 2026-08-10 | 2026-08-11 |                                      |
| M2.5 | done   | 2026-08-10 | 2026-08-11 | pre-agreed cut 1, not actually cut   |
| M2.6 | done   | 2026-08-10 | 2026-08-11 |                                      |
| M1.1a | done  | 2026-08-11 | 2026-08-11 | replan: pty harness fix, TRIAGE T4 (D71) |
| M2.7 | done   | 2026-08-11 | 2026-08-11 | replan: stale marker fix, TRIAGE T7 (D70) |
| M2.8 | done   | 2026-08-11 | 2026-08-11 | replan: star-binding discoverability, TRIAGE T5/T6 |
| M2.9 | done   | 2026-08-11 | 2026-08-11 | replan: naming cleanup, TRIAGE T1/T2/T9  |
| M3.1 | done   | 2026-08-11 | 2026-08-10 | SPEC/GLOSSARY-level gap re: Picker Activate not recording lastViewed — tracked as TRIAGE T12, not a M3.1 overclaim |
| M3.2 | done   | 2026-08-10 | 2026-08-10 |                                          |
| M3.3 | done   | 2026-08-10 | 2026-08-10 | replan: Picker never records lastViewed, TRIAGE T12 (D74) |
| M3.4 | done   | 2026-08-10 | 2026-08-10 | replan: recordViewed failure message clobber, TRIAGE T13; no bug found, test coverage added; review found coverage gap re: Picker call sites, resolved as M3.5 (D77) |
| M3.5 | done   | 2026-08-10 | 2026-08-11 | replan: Picker-path store-write-failure regression tests, TRIAGE T18 (D77) |
| M2.10 | done  | 2026-08-10 | 2026-08-11 | replan: pty star-marker column checks, TRIAGE T11        |
| M2.11 | done  | 2026-08-10 | 2026-08-10 | replan: Picker star marker (D72, TRIAGE T8), built same session; review found stale doc comment, resolved as M2.12 (D76); validate found Grep-mode live-toggle coverage gap, resolved as M2.13 (D78) |
| M2.12 | done  | 2026-08-10 | 2026-08-11 | replan: correct starMarker() doc comment, TRIAGE T17 (D76) |
| M2.13 | done  | 2026-08-10 | 2026-08-11 | replan: Grep same-frame star-toggle render test, TRIAGE T19 (D78) |
| M4.1 | done   | 2026-08-12 | 2026-08-11 | discoverability gap logged as TRIAGE T20; review found selection-index gap, TRIAGE T23 |
| M4.2 | done   | 2026-08-12 | 2026-08-11 | orphan-bucket cap + live-recompute judgment calls, D79; TRIAGE T21; review: T26, T29 |
| M4.3 | done   | 2026-08-12 | 2026-08-11 | flat-row Activate/Star wiring added beyond listing, D80; narrows TRIAGE T21; review: T23; pre-agreed cut 3, not actually cut (D93, cf. M2.5/M5.1) |
| M4.4 | done   | 2026-08-12 | 2026-08-11 | render test walks all six mode/scope combos; TRIAGE T22 |
| M4.5 | done   | 2026-08-12 | 2026-08-11 | Peek fixed via togglePeekFlatRowInto; Expand/Collapse no-ops; TRIAGE T21; review: T23, T24, T25 |
| M4.6 | done   | 2026-08-12 | 2026-08-11 | review found selection-index restore gap, TRIAGE T23                |
| M4.7 | done   | 2026-08-12 | 2026-08-11 | replan: flat-row selection re-clamp bug, TRIAGE T23/T24 (D82); drops T21 as already-handled; review: T37 |
| M4.8 | done   | 2026-08-12 | 2026-08-11 | replan: Browse cycle keys' footer/help discoverability, TRIAGE T20 (D81); review: T32/T33 |
| M4.9 | done   | 2026-08-12 | 2026-08-11 | replan: Tree's own empty-vs-no-Root message, TRIAGE T22; review: T34  |
| M4.10 | done  | 2026-08-12 | 2026-08-11 | replan: bundled test-coverage gaps, TRIAGE T25/T26/T29/T31 (D83); review: T36 |
| M4.11 | done  | 2026-08-12 | 2026-08-11 | replan: comment-literal and citation fixes, TRIAGE T27/T28; review: T35 |
| M4.12 | done  | 2026-08-11 | 2026-08-11 | replan: M4.8 test-coverage gaps, TRIAGE T32/T33 (D86)    |
| M4.13 | done  | 2026-08-11 | 2026-08-11 | replan: flatEmptyText default branch + tie-break comment, TRIAGE T34/T36 (D87); review: T39 |
| M4.14 | done  | 2026-08-11 | 2026-08-11 | replan: M4.7 acceptance-wording literal test gap, TRIAGE T37 (D88) |
| M5.1 | done   | 2026-08-11 | -         | delivered both File- and Grep-side Starred (D89, not actually cut, cf. M2.5); review found grepFiles dead code, TRIAGE T42 |
| M6.1 | done   | 2026-08-11 | -         |                                          |
| M6.2 | done   | 2026-08-11 | -         | pre-agreed cut 2, not actually cut (D91); review found untested re-Activate scroll-preserve behavior, TRIAGE T43 |
| M4.15 | done  | -        | -         | replan: openFolder flat-mode selection-reset bug, TRIAGE T38 |
| M5.2 | done   | -        | -         | replan: delete dead grepFiles, TRIAGE T42 (D94) |
| M6.3 | done   | -        | -         | replan: HistoryBack/Forward Footer/Help discoverability, TRIAGE T41; check-pty M5.3 viewer-block height widened 36→38 |
| M6.4 | done   | -        | -         | replan: D91 re-Activate scroll-preserve regression test, TRIAGE T43; confirmed no bug, coverage gap only |
| M7.1 | done   | 2026-08-11 | -         |                                        |
| M7.2 | todo   | -        | -         |                                          |
