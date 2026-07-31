package main

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
}

func newModel() model {
	return model{focus: filesPane}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.focus = m.resolveFocus(m.focus)
		return m, nil
	case tea.KeyMsg:
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

func (m model) View() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	l := computeLayout(m.width, m.height)

	body := m.renderPane(viewerPane, l.viewer)
	if l.navigatorWidth() > 0 {
		navigator := lipgloss.JoinVertical(lipgloss.Left,
			m.renderPane(rootsPane, l.roots),
			m.renderPane(filesPane, l.files),
		)
		body = lipgloss.JoinHorizontal(lipgloss.Top, navigator, body)
	}

	if l.footer.empty() {
		return body
	}
	return lipgloss.JoinVertical(lipgloss.Left, body, m.renderFooter(l.footer))
}

// renderPane draws one region at exactly r's outer size, border included.
func (m model) renderPane(p pane, r rect) string {
	if r.empty() {
		return ""
	}
	style := borderStyle
	if m.focus == p {
		style = focusedBorderStyle
	}
	// The border occupies one cell on each side of the content box.
	inner := style.Width(max(r.width-2, 0)).Height(max(r.height-2, 0))
	return inner.Render(titleStyle.Render(p.title()))
}

// renderFooter draws the footer line. Focus-aware hints and transient messages
// land here in M4.2; for now it names the focused pane and the quit key.
func (m model) renderFooter(r rect) string {
	hint := m.focus.title() + " · " + keys.Quit.Help().Key + " " + keys.Quit.Help().Desc
	return footerStyle.Width(r.width).MaxWidth(r.width).MaxHeight(r.height).Render(hint)
}
