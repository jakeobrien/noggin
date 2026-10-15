package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// restBase is Wikimedia's public REST API for parsed page HTML (S2: English
// Wikipedia only, no auth, no search). The response is Parsoid HTML, which is
// not the same markup as a rendered article page — see
// docs/dev/wikipedia.md for which attributes to key stripping off.
const restBase = "https://en.wikipedia.org/api/rest_v1/page/html/"

// userAgent identifies the spike to Wikimedia (S8). Their policy asks for an
// informative agent naming the tool and a way to reach whoever runs it, so the
// contact is overridable per operator via WIKISPIKE_CONTACT without editing
// code; the default says plainly what this is.
//
// https://foundation.wikimedia.org/wiki/Policy:Wikimedia_Foundation_User-Agent_Policy
func userAgent() string {
	contact := os.Getenv("WIKISPIKE_CONTACT")
	if contact == "" {
		contact = "local experimental spike, not distributed"
	}
	return fmt.Sprintf("noggin-wikispike/0.1 (%s)", contact)
}

// articleURL builds the REST URL for an exact article title. Wikipedia titles
// use underscores for spaces in URLs, and everything else — parentheses, `?`,
// `&`, `#` — has to be percent-escaped or the path is truncated or misread
// (S2's punctuation-heavy sample titles are exactly this case).
func articleURL(base, title string) string {
	slug := strings.ReplaceAll(strings.TrimSpace(title), " ", "_")
	return base + url.PathEscape(slug)
}

// httpClient is shared so keep-alives are reused across the run; the timeout is
// politeness, not error handling (which is out of scope this cycle).
var httpClient = &http.Client{Timeout: 30 * time.Second}

// fetchArticleHTML GETs an article's Parsoid HTML by exact title.
func fetchArticleHTML(base, title string) (string, error) {
	target := articleURL(base, title)
	req, err := http.NewRequest(http.MethodGet, target, nil)
	if err != nil {
		return "", fmt.Errorf("building request for %q: %w", title, err)
	}
	req.Header.Set("User-Agent", userAgent())
	req.Header.Set("Accept", "text/html")

	resp, err := httpClient.Do(req)
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
