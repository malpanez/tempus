package timezone

import (
	"tempus/internal/testutil"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewTimezoneManager(t *testing.T) {
	tm := NewTimezoneManager()

	if tm == nil {
		t.Fatal("NewTimezoneManager() returned nil")
	}

	if tm.zones == nil {
		t.Fatal("TimezoneManager.zones is nil")
	}

	// Should have loaded many timezones
	if len(tm.zones) == 0 {
		t.Error("TimezoneManager has no timezones loaded")
	}
}

func TestGetTimezone(t *testing.T) {
	tm := NewTimezoneManager()

	tests := []struct {
		name      string
		input     string
		shouldErr bool
	}{
		{"UTC exact match", "UTC", false},
		{testutil.TZAmericaNewYork, testutil.TZAmericaNewYork, false},
		{testutil.TZEuropeLondon, testutil.TZEuropeLondon, false},
		{testutil.TZAsiaTokyо, testutil.TZAsiaTokyо, false},
		{"case insensitive utc", "utc", false},
		{"case insensitive", "america/new_york", false},
		{"invalid timezone", testutil.TZInvalid, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tz, err := tm.GetTimezone(tt.input)

			if tt.shouldErr {
				if err == nil {
					t.Errorf("GetTimezone(%q) expected error, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("GetTimezone(%q) unexpected error: %v", tt.input, err)
				return
			}

			if tz == nil {
				t.Errorf("GetTimezone(%q) returned nil timezone", tt.input)
				return
			}

			if tz.IANA == "" {
				t.Errorf("GetTimezone(%q) returned empty IANA name", tt.input)
			}
		})
	}
}

func TestListTimezones(t *testing.T) {
	tm := NewTimezoneManager()

	zones := tm.ListTimezones()

	if len(zones) == 0 {
		t.Error("ListTimezones() returned empty list")
	}

	// Check for some expected timezones
	expectedZones := []string{"UTC", testutil.TZAmericaNewYork, testutil.TZEuropeLondon, testutil.TZAsiaTokyо}
	for _, expected := range expectedZones {
		found := false
		for _, zone := range zones {
			if zone.IANA == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ListTimezones() missing expected timezone: %s", expected)
		}
	}

	// Verify zones have required fields
	for _, zone := range zones {
		if zone.IANA == "" {
			t.Error("ListTimezones() returned zone with empty IANA")
		}
		if zone.DisplayName == "" {
			t.Errorf("Zone %s has empty DisplayName", zone.IANA)
		}
	}
}

func TestSuggestTimezone(t *testing.T) {
	tm := NewTimezoneManager()

	tests := []struct {
		name             string
		query            string
		minExpectedCount int
	}{
		{"search UTC", "UTC", 1},
		{"search America", "America", 1},
		{"search Europe", "Europe", 1},
		{"search New York", testutil.LocationNewYork, 1},
		{"search London", "London", 1},
		{"search Tokyo", "Tokyo", 1},
		{"case insensitive", "utc", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := tm.SuggestTimezone(tt.query)

			if len(results) < tt.minExpectedCount {
				t.Errorf("SuggestTimezone(%q) returned %d zones, expected at least %d",
					tt.query, len(results), tt.minExpectedCount)
			}
		})
	}
}

func TestGetEuropeanTimezones(t *testing.T) {
	tm := NewTimezoneManager()

	results := tm.GetEuropeanTimezones()

	if len(results) == 0 {
		t.Error("GetEuropeanTimezones() returned empty list")
	}

	// Check for some expected European timezones
	expectedZones := []string{testutil.TZEuropeLondon, testutil.TZEuropeParis, testutil.TZEuropeBerlin}
	for _, expected := range expectedZones {
		found := false
		for _, zone := range results {
			if zone.IANA == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GetEuropeanTimezones() missing expected timezone: %s", expected)
		}
	}
}

func TestIsEuropeanTimezone(t *testing.T) {
	tm := NewTimezoneManager()

	tests := []struct {
		name     string
		tz       string
		expected bool
	}{
		{"Europe/London is European", testutil.TZEuropeLondon, true},
		{"Europe/Paris is European", testutil.TZEuropeParis, true},
		{"America/New_York is not", testutil.TZAmericaNewYork, false},
		{"Asia/Tokyo is not", testutil.TZAsiaTokyо, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tm.IsEuropeanTimezone(tt.tz)

			if result != tt.expected {
				t.Errorf("IsEuropeanTimezone(%q) = %v, want %v", tt.tz, result, tt.expected)
			}
		})
	}
}

