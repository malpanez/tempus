package main

import (
	"strings"
	"testing"
	"time"

	"tempus/internal/calendar"
)

// ============================================================================
// Utility function tests - covering 0% functions
// ============================================================================

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		name string
		s1   string
		s2   string
		want int
	}{
		{"empty strings", "", "", 0},
		{"empty s1", "", "hello", 5},
		{"empty s2", "hello", "", 5},
		{"identical", "test", "test", 0},
		{"one char different", "test", "best", 1},
		{"completely different", "abc", "xyz", 3},
		{"different lengths", "short", "longer string", 11}, // Actual levenshtein distance
		{"insertion", "cat", "cats", 1},
		{"deletion", "cats", "cat", 1},
		{"substitution", "cat", "bat", 1},
		{"multiple operations", "kitten", "sitting", 3},
		{"case sensitive", "Test", "test", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := levenshteinDistance(tt.s1, tt.s2)
			if got != tt.want {
				t.Errorf("levenshteinDistance(%q, %q) = %d, want %d", tt.s1, tt.s2, got, tt.want)
			}
		})
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		name string
		a, b, c int
		want int
	}{
		{"a is minimum", 1, 2, 3, 1},
		{"b is minimum", 3, 1, 2, 1},
		{"c is minimum", 3, 2, 1, 1},
		{"all equal", 5, 5, 5, 5},
		{"two equal minimum", 2, 2, 3, 2},
		{"negative numbers", -5, -2, -1, -5},
		{"mixed positive and negative", -1, 0, 1, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := min(tt.a, tt.b, tt.c)
			if got != tt.want {
				t.Errorf("min(%d, %d, %d) = %d, want %d", tt.a, tt.b, tt.c, got, tt.want)
			}
		})
	}
}

func TestStripEmoji(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"no emoji", "Hello World", "Hello World"},
		{"with emoji", "💊 Medication", "Medication"},
		{"emoji in middle", "Take 💊 medicine", "Take 💊 medicine"}, // Middle emoji not stripped
		{"empty string", "", ""},
		{"leading spaces after emoji", "💊  Medication", "Medication"},
		{"leading high unicode", "¡Hola", "Hola"}, // Strips first char if > 127
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripEmoji(tt.input)
			if got != tt.want {
				t.Errorf("stripEmoji(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGenerateUID(t *testing.T) {
	// Test that it generates non-empty UIDs
	uid1 := generateUID()
	if uid1 == "" {
		t.Error("generateUID() returned empty string")
	}

	// Test that it includes @tempus
	if !strings.Contains(uid1, "@tempus") {
		t.Errorf("generateUID() = %q, should contain @tempus", uid1)
	}

	// Test that it generates unique UIDs
	uid2 := generateUID()
	if uid1 == uid2 {
		t.Error("generateUID() should generate unique UIDs")
	}
}

func TestAtoiSafe(t *testing.T) {
	// atoiSafe only handles positive integers, returns 0 for invalid/negative
	tests := []struct {
		name string
		s    string
		want int
	}{
		{"valid positive", "123", 123},
		{"valid negative returns 0", "-45", 0}, // No negative support
		{"zero", "0", 0},
		{"invalid", "abc", 0},
		{"empty", "", 0},
		{"float", "3.14", 0},
		{"spaces with number", "  42  ", 42}, // Trims spaces
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := atoiSafe(tt.s)
			if got != tt.want {
				t.Errorf("atoiSafe(%q) = %d, want %d", tt.s, got, tt.want)
			}
		})
	}
}

