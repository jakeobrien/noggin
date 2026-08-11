package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"unicode"
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

	got := filePaths(listFiles(roots, 0, activeRootScope, false))

	want := sortedCopy([]string{notesMd, readme})
	if !equalStrings(got, want) {
		t.Fatalf("listFiles() = %v, want %v", got, want)
	}
}

// TestListFilesOmitsGitignored: a .gitignore'd file is absent even though
// it's Readable by extension — rg's own ignore rules decide this, ahead of
// the allowlist (S14). S4/M6.1 leaves this half of the old
// TestListFilesOmitsGitignoredAndHidden unchanged; the hidden half is now
// the opposite assertion, split out below (TestListFilesIncludesDotfiles).
func TestListFilesOmitsGitignored(t *testing.T) {
	requireRipgrep(t)
	dir := t.TempDir()
	kept := writeFile(t, filepath.Join(dir, "kept.md"))
	// M6.1 widened rgFiles to see dotfiles at all; a .gitignore'd dotfile is
	// the case that could regress silently if --hidden ever came ahead of
	// ignore-file handling instead of alongside it — pinned here next to the
	// non-dotfile case rather than assumed from it.
	writeFileContent(t, filepath.Join(dir, ".gitignore"), "ignored.md\n.ignored-dotfile.md\n")
	writeFile(t, filepath.Join(dir, "ignored.md"))
	writeFile(t, filepath.Join(dir, ".ignored-dotfile.md"))
	roots := []root{{path: dir}}

	got := filePaths(listFiles(roots, 0, activeRootScope, false))

	want := []string{kept}
	if !equalStrings(got, want) {
		t.Fatalf("listFiles() = %v, want %v (both the plain and the dotfile .gitignore entry must stay excluded)", got, want)
	}
}

// TestListFilesIncludesDotfiles is S4/M6.1's own acceptance criterion at
// the File Picker's own layer: a dotfile and a Readable file inside a
// dot-directory — both already shown by the Files Nav tree (tree.Walk
// applies no hidden-file rule of its own) — are now found here too,
// matching that pane with no extra keystroke. A non-Readable dotfile
// (wrong extension) stays excluded, same allowlist as ever — M6.1 widens
// which files are *candidates*, not the Readable policy itself.
func TestListFilesIncludesDotfiles(t *testing.T) {
	requireRipgrep(t)
	dir := t.TempDir()
	dotfile := writeFile(t, filepath.Join(dir, ".hidden.md"))
	mkdir(t, filepath.Join(dir, ".hiddendir"))
	insideDotdir := writeFile(t, filepath.Join(dir, ".hiddendir", "inside.md"))
	writeFileContent(t, filepath.Join(dir, ".bashrc"), "not Readable: no allowlisted extension\n")
	roots := []root{{path: dir}}

	got := filePaths(listFiles(roots, 0, activeRootScope, false))

	want := sortedCopy([]string{dotfile, insideDotdir})
	if !equalStrings(got, want) {
		t.Fatalf("listFiles() = %v, want %v (dotfiles present, .bashrc excluded by the allowlist)", got, want)
	}
}

// TestListFilesHidesDotfilesWhenToggled is S4/M6.2's alt+h toggle at
// listFiles' own level: hideDotfiles=true drops both a dotfile's own base
// name and a Readable file merely living inside a dot-directory (the same
// two candidates TestListFilesIncludesDotfiles pins as present by default),
// leaving only a plain Readable file untouched. The excluded-regardless
// non-Readable dotfile is unaffected either way — the toggle narrows an
// already-Readable candidate set, it does not widen the allowlist.
func TestListFilesHidesDotfilesWhenToggled(t *testing.T) {
	requireRipgrep(t)
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".hidden.md"))
	mkdir(t, filepath.Join(dir, ".hiddendir"))
	writeFile(t, filepath.Join(dir, ".hiddendir", "inside.md"))
	plain := writeFile(t, filepath.Join(dir, "plain.md"))
	roots := []root{{path: dir}}

	got := filePaths(listFiles(roots, 0, activeRootScope, true))

	want := []string{plain}
	if !equalStrings(got, want) {
		t.Fatalf("listFiles(hideDotfiles=true) = %v, want %v", got, want)
	}
}

