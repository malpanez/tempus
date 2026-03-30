package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func NewVersionCmd(app *App, version, commit, date string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(_ *cobra.Command, _ []string) {
			if strings.TrimSpace(date) == "" {
				fmt.Fprintf(app.Stdout, "tempus %s\n", version)
			} else {
				fmt.Fprintf(app.Stdout, "tempus %s (%s) built %s\n", version, commit, date)
			}
		},
	}
}
