package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"tempus/internal/testutil"
	"testing"
	"time"

	"github.com/spf13/viper"
)

const (
	testConfigDir        = ".config"
	testTimezoneEuMadrid = testutil.TZEuropeMadrid
)

func TestLoadDefaults(t *testing.T) {
	// Create a temporary directory for config
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, testConfigDir))

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
	if cfg.DateFormat != testutil.TestDateFormatDate {
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

func TestLoadFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, testConfigDir, "tempus")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, testConfigDir))

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
	if cfg.Timezone != testTimezoneEuMadrid {
		t.Errorf("expected timezone %q, got %q", testTimezoneEuMadrid, cfg.Timezone)
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

func TestSetValidKey(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, testConfigDir))

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

func TestSetInvalidKey(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, testConfigDir))

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

func TestGetAllKeys(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, testConfigDir))

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

func TestGetInvalidKey(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, testConfigDir))

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
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, testConfigDir))

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
	configDir := filepath.Join(tmpDir, testConfigDir, "tempus")
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, testConfigDir))

	viper.Reset()
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	// Modify using Set to properly update both struct and viper
	if err := cfg.Set("language", "ga"); err != nil {
		t.Fatalf("Set(language) failed: %v", err)
	}
	if err := cfg.Set("timezone", testutil.TZEuropeDublin); err != nil {
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
	if cfg2.Timezone != testutil.TZEuropeDublin {
		t.Errorf("expected timezone 'Europe/Dublin', got %q", cfg2.Timezone)
	}
}

// TestSaveConfigPermissions is the CN-008 regression test: config.yaml must be
// mode 0600, both when Save() creates it and when Save() rewrites an existing
// file. Windows does not model POSIX permission bits, so it is skipped there.
func TestSaveConfigPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits are not modelled on Windows")
	}

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, testConfigDir, "tempus")
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, testConfigDir))

	viper.Reset()
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	configFile := filepath.Join(configDir, "config.yaml")

	// First Save() creates the file.
	if err := cfg.Save(); err != nil {
		t.Fatalf("first Save() failed: %v", err)
	}
	info, err := os.Stat(configFile)
	if err != nil {
		t.Fatalf("stat after first Save(): %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("after first Save(): config.yaml mode = %04o, want 0600", perm)
	}

	// Loosen it the way a pre-change Tempus would have left it, then rewrite:
	// Save() must tighten it back, which is what makes the fix idempotent.
	if err := os.Chmod(configFile, 0o644); err != nil {
		t.Fatalf("chmod 0644: %v", err)
	}
	if err := cfg.Save(); err != nil {
		t.Fatalf("second Save() failed: %v", err)
	}
	info, err = os.Stat(configFile)
	if err != nil {
		t.Fatalf("stat after second Save(): %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("after second Save(): config.yaml mode = %04o, want 0600", perm)
	}
}

func TestValidateTimezone(t *testing.T) {
	tests := []struct {
		name    string
		tz      string
		wantErr bool
	}{
		{"valid UTC", "UTC", false},
		{"valid America/New_York", testutil.TZAmericaNewYork, false},
		{"valid Europe/Madrid", testTimezoneEuMadrid, false},
		{"invalid timezone", testutil.TZInvalid, true},
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
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, testConfigDir))

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
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, testConfigDir))

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
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, testConfigDir))

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

func TestGetAlarmProfile(t *testing.T) {
	cfg := &Config{
		AlarmProfiles: map[string][]string{
			"adhd-default": {"-2h", "-1h", "-30m", "-10m"},
			"medication":   {"-5m", "-1m", "0m"},
			"none":         {},
		},
	}

	tests := []struct {
		name     string
		cfg      *Config
		profile  string
		expected []string
	}{
		{"known profile adhd-default", cfg, "adhd-default", []string{"-2h", "-1h", "-30m", "-10m"}},
		{"known profile medication", cfg, "medication", []string{"-5m", "-1m", "0m"}},
		{"known profile none returns empty", cfg, "none", []string{}},
		{"unknown profile returns nil", cfg, "nonexistent", nil},
		{"nil map returns nil", &Config{}, "anything", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.GetAlarmProfile(tt.profile)
			if tt.expected == nil {
				if got != nil {
					t.Errorf("GetAlarmProfile(%q) = %v, want nil", tt.profile, got)
				}
				return
			}
			if len(got) != len(tt.expected) {
				t.Fatalf("GetAlarmProfile(%q) len = %d, want %d", tt.profile, len(got), len(tt.expected))
			}
			for i, v := range got {
				if v != tt.expected[i] {
					t.Errorf("GetAlarmProfile(%q)[%d] = %q, want %q", tt.profile, i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestListAlarmProfiles(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *Config
		wantLen   int
		wantNames []string
	}{
		{
			"default profiles returns all names",
			&Config{
				AlarmProfiles: map[string][]string{
					"adhd-default":   {"-2h", "-1h", "-30m", "-10m"},
					"adhd-countdown": {"-1d", "-1h", "-15m", "-5m"},
					"medication":     {"-5m", "-1m", "0m"},
					"single":         {"-15m"},
					"none":           {},
				},
			},
			5,
			[]string{"adhd-default", "adhd-countdown", "medication", "single", "none"},
		},
		{
			"nil map returns empty slice",
			&Config{},
			0,
			nil,
		},
		{
			"single profile returns slice of length 1",
			&Config{
				AlarmProfiles: map[string][]string{
					"custom": {"-10m"},
				},
			},
			1,
			[]string{"custom"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.ListAlarmProfiles()
			if got == nil {
				t.Fatal("ListAlarmProfiles() returned nil, want non-nil slice")
			}
			if len(got) != tt.wantLen {
				t.Fatalf("ListAlarmProfiles() len = %d, want %d", len(got), tt.wantLen)
			}
			if tt.wantNames != nil {
				nameSet := make(map[string]bool, len(got))
				for _, n := range got {
					nameSet[n] = true
				}
				for _, want := range tt.wantNames {
					if !nameSet[want] {
						t.Errorf("ListAlarmProfiles() missing expected name %q", want)
					}
				}
			}
		})
	}
}

func TestSetAllFields(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, testConfigDir))

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
		{"timezone", testTimezoneEuMadrid, func(c *Config) string { return c.Timezone }},
		{"date_format", "02/01/2006", func(c *Config) string { return c.DateFormat }},
		{"time_format", "15:04:05", func(c *Config) string { return c.TimeFormat }},
		{"output_dir", "/tmp", func(c *Config) string { return c.OutputDir }},
		{"default_title", testutil.EventTitleTestEvent, func(c *Config) string { return c.DefaultTitle }},
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

func TestEnvVarOverrideTimezone(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, testConfigDir))
	t.Setenv("TEMPUS_TIMEZONE", testutil.TZAmericaNewYork)

	viper.Reset()
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Timezone != testutil.TZAmericaNewYork {
		t.Errorf("expected timezone %q, got %q", testutil.TZAmericaNewYork, cfg.Timezone)
	}
}

func TestEnvVarOverrideLanguage(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, testConfigDir))
	t.Setenv("TEMPUS_LANGUAGE", "es")

	viper.Reset()
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Language != "es" {
		t.Errorf("expected language %q, got %q", "es", cfg.Language)
	}
}

