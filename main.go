package main

import (
	"fmt"
	"os"

	"github.com/april/sid-synth/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	var filename string
	if len(os.Args) > 1 {
		filename = os.Args[1]
	}

	m, err := ui.New(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running: %v\n", err)
		os.Exit(1)
	}
}
