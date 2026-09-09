package main

// This file is C6/M1.4's own filesNav-level coverage: Browse Modes/Scopes
// with the Wikipedia Query Source active, and per-Source browse memory
// across a Wikipedia -> Root -> Wikipedia switch. Model-level wiring
// (activateWikipediaSource/activateRoot) is covered alongside the rest of
// the Wikipedia Active Source model tests in wikipedia_source_test.go.

import (
	"strings"
	"testing"
	"time"
)

// TestWikipediaBrowseModeCycleIsStarredRecentOnly is S5's own
// Wikipedia-active wording: Browse Modes are Starred and Recent only, the
// first activation landing on Starred, and Tree never appears anywhere in
// the cycle -- there is no tree to walk without a Root.
func TestWikipediaBrowseModeCycleIsStarredRecentOnly(t *testing.T) {
	f := newFilesNav().setWikipedia()
	if f.mode != starredBrowse {
		t.Fatalf("first Wikipedia visit mode = %v, want starredBrowse", f.mode)
	}
	for i, want := range []browseMode{recentBrowse, starredBrowse, recentBrowse, starredBrowse} {
		f = f.cycleMode()
		if f.mode != want {
			t.Errorf("mode after cycle %d = %v, want %v", i+1, f.mode, want)
		}
		if f.mode == treeBrowse {
			t.Fatalf("cycle %d landed on treeBrowse while Wikipedia is active", i+1)
		}
	}
}

// TestWikipediaStarredScopeExcludesOrphans checks Starred's own
// Wikipedia-active value set (S5: "Starred: Active Source | All Starred"):
// two values, wrapping, Orphans never appearing -- it names files outside
// any session Root, a question an article has no answer to.
func TestWikipediaStarredScopeExcludesOrphans(t *testing.T) {
	f := newFilesNav().setWikipedia() // first visit starts on Starred
	if f.mode != starredBrowse {
		t.Fatalf("setup: mode = %v, want starredBrowse", f.mode)
	}
	want := []string{"Active Source", "All Starred", "Active Source"}
	for i, w := range want {
		if got := f.scopeLabel(); got != w {
			t.Errorf("Wikipedia Starred scope at step %d = %q, want %q", i, got, w)
		}
		f = f.cycleScope()
	}
}

// TestWikipediaRecentScopeIsActiveSourceAllSources checks Recent's own
// Wikipedia-active value set (S5: "Recent: Active Source | All Sources"),
// the same two values a Root-active session cycles, just reachable now with
// Wikipedia as the Active Source too.
func TestWikipediaRecentScopeIsActiveSourceAllSources(t *testing.T) {
	f := newFilesNav().setWikipedia().cycleMode() // Starred -> Recent
	if f.mode != recentBrowse {
		t.Fatalf("setup: mode = %v, want recentBrowse", f.mode)
	}
	want := []string{"Active Source", "All Sources", "Active Source"}
	for i, w := range want {
		if got := f.scopeLabel(); got != w {
			t.Errorf("Wikipedia Recent scope at step %d = %q, want %q", i, got, w)
		}
		f = f.cycleScope()
	}
}

// TestWikipediaModeScopeRowShowsTheCurrentPair mirrors
// TestFilesModeScopeRowShowsTheCurrentPair for the Wikipedia-active case:
// the mode/scope row always names the actual current pair, Starred first.
func TestWikipediaModeScopeRowShowsTheCurrentPair(t *testing.T) {
	f := newFilesNav().setWikipedia().setSize(30, 8)
	cases := []struct{ mode, scope string }{
		{"Starred", "Active Source"},
		{"Recent", "Active Source"},
		{"Starred", "Active Source"},
	}
	for _, c := range cases {
		lines := strings.Split(f.view(), "\n")
		if !strings.Contains(lines[0], c.mode) || !strings.Contains(lines[0], c.scope) {
			t.Errorf("mode/scope row = %q, want it to show %q and %q", lines[0], c.mode, c.scope)
		}
		f = f.cycleMode()
	}
}

