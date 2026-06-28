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
	m.list.SetWidth(100)
	out := m.View()
	for _, want := range []string{"Select apps", "Alpha", "Beta", "system"} {
		if !strings.Contains(out, want) {
			t.Errorf("picker View() missing %q\n---\n%s", want, out)
		}
	}
}

// TestPickerModel_Toggle verifies space toggles selection and 'a' toggles all.
func TestPickerModel_Toggle(t *testing.T) {
	entries := []apps.Entry{
		{App: &app.App{Name: "Alpha", BundleID: "com.x.alpha", Path: "/Applications/Alpha.app"}},
		{App: &app.App{Name: "Beta", BundleID: "com.x.beta", Path: "/Applications/Beta.app"}},
	}
	m := newPickerModel(entries)
	// cursor at Alpha (index 0); space toggles it on
	m = step(m, tea.KeyMsg{Type: tea.KeySpace}).(*pickerModel)
	if !m.selected["/Applications/Alpha.app"] {
		t.Error("Alpha should be selected after space")
	}
	// 'a' toggles all (none fully selected -> all on)
	m = step(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}).(*pickerModel)
	if !m.selected["/Applications/Alpha.app"] || !m.selected["/Applications/Beta.app"] {
		t.Error("'a' should select all")
	}
	// 'a' again -> all off
	m = step(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}).(*pickerModel)
	if m.selected["/Applications/Alpha.app"] || m.selected["/Applications/Beta.app"] {
		t.Error("'a' should deselect all when all selected")
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
			t.Errorf("ConfirmBatch View() missing %q\n---\n%s", want, out)
		}
	}
}

// step applies a message to a model and returns the result (test helper).
func step(m tea.Model, msg tea.Msg) tea.Model {
	out, _ := m.Update(msg)
	return out
}