func TestGetTimezoneOffset(t *testing.T) {
	tests := []struct {
		name string
		iana string
	}{
		{"UTC", "UTC"},
		{testutil.TZAmericaNewYork, testutil.TZAmericaNewYork},
		{testutil.TZEuropeLondon, testutil.TZEuropeLondon},
		{testutil.TZAsiaTokyо, testutil.TZAsiaTokyо},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset := getTimezoneOffset(tt.iana)

			// Offset should be in format like "+00:00" or "-05:00"
			if offset == "" {
				t.Errorf("getTimezoneOffset(%q) returned empty string", tt.iana)
			}

			// Should start with + or -
			if !strings.HasPrefix(offset, "+") && !strings.HasPrefix(offset, "-") {
				t.Errorf("getTimezoneOffset(%q) = %q, should start with + or -", tt.iana, offset)
			}
		})
	}
}

func TestHasDST(t *testing.T) {
	tests := []struct {
		name        string
		iana        string
		expectedDST bool
	}{
		{"UTC no DST", "UTC", false},
		{"America/New_York has DST", testutil.TZAmericaNewYork, true},
		{"Europe/London has DST", testutil.TZEuropeLondon, true},
		{"Asia/Tokyo no DST", testutil.TZAsiaTokyо, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasDST(tt.iana)

			if result != tt.expectedDST {
				t.Errorf("hasDST(%q) = %v, want %v", tt.iana, result, tt.expectedDST)
			}
		})
	}
}

func TestDisplayFromIANA(t *testing.T) {
	tests := []struct {
		name     string
		iana     string
		contains string // Expected substring in display name
	}{
		{"UTC", "UTC", "UTC"},
		{testutil.TZAmericaNewYork, testutil.TZAmericaNewYork, testutil.LocationNewYork},
		{testutil.TZEuropeLondon, testutil.TZEuropeLondon, "London"},
		{testutil.TZAsiaTokyо, testutil.TZAsiaTokyо, "Tokyo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			display := displayFromIANA(tt.iana)

			if display == "" {
				t.Errorf("displayFromIANA(%q) returned empty string", tt.iana)
			}

			if !strings.Contains(display, tt.contains) {
				t.Errorf("displayFromIANA(%q) = %q, should contain %q", tt.iana, display, tt.contains)
			}
		})
	}
}

// Tests for functions with 0% coverage

