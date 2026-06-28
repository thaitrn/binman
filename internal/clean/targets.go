// Package clean defines regenerable "system junk" targets (caches, logs, Xcode
// artifacts, package-manager caches) and plans/applies their removal.
package clean

import (
	"os"
	"os/exec"
	"path/filepath"
)

// Kind distinguishes path-based (Trash) targets from command-based targets.
type Kind int

const (
	KindTrash Kind = iota
	KindCommand
)

// Target describes one cleanable category.
type Target struct {
	ID, Name, Group string
	Kind            Kind
	Paths           func(home string) []string           // KindTrash candidate paths
	Cmd             func(home string) (string, []string) // KindCommand binary + args
}

// Groups is the ordered list of selectable groups.
var Groups = []string{"default", "downloads", "xcode", "pkg"}

// Targets returns all clean targets. Group "default" (caches, logs) is on by
// default; the rest are opt-in via flags.
func Targets() []Target {
	lib := func(home, sub string) string { return filepath.Join(home, "Library", sub) }
	return []Target{
		{ID: "caches", Name: "User caches", Group: "default", Kind: KindTrash,
			Paths: func(h string) []string { return listDir(lib(h, "Caches")) }},
		{ID: "logs", Name: "User logs", Group: "default", Kind: KindTrash,
			Paths: func(h string) []string { return listDir(lib(h, "Logs")) }},

		{ID: "downloads", Name: "Old installers (Downloads)", Group: "downloads", Kind: KindTrash,
			Paths: downloadsJunk},

		{ID: "xcode-derived", Name: "Xcode DerivedData", Group: "xcode", Kind: KindTrash,
			Paths: func(h string) []string { return listDir(filepath.Join(lib(h, "Developer"), "Xcode", "DerivedData")) }},
		{ID: "xcode-devices", Name: "Xcode Device Support", Group: "xcode", Kind: KindTrash,
			Paths: xcodeDevices},
		{ID: "xcode-sim", Name: "Xcode stale simulators", Group: "xcode", Kind: KindCommand,
			Cmd: func(string) (string, []string) { return "xcrun", []string{"simctl", "delete", "unavailable"} }},

		{ID: "brew", Name: "Homebrew cleanup", Group: "pkg", Kind: KindCommand,
			Cmd: func(string) (string, []string) { return "brew", []string{"cleanup", "-s"} }},
		{ID: "npm", Name: "npm cache", Group: "pkg", Kind: KindCommand,
			Cmd: func(string) (string, []string) { return "npm", []string{"cache", "clean", "--force"} }},
		{ID: "pnpm", Name: "pnpm store", Group: "pkg", Kind: KindCommand,
			Cmd: func(string) (string, []string) { return "pnpm", []string{"store", "prune"} }},
		{ID: "pip", Name: "pip cache", Group: "pkg", Kind: KindCommand,
			Cmd: func(string) (string, []string) { return "pip", []string{"cache", "purge"} }},
		{ID: "docker", Name: "docker prune", Group: "pkg", Kind: KindCommand,
			Cmd: func(string) (string, []string) { return "docker", []string{"system", "prune", "-f"} }},
	}
}

// downloadsJunk lists installer/archive files in ~/Downloads.
func downloadsJunk(home string) []string {
	dl := filepath.Join(home, "Downloads")
	var out []string
	for _, ext := range []string{".dmg", ".pkg", ".zip", ".iso"} {
		m, _ := filepath.Glob(filepath.Join(dl, "*"+ext))
		out = append(out, m...)
	}
	return out
}

func xcodeDevices(home string) []string {
	base := filepath.Join(home, "Library", "Developer", "Xcode")
	var out []string
	for _, sub := range []string{"iOS DeviceSupport", "watchOS DeviceSupport", "tvOS DeviceSupport"} {
		out = append(out, listDir(filepath.Join(base, sub))...)
	}
	return out
}

// listDir returns absolute paths of each entry under dir (missing dir -> nil).
func listDir(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, filepath.Join(dir, e.Name()))
	}
	return out
}

// available reports whether name is on PATH.
func available(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
