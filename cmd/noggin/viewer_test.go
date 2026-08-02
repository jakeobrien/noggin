package main

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
)

const fixturePath = "testdata/fixture.md"

// frontmatterLines are the fixture's frontmatter, none of which may reach the
// screen (S7, D5).
var frontmatterLines = []string{
	"title: Fixture Document",
	"tags: [noggin, fixture]",
	"draft: true",
}

// fixtureStructure is what the fixture's markdown constructs look like once
// rendered: a heading, a table, a code fence, nested and task lists, and a
// blockquote. Colour is not asserted — under `go test` there is no TTY, so
// Glamour uses its no-TTY style and lipgloss emits no escapes
// (.charm/knowledge/pty-verification.md). Styling is verified in a pty instead.
var fixtureStructure = map[string]string{
	"heading":     "Reading Notes",
	"table row":   "Shown in the pane title",
	"code fence":  "os.ReadFile(path)",
	"nested list": "Stored as plain text, hand-editable",
	"task list":   "Find in document",
	"blockquote":  "Notes are not a database",
	"tail":        "The final line of the fixture document.",
}

func fixtureDoc(t *testing.T) document {
	t.Helper()
	d, err := loadDocument(fixturePath)
	if err != nil {
		t.Fatalf("loading %s: %v", fixturePath, err)
	}
	return d
}

// openedFixture is a model of the given size with the fixture document open.
func openedFixture(t *testing.T, width, height int) model {
	t.Helper()
	return sized(t, newModel().openDocument(fixtureDoc(t)), width, height)
}

