package main

import "fmt"

// cycleTheme is keys.Theme's action: make the next Theme in `themes` the
// active one and name it in the Footer. Every style is built from `theme` when
// it is used (theme.go), so the frame repaints in the new colours on its own;
// only the Document Viewers hold content rendered ahead of time, and restyle
// re-renders that.
func (m model) cycleTheme() model {
	theme = nextTheme(theme.Name)
	m.dv1 = m.dv1.restyle()
	if m.dv2 != nil {
		v := m.dv2.restyle()
		m.dv2 = &v
	}
	m.message = "theme: " + theme.Name
	if m.themePath != "" {
		if err := saveThemeChoice(m.themePath, theme.Name); err != nil {
			m.message += fmt.Sprintf(" (cannot save: %v)", err)
		}
	}
	return m
}

// restyle re-renders v's content (and Peek's) in the active Theme without
// moving anything. A Theme changes colours, never text or wrapping, so the
// scroll positions, a find and its current match all still address the same
// lines: they are carried across the re-render as they were, where render on
// its own would drop the find and toggleMode would re-run it.
func (v viewer) restyle() viewer {
	find := v.findState
	y, x, peekY := v.vp.YOffset(), v.vp.XOffset(), v.peek.vp.YOffset()
	v = v.render()
	v.findState = find
	// Re-draws the current match's indication over the new content; it may
	// scroll to the match, so the offsets are put back after it.
	v = v.applyCurrentMatch()
	v.vp.SetYOffset(y)
	v.vp.SetXOffset(x)
	if v.peeking() {
		v.peek.vp.SetYOffset(peekY)
	}
	return v
}
