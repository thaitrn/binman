package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"

	"github.com/thaitrn/binman/internal/app"
	"github.com/thaitrn/binman/internal/apps"
	"github.com/thaitrn/binman/internal/safety"
	"github.com/thaitrn/binman/internal/scan"
	"github.com/thaitrn/binman/internal/trash"
	"github.com/thaitrn/binman/internal/tui"
)

var uninstallApply bool

var uninstallCmd = &cobra.Command{
	Use:   "uninstall [<app>]",
	Short: "Uninstall app(s) and all their ~/Library leftovers (moved to Trash)",
	Long: `With no app argument and a real terminal: list apps → select many → confirm
→ process (Trash) → results. With an app name/path: uninstall that one app.
--apply/-y skips confirm, --dry-run/-n previews only.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runUninstall,
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
	uninstallCmd.Flags().BoolVarP(&uninstallApply, "apply", "y", false, "skip the confirm prompt and delete")
}

func runUninstall(_ *cobra.Command, args []string) error {
	interactive := term.IsTerminal(os.Stdin.Fd()) && term.IsTerminal(os.Stdout.Fd())
	if len(args) == 0 {
		if !interactive {
			return errors.New("specify an app name/path, or run 'binman uninstall' in a terminal to use the picker")
		}
		return runInteractiveBatch()
	}
	a, err := app.Resolve(args[0])
	if err != nil {
		return err
	}
	return uninstallApp(a, interactive)
}

// runInteractiveBatch is the multi-app flow:
// list → select many → confirm → process (Trash) → results.
func runInteractiveBatch() error {
	entries := apps.List(apps.DefaultDirs())
	if len(entries) == 0 {
		return errors.New("no apps found in /Applications")
	}

	// 1-2. List + multi-select.
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
	names := make([]string, len(selected))
	for i, e := range selected {
		names[i] = e.App.Name
	}

	if dryRun {
		printBatchSummary(names, all, shared)
		fmt.Fprintln(os.Stderr, "\n(--dry-run) nothing deleted.")
		return nil
	}

	// 4. Confirm.
	if !uninstallApply {
		confirmed, cerr := tui.ConfirmBatch(names, len(all), totalSize(all), shared)
		if cerr != nil {
			return cerr
		}
		if !confirmed {
			fmt.Fprintln(os.Stderr, "aborted.")
			return nil
		}
	}

	// Quit running instances before deleting their data.
	for _, e := range selected {
		if safety.IsAppRunning(e.App.Name) {
			fmt.Fprintf(os.Stderr, "quitting %s...\n", e.App.Name)
			_ = safety.QuitApp(e.App.Name)
		}
	}

	// 5-6. Process + results.
	if len(all) == 0 {
		fmt.Fprintln(os.Stderr, "no removable leftovers; nothing to do.")
		return nil
	}
	if _, err := tui.RunProgress(all, false); err != nil {
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

func printBatchSummary(names []string, all []scan.Match, shared int) {
	fmt.Fprintf(os.Stderr, "apps: %s\n", strings.Join(names, ", "))
	fmt.Fprintf(os.Stderr, "%d leftover item(s), ~%s to Trash", len(all), humanBytes(totalSize(all)))
	if shared > 0 {
		fmt.Fprintf(os.Stderr, "  (%d shared skipped)", shared)
	}
	fmt.Fprintln(os.Stderr)
}

// --- single-app path (binman uninstall <app>) ---

// uninstallApp scans one app, shows leftovers, and trashes the selected ones.
func uninstallApp(a *app.App, interactive bool) error {
	matches, err := scan.Scan(a)
	if err != nil {
		return err
	}
	actionable, system := splitMatches(matches)

	if dryRun || (!interactive && !uninstallApply) {
		printSummary(a, actionable, system)
		if dryRun {
			fmt.Fprintln(os.Stderr, "\n(--dry-run) nothing deleted. Re-run with -y to move to Trash.")
		} else {
			fmt.Fprintln(os.Stderr, "\n(non-interactive) re-run with --apply/-y to move to Trash.")
		}
		return nil
	}

	var toDelete []scan.Match
	if uninstallApply {
		toDelete = actionable
		printSummary(a, actionable, system)
	} else {
		selected, ok, cerr := tui.Confirm(fmt.Sprintf("Uninstall %s — pick leftovers to trash", a.Name), actionable, true)
		if cerr != nil {
			return cerr
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "aborted.")
			return nil
		}
		if len(selected) == 0 {
			fmt.Fprintln(os.Stderr, "nothing selected; aborted.")
			return nil
		}
		toDelete = selected
	}

	if safety.IsAppRunning(a.Name) {
		fmt.Fprintf(os.Stderr, "quitting %s...\n", a.Name)
		_ = safety.QuitApp(a.Name)
	}
	if interactive {
		if _, err := tui.RunProgress(toDelete, false); err != nil {
			return err
		}
	} else {
		deleteHeadless(toDelete)
	}
	reportSystemSkipped(system)
	return nil
}

// printSummary writes the aligned leftover table to stdout (previews / -y path).
func printSummary(a *app.App, actionable, system []scan.Match) {
	fmt.Fprintf(os.Stderr, "%s  (%s)\n\n", a.Path, a.BundleID)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	printMatches(w, actionable, system)
	_ = w.Flush()
	fmt.Fprintf(os.Stderr, "\n%d actionable item(s), ~%s to Trash.\n", len(actionable), humanBytes(totalSize(actionable)))
}

// deleteHeadless trashes paths without a TUI (non-interactive -y path).
func deleteHeadless(toDelete []scan.Match) {
	paths := make([]string, 0, len(toDelete))
	for _, m := range toDelete {
		paths = append(paths, m.Path)
	}
	results := trash.Trash(paths, false)
	var freed int64
	failed := 0
	for i, r := range results {
		if r.Err != nil {
			failed++
			fmt.Fprintf(os.Stderr, "  ✗ %s: %v\n", r.Path, r.Err)
			continue
		}
		freed += toDelete[i].Size
	}
	fmt.Fprintf(os.Stderr, "\nmoved %d item(s) (~%s) to Trash.\n", len(results)-failed, humanBytes(freed))
	if failed > 0 {
		fmt.Fprintf(os.Stderr, "%d item(s) could not be moved.\n", failed)
	}
}

func reportSystemSkipped(system []scan.Match) {
	if len(system) > 0 {
		fmt.Fprintf(os.Stderr, "%d system/sudo item(s) skipped (needs root; MVP removes user-domain only).\n", len(system))
	}
}
