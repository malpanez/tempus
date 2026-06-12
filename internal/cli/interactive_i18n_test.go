package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tempus/internal/config"
	"tempus/internal/i18n"
)

func optionLabelsByValue(t *testing.T, cfg *config.Config, lang string) map[string]string {
	t.Helper()
	tr, err := i18n.NewTranslator(lang)
	if err != nil {
		t.Fatalf("failed to create translator for %s: %v", lang, err)
	}
	opts := alarmProfileOptions(cfg, tr)
	labels := make(map[string]string, len(opts))
	for _, opt := range opts {
		labels[opt.Value] = opt.Key
	}
	return labels
}

func loadDefaultConfig(t *testing.T) *config.Config {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("language: en\n"), 0600); err != nil {
		t.Fatalf("failed to write config fixture: %v", err)
	}
	cfg, err := config.LoadFrom(path)
	if err != nil {
		t.Fatalf("failed to load default config: %v", err)
	}
	return cfg
}

// TestAlarmProfileLabelsDeriveFromConfigDefaults is the anti-drift guard:
// for every alarm profile shipped in the config defaults, the wizard label
// must contain exactly the offsets from the config definition.
func TestAlarmProfileLabelsDeriveFromConfigDefaults(t *testing.T) {
	cfg := loadDefaultConfig(t)
	if len(cfg.AlarmProfiles) == 0 {
		t.Fatal("default config has no alarm profiles")
	}

	labels := optionLabelsByValue(t, cfg, "en")

	for name, offsets := range cfg.AlarmProfiles {
		if len(offsets) == 0 {
			continue // covered by the fixed "None" entry
		}
		label, ok := labels[name]
		if !ok {
			t.Errorf("no wizard option generated for default profile %q", name)
			continue
		}
		want := "(" + strings.Join(offsets, ", ") + ")"
		if !strings.Contains(label, want) {
			t.Errorf("label for %q = %q, must contain offsets %q from config definition", name, label, want)
		}
	}

	if labels["none"] != "None" {
		t.Errorf("expected fixed None entry, got %q", labels["none"])
	}
	if labels["custom"] != "Custom..." {
		t.Errorf("expected fixed Custom... entry, got %q", labels["custom"])
	}
}

func TestAlarmProfileOptionsFromCustomConfig(t *testing.T) {
	cfg := &config.Config{
		AlarmProfiles: map[string][]string{
			"focus-sprint": {"-90m", "-3m"},
			"medication":   {"-42m", "-7m"},
			"empty":        {},
		},
	}

	labels := optionLabelsByValue(t, cfg, "en")

	if got, want := labels["focus-sprint"], "focus-sprint (-90m, -3m)"; got != want {
		t.Errorf("user-defined profile label = %q, want %q", got, want)
	}
	if got, want := labels["medication"], "Medication (-42m, -7m)"; got != want {
		t.Errorf("overridden medication label = %q, want %q (offsets must follow config)", got, want)
	}
	if _, ok := labels["empty"]; ok {
		t.Error("profiles with no offsets must not produce wizard options")
	}
	if _, ok := labels["none"]; !ok {
		t.Error("fixed None entry missing")
	}
	if _, ok := labels["custom"]; !ok {
		t.Error("fixed Custom... entry missing")
	}
}

func TestAlarmProfileDisplayNamesLocalized(t *testing.T) {
	cfg := loadDefaultConfig(t)

	tests := []struct {
		lang    string
		profile string
		want    string
	}{
		{"en", "medication", "Medication"},
		{"es", "medication", "Medicación"},
		{"pt", "medication", "Medicação"},
		{"ga", "medication", "Cógais"},
		{"es", "adhd-default", "ADHD por defecto"},
	}
	for _, tt := range tests {
		t.Run(tt.lang+"/"+tt.profile, func(t *testing.T) {
			labels := optionLabelsByValue(t, cfg, tt.lang)
			label := labels[tt.profile]
			if !strings.HasPrefix(label, tt.want+" (") {
				t.Errorf("label for %s/%s = %q, want prefix %q", tt.lang, tt.profile, label, tt.want+" (")
			}
		})
	}
}
