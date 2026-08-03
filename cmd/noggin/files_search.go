package main

import (
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sahilm/fuzzy"

	"noggin/internal/tree"
)

// pickerResultCap bounds how many results a Picker search hands back,
// shared by the File Picker's fuzzy match (M1.3) and the Grep Picker's
// search (M1.4) so both report truncation the same way (S12). No SPEC/D
// number fixes this value; 200 is a build-time judgment call, generous
// enough that a real note collection rarely hits it while still bounding
// worst case (an All Roots search over a large tree).
const pickerResultCap = 200

// pickerFile is one candidate file a Picker can list or search: its
// absolute path, and the Root it was found under. The Root travels with the
// file rather than being re-derived later — once a session has several
// Roots, a file's Root is not recoverable from its path alone (two Roots
// can each contain a same-named file), and M3.1's Result row needs it to
// show "Root + path relative to that Root".
type pickerFile struct {
	root root
	path string
}

// listFiles is the File Picker's candidate set (S3, S13): every Readable
// file under scope, found via `rg --files` (S14) with the Readable
// allowlist (tree.IsReadable) applied on top — rg itself has no notion of
// it. A Broken Root is skipped, never an error (S6, liveRootsInScope); a
// live Root that rg still cannot read (e.g. permission denied, or a TOCTOU
// race with liveRootsInScope's own stat) is likewise absorbed by rgFiles
// rather than aborting the whole scope, so one unreadable Root never blanks
// out the rest.
//
// Unranked and uncapped: M1.3 fuzzy-matches this set against the query and
// caps it. Correctness, not speed, is this task's job — a slow-Root
// complaint is whichever later milestone first hits one.
func listFiles(roots []root, active int, scope pickerScope) []pickerFile {
	var files []pickerFile
	for _, r := range liveRootsInScope(roots, active, scope) {
		for _, rel := range rgFiles(r.path) {
			if tree.IsReadable(filepath.Base(rel)) {
				files = append(files, pickerFile{root: r, path: filepath.Join(r.path, rel)})
			}
		}
	}
	return files
}

// rgFiles runs `rg --files` with dir as ripgrep's working directory,
// returning the slash-relative paths it reports (relative to dir, since
// that's where it ran) — or nil when the command fails, which is how a
// directory rg cannot read at all is absorbed into "no files from this
// Root" rather than an error (S6's shape, extended past just the
// already-stat'd-missing case liveRootsInScope filters).
//
// --no-require-git widens rg's default: by default rg only honours a
// .gitignore file when the target sits inside a git repository (verified
// against the sandbox's real rg, `.charm/cycle/findings/` — a plain,
// non-git noggin Root with its own `.gitignore` would otherwise have it
// silently ignored). A Root is a plain folder, not necessarily a git
// checkout, so this flag is what makes D28's "rg's ignore rules govern"
// hold for one regardless.
func rgFiles(dir string) []string {
	cmd := exec.Command(ripgrepBinary, "--files", "--no-require-git")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	trimmed := strings.TrimRight(string(out), "\n")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "\n")
}

// fileNameSource adapts []pickerFile to fuzzy.Source, exposing each file's
// base name only — never its path — so fuzzy.FindFrom cannot see, and so
// cannot be swayed by, path components (D31: the File Picker matches the
// file name only).
type fileNameSource []pickerFile

func (s fileNameSource) String(i int) string { return filepath.Base(s[i].path) }
func (s fileNameSource) Len() int            { return len(s) }

// fuzzyFiles ranks listFiles' candidate set against query by file name only
// (D31) and caps the result at pickerResultCap, reporting whether the cap
// truncated it (S12). fuzzy.FindFrom already returns matches best-first (its
// own bonus/penalty scoring), so ranking here is just capping that order —
// listFiles is deliberately the unranked, uncapped half of this split.
func fuzzyFiles(files []pickerFile, query string) (results []pickerFile, truncated bool) {
	matches := fuzzy.FindFrom(query, fileNameSource(files))
	n := len(matches)
	if n > pickerResultCap {
		n = pickerResultCap
		truncated = true
	}
	results = make([]pickerFile, n)
	for i := 0; i < n; i++ {
		results[i] = files[matches[i].Index]
	}
	return results, truncated
}
