# M3.5 — the image placeholder experiment (S6a / D104)

Written at C4a/M3.5. What was tried, in what order, cost, and the resulting quality.

## What was tried, in order

1. **Near-zero-cost path, first (as SPEC requires).** Parsoid HTML already carries an
   image's caption (`<figcaption>` inside `<figure typeof="mw:File/Thumb">`), alt-text
   (`img[alt]`), and filename (`img[resource]`, shaped `./File:Some_name.ext`) for free —
   verified against live fetches of all four S5 titles
   (`cmd/wikispike/fetch.go`'s `restBase`). No OCR, no extra request, matching M3.4's
   established shape exactly: a new pipeline stage, `renderImages`
   (`cmd/wikispike/images.go`), that runs before `stripChrome` and replaces each
   `[typeof^="mw:File"]` node with a plain-text placeholder — `stripChrome`'s own image
   selector (`strip.go`) is now the fallback for whatever `renderImages` couldn't recover a
   description for, same relationship `renderMath`/`stripChrome` already had for maths.
2. **Priority order (SPEC S3a), implemented as straight-line fallthrough:** caption text
   (flattened via goquery's `.Text()`, so a caption's own nested wikilinks survive as plain
   words, not markup) → `alt` attribute → filename with `File:` stripped, extension
   stripped, underscores turned back to spaces. No new capability beyond calling three
   attribute lookups already present on the parsed node.
3. **Placeholder shape:** `[Image: <description>]` — square brackets so a reader can tell
   it's a placeholder rather than body prose (SPEC's "hint of its relevance" bar), no link,
   no `![...]` markdown image syntax. The commonmark converter escapes the leading `[` in
   its output (`\[Image: ...]`), which is ordinary literal-bracket escaping, not residual
   link/image markup — confirmed no `![`, no `](`, and a balanced bracket count across all
   four regenerated samples (`why_anything.md`, `e_math.md`, `list_languages.md`,
   `major_arcana.md`, repo root, off-sandbox reference copies regenerated after M3.5).

## Effort and dependency

- No new third-party dependency (unlike M3.4's maths experiment) — everything needed
  (caption/alt/filename) was already parsed goquery attributes and text. ~100 lines of new
  code (`cmd/wikispike/images.go` + `images_test.go`), one pipeline stage added
  (`articleMarkdownFromHTML` now runs `renderImages` after `renderMath`, before
  `stripChrome`, `cmd/wikispike/main.go`), one existing comment corrected (`strip.go`'s
  image-removal note). Genuinely low-effort; no abandonment needed.

## Resulting quality, by eye

Live-fetched all four S5 titles before/after: every `<figure typeof="mw:File/Thumb">` that
used to vanish without a trace now shows its caption as a placeholder exactly where the
figure sat — e.g. `Major Arcana`: `[Image: The Major Arcana in the Rider–Waite Tarot deck]`;
`Why is there anything at all?`: four figures, each captioned (Parmenides, the Monad, Stephen
Law, Bertrand Russell), all recovered via the caption path — none of this article's figures
needed the alt/filename fallback. `E (mathematical constant)`'s figures likewise all resolved
via caption. No sample's genuine content images needed the filename fallback in practice
(figures reliably carry a caption); the fallback chain is exercised and correct
(`images_test.go`'s unit tests), just not needed live for these four titles' figures
specifically.

Mechanical re-check across the four regenerated samples: no `![`/`](` (no image markup
survives), balanced `[`/`]` (no dangling bracket), no 3+ blank-line run, no empty list item —
all clean (the one pre-existing paren imbalance in `list_languages.md`, unrelated to images,
persists unchanged from M3.4's findings).

## Discrepancy discovered mid-experiment (routed to TRIAGE, not fixed here)

Regenerating the samples surfaced a **pre-existing chrome-stripping gap**, unrelated to this
experiment's own scope: `stripChrome` (M2.2) does not match Wikipedia's "portal bar"
(`div.portal-bar`, e.g. `why_anything.md`'s trailing `Philosophy`/`Physics`/`Astronomy`/
`Religion`/`Science` list) or "sister project" side-boxes (`div.sistersitebox`, e.g.
`e_math.md`/`major_arcana.md`'s "Wikimedia Commons"/"Wikiquote" lines) — both navigational
chrome, not article content. Before this experiment, their icon images were silently deleted
by `stripChrome`'s catch-all image selector, so the leaked list items read as bare text
("Philosophy", "Physics", …) with no visible marker; now that every image gets a placeholder
first, the leak is visible as `[Image: icon]`/`[Image: Wikimedia Commons logo]` next to each
item, but the underlying leak (the whole box surviving stripping) predates this experiment
and is not new. Not fixed as part of M3.5 (out of scope: chrome selection is M2.2's/M3.3's
job, not this experiment's) — logged to `.charm/cycle/TRIAGE.md` for triage.

## Recommendation for a future real Wikipedia feature

**Worth carrying, as-is.** The caption-then-alt-then-filename rule is cheap, correct, and
measurably improves A3 for figure-bearing articles with no observed regression (`make check`,
full `go test ./...`, mechanical re-check of all four S5 samples). A real feature would want
to pair this with fixing the chrome-selector gap above (portal bars, sister-project boxes)
so the placeholder only ever marks a genuine content image, never a leaked navigational icon
— this spike deliberately left that fix for a later cycle (TRIAGE, not a "new capability"
this experiment's budget covers).
