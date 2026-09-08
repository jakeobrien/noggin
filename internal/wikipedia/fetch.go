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
const fetchBase = "https://en.wikipedia.org/api/rest_v1/page/html/"

// wikiBase is the canonical, human-facing article URL prefix S3 records —
// distinct from fetchBase, which is the REST path used only to fetch HTML.
const wikiBase = "https://en.wikipedia.org/wiki/"

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
// canonical URL. A Client is safe for concurrent use.
type Client struct {
	httpClient *http.Client

	mu    sync.Mutex
	cache map[string]Article // keyed by canonical URL
}

// NewClient returns a Client with an empty cache, ready to Fetch.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: fetchTimeout},
		cache:      make(map[string]Article),
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
func (c *Client) Fetch(title string, refresh bool) (Article, error) {
	return c.fetchAt(fetchBase, wikiBase, title, refresh)
}

// fetchAt is Fetch's implementation against overridable base URLs, so tests
// point it at an httptest server instead of the network.
func (c *Client) fetchAt(fetchBase, wikiBase, title string, refresh bool) (Article, error) {
	candidateURL := articleURLFor(wikiBase, title)

	if !refresh {
		c.mu.Lock()
		cached, ok := c.cache[candidateURL]
		c.mu.Unlock()
		if ok {
			return cached, nil
		}
	}

	html, canonicalTitle, canonicalURL, err := fetchCanonicalHTML(c.httpClient, fetchBase, wikiBase, title)
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

// fetchCanonicalHTML GETs title's Parsoid HTML, following any redirect to
// the canonical article, and returns the raw HTML plus the canonical title
// and URL taken from the final response.
func fetchCanonicalHTML(client *http.Client, fetchBase, wikiBase, title string) (html, canonicalTitle, canonicalURL string, err error) {
	target := articleURLFor(fetchBase, title)
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

	canonicalTitle, canonicalURL, err = canonicalFromURL(wikiBase, resp.Request.URL)
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
// the escaped path segment immediately before a trailing "html" segment, or
// the URL's last segment when there is none.
func canonicalFromURL(wikiBase string, finalURL *url.URL) (title, wikiURL string, err error) {
	segments := strings.Split(strings.Trim(finalURL.Path, "/"), "/")
	if len(segments) == 0 || segments[len(segments)-1] == "" {
		return "", "", fmt.Errorf("canonical URL %q has no path", finalURL)
	}
	slug := segments[len(segments)-1]
	if slug == "html" && len(segments) >= 2 {
		slug = segments[len(segments)-2]
	}
	unescaped, err := url.PathUnescape(slug)
	if err != nil {
		return "", "", fmt.Errorf("unescaping canonical title %q: %w", slug, err)
	}
	title = strings.ReplaceAll(unescaped, "_", " ")
	wikiURL = wikiBase + slug
	return title, wikiURL, nil
}
