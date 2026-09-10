package main

import (
	"strings"
	"testing"
)

// TestTintPaneBackgroundInjectsAfterEveryReset is tintPaneBackground's own
// unit coverage (style.go): the tint must reach a string built from several
// independently-closed Style.Render runs, the way a pane's real content
// (many rows, each its own Style.Render call) always is — not just the
// first run, which a single wrapping Style.Render would give for free.
func TestTintPaneBackgroundInjectsAfterEveryReset(t *testing.T) {
	// Two rows, each rendered the way rootsnav.go/files.go build one: a
	// styled run (here just bold) that closes with lipgloss's own bare
	// full reset before the next row starts.
	row := "\x1b[1mrow one\x1b[m\nrow two"
	got := tintPaneBackground(row)

	wantOccurrences := 2 // once before "row one", once after its reset
	if n := strings.Count(got, unfocusedTintSGR); n != wantOccurrences {
		t.Fatalf("tintPaneBackground(%q) = %q, want the tint injected %d times, got %d", row, got, wantOccurrences, n)
	}
	if !strings.HasPrefix(got, unfocusedTintSGR) {
		t.Errorf("tintPaneBackground(%q) = %q, want it to start with the tint", row, got)
	}
	if !strings.Contains(got, "\x1b[m\n"+unfocusedTintSGR+"row two") {
		t.Errorf("tintPaneBackground(%q) = %q, want the tint re-asserted right after the newline, ahead of the second row's plain text — not before the newline itself, which would carry the tint across it as ambient state even if the row that follows turns out to be reverse-video (a real terminal renderer serializes SGR from cumulative cell state, not from where this function grouped its escapes)", row, got)
	}
}

// TestTintPaneBackgroundLeavesReverseVideoAlone is the reverse-video half:
// a selected row (S5) or Peek's inverted title (S7) swaps whatever
// foreground/background are already active, so splicing the tint into a
// reverse-video run would make the tint itself the swapped-in text colour
// instead of the terminal's own default pair — visibly changing the
// selection bar / Peek header depending on Focus, which S11 never asks
// for. tintPaneBackground must never inject inside one.
func TestTintPaneBackgroundLeavesReverseVideoAlone(t *testing.T) {
	reversed := "\x1b[1;7mselected\x1b[m"
	got := tintPaneBackground(reversed)

	if strings.Contains(got, "\x1b[1;7m"+unfocusedTintSGR) || strings.HasPrefix(got, unfocusedTintSGR) {
		t.Fatalf("tintPaneBackground(%q) = %q, want no tint injected before or inside the reverse-video run", reversed, got)
	}
	if !strings.HasSuffix(got, "\x1b[m") {
		t.Errorf("tintPaneBackground(%q) = %q, want the run's own reset kept, untouched", reversed, got)
	}

	// After the reverse run ends, a normal row following it must still get
	// the tint — the pending injection deferred by the reverse run above
	// is not lost, only delayed to the next opportunity.
	followed := reversed + "\nnext row"
	got = tintPaneBackground(followed)
	if !strings.Contains(got, "\x1b[m\n"+unfocusedTintSGR+"next row") {
		t.Errorf("tintPaneBackground(%q) = %q, want the tint resumed for the row after the reverse-video one", followed, got)
	}
}

// TestTintPaneBackgroundDoesNotLeakAcrossANewlineIntoAReverseVideoRow pins
// a real regression this rows-with-newlines design invites: a real
// terminal renderer (this project's own pty harness confirmed it against
// ultraviolet, the actual renderer noggin draws through) serializes SGR
// from *cumulative* cell state, not from how many separate escapes this
// function emitted or in what order — so injecting the tint immediately
// before a bare newline (rather than deferring it to whatever comes after)
// would carry the tint's background straight into the very next row even
// when that row turns out to open with its own Reverse(true), corrupting
// its swap exactly the way TestTintPaneBackgroundLeavesReverseVideoAlone
// guards against for a reverse run reached directly. rootsnav.go's own
// selected-row shape (a plain row immediately followed by a reverse-video
// one) is reproduced here directly.
func TestTintPaneBackgroundDoesNotLeakAcrossANewlineIntoAReverseVideoRow(t *testing.T) {
	content := "\x1b[1mplain row\x1b[m\n\x1b[38;5;245;7mreversed row\x1b[m"
	got := tintPaneBackground(content)

	const reversedOpen = "\x1b[38;5;245;7mreversed row"
	if !strings.Contains(got, reversedOpen) {
		t.Fatalf("tintPaneBackground(%q) = %q, want the reverse-video row's own opening escape kept exactly, with nothing spliced in front of it", content, got)
	}
	if idx := strings.Index(got, reversedOpen); idx > 0 && strings.HasSuffix(got[:idx], unfocusedTintSGR) {
		t.Errorf("tintPaneBackground(%q) = %q, want no tint injected immediately before the reverse-video row's own escape", content, got)
	}
}

