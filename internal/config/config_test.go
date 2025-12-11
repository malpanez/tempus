package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestLoad_Defaults(t *testing.T) {
	// Create a temporary directory for config
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))

	// Reset viper between tests
	viper.Reset()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Check defaults
	if cfg.Language != "en" {
		t.Errorf("expected language 'en', got %q", cfg.Language)
	}
	if cfg.Timezone != "UTC" {
		t.Errorf("expected timezone 'UTC', got %q", cfg.Timezone)
	}
	if cfg.DateFormat != "2006-01-02" {
		t.Errorf("expected date_format '2006-01-02', got %q", cfg.DateFormat)
	}
	if cfg.TimeFormat != "15:04" {
		t.Errorf("expected time_format '15:04', got %q", cfg.TimeFormat)
	}
	if cfg.OutputDir != "." {
		t.Errorf("expected output_dir '.', got %q", cfg.OutputDir)
	}
	if cfg.DefaultTitle != "Event" {
		t.Errorf("expected default_title 'Event', got %q", cfg.DefaultTitle)
	}
}

func TestLoad_FromFile(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "tempus")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))

	// Write a config file
	configFile := filepath.Join(configDir, "config.yaml")
	configContent := `language: es
timezone: Europe/Madrid
date_format: "02/01/2006"
time_format: "15:04"
output_dir: "/tmp/events"
default_title: "Mi Evento"
`
	if err := os.WriteFile(configFile, []byte(configContent), 0o644); err != nil {
		t.Fatal(err)
	}

	viper.Reset()
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Language != "es" {
		t.Errorf("expected language 'es', got %q", cfg.Language)
	}
	if cfg.Timezone != "Europe/Madrid" {
		t.Errorf("expected timezone 'Europe/Madrid', got %q", cfg.Timezone)
	}
	if cfg.DateFormat != "02/01/2006" {
		t.Errorf("expected date_format '02/01/2006', got %q", cfg.DateFormat)
	}
	if cfg.OutputDir != "/tmp/events" {
		t.Errorf("expected output_dir '/tmp/events', got %q", cfg.OutputDir)
	}
	if cfg.DefaultTitle != "Mi Evento" {
		t.Errorf("expected default_title 'Mi Evento', got %q", cfg.DefaultTitle)
	}
}

func TestSet_ValidKey(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))

	viper.Reset()
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	// Set a value
	if err := cfg.Set("language", "pt"); err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	// Verify in-memory
	if cfg.Language != "pt" {
		t.Errorf("expected language 'pt', got %q", cfg.Language)
	}

	// Verify Get works
	val, err := cfg.Get("language")
	if err != nil {
		t.Errorf("Get() failed: %v", err)
	}
	if val != "pt" {
		t.Errorf("expected 'pt', got %q", val)
	}
}

func TestSet_InvalidKey(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))

	viper.Reset()
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	// Try to set an invalid key
	err = cfg.Set("invalid_key", "value")
	if err == nil {
		t.Error("expected error for invalid key, got nil")
	}
	if !strings.Contains(err.Error(), "unknown configuration key") {
		t.Errorf("expected 'unknown configuration key' error, got: %v", err)
	}
}

func TestGet_AllKeys(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))

	viper.Reset()
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	keys := []string{"language", "timezone", "date_format", "time_format", "output_dir", "default_title"}
	for _, key := range keys {
		_, err := cfg.Get(key)
		if err != nil {
			t.Errorf("Get(%q) failed: %v", key, err)
		}
	}
}

func TestGet_InvalidKey(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))

	viper.Reset()
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	_, err = cfg.Get("nonexistent")
	if err == nil {
		t.Error("expected error for invalid key, got nil")
	}
	if !strings.Contains(err.Error(), "unknown configuration key") {
		t.Errorf("expected 'unknown configuration key' error, got: %v", err)
	}
}

