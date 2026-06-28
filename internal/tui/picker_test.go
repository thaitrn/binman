package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/thaitrn/binman/internal/app"
	"github.com/thaitrn/binman/internal/apps"
)

func TestPickerModel_ViewRenders(t *testing.T) {
	entries := []apps.Entry{
		{App: &app.App{Name: "Alpha", BundleID: "com.x.alpha", Path: "/Applications/Alpha.app"}, Size: 1024},
		{App: &app.App{Name: "Beta", BundleID: "com.x.beta", Path: "/Applications/Beta.app"}, Size: 2048, Protected: true},
	}
	m := newPickerModel(entries)
	m.width, m.height = 100, 24
	out := m.View()
	for _, want := range []string{"binman", "Alpha", "Beta", "APP", "system"} {
		if !strings.Contains(out, want) {
			t.Errorf("picker View() missing %q", want)
		}
	}
}

func TestPickerModel_Toggle(t *testing.T) {
	entries := []apps.Entry{
		{App: &app.App{Name: "Alpha", BundleID: "com.x.alpha", Path: "/a.app"}},
		{App: &app.App{Name: "Beta", BundleID: "com.x.beta", Path: "/b.app"}},
	}
	m := newPickerModel(entries)
	m.width, m.height = 100, 24

	// cursor at 0; space toggles Alpha on
	m = step(m, tea.KeyMsg{Type: tea.KeySpace}).(*pickerModel)
	if !m.selected["/a.app"] {
		t.Error("Alpha should be selected after space")
	}
	// 'a': not all selected -> select all
	m = step(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}).(*pickerModel)
	if !m.selected["/a.app"] || !m.selected["/b.app"] {
		t.Error("'a' should select all")
	}
	// 'a' again: all selected -> clear
	m = step(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}).(*pickerModel)
	if len(m.selected) != 0 {
		t.Error("'a' should clear all when all selected")
	}
}

func TestPickerModel_Navigation(t *testing.T) {
	entries := []apps.Entry{
		{App: &app.App{Name: "A", BundleID: "c.a", Path: "/a.app"}},
		{App: &app.App{Name: "B", BundleID: "c.b", Path: "/b.app"}},
		{App: &app.App{Name: "C", BundleID: "c.c", Path: "/c.app"}},
	}
	m := newPickerModel(entries)
	m.width, m.height = 80, 20
	if m.cursor != 0 {
		t.Fatalf("cursor = %d, want 0", m.cursor)
	}
	m = step(m, tea.KeyMsg{Type: tea.KeyDown}).(*pickerModel)
	if m.cursor != 1 {
		t.Errorf("cursor = %d, want 1 after down", m.cursor)
	}
	m = step(m, tea.KeyMsg{Type: tea.KeyDown}).(*pickerModel)
	m = step(m, tea.KeyMsg{Type: tea.KeyDown}).(*pickerModel) // past end -> stays at last
	if m.cursor != 2 {
		t.Errorf("cursor = %d, want 2 (clamped)", m.cursor)
	}
}

func TestSelectApps_Empty(t *testing.T) {
	got, ok, err := SelectApps(nil)
	if err != nil || ok || got != nil {
		t.Errorf("SelectApps(nil) = %+v ok=%v err=%v, want nil/false/nil", got, ok, err)
	}
}

func TestConfirmBatch_ViewRenders(t *testing.T) {
	m := newBatchConfirmModel([]string{"Alpha", "Beta"}, 12, 4096, 2)
	out := m.View()
	for _, want := range []string{"Uninstall 2 app(s)", "Alpha", "Beta", "leftover", "shared"} {
		if !strings.Contains(out, want) {
			t.Errorf("ConfirmBatch View() missing %q\n%s", want, out)
		}
	}
}

// step applies a message to a model and returns the result (test helper).
func step(m tea.Model, msg tea.Msg) tea.Model {
	out, _ := m.Update(msg)
	return out
}
