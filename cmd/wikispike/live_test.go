package main

import (
	"os"
	"sync"
	"testing"
)

// This file is the M2.5 consolidation (TRIAGE T1): M1.2, M2.2, M2.3 and M2.4
// each added their own opt-in live test that independently re-fetched all
// four S5 titles, so running the whole suite (`-run Live`) made 5*4 = 20
// requests and tripped the REST API's 429. Every live test now goes through
// fetchLiveSample, which fetches each title's raw HTML at most once per test
// binary run regardless of how many tests ask for it — one request per
// title, four total, no matter how many live tests exist.
//
// WIKISPIKE_LIVE=1 go test ./cmd/wikispike -run Live -v

// skipUnlessLive is the shared opt-in guard every live test starts with, so
// `make check` stays hermetic.
func skipUnlessLive(t *testing.T) {
	t.Helper()
	if os.Getenv("WIKISPIKE_LIVE") == "" {
		t.Skip("set WIKISPIKE_LIVE=1 to hit en.wikipedia.org")
	}
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
			html, err := fetchArticleHTML(restBase, st)
			liveHTML[st] = html
			liveHTMLErr[st] = err
		}
	})
	if err := liveHTMLErr[title]; err != nil {
		t.Fatalf("fetchArticleHTML(%q): %v", title, err)
	}
	html, ok := liveHTML[title]
	if !ok {
		t.Fatalf("fetchLiveSample: %q is not one of the S5 sample titles", title)
	}
	return html
}

// liveMarkdown runs the full conversion pipeline (strip, delink, convert,
// clean up) against title's cached raw HTML — the same stages
// articleMarkdown runs, without repeating its own network fetch.
func liveMarkdown(t *testing.T, title string) string {
	t.Helper()
	markdown, err := articleMarkdownFromHTML(fetchLiveSample(t, title))
	if err != nil {
		t.Fatalf("articleMarkdownFromHTML(%q): %v", title, err)
	}
	return markdown
}
