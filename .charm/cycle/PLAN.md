# Plan — C4 recall

<!-- Per-cycle. The spec decomposed into milestones and tasks. Technically aware but
     high-level: state WHAT and the done-condition, not step-by-step HOW — trust executor
     judgment. Every task MUST have an acceptance criterion, machine-checkable where possible,
     and an effort estimate: a relative 1-5 gut-scoping (1 trivial, e.g. fixing a few typos;
     5 very high, e.g. large or complicated tasks) — not a rigorous prediction, just how the
     task looks at scoping time. Durable keys: milestones M1, M2...; tasks M1.1, M1.2...
     Never renumber; append or suffix (M1.2a) when inserting. Keys are scoped to this cycle —
     durable docs referencing them qualify with the cycle key (C3/M1.2). Note dependencies
     where ordering matters. -->

Ordering note: M1 gates almost everything. M7 (focus badges) is fully independent and can be
pulled forward if a short slice is wanted early; it also carries the Files Nav title-row change
that M4.1 builds its mode/scope row directly beneath.

## M1 — The document store

Delivers: S1

- **M1.1** — Add data-location resolution (distinct from `os.UserConfigDir()`) and move the
  starred-Roots store there as a clean break: a config-directory `roots` file is ignored, not
  read or migrated — done when: a test with a populated config-dir store and an empty data-dir
  store starts with zero starred Roots, and star/un-star round-trips through the data-dir path
  — effort: 2
