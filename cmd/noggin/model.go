package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

// model is the walking-skeleton root Bubble Tea model. It does nothing yet
// but start cleanly and quit on q / ctrl+c, restoring the terminal.
type model struct{}

func newModel() model {
	return model{}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	return "noggin\n"
}
