package cli

import (
	"bytes"
	"fmt"
	"io"

	"tempus/internal/config"
	"tempus/internal/i18n"

	"github.com/spf13/cobra"
)

type App struct {
	Config     *config.Config
	Translator *i18n.Translator
	Stdout     io.Writer
	Stderr     io.Writer
}

func SetupPersistentPreRunE(app *App) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		configPath, _ := cmd.Flags().GetString("config")
		cfg, err := config.LoadFrom(configPath)
		if err != nil {
			if configPath != "" {
				return fmt.Errorf("config file %q: %w", configPath, err)
			}
			fmt.Fprintf(app.Stderr, "Warning: could not load config (%v), using defaults\n", err)
			cfg = &config.Config{
				Language:   "en",
				Timezone:   "UTC",
				DateFormat: "2006-01-02",
				TimeFormat: "15:04",
				OutputDir:  ".",
			}
		}

		if lang, _ := cmd.Flags().GetString("language"); lang != "" {
			cfg.Language = lang
		}
		if tz, _ := cmd.Flags().GetString("timezone"); tz != "" {
			cfg.Timezone = tz
		}

		app.Config = cfg

		t, err := i18n.NewTranslator(cfg.Language)
		if err != nil {
			t, _ = i18n.NewTranslator("en")
		}
		app.Translator = t

		return nil
	}
}

func TestApp() *App {
	t, _ := i18n.NewTranslator("en")
	return &App{
		Config: &config.Config{
			Language:   "en",
			Timezone:   "UTC",
			DateFormat: "2006-01-02",
			TimeFormat: "15:04",
			OutputDir:  ".",
		},
		Translator: t,
		Stdout:     &bytes.Buffer{},
		Stderr:     &bytes.Buffer{},
	}
}
