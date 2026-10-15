// Package tree walks a root directory into the folder/file tree noggin browses:
// the file classification policy (Markdown file, Text file, binary), a
// built-in `.git` exclusion, and `.gitignore` support. It produces a [Node]
// tree the Files Nav pane renders and navigates.
package tree

import (
	"path/filepath"
	"strings"
)

// markdown is the single, trivially adjustable declaration of what noggin
// classifies as a Markdown file (GLOSSARY.md). Extensions and the
// conventional extensionless names are matched case-insensitively.
// Every other non-binary file is a Text file.
var markdown = struct {
	exts  map[string]bool
	names map[string]bool
}{
	exts: map[string]bool{
		".md":       true,
		".markdown": true,
		".mdown":    true,
		".mkd":      true,
	},
	names: map[string]bool{
		"README":    true,
		"CHANGELOG": true,
	},
}

// IsMarkdown reports whether a file with the given base name is a Markdown
// file: a `.md`/`.markdown`/`.mdown`/`.mkd` extension or an
// extensionless `README`/`CHANGELOG`, case-insensitive. `.txt` and `LICENSE`
// are Text files, not Markdown.
//
// The Pickers' search layer (cmd/noggin) applies this same predicate to
// ripgrep's file list so the classification stays declared in one place.
// Directories are not subject to it: they are always listed.
func IsMarkdown(name string) bool {
	if ext := filepath.Ext(name); ext != "" {
		return markdown.exts[strings.ToLower(ext)]
	}
	return markdown.names[strings.ToUpper(name)]
}

// Filter is the File Filter: which files the Tree lists. Directories
// are always listed. The zero value is FilterMarkdown, the session default.
type Filter int

const (
	// FilterMarkdown lists Markdown files only.
	FilterMarkdown Filter = iota
	// FilterAll lists Markdown files plus Text files; binaries stay hidden by
	// the sniff (CheckOpenable's), however they are named.
	FilterAll
)

// String is the filter's display name, as the Files Nav status row shows it.
func (f Filter) String() string {
	if f == FilterAll {
		return "All"
	}
	return "Markdown"
}

// Toggled is the other filter.
func (f Filter) Toggled() Filter {
	if f == FilterAll {
		return FilterMarkdown
	}
	return FilterAll
}

// Shows reports whether f lists the (non-directory) file called name at path.
// The Pickers apply it to ripgrep's file list, so the File Filter is decided in
// one place. The name test comes first so Markdown mode, and every Markdown file under
// All, never touches the filesystem. Under All, every other file is statted
// and (once per mtime/size) sniffed, so the first walk of a big un-ignored
// tree costs a read of each file's head.
func (f Filter) Shows(path, name string) bool {
	if IsMarkdown(name) {
		return true
	}
	return f == FilterAll && IsNotBinary(path)
}
