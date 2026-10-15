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

	"noggin/internal/tree"
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

// TestListFilesAppliesMarkdownClassification: only Markdown files come back;
// everything else rg finds — Text files (.txt, LICENSE) included, until the
// File Filter lands — is dropped (S13).
func TestListFilesAppliesMarkdownClassification(t *testing.T) {
	requireRipgrep(t)
	dir := t.TempDir()
	notesMd := writeFile(t, filepath.Join(dir, "notes.md"))
	readme := writeFile(t, filepath.Join(dir, "README"))
	writeFile(t, filepath.Join(dir, "main.go"))
	writeFile(t, filepath.Join(dir, "image.png"))
	writeFile(t, filepath.Join(dir, "notes.txt"))
	writeFile(t, filepath.Join(dir, "LICENSE"))
	roots := []root{{path: dir}}

	got := filePaths(listFiles(roots, 0, activeRootScope, pickerBounds{hideDotfiles: false, filter: tree.FilterMarkdown}))

	want := sortedCopy([]string{notesMd, readme})
	if !equalStrings(got, want) {
		t.Fatalf("listFiles() = %v, want %v", got, want)
	}
}

// TestListFilesOmitsGitignored: a .gitignore'd file is absent even though
// it's a Markdown file by extension — rg's own ignore rules decide this, ahead of
// the File Filter (S14). S4/M6.1 leaves this half of the old
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

	got := filePaths(listFiles(roots, 0, activeRootScope, pickerBounds{hideDotfiles: false, filter: tree.FilterMarkdown}))

	want := []string{kept}
	if !equalStrings(got, want) {
		t.Fatalf("listFiles() = %v, want %v (both the plain and the dotfile .gitignore entry must stay excluded)", got, want)
	}
}

// TestListFilesIncludesDotfiles is S4/M6.1's own acceptance criterion at
// the File Picker's own layer: a dotfile and a Markdown file inside a
// dot-directory — both already shown by the Files Nav tree (tree.Walk
// applies no hidden-file rule of its own) — are now found here too,
// matching that pane with no extra keystroke. A non-Markdown dotfile
// (wrong extension) stays excluded, same allowlist as ever — M6.1 widens
// which files are *candidates*, not the File Filter itself.
func TestListFilesIncludesDotfiles(t *testing.T) {
	requireRipgrep(t)
	dir := t.TempDir()
	dotfile := writeFile(t, filepath.Join(dir, ".hidden.md"))
	mkdir(t, filepath.Join(dir, ".hiddendir"))
	insideDotdir := writeFile(t, filepath.Join(dir, ".hiddendir", "inside.md"))
	writeFileContent(t, filepath.Join(dir, ".bashrc"), "not a Markdown file: no Markdown extension\n")
	roots := []root{{path: dir}}

	got := filePaths(listFiles(roots, 0, activeRootScope, pickerBounds{hideDotfiles: false, filter: tree.FilterMarkdown}))

	want := sortedCopy([]string{dotfile, insideDotdir})
	if !equalStrings(got, want) {
		t.Fatalf("listFiles() = %v, want %v (dotfiles present, .bashrc excluded by the allowlist)", got, want)
	}
}

// TestListFilesHidesDotfilesWhenToggled is S4/M6.2's alt+h toggle at
// listFiles' own level: hideDotfiles=true drops both a dotfile's own base
// name and a Markdown file merely living inside a dot-directory (the same
// two candidates TestListFilesIncludesDotfiles pins as present by default),
// leaving only a plain Markdown file untouched. The excluded-regardless
// non-Markdown dotfile is unaffected either way — the toggle narrows an
// already-Markdown candidate set, it does not widen the allowlist.
func TestListFilesHidesDotfilesWhenToggled(t *testing.T) {
	requireRipgrep(t)
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".hidden.md"))
	mkdir(t, filepath.Join(dir, ".hiddendir"))
	writeFile(t, filepath.Join(dir, ".hiddendir", "inside.md"))
	plain := writeFile(t, filepath.Join(dir, "plain.md"))
	roots := []root{{path: dir}}

	got := filePaths(listFiles(roots, 0, activeRootScope, pickerBounds{hideDotfiles: true, filter: tree.FilterMarkdown}))

	want := []string{plain}
	if !equalStrings(got, want) {
		t.Fatalf("listFiles(hideDotfiles=true) = %v, want %v", got, want)
	}
}

