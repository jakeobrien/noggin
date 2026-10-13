package main

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

const (
	thickCorner   = "┏"
	roundedCorner = "╭"
)

// titleBarSGR is the focused pane's title-bar background (Theme.Accent).
var titleBarSGR = "48;" + sgrColor(theme.Accent)

// TestFocusedPaneHasThickBorderAndTitleBar is S11's acceptance criterion:
// at every focus target, the focused pane — and only it — carries the thick
// border and the accent title bar; the other panes keep the rounded border
// and a plain title.
func TestFocusedPaneHasThickBorderAndTitleBar(t *testing.T) {
	m := sized(t, newModel().openFolder(fixtureTree()), 80, 24)
	m = m.withViewerFor(dv2Pane, newViewer())
	m = sized(t, m, 80, 24) // re-lay-out now that DV2 exists

	l := computeLayout(m.width, m.height, m.navigatorHidden, m.stale)
	dv1Rect, dv2Rect := l.splitViewer(true)

	render := func(m model, p pane) string {
		switch p {
		case rootsPane:
			return m.renderPane(p, l.roots, m.roots.view())
		case filesPane:
			return m.renderPane(p, l.files, m.files.view())
		case dv2Pane:
			content := m.composeViewerContent(dv2Pane, dv2Rect.content().width, m.viewerFor(dv2Pane).view())
			return m.renderPane(p, dv2Rect, content)
		default:
			content := m.composeViewerContent(dv1Pane, dv1Rect.content().width, m.dv1.view())
			return m.renderPane(p, dv1Rect, content)
		}
	}

	panes := []pane{rootsPane, filesPane, dv1Pane, dv2Pane}
	for _, focusTarget := range panes {
		t.Run(focusTarget.title()+" "+focusTarget.focusBadge(), func(t *testing.T) {
			m := m
			m.focus = focusTarget
			for _, p := range panes {
				rendered := render(m, p)
				focused := p == focusTarget
				if got := strings.Contains(rendered, thickCorner); got != focused {
					t.Errorf("pane %s (focus on %s): thick border present = %v, want %v", p.title(), focusTarget.title(), got, focused)
				}
				if got := strings.Contains(rendered, roundedCorner); got == focused {
					t.Errorf("pane %s (focus on %s): rounded border present = %v, want %v", p.title(), focusTarget.title(), got, !focused)
				}
				if got := strings.Contains(rendered, titleBarSGR); got != focused {
					t.Errorf("pane %s (focus on %s): title bar present = %v, want %v", p.title(), focusTarget.title(), got, focused)
				}
			}
		})
	}
}

// TestFocusedTitleBarKeepsBadgeAndFillsRow: the bar spans the whole title
// row and the focus-key badge survives inside it.
func TestFocusedTitleBarKeepsBadgeAndFillsRow(t *testing.T) {
	m := sized(t, newModel().openFolder(fixtureTree()), 80, 24)
	l := computeLayout(m.width, m.height, m.navigatorHidden, m.stale)
	rendered := m.renderPane(filesPane, l.files, m.files.view())

	titleLine := strings.Split(rendered, "\n")[1]
	if !strings.Contains(titleLine, filesPane.focusBadge()) {
		t.Errorf("focused title line = %q, want the %s badge kept", titleLine, filesPane.focusBadge())
	}
	if got, want := lipgloss.Width(titleLine), l.files.width; got != want {
		t.Errorf("focused title line width = %d, want %d (the full pane)", got, want)
	}
}

// TestFocusedPeekTitleStaysInverted: a Peeking viewer's inverted header
// (S7) is not replaced by the focus title bar, so Peek stays recognisable.
func TestFocusedPeekTitleStaysInverted(t *testing.T) {
	m := sized(t, newModel().openFolder(fixtureTree()), 80, 24)
	m.dv1 = m.dv1.startPeek(document{path: "/tmp/p.md", src: []byte("# Peeked\n")})
	m.focus = dv1Pane
	l := computeLayout(m.width, m.height, m.navigatorHidden, m.stale)
	dv1Rect, _ := l.splitViewer(false)
	rendered := m.renderPane(dv1Pane, dv1Rect, m.dv1.view())

	if strings.Contains(rendered, titleBarSGR) {
		t.Errorf("focused Peeking DV1 carries the title bar; want Peek's inverted header kept: %q", rendered)
	}
	if !strings.Contains(strings.Split(rendered, "\n")[1], "7m") {
		t.Errorf("focused Peeking DV1's title is not reverse-video: %q", rendered)
	}
}
