package main

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"noggin/internal/wikipedia"
)

// fakeWikipediaSearcher is the C6/M3.3 test double for pickerWikipediaClient:
// a fixed, query-keyed table of results or errors, so a model test can drive
// Wikipedia mode's search without ever reaching the network. calls records
// every query actually searched, so a test can assert a stale generation's
// search never ran at all, not merely that its result was discarded.
type fakeWikipediaSearcher struct {
	results map[string][]wikipedia.SearchResult
	errs    map[string]error
	calls   []string
}

func (f *fakeWikipediaSearcher) Search(query string, limit int) ([]wikipedia.SearchResult, error) {
	f.calls = append(f.calls, query)
	if err, ok := f.errs[query]; ok {
		return nil, err
	}
	return f.results[query], nil
}

// pickerWikipedia is pickerWithWikipediaActive (picker_wikipedia_mode_test.go)
// plus a fake client installed in place of newModel's real one, so this
// file's tests never construct a *wikipedia.Client that could reach out.
func pickerWikipedia(t *testing.T, fake *fakeWikipediaSearcher) model {
	t.Helper()
	m := pickerWithWikipediaActive(t)
	return m.withWikipediaClient(fake)
}

// TestPickerWikipediaModeRendersTitleAndDescription is M3.3's own
// "Results rows show title + description" acceptance criterion: a
// successful search's rows land in p.results.wikipedia in Wikipedia's own
// order, and the rendered modal carries both a title and its description
// verbatim.
func TestPickerWikipediaModeRendersTitleAndDescription(t *testing.T) {
	fake := &fakeWikipediaSearcher{results: map[string][]wikipedia.SearchResult{
		"eul": {
			{Title: "Leonhard Euler", Description: "Swiss mathematician (1707-1783)"},
			{Title: "Euler's constant", Description: ""},
		},
	}}
	m := pickerWikipedia(t, fake)

	m, seqs := typePickerQuery(t, m, "eul")
	m, searches := deliverPickerTicks(t, m, seqs)

	if searches != 1 {
		t.Fatalf("dispatched %d searches, want 1", searches)
	}
	if len(fake.calls) != 1 || fake.calls[0] != "eul" {
		t.Fatalf("Search calls = %v, want exactly one call for %q", fake.calls, "eul")
	}
	if m.picker.results.mode != wikipediaMode {
		t.Fatalf("results.mode = %v, want wikipediaMode", m.picker.results.mode)
	}
	got := m.picker.results.wikipedia
	want := fake.results["eul"]
	if len(got) != len(want) {
		t.Fatalf("results.wikipedia = %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("results.wikipedia[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}

	joined := strings.Join(plainLines(m.View().Content), "\n")
	resultsWidth := computePickerLayout(80, 24).results.content().width
	wantA := pickerWikipediaRow(got[0], resultsWidth, false, false)
	wantB := pickerWikipediaRow(got[1], resultsWidth, false, false)
	if !strings.Contains(wantA, "Leonhard Euler") {
		t.Fatalf("test fixture did not produce the expected row text: %q", wantA)
	}
	if !strings.Contains(joined, wantA) {
		t.Errorf("rendered modal is missing %q\n%s", wantA, joined)
	}
	if !strings.Contains(joined, wantB) {
		t.Errorf("rendered modal is missing %q\n%s", wantB, joined)
	}

	// The description is asserted directly against pickerWikipediaRow at a
	// width wide enough not to truncate it, since the 80x24 fixture's own
	// Results column (resultsWidth above) is too narrow to show both the
	// title and the full description without rightTruncate cutting it —
	// the row's own text-selection contract, not a rendering bug.
	wide := pickerWikipediaRow(got[0], 80, false, false)
	if !strings.Contains(wide, "Swiss mathematician (1707-1783)") {
		t.Errorf("pickerWikipediaRow at width 80 = %q, want the full description visible", wide)
	}
}

// TestPickerWikipediaModeDiscardsStaleGeneration is M2.4's stale-generation
// discard rule (already covered for File mode by TestPickerDiscardsStaleResults),
// checked for Wikipedia mode's own async search: a search dispatched for a
// query the user has since typed past must not be displayed even though it
// eventually completes.
func TestPickerWikipediaModeDiscardsStaleGeneration(t *testing.T) {
	fake := &fakeWikipediaSearcher{results: map[string][]wikipedia.SearchResult{
		"cat": {{Title: "Cat", Description: "Domestic feline"}},
	}}
	m := pickerWikipedia(t, fake)

	m, seqs := typePickerQuery(t, m, "cat")
	next, cmd := m.Update(pickerSearchMsg{seq: seqs[len(seqs)-1]})
	m = concrete(t, next)
	if cmd == nil {
		t.Fatal("the settled query dispatched no search; setup is broken")
	}
	if !m.picker.searching {
		t.Error("searching = false while a dispatched search is in flight, want true")
	}

	// The user types on (past the threshold, so a second search is also
	// scheduled) before the first search's result lands.
	m, _ = typePickerQuery(t, m, "xx")

	next, _ = m.Update(cmd())
	m = concrete(t, next)

	if m.picker.results.query != "" {
		t.Errorf("displayed stale Wikipedia results for %q; prompt is now %q",
			m.picker.results.query, m.picker.input.Value())
	}
	if len(m.picker.results.wikipedia) != 0 {
		t.Errorf("results.wikipedia = %v after a stale completion, want none", m.picker.results.wikipedia)
	}
}

// TestPickerWikipediaModeFailureKeepsPreviousResults is M3.3's own "a
// failed search ... leaves the last Results in place" acceptance criterion:
// a successful search's rows survive a later search (same or different
// query) that fails, and the failure is reported through the Picker's
// status line instead.
func TestPickerWikipediaModeFailureKeepsPreviousResults(t *testing.T) {
	fake := &fakeWikipediaSearcher{
		results: map[string][]wikipedia.SearchResult{
			"cat": {{Title: "Cat", Description: "Domestic feline"}},
		},
		errs: map[string]error{
			"catapult": &wikipedia.OfflineError{Err: errors.New("connection refused")},
		},
	}
	m := pickerWikipedia(t, fake)

	m, seqs := typePickerQuery(t, m, "cat")
	m, searches := deliverPickerTicks(t, m, seqs)
	if searches != 1 || len(m.picker.results.wikipedia) != 1 {
		t.Fatalf("setup is broken: %d searches, results = %+v", searches, m.picker.results)
	}
	before := m.picker.results.wikipedia

	m, seqs = typePickerQuery(t, m, "apult")
	m, searches = deliverPickerTicks(t, m, seqs)
	if searches != 1 {
		t.Fatalf("dispatched %d searches for the failing query, want 1", searches)
	}

	if len(m.picker.results.wikipedia) != len(before) || m.picker.results.wikipedia[0] != before[0] {
		t.Errorf("results.wikipedia after a failed search = %+v, want the previous %+v kept in place",
			m.picker.results.wikipedia, before)
	}
	if m.picker.wikipediaErr == "" {
		t.Error("wikipediaErr is empty after a failed search, want the error text set")
	}
	if got := pickerStatus(m.picker, m.ripgrep, m.hideDotfiles); got != m.picker.wikipediaErr {
		t.Errorf("pickerStatus = %q, want the wikipediaErr text %q", got, m.picker.wikipediaErr)
	}
	if m.picker.searching {
		t.Error("searching = true after a failed search landed, want false")
	}
}

// TestPickerWikipediaModeErrorClearsOnNextAttempt is wikipediaErr's own
// "goes stale the instant something changes" rule: typing further after a
// failure clears the error immediately, before the next search even
// dispatches or resolves.
func TestPickerWikipediaModeErrorClearsOnNextAttempt(t *testing.T) {
	fake := &fakeWikipediaSearcher{
		errs: map[string]error{"xx": &wikipedia.OfflineError{Err: errors.New("offline")}},
	}
	m := pickerWikipedia(t, fake)

	m, seqs := typePickerQuery(t, m, "xx")
	m, searches := deliverPickerTicks(t, m, seqs)
	if searches != 1 || m.picker.wikipediaErr == "" {
		t.Fatalf("setup is broken: %d searches, wikipediaErr = %q", searches, m.picker.wikipediaErr)
	}

	next, _ := m.Update(tea.KeyPressMsg{Code: 'y', Text: "y"})
	m = concrete(t, next)

	if m.picker.wikipediaErr != "" {
		t.Errorf("wikipediaErr = %q after typing on, want cleared", m.picker.wikipediaErr)
	}
}

// TestPickerWikipediaModePreviewShowsPlaceholder is M3.3's own "the Preview
// pane shows a short placeholder" acceptance criterion: the placeholder
// renders in Wikipedia mode regardless of the selection or of whatever a
// previous File mode search left loaded in the Preview viewport.
func TestPickerWikipediaModePreviewShowsPlaceholder(t *testing.T) {
	fake := &fakeWikipediaSearcher{results: map[string][]wikipedia.SearchResult{
		"cat": {{Title: "Cat", Description: "Domestic feline"}},
	}}
	m := pickerWikipedia(t, fake)

	m, seqs := typePickerQuery(t, m, "cat")
	m, searches := deliverPickerTicks(t, m, seqs)
	if searches != 1 {
		t.Fatalf("setup is broken: %d searches", searches)
	}

	if got := pickerPreviewContent(m.picker); got != pickerWikipediaPreviewPlaceholder {
		t.Errorf("pickerPreviewContent = %q, want the placeholder %q", got, pickerWikipediaPreviewPlaceholder)
	}
}

// TestPickerWikipediaModeSearchesWithoutRipgrep is searchable's own
// "Wikipedia mode needs no ripgrep" rule: a session with no rg on PATH can
// still search Wikipedia, unlike File/Grep (TestPickerIssuesNoSearchWithoutRipgrep).
func TestPickerWikipediaModeSearchesWithoutRipgrep(t *testing.T) {
	fake := &fakeWikipediaSearcher{results: map[string][]wikipedia.SearchResult{
		"cat": {{Title: "Cat", Description: "Domestic feline"}},
	}}
	m := pickerWikipedia(t, fake)
	m = m.withRipgrep(ripgrepAvailability{reason: "ripgrep (rg) not found on PATH"})

	m, seqs := typePickerQuery(t, m, "cat")
	m, searches := deliverPickerTicks(t, m, seqs)

	if searches != 1 {
		t.Fatalf("dispatched %d searches without rg in Wikipedia mode, want 1", searches)
	}
	if len(m.picker.results.wikipedia) != 1 {
		t.Errorf("results.wikipedia = %v, want the fake's one result", m.picker.results.wikipedia)
	}
	if got := pickerStatus(m.picker, m.ripgrep, m.hideDotfiles); got != "" {
		t.Errorf("pickerStatus = %q with rg unavailable but results present, want empty", got)
	}
}