func TestEnvVarOverrideOutputDir(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, testConfigDir))
	t.Setenv("TEMPUS_OUTPUT_DIR", "/tmp")

	viper.Reset()
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.OutputDir != "/tmp" {
		t.Errorf("expected output_dir %q, got %q", "/tmp", cfg.OutputDir)
	}
}

func TestEnvVarOverrideDateFormat(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, testConfigDir))
	t.Setenv("TEMPUS_DATE_FORMAT", "DD/MM/YYYY")

	viper.Reset()
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.DateFormat != "DD/MM/YYYY" {
		t.Errorf("expected date_format %q, got %q", "DD/MM/YYYY", cfg.DateFormat)
	}
}

func TestEnvVarOverrideTimeFormat(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, testConfigDir))
	t.Setenv("TEMPUS_TIME_FORMAT", "HH:MM:SS")

	viper.Reset()
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.TimeFormat != "HH:MM:SS" {
		t.Errorf("expected time_format %q, got %q", "HH:MM:SS", cfg.TimeFormat)
	}
}

func TestValidateOutputDir_Valid(t *testing.T) {
	dir := t.TempDir()
	if err := ValidateOutputDir(dir); err != nil {
		t.Errorf("ValidateOutputDir(%q) unexpected error: %v", dir, err)
	}
}

func TestValidateOutputDir_NonExistent(t *testing.T) {
	err := ValidateOutputDir("/nonexistent-dir-xyz")
	if err == nil {
		t.Fatal("expected error for non-existent directory")
	}
	if !strings.Contains(err.Error(), "does not exist or is not writable") {
		t.Errorf("expected 'does not exist or is not writable' in error, got: %v", err)
	}
}

func TestValidateOutputDir_NotADir(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "not-a-dir")
	if err := os.WriteFile(tmpFile, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := ValidateOutputDir(tmpFile)
	if err == nil {
		t.Fatal("expected error for file path")
	}
	if !strings.Contains(err.Error(), "is not a directory") {
		t.Errorf("expected 'is not a directory' in error, got: %v", err)
	}
}

func TestValidateOutputDir_Empty(t *testing.T) {
	err := ValidateOutputDir("")
	if err == nil {
		t.Fatal("expected error for empty string")
	}
	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("expected 'cannot be empty' in error, got: %v", err)
	}
}

func TestDetectTimezone(t *testing.T) {
	tz := DetectTimezone()
	if tz == "" {
		t.Fatal("DetectTimezone() returned empty string")
	}
	if _, err := time.LoadLocation(tz); err != nil {
		t.Errorf("DetectTimezone() returned invalid timezone %q: %v", tz, err)
	}
}

func TestDetectLanguage_Supported(t *testing.T) {
	t.Setenv("LANG", "es_ES.UTF-8")
	lang := DetectLanguage()
	if lang != "es" {
		t.Errorf("expected 'es', got %q", lang)
	}
}

func TestDetectLanguage_Unsupported(t *testing.T) {
	t.Setenv("LANG", "xx_XX.UTF-8")
	lang := DetectLanguage()
	if lang != "en" {
		t.Errorf("expected 'en' fallback, got %q", lang)
	}
}

func TestDetectLanguage_Empty(t *testing.T) {
	t.Setenv("LANG", "")
	lang := DetectLanguage()
	if lang != "en" {
		t.Errorf("expected 'en' fallback, got %q", lang)
	}
}

func TestDetectLanguage_POSIX(t *testing.T) {
	t.Setenv("LANG", "C.UTF-8")
	lang := DetectLanguage()
	if lang != "en" {
		t.Errorf("expected 'en' fallback for C locale, got %q", lang)
	}
}

func TestDefaultAlarmProfileField(t *testing.T) {
	cfg := Config{DefaultAlarmProfile: "adhd-default"}
	if cfg.DefaultAlarmProfile != "adhd-default" {
		t.Errorf("expected 'adhd-default', got %q", cfg.DefaultAlarmProfile)
	}
}
