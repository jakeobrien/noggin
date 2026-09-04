# Spec — C4a wikipedia-spike

<!-- Per-cycle. Functional, non-technical, concrete. This is the contract of intent: what the
     deliverables ARE, not how they are built. Keep it product-level so upfront errors don't
     cascade into implementation. Durable keys (S1, S2...) — never renumber. Mid-cycle
     changes go through /brainstorm amend: supersede or suffix keys, renew the agreed date;
     charm-replan then reconciles the plan. Keys are scoped to this cycle — durable docs
     referencing them qualify with the cycle key (C3/S2). -->

Status: agreed 2026-09-02 (originally 2026-08-14; amended per D104)

## Intent

This cycle answers **one question**: can a Wikipedia article be turned into markdown that is
genuinely worth reading inside noggin? It ships **no user-facing feature**. Its output is a
throwaway command-line spike, four converted articles judged by eye, a findings record, and a
**go/no-go decision** that gates the roadmap's Wikipedia and `collections` entries.

Everything about how Wikipedia would *fit* noggin — roots, panes, keys, search, starring — is
deliberately unanswered here. If the answer to the question above is "no", none of it matters.

**Amendment (2026-09-02):** reading the sample articles surfaced a "swiss-cheese" effect —
missing maths and images leave visible holes that make otherwise-readable articles feel
broken, precisely the risk A5 anticipates as a known limitation rather than an automatic
no-go. Rather than accept that limitation as final, the cycle now also runs two small, costed
experiments — a Unicode approximation for maths, and a caption-derived placeholder for
images — before the verdict is recorded, so the verdict and findings both reflect the
higher-resolution result. This is new capability being added to the contract, not a
relocated one — see D104.

## Deliverables

- **S1 — A throwaway conversion spike.** A command-line entry point, separate from the noggin
  TUI and not reachable from it, that takes a Wikipedia article title and writes a markdown
  file. It fetches the article from Wikimedia's public REST API (no auth), strips the parts
  that are not the article, and converts what remains to markdown. Clearly marked
  experimental; its presence must not change noggin's behaviour in any way.

- **S2 — Exact-title fetch only.** The spike takes a title, not a search query. It must
  handle titles containing punctuation, parentheses and spaces. No search, no
  autocompletion, no link resolution, no redirect or disambiguation handling beyond whatever
  the API does for free.

- ~~**S3**~~ — *Superseded by S3a (2026-09-02): images no longer removed without trace.*

- **S3a — The article, and nothing but the article — images excepted.** Chrome that is not
  article content is removed before conversion: navigation boxes, infoboxes, edit-section
  links, citation superscripts, reference lists, metadata tables. Content tables are kept.
  Images are removed from their position but leave a plain-text placeholder in their place,
  carrying the best available description — the figure's caption, falling back to its
  alt-text, falling back to its filename, in that priority — so a reader can tell an image
  existed and get a hint of its relevance. No image is fetched, decoded or rendered visually;
  the placeholder is text only, and any link the image carried is stripped along with every
  other link (S4) — not left clickable. Removal (of chrome and of the image itself) must not
  leave visible holes — no stray captions divorced from their image, empty rows, orphaned
  brackets or runs of blank lines where something was cut.

- **S4 — Links as plain text.** Link markup is stripped entirely; the link's text survives as
  ordinary prose. No URLs, no brackets, no styling. Following links is out of scope.

- **S5 — The judging sample.** The spike is exercised against exactly these four articles,
  chosen to span the failure modes:
  - `Why is there anything at all?` — plain prose, punctuation-heavy title
  - `E (mathematical constant)` — maths
  - `List of programming languages by type` — many subsections and lists
  - `Major Arcana` — tables

  Their converted markdown is read **in noggin's own Document Viewer**, at real pane widths,
  by placing the files in a scratch root. Judgement is by eye, by the user.

- ~~**S6**~~ — *Superseded by S6a (2026-09-02): image handling and a bounded form of maths
  approximation are now explicitly permitted, not excluded.*

- **S6a — A retune budget, widened for two costed experiments.** A "close — can it be a bit
  better?" verdict is an expected outcome, not a failure. Adjusting what gets stripped, and
  re-judging, is in scope and may be repeated. It remains bounded by the cycle: no new
  capability may be added to buy quality beyond the two named exceptions below, and no link
  resolution or per-article special-casing.

  Two named exceptions, both **exploratory and costed, not committed** — nothing decided here
  fixes the shape of the eventual real feature; a later cycle with its own agreed spec decides
  whether and how either ships:

  - **Maths.** Preferred approach: a Unicode textual approximation of formulae (fractions,
    sub/superscripts, roots, sums/integrals, Greek letters and similar), via a textual
    conversion step — this counts as "text", not "rendering", for this spec's purposes. A new
    third-party dependency is permitted for this specifically. If it is not a small-ish lift,
    fall back to preserving the formula's raw source text (already permitted, unchanged) and
    record why in findings (S7a) — do not chase the Unicode path past that point.
  - **Images.** Covered by S3a: a caption/alt/filename-derived placeholder in place of the
    stripped image, no clickable link. Try the near-zero-cost path first (stop deleting the
    image node; let the existing pipeline carry the placeholder through) before writing new
    extraction logic.

  Both experiments use the same cheapest-path-first order, and the same builder's-judgement
  stop rule as the rest of this budget: "low-effort" is the builder's call; the moment either
  experiment stops being low-effort, stop and record it in findings (S7a) rather than push on.
  Full visual/graphical rendering of maths or images remains excluded (see Out of scope).

