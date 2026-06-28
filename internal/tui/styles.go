// Package tui provides the Bubble Tea interfaces for binman.
package tui

import "github.com/charmbracelet/lipgloss"

// Catppuccin Mocha-inspired palette.
var (
	cAccent = lipgloss.Color("#89b4fa") // blue (titles, cursor, selection)
	cGreen  = lipgloss.Color("#a6e3a1") // green (selected check, size bar)
	cYellow = lipgloss.Color("#f9e2af") // yellow (shared / system warning)
	cRed    = lipgloss.Color("#f38ba8") // red (errors)
	cDim    = lipgloss.Color("#6c7086") // overlay0 (muted text)
	cBorder = lipgloss.Color("#45475a") // surface1 (borders)
	cText   = lipgloss.Color("#cdd6f4") // text (primary)
)

// Shared text styles.
var (
	titleStyle   = lipgloss.NewStyle().Foreground(cAccent).Bold(true)
	headerStyle  = lipgloss.NewStyle().Foreground(cDim)
	checkStyle   = lipgloss.NewStyle().Foreground(cGreen)
	cursorStyle  = lipgloss.NewStyle().Foreground(cAccent).Bold(true)
	sharedStyle  = lipgloss.NewStyle().Foreground(cYellow)
	helpStyle    = lipgloss.NewStyle().Foreground(cDim)
	errStyle     = lipgloss.NewStyle().Foreground(cRed)
	footerStyle  = lipgloss.NewStyle().Foreground(cAccent).Bold(true)
	dimStyle     = lipgloss.NewStyle().Foreground(cDim)
	nameStyle    = lipgloss.NewStyle().Foreground(cText)
	selectedName = lipgloss.NewStyle().Foreground(cAccent).Bold(true)
)
