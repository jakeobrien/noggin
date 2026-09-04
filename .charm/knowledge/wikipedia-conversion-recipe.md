# Wikipedia article → readable markdown: the conversion recipe

As of: C4a

Promoted from `cycle/findings/m4.1-conversion-record.md` (C4a's consolidated verdict record).
The reference for anyone converting a Wikipedia (Parsoid REST HTML) article to markdown
worth reading — what works, what doesn't, and the exact selector/cleanup list a spike arrived
at after several retune rounds. The code lives at `cmd/wikispike/`, kept as reference per
D97/D112 (a decision, not a shipped feature — see DECISIONS.md D112 for the go/no-go verdict
this record fed).

## What converts well

- **Prose, headings and structure.** Heading hierarchy survives (up to four levels seen: `##`
  – `####`), paragraphs and block quotes render normally, previously-linked phrases read as
  ordinary prose with no markdown link/URL syntax left.
- **Tables.** Content tables (`table.wikitable`) survive with correct column alignment;
  infobox/navbox/sidebar tables are gone.
- **Chrome removal generally** — see the selector list below.
- **Citation-like superscripts**, including two shapes discovered late: Wikipedia's
  maintenance-template family (`{{citation needed}}`, `{{clarification needed}}`,
  `{{dubious}}`, `{{disputed}}`) carries **no** `typeof` on its own visible `<sup>` — the
  `typeof="mw:Transclusion"` marker instead lands on a preceding, invisible
  `<link rel="mw:PageProp/Category">` sharing the same `about` id, so it has to be matched by
  its `Inline-Template` CSS class, not a typeof selector like every other chrome category
  here. The `{{rp|<page>}}` page-locator template's own `<sup>` *does* carry
  `typeof="mw:Transclusion"` directly (a real citation ref's typeof is always the compound
  `"mw:Transclusion mw:Extension/ref"`, so an exact-match selector on the bare value never
  collides with it). Neither shape fails a mechanical bracket-balance check while broken —
  only an eyeball read catches them.
- **Maths and images** — see `wikipedia-math-conversion.md` and
  `wikipedia-image-placeholders.md` for full detail.

## What converts badly, or not at all (known limitations)

- Formula raw-TeX fallback (accepted, not a defect) — see `wikipedia-math-conversion.md`.
- A half-open-interval idiom like `[0, 1)` loses its `)` to the bracket-repair pass's
  asymmetric closer-only design (closers are dropped when unmatched; openers never are, to
  avoid deleting more real text on the shape actually observed — a stray opener is a visible,
  honest artifact rather than a deleted character).
- The bracket-repair pass tracks depth across a whole flushed prose chunk, not per paragraph,
  so an earlier unmatched opener can mask a later genuinely-dangling closer. Not observed in
  any live sample, only reproduced synthetically — the mechanical balanced-count check can't
  detect it since a masked pair still nets out balanced.
- The inline-sister-project-link matcher's scope is the whole document, not just the
  External-links section its own name suggests — accepted as a broadening, not a bug: a
  full-line anchor rarely false-matches ordinary prose, and scoping it would need its own
  heading-boundary detection for no clear benefit.

## The selector/cleanup list as it ended up

Chrome removal (`cmd/wikispike/strip.go`'s `stripChromeInPlace`, in the order it runs):

1. `div.navbox` — navboxes.
2. `table.infobox, .infobox` — infoboxes.
3. `table.sidebar` — a second, distinct floated summary-box template.
4. `div.portal-bar, div.sistersitebox` — portal bars and boxed sister-project templates.
5. A text-pattern match over every `<li>`/`<p>`'s whole rendered text (via a helper that
   excludes inlined `<style>`/`<script>` first) for the sister-project template family that
   renders with no wrapping box element (e.g. inline "Media related to X at Wikimedia
   Commons" list items).
6. `.mw-editsection` — edit-section links (defensive; never present in Parsoid REST HTML).
7. Citation-like superscript selector — unions `sup[typeof*="mw:Extension/ref"]` (real
   citation refs), `sup.Inline-Template` (the maintenance-template family), and
   `sup[typeof="mw:Transclusion"]` (the `{{rp}}` page-locator template).
8. `.mw-references-wrap, [typeof*="mw:Extension/references"]`, removing the closest
   `<section>` ancestor if present — whole reference-list sections.
9. `table.metadata, table.ambox` — metadata/maintenance tables (content `table.wikitable` is
   never matched).
10. `[typeof^="mw:File"]`, `[typeof="mw:Extension/math"]`, then a final bare `img` sweep —
    whatever the image/math placeholder stages (below) didn't already replace.

Link stripping: every `<a>` unwrapped to its inner content, run *after* chrome removal so
only prose/table-cell links remain to unwrap.

Cleanup, in order, fence-aware throughout (so ASCII-art math is never touched by a
whole-string pass):

1. Per non-fenced chunk: empty/whitespace-only `()`/`[]` removal, then the asymmetric
   bracket-depth repair pass for each of `()`/`[]` independently, then per-line
   whitespace/space-before-punctuation collapse, empty list item removal, 3+ blank-line
   collapse.
2. Drop any heading with no non-blank content before the next same-or-higher-level heading or
   EOF, run once over the whole reassembled document (heading structure is document-wide, not
   chunk-local) and repeated to a fixed point — chrome removal can otherwise leave a dangling,
   content-free heading through EOF.
3. A second fence-aware blank-run collapse, since dropping a heading can create a fresh
   blank-run at its former position.

Two pre-strip substitution stages run before `stripChrome`, so their target markup is
*replaced* rather than deleted: the maths renderer and the image-placeholder renderer (both
detailed in their own knowledge files).

## Recommendation for a real Wikipedia feature

The selector list above is a solid starting point but was tuned against exactly four
articles (`Why is there anything at all?`, `E (mathematical constant)`,
`List of programming languages by type`, `Major Arcana`) — expect other articles to surface
new chrome-selector gaps the same way this cycle found portal bars, sister-project boxes and
the maintenance-template family one at a time. Treat this list as a starting selector set to
extend, not a closed one.
