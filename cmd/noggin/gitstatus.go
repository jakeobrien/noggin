package main

import (
	"os/exec"
	"path/filepath"
	"strings"
)

// gitFileStatus is one of S8's two distinguished per-file git states:
// modified or new. The zero value, gitStatusNone, covers everything else —
// clean, staged vs. unstaged, deleted, renamed — none of which S8 marks.
type gitFileStatus byte

const (
	gitStatusNone     gitFileStatus = 0
	gitStatusModified gitFileStatus = 'M'
	gitStatusAdded    gitFileStatus = 'A'
)

// gitBinary is the executable gitStatus shells out to, detected by name on
// PATH the same way ripgrepBinary is (ripgrep.go) — the same portability
// tradeoff, applied to a second external tool.
const gitBinary = "git"

// gitStatus resolves S8's per-file markers for every changed file under
// root, keyed by absolute path so a caller can look one up by
// tree.Node.Path directly. It degrades to a nil map — never an error and
// never a footer message — when git is not on PATH, root is not inside a
// git repository, or the command otherwise fails: S8 requires exactly this
// "no markers, silently" behaviour for a Root outside any repo, and reusing
// it for every other failure mode keeps the shape identical to S14's
// ripgrep detection (detectRipgrep, ripgrep.go) — detect an external tool,
// degrade quietly, never block the session over it.
//
// Status is resolved relative to the containing repository's root, not
// root itself (S8): `git status --porcelain` always reports paths relative
// to the repository toplevel regardless of the invocation directory, so a
// first call resolves that toplevel (rev-parse --show-toplevel, which also
// doubles as the "is root inside a repository at all" check) and a second
// asks for status scoped to root via a pathspec, joining the two to
// reconstruct the absolute path each entry names.
//
// --no-renames is deliberate: S8 does not mark renamed files, and turning
// off rename detection is what keeps every changed path in the output on
// its own NUL-separated entry — a detected rename instead emits a
// delete-then-add pair sharing one entry's fields, which would need a
// second parsing shape purely to throw the result away.
func gitStatus(root string) map[string]gitFileStatus {
	if _, err := exec.LookPath(gitBinary); err != nil {
		return nil
	}

	topOut, err := exec.Command(gitBinary, "-C", root, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return nil // not inside a git repository (or git otherwise failed)
	}
	toplevel := strings.TrimRight(string(topOut), "\n")

	out, err := exec.Command(gitBinary, "-C", root, "status",
		"--porcelain=v1", "--untracked-files=all", "--no-renames", "-z", "--", ".").Output()
	if err != nil {
		return nil
	}
	return parseGitStatus(toplevel, out)
}

// parseGitStatus turns one `git status --porcelain=v1 -z` invocation's
// NUL-separated entries into S8's two-state map, keyed by absolute path
// under toplevel. Each entry is exactly "XY PATH": a fixed two-character
// status, a space, then the path relative to toplevel. X is the index
// (staged) state and Y the worktree (unstaged) state; S8 does not
// distinguish them, so either position reporting '?' (untracked) or 'M'
// (modified) is enough, and a staged add (X == 'A') wins over a same-entry
// worktree modification ("AM") since the file is still new, not merely
// changed. Anything else — deleted, permission/type changes, conflicts —
// is left unmarked.
func parseGitStatus(toplevel string, out []byte) map[string]gitFileStatus {
	result := map[string]gitFileStatus{}
	for _, entry := range strings.Split(string(out), "\x00") {
		if len(entry) < 4 {
			continue // trailing empty split, or a malformed entry
		}
		x, y, relPath := entry[0], entry[1], entry[3:]

		var status gitFileStatus
		switch {
		case x == '?' || y == '?':
			status = gitStatusAdded
		case x == 'A':
			status = gitStatusAdded
		case x == 'M' || y == 'M':
			status = gitStatusModified
		default:
			continue
		}
		result[filepath.Join(toplevel, relPath)] = status
	}
	return result
}
