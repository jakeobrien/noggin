package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// root is one of the session's note sources: an absolute path plus whether it
// is starred. A launch argument produces an unstarred ad-hoc root; roots loaded
// from the store (M3.2) are starred. M3.3's Roots Nav renders these.
type root struct {
	path    string
	starred bool
}

// resolution is the outcome of applying the S1 launch contract to a command
// line: the session's roots (ad-hoc first, then starred), which one is active,
// an optional document to open, and a rejection message when the argument was
// refused. A rejection never stops the session — resolution always yields a
// runnable set of roots ("silence is not an error, but confusion is").
type resolution struct {
	roots   []root
	active  int    // index into roots, or -1 when there are none
	docPath string // absolute path of the document to open, or ""
	message string // rejection notice, or "" when the launch was clean
}

// activeRootPath is the path of the active root, or "" when none is active.
func (r resolution) activeRootPath() string {
	if r.active < 0 || r.active >= len(r.roots) {
		return ""
	}
	return r.roots[r.active].path
}

// resolveCwd finds the session's working directory for anchoring a relative
// launch argument. os.Getwd fails when the directory has been removed out
// from under the process (rare, but real on a long-lived shell) or the
// process lacks permission to stat it; falling back to the shell-set PWD
// environment variable, when it is itself absolute, still gives a usable
// anchor. When neither works there is no safe absolute path to fall back to
// (T20): the caller must not feed a bare "." into root resolution, since
// that would silently misresolve any relative argument against the actual
// process directory rather than reporting the failure.
func resolveCwd() (string, error) {
	if cwd, err := os.Getwd(); err == nil {
		return cwd, nil
	} else if pwd := os.Getenv("PWD"); filepath.IsAbs(pwd) {
		return pwd, nil
	} else {
		return "", err
	}
}

// resolveLaunch is resolveRoots' main.go-facing wrapper: it folds a failed
// working-directory lookup (cwdErr, from resolveCwd) into the launch outcome
// instead of leaving that decision to the caller. Without a safe cwd, an
// argument that actually depends on it is dropped rather than resolved
// against a bare "." (T20), with a message saying so.
//
// "Actually depends on it" is needsCwd's job (T28): a cwd failure must not
// drop an argument that does not need cwd at all — the too-many-arguments
// case is refused by resolveRoots on its own count regardless of cwd, and
// an absolute directory argument never touches cwd once resolved either.
// When needsCwd says no, args passes through to resolveRoots, but via
// resolvePathArg with needsCwd's own stat result (when it took one) rather
// than letting resolveRoots stat the same path a second time — the TOCTOU
// window a second, independent stat would open (T37).
func resolveLaunch(args []string, starred []string, cwd string, cwdErr error) resolution {
	if cwdErr == nil {
		return resolveRoots(args, starred, cwd)
	}
	needs, info, statErr, stated := needsCwd(args)
	if needs {
		res := resolveRoots(nil, starred, cwd)
		res.message = fmt.Sprintf("cannot determine working directory (%v); path argument ignored", cwdErr)
		return res
	}
	if stated {
		p := absPath(args[0], cwd)
		return resolvePathArg(args[0], p, info, statErr, cwd, baseResolution(starred))
	}
	return resolveRoots(args, starred, cwd)
}

// needsCwd reports whether resolving args genuinely depends on cwd being
// correct, so resolveLaunch knows when a resolveCwd failure actually needs
// to drop the argument (T28) rather than doing so on every failure
// regardless of relevance:
//   - zero arguments never touch cwd; more than one is refused by
//     resolveRoots on its count alone (T17), before cwd would ever be used;
//   - a relative single argument cannot be resolved to an absolute path at
//     all without cwd to anchor it (absPath, T20);
//   - an absolute single argument that turns out to be a directory never
//     touches cwd again once resolved (resolveDir does not take it);
//   - but an absolute single argument that turns out to be a file still
//     does: resolveFile always makes cwd the ad-hoc root, regardless of
//     whether the file argument itself needed cwd to be found.
//
// Only the last two cases (a single absolute argument) need to look at the
// argument itself; none needs cwd's own value to decide, so cwd is not a
// parameter here (T38) — resolveLaunch still has it in scope to hand to
// resolvePathArg afterwards.
//
// When the single-absolute-argument case takes a stat to decide, it returns
// that stat result (info, statErr, stated=true) so resolveLaunch can hand it
// to resolvePathArg instead of letting resolveRoots stat the same path again
// moments later (T37 — an independent second stat is a TOCTOU window, not
// just wasted work). stated is false whenever no stat was taken (zero or
// multiple args, or a relative single argument), in which case there is
// nothing to reuse and resolveLaunch falls through to resolveRoots as before.
func needsCwd(args []string) (needs bool, info os.FileInfo, statErr error, stated bool) {
	if len(args) != 1 {
		return false, nil, nil, false
	}
	if !filepath.IsAbs(args[0]) {
		return true, nil, nil, false
	}
	info, statErr = os.Stat(args[0])
	if statErr != nil {
		return false, nil, statErr, true // resolveRoots' own not-found rejection doesn't need cwd either
	}
	return !info.IsDir(), info, nil, true
}

