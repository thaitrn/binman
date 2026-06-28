package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/thaitrn/binman/internal/scan"
)

func TestConfirmModel_ViewRenders(t *testing.T) {
	matches := []scan.Match{
		{Path: "/Applications/Foo.app", Type: scan.TypeApp, Size: 1024},
		{Path: "/Users/x/Library/Caches/com.example.foo", Type: scan.TypeCache, Size: 2048},
		{Path: "/Users/x/Library/Group Containers/group.com.example.foo", Type: scan.TypeGroupContainer, Size: 512, Shared: true},
	}
	m := newConfirmModel("Uninstall Foo", matches, true)
	m.width = 100
	out := m.View()
	for _, want := range []string{"Uninstall Foo", "TYPE", "SIZE", "PATH", "shared", "selected"} {
		if !strings.Contains(out, want) {
			t.Errorf("View() missing %q\n---\n%s", want, out)
		}
	}
}

func TestConfirmModel_DefaultSelectionSkipsShared(t *testing.T) {
	matches := []scan.Match{
		{Path: "/a", Type: scan.TypeApp, Size: 1},
		{Path: "/g", Type: scan.TypeGroupContainer, Size: 1, Shared: true},
	}
	m := newConfirmModel("t", matches, true)
	if !m.items[0].selected {
		t.Error("non-shared item should be selected by default")
	}
	if m.items[1].selected {
		t.Error("shared item should be unselected by default")
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("hello", 10); got != "hello" {
		t.Errorf("truncate short = %q", got)
	}
	if got := truncate("helloworld", 5); got != "hell…" {
		t.Errorf("truncate long = %q", got)
	}
}

// TestConfirmModel_KeyHandling drives Update() with real KeyMsgs to verify the
// navigation/toggle/confirm/cancel logic without a live TTY.
func TestConfirmModel_KeyHandling(t *testing.T) {
	m := newConfirmModel("t", []scan.Match{
		{Path: "/a", Type: scan.TypeApp, Size: 1},
		{Path: "/b", Type: scan.TypeCache, Size: 1},
	}, true)
	if !m.items[0].selected || !m.items[1].selected {
		t.Fatal("both should be selected by default")
	}

	// move down to item 1
	m = step(m, tea.KeyMsg{Type: tea.KeyDown}).(*confirmModel)
	if m.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", m.cursor)
	}
	// space toggles item 1 off
	m = step(m, tea.KeyMsg{Type: tea.KeySpace}).(*confirmModel)
	if m.items[1].selected {
		t.Fatal("item 1 should be toggled off")
	}
	// 'a' toggles all: some selected -> deselect all
	m = step(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}).(*confirmModel)
	if m.items[0].selected || m.items[1].selected {
		t.Fatal("'a' should deselect all when some are selected")
	}
	// 'a' again: none selected -> select all
	m = step(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}).(*confirmModel)
	if !m.items[0].selected || !m.items[1].selected {
		t.Fatal("'a' should select all when none are selected")
	}
	// enter confirms and returns a quit command
	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = m2.(*confirmModel)
	if !m.confirmed {
		t.Fatal("should be confirmed after enter")
	}
	if cmd == nil {
		t.Fatal("expected a quit command after confirm")
	}

	// q cancels
	mq := newConfirmModel("t", []scan.Match{{Path: "/a", Size: 1}}, true)
	if mq2 := step(mq, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}).(*confirmModel); !mq2.canceled {
		t.Fatal("q should cancel")
	}
}

func step(m tea.Model, msg tea.Msg) tea.Model {
	out, _ := m.Update(msg)
	return out
}
