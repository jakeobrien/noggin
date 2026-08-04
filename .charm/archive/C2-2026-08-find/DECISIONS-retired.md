# Retired decisions — C2 find

Retired at close (2026-08-05) per `.charm/DECISIONS.md`'s retention sweep: each entry's
subject either no longer exists, was scoped to a now-complete task with no ongoing
constraint, or is fully self-contained in the shipped code/tests it explains. Moved here,
never deleted — see the `charm-close` report for the one-line reason each was retired. D53
and D57's still-relevant technical substance (lipgloss v2 row-rendering traps) was salvaged
into `.charm/knowledge/lipgloss-v2-row-rendering-gotchas.md` before retiring them.

## D32: S5 (match highlighting) is this cycle's pre-agreed cut

- Status: decided
- Cycle: C2
- Date: 2026-08-02
- Context: Highlighting fuzzy match positions inside a styled Result row, and a matched
  substring inside a possibly-truncated grep line, is the same ANSI-offset class of problem
  C1 already lost time to (`knowledge/viewport-search-ansi-offsets.md`).
- Decision: Keep grep in scope; if the cycle runs long, drop S5 first rather than trimming the
  Grep Picker.
- Why: Naming the cut before the pressure arrives is what stopped C1 trimming the wrong thing
  (C1/D8 pre-agreed cuts, none of which were needed). S5 is polish; grep is the substance.

## D35: C2's Picker hand-duplicates the modal dispatch again rather than building the shared abstraction

- Status: decided
- Cycle: C2
- Date: 2026-08-03
- Context: C1's code review left a roadmap item: a shared exclusive-input-mode dispatch
  abstraction across modals, "only worth it if a third modal is ever added" (C1/T31). The Picker
  is that third modal, so the condition it named has arrived.
- Decision: still hand-duplicate. M2.1/M2.3 add a third `if` gate in `Update` and a
  `handlePickerInput` in the shape of `handleFindInput`, and the roadmap item stays Unscheduled.
- Why: the three modals are less alike than the count suggests — Help swallows everything, Find
  types into one field and owns the footer, the Picker types into a field *and* drives a results
  list, a preview viewport, a mode and a Scope. The common factor is about four lines of
  ordering. Refactoring toward an abstraction while the third case is still being discovered
  would generalize from guesses; if a fourth modal arrives, the shape will be evident.

## D36: M1.2's `rg --files` runs with `--no-require-git`, and the Readable allowlist is exported from `internal/tree` rather than duplicated

- Status: decided
- Cycle: C2
- Date: 2026-08-03
- Context: Building M1.2 (scoped, allowlist-filtered file listing). Two small implementation
  calls surfaced that the plan didn't anticipate. (1) Verified against the sandbox's real rg
  (15.1.0): `rg --files` only honours a `.gitignore` file when the target directory is inside
  a git repository (has a `.git` ancestor) — a `.gitignore` in a plain, non-git folder is
  silently *not* applied by default. A noggin Root is just a folder, not necessarily a git
  checkout. (2) `internal/tree/policy.go`'s `isReadable` (D23's allowlist) is unexported, but
  `cmd/noggin`'s new search layer needs the identical check applied to rg's output — the same
  allowlist, not a second copy of it.
- Decision: (1) Pass `--no-require-git` on every `rg --files`/grep invocation the search layer
  makes, so a Root's own `.gitignore` counts regardless of whether it's a git checkout — this
  is what makes D28's "rg's ignore rules govern" hold for a Root in general, not just a Root
  that happens to be a git repo. (2) Export `isReadable` as `tree.IsReadable`; `cmd/noggin`'s
  `listFiles` (`files_search.go`) calls it directly rather than re-declaring the allowlist,
  keeping policy.go's "the one place it lives" comment true across the package boundary.
- Why: both low-risk and reversible (a CLI flag, an export), so logged rather than asked.
  Skipping (1) would mean a Picker in a non-git notes folder ignores its own `.gitignore` for
  no reason a user would expect, which is a worse surprise than the flag is a cost. Skipping
  (2) would mean the allowlist drifts the first time someone edits one copy and not the other
  — D23's whole point was there being one place to edit. M1.4 (grep) will shell out to rg the
  same way and should carry `--no-require-git` too.

