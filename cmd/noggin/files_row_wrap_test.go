package main

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"

	"noggin/internal/tree"
)

// TestFilesNavRowTruncatesRatherThanWraps is the regression test for the bug
// report: a file/folder name too long for the pane's width must truncate to
// one physical line (rightTruncate), not word-wrap into two, the same
// lipgloss v2 Width-wraps-before-MaxWidth-truncates gotcha
// (lipgloss-v2-row-rendering-gotchas.md) TestRenderStaleRowTruncatesRatherThanWraps
// already guards for the stale row. Before the fix, every wrapped row grew
// the pane's rendered height by one line, pushing the Sources Nav and Document
// Viewer panes above it off screen (windowedViewRows counts entries, not
// physical lines, so it never caught the overflow).
func TestFilesNavRowTruncatesRatherThanWraps(t *testing.T) {
	width := 20
	longName := strings.Repeat("a-very-long-file-name", 3) + ".md"
	root := &tree.Node{
		Name: "root", Path: "/root", IsDir: true,
		Children: []*tree.Node{
			{Name: longName, Path: "/root/" + longName},
		},
	}
	f := newFiles(root).setSize(width, 5)

	view := f.view()
	lines := strings.Split(view, "\n")
	if got := len(lines); got != 5 {
		t.Fatalf("view() line count = %d, want exactly 5 (f.height), got no wrap-driven growth", got)
	}
	for i, line := range lines {
		if got := lipgloss.Width(line); got != width {
			t.Errorf("line %d width = %d, want %d (line %q)", i, got, width, line)
		}
	}
}

// TestFilesNavFlatRowTruncatesRatherThanWraps is the same regression as
// TestFilesNavRowTruncatesRatherThanWraps for renderFlatRow (S6's Recent/
// Starred flat-list rows), whose left segment (star + name) can also exceed
// the pane width on its own, independent of the right-hand label's own
// leftTruncate.
func TestFilesNavFlatRowTruncatesRatherThanWraps(t *testing.T) {
	width := 20
	longName := strings.Repeat("a-very-long-file-name", 3) + ".md"
	f := newFiles(fixtureTree()).setSize(width, 5)
	row := browseRow{name: longName, path: "/elsewhere/" + longName, rootLabel: "somewhere"}

	line := f.renderFlatRow(row, false)
	if got := lipgloss.Height(line); got != 1 {
		t.Errorf("renderFlatRow height = %d, want 1 (no wrap)", got)
	}
	if got := lipgloss.Width(line); got != width {
		t.Errorf("renderFlatRow width = %d, want %d", got, width)
	}
}
