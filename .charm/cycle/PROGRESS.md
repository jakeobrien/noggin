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

Current task: M1.3

## Handoff

M1.1 done: `rootsNav` gained a `wikipedia bool` Active Source flag (`activateWikipedia`/
`activeIsWikipedia`, mutually exclusive with a valid `active` index) and `model.
activateWikipediaSource()`; every Root-only subsystem (tree walk, grep, git markers,
staleness, Refresh) already tolerated `active < 0`/`m.files.root == nil` before this change,
so none needed touching. `m.files.root` is deliberately left untouched when Wikipedia is
activated (still shows the prior Root's tree if one was loaded) — that is M1.4's job.

M1.2 done: the Sources Nav pane (`rootsnav.go`) always shows a trailing Wikipedia row
(`rowCount` = Roots + 1, never zero), removing the old "No roots…" empty state; Activate on
it calls `activateWikipediaSource`, `s` on it sets a Footer message instead of falling into
`activateRoot`/`toggleStar`; broken/starred markers are structurally unreachable for it
(`renderWikipediaRow` reads no per-row state). Both reviewed clean by code-reviewer subagent.

Logged to TRIAGE (T1, low): the pane border/footer title still reads "Roots", not "Sources"
per GLOSSARY's C6 rename — no M1.x task's acceptance criterion currently covers retitling it.

Next: M1.3 (relabel "Active Root" → "Active Source" wherever a Query Source can be active) is
queued next — needs M1.1, now done. M2 (Wikipedia client) is independent of M1 and may be
interleaved instead. Read `findings/wikimedia-api.md` before M2.2/M2.3. D120 covers the spike
extraction (M2.1).

## Status

| Key   | Status | Reviewed | Validated | Note |
|-------|--------|----------|-----------|------|
| M1.1  | done   | 2026-09-07 | -       |      |
| M1.2  | done   | 2026-09-07 | -       |      |
| M1.3  | todo   | -        | -         |      |
| M1.4  | todo   | -        | -         |      |
| M1.5  | todo   | -        | -         |      |
| M2.1  | todo   | -        | -         |      |
| M2.2  | todo   | -        | -         |      |
| M2.3  | todo   | -        | -         |      |
| M3.1  | todo   | -        | -         |      |
| M3.2  | todo   | -        | -         |      |
| M3.3  | todo   | -        | -         |      |
| M3.4  | todo   | -        | -         |      |
| M3.5  | todo   | -        | -         |      |
| M4.1  | todo   | -        | -         |      |
| M4.2  | todo   | -        | -         |      |
| M4.3  | todo   | -        | -         |      |
| M5.1  | todo   | -        | -         |      |
| M5.2  | todo   | -        | -         |      |
| M6.1  | todo   | -        | -         | pre-agreed cut 1 |
| M7.1  | todo   | -        | -         | pre-agreed cut 2 |
| M7.2  | todo   | -        | -         | pre-agreed cut 2 |
| M7.3  | todo   | -        | -         | pre-agreed cut 2 |
| M8.1  | todo   | -        | -         |      |
| M8.2  | todo   | -        | -         |      |
| M8.3  | todo   | -        | -         |      |
| M9.1  | todo   | -        | -         | pre-agreed cut 3 |
| M10.1 | todo   | -        | -         |      |
