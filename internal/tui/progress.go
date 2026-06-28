package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/thaitrn/binman/internal/human"
	"github.com/thaitrn/binman/internal/safety"
	"github.com/thaitrn/binman/internal/scan"
	"github.com/thaitrn/binman/internal/trash"
)

// Processing phases, shown to the user as discrete steps.
const (
	phaseQuit   = "quit"
	phaseTrash  = "trash"
	phaseVerify = "verify"
	phaseDone   = "done"
)

// Messages for the async phases.
type quitDoneMsg struct{ names []string }
type trashStepMsg struct {
	idx int
	res trash.Result
}
type verifyDoneMsg struct {
	verified int
	failed   int
}

// procModel drives the multi-phase processing view:
// quit apps → move items to Trash → verify removed → done.
type procModel struct {
	appNames   []string
	running    []string       // apps to quit first
	items      []scan.Match   // leftovers to trash
	results    []trash.Result // trash result per item
	idx        int            // current trash index
	verified   int            // items confirmed gone
	failedItem int            // items still present after trash
	phase      string
	dryRun     bool
	progress   progress.Model
	width      int
	height     int
	done       bool
}

func newProcModel(appNames, running []string, items []scan.Match, dryRun bool) *procModel {
	p := progress.New(progress.WithDefaultGradient())
	p.PercentageStyle = footerStyle
	return &procModel{
		appNames: appNames,
		running:  running,
		items:    items,
		results:  make([]trash.Result, len(items)),
		phase:    phaseQuit,
		dryRun:   dryRun,
		progress: p,
	}
}

func (m *procModel) Init() tea.Cmd {
	if len(m.running) > 0 {
		return m.quitCmd()
	}
	return m.startTrash()
}

func (m *procModel) startTrash() tea.Cmd {
	m.phase = phaseTrash
	if len(m.items) == 0 {
		m.phase = phaseDone
		m.done = true
		return nil
	}
	return m.trashCmd(0)
}

func (m *procModel) quitCmd() tea.Cmd {
	apps := m.running
	return func() tea.Msg {
		for _, n := range apps {
			_ = safety.QuitApp(n)
		}
		return quitDoneMsg{names: apps}
	}
}

func (m *procModel) trashCmd(i int) tea.Cmd {
	m.idx = i // index currently in flight (== completed count)
	path := m.items[i].Path
	return func() tea.Msg {
		res := trash.Trash([]string{path}, false)[0]
		return trashStepMsg{idx: i, res: res}
	}
}

func (m *procModel) verifyCmd() tea.Cmd {
	items := m.items
	return func() tea.Msg {
		v, failed := 0, 0
		for _, it := range items {
			if _, err := os.Stat(it.Path); os.IsNotExist(err) {
				v++
			} else {
				failed++
			}
		}
		return verifyDoneMsg{verified: v, failed: failed}
	}
}

func (m *procModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		if msg.Width > 4 {
			m.progress.Width = msg.Width - 4
		}
	case progress.FrameMsg:
		// Drive the progress bar animation toward the last SetPercent target.
		mod, cmd := m.progress.Update(msg)
		if pm, ok := mod.(progress.Model); ok {
			m.progress = pm
		}
		return m, cmd
	case quitDoneMsg:
		return m, m.startTrash()
	case trashStepMsg:
		m.idx = msg.idx
		m.results[msg.idx] = msg.res
		cmd := m.progress.SetPercent(float64(msg.idx+1) / float64(max(1, len(m.items))))
		if msg.idx+1 < len(m.items) {
			return m, tea.Batch(m.trashCmd(msg.idx+1), cmd)
		}
		m.phase = phaseVerify
		return m, tea.Batch(m.verifyCmd(), cmd)
	case verifyDoneMsg:
		m.verified = msg.verified
		m.failedItem = msg.failed
		m.phase = phaseDone
		m.done = true
		return m, m.progress.SetPercent(1)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter", "q", "esc":
			if m.done {
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m *procModel) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("◆ Processing") + dimStyle.Render(
		fmt.Sprintf("  uninstall %s · %d items", joinNames(m.appNames), len(m.items))))
	b.WriteString("\n\n")
	b.WriteString(m.renderSteps())
	b.WriteString("\n")
	b.WriteString(m.progress.View())
	b.WriteString("\n")
	b.WriteString(m.renderItems())
	if m.done {
		b.WriteString("\n" + m.renderSummary())
	}
	return b.String()
}

