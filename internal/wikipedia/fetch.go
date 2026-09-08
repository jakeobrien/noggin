package wikipedia

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// fetchBase is Wikimedia's REST API for parsed page HTML (findings/
// wikimedia-api.md), the same endpoint the C4a spike used
// (cmd/wikispike/fetch.go). A redirect title answers with an HTTP 307 to the
// canonical title's own REST URL, which Go's default http.Client follows
// automatically; the final request URL then carries the canonical title.
//
// A var, not a const (M2.9/TRIAGE T1, T22), solely so a test can point
// Fetch's own hardcoded production URL at an httptest server and back —
// every other test in this package exercises fetchAt directly with an
// explicit overridable base and never needs this at all.
var fetchBase = "https://en.wikipedia.org/api/rest_v1/page/html/"

// wikiBase is the canonical, human-facing article URL prefix S3 records —
// distinct from fetchBase, which is the REST path used only to fetch HTML.
// Also a var for the same test-only reason as fetchBase above.
var wikiBase = "https://en.wikipedia.org/wiki/"

// fetchTimeout bounds a single article fetch (S3: "bounds the request with a
// timeout").
const fetchTimeout = 15 * time.Second

// Article is a fetched, converted Wikipedia article (S3): the canonical
// title and URL taken from the final, post-redirect response, plus the
// converted markdown.
type Article struct {
	Title    string // canonical title, display form (spaces, not underscores)
	URL      string // canonical article URL, e.g. https://en.wikipedia.org/wiki/Leonhard_Euler
	Markdown string
}

// Client fetches Wikipedia articles by title, resolving redirects to the
// canonical article and converting to markdown, with a cache scoped to the
// Client's own lifetime (S3: "cached for the session only") keyed by
// canonical URL. A Client is safe for concurrent use, and its zero value
// (Client{}) is ready to Fetch — NewClient is a convenience, not a
// requirement (M2.4/TRIAGE T4): fetchAt lazily initializes httpClient and
// cache on first use via ensureInit.
type Client struct {
	httpClient *http.Client

	mu    sync.Mutex
	cache map[string]Article // keyed by canonical URL
}

// NewClient returns a Client with an empty cache, ready to Fetch. It builds
// through ensureInit (M2.8/TRIAGE T15) rather than constructing httpClient/
// cache itself, so there is exactly one place that decides what a freshly-
// initialized Client looks like — NewClient and a lazily-initialized
// zero-value Client{} can never drift apart.
func NewClient() *Client {
	c := &Client{}
	c.ensureInit()
	return c
}

// ensureInit lazily initializes a zero-value Client's httpClient and cache
// (M2.4/TRIAGE T4): without it, a Client{} constructed directly rather than
// via NewClient would panic on its first Fetch — a nil httpClient panics on
// the first field access inside (*http.Client).Do, and a nil cache map
// panics on the first write. Guarded by c.mu (M2.4/TRIAGE T5's "safe for
// concurrent use" claim covers this too) so concurrent first-use callers
// race safely to a single initialization rather than each allocating their
// own map/client and silently dropping one.
func (c *Client) ensureInit() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.httpClient == nil {
		c.httpClient = &http.Client{Timeout: fetchTimeout}
	}
	if c.cache == nil {
		c.cache = make(map[string]Article)
	}
}

// Fetch fetches title, following redirects to the canonical article,
// converts it to markdown, and returns the result (S3). The cache is keyed
// by canonical URL: since noggin's own records always hold the canonical
// title once an article has been opened (S3), a Fetch of an already-
// canonical title hits the cache without a network request unless refresh is
// true, which bypasses the cache and repopulates it (S3's Refresh). A title
// that turns out to be a redirect always costs one request — there is no way
// to know its canonical URL without asking.
//
// TestClientFetchViaNewClientDelegatesToFetchAt (fetch_test.go, M2.9/
// TRIAGE T1, T22) exercises this method and NewClient directly, against an
// httptest server via fetchBase/wikiBase's test-only override — every
// other test in this file exercises fetchAt itself, one layer further down.
func (c *Client) Fetch(title string, refresh bool) (Article, error) {
	return c.fetchAt(fetchBase, wikiBase, title, refresh)
}

// fetchAt is Fetch's implementation against overridable base URLs, so tests
// point it at an httptest server instead of the network. Its parameters are
// named fetchBaseURL/wikiBaseURL, not fetchBase/wikiBase (M2.5/TRIAGE T6),
// so they never shadow the package-level fetchBase/wikiBase values Fetch
// passes in for the production case — a shadowed name reads fine here but
// invites a copy-paste bug the moment code inside this function ever needs
// the real value instead of the parameter.
func (c *Client) fetchAt(fetchBaseURL, wikiBaseURL, title string, refresh bool) (Article, error) {
	c.ensureInit()
	candidateURL := articleURLFor(wikiBaseURL, title)

	if !refresh {
		c.mu.Lock()
		cached, ok := c.cache[candidateURL]
		c.mu.Unlock()
		if ok {
			return cached, nil
		}
	}

	html, canonicalTitle, canonicalURL, err := fetchCanonicalHTML(c.httpClient, fetchBaseURL, wikiBaseURL, title)
	if err != nil {
		return Article{}, err
	}
	markdown, err := ArticleMarkdownFromHTML(html)
	if err != nil {
		return Article{}, fmt.Errorf("converting %q: %w", canonicalTitle, err)
	}

	article := Article{Title: canonicalTitle, URL: canonicalURL, Markdown: markdown}
	c.mu.Lock()
	c.cache[canonicalURL] = article
	c.mu.Unlock()
	return article, nil
}

