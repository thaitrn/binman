// Package app resolves a macOS .app bundle and extracts its identity
// (bundle identifier + display name) from its Info.plist.
package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"howett.net/plist"
)

// App is a resolved macOS application bundle.
type App struct {
	Path     string // absolute path to the .app bundle
	BundleID string // CFBundleIdentifier, e.g. "com.example.app"
	Name     string // CFBundleName, or the bundle dir name if absent
}

// ErrAppNotFound is returned when an app cannot be located by name or path.
type ErrAppNotFound struct{ Input string }

func (e *ErrAppNotFound) Error() string {
	return fmt.Sprintf("app not found: %q (looked in %s)", e.Input, strings.Join(defaultAppDirs(), ", "))
}

// Resolve finds a .app by name (searched in standard app dirs) or by path,
// then reads its identity. It is the production entry point.
func Resolve(input string) (*App, error) {
	return ResolveIn(input, defaultAppDirs())
}

// ResolveIn is like Resolve but with an explicit list of search directories
// (used by tests; first match wins).
func ResolveIn(input string, dirs []string) (*App, error) {
	path, err := locate(input, dirs)
	if err != nil {
		return nil, err
	}
	info, err := readInfo(path)
	if err != nil {
		return nil, fmt.Errorf("read Info.plist for %s: %w", path, err)
	}
	name := info.Name
	if name == "" {
		name = strings.TrimSuffix(filepath.Base(path), ".app")
	}
	return &App{Path: path, BundleID: info.BundleID, Name: name}, nil
}

// locate resolves input to an existing .app bundle path, either by treating it
// as a literal path or by searching dirs for <input>.app.
func locate(input string, dirs []string) (string, error) {
	// Literal path: accept if it exists and looks like a .app bundle.
	if strings.HasSuffix(input, ".app") {
		if _, err := os.Stat(input); err == nil {
			if abs, aerr := filepath.Abs(input); aerr == nil {
				return abs, nil
			}
			return input, nil
		}
	}
	// Search by name across the provided dirs.
	name := strings.TrimSuffix(input, ".app") + ".app"
	for _, dir := range dirs {
		candidate := filepath.Join(dir, name)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", &ErrAppNotFound{Input: input}
}

// plistInfo holds the subset of Info.plist keys binman cares about.
type plistInfo struct {
	BundleID   string `plist:"CFBundleIdentifier"`
	Name       string `plist:"CFBundleName"`
	Executable string `plist:"CFBundleExecutable"`
}

// readInfo decodes the bundle's Contents/Info.plist.
func readInfo(appPath string) (plistInfo, error) {
	var info plistInfo
	f, err := os.Open(filepath.Join(appPath, "Contents", "Info.plist"))
	if err != nil {
		return info, err
	}
	defer f.Close()
	if err := plist.NewDecoder(f).Decode(&info); err != nil {
		return info, err
	}
	return info, nil
}

// defaultAppDirs lists where macOS apps commonly live, user-domain first.
func defaultAppDirs() []string {
	home, _ := os.UserHomeDir()
	return []string{
		filepath.Join(home, "Applications"),
		"/Applications",
		"/Applications/Utilities",
		"/System/Applications",
		"/System/Applications/Utilities",
	}
}
