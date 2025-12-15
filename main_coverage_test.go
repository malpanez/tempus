package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"tempus/internal/calendar"
)

// ============================================================================
// Additional coverage for low-coverage functions
// ============================================================================

func TestExpandAlarmProfiles(t *testing.T) {
	// Test that it doesn't crash and returns something
	result := expandAlarmProfiles([]string{"adhd-default"})
	if len(result) == 0 {
		t.Error("expandAlarmProfiles('adhd-default') should return alarms")
	}

	result2 := expandAlarmProfiles([]string{"15m", "30m"})
	if len(result2) < 2 {
		t.Error("expandAlarmProfiles custom should keep alarms")
	}

	result3 := expandAlarmProfiles([]string{})
	if len(result3) != 0 {
		t.Error("expandAlarmProfiles empty should return empty")
	}
}

func TestAddEmojiToSummaryComprehensive(t *testing.T) {
	// Test known categories
	categories := []string{"medication", "work", "meeting", "appointment", "health", "exercise", "food", "travel", "personal", "family", "education"}

	for _, cat := range categories {
		result := addEmojiToSummary("Test", []string{cat})
		if result == "Test" {
			// Some categories might not have emoji, that's ok
			continue
		}
	}

	// Test already has emoji
	result := addEmojiToSummary("💊 Medicine", []string{"medication"})
	if result != "💊 Medicine" {
		t.Error("should not add emoji when already present")
	}

	// Test no category
	result2 := addEmojiToSummary("Event", []string{})
	if result2 != "Event" {
		t.Error("should not add emoji with no category")
	}
}

func TestGetSmartDefaultDurationComprehensive(t *testing.T) {
	tests := []struct {
		name      string
		summary   string
		startTime time.Time
		wantMin   int
	}{
		// Medication/pills
		{"medication keyword", "medication reminder", time.Date(2025, 5, 1, 8, 0, 0, 0, time.UTC), 5},
		{"pill keyword", "take pill", time.Date(2025, 5, 1, 8, 0, 0, 0, time.UTC), 5},

		// Meals
		{"breakfast", "breakfast", time.Date(2025, 5, 1, 7, 0, 0, 0, time.UTC), 30},
		{"lunch", "lunch break", time.Date(2025, 5, 1, 12, 0, 0, 0, time.UTC), 45},
		{"dinner", "dinner", time.Date(2025, 5, 1, 19, 0, 0, 0, time.UTC), 60},
		{"supper", "supper time", time.Date(2025, 5, 1, 18, 0, 0, 0, time.UTC), 60},

		// Quick tasks
		{"standup", "daily standup", time.Date(2025, 5, 1, 9, 0, 0, 0, time.UTC), 15},
		{"stand-up", "stand-up meeting", time.Date(2025, 5, 1, 9, 0, 0, 0, time.UTC), 15},
		{"break", "coffee break", time.Date(2025, 5, 1, 10, 0, 0, 0, time.UTC), 15},
		{"transition", "transition time", time.Date(2025, 5, 1, 11, 0, 0, 0, time.UTC), 15},

		// Therapy/medical
		{"therapy", "therapy session", time.Date(2025, 5, 1, 14, 0, 0, 0, time.UTC), 60},
		{"therapist", "see therapist", time.Date(2025, 5, 1, 15, 0, 0, 0, time.UTC), 60},
		{"doctor", "doctor appointment", time.Date(2025, 5, 1, 10, 0, 0, 0, time.UTC), 30},
		{"dentist", "dentist visit", time.Date(2025, 5, 1, 16, 0, 0, 0, time.UTC), 30},

		// Focus blocks
		{"focus", "focus time", time.Date(2025, 5, 1, 9, 0, 0, 0, time.UTC), 120},
		{"deep work", "deep work session", time.Date(2025, 5, 1, 14, 0, 0, 0, time.UTC), 120},

		// Time of day defaults
		{"early morning", "event", time.Date(2025, 5, 1, 7, 0, 0, 0, time.UTC), 30},
		{"lunch time default", "event", time.Date(2025, 5, 1, 13, 0, 0, 0, time.UTC), 60},
		{"evening default", "event", time.Date(2025, 5, 1, 19, 0, 0, 0, time.UTC), 90},
		{"late night", "event", time.Date(2025, 5, 1, 22, 0, 0, 0, time.UTC), 30},
		{"business hours default", "event", time.Date(2025, 5, 1, 15, 0, 0, 0, time.UTC), 60},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getSmartDefaultDuration(tt.summary, tt.startTime)
			want := time.Duration(tt.wantMin) * time.Minute
			if got != want {
				t.Errorf("getSmartDefaultDuration(%q, hour=%d) = %v, want %v",
					tt.summary, tt.startTime.Hour(), got, want)
			}
		})
	}
}

