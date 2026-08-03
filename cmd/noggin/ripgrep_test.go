package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDetectRipgrepAbsent points PATH at a directory with no rg on it: the
// session must report unavailable with a reason, never an error or a panic.
func TestDetectRipgrepAbsent(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	got := detectRipgrep()

	if got.available {
		t.Fatalf("detectRipgrep() = %+v, want available=false", got)
	}
	if got.reason == "" {
		t.Fatalf("detectRipgrep() = %+v, want a non-empty reason", got)
	}
}

// TestDetectRipgrepPresent points PATH at a directory containing a stand-in
// rg executable: detection only needs the binary to resolve, not to run, so
// a fake, empty, executable file under that name is enough to prove the
// present case rather than depending on the sandbox's real ripgrep install.
func TestDetectRipgrepPresent(t *testing.T) {
	dir := t.TempDir()
	fake := filepath.Join(dir, ripgrepBinary)
	if err := os.WriteFile(fake, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write fake rg: %v", err)
	}
	t.Setenv("PATH", dir)

	got := detectRipgrep()

	if !got.available {
		t.Fatalf("detectRipgrep() = %+v, want available=true", got)
	}
	if got.reason != "" {
		t.Fatalf("detectRipgrep() = %+v, want an empty reason", got)
	}
}
