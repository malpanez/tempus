package timezone

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// TimezoneInfo contains information about a timezone
type TimezoneInfo struct {
	IANA        string
	DisplayName string
	Country     string
	Offset      string // computed at load-time for "now"
	DST         bool   // whether this zone observes DST (best effort)
}

// TimezoneManager handles timezone operations
type TimezoneManager struct {
	zones map[string]*TimezoneInfo // includes IANA keys + aliases pointing to same *TimezoneInfo
}

// NewTimezoneManager creates a new timezone manager
func NewTimezoneManager() *TimezoneManager {
	tm := &TimezoneManager{
		zones: make(map[string]*TimezoneInfo),
	}
	// 1) Load full IANA catalog from embedded zone1970.tab
	tm.loadFromZoneTab()
	// 2) Add/override a curated set with nicer display names + aliases
	tm.loadCommonTimezones()
	return tm
}

// GetTimezone returns timezone info by IANA name or alias
func (tm *TimezoneManager) GetTimezone(name string) (*TimezoneInfo, error) {
	// Exact
	if zone, exists := tm.zones[name]; exists {
		return zone, nil
	}
	// Case-insensitive
	for key, zone := range tm.zones {
		if strings.EqualFold(key, name) {
			return zone, nil
		}
	}
	// Try system
	if _, err := time.LoadLocation(name); err == nil {
		return &TimezoneInfo{
			IANA:        name,
			DisplayName: displayFromIANA(name),
			Country:     "Unknown",
			Offset:      getTimezoneOffset(name),
			DST:         hasDST(name),
		}, nil
	}
	return nil, fmt.Errorf("timezone not found: %s", name)
}

// ListTimezones returns unique zones by IANA (deduplicated from aliases)
func (tm *TimezoneManager) ListTimezones() []*TimezoneInfo {
	zones := make([]*TimezoneInfo, 0, len(tm.zones))
	seen := make(map[string]bool)
	for _, zone := range tm.zones {
		if zone == nil || zone.IANA == "" {
			continue
		}
		if seen[zone.IANA] {
			continue
		}
		seen[zone.IANA] = true
		zones = append(zones, zone)
	}
	// Sort by DisplayName then IANA for stable, friendly output
	sort.Slice(zones, func(i, j int) bool {
		if zones[i].DisplayName == zones[j].DisplayName {
			return zones[i].IANA < zones[j].IANA
		}
		return zones[i].DisplayName < zones[j].DisplayName
	})
	return zones
}

// GetEuropeanTimezones returns European timezones (unique)
func (tm *TimezoneManager) GetEuropeanTimezones() []*TimezoneInfo {
	all := tm.ListTimezones()
	europe := make([]*TimezoneInfo, 0, len(all))
	for _, z := range all {
		if strings.HasPrefix(z.IANA, "Europe/") || strings.HasPrefix(z.IANA, "Atlantic/") {
			europe = append(europe, z)
		}
	}
	sort.Slice(europe, func(i, j int) bool {
		if europe[i].DisplayName == europe[j].DisplayName {
			return europe[i].IANA < europe[j].IANA
		}
		return europe[i].DisplayName < europe[j].DisplayName
	})
	return europe
}

// ConvertTime converts time from one timezone to another (labels respected)
func (tm *TimezoneManager) ConvertTime(t time.Time, fromTZ, toTZ string) (time.Time, error) {
	fromLoc, err := time.LoadLocation(fromTZ)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid source timezone %s: %w", fromTZ, err)
	}
	toLoc, err := time.LoadLocation(toTZ)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid destination timezone %s: %w", toTZ, err)
	}
	timeInSource := t.In(fromLoc)
	return timeInSource.In(toLoc), nil
}

// ValidateTimezone checks if a timezone is valid
func (tm *TimezoneManager) ValidateTimezone(tz string) error {
	_, err := tm.GetTimezone(tz)
	return err
}

