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

// storedRoot is one entry in the S1/S2 Roots store: an absolute Path plus
// an optional Name (S8). Every entry in this store is, by construction, a
// starred Root — toggleStore is the only writer besides a future rename
// (M7.2). An empty Name means the Root carries no name of its own yet, so
// S8's default (the folder's base name) applies wherever a Root's label is
// shown; that default is computed from Path by the caller that needs a
// label, not stored here, since storing it would go stale the moment the
// folder itself is renamed on disk. Only a rename (M7.2) ever writes a
// non-empty Name.
type storedRoot struct {
	Path string
	Name string
}

// loadStarredRoots resolves the store path and loads it, folding a
// non-resolvable config directory into the same "no starred roots" outcome
// as a missing or garbage file. This is the entry point main.go wires into
// resolveRoots' starred argument — the full []storedRoot, Name included
// (S8/M7.2): resolveRoots/resolveLaunch/starredRoots (roots.go) carry a
// Root's Name from the store into its runtime root.name (root.displayName),
// so a Root renamed in an earlier session shows its name from the moment a
// fresh one launches, not only after an in-session rename.
func loadStarredRoots() []storedRoot {
	path, err := storePath()
	if err != nil {
		return nil
	}
	return loadStore(path)
}

// loadStore reads the starred-roots file at path: one entry per non-blank,
// non-`#`-comment line (comments and blank lines are ignored, in the spirit
// of the `.gitignore` files noggin already parses, so the format stays
// comfortable to hand-edit). Each line is an absolute path, optionally
// followed by a tab and an escaped Name (S8, escapeStoreField) — the tab and
// everything after it is new with S8; every pre-S8 line (a bare path, the
// only shape that ever existed before) still parses the same way, yielding
// an entry with an empty Name, so an existing store loads with S8's default
// names (the caller derives them from Path) exactly as it always resolved
// those paths. A missing file, an unreadable one, or a line whose path
// field is not absolute (garbage) is dropped rather than failing the whole
// load — S2: "a missing or unparseable file yields an empty root list
// rather than a failure." Paths are not stat'd here: a stored path that no
// longer exists on disk is still returned, unpruned (S2).
func loadStore(path string) []storedRoot {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var roots []storedRoot
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		p, name := line, ""
		if i := strings.IndexByte(line, '\t'); i >= 0 {
			p, name = line[:i], unescapeStoreField(line[i+1:])
		}
		if !filepath.IsAbs(p) {
			continue
		}
		roots = append(roots, storedRoot{Path: filepath.Clean(p), Name: name})
	}
	return roots
}

// saveStore writes roots to the store at path, one entry per line: a bare
// path for an unnamed Root (S8's default applies, and this is also the
// pre-S8 format — a store holding only unnamed Roots round-trips
// byte-for-byte the same as it always has), or `<path>\t<name>`
// (escapeStoreField) once a rename (M7.2) has given it a Name. Creates the
// store's directory first if it does not exist yet.
func saveStore(path string, roots []storedRoot) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var b strings.Builder
	for _, r := range roots {
		b.WriteString(r.Path)
		if r.Name != "" {
			b.WriteByte('\t')
			b.WriteString(escapeStoreField(r.Name))
		}
		b.WriteByte('\n')
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// toggleStore flips target's membership in the store at path — adds it,
// unnamed (S8: a freshly-starred Root has no Name until it is renamed), if
// absent; drops it, Name and all, if present — and writes the result
// through immediately: S2's star toggle is not batched, so a star survives
// a process restart. It returns the resulting stored roots.
func toggleStore(path, target string) ([]storedRoot, error) {
	target = filepath.Clean(target)
	current := loadStore(path)

	i := -1
	for j, r := range current {
		if r.Path == target {
			i = j
			break
		}
	}

	var updated []storedRoot
	if i >= 0 {
		updated = append(updated, current[:i]...)
		updated = append(updated, current[i+1:]...)
	} else {
		updated = append(append([]storedRoot{}, current...), storedRoot{Path: target})
	}

	if err := saveStore(path, updated); err != nil {
		return nil, err
	}
	return updated, nil
}

// renameStore sets target's Name (S8) in the store at path, leaving
// membership untouched — a rename (M7.2), not a star/unstar. name == ""
// reverts to the default (the folder's base name, computed by the caller,
// never stored). Returns the resulting stored roots; target not present is
// a no-op (defensive: every caller — model.commitRename — has already
// confirmed target is starred before a rename is even offered, so this
// should not occur in practice, but a store that has drifted out from
// under the running session must not panic or corrupt the file over it).
func renameStore(path, target, name string) ([]storedRoot, error) {
	target = filepath.Clean(target)
	current := loadStore(path)
	updated := make([]storedRoot, len(current))
	copy(updated, current)
	for i, r := range updated {
		if r.Path == target {
			updated[i].Name = name
			break
		}
	}
	if err := saveStore(path, updated); err != nil {
		return nil, err
	}
	return updated, nil
}

// escapeStoreField and unescapeStoreField guard a Roots-store Name field
// against the tab and newline bytes that would otherwise corrupt the
// store's one-entry-per-line, tab-delimited-when-named format (S8) — the
// same hazard, and the same escaping scheme, as the Document store's
// <title>/<key> fields (docstore.go's escapeDocField/unescapeDocField,
// M1.7/M1.8). Kept as this store's own copy rather than a shared call,
// since the two stores are deliberately separate formats (D61) that should
// not need to import each other's parsing details to stay correct.
func escapeStoreField(field string) string {
	r := strings.NewReplacer(`\`, `\\`, "\t", `\t`, "\n", `\n`)
	return r.Replace(field)
}

// unescapeStoreField reverses escapeStoreField on load. A lone trailing
// backslash (malformed escaping, e.g. from a hand-edited file) is left
// as-is rather than panicking or dropping the line: escapeStoreField itself
// never produces one, so this only guards against a hand-edited store.
func unescapeStoreField(field string) string {
	var b strings.Builder
	b.Grow(len(field))
	for i := 0; i < len(field); i++ {
		if field[i] == '\\' && i+1 < len(field) {
			switch field[i+1] {
			case '\\':
				b.WriteByte('\\')
				i++
				continue
			case 't':
				b.WriteByte('\t')
				i++
				continue
			case 'n':
				b.WriteByte('\n')
				i++
				continue
			}
		}
		b.WriteByte(field[i])
	}
	return b.String()
}