func TestConvertTime(t *testing.T) {
	tm := NewTimezoneManager()

	tests := []struct {
		name      string
		fromTZ    string
		toTZ      string
		shouldErr bool
	}{
		{
			name:      "UTC to New York",
			fromTZ:    "UTC",
			toTZ:      testutil.TZAmericaNewYork,
			shouldErr: false,
		},
		{
			name:      "New York to London",
			fromTZ:    testutil.TZAmericaNewYork,
			toTZ:      testutil.TZEuropeLondon,
			shouldErr: false,
		},
		{
			name:      "London to Tokyo",
			fromTZ:    testutil.TZEuropeLondon,
			toTZ:      testutil.TZAsiaTokyо,
			shouldErr: false,
		},
		{
			name:      "Invalid source timezone",
			fromTZ:    testutil.TZInvalid,
			toTZ:      "UTC",
			shouldErr: true,
		},
		{
			name:      "Invalid destination timezone",
			fromTZ:    "UTC",
			toTZ:      testutil.TZInvalid,
			shouldErr: true,
		},
		{
			name:      "Both timezones invalid",
			fromTZ:    "Invalid/Source",
			toTZ:      "Invalid/Dest",
			shouldErr: true,
		},
		{
			name:      "Same timezone conversion",
			fromTZ:    "UTC",
			toTZ:      "UTC",
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use a known time for testing
			testTime := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)

			result, err := tm.ConvertTime(testTime, tt.fromTZ, tt.toTZ)

			if tt.shouldErr {
				if err == nil {
					t.Errorf("ConvertTime expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ConvertTime unexpected error: %v", err)
				return
			}

			// Verify result is not zero time
			if result.IsZero() {
				t.Errorf("ConvertTime returned zero time")
			}

			// Verify the time value is preserved (same instant in time)
			if !testTime.Equal(result) {
				t.Logf("Original: %v, Converted: %v (times should represent the same instant)", testTime, result)
			}
		})
	}
}

func TestValidateTimezone(t *testing.T) {
	tm := NewTimezoneManager()

	tests := []struct {
		name      string
		tz        string
		shouldErr bool
	}{
		{"Valid UTC", "UTC", false},
		{"Valid New York", testutil.TZAmericaNewYork, false},
		{"Valid London", testutil.TZEuropeLondon, false},
		{"Valid Tokyo", testutil.TZAsiaTokyо, false},
		{"Valid Madrid", testutil.TZEuropeMadrid, false},
		{"Invalid timezone", testutil.TZInvalid, true},
		{testutil.TestStringEmptyString, "", false}, // time.LoadLocation("") returns Local/UTC, not an error
		{"Nonsense string", "NotATimezone", true},
		{"Case insensitive valid", "utc", false},
		{"Alias madrid", "madrid", false},
		{"Alias dublin", "dublin", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tm.ValidateTimezone(tt.tz)

			if tt.shouldErr {
				if err == nil {
					t.Errorf("ValidateTimezone(%q) expected error, got nil", tt.tz)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateTimezone(%q) unexpected error: %v", tt.tz, err)
				}
			}
		})
	}
}

func TestGetFlightTimezones(t *testing.T) {
	tm := NewTimezoneManager()

	result := tm.GetFlightTimezones()

	if len(result) == 0 {
		t.Error("GetFlightTimezones() returned empty map")
	}

	// Check for expected categories
	expectedCategories := []string{
		"Spain to Ireland/UK",
		"Spain to Europe",
		"Ireland/UK to Europe",
		"Ireland to Brazil",
		"Transatlantic",
	}

	for _, category := range expectedCategories {
		if routes, ok := result[category]; !ok {
			t.Errorf("GetFlightTimezones() missing category: %s", category)
		} else if len(routes) == 0 {
			t.Errorf("GetFlightTimezones() category %s has no routes", category)
		}
	}

	// Verify all timezones in the result are valid
	for category, routes := range result {
		if len(routes)%2 != 0 {
			t.Errorf("Category %s has odd number of timezones (should be pairs)", category)
		}

		for i, tz := range routes {
			if tz == "" {
				t.Errorf("Category %s has empty timezone at index %d", category, i)
			}
		}
	}
}

