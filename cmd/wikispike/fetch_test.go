package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// sampleTitles are the S5 judging sample, in spec order.
var sampleTitles = []string{
	"Why is there anything at all?",
	"E (mathematical constant)",
	"List of programming languages by type",
	"Major Arcana",
}

func TestFetchSendsDescriptiveUserAgent(t *testing.T) {
	t.Setenv("WIKISPIKE_CONTACT", "")

	var gotUA, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		gotPath = r.URL.EscapedPath()
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><body><p>hi</p></body></html>"))
	}))
	defer srv.Close()

	body, err := fetchArticleHTML(srv.URL+"/page/html/", "Major Arcana")
	if err != nil {
		t.Fatalf("fetchArticleHTML: %v", err)
	}
	if !strings.Contains(body, "hi") {
		t.Errorf("body not returned: %q", body)
	}
	// S8: the agent must name the tool and carry a contact, per Wikimedia's
	// User-Agent policy — not Go's default "Go-http-client/1.1".
	if !strings.HasPrefix(gotUA, "noggin-wikispike/") {
		t.Errorf("User-Agent does not name the tool: %q", gotUA)
	}
	if !strings.Contains(gotUA, "(") || !strings.Contains(gotUA, ")") {
		t.Errorf("User-Agent carries no parenthesised contact/detail: %q", gotUA)
	}
	if gotPath != "/page/html/Major_Arcana" {
		t.Errorf("path = %q, want /page/html/Major_Arcana", gotPath)
	}
}

func TestUserAgentHonoursContactOverride(t *testing.T) {
	t.Setenv("WIKISPIKE_CONTACT", "someone@example.com")
	if got := userAgent(); !strings.Contains(got, "someone@example.com") {
		t.Errorf("userAgent() = %q, want it to carry the override contact", got)
	}
}

func TestArticleURLEscapesTitles(t *testing.T) {
	for _, tc := range []struct{ title, want string }{
		{"Major Arcana", "B/Major_Arcana"},
		{"Why is there anything at all?", "B/Why_is_there_anything_at_all%3F"},
		{"E (mathematical constant)", "B/E_%28mathematical_constant%29"},
		{"List of programming languages by type", "B/List_of_programming_languages_by_type"},
	} {
		if got := articleURL("B/", tc.title); got != tc.want {
			t.Errorf("articleURL(%q) = %q, want %q", tc.title, got, tc.want)
		}
	}
}

func TestFetchReportsNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	if _, err := fetchArticleHTML(srv.URL+"/page/html/", "No Such Article Xyzzy"); err == nil {
		t.Fatal("expected an error for a 404 response")
	}
}

// TestFetchLiveSampleTitles is M1.2's acceptance check against the real API. It
// needs the network, so `make check` must stay hermetic: run it with
// WIKISPIKE_LIVE=1 go test ./cmd/wikispike -run Live -v
func TestFetchLiveSampleTitles(t *testing.T) {
	if os.Getenv("WIKISPIKE_LIVE") == "" {
		t.Skip("set WIKISPIKE_LIVE=1 to hit en.wikipedia.org")
	}
	for _, title := range sampleTitles {
		t.Run(title, func(t *testing.T) {
			html, err := fetchArticleHTML(restBase, title)
			if err != nil {
				t.Fatalf("fetchArticleHTML(%q): %v", title, err)
			}
			if len(html) < 1000 {
				t.Errorf("%q: suspiciously short response (%d bytes)", title, len(html))
			}
			if !strings.Contains(html, "<body") {
				t.Errorf("%q: response does not look like article HTML", title)
			}
		})
	}
}
