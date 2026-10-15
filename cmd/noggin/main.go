// Command noggin is a TUI for viewing markdown files with an integrated
// file/folder explorer and manager.
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
)

func main() {
	// Resolve the session's roots from the launch argument (S1). A bad path is a
	// rejection message and a session that still starts, not a process exit (T7);
	// walking the active root and opening any active document happen inside
	// applyResolution, which downgrades their failures to messages too.
	// loadStarredRoots reads the S1 store (M1.1's dataDir(), a clean break
	// from C1/D7's config-directory location — the old config-dir file is
	// never read or migrated); a fresh install with no store yet, same as a
	// garbage one, is a first run.
	cwd, cwdErr := resolveCwd()
	res := resolveLaunch(os.Args[1:], loadStarredRoots(), cwd, cwdErr)
	// storePath is resolved once here and threaded into the model so a Sources
	// Nav star toggle (M3.3) persists without re-deriving dataDir() on every
	// keypress; an unresolvable data directory (storePath's own error)
	// becomes "", the same degrade loadStarredRoots already applies.
	path, _ := storePath()
	// docsPath/loadDocuments load the S1 document store (M1.2's path-keyed
	// record, M1.4's model wiring) once here, the same one-shot-then-carry
	// pattern as the Roots store just above; every later read or write goes
	// through model.docRecord/withDocRecord, never touching docstore.go or
	// the file directly.
	docsPath, _ := docStorePath()
	// detectRipgrep is a PATH lookup, independent of the launch argument
	// above; resolving it once here keeps the Picker's later startup-time
	// query (M2.5) from re-shelling-out on every reach for it (S14).
	// The Theme chosen last session, if any, is the one to start on; the
	// default stands for a first run or an unreadable/unknown choice.
	themePath, _ := themeStorePath()
	if t, ok := loadThemeChoice(themePath); ok {
		theme = t
	}
	m := newModel().withThemePath(themePath).withStorePath(path).withDocs(docsPath, loadDocuments()).withRipgrep(detectRipgrep()).applyResolution(res)

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "noggin:", err)
		os.Exit(1)
	}
}
