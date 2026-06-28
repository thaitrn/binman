// Package tui provides the Bubble Tea interfaces for binman: an interactive
// checkbox selector for leftovers and a progress view for the Trash step.
package tui

import "github.com/charmbracelet/lipgloss"

// Shared lipgloss theme used across binman's TUI views.
var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	headerStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("245"))
	checkStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))  // green check
	cursorStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("213"))
	sharedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241")) // dimmed
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	errStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	footerStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
)
