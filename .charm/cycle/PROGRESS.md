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

Current task: M1.2

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M1.1 done: `cmd/noggin/ripgrep.go` (`detectRipgrep`, `ripgrepAvailability{available, reason}`)
is a pure `exec.LookPath("rg")` query, wired via `model.withRipgrep` and called once in
main.go alongside `withStorePath` — it does not touch `resolveLaunch`/`applyResolution`.
Verified with `rg` unresolvable (PATH pointed at an empty temp dir): binary still launches
and opens a document (manual pty smoke test); unit tests cover present/absent via a fake
executable on a scoped PATH (`t.Setenv`), no dependency on the sandbox's real `rg`.

Next step: M1.2 — scoped, allowlist-filtered file listing via `rg --files` (Active Root /
All Roots, Broken Roots skipped, Readable-only). No `os/exec` subprocess-output-parsing
precedent exists yet in the repo (`$EDITOR`'s `tea.ExecProcess` is blocking and unrelated);
this task is the first to actually run `rg` and parse its stdout.

Read before building UI (M2 onward): `findings/codebase-picker-integration.md`. Two known
new-ground gaps: (1) no async/`tea.Cmd`/debounce precedent exists anywhere in the repo —
M2.4 invents it; (2) ripgrep's raw-file line numbers do not correspond to the Document
Viewer's rendered/wrapped line numbers — D34 records how M4.3/M4.4 address that.

M6 (highlighting) is the pre-agreed cut per D32; drop it before trimming grep.

## Status

| Key  | Status | Reviewed | Validated | Note |
|------|--------|----------|-----------|------|
| M1.1 | done   | -        | -         |      |
| M1.2 | todo   | -        | -         |      |
| M1.3 | todo   | -        | -         |      |
| M1.4 | todo   | -        | -         |      |
| M2.1 | todo   | -        | -         |      |
| M2.2 | todo   | -        | -         |      |
| M2.3 | todo   | -        | -         |      |
| M2.4 | todo   | -        | -         |      |
| M2.5 | todo   | -        | -         |      |
| M3.1 | todo   | -        | -         |      |
| M3.2 | todo   | -        | -         |      |
| M3.3 | todo   | -        | -         |      |
| M4.1 | todo   | -        | -         |      |
| M4.2 | todo   | -        | -         |      |
| M4.3 | todo   | -        | -         |      |
| M4.4 | todo   | -        | -         |      |
| M5.1 | todo   | -        | -         |      |
| M5.2 | todo   | -        | -         |      |
| M5.3 | todo   | -        | -         |      |
| M6.1 | todo   | -        | -         | pre-agreed cut (D32) |
| M6.2 | todo   | -        | -         | pre-agreed cut (D32) |
| M7.1 | todo   | -        | -         |      |
| M7.2 | todo   | -        | -         |      |
