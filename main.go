package main

import (
	"os"

	"tempus/internal/cli"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = ""
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	app := &cli.App{Stdout: os.Stdout, Stderr: os.Stderr}

	cmd := &cobra.Command{
		Use:               "tempus",
		Short:             "A multilingual ICS calendar file generator",
		SilenceUsage:      true,
		PersistentPreRunE: cli.SetupPersistentPreRunE(app),
	}

	cmd.PersistentFlags().StringP("language", "l", "", "Language for output (es, en, ga, pt)")
	cmd.PersistentFlags().StringP("timezone", "t", "", "Default timezone")
	cmd.PersistentFlags().StringP("config", "c", "", "Config file path")

	cmd.AddCommand(
		cli.NewCreateCmd(app),
		cli.NewQuickCmd(app),
		cli.NewInitCmd(app),
		cli.NewBatchCmd(app),
		cli.NewLintCmd(app),
		cli.NewConfigCmd(app),
		cli.NewVersionCmd(app, version, commit, date),
		cli.NewTemplateCmd(app),
		cli.NewLocaleCmd(app),
		cli.NewTimezoneCmd(app),
		cli.NewRRuleHelperCmd(app),
	)

	return cmd
}
