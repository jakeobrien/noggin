package tree

import "testing"

// TestIsReadable locks the readable-file policy (S6, GLOSSARY.md "Readable
// file"): extensions and the conventional extensionless names both match
// case-insensitively (D23), and anything else is not readable.
func TestIsReadable(t *testing.T) {
	cases := []struct {
		name string
		want bool
	}{
		{"README", true},
		{"Readme", true},
		{"readme", true},
		{"LICENSE", true},
		{"License", true},
		{"CHANGELOG", true},
		{"changelog", true},
		{"notes.md", true},
		{"notes.MD", true},
		{"notes.Md", true},
		{"guide.markdown", true},
		{"guide.MARKDOWN", true},
		{"plain.txt", true},
		{"plain.TXT", true},
		{"main.go", false},
		{"image.png", false},
		{"README.bak", false},
		{"NOTREADME", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isReadable(c.name); got != c.want {
				t.Errorf("isReadable(%q) = %v, want %v", c.name, got, c.want)
			}
		})
	}
}
