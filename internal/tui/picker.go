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

func (i appItem) FilterValue() string { return i.entry.App.Name + " " + i.entry.App.BundleID }

// appMultiDelegate renders one row with a checkbox; selection state lives in
// the shared *selected map (toggled by the model on space/a).
type appMultiDelegate struct{ selected *map[string]bool }

func (appMultiDelegate) Height() int                             { return 1 }
func (appMultiDelegate) Spacing() int                            { return 0 }
func (appMultiDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d appMultiDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	it, ok := item.(appItem)
	if !ok {
		return
	}
	path := it.entry.App.Path
	box := "☐"
	if (*d.selected)[path] {
		box = checkStyle.Render("☑")
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
	pathW := m.Width() - 48
	if pathW < 10 {
		pathW = 10
	}
	fmt.Fprintln(w, style.Render(fmt.Sprintf("%s%s %-32s %10s  %s",
		marker, box, truncate(name, 32), human.Bytes(it.entry.Size), truncate(path, pathW))))
}

// pickerModel is the multi-select app list.
type pickerModel struct {
	list     list.Model
	entries  []apps.Entry
	selected map[string]bool
	canceled bool
}

func newPickerModel(entries []apps.Entry) *pickerModel {
	selected := make(map[string]bool)
	items := make([]list.Item, 0, len(entries))
	for _, e := range entries {
		items = append(items, appItem{entry: e})
	}
	l := list.New(items, appMultiDelegate{selected: &selected}, 80, 20)
	l.Title = "Select apps to uninstall"
	l.SetShowStatusBar(false)
	l.SetShowHelp(true)
	l.SetFilteringEnabled(false) // list all apps; no type-to-filter
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = helpStyle
	l.Styles.HelpStyle = helpStyle
	return &pickerModel{list: l, entries: entries, selected: selected}
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
			return m, tea.Quit
		case "ctrl+c", "esc":
			m.canceled = true
			return m, tea.Quit
		case "q":
			if m.list.FilterState() == list.Unfiltered {
				m.canceled = true
				return m, tea.Quit
			}
		case " ":
			m.toggleCurrent()
			return m, nil
		case "a":
			m.toggleAllVisible()
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *pickerModel) toggleCurrent() {
	it, ok := m.list.SelectedItem().(appItem)
	if !ok {
		return
	}
	p := it.entry.App.Path
	m.selected[p] = !m.selected[p]
}

// toggleAllVisible flips every item in the current (possibly filtered) view.
func (m *pickerModel) toggleAllVisible() {
	allOn := true
	for _, it := range m.list.Items() {
		if ai, ok := it.(appItem); ok && !m.selected[ai.entry.App.Path] {
			allOn = false
			break
		}
	}
	for _, it := range m.list.Items() {
		if ai, ok := it.(appItem); ok {
			m.selected[ai.entry.App.Path] = !allOn
		}
	}
}

func (m *pickerModel) View() string { return m.list.View() }

// SelectApps runs the multi-select picker and returns the chosen entries plus
// whether the user confirmed (vs canceled).
func SelectApps(entries []apps.Entry) ([]apps.Entry, bool, error) {
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
	if !ok || pm.canceled {
		return nil, false, nil
	}
	var out []apps.Entry
	for _, e := range pm.entries {
		if pm.selected[e.App.Path] {
			out = append(out, e)
		}
	}
	return out, true, nil
}

func noStyle() lipgloss.Style { return lipgloss.NewStyle() }
func selectedRowStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
}