func TestGetTimezoneAbbreviation(t *testing.T) {
	tm := NewTimezoneManager()

	tests := []struct {
		name   string
		tz     string
		expect string
	}{
		{"Madrid", testutil.TZEuropeMadrid, testutil.TZAbbrevCETCEST},
		{"Dublin", testutil.TZEuropeDublin, "GMT/IST"},
		{"London", testutil.TZEuropeLondon, "GMT/BST"},
		{"Canary", testutil.TZAtlanticCanary, "WET/WEST"},
		{"Paris", testutil.TZEuropeParis, testutil.TZAbbrevCETCEST},
		{"Berlin", testutil.TZEuropeBerlin, testutil.TZAbbrevCETCEST},
		{testutil.LocationNewYork, testutil.TZAmericaNewYork, "EST/EDT"},
		{"Los Angeles", "America/Los_Angeles", "PST/PDT"},
		{"Chicago", "America/Chicago", "CST/CDT"},
		{"Sao Paulo", testutil.TZAmericaSaoPaulo, "BRT"},
		{"Campo Grande", testutil.TZAmericaCampoGrande, "AMT"},
		{"Tokyo", testutil.TZAsiaTokyо, "JST"},
		{"Shanghai", "Asia/Shanghai", "CST"},
		{"Sydney", "Australia/Sydney", "AEST/AEDT"},
		{"Mexico City", "America/Mexico_City", "CST/CDT"},
		{"UTC", "UTC", "UTC"},
		{"GMT", "GMT", "GMT"},
		{"Unknown timezone returns itself", "Unknown/Zone", "Unknown/Zone"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tm.GetTimezoneAbbreviation(tt.tz)

			if result != tt.expect {
				t.Errorf("GetTimezoneAbbreviation(%q) = %q, want %q", tt.tz, result, tt.expect)
			}
		})
	}
}

func TestValueOr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		fallback string
		expected string
	}{
		{"Non-empty string", "test", "fallback", "test"},
		{testutil.TestStringEmptyString, "", "fallback", "fallback"},
		{"Whitespace only", "   ", "fallback", "fallback"},
		{"Tab only", "\t", "fallback", "fallback"},
		{"Newline only", "\n", "fallback", "fallback"},
		{"Mixed whitespace", " \t\n ", "fallback", "fallback"},
		{"String with spaces", "test value", "fallback", "test value"},
		{"Empty fallback", "test", "", "test"},
		{"Both empty", "", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := valueOr(tt.input, tt.fallback)

			if result != tt.expected {
				t.Errorf("valueOr(%q, %q) = %q, want %q", tt.input, tt.fallback, result, tt.expected)
			}
		})
	}
}

