package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thaitrn/binman/internal/apps"
)

// writeFile creates root/rel (making parent dirs) with n zero bytes.
func writeFile(t *testing.T, root, rel string, n int) {
	t.Helper()
	full := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, make([]byte, n), 0o644); err != nil {
		t.Fatal(err)
	}
}

// mkFakeApp creates a valid .app bundle under root/Applications with an Info.plist.
func mkFakeApp(t *testing.T, root, name, bid string) {
	t.Helper()
	c := filepath.Join(root, "Applications", name+".app", "Contents")
	if err := os.MkdirAll(c, 0o755); err != nil {
		t.Fatal(err)
	}
	plist := `<?xml version="1.0" encoding="UTF-8"?><plist version="1.0"><dict>` +
		`<key>CFBundleIdentifier</key><string>` + bid + `</string>` +
		`<key>CFBundleName</key><string>` + name + `</string>` +
		`</dict></plist>`
	if err := os.WriteFile(filepath.Join(c, "Info.plist"), []byte(plist), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestFlow_ScanSelectedOnFakeHome exercises the full list → scan chain on a
// throwaway HOME: FakeApp + leftovers are enumerated, the decoy is excluded,
// and the shared container is counted as skipped (not actionable).
func TestFlow_ScanSelectedOnFakeHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home) // scan.Scan reads $HOME

	mkFakeApp(t, home, "FakeApp", "com.example.fakeapp")
	writeFile(t, home, "Library/Caches/com.example.fakeapp/cache.bin", 1234)
	writeFile(t, home, "Library/Application Support/com.example.fakeapp/state.db", 5678)
	writeFile(t, home, "Library/Preferences/com.example.fakeapp.plist", 1)
	writeFile(t, home, "Library/Group Containers/group.com.example.fakeapp/shared", 99)
	writeFile(t, home, "Library/Caches/com.other.app/decoy", 1) // must NOT match

	entries := apps.List([]string{filepath.Join(home, "Applications")})
	var selected []apps.Entry
	for _, e := range entries {
		if e.App.Name == "FakeApp" {
			selected = append(selected, e)
		}
	}
	if len(selected) != 1 {
		t.Fatalf("FakeApp not enumerated (got %d entries): %+v", len(entries), entries)
	}

	all, shared := scanSelected(selected)
	// Expect app bundle + caches + support + preference as actionable (>=4).
	if len(all) < 4 {
		t.Errorf("expected >=4 actionable matches, got %d: %+v", len(all), all)
	}
	// Shared group container is reported but not actionable.
	if shared != 1 {
		t.Errorf("expected 1 shared skipped, got %d", shared)
	}
	// Decoy must never appear.
	for _, m := range all {
		if strings.Contains(m.Path, "com.other.app") {
			t.Errorf("decoy matched unexpectedly: %s", m.Path)
		}
	}
}