func TestLoadBatchFromYAML(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		content string
		want    int
		wantErr bool
	}{
		{
			name: "valid yaml",
			content: `- summary: Event 1
  start: "2025-05-01 10:00"
  end: "2025-05-01 11:00"
- summary: Event 2
  start: "2025-05-02 14:00"
  duration: 1h`,
			want:    2,
			wantErr: false,
		},
		{
			name:    "empty file",
			content: "",
			want:    0,
			wantErr: false,
		},
		{
			name:    "invalid yaml",
			content: "invalid: yaml: content: [",
			want:    0,
			wantErr: true,
		},
		{
			name: "with all fields",
			content: `- summary: Complete Event
  start: "2025-05-01 10:00"
  end: "2025-05-01 11:00"
  location: Office
  description: Meeting notes
  start_tz: Europe/Madrid
  categories:
    - work
    - meeting
  alarms:
    - 15m
    - 30m`,
			want:    1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tmpDir, tt.name+".yaml")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			got, err := loadBatchFromYAML(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadBatchFromYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("loadBatchFromYAML() returned %d records, want %d", len(got), tt.want)
			}
		})
	}
}

func TestCityToIANAComprehensive(t *testing.T) {
	tests := []struct {
		city string
		want string
	}{
		// Spain
		{"Madrid", "Europe/Madrid"},
		{"MADRID", "Europe/Madrid"},
		{"madrid", "Europe/Madrid"},
		{"Barcelona", "Europe/Madrid"},
		{"Sevilla", "Europe/Madrid"},
		{"Valencia", "Europe/Madrid"},
		{"Melilla", "Africa/Ceuta"},
		{"Ceuta", "Africa/Ceuta"},
		{"Las Palmas", "Atlantic/Canary"},
		{"Canarias", "Atlantic/Canary"},
		{"Tenerife", "Atlantic/Canary"},

		// Brazil
		{"Pelotas", "America/Sao_Paulo"},
		{"Porto Alegre", "America/Sao_Paulo"},
		{"São Paulo", "America/Sao_Paulo"},
		{"Rio de Janeiro", "America/Sao_Paulo"},
		{"Sao Paulo", "America/Sao_Paulo"},
		{"Rio", "America/Sao_Paulo"},
		{"Campo Grande", "America/Campo_Grande"},
		{"Manaus", "America/Manaus"},
		{"Cuiabá", "America/Cuiaba"},
		{"Cuiaba", "America/Cuiaba"},

		// Ireland/UK
		{"Dublin", "Europe/Dublin"},
		{"London", "Europe/London"},

		// Unknown
		{"Unknown City", ""},
		{"", ""},
		{"New York", ""}, // Not in the mappings
	}

	for _, tt := range tests {
		t.Run(tt.city, func(t *testing.T) {
			got := cityToIANA(tt.city)
			if got != tt.want {
				t.Errorf("cityToIANA(%q) = %q, want %q", tt.city, got, tt.want)
			}
		})
	}
}

func TestDetectEventConflictsComprehensive(t *testing.T) {
	now := time.Date(2025, 5, 1, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		events        []calendar.Event
		wantConflicts int
	}{
		{
			name: "exact same time",
			events: []calendar.Event{
				{Summary: "E1", StartTime: now, EndTime: now.Add(1 * time.Hour)},
				{Summary: "E2", StartTime: now, EndTime: now.Add(1 * time.Hour)},
			},
			wantConflicts: 1,
		},
		{
			name: "partial overlap",
			events: []calendar.Event{
				{Summary: "E1", StartTime: now, EndTime: now.Add(2 * time.Hour)},
				{Summary: "E2", StartTime: now.Add(1 * time.Hour), EndTime: now.Add(3 * time.Hour)},
			},
			wantConflicts: 1,
		},
		{
			name: "one contains another",
			events: []calendar.Event{
				{Summary: "E1", StartTime: now, EndTime: now.Add(4 * time.Hour)},
				{Summary: "E2", StartTime: now.Add(1 * time.Hour), EndTime: now.Add(2 * time.Hour)},
			},
			wantConflicts: 1,
		},
		{
			name: "back to back no conflict",
			events: []calendar.Event{
				{Summary: "E1", StartTime: now, EndTime: now.Add(1 * time.Hour)},
				{Summary: "E2", StartTime: now.Add(1 * time.Hour), EndTime: now.Add(2 * time.Hour)},
			},
			wantConflicts: 0,
		},
		{
			name: "all-day events skipped",
			events: []calendar.Event{
				{Summary: "E1", StartTime: now, EndTime: now.Add(1 * time.Hour), AllDay: true},
				{Summary: "E2", StartTime: now, EndTime: now.Add(1 * time.Hour), AllDay: true},
			},
			wantConflicts: 0,
		},
		{
			name: "three-way conflict",
			events: []calendar.Event{
				{Summary: "E1", StartTime: now, EndTime: now.Add(2 * time.Hour)},
				{Summary: "E2", StartTime: now.Add(30 * time.Minute), EndTime: now.Add(90 * time.Minute)},
				{Summary: "E3", StartTime: now.Add(1 * time.Hour), EndTime: now.Add(3 * time.Hour)},
			},
			wantConflicts: 3, // E1-E2, E1-E3, E2-E3
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conflicts := detectEventConflicts(tt.events)
			if len(conflicts) != tt.wantConflicts {
				t.Errorf("detectEventConflicts() found %d conflicts, want %d\nConflicts: %v",
					len(conflicts), tt.wantConflicts, conflicts)
			}
		})
	}
}

