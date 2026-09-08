package wikipedia

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func newTestClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: fetchTimeout},
		cache:      make(map[string]Article),
	}
}

// TestArticleURLMatchesCanonicalFromURLEscaping is C6/M3.4's own
// correctness check for ArticleURL, the Picker star's pre-fetch key
// derivation: for a title that needs URL-escaping (a literal "+", which
// url.PathEscape encodes but a naive space->underscore substitution alone
// would not), ArticleURL must produce exactly what fetchAt/canonicalFromURL
// resolve an already-canonical fetch of the same title to — the whole
// point of deriving one from the other (articleURLFor) rather than
// hand-rolling a second escaping rule that could drift from it.
func TestArticleURLMatchesCanonicalFromURLEscaping(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/C++", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html></html>"))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	c := newTestClient()
	article, err := c.fetchAt(server.URL+"/", "https://en.wikipedia.org/wiki/", "C++", false)
	if err != nil {
		t.Fatalf("fetchAt: unexpected error: %v", err)
	}
	if got := ArticleURL("C++"); got != article.URL {
		t.Errorf("ArticleURL(%q) = %q, want the same canonical URL fetchAt resolved: %q", "C++", got, article.URL)
	}
}

// TestFetchAtRedirect covers a title that redirects to the canonical
// article: the returned Article carries the canonical title/URL from the
// final response, not the requested title (M2.3's done-when, S3).
func TestFetchAtRedirect(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/page/html/Euler", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/w/rest.php/v1/page/Leonhard_Euler/html", http.StatusTemporaryRedirect)
	})
	mux.HandleFunc("/w/rest.php/v1/page/Leonhard_Euler/html", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html><body><p>Euler content</p></body></html>"))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	c := newTestClient()
	article, err := c.fetchAt(server.URL+"/page/html/", server.URL+"/wiki/", "Euler", false)
	if err != nil {
		t.Fatalf("fetchAt: unexpected error: %v", err)
	}
	if article.Title != "Leonhard Euler" {
		t.Errorf("Article.Title = %q, want %q", article.Title, "Leonhard Euler")
	}
	wantURL := server.URL + "/wiki/Leonhard_Euler"
	if article.URL != wantURL {
		t.Errorf("Article.URL = %q, want %q", article.URL, wantURL)
	}
	if article.Markdown == "" {
		t.Error("Article.Markdown is empty")
	}
}

// TestFetchAtNonRedirectedTitleSlugHTMLIsNotTruncated pins M2.5/TRIAGE T7's
// invariant directly: a non-redirected fetch whose title's own slug happens
// to render as "html" comes back whole, whatever the mechanics happen to
// be. It does *not* distinguish the current structural redirect detection
// (compare finalURL's path to the exact URL this package requested) from
// the text-heuristic it replaced (a literal trailing "html" segment means
// "redirected") — given fetchBase's own fixed ".../page/html/" shape, the
// old heuristic already produced the right answer here too, by coincidence
// rather than by design (verified: this test still passes reverted to the
// pre-M2.5 heuristic). The structural check's real benefit is not
// depending on that coincidence — it stays correct even if fetchBase's own
// shape ever changed — which no single input-based test can demonstrate;
// this test only guards the observable behavior against a regression
// either implementation could introduce.
func TestFetchAtNonRedirectedTitleSlugHTMLIsNotTruncated(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html><body><p>About markup</p></body></html>"))
	}))
	defer server.Close()

	c := newTestClient()
	article, err := c.fetchAt(server.URL+"/page/html/", server.URL+"/wiki/", "html", false)
	if err != nil {
		t.Fatalf("fetchAt: unexpected error: %v", err)
	}
	if article.Title != "html" {
		t.Errorf("Article.Title = %q, want %q (not truncated)", article.Title, "html")
	}
	wantURL := server.URL + "/wiki/html"
	if article.URL != wantURL {
		t.Errorf("Article.URL = %q, want %q", article.URL, wantURL)
	}
}

// TestFetchAtRedirectToSingleSegmentPathIsRejected covers canonicalFromURL's
// len(segments) < 2 guard in its redirected branch: a redirect whose final
// path has only one segment (so there is no title segment ahead of it to
// recover) is reported as an error, not a panic or a silently wrong title.
// Contrived — Wikimedia's real redirect shape always has at least two
// segments — but the guard exists precisely because this function trusts
// the response's shape, so it must fail cleanly if that trust is ever
// misplaced.
func TestFetchAtRedirectToSingleSegmentPathIsRejected(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/page/html/Foo", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/bare", http.StatusTemporaryRedirect)
	})
	mux.HandleFunc("/bare", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html><body><p>Unreachable in practice</p></body></html>"))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	c := newTestClient()
	_, err := c.fetchAt(server.URL+"/page/html/", server.URL+"/wiki/", "Foo", false)
	if err == nil {
		t.Fatal("fetchAt: expected an error for a redirect landing on a single-segment path, got nil")
	}
}

