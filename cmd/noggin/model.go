package main

import (
	"fmt"
	"os"
	"os/exec"

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

	// navigatorHidden is the S9 Navigator toggle: hides the whole sidebar
	// (Roots Nav and Files Nav) and gives the Document Viewer the full width
	// when set. It never touches either pane's own state (selection,
	// expansion) — only the layout computed from it (see layout()) — so
	// restoring it leaves that state exactly as it was (M5.2).
	navigatorHidden bool

	// roots is the session's note sources (S1): the Roots Nav pane and its
	// selection, doubling as the record of which root is active. message
	// carries a transient footer notice — a launch-time rejection (S1) or a
	// refused action (e.g. activating a broken root, S3); M4.2 will add
	// clear-on-next-keypress and richer hints, but showing it at all cannot
	// wait for that milestone (see D22).
	roots   rootsNav
	message string

	// helpOpen is the S12 Help Modal (M5.3): when set, it owns the keyboard
	// (handleHelpInput) and the frame gains its overlay (overlayHelp), the
	// same shape m.viewer.findEditing() gives an in-progress search — an
	// exclusive input mode layered on top of the model's other state, not a
	// pane of its own, so it never touches m.focus.
	helpOpen bool

	// ripgrep is the S14 detection result (M1.1): whether the Picker's
	// search layer has a binary to run on, and why not when it doesn't.
	// Detected once at startup (main.go) and carried here so nothing
	// downstream re-shells-out to ask; it does not affect the launch path
	// itself — a session with no ripgrep still launches and reads normally.
	ripgrep ripgrepAvailability

	// picker is the S1 modal's state (M2.1): open/closed, mode, Scope and
	// Prompt. Reachable from every Focus, so — like helpOpen — it never
	// touches m.focus; it is a layer over the frame, not a pane.
	picker picker
}

func newModel() model {
	return model{focus: filesPane, files: newFilesNav(), viewer: newViewer(), roots: newRootsNav(), picker: newPicker()}
}

// withStorePath sets the resolved starred-roots config path (M3.2's
// storePath()) that a Roots Nav star toggle persists through. main.go
// resolves it once at startup so a keypress never re-derives
// os.UserConfigDir() itself.
func (m model) withStorePath(path string) model {
	m.roots.storePath = path
	return m
}

// withRipgrep installs the S14 ripgrep detection result (detectRipgrep())
// that main.go resolves once at startup, the same one-shot-then-carry
// pattern as withStorePath.
func (m model) withRipgrep(avail ripgrepAvailability) model {
	m.ripgrep = avail
	return m
}

