package main

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestFrontmatterFields(t *testing.T) {
	for _, tc := range []struct {
		name    string
		in      string
		want    []frontmatterField
		wantOK  bool
		comment string
	}{
		{name: "source order is kept", in: "zeta: 1\nalpha: two\n", wantOK: true,
			want: []frontmatterField{{"zeta", "1"}, {"alpha", "two"}}},
		{name: "quotes are YAML syntax, not value", in: "title: \"Hi: there\"\n", wantOK: true,
			want: []frontmatterField{{"title", "Hi: there"}}},
		{name: "scalar list is comma-joined", in: "tags: [a, b]\nmore:\n  - c\n  - d\n", wantOK: true,
			want: []frontmatterField{{"tags", "a, b"}, {"more", "c, d"}}},
		{name: "empty list and null are blank", in: "tags: []\ndate:\n", wantOK: true,
			want: []frontmatterField{{"tags", ""}, {"date", ""}}},
		{name: "block scalar keeps its lines", in: "note: |\n  one\n  two\n", wantOK: true,
			want: []frontmatterField{{"note", "one\ntwo"}}},
		// Anything the card cannot show faithfully falls back to the code block.
		{name: "nested map", in: "author:\n  name: x\n"},
		{name: "list of maps", in: "links:\n  - url: x\n"},
		{name: "nested list", in: "grid: [[1, 2]]\n"},
		{name: "alias", in: "a: &x 1\nb: *x\n"},
		{name: "not a mapping", in: "- a\n- b\n"},
		{name: "comment only", in: "# nothing here\n"},
		{name: "invalid yaml", in: "title: [unclosed\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := frontmatterFields([]byte(tc.in))
			if ok != tc.wantOK {
				t.Fatalf("frontmatterFields(%q) ok = %v, want %v", tc.in, ok, tc.wantOK)
			}
			if ok && !reflect.DeepEqual(got, tc.want) {
				t.Errorf("frontmatterFields(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestFrontmatterCardFitsItsWidth(t *testing.T) {
	fields := []frontmatterField{
		{"title", "Fixture Document"},
		{"description", strings.Repeat("a long description ", 8)},
		{"note", "one\ntwo"},
	}
	for _, width := range []int{80, 54, 30} {
		for _, muted := range []bool{false, true} {
			card, ok := frontmatterCard(fields, width, muted)
			if !ok {
				t.Fatalf("width %d: no card", width)
			}
			plain := ansi.Strip(card)
			for i, line := range strings.Split(card, "\n") {
				if got := lipgloss.Width(line); got > width {
					t.Errorf("width %d muted %v: line %d is %d cells: %q", width, muted, i, got, ansi.Strip(line))
				}
			}
			// Words, not phrases: a narrow card wraps values across lines.
			words := strings.Fields(plain)
			for _, want := range []string{"title", "Fixture", "Document", "description", "note", "one", "two"} {
				if !slices.Contains(words, want) {
					t.Errorf("width %d: card is missing %q\n%s", width, want, plain)
				}
			}
		}
	}
}

func TestFrontmatterCardShrinksToItsContent(t *testing.T) {
	card, _ := frontmatterCard([]frontmatterField{{"a", "b"}}, 80, false)
	if got := lipgloss.Width(card); got > 20 {
		t.Errorf("a one-field card is %d cells wide, want it to fit its content\n%s", got, ansi.Strip(card))
	}
}

func TestFrontmatterCardDeclinesATinyWidth(t *testing.T) {
	if _, ok := frontmatterCard([]frontmatterField{{"title", "x"}}, 10, false); ok {
		t.Error("got a card at width 10, want the code-block fallback")
	}
}

// TestUncardableFrontmatterRendersAsCodeBlock: the card's fallback is the
// verbatim code block (D138), so nothing in the frontmatter is ever lost.
func TestUncardableFrontmatterRendersAsCodeBlock(t *testing.T) {
	for name, src := range map[string]string{
		"nested map": "---\nauthor:\n  name: Ada\n---\n# H\n",
		"toml":       "+++\ntitle = 'Ada'\n+++\n# H\n",
	} {
		t.Run(name, func(t *testing.T) {
			out, err := renderMarkdown([]byte(src), 60)
			if err != nil {
				t.Fatal(err)
			}
			raw := strings.Split(src, "\n")[1:3]
			for _, line := range raw {
				if line != "+++" && line != "---" && !strings.Contains(out, line) {
					t.Errorf("rendered output is missing the verbatim line %q\n%s", line, out)
				}
			}
			if strings.Contains(out, "╭") {
				t.Errorf("rendered a card for frontmatter it cannot show faithfully\n%s", out)
			}
		})
	}
}

// TestGrepLandingAcrossTheCard: the card reshapes frontmatter text ("title:"
// becomes "title"), so resolveGrepLanding maps the frontmatter and the body
// as separate segments. A raw frontmatter match the card does not reproduce
// must not shift every body match by one.
func TestGrepLandingAcrossTheCard(t *testing.T) {
	const width = 60
	src := []byte("---\ntitle: report\nstatus: draft\n---\n# doc\n\nthe title: report line\n")
	rendered, err := renderMarkdown(src, width)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, query string
		line        int
		want        int
		wantOK      bool
	}{
		{"value in the card", "report", 2, 0, true},
		{"body after the card", "report", 7, 1, true},
		// "title:" is raw frontmatter syntax; the card shows "title".
		{"body match past a frontmatter-only match", "title:", 7, 0, true},
		{"frontmatter match the card does not show", "title:", 2, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := resolveGrepLanding(src, rendered, tc.query, tc.line, width)
			if ok != tc.wantOK || (ok && got != tc.want) {
				t.Errorf("resolveGrepLanding(%q, line %d) = %d, %v; want %d, %v", tc.query, tc.line, got, ok, tc.want, tc.wantOK)
			}
		})
	}
}