## D37: The Picker result cap is 200, shared by both Pickers

- Status: decided
- Cycle: C2
- Date: 2026-08-03
- Context: Building M1.3 (fuzzy-match and cap the File Picker's candidates). S12 requires a
  cap and a truncation flag but SPEC/plan never fixed a number, and M1.4 (grep) needs the same
  cap so both Pickers report truncation consistently.
- Decision: `pickerResultCap = 200`, defined once in `files_search.go` for M1.3's `fuzzyFiles`
  and reused as-is by M1.4's grep search.
- Why: low-risk and reversible (one constant, no stored state depends on its value) — logged
  rather than asked. 200 is generous for a real note collection's file-name matches while still
  bounding a worst-case All Roots search; revisit if M1.4's grep (line-per-match, not
  file-per-match) turns out to want a different number.
- Also this task: `github.com/sahilm/fuzzy` (already present in the module graph as a
  transitive dependency of `charm.land/bubbles/v2`, per D28) is now imported directly by
  `cmd/noggin` and promoted to a direct `require` in `go.mod` via `go mod tidy` — no new
  dependency, just its existing presence made explicit.

## D39: The Picker opens on ctrl+p; Cancel (esc) dismisses it

- Status: decided
- Cycle: C2
- Date: 2026-08-03
- Context: Building M2.1 (Picker state plus open/dismiss). SPEC/PLAN name no key for opening
  the Picker (S1 only requires "openable from any Focus"); keymap.go's existing bindings
  (`/`, `e`, `ctrl+n`, `?`, `tab`) were all taken.
- Decision: a new global `keys.Picker` bound to `ctrl+p` opens it (`model.openPicker`), the
  same "reachable from every Focus" shape `Edit`/`Find`/`Navigator`/`Help` already have.
  Dismissal reuses the existing generic `keys.Cancel` (esc) rather than a new binding —
  `handlePickerInput` mirrors `handleHelpInput`'s shape (owns the keyboard while open, only
  Cancel does anything yet).
- Why: low-risk and reversible (one keymap.go entry, GLOSSARY.md already calls every binding
  "provisional") — logged rather than asked. `ctrl+p` is a common "quick open" convention
  (VS Code, Sublime, etc.) and collides with nothing already bound. M2.3 (Footer/Help Modal
  coverage) and M2.2 (layout) are the tasks that make the binding visible/discoverable; this
  task only wires the state machine.

## D40: The Picker's own non-typing keys are arrows/Tab/Shift+Tab, never a letter

- Status: decided
- Cycle: C2
- Date: 2026-08-03
- Context: Building M2.3 (always-typing input dispatch). S2 requires selection movement,
  Preview scrolling, mode switch and Scope switch to happen on keys that do not type, but
  names no specific keys. Every pane's existing movement bindings (`LineUp`/`LineDown`,
  `Top`/`Bottom`) double up a letter (`j`/`k`/`g`/`G`, GLOSSARY's "Nav Keys") with the arrow
  equivalent — reusing them as-is inside the Picker would mean those letters could never be
  typed into a query, breaking S2's "every printable keystroke goes to the Prompt" the moment
  a query contains one.
- Decision: two new bindings, `PickerUp`/`PickerDown`, bound to the arrow keys only (no
  letter alternate) for Results selection. Preview scrolling reuses the Document Viewer's
  existing `PageUp`/`PageDown`/`HalfPageUp`/`HalfPageDown` as-is — already non-printable
  (`pgup`/`pgdn`/`ctrl+u`/`ctrl+d`, `ctrl+b`/`ctrl+f`), so no second copy is needed. Mode and
  Scope switch get two more new bindings, `PickerMode` (`tab`) and `PickerScope`
  (`shift+tab`) — a modified key carries no `Text` (confirmed against `keyPress`'s own
  round-trip, keymap_test.go), so, like every other binding here, it can never collide with a
  typed character. All four new bindings are reserved as non-typing by this task
  (`handlePickerInput` recognizes and swallows them); M3.1/M3.2 give Results/Preview
  selection and scrolling their actions, M5.1/M5.2 give Mode/Scope theirs.
- Why: low-risk and reversible (four keymap.go entries, no stored state), so logged rather
  than asked. The alternative — reusing `LineUp`/`LineDown`/`Top`/`Bottom` as-is — would
  silently violate S2 for any query containing "j", "k", "g", or "G", which real search terms
  routinely do; picking disjoint keys up front avoids relitigating this once M3.1 gives
  Results selection something to actually move.

## D42: Without ripgrep the Picker still opens and says why, and every status line lives in the Prompt box

- Status: decided
- Cycle: C2
- Date: 2026-08-03
- Context: Building M2.5 (status and empty states). S14 says "the Pickers are unavailable, and
  reaching for one says why" without saying where "why" appears — refusing to open (a footer
  message, like `startEdit`'s "no document open to edit") and opening-but-inert both satisfy
  the words. Separately, S12's four status messages needed a home: the Results box (their
  natural place once M3.1 gives it rows) is `resultsPercent` of the modal's width, ~23 cells
  at 80x24, where "type at least 3 characters" wraps.