// applyResolution installs an S1 launch resolution: the session's roots, which
// one is active, an optional document, and any rejection message. Walking the
// active root or reading the document can fail; that becomes a message, never
// an exit — the session always starts (S1's "silence is not an error").
func (m model) applyResolution(res resolution) model {
	m.roots = m.roots.setRoots(res.roots, res.active)
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
// expand/collapse bindings. A read error is a footer message rather than a
// silent no-op (T15), the same "silence is not an error" shape as
// activateRoot's own refusal below and S1's launch-time version of the rule.
func (m model) activate() model {
	n := m.files.selectedNode()
	if n == nil || n.IsDir {
		return m
	}
	d, err := loadDocument(n.Path)
	if err != nil {
		m.message = fmt.Sprintf("cannot open %s: %v", n.Path, err)
		return m
	}
	return m.openDocument(d)
}

// activateRoot makes the Roots Nav's selected entry the active root,
// re-rooting the Files Nav (S4; M3.4 verifies the open document survives
// this) while leaving focus free to land in the Files Nav, same shape as
// activate() above. A broken root (missing path) is refused with a footer
// message rather than failing silently (S3), the same shape activate() now
// gives its own read-error path (T15).
func (m model) activateRoot() model {
	row, ok := m.roots.selectedRow()
	if !ok {
		return m
	}
	if row.broken {
		m.message = fmt.Sprintf("%s no longer exists; not activated", row.root.path)
		return m
	}
	node, err := tree.Walk(row.root.path)
	if err != nil {
		m.message = fmt.Sprintf("cannot read %s: %v", row.root.path, err)
		return m
	}
	m.roots = m.roots.activate(m.roots.selected)
	m.message = ""
	return m.openFolder(node)
}

// toggleStar flips the Roots Nav selection's starred mark and persists it via
// the S2 store (M3.2's toggleStore), using the storePath resolved once at
// startup (withStorePath). A config directory that could not be resolved
// (m.roots.storePath == "", the same rare case loadStarredRoots degrades on)
// declines the toggle with a message instead (T18), the same as the
// explicit-write-error branch below — a star that cannot be persisted must
// not appear to have flipped, or a restart would silently revert it with no
// explanation.
func (m model) toggleStar() model {
	row, ok := m.roots.selectedRow()
	if !ok {
		return m
	}
	if m.roots.storePath == "" {
		m.message = "cannot save starred roots: no resolvable config directory"
		return m
	}
	if _, err := toggleStore(m.roots.storePath, row.root.path); err != nil {
		m.message = fmt.Sprintf("cannot save starred roots: %v", err)
		return m
	}
	m.roots = m.roots.toggleStarred(m.roots.selected)
	return m
}

// editFinishedMsg reports that $EDITOR has exited (S10, D3). path is the
// document that was handed off, carried on the message rather than re-read
// from m.viewer at delivery time, since focus and the open document could in
// principle have changed while the terminal was away. err is any failure
// starting or running the editor.
type editFinishedMsg struct {
	path string
	err  error
}

// startEdit hands the active document to $EDITOR (falling back to vi),
// releasing the terminal for the duration and resuming when it exits (D3).
// It is global, not pane-specific: S10 names "the active document", not "the
// focused pane". With no document open it reports that in the footer rather
// than doing nothing.
func (m model) startEdit() (model, tea.Cmd) {
	if !m.viewer.doc.isOpen() {
		m.message = "no document open to edit"
		return m, nil
	}
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	path := m.viewer.doc.path
	return m, tea.ExecProcess(exec.Command(editor, path), func(err error) tea.Msg {
		return editFinishedMsg{path: path, err: err}
	})
}

// toggleNavigator flips the S9 Navigator toggle and re-lays-out the frame at
// the current terminal size. Hiding moves focus off a Navigator pane it can
// no longer land in, the same way a resize below the narrow-terminal
// threshold does (resolveFocus); showing it again leaves focus exactly where
// it was, since the Document Viewer stays visible throughout and the toggle
// never touches either Navigator pane's own selection state.
//
// It drives the same resize path a real tea.WindowSizeMsg does (both re-lay
// out the frame and can invalidate a find, see resize's doc comment), but is
// not itself a resize — resizeCauseNavigatorToggle keeps resize's own
// dismissed/cleared messages honest about which one actually happened (T45).
func (m model) toggleNavigator() model {
	m.navigatorHidden = !m.navigatorHidden
	return m.resize(m.width, m.height, resizeCauseNavigatorToggle)
}

// startFind begins an S8 search over the active document, moving focus to
// the viewer the same way activate() and activateRoot() do — the keys a
// search needs next (typing, Cancel, Activate to commit) all target it. With
// no document open it reports that instead, the same shape as startEdit.
func (m model) startFind() (model, tea.Cmd) {
	if !m.viewer.doc.isOpen() {
		m.message = "no document open to find in"
		return m, nil
	}
	next, cmd := m.viewer.startFind()
	m.viewer = next
	m.focus = m.resolveFocus(viewerPane)
	return m, cmd
}

// handleFindInput is where a keypress goes while a find query has the
// keyboard (m.viewer.findEditing()): Cancel dismisses it, Activate (enter)
// commits it, and everything else — including keys bound elsewhere, like q
// or the scroll keys — types into the query instead of acting on its usual
// binding. That exclusivity is deliberate: a query for the word "question"
// must not quit the program on its first letter.
func (m model) handleFindInput(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Cancel):
		m.viewer = m.viewer.cancelFind()
		return m, nil
	case key.Matches(msg, keys.Activate):
		m.viewer = m.viewer.commitFind()
		return m, nil
	}
	next, cmd := m.viewer.updateFindQuery(msg)
	m.viewer = next
	return m, cmd
}

