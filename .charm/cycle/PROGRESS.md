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

Current task: M1.7

## Handoff

M1.6a is done (this session): extended `resize()`'s find-dropped message to DV2
symmetrically (TRIAGE T5). `model.go`'s `resize` now captures each viewer's own
findWasEditing/findHadMatches before its `setSize` call, derives a `dv1Drop`/`dv2Drop`
string per viewer ("", "search dismissed" or "search cleared"), then combines them into
`m.message` per D59: DV1's wording stays exactly as it was (unprefixed) when DV2 drops
nothing — every pre-existing resize-message test asserting the literal text still passes
unchanged — DV2's drop is always named explicitly ("dv2 search cleared"/"dv2 search
dismissed", lowercase per D57's hint-style precedent), and a same-resize double drop joins
as `"<cause>; dv1 <drop>, dv2 <drop>"`. Added two tests to `find_test.go`:
`TestWindowResizeDropsDV2FindOnlyNamesDV2` (DV2 alone drops, message names dv2 and not dv1)
and `TestWindowResizeDroppingBothViewersFindsNamesBoth` (distinct committed finds in both
viewers, one resize, message names both). `make check` and `make check-pty` both pass.

Next: **M1.7** — extend `test/pty/harness.py`'s `KEY` table and `checks.py` with entries for
the four C3 chords (`ctrl+enter`, `ctrl+l`, `ctrl+p`, `ctrl+o`) and at least one pty check
driving `ctrl+enter` to create-and-focus DV2 and `ctrl+o` to focus it.

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
| M1.7 | todo   | -        | -         | TRIAGE T6 |
| M2.1 | todo   | -        | -         |      |
| M3.1 | todo   | -        | -         |      |
| M4.1 | todo   | -        | -         |      |
| M5.1 | todo   | -        | -         |      |
| M5.2 | todo   | -        | -         |      |
