// Package cmd wires binman's cobra subcommands and global flags.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// Version is overridden at build time via -ldflags; "dev" otherwise.
var Version = "dev"

// dryRun, when true, makes every command preview-only (no deletion/mutation).
var dryRun bool

var rootCmd = &cobra.Command{
	Use:          "binman",
	Short:        "Terminal macOS app cleaner — uninstall apps fully + clean system junk",
	Long:         "binman uninstalls macOS apps completely (.app + ~/Library leftovers → Trash) and cleans system junk. CleanMyMac-style, in your terminal.",
	Version:      Version,
	SilenceUsage: true,
}

// Execute runs the root command and exits non-zero on error.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global preview flag (default safe; destructive commands default to dry-run anyway).
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "n", false, "preview only; do not delete or change anything")
}
