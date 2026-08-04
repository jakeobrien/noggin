package main

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode/utf8"

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

// fileNameMatchPositions is which characters of name the query matched
// (S5/M6.1: "highlight matched characters within the file name"), as rune
// indexes into name, for the Results row renderer to style
// (highlightMatches, picker_view.go). nil when nothing matched, which is
// also what an empty query gives — fuzzy.Find treats an empty pattern as no
// match rather than as "everything", so a row rendered before the query
// reaches pickerQueryMinLength simply has no highlight.
//
// The positions are recomputed here rather than carried out of fuzzyFiles
// alongside each result. Carrying them would mean a second result type
// (pickerFile plus its match) threaded through pickerResults, the Preview/
// Activate accessors and every test that builds a result set by hand, to
// avoid a per-visible-row fuzzy match over one short string — the cheaper
// half of that trade is recomputation. It is exact, not an approximation:
// fuzzy.Find runs the same matcher over the same name fuzzyFiles' own
// fuzzy.FindFrom(fileNameSource) did, so the positions are the ones that
// ranked the row (pinned by TestFileNameMatchPositionsAgreeWithRanking).
//
// fuzzy's MatchedIndexes are *byte* offsets into the matched string (its
// scan loop indexes bytes, fuzzy.go's FindFromIterNoSort), while a rendered
// row is addressed in runes/cells — so they are converted here, at the one
// place that knows they came from fuzzy, rather than leaking a byte offset
// into the view layer.
func fileNameMatchPositions(name, query string) []int {
	matches := fuzzy.Find(query, []string{name})
	if len(matches) == 0 {
		return nil
	}
	byteToRune := runeIndexByByte(name)
	var positions []int
	for _, b := range matches[0].MatchedIndexes {
		if r, ok := byteToRune[b]; ok {
			positions = append(positions, r)
		}
	}
	return positions
}

// runeIndexByByte maps each rune's starting byte offset in s to its rune
// index. Both match-position functions need it: matching works in byte
// offsets (fuzzy's MatchedIndexes, strings.Index) while a rendered row is
// addressed in runes/cells, and the two only coincide for ASCII.
func runeIndexByByte(s string) map[int]int {
	byteToRune := make(map[int]int, len(s))
	for runeIdx, byteIdx := 0, 0; byteIdx < len(s); runeIdx++ {
		byteToRune[byteIdx] = runeIdx
		_, size := utf8.DecodeRuneInString(s[byteIdx:])
		byteIdx += size
	}
	return byteToRune
}

