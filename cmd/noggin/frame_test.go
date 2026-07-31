package main

import (
	"reflect"
	"strings"
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// sizes exercised by the frame tests: a standard terminal and a large one.
var frameSizes = []struct {
	name          string
	width, height int
}{
	{"standard 80x24", 80, 24},
	{"large 120x40", 120, 40},
}

func TestComputeLayoutProportions(t *testing.T) {
	for _, s := range frameSizes {
		t.Run(s.name, func(t *testing.T) {
			l := computeLayout(s.width, s.height)

			wantNav := clamp(s.width*navigatorPercent/100, minNavigatorWidth, maxNavigatorWidth)
			if l.navigatorWidth() != wantNav {
				t.Errorf("navigator width = %d, want %d", l.navigatorWidth(), wantNav)
			}
			if l.viewer.width != s.width-wantNav {
				t.Errorf("viewer width = %d, want %d", l.viewer.width, s.width-wantNav)
			}
			if l.footer.width != s.width || l.footer.height != footerHeight {
				t.Errorf("footer = %+v, want width %d height %d", l.footer, s.width, footerHeight)
			}

			body := s.height - footerHeight
			if l.viewer.height != body {
				t.Errorf("viewer height = %d, want %d", l.viewer.height, body)
			}
			if got := l.roots.height + l.files.height; got != body {
				t.Errorf("navigator panes total height = %d, want %d", got, body)
			}
			wantRoots := body * rootsPercent / 100
			if l.roots.height != wantRoots {
				t.Errorf("roots height = %d, want %d", l.roots.height, wantRoots)
			}
			if l.roots.width != wantNav || l.files.width != wantNav {
				t.Errorf("navigator pane widths = %d/%d, want %d", l.roots.width, l.files.width, wantNav)
			}
		})
	}
}

func TestViewRendersThreeRegionsAtSize(t *testing.T) {
	for _, s := range frameSizes {
		t.Run(s.name, func(t *testing.T) {
			m := sized(t, newModel(), s.width, s.height)
			view := m.View().Content

			if got := lipgloss.Height(view); got != s.height {
				t.Errorf("view height = %d, want %d", got, s.height)
			}
			for i, line := range strings.Split(view, "\n") {
				if got := lipgloss.Width(line); got != s.width {
					t.Errorf("line %d width = %d, want %d", i, got, s.width)
				}
			}

			l := computeLayout(s.width, s.height)
			lines := strings.Split(view, "\n")

			// Each Navigator pane occupies its own band of rows, so the two
			// panes' titles land at the top of their respective regions and
			// the viewer's title at the top of the body.
			for _, want := range []struct {
				row   int
				title string
			}{
				{1, rootsPane.title()},
				{l.roots.height + 1, filesPane.title()},
				{1, viewerPane.title()},
			} {
				if !strings.Contains(lines[want.row], want.title) {
					t.Errorf("row %d = %q, want it to contain %q", want.row, lines[want.row], want.title)
				}
			}

			// The Navigator column ends where the viewer begins.
			bodyRow := lines[1]
			if got := lipgloss.Width(bodyRow[:strings.Index(bodyRow, viewerPane.title())]); got < l.navigatorWidth() {
				t.Errorf("viewer starts at column %d, want at or after %d", got, l.navigatorWidth())
			}

			// The footer is the last row and names the focused pane.
			footer := lines[len(lines)-1]
			if !strings.Contains(footer, m.focus.title()) || !strings.Contains(footer, keys.Quit.Help().Key) {
				t.Errorf("footer = %q, want focus title and quit hint", footer)
			}
		})
	}
}

func TestViewReflowsOnResize(t *testing.T) {
	m := sized(t, newModel(), frameSizes[0].width, frameSizes[0].height)
	first := m.View().Content
	m = sized(t, m, frameSizes[1].width, frameSizes[1].height)
	second := m.View().Content

	if first == second {
		t.Fatal("view did not change across a resize")
	}
	if lipgloss.Width(second) != frameSizes[1].width || lipgloss.Height(second) != frameSizes[1].height {
		t.Errorf("resized view = %dx%d, want %dx%d",
			lipgloss.Width(second), lipgloss.Height(second), frameSizes[1].width, frameSizes[1].height)
	}
}

func TestFocusMovesBetweenPanesAndBack(t *testing.T) {
	m := sized(t, newModel(), 80, 24)
	start := m.focus

	var forward []pane
	for range paneOrder {
		m = press(t, m, keys.NextPane.Keys()[0])
		forward = append(forward, m.focus)
	}
	if m.focus != start {
		t.Errorf("focus after a full cycle = %v, want %v", m.focus, start)
	}
	seen := map[pane]bool{}
	for _, p := range forward {
		seen[p] = true
	}
	if len(seen) != len(paneOrder) {
		t.Errorf("forward cycle visited %d panes, want %d", len(seen), len(paneOrder))
	}

	for i := len(forward) - 2; i >= 0; i-- {
		m = press(t, m, keys.PrevPane.Keys()[0])
		if m.focus != forward[i] {
			t.Errorf("focus stepping back = %v, want %v", m.focus, forward[i])
		}
	}
	m = press(t, m, keys.PrevPane.Keys()[0])
	if m.focus != start {
		t.Errorf("focus after cycling back = %v, want %v", m.focus, start)
	}
}

func TestNarrowTerminalDropsNavigatorAndKeepsFocusVisible(t *testing.T) {
	narrow := minNavigatorWidth + minViewerWidth - 1
	l := computeLayout(narrow, 24)
	if l.navigatorWidth() != 0 {
		t.Fatalf("navigator width at %d columns = %d, want 0", narrow, l.navigatorWidth())
	}
	if l.viewer.width != narrow {
		t.Errorf("viewer width = %d, want %d", l.viewer.width, narrow)
	}

	m := sized(t, newModel(), narrow, 24)
	if m.focus != viewerPane {
		t.Errorf("focus after hiding the navigator = %v, want %v", m.focus, viewerPane)
	}
	m = press(t, m, keys.NextPane.Keys()[0])
	if m.focus != viewerPane {
		t.Errorf("focus after next-pane with no navigator = %v, want %v", m.focus, viewerPane)
	}
}

func TestQuitBindings(t *testing.T) {
	for _, k := range keys.Quit.Keys() {
		m := sized(t, newModel(), 80, 24)
		_, cmd := m.Update(keyPress(k))
		if cmd == nil {
			t.Fatalf("key %q produced no command, want quit", k)
		}
		if _, ok := cmd().(tea.QuitMsg); !ok {
			t.Errorf("key %q produced %T, want tea.QuitMsg", k, cmd())
		}
	}
}

// sized applies a window size to the model and returns the concrete result.
func sized(t *testing.T, m model, width, height int) model {
	t.Helper()
	next, _ := m.Update(tea.WindowSizeMsg{Width: width, Height: height})
	return concrete(t, next)
}

// press sends the named key to the model and returns the concrete result.
func press(t *testing.T, m model, name string) model {
	t.Helper()
	next, _ := m.Update(keyPress(name))
	return concrete(t, next)
}

func concrete(t *testing.T, m tea.Model) model {
	t.Helper()
	got, ok := m.(model)
	if !ok {
		t.Fatalf("Update returned %T, want model", m)
	}
	return got
}

// allBindings is every binding in the keymap. New bindings must be added here
// so the round-trip below covers them.
func allBindings() []key.Binding {
	return []key.Binding{
		keys.NextPane, keys.PrevPane, keys.Quit,
		keys.LineUp, keys.LineDown,
		keys.HalfPageUp, keys.HalfPageDown,
		keys.PageUp, keys.PageDown,
		keys.Top, keys.Bottom,
	}
}

// TestKeyPressRoundTrip pins keyPress to Bubble Tea's own key naming. Bubble
// Tea v2 has no public parser for the names its bindings are written in, so
// keyPress inverts Key.Keystroke by hand; this asserts the inverse is exact for
// every key the keymap actually uses, and that key.Matches — which compares
// String() to the binding — therefore recognises a synthesised key.
func TestKeyPressRoundTrip(t *testing.T) {
	for _, b := range allBindings() {
		for _, name := range b.Keys() {
			msg := keyPress(name)
			if got := msg.String(); got != name {
				t.Errorf("keyPress(%q).String() = %q, want %q", name, got, name)
			}
			if !key.Matches(msg, b) {
				t.Errorf("keyPress(%q) does not match the binding it came from", name)
			}
		}
	}
}

// TestEveryBindingIsCovered guards the round-trip above against a binding being
// added to the keymap and not to allBindings.
func TestEveryBindingIsCovered(t *testing.T) {
	want := reflect.TypeOf(keys).NumField()
	if got := len(allBindings()); got != want {
		t.Errorf("allBindings covers %d bindings, but keyMap has %d fields", got, want)
	}
}
