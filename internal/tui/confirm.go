package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/thaitrn/binman/internal/human"
	"github.com/thaitrn/binman/internal/scan"
)

// listItem pairs a leftover match with its selected state.
type listItem struct {
	match    scan.Match
	selected bool
}

// confirmModel is the interactive checkbox selector.
type confirmModel struct {
	title     string
	items     []listItem
	cursor    int
	width     int
	confirmed bool
	canceled  bool
}

func newConfirmModel(title string, matches []scan.Match, defaultOn bool) *confirmModel {
	items := make([]listItem, len(matches))
	for i, m := range matches {
		// Group/shared containers are off by default to avoid removing shared data.
		items[i] = listItem{match: m, selected: defaultOn && !m.Shared}
	}
	return &confirmModel{title: title, items: items}
}

func (m *confirmModel) Init() tea.Cmd { return nil }

func (m *confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.canceled = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case " ":
			if len(m.items) > 0 {
				m.items[m.cursor].selected = !m.items[m.cursor].selected
			}
		case "a":
			all := !anySelected(m.items)
			for i := range m.items {
				m.items[i].selected = all && !m.items[i].match.Shared
			}
		case "enter":
			m.confirmed = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *confirmModel) View() string {
	if m.width == 0 {
		return titleStyle.Render(m.title)
	}
	var b strings.Builder
	b.WriteString(titleStyle.Render(m.title))
	b.WriteString("\n\n")
	pathWidth := max(20, m.width-30)
	b.WriteString(headerStyle.Render(fmt.Sprintf("  %-11s %-9s %s", "TYPE", "SIZE", "PATH")))
	b.WriteString("\n")
	for i, it := range m.items {
		marker := "  "
		row := lipgloss.NewStyle()
		if i == m.cursor {
			marker = cursorStyle.Render("▸ ")
			row = row.Bold(true)
		}
		box := "☐"
		if it.selected {
			box = checkStyle.Render("☑")
		}
		shared := ""
		if it.match.Shared {
			shared = sharedStyle.Render(" (shared)")
		}
		line := fmt.Sprintf("%s%s %-11s %-9s %s%s",
			marker, box, it.match.Type, human.Bytes(it.match.Size), truncate(it.match.Path, pathWidth), shared)
		b.WriteString(row.Render(line))
		b.WriteString("\n")
	}
	count, size := selectedStats(m.items)
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("%s  %s",
		footerStyle.Render(fmt.Sprintf("%d/%d selected", count, len(m.items))),
		helpStyle.Render(fmt.Sprintf("· ~%s · ↑↓/jk move · space toggle · a all · enter confirm · q cancel", human.Bytes(size))),
	))
	return b.String()
}

// Confirm runs the selector TUI and returns the chosen matches plus whether
// the user confirmed (vs canceled).
func Confirm(title string, matches []scan.Match, defaultOn bool) ([]scan.Match, bool, error) {
	m := newConfirmModel(title, matches, defaultOn)
	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return nil, false, err
	}
	cm, ok := final.(*confirmModel)
	if !ok || cm.canceled || !cm.confirmed {
		return nil, false, nil
	}
	var out []scan.Match
	for _, it := range cm.items {
		if it.selected {
			out = append(out, it.match)
		}
	}
	return out, true, nil
}

func anySelected(items []listItem) bool {
	for _, it := range items {
		if it.selected {
			return true
		}
	}
	return false
}

func selectedStats(items []listItem) (count int, size int64) {
	for _, it := range items {
		if it.selected {
			count++
			size += it.match.Size
		}
	}
	return
}

// truncate shortens s to n runes, appending an ellipsis when truncated.
func truncate(s string, n int) string {
	r := []rune(s)
	if n <= 0 || len(r) <= n {
		return s
	}
	if n <= 1 {
		return "…"
	}
	return string(r[:n-1]) + "…"
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