// openHelp opens the S12 Help Modal for the focused pane's context. Unlike
// toggleNavigator, this is one-directional — handleHelpInput is the only
// path that closes it, so the main switch below only ever calls this when
// m.helpOpen is already false (the helpOpen branch there returns early
// otherwise). It never touches m.focus: the modal is a layer over the
// frame, not a pane swap, so closing it always returns focus exactly where
// it was without needing to save or restore it.
func (m model) openHelp() model {
	m.helpOpen = true
	return m
}

// handleHelpInput is where a keypress goes while the Help Modal is open
// (m.helpOpen): Help and Cancel both close it — Help because pressing the
// affordance again is the natural undo, Cancel because it is the generic
// one-key dismissal every modal in this program shares (see keymap.go's
// Cancel doc). Everything else is a no-op: the modal owns the keyboard while
// it is up, the same exclusivity handleFindInput gives an in-progress
// search, so a keypress meant for the pane underneath cannot leak through
// the modal by accident.
func (m model) handleHelpInput(msg tea.KeyPressMsg) model {
	if key.Matches(msg, keys.Help) || key.Matches(msg, keys.Cancel) {
		m.helpOpen = false
	}
	return m
}

// finishEdit resumes after $EDITOR exits, re-rendering the edited document
// from disk (D3's "no file watching" — this is the one explicit re-read) so
// changes are immediately visible. Reusing loadDocument and openDocument keeps
// this the same read-and-open path as M2.3's activate and S1's launch, rather
// than a third copy of it. A failure to launch or a since-vanished file is a
// footer message, not a crash — the same "silence is not an error" shape as
// activate() and applyResolution().
func (m model) finishEdit(msg editFinishedMsg) model {
	if msg.err != nil {
		m.message = fmt.Sprintf("edit failed: %v", msg.err)
		return m
	}
	d, err := loadDocument(msg.path)
	if err != nil {
		m.message = fmt.Sprintf("cannot open %s: %v", msg.path, err)
		return m
	}
	m.message = ""
	return m.openDocument(d)
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case editFinishedMsg:
		return m.finishEdit(msg), nil
	// The Picker's two async messages (M2.4). Neither is a keypress, so
	// neither clears m.message (S11 ties that to the user's next key, not to
	// a timer maturing or a background search reporting back), and both are
	// gated on the Picker still being open inside their handlers rather than
	// here — a message can outlive the modal that asked for it.
	case pickerSearchMsg:
		return m.dispatchPickerSearch(msg)
	case pickerResultsMsg:
		return m.applyPickerResults(msg), nil
	case tea.WindowSizeMsg:
		return m.resize(msg.Width, msg.Height, resizeCauseWindow), nil
	case tea.KeyPressMsg:
		// A transient message (D22) is displaced by the action it reports on;
		// any keypress after that clears it (S11), whether or not this one
		// sets a fresh message of its own.
		m.message = ""
		// A find query in progress owns every key but Cancel and Activate
		// (see handleFindInput); this must run before Quit and the rest below
		// get a look at the keypress.
		if m.viewer.findEditing() {
			return m.handleFindInput(msg)
		}
		if m.helpOpen {
			return m.handleHelpInput(msg), nil
		}
		// The Picker (S1) is reachable from every Focus and, once open, owns
		// the keyboard the same exclusive way a find-in-progress or the Help
		// Modal does — see handlePickerInput.
		if m.picker.open {
			return m.handlePickerInput(msg)
		}
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.NextPane):
			m.focus = m.shiftFocus(1)
			return m, nil
		case key.Matches(msg, keys.PrevPane):
			m.focus = m.shiftFocus(-1)
			return m, nil
		case key.Matches(msg, keys.Navigator):
			return m.toggleNavigator(), nil
		case key.Matches(msg, keys.Help):
			return m.openHelp(), nil
		case key.Matches(msg, keys.Picker):
			return m.openPicker()
		case key.Matches(msg, keys.Edit):
			return m.startEdit()
		case key.Matches(msg, keys.Find):
			return m.startFind()
		// FindNext and FindPrevious are global like Find itself, not scoped to
		// viewer focus: stepping through matches is meaningless anywhere else,
		// but the query lives in the Document Viewer regardless of which pane
		// currently has focus.
		case key.Matches(msg, keys.FindNext):
			if m.viewer.hasFindMatches() {
				m.viewer = m.viewer.findNext()
			}
			return m, nil
		case key.Matches(msg, keys.FindPrevious):
			if m.viewer.hasFindMatches() {
				m.viewer = m.viewer.findPrevious()
			}
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
		case rootsPane:
			if key.Matches(msg, keys.Activate) {
				return m.activateRoot(), nil
			}
			if key.Matches(msg, keys.Star) {
				return m.toggleStar(), nil
			}
			if next, handled := m.roots.handleKey(msg); handled {
				m.roots = next
				return m, nil
			}
		}
	}
	return m, nil
}

