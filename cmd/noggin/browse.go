package main

// browseMode is the Files Nav's S5 Browse Mode (GLOSSARY.md "Browse Mode"):
// how the pane lists documents. Cycled by keys.FilesMode, in this fixed
// order (Tree -> Recent -> Starred -> Tree ...); M4.2/M4.3 give Recent and
// Starred their actual listings, M4.1 only the model, the cycle keys and the
// row that shows the current pair.
type browseMode int

const (
	treeBrowse browseMode = iota
	recentBrowse
	starredBrowse
)

// label is the mode's own name, exactly as the S5 mode/scope row shows it.
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
// mode: activeRootBrowseScope means the same thing (S5's "Active Root") whichever
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
		return "All Roots"
	case allStarredBrowseScope:
		return "All Starred"
	case orphansBrowseScope:
		return "Orphans"
	default:
		return "Active Root"
	}
}

// modeScopes is S5's own table of which Browse Scope values a Browse Mode
// cycles through, in cycle order: none for Tree (the scope key is a no-op,
// filesNav.cycleScope), two for Recent, three for Starred (Orphans is the
// pre-agreed cut 3 in PLAN.md/M4.3, but the model carries it regardless —
// dropping M4.3's *listing* of it, if the cycle spills, costs nothing here).
func modeScopes(m browseMode) []browseScope {
	switch m {
	case recentBrowse:
		return []browseScope{activeRootBrowseScope, allRootsBrowseScope}
	case starredBrowse:
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
// concern — it has no scope (modeScopes(treeBrowse) is empty), and keeps
// emptyFilesText, whose "no root to browse" wording already covers Tree's
// one empty case; the caller (filesNav.flatRowsView) only ever reaches
// this while flatMode(), i.e. mode is Recent or Starred.
func flatEmptyText(mode browseMode, scope browseScope) string {
	switch mode {
	case recentBrowse:
		if scope == allRootsBrowseScope {
			return "No recently viewed documents in any Root."
		}
		return "No recently viewed documents in this Root."
	case starredBrowse:
		switch scope {
		case allStarredBrowseScope:
			return "No starred documents."
		case orphansBrowseScope:
			return "No starred documents outside a session Root."
		default:
			return "No starred documents in this Root."
		}
	default:
		return emptyFilesText
	}
}
