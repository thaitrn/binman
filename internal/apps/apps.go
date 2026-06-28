// Package apps enumerates installed macOS applications for browsing/picking.
package apps

import (
	"os"
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
// they are SIP-protected and cannot be removed anyway.
func DefaultDirs() []string {
	home, _ := os.UserHomeDir()
	return []string{
		"/Applications",
		"/Applications/Utilities",
		filepath.Join(home, "Applications"),
	}
}

// List enumerates .app bundles under dirs, resolving identity (bundle id +
// name) and on-disk size. Results are sorted by name (case-insensitive);
// symlinks are de-duplicated by their resolved path.
func List(dirs []string) []Entry {
	seen := make(map[string]bool)
	var out []Entry
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() || filepath.Ext(e.Name()) != ".app" {
				continue
			}
			path := filepath.Join(dir, e.Name())
			rp, err := filepath.EvalSymlinks(path)
			if err != nil {
				rp = path
			}
			if seen[rp] {
				continue
			}
			seen[rp] = true
			a, err := app.Resolve(path)
			if err != nil {
				continue // not a valid bundle (no Info.plist, etc.)
			}
			size, _ := scan.DirSize(path)
			out = append(out, Entry{App: a, Size: size, Protected: safety.IsForbidden(path)})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return strings.ToLower(out[i].App.Name) < strings.ToLower(out[j].App.Name)
	})
	return out
}