// grepMatchPositions is which characters of a matched line the grep query
// matched (S5/M6.2: "highlight the matched text within the line"), as rune
// indexes into text — the Grep Picker's counterpart to
// fileNameMatchPositions, returning the same shape (one position per matched
// rune) so both modes' rows share one highlighter (highlightMatches,
// picker_view.go).
//
// Matching mirrors what actually produced the row: a literal substring
// search (rg's --fixed-strings, D29 — a query's regex metacharacters are
// just characters) with smart case (rg's --smart-case: an all-lowercase
// query matches either case, any uppercase makes it case-sensitive). Every
// occurrence in the line is reported, not just the first: rg reports the
// line once however many times the query occurs in it, so emphasizing only
// one of several identical hits would misrepresent the row it labels.
//
// The smart-case path searches a folded copy and maps offsets back through
// toOrig (foldLower/origSpan, find.go) rather than folding in place, for the
// reason findMatches' own comment gives: folding is not byte-length
// preserving (U+212A, three bytes, folds to a one-byte 'k'), so offsets from
// the folded string are not offsets into text. Reusing find.go's helpers
// here keeps one implementation of that translation, and one of S8's
// smart-case rule, rather than a second copy that can drift.
func grepMatchPositions(text, query string) []int {
	if query == "" || text == "" {
		return nil
	}
	smartCase := query == strings.ToLower(query)
	haystack, needle := text, query
	var toOrig []int
	if smartCase {
		haystack, toOrig = foldLower(text)
		needle = strings.ToLower(query)
	}

	type span struct{ start, end int }
	var spans []span
	for pos := 0; ; {
		idx := strings.Index(haystack[pos:], needle)
		if idx < 0 {
			break
		}
		hStart := pos + idx
		hEnd := hStart + len(needle)
		start, end := origSpan(hStart, hEnd, smartCase, toOrig)
		spans = append(spans, span{start, end})
		pos = hEnd // non-overlapping, in haystack's own byte space
	}
	if len(spans) == 0 {
		return nil
	}

	var positions []int
	runes := runeIndexByByte(text)
	for _, s := range spans {
		for b := s.start; b < s.end; b++ {
			if r, ok := runes[b]; ok {
				positions = append(positions, r)
			}
		}
	}
	return positions
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

// grepArgMaxBytes bounds the total byte length of file-path arguments a
// single rg invocation may receive (TRIAGE T1 / D43): grepFiles used to
// pass every file in Scope as a literal CLI argument to one rg call, so an
// All Roots search over a large tree could push the argument list past the
// OS's ARG_MAX — and cmd.Output()'s resulting exec error was silently
// folded into the same "no results" path as a genuine no-match, with no
// signal the search hadn't actually run. batchFilePaths splits the file
// list into groups that stay comfortably under this limit, well below
// real-world ARG_MAX values on both Linux (single-digit MB) and macOS
// (historically ~256KB), so the failure mode is avoided rather than
// detected after the fact. It's a var, not a const, purely so a test can
// shrink it to force multiple batches without needing an unwieldy number
// of real files.
//
// Test-only mutation contract (TRIAGE T18): tests that need a small batch
// size mutate this package-level var directly (save the original, shrink
// it, defer restoring it — see TestGrepFilesBatchesAcrossArgSizeLimit).
// That save/shrink/restore pattern is not goroutine-safe: no test in this
// package may call t.Parallel() while any test still mutates
// grepArgMaxBytes this way, or the mutation could leak across concurrently
// running tests. Threading the limit through grepFiles as a parameter
// instead would remove this hazard entirely, but isn't worth the call-chain
// churn (grepFiles, batchFilePaths and their one production caller) unless
// this package actually adopts parallel tests.
var grepArgMaxBytes = 128 * 1024

// batchFilePaths splits files into groups whose combined path byte length
// stays at or under maxBytes per group, preserving files' relative order
// within and across batches — grepFiles's cap/truncation contract does not
// depend on batch boundaries, but a stable order keeps behaviour
// deterministic. Never produces an empty batch; a single file whose own
// path exceeds maxBytes still gets a (oversized) batch of its own rather
// than being dropped.
func batchFilePaths(files []pickerFile, maxBytes int) [][]pickerFile {
	var batches [][]pickerFile
	var current []pickerFile
	size := 0
	for _, f := range files {
		added := len(f.path) + 1 // +1: argv separator overhead, not exact
		if len(current) > 0 && size+added > maxBytes {
			batches = append(batches, current)
			current = nil
			size = 0
		}
		current = append(current, f)
		size += added
	}
	if len(current) > 0 {
		batches = append(batches, current)
	}
	return batches
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
// The file list is run through rg in batches (batchFilePaths, grepBatch)
// rather than one invocation (D43/M1.4a) — see grepArgMaxBytes. Batches are
// merged in order and pickerResultCap/truncation (S12, matching fuzzyFiles)
// is applied across that merged total, not per batch, so a match count that
// only exceeds the cap once batches are combined is still reported
// correctly as truncated. An empty file set or an empty query short-circuits
// to no results without invoking rg at all — an empty query would otherwise
// be an empty-pattern rg run, which matches every line of every file, not
// "nothing to search".
func grepFiles(roots []root, active int, scope pickerScope, query string) (results []grepResult, truncated bool) {
	files := listFiles(roots, active, scope)
	if len(files) == 0 || query == "" {
		return nil, false
	}

	for _, batch := range batchFilePaths(files, grepArgMaxBytes) {
		for _, r := range grepBatch(batch, query) {
			if len(results) == pickerResultCap {
				return results, true
			}
			results = append(results, r)
		}
	}
	return results, false
}

// grepBatch runs a single rg invocation over exactly the given files (one
// batch from batchFilePaths) and returns every matching line, uncapped —
// grepFiles is what applies pickerResultCap, across the merged results of
// every batch.
func grepBatch(files []pickerFile, query string) []grepResult {
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
		// rg exits non-zero both for "no matches in this batch" and for a
		// run it could not complete; either is absorbed as "no results from
		// this batch" (same shape rgFiles gives an unreadable Root, M1.2),
		// not an error surfaced to the user. grepArgMaxBytes' batching is
		// what keeps this from being an ARG_MAX exec failure in practice,
		// rather than something detected and distinguished here.
		return nil
	}

	var results []grepResult
	for _, line := range strings.Split(strings.TrimRight(string(out), "\n"), "\n") {
		var msg rgJSONMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil || msg.Type != "match" {
			continue
		}
		file, ok := byPath[msg.Data.Path.Text]
		if !ok {
			continue
		}
		results = append(results, grepResult{
			file: file,
			line: msg.Data.LineNumber,
			text: strings.TrimRight(msg.Data.Lines.Text, "\n"),
		})
	}
	return results
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
