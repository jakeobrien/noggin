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
			l := computeLayout(s.width, s.height, false, false)

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

			l := computeLayout(s.width, s.height, false, false)
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
				{1, dv1Pane.title()},
			} {
				if !strings.Contains(lines[want.row], want.title) {
					t.Errorf("row %d = %q, want it to contain %q", want.row, lines[want.row], want.title)
				}
			}

			// The Navigator column ends where the viewer begins.
			bodyRow := lines[1]
			if got := lipgloss.Width(bodyRow[:strings.Index(bodyRow, dv1Pane.title())]); got < l.navigatorWidth() {
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

// TestFocusJumpsReachEveryVisiblePane covers S5's replacement for Tab-
// cycling (removed as of C3/M1.1): FocusRoots/FocusFiles jump Focus
// directly, from any starting pane, to a pane that stays put otherwise —
// there is no cycle to walk any more, just direct jumps (focus_test.go
// covers the jumps themselves in full; this only pins that Tab/Shift+Tab
// no longer do anything).
func TestTabAndShiftTabAreUnbound(t *testing.T) {
	m := sized(t, newModel(), 80, 24)
	start := m.focus

	m = press(t, m, "tab")
	if m.focus != start {
		t.Errorf("focus after tab = %v, want unchanged %v (Tab-cycling is removed, S5)", m.focus, start)
	}
	m = press(t, m, "shift+tab")
	if m.focus != start {
		t.Errorf("focus after shift+tab = %v, want unchanged %v (Tab-cycling is removed, S5)", m.focus, start)
	}
}

func TestNarrowTerminalDropsNavigatorAndKeepsFocusVisible(t *testing.T) {
	narrow := minNavigatorWidth + minViewerWidth - 1
	l := computeLayout(narrow, 24, false, false)
	if l.navigatorWidth() != 0 {
		t.Fatalf("navigator width at %d columns = %d, want 0", narrow, l.navigatorWidth())
	}
	if l.viewer.width != narrow {
		t.Errorf("viewer width = %d, want %d", l.viewer.width, narrow)
	}

	m := sized(t, newModel(), narrow, 24)
	if m.focus != dv1Pane {
		t.Errorf("focus after hiding the navigator = %v, want %v", m.focus, dv1Pane)
	}
	// FocusFiles cannot reveal a Navigator hidden by lack of width (unlike
	// the S9 toggle, model.focusPane only ever un-hides that toggle, never
	// widens the terminal) — resolveFocus falls back to dv1Pane, the one
	// pane still visible, so the jump is a no-op here.
	m = press(t, m, keys.FocusFiles.Keys()[0])
	if m.focus != dv1Pane {
		t.Errorf("focus after FocusFiles on a too-narrow terminal = %v, want unchanged %v", m.focus, dv1Pane)
	}
}

// TestSplitViewerEqualHalvesWhenDV2Exists is C3/M1.2's own layout test:
// splitViewer divides the viewer region into two regions summing to its
// full width, equal or off by exactly one column (D54: no minimum-width
// guard, so this holds at any size), and unchanged height.
func TestSplitViewerEqualHalvesWhenDV2Exists(t *testing.T) {
	for _, s := range frameSizes {
		t.Run(s.name, func(t *testing.T) {
			l := computeLayout(s.width, s.height, false, false)
			dv1, dv2 := l.splitViewer(true)

			if got := dv1.width + dv2.width; got != l.viewer.width {
				t.Errorf("dv1 width %d + dv2 width %d = %d, want %d", dv1.width, dv2.width, got, l.viewer.width)
			}
			if diff := dv1.width - dv2.width; diff < 0 || diff > 1 {
				t.Errorf("dv1 width %d, dv2 width %d: not equal or off by one", dv1.width, dv2.width)
			}
			if dv1.height != l.viewer.height || dv2.height != l.viewer.height {
				t.Errorf("dv1/dv2 height = %d/%d, want both %d", dv1.height, dv2.height, l.viewer.height)
			}
		})
	}
}

// TestSplitViewerDV1FullWidthWithoutDV2 confirms the other half of C3/M1.2:
// DV1 takes the whole viewer region, and DV2's is the zero rect, when DV2
// does not exist.
func TestSplitViewerDV1FullWidthWithoutDV2(t *testing.T) {
	for _, s := range frameSizes {
		t.Run(s.name, func(t *testing.T) {
			l := computeLayout(s.width, s.height, false, false)
			dv1, dv2 := l.splitViewer(false)

			if dv1 != l.viewer {
				t.Errorf("dv1 = %+v, want the full viewer region %+v", dv1, l.viewer)
			}
			if !dv2.empty() {
				t.Errorf("dv2 = %+v, want empty", dv2)
			}
		})
	}
}

// TestFrameRendersBothViewersWhenDV2Exists drives the split through the
// actual model/frame path (rather than the pure layout function above):
// once DV2 exists, both panes render their own title in the body row,
// and the frame's overall width/height are unaffected.
func TestFrameRendersBothViewersWhenDV2Exists(t *testing.T) {
	m := sized(t, newModel(), frameSizes[0].width, frameSizes[0].height)
	m = m.withViewerFor(dv2Pane, newViewer())
	m = sized(t, m, frameSizes[0].width, frameSizes[0].height) // re-lay-out now that dv2 exists

	view := m.View().Content
	if got := lipgloss.Height(view); got != frameSizes[0].height {
		t.Errorf("view height = %d, want %d", got, frameSizes[0].height)
	}
	for i, line := range strings.Split(view, "\n") {
		if got := lipgloss.Width(line); got != frameSizes[0].width {
			t.Errorf("line %d width = %d, want %d", i, got, frameSizes[0].width)
		}
	}

	titleRow := strings.Split(view, "\n")[1]
	if got := strings.Count(titleRow, dv1Pane.title()); got != 2 {
		t.Errorf("title row contains %q %d times, want 2 (DV1 and DV2 each draw one)", dv1Pane.title(), got)
	}
	if !m.visible(dv2Pane) {
		t.Error("dv2Pane not visible once dv2 exists at this size")
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

// typeText sends s to the model one rune at a time, as a find query's
// keystrokes arrive (find.go's updateFindQuery) rather than as a keymap.go
// binding — there is no binding for arbitrary typed text, so this builds the
// tea.KeyPressMsg directly instead of going through keyPress, the same way
// production code never does but a query's actual characters are not one of
// the named keys D2 confines to keymap.go.
func typeText(t *testing.T, m model, s string) model {
	t.Helper()
	for _, r := range s {
		next, _ := m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		m = concrete(t, next)
	}
	return m
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
		keys.Quit,
		keys.Star,
		keys.Expand, keys.Collapse, keys.Activate,
		keys.LineUp, keys.LineDown,
		keys.HalfPageUp, keys.HalfPageDown,
		keys.PageUp, keys.PageDown,
		keys.Top, keys.Bottom,
		keys.Edit, keys.RenderMode,
		keys.Find, keys.FindNext, keys.FindPrevious, keys.Navigator, keys.Cancel,
		keys.FocusRoots, keys.FocusFiles, keys.FocusDV1, keys.FocusDV2, keys.CloseDV2, keys.Refresh,
		keys.Help, keys.Picker, keys.PickerRestore,
		keys.PickerUp, keys.PickerDown, keys.PickerMode, keys.PickerScope,
		keys.ToggleHidden, keys.PickerStar,
		keys.Peek,
		keys.ActivateDV2, keys.PeekDV2,
		keys.FilesMode, keys.FilesScope,
		keys.HistoryBack, keys.HistoryForward,
		keys.Rename,
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
