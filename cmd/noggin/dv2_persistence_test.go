package main

import (
	"os"
	"path/filepath"
	"testing"
)

// C3/M4.1 (S8): DV2 is session-only. store.go's roots file — one absolute
// starred path per line, written only by saveStore/toggleStore — is the only
// thing noggin ever writes to disk (grep confirms no other os.WriteFile call
// exists in the package); DV2's existence, Active Document and Peek/Find
// state have no persistence path to leak through in the first place. This
// test pins that structurally: it drives DV2 through document, Peek and Find
// state, confirms the store file on disk is untouched by any of it, and
// confirms a fresh model relaunched against that same store — the real
// main.go sequence (loadStarredRoots -> newModel().withStorePath(...).
// applyResolution(resolveLaunch(...))), with loadStore standing in for
// loadStarredRoots since dataDir() (M1.1/S1) is not under test control —
// comes up with no DV2 at all, while the starred root gathered across DV2's
// whole lifetime survives exactly as a real restart would replay it.

// launchFixture is one simulated process: reads storePath the way main.go
// does, resolves the launch argument (root, always) and applies it.
func launchFixture(t *testing.T, storePath, root string) model {
	t.Helper()
	starred := loadStore(storePath)
	res := resolveLaunch([]string{root}, starred, filepath.Dir(root), nil)
	return sized(t, newModel().withStorePath(storePath).applyResolution(res), 80, 24)
}

func TestDV2NeverPersistedAcrossRestart(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "roots")
	root := mkdir(t, filepath.Join(dir, "root"))
	activePath := writeFileContent(t, filepath.Join(root, "active.md"), "# Active\n")
	sidePath := writeFileContent(t, filepath.Join(root, "side.md"), "Side\nmarker line\n")

	// "First process": open the root, load DV1's Active Document.
	m := launchFixture(t, storePath, root)
	doc, err := loadDocument(activePath)
	if err != nil {
		t.Fatalf("loadDocument(active): %v", err)
	}
	m = m.openDocument(doc)

	// Persisted state gathered during this same session, before DV2 exists:
	// star the root via the real model action (not a direct toggleStore
	// call), so the store file reflects an actual product action.
	m = m.focusPane(rootsPane)
	m = m.toggleStar()
	if m.message != "" {
		t.Fatalf("toggleStar: unexpected message %q", m.message)
	}
	preDV2Store, err := os.ReadFile(storePath)
	if err != nil {
		t.Fatalf("read store after star: %v", err)
	}

	// Now create DV2 with its own Active Document, then drive it into Peek
	// and committed Find state — every kind of DV2 state S8 says must not
	// persist.
	side, err := loadDocument(sidePath)
	if err != nil {
		t.Fatalf("loadDocument(side): %v", err)
	}
	m = m.openDocumentInto(dv2Pane, side)
	if m.dv2 == nil {
		t.Fatal("setup: DV2 was not created")
	}
	m = press(t, m, "/")
	m = typeText(t, m, "marker")
	m = press(t, m, "enter")
	if !m.dv2.hasFindMatches() {
		t.Fatal("setup: DV2's find did not commit")
	}

	// DV2's whole lifetime above must not have touched the store file at
	// all: it is byte-for-byte the same as right after the star, before DV2
	// existed.
	postDV2Store, err := os.ReadFile(storePath)
	if err != nil {
		t.Fatalf("read store after DV2 activity: %v", err)
	}
	if string(postDV2Store) != string(preDV2Store) {
		t.Errorf("store file changed while DV2 was active:\nbefore: %q\nafter:  %q",
			preDV2Store, postDV2Store)
	}

	// "Restart": a brand new model against the same on-disk store and the
	// same launch argument.
	restarted := launchFixture(t, storePath, root)

	if restarted.dv2 != nil {
		t.Error("DV2 exists in a freshly relaunched model — it was persisted")
	}
	if restarted.dv1.hasFindMatches() || restarted.dv1.findEditing() {
		t.Error("Find state leaked into the restarted model's DV1")
	}
	if restarted.dv1.peeking() {
		t.Error("Peek state leaked into the restarted model's DV1")
	}

	// The starred root, however, is exactly what a real restart should
	// still see — DV2 having existed must not have disturbed the roots
	// state that genuinely is persisted (S2).
	var found bool
	for _, r := range restarted.roots.rootList() {
		if r.path == root {
			found = true
			if !r.starred {
				t.Error("root is present after restart but not marked starred")
			}
		}
	}
	if !found {
		t.Fatalf("restarted model has no root for %s", root)
	}
}