// SuggestTimezone suggests a timezone based on partial input
func (tm *TimezoneManager) SuggestTimezone(input string) []*TimezoneInfo {
	q := strings.ToLower(strings.TrimSpace(input))
	if q == "" {
		return nil
	}
	results := make([]*TimezoneInfo, 0, 10)
	seen := map[string]bool{}
	for key, zone := range tm.zones {
		if zone == nil {
			continue
		}
		if strings.Contains(strings.ToLower(key), q) ||
			strings.Contains(strings.ToLower(zone.DisplayName), q) ||
			strings.Contains(strings.ToLower(zone.IANA), q) ||
			strings.Contains(strings.ToLower(zone.Country), q) {
			if !seen[zone.IANA] {
				results = append(results, zone)
				seen[zone.IANA] = true
				if len(results) >= 10 {
					break
				}
			}
		}
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].DisplayName == results[j].DisplayName {
			return results[i].IANA < results[j].IANA
		}
		return results[i].DisplayName < results[j].DisplayName
	})
	return results
}

// ---------- Loaders ----------

// loadFromZoneTab loads the full IANA catalog from the embedded zone1970.tab.
func (tm *TimezoneManager) loadFromZoneTab() {
	rows := parseZone1970Tab()
	if len(rows) == 0 {
		return
	}
	for _, r := range rows {
		tz := strings.TrimSpace(r.TZ)
		if tz == "" {
			continue
		}
		if _, exists := tm.zones[tz]; exists {
			continue
		}
		info := &TimezoneInfo{
			IANA:        tz,
			DisplayName: displayFromIANA(tz),        // city only
			Country:     countryNameFromCodes(r.CC), // "IE", "BR,AR", etc.
			Offset:      getTimezoneOffset(tz),
			DST:         hasDST(tz),
		}
		tm.zones[tz] = info
	}
}

// loadCommonTimezones adds a curated subset + friendly aliases (overrides DisplayName)
func (tm *TimezoneManager) loadCommonTimezones() {
	type seed struct {
		IANA, Display, Country string
	}
	seeds := []seed{
		// Spain and territories (Display = city/area only; Country separate)
		{"Europe/Madrid", "Madrid", "Spain"},
		{"Atlantic/Canary", "Canary Islands", "Spain"},
		{"Africa/Ceuta", "Ceuta / Melilla", "Spain"},

		// Ireland and UK
		{"Europe/Dublin", "Dublin", "Ireland"},
		{"Europe/London", "London", "United Kingdom"},

		// Other European capitals
		{"Europe/Paris", "Paris", "France"},
		{"Europe/Berlin", "Berlin", "Germany"},
		{"Europe/Rome", "Rome", "Italy"},
		{"Europe/Lisbon", "Lisbon", "Portugal"},
		{"Europe/Amsterdam", "Amsterdam", "Netherlands"},
		{"Europe/Brussels", "Brussels", "Belgium"},
		{"Europe/Vienna", "Vienna", "Austria"},
		{"Europe/Zurich", "Zurich", "Switzerland"},

		// Americas (incl. Brazil you asked about)
		{"America/New_York", "New York", "United States"},
		{"America/Los_Angeles", "Los Angeles", "United States"},
		{"America/Chicago", "Chicago", "United States"},
		{"America/Mexico_City", "Mexico City", "Mexico"},
		{"America/Sao_Paulo", "São Paulo", "Brazil"},
		{"America/Campo_Grande", "Campo Grande", "Brazil"},

		// APAC
		{"Asia/Tokyo", "Tokyo", "Japan"},
		{"Asia/Shanghai", "Shanghai", "China"},
		{"Australia/Sydney", "Sydney", "Australia"},

		// UTC/GMT
		{"UTC", "UTC", "Universal"},
		{"GMT", "GMT", "Universal"},
	}

	for _, s := range seeds {
		info := &TimezoneInfo{
			IANA:        s.IANA,
			DisplayName: s.Display,
			Country:     s.Country,
			Offset:      getTimezoneOffset(s.IANA),
			DST:         hasDST(s.IANA),
		}
		tm.zones[s.IANA] = info
	}

	// Aliases (lowercase keys) -> map to the same *TimezoneInfo
	alias := func(key, iana string) { tm.zones[strings.ToLower(key)] = tm.zones[iana] }

	// Spain
	alias("madrid", "Europe/Madrid")
	alias("spain", "Europe/Madrid")
	alias("canarias", "Atlantic/Canary")
	alias("canary", "Atlantic/Canary")
	alias("ceuta", "Africa/Ceuta")
	alias("melilla", "Africa/Ceuta")

	// Ireland / UK
	alias("dublin", "Europe/Dublin")
	alias("ireland", "Europe/Dublin")
	alias("london", "Europe/London")
	alias("uk", "Europe/London")

	// Brazil (requested)
	alias("sao paulo", "America/Sao_Paulo")
	alias("são paulo", "America/Sao_Paulo")
	alias("porto alegre", "America/Sao_Paulo")
	alias("pelotas", "America/Sao_Paulo")
	alias("florianopolis", "America/Sao_Paulo")
	alias("florianópolis", "America/Sao_Paulo")
	alias("rio", "America/Sao_Paulo")
	alias("rio de janeiro", "America/Sao_Paulo")
	alias("campo grande", "America/Campo_Grande")
	alias("campo_grande", "America/Campo_Grande")
	alias("ponta pora", "America/Campo_Grande")
	alias("ponta porã", "America/Campo_Grande")
	alias("ponta-pora", "America/Campo_Grande")
	alias("dourados", "America/Campo_Grande")

	// Common references
	alias("utc", "UTC")
	alias("gmt", "GMT")
}

