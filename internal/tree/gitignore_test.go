package tree

import "testing"

// TestIgnoreMatch exercises the gitignore subset the walk relies on: plain
// names at any depth, extension globs, anchoring, directory-only, negation and
// `**`.
func TestIgnoreMatch(t *testing.T) {
	cases := []struct {
		name       string
		gitignore  string
		path       string
		isDir      bool
		wantIgnore bool
	}{
		{"plain name at root", "ignored.md\n", "ignored.md", false, true},
		{"plain name nested", "ignored.md\n", "a/b/ignored.md", false, true},
		{"non-match", "ignored.md\n", "keep.md", false, false},
		{"extension glob", "*.log\n", "a/b/c.log", false, true},
		{"extension glob non-match", "*.log\n", "a/b/c.txt", false, false},
		{"anchored matches at base", "/build\n", "build", true, true},
		{"anchored not nested", "/build\n", "a/build", true, false},
		{"embedded slash anchors", "docs/tmp\n", "docs/tmp", true, true},
		{"embedded slash not elsewhere", "docs/tmp\n", "a/docs/tmp", true, false},
		{"dir-only ignores dir", "build/\n", "build", true, true},
		{"dir-only spares file", "build/\n", "build", false, false},
		{"comment is inert", "# ignored.md\n", "ignored.md", false, false},
		{"blank line inert", "\n\n", "ignored.md", false, false},
		{"question mark one char", "file?.md\n", "file1.md", false, true},
		{"question mark not slash", "a?b\n", "a/b", false, false},
		{"char class", "[abc].md\n", "b.md", false, true},
		{"char class negation", "[!abc].md\n", "b.md", false, false},
		{"double star spans dirs", "**/tmp\n", "a/b/tmp", true, true},
		{"trailing double star", "logs/**\n", "logs/a/b.log", false, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			files := []ignoreFile{parseIgnore("", []byte(c.gitignore))}
			if got := ignored(files, c.path, c.isDir); got != c.wantIgnore {
				t.Errorf("ignored(%q, dir=%v) = %v, want %v", c.path, c.isDir, got, c.wantIgnore)
			}
		})
	}
}

// TestIgnoreNegation checks last-match-wins re-inclusion within one file.
func TestIgnoreNegation(t *testing.T) {
	files := []ignoreFile{parseIgnore("", []byte("*.md\n!keep.md\n"))}
	if !ignored(files, "drop.md", false) {
		t.Error("drop.md should be ignored by *.md")
	}
	if ignored(files, "keep.md", false) {
		t.Error("keep.md should be re-included by !keep.md")
	}
}

// TestIgnoreNestedPrecedence checks a deeper `.gitignore` overriding a shallow
// one, and that a file's patterns only apply under its own directory.
func TestIgnoreNestedPrecedence(t *testing.T) {
	files := []ignoreFile{
		parseIgnore("", []byte("*.md\n")),
		parseIgnore("docs", []byte("!keep.md\n")),
	}
	if !ignored(files, "top.md", false) {
		t.Error("top.md should be ignored by root *.md")
	}
	if ignored(files, "docs/keep.md", false) {
		t.Error("docs/keep.md should be re-included by docs/.gitignore")
	}
	if !ignored(files, "docs/other.md", false) {
		t.Error("docs/other.md should still be ignored by root *.md")
	}
}
