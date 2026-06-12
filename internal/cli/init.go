package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"tempus/internal/config"
)

func NewInitCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Configure Tempus interactively",
		Long:  "Run an interactive wizard to set up timezone, language, output directory, and default alarm profile.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(app)
		},
	}
}

func runInit(app *App) error {
	configDir, err := config.ConfigDir()
	if err != nil {
		return err
	}
	configFile := filepath.Join(configDir, "config.yaml")

	if _, err := os.Stat(configFile); err == nil {
		var overwrite bool
		form := huh.NewForm(huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("Config already exists at %s. Overwrite?", configFile)).
				Affirmative("Yes").
				Negative("No").
				Value(&overwrite),
		))
		if err := form.Run(); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				return fmt.Errorf("init aborted")
			}
			return fmt.Errorf("interactive form: %w", err)
		}
		if !overwrite {
			PrintOK(app.Stdout, "Config unchanged. Use 'tempus config set <key> <value>' for individual changes.\n")
			return nil
		}
	}

	timezone := config.DetectTimezone()
	tzForm := huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title("Timezone").
			Value(&timezone).
			Validate(func(s string) error {
				return config.ValidateTimezone(s)
			}),
	))
	if err := tzForm.Run(); err != nil {
		return nil
	}

	detectedLang := config.DetectLanguage()
	language := detectedLang
	langForm := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().
			Title("Language").
			Options(
				huh.NewOption("English", "en"),
				huh.NewOption("Spanish", "es"),
				huh.NewOption("Portuguese", "pt"),
				huh.NewOption("Galician", "ga"),
			).
			Value(&language),
	))
	if err := langForm.Run(); err != nil {
		return nil
	}

	outputDir := "."
	dirForm := huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title("Output directory").
			Value(&outputDir).
			Validate(func(s string) error {
				return config.ValidateOutputDir(s)
			}),
	))
	if err := dirForm.Run(); err != nil {
		return nil
	}

	alarmProfile := "adhd-default"
	profileForm := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().
			Title("Default alarm profile").
			Options(
				huh.NewOption("ADHD Default (-2h, -30m, -5m)", "adhd-default"),
				huh.NewOption("ADHD Countdown (-1h, -45m, -30m, -15m, -5m)", "adhd-countdown"),
				huh.NewOption("Medication (-30m, -5m)", "medication"),
				huh.NewOption("Single (-15m)", "single"),
				huh.NewOption("None", "none"),
			).
			Value(&alarmProfile),
	))
	if err := profileForm.Run(); err != nil {
		return nil
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	for _, kv := range [][2]string{
		{"timezone", timezone},
		{"language", language},
		{"output_dir", outputDir},
		{"default_alarm_profile", alarmProfile},
	} {
		if err := cfg.Set(kv[0], kv[1]); err != nil {
			return err
		}
	}

	PrintOK(app.Stdout, "\n> Config saved to %s\n\n", configFile)
	PrintOK(app.Stdout, "  Timezone:      %s\n", timezone)
	PrintOK(app.Stdout, "  Language:      %s\n", language)
	PrintOK(app.Stdout, "  Output dir:    %s\n", outputDir)
	PrintOK(app.Stdout, "  Alarm profile: %s\n", alarmProfile)
	PrintOK(app.Stdout, "\nNext: tempus create --start today --duration 1h\n")
	return nil
}