// TestFetchAtNotFound covers a 404 (article gone), reported distinctly from
// an offline failure (M2.3's done-when, S3).
func TestFetchAtNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := newTestClient()
	_, err := c.fetchAt(server.URL+"/page/html/", server.URL+"/wiki/", "Nonexistent Article", false)
	if err == nil {
		t.Fatal("fetchAt: expected an error, got nil")
	}
	var svcErr *ServiceError
	if !errors.As(err, &svcErr) {
		t.Fatalf("fetchAt: error = %v (%T), want *ServiceError", err, err)
	}
	if svcErr.StatusCode != http.StatusNotFound {
		t.Errorf("ServiceError.StatusCode = %d, want %d", svcErr.StatusCode, http.StatusNotFound)
	}
}

// TestFetchAtTimeout covers a request that exceeds the client's timeout,
// reported as an offline-shaped failure since the request never completed
// (M2.3's done-when, S3).
func TestFetchAtTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.Write([]byte("<html><body><p>too slow</p></body></html>"))
	}))
	defer server.Close()

	c := &Client{httpClient: &http.Client{Timeout: 5 * time.Millisecond}, cache: make(map[string]Article)}
	_, err := c.fetchAt(server.URL+"/page/html/", server.URL+"/wiki/", "Slow Article", false)
	if err == nil {
		t.Fatal("fetchAt: expected a timeout error, got nil")
	}
	var offErr *OfflineError
	if !errors.As(err, &offErr) {
		t.Fatalf("fetchAt: error = %v (%T), want *OfflineError", err, err)
	}
}

// TestFetchAtCacheHit covers a repeat fetch of an already-canonical title
// hitting the session cache without a second request (M2.3's done-when, S3).
func TestFetchAtCacheHit(t *testing.T) {
	var requests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		w.Write([]byte("<html><body><p>Stable content</p></body></html>"))
	}))
	defer server.Close()

	c := newTestClient()
	first, err := c.fetchAt(server.URL+"/page/html/", server.URL+"/wiki/", "Stable Article", false)
	if err != nil {
		t.Fatalf("fetchAt (first): unexpected error: %v", err)
	}
	second, err := c.fetchAt(server.URL+"/page/html/", server.URL+"/wiki/", "Stable Article", false)
	if err != nil {
		t.Fatalf("fetchAt (second): unexpected error: %v", err)
	}
	if got := atomic.LoadInt32(&requests); got != 1 {
		t.Fatalf("server received %d requests, want 1 (second fetch should be a cache hit)", got)
	}
	if second != first {
		t.Errorf("fetchAt (second) = %+v, want identical to first %+v", second, first)
	}
}

// TestFetchAtRefreshBypass covers refresh bypassing the cache and making a
// second request, repopulating the cache (M2.3's done-when, S3's Refresh).
func TestFetchAtRefreshBypass(t *testing.T) {
	var requests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&requests, 1)
		w.Write([]byte("<html><body><p>Version " + strconv.Itoa(int(n)) + "</p></body></html>"))
	}))
	defer server.Close()

	c := newTestClient()
	first, err := c.fetchAt(server.URL+"/page/html/", server.URL+"/wiki/", "Refreshed Article", false)
	if err != nil {
		t.Fatalf("fetchAt (initial): unexpected error: %v", err)
	}
	second, err := c.fetchAt(server.URL+"/page/html/", server.URL+"/wiki/", "Refreshed Article", true)
	if err != nil {
		t.Fatalf("fetchAt (refresh): unexpected error: %v", err)
	}
	if got := atomic.LoadInt32(&requests); got != 2 {
		t.Fatalf("server received %d requests, want 2 (refresh should bypass the cache)", got)
	}
	if second.Markdown == first.Markdown {
		t.Error("fetchAt (refresh): markdown unchanged, expected the second request's content")
	}

	// The refreshed result replaces the cached entry: a following
	// non-refresh fetch must return it, not the stale first version.
	third, err := c.fetchAt(server.URL+"/page/html/", server.URL+"/wiki/", "Refreshed Article", false)
	if err != nil {
		t.Fatalf("fetchAt (post-refresh): unexpected error: %v", err)
	}
	if got := atomic.LoadInt32(&requests); got != 2 {
		t.Fatalf("server received %d requests, want 2 (post-refresh fetch should be a cache hit)", got)
	}
	if third != second {
		t.Errorf("fetchAt (post-refresh) = %+v, want the refreshed article %+v", third, second)
	}
}

