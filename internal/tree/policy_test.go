package tree

import "testing"

// TestIsMarkdown locks the Markdown classification (D140, GLOSSARY.md
// "Markdown file"): extensions and the conventional extensionless names both
// match case-insensitively (D23); `.txt` and `LICENSE` are Text, not Markdown.
func TestIsMarkdown(t *testing.T) {
	cases := []struct {
		name string
		want bool
	}{
		{"README", true},
		{"Readme", true},
		{"readme", true},
		{"CHANGELOG", true},
		{"changelog", true},
		{"notes.md", true},
		{"notes.MD", true},
		{"notes.Md", true},
		{"guide.markdown", true},
		{"guide.MARKDOWN", true},
		{"guide.mdown", true},
		{"guide.MKD", true},
		{"guide.mkd", true},
		{"plain.txt", false},
		{"plain.TXT", false},
		{"LICENSE", false},
		{"License", false},
		{"main.go", false},
		{"image.png", false},
		{"README.bak", false},
		{"NOTREADME", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := IsMarkdown(c.name); got != c.want {
				t.Errorf("IsMarkdown(%q) = %v, want %v", c.name, got, c.want)
			}
		})
	}
}