func TestStripFrontmatter(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want string
	}{
		{"leading block", "---\ntitle: x\n---\n# H\n", "# H\n"},
		{"crlf", "---\r\ntitle: x\r\n---\r\n# H\r\n", "# H\r\n"},
		{"empty block", "---\n---\n# H\n", "# H\n"},
		{"trailing spaces on fences", "--- \ntitle: x\n---\t\n# H\n", "# H\n"},
		{"no frontmatter", "# H\n\ntext\n", "# H\n\ntext\n"},
		{"unclosed block", "---\ntitle: x\n", "---\ntitle: x\n"},
		// Glow slices between the first two fences found anywhere, which eats
		// the head of a document whose thematic breaks are `---`.
		{"thematic breaks mid-document", "# H\n\n---\n\ntext\n\n---\n\nmore\n", "# H\n\n---\n\ntext\n\n---\n\nmore\n"},
		{"fence-like text indented", "text\n\n    ---\n    ---\n", "text\n\n    ---\n    ---\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := string(stripFrontmatter([]byte(tc.in))); got != tc.want {
				t.Errorf("stripFrontmatter(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestExpandTabs(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want string
	}{
		{"leading tab", "\tx", "    x"},
		{"two leading tabs", "\t\tx", "        x"},
		{"partial stop", "ab\tx", "ab  x"},
		{"full stop", "abcd\tx", "abcd    x"},
		{"column resets each line", "ab\tx\n\ty", "ab  x\n    y"},
		{"no tabs is untouched", "plain text\n", "plain text\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := string(expandTabs([]byte(tc.in))); got != tc.want {
				t.Errorf("expandTabs(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestRenderedFixtureHidesFrontmatterAndKeepsStructure(t *testing.T) {
	const width = 54
	out, err := renderMarkdown(fixtureDoc(t).src, width)
	if err != nil {
		t.Fatalf("renderMarkdown: %v", err)
	}

	for _, line := range frontmatterLines {
		if strings.Contains(out, line) {
			t.Errorf("rendered document contains frontmatter %q", line)
		}
	}
	for name, want := range fixtureStructure {
		if !strings.Contains(out, want) {
			t.Errorf("rendered document is missing its %s (%q)", name, want)
		}
	}
	for i, line := range strings.Split(out, "\n") {
		if got := lipgloss.Width(line); got > width {
			t.Errorf("rendered line %d is %d cells wide, want at most %d: %q", i, got, width, line)
		}
	}
	// A tab's width is the terminal's business, not the string's, so a rendered
	// line containing one is a line whose measured width is a lie.
	if n := strings.Count(out, "\t"); n != 0 {
		t.Errorf("rendered document contains %d literal tabs, want 0", n)
	}
}

func TestViewerShowsTheOpenDocument(t *testing.T) {
	const w, h = 80, 24
	m := openedFixture(t, w, h)
	view := m.View().Content

	if got := lipgloss.Height(view); got != h {
		t.Errorf("view height = %d, want %d", got, h)
	}
	for i, line := range strings.Split(view, "\n") {
		if got := lipgloss.Width(line); got != w {
			t.Errorf("line %d width = %d, want %d", i, got, w)
		}
	}

	// The document's opening is on screen and the pane says which file it is.
	if !strings.Contains(view, fixtureStructure["heading"]) {
		t.Errorf("view does not show the document's first heading %q", fixtureStructure["heading"])
	}
	if !strings.Contains(view, "fixture.md") {
		t.Error("view does not name the open document")
	}
	for _, line := range frontmatterLines {
		if strings.Contains(view, line) {
			t.Errorf("view contains frontmatter %q", line)
		}
	}
	if strings.Contains(view, emptyViewerText) {
		t.Error("view shows the empty state with a document open")
	}
}

func TestEmptyStateWithNoDocumentOpen(t *testing.T) {
	m := sized(t, newModel(), 80, 24)
	if !strings.Contains(m.View().Content, emptyViewerText) {
		t.Errorf("view with no document open does not contain %q", emptyViewerText)
	}
	if got := m.paneTitle(viewerPane); got != viewerPane.title() {
		t.Errorf("viewer title with no document = %q, want %q", got, viewerPane.title())
	}
}

func TestScrollActionsMoveTheViewport(t *testing.T) {
	base := openedFixture(t, 80, 24)
	page := base.viewer.height()
	if total := base.viewer.lineCount(); total <= page {
		t.Fatalf("fixture renders %d lines into a %d-line viewport; nothing to scroll", total, page)
	}
	bottom := base.viewer.lineCount() - page

	for _, tc := range []struct {
		name    string
		binding key.Binding
		start   int
		want    int
	}{
		{"line down", keys.LineDown, 0, 1},
		{"line up", keys.LineUp, 5, 4},
		{"half page down", keys.HalfPageDown, 0, page / 2},
		{"half page up", keys.HalfPageUp, page, page - page/2},
		{"page down", keys.PageDown, 0, page},
		{"page up", keys.PageUp, page, 0},
		{"bottom", keys.Bottom, 0, bottom},
		{"top", keys.Top, 5, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Every key the action is bound to must perform it.
			for _, k := range tc.binding.Keys() {
				m := base
				m.viewer = m.viewer.withOffset(tc.start)
				m = press(t, m, k)
				if got := m.viewer.offset(); got != tc.want {
					t.Errorf("%q from offset %d left offset %d, want %d", k, tc.start, got, tc.want)
				}
			}
		})
	}

	// The last screen is reachable and is not the first screen.
	m := press(t, base, keys.Bottom.Keys()[0])
	if !m.viewer.atBottom() {
		t.Error("bottom key did not reach the bottom of the document")
	}
	if !strings.Contains(m.View().Content, fixtureStructure["tail"]) {
		t.Errorf("view at the bottom does not show %q", fixtureStructure["tail"])
	}
	if m.View().Content == base.View().Content {
		t.Error("scrolling to the bottom did not change the view")
	}
}

func TestScrollKeysOnlyActWhenTheViewerHasFocus(t *testing.T) {
	m := openedFixture(t, 80, 24)
	if m.focus != viewerPane {
		t.Fatalf("focus after opening a document = %v, want %v", m.focus, viewerPane)
	}
	m = press(t, m, keys.PrevPane.Keys()[0])
	if m.focus == viewerPane {
		t.Fatal("focus did not leave the viewer")
	}

	for _, k := range keys.LineDown.Keys() {
		next := press(t, m, k)
		if next.viewer.offset() != 0 {
			t.Errorf("%q scrolled the viewer while %v had focus", k, m.focus)
		}
	}
}

func TestViewerRewrapsOnResize(t *testing.T) {
	wide := openedFixture(t, 120, 40)
	narrow := sized(t, wide, 80, 24)

	for _, tc := range []struct {
		name          string
		m             model
		width, height int
	}{
		{"wide", wide, 120, 40},
		{"narrow", narrow, 80, 24},
	} {
		want := computeLayout(tc.width, tc.height, false).viewer.content()
		if got := tc.m.viewer.renderedWidth; got != want.width {
			t.Errorf("%s: rendered at width %d, want %d", tc.name, got, want.width)
		}
		if got := tc.m.viewer.height(); got != want.height {
			t.Errorf("%s: viewport height = %d, want %d", tc.name, got, want.height)
		}
	}

	if narrow.viewer.lineCount() <= wide.viewer.lineCount() {
		t.Errorf("narrowing the viewer did not re-wrap the document: %d lines at 120 columns, %d at 80",
			wide.viewer.lineCount(), narrow.viewer.lineCount())
	}
}

// TestFrameStaysWithinTheTerminalWhileScrolling walks the whole fixture a page
// at a time and holds the frame to exactly the terminal's size at every stop.
//
// This is the assertion that catches content whose measured width is a lie:
// lipgloss expands a literal tab to four spaces when it renders, so a
// tab-bearing line the viewport sliced as three cells shorter than it draws
// wraps inside the pane box and pushes the frame a row past the terminal. A
// single screenful cannot see it — the offending row has to be scrolled to.
func TestFrameStaysWithinTheTerminalWhileScrolling(t *testing.T) {
	const w, h = 80, 24
	m := openedFixture(t, w, h)
	for step := 0; ; step++ {
		view := m.View().Content
		if got := lipgloss.Height(view); got != h {
			t.Fatalf("step %d: view height = %d, want %d", step, got, h)
		}
		for i, line := range strings.Split(view, "\n") {
			if got := lipgloss.Width(line); got != w {
				t.Fatalf("step %d row %d: width = %d, want %d: %q", step, i, got, w, line)
			}
		}
		if m.viewer.atBottom() {
			return
		}
		m = press(t, m, keys.PageDown.Keys()[0])
	}
}