// TestWikipediaRecentModeRendersArticleRowByTitle is C6/M6.1's own
// filesNav-level wiring check, mirroring TestFilesNavRecentModeRendersFlatRows:
// with Wikipedia the Active Source, Recent mode's Active Source scope
// renders a recently-viewed article by its title, labelled Wikipedia
// (renderFlatRow's rootLabel column), never a URL-derived name.
func TestWikipediaRecentModeRendersArticleRowByTitle(t *testing.T) {
	article := fixtureArticle()
	docs := map[string]docRecord{
		article.path: {LastViewed: time.Now(), Article: true, Title: article.title},
	}
	f := newFilesNav().setWikipedia().withRecentSource(docs, nil, "").setSize(40, 8)
	f = f.cycleMode() // Starred -> Recent
	if f.mode != recentBrowse {
		t.Fatalf("setup: mode = %v, want recentBrowse", f.mode)
	}

	view := f.view()
	if !strings.Contains(view, article.title) {
		t.Errorf("Wikipedia Recent view does not show the article's title:\n%s", view)
	}
	if !strings.Contains(view, wikipediaLabel) {
		t.Errorf("Wikipedia Recent view does not label the row %q:\n%s", wikipediaLabel, view)
	}
}

// TestFilesNavMemoryAcrossWikipediaRootWikipediaSwitch is C6/M1.4's own
// per-Source browse memory clause: a mode/scope pair left on Wikipedia
// survives a detour through a Root and back, and the Root's own pair (set
// independently, before the Wikipedia round trip) is itself untouched by
// Wikipedia ever having been visited in between -- the same "per Source,
// not shared" guarantee TestBrowseMemoryModePerRootScopePerRootMode already
// gives two Roots, now covering Wikipedia's own memory key
// (wikipediaSourceKey) too.
func TestFilesNavMemoryAcrossWikipediaRootWikipediaSwitch(t *testing.T) {
	root := fixtureTree()

	f := newFilesNav().setWikipedia()
	f = f.cycleMode()  // Starred -> Recent
	f = f.cycleScope() // Active Source -> All Sources
	if f.mode != recentBrowse || f.scope != allRootsBrowseScope {
		t.Fatalf("setup: Wikipedia's pair = (%v, %v), want (Recent, All Sources)", f.mode, f.scope)
	}

	// The Root has never been shown: starts fresh on Tree, not carrying
	// Wikipedia's own pair over by accident.
	f = f.setRoot(root)
	if f.mode != treeBrowse {
		t.Fatalf("Root's first visit mode = %v, want treeBrowse", f.mode)
	}
	f = f.cycleMode() // Tree -> Recent, and leave it there with its own scope
	if f.mode != recentBrowse || f.scope != activeRootBrowseScope {
		t.Fatalf("setup: Root's pair = (%v, %v), want (Recent, Active Source)", f.mode, f.scope)
	}

	// Back to Wikipedia: its own pair, not the Root's fresh-visit default.
	f = f.setWikipedia()
	if !f.wikipedia || f.root != nil {
		t.Fatalf("setWikipedia did not restore Wikipedia: wikipedia=%v root=%v", f.wikipedia, f.root)
	}
	if f.mode != recentBrowse {
		t.Errorf("Wikipedia's mode after Wikipedia\u2192Root\u2192Wikipedia = %v, want recentBrowse", f.mode)
	}
	if f.scope != allRootsBrowseScope {
		t.Errorf("Wikipedia's scope after Wikipedia\u2192Root\u2192Wikipedia = %v, want allRootsBrowseScope", f.scope)
	}

	// And the Root again: its own pair, unaffected by the Wikipedia visit
	// in between.
	f = f.setRoot(root)
	if f.mode != recentBrowse || f.scope != activeRootBrowseScope {
		t.Errorf("Root's pair after the round trip = (%v, %v), want (Recent, Active Source)", f.mode, f.scope)
	}
}
