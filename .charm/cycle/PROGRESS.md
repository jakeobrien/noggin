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

Current task: M2.1

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M1.4 done, closing M1 (the whole headless search layer): `files_search.go` gained
`grepResult{file pickerFile, line int, text string}` and `grepFiles(roots, active, scope,
query) (results []grepResult, truncated bool)`. Key call: `grepFiles` runs rg over
`listFiles`' *explicit* file list rather than letting rg walk Scope's directories itself
(D38, a deviation from M1.3's handoff assumption) — verified against the sandbox's real rg
(15.1.0) that rg only honours `.gitignore`/hidden-file rules for directories it walks, not
for files named explicitly on its command line, so a second directory-walking rg call for
grep would have leaked ignored files that the File Picker's own `listFiles` already excludes.
Literal matching is rg's `--fixed-strings`; smart case is rg's own `--smart-case`, verified
to give `find.go`'s exact S8 semantic (all-lowercase matches either case, any uppercase is
case-sensitive) — no reimplementation needed. Output is parsed from rg's `--json` JSON-Lines
format (`rgJSONMessage`), not plain `path:line:text` text, since a colon can appear in either
a path or a line's text and JSON does not leave that ambiguous. Cap/truncation is the same
shape as `fuzzyFiles` (`pickerResultCap`, D37), so both Pickers report it identically.
Verified: a query with regex metacharacters (`v1.2`) matches only its literal occurrence, not
`v1x2`; an all-lowercase query matches mixed-case text while an uppercase query does not match
a differently-cased occurrence; a non-Readable file in Scope is never searched even when it
contains the query; a match count over `pickerResultCap` comes back capped with the flag set.
`go test ./...`, `make check`, `make check-pty` all green (see D38 for the full reasoning).

Next step: M2.1 — Picker state plus open/dismiss, reachable from every Focus including while
the Navigator is hidden, always opening as File Picker / Active Root / empty Prompt. This is
the first UI-layer task (M2), building on M1's now-complete headless search layer
(`fuzzyFiles`, `grepFiles`, `listFiles`, `pickerScope`) — read
`findings/codebase-picker-integration.md` before starting (modal dispatch, compositing, the
Find-state surface a grep hand-off will reuse) if it hasn't been read yet this cycle. Two
known new-ground gaps ahead, not this task's problem yet: (1) no async/`tea.Cmd`/debounce
precedent exists anywhere in the repo — M2.4 invents it; (2) ripgrep's raw-file line numbers
don't correspond to the Document Viewer's rendered/wrapped line numbers — D34 records how
M4.3/M4.4 address that.

M6 (highlighting) is the pre-agreed cut per D32; drop it before trimming grep.

## Status

| Key  | Status | Reviewed | Validated | Note |
|------|--------|----------|-----------|------|
| M1.1 | done   | -        | -         |      |
| M1.2 | done   | -        | -         |      |
| M1.3 | done   | -        | -         |      |
| M1.4 | done   | -        | -         |      |
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
