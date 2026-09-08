package wikipedia

import (
	"net/http"
	"os"
	"sync"
	"testing"
	"time"
)

// This file mirrors cmd/wikispike's live_test.go (same opt-in guard, same
// once-per-title fetch cache across the whole test binary run) so the stage
// tests moved here in M2.1 (D120) keep exercising real article HTML without
// re-fetching per test.
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

var liveHTTPClient = &http.Client{Timeout: 30 * time.Second}

// fetchLiveArticleHTML GETs an article's raw Parsoid HTML by exact title
// (M2.6/TRIAGE T10): several stage tests (convert_test.go, images_test.go,
// strip_test.go) need the *raw* HTML directly, ahead of conversion, so this
// cannot simply call Client.Fetch/fetchAt — both convert to markdown before
// returning. It instead calls fetchCanonicalHTML, the same unexported
// fetch-and-follow-redirects step fetchAt itself uses (fetch.go), so this
// fixture shares the package's one real HTTP-fetch implementation rather
// than keeping its own copy of the request-building/status-check/body-read
// logic — the duplication M2.1's handoff called out as provisional, resolved
// now that M2.3's real fetch has landed in this package. Canonical
// title/URL are discarded here; the stage tests only need the HTML.
func fetchLiveArticleHTML(title string) (string, error) {
	html, _, _, err := fetchCanonicalHTML(liveHTTPClient, fetchBase, wikiBase, title)
	return html, err
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
