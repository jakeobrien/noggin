# M3.4 — the maths Unicode experiment (S6a / D104)

Written at C4a/M3.4. What was tried, in what order, cost, and the resulting quality.

## What was tried, in order

1. **Unicode textual approximation, first (as SPEC requires).** Wikimedia's Parsoid HTML
   already carries each formula's raw TeX source for free, in
   `data-mw='{"body":{"extsrc":"..."}}'` on the `typeof="mw:Extension/math"` span (verified
   against a live fetch of `E (mathematical constant)`, `cmd/wikispike/fetch.go`'s
   `restBase`) — no OCR, no extra request. Fed that source to
   [`github.com/doug/termtex`](https://github.com/doug/termtex) (Apache-2.0, real published
   module, `go get` verified), which renders LaTeX-flavoured math to Unicode text, including
   `\frac`, sub/superscripts, `\sqrt`, `\sum`/`\int`/`\lim`, and Greek letters — exactly
   SPEC's named list.
2. **The catch, found empirically, not assumed:** termtex renders fractions, multi-glyph
   superscripts without a dedicated Unicode codepoint, and big-operator forms as a multi-row,
   box-drawing ASCII-art block (its README's own gallery shows this — e.g. the quadratic
   formula renders as three stacked lines). That's fine for a formula that owns its own line,
   but Parsoid puts these spans **mid-sentence** — checked directly against a live fetch: none
   of `E (mathematical constant)`'s 31 "display: block"-styled formulas sit inside their own
   `<dl>`/`<dd>` line; they're inline spans in running prose exactly like the "inline"-styled
   ones. A multi-line result inserted there would tear the sentence apart, not repair it.
3. **The per-formula rule adopted (`cmd/wikispike/math.go`'s `renderMath`):** try termtex;
   keep the Unicode result only if it stays on one line; otherwise fall back to the raw TeX
   source text (already permitted by SPEC, unchanged) so the sentence still reads. This
   applies the "small-ish lift" test in D104/S6a at formula granularity rather than
   all-or-nothing for the whole experiment — no new capability beyond calling the library and
   checking line count.

## Effort and dependency

- One new third-party dependency: `github.com/doug/termtex` (permitted specifically for this
  experiment, D104). ~110 lines of new code (`cmd/wikispike/math.go` + `math_test.go`), one
  pipeline stage added (`articleMarkdownFromHTML` now runs `renderMath` before `stripChrome`,
  `cmd/wikispike/main.go`), one existing comment corrected (`strip.go`'s math-removal note).
  No new extraction logic against Wikipedia beyond reading `data-mw` — that path already
  existed in Parsoid's response. Genuinely small-ish; no abandonment needed.

## Resulting quality, by eye

Live-fetched `E (mathematical constant)` before/after (`/tmp/e_math_new.md` vs. the
pre-M3.4 sample in the repo root): every previously-orphaned sentence — "denoted." →
"denoted γ.", "One has where is..." → "One has e = exp(1), where exp is...", "the limit an
expression" → "the limit \lim_{n\to \infty}\left(1+\frac 1n\right)^n, an expression" — now
reads as a complete sentence. No mid-sentence gap remains; SPEC's mechanical bars (no `()`
survives, no 3+ blank-line run, balanced brackets) hold across all four S5 samples,
re-checked after this change.

Tally across the live article's 104 formula occurrences (`data-mw` present for all 104):
**63 (61%) got the Unicode approximation** — single-line results such as `\gamma` → `γ`,
`\tau` → `τ`, `x_0` → `x₀`, `e=\exp(1),` → `e = exp (1),`; **41 (39%) fell back to raw TeX
text** — anything with `\frac`, `\sum`/`\lim` combined with a fraction, `\binom` (unsupported
by termtex and garbled if not caught), or a multi-glyph superscript without its own Unicode
codepoint (e.g. `e^{i\pi}+1=0` renders as two stacked lines in termtex, so it falls back to
its raw form rather than break the sentence). The raw-TeX fallback text is legible as
"formula-shaped" text (e.g. `\frac{1}{2}`) but not pretty; that's within SPEC's "ugly is
acceptable, unreadable is a defect" bar for A3 purposes (the bar there is "no gap", not "no
ugliness").

## Recommendation for a future real Wikipedia feature

**Worth carrying, in modified form.** The per-formula single-line-or-raw-text rule is cheap,
correct, and measurably improves A3 with no observed regression (verified: `make check`,
full `go test ./...`, and a mechanical re-check of all four S5 samples for blank-line runs,
empty list items, and bracket balance — all clean, one pre-existing unrelated paren
imbalance in `list_languages.md`, unchanged from before this experiment). A real feature
would want to go further than this spike attempted — e.g. reserving termtex's multi-line
output for formulas Parsoid *does* place on their own line (a `<dl>/<dd>` block, which this
article's samples never used, but other articles may), rendered as a fenced code block —
which this cycle deliberately did not build (out of scope for a "small-ish lift" spike; the
uniform single-line-or-fallback rule was chosen specifically to avoid that added complexity).
