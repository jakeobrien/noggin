package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// rootPaths is the path of each resolved root, in order.
func rootPaths(res resolution) []string {
	paths := make([]string, len(res.roots))
	for i, r := range res.roots {
		paths[i] = r.path
	}
	return paths
}

// mkdir makes dir under a fresh temp base and returns its cleaned path.
func mkdir(t *testing.T, path string) string {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	return path
}

// writeFile creates a file with some content and returns its path.
func writeFile(t *testing.T, path string) string {
	t.Helper()
	if err := os.WriteFile(path, []byte("# note\n"), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return path
}

// TestResolveRoots walks the S1 launch contract: every argument shape it
// accepts or refuses, always yielding a session that still starts.
func TestResolveRoots(t *testing.T) {
	tests := []struct {
		name  string
		build func(t *testing.T) (args, starred []string, cwd string)
		check func(t *testing.T, res resolution)
	}{
		{
			name: "bare launch resolves starred roots, first active",
			build: func(t *testing.T) ([]string, []string, string) {
				base := t.TempDir()
				a := mkdir(t, filepath.Join(base, "a"))
				b := mkdir(t, filepath.Join(base, "b"))
				return nil, []string{a, b}, base
			},
			check: func(t *testing.T, res resolution) {
				wantLen(t, res, 2)
				wantActiveStarred(t, res, 0)
				wantNoDoc(t, res)
				wantClean(t, res)
			},
		},
		{
			name: "bare launch with no starred roots is a first run",
			build: func(t *testing.T) ([]string, []string, string) {
				return nil, nil, t.TempDir()
			},
			check: func(t *testing.T, res resolution) {
				wantLen(t, res, 0)
				if res.active != -1 {
					t.Errorf("active = %d, want -1 (no roots)", res.active)
				}
				wantNoDoc(t, res)
				wantClean(t, res)
			},
		},
		{
			name: "dir argument becomes an ad-hoc root, listed first and active",
			build: func(t *testing.T) ([]string, []string, string) {
				base := t.TempDir()
				proj := mkdir(t, filepath.Join(base, "proj"))
				star := mkdir(t, filepath.Join(base, "star"))
				return []string{proj}, []string{star}, base
			},
			check: func(t *testing.T, res resolution) {
				wantLen(t, res, 2)
				if res.active != 0 || res.roots[0].starred {
					t.Errorf("active = %d starred=%v, want the ad-hoc root at 0", res.active, res.roots[0].starred)
				}
				wantNoDoc(t, res)
				wantClean(t, res)
			},
		},
		{
			name: "file argument roots at cwd and opens the file",
			build: func(t *testing.T) ([]string, []string, string) {
				cwd := t.TempDir()
				writeFile(t, filepath.Join(cwd, "note.md"))
				return []string{"note.md"}, nil, cwd // relative, joined against cwd
			},
			check: func(t *testing.T, res resolution) {
				wantLen(t, res, 1)
				if res.active != 0 || res.roots[0].starred {
					t.Errorf("want the cwd as the ad-hoc active root, got %+v", res.roots)
				}
				if got, want := res.docPath, filepath.Join(res.roots[0].path, "note.md"); got != want {
					t.Errorf("docPath = %q, want %q", got, want)
				}
				wantClean(t, res)
			},
		},
		{
			name: "dir already a starred root is not duplicated and becomes active",
			build: func(t *testing.T) ([]string, []string, string) {
				base := t.TempDir()
				a := mkdir(t, filepath.Join(base, "a"))
				b := mkdir(t, filepath.Join(base, "b"))
				return []string{b}, []string{a, b}, base
			},
			check: func(t *testing.T, res resolution) {
				wantLen(t, res, 2) // no duplicate added
				wantActiveStarred(t, res, 1)
				wantNoDoc(t, res)
				wantClean(t, res)
			},
		},
		{
			name: "non-existent path is rejected but the session still starts",
			build: func(t *testing.T) ([]string, []string, string) {
				base := t.TempDir()
				star := mkdir(t, filepath.Join(base, "star"))
				return []string{filepath.Join(base, "nope")}, []string{star}, base
			},
			check: func(t *testing.T, res resolution) {
				wantLen(t, res, 1)
				wantActiveStarred(t, res, 0)
				wantNoDoc(t, res)
				wantRejected(t, res)
			},
		},
		{
			name: "subdirectory of a starred root activates the covering root",
			build: func(t *testing.T) ([]string, []string, string) {
				base := t.TempDir()
				star := mkdir(t, filepath.Join(base, "star"))
				sub := mkdir(t, filepath.Join(star, "sub"))
				return []string{sub}, []string{star}, base
			},
			check: func(t *testing.T, res resolution) {
				wantLen(t, res, 1) // no ad-hoc added
				wantActiveStarred(t, res, 0)
				wantNoDoc(t, res)
				wantRejected(t, res)
			},
		},
		{
			name: "parent of a starred root is rejected, session still starts",
			build: func(t *testing.T) ([]string, []string, string) {
				base := t.TempDir()
				parent := mkdir(t, filepath.Join(base, "parent"))
				// The starred root lives inside the argument directory; it need
				// not exist on disk (S2 keeps missing paths).
				child := filepath.Join(parent, "child")
				return []string{parent}, []string{child}, base
			},
			check: func(t *testing.T, res resolution) {
				wantLen(t, res, 1) // still just the starred root, no ad-hoc
				wantActiveStarred(t, res, 0)
				wantNoDoc(t, res)
				wantRejected(t, res)
			},
		},
		{
			name: "more than one argument is rejected, session still starts",
			build: func(t *testing.T) ([]string, []string, string) {
				base := t.TempDir()
				star := mkdir(t, filepath.Join(base, "star"))
				one := mkdir(t, filepath.Join(base, "one"))
				two := mkdir(t, filepath.Join(base, "two"))
				return []string{one, two}, []string{star}, base
			},
			check: func(t *testing.T, res resolution) {
				wantLen(t, res, 1)
				wantActiveStarred(t, res, 0)
				wantNoDoc(t, res)
				wantRejected(t, res)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args, starred, cwd := tt.build(t)
			res := resolveRoots(args, starred, cwd)
			tt.check(t, res)
		})
	}
}

func wantLen(t *testing.T, res resolution, n int) {
	t.Helper()
	if len(res.roots) != n {
		t.Fatalf("roots = %v, want %d root(s)", rootPaths(res), n)
	}
}

// TestTooManyArgumentsMessageDoesNotImplyFirstWasHonored is T17: the
// rejection must not read as if args[0] was used and only the rest were
// dropped, since none of them become a root.
func TestTooManyArgumentsMessageDoesNotImplyFirstWasHonored(t *testing.T) {
	base := t.TempDir()
	one := mkdir(t, filepath.Join(base, "one"))
	two := mkdir(t, filepath.Join(base, "two"))

	res := resolveRoots([]string{one, two}, nil, base)

	if strings.Contains(res.message, "ignored") && strings.Contains(res.message, "extra") {
		t.Errorf("message = %q, still implies args[0] was honored and only extras were dropped", res.message)
	}
	if res.message == "" {
		t.Fatal("want a rejection message")
	}
	if len(res.roots) != 0 {
		t.Errorf("roots = %v, want none (no starred roots, no args honored)", rootPaths(res))
	}
}

// TestResolveCwdFallsBackToPWD is T20: when os.Getwd fails (e.g. the working
// directory was removed out from under the process), resolveCwd falls back to
// the shell-set PWD rather than leaving the caller to feed a bare "." into
// root resolution.
func TestResolveCwdFallsBackToPWD(t *testing.T) {
	removed := mkdir(t, filepath.Join(t.TempDir(), "gone"))
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	if err := os.Chdir(removed); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	if err := os.RemoveAll(removed); err != nil {
		t.Fatalf("RemoveAll: %v", err)
	}
	if _, err := os.Getwd(); err == nil {
		t.Skip("os.Getwd did not fail after its directory was removed on this platform")
	}

	fallback := filepath.Join(t.TempDir(), "pwd-fallback")
	t.Setenv("PWD", fallback)

	cwd, err := resolveCwd()
	if err != nil {
		t.Fatalf("resolveCwd: %v", err)
	}
	if cwd != fallback {
		t.Errorf("cwd = %q, want the PWD fallback %q", cwd, fallback)
	}
}

// TestResolveLaunchDropsArgOnUnresolvableCwd is T20's main.go-facing half:
// when neither os.Getwd nor PWD yields a usable directory, a relative launch
// argument must not be silently resolved against "." — it is dropped and the
// session starts on the starred roots alone with a message explaining why.
func TestResolveLaunchDropsArgOnUnresolvableCwd(t *testing.T) {
	base := t.TempDir()
	star := mkdir(t, filepath.Join(base, "star"))
	cwdErr := errors.New("getwd: no such file or directory")

	res := resolveLaunch([]string{"note.md"}, []string{star}, "", cwdErr)

	wantLen(t, res, 1)
	wantActiveStarred(t, res, 0)
	wantNoDoc(t, res)
	if res.message == "" {
		t.Error("want a message explaining the dropped argument")
	}
}

// wantActiveStarred asserts the active root is the starred one at index i.
func wantActiveStarred(t *testing.T, res resolution, i int) {
	t.Helper()
	if res.active != i {
		t.Errorf("active = %d, want %d", res.active, i)
		return
	}
	if !res.roots[i].starred {
		t.Errorf("active root %s is not starred", res.roots[i].path)
	}
}

func wantNoDoc(t *testing.T, res resolution) {
	t.Helper()
	if res.docPath != "" {
		t.Errorf("docPath = %q, want no document", res.docPath)
	}
}

func wantClean(t *testing.T, res resolution) {
	t.Helper()
	if res.message != "" {
		t.Errorf("message = %q, want a clean launch", res.message)
	}
}

func wantRejected(t *testing.T, res resolution) {
	t.Helper()
	if res.message == "" {
		t.Error("want a rejection message, got none")
	}
}
