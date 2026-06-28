package cmd

import (
	"errors"
	"fmt"
	"os"
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
	Short: "Uninstall an app and all its ~/Library leftovers (moved to Trash)",
	Long: `Find an app by name or path, scan ~/Library for its leftovers, then move
everything to the Trash (undoable via Put Back). With no app argument and a real
terminal, an interactive app picker opens first. --apply/-y deletes without
prompting, --dry-run/-n previews only.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runUninstall,
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
	uninstallCmd.Flags().BoolVarP(&uninstallApply, "apply", "y", false, "skip the confirm prompt and delete")
}

func runUninstall(_ *cobra.Command, args []string) error {
	interactive := term.IsTerminal(os.Stdin.Fd()) && term.IsTerminal(os.Stdout.Fd())

	var a *app.App
	if len(args) == 0 {
		// No app given: open the picker (interactive only).
		if !interactive {
			return errors.New("specify an app name/path, or run 'binman uninstall' in a terminal to use the picker")
		}
		entries := apps.List(apps.DefaultDirs())
		if len(entries) == 0 {
			return errors.New("no apps found in /Applications")
		}
		picked, ok, perr := tui.PickApp(entries)
		if perr != nil {
			return perr
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "aborted.")
			return nil
		}
		a = picked.App
	} else {
		resolved, err := app.Resolve(args[0])
		if err != nil {
			return err
		}
		a = resolved
	}
	return uninstallApp(a, interactive)
}

// uninstallApp scans an app, shows leftovers, and trashes the selected ones.
func uninstallApp(a *app.App, interactive bool) error {
	matches, err := scan.Scan(a)
	if err != nil {
		return err
	}
	actionable, system := splitMatches(matches)

	// Preview-only paths: explicit --dry-run, or non-interactive without -y.
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

	// Gracefully quit a running instance before deleting its data.
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