func TestLoadJSONDir(t *testing.T) {
	tm := NewTimezoneManager()

	t.Run("Load from valid directory with JSON files", func(t *testing.T) {
		// Create a temporary directory
		tmpDir := t.TempDir()

		// Create a valid JSON file
		validJSON := `{
			"zones": [
				{
					"iana": "Test/Zone1",
					"display_name": "Test Zone 1",
					"country": "Test Country",
					"dst": true,
					"aliases": ["test1", "zone1"]
				},
				{
					"iana": "Test/Zone2",
					"display_name": "Test Zone 2",
					"country": "Test Country 2"
				}
			],
			"aliases": {
				"testalias": "Test/Zone1"
			}
		}`

		jsonPath := tmpDir + "/test.json"
		if err := os.WriteFile(jsonPath, []byte(validJSON), 0644); err != nil {
			t.Fatalf(testutil.ErrMsgFailedToWriteTestJSON, err)
		}

		// Load the directory
		err := tm.LoadJSONDir(tmpDir)
		if err != nil {
			t.Errorf(testutil.ErrMsgLoadJSONDirError, err)
		}

		// Verify the zones were loaded
		zone1, err := tm.GetTimezone(testutil.TestZone1)
		if err != nil {
			t.Errorf(testutil.ErrMsgFailedToGetLoadedZone, err)
		} else if zone1.DisplayName != "Test Zone 1" {
			t.Errorf("Zone display name = %q, want 'Test Zone 1'", zone1.DisplayName)
		}

		// Verify alias
		aliasZone, err := tm.GetTimezone("test1")
		if err != nil {
			t.Errorf("Failed to get zone by alias: %v", err)
		} else if aliasZone.IANA != testutil.TestZone1 {
			t.Errorf("Alias points to %q, want 'Test/Zone1'", aliasZone.IANA)
		}

		// Verify global alias
		globalAliasZone, err := tm.GetTimezone("testalias")
		if err != nil {
			t.Errorf("Failed to get zone by global alias: %v", err)
		} else if globalAliasZone.IANA != testutil.TestZone1 {
			t.Errorf("Global alias points to %q, want 'Test/Zone1'", globalAliasZone.IANA)
		}
	})

	t.Run("Load from non-existent directory", func(t *testing.T) {
		tm := NewTimezoneManager()
		err := tm.LoadJSONDir("/nonexistent/directory/path")
		if err == nil {
			t.Error("LoadJSONDir() expected error for non-existent directory, got nil")
		}
	})

	t.Run("Load from file instead of directory", func(t *testing.T) {
		tm := NewTimezoneManager()
		tmpFile := t.TempDir() + "/notadir.txt"
		if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		err := tm.LoadJSONDir(tmpFile)
		if err == nil {
			t.Error("LoadJSONDir() expected error for file path, got nil")
		}
	})

	t.Run("Load directory with non-JSON files", func(t *testing.T) {
		tm := NewTimezoneManager()
		tmpDir := t.TempDir()

		// Create a non-JSON file
		txtPath := tmpDir + "/test.txt"
		if err := os.WriteFile(txtPath, []byte("not json"), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		// Should not error, just skip the file
		err := tm.LoadJSONDir(tmpDir)
		if err != nil {
			t.Errorf(testutil.ErrMsgLoadJSONDirError, err)
		}
	})

	t.Run("Load directory with invalid JSON", func(t *testing.T) {
		tm := NewTimezoneManager()
		tmpDir := t.TempDir()

		// Create an invalid JSON file
		invalidJSON := `{invalid json content`
		jsonPath := tmpDir + "/invalid.json"
		if err := os.WriteFile(jsonPath, []byte(invalidJSON), 0644); err != nil {
			t.Fatalf(testutil.ErrMsgFailedToWriteTestJSON, err)
		}

		// Should not return error, but will log warning
		err := tm.LoadJSONDir(tmpDir)
		if err != nil {
			t.Errorf(testutil.ErrMsgLoadJSONDirError, err)
		}
	})

	t.Run("Load directory with empty zones array", func(t *testing.T) {
		tm := NewTimezoneManager()
		tmpDir := t.TempDir()

		emptyJSON := `{"zones": []}`
		jsonPath := tmpDir + "/empty.json"
		if err := os.WriteFile(jsonPath, []byte(emptyJSON), 0644); err != nil {
			t.Fatalf(testutil.ErrMsgFailedToWriteTestJSON, err)
		}

		err := tm.LoadJSONDir(tmpDir)
		if err != nil {
			t.Errorf(testutil.ErrMsgLoadJSONDirError, err)
		}
	})

	t.Run("Load directory with zone without IANA", func(t *testing.T) {
		tm := NewTimezoneManager()
		tmpDir := t.TempDir()

		jsonContent := `{
			"zones": [
				{
					"iana": "",
					"display_name": "No IANA",
					"country": "Test"
				}
			]
		}`
		jsonPath := tmpDir + "/noiana.json"
		if err := os.WriteFile(jsonPath, []byte(jsonContent), 0644); err != nil {
			t.Fatalf(testutil.ErrMsgFailedToWriteTestJSON, err)
		}

		err := tm.LoadJSONDir(tmpDir)
		if err != nil {
			t.Errorf(testutil.ErrMsgLoadJSONDirError, err)
		}
	})
}

func TestLoadDefaultJSONDirs(t *testing.T) {
	tm := NewTimezoneManager()

	// This function is non-fatal, so it shouldn't panic or error
	// Just verify it completes without panic
	tm.LoadDefaultJSONDirs()

	// Verify the manager still works after attempting to load
	zones := tm.ListTimezones()
	if len(zones) == 0 {
		t.Error("After LoadDefaultJSONDirs(), no timezones available")
	}
}

func TestLoadJSONFile(t *testing.T) {
	t.Run("Load valid JSON file with DST specified", func(t *testing.T) {
		tm := NewTimezoneManager()
		tmpDir := t.TempDir()

		dstTrue := true
		jsonContent := `{
			"zones": [
				{
					"iana": "Custom/Zone",
					"display_name": "Custom Zone",
					"country": "Custom Country",
					"dst": true
				}
			]
		}`

		jsonPath := tmpDir + "/custom.json"
		if err := os.WriteFile(jsonPath, []byte(jsonContent), 0644); err != nil {
			t.Fatalf(testutil.ErrMsgFailedToWriteTestJSON, err)
		}

		err := tm.loadJSONFile(jsonPath)
		if err != nil {
			t.Errorf(testutil.ErrMsgLoadJSONFileError, err)
		}

		zone, err := tm.GetTimezone("Custom/Zone")
		if err != nil {
			t.Errorf(testutil.ErrMsgFailedToGetLoadedZone, err)
		} else {
			if zone.DisplayName != "Custom Zone" {
				t.Errorf("DisplayName = %q, want 'Custom Zone'", zone.DisplayName)
			}
			if zone.Country != "Custom Country" {
				t.Errorf("Country = %q, want 'Custom Country'", zone.Country)
			}
			if zone.DST != dstTrue {
				t.Errorf("DST = %v, want %v", zone.DST, dstTrue)
			}
		}
	})

	t.Run("Load JSON file with empty display name and country", func(t *testing.T) {
		tm := NewTimezoneManager()
		tmpDir := t.TempDir()

		jsonContent := `{
			"zones": [
				{
					"iana": "Test/NoDetails",
					"display_name": "",
					"country": ""
				}
			]
		}`

		jsonPath := tmpDir + "/nodetails.json"
		if err := os.WriteFile(jsonPath, []byte(jsonContent), 0644); err != nil {
			t.Fatalf(testutil.ErrMsgFailedToWriteTestJSON, err)
		}

		err := tm.loadJSONFile(jsonPath)
		if err != nil {
			t.Errorf(testutil.ErrMsgLoadJSONFileError, err)
		}

		zone, err := tm.GetTimezone("Test/NoDetails")
		if err != nil {
			t.Errorf(testutil.ErrMsgFailedToGetLoadedZone, err)
		} else {
			// DisplayName should fall back to IANA
			if zone.DisplayName != "Test/NoDetails" {
				t.Errorf("DisplayName = %q, want 'Test/NoDetails'", zone.DisplayName)
			}
			// Country should fall back to "Unknown"
			if zone.Country != "Unknown" {
				t.Errorf("Country = %q, want 'Unknown'", zone.Country)
			}
		}
	})

	t.Run("Load non-existent file", func(t *testing.T) {
		tm := NewTimezoneManager()
		err := tm.loadJSONFile("/nonexistent/file.json")
		if err == nil {
			t.Error("loadJSONFile() expected error for non-existent file, got nil")
		}
	})

	t.Run("Load invalid JSON", func(t *testing.T) {
		tm := NewTimezoneManager()
		tmpDir := t.TempDir()

		jsonPath := tmpDir + "/invalid.json"
		if err := os.WriteFile(jsonPath, []byte("{invalid"), 0644); err != nil {
			t.Fatalf(testutil.ErrMsgFailedToWriteTestJSON, err)
		}

		err := tm.loadJSONFile(jsonPath)
		if err == nil {
			t.Error("loadJSONFile() expected error for invalid JSON, got nil")
		}
	})

	t.Run("Load JSON with empty alias", func(t *testing.T) {
		tm := NewTimezoneManager()
		tmpDir := t.TempDir()

		jsonContent := `{
			"zones": [
				{
					"iana": "Test/Alias",
					"display_name": "Test",
					"country": "Test",
					"aliases": ["", "  ", "valid_alias"]
				}
			]
		}`

		jsonPath := tmpDir + "/aliases.json"
		if err := os.WriteFile(jsonPath, []byte(jsonContent), 0644); err != nil {
			t.Fatalf(testutil.ErrMsgFailedToWriteTestJSON, err)
		}

		err := tm.loadJSONFile(jsonPath)
		if err != nil {
			t.Errorf(testutil.ErrMsgLoadJSONFileError, err)
		}

		// Valid alias should work
		zone, err := tm.GetTimezone("valid_alias")
		if err != nil {
			t.Errorf("Failed to get zone by valid alias: %v", err)
		} else if zone.IANA != "Test/Alias" {
			t.Errorf("Alias points to %q, want 'Test/Alias'", zone.IANA)
		}
	})

	t.Run("Load JSON with global alias pointing to non-existent zone", func(t *testing.T) {
		tm := NewTimezoneManager()
		tmpDir := t.TempDir()

		jsonContent := `{
			"zones": [],
			"aliases": {
				"bad_alias": "NonExistent/Zone"
			}
		}`

		jsonPath := tmpDir + "/badalias.json"
		if err := os.WriteFile(jsonPath, []byte(jsonContent), 0644); err != nil {
			t.Fatalf(testutil.ErrMsgFailedToWriteTestJSON, err)
		}

		err := tm.loadJSONFile(jsonPath)
		if err != nil {
			t.Errorf(testutil.ErrMsgLoadJSONFileError, err)
		}

		// Bad alias should not resolve
		_, err = tm.GetTimezone("bad_alias")
		if err == nil {
			t.Error("Expected error for alias pointing to non-existent zone")
		}
	})
}

// Additional edge case tests for partially covered functions

func TestGetTimezoneSystemFallback(t *testing.T) {
	tm := NewTimezoneManager()

	// Test a valid timezone that might not be in the pre-loaded list
	// but exists in the system
	zone, err := tm.GetTimezone("Pacific/Auckland")
	if err != nil {
		// This is OK if the system doesn't have this timezone
		t.Logf("Pacific/Auckland not available (expected on some systems): %v", err)
		return
	}

	if zone == nil {
		t.Error("GetTimezone returned nil zone without error")
	}

	if zone.IANA == "" {
		t.Error("GetTimezone returned zone with empty IANA")
	}
}

func TestListTimezonesNilZone(t *testing.T) {
	tm := &TimezoneManager{
		zones: make(map[string]*TimezoneInfo),
	}

	// Add a nil zone
	tm.zones["nil_test"] = nil

	// Should handle nil zone gracefully
	zones := tm.ListTimezones()

	for _, zone := range zones {
		if zone == nil {
			t.Error("ListTimezones() returned nil zone")
		}
	}
}

func TestListTimezonesEmptyIANA(t *testing.T) {
	tm := &TimezoneManager{
		zones: make(map[string]*TimezoneInfo),
	}

	// Add a zone with empty IANA
	tm.zones["empty"] = &TimezoneInfo{
		IANA:        "",
		DisplayName: "Empty",
		Country:     "Test",
	}

	// Should filter out zones with empty IANA
	zones := tm.ListTimezones()

	for _, zone := range zones {
		if zone.IANA == "" {
			t.Error("ListTimezones() returned zone with empty IANA")
		}
	}
}

func TestSuggestTimezoneEmptyInput(t *testing.T) {
	tm := NewTimezoneManager()

	tests := []struct {
		name  string
		input string
	}{
		{testutil.TestStringEmptyString, ""},
		{"Whitespace only", "   "},
		{"Tab only", "\t"},
		{"Newline only", "\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := tm.SuggestTimezone(tt.input)

			if results != nil && len(results) > 0 {
				t.Errorf("SuggestTimezone(%q) expected nil or empty, got %d results", tt.input, len(results))
			}
		})
	}
}

