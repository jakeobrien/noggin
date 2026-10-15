package tree

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Node is one entry in the browsable tree: a directory (with Children) or a
// listed file (Children nil). Path is absolute so the viewer can load a file
// without knowing the root.
type Node struct {
	Name     string
	Path     string
	IsDir    bool
	Children []*Node

	// Ignored marks an entry a `.gitignore` excludes (or one under such a
	// directory). Only ever set when the walk's Options.ShowIgnored kept it;
	// otherwise such entries are not listed at all.
	Ignored bool
}

// Options bound what Walk and ListDir list. The zero value is
// the session default: Markdown files only, dotfiles shown, gitignored
// entries hidden.
type Options struct {
	// Filter is the File Filter: which non-directory files are listed.
	Filter Filter

	// HideDotfiles drops every entry whose name starts with "." (and so
	// everything under a dot-directory).
	HideDotfiles bool

	// ShowIgnored keeps entries a `.gitignore` excludes, marking them
	// Ignored, except directories named in heavyDirs, which stay excluded.
	ShowIgnored bool
}

// heavyDirs are directory names that stay excluded even when ShowIgnored
// reveals gitignored entries: dependency and cache folders big enough
// to make an eager walk crawl, and never what someone browsing notes wants.
// Matched by name at any depth, whether or not a `.gitignore` names them.
// Declared once here; the Picker's ripgrep listing excludes the same set
// (HeavyDirs).
var heavyDirs = []string{
	"node_modules",
	".venv",
	"venv",
	"__pycache__",
	".mypy_cache",
	".pytest_cache",
	".tox",
	".gradle",
	".next",
	"target",
}

// HeavyDirs returns the directory names that stay excluded while gitignored
// entries are shown (see heavyDirs).
func HeavyDirs() []string { return append([]string(nil), heavyDirs...) }

func isHeavyDir(name string) bool {
	for _, d := range heavyDirs {
		if d == name {
			return true
		}
	}
	return false
}

// Walk builds the browsable tree rooted at root, listing what opts allows: the
// files its Filter shows, dotfiles unless HideDotfiles, and gitignored
// entries only when ShowIgnored (marked Ignored, heavyDirs excepted). `.git`
// is always excluded; directories are always listed even when they hold
// nothing listable. Children are sorted directories-first, then by name.
//
// root must be a directory; a file argument is M3.1's concern (cwd becomes the
// root), not the walk's. Walk does not stat root itself to special-case a
// non-directory: walkDir's os.ReadDir fails naturally (ENOTDIR) and that error
// propagates, which is all a caller-contract violation needs.
func Walk(root string, opts Options) (*Node, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	node := &Node{Name: filepath.Base(abs), Path: abs, IsDir: true}
	if err := walkDir(node, "", nil, opts); err != nil {
		return nil, err
	}
	return node, nil
}

// Entry is one entry Walk lists in a directory: a child's name and whether it
// is a directory. It is what ListDir reports, one level deep.
type Entry struct {
	Name    string
	IsDir   bool
	Ignored bool
}

// ListDir reports the entries Walk would give dir's node as Children, in the
// same order, without recursing: the same options, `.git` exclusion and
// `.gitignore` files from root down to dir (dir's own
// included). dir must be root or a directory under it that Walk would list.
// A missing or unreadable dir is an error.
//
// It exists so a caller holding an already-walked tree can re-check one
// directory cheaply (the staleness check, cmd/noggin) and get the answer the
// tree would have, rather than re-walking the whole Root to find out.
func ListDir(root, dir string, opts Options) ([]Entry, error) {
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
	// readVisible. Whether dir itself sits at or under an ignored directory
	// is worked out on the same way down, so its entries inherit Ignored the
	// way walkDir passes it on.
	var stack []ignoreFile
	dirIgnored := false
	ancestor, ancestorRel := root, ""
	if rel != "" {
		for _, part := range strings.Split(rel, "/") {
			if data, err := os.ReadFile(filepath.Join(ancestor, ".gitignore")); err == nil {
				stack = append(stack, parseIgnore(ancestorRel, data))
			}
			partRel := childRel(ancestorRel, part)
			if !dirIgnored && ignored(stack, partRel, true) {
				dirIgnored = true
			}
			ancestor = filepath.Join(ancestor, part)
			ancestorRel = partRel
		}
	}

	entries, _, err := readVisible(dir, rel, stack, dirIgnored, opts)
	if err != nil {
		return nil, err
	}
	sortEntries(entries)
	return entries, nil
}

// readVisible lists dir (at slash path rel from the walk root) and keeps what
// Walk shows: `.git` is dropped; dotfiles are dropped under HideDotfiles;
// anything a `.gitignore` in stack or in dir itself ignores — or everything,
// when dirIgnored says dir itself is ignored — is dropped unless ShowIgnored,
// which keeps it marked Ignored (heavyDirs excepted); directories are
// otherwise always kept, files only if the Filter shows them.
// It returns the entries unsorted, plus stack extended with dir's own
// `.gitignore` for the caller to pass down to dir's children.
func readVisible(dir, rel string, stack []ignoreFile, dirIgnored bool, opts Options) ([]Entry, []ignoreFile, error) {
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
		if opts.HideDotfiles && strings.HasPrefix(name, ".") {
			continue
		}
		isDir := entry.IsDir()
		ign := dirIgnored || ignored(stack, childRel(rel, name), isDir)
		if ign && !opts.ShowIgnored {
			continue
		}
		if opts.ShowIgnored && isDir && isHeavyDir(name) {
			continue
		}
		if isDir || opts.Filter.Shows(filepath.Join(dir, name), name) {
			out = append(out, Entry{Name: name, IsDir: isDir, Ignored: ign})
		}
	}
	return out, stack, nil
}

// walkDir populates dir.Children. rel is dir's slash path relative to the walk
// root ("" at the root); stack is the `.gitignore` files accumulated from the
// root down to dir's parent.
func walkDir(dir *Node, rel string, stack []ignoreFile, opts Options) error {
	entries, stack, err := readVisible(dir.Path, rel, stack, dir.Ignored, opts)
	if err != nil {
		return err
	}
	sortEntries(entries)

	for _, e := range entries {
		child := &Node{
			Name:    e.Name,
			Path:    filepath.Join(dir.Path, e.Name),
			IsDir:   e.IsDir,
			Ignored: e.Ignored,
		}
		if e.IsDir {
			if err := walkDir(child, childRel(rel, e.Name), stack, opts); err != nil {
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
