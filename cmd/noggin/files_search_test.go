package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"testing"
)

// requireRipgrep skips a test when the real ripgrep binary isn't on PATH.
// Unlike ripgrep_test.go's detection tests (which only need something
// resolvable by that name), these tests exercise rg's actual `--files`
// behaviour — a fake stand-in executable can't stand in for that, so a
// missing real rg means "can't verify this here", not "this fails".
func requireRipgrep(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath(ripgrepBinary); err != nil {
		t.Skip("ripgrep (rg) not on PATH; skipping rg-backed search tests")
	}
}

// writeFileContent creates a file with explicit content (unlike roots_test.go's
// writeFile, fixed to "# note\n"), for fixtures where the content itself is
// what's under test (a .gitignore's pattern lines).
func writeFileContent(t *testing.T, path, content string) string {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return path
}

// filePaths is the path of each pickerFile, sorted, so a result set can be
// asserted as a set rather than an order (rg's own --files order is not a
// contract listFiles makes any promise about — M1.3 is what orders results).
func filePaths(files []pickerFile) []string {
	paths := make([]string, len(files))
	for i, f := range files {
		paths[i] = f.path
	}
	sort.Strings(paths)
	return paths
}

// TestListFilesAppliesReadableAllowlist: only allowlist extensions/names
// come back; everything else rg finds is dropped (S13).
func TestListFilesAppliesReadableAllowlist(t *testing.T) {
	requireRipgrep(t)
	dir := t.TempDir()
	notesMd := writeFile(t, filepath.Join(dir, "notes.md"))
	readme := writeFile(t, filepath.Join(dir, "README"))
	writeFile(t, filepath.Join(dir, "main.go"))
	writeFile(t, filepath.Join(dir, "image.png"))
	roots := []root{{path: dir}}

	got := filePaths(listFiles(roots, 0, activeRootScope))

	want := sortedCopy([]string{notesMd, readme})
	if !equalStrings(got, want) {
		t.Fatalf("listFiles() = %v, want %v", got, want)
	}
}

// TestListFilesOmitsGitignoredAndHidden: a .gitignore'd file and hidden
// dotfiles/dot-directories are absent, even though they're Readable by
// extension — rg's own ignore rules decide this, ahead of the allowlist
// (S14).
func TestListFilesOmitsGitignoredAndHidden(t *testing.T) {
	requireRipgrep(t)
	dir := t.TempDir()
	kept := writeFile(t, filepath.Join(dir, "kept.md"))
	writeFileContent(t, filepath.Join(dir, ".gitignore"), "ignored.md\n")
	writeFile(t, filepath.Join(dir, "ignored.md"))
	writeFile(t, filepath.Join(dir, ".hidden.md"))
	mkdir(t, filepath.Join(dir, ".hiddendir"))
	writeFile(t, filepath.Join(dir, ".hiddendir", "inside.md"))
	roots := []root{{path: dir}}

	got := filePaths(listFiles(roots, 0, activeRootScope))

	want := []string{kept}
	if !equalStrings(got, want) {
		t.Fatalf("listFiles() = %v, want %v", got, want)
	}
}

// TestListFilesAllRootsCoversLiveRootsAndSkipsBroken: All Roots scope finds
// Readable files across every live Root and silently omits a Broken one,
// without an error (S6).
func TestListFilesAllRootsCoversLiveRootsAndSkipsBroken(t *testing.T) {
	requireRipgrep(t)
	a := mkdir(t, filepath.Join(t.TempDir(), "a"))
	aFile := writeFile(t, filepath.Join(a, "one.md"))
	b := mkdir(t, filepath.Join(t.TempDir(), "b"))
	bFile := writeFile(t, filepath.Join(b, "two.txt"))
	missing := filepath.Join(t.TempDir(), "gone")
	roots := []root{{path: a}, {path: missing}, {path: b}}

	got := filePaths(listFiles(roots, 0, allRootsScope))

	want := sortedCopy([]string{aFile, bFile})
	if !equalStrings(got, want) {
		t.Fatalf("listFiles(allRootsScope) = %v, want %v", got, want)
	}
}

// TestListFilesActiveRootScopeIgnoresOtherRoots: Active Root scope searches
// only the active Root, even when other live Roots exist in the session.
func TestListFilesActiveRootScopeIgnoresOtherRoots(t *testing.T) {
	requireRipgrep(t)
	a := mkdir(t, filepath.Join(t.TempDir(), "a"))
	writeFile(t, filepath.Join(a, "one.md"))
	b := mkdir(t, filepath.Join(t.TempDir(), "b"))
	bFile := writeFile(t, filepath.Join(b, "two.md"))
	roots := []root{{path: a}, {path: b}}

	got := filePaths(listFiles(roots, 1, activeRootScope))

	want := []string{bFile}
	if !equalStrings(got, want) {
		t.Fatalf("listFiles(active=1) = %v, want %v", got, want)
	}
}

// TestListFilesEachResultCarriesItsRoot: a result records which Root it
// came from, not just its own path — M3.1 needs this to show "Root + path
// relative to that Root" for same-named files in different Roots.
func TestListFilesEachResultCarriesItsRoot(t *testing.T) {
	requireRipgrep(t)
	a := mkdir(t, filepath.Join(t.TempDir(), "a"))
	writeFile(t, filepath.Join(a, "same.md"))
	b := mkdir(t, filepath.Join(t.TempDir(), "b"))
	writeFile(t, filepath.Join(b, "same.md"))
	roots := []root{{path: a}, {path: b}}

	files := listFiles(roots, 0, allRootsScope)

	if len(files) != 2 {
		t.Fatalf("listFiles() returned %d files, want 2", len(files))
	}
	for _, f := range files {
		if f.root.path != a && f.root.path != b {
			t.Errorf("file %s carries root %q, want %s or %s", f.path, f.root.path, a, b)
		}
	}
}

func sortedCopy(s []string) []string {
	out := append([]string(nil), s...)
	sort.Strings(out)
	return out
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
