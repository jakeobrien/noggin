# Wikipedia: endpoints and conversion

The Wikipedia Source lives in `internal/wikipedia`. `cmd/wikispike` is a thin CLI over the same package (`wikispike <title> <out>`) for converting articles to files and judging conversion quality in noggin itself. Use it when changing the pipeline. Live tests need `en.wikipedia.org` reachable; in the sandbox it is allow-listed (`docs/agents/sbx_policy_decisions.md`).

## Endpoints

Both are plain GET with no auth, so tests fake them with `httptest`; no test should reach the network outside `live_test.go`.

**Title search**: `GET https://en.wikipedia.org/w/rest.php/v1/search/title?q=<query>&limit=<n>`

- Prefix/title matching, not full text. Surface results in Wikipedia's order; don't re-sort.
- Response: `{"pages":[{"id","key","title","excerpt","matched_title","description","thumbnail"}]}`. `key` is URL form (`Leonhard_Euler`), `title` display form; `description` may be `null`.

**Article HTML**: `GET https://en.wikipedia.org/api/rest_v1/page/html/<title>`

- Returns Parsoid HTML, which isn't the markup of a rendered article page.
- A redirect title answers **HTTP 307** to `.../w/rest.php/v1/page/<Canonical_Title>/html?redirect=no`. Go's client follows it, so `resp.Request.URL` carries the canonical key, giving the canonical URL (`https://en.wikipedia.org/wiki/<key>`) and title. The HTML's `<title>` and `<link rel="dc:isVersionOf">` are a second source if ever needed.

Wikimedia's User-Agent policy requires a descriptive UA with contact info (`userAgent()`). Keep request rates modest: the Wikipedia Picker's longer debounce is the throttle; don't stack a second one.

## Conversion pipeline

Order matters: maths and images are *replaced* by placeholder stages before chrome stripping deletes whatever is left, then links are unwrapped, then markdown cleanup runs. The selector list and cleanup passes are documented where they live (`strip.go`, `htmldoc.go`, `cleanup.go`).

Parsoid quirks worth knowing:

- **Maintenance templates** (`{{citation needed}}` and kin) carry no `typeof` on their visible `<sup>`; match them by the `Inline-Template` class. `{{rp|…}}` page locators do carry bare `typeof="mw:Transclusion"`. Neither breaks a bracket-balance check while leaking; only reading the output catches them.
- **Maths**: each formula's raw TeX is in `data-mw` (`body.extsrc`) on the `mw:Extension/math` span. Display formulas are marked by `attrs.display == "block"` and the `mwe-math-element-block` class.
- **Images**: caption (`figcaption`), alt text and filename (`img[resource]`) are all present. Strip citation superscripts from a caption *before* reading it, or ref digits leak into the placeholder.

## Maths: termtex and its limit

`github.com/doug/termtex` renders TeX to Unicode text, but fractions, big operators and multi-glyph superscripts come out as multi-row box drawing. Parsoid puts most formulas mid-sentence, so the rule is: inline formulas keep termtex's output only if it stays on one line, otherwise fall back to raw TeX; display formulas get a fenced block and keep the full multi-line rendering. Cleanup is fence-aware so it never mangles that ASCII art. On `E (mathematical constant)`, about 61% of formulas got Unicode and 39% fell back to raw TeX. That's an accepted limitation.

## Known limitations

- The selectors were tuned against four articles (`Why is there anything at all?`, `E (mathematical constant)`, `List of programming languages by type`, `Major Arcana`). Expect new articles to surface new chrome; treat the list as one to extend.
- Image placeholders make chrome gaps loud: a navigational icon that used to vanish now shows as `[Image: …]`. That's a signal to extend the chrome selectors, not a placeholder bug.
- Bracket repair drops unmatched closers but never openers, so a half-open interval like `[0, 1)` loses its `)`. Depth is tracked per prose chunk, not per paragraph, so an earlier stray opener can mask a later stray closer.
- The inline sister-project-link matcher scans the whole document, not just External links.
