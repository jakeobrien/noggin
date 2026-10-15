package tree

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// writeTree materialises a fixture tree under a fresh temp dir. Keys are slash
// paths; a key ending in "/" is an empty directory, otherwise a file (its
// parent dirs are created as needed). It returns the root.
func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, content := range files {
		p := filepath.Join(root, filepath.FromSlash(name))
		if strings.HasSuffix(name, "/") {
			if err := os.MkdirAll(p, 0o755); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// collect returns the listed files and the directories of a walked tree, each
// as sorted slash paths relative to root, so a walk can be asserted as a set.
func collect(root *Node, base string) (files, dirs []string) {
	for _, c := range root.Children {
		rel := c.Name
		if base != "" {
			rel = base + "/" + c.Name
		}
		if c.IsDir {
			dirs = append(dirs, rel)
			f, d := collect(c, rel)
			files = append(files, f...)
			dirs = append(dirs, d...)
		} else {
			files = append(files, rel)
		}
	}
	sort.Strings(files)
	sort.Strings(dirs)
	return files, dirs
}

// TestWalkMarkdownSet is the M2.1 acceptance criterion: over a fixture tree the
// walk returns exactly the expected Markdown file set (LICENSE and notes.txt are Text, not listed until the File Filter lands) — a gitignored `.md`
// excluded, `.git/` absent, `.charm/notes.md` present, `main.go` absent, a
// `README` present — plus the surrounding policy (dotdirs descended, a
// gitignored directory pruned, empty folders still listed).
func TestWalkMarkdownSet(t *testing.T) {
	root := writeTree(t, map[string]string{
		".gitignore":       "ignored.md\nbuild/\n",
		"README":           "# readme",
		"LICENSE":          "license",
		"notes.txt":        "plain text",
		"main.go":          "package main",
		"ignored.md":       "gitignored",
		".git/HEAD":        "ref: refs/heads/main",
		".charm/notes.md":  "# notes",
		"docs/guide.md":    "# guide",
		"docs/diagram.png": "not text",
		"build/output.md":  "generated",
		"empty/":           "",
	})

	node, err := Walk(root, FilterMarkdown)
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}

	gotFiles, gotDirs := collect(node, "")

	wantFiles := []string{
		"README",
		".charm/notes.md",
		"docs/guide.md",
	}
	sort.Strings(wantFiles)
	if !reflect.DeepEqual(gotFiles, wantFiles) {
		t.Errorf("markdown files:\n got %v\nwant %v", gotFiles, wantFiles)
	}

	wantDirs := []string{".charm", "docs", "empty"}
	if !reflect.DeepEqual(gotDirs, wantDirs) {
		t.Errorf("directories:\n got %v\nwant %v", gotDirs, wantDirs)
	}
}

// TestWalkChildOrder checks the documented ordering: directories first, then
// files, each alphabetically.
func TestWalkChildOrder(t *testing.T) {
	root := writeTree(t, map[string]string{
		"b.md":      "",
		"a.md":      "",
		"zdir/x.md": "",
		"mdir/y.md": "",
		"README":    "",
	})
	node, err := Walk(root, FilterMarkdown)
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	var got []string
	for _, c := range node.Children {
		got = append(got, c.Name)
	}
	want := []string{"mdir", "zdir", "README", "a.md", "b.md"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("child order: got %v want %v", got, want)
	}
}

// TestWalkPathsAbsolute confirms nodes carry an absolute, loadable path.
func TestWalkPathsAbsolute(t *testing.T) {
	root := writeTree(t, map[string]string{"README": "hi"})
	node, err := Walk(root, FilterMarkdown)
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	if len(node.Children) != 1 {
		t.Fatalf("want 1 child, got %d", len(node.Children))
	}
	p := node.Children[0].Path
	if !filepath.IsAbs(p) {
		t.Errorf("child path not absolute: %q", p)
	}
	if _, err := os.Stat(p); err != nil {
		t.Errorf("child path not loadable: %v", err)
	}
}

// entriesOf is n's Children as the Entries ListDir reports, in order.
func entriesOf(n *Node) []Entry {
	out := []Entry{}
	for _, c := range n.Children {
		out = append(out, Entry{Name: c.Name, IsDir: c.IsDir})
	}
	return out
}

// findDir returns the directory node at slash path rel under root.
func findDir(t *testing.T, root *Node, rel string) *Node {
	t.Helper()
	n := root
	for _, part := range strings.Split(rel, "/") {
		var next *Node
		for _, c := range n.Children {
			if c.IsDir && c.Name == part {
				next = c
			}
		}
		if next == nil {
			t.Fatalf("no directory %q in walked tree", rel)
		}
		n = next
	}
	return n
}

// TestListDirMatchesWalk pins ListDir's one contract: for any directory Walk
// lists, ListDir reports exactly that directory's Children, in the same order
// — the same file-classification policy, `.git` exclusion and `.gitignore` files
// from the root down to (and including) the directory itself.
func TestListDirMatchesWalk(t *testing.T) {
	root := writeTree(t, map[string]string{
		".gitignore":               "*.log\nbuild/\n",
		".git/HEAD":                "ref",
		"README":                   "",
		"top.md":                   "",
		"image.png":                "",
		"debug.log":                "",
		"build/out.md":             "",
		"empty/":                   "",
		"notes/.gitignore":         "secret.md\n",
		"notes/a.md":               "",
		"notes/secret.md":          "",
		"notes/trace.log":          "",
		"notes/deep/b.md":          "",
		"notes/deep/secret.md":     "",
		"notes/deep/.DS_Store":     "",
		"notes/deep/c.txt":         "",
		"notes/deep/nested/":       "",
		"notes/deep/build/hide.md": "",
	})
	walked, err := Walk(root, FilterMarkdown)
	if err != nil {
		t.Fatal(err)
	}

	for _, rel := range []string{"", "notes", "notes/deep", "empty"} {
		dir := walked
		if rel != "" {
			dir = findDir(t, walked, rel)
		}
		got, err := ListDir(walked.Path, dir.Path, FilterMarkdown)
		if err != nil {
			t.Fatalf("ListDir(%q): %v", rel, err)
		}
		if got == nil {
			got = []Entry{}
		}
		if want := entriesOf(dir); !reflect.DeepEqual(got, want) {
			t.Errorf("ListDir(%q) = %v, want Walk's children %v", rel, got, want)
		}
	}
}

// TestListDirMissingDirectoryErrors: a directory that no longer exists is an
// error, not an empty listing, so a caller can tell "gone" from "emptied".
func TestListDirMissingDirectoryErrors(t *testing.T) {
	root := t.TempDir()
	if _, err := ListDir(root, filepath.Join(root, "gone"), FilterMarkdown); err == nil {
		t.Error("ListDir on a missing directory returned nil error, want one")
	}
}
