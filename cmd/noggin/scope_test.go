package main

import (
	"path/filepath"
	"testing"
)

// TestLiveRootsInScopeActiveRoot: Active Root scope yields exactly the one
// active root, ignoring the rest of the session's roots entirely.
func TestLiveRootsInScopeActiveRoot(t *testing.T) {
	a := mkdir(t, filepath.Join(t.TempDir(), "a"))
	b := mkdir(t, filepath.Join(t.TempDir(), "b"))
	roots := []root{{path: a}, {path: b}}

	got := liveRootsInScope(roots, 1, activeRootScope)

	if len(got) != 1 || got[0].path != b {
		t.Fatalf("liveRootsInScope(active=1, activeRootScope) = %v, want just %s", got, b)
	}
}

// TestLiveRootsInScopeActiveRootBroken: a broken Active Root yields no
// roots to search, not an error.
func TestLiveRootsInScopeActiveRootBroken(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "gone")
	roots := []root{{path: missing}}

	got := liveRootsInScope(roots, 0, activeRootScope)

	if got != nil {
		t.Fatalf("liveRootsInScope(broken active) = %v, want nil", got)
	}
}

// TestLiveRootsInScopeActiveRootNoneActive covers an empty session (active
// == -1): Active Root scope has nothing to search.
func TestLiveRootsInScopeActiveRootNoneActive(t *testing.T) {
	if got := liveRootsInScope(nil, -1, activeRootScope); got != nil {
		t.Fatalf("liveRootsInScope(no roots) = %v, want nil", got)
	}
}

// TestLiveRootsInScopeAllRootsSkipsBroken: All Roots covers every live root
// and omits a Broken one, in session order, without erroring (S6).
func TestLiveRootsInScopeAllRootsSkipsBroken(t *testing.T) {
	a := mkdir(t, filepath.Join(t.TempDir(), "a"))
	missing := filepath.Join(t.TempDir(), "gone")
	c := mkdir(t, filepath.Join(t.TempDir(), "c"))
	roots := []root{{path: a}, {path: missing}, {path: c}}

	got := liveRootsInScope(roots, 0, allRootsScope)

	if len(got) != 2 || got[0].path != a || got[1].path != c {
		t.Fatalf("liveRootsInScope(allRootsScope) = %v, want [%s %s]", got, a, c)
	}
}

// TestLiveRootsInScopeAllRootsEmpty: no roots at all is "nothing to search",
// not an error, matching the broken-active case's shape.
func TestLiveRootsInScopeAllRootsEmpty(t *testing.T) {
	if got := liveRootsInScope(nil, -1, allRootsScope); len(got) != 0 {
		t.Fatalf("liveRootsInScope(no roots, allRootsScope) = %v, want empty", got)
	}
}
