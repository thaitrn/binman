package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/thaitrn/binman/internal/safety"
	"github.com/thaitrn/binman/internal/scan"
)

// humanBytes formats a byte count in binary units (KiB, MiB, ...).
func humanBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

// totalSize sums the Size of all matches.
func totalSize(ms []scan.Match) int64 {
	var s int64
	for _, m := range ms {
		s += m.Size
	}
	return s
}

// splitMatches separates actionable (user-domain, not SIP-protected) leftovers
// from system/sudo or protected ones, which the MVP reports but does not remove.
func splitMatches(ms []scan.Match) (actionable, system []scan.Match) {
	for _, m := range ms {
		switch {
		case m.NeedsSudo, isProtected(m):
			system = append(system, m)
		default:
			actionable = append(actionable, m)
		}
	}
	return actionable, system
}

// isProtected reports whether a match's path is SIP-protected.
func isProtected(m scan.Match) bool {
	return safety.IsForbidden(m.Path)
}

// printMatches prints an aligned table of actionable leftovers, then any
// system/sudo leftovers that will be skipped.
func printMatches(w *tabwriter.Writer, actionable, system []scan.Match) {
	fmt.Fprintf(w, "TYPE\tSIZE\tPATH\n")
	for _, m := range actionable {
		flag := ""
		if m.Shared {
			flag = " (shared)"
		}
		fmt.Fprintf(w, "%s%s\t%s\t%s\n", m.Type, flag, humanBytes(m.Size), m.Path)
	}
	if len(system) > 0 {
		fmt.Fprintf(w, "\t\t\n-- skipped (system / sudo / protected) --\t\t\n")
		for _, m := range system {
			fmt.Fprintf(w, "%s\t%s\t%s\n", m.Type, humanBytes(m.Size), m.Path)
		}
	}
}
