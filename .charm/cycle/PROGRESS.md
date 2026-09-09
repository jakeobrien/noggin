# Progress — C6 sources

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

Current task: M8.2

## Handoff

M6.1 (Recent for articles) is done: `recordArticleViewed` (model.go) now stamps Article/Title
on Activate (not Peek), `recentRows` (browse_list.go) takes a `wikipedia bool` param mirroring
`starredRows`, buckets article records under `wikipediaSourceKey` instead of `rootForPath`,
skips the disk-existence check for them (a URL always looks "missing" to `pathMissing`), and
builds rows via the renamed `buildDocBrowseRow` (was `buildStarredBrowseRow`, now shared by both
Recent and Starred). The D133 gotcha (toggleDocStar's flat-mode branch) turned out to need no
code change: `recordArticleViewed` always writes a fully-stamped record, so by the time an
article can reach a flat row at all it is already correct — only the stale doc comment there
was updated to explain why. Reviewed by a fresh-context code-reviewer subagent pass (2026-09-09, Status table); two
low-severity stale-comment findings fixed, no functional issues. Not yet validated
(charm-validate) — only reviewed so far. `make check` green; `go test ./...` covers recording,
non-recording-on-Peek, ordering, the per-Source cap and the Wikipedia label, all new.

M7.1 (Root names' store layer) is done: store.go's Roots store now persists an optional per-
Root `Name` (S8) alongside its `Path` — `storedRoot{Path, Name}` replaces the old bare
`[]string`. Format stays hand-editable and backward-compatible: a line is `<path>` (unnamed —
S8's default, the folder's base name, applies) or `<path>\t<name>` (name escaped via a new
`escapeStoreField`/`unescapeStoreField`, a store.go-local copy of docstore.go's tab/newline/
backslash scheme — kept separate per D61, the two stores' formats deliberately don't share
parsing code). `saveStore` never writes the trailing tab for an unnamed Root, so an
all-unnamed store round-trips byte-for-byte identical to the pre-S8 format; a legacy plain-
path-per-line store loads straight into unnamed entries. `toggleStore` adds a freshly-starred
Root unnamed and drops a Root's Name entirely on unstar (re-starring starts fresh, not with
the old name).

Deliberately NOT done in M7.1 (left for M7.2, which needs it): threading Name through the
launch-resolution pipeline (`resolveRoots`/`resolveLaunch`/`starredRoots` in roots.go still
take/produce plain `[]string`/`root{path, starred}` with no `name` field) or into the runtime
`root` struct at all. `loadStarredRoots()` kept its existing `[]string` contract (a new
`storedRootPaths` helper extracts paths from `loadStore`'s `[]storedRoot`) specifically so
roots.go's ~15 existing call sites (resolveRoots/resolveLaunch/baseResolution, exercised by a
large table-driven roots_test.go) stay untouched — none of that logic needs a Root's name to
decide which root is active or how a launch argument resolves. M7.2's rename UI is the natural
point to add `root.name`, wire it from the store at startup, and read/write it live; M7.1 only
had to prove the store itself can carry and round-trip a name. Round-trip coverage: named +
unnamed Roots in the same store, a Name containing the escape scheme's own delimiter (tab),
a newline and a backslash, a legacy-format (bare-path) store, and toggle-on/off interaction
with a Name. `make check` green. Not yet reviewed or validated.

M7.2 (inline rename in Sources Nav) is done, completing a session that a previous run had
left in-progress and uncommitted: `keys.Rename` ("c", the obvious mnemonics `r`/`n`/`N` being
taken by FocusRoots/FindNext/FindPrevious) opens an inline edit on the Sources Nav's selected
row (rootsnav.go's `renderRow`/`startRename`), Enter commits (`model.commitRename` →
`renameStore`), Esc cancels, and an empty/whitespace-only name reverts to the default (the
folder base name via `root.displayName`). The Wikipedia row and an unstarred ad-hoc Root are
both refused with a Footer message (mirroring Star's own refusals); starring an ad-hoc Root
first (already unnamed per M7.1's `toggleStore`) then lets it be renamed, seeded with its
default name. `keys.Rename` is in rootsPane's Footer hints and Help Modal only. The session
also finished M7.1's deliberately-deferred wiring: `loadStarredRoots` now returns the full
`[]storedRoot` (Name included) instead of bare paths, and `resolveRoots`/`resolveLaunch`/
`starredRoots` (roots.go) thread each Root's persisted Name into the runtime `root.name` —
so a Root renamed in an earlier session shows its name from the moment a fresh one launches,
not only after an in-session rename (`TestRenamePersistsAcrossReload`). `rename_test.go` (new)
covers every path in the acceptance criterion: seed-from-current-name, commit, cancel, empty/
whitespace revert, both refusals, post-star allowance, keyboard exclusivity while editing,
cross-session persistence, Footer-hints/Help-Modal presence, and the footer/row rendering
switch. `make check` green (build, vet, fmt, `go test ./...`); no findings, no TRIAGE items,
no DECISIONS entries needed — the implementation followed M7.1's handoff plan directly with
no judgment calls beyond what's already in the code's own doc comments. Not yet reviewed or
validated.

M7.3 (Root-label helper) is done: `root.displayName()` (roots.go, already introduced by
M7.1/M7.2) is now the one helper feeding all three places S6/S3/S8 show a Root's label —
Sources Nav line 1 (already wired by M7.2), `buildBrowseRow`'s `rootLabel` field
(browse_list.go:67, cross-Source Files Nav rows in Recent/Starred, consumed verbatim by
`filesNav.renderFlatRow`), and `pickerRootLabel` (picker_view.go:274, File Picker Results).
Both remaining call sites previously called `filepath.Base(<root path>)` directly instead of
`displayName()`, silently dropping a Root's Name outside the Sources Nav; changing each to
`r.displayName()` / `f.root.displayName()` was the entire code change — no new field, no
new type, `displayName`'s own fallback (empty Name → folder base name) made both call sites
behave identically to before for every existing (unnamed) fixture, confirmed by the full
existing suite passing unmodified. New `rootlabel_test.go` pins the helper itself and all
three call sites against a renamed Root; each of the three was confirmed to actually fail
against the pre-fix code (temporary revert, mutation-tested, then restored) before being
trusted. Grepped the whole package after: no `filepath.Base(<root path>)` label computation
remains outside `displayName()` (the only other `filepath.Base` calls left are on a
document's own path/file name — document.go's `label()`, files_search.go's fuzzy-match
source — which are a different thing, a file's own name, not a Root's). `make check` green;
`go test ./... -count=1` (uncached) green. Not yet reviewed or validated.

M8.1 (half-page paging) is done: `keys.PageUp`/`keys.PageDown` (pgup/pgdn, plus their
ctrl+b/ctrl+f alt keys) now move a Document Viewer by the same distance as `HalfPageUp`/
`HalfPageDown` (ctrl+u/ctrl+d) — viewer.go's `scrollActions` table changes only the two
move funcs (`vp.HalfPageUp()`/`vp.HalfPageDown()` instead of `vp.PageUp()`/`vp.PageDown()`);
the bindings themselves (keymap.go) are untouched, so nothing else sharing them moves. The
Picker's own Preview reuses the identical `keys.PageUp`/`keys.PageDown` bindings in a wholly
separate call site (picker.go's `handlePickerInput`) that still calls `.PageUp()`/`.PageDown()`
directly — confirmed untouched by both a code read and `TestPickerPreviewScrollsOnScrollKeys`
still passing; Files Nav never bound PageUp/PageDown at all (`files.go` only handles
LineUp/LineDown), so it needed no test and has nothing to regress. `TestScrollActionsMoveThe
Viewport`'s page-up/page-down cases were updated to the new expected offsets, plus a new
subtest asserting pgdn and ctrl+d land on the exact same offset from the same start (the
acceptance criterion's own wording). New `TestHelpModalListsPageAndHalfPageTogetherForThe
Viewer` (help_test.go) pins the Help Modal showing all four scroll-distance bindings as one
contiguous block for both dv1Pane/dv2Pane. Left `keys.PageUp`/`PageDown`'s shared `WithHelp`
text ("page up"/"page down") unchanged — a per-pane fork of that description was judged not
worth it for an effort-1 task; both viewer.go's and keymap.go's comments were updated instead
to explain the two call sites now move by different distances on purpose. `make check` and
`go test ./... -count=1` green. Not yet reviewed or validated.

M6.1, M7.1, M7.2 and M7.3 were validated fresh-context (charm-validate, 2026-09-08, via a
read-only validator subagent): `make check` and `go test ./... -count=1` both green; each
task's acceptance criterion independently re-derived from code (not taken from this Handoff's
own prose) — M6.1's Activate/Peek recording split traced through `applyArticleFetch`'s two
branches to confirm no other path reaches `recordArticleViewed`; M7.1's store round-trip
confirmed for named/unnamed/legacy/escaped-delimiter cases; M7.2's two refusal conditionals
(unstarred ad-hoc, Wikipedia) traced directly in `model.go`, not just trusted from tests;
M7.3's "single helper" claim re-checked with an independent `filepath.Base` grep across the
whole repo (found the handoff's own exception list under-enumerates the survivors — `tree.go`,
`picker_view.go:357`, `browse_list.go:66` — but all are non-Root-label uses, so the claim holds
functionally; cosmetic only, not worth a TRIAGE item). All four: PASS, no gaps, no new
findings. TRIAGE.md still holds T24/T25 (stale doc-comment findings from M7.1/M7.2's code
review, unrelated to functional correctness) — open, awaiting the next replan pass, not
blocking.

Cursor is now M8.2 (location line), the next M8 task (`needs: M4.1, M7.3`, both already done).
M7's three tasks (S8, Root names) are now all done and validated; M8.1 is done (not yet
reviewed/validated); M1.11, M2.10 and M5.3 remain independent replan-added hygiene tasks,
pull-forward-anytime. Working tree is clean except this session's own tracking-doc edits
(M8.1's code+test changes are committed).

No blockers.

## Status

| Key   | Status | Reviewed | Validated | Note |
|-------|--------|----------|-----------|------|
| M1.1  | done   | 2026-09-08 | 2026-09-08 |      |
| M1.2  | done   | 2026-09-08 | 2026-09-08 |      |
| M1.3  | done   | 2026-09-08 | 2026-09-08 |      |
| M1.4  | done   | 2026-09-08 | 2026-09-08 |      |
| M1.5  | done   | 2026-09-08 | 2026-09-08 |      |
| M2.1  | done   | 2026-09-08 | 2026-09-08 |      |
| M2.2  | done   | 2026-09-08 | 2026-09-08 |      |
| M2.3  | done   | 2026-09-08 | 2026-09-08 |      |
| M1.6  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T1, T2) |
| M1.7  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T3) |
| M2.4  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T4, T5) |
| M2.5  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T6, T7, T8) |
| M2.6  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T10) |
| M2.7  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T11, T12) |
| M1.8  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T17) |
| M2.8  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T15) |
| M3.1  | done   | 2026-09-08 | 2026-09-08 |      |
| M3.2  | done   | 2026-09-08 | 2026-09-08 |      |
| M3.3  | done   | 2026-09-08 | 2026-09-08 |      |
| M3.4  | done   | 2026-09-08 | 2026-09-08 |      |
| M3.5  | done   | 2026-09-08 | 2026-09-08 |      |
| M4.1  | done   | 2026-09-08 | 2026-09-08 |      |
| M4.2  | done   | 2026-09-08 | 2026-09-08 |      |
| M4.3  | done   | 2026-09-08 | 2026-09-08 |      |
| M5.1  | done   | 2026-09-08 | 2026-09-08 |      |
| M5.2  | done   | 2026-09-08 | 2026-09-08 |      |
| M1.9  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T18) |
| M1.10 | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T2) |
| M2.9  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T1, T22) |
| M2.10 | todo   | -        | -         | added by replan 2026-09-08 (TRIAGE T2) |
| M3.6  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T20) |
| M3.7  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T21) |
| M4.4  | done   | 2026-09-08 | 2026-09-08 | added by replan 2026-09-08 (TRIAGE T23) |
| M1.11 | todo   | -        | -         | added by replan 2026-09-08 (TRIAGE T1, T3) |
| M5.3  | todo   | -        | -         | added by replan 2026-09-08 (TRIAGE T4) |
| M6.1  | done   | 2026-09-09 | 2026-09-08 | pre-agreed cut 1 |
| M7.1  | done   | 2026-09-08 | 2026-09-08 | pre-agreed cut 2 |
| M7.2  | done   | 2026-09-08 | 2026-09-08 | pre-agreed cut 2 |
| M7.3  | done   | 2026-09-08 | 2026-09-08 | pre-agreed cut 2 |
| M8.1  | done   | -        | -         |      |
| M8.2  | todo   | -        | -         | scope widened by replan 2026-09-08: S9a attribution line |
| M8.3  | todo   | -        | -         |      |
| M9.1  | todo   | -        | -         | pre-agreed cut 3 |
| M10.1 | todo   | -        | -         |      |
