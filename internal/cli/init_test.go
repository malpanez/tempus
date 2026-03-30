package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCmdRegistered(t *testing.T) {
	app := TestApp()
	cmd := NewInitCmd(app)
	if cmd == nil {
		t.Fatal("NewInitCmd() returned nil")
	}
	if cmd.Use != "init" {
		t.Errorf("expected Use == 'init', got %q", cmd.Use)
	}
}

func TestInitCmdHelp(t *testing.T) {
	app := TestApp()
	cmd := NewInitCmd(app)
	if cmd == nil {
		t.Fatal("NewInitCmd() returned nil")
	}
	if cmd.Use != "init" {
		t.Errorf("expected Use == 'init', got %q", cmd.Use)
	}
	if !strings.Contains(cmd.Short, "Configure Tempus") && !strings.Contains(cmd.Short, "interactively") {
		t.Errorf("expected Short to contain 'Configure Tempus' or 'interactively', got %q", cmd.Short)
	}
	if cmd.RunE == nil {
		t.Error("init command should have RunE function")
	}
}

func TestInitCmdExistingConfigNoOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	configDir := filepath.Join(tmpDir, "tempus")
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	configFile := filepath.Join(configDir, "config.yaml")
	original := []byte("timezone: UTC\nlanguage: en\n")
	if err := os.WriteFile(configFile, original, 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	app := TestApp()
	err := runInit(app)
	if err != nil {
		t.Fatalf("runInit returned error: %v", err)
	}

	after, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config after runInit: %v", err)
	}
	if string(after) != string(original) {
		t.Errorf("config was modified when huh form returned error (non-TTY)\nbefore: %s\nafter: %s", original, after)
	}
}

func TestInitCmdNoExistingConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	app := TestApp()
	err := runInit(app)
	if err != nil {
		t.Fatalf("runInit returned error on fresh config: %v", err)
	}
}
