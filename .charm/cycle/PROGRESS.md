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

Current task: M2.7

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M1.1a done — `test/pty/checks.py`'s `config_env` (D71) now resolves the store path via
`$XDG_DATA_HOME` (darwin: `$HOME/Library/Application Support`, matching `store.go`'s
`dataDir()`) instead of the pre-M1.1 `$XDG_CONFIG_HOME`/`.config` path, returning
`XDG_DATA_HOME` in the env dict rather than `XDG_CONFIG_HOME`. Confirmed the bug was real and
is now fixed: reverting the edit and rerunning `make check-pty` reproduces D71 exactly (M3.3's
root-star-persistence checks fail, then M3.2's restart fixture crashes with `FileNotFoundError`
on the never-created config-dir store file); with the fix, every M3.2/M3.3/M3.4 check passes.
`make check` unaffected (was already green).
One unrelated pre-existing failure surfaced while running the full `make check-pty` suite: the
two `C2a/M4.1` git-status-marker checks (`test/pty/checks.py:582,584`) now fail because M2.3's
star-marker column (`files.go:367-369`) shifted the rendered spacing between the status letter
and the file name (`"A new.md"` is now `"A   new.md"`) — logged as TRIAGE T11 rather than
fixed here, out of scope for M1.1a's own task.
Gotcha still live from the previous replan: gate is `make check`; `make check-pty` is the
separate pty layer (now green apart from the newly-logged T11).
Next: **M2.7** — Files Nav marker sync when starring from a Document Viewer or the Picker
(TRIAGE T7 / D70); M2.1, M2.2 and M2.5 (its dependencies) are all done.

## Status

| Key  | Status | Reviewed | Validated | Note                                     |
|------|--------|----------|-----------|------------------------------------------|
| M1.1 | done   | 2026-08-10 | 2026-08-10 |                                       |
| M1.2 | done   | 2026-08-10 | 2026-08-10 |                                       |
| M1.3 | done   | 2026-08-10 | 2026-08-10 |                                       |
| M1.4 | done   | 2026-08-10 | 2026-08-10 |                                       |
| M2.1 | done   | 2026-08-10 | -         |                                          |
| M2.2 | done   | 2026-08-10 | -         |                                          |
| M2.3 | done   | 2026-08-10 | -         |                                          |
| M2.4 | done   | 2026-08-10 | -         |                                          |
| M2.5 | done   | 2026-08-10 | -         | pre-agreed cut 1                         |
| M2.6 | done   | 2026-08-10 | -         |                                          |
| M1.1a | done  | -        | -         | replan: pty harness fix, TRIAGE T4 (D71) |
| M2.7 | todo   | -        | -         | replan: stale marker fix, TRIAGE T7 (D70) |
| M2.8 | todo   | -        | -         | replan: star-binding discoverability, TRIAGE T5/T6 |
| M2.9 | todo   | -        | -         | replan: naming cleanup, TRIAGE T1/T2/T9  |
| M3.1 | done   | -        | -         |                                          |
| M3.2 | todo   | -        | -         |                                          |
| M4.1 | todo   | -        | -         |                                          |
| M4.2 | todo   | -        | -         |                                          |
| M4.3 | todo   | -        | -         | `Orphans` scope is pre-agreed cut 3      |
| M4.4 | todo   | -        | -         |                                          |
| M4.5 | todo   | -        | -         |                                          |
| M4.6 | todo   | -        | -         |                                          |
| M5.1 | todo   | -        | -         | Grep-side `Starred` is pre-agreed cut 4  |
| M6.1 | todo   | -        | -         |                                          |
| M6.2 | todo   | -        | -         | pre-agreed cut 2                         |
| M7.1 | todo   | -        | -         |                                          |
| M7.2 | todo   | -        | -         |                                          |
