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

// TestSearchTitlesUserAgent covers noggin's descriptive User-Agent header
// (M2.2's done-when, S8).
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
