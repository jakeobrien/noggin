package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// requireGit skips a test when the real git binary isn't on PATH — the same
// "can't verify this here, not a failure" shape requireRipgrep uses
// (files_search_test.go) for its own external-tool dependency.
func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath(gitBinary); err != nil {
		t.Skip("git not on PATH; skipping git-backed status tests")
	}
}

// runGit runs a git subcommand in dir, failing the test on error. Tests use
// it to build a real repository and real working-tree changes under it —
// gitStatus shells out to the real binary, so its own tests exercise the
// real binary too, not a stand-in for its output format.
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command(gitBinary, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=noggin-test", "GIT_AUTHOR_EMAIL=noggin-test@example.com",
		"GIT_COMMITTER_NAME=noggin-test", "GIT_COMMITTER_EMAIL=noggin-test@example.com",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// newGitRepo initializes a repository at t.TempDir() with tracked.txt and
// sub/nested.txt committed, then modifies tracked.txt, adds an untracked
// sub/untracked.txt, and stages a wholly new sub/staged.txt — one fixture
// covering both of S8's marked states plus a couple of shapes it must NOT
// mark (clean files, and — since AM below double-checks "new" wins over
// "modified" for the same path — a staged-then-further-edited file).
func newGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-q")
	if err := os.MkdirAll(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatalf("mkdir sub: %v", err)
	}
	write := func(rel, content string) {
		if err := os.WriteFile(filepath.Join(dir, rel), []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}
	write("tracked.txt", "one\n")
	write("clean.txt", "stays clean\n")
	write("sub/nested.txt", "two\n")
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-q", "-m", "init")

	write("tracked.txt", "one\nmodified\n")    // unstaged modification -> M
	write("sub/untracked.txt", "new\n")        // untracked -> A
	write("sub/addedmodified.txt", "staged\n") // staged add, then edited below
	runGit(t, dir, "add", "sub/addedmodified.txt")
	write("sub/addedmodified.txt", "staged\nplus more\n") // AM -> still A, not M

	return dir
}

func TestGitStatusMarksModifiedAndAddedFiles(t *testing.T) {
	requireGit(t)
	dir := newGitRepo(t)

	got := gitStatus(dir)

	want := map[string]gitFileStatus{
		filepath.Join(dir, "tracked.txt"):              gitStatusModified,
		filepath.Join(dir, "sub", "untracked.txt"):     gitStatusAdded,
		filepath.Join(dir, "sub", "addedmodified.txt"): gitStatusAdded,
	}
	for path, status := range want {
		if got[path] != status {
			t.Errorf("gitStatus[%s] = %q, want %q", path, got[path], status)
		}
	}
	if s, marked := got[filepath.Join(dir, "clean.txt")]; marked && s != gitStatusNone {
		t.Errorf("clean.txt marked %q, want unmarked", s)
	}
	if s, marked := got[filepath.Join(dir, "sub", "nested.txt")]; marked && s != gitStatusNone {
		t.Errorf("sub/nested.txt marked %q, want unmarked", s)
	}
}

// TestGitStatusScopesToRootEvenInsideALargerRepo checks the pathspec
// scoping: a repo with changes both inside and outside a given root only
// reports the ones inside it, which is what lets gitStatus be called once
// per Root rather than once per repository.
func TestGitStatusScopesToRootEvenInsideALargerRepo(t *testing.T) {
	requireGit(t)
	dir := t.TempDir()
	runGit(t, dir, "init", "-q")
	if err := os.MkdirAll(filepath.Join(dir, "inside"), 0o755); err != nil {
		t.Fatalf("mkdir inside: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "inside", "a.txt"), []byte("a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-q", "-m", "init")
	// A change outside "inside" and one inside it.
	if err := os.WriteFile(filepath.Join(dir, "outside.txt"), []byte("new\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "inside", "b.txt"), []byte("new\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := gitStatus(filepath.Join(dir, "inside"))
	if got[filepath.Join(dir, "inside", "b.txt")] != gitStatusAdded {
		t.Errorf("inside/b.txt not marked added: %v", got)
	}
	if _, ok := got[filepath.Join(dir, "outside.txt")]; ok {
		t.Errorf("gitStatus(inside) reported a file outside root: %v", got)
	}
}

// TestGitStatusOutsideARepositoryIsSilent is S8/M4.1's other half: a Root
// with no containing repository returns no markers and (implicitly, since
// gitStatus never returns an error at all) raises no error.
func TestGitStatusOutsideARepositoryIsSilent(t *testing.T) {
	requireGit(t)
	dir := t.TempDir() // never git-init'd
	got := gitStatus(dir)
	if len(got) != 0 {
		t.Errorf("gitStatus outside any repo = %v, want empty", got)
	}
}

func TestGitStatusMissingGitBinaryIsSilent(t *testing.T) {
	// Point PATH at nothing so exec.LookPath(gitBinary) fails, regardless of
	// whether the real git happens to be installed on this machine.
	t.Setenv("PATH", t.TempDir())
	got := gitStatus(t.TempDir())
	if got != nil {
		t.Errorf("gitStatus with no git on PATH = %v, want nil", got)
	}
}

// TestParseGitStatusPriorityAndIgnoredCodes exercises the parser directly
// against synthetic porcelain -z output, independent of a real git
// invocation, so the M/A/none decision table is pinned down explicitly.
func TestParseGitStatusPriorityAndIgnoredCodes(t *testing.T) {
	entries := []string{
		" M modified.txt",   // unstaged modification -> M
		"M  staged-mod.txt", // staged modification -> M
		"?? untracked.txt",  // untracked -> A
		"A  added.txt",      // staged add -> A
		"AM addedmod.txt",   // staged add + further edit -> A, not M
		" D deleted.txt",    // deleted -> unmarked
		"D  staged-del.txt", // staged deletion -> unmarked
	}
	out := []byte(joinNUL(entries))

	got := parseGitStatus("/repo", out)
	want := map[string]gitFileStatus{
		"/repo/modified.txt":   gitStatusModified,
		"/repo/staged-mod.txt": gitStatusModified,
		"/repo/untracked.txt":  gitStatusAdded,
		"/repo/added.txt":      gitStatusAdded,
		"/repo/addedmod.txt":   gitStatusAdded,
	}
	for path, status := range want {
		if got[path] != status {
			t.Errorf("parseGitStatus[%s] = %q, want %q", path, got[path], status)
		}
	}
	for _, unmarked := range []string{"/repo/deleted.txt", "/repo/staged-del.txt"} {
		if s, ok := got[unmarked]; ok {
			t.Errorf("%s marked %q, want absent", unmarked, s)
		}
	}
}

func joinNUL(entries []string) string {
	s := ""
	for _, e := range entries {
		s += e + "\x00"
	}
	return s
}

// TestFilesNavRendersGitStatusMarkers checks S8's rendering contract at the
// pane level: a file's marker appears in its row, a clean file's doesn't,
// and a directory is never marked even if it happens to be a key in status
// (aggregation is out of scope; the render must not do it by accident).
func TestFilesNavRendersGitStatusMarkers(t *testing.T) {
	root := fixtureTree()
	f := newFiles(root).withStatus(map[string]gitFileStatus{
		"/root/README.md": gitStatusModified,
		"/root/notes.txt": gitStatusAdded,
		"/root/docs":      gitStatusModified, // a dir; must be ignored by render
	}).setSize(30, 6)

	view := f.view()
	lines := strings.Split(view, "\n")
	readmeLine := lineContaining(lines, "README.md")
	notesLine := lineContaining(lines, "notes.txt")
	docsLine := lineContaining(lines, "docs")

	// Three spaces, not one, between the marker letter and the name: the
	// status column's own trailing space ("M "/"A ") plus C4/M2.3's star
	// marker column ("  " unstarred) sitting between it and the name.
	if !strings.Contains(readmeLine, "M   README.md") {
		t.Errorf("README.md row = %q, want an M marker before the name", readmeLine)
	}
	if !strings.Contains(notesLine, "A   notes.txt") {
		t.Errorf("notes.txt row = %q, want an A marker before the name", notesLine)
	}
	if strings.Contains(docsLine, "M ") {
		t.Errorf("docs (a directory) row = %q, want no git status marker", docsLine)
	}
}

func lineContaining(lines []string, substr string) string {
	for _, l := range lines {
		if strings.Contains(l, substr) {
			return l
		}
	}
	return ""
}

// TestActivateRootShowsGitStatusMarkersAtLaunchAndOnSwitch is M4.1's
// acceptance criterion end-to-end, through the real model path rather than
// filesNav directly: a Root inside a git repo shows M/A markers matching its
// actual working-tree status once activated, and a Root outside any
// repository shows none, silently — no footer message, session still runs.
func TestActivateRootShowsGitStatusMarkersAtLaunchAndOnSwitch(t *testing.T) {
	requireGit(t)
	repo := newGitRepo(t)                                  // has tracked.txt (M) and sub/untracked.txt (A)
	plain := mkdir(t, filepath.Join(t.TempDir(), "plain")) // no .git at all
	writeFile(t, filepath.Join(plain, "note.md"))

	m := sized(t, newModel(), 80, 24)
	m.roots = m.roots.setRoots([]root{{path: repo}, {path: plain}}, -1)

	m = m.activateRoot() // selection starts on row 0 (repo)
	view := m.files.view()
	if !strings.Contains(lineContaining(strings.Split(view, "\n"), "tracked.txt"), "M   tracked.txt") {
		t.Errorf("repo Files Nav view = %q, want an M marker on tracked.txt", view)
	}

	m.roots.selected = 1
	m = m.activateRoot() // switch to the non-repo Root
	if m.message != "" {
		t.Errorf("activating a non-repo Root produced a message %q, want silence", m.message)
	}
	view = m.files.view()
	if strings.Contains(view, "M ") || strings.Contains(view, "A ") {
		t.Errorf("non-repo Files Nav view = %q, want no git status markers", view)
	}
	if !strings.Contains(view, "note.md") {
		t.Errorf("non-repo Files Nav view = %q, want note.md still listed", view)
	}
}
