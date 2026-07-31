package main

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"noggin/internal/tree"
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

// title is the pane's heading, shown in its border and the footer. It is a
// short display label, not the GLOSSARY.md term ("Roots Nav", "Files Nav",
// "Document Viewer") — those name the pane in prose, but a title lives on a
// single content row shared with the Active Document's own label
// (paneTitle), so it stays terse on purpose rather than drifting to match the
// glossary.
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
	files  filesNav
	viewer viewer

	// roots is the session's note sources (S1); activeRoot indexes the one the
	// Files Nav is browsing, or -1 when none. message carries a launch-time
	// rejection notice until M4.2's footer surfaces it.
	roots      []root
	activeRoot int
	message    string
}

func newModel() model {
	return model{focus: filesPane, files: newFilesNav(), viewer: newViewer(), activeRoot: -1}
}

// applyResolution installs an S1 launch resolution: the session's roots, which
// one is active, an optional document, and any rejection message. Walking the
// active root or reading the document can fail; that becomes a message, never
// an exit — the session always starts (S1's "silence is not an error").
func (m model) applyResolution(res resolution) model {
	m.roots = res.roots
	m.activeRoot = res.active
	m.message = res.message

	if path := res.activeRootPath(); path != "" {
		if node, err := tree.Walk(path); err != nil {
			m.message = fmt.Sprintf("cannot read %s: %v", path, err)
		} else {
			m = m.openFolder(node)
		}
	}
	if res.docPath != "" {
		if doc, err := loadDocument(res.docPath); err != nil {
			m.message = fmt.Sprintf("cannot open %s: %v", res.docPath, err)
		} else {
			m = m.openDocument(doc)
		}
	}
	return m
}

// openFolder points the Files Nav at root and puts focus on it, which is what
// browsing a folder means.
func (m model) openFolder(root *tree.Node) model {
	m.files = m.files.setRoot(root)
	m.focus = m.resolveFocus(filesPane)
	return m
}

// openDocument makes d the active document and puts focus on it, which is what
// asking to read a document means.
func (m model) openDocument(d document) model {
	m.viewer = m.viewer.open(d)
	m.focus = m.resolveFocus(viewerPane)
	return m
}

// activate opens the Files Nav's selected file as the active document,
// moving focus to the viewer (M2.3). A folder or an empty pane is a no-op —
// activating a folder is left for later, per the Files Nav's own
// expand/collapse bindings. A read error leaves the session exactly as it
// was rather than exiting it (S1's "silence is not an error" principle,
// generalized here ahead of M3.1's launch-time version of the same rule).
func (m model) activate() model {
	n := m.files.selectedNode()
	if n == nil || n.IsDir {
		return m
	}
	d, err := loadDocument(n.Path)
	if err != nil {
		return m
	}
	return m.openDocument(d)
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.focus = m.resolveFocus(m.focus)
		l := computeLayout(m.width, m.height)
		viewerContent := l.viewer.content()
		m.viewer = m.viewer.setSize(viewerContent.width, viewerContent.height)
		filesContent := l.files.content()
		m.files = m.files.setSize(filesContent.width, filesContent.height)
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
		// Movement keys belong to whichever pane has focus: the viewer scrolls,
		// the Files Nav moves its selection and opens/closes folders.
		switch m.focus {
		case viewerPane:
			if next, scrolled := m.viewer.scroll(msg); scrolled {
				m.viewer = next
				return m, nil
			}
		case filesPane:
			if key.Matches(msg, keys.Activate) {
				return m.activate(), nil
			}
			if next, handled := m.files.handleKey(msg); handled {
				m.files = next
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
			m.renderPane(filesPane, l.files, m.files.view()),
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
