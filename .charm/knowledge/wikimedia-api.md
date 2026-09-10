# Wikimedia endpoints: title search and canonical-URL resolution

As of: C6

Promoted from `cycle/findings/wikimedia-api.md` (written at C6/M2.2-M2.3, verified live
2026-09-07). Cited by `internal/wikipedia/search.go` and `internal/wikipedia/fetch.go`.
Wikimedia domains need explicit sandbox allow-listing — see
`knowledge/sandbox-network-allowlist.md`. The C4a spike (`cmd/wikispike/fetch.go:39-59`) only
ever fetched by exact title; search and redirect resolution were new to C6.

## Title search

`GET https://en.wikipedia.org/w/rest.php/v1/search/title?q=<query>&limit=<n>`

- Prefix/title matching, not full-text.
- Response `{"pages":[{ "id", "key", "title", "excerpt", "matched_title", "description",
  "thumbnail" }]}`; `description` is a short description and can be `null`. `key` is the
  URL-form title (`Leonhard_Euler`); `title` is display form.
- Order is Wikipedia's own ranking — surface as-is, don't re-sort.
- Sample: `q=eul&limit=3` → `Leonhard Euler` / "Swiss mathematician (1707–1783)",
  `Euler's constant` / "Difference between logarithm and harmonic series", …

## Article HTML and canonical resolution

- `GET https://en.wikipedia.org/api/rest_v1/page/html/<title>` answers a redirect title with
  **HTTP 307** to `https://en.wikipedia.org/w/rest.php/v1/page/<Canonical_Title>/html?redirect=no`
  (observed: `Euler` → `Leonhard_Euler`). Go's default `http.Client` follows it; the final
  request URL (`resp.Request.URL`) therefore carries the canonical key, which yields both the
  canonical URL (`https://en.wikipedia.org/wiki/<key>`) and, after `_`→space and unescaping,
  the canonical title. The Parsoid HTML also carries `<title>` and
  `<link rel="dc:isVersionOf">` if a second source for the same fact is ever wanted.
- English Wikipedia only. Wikimedia's User-Agent policy requires a descriptive UA with contact
  info; the spike's `userAgent()` (`cmd/wikispike/fetch.go:20-27`) already complies and was
  reused with a `noggin/…` product token.

## Notes for a future caller (e.g. a dictionary Query Source against a similar REST API)

- Both endpoints are plain GET + JSON/HTML with no auth — trivially fakeable with `httptest`;
  no test should ever reach the network for them.
- Wikimedia asks clients to keep request rates modest; a longer per-mode search debounce is a
  reasonable answer — avoid adding a second, separate throttle on top of it.
