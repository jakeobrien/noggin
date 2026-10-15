package main

// browseMode is the Files Nav's S5 Browse Mode (GLOSSARY.md "Browse Mode"):
// how the pane lists documents. Cycled by keys.FilesMode, in this fixed
// order (Tree -> Recent -> Starred -> Tree ...) while a Root is (or
// nothing is) the Active Source; M4.2/M4.3 give Recent and Starred their
// actual listings, M4.1 only the model, the cycle keys and the row that
// shows the current pair. While the Wikipedia Query Source is active
// (C6/M1.4) there is no tree to walk, so Tree drops out of the cycle
// entirely and it toggles directly between Recent and Starred instead
// (filesNav.cycleMode), landing on Starred rather than Tree the first time
// a session shows Wikipedia (filesNav.setWikipedia).
type browseMode int

const (
	treeBrowse browseMode = iota
	recentBrowse
	starredBrowse
)

// label is the mode's own name, exactly as the S5 mode/scope row shows it.
//
// While the Wikipedia Query Source is active (C6/M1.4) the mode field is
// never treeBrowse — cycleMode excludes it from the cycle entirely, since
// there is no tree to walk without a Root — so this default case is only
// ever reached by treeBrowse itself while a Root is (or nothing is) active.
func (m browseMode) label() string {
	switch m {
	case recentBrowse:
		return "Recent"
	case starredBrowse:
		return "Starred"
	default:
		return "Tree"
	}
}

// browseScope is the Files Nav's S5 Browse Scope (GLOSSARY.md "Browse
// Scope"): what a Browse Mode draws from. Distinct from a Picker's own
// pickerScope (scope.go) — same underlying idea, deliberately not shared,
// since the two axes have different value sets (Starred's three vs. a
// Picker's two, S8's future third notwithstanding) and evolve under
// different tasks.
//
// A single enum spans every mode's scope values rather than one type per
// mode: activeRootBrowseScope means the same thing (S5's "Active Source") whichever
// of Recent/Starred it is read against, so modeScopes below is what actually
// bounds which values are reachable from a given mode — a browseScope value
// meaningless for the current mode (e.g. allRootsBrowseScope while Starred) is
// simply never produced by cycleScope or read by scopeLabel, which special-
// cases treeBrowse before consulting the value at all.
type browseScope int

const (
	activeRootBrowseScope browseScope = iota
	allRootsBrowseScope
	allStarredBrowseScope
	orphansBrowseScope
)

// label is the scope's own name, exactly as the S5 mode/scope row shows it,
// for whichever mode is showing it (Tree never calls this — see
// filesNav.scopeLabel).
func (s browseScope) label() string {
	switch s {
	case allRootsBrowseScope:
		return "All Sources"
	case allStarredBrowseScope:
		return "All Starred"
	case orphansBrowseScope:
		return "Orphans"
	default:
		return "Active Source"
	}
}

// modeScopes is S5's own table of which Browse Scope values a Browse Mode
// cycles through, in cycle order: none for Tree (the scope key is a no-op,
// filesNav.cycleScope), two for Recent, three for Starred (Orphans is the
// pre-agreed cut 3 in PLAN.md/M4.3, but the model carries it regardless —
// dropping M4.3's *listing* of it, if the cycle spills, costs nothing here)
// — except while the Wikipedia Query Source is active (C6/M1.4), whose own
// Starred value set drops Orphans entirely (S5: "Starred: Active Source |
// All Starred"): Orphans names files outside any session Root, a question
// Wikipedia's own document kind (an article, never under a Root at all) has
// no useful answer to. wikipedia only ever changes Starred's own set; every
// other mode's cycle is identical either way.
func modeScopes(m browseMode, wikipedia bool) []browseScope {
	switch m {
	case recentBrowse:
		return []browseScope{activeRootBrowseScope, allRootsBrowseScope}
	case starredBrowse:
		if wikipedia {
			return []browseScope{activeRootBrowseScope, allStarredBrowseScope}
		}
		return []browseScope{activeRootBrowseScope, allStarredBrowseScope, orphansBrowseScope}
	default:
		return nil
	}
}

// flatEmptyText is M4.4's own per-combination wording (S6: "every mode/
// scope combination with nothing to show displays a brief message saying
// so"): each of Recent's and Starred's five scope pairs gets its own
// sentence naming what was searched, rather than recentEmptyText/
// starredEmptyText's pre-M4.4 placeholder pair that read identically across
// every scope of their own mode. Tree itself is not this function's
// concern — it has no scope (modeScopes(treeBrowse) is empty), and instead
// gets its own two-message split directly in filesNav.rowsView (M4.9,
// TRIAGE T22: emptyFilesText for no Root open, emptyTreeText for a Root
// open but empty); the caller here (filesNav.flatRowsView) only ever
// reaches this while flatMode(), i.e. mode is Recent or Starred — this
// func's own default case is unreachable in practice, just a defensive
// fallback. It returns emptyTreeText, not emptyFilesText (TRIAGE T34):
// emptyFilesText's "no Root open" case is already rowsView's own Tree
// branch's concern (its f.root == nil check, well before any call into
// this function could occur); a Tree-mode caller landing here regardless
// would only be asking this func the flat modes' own question — "nothing
// to show" — which is what emptyTreeText's "open but empty" wording
// names.
//
// Active Source's own two messages (Recent's default case, Starred's own
// default case) read "Source", not "Root": C6/M1.4 makes Wikipedia reachable
// there too (Starred's set keeps Active Source while Wikipedia is active,
// modeScopes above), so wording naming a Root specifically would misdescribe
// it. Orphans keeps naming a Root — it is unreachable while Wikipedia is
// active (modeScopes) and, unlike Active Source, names a Root-only concept
// even when it is reachable (a document outside any session Root).
func flatEmptyText(mode browseMode, scope browseScope) string {
	switch mode {
	case recentBrowse:
		if scope == allRootsBrowseScope {
			return "No recently viewed documents in any Source."
		}
		return "No recently viewed documents in this Source."
	case starredBrowse:
		switch scope {
		case allStarredBrowseScope:
			return "No starred documents."
		case orphansBrowseScope:
			return "No starred documents outside a session Root."
		default:
			return "No starred documents in this Source."
		}
	default:
		return emptyTreeText
	}
}
