package main

import (
	"encoding/json"
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

// grepResult is one matching line the Grep Picker can show (S4): which file
// (with its Root, like pickerFile carries) it was found in, its 1-based
// line number and the line's own text. The line number is rg's raw-file
// count, not the Document Viewer's rendered/wrapped line index — D34 records
// how M4.3/M4.4 bridge that.
type grepResult struct {
	file pickerFile
	line int
	text string
}

// grepFiles searches query literally (S4: rg's --fixed-strings, so regex
// metacharacters in a half-typed query are just characters — D29) with
// smart case across the Readable files in scope: rg's own --smart-case
// flag implements exactly find.go's S8 semantic (all-lowercase matches
// either case; any uppercase makes it case-sensitive), so this reuses that
// behaviour by name rather than reimplementing find.go's fold-and-compare.
//
// It runs rg over the *explicit* file list listFiles already produced,
// rather than letting rg walk Scope's directories itself the way rgFiles
// does for the file list: verified against the sandbox's real rg (15.1.0),
// rg does not apply `.gitignore`/hidden-file rules to files named explicitly
// on its command line — only to directories it walks. listFiles has
// already applied every one of those rules (plus the Readable allowlist and
// Broken-Root skipping, M1.2) to produce that list, so searching exactly
// those files is what makes grep agree with the File Picker's own view of
// Scope, rather than re-deriving (and risking disagreeing with) it.
//
// Returns one grepResult per matching line, in rg's own order, capped at
// pickerResultCap with the same truncation-flag shape as fuzzyFiles (S12).
// An empty file set or an empty query short-circuits to no results without
// invoking rg at all — an empty query would otherwise be an empty-pattern
// rg run, which matches every line of every file, not "nothing to search".
func grepFiles(roots []root, active int, scope pickerScope, query string) (results []grepResult, truncated bool) {
	files := listFiles(roots, active, scope)
	if len(files) == 0 || query == "" {
		return nil, false
	}

	byPath := make(map[string]pickerFile, len(files))
	args := make([]string, 0, len(files)+5)
	args = append(args, "--fixed-strings", "--smart-case", "--json", "--", query)
	for _, f := range files {
		byPath[f.path] = f
		args = append(args, f.path)
	}

	cmd := exec.Command(ripgrepBinary, args...)
	out, err := cmd.Output()
	if err != nil {
		// rg exits non-zero both for "no matches" and for a run it could not
		// complete; either is absorbed as "no results" (same shape rgFiles
		// gives an unreadable Root, M1.2), not an error surfaced to the user.
		return nil, false
	}

	for _, line := range strings.Split(strings.TrimRight(string(out), "\n"), "\n") {
		var msg rgJSONMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil || msg.Type != "match" {
			continue
		}
		file, ok := byPath[msg.Data.Path.Text]
		if !ok {
			continue
		}
		if len(results) == pickerResultCap {
			truncated = true
			break
		}
		results = append(results, grepResult{
			file: file,
			line: msg.Data.LineNumber,
			text: strings.TrimRight(msg.Data.Lines.Text, "\n"),
		})
	}
	return results, truncated
}

// rgJSONMessage is the subset of one line of rg's `--json` JSON-Lines output
// (docs.rs/grep-printer, the "match" message shape) grepFiles needs: which
// file, which line number, the line's own text. Parsed instead of rg's
// plain `path:line:text` text output because a colon can legitimately
// appear in either a path or a line's text, which that format leaves
// ambiguous and JSON does not.
type rgJSONMessage struct {
	Type string `json:"type"`
	Data struct {
		Path struct {
			Text string `json:"text"`
		} `json:"path"`
		Lines struct {
			Text string `json:"text"`
		} `json:"lines"`
		LineNumber int `json:"line_number"`
	} `json:"data"`
}
