package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/thaitrn/binman/internal/human"
	"github.com/thaitrn/binman/internal/safety"
	"github.com/thaitrn/binman/internal/scan"
)

const nameW = 22

// View renders the two-pane layout: title, [apps | details], status bar.
func (m *pickerModel) View() string {
	if m.width == 0 {
		return titleStyle.Render("◆ binman")
	}
	w, h := m.width, m.height

	// Title line.
	total := m.sumAll()
	title := titleStyle.Render("◆ binman") +
		dimStyle.Render(fmt.Sprintf("  %d apps · %s total", len(m.entries), human.Bytes(total)))

	bodyH := h - 3 // title(1) + status(2)
	if bodyH < 6 {
		bodyH = 6
	}
	leftW := w * 45 / 100
	if leftW > 56 {
		leftW = 56
	}
	if leftW < 40 {
		leftW = 40
	}
	if leftW > w-24 {
		leftW = w - 24
	}
	rightW := w - leftW - 1
	if rightW < 20 {
		rightW = 20
	}

	left := m.renderLeft(leftW, bodyH)
	right := m.renderRight(rightW, bodyH)
	body := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	sep := dimStyle.Render(strings.Repeat("─", w))
	status := m.renderStatus(w)
	return lipgloss.JoinVertical(lipgloss.Left, title, body, sep, status)
}

// renderLeft renders the bordered app list with size bars.
func (m *pickerModel) renderLeft(outerW, outerH int) string {
	contentW := outerW - 2
	contentH := outerH - 2
	barW := contentW - nameW - 15
	if barW < 4 {
		barW = 4
	}

	var rows []string
	hdr := "    " + padRight("APP", nameW) + " " + strings.Repeat(" ", barW) + " " + fmt.Sprintf("%9s", "SIZE")
	rows = append(rows, headerStyle.Render(hdr))

	vis := m.visible()
	for _, i := range vis {
		rows = append(rows, m.renderAppRow(i, contentW, barW))
	}
	// pad to content height
	for len(rows) < contentH {
		rows = append(rows, strings.Repeat(" ", contentW))
	}
	if len(rows) > contentH {
		rows = rows[:contentH]
	}
	inner := strings.Join(rows, "\n")
	return pane(inner, outerW, outerH)
}

func (m *pickerModel) renderAppRow(i, contentW, barW int) string {
	e := m.entries[i]
	marker := "  "
	nameSt := nameStyle
	if i == m.cursor {
		marker = cursorStyle.Render("▸ ")
		nameSt = selectedName
	}
	box := "☐"
	if e.Protected {
		box = sharedStyle.Render("⊘")
	} else if m.selected[e.App.Path] {
		box = checkStyle.Render("☑")
	}
	name := nameSt.Render(padRight(truncate(e.App.Name, nameW), nameW))
	bar := renderBar(e.Size, m.maxSize, barW)
	size := dimStyle.Render(fmt.Sprintf("%9s", human.Bytes(e.Size)))
	return marker + box + " " + name + " " + bar + " " + size
}