func TestGetOrDefault(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))

	viper.Reset()
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	// Test with valid key
	val := cfg.GetOrDefault("language", "fallback")
	if val == "fallback" {
		t.Error("expected actual value, got fallback")
	}

	// Test with invalid key
	val = cfg.GetOrDefault("nonexistent", "fallback")
	if val != "fallback" {
		t.Errorf("expected 'fallback', got %q", val)
	}
}

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "tempus")
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))

	viper.Reset()
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	// Modify using Set to properly update both struct and viper
	if err := cfg.Set("language", "ga"); err != nil {
		t.Fatalf("Set(language) failed: %v", err)
	}
	if err := cfg.Set("timezone", "Europe/Dublin"); err != nil {
		t.Fatalf("Set(timezone) failed: %v", err)
	}

	// Verify file exists
	configFile := filepath.Join(configDir, "config.yaml")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Error("config file was not created")
	}

	// Load again and verify
	viper.Reset()
	cfg2, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	if cfg2.Language != "ga" {
		t.Errorf("expected language 'ga', got %q", cfg2.Language)
	}
	if cfg2.Timezone != "Europe/Dublin" {
		t.Errorf("expected timezone 'Europe/Dublin', got %q", cfg2.Timezone)
	}
}

func TestValidateTimezone(t *testing.T) {
	tests := []struct {
		name    string
		tz      string
		wantErr bool
	}{
		{"valid UTC", "UTC", false},
		{"valid America/New_York", "America/New_York", false},
		{"valid Europe/Madrid", "Europe/Madrid", false},
		{"invalid timezone", "Invalid/Timezone", true},
		{"empty timezone", "", true},
		{"whitespace only", "   ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTimezone(tt.tz)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTimezone(%q) error = %v, wantErr %v", tt.tz, err, tt.wantErr)
			}
		})
	}
}

func TestValidateLanguage(t *testing.T) {
	tests := []struct {
		name    string
		lang    string
		wantErr bool
	}{
		{"valid en", "en", false},
		{"valid es", "es", false},
		{"valid EN uppercase", "EN", false},
		{"valid pt", "pt", false},
		{"valid ga", "ga", false},
		{"invalid language", "invalid", true},
		{"empty language", "", true},
		{"whitespace only", "   ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLanguage(tt.lang)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLanguage(%q) error = %v, wantErr %v", tt.lang, err, tt.wantErr)
			}
		})
	}
}

func TestGetConfigDir(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))

	dir, err := getConfigDir()
	if err != nil {
		t.Fatalf("getConfigDir() failed: %v", err)
	}

	if dir == "" {
		t.Error("expected non-empty config dir")
	}

	// Should contain "tempus" in the path
	if !strings.Contains(dir, "tempus") {
		t.Errorf("expected config dir to contain 'tempus', got: %s", dir)
	}
}

func TestConfigDir(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))

	dir, err := ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() failed: %v", err)
	}

	if dir == "" {
		t.Error("expected non-empty config dir")
	}

	// Should match getConfigDir
	expectedDir, _ := getConfigDir()
	if dir != expectedDir {
		t.Errorf("ConfigDir() = %q, want %q", dir, expectedDir)
	}
}

func TestList(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))

	viper.Reset()
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	// List should not return an error
	if err := cfg.List(); err != nil {
		t.Errorf("List() failed: %v", err)
	}
}

func TestSet_AllFields(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))

	viper.Reset()
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		key   string
		value string
		check func(*Config) string
	}{
		{"language", "es", func(c *Config) string { return c.Language }},
		{"timezone", "Europe/Madrid", func(c *Config) string { return c.Timezone }},
		{"date_format", "02/01/2006", func(c *Config) string { return c.DateFormat }},
		{"time_format", "15:04:05", func(c *Config) string { return c.TimeFormat }},
		{"output_dir", "/tmp", func(c *Config) string { return c.OutputDir }},
		{"default_title", "Test Event", func(c *Config) string { return c.DefaultTitle }},
	}

	for _, tt := range tests {
		t.Run("set_"+tt.key, func(t *testing.T) {
			if err := cfg.Set(tt.key, tt.value); err != nil {
				t.Fatalf("Set(%q, %q) failed: %v", tt.key, tt.value, err)
			}

			actual := tt.check(cfg)
			if actual != tt.value {
				t.Errorf("expected %q, got %q", tt.value, actual)
			}
		})
	}
}
