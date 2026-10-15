package main

import (
	"os"
	"path/filepath"
	"strings"
)

// themeStoreName is the persisted Theme choice's file name, alongside the
// starred-Roots and document stores under the same data directory (dataDir):
// one plain file holding the Theme's Name, hand-editable.
const themeStoreName = "theme"

// themeStorePath resolves the Theme store's path via dataDir. An error (no
// resolvable data directory) degrades like the other stores: the choice
// simply is not persisted.
func themeStorePath() (string, error) {
	dir, err := dataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, storeSubdir, themeStoreName), nil
}

// loadThemeChoice reads the Theme store at path and returns the Theme it
// names. ok is false for a missing, unreadable or empty file, or a name that
// is no longer one of `themes` (a Theme since renamed or removed): the caller
// keeps the default rather than failing the launch.
func loadThemeChoice(path string) (Theme, bool) {
	if path == "" {
		return Theme{}, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Theme{}, false
	}
	name := strings.TrimSpace(string(data))
	for _, t := range themes {
		if t.Name == name {
			return t, true
		}
	}
	return Theme{}, false
}

// saveThemeChoice writes name to the Theme store at path, creating its
// directory first if needed.
func saveThemeChoice(path, name string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(name+"\n"), 0o644)
}