- Decision: (1) the Picker opens normally with no rg and reports `Picker unavailable: <reason>`
  in the modal, dispatching no search (`model.searchable` gates both search paths). (2) all
  status text renders as one line inside the Prompt box, under the textinput, with the
  mode/Scope indicator as that box's title (`promptOuterHeight` 4 → 5); `pickerStatus` picks
  exactly one message by priority (unavailable → below threshold → searching → no results →
  truncated).
- Why: low-risk, reversible, no spec surface — logged not asked. Opening keeps S1's "reachable
  from every Focus" literally true and puts the explanation where the user looked, rather than
  in a footer line the next keystroke clears (D22); it also keeps M2.1–M2.4's tests meaningful
  on a machine without rg. The Prompt box is the only region spanning the modal's full width,
  so it is the only place the longest message fits unwrapped at ordinary terminal sizes, and
  a caption directly under the query reads as being *about* that query. `pickerStatus` compares
  the results' own (query, mode, Scope) labels against the Prompt's rather than assuming they
  match, which is what makes "searching…" distinguishable from "no results" under M2.4's async
  fill.

## D43: `grepFiles`' unbounded rg argument list is a real gap, promoted to a plan task

- Status: decided
- Cycle: C2
- Date: 2026-08-03
- Context: TRIAGE T1 (review, medium): `grepFiles` (cmd/noggin/files_search.go:146-166) passes
  every file in Scope as a literal CLI argument to `rg`; an All Roots search over a large tree
  can push the argument list past the OS's ARG_MAX, and `cmd.Output()`'s error in that case is
  folded into the same "no results" path as a genuine no-match — silent, no signal the search
  didn't run.
- Decision: Accept as a plan task (M1.4a) to batch the file list across multiple `rg`
  invocations (or otherwise bound argument size), merging results under the existing cap/
  truncation contract. Not a hand-wave: it is a correctness gap in already-shipped, validated
  code, not a style nit.
- Why: medium risk (silent failure, no user-visible signal) on a path (All Roots grep over a
  realistically large collection) the cycle-level acceptance explicitly calls out — worth a
  tracked fix rather than a dropped note.

## D44: `listFiles` absorbing a live-but-unreadable Root folds into D36, no separate task

