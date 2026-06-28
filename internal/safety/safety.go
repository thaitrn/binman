// Package safety enforces deletion safety: SIP-protected paths are never
// deleted, Apple first-party bundle ids are excluded, and running apps can be
// detected and gracefully quit.
package safety

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// forbiddenRoots are SIP-protected / system locations binman must never delete.
// Touching these can brick the OS; they are hard-refused regardless of flags.
var forbiddenRoots = []string{
	"/System",
	"/usr",
	"/bin",
	"/sbin",
	"/private/var/db",
	"/private/var/log",
	"/Library/Preferences/SystemConfiguration",
}

// IsForbidden reports whether p is under a SIP-protected root.
func IsForbidden(p string) bool {
	clean := filepath.Clean(p)
	for _, root := range forbiddenRoots {
		if clean == root || strings.HasPrefix(clean, root+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

// IsAppleBundleID reports whether bid is an Apple first-party id (com.apple.*).
// Used to avoid matching Apple framework/shared files during scans and cleans.
func IsAppleBundleID(bid string) bool {
	return bid == "com.apple" || strings.HasPrefix(bid, "com.apple.")
}

// IsAppRunning reports whether a process matching name is currently running.
func IsAppRunning(name string) bool {
	if name == "" {
		return false
	}
	out, err := exec.Command("pgrep", "-x", name).Output()
	if err != nil {
		return false // pgrep exits non-zero when no match -> not running
	}
	return strings.TrimSpace(string(out)) != ""
}

// QuitApp asks the app to quit gracefully via AppleScript. It is a no-op (with
// no error) if the app is not running.
func QuitApp(name string) error {
	if name == "" {
		return nil
	}
	// osascript errors if the app is not running; that is expected, ignore it.
	_ = exec.Command("osascript", "-e", fmt.Sprintf(`quit app %q`, name)).Run()
	return nil
}
