# Wikimedia endpoints for the Wikipedia Source

Cycle: C6. Verified live 2026-09-07 from this sandbox (Wikimedia domains are allow-listed —
see `knowledge/sandbox-network-allowlist.md`). The C4a spike (`cmd/wikispike/fetch.go:39-59`)
only fetches by exact title; search and redirect resolution are new to C6.

## Title search (S2)

`GET https://en.wikipedia.org/w/rest.php/v1/search/title?q=<query>&limit=<n>`

- Prefix/title matching, not full-text — exactly S2's "article titles" semantics.
- Response `{"pages":[{ "id", "key", "title", "excerpt", "matched_title", "description",
  "thumbnail" }]}`; `description` is the short description S2 wants and can be `null`.
  `key` is the URL-form title (`Leonhard_Euler`); `title` is display form.
- Order is Wikipedia's ranking — surface as-is (S2: "in Wikipedia's order").
- Sample: `q=eul&limit=3` → `Leonhard Euler` / "Swiss mathematician (1707–1783)",
  `Euler's constant` / "Difference between logarithm and harmonic series", …

## Article HTML and canonical resolution (S3)

- The spike's endpoint `GET https://en.wikipedia.org/api/rest_v1/page/html/<title>` answers a
  redirect title with **HTTP 307** to
  `https://en.wikipedia.org/w/rest.php/v1/page/<Canonical_Title>/html?redirect=no`
  (observed: `Euler` → `Leonhard_Euler`). Go's default `http.Client` follows it; the final
  request URL (`resp.Request.URL`) therefore carries the canonical key, which yields both the
  canonical URL (`https://en.wikipedia.org/wiki/<key>`) and, after `_`→space and unescaping,
  the canonical title. The Parsoid HTML also carries `<title>` and
  `<link rel="dc:isVersionOf">` if a second source is wanted.
- English Wikipedia only (spec). Wikimedia's User-Agent policy requires a descriptive UA with
  contact info; the spike's `userAgent()` (`cmd/wikispike/fetch.go:20-27`) already complies and
  can be reused with a `noggin/…` product token.

## Notes for the builder

- Both endpoints are plain GET + JSON/HTML with no auth — trivially fakeable with `httptest`;
  no test should reach the network.
- Wikimedia asks clients to keep request rates modest; S13's longer Wikipedia debounce is the
  spec's answer — do not add a second throttle.
