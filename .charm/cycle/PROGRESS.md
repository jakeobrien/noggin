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

Current task: M1.3

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M1.2 done: two new headless files in `cmd/noggin` — `scope.go` (`pickerScope`,
`liveRootsInScope(roots, active, scope)`, shared by M1.2 and M1.4/M5.2 later) and
`files_search.go` (`pickerFile{root, path}`, `listFiles`, `rgFiles`). `listFiles` runs
`rg --files --no-require-git` per live root (cwd = the root, so rg's own output is already
root-relative) and filters through `tree.IsReadable` (exported from `internal/tree/policy.go`
for this — see D36). Broken Roots are skipped via `pathMissing` (rootsnav.go's own
definition, reused rather than re-derived) before rg ever runs on them; a live root rg still
can't read is absorbed the same way rather than erroring. Verified: allowlist-only results,
`.gitignore`'d + hidden files absent (needed `--no-require-git` — D36, rg only honours
`.gitignore` inside a git repo by default, verified against the sandbox's real rg 15.1.0),
All Roots covers every live root and omits a broken one without erroring, Active Root scope
ignores other roots, and each result carries the root it came from (for M3.1's row display).
`go test ./...`, `make check`, `make check-pty` all green (see D36 for the two build-time
decisions this task made).

Next step: M1.3 — fuzzy-match and rank `listFiles`' output against file names only
(`filepath.Base`, never path components — D31), capped, with a truncation flag. `listFiles`
is intentionally unranked/uncapped; M1.3 is where both land. `sahilm/fuzzy` is already a
transitive dependency (D28's note) — check whether it's already vendored/available before
reaching for scoring by hand.

Read before building UI (M2 onward): `findings/codebase-picker-integration.md`. Two known
new-ground gaps: (1) no async/`tea.Cmd`/debounce precedent exists anywhere in the repo —
M2.4 invents it; (2) ripgrep's raw-file line numbers do not correspond to the Document
Viewer's rendered/wrapped line numbers — D34 records how M4.3/M4.4 address that. M1.4 (grep)
should reuse `scope.go`'s `liveRootsInScope` and carry `--no-require-git` too (D36).

M6 (highlighting) is the pre-agreed cut per D32; drop it before trimming grep.

## Status

| Key  | Status | Reviewed | Validated | Note |
|------|--------|----------|-----------|------|
| M1.1 | done   | -        | -         |      |
| M1.2 | done   | -        | -         |      |
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
