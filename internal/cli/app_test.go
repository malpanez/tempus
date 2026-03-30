package cli

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func TestTestApp(t *testing.T) {
	app := TestApp()
	if app.Config == nil {
		t.Fatal("TestApp().Config is nil")
	}
	if app.Translator == nil {
		t.Fatal("TestApp().Translator is nil")
	}
	if app.Stdout == nil {
		t.Fatal("TestApp().Stdout is nil")
	}
	if app.Stderr == nil {
		t.Fatal("TestApp().Stderr is nil")
	}
}

func TestSetupPersistentPreRunE(t *testing.T) {
	app := &App{
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}

	cmd := &cobra.Command{
		Use: "test",
		PersistentPreRunE: SetupPersistentPreRunE(app),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	cmd.PersistentFlags().StringP("language", "l", "", "Language")
	cmd.PersistentFlags().StringP("timezone", "t", "", "Timezone")
	cmd.PersistentFlags().StringP("config", "c", "", "Config file path")

	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if app.Config == nil {
		t.Fatal("app.Config is nil after PersistentPreRunE")
	}
	if app.Translator == nil {
		t.Fatal("app.Translator is nil after PersistentPreRunE")
	}
}

func TestSetupPersistentPreRunEWithFlags(t *testing.T) {
	app := &App{
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}

	cmd := &cobra.Command{
		Use: "test",
		PersistentPreRunE: SetupPersistentPreRunE(app),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	cmd.PersistentFlags().StringP("language", "l", "", "Language")
	cmd.PersistentFlags().StringP("timezone", "t", "", "Timezone")
	cmd.PersistentFlags().StringP("config", "c", "", "Config file path")

	cmd.SetArgs([]string{"--language", "es", "--timezone", "Europe/Madrid"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if app.Config.Language != "es" {
		t.Errorf("language = %q, want %q", app.Config.Language, "es")
	}
	if app.Config.Timezone != "Europe/Madrid" {
		t.Errorf("timezone = %q, want %q", app.Config.Timezone, "Europe/Madrid")
	}
}
