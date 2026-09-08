package main

import (
	"reflect"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"noggin/internal/wikipedia"
)

// pickerWikipediaWithDocs is pickerWithWikipediaActive
// (picker_wikipedia_mode_test.go) plus a resolvable S1 document store, so
// this file's star tests can round-trip a real docRecord write/persist —
// pickerWithWikipediaActive's own newModel() has no docsPath, which would
// make every withDocRecord decline with "no resolvable data directory".
func pickerWikipediaWithDocs(t *testing.T, storePath string) model {
	t.Helper()
	m := sized(t, newModel().withDocs(storePath, loadDocStore(storePath)), 80, 24)
	m = m.activateWikipediaSource()
	m, _ = m.openPicker()
	return m
}

// TestPickerStarTogglesSelectedWikipediaResult is M3.4's own acceptance
// criterion: the Picker's star key on a Wikipedia Result toggles the
// article's record in the Document Store (keyed by wikipedia.ArticleURL,
// M1.5/M4.1's "the store's key is the article's canonical URL", derived
// from the title since a title-search SearchResult carries no URL of its
// own — see togglePickerStar's own doc comment), stamps Article/Title so
// the record reads like one M4.1 would later write, and — mirroring
// TestPickerStarTogglesSelectedFileResult's own "leaves the Result in
// place" assertion — never touches p.results or p.selected.
func TestPickerStarTogglesSelectedWikipediaResult(t *testing.T) {
	dir := t.TempDir()
	storePath := dir + "/documents"
	m := pickerWikipediaWithDocs(t, storePath)

	m.picker.input.SetValue("eul")
	result := wikipedia.SearchResult{Title: "Leonhard Euler", Description: "Swiss mathematician (1707-1783)"}
	m.picker.results = pickerResults{
		query:     "eul",
		mode:      wikipediaMode,
		wikipedia: []wikipedia.SearchResult{result},
	}
	m.picker.selected = 0
	before := m.picker

	key := wikipedia.ArticleURL(result.Title)

	m = press(t, m, keys.PickerStar.Keys()[0])

	got := m.docRecord(key)
	if !got.Starred || !got.Article || got.Title != result.Title {
		t.Errorf("docRecord(%q) after Picker star = %+v, want Starred/Article/Title=%q", key, got, result.Title)
	}

	m = press(t, m, keys.PickerStar.Keys()[0])
	got = m.docRecord(key)
	if got.Starred {
		t.Errorf("docRecord(%q) after un-starring = %+v, want Starred cleared", key, got)
	}
	// S4/M1.5: un-starring an article record leaves its pointer fields
	// alone, the same way a file record's own fields never revert to "no
	// record" on un-star.
	if !got.Article || got.Title != result.Title {
		t.Errorf("docRecord(%q) after un-starring = %+v, want Article/Title preserved", key, got)
	}

	if !reflect.DeepEqual(m.picker.results, before.results) {
		t.Errorf("picker.results after star/un-star round-trip = %+v, want unchanged %+v", m.picker.results, before.results)
	}
	if m.picker.selected != before.selected {
		t.Errorf("picker.selected after star/un-star round-trip = %d, want unchanged %d", m.picker.selected, before.selected)
	}

	restarted := newModel().withDocs(storePath, loadDocStore(storePath))
	if got := restarted.docRecord(key); got.Starred {
		t.Errorf("docRecord(%q) after restart = %+v, want unstarred", key, got)
	} else if !got.Article || got.Title != result.Title {
		t.Errorf("docRecord(%q) after restart = %+v, want the article pointer preserved", key, got)
	}
}

// TestPickerStarNoOpWithNothingSelectedInWikipediaMode mirrors
// TestPickerStarNoOpWithNothingSelected for wikipediaMode: no search has
// run yet (zero-value results), so there is no Result to act on — a
// silent no-op, not a message, a crash, or a spurious docs entry.
func TestPickerStarNoOpWithNothingSelectedInWikipediaMode(t *testing.T) {
	dir := t.TempDir()
	storePath := dir + "/documents"
	m := pickerWikipediaWithDocs(t, storePath)

	m = press(t, m, keys.PickerStar.Keys()[0])

	if m.message != "" {
		t.Errorf("Picker star with nothing selected: want silent no-op, got message %q", m.message)
	}
	if len(m.docs) != 0 {
		t.Errorf("docs after Picker star with nothing selected = %+v, want untouched", m.docs)
	}
}

// TestPickerStarTogglingUpdatesTheRenderedMarkerLiveInWikipediaMode is
// D72's live-marker rule (TestPickerStarTogglingUpdatesTheRenderedMarkerLive,
// TestPickerStarTogglingUpdatesTheRenderedMarkerLiveInGrepMode), checked for
// wikipediaMode's own row (pickerWikipediaRow, M3.4): toggling PickerStar
// shows the marker in the very next View(), with no separate refresh step.
func TestPickerStarTogglingUpdatesTheRenderedMarkerLiveInWikipediaMode(t *testing.T) {
	dir := t.TempDir()
	storePath := dir + "/documents"
	m := pickerWikipediaWithDocs(t, storePath)

	m.picker.input.SetValue("eul")
	m.picker.results = pickerResults{
		query: "eul",
		mode:  wikipediaMode,
		wikipedia: []wikipedia.SearchResult{
			{Title: "Leonhard Euler", Description: "Swiss mathematician (1707-1783)"},
		},
	}
	m.picker.selected = 0

	resultRow := func(content string) string {
		for _, l := range strings.Split(content, "\n") {
			if strings.Contains(l, "Leonhard Euler") {
				return l
			}
		}
		return ""
	}

	before := ansi.Strip(m.View().Content)
	if row := resultRow(before); row == "" || strings.Contains(row, "\u2605") {
		t.Fatalf("Results row before starring = %q, want the article row present with no star marker\n%s", row, before)
	}

	m = press(t, m, keys.PickerStar.Keys()[0])

	after := ansi.Strip(m.View().Content)
	if row := resultRow(after); !strings.Contains(row, "\u2605") {
		t.Errorf("Results row after starring the selected Result = %q, want it to carry the \"\u2605\" marker\n%s", row, after)
	}

	m = press(t, m, keys.PickerStar.Keys()[0])
	cleared := ansi.Strip(m.View().Content)
	if row := resultRow(cleared); strings.Contains(row, "\u2605") {
		t.Errorf("Results row after un-starring = %q, want the star marker gone\n%s", row, cleared)
	}
}
