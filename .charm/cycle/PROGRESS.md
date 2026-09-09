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

Current task: M10.1

## Handoff

M9.1 (Focus by background, S11, pre-agreed cut 3) is done. Unfocused panes (rootsPane,
filesPane, dv1Pane, dv2Pane) now render on ANSI256 236, a subtly-lighter-than-black tint; the
focused pane keeps the terminal's own background; the border-colour cue (borderStyle vs
focusedBorderStyle) is untouched, per S11.

The pane's content is many independently-`Style.Render`ed runs (each nav row its own call;
DV1/DV2's open-document case is glamour-rendered markdown), so a single wrapping
`Style.Background(...)` fails outright (`lipgloss-v2-row-rendering-gotchas.md`'s own finding —
any nested `Style.Render` always closes with a bare full reset). The fix is a new function,
`tintPaneBackground` (style.go): an SGR-tracking pass over already-rendered content that
(re-)injects the tint after every reset, skips it entirely while `SGR 7` (reverse video) is in
force (a selected row, S5; Peek's inverted title, S7 — an explicit background there would
become the swapped-in text colour instead of the terminal's default pair), and — the one
genuinely new gotcha this task found — explicitly *clears* the tint immediately before every
`"\n"` it wrote, not just declines to write more: `renderBox`'s own outer `Style.Render` (the
border-drawing wrap) re-asserts whatever SGR was merely *active* at a line break, for its own
unrelated soft-wrap-continuity reasons, which was leaking the tint into the very next row's own
reverse-video escape until this was added — confirmed live via `make check-pty`, not visible
from `tintPaneBackground`'s own isolated output (see `lipgloss-reflow-sgr-continuity.md`,
condensed as this cycle's finding; D134 records the design/colour decision). `renderPane` calls
it on the title and content separately, only when `m.focus != p`.

Tests: `focus_background_test.go` — direct `tintPaneBackground` unit coverage (reset-recovery
across rows, reverse-video left alone, the cross-"\n"-into-reverse-video regression pinned
directly, and the same regression re-pinned end to end through `renderPane` since the isolated
function alone didn't show it) plus `TestUnfocusedPanesCarryTheS11BackgroundTint` — the
acceptance criterion itself, asserting the tint on every unfocused pane and its absence on the
focused one across all four focus targets. 6 new `make check-pty` checks (tint reaches the
wire for an unfocused Sources/DV1 pane at startup; moving Focus redraws both sides correctly;
the now-unfocused Files Nav's selected row keeps its reverse-video bar untinted). `make check`
and `make check-pty` (268 checks) both green.

Cursor moves to M10.1 (README, needs M1–M9 as built) — the cycle's last task. Not yet reviewed
or validated. No blockers. T26 (open, high, TRIAGE.md, filed against M8.2's location.go) is a
real bug but not a blocker — still carried over, unrelated to M9.1.

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
| M2.10 | done   | -        | -         | added by replan 2026-09-08 (TRIAGE T2) |
| M3.6  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T20) |
| M3.7  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T21) |
| M4.4  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T23) |
| M1.11 | done   | -        | -         | added by replan 2026-09-08 (TRIAGE T1, T3) |
| M5.3  | done   | -        | -         | added by replan 2026-09-08 (TRIAGE T4) |
| M6.1  | done   | 2026-09-09 | 2026-09-08 | pre-agreed cut 1 |
| M7.1  | done   | 2026-09-08 | 2026-09-08 | pre-agreed cut 2 |
| M7.2  | done   | 2026-09-08 | 2026-09-08 | pre-agreed cut 2 |
| M7.3  | done   | 2026-09-08 | 2026-09-08 | pre-agreed cut 2 |
| M8.1  | done   | 2026-09-09 | 2026-09-09 |      |
| M8.2  | done   | 2026-09-09 | 2026-09-09 | scope widened by replan 2026-09-08: S9a attribution line |
| M8.3  | done   | 2026-09-09 | 2026-09-09 |      |
| M9.1  | done   | -        | -         | pre-agreed cut 3 |
| M10.1 | todo   | -        | -         |      |
