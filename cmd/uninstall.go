package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/thaitrn/binman/internal/app"
	"github.com/thaitrn/binman/internal/safety"
	"github.com/thaitrn/binman/internal/scan"
	"github.com/thaitrn/binman/internal/trash"
)

var uninstallApply bool

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <app>",
	Short: "Uninstall an app and all its ~/Library leftovers (moved to Trash)",
	Long: `Find an app by name or path, scan ~/Library for its leftovers, show a table,
and move everything to the Trash (undoable via Put Back). Defaults to a preview;
pass --apply/-y to actually delete, or answer the prompt.`,
	Args: cobra.ExactArgs(1),
	RunE: runUninstall,
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
	uninstallCmd.Flags().BoolVarP(&uninstallApply, "apply", "y", false, "skip the confirm prompt and delete")
}

func runUninstall(_ *cobra.Command, args []string) error {
	a, err := app.Resolve(args[0])
	if err != nil {
		return err
	}
	matches, err := scan.Scan(a)
	if err != nil {
		return err
	}
	actionable, system := splitMatches(matches)

	fmt.Fprintf(os.Stderr, "%s  (%s)\n\n", a.Path, a.BundleID)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	printMatches(w, actionable, system)
	_ = w.Flush()

	fmt.Fprintf(os.Stderr, "\n%d item(s), ~%s will be moved to Trash.\n", len(actionable), humanBytes(totalSize(actionable)))

	// Decide whether to delete: global --dry-run wins; else -y or interactive confirm.
	switch {
	case dryRun:
		fmt.Fprintln(os.Stderr, "(--dry-run) nothing deleted.")
		return nil
	case uninstallApply:
		// proceed to deletion
	case confirmPrompt(fmt.Sprintf("Move these %d item(s) to Trash? [y/N]", len(actionable))):
		// proceed to deletion
	default:
		fmt.Fprintln(os.Stderr, "aborted.")
		return nil
	}

	return deleteActionable(a, actionable, system)
}

// deleteActionable quits a running app, trashes actionable paths, and reports.
func deleteActionable(a *app.App, actionable, system []scan.Match) error {
	if safety.IsAppRunning(a.Name) {
		fmt.Fprintf(os.Stderr, "quitting %s...\n", a.Name)
		_ = safety.QuitApp(a.Name)
	}

	paths := make([]string, 0, len(actionable))
	for _, m := range actionable {
		paths = append(paths, m.Path)
	}
	results := trash.Trash(paths, false)

	var freed int64
	var failed int
	for i, r := range results {
		if r.Err != nil {
			failed++
			fmt.Fprintf(os.Stderr, "  ✗ %s: %v\n", r.Path, r.Err)
			continue
		}
		freed += actionable[i].Size
	}
	fmt.Fprintf(os.Stderr, "\nmoved %d item(s) (~%s) to Trash.\n", len(results)-failed, humanBytes(freed))
	if failed > 0 {
		fmt.Fprintf(os.Stderr, "%d item(s) could not be moved; see errors above.\n", failed)
	}
	if len(system) > 0 {
		fmt.Fprintf(os.Stderr, "%d system/sudo item(s) skipped (needs root; MVP removes user-domain only).\n", len(system))
	}
	return nil
}

// confirmPrompt asks a yes/no question on stdin; default is no.
func confirmPrompt(question string) bool {
	fmt.Fprintf(os.Stderr, "%s ", question)
	r := bufio.NewReader(os.Stdin)
	line, err := r.ReadString('\n')
	if err != nil {
		return false
	}
	ans := strings.ToLower(strings.TrimSpace(line))
	return ans == "y" || ans == "yes"
}
