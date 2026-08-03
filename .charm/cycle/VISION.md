# Vision — C2 find

<!-- Per-cycle. The raw, high-level goals for THIS cycle. Not formalized, possibly un-vetted;
     captured from the human during brainstorming. May contain contradictions and open ends —
     SPEC.md is where they get resolved. Free-form. Intent that reaches beyond this cycle
     belongs in ROADMAP.md, not here. The cycle key (C<n>) comes from the roadmap ledger. -->

## Goals

Telescope-grade fuzzy finding, as a modal UI over the whole session:

- **Fuzzy filename find** — the core value; if only one thing shipped, this is it.
- **Grep** — search text across files. Wanted in the same cycle even at the cost of a
  longer cycle; the pair is the unit of value, not filename-find alone.
- Both scoped to the Active Root or All Roots.
- Preview of the selected result before committing to it.

## Non-goals / boundaries

- **The allowlist holds.** Every C2 feature — grep included — stays inside C1's
  markdown-ish Readable file allowlist. D9 (general text/source viewing) is not answered
  by this cycle, by the back door or otherwise.
- C1's in-document `/` search is **not** replaced. `NOTES.md`'s "Document Search" (a fuzzy
  finder over the Active Document) is considered already covered by it and is dropped.

## Notes

- `ripgrep` may be required as an external dependency. Accepted.
- Naming needs settling this cycle and then holding: `/`-search vs. the new modal
  fuzzy-finding UI have been used loosely so far. Adopt a convention, retrofit anything
  from C1 that doesn't conform, stick with it.
- Roadmap items deliberately still out (from C1's Unscheduled): send results to the Files
  Nav ("browse-by-results"), scope = documents opened this session, repeat-last-search.