// TestListFilesStillExcludesGitDirectory checks the flip side of M6.1's
// --hidden flag: `.git` must stay excluded unconditionally (S4), the same
// as tree.Walk's own by-name check, not merely as a side effect of the
// hidden-file rule --hidden just widened past. Uses a Readable-by-extension
// name inside `.git` so a regression (dropping the `.git` glob but keeping
// --hidden) would otherwise list it.
func TestListFilesStillExcludesGitDirectory(t *testing.T) {
	requireRipgrep(t)
	dir := t.TempDir()
	kept := writeFile(t, filepath.Join(dir, "kept.md"))
	gitDir := mkdir(t, filepath.Join(dir, ".git"))
	writeFile(t, filepath.Join(gitDir, "COMMIT_EDITMSG.md"))
	// A .git nested under a subfolder (e.g. a vendored/nested checkout) must
	// stay excluded too — the glob's doc comment claims "any depth", matching
	// tree.Walk's by-name check regardless of nesting; pinned here rather
	// than left to only the depth-1 case above.
	sub := mkdir(t, filepath.Join(dir, "sub"))
	nestedGitDir := mkdir(t, filepath.Join(sub, ".git"))
	writeFile(t, filepath.Join(nestedGitDir, "COMMIT_EDITMSG.md"))
	roots := []root{{path: dir}}

	got := filePaths(listFiles(roots, 0, activeRootScope, false))

	want := []string{kept}
	if !equalStrings(got, want) {
		t.Fatalf("listFiles() = %v, want %v (.git must stay excluded at any depth)", got, want)
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

	got := filePaths(listFiles(roots, 0, allRootsScope, false))

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

	got := filePaths(listFiles(roots, 1, activeRootScope, false))

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

	files := listFiles(roots, 0, allRootsScope, false)

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

	results, truncated := grepFiles(roots, 0, activeRootScope, "v1.2", false)

	if truncated {
		t.Fatalf("grepFiles() truncated = true, want false")
	}
	want := []string{filepath.Join(dir, "notes.md") + ":1:see v1.2 here"}
	if got := grepLines(results); !equalStrings(got, want) {
		t.Fatalf("grepFiles(\"v1.2\") = %v, want %v", got, want)
	}
}

// TestGrepFilesFindsMatchesInsideDotfiles is S4/M6.1's Grep Picker half:
// once listFiles includes a dotfile (TestListFilesIncludesDotfiles above),
// grepFiles searches it too with no extra keystroke — it runs over exactly
// that file list (grepFiles' own doc comment), so this is really pinning
// that grepBatch's explicit-file-argument invocation of rg does not itself
// re-apply a hidden-file filter on top (verified against the sandbox's real
// rg: an explicitly-named file is always searched regardless of --hidden).
func TestGrepFilesFindsMatchesInsideDotfiles(t *testing.T) {
	requireRipgrep(t)
	dir := t.TempDir()
	writeFileContent(t, filepath.Join(dir, ".hidden.md"), "needle in a dotfile\n")
	roots := []root{{path: dir}}

	results, truncated := grepFiles(roots, 0, activeRootScope, "needle", false)

	if truncated {
		t.Fatalf("grepFiles() truncated = true, want false")
	}
	want := []string{filepath.Join(dir, ".hidden.md") + ":1:needle in a dotfile"}
	if got := grepLines(results); !equalStrings(got, want) {
		t.Fatalf("grepFiles(\"needle\") = %v, want %v", got, want)
	}
}

// TestGrepFilesHidesDotfilesWhenToggled is S4/M6.2's Grep Picker half:
// hideDotfiles=true drops a dotfile from grepFiles' own candidate set (it
// reuses listFiles, TestListFilesHidesDotfilesWhenToggled's own guarantee)
// just as it does for the File Picker, so a match that only exists inside a
// dotfile is not found at all — not merely hidden after being found — while
// a match in a plain file is unaffected.
func TestGrepFilesHidesDotfilesWhenToggled(t *testing.T) {
	requireRipgrep(t)
	dir := t.TempDir()
	writeFileContent(t, filepath.Join(dir, ".hidden.md"), "needle in a dotfile\n")
	writeFileContent(t, filepath.Join(dir, "plain.md"), "needle in a plain file\n")
	roots := []root{{path: dir}}

	results, truncated := grepFiles(roots, 0, activeRootScope, "needle", true)

	if truncated {
		t.Fatalf("grepFiles(hideDotfiles=true) truncated = true, want false")
	}
	want := []string{filepath.Join(dir, "plain.md") + ":1:needle in a plain file"}
	if got := grepLines(results); !equalStrings(got, want) {
		t.Fatalf("grepFiles(hideDotfiles=true) = %v, want %v", got, want)
	}
}

// TestIsDotfilePath pins the path-component rule directly: a match on any
// component between the Root and the file (not just the file's own base
// name), no match for an all-plain path, and no false match from a `.` that
// merely appears mid-component (e.g. "a.b/c.md") rather than starting one.
func TestIsDotfilePath(t *testing.T) {
	cases := []struct {
		rel  string
		want bool
	}{
		{"plain.md", false},
		{"docs/plain.md", false},
		{".hidden.md", true},
		{".hiddendir/inside.md", true},
		{"docs/.hiddendir/inside.md", true},
		{"a.b/c.md", false},
	}
	for _, c := range cases {
		if got := isDotfilePath(c.rel); got != c.want {
			t.Errorf("isDotfilePath(%q) = %v, want %v", c.rel, got, c.want)
		}
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

	lower, _ := grepFiles(roots, 0, activeRootScope, "hello", false)
	if len(lower) != 1 {
		t.Fatalf("grepFiles(\"hello\") = %v, want one match against mixed-case text", grepLines(lower))
	}

	upper, _ := grepFiles(roots, 0, activeRootScope, "HELLO", false)
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

	results, _ := grepFiles(roots, 0, activeRootScope, "needle", false)

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

	results, truncated := grepFiles(notesWithLines(t, pickerResultCap+50), 0, activeRootScope, "needle", false)
	if !truncated {
		t.Fatalf("grepFiles() truncated = false, want true")
	}
	if len(results) != pickerResultCap {
		t.Fatalf("grepFiles() returned %d results, want %d", len(results), pickerResultCap)
	}

	results, truncated = grepFiles(notesWithLines(t, pickerResultCap), 0, activeRootScope, "needle", false)
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

	results, truncated := grepFiles(roots, 0, activeRootScope, "needle", false)
	if truncated {
		t.Fatalf("grepFiles() truncated = true, want false")
	}
	if got := grepLines(results); !equalStrings(got, want) {
		t.Fatalf("grepFiles() with batching = %v, want %v", got, want)
	}

	noMatch, truncated := grepFiles(roots, 0, activeRootScope, "absent", false)
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
//
// D56 (TRIAGE T23): today grepFiles checks the cap per-appended-result
// inside the innermost loop regardless of batch boundaries, so this test
// can't yet exercise a path TestGrepFilesCapsAndReportsTruncation doesn't
// already cover — kept anyway as a forward-looking regression guard: it
// starts covering real distinct behavior the moment a future batching
// refactor moves the cap check outside the per-result loop.
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

	results, truncated := grepFiles(roots, 0, activeRootScope, "needle", false)
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

// TestFileNameMatchPositionsSpellTheQuery pins fileNameMatchPositions'
// contract the view layer depends on (M6.1): the positions are rune indexes
// into the *name*, in ascending order, and the characters they point at are
// exactly the query's own characters (case-insensitively, since the fuzzy
// match is). Asserting the characters rather than a hard-coded index list is
// what catches the byte-vs-rune confusion the conversion exists for —
// fuzzy's MatchedIndexes are byte offsets, so a name with a multibyte
// character (café) would silently shift every later highlight by one column
// if they were used as-is.
func TestFileNameMatchPositionsSpellTheQuery(t *testing.T) {
	cases := []struct{ name, query string }{
		{"report.md", "rep"},     // contiguous prefix
		{"report.md", "rpt"},     // scattered
		{"Report.md", "rep"},     // case-insensitive
		{"café-notes.md", "cfn"}, // multibyte before the later matches
		{"café-notes.md", "cén"}, // the multibyte rune itself matched
	}
	for _, c := range cases {
		positions := fileNameMatchPositions(c.name, c.query)
		runes := []rune(c.name)
		query := []rune(strings.ToLower(c.query))
		if len(positions) != len(query) {
			t.Errorf("fileNameMatchPositions(%q, %q) = %v, want %d positions", c.name, c.query, positions, len(query))
			continue
		}
		last := -1
		for i, p := range positions {
			if p <= last || p >= len(runes) {
				t.Errorf("fileNameMatchPositions(%q, %q) = %v: position %d out of order or out of range", c.name, c.query, positions, p)
				break
			}
			last = p
			if got := []rune(strings.ToLower(string(runes[p])))[0]; got != query[i] {
				t.Errorf("fileNameMatchPositions(%q, %q) = %v: position %d points at %q, want %q", c.name, c.query, positions, p, got, query[i])
			}
		}
	}
	if positions := fileNameMatchPositions("report.md", ""); positions != nil {
		t.Errorf("fileNameMatchPositions() with an empty query = %v, want nil", positions)
	}
	if positions := fileNameMatchPositions("report.md", "zzz"); positions != nil {
		t.Errorf("fileNameMatchPositions() with a non-matching query = %v, want nil", positions)
	}
}

// TestFileNameMatchPositionsCoverEveryRankedResult is the anti-drift guard
// for recomputing match positions at render time instead of carrying them
// out of fuzzyFiles (see fileNameMatchPositions' own comment): every row
// fuzzyFiles ranked must be highlightable, i.e. the recomputed match must
// exist for each one. If the two ever stopped agreeing — a different matcher,
// a different string fed to it — a ranked row would render with no highlight,
// and this fails.
func TestFileNameMatchPositionsCoverEveryRankedResult(t *testing.T) {
	files := []pickerFile{
		pickerFileAt("/notes/report.md"),
		pickerFileAt("/notes/Retrospective.md"),
		pickerFileAt("/notes/rough-plan-tuesday.md"),
		pickerFileAt("/notes/unrelated.md"),
	}
	results, _ := fuzzyFiles(files, "rpt")
	if len(results) == 0 {
		t.Fatal("fuzzyFiles() ranked nothing; fixture is broken")
	}
	for _, f := range results {
		name := filepath.Base(f.path)
		if positions := fileNameMatchPositions(name, "rpt"); len(positions) == 0 {
			t.Errorf("fuzzyFiles ranked %q but fileNameMatchPositions(%q, %q) found no match", f.path, name, "rpt")
		}
	}
}

// TestGrepMatchPositionsAreRuneIndexesUnderFolding guards grepMatchPositions'
// reuse of find.go's fold-and-map translation (foldLower/origSpan): the Kelvin
// sign U+212A is three bytes and folds to a one-byte 'k', so a smart-case
// search over a line containing one produces byte offsets in the *folded*
// string that are two short of the line's own — and rune positions two
// columns off — unless they are mapped back. The reported positions must
// point at the matched characters themselves, whatever precedes them.
func TestGrepMatchPositionsAreRuneIndexesUnderFolding(t *testing.T) {
	cases := []struct {
		text, query string
		want        []int
	}{
		// runes: 0 'K'(U+212A), 1 ' ', 2 'a', 3 'b', 4 'c', 5 ' ', 6-11 "report"
		{"\u212A abc report", "report", []int{6, 7, 8, 9, 10, 11}},
		// café: the é is one rune, two bytes, before the match
		{"café report", "report", []int{5, 6, 7, 8, 9, 10}},
		{"no match here", "report", nil},
		{"report", "", nil},
	}
	for _, c := range cases {
		got := grepMatchPositions(c.text, c.query)
		if !equalInts(got, c.want) {
			t.Errorf("grepMatchPositions(%q, %q) = %v, want %v", c.text, c.query, got, c.want)
			continue
		}
		runes := []rune(c.text)
		for i, p := range got {
			if p >= len(runes) {
				t.Errorf("grepMatchPositions(%q, %q) = %v: position %d out of range", c.text, c.query, got, p)
				continue
			}
			if want := []rune(strings.ToLower(c.query))[i]; unicode.ToLower(runes[p]) != want {
				t.Errorf("grepMatchPositions(%q, %q) = %v: position %d points at %q, want %q", c.text, c.query, got, p, runes[p], want)
			}
		}
	}
}

// starredPickerFileByPath finds one starredPickerFiles result by path, so a
// test can assert on a single candidate's own fields (root/orphan) rather
// than just membership.
func starredPickerFileByPath(files []pickerFile, path string) (pickerFile, bool) {
	for _, f := range files {
		if f.path == path {
			return f, true
		}
	}
	return pickerFile{}, false
}

// TestStarredPickerFilesIncludesOrphansAndRootCoveredFiles is S8/M5.1's own
// acceptance wording, first clause: Starred scope's candidate set (S8: "the
// whole store — every starred document, including orphans") reaches a
// starred document under no session Root exactly as it reaches one under a
// Root — mirroring starredRows' own
// TestStarredRowsAllStarredIncludesOrphansAndOtherRoots (browse_starred_test.go),
// the Files Nav's identical S1-store-wide guarantee.
func TestStarredPickerFilesIncludesOrphansAndRootCoveredFiles(t *testing.T) {
	dir := t.TempDir()
	rootDir := mkdir(t, filepath.Join(dir, "root-a"))
	orphanDir := mkdir(t, filepath.Join(dir, "elsewhere"))
	inRoot := writeFileContent(t, filepath.Join(rootDir, "a.md"), "content\n")
	orphan := writeFileContent(t, filepath.Join(orphanDir, "orphan.md"), "content\n")
	// An unstarred document under the same Root must not leak in.
	writeFileContent(t, filepath.Join(rootDir, "unstarred.md"), "content\n")

	docs := map[string]docRecord{
		inRoot: {Starred: true},
		orphan: {Starred: true},
	}
	roots := []root{{path: rootDir}}

	got := filePaths(starredPickerFiles(docs, roots, false))
	want := sortedCopy([]string{inRoot, orphan})
	if !equalStrings(got, want) {
		t.Fatalf("starredPickerFiles() = %v, want %v", got, want)
	}

	rootFile, ok := starredPickerFileByPath(starredPickerFiles(docs, roots, false), inRoot)
	if !ok || rootFile.orphan || rootFile.root.path != rootDir {
		t.Errorf("Root-covered candidate = %+v, want orphan=false root.path=%q", rootFile, rootDir)
	}
	orphanFile, ok := starredPickerFileByPath(starredPickerFiles(docs, roots, false), orphan)
	if !ok || !orphanFile.orphan || orphanFile.root.path != "" {
		t.Errorf("orphan candidate = %+v, want orphan=true, zero-value root", orphanFile)
	}
}

// TestStarredPickerFilesExcludesMissingFiles: unlike Browse's own
// starredRows (S3: a broken star is still listed, marked), the Picker's
// Starred candidate set drops a starred record whose file no longer exists
// — there is nothing for a File-mode match or a Grep search to do with a
// path that does not resolve to real bytes on disk (starredPickerFiles' own
// doc comment).
func TestStarredPickerFilesExcludesMissingFiles(t *testing.T) {
	dir := t.TempDir()
	present := writeFileContent(t, filepath.Join(dir, "present.md"), "content\n")
	missing := filepath.Join(dir, "gone.md")

	docs := map[string]docRecord{
		present: {Starred: true},
		missing: {Starred: true},
	}

	got := filePaths(starredPickerFiles(docs, nil, false))
	want := []string{present}
	if !equalStrings(got, want) {
		t.Fatalf("starredPickerFiles() = %v, want %v (missing file dropped)", got, want)
	}
}

// TestStarredPickerFilesAppliesReadableAllowlist mirrors
// TestListFilesAppliesReadableAllowlist: Starred scope's candidate set is
// still bound by S13's Readable allowlist, the same as every other scope's.
func TestStarredPickerFilesAppliesReadableAllowlist(t *testing.T) {
	dir := t.TempDir()
	notesMd := writeFileContent(t, filepath.Join(dir, "notes.md"), "content\n")
	binPath := writeFileContent(t, filepath.Join(dir, "image.png"), "content\n")

	docs := map[string]docRecord{
		notesMd: {Starred: true},
		binPath: {Starred: true},
	}

	got := filePaths(starredPickerFiles(docs, nil, false))
	want := []string{notesMd}
	if !equalStrings(got, want) {
		t.Fatalf("starredPickerFiles() = %v, want %v (non-Readable file dropped)", got, want)
	}
}

// TestStarredPickerFilesAppliesHideDotfiles mirrors
// TestListFilesHonoursHideDotfilesToggle: a starred document under a
// dot-directory or with a dotfile name is dropped when hideDotfiles is set,
// the same session-scoped toggle every other candidate set already honours.
func TestStarredPickerFilesAppliesHideDotfiles(t *testing.T) {
	dir := t.TempDir()
	plain := writeFileContent(t, filepath.Join(dir, "notes.md"), "content\n")
	dotdir := mkdir(t, filepath.Join(dir, ".config"))
	dotted := writeFileContent(t, filepath.Join(dotdir, "hidden.md"), "content\n")

	docs := map[string]docRecord{
		plain:  {Starred: true},
		dotted: {Starred: true},
	}

	shown := filePaths(starredPickerFiles(docs, nil, false))
	if want := sortedCopy([]string{plain, dotted}); !equalStrings(shown, want) {
		t.Fatalf("starredPickerFiles(hideDotfiles=false) = %v, want %v", shown, want)
	}

	hidden := filePaths(starredPickerFiles(docs, nil, true))
	if want := []string{plain}; !equalStrings(hidden, want) {
		t.Fatalf("starredPickerFiles(hideDotfiles=true) = %v, want %v (dot-directory file dropped)", hidden, want)
	}
}

// TestStarredPickerFilesExcludesUnstarredDocuments confirms the candidate
// set reads Starred alone, the same S1-field independence starredRows'
// own doc comment already states for Browse.
func TestStarredPickerFilesExcludesUnstarredDocuments(t *testing.T) {
	dir := t.TempDir()
	starred := writeFileContent(t, filepath.Join(dir, "starred.md"), "content\n")
	unstarred := writeFileContent(t, filepath.Join(dir, "unstarred.md"), "content\n")

	docs := map[string]docRecord{
		starred:   {Starred: true},
		unstarred: {Starred: false},
	}

	got := filePaths(starredPickerFiles(docs, nil, false))
	want := []string{starred}
	if !equalStrings(got, want) {
		t.Fatalf("starredPickerFiles() = %v, want %v (unstarred document excluded)", got, want)
	}
}
