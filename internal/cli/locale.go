package cli

import (
	"fmt"
	"os"
	"strings"

	"tempus/internal/i18n"

	"github.com/spf13/cobra"
)

func NewLocaleCmd(app *App) *cobra.Command {
	root := &cobra.Command{
		Use:   "locale",
		Short: "Inspect available locales",
	}

	root.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List available locales",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLocaleList(app, cmd, args)
		},
	})

	return root
}

func runLocaleList(app *App, _ *cobra.Command, _ []string) error {
	w := app.Stdout
	if w == nil {
		w = os.Stdout
	}
	locales := i18n.Locales()
	if len(locales) == 0 {
		fmt.Fprintln(w, "No locales found.")
		return nil
	}
	fmt.Fprintf(w, "%-8s %-9s %s\n", "Code", "Embedded", "Disk Paths")
	for _, loc := range locales {
		embedded := "no"
		if loc.Embedded {
			embedded = "yes"
		}
		paths := "-"
		if len(loc.DiskPaths) > 0 {
			paths = strings.Join(loc.DiskPaths, ", ")
		}
		fmt.Fprintf(w, "%-8s %-9s %s\n", loc.Code, embedded, paths)
	}
	return nil
}
