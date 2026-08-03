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

Current task: M2.5

## Handoff

<!-- Overwrite each session. 5–15 lines. Enough for a fresh agent to continue:
     state of the current task, in-flight changes, surprises/gotchas, next step, blockers. -->

M2.4 done: async, debounced search dispatch — the codebase's first genuinely concurrent
`tea.Cmd`. Design is D41: one monotonic `picker.seq` generation counter; every change to what
would be searched bumps it and schedules `tea.Tick(120ms)` carrying that seq
(`pickerSearchMsg`); a matured timer dispatches the real search only if its seq is still
current (that is the debounce), and a returning `pickerResultsMsg` is applied only if its seq
is still current (that is the stale-discard). `openPicker`/`dismissPicker` bump seq rather than
zeroing it, so a search in flight across a dismissal cannot be accepted by the next Picker —
the one field S8's "replace the whole struct" hygiene deliberately carries over, documented as
carrying no user-visible information. `pickerSearchCmd` is a pure function of copied-in values
(roots, active, mode, scope, query) because it runs concurrently with `Update`.

New state: `picker.seq`, `picker.searching`, `picker.results pickerResults{query, mode, scope,
files, grep, truncated}`. Results are labelled with what they answer *for* — M2.5's status /
empty states read those labels rather than assuming results match the current Prompt, and
`searching` is what distinguishes "found nothing" from "still looking". Nothing renders any of
it yet (M2.5/M3.1 do). `rootsNav.rootList()` is new: the search layer's `(roots, active)` shape
read from the pane that already owns the session's roots, so no second copy can drift.

Gotchas: an empty Prompt is refused in both `queuePickerSearch` and `dispatchPickerSearch` (rg
with an empty pattern matches every line of every file); a keystroke the textinput consumes
without changing the query (cursor moves) must not bump seq, or it would discard a live search
for a still-current query — both have tests. The 3-character threshold is *not* here; it is
M2.5's criterion.

Tests (`picker_test.go`, 5 new): no sleeps, no goroutines, no test seam — see
`findings/async-cmd-testing.md`. Ticks are synthesized from each keystroke's recorded seq and
played back; "a search was issued" is observable as `Update` returning a non-nil Cmd for a
`pickerSearchMsg`; the search Cmd is called inline against a real rg over a `t.TempDir()` Root.
Evidence for the criterion: typing "repor" (5 keystrokes) dispatches exactly 1 search and the
displayed results answer "repor". `make check` and `make check-pty` both green.

Next step: M2.5 — status and empty states (mode/Scope indicator, sub-threshold message, no
results, truncated-at-cap, ripgrep-unavailable), including "no search issued for a 1- or
2-character query". The threshold belongs in `queuePickerSearch`/`dispatchPickerSearch`
alongside the existing empty-Prompt refusal; `m.ripgrep` (model.go) already carries S14's
availability and reason.

## Status

| Key  | Status | Reviewed | Validated | Note |
|------|--------|----------|-----------|------|
| M1.1 | done   | 2026-08-03 | 2026-08-03 |  |
| M1.2 | done   | 2026-08-03 | 2026-08-03 |  |
| M1.3 | done   | 2026-08-03 | 2026-08-03 |  |
| M1.4 | done   | 2026-08-03 | 2026-08-03 |  |
| M2.1 | done   | -        | -         |      |
| M2.2 | done   | -        | -         |      |
| M2.3 | done   | -        | -         |      |
| M2.4 | done   | -        | -         |      |
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