// resolveRoots applies the S1 launch contract. args is the command line beyond
// the program name; starred is the stored starred-root paths (M3.2 supplies
// them, empty until then); cwd anchors a relative argument. It never touches
// the filesystem beyond a stat of the single argument, so the resulting active
// root is a path for the caller to walk — the side-effecting steps (walking a
// tree, reading a document) stay out of this pure decision.
func resolveRoots(args []string, starred []string, cwd string) resolution {
	base := baseResolution(starred)

	switch {
	case len(args) == 0:
		return base
	case len(args) > 1:
		// More than one path: refused outright, not just trimmed to the first —
		// none of them become a root (T17), the session still starts on the
		// starred roots alone.
		base.message = fmt.Sprintf("noggin takes a single path argument; got %d, none used", len(args))
		return base
	}

	p := absPath(args[0], cwd)
	info, err := os.Stat(p)
	return resolvePathArg(args[0], p, info, err, cwd, base)
}

// baseResolution is the resolution every path-argument branch starts from:
// the starred roots, with the first one active when there are any (-1
// otherwise). "-1 when there are none to be active; otherwise the first" is
// safe because S1's order puts the ad-hoc root, if any, ahead of these
// anyway, so this only matters for the bare-launch case where roots is
// starred roots alone. Shared by resolveRoots and resolveLaunch's
// already-stat'd single-argument path (T37), so both build the same base.
func baseResolution(starred []string) resolution {
	roots := starredRoots(starred)
	active := -1
	if len(roots) > 0 {
		active = 0
	}
	return resolution{roots: roots, active: active}
}

// resolvePathArg finishes resolving a single path argument once its stat
// result (info, err) is already known — the not-found/directory/file
// branching resolveRoots itself performs right after its own stat, and that
// resolveLaunch also drives with needsCwd's stat instead of taking a second,
// independent one on the same path (T37).
func resolvePathArg(arg, p string, info os.FileInfo, err error, cwd string, base resolution) resolution {
	if err != nil {
		base.message = fmt.Sprintf("cannot open %s: no such path", arg)
		return base
	}
	if info.IsDir() {
		return resolveDir(p, base)
	}
	return resolveFile(p, cwd, base.roots)
}

// resolveDir handles a directory argument against the starred roots in base.
func resolveDir(p string, base resolution) resolution {
	// Already a starred root: no duplicate, it simply becomes active.
	if i := indexOfPath(base.roots, p); i >= 0 {
		base.active = i
		return base
	}
	// Inside a starred root: refused, the covering root becomes active instead.
	if i := coveringRoot(base.roots, p); i >= 0 {
		base.active = i
		base.message = fmt.Sprintf("%s is inside starred root %s; opening it instead", p, base.roots[i].path)
		return base
	}
	// A parent of a starred root: refused, the session still starts unchanged.
	if i := coveredRoot(base.roots, p); i >= 0 {
		base.message = fmt.Sprintf("%s contains starred root %s; not added", p, base.roots[i].path)
		return base
	}
	// A fresh ad-hoc root: listed first and active alongside the starred roots.
	return prependAdhoc(p, base.roots)
}

// resolveFile handles a file argument: the cwd becomes the ad-hoc root and the
// file opens as the active document, whether or not it falls inside any root.
func resolveFile(file, cwd string, starred []root) resolution {
	adhoc := filepath.Clean(cwd)
	// If the cwd is already a starred root, keep roots unique by exact path
	// rather than adding a duplicate ad-hoc entry.
	if i := indexOfPath(starred, adhoc); i >= 0 {
		return resolution{roots: starred, active: i, docPath: file}
	}
	res := prependAdhoc(adhoc, starred)
	res.docPath = file
	return res
}

// prependAdhoc puts an unstarred ad-hoc root at the head of the starred roots,
// active on launch.
func prependAdhoc(path string, starred []root) resolution {
	roots := make([]root, 0, len(starred)+1)
	roots = append(roots, root{path: path})
	roots = append(roots, starred...)
	return resolution{roots: roots, active: 0}
}

// starredRoots turns stored paths into starred roots, cleaned for exact-path
// comparison. Missing paths are kept (S2): this does not stat them.
func starredRoots(paths []string) []root {
	roots := make([]root, 0, len(paths))
	for _, p := range paths {
		roots = append(roots, root{path: filepath.Clean(p), starred: true})
	}
	return roots
}

// absPath resolves path against cwd when it is relative, then cleans it, so
// every root is comparable by exact string.
func absPath(path, cwd string) string {
	if !filepath.IsAbs(path) {
		path = filepath.Join(cwd, path)
	}
	return filepath.Clean(path)
}

func indexOfPath(roots []root, path string) int {
	for i, r := range roots {
		if r.path == path {
			return i
		}
	}
	return -1
}

// coveringRoot returns the index of a root that strictly contains p, or -1.
func coveringRoot(roots []root, p string) int {
	for i, r := range roots {
		if strictlyInside(r.path, p) {
			return i
		}
	}
	return -1
}

// coveredRoot returns the index of a root strictly inside p, or -1.
func coveredRoot(roots []root, p string) int {
	for i, r := range roots {
		if strictlyInside(p, r.path) {
			return i
		}
	}
	return -1
}

// strictlyInside reports whether child is a proper descendant of parent. Both
// are expected to be cleaned absolute paths.
func strictlyInside(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
