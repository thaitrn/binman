package app

import (
	"os"
	"path/filepath"
	"testing"
)

// minimalXMLPlist is a valid XML property list used as a fake Info.plist.
const minimalXMLPlist = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
	<key>CFBundleIdentifier</key><string>com.example.fakeapp</string>
	<key>CFBundleName</key><string>FakeApp</string>
	<key>CFBundleExecutable</key><string>FakeApp</string>
</dict></plist>`

// writeFakeApp creates a temp apps dir containing FakeApp.app with an Info.plist.
func writeFakeApp(t *testing.T) (appsDir, appPath string) {
	t.Helper()
	appsDir = t.TempDir()
	appPath = filepath.Join(appsDir, "FakeApp.app", "Contents")
	if err := os.MkdirAll(appPath, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(appPath, "Info.plist"), []byte(minimalXMLPlist), 0o644); err != nil {
		t.Fatalf("write plist: %v", err)
	}
	return appsDir, filepath.Join(appsDir, "FakeApp.app")
}

func TestResolveIn_ByName(t *testing.T) {
	appsDir, appPath := writeFakeApp(t)
	got, err := ResolveIn("FakeApp", []string{appsDir})
	if err != nil {
		t.Fatalf("ResolveIn: %v", err)
	}
	if got.Path != appPath {
		t.Errorf("Path = %q, want %q", got.Path, appPath)
	}
	if got.BundleID != "com.example.fakeapp" {
		t.Errorf("BundleID = %q, want com.example.fakeapp", got.BundleID)
	}
	if got.Name != "FakeApp" {
		t.Errorf("Name = %q, want FakeApp", got.Name)
	}
}

func TestResolveIn_ByPath(t *testing.T) {
	appsDir, appPath := writeFakeApp(t)
	got, err := ResolveIn(appPath, []string{appsDir})
	if err != nil {
		t.Fatalf("ResolveIn by path: %v", err)
	}
	if got.Path != appPath {
		t.Errorf("Path = %q, want %q", got.Path, appPath)
	}
}

func TestResolveIn_NotFound(t *testing.T) {
	appsDir, _ := writeFakeApp(t)
	if _, err := ResolveIn("DoesNotExist", []string{appsDir}); err == nil {
		t.Fatal("expected ErrAppNotFound, got nil")
	}
}

// TestResolveIn_NameFallback verifies the display name is derived from the
// bundle dir when CFBundleName is absent.
func TestResolveIn_NameFallback(t *testing.T) {
	appsDir := t.TempDir()
	contents := filepath.Join(appsDir, "NoName.app", "Contents")
	if err := os.MkdirAll(contents, 0o755); err != nil {
		t.Fatal(err)
	}
	noName := `<?xml version="1.0" encoding="UTF-8"?>
<plist version="1.0"><dict>
	<key>CFBundleIdentifier</key><string>com.example.noname</string>
</dict></plist>`
	if err := os.WriteFile(filepath.Join(contents, "Info.plist"), []byte(noName), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := ResolveIn("NoName", []string{appsDir})
	if err != nil {
		t.Fatalf("ResolveIn: %v", err)
	}
	if got.Name != "NoName" {
		t.Errorf("Name = %q, want NoName (derived)", got.Name)
	}
}
