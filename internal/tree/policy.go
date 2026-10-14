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
// classifies as a Markdown file (D140, GLOSSARY.md). Extensions and the
// conventional extensionless names are matched case-insensitively (D23).
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
// file (D140): a `.md`/`.markdown`/`.mdown`/`.mkd` extension or an
// extensionless `README`/`CHANGELOG`, case-insensitive. `.txt` and `LICENSE`
// are Text files, not Markdown.
//
// Until the File Filter lands the Tree and the Pickers list Markdown files
// only; the Picker's search layer (cmd/noggin) applies this same predicate to
// ripgrep's file list so the classification stays declared in one place.
// Directories are not subject to it: they are always listed.
func IsMarkdown(name string) bool {
	if ext := filepath.Ext(name); ext != "" {
		return markdown.exts[strings.ToLower(ext)]
	}
	return markdown.names[strings.ToUpper(name)]
}
