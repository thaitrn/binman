// Package apps enumerates installed macOS applications for browsing/picking.
package apps

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/thaitrn/binman/internal/app"
	"github.com/thaitrn/binman/internal/safety"
	"github.com/thaitrn/binman/internal/scan"
)

// Entry is an installed app with its on-disk size and protection flag.
type Entry struct {
	App       *app.App
	Size      int64
	Protected bool // SIP-protected (e.g. Apple system app); not removable
}

// DefaultDirs lists where user-installed apps live. /System apps are excluded:
// they are SIP-protected and cannot be removed.
func DefaultDirs() []string {
	home, _ := os.UserHomeDir()
	return []string{"/Applications", filepath.Join(home, "Applications")}
}

// List enumerates .app bundles under dirs and returns them sorted by name.
// Enumeration combines a readdir recursion (reliable for top-level + nested
// suite apps, includes symlinked .app leaves) with a Spotlight (mdfind) safety
// net, because macOS /Applications readdir can be inconsistent across APIs.
func List(dirs []string) []Entry {
	seen := make(map[string]bool)
	var out []Entry
	for _, p := range discover(dirs) {
		rp, err := filepath.EvalSymlinks(p)
		if err != nil {
			rp = filepath.Clean(p)
		}
		if seen[rp] {
			continue
		}
		seen[rp] = true
		a, err := app.Resolve(p)
		if err != nil {
			continue // not a valid bundle
		}
		size, _ := scan.DirSize(p)
		protected := safety.IsForbidden(rp) || safety.IsForbidden(filepath.Clean(p))
		out = append(out, Entry{App: a, Size: size, Protected: protected})
	}
	sort.Slice(out, func(i, j int) bool {
		return strings.ToLower(out[i].App.Name) < strings.ToLower(out[j].App.Name)
	})
	return out
}

// discover returns candidate .app paths via readdir recursion of dirs plus a
// Spotlight query, filtered to real app bundles under one of the requested roots.
func discover(dirs []string) []string {
	set := make(map[string]bool)
	add := func(p string) {
		if filepath.Ext(filepath.Base(p)) != ".app" {
			return
		}
		if strings.Contains(p, "/Contents/") {
			return // helper app inside another bundle
		}
		if !underAny(p, dirs) {
			return // keep enumeration scoped to requested roots
		}
		set[filepath.Clean(p)] = true
	}
	for _, d := range dirs {
		collect(d, add)
	}
	for _, p := range mdfindApps() {
		add(p)
	}
	out := make([]string, 0, len(set))
	for p := range set {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

// underAny reports whether p is within one of dirs.
func underAny(p string, dirs []string) bool {
	cp := filepath.Clean(p)
	for _, d := range dirs {
		cd := filepath.Clean(d)
		if cp == cd || strings.HasPrefix(cp, cd+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

// collect recursively lists .app under root using os.ReadDir (filepath.WalkDir
// truncates on macOS /Applications). Descends only into real (non-symlink)
// directories; records .app leaves including symlinks.
func collect(root string, add func(string)) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return
	}
	for _, e := range entries {
		name := e.Name()
		path := filepath.Join(root, name)
		if filepath.Ext(name) == ".app" {
			add(path) // leaf bundle (incl. symlinked apps)
			continue
		}
		if e.IsDir() {
			collect(path, add) // real subfolder (suite installers, etc.)
		}
		// symlinked folders are not followed (avoids scanning /System)
	}
}

// mdfindApps queries Spotlight for application bundles (safety net).
func mdfindApps() []string {
	out, err := exec.Command("mdfind", "kMDItemKind == 'Application'").Output()
	if err != nil {
		return nil
	}
	var apps []string
	for _, line := range strings.Split(string(out), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			apps = append(apps, line)
		}
	}
	return apps
}
