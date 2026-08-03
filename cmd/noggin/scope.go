package main

// pickerScope selects which Roots a Picker searches (S6, GLOSSARY.md
// "Scope"): the session's single Active Root, or every Root at once. Shared
// by the File Picker's listing (M1.2) and the Grep Picker's search (M1.4),
// and by the in-Picker switch between them (M5.2) later.
type pickerScope int

const (
	activeRootScope pickerScope = iota
	allRootsScope
)

// liveRootsInScope resolves scope against the session's roots and which one
// is active (the same (roots, active) shape resolution and rootsNav already
// carry), dropping a Broken Root silently rather than erroring (S6: "All
// Roots silently skips Broken Roots — no error row"). "Broken" reuses
// pathMissing (rootsnav.go), the session's one existing definition of it,
// rather than a second stat-based check with its own opinion.
//
// Active Root scope with no active root (active out of range, e.g. an empty
// session) yields no roots to search, the same "nothing to search" outcome
// an All-Roots session with zero live roots gives.
func liveRootsInScope(roots []root, active int, scope pickerScope) []root {
	if scope == activeRootScope {
		if active < 0 || active >= len(roots) || pathMissing(roots[active].path) {
			return nil
		}
		return roots[active : active+1]
	}
	live := make([]root, 0, len(roots))
	for _, r := range roots {
		if !pathMissing(r.path) {
			live = append(live, r)
		}
	}
	return live
}
