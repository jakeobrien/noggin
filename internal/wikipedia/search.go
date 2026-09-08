package wikipedia

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// searchBase is Wikimedia's title-search endpoint (S2): prefix/title
// matching over article titles, not full-text search — see
// .charm/cycle/findings/wikimedia-api.md.
const searchBase = "https://en.wikipedia.org/w/rest.php/v1/search/title"

// searchHTTPClient is scoped to search. M2.3's canonical fetch bounds its own
// requests with a timeout per its own acceptance criterion, so this is not
// shared with it.
var searchHTTPClient = &http.Client{Timeout: 10 * time.Second}

// SearchResult is one row of a Wikipedia title search (S2): the article
// title plus, when Wikipedia supplies one, its short description.
type SearchResult struct {
	Title       string
	Description string // empty when Wikipedia has none for this title
}

// ServiceError reports that Wikimedia answered the request but with a
// non-2xx status — S2's "service error", as distinct from being offline.
// Shared by both of this package's Wikimedia callers (M2.5/TRIAGE T8): title
// search (this file) and canonical article fetch (fetch.go).
type ServiceError struct {
	StatusCode int
	Status     string
}

func (e *ServiceError) Error() string {
	return fmt.Sprintf("wikipedia: %s", e.Status)
}

// OfflineError reports that the request never reached Wikimedia at all
// (DNS failure, connection refused, timeout mid-dial) — S2's "offline", as
// distinct from a service error. Shared by both of this package's Wikimedia
// callers (M2.5/TRIAGE T8): title search (this file) and canonical article
// fetch (fetch.go).
type OfflineError struct {
	Err error
}

func (e *OfflineError) Error() string {
	return fmt.Sprintf("wikipedia: offline: %v", e.Err)
}

func (e *OfflineError) Unwrap() error { return e.Err }

// userAgent identifies noggin to Wikimedia (S8) per their User-Agent policy,
// which asks for an informative agent naming the tool and a way to reach
// whoever runs it. NOGGIN_WIKIPEDIA_CONTACT overrides the contact detail per
// operator without editing code; the default says plainly what is missing.
//
// https://foundation.wikimedia.org/wiki/Policy:Wikimedia_Foundation_User-Agent_Policy
func userAgent() string {
	contact := os.Getenv("NOGGIN_WIKIPEDIA_CONTACT")
	if contact == "" {
		contact = "no contact configured, see NOGGIN_WIKIPEDIA_CONTACT"
	}
	return fmt.Sprintf("noggin/0.1 (%s)", contact)
}

// Search queries Wikimedia's title-search endpoint for article titles
// matching query (prefix/title matching, not full-text — S2), bounded to at
// most limit results, in Wikipedia's own ranking order.
func Search(query string, limit int) ([]SearchResult, error) {
	return searchTitles(searchBase, query, limit)
}

// searchResponse mirrors the title-search endpoint's JSON shape; only the
// fields S2 needs are decoded (findings/wikimedia-api.md). Description is a
// pointer because Wikimedia sends `null` for titles it has none for.
type searchResponse struct {
	Pages []struct {
		Title       string  `json:"title"`
		Description *string `json:"description"`
	} `json:"pages"`
}

// searchTitles is Search's implementation against an overridable base URL,
// so tests can point it at an httptest server instead of the network.
func searchTitles(base, query string, limit int) ([]SearchResult, error) {
	target, err := searchURL(base, query, limit)
	if err != nil {
		return nil, fmt.Errorf("building search request for %q: %w", query, err)
	}

	req, err := http.NewRequest(http.MethodGet, target, nil)
	if err != nil {
		return nil, fmt.Errorf("building search request for %q: %w", query, err)
	}
	req.Header.Set("User-Agent", userAgent())
	req.Header.Set("Accept", "application/json")

	resp, err := searchHTTPClient.Do(req)
	if err != nil {
		return nil, &OfflineError{Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &ServiceError{StatusCode: resp.StatusCode, Status: resp.Status}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading search response for %q: %w", query, err)
	}

	var parsed searchResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("parsing search response for %q: %w", query, err)
	}

	results := make([]SearchResult, 0, len(parsed.Pages))
	for _, p := range parsed.Pages {
		r := SearchResult{Title: p.Title}
		if p.Description != nil {
			r.Description = *p.Description
		}
		results = append(results, r)
	}
	return results, nil
}

// searchURL builds the title-search request URL for query, bounded to limit
// results.
func searchURL(base, query string, limit int) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("parsing search base %q: %w", base, err)
	}
	q := u.Query()
	q.Set("q", query)
	q.Set("limit", fmt.Sprintf("%d", limit))
	u.RawQuery = q.Encode()
	return u.String(), nil
}