// resizeCause names what drove a call to resize, so its dismissed/cleared
// find messages (below) can say what actually happened instead of
// hardcoding "window resized" even when the real trigger was a Navigator
// toggle (T45) — both drive the frame through this one path (see resize's
// doc comment), but they are not the same event, and the footer must not
// claim one for the other.
type resizeCause string

const (
	resizeCauseWindow          resizeCause = "window resized"
	resizeCauseNavigatorToggle resizeCause = "sidebar toggled"
)

// resize sets the terminal size and re-lays-out the frame: resolving focus
// onto a still-visible pane, then sizing each pane's content box from the
// resulting layout (which also reflects the S9 Navigator toggle, see
// layout()). Both a real terminal resize and toggleNavigator drive the frame
// through this one path, so they reflow identically; cause records which of
// the two actually happened, for the messages below.
func (m model) resize(width, height int, cause resizeCause) model {
	m.width, m.height = width, height
	m.focus = m.resolveFocus(m.focus)
	l := m.layout()
	viewerContent := l.viewer.content()
	// A width change re-renders the viewer (setSize), and a re-render
	// silently drops any find in progress (viewer.render's doc comment) --
	// fine when nothing was actively being typed, but a query mid-edit
	// vanishing with no explanation violates S1/S8's "confusion is not
	// allowed" (T34). Catch it here, the one path both a real terminal
	// resize and toggleNavigator drive through, by comparing editing state
	// before and after: toggleNavigator can never trigger this in practice
	// (handleFindInput owns every key but Cancel/Activate while editing, so
	// the Navigator binding never reaches it), but a tea.WindowSizeMsg is
	// not gated the same way.
	//
	// The same re-render just as silently drops a *committed* find (T43):
	// editing is already false post-Enter, so the check above alone misses
	// it, but applyCurrentMatch's indication and FindNext/FindPrevious's
	// traversal are exactly as lost as an in-progress query would be.
	// hasFindMatches() catches that case the same way, checked after the
	// editing case so a find that was both editing and matched (the normal
	// case) reports the more specific "search dismissed" message once, not
	// both. Either message leads with cause (T45) rather than assuming a
	// resize: toggleNavigator can in fact reach the committed-find branch
	// (unlike the mid-edit one above, nothing gates it), so the message must
	// name whichever of the two actually fired.
	findWasEditing := m.viewer.findEditing()
	findHadMatches := m.viewer.hasFindMatches()
	m.viewer = m.viewer.setSize(viewerContent.width, viewerContent.height)
	switch {
	case findWasEditing && !m.viewer.findEditing():
		m.message = fmt.Sprintf("%s; search dismissed", cause)
	case findHadMatches && !m.viewer.hasFindMatches():
		m.message = fmt.Sprintf("%s; search cleared", cause)
	}
	filesContent := l.files.content()
	m.files = m.files.setSize(filesContent.width, filesContent.height)
	rootsContent := l.roots.content()
	m.roots = m.roots.setSize(rootsContent.width, rootsContent.height)
	return m
}

