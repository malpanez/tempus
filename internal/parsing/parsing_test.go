package parsing

import (
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	t.Run("all-day event", func(t *testing.T) {
		result, err := Parse(ParseOptions{
			StartDate: "2025-01-15",
			AllDay:    true,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
		if !result.Start.Equal(expected) {
			t.Errorf("start = %v, want %v", result.Start, expected)
		}
		expectedEnd := expected.AddDate(0, 0, 1)
		if !result.End.Equal(expectedEnd) {
			t.Errorf("end = %v, want %v", result.End, expectedEnd)
		}
	})

	t.Run("timed event with duration", func(t *testing.T) {
		result, err := Parse(ParseOptions{
			StartDate: "2025-01-15",
			StartTime: "10:00",
			Duration:  "1h",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expectedStart := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
		if !result.Start.Equal(expectedStart) {
			t.Errorf("start = %v, want %v", result.Start, expectedStart)
		}
		expectedEnd := expectedStart.Add(1 * time.Hour)
		if !result.End.Equal(expectedEnd) {
			t.Errorf("end = %v, want %v", result.End, expectedEnd)
		}
	})

	t.Run("timed event with end time", func(t *testing.T) {
		result, err := Parse(ParseOptions{
			StartDate: "2025-01-15",
			StartTime: "10:00",
			EndTime:   "11:30",
			EndDate:   "2025-01-15",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expectedEnd := time.Date(2025, 1, 15, 11, 30, 0, 0, time.UTC)
		if !result.End.Equal(expectedEnd) {
			t.Errorf("end = %v, want %v", result.End, expectedEnd)
		}
	})

	t.Run("slash dates normalized", func(t *testing.T) {
		result, err := Parse(ParseOptions{
			StartDate: "2025/01/15",
			StartTime: "10:00",
			Duration:  "1h",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expectedStart := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
		if !result.Start.Equal(expectedStart) {
			t.Errorf("start = %v, want %v", result.Start, expectedStart)
		}
	})

	t.Run("missing required fields", func(t *testing.T) {
		_, err := Parse(ParseOptions{})
		if err == nil {
			t.Fatal("expected error for empty options")
		}
	})

	t.Run("all-day with end date", func(t *testing.T) {
		result, err := Parse(ParseOptions{
			StartDate: "2025-01-15",
			EndDate:   "2025-01-17",
			AllDay:    true,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expectedEnd := time.Date(2025, 1, 18, 0, 0, 0, 0, time.UTC)
		if !result.End.Equal(expectedEnd) {
			t.Errorf("end = %v, want %v", result.End, expectedEnd)
		}
	})

	t.Run("default duration 1h when no end or duration", func(t *testing.T) {
		result, err := Parse(ParseOptions{
			StartDate: "2025-01-15",
			StartTime: "10:00",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expectedEnd := time.Date(2025, 1, 15, 11, 0, 0, 0, time.UTC)
		if !result.End.Equal(expectedEnd) {
			t.Errorf("end = %v, want %v", result.End, expectedEnd)
		}
	})
}

func TestNormalizeDateTimeInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"already normalized", "2025-01-15 10:00", "2025-01-15 10:00"},
		{"slash date", "2025/01/15 10:00", "2025-01-15 10:00"},
		{"missing zero month", "2025-1-15 10:00", "2025-01-15 10:00"},
		{"missing zero day", "2025-01-5 10:00", "2025-01-05 10:00"},
		{"time without colon", "2025-01-15 0900", "2025-01-15 09:00"},
		{"single digit hour", "2025-01-15 9:00", "2025-01-15 09:00"},
		{"date only", "2025/1/5", "2025-01-05"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := NormalizeDateTimeInput(tc.input)
			if got != tc.want {
				t.Errorf("NormalizeDateTimeInput(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestParseDateTimeWithTZ(t *testing.T) {
	t.Run("date only no tz", func(t *testing.T) {
		result, err := ParseDateTimeWithTZ("2025-01-15", "", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Year() != 2025 || result.Month() != 1 || result.Day() != 15 {
			t.Errorf("unexpected result: %v", result)
		}
	})

	t.Run("date and time with tz", func(t *testing.T) {
		result, err := ParseDateTimeWithTZ("2025-01-15", "10:00", "UTC")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Hour() != 10 {
			t.Errorf("hour = %d, want 10", result.Hour())
		}
	})

	t.Run("invalid timezone", func(t *testing.T) {
		_, err := ParseDateTimeWithTZ("2025-01-15", "10:00", "Invalid/TZ")
		if err == nil {
			t.Fatal("expected error for invalid timezone")
		}
	})
}

func TestAddDurationToStart(t *testing.T) {
	t.Run("basic addition", func(t *testing.T) {
		result := AddDurationToStart("2025-01-15 10:00", "", 1*time.Hour)
		if result != "2025-01-15 11:00" {
			t.Errorf("got %q, want %q", result, "2025-01-15 11:00")
		}
	})

	t.Run("with timezone", func(t *testing.T) {
		result := AddDurationToStart("2025-01-15 10:00", "UTC", 30*time.Minute)
		if result != "2025-01-15 10:30" {
			t.Errorf("got %q, want %q", result, "2025-01-15 10:30")
		}
	})

	t.Run("invalid input", func(t *testing.T) {
		result := AddDurationToStart("invalid", "", 1*time.Hour)
		if result != "" {
			t.Errorf("got %q, want empty string", result)
		}
	})
}

func TestParseExDateValues(t *testing.T) {
	t.Run("date only", func(t *testing.T) {
		dates, err := ParseExDateValues([]string{"2025-01-15"}, "", true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(dates) != 1 {
			t.Fatalf("expected 1 date, got %d", len(dates))
		}
		if dates[0].Day() != 15 {
			t.Errorf("day = %d, want 15", dates[0].Day())
		}
	})

	t.Run("date and time", func(t *testing.T) {
		dates, err := ParseExDateValues([]string{"2025-01-15 10:00"}, "", false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(dates) != 1 {
			t.Fatalf("expected 1 date, got %d", len(dates))
		}
	})

	t.Run("empty values skipped", func(t *testing.T) {
		dates, err := ParseExDateValues([]string{"", "  ", "2025-01-15"}, "", true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(dates) != 1 {
			t.Fatalf("expected 1 date, got %d", len(dates))
		}
	})

	t.Run("invalid date", func(t *testing.T) {
		_, err := ParseExDateValues([]string{"not-a-date"}, "", true)
		if err == nil {
			t.Fatal("expected error for invalid date")
		}
	})

	t.Run("T separator", func(t *testing.T) {
		dates, err := ParseExDateValues([]string{"2025-01-15T10:00"}, "", false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(dates) != 1 {
			t.Fatalf("expected 1 date, got %d", len(dates))
		}
	})
}

func TestNormalizeClockOnlyDateTimes(t *testing.T) {
	t.Run("clock-only start", func(t *testing.T) {
		values := map[string]string{
			"start": "10:00",
			"tz":    "UTC",
		}
		NormalizeClockOnlyDateTimes(values, "start", "", "tz")
		if values["start"] == "10:00" {
			t.Error("start should have been prepended with today's date")
		}
	})

	t.Run("already full datetime", func(t *testing.T) {
		values := map[string]string{
			"start": "2025-01-15 10:00",
			"tz":    "UTC",
		}
		NormalizeClockOnlyDateTimes(values, "start", "", "tz")
		if values["start"] != "2025-01-15 10:00" {
			t.Error("full datetime should not be modified")
		}
	})

	t.Run("empty start key", func(t *testing.T) {
		values := map[string]string{
			"start": "10:00",
		}
		NormalizeClockOnlyDateTimes(values, "", "", "tz")
		if values["start"] != "10:00" {
			t.Error("should not modify when startKey is empty")
		}
	})
}

func TestNormalizeEndTimeFromDuration(t *testing.T) {
	t.Run("end from duration", func(t *testing.T) {
		values := map[string]string{
			"start":    "2025-01-15 10:00",
			"end":      "",
			"duration": "1h",
			"tz":       "",
		}
		NormalizeEndTimeFromDuration(values, "start", "end", "duration", "tz", "30m")
		if values["end"] != "2025-01-15 11:00" {
			t.Errorf("end = %q, want %q", values["end"], "2025-01-15 11:00")
		}
	})

	t.Run("default duration", func(t *testing.T) {
		values := map[string]string{
			"start":    "2025-01-15 10:00",
			"end":      "",
			"duration": "",
			"tz":       "",
		}
		NormalizeEndTimeFromDuration(values, "start", "end", "duration", "tz", "30m")
		if values["end"] != "2025-01-15 10:30" {
			t.Errorf("end = %q, want %q", values["end"], "2025-01-15 10:30")
		}
	})

	t.Run("duration in end field", func(t *testing.T) {
		values := map[string]string{
			"start":    "2025-01-15 10:00",
			"end":      "2h",
			"duration": "",
			"tz":       "",
		}
		NormalizeEndTimeFromDuration(values, "start", "end", "duration", "tz", "30m")
		if values["end"] != "2025-01-15 12:00" {
			t.Errorf("end = %q, want %q", values["end"], "2025-01-15 12:00")
		}
	})

	t.Run("empty start", func(t *testing.T) {
		values := map[string]string{
			"start": "",
			"end":   "",
		}
		NormalizeEndTimeFromDuration(values, "start", "end", "duration", "tz", "30m")
		if values["end"] != "" {
			t.Errorf("end should remain empty when start is empty")
		}
	})
}

func TestParseEndTime(t *testing.T) {
	start := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)

	t.Run("duration string", func(t *testing.T) {
		end, err := parseEndTime(start, "1h", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := start.Add(1 * time.Hour)
		if !end.Equal(expected) {
			t.Errorf("end = %v, want %v", end, expected)
		}
	})

	t.Run("explicit datetime", func(t *testing.T) {
		end, err := parseEndTime(start, "2025-01-15 11:30", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := time.Date(2025, 1, 15, 11, 30, 0, 0, time.UTC)
		if !end.Equal(expected) {
			t.Errorf("end = %v, want %v", end, expected)
		}
	})

	t.Run("zero duration", func(t *testing.T) {
		_, err := parseEndTime(start, "0m", "")
		if err == nil {
			t.Fatal("expected error for zero duration")
		}
	})
}

func TestParseDurationEnd(t *testing.T) {
	start := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)

	t.Run("valid duration", func(t *testing.T) {
		end, err := parseDurationEnd(start, "2h")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := start.Add(2 * time.Hour)
		if !end.Equal(expected) {
			t.Errorf("end = %v, want %v", end, expected)
		}
	})

	t.Run("invalid duration", func(t *testing.T) {
		_, err := parseDurationEnd(start, "not-a-duration")
		if err == nil {
			t.Fatal("expected error for invalid duration")
		}
	})

	t.Run("zero duration", func(t *testing.T) {
		_, err := parseDurationEnd(start, "0m")
		if err == nil {
			t.Fatal("expected error for zero duration")
		}
	})
}

func TestBuildDateTimeStr(t *testing.T) {
	tests := []struct {
		date, time, want string
	}{
		{"2025-01-15", "10:00", "2025-01-15 10:00"},
		{"2025-01-15", "", "2025-01-15"},
		{"", "10:00", "10:00"},
		{"", "", ""},
		{"2025-01-15 10:00", "ignored", "2025-01-15 10:00"},
	}
	for _, tc := range tests {
		got := buildDateTimeStr(tc.date, tc.time)
		if got != tc.want {
			t.Errorf("buildDateTimeStr(%q, %q) = %q, want %q", tc.date, tc.time, got, tc.want)
		}
	}
}

func TestNormalizeTimeInput(t *testing.T) {
	got := normalizeTimeInput("10:00", "UTC", "")
	if got == "10:00" {
		t.Error("clock-only should be prepended with today's date")
	}
	got = normalizeTimeInput("2025-01-15 10:00", "UTC", "")
	if got != "2025-01-15 10:00" {
		t.Error("full datetime should not be modified")
	}
	got = normalizeTimeInput("", "UTC", "")
	if got != "" {
		t.Error("empty should stay empty")
	}
}

func TestParseAllDayErrors(t *testing.T) {
	t.Run("empty start", func(t *testing.T) {
		_, err := Parse(ParseOptions{AllDay: true})
		if err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("invalid start", func(t *testing.T) {
		_, err := Parse(ParseOptions{StartDate: "not-a-date", AllDay: true})
		if err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("invalid end date", func(t *testing.T) {
		_, err := Parse(ParseOptions{StartDate: "2025-01-15", EndDate: "not-a-date", AllDay: true})
		if err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("end before start", func(t *testing.T) {
		_, err := Parse(ParseOptions{StartDate: "2025-01-15", EndDate: "2025-01-10", AllDay: true})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestParseTimedErrors(t *testing.T) {
	t.Run("empty start", func(t *testing.T) {
		_, err := Parse(ParseOptions{})
		if err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("invalid start format", func(t *testing.T) {
		_, err := Parse(ParseOptions{StartDate: "not-a-date", StartTime: "10:00"})
		if err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("invalid end time", func(t *testing.T) {
		_, err := Parse(ParseOptions{StartDate: "2025-01-15", StartTime: "10:00", EndDate: "2025-01-15", EndTime: "not-time"})
		if err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("end before start", func(t *testing.T) {
		_, err := Parse(ParseOptions{StartDate: "2025-01-15", StartTime: "10:00", EndDate: "2025-01-15", EndTime: "09:00"})
		if err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("invalid duration", func(t *testing.T) {
		_, err := Parse(ParseOptions{StartDate: "2025-01-15", StartTime: "10:00", Duration: "not-a-duration"})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestGetValueIfKeyNotEmpty(t *testing.T) {
	values := map[string]string{"key": "value"}
	if got := GetValueIfKeyNotEmpty(values, "key"); got != "value" {
		t.Errorf("got %q, want %q", got, "value")
	}
	if got := GetValueIfKeyNotEmpty(values, ""); got != "" {
		t.Errorf("empty key: got %q, want empty", got)
	}
}

func TestNormalizeClockOnlyDateTimesEndKey(t *testing.T) {
	t.Run("clock-only end treated as duration", func(t *testing.T) {
		values := map[string]string{
			"start": "2025-01-15 10:00",
			"end":   "11:00",
			"tz":    "UTC",
		}
		NormalizeClockOnlyDateTimes(values, "start", "end", "tz")
		if values["end"] != "11:00" {
			t.Errorf("11:00 is a valid duration, should not be prepended, got %q", values["end"])
		}
	})

	t.Run("duration in end not modified", func(t *testing.T) {
		values := map[string]string{
			"start": "2025-01-15 10:00",
			"end":   "1h",
			"tz":    "UTC",
		}
		NormalizeClockOnlyDateTimes(values, "start", "end", "tz")
	})
}

func TestSetEndFromDurationInvalid(t *testing.T) {
	values := map[string]string{
		"start": "2025-01-15 10:00",
		"end":   "",
	}
	SetEndFromDuration(values, "2025-01-15 10:00", "", "invalid", "end", "duration")
	if values["end"] != "" {
		t.Error("end should remain empty for invalid duration")
	}
}

func TestTrySetEndFromDurationInEndEmpty(t *testing.T) {
	values := map[string]string{}
	result := TrySetEndFromDurationInEnd(values, "2025-01-15 10:00", "", "", "end", "duration")
	if result {
		t.Error("should return false for empty end")
	}
}

func TestParseExDateValuesWithTZ(t *testing.T) {
	dates, err := ParseExDateValues([]string{"2025-01-15 10:00"}, "UTC", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dates) != 1 {
		t.Fatalf("expected 1 date, got %d", len(dates))
	}
}

func TestParseAllDayEndViaEndTime(t *testing.T) {
	result, err := Parse(ParseOptions{
		StartDate: "2025-01-15",
		EndTime:   "2025-01-17",
		AllDay:    true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedEnd := time.Date(2025, 1, 18, 0, 0, 0, 0, time.UTC)
	if !result.End.Equal(expectedEnd) {
		t.Errorf("end = %v, want %v", result.End, expectedEnd)
	}
}
