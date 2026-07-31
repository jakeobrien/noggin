package tree

import (
	"regexp"
	"strings"
)

// A focused `.gitignore` matcher, not a full reimplementation of git's
// wildmatch. It hand-rolls the subset a read-only browse needs — per-directory
// `.gitignore` files, comments, negation (`!`), directory-only (trailing `/`),
// anchoring (a leading or embedded `/`), and the wildcards `*`, `?`, `[...]`
// and `**` — rather than take a dependency (the dep graph is a tracked
// concern, see .charm/cycle/PROGRESS.md) or shell out to `git check-ignore`
// (which only works inside a real git worktree, wrong for arbitrary roots).
//
// See .charm/cycle/findings/gitignore-matching.md for the semantics covered and
// the ones deliberately left out.

// ignorePattern is one compiled `.gitignore` line.
type ignorePattern struct {
	negate  bool           // a `!` line re-includes what an earlier line ignored
	dirOnly bool           // a trailing `/` matches directories only
	re      *regexp.Regexp // matched against a path relative to the owning file's dir
}

// ignoreFile is the compiled patterns of one `.gitignore`, tagged with the
// directory it lives in (as a slash path relative to the walk root, "" at root)
// so its patterns match against paths relative to that directory.
type ignoreFile struct {
	base     string
	patterns []ignorePattern
}

// parseIgnore compiles the lines of a `.gitignore` living in dir base (a slash
// path relative to the walk root).
func parseIgnore(base string, data []byte) ignoreFile {
	f := ignoreFile{base: base}
	for _, line := range strings.Split(string(data), "\n") {
		if p, ok := compilePattern(line); ok {
			f.patterns = append(f.patterns, p)
		}
	}
	return f
}

// compilePattern turns one `.gitignore` line into a matcher, reporting ok=false
// for blank and comment lines that match nothing.
func compilePattern(line string) (ignorePattern, bool) {
	// Trailing whitespace is not significant unless escaped; leading is.
	line = strings.TrimRight(line, " \t")
	if line == "" || strings.HasPrefix(line, "#") {
		return ignorePattern{}, false
	}

	var p ignorePattern
	if strings.HasPrefix(line, "!") {
		p.negate = true
		line = line[1:]
	}
	// An escaped leading '#' or '!' is a literal.
	if strings.HasPrefix(line, `\#`) || strings.HasPrefix(line, `\!`) {
		line = line[1:]
	}

	if strings.HasSuffix(line, "/") {
		p.dirOnly = true
		line = strings.TrimSuffix(line, "/")
	}
	if line == "" {
		return ignorePattern{}, false
	}

	// A slash anywhere but the trailing position anchors the pattern to the
	// `.gitignore`'s own directory; otherwise it matches at any depth below it.
	anchored := strings.Contains(line, "/")
	line = strings.TrimPrefix(line, "/")

	var b strings.Builder
	b.WriteString("^")
	if !anchored {
		b.WriteString("(?:.*/)?")
	}
	b.WriteString(translateGlob(line))
	b.WriteString("$")

	p.re = regexp.MustCompile(b.String())
	return p, true
}

// translateGlob converts a gitignore glob body to a regular expression,
// covering `*`, `?`, `[...]` and `**`. `*` and `?` do not cross `/`; `**`
// does.
func translateGlob(glob string) string {
	var b strings.Builder
	runes := []rune(glob)
	for i := 0; i < len(runes); i++ {
		c := runes[i]
		switch c {
		case '*':
			if i+1 < len(runes) && runes[i+1] == '*' {
				i++ // consume the second '*'
				// `**/` collapses to "zero or more directories".
				if i+1 < len(runes) && runes[i+1] == '/' {
					i++ // consume the '/'
					b.WriteString("(?:.*/)?")
				} else {
					b.WriteString(".*")
				}
			} else {
				b.WriteString("[^/]*")
			}
		case '?':
			b.WriteString("[^/]")
		case '[':
			// Copy the character class through, translating a leading '!'
			// negation to regexp's '^'. If it is unterminated, treat '[' as a
			// literal.
			j := i + 1
			if j < len(runes) && (runes[j] == '!' || runes[j] == '^') {
				j++
			}
			if j < len(runes) && runes[j] == ']' {
				j++
			}
			for j < len(runes) && runes[j] != ']' {
				j++
			}
			if j >= len(runes) {
				b.WriteString(`\[`)
				continue
			}
			class := string(runes[i+1 : j])
			if strings.HasPrefix(class, "!") {
				class = "^" + class[1:]
			}
			b.WriteString("[" + class + "]")
			i = j
		default:
			b.WriteString(regexp.QuoteMeta(string(c)))
		}
	}
	return b.String()
}

// ignored reports whether the entry at rel (a slash path relative to the walk
// root) is excluded by the accumulated `.gitignore` files. Later files (deeper
// directories) and later lines win, and a negation re-includes — matching
// git's last-match-wins precedence.
func ignored(files []ignoreFile, rel string, isDir bool) bool {
	result := false
	for _, f := range files {
		sub := rel
		if f.base != "" {
			// Only entries under this file's directory are subject to it.
			if !strings.HasPrefix(rel, f.base+"/") {
				continue
			}
			sub = rel[len(f.base)+1:]
		}
		for _, p := range f.patterns {
			if p.dirOnly && !isDir {
				continue
			}
			if p.re.MatchString(sub) {
				result = !p.negate
			}
		}
	}
	return result
}
