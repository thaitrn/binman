package tui

import (
	"strings"
	"testing"

	"github.com/thaitrn/binman/internal/scan"
	"github.com/thaitrn/binman/internal/trash"
)

func TestProcModel_ViewRenders(t *testing.T) {
	items := []scan.Match{
		{Path: "/Applications/Alpha.app", Type: scan.TypeApp, Size: 1024},
		{Path: "/Users/x/Library/Caches/com.x.alpha", Type: scan.TypeCache, Size: 2048},
	}
	m := newProcModel([]string{"Alpha"}, []string{"Alpha"}, items, false)
	m.width, m.height = 90, 24
	m.progress.Width = 86
	out := m.View()
	for _, want := range []string{"Processing", "Quit apps", "Trash", "Verify", "Done", "Alpha.app"} {
		if !strings.Contains(out, want) {
			t.Errorf("proc View() missing %q\n%s", want, out)
		}
	}
}

func TestProcModel_PhaseStates(t *testing.T) {
	if phaseState(phaseQuit, phaseQuit) != 1 {
		t.Error("current phase should be active (1)")
	}
	if phaseState(phaseTrash, phaseQuit) != 2 {
		t.Error("past phase should be done (2)")
	}
	if phaseState(phaseTrash, phaseVerify) != 0 {
		t.Error("future phase should be pending (0)")
	}
}

func TestProcModel_Summary(t *testing.T) {
	items := []scan.Match{
		{Path: "/a", Size: 1024},
		{Path: "/b", Size: 2048},
	}
	m := newProcModel([]string{"A"}, nil, items, false)
	m.results = []trash.Result{{Path: "/a"}, {Path: "/b"}}
	m.verified = 2
	m.phase = phaseDone
	m.done = true
	out := m.renderSummary()
	if !strings.Contains(out, "moved 2") || !strings.Contains(out, "verified 2") {
		t.Errorf("summary missing totals: %s", out)
	}
}
