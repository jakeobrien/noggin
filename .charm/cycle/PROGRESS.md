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

Current task: M5.2

## Handoff

M5.1 is done (this session): the staleness indicator (S10/C2a) moved off the Footer's
appended hint suffix into its own dedicated row above the Footer, per S9. `layout.go`:
added `staleRowHeight` (1) and a `staleRow rect` field; `computeLayout` gained a `stale bool`
parameter — it reserves the Footer's own height first, then gives the stale row whatever
remains (dropped silently at a terminal too short for both, same precedence the frame's
other regions already follow at extreme sizes; D60). `model.layout()` now passes `m.stale`
through. `model.frame()` inserts `renderStaleRow(l.staleRow)` between the body and the
Footer when non-empty. `renderFooter` no longer looks at `m.stale` at all — the old
width-reservation-for-a-suffix logic and `appendStaleSuffix` are both gone; the row and a
transient `m.message` now coexist (D60) since they no longer share one line. New
`staleRowStyle` (style.go): `Foreground(214).Reverse(true)`, reusing the Help Modal border's
own highlight color (D60) for S9's "highlight color, inverse text". `renderStaleRow`
pre-truncates with `rightTruncate` before `Width`/`MaxWidth` (the v2 wrap-before-truncate
gotcha, lipgloss-v2-row-rendering-gotchas.md) and fills the full row width as a highlight
bar even when the message is shorter. Rewrote `staleness_test.go`'s old Footer-suffix tests
(`TestFooterShowsStaleIndicator*`, `TestFooterMessageTakesPriorityOverStaleIndicator`,
`TestAppendStaleSuffixOmitsWhenItWouldOverflow`) as row/layout tests
(`TestStaleRowAbsentWhenNotStale`/`TestStaleRowPresentWhenStale` are the plan's own
done-when, verbatim — driven at 80 cols, a width the old suffix design never rendered at all
above ~200 cols per D53); `computeLayout` call sites across `frame_test.go`/`viewer_test.go`
gained the new `stale` argument. Updated `test/pty/checks.py`'s M10.2 live-wallclock check
to read the new row (`raw_lines()[-2]`) instead of the Footer line, plus a companion check
that the message no longer appears on the Footer itself. Verified the new tests actually
bite by regression injection (short-circuited `computeLayout`'s stale-height reservation —
three tests failed with the expected messages; reverted clean). `make check` and
`make check-pty` both pass. Logged D60 (color reuse, message coexistence, low-height
precedence — the three things S9 left for this task to decide).

Next: **M5.2** — extend the staleness baseline/check to track DV1's and DV2's Active
Documents (whichever exist) instead of one, still surfaced as the single combined signal
(effort 2, needs M1.1, M5.1).

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
| M1.6a| done   | -        | -         | TRIAGE T5, D59 |
| M1.7 | done   | -        | -         | TRIAGE T6 |
| M2.1 | done   | -        | -         |      |
| M3.1 | done   | -        | -         |      |
| M4.1 | done   | -        | -         |      |
| M5.1 | done   | -        | -         |      |
| M5.2 | todo   | -        | -         |      |
