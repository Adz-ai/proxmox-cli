package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Options configures the TUI.
type Options struct {
	// Source provides cluster state and guest actions.
	Source DataSource
	// ContextName is shown in the title bar.
	ContextName string
	// Refresh is the auto-refresh interval; defaults to five seconds.
	Refresh time.Duration
}

// Run starts the TUI and blocks until the user quits.
func Run(opts Options) error {
	program := tea.NewProgram(NewModel(opts), tea.WithAltScreen())
	_, err := program.Run()
	return err
}
