package main

import (
	"fmt"
	"os"

	"bitcrafttsbc/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	model := tui.New()
	program := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "起動失敗: %v\n", err)
		os.Exit(1)
	}
}
