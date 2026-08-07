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

Current task: none — cycle-end validation pass done; awaiting `charm-replan` to triage T13
(and the rest of TRIAGE.md) before this cycle can close.

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
`exec.LookPath`, same degrade-quietly shape as C2/S14's `detectRipgrep`) — `rev-parse
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

M6.2 done: `keys.ToggleHidden` (`alt+h`, keymap.go) is a non-typing key inside the Picker
(handlePickerInput), handled by the new `model.toggleHiddenFiles()` (picker.go) exactly like
`switchPickerMode`/`switchPickerScope` already are — flips `model.hideDotfiles` and re-runs
the current query through `queuePickerSearch`. `hideDotfiles` lives on `model` (model.go), not
`picker`, per the foot-gun this handoff flagged: `resetPicker` replaces the whole `picker`
struct with `newPicker()`'s zero value on every open/close, so anything needing to survive
that within a session cannot live there — the same reason `navigatorHidden` (S9) already
lives on `model`. `listFiles`/`grepFiles` (files_search.go) both gained a `hideDotfiles bool`
parameter (grepFiles' own reuse of listFiles' candidate set means one filter point covers both
modes): true drops a candidate if any path component between the Root and the file starts
with `.` (new `isDotfilePath`, matching what M6.1's `--hidden` widened `rgFiles` to walk into,
not just the file's own base name). `pickerResults` gained a matching `hideDotfiles` label and
`resultsCurrent` compares it alongside query/mode/scope — the same "do these results still
answer what's currently asked" contract mode/Scope switches already rely on, now covering the
toggle too, threaded through `pickerPreviewCandidate`/`pickerPreviewTarget`/
`pickerGrepCandidate`/`syncPreview`/`pickerStatus`/`pickerResultsContent`/`renderPicker`/
`overlayPicker`. `pickerModeScope`'s title indicator (already showing mode/Scope) gained a
" · Hidden" suffix when active — not spec-mandated, but the toggle's only user-visible
confirmation beyond the Results list itself changing, and cheap given the title already
answers the same kind of question.
Tests: `cmd/noggin/files_search_test.go` — `TestListFilesHidesDotfilesWhenToggled`,
`TestGrepFilesHidesDotfilesWhenToggled`, `TestIsDotfilePath` (path-component rule directly,
including the `a.b/c.md` non-dot-prefix negative case). `cmd/noggin/picker_test.go`'s
`TestPickerAltHTogglesDotfilesInBothModesAndSurvivesReopen` is the cycle's stated acceptance
criterion verbatim: hide/restore in File mode, survives Cancel + reopen, then Grep mode too.
`frame_test.go`'s `allBindings()` updated for the new field. New `test/pty/checks.py` section
(`C2a/M6.2`) drives the real binary in both modes. It hit the exact ordering hazard
`picker-pty-async-ordering.md` (D58) already documented — toggling immediately after a landed
search intermittently misread the title — confirmed by reproducing it, then fixed the way that
finding prescribes: every `alt+h` in the pty section is sent on an empty Prompt, before that
Picker-open's first search, never right after one lands; 4+ consecutive full `make check-pty`
runs clean after the reorder. `make check` and `make check-pty` both green. No code-reviewer
pass run yet this task (see Status table).
Found but not fixed (logged to TRIAGE T6, not this task's scope): README.md's Picker section
still only describes dotfiles as included by default, with no mention of the new toggle —
M9.1 owns the Footer/Help Modal half of S4's sweep but nothing owns README, same shape as T1.

M7.1/M7.2 done together: Peek (`p`, `keys.Peek`) shows the Files Nav's selected non-Active
file in the Document Viewer via a new `viewer.peek` (`peekState`, viewer.go) — its own
read-only `viewport.Model`, sized/rendered alongside the Active Document's on every width
change (`render`/`renderPeek`) but never touching `v.vp`'s own scroll, which is what lets a
clear restore it untouched. `model.togglePeek` (model.go) starts it (no-op on a folder or the
Active Document's own row, per S7) and is also its own "second p" clear path.
`model.setFocus` (new — replaces every direct `m.focus = ...` assignment: openFolder,
openDocument, focusPane, resize, and Update's shiftFocus cases) is the one place a Focus
change clears Peek, since Peek can only ever exist while focus sits on `filesPane`;
`refresh()` and `openPicker()` clear it explicitly for their own triggers, and a Files Nav
selection move (LineUp/LineDown only — Expand/Collapse leave the selected node itself
unchanged) clears it in Update's `filesPane` case. `e` and `/` are silent no-ops while Peeking
(`startEdit`/`startFind`); scroll keys never reach the viewer at all during Peek since focus
never leaves `filesPane` — no separate scroll guard was needed. The inverted header is a new
`peekTitleStyle` (`titleStyle.Reverse(true)`); `renderBox` (picker_view.go) now takes an
already-styled title instead of styling it itself, since Peek needs a style other panes don't
— its three non-frame call sites (Prompt/Results/Preview) now wrap their titles in
`titleStyle.Render(...)` themselves.

M7.1 and M7.2 were built and verified together, not sequentially, despite the plan's split:
M7.1's own "cannot be scrolled" claim only holds if a Focus change clears Peek first
(otherwise `tab` then a scroll key would silently scroll the Active Document's real viewport
behind the still-shown Peek content) — the two milestones turned out to be one mechanism, not
two independently-choosable ones. Tests: `cmd/noggin/peek_test.go` (M7.1's own acceptance:
shows/header/focus-unchanged/active-doc-untouched, no-op on the Active Document and on a
folder, cannot scroll/find/edit) and `cmd/noggin/peek_clear_test.go` (M7.2's acceptance: each
of the five clearing triggers, scroll position preserved, view actually shows the Active
Document again, and the no-Active-Document case reverting to the empty Viewer).
`test/pty/checks.py`'s new `C2a/M7.1` section verifies the inverted header reaches the wire
as real reverse video (`\x1b[1;7m`, Bold+Reverse combined — not the bare `\x1b[7m` M2.2's own
check gets away with) and that a scroll-only key leaves the frame unchanged while Peeking.
`make check` and `make check-pty` both green (200 pty checks). No findings; no new triage
items.

M8.1 done: each Roots Nav entry now renders across two physical lines (`rootsnav.go`'s
`renderRow`, `rootRowHeight = 2`) — line 1 carries the active (`●`) / starred (`★`) markers,
the base name (now bold unconditionally, not just when active — see D47) and the
`(missing)` broken suffix; line 2 carries the entry's full path alone (not just its parent,
unlike the single-line layout this retires), unbold, left-truncated the same way. `nav.go`'s
single-line `windowedView` is now a one-line special case of a new, more general
`windowedViewRows(width, height, selected, count, rowHeight, renderRow)`, which windows and
scrolls in whole entries — never splitting one across the top/bottom edge — so
`rootsNav.view()` and `filesNav.view()` (unchanged, rowHeight=1) share the same scrolling
primitive per T16's own precedent. `move()`/`moveSelection` needed no change: selection was
already an entry index, not a line index, so "one keypress moves one entry" was true by
construction once rendering caught up.

D47 records a real, deliberate behaviour change surfaced by implementing S5 literally: the
pre-S5 layout used Bold to mean "this is the active root" (redundant with the `●` marker);
S5's "name bold unconditionally" removes that second cue, leaving `●` as the sole
active-root signal. Not a regression — a consequence of the spec text, logged since a fresh
session diffing `rootsnav.go` might otherwise wonder about it.

Tests: `cmd/noggin/rootsnav_test.go` — the four pre-existing render tests
(`TestRootsNavDistinguishesSameNamedRootsByParent`, `TestRootsNavMarksMissingPathBroken`,
`TestRootsNavHighlightsActiveRoot`, `TestRootsNavStarredMarked`) updated for the two-line
line-index shape; three new ones for M8.1's own acceptance
(`TestRootsNavRendersTwoLinesPerEntry`, `TestRootsNavSelectionHighlightsBothLines`,
`TestRootsNavScrollsInWholeEntries` — the last pins "never split an entry across the
screen edge" with a height that only fits one entry at a time). One pre-existing Picker test
(`TestPickerActivateGrepResultInsideStrippedHTMLCommentDegradesToPlainOpen`) had its
no-reverse-video assertion rescoped from the whole frame to `viewer.view()` alone — S5's
active+selected row's unbold path line now legitimately renders a bare `Reverse(true)` the
old single-line layout's Bold+Reverse combination never produced as a bare-comparable
substring, an unrelated-pane false positive the assertion was never actually testing for.

New `test/pty/checks.py` section (`C2a/M8.1`): confirms the name/path land on two separate
screen rows, that a path line never repeats the line-1 markers, that bold and reverse SGR
both reach the wire on the pane's first paint (`Screen.raw_lines()` drops all SGR — see
`pty-harness-ESC-M-D-gap.md` — so *which* line carries *which* attribute is a `go test`-only
assertion, not a pty one), and a functional Down-then-Activate check that selection actually
tracks whole entries. Building this section found and fixed a real harness gap (bare `ESC
M`/`ESC D` were unhandled, corrupting rows exactly like a product bug would) that was also
masking a pre-existing, always-vacuously-true assertion in the M6.2 Grep `alt+h` pty check
(comparing against the Prompt's own query echo, not an actual Result) — both are written up
in `pty-harness-ESC-M-D-gap.md` since they're reusable harness/testing knowledge, not
M8.1-specific. The M3.3 Roots Nav pty fixture's own three-root scenario needed a taller
Session (24 → 30 rows) to keep all three fixture roots visible under S5's halved per-height
row capacity — an instance of S5's own accepted "fewer visible roots" tradeoff, not a bug —
and its `row_with` marker-lookup helper gained a `row_pair_with` sibling that pairs a
path-line match with its own name line (never the next entry's, which risked a false-positive
marker read).

`make check` and `make check-pty` both green (206 pty checks, repeated runs clean — the
`alt+h`/star-toggle flakiness the harness fix removed was checked over multiple repeated
`make check-pty` runs, not just once). No triage items opened for this task.

M9.1 done: swept every binding S4/S6/S7 added, changed or retired this cycle into the
Footer hints (`paneHints`/`pickerHints`, footer.go) and Help Modal (`helpBindings`,
help.go), the two artifacts S9 names by name. FocusRoots/FocusFiles (S6, M2.1) and Refresh
(S2, M5.1) are global — reachable from Roots Nav, Files Nav and Document Viewer alike — so
all three now appear in every pane's footer hints and Help Modal list, the same
aware-regardless-of-focus treatment Edit/Navigator/Picker already got. ToggleHidden (S4,
M6.2) only does anything once the Picker is open, so it joins `pickerHints` (footer, where
it was missing entirely) and every pane's `helpBindings` (Help Modal is unreachable while
the Picker owns the keyboard, so these three panes are the only place it can ever be
documented — the same reasoning already covering PickerUp/PickerDown/PickerMode/
PickerScope). Peek (S7, M7.1/M7.2) is Files-Nav-only by design, so it was added only to
filesPane's footer hints and Help Modal list, deliberately absent from rootsPane/
viewerPane — verified against model.go's Update switch: Peek is matched only inside the
filesPane case, ToggleHidden only inside handlePickerInput, the other three inside the
global pre-focus switch, so each binding's new placement matches where it actually does
something. No remaining `ctrl+p` reference in either artifact (grep-verified); README's own
three ctrl+p mentions (T1) are out of this task's scope per that TRIAGE item's own reasoning,
now recorded explicitly as D48 rather than left implicit.

Growing the Help Modal's binding lists pushed filesPane's rendered box past 24 rows for the
first time (25, from 21 entries) — `TestHelpOverlayLeavesTheFrameVisibleAroundIt` (go test)
and its pty twin (`test/pty/checks.py`'s M5.3 section) both asserted the frame's top
border/footer were byte-unchanged outside the modal, which only holds if the modal doesn't
get clamped by `overlayHelp`'s MaxHeight; both were widened to a 30-row fixture so the
assertion keeps testing what it claims to (unclamped compositing), not the already-separately-
covered clamped case (`TestHelpOverlayNeverExceedsTheTerminal`, 40x10).

Tests: `cmd/noggin/footer_test.go` — `TestPaneHintsNameFocusAndRefreshFromEveryPane`,
`TestPaneHintsNamePeekOnlyInFilesPane` (positive for filesPane, negative for the other two),
`TestPickerHintTextNamesToggleHidden`, `TestNoFooterHintNamesCtrlP`.
`cmd/noggin/help_test.go` — `TestHelpBindingsNameFocusRefreshAndToggleHiddenFromEveryPane`,
`TestHelpBindingsListPeekOnlyForFilesPane`, `TestNoHelpModalNamesCtrlP`. These are the
cycle's stated acceptance criterion verbatim: `?` in each pane lists exactly that pane's
current bindings, and no ctrl+p anywhere in either artifact. `make check` and `make
check-pty` both green (full pty suite, 206 checks).

A fresh-context code-reviewer pass found no blocker/high/medium findings. Two low-severity
comment/capacity nits (a spec-item mislabel folding Refresh's S2 under S6; an off-by-one
allocation-capacity hint in help.go) were cheap and fixed inline before this handoff. A
third low finding — the M5.3 viewer-context pty block (untouched by this diff) already runs
against a Help Modal clamped past 24 rows, passing today only by luck of binding order — is
logged as TRIAGE T17, not this task's own regression but newly exposed by this task's own
growth of the list; the reviewer's assessment that "no ctrl+p anywhere" is Footer/Help-Modal-
scoped (not README) matches the project's own prior TRIAGE T1/T6/T15 reasoning, now recorded
as DECISIONS.md D48 rather than left implicit.

M10.1 done: the periodic disk check (S10) lives in a new `cmd/noggin/staleness.go`. A
`staleBaseline` (dirModTimes: one `os.Stat` per directory the already-loaded tree knows
about — root plus every `tree.Node` under it, traversed in memory, never `os.ReadDir` — plus
the Active Document's own path/mtime/size) is the "known good" snapshot; `resetStaleBaseline`
(re)establishes it and bumps a generation counter from every place that already (re)reads the
Active Root's tree or the Active Document from disk — `openFolder`, `openDocument`, `refresh`
— so the indicator is never stuck true past the user's own next load/refresh. A `tea.Tick`
every `staleCheckInterval` (6s, inside the spec's 5–10s band) fires `staleCheckMsg`;
`runStaleCheck` dispatches the actual Stat work as its own Cmd against whichever baseline/
generation are current *at that moment* (D49 records why: gating on the schedule-time
generation, the way picker.go's own debounce discards a superseded timer, would let a
reset landing mid-flight push real detection out to roughly two intervals instead of the one
M10.1's acceptance names) — `applyStaleResult` still discards a result whose generation has
since been superseded, so a slow check can't stomp a state that moved on while it ran. The
heuristic is deliberately not exhaustive (documented in staleness.go and D49): an in-place
edit to a *different*, non-Active file that touches neither its own directory's mtime nor the
Active Document goes unnoticed until the next refresh — catching that would mean statting
every file, exactly the cost S10 forbids. `git status` (already shelled out to for S8, D46)
was considered and rejected as the general tree-changed signal — it would leave non-git Roots
with no tree coverage at all, a bigger gap than the heuristic's own.

The indicator surfaces in the Footer (D49, since SPEC.md left the surface "to the plan"): a
fixed-wording suffix (`staleMessage`, "something changed on disk; press R to refresh" —
never what changed) appended to whatever hint line is already showing (pane hints or the
open Picker's, `appendStaleSuffix`) rather than replacing it, since S10 is a standing
condition until refreshed, not a one-shot D22 message — but a transient `m.message` and an
in-progress find still win over it, and the suffix is simply dropped (not truncated into) if
it would overflow the current width.

Tests: `cmd/noggin/staleness_test.go` (new) — `staleBaseline`/`snapshotStaleness`/`stale()` at
the unit level (nested file added inside a known subdirectory, a known subdirectory removed,
the Active Document edited in place or deleted, changes entirely outside the snapshotted tree
ignored, an unstattable document left untracked rather than half-recorded); the model-level
wiring (`openFolder`/`openDocument` establishing and advancing the baseline, `runStaleCheck`
dispatching against current state regardless of the timer's own age, `applyStaleResult`
dropping a superseded result, `Init` scheduling the first tick, `Update`'s own message-case
round trip, `refresh` clearing the indicator and rebaselining); and the Footer surfacing rules
(appended to pane hints, appended to Picker hints, losing to a transient message, omitted
rather than overflowing a narrow width). All async Cmds are driven synchronously per
async-cmd-testing.md's own recipe — no real `tea.Tick` is ever invoked in a test, only the
Stat-work Cmd `runStaleCheck` dispatches, which is safe to call directly. `make check` green.
No new `test/pty/checks.py` section, deliberately, matching M6.1's own precedent: the surfaced
text is plain, unstyled `footerStyle` output with no rendering ambiguity already untested by
`go test`, and the real behaviour needs a genuine ~6s wall-clock wait to observe, which would
add real time to every future `make check-pty` run for a claim `go test` already covers
deterministically. The claim was verified once, out-of-band, against the real binary via a
scratch harness script (not committed): at 300 columns (wide enough that the already-long
Files-pane hint line plus the suffix still fits), creating a nested file under a fresh Root
surfaced the indicator at 5.7s (inside the 6s–ish interval, accounting for scheduling jitter),
and pressing `R` cleared it on the same real process. `make check-pty`'s existing 206 checks
still pass unchanged.

A fresh-context code-reviewer pass (Pass 1 standards + Pass 2 spec-fidelity) found no
blocker/high findings and confirmed the generation-discard logic, the "no tree.Walk anywhere
reachable from the check path", Peek/Picker/Help-Modal non-interference, and the
openDocument-ordering claim (baseline reads the *new* document, not the old one) all hold by
direct code reading. Two of its low-severity findings were fixed inline before this handoff
since they were cheap and materially tightened claims already made in the code/D49: (1) a
document whose Stat fails right at snapshot time is now left untracked rather than half-
recorded with a zero-value ModTime that could never compare equal and would pin `stale()`
true forever; (2) `runStaleCheck` no longer gates on the schedule-time generation at all — see
D49's rewritten detection rationale above, this is the two-intervals-worst-case fix, not a
cosmetic change. The remaining low-severity items (no benchmark substantiating the "must not
perceptibly cost anything on a large Root" clause; a couple of untested-but-benign edge
combinations) are accepted as-is, consistent with the rest of this codebase having no
benchmark tests either, and are not logged to TRIAGE since none are actionable defects.

This was the cycle's last planned milestone (PLAN.md has no M11+). All ten milestones are
marked `done`.

Cycle-end `charm-validate` pass (2026-08-08, fresh-context validator subagent, not the
builder) covered M7.1, M7.2, M8.1, M9.1, M10.1 (the five never-validated tasks) plus a
post-refactor re-check of M2.1/M5.1 and the cycle-level SPEC.md acceptance list.
`make check` and `make check-pty` both green. **M8.1, M9.1, M10.1 verified and stamped
Validated below** — acceptance criteria independently confirmed against code/tests, no
regressions. M2.1/M5.1's existing 2026-08-06 Validated stamps hold: M7.1's `setFocus`
refactor (touching `focusPane`/`refresh`'s focus assignment) is a superset of the prior
behaviour (adds only an orthogonal `clearPeek()` side effect) — traced and confirmed, no
stale-date correction needed.

**M7.1 and M7.2 did NOT pass validation cleanly — left at `Validated: -`.** The validator
independently confirmed TRIAGE T13 (already logged, high severity, not new): `FindNext`/
`FindPrevious` (`n`/`N`, model.go:620-627) are gated only by `hasFindMatches()`, not by
`m.viewer.peeking()`, unlike `startEdit`/`startFind` a few lines above which both got explicit
`peeking()` guards in the same diff. If a Find was already committed on the Active Document
before Peek starts, `n`/`N` pressed while Peeking silently drifts `v.vp`'s real YOffset via
`applyCurrentMatch`→`EnsureVisible` (invisible on screen, since the view shows `peek.vp`), and
`clearPeek()` (viewer.go:139-142) never touches `v.vp`, so the drifted offset survives the
clear. This violates S7's unqualified "cannot be searched with Find" and the cycle-level
acceptance bullet "moving the selection restores the previous document at its previous scroll
position" in that specific, reachable scenario (M7.1's own PLAN.md wording only tests the `/`
key narrowly and technically passes; the broader SPEC.md guarantee does not). No fix applied
during this validation pass (validate does not fix). No new TRIAGE item needed — T13 already
cites this exact defect with matching file:line evidence.

TRIAGE.md is **not empty** (T1-T18, all `open`) — this blocks cycle-close regardless of the
above; the next step is `charm-replan` to triage the inbox (T13 first, given it changes a
task's pass/fail verdict), decide whether T13 becomes its own task or folds into a reopened
M7.1/M7.2, then re-run validate on whatever gets touched.

Sequencing rationale (not yet worth a DECISIONS.md entry, but useful to a fresh session):
M3 (tree-state memory, S3) and M4 (git markers, S8) were ordered ahead of M5 (refresh, S2)
because S2 explicitly reuses S3's per-Root preserve mechanism and recomputes S8's markers in
the same pass — confirmed by M5.1 landing exactly that way with no rework needed. M9
(footer/help, S9) is last of the binding-affecting work since it is a sweep over bindings
M2/M5/M6/M7 define. M10 (staleness, S10) is last per the spec's own "pre-agreed cut, dropped
without ceremony".

Known foot-guns flagged by research, worth re-reading before M9:
- M9 (footer/help sweep) still owes `r`/`f` (M2.1), `R` (M5.1), `alt+h` (M6.2) and now `p`
  (M7.1/M7.2) their footer hints and Help Modal entries — none add one yet, deliberately
  deferred to M9.

## Status

| Key   | Status | Reviewed | Validated | Note |
|-------|--------|----------|-----------|------|
| M1.1  | done   | 2026-08-06 | 2026-08-06 |      |
| M2.1  | done   | 2026-08-06 | 2026-08-06 |      |
| M2.2  | done   | 2026-08-06 | 2026-08-06 |      |
| M3.1  | done   | 2026-08-06 | 2026-08-06 |      |
| M4.1  | done   | 2026-08-06 | 2026-08-06 |      |
| M5.1  | done   | 2026-08-06 | 2026-08-06 |      |
| M6.1  | done   | 2026-08-06 | 2026-08-06 |      |
| M6.2  | done   | 2026-08-06 | 2026-08-06 |      |
| M7.1  | done   | 2026-08-07 | -         | see T13: broader S7 "cannot be searched with Find" gap (n/N ungated by peeking()) |
| M7.2  | done   | 2026-08-07 | -         | see T13: scroll-restore-on-clear violated in the same scenario |
| M8.1  | done   | 2026-08-07 | 2026-08-08 |      |
| M9.1  | done   | 2026-08-07 | 2026-08-08 |      |
| M10.1 | done   | 2026-08-08 | 2026-08-08 | perf clause ("no perceptible cost on a large Root") verified architecturally, not benchmarked |
