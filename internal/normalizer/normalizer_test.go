package normalizer

import (
	"tempus/internal/testutil"
	"strings"
	"testing"
	"time"
)

func TestPrependToday(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		timezone string
		wantTime bool // true if result should contain time component
	}{
		{"clock only", "10:30", "UTC", true},
		{testutil.TestNameFullDatetime, "2025-06-15 10:30", "UTC", true},
		{testutil.TestNameDateOnly, "2025-06-15", "UTC", false},
		{"empty", "", "UTC", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PrependToday(tt.input, tt.timezone)

			if tt.input == "" {
				if result != "" {
					t.Errorf("PrependToday(%q) = %q, want empty", tt.input, result)
				}
				return
			}

			// If input was clock-only, result should have today's date
			if strings.Contains(tt.input, ":") && !strings.Contains(tt.input, " ") {
				if !strings.Contains(result, " ") {
					t.Errorf("PrependToday(%q) = %q, should contain date", tt.input, result)
				}
			}

			// If input already had date, should be unchanged
			if strings.Contains(tt.input, "-") {
				if result != tt.input {
					t.Errorf("PrependToday(%q) = %q, want %q", tt.input, result, tt.input)
				}
			}
		})
	}
}

func TestParseHumanDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{"45m", 45 * time.Minute, false},
		{"1h", 1 * time.Hour, false},
		{"1h30m", 90 * time.Minute, false},
		{"90", 90 * time.Minute, false},
		{"1:30", 90 * time.Minute, false},
		{"2:15", 135 * time.Minute, false},
		{"", 0, true},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseHumanDuration(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseHumanDuration(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("ParseHumanDuration(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseDateTime(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		timezone string
		wantErr  bool
	}{
		{testutil.TestNameFullDatetime, "2025-06-15 14:30", "UTC", false},
		{testutil.TestNameDateOnly, "2025-06-15", "UTC", false},
		{"time only", "14:30", "UTC", false},
		{"empty", "", "UTC", true},
		{"invalid", "invalid", "UTC", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDateTime(tt.input, tt.timezone)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDateTime(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && result.IsZero() {
				t.Errorf("ParseDateTime(%q) returned zero time", tt.input)
			}
		})
	}
}

func TestNormalizeEndTimeFromDuration(t *testing.T) {
	tests := []struct {
		name     string
		start    string
		end      string
		duration string
		timezone string
		wantErr  bool
	}{
		{"end provided", testutil.DateTime20250615_1000, "2025-06-15 11:00", "1h", "UTC", false},
		{"duration used", testutil.DateTime20250615_1000, "", "45m", "UTC", false},
		{"empty duration", testutil.DateTime20250615_1000, "", "", "UTC", false},
		{"invalid start", "invalid", "", "1h", "UTC", true},
		{"invalid duration", testutil.DateTime20250615_1000, "", "invalid", "UTC", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NormalizeEndTimeFromDuration(tt.start, tt.end, tt.duration, tt.timezone)
			if (err != nil) != tt.wantErr {
				t.Errorf("NormalizeEndTimeFromDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If end was provided, it should be returned unchanged
			if tt.end != "" && result != tt.end {
				t.Errorf("NormalizeEndTimeFromDuration() = %q, want %q", result, tt.end)
			}

			// If duration was used, result should not be empty
			if tt.end == "" && tt.duration != "" && !tt.wantErr && result == "" {
				t.Errorf("NormalizeEndTimeFromDuration() returned empty, expected calculated end time")
			}
		})
	}
}

func TestNormalizeValuesForTemplate(t *testing.T) {
	tests := []struct {
		name     string
		values   map[string]string
		wantErr  bool
		checkKey string
		wantVal  string
	}{
		{
			name: "normalize clock time",
			values: map[string]string{
				"start_time": "10:30",
				"timezone":   "UTC",
			},
			wantErr:  false,
			checkKey: "start_time",
		},
		{
			name: "calculate end from duration",
			values: map[string]string{
				"start_time": testutil.DateTime20250615_1000,
				"end_time":   "",
				"duration":   "1h",
				"timezone":   "UTC",
			},
			wantErr:  false,
			checkKey: "end_time",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NormalizeValuesForTemplate(tt.values, "start_time", "end_time", "duration", "timezone")
			if (err != nil) != tt.wantErr {
				t.Errorf("NormalizeValuesForTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.checkKey != "" {
				val := tt.values[tt.checkKey]
				if val == "" {
					t.Errorf("NormalizeValuesForTemplate() did not set %q", tt.checkKey)
				}
			}
		})
	}
}