- Status: decided
- Cycle: C2
- Date: 2026-08-03
- Context: TRIAGE T2 (review, low): `listFiles`'s doc comment (cmd/noggin/files_search.go:40-46)
  documents that a live Root `rg` still cannot read (e.g. permission denied, or a TOCTOU race
  with `liveRootsInScope`'s own stat) is absorbed the same way a Broken Root is silently
  skipped — an extension beyond SPEC S6's literal text, logged nowhere unlike this milestone's
  other implementation calls (D36, D37, D38).
- Decision: Recorded here rather than as a plan task — the design is already implemented and
  correct, this closes the missing-decision-entry gap D36 should have covered. No code change.
- Why: low risk, no user-facing ambiguity (an unreadable Root produces no rows and no error,
  consistent with a Broken Root); the only gap was a documentation trail, now closed.

## D45: Grep row highlighting (M6.1/M6.2) test-coverage asymmetry, small task added

- Status: decided
- Cycle: C2
- Date: 2026-08-03
- Context: TRIAGE T3 (review, low): `TestFuzzyFilesCapsAndReportsTruncation` covers the
  exact-cap/not-truncated boundary as well as the over-cap case; `TestGrepFilesCapsAndReportsTruncation`
  (cmd/noggin/files_search_test.go:197-227,294-321) only covers the over-cap case — an
  asymmetry between the two Pickers' shared cap/truncation contract (S12/D37).
- Decision: Accept as a plan task (M1.4b) to add the same boundary case to the grep test.
- Why: cheap (effort 1), and the missing case is exactly the kind of off-by-one a cap
  contract shared by two code paths should verify symmetrically.

## D46: Help Modal's Picker-binding gap (T5) promoted to a plan task; M2.3 stays unvalidated until it lands

- Status: decided
- Cycle: C2
- Date: 2026-08-03
- Context: TRIAGE T5 (review, medium), corroborated by validate (PROGRESS.md M2.3 note): S2/
  M2.3's "Footer and Help Modal cover the Picker's actions" is only half met — `helpBindings`
  (cmd/noggin/help.go:14-40) lists the key that *opens* the Picker on every pane, but never the
  Picker's own in-Picker actions (PickerUp/Down/Mode/Scope) anywhere, and the Help Modal is by
  design unreachable while the Picker is open, so those actions live only in the footer's
  `pickerHintText`, which can itself truncate under a narrow terminal with no Help fallback.
- Decision: Accept as a plan task (M2.3a) to extend Help coverage of the Picker's own bindings
  from the panes Help remains reachable from, rather than declare the gap intentional. M2.3
  stays `Validated: -` until M2.3a lands and is validated.
- Why: medium risk — a documented, user-facing acceptance criterion (S2) is only half
  satisfied, and the footer fallback is itself fragile (truncation, no narrow-terminal escape
  hatch). Worth the small fix rather than retroactively narrowing S2's scope.

## D47: `picker.input` keeps its Go identifier; the "Prompt" glossary term stays UI-facing only

- Status: decided
- Cycle: C2
- Date: 2026-08-03
- Context: TRIAGE T6 (review, low): the GLOSSARY.md term "Prompt" is used throughout doc
  comments and status text, but the field holding it (`picker.input`, cmd/noggin/picker.go:37-49)
  is named `input`, not `prompt` — the one place this milestone's code diverges from the
  pattern where `mode`/`scope` match `pickerMode`/`pickerScope` exactly.
- Decision: Wontfix. Keep `input` as the field name.

## D48: `dismissPicker`'s one-sided preview-size carry-over is a real regression, promoted to a plan task

- Status: decided
- Cycle: C2
- Date: 2026-08-03
- Context: TRIAGE T6 (review, high): `dismissPicker` (cmd/noggin/picker.go:190-215) resets
  `m.picker` to `newPicker()`'s fresh 0x0 preview viewport and never restores the previous
  preview Width/Height, unlike `openPicker` (picker.go:158-175), which explicitly carries them
  over per its own "no memory except terminal-size fields" comment. `model.resize` only
  re-sizes the preview on `tea.WindowSizeMsg`/navigator-toggle, so opening the Picker a second
  time without an intervening resize leaves the Preview pane rendering empty (0x0) for the rest
  of the session — M3.2/S9 ("Preview pane renders … scrollable") silently breaks on the second
  open. No existing test exercises dismiss-then-reopen without a resize.
- Decision: Accept as a plan task (M3.2a): `dismissPicker` preserves/restores
  `preview.Width()/Height()` the same way `openPicker` does, plus a regression test for
  dismiss→reopen with no intervening resize. Flagged to land before M4.2, which reuses the same
  preview mechanism for grep's matched-line view.
- Why: high risk (a documented S9 acceptance criterion breaks on ordinary repeated use, not an
  edge case) on a mechanism M4.2/M4.3 are about to build on — worth fixing at the source before
  a second Picker mode inherits the same bug.

## D49: `activatePicker`'s untested read-error/grep-no-op paths and a weak expansion assertion, promoted to small plan tasks

- Status: decided
- Cycle: C2
- Date: 2026-08-03
- Context: TRIAGE T7/T8/T9 (review, medium), all in `cmd/noggin/picker.go`'s `activatePicker`
  (M3.3) and its test coverage: (T7) the read-error path (file removed/unreadable between
  listing and Activate) mirrors `activate()`'s tested error shape but has no picker_test.go
  equivalent, despite M3.3's own doc comment calling out the behavior explicitly; (T8) the
  claim "Activate is a no-op in Grep mode" is asserted only in a code comment, untested, right
  before M4.1 is about to make Grep mode real; (T9) `TestPickerActivateOpensResultAsActiveDocumentOnly`
  asserts the Files Nav tree's expansion is unchanged by comparing set *cardinality*
  (`len(got.files.expanded) != wantExpanded`) rather than identity, so a regression that swapped
  one expanded node for another at the same count would still pass.
- Decision: Accept as two small plan tasks: M3.3a (T7 read-error test + T8 grep-no-op test) and
  M3.3b (T9's key/set equality fix). Both are test-only additions/fixes, no production code
  change implied.
- Why: medium risk each — T7/T8 are unverified assumptions about already-shipped behavior that
  M4.1 is about to build directly on top of (grep mode's Activate semantics in particular); T9
  is a real gap in the exact acceptance criterion M3.3 claims to satisfy. Cheap to close now,
  before M4.x compounds on unverified ground.

## D50: Root-name collisions across Picker rows are in scope for C2, promoted to a plan task

- Status: decided
- Cycle: C2
- Date: 2026-08-03
- Context: TRIAGE T10 (review, medium): `pickerFileRow` (cmd/noggin/picker_view.go, ~208-220)
  labels a Result's Root with `filepath.Base(f.root.path)` only, so two Roots sharing a base
  directory name (different parents) render identical, ambiguous rows for same-named files in
  each — exactly the ambiguity S3 asks the Root label to resolve ("shows the file's Root and
  its path relative to that Root, so same-named files are distinguishable"). The code comment
  deferred this as out of this cycle's fixtures, a build-time call rather than something S3's
  text itself scopes out. `rootsNav.renderRow` already carries a left-truncated-parent
  disambiguation strategy for this exact case, unused here.
- Decision: In scope. Accept as a plan task (M3.1a): reuse `rootsNav.renderRow`'s
  disambiguation strategy in `pickerFileRow` rather than leaving the gap as an
  implementation-comment-only call or pushing it to Unscheduled.
- Why: this isn't a rare edge for noggin's own north star — personal notes, work notes and
  per-project docs plausibly share folder basenames (`notes`, `docs`) across different parents
  — and the fix reuses an existing, already-correct strategy rather than authoring a new one,
  so the cost is low relative to closing a real gap in S3's stated purpose.
- Why: low risk/value — the field wraps `bubbles/textinput.Model` generically, and `input`
  names what it *is* (a textinput wrapper) while "Prompt" names what it *means* in the UI;
  the two are allowed to diverge where a field wraps a third-party type generically. Renaming
  buys naming consistency at the cost of a slightly less honest field name.

## D51: `openPicker`/`dismissPicker`'s reset-sequence duplication (T14) promoted to a plan task

- Status: decided
- Cycle: C2
- Date: 2026-08-03
- Context: TRIAGE T14 (review, medium): `openPicker` and `dismissPicker`
  (cmd/noggin/picker.go:161-176, 199-206) now duplicate the identical 4-line "bump seq,
  snapshot preview size, replace with newPicker(), restore seq, restore preview size"
  sequence verbatim. This is the same one-path-forgot-the-carry-over shape D48/M3.2a already
  fixed once, in only one of the two paths — duplicating rather than sharing the sequence
  leaves a future third reset path (anticipated around M4.x, which reuses the Preview
  mechanism for grep) free to independently regress D48 again.
- Decision: Accept as a plan task (M2.1a): extract a shared reset helper both `openPicker`
  and `dismissPicker` call, so the carry-over logic exists in exactly one place.
- Why: medium risk — not a live bug today (both paths currently agree), but the exact failure
  mode D48 already cost a cycle to find once, on a mechanism about to gain a third caller.
  Cheap to close before M4.x compounds it.

## D52: `pickerRootLabel`'s double-counted gap budget (T13/T15/T16) promoted to a plan task

- Status: decided
- Cycle: C2
- Date: 2026-08-03
- Context: TRIAGE T13/T15/T16 (review, medium+low), all in `pickerRootLabel`
  (cmd/noggin/picker_view.go:212-214) and its test coverage: (T13) the `budget` calculation
  subtracts the connecting gap's width twice — once explicitly, once already embedded in the
  `rest` argument (`"  "+rel`, picker_view.go:232-236) — under-computing the parent-path
  truncation budget by one gap width (~2 cols) on every row, so M3.1a/D50's disambiguation fix
  truncates the parent segment more aggressively than the design intends, without overflowing;
  (T15) `TestPickerFileResultsShowRootAndRelativePath` derives its expected strings by calling
  `pickerRootLabel` itself, so the assertion would pass even with this bug present; (T16)
  M3.1a's own disambiguation test runs at 140x30 instead of the project's standard 80x24
  fixture, which masks the bug and leaves S3's disambiguation guarantee unverified at a
  realistic terminal size.
- Decision: Accept as a plan task (M3.1b): fix the budget formula, add a test that pins the
  full rendered row width against the layout's content width, pin at least one label string
  independently of `pickerRootLabel`, and confirm M3.1a's disambiguation test at 80x24.
- Why: medium risk — a shipped S3 acceptance criterion (row disambiguation) is measurably less
  effective than designed at the project's own standard terminal size, and the gap was hidden
  by two independent test-quality issues rather than caught by them. Cheap fix, reuses the
  already-agreed disambiguation approach from D50.

## D53: Grep Picker row format — no Root disambiguation, right-truncate before styling

- Status: decided
- Cycle: C2
- Date: 2026-08-03
- Context: M4.1 (S4) needed a concrete Grep Result row shape and a truncation strategy for
  arbitrary-length matched lines. Two things weren't pinned by the spec: (1) whether
  `pickerGrepRow` should reuse `pickerFileRow`'s Root-disambiguation treatment
  (`pickerRootLabel`, M3.1a/D50), and (2) how to bound a match line's length — `pickerFileRow`
  gets this for free because its budget math constructs a row that already fits; a grep
  match's line text is arbitrary length and does not. Naively relying on
  `lipgloss.Style.Width(...).MaxWidth(...)` (the pattern `pickerFileRow` already uses) does
  not work for this case: lipgloss word-wraps under `Width` before `MaxWidth` ever truncates
  (charm.land/lipgloss/v2 style.go), so an over-length match would silently wrap into several
  Results rows rather than truncate to one, breaking "one row per matching line" (M4.1's own
  acceptance criterion).
- Decision: `pickerGrepRow` (picker_view.go) renders `relPath:line  text`, relative to the
  match's own Root but without `pickerRootLabel`'s parent-path disambiguation — S4 only
  specifies "the file, the line number and the line", not Root disambiguation, and the line
  text itself is normally enough to tell same-named-file matches apart. Long lines are
  right-truncated (a new `rightTruncate`, the mirror of `rootsnav.go`'s `leftTruncate`) to the
  row's width *before* the row is styled, so the string handed to `Style.Width` is already
  short enough that its word-wrap is a no-op.
- Why: low-medium risk, build-time judgment call within M4.1's scope, not a spec ambiguity
  worth blocking on — recorded because the lipgloss Width/MaxWidth wrap-before-truncate
  behaviour is a easy-to-reintroduce trap for any future row renderer (M4.2's highlighted
  preview line, M6's match-highlighted rows) that assembles a row from unbounded content.

## D54: Grep preview positioning reuses find.go's line-level StyleLineFunc, keyed on (path, line)

- Status: decided
- Cycle: C2
- Date: 2026-08-03
- Context: M4.2 (S9) needed the Preview pane to scroll to and highlight a Grep Result's matched
  line. Two things weren't pinned by the plan: (1) whether to use `viewport.Model`'s own
  `SetHighlights`/`HighlightNext` or the `EnsureVisible`+`StyleLineFunc` pair find.go already
  uses for the Document Viewer's own search (S8), and (2) M3.2's `syncPreview` re-targets only
  when the *path* changes (`previewPath`) — insufficient for grep, where the Results cursor can
  move between two matches in the *same* file, which must still re-scroll/re-highlight.
- Decision: (1) Reuse the `EnsureVisible`/`StyleLineFunc` line-level approach, applied directly
  as `SetYOffset(line-1)` (an "opens scrolled to" jump, not an incremental "ensure visible from
  wherever it is now" the Document Viewer's search needs) plus a whole-line reverse-video
  `StyleLineFunc` — not `SetHighlights`, even though `.charm/knowledge/viewport-search-ansi-
  offsets.md`'s drift bug is specific to Glamour-styled content and the Preview pane's raw,
  unstyled text is not actually at risk of it: matching find.go's established pattern rather
  than introducing a second, narrower-scoped exception keeps the codebase's "how do we address
  a line in a viewport" story single. (2) Added a `previewLine` field alongside `previewPath`;
  `syncPreview`'s re-targeting key is now the pair, so a same-file/different-line Grep move
  re-scrolls and re-highlights without re-reading the file, while a File Result's `previewLine`
  is always 0 (M3.2's own contract untouched).
- Why: medium value, low risk — (1) is consistency with existing, already-verified practice
  rather than a new risk; (2) is a real gap this build would otherwise have shipped (a same-file
  Grep re-select silently not moving the Preview), caught by writing
  `TestPickerGrepPreviewRetargetsWithinTheSameFile` before assuming path-only comparison was
  enough.

## D55: Grep row rendering hygiene (T21, T22) promoted to a plan task

- Status: decided
- Cycle: C2
- Date: 2026-08-04
- Context: TRIAGE T21 (review, medium): `pickerGrepRow` (M4.1, picker_view.go:243-247,302-306)
  duplicates `pickerFileRow`'s exact 4-line style/`.Width(width).MaxWidth(width).Render(line)`
  tail verbatim, the same duplicated-sequence shape D51 already promoted to a plan task one
  commit earlier for `openPicker`/`dismissPicker` — and M6's highlighted rows are an
  anticipated third caller. TRIAGE T22 (review, medium): the grep row test
  (`TestPickerGrepResultsShowFileLineNumberAndText`, picker_view_test.go:295-322) derives its
  main expected strings by calling `pickerGrepRow` itself — the same tautological-assertion
  shape T15/D52 just fixed for the file row test — and never independently checks the match's
  line text is actually rendered.
- Decision: Accept as a plan task (M4.2a): extract a shared `styleRow` helper for both row
  renderers, and add a hand-computed expected string plus a rendered-text assertion to the
  grep row test. Bundled as one task since both were found in the same review pass over the
  same code.
- Why: medium risk on both — the styling duplication is about to gain a third independent
  drift point at M6, and the tautological test means a real text-dropping bug in
  `pickerGrepRow` would currently pass undetected. Cheap to close now, reusing the
  already-agreed patterns from D51 and D52.

## D56: Keep the batch-spread grep cap test as a forward-looking regression guard (T23)

- Status: decided
- Cycle: C2
- Date: 2026-08-04
- Context: TRIAGE T23 (review, low): `TestGrepFilesBatchedResultsExactlyAtCapAreNotTruncated`
  (M1.4c, files_search_test.go:367-401) aims to cover a batch-spread cap boundary distinct
  from the existing single-batch test, but `grepFiles`'s cap check runs per-appended-result
  inside the innermost loop regardless of batch boundaries (files_search.go:211) — so today it
  can't exercise a path the single-batch test doesn't already cover.
- Decision: Keep the test rather than drop or rewrite it; add a one-line comment (M7.1b)
  explaining it guards against a future batching refactor that moves the cap check outside
  the per-result loop, at which point it would start covering real distinct behavior.
- Why: low risk either way — the test is harmless and cheap to keep, and dropping it would
  lose a regression guard that becomes meaningful the moment the implementation it shadows
  changes shape. Not worth the churn of rewriting to force distinctness today.

## D57: Recompute Picker match positions at render time; highlight with bold only (M6.1, M6.2)

- Status: decided
- Cycle: C2
- Date: 2026-08-04
- Context: S5/M6.1 needs the fuzzy-matched character positions of a file name inside a
  rendered Results row. `fuzzyFiles` (files_search.go) discards `fuzzy.Match.MatchedIndexes`,
  so the positions were either carried out of the search alongside each result or recomputed
  at render time. Carrying them means a second result type (pickerFile + its match) threaded
  through `pickerResults`, the Preview/Activate accessors and every test that hand-builds a
  result set. Separately, the highlight has to read on both a plain row and the selected row's
  reverse-video bar (`pickerSelectedStyle`).
- Decision: Recompute via `fileNameMatchPositions` (files_search.go) — one `fuzzy.Find` over
  one short name per *visible* row — and guard the two against drift with a test asserting
  every row `fuzzyFiles` ranked is highlightable. Style matches with bold only, derived from
  the row's own base style (`base.Bold(true)`) rather than a standalone style, and render every
  run (matched and unmatched) with its style fully specified.
- Why: recomputation is exact, not approximate (same matcher, same string), so the only real
  cost is a bounded per-render match the drift test pins. Bold composes with reverse video,
  where a foreground colour would land on the selected row's text colour and can vanish;
  fully-specified runs avoid lipgloss's inner-style full reset (`ESC[m`) cancelling the
  selection bar mid-row — the failure mode
  `TestPickerFileRowHighlightKeepsSelectionBar` now pins.
- Extended at M6.2: the same shape covers grep rows. `grepMatchPositions` (files_search.go)
  reports rune positions for a literal, smart-case query occurrence — every occurrence in the
  line, since rg reports the line once however often it occurs — reusing find.go's
  `foldLower`/`origSpan` so S8's smart-case rule and its fold-length offset translation have one
  implementation. `highlightMatches` now takes row *columns*, with each mode supplying its own
  mapping (`suffixColumns` for file names, prefix offset + `clipColumns` for grep rows, whose
  match can be partly or wholly cut by `rightTruncate`).

## D58: M7.1's pty Preview-highlight check is tautological; M7.1 stays done, fix promoted to a plan task (T26)

- Status: decided
- Cycle: C2
- Date: 2026-08-04
- Context: TRIAGE T26 (review, medium): `test/pty/checks.py`'s M7.1 check "the Preview
  highlights the matched line (reverse video on the wire)" (lines 907-911) asserts only that
  `\x1b[7m` appears somewhere in the frame's byte stream. `pickerSelectedStyle`
  (picker_view.go:197, `Reverse(true)`) already renders the *selected Results row* in reverse
  video regardless of Preview behavior, so the check would still pass if the Preview's
  matched-line highlight were removed or broken — it doesn't prove what its name claims. Also
  redundant as written: `out` is `s.send()`'s return value, which is `self.pump()`'s return —
  the harness's own cumulative `self.out` buffer (test/pty/harness.py:74-76) — so `out` and
  `s.out` are the same object at that call site and the `or` branch is dead.
- Decision: M7.1 stays `done` — its own acceptance criterion ("`make check-pty` passes with
  new Picker checks ... including one that asserts the frame after dismissal matches the
  frame before opening") is satisfied as literally written and doesn't require this specific
  check to be non-tautological. Promote the fix to a small opportunistic task (M7.1c) rather
  than reopening M7.1: scope the assertion to something that can't pass without genuine
  Preview-side highlighting, and drop the redundant `or s.out`.
- Why: medium risk on the gap (a real regression in Preview highlighting could ship
  undetected by this pty check specifically), but low value to reopen M7.1 for — the
  underlying behavior is already independently verified by
  `TestPickerGrepPreviewOpensScrolledToMatchedLineWithHighlight` (a unit test, not pty-level).
  Cheap to close with a follow-up task before cycle close.
