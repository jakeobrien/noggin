package tree

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Node is one entry in the browsable tree: a directory (with Children) or a
// readable file (Children nil). Path is absolute so the viewer can load a file
// without knowing the root.
type Node struct {
	Name     string
	Path     string
	IsDir    bool
	Children []*Node
}

// Walk builds the browsable tree rooted at root, applying the readable-file
// policy (S6) and honouring `.gitignore`. `.git` is always excluded; dotfiles
// and hidden directories are otherwise kept; directories are always listed even
// when they hold nothing readable. Children are sorted directories-first, then
// by name.
//
// root must be a directory; a file argument is M3.1's concern (cwd becomes the
// root), not the walk's. Walk does not stat root itself to special-case a
// non-directory: walkDir's os.ReadDir fails naturally (ENOTDIR) and that error
// propagates, which is all a caller-contract violation needs.
func Walk(root string) (*Node, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	node := &Node{Name: filepath.Base(abs), Path: abs, IsDir: true}
	if err := walkDir(node, "", nil); err != nil {
		return nil, err
	}
	return node, nil
}

// Entry is one entry Walk lists in a directory: a child's name and whether it
// is a directory. It is what ListDir reports, one level deep.
type Entry struct {
	Name  string
	IsDir bool
}

// ListDir reports the entries Walk would give dir's node as Children, in the
// same order, without recursing: the same readable-file policy, `.git`
// exclusion and `.gitignore` files from root down to dir (dir's own
// included). dir must be root or a directory under it that Walk would list.
// A missing or unreadable dir is an error.
//
// It exists so a caller holding an already-walked tree can re-check one
// directory cheaply (the staleness check, cmd/noggin) and get the answer the
// tree would have, rather than re-walking the whole Root to find out.
func ListDir(root, dir string) ([]Entry, error) {
	rel, err := filepath.Rel(root, dir)
	if err != nil {
		return nil, err
	}
	rel = filepath.ToSlash(rel)
	if rel == "." {
		rel = ""
	}

	// The ancestors' `.gitignore` files, root first, exactly as walkDir would
	// have accumulated them on the way down; dir's own is added by
	// readVisible.
	var stack []ignoreFile
	ancestor, ancestorRel := root, ""
	if rel != "" {
		parts := strings.Split(rel, "/")
		for i, part := range parts {
			if data, err := os.ReadFile(filepath.Join(ancestor, ".gitignore")); err == nil {
				stack = append(stack, parseIgnore(ancestorRel, data))
			}
			if i == len(parts)-1 {
				break
			}
			ancestor = filepath.Join(ancestor, part)
			ancestorRel = strings.Join(parts[:i+1], "/")
		}
	}

	entries, _, err := readVisible(dir, rel, stack)
	if err != nil {
		return nil, err
	}
	sortEntries(entries)
	return entries, nil
}

// readVisible lists dir (at slash path rel from the walk root) and keeps what
// Walk shows: `.git` and anything a `.gitignore` in stack or in dir itself
// ignores are dropped, directories are always kept, files only if Markdown.
// It returns the entries unsorted, plus stack extended with dir's own
// `.gitignore` for the caller to pass down to dir's children.
func readVisible(dir, rel string, stack []ignoreFile) ([]Entry, []ignoreFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, err
	}

	// A `.gitignore` in this directory applies to this directory's entries and
	// everything below, appended after the ancestors so its lines win.
	if data, err := os.ReadFile(filepath.Join(dir, ".gitignore")); err == nil {
		stack = append(stack, parseIgnore(rel, data))
	}

	var out []Entry
	for _, entry := range entries {
		name := entry.Name()
		if name == ".git" {
			continue // built-in exclusion, independent of any `.gitignore`
		}
		isDir := entry.IsDir()
		if ignored(stack, childRel(rel, name), isDir) {
			continue
		}
		if isDir || IsMarkdown(name) {
			out = append(out, Entry{Name: name, IsDir: isDir})
		}
	}
	return out, stack, nil
}

// walkDir populates dir.Children. rel is dir's slash path relative to the walk
// root ("" at the root); stack is the `.gitignore` files accumulated from the
// root down to dir's parent.
func walkDir(dir *Node, rel string, stack []ignoreFile) error {
	entries, stack, err := readVisible(dir.Path, rel, stack)
	if err != nil {
		return err
	}
	sortEntries(entries)

	for _, e := range entries {
		child := &Node{
			Name:  e.Name,
			Path:  filepath.Join(dir.Path, e.Name),
			IsDir: e.IsDir,
		}
		if e.IsDir {
			if err := walkDir(child, childRel(rel, e.Name), stack); err != nil {
				return err
			}
		}
		dir.Children = append(dir.Children, child)
	}
	return nil
}

// childRel is name's slash path from the walk root, given its parent's.
func childRel(rel, name string) string {
	if rel == "" {
		return name
	}
	return rel + "/" + name
}

// sortEntries orders entries the way Walk orders Children: directories first,
// then by name.
func sortEntries(entries []Entry) {
	sort.SliceStable(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]
		if a.IsDir != b.IsDir {
			return a.IsDir // directories first
		}
		return a.Name < b.Name
	})
}
