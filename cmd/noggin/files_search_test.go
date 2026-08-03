package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
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

// pickerFileAt builds a pickerFile from a bare path for fuzzyFiles tests —
// no rg, no fixture tree, since fuzzyFiles is a pure function over whatever
// listFiles would have produced (M1.2 already covers how that set is built).
func pickerFileAt(path string) pickerFile {
	return pickerFile{root: root{path: filepath.Dir(path)}, path: path}
}

// TestFuzzyFilesRanksStrongerNameMatchAbove: a file name that matches the
// query tightly outranks one that only matches loosely, regardless of the
// order they were supplied in (S3: "ordered by match strength").
func TestFuzzyFilesRanksStrongerNameMatchAbove(t *testing.T) {
	weak := pickerFileAt("/notes/xnyotyeysz.md") // scattered, weak match
	strong := pickerFileAt("/notes/notes.md")    // contiguous, strong match
	files := []pickerFile{weak, strong}

	results, truncated := fuzzyFiles(files, "notes")

	if truncated {
		t.Fatalf("fuzzyFiles() truncated = true, want false")
	}
	if len(results) != 2 || results[0].path != strong.path || results[1].path != weak.path {
		t.Fatalf("fuzzyFiles() = %v, want [%s, %s]", filePaths(results), strong.path, weak.path)
	}
}

// TestFuzzyFilesIgnoresPathComponents: a file whose directory contains the
// query but whose base name does not is not matched, while a file whose
// base name matches is found even though nothing in its directory does
// (D31: "the File Picker matches the file name only, never path
// components").
func TestFuzzyFilesIgnoresPathComponents(t *testing.T) {
	pathOnly := pickerFileAt("/journal/unrelated.md") // dir matches, name does not
	nameOnly := pickerFileAt("/misc/journal.md")      // name matches, dir does not
	files := []pickerFile{pathOnly, nameOnly}

	results, _ := fuzzyFiles(files, "journal")

	if len(results) != 1 || results[0].path != nameOnly.path {
		t.Fatalf("fuzzyFiles() = %v, want only %s", filePaths(results), nameOnly.path)
	}
}

// TestFuzzyFilesCapsAndReportsTruncation: a result set larger than
// pickerResultCap comes back capped at that length with the truncation
// flag set; a set at or under the cap comes back untouched and unflagged
// (S12).
func TestFuzzyFilesCapsAndReportsTruncation(t *testing.T) {
	var many []pickerFile
	for i := 0; i < pickerResultCap+50; i++ {
		many = append(many, pickerFileAt(filepath.Join("/notes", fmt.Sprintf("match-%03d.md", i))))
	}

	results, truncated := fuzzyFiles(many, "match")

	if !truncated {
		t.Fatalf("fuzzyFiles() truncated = false, want true")
	}
	if len(results) != pickerResultCap {
		t.Fatalf("fuzzyFiles() returned %d results, want %d", len(results), pickerResultCap)
	}

	results, truncated = fuzzyFiles(many[:pickerResultCap], "match")
	if truncated {
		t.Fatalf("fuzzyFiles() truncated = true at exactly the cap, want false")
	}
	if len(results) != pickerResultCap {
		t.Fatalf("fuzzyFiles() returned %d results, want %d", len(results), pickerResultCap)
	}
}

// grepLines is the (path, line, text) of each grepResult, sorted by path
// then line, so a result set can be asserted as a set rather than rg's own
// per-file argument order.
func grepLines(results []grepResult) []string {
	lines := make([]string, len(results))
	for i, r := range results {
		lines[i] = fmt.Sprintf("%s:%d:%s", r.file.path, r.line, r.text)
	}
	sort.Strings(lines)
	return lines
}