// TestFetchAtZeroValueClientDoesNotPanic is M2.4/TRIAGE T4's acceptance
// criterion: a Client constructed as a bare zero value (Client{}), not via
// NewClient, must not panic on its first Fetch — a nil httpClient panics
// inside (*http.Client).Do's first field access, and a nil cache map panics
// on the write that populates it after a successful fetch. ensureInit's
// lazy initialization must cover both.
func TestFetchAtZeroValueClientDoesNotPanic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html><body><p>Zero-value content</p></body></html>"))
	}))
	defer server.Close()

	var c Client // deliberately not NewClient()
	article, err := c.fetchAt(server.URL+"/page/html/", server.URL+"/wiki/", "Zero Value Article", false)
	if err != nil {
		t.Fatalf("fetchAt on a zero-value Client: unexpected error: %v", err)
	}
	if article.Markdown == "" {
		t.Error("fetchAt on a zero-value Client: Article.Markdown is empty")
	}

	// A second fetch of the same title should now hit the lazily-created
	// cache without panicking either.
	if _, err := c.fetchAt(server.URL+"/page/html/", server.URL+"/wiki/", "Zero Value Article", false); err != nil {
		t.Fatalf("fetchAt (second, cached): unexpected error: %v", err)
	}
}

// TestClientFetchViaNewClientDelegatesToFetchAt is M2.9's acceptance
// criterion (TRIAGE T1, T22): NewClient() and (*Client).Fetch are this
// package's real exported entry points for fetching an article, but every
// other test in this file calls fetchAt directly with an explicit
// overridable base, so a regression in Fetch's own one-line delegation
// (`return c.fetchAt(fetchBase, wikiBase, title, refresh)` — e.g. the two
// base arguments swapped, or refresh silently dropped) would go uncaught.
// This test points the package's normally-hardcoded fetchBase/wikiBase at
// an httptest server, restores them on return, and calls Fetch on a
// NewClient()-built Client — exercising the real construction path
// (ensureInit via NewClient, not newTestClient's direct struct literal)
// alongside the real exported method.
func TestClientFetchViaNewClientDelegatesToFetchAt(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/page/html/Euler", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/w/rest.php/v1/page/Leonhard_Euler/html", http.StatusTemporaryRedirect)
	})
	mux.HandleFunc("/w/rest.php/v1/page/Leonhard_Euler/html", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html><body><p>Euler content</p></body></html>"))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	origFetchBase, origWikiBase := fetchBase, wikiBase
	fetchBase, wikiBase = server.URL+"/page/html/", server.URL+"/wiki/"
	defer func() { fetchBase, wikiBase = origFetchBase, origWikiBase }()

	c := NewClient()
	article, err := c.Fetch("Euler", false)
	if err != nil {
		t.Fatalf("Fetch: unexpected error: %v", err)
	}
	if article.Title != "Leonhard Euler" {
		t.Errorf("Article.Title = %q, want %q", article.Title, "Leonhard Euler")
	}
	wantURL := server.URL + "/wiki/Leonhard_Euler"
	if article.URL != wantURL {
		t.Errorf("Article.URL = %q, want %q", article.URL, wantURL)
	}
	if article.Markdown == "" {
		t.Error("Article.Markdown is empty")
	}
}

// TestFetchAtConcurrentUseIsRaceFree backs the Client doc comment's "safe
// for concurrent use" claim (M2.4/TRIAGE T5) with an actual test: many
// goroutines calling fetchAt concurrently — on a zero-value Client, so
// ensureInit's lazy first-use initialization is itself exercised under
// race — against a shared title (exercising concurrent cache reads/writes
// on the same key) and distinct titles (exercising concurrent cache writes
// on different keys). Meaningful only under `go test -race`.
func TestFetchAtConcurrentUseIsRaceFree(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html><body><p>Concurrent content</p></body></html>"))
	}))
	defer server.Close()

	var c Client // deliberately not NewClient()
	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		i := i
		go func() {
			defer wg.Done()
			title := "Shared Article"
			if i%2 == 0 {
				title = "Distinct Article " + strconv.Itoa(i)
			}
			if _, err := c.fetchAt(server.URL+"/page/html/", server.URL+"/wiki/", title, i%5 == 0); err != nil {
				t.Errorf("fetchAt(%q): unexpected error: %v", title, err)
			}
		}()
	}
	wg.Wait()
}
