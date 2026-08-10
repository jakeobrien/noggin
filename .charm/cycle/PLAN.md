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
  — effort: 3 (needs: M4.1, M3.1)
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

## M5 — `Starred` as a Picker scope

Delivers: S8

- **M5.1** — Add `Starred` as a third value on the Picker's scope axis in both the File and the
  Grep Picker, resolving to the whole store (every starred document, orphans included) rather
  than a set of Roots, with a three-way cycle and a label for it — done when: the scope key
  cycles all three values in both modes, a Grep and a File search under `Starred` return
  matches from a starred document under no session Root, and switching Picker mode never
  rewrites the scope — effort: 3 (needs: M1.4, M2.1)
  — *Grep-side `Starred` is pre-agreed cut 4.*

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
