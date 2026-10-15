package tree

import (
	"reflect"
	"slices"
	"testing"
)

// D141: Options.HideDotfiles and Options.ShowIgnored.

func optionsFixture(t *testing.T) string {
	return writeTree(t, map[string]string{
		".gitignore":                 "ignored.md\nbuild/\nnode_modules/\n",
		"README.md":                  "# readme",
		".hidden.md":                 "# hidden",
		".notes/inner.md":            "# inner",
		"ignored.md":                 "gitignored",
		"build/out.md":               "generated",
		"build/deep/more.md":         "generated",
		"node_modules/pkg/README.md": "dependency",
		"target/notes.md":            "not ignored, but a heavy dir name",
		".git/HEAD":                  "ref",
	})
}

func TestWalkDefaultOptionsKeepDotfilesAndDropIgnored(t *testing.T) {
	node, err := Walk(optionsFixture(t), Options{})
	if err != nil {
		t.Fatal(err)
	}
	files, _ := collect(node, "")
	want := []string{".hidden.md", ".notes/inner.md", "README.md", "target/notes.md"}
	if !reflect.DeepEqual(files, want) {
		t.Errorf("files = %v, want %v", files, want)
	}
}

func TestWalkHideDotfiles(t *testing.T) {
	node, err := Walk(optionsFixture(t), Options{HideDotfiles: true})
	if err != nil {
		t.Fatal(err)
	}
	files, dirs := collect(node, "")
	if want := []string{"README.md", "target/notes.md"}; !reflect.DeepEqual(files, want) {
		t.Errorf("files = %v, want %v", files, want)
	}
	if slices.Contains(dirs, ".notes") {
		t.Errorf("dirs = %v, want no dot-directory", dirs)
	}
}

func TestWalkShowIgnoredMarksAndExcludesHeavyDirs(t *testing.T) {
	node, err := Walk(optionsFixture(t), Options{ShowIgnored: true})
	if err != nil {
		t.Fatal(err)
	}
	files, dirs := collect(node, "")
	want := []string{".hidden.md", ".notes/inner.md", "README.md", "build/deep/more.md", "build/out.md", "ignored.md"}
	if !reflect.DeepEqual(files, want) {
		t.Errorf("files = %v, want %v", files, want)
	}
	for _, d := range dirs {
		if d == "node_modules" || d == "target" || d == ".git" {
			t.Errorf("dirs = %v, want %s excluded", dirs, d)
		}
	}

	ignoredByName := map[string]bool{}
	var mark func(n *Node, base string)
	mark = func(n *Node, base string) {
		for _, c := range n.Children {
			rel := c.Name
			if base != "" {
				rel = base + "/" + c.Name
			}
			ignoredByName[rel] = c.Ignored
			mark(c, rel)
		}
	}
	mark(node, "")
	for rel, want := range map[string]bool{
		"ignored.md":         true,
		"build":              true,
		"build/out.md":       true,
		"build/deep":         true,
		"build/deep/more.md": true,
		"README.md":          false,
		".hidden.md":         false,
	} {
		if got := ignoredByName[rel]; got != want {
			t.Errorf("%s Ignored = %v, want %v", rel, got, want)
		}
	}
}

func TestListDirMatchesWalkUnderEveryOption(t *testing.T) {
	root := optionsFixture(t)
	for _, opts := range []Options{{}, {HideDotfiles: true}, {ShowIgnored: true}, {ShowIgnored: true, HideDotfiles: true, Filter: FilterAll}} {
		walked, err := Walk(root, opts)
		if err != nil {
			t.Fatal(err)
		}
		var check func(n *Node)
		check = func(n *Node) {
			got, err := ListDir(walked.Path, n.Path, opts)
			if err != nil {
				t.Fatalf("%+v: ListDir(%s): %v", opts, n.Path, err)
			}
			if want := entriesOf(n); !reflect.DeepEqual(append([]Entry{}, got...), want) {
				t.Errorf("%+v: ListDir(%s) = %v, want Walk's %v", opts, n.Path, got, want)
			}
			for _, c := range n.Children {
				if c.IsDir {
					check(c)
				}
			}
		}
		check(walked)
	}
}
