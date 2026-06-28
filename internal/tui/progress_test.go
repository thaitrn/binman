package tui

import (
	"strings"
	"testing"

	"github.com/thaitrn/binman/internal/scan"
	"github.com/thaitrn/binman/internal/trash"
)

func TestProgressModel_InProgressView(t *testing.T) {
	m := newProgressModel([]scan.Match{{Path: "/a", Size: 1024}, {Path: "/b", Size: 2048}}, false)
	m.width = 60
	m.progress.Width = 56
	m.results = []trash.Result{{Path: "/a"}} // 1 of 2 done
	out := m.View()
	if !strings.Contains(out, "1/2") {
		t.Errorf("in-progress view missing count:\n%s", out)
	}
}

func TestProgressModel_SummaryView(t *testing.T) {
	m := newProgressModel([]scan.Match{{Path: "/a", Size: 1024}, {Path: "/b", Size: 2048}}, false)
	m.done = true
	m.results = []trash.Result{{Path: "/a"}, {Path: "/b"}}
	out := m.View()
	if !strings.Contains(out, "Done") || !strings.Contains(out, "freed") {
		t.Errorf("summary view missing text:\n%s", out)
	}
	if !strings.Contains(out, "3.0 KiB") { // 1024+2048
		t.Errorf("summary freed size wrong:\n%s", out)
	}
}
