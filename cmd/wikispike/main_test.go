package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWritesFile(t *testing.T) {
	// A title with spaces and punctuation, and an output directory that does
	// not exist yet — both are the normal case for this spike (S2, M3.1).
	out := filepath.Join(t.TempDir(), "scratch", "why.md")
	var log strings.Builder

	if err := run([]string{"Why is there anything at all?", out}, &log); err != nil {
		t.Fatalf("run: %v", err)
	}

	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("output file is empty")
	}
	if !strings.Contains(string(got), "Why is there anything at all?") {
		t.Errorf("output does not mention the title:\n%s", got)
	}
	if !strings.Contains(log.String(), out) {
		t.Errorf("operator output does not name the file written: %q", log.String())
	}
}

func TestRunRejectsWrongArgCount(t *testing.T) {
	for name, args := range map[string][]string{
		"none":     {},
		"one":      {"Major Arcana"},
		"three":    {"Major Arcana", "a.md", "b.md"},
		"no title": {"", "a.md"},
	} {
		t.Run(name, func(t *testing.T) {
			var log strings.Builder
			if err := run(args, &log); err == nil {
				t.Fatalf("expected an error for args %q", args)
			}
		})
	}
}
