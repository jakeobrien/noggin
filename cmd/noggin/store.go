package main

import (
	"os"
	"path/filepath"
	"strings"
)

// storeSubdir and storeName locate the starred-roots file under the user's
// config directory (D7: os.UserConfigDir(), e.g. ~/.config/noggin/roots on
// Linux). One plain file, hand-editable, so a moved root can be repaired
// directly in an editor (S2).
const (
	storeSubdir = "noggin"
	storeName   = "roots"
)

// storePath resolves the starred-roots file's path. An error here (no
// resolvable config directory) is treated by callers the same as an absent
// store: loadStarredRoots yields no roots rather than failing the launch.
func storePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, storeSubdir, storeName), nil
}

// loadStarredRoots resolves the store path and loads it, folding a
// non-resolvable config directory into the same "no starred roots" outcome
// as a missing or garbage file. This is the entry point main.go wires into
// resolveRoots' starred argument.
func loadStarredRoots() []string {
	path, err := storePath()
	if err != nil {
		return nil
	}
	return loadStore(path)
}

// loadStore reads the starred-roots file at path: one absolute path per
// non-blank, non-`#`-comment line (comments and blank lines are ignored, in
// the spirit of the `.gitignore` files noggin already parses, so the format
// stays comfortable to hand-edit). A missing file, an unreadable one, or one
// whose lines are not absolute paths (garbage) all yield an empty list
// rather than an error — S2: "a missing or unparseable file yields an empty
// root list rather than a failure." Paths are not stat'd here: a stored path
// that no longer exists on disk is still returned, unpruned (S2).
func loadStore(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var paths []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || !filepath.IsAbs(line) {
			continue
		}
		paths = append(paths, filepath.Clean(line))
	}
	return paths
}

// saveStore writes paths to the store at path, one absolute path per line,
// creating the config directory first if it does not exist yet.
func saveStore(path string, paths []string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var b strings.Builder
	for _, p := range paths {
		b.WriteString(p)
		b.WriteByte('\n')
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// toggleStore flips target's membership in the store at path — adds it if
// absent, drops it if present — and writes the result through immediately:
// S2's star toggle is not batched, so a star survives a process restart. It
// returns the resulting starred-paths list.
func toggleStore(path, target string) ([]string, error) {
	target = filepath.Clean(target)
	current := loadStore(path)

	i := -1
	for j, p := range current {
		if p == target {
			i = j
			break
		}
	}

	var updated []string
	if i >= 0 {
		updated = append(updated, current[:i]...)
		updated = append(updated, current[i+1:]...)
	} else {
		updated = append(append([]string{}, current...), target)
	}

	if err := saveStore(path, updated); err != nil {
		return nil, err
	}
	return updated, nil
}