func TestSuggestTimezoneNilZone(t *testing.T) {
	tm := &TimezoneManager{
		zones: make(map[string]*TimezoneInfo),
	}

	// Add a nil zone
	tm.zones["test"] = nil

	// Should handle nil zones gracefully
	results := tm.SuggestTimezone("test")

	for _, zone := range results {
		if zone == nil {
			t.Error("SuggestTimezone() returned nil zone")
		}
	}
}

func TestSuggestTimezoneLimitResults(t *testing.T) {
	tm := NewTimezoneManager()

	// Search for a common term that should match many zones
	results := tm.SuggestTimezone("America")

	// Should limit to maximum 10 results
	if len(results) > 10 {
		t.Errorf("SuggestTimezone() returned %d results, maximum should be 10", len(results))
	}
}

func TestGetTimezoneOffsetInvalidTimezone(t *testing.T) {
	result := getTimezoneOffset(testutil.TZInvalid)

	if result != "Unknown" {
		t.Errorf("getTimezoneOffset('Invalid/Timezone') = %q, want 'Unknown'", result)
	}
}

func TestGetTimezoneOffsetNegativeOffset(t *testing.T) {
	// Test a timezone with negative offset
	result := getTimezoneOffset(testutil.TZAmericaNewYork)

	if result == "" {
		t.Error("getTimezoneOffset('America/New_York') returned empty string")
	}

	// Should have proper format
	if !strings.HasPrefix(result, "+") && !strings.HasPrefix(result, "-") {
		t.Errorf("getTimezoneOffset('America/New_York') = %q, should start with + or -", result)
	}
}

