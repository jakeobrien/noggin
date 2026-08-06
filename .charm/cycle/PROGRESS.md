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

Current task: M4.1

## Handoff

M1.1 done: `make install` (new target) copies the already-built `bin/noggin` to
`$(PREFIX)/bin/noggin`, `PREFIX ?= /usr/local`, and deliberately has no `build` prerequisite
(Makefile:14-15,29-33). Verified: `make build && make install PREFIX=/tmp/x` places a working
binary there; `make -n install` shows no `go build` step; default `PREFIX=/usr/local` also
works (needs root in this sandbox since `/usr/local` is root-owned here — a sandbox fact, not
a product concern). No findings or triage items from this task.

M2.1 done: added `keys.FocusRoots` ("r") and `keys.FocusFiles` ("f") (keymap.go) and
`model.focusPane(target pane)` (model.go, next to toggleNavigator), wired as two more global
cases in Update's main switch alongside Navigator/Edit/Find — which is what makes them
automatically inert while the Picker is open (the `m.picker.open` gate returns before that
switch is reached, same shield those three already had). `focusPane` un-hides the Navigator
via the same `resize`/`resizeCauseNavigatorToggle` path `toggleNavigator` uses, but
one-directionally (only ever clearing `navigatorHidden`, never setting it), then sets focus
through `resolveFocus` rather than a bare assignment so a too-narrow terminal (no Navigator
column regardless of the toggle, per computeLayout) can't land focus on an invisible pane.
Tests: `cmd/noggin/focus_test.go` (new) covers all three panes as the press-origin, the
reveal-hidden-Navigator case, the never-hide case, and the Picker-open inertness case;
`frame_test.go`'s `allBindings()`/`TestEveryBindingIsCovered` updated for the two new fields.
`make check` and `make check-pty` both green. Footer hints and the Help Modal do not yet
mention `r`/`f` — deliberately deferred to M9 (S9 sweep), not missed.

M2.2 done: `keys.Picker` is now `key.WithKeys("space")` (keymap.go) — `ctrl+p` is gone, not
kept as an alias, and no other binding claims it, so pressing it now falls through every
switch unmatched (D2's "keymap.go is the only file naming a key" makes that structural, not
just observed). Two test-only key-encoding tables needed the new key added to stay in sync,
per their own comments: `cmd/noggin/keymap_test.go`'s `namedKeys` (`"space": tea.KeySpace`,
keymap.go's exempted test-only carve-out) and `test/pty/checks.py`'s `KEY` table (the pty
harness's real-bytes-on-the-wire carve-out, same rationale, different language boundary) —
neither is scope creep, both are the existing carve-out mechanism doing its job for a changed
binding. `TestPickerOpensFromEveryFocusAndNavigatorHidden` (picker_test.go) already drives the
binding indirectly via `keys.Picker.Keys()[0]`, so it picked up the change with no edit and
confirms the Picker opens from every Focus on the new key; the pty check formerly named
"ctrl+p opens the Picker modal" is renamed and now sends the space byte, still green.
`make check` and `make check-pty` both green.

Found but not fixed (logged to TRIAGE T1, not this task's scope): README.md still documents
`ctrl+p` opening the Picker in three places — no plan task owns README sync the way M9.1 owns
Footer/Help Modal, so it's flagged for replan rather than silently touched.

M3.1 done: `filesNav` gained `memory map[string]filesNavMemory` (files.go), keyed by a Root's
own `Path` (not the `*tree.Node` pointer — `activateRoot` re-walks a fresh Node every
activation even when nothing on disk changed, exactly the flagged foot-gun). `setRoot` now
calls `remember` on the outgoing root before switching (snapshots `expanded` by reference —
safe because `withExpanded` only ever builds new maps, never mutates in place — plus the
selected node's `Path`, not its index), then either restores a previously-visited Root's
memory or resets fresh for a first visit, same as before M3.1. Restoring recomputes `selected`
via a new `indexOfPath` lookup against the restored expansion's visible rows rather than
trusting a stale int, which is what makes it correct across however many other Roots were
visited in between. Tests: `TestFilesNavRestoresPerRootStateAcrossSwitches` (files_test.go,
filesNav-level, exercises setRoot directly) and `TestActivateRootRestoresFilesNavPerRoot`
(rootsnav_test.go, model-level through `activateRoot` — the actual keypress path); both are
the cycle's stated acceptance criterion verbatim (expand/select in A, switch to B, switch back
to A, same state). `TestActivateRootLeavesOpenDocumentUntouched`'s comment updated to note its
B-is-a-first-visit case is now one of two paths, not the only one. `make check` and
`make check-pty` both green (full pty suite, no new pty check added — this is pure model
state with no rendering ambiguity, already deterministic under `go test`). No findings; no new
triage items.

Next step: M4.1 — per-file git status markers (M/A) in the Files Nav. New territory: no git
plumbing exists yet in the tree package (`internal/tree`) or `cmd/noggin`; needs a way to
resolve each Root's containing repository (if any) and its working-tree status, without
erroring on a Root outside any repo. Likely shells out to `git status --porcelain` or reads
`.git` plumbing directly — worth a quick research pass on which is more robust before
committing to an approach, given the security-review's eye on shelling out (cf. how
`internal/tree/gitignore.go` deliberately hand-rolled `.gitignore` instead of shelling to
`git check-ignore`, per `.charm/knowledge/gitignore-matching.md`).

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

## Status

| Key   | Status | Reviewed | Validated | Note |
|-------|--------|----------|-----------|------|
| M1.1  | done   | -        | -         |      |
| M2.1  | done   | -        | -         |      |
| M2.2  | done   | -        | -         |      |
| M3.1  | done   | -        | -         |      |
| M4.1  | todo   | -        | -         |      |
| M5.1  | todo   | -        | -         |      |
| M6.1  | todo   | -        | -         |      |
| M6.2  | todo   | -        | -         |      |
| M7.1  | todo   | -        | -         |      |
| M7.2  | todo   | -        | -         |      |
| M8.1  | todo   | -        | -         |      |
| M9.1  | todo   | -        | -         |      |
| M10.1 | todo   | -        | -         |      |
