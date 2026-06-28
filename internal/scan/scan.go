// Package scan finds the leftover files a macOS app scatters across
// ~/Library (and /Library), keyed by bundle identifier and app name.
package scan

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/thaitrn/binman/internal/app"
)

// Type classifies a leftover match.
type Type string

const (
	TypeApp            Type = "app"
	TypeSupport        Type = "support"
	TypeCache          Type = "cache"
	TypePreference     Type = "preference"
	TypeContainer      Type = "container"
	TypeGroupContainer Type = "group"
	TypeSavedState     Type = "savedstate"
	TypeLog            Type = "log"
	TypeAgent          Type = "agent"
	TypeHTTP           Type = "http"
	TypeWebKit         Type = "webkit"
)

// matchMode controls how a directory entry name is matched to the target app.
type matchMode int

const (
	modeExactBID matchMode = iota // name == bid or name starts with "bid."
	modeBIDOrName                 // exact bid, or contains app name (len>=3)
	modeName                      // contains app name (len>=3)
	modeVendor                    // name starts with "<vendor>."
)

// Match is one discovered leftover file or directory.
type Match struct {
	Path      string
	Type      Type
	Size      int64
	NeedsSudo bool
	Shared    bool // group container shared across apps (uncheck by default)
}

// Scan finds leftovers for a using the real user home and system root.
func Scan(a *app.App) ([]Match, error) {
	home, _ := os.UserHomeDir()
	return ScanIn(a, home, "")
}

// ScanIn is the testable core: home is the user home; sysRoot is "" (real
// system) or a sandbox root so tests never touch the real /Library.
func ScanIn(a *app.App, home, sysRoot string) ([]Match, error) {
	vendor := vendorPrefix(a.BundleID)

	appSize, _ := DirSize(a.Path)
	out := []Match{{Path: a.Path, Type: TypeApp, Size: appSize}}
	seen := map[string]bool{normalize(a.Path): true}

	for _, loc := range locations(home, sysRoot) {
		names, err := readDirNames(loc.base)
		if err != nil {
			continue // missing/unreadable dir is not an error
		}
		for _, name := range names {
			if !matchEntry(loc.mode, name, a.BundleID, a.Name, vendor) {
				continue
			}
			full := filepath.Join(loc.base, name)
			key := normalize(full)
			if seen[key] {
				continue
			}
			seen[key] = true
			size, _ := DirSize(full)
			m := Match{Path: full, Type: loc.typ, Size: size, NeedsSudo: loc.sudo}
			if loc.typ == TypeGroupContainer {
				m.Shared = true // reported but not selected by default
			}
			out = append(out, m)
		}
	}
	return out, nil
}

// matchEntry decides whether a directory entry belongs to the target app.
func matchEntry(mode matchMode, name, bid, appName, vendor string) bool {
	low := strings.ToLower(name)
	switch mode {
	case modeExactBID:
		return name == bid || strings.HasPrefix(name, bid+".")
	case modeBIDOrName:
		if name == bid {
			return true
		}
		return len(appName) >= 3 && strings.Contains(low, strings.ToLower(appName))
	case modeName:
		return len(appName) >= 3 && strings.Contains(low, strings.ToLower(appName))
	case modeVendor:
		return vendor != "" && strings.HasPrefix(name, vendor+".")
	}
	return false
}

// vendorPrefix returns the first two dot-segments of a bundle id ("com.vendor"),
// used to match LaunchAgents/LaunchDaemons and group containers.
func vendorPrefix(bid string) string {
	parts := strings.SplitN(bid, ".", 3)
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return bid
}

// DirSize sums file sizes recursively; unreadable entries are skipped silently.
// Exported so other packages (e.g. clean) can reuse it.
func DirSize(path string) (int64, error) {
	var total int64
	err := filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if fi, e := d.Info(); e == nil {
			total += fi.Size()
		}
		return nil
	})
	return total, err
}

func readDirNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names, nil
}

// normalize cleans a path and resolves symlinks (macOS /tmp -> /private/tmp).
func normalize(p string) string {
	c := filepath.Clean(p)
	if r, err := filepath.EvalSymlinks(c); err == nil {
		return r
	}
	return c
}
