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

Current task: M1.4

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M1.3 done: `files_search.go` gained `pickerResultCap = 200` (D37 — no SPEC/plan number
existed, both Pickers will share this constant), `fileNameSource` (a `fuzzy.Source` adapter
over `[]pickerFile` exposing only `filepath.Base`, so path components are structurally
invisible to matching — D31), and `fuzzyFiles(files, query) (results []pickerFile,
truncated bool)`. `github.com/sahilm/fuzzy` was already in the module graph (transitive via
`charm.land/bubbles/v2`, D28) but not imported by `cmd/noggin`; `go get` + `go mod tidy`
promoted it to a direct `require` — no new dependency, see D37. `fuzzy.FindFrom` already
returns best-first, so ranking is just capping its output; `fuzzyFiles` does not enforce the
3-character minimum (that's M2.4/M2.5's UI-layer job, not this headless layer's). Verified:
a tight name match outranks a scattered one (had to pick a genuinely-scattered fixture name —
a string with no valid subsequence match returns zero results, not a low-ranked one), a file
whose directory matches the query but whose base name doesn't is excluded while the reverse
is included, and a set of `pickerResultCap+50` returns exactly capped at `pickerResultCap`
with `truncated=true` (and `truncated=false` at exactly the cap). `go test ./...`, `make
check`, `make check-pty` all green.

Next step: M1.4 — literal (`--fixed-strings`) scoped grep with smart case, one result per
matching line (file, line number, line text), capped with `pickerResultCap` and the same
truncation-flag shape as `fuzzyFiles`. Reuse `scope.go`'s `liveRootsInScope` and carry
`--no-require-git` (D36) the same way `rgFiles` does. Smart case should match C1's Find
behaviour — check `find.go` for how that's implemented before reinventing it. Only Readable
files are searched (reuse `tree.IsReadable` the way `listFiles` does, or grep `listFiles`'
output directly rather than letting rg loose on the whole tree).

Read before building UI (M2 onward): `findings/codebase-picker-integration.md`. Two known
new-ground gaps: (1) no async/`tea.Cmd`/debounce precedent exists anywhere in the repo —
M2.4 invents it; (2) ripgrep's raw-file line numbers do not correspond to the Document
Viewer's rendered/wrapped line numbers — D34 records how M4.3/M4.4 address that.

M6 (highlighting) is the pre-agreed cut per D32; drop it before trimming grep.

## Status

| Key  | Status | Reviewed | Validated | Note |
|------|--------|----------|-----------|------|
| M1.1 | done   | -        | -         |      |
| M1.2 | done   | -        | -         |      |
| M1.3 | done   | -        | -         |      |
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
