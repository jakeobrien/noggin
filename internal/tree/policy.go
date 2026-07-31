// Package tree walks a root directory into the folder/file tree noggin browses:
// the readable-file policy (which files count as documents), a built-in `.git`
// exclusion, and `.gitignore` support. It produces a [Node] tree the Files Nav
// pane (M2.2) renders and navigates; M2.1 is the walk and the policy alone.
package tree

import (
	"path/filepath"
	"strings"
)

// readable is the single, trivially adjustable declaration of what noggin
// considers a readable document (S6). Everything else the walk finds is a
// folder (always shown) or ignored. Extensions are matched case-insensitively;
// names match conventional extensionless documents.
//
// This is deliberately the one place the allowlist lives — retuning what counts
// as readable is an edit here and nowhere else.
var readable = struct {
	exts  map[string]bool
	names map[string]bool
}{
	exts: map[string]bool{
		".md":       true,
		".markdown": true,
		".txt":      true,
	},
	names: map[string]bool{
		"README":    true,
		"LICENSE":   true,
		"CHANGELOG": true,
	},
}

// isReadable reports whether a file with the given base name is a document
// noggin will show. Directories are not subject to this — they are always
// listed (S6: folders with nothing readable are still shown).
func isReadable(name string) bool {
	if ext := filepath.Ext(name); ext != "" {
		return readable.exts[strings.ToLower(ext)]
	}
	return readable.names[name]
}
