package cli

import (
	"strings"
	"tempus/internal/calendar"
	"tempus/internal/testutil"
	"testing"
	"time"
)

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
		{"different lengths", "short", "longer string", 11},
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

func TestMinInt(t *testing.T) {
	tests := []struct {
		name    string
		a, b, c int
		want    int
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
			got := minInt(tt.a, tt.b, tt.c)
			if got != tt.want {
				t.Errorf("minInt(%d, %d, %d) = %d, want %d", tt.a, tt.b, tt.c, got, tt.want)
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
		{"no emoji", testutil.EventTitleHelloWorld, testutil.EventTitleHelloWorld},
		{"with emoji", "\U0001f48a Medication", "Medication"},
		{"emoji in middle", "Take \U0001f48a medicine", "Take \U0001f48a medicine"},
		{testutil.TestNameEmptyString, "", ""},
		{"leading spaces after emoji", "\U0001f48a  Medication", "Medication"},
		{"leading high unicode", "\u00a1Hola", "Hola"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripEmoji(tt.input)
			if got != tt.want {
				t.Errorf("StripEmoji(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGenerateUID(t *testing.T) {
	uid1 := GenerateUID()
	if uid1 == "" {
		t.Error("GenerateUID() returned empty string")
	}

	if !strings.Contains(uid1, "@tempus") {
		t.Errorf("GenerateUID() = %q, should contain @tempus", uid1)
	}

	uid2 := GenerateUID()
	if uid1 == uid2 {
		t.Error("GenerateUID() should generate unique UIDs")
	}
}

func TestDetectEventConflicts(t *testing.T) {
	now := time.Date(2025, 5, 1, 10, 0, 0, 0, time.UTC)

	events := []calendar.Event{
		{
			Summary:   testutil.EventTitleEvent1,
			StartTime: now,
			EndTime:   now.Add(1 * time.Hour),
		},
		{
			Summary:   testutil.EventTitleEvent2,
			StartTime: now.Add(30 * time.Minute),
			EndTime:   now.Add(90 * time.Minute),
		},
		{
			Summary:   "Event 3",
			StartTime: now.Add(2 * time.Hour),
			EndTime:   now.Add(3 * time.Hour),
		},
	}

	conflicts := DetectEventConflicts(events)
	if len(conflicts) == 0 {
		t.Error("DetectEventConflicts() should find conflicts")
	}

	noConflictEvents := []calendar.Event{
		{
			Summary:   testutil.EventTitleEvent1,
			StartTime: now,
			EndTime:   now.Add(1 * time.Hour),
		},
		{
			Summary:   testutil.EventTitleEvent2,
			StartTime: now.Add(2 * time.Hour),
			EndTime:   now.Add(3 * time.Hour),
		},
	}

	noConflicts := DetectEventConflicts(noConflictEvents)
	if len(noConflicts) != 0 {
		t.Errorf("DetectEventConflicts() should not find conflicts, but found %d", len(noConflicts))
	}

	emptyConflicts := DetectEventConflicts([]calendar.Event{})
	if len(emptyConflicts) != 0 {
		t.Error("DetectEventConflicts() with empty slice should return no conflicts")
	}

	singleConflicts := DetectEventConflicts(events[:1])
	if len(singleConflicts) != 0 {
		t.Error("DetectEventConflicts() with single event should return no conflicts")
	}
}

func TestDetectOverwhelmDays(t *testing.T) {
	now := time.Date(2025, 5, 1, 10, 0, 0, 0, time.UTC)
	threshold := 3

	events := []calendar.Event{
		{Summary: testutil.EventTitleEvent1, StartTime: now, EndTime: now.Add(1 * time.Hour)},
		{Summary: testutil.EventTitleEvent2, StartTime: now.Add(2 * time.Hour), EndTime: now.Add(3 * time.Hour)},
		{Summary: "Event 3", StartTime: now.Add(4 * time.Hour), EndTime: now.Add(5 * time.Hour)},
		{Summary: "Event 4", StartTime: now.Add(6 * time.Hour), EndTime: now.Add(7 * time.Hour)},
	}

	overwhelmDays := DetectOverwhelmDays(events, threshold)
	if len(overwhelmDays) == 0 {
		t.Error("DetectOverwhelmDays() should find overwhelmed days")
	}

	belowThreshold := DetectOverwhelmDays(events[:2], threshold)
	if len(belowThreshold) != 0 {
		t.Error("DetectOverwhelmDays() should not find overwhelmed days when below threshold")
	}

	empty := DetectOverwhelmDays([]calendar.Event{}, threshold)
	if len(empty) != 0 {
		t.Error("DetectOverwhelmDays() with empty slice should return no overwhelmed days")
	}

	zeroThreshold := DetectOverwhelmDays(events, 0)
	if len(zeroThreshold) != 0 {
		t.Error("DetectOverwhelmDays() with threshold 0 uses default (8), so 4 events shouldn't be overwhelmed")
	}
}

func TestGeneratePrepTimeEvents(t *testing.T) {
	meetingEvent := calendar.Event{
		Summary:   testutil.EventTitleTeamMeeting,
		StartTime: time.Date(2025, 5, 1, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2025, 5, 1, 11, 0, 0, 0, time.UTC),
		StartTZ:   testutil.TZEuropeMadrid,
		EndTZ:     testutil.TZEuropeMadrid,
	}

	events := []calendar.Event{meetingEvent}
	prepEvents := GeneratePrepTimeEvents(events)

	if len(prepEvents) != 1 {
		t.Errorf("GeneratePrepTimeEvents() returned %d events, want 1", len(prepEvents))
		return
	}

	prepEvent := prepEvents[0]

	if !prepEvent.EndTime.Equal(meetingEvent.StartTime) {
		t.Errorf("prep event should end when main event starts")
	}

	duration := prepEvent.EndTime.Sub(prepEvent.StartTime)
	expectedDuration := 15 * time.Minute
	if duration != expectedDuration {
		t.Errorf("prep event duration = %v, want %v", duration, expectedDuration)
	}

	if !strings.Contains(prepEvent.Summary, "Preparation") {
		t.Errorf("prep event summary should contain 'Preparation', got %q", prepEvent.Summary)
	}

	doctorEvent := calendar.Event{
		Summary:   "Doctor Appointment",
		StartTime: time.Date(2025, 5, 1, 14, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2025, 5, 1, 15, 0, 0, 0, time.UTC),
		StartTZ:   testutil.TZEuropeMadrid,
	}
	medicalPrep := GeneratePrepTimeEvents([]calendar.Event{doctorEvent})
	if len(medicalPrep) != 1 {
		t.Error("doctor appointment should generate prep event")
	} else {
		medDuration := medicalPrep[0].EndTime.Sub(medicalPrep[0].StartTime)
		if medDuration != 20*time.Minute {
			t.Errorf("medical prep duration = %v, want 20m", medDuration)
		}
	}

	focusEvent := calendar.Event{
		Summary:   "Focus Block",
		StartTime: time.Date(2025, 5, 1, 9, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2025, 5, 1, 10, 30, 0, 0, time.UTC),
	}
	focusPrep := GeneratePrepTimeEvents([]calendar.Event{focusEvent})
	if len(focusPrep) != 1 {
		t.Error("focus block should generate transition event")
	} else {
		if !focusPrep[0].StartTime.Equal(focusEvent.EndTime) {
			t.Error("transition should start when focus block ends")
		}
	}

	regularEvent := calendar.Event{
		Summary:   "Regular Event",
		StartTime: time.Date(2025, 5, 1, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2025, 5, 1, 11, 0, 0, 0, time.UTC),
	}
	regularPrep := GeneratePrepTimeEvents([]calendar.Event{regularEvent})
	if len(regularPrep) != 0 {
		t.Error("regular event should not generate prep events")
	}

	allDayEvent := calendar.Event{
		Summary:   testutil.EventTitleTeamMeeting,
		StartTime: time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2025, 5, 2, 0, 0, 0, 0, time.UTC),
		AllDay:    true,
	}
	allDayPrep := GeneratePrepTimeEvents([]calendar.Event{allDayEvent})
	if len(allDayPrep) != 0 {
		t.Error("all-day events should not generate prep events")
	}

	emptyPrepEvents := GeneratePrepTimeEvents([]calendar.Event{})
	if len(emptyPrepEvents) != 0 {
		t.Error("GeneratePrepTimeEvents() with empty slice should return no events")
	}
}

func TestGetSmartDefaultDuration(t *testing.T) {
	tests := []struct {
		name      string
		summary   string
		startTime time.Time
		wantMin   int
	}{
		{"medication", "Take medication", time.Date(2025, 5, 1, 8, 0, 0, 0, time.UTC), 5},
		{"medication with emoji", "\U0001f48a Medicine", time.Date(2025, 5, 1, 8, 0, 0, 0, time.UTC), 5},
		{"focus block", "Focus time", time.Date(2025, 5, 1, 9, 0, 0, 0, time.UTC), 120},
		{"doctor appointment", "Doctor appointment", time.Date(2025, 5, 1, 14, 0, 0, 0, time.UTC), 30},
		{"transition", "Transition period", time.Date(2025, 5, 1, 10, 0, 0, 0, time.UTC), 15},
		{"default business hours", "Random event", time.Date(2025, 5, 1, 12, 0, 0, 0, time.UTC), 60},
		{"empty business hours", "", time.Date(2025, 5, 1, 12, 0, 0, 0, time.UTC), 60},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetSmartDefaultDuration(tt.summary, tt.startTime)
			wantDuration := time.Duration(tt.wantMin) * time.Minute
			if got != wantDuration {
				t.Errorf("GetSmartDefaultDuration(%q, %v) = %v, want %v",
					tt.summary, tt.startTime, got, wantDuration)
			}
		})
	}
}

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
		{"already has emoji", "\U0001f48a Medicine", []string{"medication"}, false},
		{"multiple categories", "Event", []string{"work", "meeting"}, true},
		{"empty summary", "", []string{"work"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AddEmojiToSummary(tt.summary, tt.categories)
			hasEmoji := got != tt.summary
			if hasEmoji != tt.wantEmoji {
				t.Errorf("AddEmojiToSummary(%q, %v) hasEmoji = %v, want %v",
					tt.summary, tt.categories, hasEmoji, tt.wantEmoji)
			}
		})
	}
}

func TestNormalizeAndSpellCheck(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty string", "", ""},
		{"no corrections needed", "Hello World", "Hello World"},
		{"simple text", "Team Meeting", "Team Meeting"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeAndSpellCheck(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeAndSpellCheck(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateCategoryWithSuggestion(t *testing.T) {
	tests := []struct {
		name     string
		category string
		want     string
	}{
		{"exact match lowercase", "work", "Work"},
		{"exact match mixed case", "Work", "Work"},
		{"medication alias", "meds", "Medication"},
		{"typo close to work", "wrk", "Work"},
		{"unknown category", "xyzabc123", "xyzabc123"},
		{"empty", "", ""},
		{"deep work alias", "deep work", "Focus"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateCategoryWithSuggestion(tt.category)
			if got != tt.want {
				t.Errorf("ValidateCategoryWithSuggestion(%q) = %q, want %q", tt.category, got, tt.want)
			}
		})
	}
}
