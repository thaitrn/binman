package scan

import (
	"os"
	"path/filepath"
)

// location is one scanned directory plus the rule for matching its entries
// to the target app.
type location struct {
	base string // absolute directory to enumerate
	typ  Type
	mode matchMode
	sudo bool // true for /Library (system) entries
}

// locations builds the canonical leftover location list. home is the user home
// (for ~/Library); sysRoot is "" for the real system root or a sandbox root in
// tests (so tests never touch the real /Library).
func locations(home, sysRoot string) []location {
	user := func(sub string) string { return filepath.Join(home, "Library", sub) }
	sys := func(sub string) string { return systemPath(sysRoot, filepath.Join("Library", sub)) }
	return []location{
		// --- user domain (~/Library) ---
		{user("Application Support"), TypeSupport, modeBIDOrName, false},
		{user("Caches"), TypeCache, modeBIDOrName, false},
		{user("Preferences"), TypePreference, modeExactBID, false},
		{user("Saved Application State"), TypeSavedState, modeExactBID, false},
		{user("Logs"), TypeLog, modeName, false},
		{user("HTTPStorages"), TypeHTTP, modeExactBID, false},
		{user("WebKit"), TypeWebKit, modeBIDOrName, false},
		{user("Containers"), TypeContainer, modeExactBID, false},
		{user("Group Containers"), TypeGroupContainer, modeBIDOrName, false},
		{user("LaunchAgents"), TypeAgent, modeVendor, false},
		// --- system domain (/Library, needs sudo) ---
		{sys("Application Support"), TypeSupport, modeBIDOrName, true},
		{sys("Caches"), TypeCache, modeBIDOrName, true},
		{sys("LaunchAgents"), TypeAgent, modeVendor, true},
		{sys("LaunchDaemons"), TypeAgent, modeVendor, true},
		{sys("Preferences"), TypePreference, modeExactBID, true},
	}
}

// systemPath returns an absolute system path: the real root when sysRoot is "",
// otherwise rooted under the sandbox (used by tests).
func systemPath(sysRoot, sub string) string {
	if sysRoot == "" {
		return string(os.PathSeparator) + sub
	}
	return filepath.Join(sysRoot, sub)
}
