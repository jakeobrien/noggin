// Command noggin is a TUI for viewing markdown files with an integrated
// file/folder explorer and manager.
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Decide the markdown style while stdin is still ours to query; see
	// markdownStyle for why it cannot wait until the first render.
	markdownStyle()

	m := newModel()

	// Temporary launch handling: a bare path opens that document. M3.1 replaces
	// this wholesale with the S1 launch contract, where arguments resolve the
	// session's roots and every rejection still yields a running session (D11).
	if len(os.Args) > 1 {
		doc, err := loadDocument(os.Args[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, "noggin:", err)
			os.Exit(1)
		}
		m = m.openDocument(doc)
	}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "noggin:", err)
		os.Exit(1)
	}
}