// renderSteps shows the four phases with status glyphs.
func (m *procModel) renderSteps() string {
	var lines []string
	lines = append(lines, m.stepLine(phaseQuit, "Quit apps", joinNames(m.running)))
	lines = append(lines, m.stepLine(phaseTrash, "Trash", fmt.Sprintf("%d/%d", m.trashProgress(), len(m.items))))
	lines = append(lines, m.stepLine(phaseVerify, "Verify", m.verifyDetail()))
	lines = append(lines, m.stepLine(phaseDone, "Done", ""))
	return strings.Join(lines, "\n")
}

// stepLine renders one phase line with a state glyph based on the current phase.
func (m *procModel) stepLine(phase, label, detail string) string {
	state := phaseState(m.phase, phase)
	glyph := dimStyle.Render("·")
	style := dimStyle
	switch state {
	case 1:
		glyph = cursorStyle.Render("▸")
		style = selectedName
	case 2:
		glyph = checkStyle.Render("✓")
		style = nameStyle
	}
	line := fmt.Sprintf("%s %-9s", glyph, label)
	if detail != "" {
		line += " " + dimStyle.Render(detail)
	}
	return style.Render(line)
}

// phaseState returns 0 pending / 1 active / 2 done for `phase` given current.
func phaseState(current, phase string) int {
	order := map[string]int{phaseQuit: 0, phaseTrash: 1, phaseVerify: 2, phaseDone: 3}
	c, p := order[current], order[phase]
	switch {
	case p < c:
		return 2
	case p == c:
		return 1
	}
	return 0
}

func (m *procModel) trashProgress() int {
	if m.phase == phaseTrash {
		return m.idx // completed count (current index is in flight)
	}
	if phaseState(m.phase, phaseTrash) == 2 {
		return len(m.items)
	}
	return 0
}

func (m *procModel) verifyDetail() string {
	switch m.phase {
	case phaseVerify:
		return "checking…"
	case phaseDone:
		return fmt.Sprintf("%d verified", m.verified)
	}
	return ""
}

// renderItems shows the leftover list with per-item status glyphs.
func (m *procModel) renderItems() string {
	if len(m.items) == 0 {
		return dimStyle.Render("  no items")
	}
	maxH := m.height - 12
	if maxH < 3 {
		maxH = 3
	}
	off := 0
	if m.idx-maxH/2 > 0 {
		off = m.idx - maxH/2
	}
	end := off + maxH
	if end > len(m.items) {
		end = len(m.items)
	}
	var lines []string
	for i := off; i < end; i++ {
		lines = append(lines, m.renderItemRow(i))
	}
	return strings.Join(lines, "\n")
}

func (m *procModel) renderItemRow(i int) string {
	it := m.items[i]
	name := filepath.Base(it.Path)
	if len(name) > 34 {
		name = name[:31] + "…"
	}
	processed := (m.phase == phaseTrash && i < m.idx) || m.phase == phaseVerify || m.phase == phaseDone
	glyph := dimStyle.Render("·")
	row := dimStyle
	switch {
	case m.phase == phaseTrash && i == m.idx:
		glyph = cursorStyle.Render("▸")
		row = selectedName
	case processed && i < len(m.results) && m.results[i].Err != nil:
		glyph = errStyle.Render("✗")
		row = nameStyle
	case processed:
		glyph = checkStyle.Render("✓")
		row = nameStyle
	}
	return row.Render(fmt.Sprintf("  %s %-11s %-34s %9s",
		glyph, it.Type, name, human.Bytes(it.Size)))
}

func (m *procModel) renderSummary() string {
	var freed int64
	moved, failed := 0, 0
	for i, r := range m.results {
		if r.Err != nil {
			failed++
			continue
		}
		moved++
		freed += m.items[i].Size
	}
	line := footerStyle.Render(fmt.Sprintf("moved %d · verified %d · ~%s freed",
		moved, m.verified, human.Bytes(freed)))
	if m.failedItem > 0 || failed > 0 {
		line += "  " + errStyle.Render(fmt.Sprintf("%d not removed", m.failedItem+failed))
	}
	return line + "  " + helpStyle.Render("enter to exit")
}

func joinNames(names []string) string {
	if len(names) == 0 {
		return "—"
	}
	out := make([]string, len(names))
	for i, n := range names {
		out[i] = truncate(n, 18)
	}
	return strings.Join(out, ", ")
}

// RunProcessing runs the processing view for the given apps/items and returns
// the per-item trash results.
func RunProcessing(appNames, running []string, items []scan.Match) ([]trash.Result, error) {
	m := newProcModel(appNames, running, items, false)
	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return nil, err
	}
	return final.(*procModel).results, nil
}
