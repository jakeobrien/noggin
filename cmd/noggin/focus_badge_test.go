package main

import (
	"strings"
	"testing"
)

// TestPaneFocusBadgesShowAtNormalWidthRegardlessOfFocus is M7.1's acceptance
// criterion: at a normal terminal width every pane's title row carries its
// own focus-key badge ([r], [f], [o], [^o]), whether or not that pane
// currently holds Focus. Focus is parked on filesPane so the other three
// checks all cover unfocused panes.
func TestPaneFocusBadgesShowAtNormalWidthRegardlessOfFocus(t *testing.T) {
	m := sized(t, newModel().openFolder(fixtureTree()), 80, 24)
	m = m.withViewerFor(dv2Pane, newViewer())
	m = sized(t, m, 80, 24) // re-lay-out now that DV2 exists
	m.focus = filesPane

	l := computeLayout(m.width, m.height, m.navigatorHidden, m.stale)
	lines := strings.Split(m.View().Content, "\n")

	for _, want := range []struct {
		name  string
		row   int
		pane  pane
		badge string
	}{
		{"rootsPane", 1, rootsPane, "[r]"},
		{"filesPane", l.roots.height + 1, filesPane, "[f]"},
		{"dv1Pane", 1, dv1Pane, "[o]"},
		{"dv2Pane", 1, dv2Pane, "[^o]"},
	} {
		t.Run(want.name, func(t *testing.T) {
			if got := want.pane.focusBadge(); got != want.badge {
				t.Fatalf("focusBadge() = %q, want %q", got, want.badge)
			}
			line := lines[want.row]
			if !strings.Contains(line, want.pane.title()) {
				t.Errorf("row %d = %q, want it to contain title %q", want.row, line, want.pane.title())
			}
			if !strings.Contains(line, want.badge) {
				t.Errorf("row %d = %q, want it to contain badge %q", want.row, line, want.badge)
			}
			if want.pane != filesPane && m.focus == want.pane {
				t.Fatalf("test bug: %s unexpectedly holds Focus", want.name)
			}
		})
	}
}

// TestRenderPaneDropsTheBadgeNotTheTitleWhenTooNarrow covers the other half
// of M7.1's acceptance criterion: a pane too narrow for title plus badge
// drops the badge, not the title. rootsPane's title ("Sources", 7 columns)
// plus its badge ("[r]", 3 columns) needs 11 interior columns (7 + 1
// separator + 3); interior width is r.width - 2*paneBorderWidth.
func TestRenderPaneDropsTheBadgeNotTheTitleWhenTooNarrow(t *testing.T) {
	m := sized(t, newModel(), 80, 24)

	fits := m.renderPane(rootsPane, rect{width: 13, height: 3}, "") // interior 11: exact fit
	if !strings.Contains(fits, "Sources") {
		t.Errorf("at exact-fit width, rendered pane = %q, want it to contain the title", fits)
	}
	if !strings.Contains(fits, "[r]") {
		t.Errorf("at exact-fit width, rendered pane = %q, want it to contain the badge", fits)
	}

	tooNarrow := m.renderPane(rootsPane, rect{width: 12, height: 3}, "") // interior 10: title only
	if !strings.Contains(tooNarrow, "Sources") {
		t.Errorf("at too-narrow width, rendered pane = %q, want the title kept", tooNarrow)
	}
	if strings.Contains(tooNarrow, "[r]") {
		t.Errorf("at too-narrow width, rendered pane = %q, want the badge dropped", tooNarrow)
	}
}

// TestTitleRowDropsTheBadgeAtExactlyOneColumnShort pins titleRow's own
// boundary directly, in case a future title/badge change shifts it: one
// column short of "title + separator + badge" drops the badge; exactly
// enough keeps both.
func TestTitleRowDropsTheBadgeAtExactlyOneColumnShort(t *testing.T) {
	title, badge := "Sources", "[r]" // 7 + 1 + 3 = 11
	if got := titleRow(title, badge, 11); got != title+" "+badge {
		t.Errorf("titleRow at exact width = %q, want title and badge joined by one space of padding", got)
	}
	if got := titleRow(title, badge, 10); got != title {
		t.Errorf("titleRow one column short = %q, want the bare title with the badge dropped", got)
	}
}
