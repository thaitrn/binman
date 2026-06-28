package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/thaitrn/binman/internal/human"
	"github.com/thaitrn/binman/internal/scan"
	"github.com/thaitrn/binman/internal/trash"
)

// stepMsg reports the result of trashing one item.
type stepMsg struct {
	idx int
	res trash.Result
}

type allDoneMsg struct{}

// progressModel drives the Trash step with a progress bar and a summary.
type progressModel struct {
	items    []scan.Match
	results  []trash.Result
	dryRun   bool
	progress progress.Model
	width    int
	done     bool
}

func newProgressModel(items []scan.Match, dryRun bool) *progressModel {
	p := progress.New(progress.WithDefaultGradient())
	p.PercentageStyle = footerStyle
	return &progressModel{items: items, dryRun: dryRun, progress: p}
}

func (m *progressModel) Init() tea.Cmd {
	if len(m.items) == 0 {
		return func() tea.Msg { return allDoneMsg{} }
	}
	return trashStep(0, m.items[0].Path, m.dryRun)
}

// trashStep returns a command that trashes one path and reports back.
func trashStep(idx int, path string, dryRun bool) tea.Cmd {
	return func() tea.Msg {
		res := trash.Trash([]string{path}, dryRun)[0]
		return stepMsg{idx: idx, res: res}
	}
}

func (m *progressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		if msg.Width > 4 {
			m.progress.Width = msg.Width - 4
		}
	case stepMsg:
		m.results = append(m.results, msg.res)
		_ = m.progress.SetPercent(float64(len(m.results)) / float64(len(m.items)))
		if len(m.results) >= len(m.items) {
			return m, func() tea.Msg { return allDoneMsg{} }
		}
		return m, trashStep(len(m.results), m.items[len(m.results)].Path, m.dryRun)
	case allDoneMsg:
		m.done = true
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q", "enter", "esc":
			if m.done {
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m *progressModel) View() string {
	var b strings.Builder
	if !m.done {
		cur := ""
		if len(m.results) < len(m.items) {
			cur = m.items[len(m.results)].Path
		}
		b.WriteString(titleStyle.Render(fmt.Sprintf("Trashing %d/%d items", len(m.results), len(m.items))))
		b.WriteString("\n\n")
		b.WriteString(m.progress.View() + "\n\n")
		b.WriteString(helpStyle.Render(truncate(cur, max(40, m.width-2))))
		return b.String()
	}
	var freed int64
	failed := 0
	for i, r := range m.results {
		if r.Err != nil {
			failed++
			continue
		}
		freed += m.items[i].Size
	}
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("Done"))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("  moved %d item(s) · ~%s freed\n", len(m.results)-failed, human.Bytes(freed)))
	if failed > 0 {
		b.WriteString(errStyle.Render(fmt.Sprintf("  %d item(s) failed\n", failed)))
	}
	b.WriteString("\n" + helpStyle.Render("press enter to exit"))
	return b.String()
}

// RunProgress runs the Trash step as a TUI and returns the per-path results.
func RunProgress(items []scan.Match, dryRun bool) ([]trash.Result, error) {
	m := newProgressModel(items, dryRun)
	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return nil, err
	}
	return final.(*progressModel).results, nil
}