func TestPrintErr(t *testing.T) {
	// This function prints to stderr, so we just test it doesn't panic
	printErr("test error message")
	printErr("")
	printErr("error with special chars: 💊 😀")
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", "Hello World", "hello-world"},
		{"uppercase", "TEST STRING", "test-string"},
		{"with underscores", "test_name_here", "test-name-here"},
		{"multiple spaces", "hello    world", "hello-world"},
		{"leading/trailing spaces", "  hello world  ", "hello-world"},
		{"special chars", "hello@world!test", "hello-world-test"},
		{"numbers", "test123", "test123"},
		{"mixed", "Test_123 Hello!", "test-123-hello"},
		{"empty", "", ""},
		{"only special chars", "@#$%", "event"}, // Returns "event" for empty/special-only
		{"hyphen already exists", "hello-world", "hello-world"},
		{"consecutive hyphens", "hello--world", "hello-world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slugify(tt.input)
			if got != tt.want {
				t.Errorf("slugify(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ============================================================================
// Event processing functions with 0% coverage
// ============================================================================

func TestDetectEventConflicts(t *testing.T) {
	now := time.Date(2025, 5, 1, 10, 0, 0, 0, time.UTC)

	events := []calendar.Event{
		{
			Summary:   "Event 1",
			StartTime: now,
			EndTime:   now.Add(1 * time.Hour),
		},
		{
			Summary:   "Event 2",
			StartTime: now.Add(30 * time.Minute),
			EndTime:   now.Add(90 * time.Minute),
		},
		{
			Summary:   "Event 3",
			StartTime: now.Add(2 * time.Hour),
			EndTime:   now.Add(3 * time.Hour),
		},
	}

	conflicts := detectEventConflicts(events)

	// Should detect conflict between Event 1 and Event 2
	if len(conflicts) == 0 {
		t.Error("detectEventConflicts() should find conflicts")
	}

	// Test with no conflicts
	noConflictEvents := []calendar.Event{
		{
			Summary:   "Event 1",
			StartTime: now,
			EndTime:   now.Add(1 * time.Hour),
		},
		{
			Summary:   "Event 2",
			StartTime: now.Add(2 * time.Hour),
			EndTime:   now.Add(3 * time.Hour),
		},
	}

	noConflicts := detectEventConflicts(noConflictEvents)
	if len(noConflicts) != 0 {
		t.Errorf("detectEventConflicts() should not find conflicts, but found %d", len(noConflicts))
	}

	// Test with empty slice
	emptyConflicts := detectEventConflicts([]calendar.Event{})
	if len(emptyConflicts) != 0 {
		t.Error("detectEventConflicts() with empty slice should return no conflicts")
	}

	// Test with single event
	singleConflicts := detectEventConflicts(events[:1])
	if len(singleConflicts) != 0 {
		t.Error("detectEventConflicts() with single event should return no conflicts")
	}
}

func TestDetectOverwhelmDays(t *testing.T) {
	now := time.Date(2025, 5, 1, 10, 0, 0, 0, time.UTC)
	threshold := 3

	// Create events on same day
	events := []calendar.Event{
		{
			Summary:   "Event 1",
			StartTime: now,
			EndTime:   now.Add(1 * time.Hour),
		},
		{
			Summary:   "Event 2",
			StartTime: now.Add(2 * time.Hour),
			EndTime:   now.Add(3 * time.Hour),
		},
		{
			Summary:   "Event 3",
			StartTime: now.Add(4 * time.Hour),
			EndTime:   now.Add(5 * time.Hour),
		},
		{
			Summary:   "Event 4",
			StartTime: now.Add(6 * time.Hour),
			EndTime:   now.Add(7 * time.Hour),
		},
	}

	overwhelmDays := detectOverwhelmDays(events, threshold)

	// Should detect the day as overwhelmed (4 events > 3 threshold)
	if len(overwhelmDays) == 0 {
		t.Error("detectOverwhelmDays() should find overwhelmed days")
	}

	// Test with events below threshold
	belowThreshold := detectOverwhelmDays(events[:2], threshold)
	if len(belowThreshold) != 0 {
		t.Error("detectOverwhelmDays() should not find overwhelmed days when below threshold")
	}

	// Test with empty slice
	empty := detectOverwhelmDays([]calendar.Event{}, threshold)
	if len(empty) != 0 {
		t.Error("detectOverwhelmDays() with empty slice should return no overwhelmed days")
	}

	// Test with zero threshold (uses default of 8)
	zeroThreshold := detectOverwhelmDays(events, 0)
	if len(zeroThreshold) != 0 {
		t.Error("detectOverwhelmDays() with threshold 0 uses default (8), so 4 events shouldn't be overwhelmed")
	}
}

func TestGeneratePrepTimeEvents(t *testing.T) {
	// This function auto-detects events that need prep time based on keywords
	// and generates prep events before them

	// Test meeting (should get 15min prep)
	meetingEvent := calendar.Event{
		Summary:   "Team Meeting",
		StartTime: time.Date(2025, 5, 1, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2025, 5, 1, 11, 0, 0, 0, time.UTC),
		StartTZ:   "Europe/Madrid",
		EndTZ:     "Europe/Madrid",
	}

	events := []calendar.Event{meetingEvent}
	prepEvents := generatePrepTimeEvents(events)

	// Should generate one prep event
	if len(prepEvents) != 1 {
		t.Errorf("generatePrepTimeEvents() returned %d events, want 1", len(prepEvents))
		return
	}

	prepEvent := prepEvents[0]

	// Check prep event is before main event
	if !prepEvent.EndTime.Equal(meetingEvent.StartTime) {
		t.Errorf("prep event should end when main event starts")
	}

	// Check prep event duration (meetings get 15min)
	duration := prepEvent.EndTime.Sub(prepEvent.StartTime)
	expectedDuration := 15 * time.Minute
	if duration != expectedDuration {
		t.Errorf("prep event duration = %v, want %v", duration, expectedDuration)
	}

	// Check summary contains preparation indicator
	if !strings.Contains(prepEvent.Summary, "Preparation") {
		t.Errorf("prep event summary should contain 'Preparation', got %q", prepEvent.Summary)
	}

	// Test medical event (should get 20min prep)
	doctorEvent := calendar.Event{
		Summary:   "Doctor Appointment",
		StartTime: time.Date(2025, 5, 1, 14, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2025, 5, 1, 15, 0, 0, 0, time.UTC),
		StartTZ:   "Europe/Madrid",
	}
	medicalPrep := generatePrepTimeEvents([]calendar.Event{doctorEvent})
	if len(medicalPrep) != 1 {
		t.Error("doctor appointment should generate prep event")
	} else {
		medDuration := medicalPrep[0].EndTime.Sub(medicalPrep[0].StartTime)
		if medDuration != 20*time.Minute {
			t.Errorf("medical prep duration = %v, want 20m", medDuration)
		}
	}

	// Test focus block (should get transition AFTER, not before)
	focusEvent := calendar.Event{
		Summary:   "Focus Block",
		StartTime: time.Date(2025, 5, 1, 9, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2025, 5, 1, 10, 30, 0, 0, time.UTC),
	}
	focusPrep := generatePrepTimeEvents([]calendar.Event{focusEvent})
	if len(focusPrep) != 1 {
		t.Error("focus block should generate transition event")
	} else {
		// Transition should start when focus block ends
		if !focusPrep[0].StartTime.Equal(focusEvent.EndTime) {
			t.Error("transition should start when focus block ends")
		}
	}

	// Test event without prep keywords
	regularEvent := calendar.Event{
		Summary:   "Regular Event",
		StartTime: time.Date(2025, 5, 1, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2025, 5, 1, 11, 0, 0, 0, time.UTC),
	}
	regularPrep := generatePrepTimeEvents([]calendar.Event{regularEvent})
	if len(regularPrep) != 0 {
		t.Error("regular event should not generate prep events")
	}

	// Test all-day event (should not get prep)
	allDayEvent := calendar.Event{
		Summary:   "Team Meeting",
		StartTime: time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2025, 5, 2, 0, 0, 0, 0, time.UTC),
		AllDay:    true,
	}
	allDayPrep := generatePrepTimeEvents([]calendar.Event{allDayEvent})
	if len(allDayPrep) != 0 {
		t.Error("all-day events should not generate prep events")
	}

	// Test with empty slice
	emptyPrepEvents := generatePrepTimeEvents([]calendar.Event{})
	if len(emptyPrepEvents) != 0 {
		t.Error("generatePrepTimeEvents() with empty slice should return no events")
	}
}

// ============================================================================
// Smart duration detection
// ============================================================================

func TestGetSmartDefaultDuration(t *testing.T) {
	tests := []struct {
		name      string
		summary   string
		startTime time.Time
		wantMin   int
	}{
		{"medication", "Take medication", time.Date(2025, 5, 1, 8, 0, 0, 0, time.UTC), 5},
		{"medication with emoji", "💊 Medicine", time.Date(2025, 5, 1, 8, 0, 0, 0, time.UTC), 5},
		{"focus block", "Focus time", time.Date(2025, 5, 1, 9, 0, 0, 0, time.UTC), 120}, // 2 hours
		{"doctor appointment", "Doctor appointment", time.Date(2025, 5, 1, 14, 0, 0, 0, time.UTC), 30},
		{"transition", "Transition period", time.Date(2025, 5, 1, 10, 0, 0, 0, time.UTC), 15},
		{"default business hours", "Random event", time.Date(2025, 5, 1, 12, 0, 0, 0, time.UTC), 60},
		{"empty business hours", "", time.Date(2025, 5, 1, 12, 0, 0, 0, time.UTC), 60},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getSmartDefaultDuration(tt.summary, tt.startTime)
			wantDuration := time.Duration(tt.wantMin) * time.Minute
			if got != wantDuration {
				t.Errorf("getSmartDefaultDuration(%q, %v) = %v, want %v",
					tt.summary, tt.startTime, got, wantDuration)
			}
		})
	}
}

// ============================================================================
// Emoji and category functions
// ============================================================================

func TestAddEmojiToSummary(t *testing.T) {
	tests := []struct {
		name       string
		summary    string
		categories []string
		wantEmoji  bool
	}{
		{"medication", "Take pills", []string{"medication"}, true},
		{"work", "Team meeting", []string{"work"}, true},
		{"appointment", "Doctor visit", []string{"appointment"}, true},
		{"health", "Checkup", []string{"health"}, true},
		{"no category", "Event", []string{}, false},
		{"already has emoji", "💊 Medicine", []string{"medication"}, false},
		{"multiple categories", "Event", []string{"work", "meeting"}, true},
		{"empty summary", "", []string{"work"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := addEmojiToSummary(tt.summary, tt.categories)
			hasEmoji := got != tt.summary
			if hasEmoji != tt.wantEmoji {
				t.Errorf("addEmojiToSummary(%q, %v) hasEmoji = %v, want %v",
					tt.summary, tt.categories, hasEmoji, tt.wantEmoji)
			}
		})
	}
}

// ============================================================================
// Command creation tests for 0% coverage commands
// ============================================================================

func TestNewTimezoneCmd(t *testing.T) {
	cmd := newTimezoneCmd()
	if cmd == nil {
		t.Fatal("newTimezoneCmd() returned nil")
	}
	if cmd.Use != "timezone" {
		t.Errorf("Use = %q, want %q", cmd.Use, "timezone")
	}

	// Check subcommands
	subcommands := cmd.Commands()
	if len(subcommands) != 2 {
		t.Errorf("expected 2 subcommands, got %d", len(subcommands))
	}

	var hasList, hasInfo bool
	for _, sub := range subcommands {
		if strings.HasPrefix(sub.Use, "list") {
			hasList = true
		}
		if strings.HasPrefix(sub.Use, "info") {
			hasInfo = true
		}
	}
	if !hasList {
		t.Error("timezone command missing 'list' subcommand")
	}
	if !hasInfo {
		t.Error("timezone command missing 'info' subcommand")
	}
}

func TestNewRRuleHelperCmd(t *testing.T) {
	cmd := newRRuleHelperCmd()
	if cmd == nil {
		t.Fatal("newRRuleHelperCmd() returned nil")
	}
	if cmd.Use != "rrule" {
		t.Errorf("Use = %q, want %q", cmd.Use, "rrule")
	}
	if cmd.RunE == nil {
		t.Error("rrule command should have RunE function")
	}
}

func TestNewLocaleCmd(t *testing.T) {
	cmd := newLocaleCmd()
	if cmd == nil {
		t.Fatal("newLocaleCmd() returned nil")
	}
	if cmd.Use != "locale" {
		t.Errorf("Use = %q, want %q", cmd.Use, "locale")
	}

	// Check subcommands
	subcommands := cmd.Commands()
	if len(subcommands) != 1 {
		t.Errorf("expected 1 subcommand, got %d", len(subcommands))
	}

	listCmd := subcommands[0]
	if !strings.HasPrefix(listCmd.Use, "list") {
		t.Error("locale command should have 'list' subcommand")
	}
}

func TestNewTemplateCmd(t *testing.T) {
	cmd := newTemplateCmd()
	if cmd == nil {
		t.Fatal("newTemplateCmd() returned nil")
	}
	if cmd.Use != "template" {
		t.Errorf("Use = %q, want %q", cmd.Use, "template")
	}

	// Check subcommands exist
	subcommands := cmd.Commands()
	if len(subcommands) < 5 {
		t.Errorf("expected at least 5 subcommands, got %d", len(subcommands))
	}

	var hasList, hasDescribe, hasCreate, hasValidate, hasInit bool
	for _, sub := range subcommands {
		use := sub.Use
		if strings.HasPrefix(use, "list") {
			hasList = true
		}
		if strings.HasPrefix(use, "describe") {
			hasDescribe = true
		}
		if strings.HasPrefix(use, "create") {
			hasCreate = true
		}
		if strings.HasPrefix(use, "validate") {
			hasValidate = true
		}
		if strings.HasPrefix(use, "init") {
			hasInit = true
		}
	}

	if !hasList {
		t.Error("template command missing 'list' subcommand")
	}
	if !hasDescribe {
		t.Error("template command missing 'describe' subcommand")
	}
	if !hasCreate {
		t.Error("template command missing 'create' subcommand")
	}
	if !hasValidate {
		t.Error("template command missing 'validate' subcommand")
	}
	if !hasInit {
		t.Error("template command missing 'init' subcommand")
	}
}

// ============================================================================
// Batch template functions (0% coverage)
// ============================================================================

func TestGetBasicTemplate(t *testing.T) {
	content := getBasicTemplate()
	if content == "" {
		t.Error("getBasicTemplate() returned empty string")
	}
	if !strings.Contains(content, "summary") {
		t.Error("basic template should contain 'summary' field")
	}
	if !strings.Contains(content, "start") {
		t.Error("basic template should contain 'start' field")
	}
}

func TestGetADHDRoutineTemplate(t *testing.T) {
	content := getADHDRoutineTemplate()
	if content == "" {
		t.Error("getADHDRoutineTemplate() returned empty string")
	}
	// Should contain typical ADHD routine events
	if !strings.Contains(content, "Morning") && !strings.Contains(content, "medication") {
		t.Error("ADHD routine template should contain morning or medication events")
	}
}

func TestGetMedicationTemplate(t *testing.T) {
	content := getMedicationTemplate()
	if content == "" {
		t.Error("getMedicationTemplate() returned empty string")
	}
	if !strings.Contains(content, "medication") && !strings.Contains(content, "Medication") {
		t.Error("medication template should contain medication-related content")
	}
}

func TestGetWorkMeetingsTemplate(t *testing.T) {
	content := getWorkMeetingsTemplate()
	if content == "" {
		t.Error("getWorkMeetingsTemplate() returned empty string")
	}
	if !strings.Contains(content, "meeting") && !strings.Contains(content, "Meeting") {
		t.Error("work meetings template should contain meeting-related content")
	}
}

func TestGetMedicalTemplate(t *testing.T) {
	content := getMedicalTemplate()
	if content == "" {
		t.Error("getMedicalTemplate() returned empty string")
	}
	if !strings.Contains(content, "appointment") && !strings.Contains(content, "medical") {
		t.Error("medical template should contain appointment or medical content")
	}
}

func TestGetTravelTemplate(t *testing.T) {
	content := getTravelTemplate()
	if content == "" {
		t.Error("getTravelTemplate() returned empty string")
	}
	// Travel template might contain flight or travel-related terms
	if !strings.Contains(content, "flight") && !strings.Contains(content, "Flight") &&
		!strings.Contains(content, "travel") && !strings.Contains(content, "Travel") {
		t.Error("travel template should contain flight or travel-related content")
	}
}

func TestGetFamilyTemplate(t *testing.T) {
	content := getFamilyTemplate()
	if content == "" {
		t.Error("getFamilyTemplate() returned empty string")
	}
	// Family template might contain family-related events
}

func TestGetBatchTemplateContent(t *testing.T) {
	tests := []struct {
		name        string
		templateKey string
		wantEmpty   bool
		wantErr     bool
	}{
		{"basic", "basic", false, false},
		{"adhd-routine", "adhd-routine", false, false},
		{"medication", "medication", false, false},
		{"work-meetings", "work-meetings", false, false},
		{"medical", "medical", false, false},
		{"travel", "travel", false, false},
		{"family", "family", false, false},
		{"unknown", "unknown-template", true, true},
		{"empty", "", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getBatchTemplateContent(tt.templateKey)
			isEmpty := got == ""
			hasErr := err != nil
			if isEmpty != tt.wantEmpty {
				t.Errorf("getBatchTemplateContent(%q) isEmpty = %v, want %v",
					tt.templateKey, isEmpty, tt.wantEmpty)
			}
			if hasErr != tt.wantErr {
				t.Errorf("getBatchTemplateContent(%q) hasErr = %v, want %v",
					tt.templateKey, hasErr, tt.wantErr)
			}
		})
	}
}

// ============================================================================
// RRULE interpretation (0% coverage)
// ============================================================================

func TestInterpretRRule(t *testing.T) {
	tests := []struct {
		name      string
		rrule     string
		wantEmpty bool
	}{
		{"daily", "FREQ=DAILY", false},
		{"weekly", "FREQ=WEEKLY;BYDAY=MO,WE,FR", false},
		{"monthly", "FREQ=MONTHLY;BYMONTHDAY=15", false},
		{"yearly", "FREQ=YEARLY", false},
		{"with count", "FREQ=DAILY;COUNT=10", false},
		{"with until", "FREQ=WEEKLY;UNTIL=20251231", false},
		{"complex", "FREQ=MONTHLY;BYDAY=1MO;COUNT=12", false},
		{"empty", "", false}, // Returns default message even for empty
		{"invalid", "INVALID", false}, // Should still return something
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := interpretRRule(tt.rrule)
			isEmpty := got == ""
			if isEmpty != tt.wantEmpty {
				t.Errorf("interpretRRule(%q) isEmpty = %v, want %v, got %q",
					tt.rrule, isEmpty, tt.wantEmpty, got)
			}
		})
	}
}
