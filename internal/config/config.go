package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"

	"tempus/internal/i18n"
)

type Config struct {
	Language         string              `mapstructure:"language" json:"language"`
	Timezone         string              `mapstructure:"timezone" json:"timezone"`
	DateFormat       string              `mapstructure:"date_format" json:"date_format"`
	TimeFormat       string              `mapstructure:"time_format" json:"time_format"`
	OutputDir        string              `mapstructure:"output_dir" json:"output_dir"`
	DefaultTitle     string              `mapstructure:"default_title" json:"default_title"`
	AlarmProfiles    map[string][]string `mapstructure:"alarm_profiles" json:"alarm_profiles"`
	SpellCorrections map[string]string   `mapstructure:"spell_corrections" json:"spell_corrections"`
}

var defaultConfig = Config{
	Language:     "en",
	Timezone:     "UTC",
	DateFormat:   "2006-01-02",
	TimeFormat:   "15:04",
	OutputDir:    ".",
	DefaultTitle: "Event",
	AlarmProfiles: map[string][]string{
		// Evidence-based ADHD profiles (neuroscience research 2024-2025)
		// Spacing based on working memory & prospective memory studies
		"adhd-default":   {"-2h", "-1h", "-30m", "-10m"}, // Optimal spacing for regular events
		"adhd-countdown": {"-1d", "-1h", "-15m", "-5m"},  // For important deadlines/appointments
		"medication":     {"-5m", "-1m", "0m"},           // Triple reminder for medication
		"single":         {"-15m"},                       // Standard single reminder
		"none":           {},                             // No alarms
	},
	SpellCorrections: map[string]string{
		"meetng":       "meeting",
		"meetting":     "meeting",
		"meting":       "meeting",
		"appointmnt":   "appointment",
		"apointment":   "appointment",
		"appointement": "appointment",
		"medicaton":    "medication",
		"mediction":    "medication",
		"medikation":   "medication",
		"breakfst":     "breakfast",
		"brekfast":     "breakfast",
		"brek":         "break",
		"brk":          "break",
		"dinr":         "dinner",
		"diner":        "dinner",
		"prepartion":   "preparation",
		"preperation":  "preparation",
		"doctr":        "doctor",
		"docter":       "doctor",
		"therepy":      "therapy",
		"theraphy":     "therapy",
		"sesion":       "session",
		"sesson":       "session",
		"excersize":    "exercise",
		"excercise":    "exercise",
	},
}

// Load loads configuration from file or creates defaults in memory.
// It reads ~/.config/tempus/config.yaml (or OS-specific dir) with a fallback to current dir.
func Load() (*Config, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return nil, err
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir)
	viper.AddConfigPath(".")

	// Defaults
	viper.SetDefault("language", defaultConfig.Language)
	viper.SetDefault("timezone", defaultConfig.Timezone)
	viper.SetDefault("date_format", defaultConfig.DateFormat)
	viper.SetDefault("time_format", defaultConfig.TimeFormat)
	viper.SetDefault("output_dir", defaultConfig.OutputDir)
	viper.SetDefault("default_title", defaultConfig.DefaultTitle)
	viper.SetDefault("alarm_profiles", defaultConfig.AlarmProfiles)
	viper.SetDefault("spell_corrections", defaultConfig.SpellCorrections)

	// Try to read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
		// Config file not found: continue with defaults
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Set sets a configuration value and persists it to disk.
func (c *Config) Set(key, value string) error {
	viper.Set(key, value)

	// Update struct fields for the running process
	switch key {
	case "language":
		c.Language = value
	case "timezone":
		c.Timezone = value
	case "date_format":
		c.DateFormat = value
	case "time_format":
		c.TimeFormat = value
	case "output_dir":
		c.OutputDir = value
	case "default_title":
		c.DefaultTitle = value
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	return c.Save()
}

// Get returns a configuration value by key.
func (c *Config) Get(key string) (string, error) {
	switch key {
	case "language":
		return c.Language, nil
	case "timezone":
		return c.Timezone, nil
	case "date_format":
		return c.DateFormat, nil
	case "time_format":
		return c.TimeFormat, nil
	case "output_dir":
		return c.OutputDir, nil
	case "default_title":
		return c.DefaultTitle, nil
	default:
		return "", fmt.Errorf("unknown configuration key: %s", key)
	}
}

// GetOrDefault returns the value for key, or def if empty/unknown.
func (c *Config) GetOrDefault(key, def string) string {
	v, err := c.Get(key)
	if err != nil || strings.TrimSpace(v) == "" {
		return def
	}
	return v
}

// List prints all configuration values to stdout.
func (c *Config) List() error {
	fmt.Printf("language: %s\n", c.Language)
	fmt.Printf("timezone: %s\n", c.Timezone)
	fmt.Printf("date_format: %s\n", c.DateFormat)
	fmt.Printf("time_format: %s\n", c.TimeFormat)
	fmt.Printf("output_dir: %s\n", c.OutputDir)
	fmt.Printf("default_title: %s\n", c.DefaultTitle)
	return nil
}

// Save persists the current in-memory configuration to disk.
func (c *Config) Save() error {
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		return err
	}
	configFile := filepath.Join(configDir, "config.yaml")
	return viper.WriteConfigAs(configFile)
}

// getConfigDir returns the platform-appropriate config directory:
//   - Linux/macOS: $XDG_CONFIG_HOME/tempus or ~/.config/tempus
//   - Windows: %AppData%\Tempus
//
// Falls back to ~/.tempus if UserConfigDir is unavailable.
func getConfigDir() (string, error) {
	// Check XDG_CONFIG_HOME first (respects test environment variables)
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "tempus"), nil
	}

	// Use os.UserConfigDir() for platform-specific defaults
	if base, err := os.UserConfigDir(); err == nil && strings.TrimSpace(base) != "" {
		return filepath.Join(base, "tempus"), nil
	}

	// Final fallback to ~/.tempus
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".tempus"), nil
}

// ConfigDir returns the directory used to store Tempus configuration files.
func ConfigDir() (string, error) {
	return getConfigDir()
}

// GetAlarmProfile returns the alarm triggers for a named profile.
// Returns nil if the profile doesn't exist.
func (c *Config) GetAlarmProfile(name string) []string {
	if c.AlarmProfiles == nil {
		return nil
	}
	profile, exists := c.AlarmProfiles[name]
	if !exists {
		return nil
	}
	return profile
}

// ListAlarmProfiles returns all available alarm profile names.
func (c *Config) ListAlarmProfiles() []string {
	if c.AlarmProfiles == nil {
		return []string{}
	}
	profiles := make([]string, 0, len(c.AlarmProfiles))
	for name := range c.AlarmProfiles {
		profiles = append(profiles, name)
	}
	return profiles
}

// ValidateTimezone checks the TZ identifier using the system tz database.
func ValidateTimezone(tz string) error {
	if strings.TrimSpace(tz) == "" {
		return fmt.Errorf("timezone cannot be empty")
	}
	if _, err := time.LoadLocation(tz); err != nil {
		return fmt.Errorf("invalid timezone %q: %w", tz, err)
	}
	return nil
}

// ValidateLanguage checks if a language code is supported.
func ValidateLanguage(lang string) error {
	normalized := strings.ToLower(strings.TrimSpace(lang))
	if normalized == "" {
		return fmt.Errorf("language cannot be empty")
	}
	if i18n.IsSupportedLanguage(normalized) {
		return nil
	}
	return fmt.Errorf("unsupported language: %s (supported: %v)", lang, i18n.SupportedLanguages())
}