// renderRight renders the details + leftover preview for the highlighted app.
func (m *pickerModel) renderRight(outerW, outerH int) string {
	contentW := outerW - 2
	contentH := outerH - 2
	var b strings.Builder
	if len(m.entries) == 0 {
		b.WriteString(dimStyle.Render("No apps."))
		return pane(b.String(), outerW, outerH)
	}
	e := m.entries[m.cursor]
	b.WriteString(titleStyle.Render(truncate(e.App.Name, contentW)))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render(truncate(e.App.BundleID, contentW)))
	b.WriteString("\n")
	if e.Protected {
		b.WriteString(sharedStyle.Render("system app — cannot remove"))
	} else {
		b.WriteString(dimStyle.Render(truncate(e.App.Path, contentW)))
	}
	b.WriteString("\n")
	b.WriteString(dimStyle.Render(strings.Repeat("─", contentW)))
	b.WriteString("\n")

	// Leftover preview (from the async scan cache; "scanning…" until ready).
	leftovers := m.cache[e.App.Path]
	if leftovers == nil {
		b.WriteString(dimStyle.Render("  scanning leftovers…"))
		b.WriteString("\n")
		return pane(b.String(), outerW, outerH)
	}
	var act []scan.Match
	var shared int
	var sum int64
	for _, mm := range leftovers {
		if mm.NeedsSudo || safety.IsForbidden(mm.Path) {
			continue
		}
		if mm.Shared {
			shared++
			sum += mm.Size
			continue
		}
		act = append(act, mm)
		sum += mm.Size
	}
	b.WriteString(headerStyle.Render(fmt.Sprintf("Leftovers  ~%s", human.Bytes(sum))))
	b.WriteString("\n")
	max := contentH - 6
	if max < 3 {
		max = 3
	}
	shown := 0
	for _, mm := range act {
		if shown >= max {
			b.WriteString(dimStyle.Render(fmt.Sprintf("  + %d more", len(act)-shown)))
			break
		}
		b.WriteString(fmt.Sprintf("  %s %s",
			dimStyle.Render(padRight(string(mm.Type), 12)),
			dimStyle.Render(fmt.Sprintf("%8s", human.Bytes(mm.Size)))))
		b.WriteString("\n")
		shown++
	}
	if shared > 0 {
		b.WriteString(sharedStyle.Render(fmt.Sprintf("  + %d shared (off)", shared)))
		b.WriteString("\n")
	}
	if len(act) == 0 && shared == 0 {
		b.WriteString(dimStyle.Render("  clean — no leftovers"))
		b.WriteString("\n")
	}

	return pane(b.String(), outerW, outerH)
}

// renderStatus renders the bottom status line.
func (m *pickerModel) renderStatus(w int) string {
	left := footerStyle.Render(fmt.Sprintf("%d selected · ~%s",
		m.selectedCount(), human.Bytes(m.selectedSize())))
	keys := helpStyle.Render("↑↓/jk move · space toggle · a all · enter confirm · q/esc quit")
	gap := w - lipgloss.Width(left) - lipgloss.Width(keys)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + keys
}

// renderBar renders a proportional size bar: green fill on dim track.
func renderBar(size, max int64, w int) string {
	if max <= 0 {
		return dimStyle.Render(strings.Repeat("·", w))
	}
	filled := int(float64(w) * float64(size) / float64(max))
	if filled > w {
		filled = w
	}
	if filled < 0 {
		filled = 0
	}
	fill := lipgloss.NewStyle().Foreground(cGreen).Render(strings.Repeat("█", filled))
	track := dimStyle.Render(strings.Repeat("░", w-filled))
	return fill + track
}

// pane wraps content in a rounded border of the given outer size.
func pane(content string, outerW, outerH int) string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(cBorder).
		Width(outerW - 2).
		Height(outerH - 2).
		Render(content)
}

// visible returns the indices currently shown in the left pane.
func (m *pickerModel) visible() []int {
	h := m.listHeight()
	end := m.offset + h
	if end > len(m.entries) {
		end = len(m.entries)
	}
	out := make([]int, 0, h)
	for i := m.offset; i < end; i++ {
		out = append(out, i)
	}
	return out
}

func (m *pickerModel) sumAll() int64 {
	var s int64
	for _, e := range m.entries {
		s += e.Size
	}
	return s
}

func (m *pickerModel) selectedCount() int {
	c := 0
	for _, e := range m.entries {
		if m.selected[e.App.Path] {
			c++
		}
	}
	return c
}

func (m *pickerModel) selectedSize() int64 {
	var s int64
	for _, e := range m.entries {
		if m.selected[e.App.Path] {
			s += e.Size
		}
	}
	return s
}

// padRight pads s with spaces to width n (no truncation here; caller truncates).
func padRight(s string, n int) string {
	r := []rune(s)
	if len(r) >= n {
		return string(r[:n])
	}
	return s + strings.Repeat(" ", n-len(r))
}