- ~~**S7**~~ — *Superseded by S7a (2026-09-02): findings must also cover S6a's two experiments.*

- **S7a — A recorded verdict, plus two costed experiments.** The cycle ends with an explicit
  **go / no-go / retune-exhausted** decision logged in `DECISIONS.md`, plus a findings record
  of what converts well, what converts badly, and what could not be fixed within the retune
  budget — enough that a later cycle can act on it without re-running the experiment. The
  findings record additionally covers, for each of S6a's two experiments: what was tried, in
  what order; the effort actually spent; any dependency added; the resulting quality by eye;
  and a recommendation on whether it is worth carrying into a future real Wikipedia feature.

- **S8 — Polite by default.** The spike identifies itself to Wikimedia with a descriptive
  User-Agent, per their published policy.

## Out of scope

- Any Wikipedia capability inside the noggin TUI: no root, no pane, no key, no search, no
  reading Wikipedia from the app. *(Roadmap: "Wikipedia as a real collection inside noggin",
  gated on this cycle's verdict.)*
- Search of any kind, including title completion. *(Same roadmap entry.)*
- Starring articles; anything that spans collections. *(Roadmap: `C6 collections`.)*
- Following links, wiki-links or URLs. *(Roadmap Unscheduled, existing entry.)*
- ~~Rendering images. Not deferred — not wanted.~~ *Superseded by S3a/S6a (2026-09-02):
  image placeholder text is now in scope as a costed experiment. Visual/graphical image
  rendering remains out — no image is ever fetched, decoded or displayed.*
- ~~Rendering maths. Formulae Wikipedia serves as images will not survive. A low-effort
  textual fallback is allowed under S6; building maths rendering is not.~~ *Superseded by
  S6a (2026-09-02): Unicode textual approximation is now the preferred approach, as a
  costed experiment. Full visual/graphical maths rendering remains out.*
- A live/clickable link on the image placeholder (or any other stripped link). Stripped, not
  rendered, this cycle — surfacing it is future work, alongside link-following generally.
  *(Same roadmap entry as "Following links, wiki-links or URLs".)*
- Caching, offline reading, refresh semantics, rate limiting beyond politeness.
- Error and edge-case handling: no network, missing article, disambiguation page, redirects.
- Attribution / licence text on rendered articles. Nothing is distributed this cycle.
  *(Open question OQ2 — must be settled before any released build carries this feature.)*
- Any decision about the roots/collections model, typed roots, or a tab/channel navigation
  layer. Deliberately left open. *(Roadmap: `C6 collections`.)*
- Non-English Wikipedia, and any other Wikimedia project.

## Acceptance (cycle-level)

The cycle is done when the four sample articles have been converted, read in noggin, and
judged against the criteria below, and the verdict is recorded.

The user's readability bar, in priority order:

- **A1** — Text is readable: wrapped to the pane, scrollable, nothing overlapping or
  truncated.
- **A2** — Headings and sections are legible and visually parsable — the article's structure
  is apparent at a glance.
- **A3** — No large gaps or artifacts where images, formulae, footnotes or stripped chrome
  used to be.
- **A4** — Tables are readable. Ugly is acceptable; unreadable is a defect. *(Nice to have —
  the cycle can pass without it.)*

- **A5 — The verdict is made across the sample, not per article.** An article that reads
  badly *only* because of content noggin deliberately does not render — maths above all — is
  recorded as a **known limitation** against that article, and does not by itself force a
  no-go. It forces a no-go only if the same failure would plainly recur across ordinary
  articles. Concretely: if `E (mathematical constant)` reads as swiss cheese while the other
  three clear A1–A3, that is a *go with a known limitation*, not a no-go.

Explicitly **not** deal-breakers: non-rendering maths, non-rendering images, missing
footnotes, ugly tables.

The overall bar, in the user's words: **"I would read this here instead of opening a
browser."**

A **no-go** verdict is a legitimate, successful outcome of this cycle — the experiment
answered its question. It is not a failed cycle.

## Open questions

- **OQ1** — Does the spike code survive the cycle, or is it deleted at close? The conversion
  logic is the model-agnostic keeper if the verdict is *go*; the command-line wrapper is not.
  Proposal: keep the conversion, drop the wrapper, decide at `/charm-close`.
- **OQ2** — CC BY-SA attribution: local personal use appears not to trigger redistribution
  obligations, but this was not confirmed against a citable source. Must be settled before
  noggin is ever distributed with a Wikipedia feature. Not blocking this cycle.
- ~~**OQ3**~~ — *Resolved 2026-08-14:* reading is the risk, arrival is known-solvable, so the
  scratch-root proxy measures the right thing. The spike stands as specified.
- **OQ4** — No hard retune bound (round count) was set; the cycle boundary and the
  builder's "stop when it stops being low-effort" judgement are the only limits. Accepted
  deliberately. If retuning starts to sprawl, that is a `charm-replan` trigger.
