package scan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thaitrn/binman/internal/app"
)

// writeFile writes data into rel (creating parent dirs) under root.
func writeFile(t *testing.T, root, rel string, data []byte) {
	t.Helper()
	full := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", rel, err)
	}
	if err := os.WriteFile(full, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}

// buildFakeLibrary creates a temp user home with known leftovers + a decoy,
// and returns the home path and the fake app descriptor.
func buildFakeLibrary(t *testing.T) (home, sysRoot string, a *app.App) {
	t.Helper()
	home = t.TempDir()
	// the .app bundle itself (must exist for dirSize to work)
	writeFile(t, home, "FakeApp.app/Contents/FakeApp", []byte("binary"))

	// matching leftovers (user domain)
	writeFile(t, home, "Library/Application Support/com.example.fakeapp/state.db", make([]byte, 2048))
	writeFile(t, home, "Library/Caches/com.example.fakeapp/cache.bin", make([]byte, 1024))
	writeFile(t, home, "Library/Preferences/com.example.fakeapp.plist", []byte("plist"))
	writeFile(t, home, "Library/Saved Application State/com.example.fakeapp.savedState/data", []byte("x"))
	writeFile(t, home, "Library/Logs/FakeApp/log.txt", []byte("log"))
	writeFile(t, home, "Library/Containers/com.example.fakeapp/Data/x", []byte("c"))
	writeFile(t, home, "Library/Group Containers/group.com.example.fakeapp.suite/s", []byte("g"))
	writeFile(t, home, "Library/LaunchAgents/com.example.Helper.plist", []byte("agent"))

	// decoys that must NOT match
	writeFile(t, home, "Library/Caches/com.other.app/decoy", []byte("nope"))
	writeFile(t, home, "Library/Application Support/Unrelated/decoy", []byte("nope"))

	sysRoot = t.TempDir() // empty -> system /Library entries are skipped
	a = &app.App{Path: filepath.Join(home, "FakeApp.app"), BundleID: "com.example.fakeapp", Name: "FakeApp"}
	return home, sysRoot, a
}

func TestScanIn_FindsLeftovers(t *testing.T) {
	home, sysRoot, a := buildFakeLibrary(t)
	matches, err := ScanIn(a, home, sysRoot)
	if err != nil {
		t.Fatalf("ScanIn: %v", err)
	}
	got := map[Type]bool{}
	for _, m := range matches {
		got[m.Type] = true
	}
	for _, want := range []Type{TypeApp, TypeSupport, TypeCache, TypePreference, TypeSavedState, TypeLog, TypeContainer, TypeGroupContainer, TypeAgent} {
		if !got[want] {
			t.Errorf("expected match of type %q, not found", want)
		}
	}
}

func TestScanIn_NoFalsePositive(t *testing.T) {
	home, sysRoot, a := buildFakeLibrary(t)
	matches, _ := ScanIn(a, home, sysRoot)
	for _, m := range matches {
		if m.Type != TypeApp && containsSegment(m.Path, "com.other.app") {
			t.Errorf("decoy matched unexpectedly: %s", m.Path)
		}
		if containsSegment(m.Path, "Unrelated") {
			t.Errorf("unrelated dir matched: %s", m.Path)
		}
	}
}

func TestScanIn_GroupContainerFlaggedShared(t *testing.T) {
	home, sysRoot, a := buildFakeLibrary(t)
	matches, _ := ScanIn(a, home, sysRoot)
	for _, m := range matches {
		if m.Type == TypeGroupContainer && !m.Shared {
			t.Errorf("group container %s should be Shared", m.Path)
		}
	}
}

func TestScanIn_SystemPathsFlagSudo(t *testing.T) {
	// Put a leftover in the sandboxed system root and confirm NeedsSudo is set.
	home, sysRoot, a := buildFakeLibrary(t)
	writeFile(t, sysRoot, "Library/LaunchDaemons/com.example.daemon.plist", []byte("d"))
	matches, _ := ScanIn(a, home, sysRoot)
	found := false
	for _, m := range matches {
		if m.Type == TypeAgent && m.NeedsSudo {
			found = true
		}
	}
	if !found {
		t.Error("expected a system (NeedsSudo) launch daemon match")
	}
}

func TestScanIn_Sizes(t *testing.T) {
	home, sysRoot, a := buildFakeLibrary(t)
	matches, _ := ScanIn(a, home, sysRoot)
	for _, m := range matches {
		if m.Type == TypeCache && m.Size != 1024 {
			t.Errorf("cache size = %d, want 1024", m.Size)
		}
		if m.Type == TypeSupport && m.Size != 2048 {
			t.Errorf("support size = %d, want 2048", m.Size)
		}
	}
}

// containsSegment reports whether seg appears as a substring of path.
func containsSegment(path, seg string) bool {
	return strings.Contains(path, seg)
}