// ---------- Helpers ----------

func displayFromIANA(iana string) string {
	// "Area/City_Name" -> "City Name"
	if i := strings.LastIndex(iana, "/"); i >= 0 {
		part := iana[i+1:]
		return strings.ReplaceAll(part, "_", " ")
	}
	return strings.ReplaceAll(iana, "_", " ")
}

func countryNameFromCodes(cc string) string {
	parts := strings.Split(cc, ",")
	names := make([]string, 0, len(parts))
	for _, p := range parts {
		code := strings.TrimSpace(p)
		if full, ok := countryNames[code]; ok {
			names = append(names, full)
		} else if code != "" {
			names = append(names, code) // fallback to code
		}
	}
	return strings.Join(names, ", ")
}

// getTimezoneOffset calculates the current offset for a timezone
func getTimezoneOffset(tzName string) string {
	loc, err := time.LoadLocation(tzName)
	if err != nil {
		return "Unknown"
	}
	now := time.Now().In(loc)
	_, offset := now.Zone()
	hours := offset / 3600
	minutes := (offset % 3600) / 60
	sign := "+"
	if offset < 0 {
		sign = "-"
		hours = -hours
		minutes = -minutes
	}
	return fmt.Sprintf("%s%02d:%02d", sign, hours, minutes)
}

// hasDST tries to detect if a zone observes DST by comparing Jan/Jul offsets.
func hasDST(tzName string) bool {
	loc, err := time.LoadLocation(tzName)
	if err != nil {
		return false
	}
	year := time.Now().Year()
	january := time.Date(year, time.January, 1, 0, 0, 0, 0, loc)
	july := time.Date(year, time.July, 1, 0, 0, 0, 0, loc)
	_, offJan := january.Zone()
	_, offJul := july.Zone()
	return offJan != offJul
}