// articleSlug turns a display title into Wikipedia's URL slug form: spaces
// become underscores (cmd/wikispike/fetch.go's same rule).
func articleSlug(title string) string {
	return strings.ReplaceAll(strings.TrimSpace(title), " ", "_")
}

// articleURLFor builds an exact-title URL under base — the REST fetch URL
// when base is fetchBase, or the canonical wiki URL when base is wikiBase;
// both use the same slug-escaping rule.
func articleURLFor(base, title string) string {
	return base + url.PathEscape(articleSlug(title))
}

// ArticleURL builds the wiki URL an exact, non-redirect title would resolve
// to — the same escaping fetchAt's own request and canonicalFromURL's
// response both go through (articleURLFor(wikiBase, title)), exported for a
// caller that needs a Document Store key for a title it has not fetched yet
// (cmd/noggin's Picker star, C6/M3.4): a title-search SearchResult carries
// no URL of its own, so the Picker derives one from the title alone rather
// than requiring a Fetch first, the same trade D122 already accepts for the
// fetch cache ("noggin's own records always hold the canonical title once
// an article has been opened"). This is exact when title is already
// canonical, which title search's own results always are unless the hit
// itself happens to be a redirect page — that rare case is reconciled the
// same way any non-canonical title is: fetching it (M4.2) resolves the true
// canonical URL, which is what Activate/Peek then record.
func ArticleURL(title string) string {
	return articleURLFor(wikiBase, title)
}

// fetchCanonicalHTML GETs title's Parsoid HTML, following any redirect to
// the canonical article, and returns the raw HTML plus the canonical title
// and URL taken from the final response.
func fetchCanonicalHTML(client *http.Client, fetchBaseURL, wikiBaseURL, title string) (html, canonicalTitle, canonicalURL string, err error) {
	target := articleURLFor(fetchBaseURL, title)
	req, err := http.NewRequest(http.MethodGet, target, nil)
	if err != nil {
		return "", "", "", fmt.Errorf("building request for %q: %w", title, err)
	}
	req.Header.Set("User-Agent", userAgent())
	req.Header.Set("Accept", "text/html")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", "", &OfflineError{Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", "", &ServiceError{StatusCode: resp.StatusCode, Status: resp.Status}
	}

	canonicalTitle, canonicalURL, err = canonicalFromURL(wikiBaseURL, req.URL, resp.Request.URL)
	if err != nil {
		return "", "", "", fmt.Errorf("resolving canonical URL for %q: %w", title, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", "", fmt.Errorf("reading %q: %w", title, err)
	}
	return string(body), canonicalTitle, canonicalURL, nil
}

// canonicalFromURL extracts the canonical title and wiki article URL from
// finalURL — the request URL after Go's client has followed any redirect
// (findings/wikimedia-api.md). A redirect to a canonical title lands on
// .../rest.php/v1/page/<Canonical_Title>/html; an already-canonical
// request's own URL is .../rest_v1/page/<Title> — in both cases the slug is
// the escaped path segment immediately before a trailing "html" segment
// that only the redirect shape has, or the URL's last segment otherwise.
//
// Which shape applies is decided structurally (M2.5/TRIAGE T7), by
// comparing finalURL's path against requestedURL's own — the exact request
// this package built and sent, before any redirect — rather than by
// guessing from the slug text itself. An earlier version treated a literal
// trailing "html" segment as always meaning "redirected": for today's two
// documented Wikimedia shapes that guess happens to be safe for every
// title (fetchBase's own fixed ".../page/html/" segment sits immediately
// before the title in the non-redirected case, so even a title whose own
// slug is "html" reads back correctly by coincidence, not by design) — but
// it depends on fetchBase always having that exact shape, a fact this
// function had no way to enforce or even see. Comparing paths instead
// makes correctness a structural guarantee rather than a coincidence of
// the current base URL, at the cost of the one extra parameter.
func canonicalFromURL(wikiBaseURL string, requestedURL, finalURL *url.URL) (title, wikiURL string, err error) {
	segments := strings.Split(strings.Trim(finalURL.Path, "/"), "/")
	if len(segments) == 0 || segments[len(segments)-1] == "" {
		return "", "", fmt.Errorf("canonical URL %q has no path", finalURL)
	}
	slug := segments[len(segments)-1]
	if finalURL.Path != requestedURL.Path {
		if len(segments) < 2 {
			return "", "", fmt.Errorf("redirected canonical URL %q is missing its title segment", finalURL)
		}
		slug = segments[len(segments)-2]
	}
	unescaped, err := url.PathUnescape(slug)
	if err != nil {
		return "", "", fmt.Errorf("unescaping canonical title %q: %w", slug, err)
	}
	title = strings.ReplaceAll(unescaped, "_", " ")
	wikiURL = wikiBaseURL + slug
	return title, wikiURL, nil
}
