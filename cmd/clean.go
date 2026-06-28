package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/thaitrn/binman/internal/clean"
)

var (
	cleanApply     bool
	cleanAll       bool
	cleanXcode     bool
	cleanPkg       bool
	cleanDownloads bool
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean system junk (caches, logs, Xcode, package-manager caches)",
	Long: `Remove regenerable junk: user caches and logs by default. Add categories with
flags. Defaults to a dry-run preview; pass --apply/-y to actually delete
(caches/logs/Xcode move to Trash; package managers run their cleanup commands).`,
	RunE: runClean,
}

func init() {
	rootCmd.AddCommand(cleanCmd)
	cleanCmd.Flags().BoolVarP(&cleanApply, "apply", "y", false, "execute (default is dry-run)")
	cleanCmd.Flags().BoolVar(&cleanAll, "all", false, "include every category")
	cleanCmd.Flags().BoolVar(&cleanXcode, "xcode", false, "include Xcode artifacts")
	cleanCmd.Flags().BoolVar(&cleanPkg, "pkg", false, "include package-manager caches (brew/npm/pnpm/pip/docker)")
	cleanCmd.Flags().BoolVar(&cleanDownloads, "downloads", false, "include old installers in ~/Downloads")
}

func runClean(_ *cobra.Command, _ []string) error {
	home, _ := os.UserHomeDir()

	groups := []string{"default"}
	if cleanDownloads || cleanAll {
		groups = append(groups, "downloads")
	}
	if cleanXcode || cleanAll {
		groups = append(groups, "xcode")
	}
	if cleanPkg || cleanAll {
		groups = append(groups, "pkg")
	}

	execute := cleanApply && !dryRun
	mode := "DRY-RUN"
	if execute {
		mode = "APPLY"
	}
	fmt.Fprintf(os.Stderr, "binman clean [%s] — groups: %s\n\n", mode, strings.Join(groups, ", "))

	clean.Apply(clean.Plan(home, groups), !execute)

	if !execute {
		fmt.Fprintln(os.Stderr, "\n(dry-run) nothing changed. Re-run with --apply/-y to execute.")
	}
	return nil
}
