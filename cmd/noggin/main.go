// Command noggin is a TUI for viewing markdown files with an integrated
// file/folder explorer and manager.
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "noggin:", err)
		os.Exit(1)
	}
}
