package main

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
)

const fixturePath = "testdata/fixture.md"

// frontmatterLines are the fixture's frontmatter as its Frontmatter Card shows
// it: each key, then its value, in source order.
var frontmatterLines = []string{
	"title", "Fixture Document",
	"tags", "noggin, fixture",
	"draft", "true",
}

// fixtureStructure is what the fixture's markdown constructs look like once
// rendered: a heading, a table, a code fence, nested and task lists, and a
// blockquote. Colour is not asserted — under `go test` there is no TTY, so
// Glamour uses its no-TTY style and lipgloss emits no escapes
// (docs/dev/pty-harness.md). Styling is verified in a pty instead.
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

func TestFenceFrontmatter(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want string
	}{
		{"leading block", "---\ntitle: x\n---\n# H\n", "```yaml\ntitle: x\n```\n# H\n"},
		{"crlf", "---\r\ntitle: x\r\n---\r\n# H\r\n", "```yaml\r\ntitle: x\r\n```\r\n# H\r\n"},
		{"yaml document-end closer", "---\ntitle: x\n...\n# H\n", "```yaml\ntitle: x\n```\n# H\n"},
		{"toml", "+++\ntitle = 'x'\n+++\n# H\n", "```toml\ntitle = 'x'\n```\n# H\n"},
		{"toml block does not close on ---", "+++\na = 1\n---\n+++\n# H\n", "```toml\na = 1\n---\n```\n# H\n"},
		{"byte order mark", "\ufeff---\ntitle: x\n---\n# H\n", "```yaml\ntitle: x\n```\n# H\n"},
		{"trailing spaces on fences", "--- \ntitle: x\n---\t\n# H\n", "```yaml\ntitle: x\n```\n# H\n"},
		{"closer at end of file", "---\ntitle: x\n---", "```yaml\ntitle: x\n```"},
		// An empty block becomes blank lines, not an empty code box.
		{"empty block", "---\n---\n# H\n", "\n\n# H\n"},
		{"blank-only block", "---\n\n---\n# H\n", "\n\n\n# H\n"},
		// A backtick fence in a value must not close the code block early.
		{"backticks in the block", "---\nx: |\n  ```\n---\n", "````yaml\nx: |\n  ```\n````\n"},
		{"no frontmatter", "# H\n\ntext\n", "# H\n\ntext\n"},
		{"no frontmatter keeps its byte order mark", "\ufeff# H\n", "\ufeff# H\n"},
		{"unclosed block", "---\ntitle: x\n", "---\ntitle: x\n"},
		{"four dashes is a thematic break", "----\ntitle: x\n----\n", "----\ntitle: x\n----\n"},
		{"blank line before opener", "\n---\ntitle: x\n---\n", "\n---\ntitle: x\n---\n"},
		// Glow slices between the first two fences found anywhere, which eats
		// the head of a document whose thematic breaks are `---`.
		{"thematic breaks mid-document", "# H\n\n---\n\ntext\n\n---\n\nmore\n", "# H\n\n---\n\ntext\n\n---\n\nmore\n"},
		{"fence-like text indented", "text\n\n    ---\n    ---\n", "text\n\n    ---\n    ---\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := string(fenceFrontmatter([]byte(tc.in)))
			if got != tc.want {
				t.Errorf("fenceFrontmatter(%q) = %q, want %q", tc.in, got, tc.want)
			}
			// Line numbers must survive the rewrite: resolveGrepLanding maps
			// ripgrep's raw-file lines straight onto this text (ADR-0002).
			if in, out := strings.Count(tc.in, "\n"), strings.Count(got, "\n"); in != out {
				t.Errorf("fenceFrontmatter changed the line count from %d to %d", in, out)
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

func TestRenderedFixtureShowsFrontmatterAndKeepsStructure(t *testing.T) {
	const width = 54
	out, err := renderMarkdown(fixtureDoc(t).src, width)
	if err != nil {
		t.Fatalf("renderMarkdown: %v", err)
	}

	// In order, ahead of the document's own title: as a card, not
	// CommonMark's thematic break + paragraph + setext H2, nor the raw YAML.
	if strings.Contains(out, "title: Fixture Document") {
		t.Error("rendered document shows raw YAML, want the Frontmatter Card")
	}
	prev := -1
	for _, line := range append(frontmatterLines, fixtureStructure["heading"]) {
		i := strings.Index(out, line)
		if i < 0 {
			t.Errorf("rendered document is missing %q", line)
			continue
		}
		if i < prev {
			t.Errorf("%q is rendered out of order", line)
		}
		prev = i
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
		if !strings.Contains(view, line) {
			t.Errorf("view does not show frontmatter %q", line)
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
	if got := m.paneTitle(dv1Pane); got != dv1Pane.title() {
		t.Errorf("viewer title with no document = %q, want %q", got, dv1Pane.title())
	}
}

func TestScrollActionsMoveTheViewport(t *testing.T) {
	base := openedFixture(t, 80, 24)
	page := base.dv1.height()
	if total := base.dv1.lineCount(); total <= page {
		t.Fatalf("fixture renders %d lines into a %d-line viewport; nothing to scroll", total, page)
	}
	bottom := base.dv1.lineCount() - page

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
		// M8.1: pgup/pgdn (and their ctrl+b/ctrl+f alt keys) now land on the
		// same offset as ctrl+u/ctrl+d in a Document Viewer, not a full page.
		{"page down", keys.PageDown, 0, page / 2},
		{"page up", keys.PageUp, page, page - page/2},
		{"bottom", keys.Bottom, 0, bottom},
		{"top", keys.Top, 5, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Every key the action is bound to must perform it.
			for _, k := range tc.binding.Keys() {
				m := base
				m.dv1 = m.dv1.withOffset(tc.start)
				m = press(t, m, k)
				if got := m.dv1.offset(); got != tc.want {
					t.Errorf("%q from offset %d left offset %d, want %d", k, tc.start, got, tc.want)
				}
			}
		})
	}

	// M8.1's own acceptance wording: pgdn and ctrl+d must land on the exact
	// same offset from the same start, not merely the same formula.
	t.Run("pgdn lands on the same offset as ctrl+d", func(t *testing.T) {
		m := base
		m.dv1 = m.dv1.withOffset(0)
		afterPgdn := press(t, m, keys.PageDown.Keys()[0]).dv1.offset()

		m = base
		m.dv1 = m.dv1.withOffset(0)
		afterCtrlD := press(t, m, keys.HalfPageDown.Keys()[0]).dv1.offset()

		if afterPgdn != afterCtrlD {
			t.Errorf("pgdn landed at offset %d, ctrl+d at %d, want equal", afterPgdn, afterCtrlD)
		}
	})

	// The last screen is reachable and is not the first screen.
	m := press(t, base, keys.Bottom.Keys()[0])
	if !m.dv1.atBottom() {
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
	if m.focus != dv1Pane {
		t.Fatalf("focus after opening a document = %v, want %v", m.focus, dv1Pane)
	}
	// Tab-cycling is gone (S5, C3/M1.1); FocusRoots is one of its
	// direct-jump replacements for moving Focus off the viewer.
	m = press(t, m, keys.FocusRoots.Keys()[0])
	if m.focus == dv1Pane {
		t.Fatal("focus did not leave the viewer")
	}

	for _, k := range keys.LineDown.Keys() {
		next := press(t, m, k)
		if next.dv1.offset() != 0 {
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
		want := computeLayout(tc.width, tc.height, false, false).viewer.content()
		if got := tc.m.dv1.renderedWidth; got != want.width {
			t.Errorf("%s: rendered at width %d, want %d", tc.name, got, want.width)
		}
		// M8.2/S9: openedFixture's document is open, so the location line
		// costs the viewport one row of want.height's own content box.
		if got := tc.m.dv1.height(); got != want.height-1 {
			t.Errorf("%s: viewport height = %d, want %d", tc.name, got, want.height-1)
		}
	}

	if narrow.dv1.lineCount() <= wide.dv1.lineCount() {
		t.Errorf("narrowing the viewer did not re-wrap the document: %d lines at 120 columns, %d at 80",
			wide.dv1.lineCount(), narrow.dv1.lineCount())
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
		if m.dv1.atBottom() {
			return
		}
		m = press(t, m, keys.PageDown.Keys()[0])
	}
}