- **M1.1a** — Update `make check-pty`'s harness (`test/pty/checks.py`'s `config_env`) to
  resolve the starred-Roots store path the same way the shipped code does post-M1.1
  (`dataDir()`/`storePath()`, not the old config-directory path) — done when: `make check-pty`'s
  Roots pane listing/marking/persistence checks (C3's M3.3/M3.4) pass again — effort: 2
  (needs: M1.1) — from TRIAGE T4 (D71).
- **M1.2** — A path-keyed document store: one record per absolute document path carrying
  `starred` and `lastViewed`, with load/save that tolerates a missing or malformed file the way
  the Roots store does, and no attachment to Roots — done when: unit tests cover round-trip,
  absent-file, and garbage-line cases, and reading a record for a path under no Root succeeds
  — effort: 3 (needs: M1.1)
- **M1.3** — Bound the store: evict oldest `lastViewed` first, never evict a starred record,
  and eviction clears the timestamp while leaving the star — done when: a test loads the store
  past its cap and confirms the oldest unstarred records are gone, every starred record
  survives, and a starred record that was evicted from recency keeps its star with no timestamp
  — effort: 2 (needs: M1.2)
- **M1.4** — Wire the store into the model lifecycle: loaded once at startup, persisted
  immediately on change, single accessor pair used by every reader/writer — done when: the
  model exposes store reads to Files Nav / viewers / Picker without any of them touching the
  file directly, and a restart test sees the previous session's records — effort: 2
  (needs: M1.2)

## M2 — Starring a document, and seeing it

Delivers: S2, S3

- **M2.1** — Star / un-star the Files Nav's selected file, persisted — done when: a test stars
  a file, rebuilds the model against the same store, and the star is still set; un-starring
  clears it — effort: 2 (needs: M1.4)
- **M2.2** — Star / un-star from a focused Document Viewer, acting on *that* viewer's Active
  Document — done when: a dual-viewer test stars from DV1 and from DV2 and confirms two
  different documents were starred, with neither action affecting the other viewer — effort: 2
  (needs: M1.4)
- **M2.3** — Star marker on Files Nav rows in every browse mode except `Starred` — done when: a
  render test shows the marker on a starred file's Tree row, absent on an unstarred one, and
  absent from every row of a `Starred` list — effort: 2 (needs: M2.1; marker-in-Starred clause
  needs M4.3)
- **M2.4** — Star indicator to the left of the document title in each Document Viewer's title
  row, legible with the Navigator hidden — done when: a render test with the Navigator hidden
  shows the indicator in DV1's and DV2's title rows for starred documents only, and the title
  row still degrades sanely at narrow widths — effort: 2 (needs: M1.4)
- **M2.5** — Star / un-star the Picker's selected Result; un-starring leaves the Result in place
  with its marker cleared — done when: a Picker test un-stars the selected Result and confirms
  the row and cursor position are unchanged while the marker is gone — effort: 2
  (needs: M1.4) — *pre-agreed cut 1: drop this task first if the cycle spills.*
- **M2.6** — Confirm the Root lifecycle never destroys document state: un-starring a Root leaves
  its documents' stars and timestamps intact — done when: a test un-stars a Root holding
  starred and recently-viewed documents and confirms every document record is byte-identical
  afterwards — effort: 1 (needs: M2.1, M3.1)
- **M2.7** — Keep the Files Nav's starred marker in sync when a document is starred/un-starred
  from a focused Document Viewer or the Picker, not only from the Files Nav itself (`m.docs`
  mutators other than `toggleDocStar` currently leave `m.files.starred` stale) — done when: a
  test stars the Active Document from DV1 while the same file is a Files Nav row and confirms
  the row's marker updates without a manual refresh, and likewise for a Picker star toggle
  — effort: 2 (needs: M2.1, M2.2, M2.5) — from TRIAGE T7 (D70).
- **M2.8** — Add `keys.Star` and `keys.PickerStar` to the Footer/Help Modal hint lists for the
  panes and Picker where each applies, so both star actions are discoverable outside the Roots
  Nav — done when: `keys.Star` appears in the Files Nav's and each Document Viewer's
  `paneHints`/Help Modal listing, `keys.PickerStar` appears in the Picker's, and
  `TestHelpModalListsThePaneOwnBindings`'s existing pins still pass (reworded if needed)
  — effort: 2 (needs: M2.1, M2.2, M2.5) — from TRIAGE T5, T6.
- **M2.9** — Naming cleanup: rename `TestLoadStarredRootsNoConfigDir` →
  `TestLoadStarredRootsNoDataDir`, `evictDocStore`'s `cap` parameter → `limit`, and
  `toggleStarPicker` → `togglePickerStar`, for consistency with each area's own conventions
  — done when: `make check` is green with the new names and no reference to the old names
  remains — effort: 1 (needs: M1.1, M1.3, M2.5) — from TRIAGE T1, T2, T9.
- **M2.10** — Update `make check-pty`'s C2a/M4.1 git-status-marker checks for the two-char
  star-marker column M2.3 inserted between the status marker and the file name (row layout is
  now status + star + name, so an unstarred modified file renders `"M   tracked.md"`, not
  `"M tracked.md"`) — done when: `make check-pty` passes against the current row layout
  without weakening what the checks verify — effort: 1 (needs: M2.3) — from TRIAGE T11.
- **M2.11** — Add S3's third star-marker surface: a starred Result's row (File and Grep
  Picker) shows the same `\u2605 `/`  ` marker Files Nav and Roots Nav rows already use, computed
  live from the S1 document store on every render (no cache to go stale, unlike M2.7's Files
  Nav marker) — done when: a render test shows the marker on a starred Result's row in both
  Picker modes, absent on an unstarred Result's row with the rest of the row byte-identical,
  and confirms toggling `PickerStar` shows/clears the marker in the same frame with no
  separate refresh — effort: 2 (needs: M2.5) — from TRIAGE T8 (D72).
- **M2.12** — Correct `starMarker()`'s doc comment, which currently claims the two-column
  marker field is already shared by files.go, rootsnav.go, model.go's Document Viewer title
  and Picker rows — in reality only Picker rows call it; the other three independently render
  an identical literal — done when: the comment no longer overclaims consolidation that hasn't
  happened, and `make check` is green — effort: 1 (needs: M2.11) — from TRIAGE T17 (D76).
- **M2.13** — Add a Grep Picker variant of M2.11's same-frame star-marker toggle render test,
  closing the mode-parity gap in its acceptance wording — done when: a new test drives a real
  `m.View()` render through a `PickerStar` toggle in Grep mode and confirms the marker
  appears/clears in the same frame, mirroring the existing File Picker test — effort: 1
  (needs: M2.11) — from TRIAGE T19 (D78).

## M3 — Recency is recorded

Delivers: S4

- **M3.1** — Record `lastViewed` when a document is Activated in DV1 or DV2, or opened via the
  `$EDITOR` handoff; Peek never records — done when: tests confirm a timestamp after Activate
  in each viewer and after an edit handoff, and confirm a Peek (including a `ctrl+` peek into
  DV2) leaves the record untouched — effort: 2 (needs: M1.4)
- **M3.2** — Record regardless of Root provenance: documents under ad-hoc, un-starred, or
  absent Roots all get timestamps — done when: a test Activates a document from an ad-hoc Root
  and one whose Root is not in the session at all, and both records persist — effort: 1
  (needs: M3.1)
- **M3.3** — Record `lastViewed` when a document is Activated via the File or Grep Picker into
  DV1 or DV2, not only via the Files Nav's own Activate — `activatePickerInto` and
  `activateGrepPickerInto` currently call `openDocumentInto` directly, bypassing `activateInto`
  and therefore `recordViewed` entirely — done when: a test Activates a File Picker Result and
  a Grep Picker Result, into both DV1 and DV2, and confirms a timestamp lands in each case
  — effort: 2 (needs: M3.1) — from TRIAGE T12 (D74).
- **M3.4** — `recordViewed`'s store-write failure path must not silently overwrite a
  just-set Activate/Edit success message with an unrelated store-save error — done when: a
  test forces the store write to fail during an Activate or Edit and confirms the resulting
  message reflects the store failure without discarding it as an untraceable overwrite —
  effort: 1 (needs: M3.1) — from TRIAGE T13.
- **M3.5** — Extend M3.4's store-write-failure regression coverage to the Picker's own
  `recordViewed` call sites (`activatePickerInto`, `activateGrepPickerInto`), mirroring the
  existing DV1/DV2 tests — done when: a File Picker and a Grep Picker variant of
  `TestActivateStoreWriteFailureSurfacesMessage` exist and pass — effort: 1 (needs: M3.3, M3.4)
  — from TRIAGE T18 (D77).

## M4 — Files Nav browse modes

Delivers: S5, S6, S7 (and S2's Starred-list removal clause)

- **M4.1** — Mode and scope model plus the two cycle keys (Files-Nav-Focus-only, no modal),
  and a mode/scope row directly below the Files Nav title row — done when: from Files Nav Focus
  one key cycles `Tree`→`Recent`→`Starred` and the other cycles that mode's scope values
  (`Tree` reads `-` and its scope key is a no-op), the row shows the current pair, the keys do
  nothing from any other Focus, and the file list is one row shorter in every mode — effort: 3
  (needs: M1.4)
- **M4.2** — `Recent` mode lists: flat rows with file name left and Root base name (or, for a
  document under no session Root, the left-truncated parent path) right; at most 100 documents
  per Root; `All Roots` merges by timestamp with no further cap; records whose file is gone are
  silently omitted — done when: tests cover ordering by timestamp, the per-Root cap, the merged
  `All Roots` ordering, the two right-hand label forms, and omission of a deleted file
  — effort: 3 (needs: M4.1, M3.1, M3.3)
- **M4.3** — `Starred` mode lists across `Active Root` | `All Starred` | `Orphans`, where a
  broken starred document is listed and marked but cannot be Activated, and un-starring a
  displayed row removes it immediately — done when: tests confirm each scope's membership
  (including a starred document under no Root appearing in `All Starred` and `Orphans` only),
  that activating a broken star is refused the way a Broken Root is, and that un-starring drops
  the row — effort: 3 (needs: M4.1, M2.1) — *`Orphans` is pre-agreed cut 3.*
- **M4.4** — Every mode/scope combination with nothing to show displays a brief message saying
  so — done when: a render test walks all six combinations empty and each shows a message
  rather than a blank pane — effort: 1 (needs: M4.2, M4.3)
- **M4.5** — Flat-list rows support the Files Nav's document actions — Activate, Peek, Edit and
  their `ctrl+` DV2 targeting — while Expand/Collapse are no-ops — done when: tests activate and
  peek a `Recent` row into DV1 and into DV2, and confirm Expand/Collapse change nothing
  — effort: 2 (needs: M4.2)
- **M4.6** — Session-only browse memory: mode remembered per Root, scope remembered per
  (Root, mode); every session starts on `Tree` — done when: a test sets distinct mode/scope
  pairs on two Roots, switches away and back to each, gets the last pair for each, and a fresh
  model starts on `Tree` — effort: 2 (needs: M4.1)
- **M4.7** — Re-clamp the flat-list selection index wherever the effective row set can shrink:
  `cycleMode`, `cycleScope`, `setRoot`'s flat-mode memory restore, and un-starring the selected
  row out of a Starred list — and correct `selectedFlatRow`'s doc comment, which still claims
  Peek/Edit/DV2-targeting await M4.5's fix — done when: selecting a lower Tree row then cycling
  to a shorter Recent/Starred list leaves Activate/Peek/Star acting on a valid row (not a
  silent no-op), the same holds after un-starring shrinks the Starred row set, and the comment
  no longer describes a pre-M4.5 state — effort: 3 (needs: M4.5, M4.6) — from TRIAGE T23, T24
  (D82); resolves TRIAGE T21 (dropped as already-handled — see D82).
- **M4.8** — Add `keys.FilesMode` and `keys.FilesScope` to the Files Nav's Footer hints and
  Help Modal listing, mirroring M2.8's precedent for `keys.Star` — done when: both bindings
  appear in `filesPane`'s `paneHints` and the Help Modal listing, and
  `TestHelpModalListsThePaneOwnBindings`'s existing pins still pass (reworded if needed)
  — effort: 2 (needs: M4.1) — from TRIAGE T20 (D81).
- **M4.9** — Give Tree mode its own empty-state message, distinct from "no Root open," for an
  open Root with zero visible rows (empty folder, or every top-level entry gitignored) —
  matching the per-scope wording M4.4 already gave Recent/Starred — done when: a render test
  shows a different message for an open-but-empty Root than for no Root open at all — effort: 1
  (needs: M4.4) — from TRIAGE T22.
- **M4.10** — Close four review-flagged test-coverage gaps: a tie-break test for
  `sortByRecency` and `starredRows`' name sort (two records sharing a timestamp/name), an
  orphan-bucket test with >100 documents confirming the cap applies there too, a `view()`
  height==1 test, and a star-marker assertion on a starred Recent-mode row — done when: all
  four new tests exist and pass — effort: 2 (needs: M4.2, M4.3, M4.4) — from TRIAGE T25, T26,
  T29, T31 (D83).
- **M4.11** — Fix broken `\u2014` comment literals (real text, not a processed escape, inside
  `//` comments) and the wrong "scope.go" citation (should read browse.go) across
  `browse_list.go` and its test — done when: no comment contains a literal `\uXXXX` escape,
  both citations read `browse.go`, and `make check` is green — effort: 1 — from TRIAGE T27, T28.
- **M4.12** — Close M4.8's undertested additions: strengthen
  `TestHelpModalListsThePaneOwnBindings`'s `filesPane` case with an identity-based assertion
  (matching `b.Help().Key`/`.Desc`, as `TestHelpBindingsListPickersOwnBindingsFromEveryPane`
  already does) so it can catch a regression dropping `keys.FilesMode`/`keys.FilesScope` from
  `filesPane`'s `helpBindings` — currently indistinguishable from `keys.PickerMode`/
  `keys.PickerScope`'s identical "mode"/"scope" text, which are already listed unconditionally
  for `filesPane` — and add a new footer hint test (`TestPaneHintsNameFilesModeAndScopeOnly-
  InFilesPane`, mirroring `TestPaneHintsNamePeekOnlyInFilesPane`'s shape) asserting
  `keys.FilesMode`/`keys.FilesScope` actually appear in `filesPane`'s hint text — done when:
  both the strengthened Help Modal assertion and the new footer test exist and pass — effort:
  2 (needs: M4.8) — from TRIAGE T32, T33 (D86).
- **M4.13** — Fix `flatEmptyText`'s `default` branch, whose own doc comment (post-M4.9) now
  calls it an unreachable defensive fallback while it still returns `emptyFilesText` ("no Root
  open") rather than `emptyTreeText` ("open but empty"), the message a Tree-mode caller would
  actually want if the branch were ever hit; correct the returned value or reframe the comment
  accurately, whichever stays true. Also add a one-line comment to
  `TestRecencyAndStarredNameTiesBreakByPath` noting its tie-break assertion holds because of
  the literal subdirectory names `"a"`/`"b"`, not `t.TempDir()` ordering — done when:
  `flatEmptyText`'s default case and its comment agree, the tie-break test's comment is added,
  and `make check` is green — effort: 1 (needs: M4.9) — from TRIAGE T34, T36 (D87).
- **M4.14** — Close M4.7's acceptance-criterion literal gap: add regression tests exercising
  `cycleScope` shrinking the flat row set and `setRoot`'s flat-mode memory-restore branch,
  mirroring the existing `cycleMode`/un-star coverage already in `flat_selection_clamp_test.go`
  — done when: two new tests assert `f.selected` stays valid after each of those two
  transitions, and pass — effort: 2 (needs: M4.7) — from TRIAGE T37 (D88).
- **M4.15** — Fix `openFolder`'s selection reset when switching the Active Root back onto one
  previously left in a flat browse mode: `filesNav.setRoot`'s own clamp runs against the
  *outgoing* Root's still-stale `activeRootPath` before `model.refreshFilesRecentSource`
  re-points it and re-clamps against the correct list, so the restore lands on row 0 instead
  of the nearest surviving row from M4.6's memory — done when: a test drives `openFolder`
  itself (not `setRoot` in isolation) switching away from and back to a Root in flat mode and
  confirms selection lands on the memory-restored row, not index 0 — effort: 2
  (needs: M4.6, M4.7) — from TRIAGE T38.

## M5 — `Starred` as a Picker scope

Delivers: S8

- **M5.1** — Add `Starred` as a third value on the Picker's scope axis in both the File and the
  Grep Picker, resolving to the whole store (every starred document, orphans included) rather
  than a set of Roots, with a three-way cycle and a label for it — done when: the scope key
  cycles all three values in both modes, a Grep and a File search under `Starred` return
  matches from a starred document under no session Root, and switching Picker mode never
  rewrites the scope — effort: 3 (needs: M1.4, M2.1)
  — *Grep-side `Starred` is pre-agreed cut 4.*
- **M5.2** — Delete the now-dead `grepFiles` (superseded by M5.1's `grepPickerFiles` split;
  `pickerSearchCmd` no longer calls it) and repoint `files_search_test.go`'s existing coverage
  at `grepPickerFiles(listFiles(...), query)` directly, so no stale caller or doc-comment
  claim survives — done when: `grepFiles` has no remaining references, its former tests pass
  against the new call shape, and `make check` is green — effort: 2 (needs: M5.1) — from
  TRIAGE T42 (D94).

## M6 — Per-viewer document history

Delivers: S9

- **M6.1** — Each Document Viewer keeps its own session-only back/forward history with browser
  semantics, traversable by two keys: activating after going back discards forward entries, the
  same document may recur but never adjacently, the `$EDITOR` handoff and Refresh create no
  entries, and traversal never changes the Active Root — done when: tests drive DV1 and DV2
  through independent back/forward sequences, confirm forward-truncation, no adjacent
  duplicates, no entry from edit or Refresh, an unchanged Active Root when history walks into
  another Root's document, and that DV2's history dies with DV2 — effort: 3
- **M6.2** — Going back or forward restores that document's most recent scroll position,
  remembered per document rather than per history entry — done when: a test scrolls a document,
  navigates away and back, and lands at the remembered offset — and revisiting via a *different*
  history entry for the same document lands at the same offset — effort: 2 (needs: M6.1)
  — *pre-agreed cut 2.*
- **M6.3** — Add `keys.HistoryBack`/`keys.HistoryForward` to the Document Viewer's Footer
  hints and Help Modal binding list, mirroring M2.8's and M4.8's precedent for newly
  introduced bindings — done when: both bindings appear in the viewer's `paneHints` and Help
  Modal listing, and `TestHelpModalListsThePaneOwnBindings`'s existing pins still pass
  (reworded if needed) — effort: 2 (needs: M6.1) — from TRIAGE T41.
- **M6.4** — Add a regression test for D91's flagged-but-untested behavior: scroll a document,
  then re-Activate that same document with no navigation in between, and confirm the offset
  survives instead of resetting to top, mirroring `history_scroll_test.go`'s existing style —
  done when: the test exists and passes — effort: 1 (needs: M6.2) — from TRIAGE T43.

## M7 — Pane focus badges

Delivers: S10

- **M7.1** — Each pane shows its focus key right-aligned on its existing title row (`[r]`,
  `[f]`, `[o]`, `[^o]`), whether or not the pane holds Focus, dropped entirely when the pane is
  too narrow for title plus badge — done when: render tests show all four badges at a normal
  width, the badge present on unfocused panes, and the badge (not the title) dropped at a width
  that fits only the title — effort: 2
- **M7.2** — Remove the focus-key hints from the Footer, leaving Refresh, Help, Quit,
  pane-local hints and transient messages untouched — done when: no `keys.Focus*` binding
  appears in any Footer hint list and a Footer render test at a wide width shows the remaining
  elements unchanged — effort: 1 (needs: M7.1)
  <!-- S10 says four hints; only FocusRoots/FocusFiles/FocusDV1 are actually listed
       (cmd/noggin/footer.go:36,38,44) — FocusDV2 was never in the Footer. Three removals,
       same outcome. -->
