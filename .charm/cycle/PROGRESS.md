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

Current task: M6.2

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

M4.1 done: new `cmd/noggin/gitstatus.go` shells out to the real `git` binary (detected via
`exec.LookPath`, same degrade-quietly shape as S14's `detectRipgrep`) — `rev-parse
--show-toplevel` (doubles as the "inside a repo at all" check) then `status --porcelain=v1
--untracked-files=all --no-renames -z -- .`, scoped to the Root via `-C root` + pathspec `.`.
Porcelain paths are always toplevel-relative regardless of invocation dir (verified
empirically against git 2.53 — see D46), so `parseGitStatus` joins them onto the resolved
toplevel to get absolute paths matching `tree.Node.Path` directly. `--no-renames` keeps every
changed path a flat single entry, matching S8's "no renamed marking". Any failure (no git, not
a repo, anything else) degrades to a nil map, never an error — `model.openFolder` calls it
fresh on every load (launch and every Root switch, not cached like M3.1's memory) and installs
it via `filesNav.withStatus`; `renderRow` prefixes a file row (never a directory) with `M `/
`A `/`  ` in the same 2-char-column style the expand/collapse marker already uses. Full
design rationale in DECISIONS.md D46.
Tests: `cmd/noggin/gitstatus_test.go` (new) — `parseGitStatus` unit-level against synthetic
porcelain output (M/A priority table, deleted/staged-deletion left unmarked); `gitStatus`
end-to-end against real temp git repos (modified/untracked/staged-add-then-edited, pathspec
scoping inside a larger repo, outside-any-repo silence, missing-git-binary silence);
`filesNav` render-level (marker on the right row, never on a directory); and
`TestActivateRootShowsGitStatusMarkersAtLaunchAndOnSwitch` driving the real `model.activateRoot`
path. New `test/pty/checks.py` section (labelled `C2a/M4.1` — the bare `M4.1` label already
belongs to an earlier cycle's edit-round-trip checks in this same file) spins up a real temp
git repo and a real non-repo dir and confirms M/A markers reach the actual terminal, and that a
non-repo Root's footer never shows an error. `make check` and `make check-pty` both green,
including the 6 new pty checks. No findings; no new triage items — T1/T2 unchanged.

M5.1 done: `keys.Refresh` ("R", keymap.go) is global like Edit/Find/Navigator, reachable from
every pane and inert while the Picker/Help Modal/an in-progress find own the keyboard (same
early-return gates FocusRoots/FocusFiles already relied on). `model.refresh()` (model.go) is the
one place both halves happen together: it re-`tree.Walk`s `m.files.root.Path` (a nil root, i.e.
no launch argument and no roots at all, is left untouched — nothing to refresh) and, on
success, calls the new `filesNav.refreshRoot` plus `gitStatus(node.Path)` in the same branch, so
S8's markers are never stale behind the tree they describe. `refreshRoot` (files.go) preserves
the expansion set as-is (a dead path is just absent from `visibleRows`, harmless) and moves
selection via the new `nearestSurvivingIndex`: walks the selected node's path upward through
`filepath.Dir` until a still-visible row matches or the walk reaches `f.root.Path` (falls back
to index 0) — correct in one pass because every ancestor a visible row sits under is, by
construction, already expanded, so the walk only ever asks "does this path still exist", never
"is it visible". `indexOfPath` (M3.1's own restore-from-memory lookup) now shares the same
`rowIndexOf` primitive rather than duplicating the loop. The Document Viewer's own half:
`viewer.refreshDoc` (viewer.go) re-renders in place without `open()`'s `GotoTop()` — bubbles
v2's `viewport.SetContentLines` already clamps `YOffset` down to the new end if the document
shrank past it, so no extra clamping code was needed, only *not* resetting to the top; a
read failure instead calls the new `viewer.close()`, which empties the pane. A committed find's
matches being dropped by the re-render (the same T43 mechanism `resize` already has) now also
surfaces a "refreshed; search cleared" footer message rather than losing them silently, for the
same reason `resize` bothers to.
Tests: `cmd/noggin/refresh_test.go` (new) — `filesNav.refreshRoot`/`nearestSurvivingIndex` at
the unit level (unchanged tree, ancestor substitution, whole-branch-gone fallback);
`viewer.refreshDoc`/`close` (scroll preserved, clamped on shrink, empties on close); and
model-level end-to-end covering the cycle's stated acceptance verbatim (create/edit/delete
under the Active Root then `R`, nearest-surviving-ancestor selection, preserved scroll, git
markers recomputed, Active-Document-deleted empties the viewer, no-root-at-all is a no-op, and
the committed-find-cleared message). New `test/pty/checks.py` section (`C2a/M5.1`) drives the
real binary: a file added/removed on disk without relaunching, a scroll position preserved
across a refresh that only grows the document (with `Bottom` confirming the grown content is
reachable), and the Active Document vanishing emptying the viewer. `make check` and
`make check-pty` both green (183 pty checks, 0 failures). A fresh-context code-reviewer pass
found no blocker/high findings; five low-severity items logged as TRIAGE T3–T5 (T1/T2
unchanged) — a message-wording nit was fixed inline, the rest (unbounded `expanded` map growth
over many refreshes, no explicit rootsPane-focused test, no explicit multi-Root
"nothing-else-touched" test) are left for replan to triage, not blocking.

M6.1 done: the entire fix is one call site — `rgFiles` (files_search.go), the File Picker's
`rg --files` invocation, gained `--hidden --glob '!.git'`. `--hidden` widens rg past its own
default dotfile/dot-directory exclusion (verified empirically against the sandbox's real rg:
without it, dotfiles were never even candidates before the Readable allowlist got a look);
the `.git` glob is necessary once `--hidden` is on, since rg has no built-in "`.git` is
special" rule the way `tree.Walk`'s own by-name check does — verified the bare pattern (no
leading `/`) excludes a `.git` at any depth, matching `tree.Walk`. `grepFiles`/`grepBatch`
needed *no* change: they run rg over the explicit file list `listFiles` already produced, and
rg does not re-apply hidden-file filtering to files named explicitly on its command line
(verified against the real binary) — so once `listFiles` includes a dotfile, Grep mode finds
it automatically. `.gitignore`/`.ignore` handling is untouched by either new flag (orthogonal
to both — confirmed by test, both a plain and a dotfile `.gitignore` entry still exclude).
Tests: `cmd/noggin/files_search_test.go` — the old combined `TestListFilesOmitsGitignoredAndHidden`
split into `TestListFilesOmitsGitignored` (unchanged half, now also covers a `.gitignore`'d
dotfile) and the new `TestListFilesIncludesDotfiles` (S4's positive assertion: a dotfile and a
Readable file inside a dot-directory now found, a non-Readable dotfile still excluded by the
allowlist — M6.1 widens candidates, not the Readable policy); `TestListFilesStillExcludesGitDirectory`
(top-level and nested `.git`); `TestGrepFilesFindsMatchesInsideDotfiles` (Grep mode's own half).
`cmd/noggin/picker_test.go`'s new `TestPickerFindsADotfileInBothModesWithNoExtraKeystroke` is
the cycle's stated acceptance criterion verbatim, driven through the real Picker model against
the real `rg` binary in both modes. `README.md`'s "Requires ripgrep" section updated to drop
the now-false "Picker disagrees with the Files Nav at the edges" caveat. `make check` and
`make check-pty` both green (no new pty check — this is which files rg returns, already
exercised against the real binary by `go test`, not a rendering-ambiguity case like M4.1's git
markers). A fresh-context code-reviewer pass found no blocker/high/medium findings; two low
severity test-coverage gaps it flagged (nested-`.git` depth, dotfile-under-`.gitignore`) were
backfilled inline before this handoff, since they were cheap and directly strengthened claims
already made in code comments — nothing left to log.

Next step: M6.2 — a session-scoped `alt+h` toggle hiding/showing dotfiles in both Picker modes
(needs M6.1, done). Must live on `model`, not `picker` (see foot-gun below) — `resetPicker`
(picker.go) wipes anything stored on `picker` every close/reopen, but the toggle must survive
that within a session and only reset on relaunch. The toggle presumably threads through to
`listFiles`/`grepFiles` (files_search.go) as an extra parameter (skip dotfile candidates when
hidden) rather than a second rg flag, since M6.1's `rgFiles` has no notion of the toggle's
state and re-deriving it from `rg` invocation flags per search would duplicate the Readable-
allowlist-style filtering `listFiles` already does in Go, not in rg.

Sequencing rationale (not yet worth a DECISIONS.md entry, but useful to a fresh session):
M3 (tree-state memory, S3) and M4 (git markers, S8) were ordered ahead of M5 (refresh, S2)
because S2 explicitly reuses S3's per-Root preserve mechanism and recomputes S8's markers in
the same pass — confirmed by M5.1 landing exactly that way with no rework needed. M9
(footer/help, S9) is last of the binding-affecting work since it is a sweep over bindings
M2/M5/M6/M7 define. M10 (staleness, S10) is last per the spec's own "pre-agreed cut, dropped
without ceremony".

Known foot-guns flagged by research, worth re-reading before M6.2/M7:
- A dotfile-toggle (M6.2) must live on `model`, not `picker` — `resetPicker` wipes anything
  stored on `picker` every close/reopen.
- Peek (M7) touches many call sites (shiftFocus, openPicker, refresh, selection change) — each
  is a place its "clear on X" rule can be forgotten. M5.1's `refresh()` is now one more call
  site Peek (M7) will need its own "clear on refresh" rule at, per S7's own bullet list.
- M9 (footer/help sweep) still owes `r`/`f` (M2.1) and now `R` (M5.1) their footer hints and
  Help Modal entries — none of the three add one yet, deliberately deferred to M9 (see M2.1's
  own handoff note above). M6.2's `alt+h` will be a fourth.

## Status

| Key   | Status | Reviewed | Validated | Note |
|-------|--------|----------|-----------|------|
| M1.1  | done   | 2026-08-06 | 2026-08-06 |      |
| M2.1  | done   | 2026-08-06 | 2026-08-06 |      |
| M2.2  | done   | 2026-08-06 | 2026-08-06 |      |
| M3.1  | done   | 2026-08-06 | 2026-08-06 |      |
| M4.1  | done   | -        | -         |      |
| M5.1  | done   | 2026-08-06 | -         |      |
| M6.1  | done   | 2026-08-06 | -         |      |
| M6.2  | todo   | -        | -         |      |
| M7.1  | todo   | -        | -         |      |
| M7.2  | todo   | -        | -         |      |
| M8.1  | todo   | -        | -         |      |
| M9.1  | todo   | -        | -         |      |
| M10.1 | todo   | -        | -         |      |