// TestGrepFilesMatchesLiterally: a query containing regex metacharacters
// matches only their literal occurrence, not what they'd mean as regex
// (S4, D29: --fixed-strings).
func TestGrepFilesMatchesLiterally(t *testing.T) {
	requireRipgrep(t)
	dir := t.TempDir()
	writeFileContent(t, filepath.Join(dir, "notes.md"), "see v1.2 here\nnot v1x2 though\n")
	roots := []root{{path: dir}}

	results, truncated := grepFiles(roots, 0, activeRootScope, "v1.2")

	if truncated {
		t.Fatalf("grepFiles() truncated = true, want false")
	}
	want := []string{filepath.Join(dir, "notes.md") + ":1:see v1.2 here"}
	if got := grepLines(results); !equalStrings(got, want) {
		t.Fatalf("grepFiles(\"v1.2\") = %v, want %v", got, want)
	}
}

// TestGrepFilesIsSmartCase: an all-lowercase query matches mixed-case text,
// but a query containing uppercase matches only that exact case (S4,
// matching find.go's S8 smart-case semantic via rg's own --smart-case).
func TestGrepFilesIsSmartCase(t *testing.T) {
	requireRipgrep(t)
	dir := t.TempDir()
	writeFileContent(t, filepath.Join(dir, "notes.md"), "Hello World\n")
	roots := []root{{path: dir}}

	lower, _ := grepFiles(roots, 0, activeRootScope, "hello")
	if len(lower) != 1 {
		t.Fatalf("grepFiles(\"hello\") = %v, want one match against mixed-case text", grepLines(lower))
	}

	upper, _ := grepFiles(roots, 0, activeRootScope, "HELLO")
	if len(upper) != 0 {
		t.Fatalf("grepFiles(\"HELLO\") = %v, want no match against differently-cased text", grepLines(upper))
	}
}

// TestGrepFilesOnlySearchesReadableFiles: a non-Readable file in Scope
// (S13) is never searched, even though it contains the query — grepFiles
// only ever sees what listFiles already allowed through.
func TestGrepFilesOnlySearchesReadableFiles(t *testing.T) {
	requireRipgrep(t)
	dir := t.TempDir()
	writeFileContent(t, filepath.Join(dir, "notes.md"), "needle here\n")
	writeFileContent(t, filepath.Join(dir, "image.png"), "needle here too\n")
	roots := []root{{path: dir}}

	results, _ := grepFiles(roots, 0, activeRootScope, "needle")

	want := []string{filepath.Join(dir, "notes.md") + ":1:needle here"}
	if got := grepLines(results); !equalStrings(got, want) {
		t.Fatalf("grepFiles(\"needle\") = %v, want %v", got, want)
	}
}

// TestGrepFilesCapsAndReportsTruncation: a match count larger than
// pickerResultCap comes back capped at that length with the truncation
// flag set (S12), the same shape fuzzyFiles gives the File Picker; a match
// count exactly at the cap comes back untouched and unflagged (TRIAGE T3 /
// D45, matching TestFuzzyFilesCapsAndReportsTruncation's existing coverage
// of that boundary).
func TestGrepFilesCapsAndReportsTruncation(t *testing.T) {
	requireRipgrep(t)
	notesWithLines := func(t *testing.T, n int) []root {
		t.Helper()
		dir := t.TempDir()
		var lines strings.Builder
		for i := 0; i < n; i++ {
			fmt.Fprintf(&lines, "needle line %d\n", i)
		}
		writeFileContent(t, filepath.Join(dir, "notes.md"), lines.String())
		return []root{{path: dir}}
	}

	results, truncated := grepFiles(notesWithLines(t, pickerResultCap+50), 0, activeRootScope, "needle")
	if !truncated {
		t.Fatalf("grepFiles() truncated = false, want true")
	}
	if len(results) != pickerResultCap {
		t.Fatalf("grepFiles() returned %d results, want %d", len(results), pickerResultCap)
	}

	results, truncated = grepFiles(notesWithLines(t, pickerResultCap), 0, activeRootScope, "needle")
	if truncated {
		t.Fatalf("grepFiles() truncated = true at exactly the cap, want false")
	}
	if len(results) != pickerResultCap {
		t.Fatalf("grepFiles() returned %d results, want %d", len(results), pickerResultCap)
	}
}

