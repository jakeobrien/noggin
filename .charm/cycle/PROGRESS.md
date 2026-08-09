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

Current task: M2.1

## Handoff

M1.7 is done (this session, TRIAGE T6): added the four C3 chords to `test/pty/checks.py`'s
`KEY` table — `ActivateDV2CtrlEnter` (`\x1b[13;5u`, the CSI u form; see
`cycle/findings/pty-ctrl-enter-encoding.md` for why a raw byte can't stand in for
`ctrl+enter` here), `ActivateDV2CtrlL` (`\x0c`), `PeekDV2` (`\x10`), `FocusDV2` (`\x0f`) —
and a new "C3/M1.7" pty section exercising all four against a 3-file fixture
(alpha/bravo/charlie.md): ctrl+enter creates+focuses DV2 with the Files Nav selection;
ctrl+l (fallback) does the same, replacing DV2's content; ctrl+p Peeks into DV2 without
moving Focus; ctrl+o moves Focus to DV2, which also clears the Peek per S7's own
Focus-change rule (verified against the model layer's `peek_clear_test.go`) and is
confirmed by a subsequent `G` scroll landing on DV2, not DV1. Since DV1 and DV2 share the
same pane title ("Document"), the checks disambiguate by content/scroll behaviour instead
of the footer's pane name. Verified the checks actually bite by regression injection: 
temporarily dropping `ctrl+o` from `keys.FocusDV2` failed exactly the two checks built on
it; temporarily dropping `ctrl+enter` from `keys.ActivateDV2`'s key list (leaving
`ctrl+l`) failed exactly the two built on the CSI u KEY entry — both reverted clean.
`make check` and `make check-pty` both pass.

Next: **M2.1** — bind `ctrl+x` globally (any pane's Focus, not while Picker or Help Modal
owns the keyboard) to close DV2.

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
| M2.1 | todo   | -        | -         |      |
| M3.1 | todo   | -        | -         |      |
| M4.1 | todo   | -        | -         |      |
| M5.1 | todo   | -        | -         |      |
| M5.2 | todo   | -        | -         |      |