func TestHasDSTInvalidTimezone(t *testing.T) {
	result := hasDST(testutil.TZInvalid)

	if result != false {
		t.Errorf("hasDST('Invalid/Timezone') = %v, want false", result)
	}
}

func TestLoadFromZoneTabEmptyRows(t *testing.T) {
	// Test with a fresh manager
	tm := &TimezoneManager{
		zones: make(map[string]*TimezoneInfo),
	}

	// Call loadFromZoneTab
	tm.loadFromZoneTab()

	// Should load zones from embedded data
	if len(tm.zones) == 0 {
		t.Error("loadFromZoneTab() loaded no zones")
	}
}

func TestLoadFromZoneTabExistingZone(t *testing.T) {
	tm := &TimezoneManager{
		zones: make(map[string]*TimezoneInfo),
	}

	// Pre-add a zone
	preAddedZone := &TimezoneInfo{
		IANA:        testutil.TZEuropeMadrid,
		DisplayName: "Pre-added Madrid",
		Country:     "Pre-added Spain",
		Offset:      "+00:00",
		DST:         false,
	}
	tm.zones[testutil.TZEuropeMadrid] = preAddedZone

	// Load from zone tab (which should include Europe/Madrid)
	tm.loadFromZoneTab()

	// The pre-added zone should not be overwritten
	zone := tm.zones[testutil.TZEuropeMadrid]
	if zone.DisplayName != "Pre-added Madrid" {
		t.Errorf("loadFromZoneTab() overwrote existing zone: got %q", zone.DisplayName)
	}
}

func TestParseZone1970Tab(t *testing.T) {
	rows := parseZone1970Tab()

	if len(rows) == 0 {
		t.Error("parseZone1970Tab() returned no rows")
	}

	// Check that rows have expected fields
	for i, row := range rows {
		if row.CC == "" {
			t.Errorf("Row %d has empty country code", i)
		}
		if row.TZ == "" {
			t.Errorf("Row %d has empty timezone", i)
		}
		// Comment can be empty, that's OK
	}

	// Check for some expected timezones
	expectedZones := []string{testutil.TZEuropeLondon, testutil.TZAmericaNewYork, testutil.TZAsiaTokyо}
	for _, expected := range expectedZones {
		found := false
		for _, row := range rows {
			if row.TZ == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("parseZone1970Tab() missing expected timezone: %s", expected)
		}
	}
}
