package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"noggin/internal/tree"
)

// normalizedRows strips ANSI styling from a rendered pane and collapses it
// to a single space-joined line, trimming each row's own padding first — a
// narrow pane (e.g. the Files Nav's ~20-24 content columns in these tests,
// layout.go's navigatorPercent/minNavigatorWidth) word-wraps a longer empty
// message across more than one row (lipgloss-v2-row-rendering-gotchas.md:
// Width/Height wraps before it truncates), which would otherwise defeat a
// plain strings.Contains check against the raw multi-line view.
func normalizedRows(view string) string {
	var words []string
	for _, line := range strings.Split(ansi.Strip(view), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			words = append(words, line)
		}
	}
	return strings.Join(words, " ")
}

// TestFilesNavEveryModeScopeCombinationEmptyShowsAMessage is M4.4's own
// acceptance wording: every one of S5's six mode/scope combinations (Tree's
// one scope-less state; Recent's Active Root and All Roots; Starred's
// Active Root, All Starred and Orphans) renders a brief, non-blank message
// of its own when it has nothing to show, rather than sharing one
// placeholder across a whole mode (recentEmptyText/starredEmptyText's
// pre-M4.4 shape) or leaving the row list blank. The Tree combo's own
// message is emptyTreeText, not emptyFilesText (M4.9, TRIAGE T22): this
// fixture's Root is open (setRoot(node) below) with an empty folder under
// it, the open-but-empty case emptyTreeText names — emptyFilesText is
// reserved for no Root open at all, which TestTreeEmptyMessageDiffersFromNoRootOpen
// covers directly.
func TestFilesNavEveryModeScopeCombinationEmptyShowsAMessage(t *testing.T) {
	dir := t.TempDir()
	rootDir := mkdir(t, filepath.Join(dir, "project"))

	node, err := tree.Walk(rootDir)
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}

	base := newFilesNav().setRoot(node).setSize(40, 10)
	base = base.withRecentSource(map[string]docRecord{}, []root{{path: rootDir}}, rootDir)

	combos := []struct {
		name  string
		mode  browseMode
		scope browseScope
	}{
		{"Tree", treeBrowse, activeRootBrowseScope},
		{"Recent/Active Root", recentBrowse, activeRootBrowseScope},
		{"Recent/All Roots", recentBrowse, allRootsBrowseScope},
		{"Starred/Active Root", starredBrowse, activeRootBrowseScope},
		{"Starred/All Starred", starredBrowse, allStarredBrowseScope},
		{"Starred/Orphans", starredBrowse, orphansBrowseScope},
	}

	seen := map[string]string{}
	for _, c := range combos {
		f := base
		f.mode = c.mode
		f.scope = c.scope

		view := f.view()
		rowsPortion := normalizedRows(f.rowsView(f.height - 1)) // exclude the mode/scope row

		if rowsPortion == "" {
			t.Errorf("%s: empty view, want a message rather than a blank pane:\n%q", c.name, view)
			continue
		}

		var want string
		if c.mode == treeBrowse {
			want = emptyTreeText
		} else {
			want = flatEmptyText(c.mode, c.scope)
		}
		if !strings.Contains(rowsPortion, want) {
			t.Errorf("%s: rows = %q, want it to contain %q", c.name, rowsPortion, want)
		}

		if prior, ok := seen[want]; ok && prior != c.name {
			t.Errorf("%s and %s share the same empty message %q, want each combination's own", c.name, prior, want)
		}
		seen[want] = c.name
	}
}

// TestTreeEmptyMessageDiffersFromNoRootOpen is M4.9's own acceptance
// wording directly (TRIAGE T22): Tree mode's empty state has two distinct
// causes — no Root open at all, and a Root open with nothing visible under
// it (an empty folder, or one whose every top-level entry the tree walk
// excluded) — and each must show its own message, not share emptyFilesText
// across both the way rowsView did before this task.
func TestTreeEmptyMessageDiffersFromNoRootOpen(t *testing.T) {
	noRoot := newFilesNav().setSize(40, 10)
	if got := normalizedRows(noRoot.rowsView(noRoot.height)); !strings.Contains(got, emptyFilesText) {
		t.Errorf("no Root open: rows = %q, want it to contain %q", got, emptyFilesText)
	}

	dir := t.TempDir()
	empty := mkdir(t, filepath.Join(dir, "empty-project"))
	node, err := tree.Walk(empty)
	if err != nil {
		t.Fatalf("tree.Walk: %v", err)
	}
	openButEmpty := newFilesNav().setRoot(node).setSize(40, 10)
	if got := normalizedRows(openButEmpty.rowsView(openButEmpty.height)); !strings.Contains(got, emptyTreeText) {
		t.Errorf("Root open but empty: rows = %q, want it to contain %q", got, emptyTreeText)
	}

	if emptyFilesText == emptyTreeText {
		t.Fatal("emptyFilesText and emptyTreeText are identical, want distinct wording for the two causes")
	}
	if strings.Contains(normalizedRows(openButEmpty.rowsView(openButEmpty.height)), emptyFilesText) {
		t.Error("Root open but empty: rows unexpectedly contain the no-Root-open message")
	}
	if strings.Contains(normalizedRows(noRoot.rowsView(noRoot.height)), emptyTreeText) {
		t.Error("no Root open: rows unexpectedly contain the open-but-empty message")
	}
}
