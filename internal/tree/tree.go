package tree

import (
	"os"
	"path/filepath"
	"sort"
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

// walkDir populates dir.Children. rel is dir's slash path relative to the walk
// root ("" at the root); stack is the `.gitignore` files accumulated from the
// root down to dir's parent.
func walkDir(dir *Node, rel string, stack []ignoreFile) error {
	entries, err := os.ReadDir(dir.Path)
	if err != nil {
		return err
	}

	// A `.gitignore` in this directory applies to this directory's entries and
	// everything below, appended after the ancestors so its lines win.
	if data, err := os.ReadFile(filepath.Join(dir.Path, ".gitignore")); err == nil {
		stack = append(stack, parseIgnore(rel, data))
	}

	for _, entry := range entries {
		name := entry.Name()
		if name == ".git" {
			continue // built-in exclusion, independent of any `.gitignore`
		}
		childRel := name
		if rel != "" {
			childRel = rel + "/" + name
		}
		isDir := entry.IsDir()
		if ignored(stack, childRel, isDir) {
			continue
		}

		child := &Node{
			Name:  name,
			Path:  filepath.Join(dir.Path, name),
			IsDir: isDir,
		}
		if isDir {
			if err := walkDir(child, childRel, stack); err != nil {
				return err
			}
			dir.Children = append(dir.Children, child)
		} else if IsReadable(name) {
			dir.Children = append(dir.Children, child)
		}
	}

	sort.SliceStable(dir.Children, func(i, j int) bool {
		a, b := dir.Children[i], dir.Children[j]
		if a.IsDir != b.IsDir {
			return a.IsDir // directories first
		}
		return a.Name < b.Name
	})
	return nil
}
