package apps

import (
	"os"
	"path/filepath"
	"testing"
)

func mkApp(t *testing.T, dir, appdir, bid, name string, sizeBytes int) {
	t.Helper()
	contents := filepath.Join(dir, appdir, "Contents")
	if err := os.MkdirAll(contents, 0o755); err != nil {
		t.Fatal(err)
	}
	plist := `<?xml version="1.0" encoding="UTF-8"?><plist version="1.0"><dict>` +
		`<key>CFBundleIdentifier</key><string>` + bid + `</string>` +
		`<key>CFBundleName</key><string>` + name + `</string>` +
		`</dict></plist>`
	if err := os.WriteFile(filepath.Join(contents, "Info.plist"), []byte(plist), 0o644); err != nil {
		t.Fatal(err)
	}
	if sizeBytes > 0 {
		if err := os.WriteFile(filepath.Join(contents, "payload"), make([]byte, sizeBytes), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestList(t *testing.T) {
	dir := t.TempDir()
	mkApp(t, dir, "Beta.app", "com.x.beta", "Beta", 200)
	mkApp(t, dir, "Alpha.app", "com.x.alpha", "Alpha", 100)
	// invalid bundle (no Info.plist) -> skipped
	if err := os.MkdirAll(filepath.Join(dir, "Broken.app", "Contents"), 0o755); err != nil {
		t.Fatal(err)
	}
	// non-.app entry -> ignored
	if err := os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	entries := List([]string{dir})
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2: %+v", len(entries), entries)
	}
	if entries[0].App.Name != "Alpha" || entries[1].App.Name != "Beta" {
		t.Errorf("not sorted: %s, %s", entries[0].App.Name, entries[1].App.Name)
	}
	// sizes computed
	var alpha *Entry
	for i := range entries {
		if entries[i].App.Name == "Alpha" {
			alpha = &entries[i]
		}
	}
	if alpha == nil || alpha.Size < 100 {
		t.Errorf("Alpha size = %v", alpha)
	}
}

func TestList_MissingDir(t *testing.T) {
	entries := List([]string{"/does/not/exist/xyz"})
	if len(entries) != 0 {
		t.Errorf("expected 0 entries for missing dir, got %d", len(entries))
	}
}
