package tui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/thaitrn/binman/internal/apps"
	"github.com/thaitrn/binman/internal/human"
)

// appItem adapts an apps.Entry for bubbles/list.
type appItem struct{ entry apps.Entry }

func (i appItem) Title() string       { return i.entry.App.Name }
func (i appItem) Description() string { return human.Bytes(i.entry.Size) + "  " + i.entry.App.Path }
func (i appItem) FilterValue() string { return i.entry.App.Name + " " + i.entry.App.BundleID }

// appDelegate renders one row: name, size, path; marks system apps.
type appDelegate struct{}

func (appDelegate) Height() int                             { return 1 }
func (appDelegate) Spacing() int                            { return 0 }
func (appDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d appDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	it, ok := item.(appItem)
	if !ok {
		return
	}
	marker := "  "
	style := noStyle()
	if index == m.Index() {
		marker = cursorStyle.Render("▸ ")
		style = selectedRowStyle()
	}
	name := it.entry.App.Name
	if it.entry.Protected {
		name += sharedStyle.Render(" (system)")
	}
	// name (34) | size (10) | path (rest)
	pathW := m.Width() - 48
	if pathW < 10 {
		pathW = 10
	}
	fmt.Fprintln(w, style.Render(fmt.Sprintf("%s%-32s %10s  %s",
		marker, truncate(name, 32), human.Bytes(it.entry.Size), truncate(it.entry.App.Path, pathW))))
}

// pickerModel is the app-selection list.
type pickerModel struct {
	list     list.Model
	choice   *apps.Entry
	canceled bool
}

func newPickerModel(entries []apps.Entry) *pickerModel {
	items := make([]list.Item, 0, len(entries))
	for _, e := range entries {
		items = append(items, appItem{entry: e})
	}
	l := list.New(items, appDelegate{}, 80, 20)
	l.Title = "Select an app to uninstall"
	l.SetShowStatusBar(false)
	l.SetShowHelp(true)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = helpStyle
	l.Styles.HelpStyle = helpStyle
	return &pickerModel{list: l}
}

func (m *pickerModel) Init() tea.Cmd { return nil }

func (m *pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height)
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if it, ok := m.list.SelectedItem().(appItem); ok {
				e := it.entry
				m.choice = &e
			}
			return m, tea.Quit
		case "ctrl+c":
			m.canceled = true
			return m, tea.Quit
		case "q":
			// quit on q only when not typing a filter
			if m.list.FilterState() == list.Unfiltered {
				m.canceled = true
				return m, tea.Quit
			}
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *pickerModel) View() string { return m.list.View() }

// PickApp runs the picker TUI and returns the chosen entry and whether the user
// confirmed (vs canceled).
func PickApp(entries []apps.Entry) (*apps.Entry, bool, error) {
	if len(entries) == 0 {
		return nil, false, nil
	}
	m := newPickerModel(entries)
	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return nil, false, err
	}
	pm, ok := final.(*pickerModel)
	if !ok || pm.canceled || pm.choice == nil {
		return nil, false, nil
	}
	return pm.choice, true, nil
}

// noStyle / selectedRowStyle are small helpers to avoid re-allocating styles.
func noStyle() lipgloss.Style { return lipgloss.NewStyle() }
func selectedRowStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
}
