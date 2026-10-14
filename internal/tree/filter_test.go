package tree

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// filterFixture is a root holding a Markdown file, Text files (a code file, a
// `.txt`, an extensionless one), a binary named like a Text file and one named
// like an image, and a subfolder with a Text file.
func filterFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	write := func(rel, content string) {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("doc.md", "# doc\n")
	write("main.go", "package main\n")
	write("notes.txt", "notes\n")
	write("Makefile", "all:\n")
	write("data.txt.bin", "ab\x00cd")
	write("pic.png", "\x89PNG\x00\x00")
	write("sub/script.sh", "#!/bin/sh\n")
	write("sub/blob.dat", "\x00\x01\x02")
	return root
}

func childNames(n *Node) []string {
	var out []string
	for _, c := range n.Children {
		out = append(out, c.Name)
	}
	return out
}

func TestFilterMarkdownListsOnlyMarkdownFiles(t *testing.T) {
	node, err := Walk(filterFixture(t), FilterMarkdown)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := childNames(node), []string{"sub", "doc.md"}; !slices.Equal(got, want) {
		t.Errorf("Markdown filter lists %v, want %v", got, want)
	}
}

func TestFilterAllListsTextFilesAndHidesBinaries(t *testing.T) {
	node, err := Walk(filterFixture(t), FilterAll)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"sub", "Makefile", "doc.md", "main.go", "notes.txt"}
	if got := childNames(node); !slices.Equal(got, want) {
		t.Errorf("All filter lists %v, want %v (binaries hidden whatever their name)", got, want)
	}
	if got, want := childNames(node.Children[0]), []string{"script.sh"}; !slices.Equal(got, want) {
		t.Errorf("sub lists %v under All, want %v", got, want)
	}
}

// A binary named like Markdown is still Markdown by name, as before: the
// sniff only decides among non-Markdown names, and opening it is refused.
func TestFilterAllStillListsMarkdownNamedFiles(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "odd.md"), []byte("a\x00b"), 0o644); err != nil {
		t.Fatal(err)
	}
	node, err := Walk(root, FilterAll)
	if err != nil {
		t.Fatal(err)
	}
	if got := childNames(node); !slices.Equal(got, []string{"odd.md"}) {
		t.Errorf("All lists %v, want odd.md", got)
	}
}

func TestFilterAllListsAnOversizeTextFile(t *testing.T) {
	root := t.TempDir()
	big := make([]byte, MaxOpenSize+1)
	for i := range big {
		big[i] = 'x'
	}
	if err := os.WriteFile(filepath.Join(root, "big.log"), big, 0o644); err != nil {
		t.Fatal(err)
	}
	node, err := Walk(root, FilterAll)
	if err != nil {
		t.Fatal(err)
	}
	if got := childNames(node); !slices.Equal(got, []string{"big.log"}) {
		t.Errorf("All lists %v, want big.log (refused on open, not hidden)", got)
	}
}

func TestFilterAllHidesASymlinkedDirectory(t *testing.T) {
	root := filterFixture(t)
	if err := os.Symlink(filepath.Join(root, "sub"), filepath.Join(root, "link")); err != nil {
		t.Skip("symlinks unavailable:", err)
	}
	node, err := Walk(root, FilterAll)
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(childNames(node), "link") {
		t.Errorf("All lists a symlink to a directory as a file: %v", childNames(node))
	}
}

func TestFilterAllHidesADanglingSymlink(t *testing.T) {
	root := t.TempDir()
	if err := os.Symlink(filepath.Join(root, "gone"), filepath.Join(root, "dangling")); err != nil {
		t.Skip("symlinks unavailable:", err)
	}
	node, err := Walk(root, FilterAll)
	if err != nil {
		t.Fatal(err)
	}
	if got := childNames(node); len(got) != 0 {
		t.Errorf("All lists %v, want the dangling symlink hidden", got)
	}
}

func TestListDirMatchesWalkUnderEveryFilter(t *testing.T) {
	root := filterFixture(t)
	for _, f := range []Filter{FilterMarkdown, FilterAll} {
		walked, err := Walk(root, f)
		if err != nil {
			t.Fatal(err)
		}
		for _, dir := range []*Node{walked, walked.Children[0]} {
			got, err := ListDir(walked.Path, dir.Path, f)
			if err != nil {
				t.Fatal(err)
			}
			var want []Entry
			for _, c := range dir.Children {
				want = append(want, Entry{Name: c.Name, IsDir: c.IsDir})
			}
			if !slices.Equal(got, want) {
				t.Errorf("filter %v: ListDir(%s) = %v, want Walk's %v", f, dir.Name, got, want)
			}
		}
	}
}

func TestFilterNamesAndToggle(t *testing.T) {
	if FilterMarkdown.String() != "Markdown" || FilterAll.String() != "All" {
		t.Errorf("names = %q, %q", FilterMarkdown, FilterAll)
	}
	if FilterMarkdown.Toggled() != FilterAll || FilterAll.Toggled() != FilterMarkdown {
		t.Error("Toggled does not swap the two filters")
	}
	var zero Filter
	if zero != FilterMarkdown {
		t.Error("the zero Filter must be Markdown, the session default")
	}
}