// TestGrepFilesBatchesAcrossArgSizeLimit: with grepArgMaxBytes shrunk to
// force one file per rg invocation, grepFiles still finds matches in files
// from every batch, merged as if it had run in a single invocation
// (TRIAGE T1 / D43, M1.4a) — a small injected limit exercises the same
// batching boundary a real ARG_MAX overflow would, cheaply. A genuine
// no-match under the same batching still reports "no results", not
// something conflated with a batch that failed to run.
func TestGrepFilesBatchesAcrossArgSizeLimit(t *testing.T) {
	requireRipgrep(t)
	dir := t.TempDir()
	const fileCount = 6
	var want []string
	for i := 0; i < fileCount; i++ {
		p := writeFileContent(t, filepath.Join(dir, fmt.Sprintf("file-%d.md", i)), fmt.Sprintf("needle %d\n", i))
		want = append(want, fmt.Sprintf("%s:1:needle %d", p, i))
	}
	sort.Strings(want)
	roots := []root{{path: dir}}

	orig := grepArgMaxBytes
	grepArgMaxBytes = 1 // forces a new batch for every file
	defer func() { grepArgMaxBytes = orig }()

	results, truncated := grepFiles(roots, 0, activeRootScope, "needle")
	if truncated {
		t.Fatalf("grepFiles() truncated = true, want false")
	}
	if got := grepLines(results); !equalStrings(got, want) {
		t.Fatalf("grepFiles() with batching = %v, want %v", got, want)
	}

	noMatch, truncated := grepFiles(roots, 0, activeRootScope, "absent")
	if truncated {
		t.Fatalf("grepFiles(\"absent\") truncated = true, want false")
	}
	if len(noMatch) != 0 {
		t.Fatalf("grepFiles(\"absent\") = %v, want no results", grepLines(noMatch))
	}
}

// TestGrepFilesBatchedResultsExactlyAtCapAreNotTruncated: M1.4a (batching)
// and M1.4b (the exact-cap boundary) each cover their own concern in
// isolation only; this pins the combination (TRIAGE T19) — a match count
// landing exactly at pickerResultCap while spread across multiple
// grepArgMaxBytes batches still reports truncated = false, not a false
// positive from batch-boundary accounting (e.g. checking the cap only at
// the end of a batch rather than as each result is appended).
func TestGrepFilesBatchedResultsExactlyAtCapAreNotTruncated(t *testing.T) {
	requireRipgrep(t)
	dir := t.TempDir()
	const fileCount = 4
	if pickerResultCap%fileCount != 0 {
		t.Fatalf("pickerResultCap (%d) not evenly divisible by fileCount (%d)", pickerResultCap, fileCount)
	}
	linesPerFile := pickerResultCap / fileCount
	for i := 0; i < fileCount; i++ {
		var lines strings.Builder
		for j := 0; j < linesPerFile; j++ {
			fmt.Fprintf(&lines, "needle %d-%d\n", i, j)
		}
		writeFileContent(t, filepath.Join(dir, fmt.Sprintf("file-%d.md", i)), lines.String())
	}
	roots := []root{{path: dir}}

	orig := grepArgMaxBytes
	grepArgMaxBytes = 1 // forces a new batch per file, spreading the cap-sized total across fileCount batches
	defer func() { grepArgMaxBytes = orig }()

	results, truncated := grepFiles(roots, 0, activeRootScope, "needle")
	if truncated {
		t.Fatalf("grepFiles() truncated = true at exactly the cap spread across batches, want false")
	}
	if len(results) != pickerResultCap {
		t.Fatalf("grepFiles() returned %d results, want %d", len(results), pickerResultCap)
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
