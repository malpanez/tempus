package timezone

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// jsonZone is a single zone entry in a JSON file.
type jsonZone struct {
	IANA        string   `json:"iana"`
	DisplayName string   `json:"display_name"`
	Country     string   `json:"country"`
	DST         *bool    `json:"dst,omitempty"`     // optional; if nil we detect
	Aliases     []string `json:"aliases,omitempty"` // per-zone extra aliases
}

// jsonFile is the structure of a timezones JSON.
type jsonFile struct {
	Zones   []jsonZone        `json:"zones"`
	Aliases map[string]string `json:"aliases,omitempty"` // global alias -> IANA
}

// LoadJSONDir loads all *.json files from dir (non-recursive).
func (tm *TimezoneManager) LoadJSONDir(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		// non-fatal for callers; just skip if directory is missing
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(e.Name()), ".json") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		if err := tm.loadJSONFile(path); err != nil {
			// Keep going, but report the error to stderr
			fmt.Fprintf(os.Stderr, "warning: failed to load %s: %v\n", path, err)
		}
	}
	return nil
}

// LoadDefaultJSONDirs tries commonly useful directories (non-fatal).
// 1) User config dir: <UserConfigDir>/tempus/timezones
// 2) Repo-local examples: internal/timezone/json
func (tm *TimezoneManager) LoadDefaultJSONDirs() {
	// User config dir
	if cdir, err := os.UserConfigDir(); err == nil {
		_ = tm.LoadJSONDir(filepath.Join(cdir, "tempus", "timezones"))
	}
	// Repo-local examples (useful in dev)
	_ = tm.LoadJSONDir(filepath.Join("internal", "timezone", "json"))
}

func (tm *TimezoneManager) loadJSONFile(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var jf jsonFile
	if err := json.Unmarshal(b, &jf); err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	tm.processZones(jf.Zones)
	tm.processGlobalAliases(jf.Aliases)
	return nil
}

func (tm *TimezoneManager) processZones(zones []jsonZone) {
	for _, z := range zones {
		if strings.TrimSpace(z.IANA) == "" {
			continue
		}
		ti := tm.createTimezoneInfo(z)
		tm.zones[z.IANA] = ti
		tm.processZoneAliases(z.Aliases, ti)
	}
}

func (tm *TimezoneManager) createTimezoneInfo(z jsonZone) *TimezoneInfo {
	dst := hasDST(z.IANA)
	if z.DST != nil {
		dst = *z.DST
	}
	return &TimezoneInfo{
		IANA:        z.IANA,
		DisplayName: valueOr(z.DisplayName, z.IANA),
		Country:     valueOr(z.Country, "Unknown"),
		Offset:      getTimezoneOffset(z.IANA),
		DST:         dst,
	}
}

func (tm *TimezoneManager) processZoneAliases(aliases []string, ti *TimezoneInfo) {
	for _, a := range aliases {
		if strings.TrimSpace(a) == "" {
			continue
		}
		tm.zones[strings.ToLower(a)] = ti
	}
}

func (tm *TimezoneManager) processGlobalAliases(aliases map[string]string) {
	for alias, iana := range aliases {
		if zi, ok := tm.zones[iana]; ok {
			tm.zones[strings.ToLower(alias)] = zi
		}
	}
}

func valueOr(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
}

// (helpers reused from timezone.go) getTimezoneOffset and hasDST declared there.
// We rely on them; no duplication here.
