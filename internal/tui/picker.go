package tui

import (
	"sort"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/thaitrn/binman/internal/apps"
	"github.com/thaitrn/binman/internal/scan"
)

// scannedMsg delivers the leftover scan result for one app (async).
type scannedMsg struct {
	idx     int
	path    string
	matches []scan.Match
}

// pickerModel is the two-pane app uninstaller: left = app list with size bars,
// right = details + live leftover preview of the highlighted app.
type pickerModel struct {
	entries  []apps.Entry // sorted by size desc
	maxSize  int64
	selected map[string]bool
	cursor   int
	offset   int // first visible row
	width    int
	height   int
	cache    map[string][]scan.Match // leftover scan cache, by app path
	scanIdx  int                     // app index currently being scanned
	canceled bool
}

func newPickerModel(entries []apps.Entry) *pickerModel {
	es := make([]apps.Entry, len(entries))
	copy(es, entries)
	sort.SliceStable(es, func(i, j int) bool {
		if es[i].Size != es[j].Size {
			return es[i].Size > es[j].Size
		}
		return es[i].App.Name < es[j].App.Name
	})
	var max int64
	for _, e := range es {
		if e.Size > max {
			max = e.Size
		}
	}
	return &pickerModel{
		entries:  es,
		maxSize:  max,
		selected: make(map[string]bool),
		cache:    make(map[string][]scan.Match),
	}
}

func (m *pickerModel) Init() tea.Cmd { return m.scanCmd() }

func (m *pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.clampOffset()
		return m, m.scanCmd()
	case scannedMsg:
		m.cache[msg.path] = msg.matches
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			m.canceled = true
			return m, tea.Quit
		case "enter":
			return m, tea.Quit
		case "up", "k":
			m.moveCursor(-1)
			return m, m.scanCmd()
		case "down", "j":
			m.moveCursor(1)
			return m, m.scanCmd()
		case "pgup":
			m.moveCursor(-m.listHeight())
			return m, m.scanCmd()
		case "pgdown":
			m.moveCursor(m.listHeight())
			return m, m.scanCmd()
		case " ":
			m.toggleCurrent()
		case "g":
			m.setCursor(0)
			return m, m.scanCmd()
		case "G":
			m.setCursor(len(m.entries) - 1)
			return m, m.scanCmd()
		case "a":
			m.toggleAll()
		}
	}
	return m, nil
}

// scanCmd returns a command that scans the highlighted app's leftovers, unless
// already cached. The scan runs asynchronously so the UI never blocks.
func (m *pickerModel) scanCmd() tea.Cmd {
	if len(m.entries) == 0 {
		return nil
	}
	idx := m.cursor
	a := m.entries[idx].App
	if _, ok := m.cache[a.Path]; ok {
		return nil
	}
	m.scanIdx = idx
	path := a.Path
	app := a
	return func() tea.Msg {
		ms, _ := scan.Scan(app)
		return scannedMsg{idx: idx, path: path, matches: ms}
	}
}

func (m *pickerModel) moveCursor(delta int) { m.setCursor(m.cursor + delta) }

func (m *pickerModel) setCursor(i int) {
	if len(m.entries) == 0 {
		return
	}
	if i < 0 {
		i = 0
	}
	if i > len(m.entries)-1 {
		i = len(m.entries) - 1
	}
	m.cursor = i
	m.clampOffset()
}

func (m *pickerModel) clampOffset() {
	h := m.listHeight()
	if h <= 0 {
		return
	}
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+h {
		m.offset = m.cursor - h + 1
	}
	if m.offset < 0 {
		m.offset = 0
	}
}

func (m *pickerModel) toggleCurrent() {
	if len(m.entries) == 0 {
		return
	}
	p := m.entries[m.cursor].App.Path
	m.selected[p] = !m.selected[p]
}

func (m *pickerModel) toggleAll() {
	allOn := len(m.selected) >= len(m.entries)
	m.selected = make(map[string]bool)
	if !allOn {
		for _, e := range m.entries {
			m.selected[e.App.Path] = true
		}
	}
}

// listHeight is the number of app rows that fit in the left pane.
func (m *pickerModel) listHeight() int {
	h := m.height - 6 // title(1) + header(1) + status(2) + borders(2)
	if h < 1 {
		return 1
	}
	return h
}

// SelectApps runs the picker TUI and returns the chosen entries.
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
