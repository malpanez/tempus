package cli

import (
	"fmt"
	"sort"
	"strings"

	"tempus/internal/config"

	"github.com/spf13/cobra"
)

func NewConfigCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage tempus configuration",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "set <key> <value>",
			Short: "Set a configuration value",
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				return runConfigSet(app, cmd, args)
			},
		},
		&cobra.Command{
			Use:   "list",
			Short: "List all configuration values",
			RunE: func(cmd *cobra.Command, args []string) error {
				return runConfigList(app, cmd, args)
			},
		},
		&cobra.Command{
			Use:   "alarm-profiles",
			Short: "List available alarm profiles",
			RunE: func(cmd *cobra.Command, args []string) error {
				return runConfigAlarmProfiles(app, cmd, args)
			},
		},
	)

	return cmd
}

func runConfigSet(app *App, _ *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if err := cfg.Set(args[0], args[1]); err != nil {
		return err
	}
	PrintOK(app.Stdout, "Config updated: %s = %s\n", args[0], args[1])
	return nil
}

func runConfigList(_ *App, _ *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	return cfg.List()
}

func runConfigAlarmProfiles(_ *App, _ *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if len(cfg.AlarmProfiles) == 0 {
		fmt.Println("No alarm profiles configured.")
		return nil
	}

	fmt.Println("Available alarm profiles:")
	fmt.Println()

	names := cfg.ListAlarmProfiles()
	sort.Strings(names)

	for _, name := range names {
		profile := cfg.GetAlarmProfile(name)
		if profile == nil {
			continue
		}

		fmt.Printf("  %s:\n", name)
		if len(profile) == 0 {
			fmt.Println("    (no alarms)")
		} else {
			for _, trigger := range profile {
				fmt.Printf("    - %s\n", trigger)
			}
		}
		fmt.Println()
	}

	fmt.Println("Usage in batch files:")
	fmt.Printf("  CSV:  alarms column with 'profile:adhd-triple'\n")
	fmt.Printf("  JSON: \"alarms\": [\"profile:medication\"]\n")
	fmt.Printf("  YAML: alarms: [profile:single]\n")

	return nil
}

func findSubcommand(parent *cobra.Command, name string) *cobra.Command {
	for _, cmd := range parent.Commands() {
		if strings.HasPrefix(cmd.Use, name) {
			return cmd
		}
	}
	return nil
}
