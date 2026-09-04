# Converting Wikipedia's maths to text: termtex, Parsoid's display signal, and their limits

As of: C4a

Promoted from `cycle/findings/maths-experiment.md` (written at C4a/M3.4, addended post-M4.1).
Facts and citations from the wikipedia-conversion-spike (`cmd/wikispike/`, kept as reference
per D97/D112 — not integrated, no product feature exists yet).

## What was tried, in order

1. **Unicode textual approximation, first.** Wikimedia's Parsoid HTML carries each formula's
   raw TeX source for free, in `data-mw='{"body":{"extsrc":"..."}}'` on the
   `typeof="mw:Extension/math"` span (`cmd/wikispike/fetch.go`'s `restBase`) — no OCR, no
   extra request. Fed to [`github.com/doug/termtex`](https://github.com/doug/termtex)
   (Apache-2.0), which renders LaTeX-flavoured math to Unicode text: `\frac`,
   sub/superscripts, `\sqrt`, `\sum`/`\int`/`\lim`, Greek letters.
2. **The catch, found empirically:** termtex renders fractions, multi-glyph superscripts
   without a dedicated Unicode codepoint, and big-operator forms as a multi-row,
   box-drawing ASCII-art block. Parsoid puts most formula spans **mid-sentence** — a
   multi-line result inserted there tears the sentence apart rather than repairing it.
3. **Per-formula rule adopted (`cmd/wikispike/math.go`'s `renderMath`):** try termtex; keep
   the Unicode result only if it stays on one line; otherwise fall back to the formula's raw
   TeX source text so the sentence still reads.
4. **Addendum, found later (post-M4.1 fix-up):** Parsoid *does* mark a formula
   display/block-style with `data-mw.attrs.display == "block"` and a
   `mwe-math-element-block` class (vs. `mwe-math-element-inline`) — Wikipedia's own CSS gives
   that formula its own visual line even though the Parsoid span stays nested inside the same
   running `<p>`. `renderMath` branches on this signal: display/block formulas are wrapped in
   a `<pre>` element (invalid nesting inside the source `<p>`, but exactly what makes the
   html-to-markdown converter's own HTML5 parser auto-split the paragraph around it on its
   next parse — the same mechanism a browser uses) and keep termtex's full, possibly
   multi-line rendering, since there is no sentence left to tear apart. A companion fix was
   needed in the cleanup pass (`cleanup.go`'s `cleanupMarkdown`): its whole-string
   bracket/blank-run/whitespace passes would corrupt fenced ASCII-art unless fenced regions
   are carved out first.

## Resulting quality (live `E (mathematical constant)`, 104 formula occurrences)

**61% got the Unicode approximation** (single-line results: `\gamma` → `γ`, `x_0` → `x₀`).
**39% fell back to raw TeX text** (anything with `\frac`, `\sum`/`\lim` combined with a
fraction, `\binom` — unsupported by termtex and garbled if not caught — or a multi-glyph
superscript with no Unicode codepoint). No mid-sentence gap remains anywhere in the sample;
raw-TeX fallback is legible as "formula-shaped" text but not pretty.

## Recommendation for a real Wikipedia feature

Worth carrying, in modified form. The per-formula single-line-or-raw-text rule is cheap and
regression-free. A real feature should also give termtex's multi-line output its own fenced
code block for display/block formulas from the start (this spike's addendum bolted that on
after the fact) rather than defaulting every display formula through the inline rule first.