// TestTintPaneBackgroundNotFooledByExtendedColorParams is TRIAGE T29's
// regression test (PLAN M9.2), extended by TRIAGE T34/PLAN M9.3 to also
// cover component value 0: the SGR token scan is positional, not
// per-token string equality, so a 256-colour index or truecolor component
// equal to 7, 27 or 0 that happens to follow a 38/48 (extended foreground/
// background) token must never be misread as a reverse-video set/clear or
// a reset — the run must still get the tint, exactly as any other
// non-reverse run does.
func TestTintPaneBackgroundNotFooledByExtendedColorParams(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{"256-colour background index 7", "\x1b[48;5;7mrow\x1b[m"},
		{"256-colour foreground index 27", "\x1b[38;5;27mrow\x1b[m"},
		{"truecolor background component 7", "\x1b[48;2;7;100;200mrow\x1b[m"},
		{"truecolor foreground component 27", "\x1b[38;2;100;27;200mrow\x1b[m"},
		{"truecolor background component 0", "\x1b[48;2;0;100;200mrow\x1b[m"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := tintPaneBackground(c.in)
			if !strings.HasPrefix(got, unfocusedTintSGR) {
				t.Errorf("tintPaneBackground(%q) = %q, want it tinted (prefixed with the tint), not mistaken for reverse-video", c.in, got)
			}
		})
	}
}

// TestUnfocusedFilesNavSelectedRowStaysCleanThroughRenderPane is the
// full-pipeline counterpart to TestTintPaneBackgroundDoesNotLeakAcrossA...:
// tintPaneBackground's own output passes through a *second*,
// outer Style.Render inside renderBox (the box that draws the border), and
// that outer call's own word-wrap step is what actually re-injected the
// tint into a reverse-video row in practice (confirmed live via
// `make check-pty` during this task) even though tintPaneBackground's own
// isolated output never carried the leak — the corruption only appeared
// after the second Style.Render. filesnav.go's real selected-row shape
// (an unstyled mode/scope row followed by the selected, reverse-video row)
// is exercised end to end here, through renderPane itself, not just
// tintPaneBackground in isolation.
func TestUnfocusedFilesNavSelectedRowStaysCleanThroughRenderPane(t *testing.T) {
	m := sized(t, newModel().openFolder(fixtureTree()), 80, 24)
	m.focus = dv1Pane // Files Nav unfocused, tint applies
	l := computeLayout(m.width, m.height, m.navigatorHidden, m.stale)

	rendered := m.renderPane(filesPane, l.files, m.files.view())

	if !strings.Contains(rendered, "7m\u25b8     docs") {
		t.Fatalf("renderPane(filesPane, ...) = %q, want the selected row (\"docs\", the fixture tree's first entry) rendered reverse-video", rendered)
	}
	if strings.Contains(rendered, unfocusedTintSGR+"\x1b[7m") || strings.Contains(rendered, "\x1b[7m"+unfocusedTintSGR) {
		t.Errorf("renderPane(filesPane, ...) = %q, want the tint never combined with the selected row's own reverse-video escape", rendered)
	}
	if !strings.Contains(rendered, unfocusedTintSGR+"\x1b[38;5;245mTree") {
		t.Errorf("renderPane(filesPane, ...) = %q, want the (non-reverse) mode/scope row right above the selection still tinted", rendered)
	}
}

// TestUnfocusedPanesCarryTheS11BackgroundTint is M9.1's own acceptance
// criterion: at every one of the four focus targets, the tint reaches the
// other three panes' rendered box (title and content alike) and is absent
// from the one currently holding Focus.
func TestUnfocusedPanesCarryTheS11BackgroundTint(t *testing.T) {
	m := sized(t, newModel().openFolder(fixtureTree()), 80, 24)
	m = m.withViewerFor(dv2Pane, newViewer())
	m = sized(t, m, 80, 24) // re-lay-out now that DV2 exists

	l := computeLayout(m.width, m.height, m.navigatorHidden, m.stale)
	dv1Rect, dv2Rect := l.splitViewer(true)

	renderPane := func(m model, p pane) string {
		switch p {
		case rootsPane:
			return m.renderPane(p, l.roots, m.roots.view())
		case filesPane:
			return m.renderPane(p, l.files, m.files.view())
		case dv2Pane:
			content := m.composeViewerContent(dv2Pane, dv2Rect.content().width, m.viewerFor(dv2Pane).view())
			return m.renderPane(p, dv2Rect, content)
		default: // dv1Pane
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
				rendered := renderPane(m, p)
				hasTint := strings.Contains(rendered, unfocusedTintSGR)
				wantTint := p != focusTarget
				if hasTint != wantTint {
					t.Errorf("pane %s with focus on %s: tint present = %v, want %v", p.title(), focusTarget.title(), hasTint, wantTint)
				}
			}
		})
	}
}