// GetFlightTimezones returns common timezone pairs for flights
func (tm *TimezoneManager) GetFlightTimezones() map[string][]string {
	return map[string][]string{
		"Spain to Ireland/UK": {
			"Europe/Madrid", "Europe/Dublin",
			"Europe/Madrid", "Europe/London",
			"Atlantic/Canary", "Europe/Dublin",
			"Atlantic/Canary", "Europe/London",
		},
		"Spain to Europe": {
			"Europe/Madrid", "Europe/Paris",
			"Europe/Madrid", "Europe/Berlin",
			"Europe/Madrid", "Europe/Rome",
		},
		"Ireland/UK to Europe": {
			"Europe/Dublin", "Europe/Paris",
			"Europe/London", "Europe/Berlin",
			"Europe/Dublin", "Europe/Madrid",
		},
		"Ireland to Brazil": {
			"Europe/Dublin", "America/Sao_Paulo",
			"Europe/Dublin", "America/Campo_Grande",
		},
		"Transatlantic": {
			"Europe/Madrid", "America/New_York",
			"Europe/Dublin", "America/New_York",
			"Europe/London", "America/New_York",
		},
	}
}

// IsEuropeanTimezone checks if a timezone is in Europe
func (tm *TimezoneManager) IsEuropeanTimezone(tz string) bool {
	return strings.HasPrefix(tz, "Europe/") ||
		strings.HasPrefix(tz, "Atlantic/") ||
		tz == "GMT" || tz == "UTC"
}

// GetTimezoneAbbreviation returns a typical abbreviation for a timezone
func (tm *TimezoneManager) GetTimezoneAbbreviation(tz string) string {
	abbreviations := map[string]string{
		"Europe/Madrid":        "CET/CEST",
		"Europe/Dublin":        "GMT/IST",
		"Europe/London":        "GMT/BST",
		"Atlantic/Canary":      "WET/WEST",
		"Europe/Paris":         "CET/CEST",
		"Europe/Berlin":        "CET/CEST",
		"America/New_York":     "EST/EDT",
		"America/Los_Angeles":  "PST/PDT",
		"America/Chicago":      "CST/CDT",
		"America/Sao_Paulo":    "BRT", // Brazil (no DST currently)
		"America/Campo_Grande": "AMT",
		"Asia/Tokyo":           "JST",
		"Asia/Shanghai":        "CST",
		"Australia/Sydney":     "AEST/AEDT",
		"America/Mexico_City":  "CST/CDT",
		"UTC":                  "UTC",
		"GMT":                  "GMT",
	}
	if abbr, ok := abbreviations[tz]; ok {
		return abbr
	}
	return tz
}

