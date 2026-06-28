package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/x/term"

	"github.com/thaitrn/binman/internal/apps"
	"github.com/thaitrn/binman/internal/safety"
	"github.com/thaitrn/binman/internal/scan"
	"github.com/thaitrn/binman/internal/tui"
)

// runPickerFlow is binman's default flow:
// list → select many → confirm → process (Trash) → results.
func runPickerFlow() error {
	interactive := term.IsTerminal(os.Stdin.Fd()) && term.IsTerminal(os.Stdout.Fd())
	if !interactive {
		return errors.New("binman needs an interactive terminal — run it directly in your terminal")
	}

	// 1-2. List + multi-select.
	entries := apps.List(apps.DefaultDirs())
	if len(entries) == 0 {
		return errors.New("no apps found in /Applications")
	}
	selected, ok, err := tui.SelectApps(entries)
	if err != nil {
		return err
	}
	if !ok || len(selected) == 0 {
		fmt.Fprintln(os.Stderr, "aborted.")
		return nil
	}

	// 3. Scan aggregate leftovers (user-domain; shared excluded by default).
	all, shared := scanSelected(selected)
	names := appNames(selected)

	if dryRun {
		printBatchSummary(names, all, shared)
		fmt.Fprintln(os.Stderr, "\n(--dry-run) nothing deleted.")
		return nil
	}

	// 4. Confirm.
	if !apply {
		confirmed, cerr := tui.ConfirmBatch(names, len(all), totalSize(all), shared)
		if cerr != nil {
			return cerr
		}
		if !confirmed {
			fmt.Fprintln(os.Stderr, "aborted.")
			return nil
		}
	}

	// 5-6. Process (quit → trash → verify → done) + results.
	var running []string
	for _, e := range selected {
		if safety.IsAppRunning(e.App.Name) {
			running = append(running, e.App.Name)
		}
	}
	if len(all) == 0 && len(running) == 0 {
		fmt.Fprintln(os.Stderr, "no removable leftovers; nothing to do.")
		return nil
	}
	if _, err := tui.RunProcessing(names, running, all); err != nil {
		return err
	}
	return nil
}

// scanSelected scans each app and aggregates user-domain, non-shared matches.
func scanSelected(selected []apps.Entry) (all []scan.Match, shared int) {
	for _, e := range selected {
		ms, _ := scan.Scan(e.App)
		for _, m := range ms {
			if m.NeedsSudo || safety.IsForbidden(m.Path) {
				continue // system/sudo leftovers: not removable in MVP
			}
			if m.Shared {
				shared++
				continue // group containers off by default
			}
			all = append(all, m)
		}
	}
	return all, shared
}

func appNames(selected []apps.Entry) []string {
	names := make([]string, len(selected))
	for i, e := range selected {
		names[i] = e.App.Name
	}
	return names
}

func printBatchSummary(names []string, all []scan.Match, shared int) {
	fmt.Fprintf(os.Stderr, "apps: %s\n", strings.Join(names, ", "))
	fmt.Fprintf(os.Stderr, "%d leftover item(s), ~%s to Trash", len(all), humanBytes(totalSize(all)))
	if shared > 0 {
		fmt.Fprintf(os.Stderr, "  (%d shared skipped)", shared)
	}
	fmt.Fprintln(os.Stderr)
}
