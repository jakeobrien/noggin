// Command noggin is a TUI for viewing markdown files with an integrated
// file/folder explorer and manager.
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"noggin/internal/tree"
)

func main() {
	// Decide the markdown style while stdin is still ours to query; see
	// markdownStyle for why it cannot wait until the first render.
	markdownStyle()

	m := newModel()

	// Temporary launch handling: a directory argument browses that folder in the
	// Files Nav, a file argument opens that document. M3.1 replaces this
	// wholesale with the S1 launch contract, where arguments resolve the
	// session's roots and every rejection still yields a running session (D11).
	if len(os.Args) > 1 {
		path := os.Args[1]
		info, err := os.Stat(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, "noggin:", err)
			os.Exit(1)
		}
		if info.IsDir() {
			root, err := tree.Walk(path)
			if err != nil {
				fmt.Fprintln(os.Stderr, "noggin:", err)
				os.Exit(1)
			}
			m = m.openFolder(root)
		} else {
			doc, err := loadDocument(path)
			if err != nil {
				fmt.Fprintln(os.Stderr, "noggin:", err)
				os.Exit(1)
			}
			m = m.openDocument(doc)
		}
	}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "noggin:", err)
		os.Exit(1)
	}
}
