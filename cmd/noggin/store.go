package main

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// storeSubdir and storeName locate the starred-roots file under noggin's data
// directory (dataDir, S1). One plain file, hand-editable, so a moved root can
// be repaired directly in an editor (S2). This is a clean break from C1/D7's
// config-directory location: an existing config-directory `roots` file is
// never read or migrated here (S1) — main.go's loadStarredRoots is the only
// reader, and it now resolves exclusively through dataDir.
const (
	storeSubdir = "noggin"
	storeName   = "roots"
)

// dataDir resolves the root directory for noggin's persisted application
// data (the starred-Roots store here; M1.2's document store next), distinct
// from os.UserConfigDir()'s user-facing configuration directory (S1). Data
// is not meant to be hand-edited day to day the way config is, so it lives
// in its own platform-conventional location — mirroring the split the
// standard library's own os.UserCacheDir() already makes from
// os.UserConfigDir(): $XDG_DATA_HOME (or $HOME/.local/share) on Unix,
// ~/Library/Application Support on Darwin, %LocalAppData% on Windows.
func dataDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		dir := os.Getenv("LocalAppData")
		if dir == "" {
			return "", errors.New("%LocalAppData% is not defined")
		}
		return dir, nil

	case "darwin", "ios":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Library", "Application Support"), nil

	case "plan9":
		home := os.Getenv("home")
		if home == "" {
			return "", errors.New("$home is not defined")
		}
		return filepath.Join(home, "lib"), nil

	default: // Unix
		if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
			if !filepath.IsAbs(dir) {
				return "", errors.New("path in $XDG_DATA_HOME is relative")
			}
			return dir, nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".local", "share"), nil
	}
}

// storePath resolves the starred-roots file's path via dataDir. An error
// here (no resolvable data directory) is treated by callers the same as an
// absent store: loadStarredRoots yields no roots rather than failing the
// launch.
func storePath() (string, error) {
	dir, err := dataDir()
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
