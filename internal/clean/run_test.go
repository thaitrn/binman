package clean

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFileB(t *testing.T, root, rel string, n int) {
	t.Helper()
	full := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, make([]byte, n), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestPlan_DefaultAndDownloads(t *testing.T) {
	home := t.TempDir()
	writeFileB(t, home, "Library/Caches/com.example/cache.bin", 1000)
	writeFileB(t, home, "Library/Logs/App/log", 500)
	writeFileB(t, home, "Downloads/setup.dmg", 2000)
	writeFileB(t, home, "Downloads/keep.txt", 999) // not junk
	writeFileB(t, home, "Downloads/archive.zip", 300)

	reps := Plan(home, []string{"default", "downloads"})
	byID := map[string]Report{}
	for _, r := range reps {
		byID[r.Target.ID] = r
	}
	if byID["caches"].Size != 1000 {
		t.Errorf("caches=%d want 1000", byID["caches"].Size)
	}
	if byID["logs"].Size != 500 {
		t.Errorf("logs=%d want 500", byID["logs"].Size)
	}
	dl := byID["downloads"]
	if len(dl.Paths) != 2 {
		t.Errorf("downloads paths=%d want 2 (.dmg,.zip)", len(dl.Paths))
	}
	if dl.Size != 2300 {
		t.Errorf("downloads size=%d want 2300", dl.Size)
	}
}

func TestPlan_GroupFilter(t *testing.T) {
	home := t.TempDir()
	writeFileB(t, home, "Library/Caches/com.example/c", 100)
	for _, r := range Plan(home, []string{"xcode"}) {
		if r.Target.Group != "xcode" {
			t.Errorf("non-xcode target %q included in xcode group", r.Target.ID)
		}
	}
}

func TestPlan_CommandReport(t *testing.T) {
	home := t.TempDir()
	var brew *Report
	for _, r := range Plan(home, []string{"pkg"}) {
		if r.Target.ID == "brew" {
			b := r
			brew = &b
		}
	}
	if brew == nil {
		t.Fatal("brew target missing from pkg group")
	}
	if brew.Command != "brew cleanup -s" {
		t.Errorf("brew command = %q", brew.Command)
	}
}
