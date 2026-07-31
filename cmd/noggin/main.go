// Command noggin is a TUI for viewing markdown files with an integrated
// file/folder explorer and manager.
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
)

func main() {
	// Decide the markdown style while stdin is still ours to query; see
	// markdownStyle for why it cannot wait until the first render.
	markdownStyle()

	// Resolve the session's roots from the launch argument (S1). A bad path is a
	// rejection message and a session that still starts, not a process exit (T7);
	// walking the active root and opening any active document happen inside
	// applyResolution, which downgrades their failures to messages too.
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	// loadStarredRoots reads the S2 store (D7's os.UserConfigDir()); a fresh
	// install with no store yet, same as a garbage one, is a first run.
	res := resolveRoots(os.Args[1:], loadStarredRoots(), cwd)
	// storePath is resolved once here and threaded into the model so a Roots
	// Nav star toggle (M3.3) persists without re-deriving os.UserConfigDir()
	// on every keypress; an unresolvable config directory (storePath's own
	// error) becomes "", the same degrade loadStarredRoots already applies.
	path, _ := storePath()
	m := newModel().withStorePath(path).applyResolution(res)

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "noggin:", err)
		os.Exit(1)
	}
}
