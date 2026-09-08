package wikipedia

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// This file mirrors cmd/wikispike's live_test.go (same opt-in guard, same
// once-per-title fetch cache across the whole test binary run) so the stage
// tests moved here in M2.1 (D120) keep exercising real article HTML without
// re-fetching per test. It has its own minimal exact-title fetch rather than
// importing cmd/wikispike's (a `main` package can't be imported) — this is
// test-only fixture plumbing, not the package's production fetch path; M2.3
// adds the real canonical fetch this package will actually ship.
//
// WIKISPIKE_LIVE=1 go test ./internal/wikipedia -run Live -v

// skipUnlessLive is the shared opt-in guard every live test starts with, so
// `make check` stays hermetic (D101).
func skipUnlessLive(t *testing.T) {
	t.Helper()
	if os.Getenv("WIKISPIKE_LIVE") == "" {
		t.Skip("set WIKISPIKE_LIVE=1 to hit en.wikipedia.org")
	}
}

// liveRestBase is the same Wikimedia REST endpoint cmd/wikispike's fetch.go
// uses (S2: English Wikipedia only, no auth, no search).
const liveRestBase = "https://en.wikipedia.org/api/rest_v1/page/html/"

var liveHTTPClient = &http.Client{Timeout: 30 * time.Second}

// fetchLiveArticleHTML GETs an article's Parsoid HTML by exact title —
// test-only, mirrors cmd/wikispike/fetch.go's fetchArticleHTML.
func fetchLiveArticleHTML(title string) (string, error) {
	slug := strings.ReplaceAll(strings.TrimSpace(title), " ", "_")
	target := liveRestBase + url.PathEscape(slug)
	req, err := http.NewRequest(http.MethodGet, target, nil)
	if err != nil {
		return "", fmt.Errorf("building request for %q: %w", title, err)
	}
	req.Header.Set("User-Agent", "noggin-wikipedia-pkg-test/0.1 (local experimental test, not distributed)")
	req.Header.Set("Accept", "text/html")

	resp, err := liveHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetching %q: %w", title, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetching %q: %s returned %s", title, target, resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading %q: %w", title, err)
	}
	return string(body), nil
}

var (
	liveHTMLOnce sync.Once
	liveHTML     map[string]string
	liveHTMLErr  map[string]error
)

// fetchLiveSample returns title's raw fetched HTML, fetching all four S5
// sample titles once — on whichever live test runs first — and caching the
// rest for every other live test in the same run.
func fetchLiveSample(t *testing.T, title string) string {
	t.Helper()
	liveHTMLOnce.Do(func() {
		liveHTML = make(map[string]string, len(sampleTitles))
		liveHTMLErr = make(map[string]error, len(sampleTitles))
		for _, st := range sampleTitles {
			html, err := fetchLiveArticleHTML(st)
			liveHTML[st] = html
			liveHTMLErr[st] = err
		}
	})
	if err := liveHTMLErr[title]; err != nil {
		t.Fatalf("fetchLiveArticleHTML(%q): %v", title, err)
	}
	html, ok := liveHTML[title]
	if !ok {
		t.Fatalf("fetchLiveSample: %q is not one of the S5 sample titles", title)
	}
	return html
}

// liveMarkdown runs the full conversion pipeline against title's cached raw
// HTML — the same stages ArticleMarkdownFromHTML runs, without repeating its
// own network fetch.
func liveMarkdown(t *testing.T, title string) string {
	t.Helper()
	markdown, err := ArticleMarkdownFromHTML(fetchLiveSample(t, title))
	if err != nil {
		t.Fatalf("ArticleMarkdownFromHTML(%q): %v", title, err)
	}
	return markdown
}
