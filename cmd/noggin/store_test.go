package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// TestLoadStoreMissingFile: a store file that has never been written yields
// an empty list, not an error (S2).
func TestLoadStoreMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "roots")
	got := loadStore(path)
	if len(got) != 0 {
		t.Errorf("loadStore(missing) = %v, want empty", got)
	}
}

// TestLoadStoreGarbageFile: unparseable content — no line is an absolute
// path — yields an empty list, not an error (S2).
func TestLoadStoreGarbageFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "roots")
	garbage := []byte("not a path\n\x00\x01binary noise\n---\nrelative/path\n")
	if err := os.WriteFile(path, garbage, 0o644); err != nil {
		t.Fatal(err)
	}
	got := loadStore(path)
	if len(got) != 0 {
		t.Errorf("loadStore(garbage) = %v, want empty", got)
	}
}

// TestLoadStoreIgnoresCommentsAndBlanks checks the hand-editable format: blank
// lines and `#` comments are skipped, real entries survive.
func TestLoadStoreIgnoresCommentsAndBlanks(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a")
	b := filepath.Join(dir, "b")
	path := filepath.Join(dir, "roots")
	content := "# starred roots\n\n" + a + "\n\n# another\n" + b + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	got := loadStore(path)
	want := []string{a, b}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("loadStore = %v, want %v", got, want)
	}
}

// TestLoadStoreKeepsMissingPaths: a stored path that no longer exists on disk
// is still loaded, not pruned (S2) — loadStore never stats.
func TestLoadStoreKeepsMissingPaths(t *testing.T) {
	dir := t.TempDir()
	gone := filepath.Join(dir, "gone")
	path := filepath.Join(dir, "roots")
	if err := saveStore(path, []string{gone}); err != nil {
		t.Fatal(err)
	}
	got := loadStore(path)
	want := []string{gone}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("loadStore = %v, want %v (unpruned)", got, want)
	}
}

// TestSaveStoreCreatesConfigDir: saveStore makes the parent directory when it
// does not exist yet, so a first-ever star does not fail.
func TestSaveStoreCreatesConfigDir(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "noggin", "roots")
	root := filepath.Join(t.TempDir(), "a")
	if err := saveStore(path, []string{root}); err != nil {
		t.Fatalf("saveStore: %v", err)
	}
	got := loadStore(path)
	if want := []string{root}; !reflect.DeepEqual(got, want) {
		t.Errorf("loadStore after save = %v, want %v", got, want)
	}
}

// TestToggleStorePersistsAcrossReload is the acceptance criterion in prose: a
// star survives a process restart. toggleStore writes through immediately, so
// a fresh loadStore call (standing in for a new process) sees it.
func TestToggleStorePersistsAcrossReload(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "project")
	path := filepath.Join(dir, "roots")

	updated, err := toggleStore(path, root)
	if err != nil {
		t.Fatalf("toggleStore(star): %v", err)
	}
	if want := []string{root}; !reflect.DeepEqual(updated, want) {
		t.Fatalf("toggleStore(star) = %v, want %v", updated, want)
	}

	// A fresh load, as a new process would perform on launch.
	reloaded := loadStore(path)
	if want := []string{root}; !reflect.DeepEqual(reloaded, want) {
		t.Fatalf("loadStore after restart = %v, want %v", reloaded, want)
	}
}

// TestToggleStoreUnstars checks the other half of "toggle": starring an
// already-starred path removes it, and that removal is written through too.
func TestToggleStoreUnstars(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a")
	b := filepath.Join(dir, "b")
	path := filepath.Join(dir, "roots")
	if err := saveStore(path, []string{a, b}); err != nil {
		t.Fatal(err)
	}

	updated, err := toggleStore(path, a)
	if err != nil {
		t.Fatalf("toggleStore(unstar): %v", err)
	}
	want := []string{b}
	if !reflect.DeepEqual(updated, want) {
		t.Fatalf("toggleStore(unstar) = %v, want %v", updated, want)
	}
	if reloaded := loadStore(path); !reflect.DeepEqual(reloaded, want) {
		t.Fatalf("loadStore after unstar = %v, want %v", reloaded, want)
	}
}

