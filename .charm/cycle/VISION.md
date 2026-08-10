# Vision — C4 recall

<!-- Per-cycle. The raw, high-level goals for THIS cycle. Not formalized, possibly un-vetted;
     captured from the human during brainstorming. May contain contradictions and open ends —
     SPEC.md is where they get resolved. Free-form. Intent that reaches beyond this cycle
     belongs in ROADMAP.md, not here. The cycle key (C<n>) comes from the roadmap ledger. -->

## Goals

The session remembers. Today noggin knows only what you tell it right now: which Root is
active, which document is open. Nothing carries across a switch, and nothing carries across a
restart except the Roots themselves. This cycle gives the *documents* the same durability the
Roots already have.

- **Starred documents** — star a document; it survives the session, like a starred Root.
- **Recent documents** — noggin records when a document was last viewed and can list documents
  by recency.
- **A way to browse by them** — the Files Nav stops being tree-only and gains browse modes:
  Tree, Recent, Starred. Recency and starring are useless without a surface that shows them.
- **Starred as a Picker scope** — the Picker gains "Starred" alongside its existing
  File/Grep mode and Active Root/All Roots scope axes.
- **Document history** — back/forward through the documents opened this session.
- **Polish: pane focus-key badges** — each pane shows the key that focuses it, right-aligned on
  its existing title row (`Files` left, `[f]` right). The Footer then drops its focus-key hints,
  which are the least contextual thing on it.

## Non-goals / boundaries

- **Wiki-links are out.** They need some notion of "which link is current", i.e. the document
  cursor C1/D4 deliberately refused. Deferred to Unscheduled, to be revisited as one piece
  alongside the other cursor-dependent features (Select mode + copy, URL following).
- **Recency, not frecency.** Ranked by last-viewed timestamp only; no visit counts, no decay.
- **Peek does not count as viewing.** Activate and the `$EDITOR` handoff do.
- **Tags are not this cycle** — they are C5's Browse-By-adjacent feature, and the "Results"
  browse mode that `NOTES.md` pairs with them stays out with them.

## Notes

- `NOTES.md` describes browse modes as a single list behind a modal "Browse By Picker"
  (Tree / Recent / Recent-All / Starred / Starred-All / Results). Worth evaluating instead:
  direct key commands, and splitting the modes into independent axes the way the Picker does
  (scope: Active Root vs All Roots; mode: Tree vs Recent vs Starred).
- C2/D31 declared the Picker stateless and query-only, with "remembering is C4's job". Adding
  a Starred scope is a filter over what the Picker searches, not Picker memory — but the
  boundary is worth stating so it is not read as a contradiction.
- The starred-Roots store already exists in the config directory, hand-editable (C1/D7), with
  a standing Unscheduled item to move it to a data location. Per-document timestamps are
  unambiguously data, so this cycle forces the question.
