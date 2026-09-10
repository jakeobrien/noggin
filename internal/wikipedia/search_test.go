package wikipedia

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestSearchTitlesResults covers a mix of results with and without a
// description (M2.2's done-when), and asserts Wikipedia's response order is
// preserved (S2: "in Wikipedia's order").
func TestSearchTitlesResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"pages":[
			{"title":"Leonhard Euler","description":"Swiss mathematician (1707-1783)"},
			{"title":"Euler's constant","description":null}
		]}`))
	}))
	defer server.Close()

	got, err := searchTitles(server.URL, "eul", 3)
	if err != nil {
		t.Fatalf("searchTitles: unexpected error: %v", err)
	}
	want := []SearchResult{
		{Title: "Leonhard Euler", Description: "Swiss mathematician (1707-1783)"},
		{Title: "Euler's constant", Description: ""},
	}
	if len(got) != len(want) {
		t.Fatalf("searchTitles: got %d results, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("searchTitles[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

// TestSearchTitlesEmpty covers an empty result list.
func TestSearchTitlesEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"pages":[]}`))
	}))
	defer server.Close()

	got, err := searchTitles(server.URL, "zzznoresults", 3)
	if err != nil {
		t.Fatalf("searchTitles: unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("searchTitles: got %d results, want 0: %+v", len(got), got)
	}
}

// TestSearchTitlesServiceError covers Wikimedia answering with a non-2xx
// status — S2's "service error", which must be distinguishable from being
// offline (M2.2's done-when).
func TestSearchTitlesServiceError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer server.Close()

	_, err := searchTitles(server.URL, "eul", 3)
	if err == nil {
		t.Fatal("searchTitles: expected an error, got nil")
	}
	var svcErr *ServiceError
	if !errors.As(err, &svcErr) {
		t.Fatalf("searchTitles: error = %v (%T), want *ServiceError", err, err)
	}
	if svcErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("ServiceError.StatusCode = %d, want %d", svcErr.StatusCode, http.StatusInternalServerError)
	}
	var offErr *OfflineError
	if errors.As(err, &offErr) {
		t.Fatalf("searchTitles: service error also matched *OfflineError: %v", err)
	}
}

// TestSearchTitlesOffline covers a connection failure (the request never
// reaching Wikimedia at all) — S2's "offline", distinguished from a service
// error (M2.2's done-when). Closing the server before the request means the
// listening socket is gone, so the client fails to connect.
func TestSearchTitlesOffline(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	url := server.URL
	server.Close()

	_, err := searchTitles(url, "eul", 3)
	if err == nil {
		t.Fatal("searchTitles: expected an error, got nil")
	}
	var offErr *OfflineError
	if !errors.As(err, &offErr) {
		t.Fatalf("searchTitles: error = %v (%T), want *OfflineError", err, err)
	}
	var svcErr *ServiceError
	if errors.As(err, &svcErr) {
		t.Fatalf("searchTitles: offline error also matched *ServiceError: %v", err)
	}
}

// TestClientSearchDelegatesToSearchTitles is M2.9's acceptance criterion
// (TRIAGE T1, T22): the package-level Search function and (*Client).Search
// are this package's real exported entry points for title search, but
// every other test in this file calls searchTitles directly with an
// explicit overridable base, so a regression in either's one-line
// delegation (`Search(query, limit) { return searchTitles(searchBase,
// query, limit) }`, `(*Client).Search(query, limit) { return Search(query,
// limit) }` — e.g. a dropped or swapped argument) would go uncaught. This
// test points the package's normally-hardcoded searchBase at an httptest
// server, restores it on return, and calls Search on a NewClient()-built
// Client — exercising both delegation hops from the real exported method
// down to searchTitles.
//
// Test-only mutation contract: same hazard class and rule as
// TestClientFetchViaNewClientDelegatesToFetchAt's doc comment in
// fetch_test.go (the canonical statement, TRIAGE T2) — this test mutates
// the package-level searchBase var the same way (save the original, point
// it at the httptest server, defer restoring it), so that test's
// no-t.Parallel()-while-either-mutates-its-base-var rule applies here too.
func TestClientSearchDelegatesToSearchTitles(t *testing.T) {
	var gotQuery, gotLimit string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query().Get("q")
		gotLimit = r.URL.Query().Get("limit")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"pages":[
			{"title":"Leonhard Euler","description":"Swiss mathematician (1707-1783)"}
		]}`))
	}))
	defer server.Close()

	origSearchBase := searchBase
	searchBase = server.URL
	defer func() { searchBase = origSearchBase }()

	c := NewClient()
	got, err := c.Search("eul", 3)
	if err != nil {
		t.Fatalf("Client.Search: unexpected error: %v", err)
	}

	// The request that actually reached the server must carry Search's own
	// arguments unaltered through both delegation hops (Client.Search ->
	// package Search -> searchTitles) — checked directly against the
	// request, not just inferred from a successful response, so a mutation
	// that drops or corrupts an argument along the way (this handler does
	// not itself filter by query/limit) cannot slip past unnoticed.
	if gotQuery != "eul" {
		t.Errorf("request query param %q = %q, want %q", "q", gotQuery, "eul")
	}
	if gotLimit != "3" {
		t.Errorf("request query param %q = %q, want %q", "limit", gotLimit, "3")
	}

	want := []SearchResult{{Title: "Leonhard Euler", Description: "Swiss mathematician (1707-1783)"}}
	if len(got) != len(want) {
		t.Fatalf("Client.Search: got %d results, want %d: %+v", len(got), len(want), got)
	}
	if got[0] != want[0] {
		t.Errorf("Client.Search[0] = %+v, want %+v", got[0], want[0])
	}
}

// TestSearchTitlesUserAgent covers noggin's descriptive User-Agent header
// (M2.2's done-when, S3).
func TestSearchTitlesUserAgent(t *testing.T) {
	var gotUA string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"pages":[]}`))
	}))
	defer server.Close()

	if _, err := searchTitles(server.URL, "eul", 3); err != nil {
		t.Fatalf("searchTitles: unexpected error: %v", err)
	}
	if gotUA == "" {
		t.Fatal("searchTitles: request carried no User-Agent header")
	}
	if gotUA != userAgent() {
		t.Errorf("User-Agent = %q, want %q", gotUA, userAgent())
	}
}