// TestLoadStarredRootsNoDataDir checks the main.go entry point degrades to
// "no starred roots" rather than panicking when the data directory cannot be
// resolved. dataDir() consults $HOME/$XDG_DATA_HOME on unix; clearing both
// forces its error path.
func TestLoadStarredRootsNoDataDir(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("HOME", "")
	got := loadStarredRoots()
	if len(got) != 0 {
		t.Errorf("loadStarredRoots() = %v, want empty when data dir is unresolvable", got)
	}
}

// TestDataDirUsesXDGDataHome pins dataDir()'s Unix resolution: an explicit
// $XDG_DATA_HOME wins over $HOME/.local/share.
func TestDataDirUsesXDGDataHome(t *testing.T) {
	want := t.TempDir()
	t.Setenv("XDG_DATA_HOME", want)
	got, err := dataDir()
	if err != nil {
		t.Fatalf("dataDir: %v", err)
	}
	if got != want {
		t.Errorf("dataDir() = %q, want %q", got, want)
	}
}

// TestDataDirFallsBackToHome pins dataDir()'s fallback when $XDG_DATA_HOME is
// unset: $HOME/.local/share.
func TestDataDirFallsBackToHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("HOME", home)
	got, err := dataDir()
	if err != nil {
		t.Fatalf("dataDir: %v", err)
	}
	want := filepath.Join(home, ".local", "share")
	if got != want {
		t.Errorf("dataDir() = %q, want %q", got, want)
	}
}

// TestStoreIgnoresConfigDirCleanBreak is M1.1's acceptance criterion: a
// populated config-directory store (the pre-C4/D7 location) is never read,
// and starring through the current data-dir-resolved storePath round-trips
// across a fresh load exactly as a new process would see it.
func TestStoreIgnoresConfigDirCleanBreak(t *testing.T) {
	configHome := t.TempDir()
	dataHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("XDG_DATA_HOME", dataHome)

	// Populate the old config-dir store directly — simulating a pre-existing
	// install — without going through any of this package's data-dir-based
	// helpers.
	configStore := filepath.Join(configHome, storeSubdir, storeName)
	if err := os.MkdirAll(filepath.Dir(configStore), 0o755); err != nil {
		t.Fatal(err)
	}
	stale := filepath.Join(configHome, "stale-root")
	if err := os.WriteFile(configStore, []byte(stale+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// The data-dir store is empty (never written). loadStarredRoots must see
	// zero starred roots, not the config-dir file's contents.
	if got := loadStarredRoots(); len(got) != 0 {
		t.Fatalf("loadStarredRoots() = %v, want empty (config-dir store must be ignored)", got)
	}

	// Star/un-star round-trips through the data-dir path.
	path, err := storePath()
	if err != nil {
		t.Fatalf("storePath: %v", err)
	}
	want := filepath.Join(dataHome, storeSubdir, storeName)
	if path != want {
		t.Fatalf("storePath() = %q, want %q", path, want)
	}

	root := filepath.Join(dataHome, "a-root")
	if _, err := toggleStore(path, root); err != nil {
		t.Fatalf("toggleStore(star): %v", err)
	}
	if got := loadStarredRoots(); !reflect.DeepEqual(got, []string{root}) {
		t.Fatalf("loadStarredRoots() after star = %v, want [%s]", got, root)
	}

	if _, err := toggleStore(path, root); err != nil {
		t.Fatalf("toggleStore(unstar): %v", err)
	}
	if got := loadStarredRoots(); len(got) != 0 {
		t.Fatalf("loadStarredRoots() after unstar = %v, want empty", got)
	}
}
