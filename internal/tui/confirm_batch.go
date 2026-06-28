package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/thaitrn/binman/internal/human"
)

// batchConfirmModel shows the aggregate of a multi-app uninstall and asks the
// user to confirm before anything is deleted.
type batchConfirmModel struct {
	names     []string
	items     int
	size      int64
	shared    int
	confirmed bool
	canceled  bool
	width     int
}

func newBatchConfirmModel(names []string, items int, size int64, shared int) *batchConfirmModel {
	return &batchConfirmModel{names: names, items: items, size: size, shared: shared}
}

func (m *batchConfirmModel) Init() tea.Cmd { return nil }

func (m *batchConfirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.confirmed = true
			return m, tea.Quit
		case "ctrl+c", "q", "esc":
			m.canceled = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *batchConfirmModel) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(fmt.Sprintf("Uninstall %d app(s)", len(m.names))))
	b.WriteString("\n\n")
	for _, n := range m.names {
		b.WriteString("  • " + truncate(n, 60) + "\n")
	}
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  %s leftover item(s), ~%s → Trash\n",
		footerStyle.Render(fmt.Sprintf("%d", m.items)), human.Bytes(m.size)))
	if m.shared > 0 {
		b.WriteString(sharedStyle.Render(fmt.Sprintf("  %d shared container(s) skipped (off by default)\n", m.shared)))
	}
	b.WriteString("\n" + helpStyle.Render("enter confirm · esc/q cancel"))
	return b.String()
}

// ConfirmBatch runs the aggregate confirm screen and returns whether the user
// confirmed. Names is the list of app names to display.
func ConfirmBatch(names []string, items int, size int64, shared int) (bool, error) {
	m := newBatchConfirmModel(names, items, size, shared)
	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return false, err
	}
	bm, ok := final.(*batchConfirmModel)
	return ok && bm.confirmed, nil
}
