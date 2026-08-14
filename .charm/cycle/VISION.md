# Vision — C4a wikipedia-spike

<!-- Raw goals for THIS cycle, captured during /brainstorm. Unpolished is fine; it is input
     to SPEC.md, not a contract. Intent reaching past this cycle belongs in ROADMAP.md. -->

## Raw pitch (verbatim intent, 2026-08-14)

- Add a **tab-based modal navigation paradigm** to noggin. Tabs occupy the top row of the
  TUI, selected with number keys.
- `1` = `Notes` — everything noggin is today.
- `2` = `Wikipedia` — TUI-based **search and reading of Wikipedia articles**, rendered as
  markdown through the same glow/glamour pipeline as Notes.
- Method: standard Wikimedia REST API (`/api/rest_v1/page/html/{title}`) → goquery cleanup →
  Go `html-to-markdown` (see `wikipedia-to-markdown.md` at repo root).
- Out of scope this cycle: parsing/following links inside an article; rendering images
  (not even deferred — not wanted); starring articles (**on the roadmap**).
- Likely crossover with future Notes-tab features.
- "Tab" is not settled terminology — open to a better fit.

## North star dilation (2026-08-14, Q1)

The original north star was a deliberately narrow subset — a hedge against the value not
materialising. Confidence is now high, so the vision dilates:

> noggin is a multi-modal **thoughtspace**: navigate seamlessly between sources of
> knowledge, synthesize them, and remember them. Search, navigation and consumption —
> never editing. Light text input at most; it works *with* your editor, not instead of it.

Sources are chosen for a specific value, not accumulated on whim:

| Source | Value it carries |
|---|---|
| Notes | personally-relevant information; the **synthesis target** |
| Wikipedia | authoritative, highly trustable information |
| ~~LLM chat~~ | **cut** — chat *creates* documents; it belongs outside noggin, like editing. Handoff to a separate chat TUI, and read its saved conversations back as an ordinary Root |
| Dictionary / thesaurus (future) | the universal glossary — the medium everything else is written in |

Two features reach *across* sources and are where the value compounds: **starring** and
**link-following** — connections and collections that span scopes.

## Open threads (from interview)

- Roadmap placement: insert as C5, push `organize` to C6 (agreed).
- Appetite: **(a)** — cheapest viable Wikipedia reader; a prototype/kill-gate is expected,
  and the feature is allowed to not survive.
## Resolved (2026-08-14)

- **Typed Roots, not channels.** Wikipedia is a new *kind of Root*, not a new navigation
  paradigm. No top-row bar, no number keys, no new terminology (`tab`/`channel` dropped).
  The product sharpens from "everything I use to think" — messy, scales irregularly — to
  **a browser of document collections**: tightly scoped, and able to grow to man pages,
  cheat sheets, docsets, dictionaries.
- Wikipedia root: **built-in, always present, English only**, not addable/removable.
- Links render as **plain text** (markup stripped). Following links stays out of scope.
- Appetite **(a)**: cheapest viable reader. Explicit kill-gate — the feature may not survive.
- Kill-gate sample articles (judged by eye, by the user):
  - `Why_is_there_anything_at_all%3F` (simple prose)
  - `E_(mathematical_constant)` (maths)
  - `List_of_programming_languages_by_type` (many subsections / lists)
  - `Major_Arcana` (tables)