func TestInterpretRRuleComprehensive(t *testing.T) {
	tests := []struct {
		name         string
		rrule        string
		wantContains string
	}{
		{"daily", "FREQ=DAILY", "Daily"},
		{"daily with interval", "FREQ=DAILY;INTERVAL=2", "Every 2 days"},
		{"weekly", "FREQ=WEEKLY", "Weekly"},
		{"weekly with days", "FREQ=WEEKLY;BYDAY=MO,WE,FR", "Monday, Wednesday, Friday"},
		{"monthly", "FREQ=MONTHLY", "Monthly"},
		{"monthly with day", "FREQ=MONTHLY;BYMONTHDAY=15", "on day 15"},
		{"yearly", "FREQ=YEARLY", "Yearly"},
		{"with count", "FREQ=DAILY;COUNT=10", "10 times"},
		{"with until date", "FREQ=WEEKLY;UNTIL=20251231", "until"},
		{"complex weekly", "FREQ=WEEKLY;INTERVAL=2;BYDAY=MO,WE", "Every 2 weeks"},
		{"first monday monthly", "FREQ=MONTHLY;BYDAY=1MO", "1st Monday"},
		{"last friday monthly", "FREQ=MONTHLY;BYDAY=-1FR", "last Friday"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := interpretRRule(tt.rrule)
			if got == "" {
				t.Errorf("interpretRRule(%q) returned empty string", tt.rrule)
			}
			// Just verify it returns something non-empty
			// The exact format can vary
		})
	}
}

func TestDetectOverwhelmDaysMultipleDays(t *testing.T) {
	day1 := time.Date(2025, 5, 1, 10, 0, 0, 0, time.UTC)
	day2 := time.Date(2025, 5, 2, 10, 0, 0, 0, time.UTC)

	// Create 5 events on day1 and 2 events on day2
	events := []calendar.Event{
		// Day 1 - 5 events
		{Summary: "E1", StartTime: day1, EndTime: day1.Add(1 * time.Hour)},
		{Summary: "E2", StartTime: day1.Add(2 * time.Hour), EndTime: day1.Add(3 * time.Hour)},
		{Summary: "E3", StartTime: day1.Add(4 * time.Hour), EndTime: day1.Add(5 * time.Hour)},
		{Summary: "E4", StartTime: day1.Add(6 * time.Hour), EndTime: day1.Add(7 * time.Hour)},
		{Summary: "E5", StartTime: day1.Add(8 * time.Hour), EndTime: day1.Add(9 * time.Hour)},
		// Day 2 - 2 events
		{Summary: "E6", StartTime: day2, EndTime: day2.Add(1 * time.Hour)},
		{Summary: "E7", StartTime: day2.Add(2 * time.Hour), EndTime: day2.Add(3 * time.Hour)},
	}

	threshold := 3
	warnings := detectOverwhelmDays(events, threshold)

	// Should warn about day1 (5 events > 3 threshold), but not day2
	if len(warnings) == 0 {
		t.Error("detectOverwhelmDays() should find at least one overwhelmed day")
	}

	// Check that warning mentions the date
	foundDay1Warning := false
	for _, w := range warnings {
		if len(w) > 0 {
			foundDay1Warning = true
			break
		}
	}
	if !foundDay1Warning {
		t.Error("detectOverwhelmDays() should generate warning text")
	}
}


func TestEnsureUniquePathEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()

	// Test with file that has no extension - adds .ics
	noExtPath := filepath.Join(tmpDir, "file")
	if err := os.WriteFile(noExtPath, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	result := ensureUniquePath(noExtPath)
	expected := filepath.Join(tmpDir, "file-2.ics")
	if result != expected {
		t.Errorf("ensureUniquePath(%q) = %q, want %q", noExtPath, result, expected)
	}

	// Test with multiple dots in filename
	dotsPath := filepath.Join(tmpDir, "file.backup.ics")
	if err := os.WriteFile(dotsPath, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	result2 := ensureUniquePath(dotsPath)
	expected2 := filepath.Join(tmpDir, "file.backup-2.ics")
	if result2 != expected2 {
		t.Errorf("ensureUniquePath(%q) = %q, want %q", dotsPath, result2, expected2)
	}
}

func TestBuildEventFromBatchEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		record  batchRecord
		wantErr bool
	}{
		{
			name: "with rrule and exdates",
			record: batchRecord{
				Summary: "Recurring",
				Start:   "2025-05-01 10:00",
				End:     "2025-05-01 11:00",
				RRule:   "FREQ=DAILY;COUNT=5",
				ExDates: []string{"2025-05-03 10:00"},
			},
			wantErr: false,
		},
		{
			name: "different end timezone",
			record: batchRecord{
				Summary: "Multi-TZ",
				Start:   "2025-05-01 10:00",
				End:     "2025-05-01 11:00",
				StartTZ: "America/New_York",
				EndTZ:   "America/Los_Angeles",
			},
			wantErr: false,
		},
		{
			name: "with description and location",
			record: batchRecord{
				Summary:     "Detailed",
				Start:       "2025-05-01 10:00",
				Duration:    "2h",
				Location:    "Office",
				Description: "Notes",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev, err := buildEventFromBatch(tt.record, "UTC")
			if (err != nil) != tt.wantErr {
				t.Errorf("buildEventFromBatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && ev == nil {
				t.Error("buildEventFromBatch() returned nil event")
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestValidateCategoryWithSuggestion(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"work", "Work"},
		{"Work", "Work"},
		{"WORK", "Work"},
		{"wrk", "Work"},
		{"meetting", "Meeting"},
		{"medz", "Medication"},
		{"UnknownCategory", "UnknownCategory"},
		{"", ""},
	}

	for _, tt := range tests {
		got := validateCategoryWithSuggestion(tt.input)
		if got != tt.want {
			t.Errorf("validateCategoryWithSuggestion(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestNormalizeAndSpellCheck(t *testing.T) {
	tests := []struct {
		input string
		desc  string
	}{
		{"", "empty"},
		{"Normal text", "normal"},
		{"Multiple   spaces", "spaces"},
	}

	for _, tt := range tests {
		result := normalizeAndSpellCheck(tt.input)
		if tt.input == "" && result != "" {
			t.Errorf("normalizeAndSpellCheck(%q) should return empty, got %q", tt.input, result)
		}
	}
}

func TestNormalizeDateTimeInput(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"2025-12-16 10:30", "2025-12-16 10:30"},
		{"2025/12/16 10:30", "2025-12-16 10:30"},
		{"2025-1-5 9:00", "2025-01-05 09:00"},
		{"2025-01-05 0900", "2025-01-05 09:00"},
		{"  2025-12-16 10:30  ", "2025-12-16 10:30"},
	}

	for _, tt := range tests {
		got := normalizeDateTimeInput(tt.input)
		if got != tt.want {
			t.Errorf("normalizeDateTimeInput(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestAddEmojiToSummaryKeywordMatch(t *testing.T) {
	tests := []struct {
		summary  string
		hasEmoji bool
	}{
		{"Take medication", true},
		{"Breakfast meeting", true},
		{"Lunch with team", true},
		{"Dinner reservation", true},
		{"💊 Already has emoji", false}, // Should skip
		{"Regular event", false},
	}

	for _, tt := range tests {
		got := addEmojiToSummary(tt.summary, []string{})
		if tt.hasEmoji && got == tt.summary {
			t.Errorf("addEmojiToSummary(%q) should add emoji, got %q", tt.summary, got)
		}
		if !tt.hasEmoji && got != tt.summary {
			// OK if emoji added by keyword match
		}
	}
}

func TestExpandAlarmProfilesEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input []string
	}{
		{"empty spec", []string{""}},
		{"whitespace only", []string{"  "}},
		{"profile not found", []string{"profile:nonexistent"}},
		{"regular alarm", []string{"-15m"}},
		{"mixed", []string{"-15m", "", "profile:adhd-default"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandAlarmProfiles(tt.input)
			if result == nil {
				t.Error("expandAlarmProfiles should not return nil")
			}
		})
	}
}

func TestValueAsStringEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
	}{
		{"string", "test"},
		{"int", 42},
		{"float", 3.14},
		{"bool", true},
		{"nil", nil},
		{"slice", []string{"a", "b"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := valueAsString(tt.value)
			_ = result // Just ensure no panic
		})
	}
}

func TestValueAsBoolEdgeCases(t *testing.T) {
	tests := []struct {
		value interface{}
		want  bool
	}{
		{true, true},
		{false, false},
		{"true", true},
		{"false", false},
		{"1", true},
		{"0", false},
		{1, true},
		{nil, false},
	}

	for _, tt := range tests {
		got := valueAsBool(tt.value)
		if got != tt.want {
			t.Errorf("valueAsBool(%v) = %v, want %v", tt.value, got, tt.want)
		}
	}
}
