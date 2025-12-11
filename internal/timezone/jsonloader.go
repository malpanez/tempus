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

	// Zones
	for _, z := range jf.Zones {
		if strings.TrimSpace(z.IANA) == "" {
			continue
		}
		offset := getTimezoneOffset(z.IANA)
		dst := false
		if z.DST != nil {
			dst = *z.DST
		} else {
			dst = hasDST(z.IANA)
		}
		ti := &TimezoneInfo{
			IANA:        z.IANA,
			DisplayName: valueOr(z.DisplayName, z.IANA),
			Country:     valueOr(z.Country, "Unknown"),
			Offset:      offset,
			DST:         dst,
		}
		// Add/overwrite canonical IANA entry
		tm.zones[z.IANA] = ti

		// Per-zone aliases
		for _, a := range z.Aliases {
			if strings.TrimSpace(a) == "" {
				continue
			}
			tm.zones[strings.ToLower(a)] = ti
		}
	}

	// Global aliases
	for alias, iana := range jf.Aliases {
		if zi, ok := tm.zones[iana]; ok {
			tm.zones[strings.ToLower(alias)] = zi
		}
	}
	return nil
}

func valueOr(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
}

// (helpers reused from timezone.go) getTimezoneOffset and hasDST declared there.
// We rely on them; no duplication here.
