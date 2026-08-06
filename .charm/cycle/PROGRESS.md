# Progress — C2a polish

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

M1.1 done: `make install` (new target) copies the already-built `bin/noggin` to
`$(PREFIX)/bin/noggin`, `PREFIX ?= /usr/local`, and deliberately has no `build` prerequisite
(Makefile:14-15,29-33). Verified: `make build && make install PREFIX=/tmp/x` places a working
binary there; `make -n install` shows no `go build` step; default `PREFIX=/usr/local` also
works (needs root in this sandbox since `/usr/local` is root-owned here — a sandbox fact, not
a product concern). No findings or triage items from this task.

Sequencing rationale (not yet worth a DECISIONS.md entry, but useful to a fresh session):
M3 (tree-state memory, S3) and M4 (git markers, S8) are ordered ahead of M5 (refresh, S2)
because S2 explicitly reuses S3's per-Root preserve mechanism and recomputes S8's markers in
the same pass — building refresh first would mean redoing it once those land. M9 (footer/help,
S9) is last of the binding-affecting work since it is a sweep over bindings M2/M5/M6/M7 define.
M10 (staleness, S10) is last per the spec's own "pre-agreed cut, dropped without ceremony".

Known foot-guns flagged by research, worth re-reading before M3/M5/M7:
- Files Nav `selected` is a flattened-row index today, not a stable node key — M3/M5 need to
  key state by path, not index (files.go).
- `viewer.open`/`filesNav.setRoot` both reset state (scroll-to-top / collapse-all) — M5 must
  bypass them, not call them, to preserve position.
- A dotfile-toggle (M6.2) must live on `model`, not `picker` — `resetPicker` wipes anything
  stored on `picker` every close/reopen.
- Peek (M7) touches many call sites (shiftFocus, openPicker, refresh, selection change) — each
  is a place its "clear on X" rule can be forgotten.

Next step: M2.1 — `r`/`f` focus bindings; check keymap.go and nav.go for existing focus-shift
plumbing before adding new bindings.

## Status

| Key   | Status | Reviewed | Validated | Note |
|-------|--------|----------|-----------|------|
| M1.1  | done   | -        | -         |      |
| M2.1  | todo   | -        | -         |      |
| M2.2  | todo   | -        | -         |      |
| M3.1  | todo   | -        | -         |      |
| M4.1  | todo   | -        | -         |      |
| M5.1  | todo   | -        | -         |      |
| M6.1  | todo   | -        | -         |      |
| M6.2  | todo   | -        | -         |      |
| M7.1  | todo   | -        | -         |      |
| M7.2  | todo   | -        | -         |      |
| M8.1  | todo   | -        | -         |      |
| M9.1  | todo   | -        | -         |      |
| M10.1 | todo   | -        | -         |      |