// TestListFilesStillExcludesGitDirectory checks the flip side of M6.1's
// --hidden flag: `.git` must stay excluded unconditionally (S4), the same
// as tree.Walk's own by-name check, not merely as a side effect of the
// hidden-file rule --hidden just widened past. Uses a Markdown-by-extension
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

	got := filePaths(listFiles(roots, 0, activeRootScope, pickerBounds{hideDotfiles: false, filter: tree.FilterMarkdown}))

	want := []string{kept}
	if !equalStrings(got, want) {
		t.Fatalf("listFiles() = %v, want %v (.git must stay excluded at any depth)", got, want)
	}
}

// TestListFilesAllRootsCoversLiveRootsAndSkipsBroken: All Roots scope finds
// Markdown files across every live Root and silently omits a Broken one,
// without an error (S6).
func TestListFilesAllRootsCoversLiveRootsAndSkipsBroken(t *testing.T) {
	requireRipgrep(t)
	a := mkdir(t, filepath.Join(t.TempDir(), "a"))
	aFile := writeFile(t, filepath.Join(a, "one.md"))
	b := mkdir(t, filepath.Join(t.TempDir(), "b"))
	bFile := writeFile(t, filepath.Join(b, "two.md"))
	missing := filepath.Join(t.TempDir(), "gone")
	roots := []root{{path: a}, {path: missing}, {path: b}}

	got := filePaths(listFiles(roots, 0, allRootsScope, pickerBounds{hideDotfiles: false, filter: tree.FilterMarkdown}))

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

	got := filePaths(listFiles(roots, 1, activeRootScope, pickerBounds{hideDotfiles: false, filter: tree.FilterMarkdown}))

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

	files := listFiles(roots, 0, allRootsScope, pickerBounds{hideDotfiles: false, filter: tree.FilterMarkdown})

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
// (the File Picker matches the file name only, never path
// components).
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
// (S12/C2).
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

// TestGrepPickerFilesMatchesLiterally: a query containing regex metacharacters
// matches only their literal occurrence, not what they'd mean as regex
// (S4: --fixed-strings).
func TestGrepPickerFilesMatchesLiterally(t *testing.T) {
	requireRipgrep(t)
	dir := t.TempDir()
	writeFileContent(t, filepath.Join(dir, "notes.md"), "see v1.2 here\nnot v1x2 though\n")
	roots := []root{{path: dir}}

	results, truncated := grepPickerFiles(listFiles(roots, 0, activeRootScope, pickerBounds{hideDotfiles: false, filter: tree.FilterMarkdown}), "v1.2")

	if truncated {
		t.Fatalf("grepPickerFiles() truncated = true, want false")
	}
	want := []string{filepath.Join(dir, "notes.md") + ":1:see v1.2 here"}
	if got := grepLines(results); !equalStrings(got, want) {
		t.Fatalf("grepPickerFiles(\"v1.2\") = %v, want %v", got, want)
	}
}

// TestGrepPickerFilesFindsMatchesInsideDotfiles is S4/M6.1's Grep Picker
// half: once listFiles includes a dotfile (TestListFilesIncludesDotfiles
// above), grepPickerFiles searches it too with no extra keystroke — it runs
// over exactly that file list (grepPickerFiles' own doc comment), so this
// is really pinning that grepBatch's explicit-file-argument invocation of
// rg does not itself re-apply a hidden-file filter on top (verified
// against the sandbox's real rg: an explicitly-named file is always
// searched regardless of --hidden).
func TestGrepPickerFilesFindsMatchesInsideDotfiles(t *testing.T) {
	requireRipgrep(t)
	dir := t.TempDir()
	writeFileContent(t, filepath.Join(dir, ".hidden.md"), "needle in a dotfile\n")
	roots := []root{{path: dir}}

	results, truncated := grepPickerFiles(listFiles(roots, 0, activeRootScope, pickerBounds{hideDotfiles: false, filter: tree.FilterMarkdown}), "needle")

	if truncated {
		t.Fatalf("grepPickerFiles() truncated = true, want false")
	}
	want := []string{filepath.Join(dir, ".hidden.md") + ":1:needle in a dotfile"}
	if got := grepLines(results); !equalStrings(got, want) {
		t.Fatalf("grepPickerFiles(\"needle\") = %v, want %v", got, want)
	}
}

// TestGrepPickerFilesHidesDotfilesWhenToggled is S4/M6.2's Grep Picker half:
// hideDotfiles=true drops a dotfile from listFiles' own candidate set
// before it ever reaches grepPickerFiles (TestListFilesHidesDotfilesWhen-
// Toggled's own guarantee) just as it does for the File Picker, so a match
// that only exists inside a dotfile is not found at all — not merely
// hidden after being found — while a match in a plain file is unaffected.
func TestGrepPickerFilesHidesDotfilesWhenToggled(t *testing.T) {
	requireRipgrep(t)
	dir := t.TempDir()
	writeFileContent(t, filepath.Join(dir, ".hidden.md"), "needle in a dotfile\n")
	writeFileContent(t, filepath.Join(dir, "plain.md"), "needle in a plain file\n")
	roots := []root{{path: dir}}

	results, truncated := grepPickerFiles(listFiles(roots, 0, activeRootScope, pickerBounds{hideDotfiles: true, filter: tree.FilterMarkdown}), "needle")

	if truncated {
		t.Fatalf("grepPickerFiles(hideDotfiles=true) truncated = true, want false")
	}
	want := []string{filepath.Join(dir, "plain.md") + ":1:needle in a plain file"}
	if got := grepLines(results); !equalStrings(got, want) {
		t.Fatalf("grepPickerFiles(hideDotfiles=true) = %v, want %v", got, want)
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

// TestGrepPickerFilesIsSmartCase: an all-lowercase query matches mixed-case
// text, but a query containing uppercase matches only that exact case (S4,
// matching find.go's S8 smart-case semantic via rg's own --smart-case).
func TestGrepPickerFilesIsSmartCase(t *testing.T) {
	requireRipgrep(t)
	dir := t.TempDir()
	writeFileContent(t, filepath.Join(dir, "notes.md"), "Hello World\n")
	roots := []root{{path: dir}}
	files := listFiles(roots, 0, activeRootScope, pickerBounds{hideDotfiles: false, filter: tree.FilterMarkdown})

	lower, _ := grepPickerFiles(files, "hello")
	if len(lower) != 1 {
		t.Fatalf("grepPickerFiles(\"hello\") = %v, want one match against mixed-case text", grepLines(lower))
	}

	upper, _ := grepPickerFiles(files, "HELLO")
	if len(upper) != 0 {
		t.Fatalf("grepPickerFiles(\"HELLO\") = %v, want no match against differently-cased text", grepLines(upper))
	}
}

// TestGrepPickerFilesOnlySearchesMarkdownFiles: a non-Markdown file in Scope
// (S13) is never searched, even though it contains the query —
// grepPickerFiles only ever sees what listFiles already allowed through.
func TestGrepPickerFilesOnlySearchesMarkdownFiles(t *testing.T) {
	requireRipgrep(t)
	dir := t.TempDir()
	writeFileContent(t, filepath.Join(dir, "notes.md"), "needle here\n")
	writeFileContent(t, filepath.Join(dir, "image.png"), "needle here too\n")
	roots := []root{{path: dir}}

	results, _ := grepPickerFiles(listFiles(roots, 0, activeRootScope, pickerBounds{hideDotfiles: false, filter: tree.FilterMarkdown}), "needle")

	want := []string{filepath.Join(dir, "notes.md") + ":1:needle here"}
	if got := grepLines(results); !equalStrings(got, want) {
		t.Fatalf("grepPickerFiles(\"needle\") = %v, want %v", got, want)
	}
}

// TestGrepPickerFilesCapsAndReportsTruncation: a match count larger than
// pickerResultCap comes back capped at that length with the truncation
// flag set (S12/C2), the same shape fuzzyFiles gives the File Picker; a match
// count exactly at the cap comes back untouched and unflagged (TRIAGE T3,
// matching TestFuzzyFilesCapsAndReportsTruncation's existing coverage of
// that boundary).
func TestGrepPickerFilesCapsAndReportsTruncation(t *testing.T) {
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

	results, truncated := grepPickerFiles(listFiles(notesWithLines(t, pickerResultCap+50), 0, activeRootScope, pickerBounds{filter: tree.FilterMarkdown}), "needle")
	if !truncated {
		t.Fatalf("grepPickerFiles() truncated = false, want true")
	}
	if len(results) != pickerResultCap {
		t.Fatalf("grepPickerFiles() returned %d results, want %d", len(results), pickerResultCap)
	}

	results, truncated = grepPickerFiles(listFiles(notesWithLines(t, pickerResultCap), 0, activeRootScope, pickerBounds{filter: tree.FilterMarkdown}), "needle")
	if truncated {
		t.Fatalf("grepPickerFiles() truncated = true at exactly the cap, want false")
	}
	if len(results) != pickerResultCap {
		t.Fatalf("grepPickerFiles() returned %d results, want %d", len(results), pickerResultCap)
	}
}

// TestGrepPickerFilesBatchesAcrossArgSizeLimit: with grepArgMaxBytes shrunk
// to force one file per rg invocation, grepPickerFiles still finds matches
// in files from every batch, merged as if it had run in a single invocation
// (TRIAGE T1, M1.4a) — a small injected limit exercises the same
// batching boundary a real ARG_MAX overflow would, cheaply. A genuine
// no-match under the same batching still reports "no results", not
// something conflated with a batch that failed to run.
func TestGrepPickerFilesBatchesAcrossArgSizeLimit(t *testing.T) {
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
	files := listFiles(roots, 0, activeRootScope, pickerBounds{hideDotfiles: false, filter: tree.FilterMarkdown})

	orig := grepArgMaxBytes
	grepArgMaxBytes = 1 // forces a new batch for every file
	defer func() { grepArgMaxBytes = orig }()

	results, truncated := grepPickerFiles(files, "needle")
	if truncated {
		t.Fatalf("grepPickerFiles() truncated = true, want false")
	}
	if got := grepLines(results); !equalStrings(got, want) {
		t.Fatalf("grepPickerFiles() with batching = %v, want %v", got, want)
	}

	noMatch, truncated := grepPickerFiles(files, "absent")
	if truncated {
		t.Fatalf("grepPickerFiles(\"absent\") truncated = true, want false")
	}
	if len(noMatch) != 0 {
		t.Fatalf("grepPickerFiles(\"absent\") = %v, want no results", grepLines(noMatch))
	}
}

// TestGrepPickerFilesBatchedResultsExactlyAtCapAreNotTruncated: M1.4a
// (batching) and M1.4b (the exact-cap boundary) each cover their own
// concern in isolation only; this pins the combination (TRIAGE T19) — a
// match count landing exactly at pickerResultCap while spread across
// multiple grepArgMaxBytes batches still reports truncated = false, not a
// false positive from batch-boundary accounting (e.g. checking the cap
// only at the end of a batch rather than as each result is appended).
//
// TRIAGE T23: today grepPickerFiles checks the cap per-appended-
// result inside the innermost loop regardless of batch boundaries, so this
// test can't yet exercise a path TestGrepPickerFilesCapsAndReportsTruncation
// doesn't already cover — kept anyway as a forward-looking regression
// guard: it starts covering real distinct behavior the moment a future
// batching refactor moves the cap check outside the per-result loop.
func TestGrepPickerFilesBatchedResultsExactlyAtCapAreNotTruncated(t *testing.T) {
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

	results, truncated := grepPickerFiles(listFiles(roots, 0, activeRootScope, pickerBounds{hideDotfiles: false, filter: tree.FilterMarkdown}), "needle")
	if truncated {
		t.Fatalf("grepPickerFiles() truncated = true at exactly the cap spread across batches, want false")
	}
	if len(results) != pickerResultCap {
		t.Fatalf("grepPickerFiles() returned %d results, want %d", len(results), pickerResultCap)
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

// TestStarredPickerFilesIgnoresTheFileFilter: Starred scope ignores the File
// Filter, so a starred Text file stays reachable; only a binary, which
// could not be opened, is dropped.
func TestStarredPickerFilesIgnoresTheFileFilter(t *testing.T) {
	dir := t.TempDir()
	notesMd := writeFileContent(t, filepath.Join(dir, "notes.md"), "content\n")
	textPath := writeFileContent(t, filepath.Join(dir, "main.go"), "package main\n")
	binPath := writeFileContent(t, filepath.Join(dir, "image.png"), "PNG\x00\x01\n")

	docs := map[string]docRecord{
		notesMd:  {Starred: true},
		textPath: {Starred: true},
		binPath:  {Starred: true},
	}

	got := filePaths(starredPickerFiles(docs, nil, false))
	want := []string{textPath, notesMd}
	if !equalStrings(got, want) {
		t.Fatalf("starredPickerFiles() = %v, want %v (Text file kept, binary dropped)", got, want)
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

// TestStarredPickerFilesIncludesStarredArticle is C6/M3.5's own unit-level
// check for starredPickerFiles' article branch (rule 2, S6: "File mode's
// Starred scope includes starred articles"): a starred article record
// (docRecord.Article) surfaces as a pickerFile{article: true} pseudo-
// candidate carrying its title and its docs-map key as path, bypassing
// every file-specific filter above it (pathMissing, tree.IsMarkdown,
// hideDotfiles) since none of them describe a URL, alongside an ordinary
// starred file from the same store — mirroring
// TestStarredPickerFilesIncludesOrphansAndRootCoveredFiles's own
// "assert on a single candidate's own fields" shape for the file/orphan
// pair.
func TestStarredPickerFilesIncludesStarredArticle(t *testing.T) {
	dir := t.TempDir()
	dotdir := mkdir(t, filepath.Join(dir, ".hidden"))
	starredFile := writeFileContent(t, filepath.Join(dotdir, "report.md"), "content\n")
	articleKey := "https://en.wikipedia.org/wiki/Leonhard_Euler"

	docs := map[string]docRecord{
		starredFile: {Starred: true},
		articleKey:  {Starred: true, Article: true, Title: "Leonhard Euler"},
	}

	// hideDotfiles=true would drop the dotfile-directory file above (a
	// deliberate collision with the article check below: the article must
	// survive this toggle exactly as it survives every other file-specific
	// filter, since none of them are about a URL at all).
	got := starredPickerFiles(docs, nil, true)
	if len(got) != 1 {
		t.Fatalf("starredPickerFiles(hideDotfiles=true) = %+v, want exactly the article (the dotfile-path file excluded)", got)
	}
	article := got[0]
	if !article.article || article.title != "Leonhard Euler" || article.path != articleKey {
		t.Errorf("article candidate = %+v, want article=true title=%q path=%q", article, "Leonhard Euler", articleKey)
	}
	if article.orphan || article.root.path != "" {
		t.Errorf("article candidate = %+v, want orphan=false and a zero-value root (neither concept applies to an article)", article)
	}
}

// TestFileNameSourceUsesArticleTitleForFuzzyRanking is C6/M3.5's own
// unit-level check for fileNameSource's article branch (S6: "matched on
// title"): an article pseudo-candidate's fuzzy-ranking text is its title,
// not filepath.Base of its URL (which would read as an underscore-joined
// slug, not the natural-language title a query is actually typed against).
func TestFileNameSourceUsesArticleTitleForFuzzyRanking(t *testing.T) {
	source := fileNameSource([]pickerFile{
		{path: "/tmp/report.md"},
		{path: "https://en.wikipedia.org/wiki/Leonhard_Euler", article: true, title: "Leonhard Euler"},
	})
	if got := source.String(0); got != "report.md" {
		t.Errorf("fileNameSource.String(file) = %q, want %q", got, "report.md")
	}
	if got := source.String(1); got != "Leonhard Euler" {
		t.Errorf("fileNameSource.String(article) = %q, want the title %q, not a path-derived name", got, "Leonhard Euler")
	}
}
