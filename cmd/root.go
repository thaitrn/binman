// Package cmd wires binman's commands and flags.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// Version is overridden at build time via -ldflags; "dev" otherwise.
var Version = "dev"

// dryRun (global, -n) makes every command preview-only.
var dryRun bool

// apply (-y) skips the confirm screen and deletes.
var apply bool

var rootCmd = &cobra.Command{
	Use:   "binman",
	Short: "Terminal macOS app uninstaller — list, select, and remove apps + leftovers",
	Long: `binman lists your installed apps, lets you pick one or many, and uninstalls
them completely (.app + ~/Library leftovers → Trash). CleanMyMac-style, in your
terminal, undoable via Put Back. System apps (/System) are skipped.

Just run it:
  binman             # list → select many → confirm → Trash → results
  binman -n          # preview only (dry-run)
  binman -y          # skip the confirm screen
  binman clean       # (optional) clean system junk (caches/logs/Xcode/pkg)`,
	Version:      Version,
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE: func(_ *cobra.Command, _ []string) error {
		return runPickerFlow()
	},
}

// Execute runs the root command and exits non-zero on error.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "n", false, "preview only; do not delete or change anything")
	rootCmd.Flags().BoolVarP(&apply, "apply", "y", false, "skip the confirm screen and delete")
}
