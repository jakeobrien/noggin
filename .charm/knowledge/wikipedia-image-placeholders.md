# Converting Wikipedia's images to text placeholders: Parsoid's free metadata

As of: C4a

Promoted from `cycle/findings/images-experiment.md` (written at C4a/M3.5). Facts and
citations from the wikipedia-conversion-spike (`cmd/wikispike/`, kept as reference per
D97/D112 — not integrated, no product feature exists yet).

## What was tried, in order

1. **Near-zero-cost path, first.** Parsoid HTML already carries an image's caption
   (`<figcaption>` inside `<figure typeof="mw:File/Thumb">`), alt-text (`img[alt]`), and
   filename (`img[resource]`, shaped `./File:Some_name.ext`) for free — no OCR, no extra
   request. A new pipeline stage, `renderImages` (`cmd/wikispike/images.go`), runs before
   `stripChrome` and replaces each `[typeof^="mw:File"]` node with a plain-text placeholder;
   `stripChrome`'s own image selector is the fallback for whatever `renderImages` couldn't
   recover a description for.
2. **Priority order, straight-line fallthrough:** caption text (flattened via goquery's
   `.Text()`, so a caption's own nested wikilinks survive as plain words) → `alt` attribute →
   filename with `File:` stripped, extension stripped, underscores turned back to spaces.
3. **Placeholder shape:** `` `[Image: <description>]` `` — square brackets so a reader can
   tell it's a placeholder, wrapped in an actual `<code>` element so the html-to-markdown
   converter handles its own bracket-escaping consistently. No link, no `![...]` markdown
   image syntax.
4. **A caption-leak gotcha, found live (folded in as M3.5a):** `renderImages` must strip
   citation-superscript markup (`sup[typeof*="mw:Extension/ref"]` and its wider family, see
   `wikipedia-conversion-recipe.md`) from the caption text *before* reading it, or a
   placeholder can leak ref-marker digits (e.g. `[Image: ... Parmenides (c. 515 BC).[1][2]]`)
   — order of operations across pipeline stages matters even when each stage looks
   independent.

## Resulting quality

Live-fetched across all four S5 titles: every `<figure typeof="mw:File/Thumb">` that used to
vanish without a trace now shows its caption as a placeholder exactly where the figure sat.
No sample's genuine content images needed the alt-text/filename fallback in practice — the
fallback chain is exercised only by unit tests, not by any of these four titles' real figures
— worth knowing before assuming the fallback path is well-exercised by live data alone.

## A gotcha this experiment surfaced (not itself a defect here)

Turning every stripped image into a *visible* placeholder converts a previously-silent chrome
gap into a visible one: navigational icon images (portal bars, sister-project boxes) that
used to vanish invisibly now surface as `[Image: icon]`-style leaks once every image gets a
placeholder. Any future work reusing this pattern should expect the same effect — a
placeholder step makes chrome-selector gaps loud, which is a feature for catching them, not
a regression in the placeholder step itself.

## Recommendation for a real Wikipedia feature

Worth carrying, as-is. Cheap (no new dependency), correct, and the fallback chain is already
unit-tested even though live data didn't need it. Pair with keeping chrome-selector coverage
current (see `wikipedia-conversion-recipe.md`) so a placeholder only ever marks a genuine
content image, never a leaked navigational icon.