// layout is the frame's regions at the model's current terminal size and
// Navigator visibility (S9).
func (m model) layout() layout {
	return computeLayout(m.width, m.height, m.navigatorHidden)
}

// visible reports whether a pane has a region to draw in at the current size.
// A terminal too narrow for the Navigator column, or the Navigator toggle
// (S9) hiding it, both hide its panes (see computeLayout), and focus must
// not land in one.
func (m model) visible(p pane) bool {
	l := m.layout()
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
			Foreground(dimColor)
)

// View returns the frame. Bubble Tea v2 derives terminal state (alt screen,
// cursor, mouse mode, window title) from the returned View on every render, so
// every field left at its zero value is a deliberate "off": noggin draws inline
// with no cursor, exactly as it did under v1.
func (m model) View() tea.View {
	return tea.NewView(m.render())
}

// render is the frame (frame()) plus, when the Help Modal or the Picker is
// open, its overlay on top (overlayHelp / overlayPicker) — the one place
// that composition happens, so View() itself stays a one-liner. The two
// modals are mutually exclusive by construction (Update's handleHelpInput/
// handlePickerInput gating, model.go's Update), so at most one overlay ever
// applies.
func (m model) render() string {
	frame := m.frame()
	switch {
	case m.width <= 0 || m.height <= 0:
		return frame
	case m.helpOpen:
		return overlayHelp(frame, m.width, m.height, m.focus)
	case m.picker.open:
		return overlayPicker(frame, m.width, m.height, m.picker)
	default:
		return frame
	}
}

// frame composes the three regions into the string View wraps.
func (m model) frame() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	l := m.layout()

	body := m.renderPane(viewerPane, l.viewer, m.viewer.view())
	if l.navigatorWidth() > 0 {
		navigator := lipgloss.JoinVertical(lipgloss.Left,
			m.renderPane(rootsPane, l.roots, m.roots.view()),
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
	style := borderStyle
	if m.focus == p {
		style = focusedBorderStyle
	}
	return renderBox(style, r, m.paneTitle(p), content)
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

// renderFooter draws the footer line: an in-progress find query (M5.1)
// first, since it needs the keyboard and so must be shown; otherwise a
// transient message (D22) when one is set; otherwise the open Picker's own
// hints (pickerHintText, M2.3) when it is open, or the focused pane's key
// hints (hintText) otherwise, truncated to fit (S11). Update clears
// m.message on every keypress, so a message shown here never survives past
// the next key.
//
// The find prompt is not run through footerStyle's dim foreground: it is
// already a complete rendering (textinput.Model.View(), cursor included) and
// muting its own styling would fight it rather than match the other two
// cases, which are plain, unstyled text until footerStyle colours them.
func (m model) renderFooter(r rect) string {
	if m.viewer.findEditing() {
		return lipgloss.NewStyle().Width(r.width).MaxWidth(r.width).MaxHeight(r.height).
			Render(m.viewer.findView(r.width))
	}
	text := hintText(m.focus, r.width)
	// The open Picker (S2, M2.3) owns the keyboard the same exclusive way a
	// find-in-progress does, but unlike Find it does not take over the
	// footer row itself (its own Prompt already renders inside the modal) --
	// it only swaps which hints the footer shows, from the pane underneath's
	// to the Picker's own.
	if m.picker.open {
		text = pickerHintText(r.width)
	}
	if m.message != "" {
		text = m.message
	}
	return footerStyle.Width(r.width).MaxWidth(r.width).MaxHeight(r.height).Render(text)
}
