# Vision — C6 sources

<!-- Per-cycle. The raw, high-level goals for THIS cycle. Not formalized, possibly un-vetted;
     captured from the human during brainstorming. May contain contradictions and open ends —
     SPEC.md is where they get resolved. Free-form. Intent that reaches beyond this cycle
     belongs in ROADMAP.md, not here. The cycle key (C<n>) comes from the roadmap ledger. -->

## Goals

- Bring Wikipedia into noggin as a first-class thing you can search, read, star and revisit —
  **not** as a separate TUI mode, but by widening what a "collection" is so Wikipedia sits in
  the same Navigator / Files Nav / Picker / Document Viewer paradigm as a folder Root.
  - Reuses components, but more importantly reuses the *paradigm* (UX win).
  - Unlocks cross-class viewing in one context: local md beside a Wikipedia article in DV1/DV2 —
    noggin's core use case (find, read, compare text artifacts).
- Redefine collections/items so the model extends to the next class (dictionary/thesaurus,
  on the roadmap) without another redefinition. The dictionary itself is *not* built.
- While Roots are being refactored: persisted (non-ad-hoc) Roots get an editable name,
  defaulting to the folder name.
- A batch of small polish items ridden along in the same cycle (see Notes).

## Non-goals / boundaries

- No new TUI "mode" for Wikipedia.
- Dictionary collection: shapes the model, is not delivered.
- Wikipedia is excluded from every grep operation (would require persisting or on-demand
  fetching every starred article — not worth it).
- Starred Wikipedia articles are *not* in scope for an "All Roots" File search; the only
  cross-class File search is the Picker's Starred scope.
- Wikipedia is not a "Tree": it cannot be browsed, only searched, then starred / seen in Recent.

## Notes

Raw model as pitched:

- New collection classes differ from Roots: cannot be created or removed, exactly one of each,
  source of truth is not local.
- The deeper commonality: *neither* class persists its source of truth. A Root queries the
  filesystem the way Wikipedia queries a service — content is always fetched, never stored.
  The real difference is bounded-and-crawlable (file tree) vs. finite-but-unbounded (Wikipedia,
  dictionary). Hence: no tree browse, no grep; documents are reached only via search, then
  starred (stores a pointer: URL + title) or seen in Recent for a time.
- Terminology candidate: "collection" → **"source"**; Root → "file source"; Wikipedia/dictionary
  → "query source". To be decided.
- Files Nav when Wikipedia is the Active Source: no Tree mode; Starred shows starred articles;
  Recent shows recently viewed articles with a sensible time and/or count cap — maybe
  session-scoped.
- Picker: Wikipedia search as its own Picker mode reachable from any Root (leaning this way)
  rather than only as "File · Active Root" when Wikipedia is active.
- Polish riders:
  - pgup/pgdn scroll half a page, not a full page
  - inactive-pane dimming (darker active background, or lighter inactive panes)
  - lighter text: Roots Nav path line; the non-matching portion of Picker result paths
  - drop the root path fragment from Picker results
  - Document Viewer bottom line, lightened: root name + file path relative to the root
  - Picker minimum query length 3 → 2; consider a keystroke-to-search delay (debounce)
- Roadmap: un-plan the old C5 "organize" (tags/outline/file management) → Unscheduled. This
  is what C6 "collections" was aimed at, so it takes the C6 key (D94 already pointed there);
  C5 is withdrawn, never reused.
- Terminology settled in conversation: Source (umbrella), Root (unchanged), Query Source;
  Roots Nav → Sources Nav; Collection retired as an in-app term.
- Agreed in conversation: title search only; no Preview for articles; canonical article via
  redirects, English only; Recent for articles identical to files; "lighter" text means
  de-emphasized (less contrast); focus indication via background (evaluation, pre style pass);
  half-page paging DV-only; cut order = article Preview (already none) → articles in Recent →
  Root names → background focus.