// Minimal ISO country code -> name mapping (fallback to code if missing).
var countryNames = map[string]string{
	"AD": "Andorra", "AE": "United Arab Emirates", "AF": "Afghanistan",
	"AG": "Antigua and Barbuda", "AI": "Anguilla", "AL": "Albania",
	"AM": "Armenia", "AO": "Angola", "AQ": "Antarctica", "AR": "Argentina",
	"AT": "Austria", "AU": "Australia", "AW": "Aruba", "AX": "Aland Islands",
	"AZ": "Azerbaijan",

	"BA": "Bosnia and Herzegovina", "BB": "Barbados", "BD": "Bangladesh",
	"BE": "Belgium", "BF": "Burkina Faso", "BG": "Bulgaria", "BH": "Bahrain",
	"BI": "Burundi", "BJ": "Benin", "BM": "Bermuda", "BN": "Brunei Darussalam",
	"BO": "Bolivia", "BR": "Brazil", "BS": "Bahamas", "BT": "Bhutan",
	"BW": "Botswana", "BY": "Belarus", "BZ": "Belize",

	"CA": "Canada", "CD": "Congo (DRC)", "CF": "Central African Republic",
	"CG": "Congo", "CH": "Switzerland", "CI": "Cote d’Ivoire", "CL": "Chile",
	"CM": "Cameroon", "CN": "China", "CO": "Colombia", "CR": "Costa Rica",
	"CU": "Cuba", "CV": "Cabo Verde", "CY": "Cyprus", "CZ": "Czechia",

	"DE": "Germany", "DK": "Denmark", "DO": "Dominican Republic",
	"DZ": "Algeria",

	"EC": "Ecuador", "EE": "Estonia", "EG": "Egypt", "ES": "Spain",
	"ET": "Ethiopia",

	"FI": "Finland", "FJ": "Fiji", "FK": "Falkland Islands", "FO": "Faroe Islands",
	"FR": "France",

	"GB": "United Kingdom", "GD": "Grenada", "GE": "Georgia", "GF": "French Guiana",
	"GG": "Guernsey", "GH": "Ghana", "GI": "Gibraltar", "GL": "Greenland",
	"GP": "Guadeloupe", "GR": "Greece", "GT": "Guatemala", "GU": "Guam",
	"GY": "Guyana",

	"HK": "Hong Kong", "HN": "Honduras", "HR": "Croatia", "HT": "Haiti",
	"HU": "Hungary",

	"IE": "Ireland", "IL": "Israel", "IM": "Isle of Man", "IN": "India",
	"IQ": "Iraq", "IR": "Iran", "IS": "Iceland", "IT": "Italy",

	"JM": "Jamaica", "JO": "Jordan", "JP": "Japan",

	"KE": "Kenya", "KG": "Kyrgyzstan", "KH": "Cambodia", "KR": "Korea (South)",
	"KW": "Kuwait", "KY": "Cayman Islands", "KZ": "Kazakhstan",

	"LA": "Laos", "LB": "Lebanon", "LI": "Liechtenstein", "LK": "Sri Lanka",
	"LT": "Lithuania", "LU": "Luxembourg", "LV": "Latvia", "LY": "Libya",

	"MA": "Morocco", "MC": "Monaco", "MD": "Moldova", "ME": "Montenegro",
	"MF": "Saint Martin", "MG": "Madagascar", "MK": "North Macedonia",
	"MM": "Myanmar", "MN": "Mongolia", "MO": "Macao", "MQ": "Martinique",
	"MT": "Malta", "MU": "Mauritius", "MX": "Mexico", "MY": "Malaysia",

	"NA": "Namibia", "NC": "New Caledonia", "NG": "Nigeria", "NI": "Nicaragua",
	"NL": "Netherlands", "NO": "Norway", "NP": "Nepal", "NZ": "New Zealand",

	"PA": "Panama", "PE": "Peru", "PF": "French Polynesia", "PG": "Papua New Guinea",
	"PH": "Philippines", "PK": "Pakistan", "PL": "Poland", "PM": "Saint Pierre and Miquelon",
	"PR": "Puerto Rico", "PT": "Portugal", "PY": "Paraguay",

	"QA": "Qatar",

	"RE": "Reunion", "RO": "Romania", "RS": "Serbia", "RU": "Russia",
	"RW": "Rwanda",

	"SA": "Saudi Arabia", "SC": "Seychelles", "SE": "Sweden", "SG": "Singapore",
	"SI": "Slovenia", "SJ": "Svalbard and Jan Mayen", "SK": "Slovakia",
	"SM": "San Marino", "SN": "Senegal", "SR": "Suriname",
	"SV": "El Salvador", "SY": "Syria",

	"TC": "Turks and Caicos Islands", "TH": "Thailand", "TN": "Tunisia",
	"TR": "Turkiye", "TT": "Trinidad and Tobago", "TW": "Taiwan", "TZ": "Tanzania",

	"UA": "Ukraine", "UG": "Uganda", "US": "United States", "UY": "Uruguay",
	"UZ": "Uzbekistan",

	"VA": "Vatican City", "VE": "Venezuela", "VG": "Virgin Islands (UK)",
	"VI": "Virgin Islands (US)", "VN": "Viet Nam",

	"WF": "Wallis and Futuna",

	"YE": "Yemen", "YT": "Mayotte", "ZA": "South Africa", "ZM": "Zambia",
	"ZW": "Zimbabwe",
}
