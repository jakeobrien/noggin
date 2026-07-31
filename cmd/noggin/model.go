package main

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// pane identifies a focusable region of the frame.
type pane int

const (
	rootsPane pane = iota
	filesPane
	viewerPane
)

// paneOrder is the focus cycle order: down the Navigator column, then out to
// the Document Viewer.
var paneOrder = []pane{rootsPane, filesPane, viewerPane}

func (p pane) title() string {
	switch p {
	case rootsPane:
		return "Roots"
	case filesPane:
		return "Files"
	default:
		return "Document"
	}
}

// model is the root Bubble Tea model: the frame (Navigator column, Document
// Viewer, Footer), keyboard focus, and the terminal size the frame reflows to.
type model struct {
	width  int
	height int
	focus  pane
	viewer viewer
}

func newModel() model {
	return model{focus: filesPane, viewer: newViewer()}
}

// openDocument makes d the active document and puts focus on it, which is what
// asking to read a document means.
func (m model) openDocument(d document) model {
	m.viewer = m.viewer.open(d)
	m.focus = m.resolveFocus(viewerPane)
	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.focus = m.resolveFocus(m.focus)
		content := computeLayout(m.width, m.height).viewer.content()
		m.viewer = m.viewer.setSize(content.width, content.height)
		return m, nil
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.NextPane):
			m.focus = m.shiftFocus(1)
			return m, nil
		case key.Matches(msg, keys.PrevPane):
			m.focus = m.shiftFocus(-1)
			return m, nil
		}
		// Scroll keys belong to whichever pane has focus; only the viewer
		// claims them so far.
		if m.focus == viewerPane {
			if next, scrolled := m.viewer.scroll(msg); scrolled {
				m.viewer = next
				return m, nil
			}
		}
	}
	return m, nil
}

// visible reports whether a pane has a region to draw in at the current size.
// A terminal too narrow for the Navigator column hides its panes (see
// computeLayout), and focus must not land in one.
func (m model) visible(p pane) bool {
	l := computeLayout(m.width, m.height)
	switch p {
	case rootsPane:
		return !l.roots.empty()
	case filesPane:
		return !l.files.empty()
	default:
		return !l.viewer.empty()
	}
}

// shiftFocus moves focus by delta positions through paneOrder, wrapping, and
// skipping panes that are not currently visible.
func (m model) shiftFocus(delta int) pane {
	start := indexOfPane(m.focus)
	n := len(paneOrder)
	for i := 1; i <= n; i++ {
		candidate := paneOrder[((start+delta*i)%n+n)%n]
		if m.visible(candidate) {
			return candidate
		}
	}
	return m.focus
}

// resolveFocus keeps focus on a visible pane across a resize.
func (m model) resolveFocus(want pane) pane {
	if m.visible(want) {
		return want
	}
	for _, p := range paneOrder {
		if m.visible(p) {
			return p
		}
	}
	return want
}

func indexOfPane(p pane) int {
	for i, candidate := range paneOrder {
		if candidate == p {
			return i
		}
	}
	return 0
}

var (
	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240"))
	focusedBorderStyle = borderStyle.
				BorderForeground(lipgloss.Color("205"))
	titleStyle  = lipgloss.NewStyle().Bold(true)
	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))
)

// View returns the frame. Bubble Tea v2 derives terminal state (alt screen,
// cursor, mouse mode, window title) from the returned View on every render, so
// every field left at its zero value is a deliberate "off": noggin draws inline
// with no cursor, exactly as it did under v1.
func (m model) View() tea.View {
	return tea.NewView(m.frame())
}

// frame composes the three regions into the string View wraps.
func (m model) frame() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	l := computeLayout(m.width, m.height)

	body := m.renderPane(viewerPane, l.viewer, m.viewer.view())
	if l.navigatorWidth() > 0 {
		navigator := lipgloss.JoinVertical(lipgloss.Left,
			m.renderPane(rootsPane, l.roots, ""),
			m.renderPane(filesPane, l.files, ""),
		)
		body = lipgloss.JoinHorizontal(lipgloss.Top, navigator, body)
	}

	if l.footer.empty() {
		return body
	}
	return lipgloss.JoinVertical(lipgloss.Left, body, m.renderFooter(l.footer))
}

// renderPane draws one region at exactly r's outer size, border included: the
// pane's title on the first content row, then content, which is expected to be
// exactly r.content() in size.
//
// lipgloss v2 counts the border inside Width/Height (PR #451), so the region's
// outer size is passed straight through rather than pre-deducting the border as
// v1 required.
func (m model) renderPane(p pane, r rect, content string) string {
	if r.empty() {
		return ""
	}
	style := borderStyle
	if m.focus == p {
		style = focusedBorderStyle
	}
	box := style.Width(r.width).Height(r.height)
	title := titleStyle.Render(m.paneTitle(p))
	if content == "" {
		return box.Render(title)
	}
	return box.Render(title + "\n" + content)
}

// paneTitle is a pane's heading: its name, plus what it is currently showing
// where that is worth saying.
func (m model) paneTitle(p pane) string {
	if p == viewerPane {
		if label := m.viewer.label(); label != "" {
			return p.title() + ": " + label
		}
	}
	return p.title()
}

// renderFooter draws the footer line. Focus-aware hints and transient messages
// land here in M4.2; for now it names the focused pane and the quit key.
func (m model) renderFooter(r rect) string {
	hint := m.focus.title() + " · " + keys.Quit.Help().Key + " " + keys.Quit.Help().Desc
	return footerStyle.Width(r.width).MaxWidth(r.width).MaxHeight(r.height).Render(hint)
}
