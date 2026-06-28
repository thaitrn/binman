package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/thaitrn/binman/internal/apps"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed apps with their sizes",
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(_ *cobra.Command, _ []string) error {
	entries := apps.List(apps.DefaultDirs())
	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "no apps found in /Applications")
		return nil
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tSIZE\tPATH")
	var total int64
	for _, e := range entries {
		mark := ""
		if e.Protected {
			mark = " (system)"
		}
		fmt.Fprintf(w, "%s%s\t%s\t%s\n", e.App.Name, mark, humanBytes(e.Size), e.App.Path)
		total += e.Size
	}
	_ = w.Flush()
	fmt.Fprintf(os.Stderr, "\n%d apps, ~%s total.\n", len(entries), humanBytes(total))
	return nil
}
