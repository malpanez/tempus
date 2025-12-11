package calendar

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestNewCalendar(t *testing.T) {
	cal := NewCalendar()

	if cal == nil {
		t.Fatal("NewCalendar() returned nil")
	}

	if cal.ProdID == "" {
		t.Error("Calendar has empty ProdID")
	}

	if cal.Version != "2.0" {
		t.Errorf("Calendar Version = %s, want 2.0", cal.Version)
	}

	if cal.CalScale != "GREGORIAN" {
		t.Errorf("Calendar CalScale = %s, want GREGORIAN", cal.CalScale)
	}

	if cal.Method != "PUBLISH" {
		t.Errorf("Calendar Method = %s, want PUBLISH", cal.Method)
	}

	if cal.Events == nil {
		t.Error("Calendar Events is nil")
	}
}

func TestNewEvent(t *testing.T) {
	summary := "Test Event"
	start := time.Now()
	end := start.Add(1 * time.Hour)

	event := NewEvent(summary, start, end)

	if event == nil {
		t.Fatal("NewEvent() returned nil")
	}

	if event.Summary != summary {
		t.Errorf("Event Summary = %q, want %q", event.Summary, summary)
	}

	if !event.StartTime.Equal(start) {
		t.Errorf("Event StartTime = %v, want %v", event.StartTime, start)
	}

	if !event.EndTime.Equal(end) {
		t.Errorf("Event EndTime = %v, want %v", event.EndTime, end)
	}

	if event.UID == "" {
		t.Error("Event UID is empty")
	}

	if event.Status != "CONFIRMED" {
		t.Errorf("Event Status = %s, want CONFIRMED", event.Status)
	}

	if event.Created.IsZero() {
		t.Error("Event Created time is zero")
	}

	if event.LastMod.IsZero() {
		t.Error("Event LastMod time is zero")
	}
}

func TestAddEvent(t *testing.T) {
	cal := NewCalendar()
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))

	cal.AddEvent(event)

	if len(cal.Events) != 1 {
		t.Errorf("Calendar has %d events, want 1", len(cal.Events))
	}

	if cal.Events[0].Summary != "Test" {
		t.Errorf("Event summary = %q, want %q", cal.Events[0].Summary, "Test")
	}
}

func TestSetters(t *testing.T) {
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))

	event.SetStartTimezone("America/New_York")
	if event.StartTZ != "America/New_York" {
		t.Errorf("StartTZ = %s, want America/New_York", event.StartTZ)
	}

	event.SetEndTimezone("Europe/London")
	if event.EndTZ != "Europe/London" {
		t.Errorf("EndTZ = %s, want Europe/London", event.EndTZ)
	}

	event.AllDay = true
	if !event.AllDay {
		t.Error("AllDay should be true")
	}

	cal := NewCalendar()
	cal.DefaultTZ = "UTC"
	if cal.DefaultTZ != "UTC" {
		t.Errorf("DefaultTZ = %s, want UTC", cal.DefaultTZ)
	}
}

func TestAddAttendee(t *testing.T) {
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))

	event.AddAttendee("alice@example.com")
	event.AddAttendee("bob@example.com")

	if len(event.Attendees) != 2 {
		t.Errorf("Event has %d attendees, want 2", len(event.Attendees))
	}

	if event.Attendees[0] != "alice@example.com" {
		t.Errorf("Attendee[0] = %s, want alice@example.com", event.Attendees[0])
	}
}

func TestAddCategory(t *testing.T) {
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))

	event.AddCategory("Work")
	event.AddCategory("Meeting")

	if len(event.Categories) != 2 {
		t.Errorf("Event has %d categories, want 2", len(event.Categories))
	}

	if event.Categories[0] != "Work" {
		t.Errorf("Category[0] = %s, want Work", event.Categories[0])
	}
}

func TestAddAlarm(t *testing.T) {
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))

	alarm := Alarm{
		Action:            "DISPLAY",
		Description:       "Reminder",
		TriggerIsRelative: true,
		TriggerDuration:   -15 * time.Minute,
	}

	event.Alarms = append(event.Alarms, alarm)

	if len(event.Alarms) != 1 {
		t.Errorf("Event has %d alarms, want 1", len(event.Alarms))
	}

	if event.Alarms[0].Description != "Reminder" {
		t.Errorf("Alarm description = %s, want Reminder", event.Alarms[0].Description)
	}
}

func TestCalendarToICSBasic(t *testing.T) {
	cal := NewCalendar()
	start := time.Date(2025, 11, 15, 10, 0, 0, 0, time.UTC)
	end := start.Add(1 * time.Hour)

	event := NewEvent("Meeting", start, end)
	cal.AddEvent(event)

	ics := cal.ToICS()

	// Check for required ICS components
	requiredFields := []string{
		"BEGIN:VCALENDAR",
		"VERSION:2.0",
		"PRODID:",
		"CALSCALE:GREGORIAN",
		"BEGIN:VEVENT",
		"UID:",
		"SUMMARY:Meeting",
		"DTSTART:",
		"DTEND:",
		"STATUS:CONFIRMED",
		"END:VEVENT",
		"END:VCALENDAR",
	}

	for _, field := range requiredFields {
		if !strings.Contains(ics, field) {
			t.Errorf("ICS missing required field: %s", field)
		}
	}
}

func TestCalendarToICSIncludesGoogleFriendlyMetadata(t *testing.T) {
	cal := NewCalendar()
	cal.Name = "Consulta medica"
	cal.DefaultTZ = "Europe/Madrid"

	start := time.Date(2025, time.September, 10, 12, 0, 0, 0, time.FixedZone("CEST", 2*60*60))
	end := start.Add(45 * time.Minute)

	event := NewEvent("Consulta medica", start, end)
	event.SetStartTimezone("Europe/Madrid")
	event.SetEndTimezone("Europe/Madrid")
	cal.AddEvent(event)

	ics := cal.ToICS()

	if !strings.Contains(ics, "METHOD:PUBLISH") {
		t.Fatalf("expected METHOD:PUBLISH header, got:\n%s", ics)
	}
	if !strings.Contains(ics, "X-WR-CALNAME:Consulta medica") {
		t.Fatalf("expected calendar name header, got:\n%s", ics)
	}
	if !strings.Contains(ics, "X-WR-TIMEZONE:Europe/Madrid") {
		t.Fatalf("expected X-WR-TIMEZONE header, got:\n%s", ics)
	}
	if !strings.Contains(ics, "DTSTART;TZID=Europe/Madrid:20250910T120000") {
		t.Fatalf("expected DTSTART with timezone, got:\n%s", ics)
	}
	if !strings.Contains(ics, "DTEND;TZID=Europe/Madrid:20250910T124500") {
		t.Fatalf("expected DTEND with timezone, got:\n%s", ics)
	}
}

func TestAllDayEvent(t *testing.T) {
	cal := NewCalendar()
	start := time.Date(2025, 11, 15, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	event := NewEvent("All Day Event", start, end)
	event.AllDay = true
	cal.AddEvent(event)

	ics := cal.ToICS()

	// All-day events use DTSTART;VALUE=DATE format
	if !strings.Contains(ics, "DTSTART;VALUE=DATE:20251115") {
		t.Errorf("All-day event should have VALUE=DATE format")
	}
}

func TestEventWithDescription(t *testing.T) {
	cal := NewCalendar()
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))
	event.Description = "This is a test description"
	cal.AddEvent(event)

	ics := cal.ToICS()

	if !strings.Contains(ics, "DESCRIPTION:This is a test description") {
		t.Error("ICS should contain description")
	}
}

func TestEventWithLocation(t *testing.T) {
	cal := NewCalendar()
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))
	event.Location = "Room 101"
	cal.AddEvent(event)

	ics := cal.ToICS()

	if !strings.Contains(ics, "LOCATION:Room 101") {
		t.Error("ICS should contain location")
	}
}

func TestEventWithPriority(t *testing.T) {
	cal := NewCalendar()
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))
	event.Priority = 5
	cal.AddEvent(event)

	ics := cal.ToICS()

	if !strings.Contains(ics, "PRIORITY:5") {
		t.Error("ICS should contain priority")
	}
}

func TestEventWithRecurrence(t *testing.T) {
	cal := NewCalendar()
	event := NewEvent("Recurring", time.Now(), time.Now().Add(1*time.Hour))
	event.RRule = "FREQ=DAILY;COUNT=5"
	cal.AddEvent(event)

	ics := cal.ToICS()

	if !strings.Contains(ics, "RRULE:FREQ=DAILY;COUNT=5") {
		t.Error("ICS should contain RRULE")
	}
}

func TestEventWithAlarms(t *testing.T) {
	cal := NewCalendar()
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))

	alarm := Alarm{
		Action:            "DISPLAY",
		Description:       "15 minute reminder",
		TriggerIsRelative: true,
		TriggerDuration:   -15 * time.Minute,
	}
	event.Alarms = append(event.Alarms, alarm)
	cal.AddEvent(event)

	ics := cal.ToICS()

	requiredAlarmFields := []string{
		"BEGIN:VALARM",
		"ACTION:DISPLAY",
		"DESCRIPTION:15 minute reminder",
		"TRIGGER:-PT15M",
		"END:VALARM",
	}

	for _, field := range requiredAlarmFields {
		if !strings.Contains(ics, field) {
			t.Errorf("Alarm missing field: %s", field)
		}
	}
}

func TestMultipleEvents(t *testing.T) {
	cal := NewCalendar()

	for i := 0; i < 3; i++ {
		start := time.Now().Add(time.Duration(i) * 24 * time.Hour)
		event := NewEvent(fmt.Sprintf("Event %d", i), start, start.Add(1*time.Hour))
		cal.AddEvent(event)
	}

	if len(cal.Events) != 3 {
		t.Errorf("Calendar has %d events, want 3", len(cal.Events))
	}

	ics := cal.ToICS()

	// Should have 3 VEVENT blocks
	count := strings.Count(ics, "BEGIN:VEVENT")
	if count != 3 {
		t.Errorf("ICS has %d VEVENTs, want 3", count)
	}
}

func TestEscapeICSText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{"comma", "Hello, World", "Hello\\, World"},
		{"semicolon", "Time: 10:30", "Time: 10:30"}, // Colon doesn't need escaping
		{"newline", "Line1\nLine2", "Line1\\nLine2"},
		{"backslash", "Path\\to\\file", "Path\\\\to\\\\file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cal := NewCalendar()
			event := NewEvent(tt.input, time.Now(), time.Now().Add(1*time.Hour))
			cal.AddEvent(event)

			ics := cal.ToICS()

			if !strings.Contains(ics, tt.contains) {
				t.Errorf("ICS should contain escaped text: %s", tt.contains)
			}
		})
	}
}

func TestUIDGeneration(t *testing.T) {
	event1 := NewEvent("Test1", time.Now(), time.Now().Add(1*time.Hour))
	event2 := NewEvent("Test2", time.Now(), time.Now().Add(1*time.Hour))

	if event1.UID == event2.UID {
		t.Error("Two events should have different UIDs")
	}

	if !strings.HasSuffix(event1.UID, "@tempus") {
		t.Errorf("UID should end with @tempus, got: %s", event1.UID)
	}
}

func TestCategories(t *testing.T) {
	cal := NewCalendar()
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))
	event.AddCategory("Work")
	event.AddCategory("Important")
	cal.AddEvent(event)

	ics := cal.ToICS()

	if !strings.Contains(ics, "CATEGORIES:Work,Important") {
		t.Error("ICS should contain categories")
	}
}

func TestAttendees(t *testing.T) {
	cal := NewCalendar()
	event := NewEvent("Meeting", time.Now(), time.Now().Add(1*time.Hour))
	event.AddAttendee("alice@example.com")
	event.AddAttendee("bob@example.com")
	cal.AddEvent(event)

	ics := cal.ToICS()

	if !strings.Contains(ics, "ATTENDEE:") {
		t.Error("ICS should contain attendees")
	}
}

func TestLineFolding(t *testing.T) {
	cal := NewCalendar()

	// Create a very long description to trigger line folding (>75 chars)
	longDesc := strings.Repeat("A", 100)
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))
	event.Description = longDesc
	cal.AddEvent(event)

	ics := cal.ToICS()

	// RFC 5545 requires lines to be folded at 75 octets
	// Check that we don't have lines longer than 75 chars (excluding CRLF)
	lines := strings.Split(ics, "\n")
	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		if len(line) > 75 {
			t.Errorf("Line too long (%d chars): %s", len(line), line)
		}
	}
}

// ========================================
// Test escapeText function edge cases
// ========================================

func TestEscapeText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"no special chars", "Hello World", "Hello World"},
		{"comma", "Hello, World", "Hello\\, World"},
		{"semicolon", "Time; Space", "Time\\; Space"},
		{"backslash", "Path\\to\\file", "Path\\\\to\\\\file"},
		{"newline", "Line1\nLine2", "Line1\\nLine2"},
		{"CRLF", "Line1\r\nLine2", "Line1\\nLine2"},
		{"CR only", "Line1\rLine2", "Line1Line2"},
		{"all special chars", "Test\\;,\n", "Test\\\\\\;\\,\\n"},
		{"multiple newlines", "A\nB\nC", "A\\nB\\nC"},
		{"backslash before comma", "\\,test", "\\\\\\,test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeText(tt.input)
			if result != tt.expected {
				t.Errorf("escapeText(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// ========================================
// Test normalizeUserNewlines function
// ========================================

func TestNormalizeUserNewlines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"no newlines", "Hello World", "Hello World"},
		{"escaped newline", "Line1\\nLine2", "Line1\nLine2"},
		{"multiple escaped newlines", "A\\nB\\nC", "A\nB\nC"},
		{"real newline unchanged", "A\nB", "A\nB"},
		{"mixed", "A\\nB\nC", "A\nB\nC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeUserNewlines(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeUserNewlines(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// ========================================
// Test foldICalLine function edge cases
// ========================================

func TestFoldICalLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		limit    int
		expected int // expected number of segments
	}{
		{"empty string", "", 75, 1},
		{"short line", "SHORT", 75, 1},
		{"exactly limit", strings.Repeat("A", 75), 75, 1},
		{"one over limit", strings.Repeat("A", 76), 75, 2},
		{"double limit", strings.Repeat("A", 150), 75, 2},
		{"zero limit", "test", 0, 1}, // should not fold
		{"negative limit", "test", -1, 1}, // should not fold
		{"unicode chars", "Hello世界" + strings.Repeat("A", 70), 75, 2},
		{"long line 200 chars", strings.Repeat("B", 200), 75, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := foldICalLine(tt.input, tt.limit)
			if len(result) != tt.expected {
				t.Errorf("foldICalLine(%q, %d) returned %d segments, want %d",
					tt.input[:min(len(tt.input), 20)], tt.limit, len(result), tt.expected)
			}
			// Verify reconstruction matches original
			if tt.limit > 0 && len(tt.input) > tt.limit {
				reconstructed := strings.Join(result, "")
				if reconstructed != tt.input {
					t.Errorf("Reconstructed string doesn't match original")
				}
			}
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ========================================
// Test formatICSDuration function
// ========================================

func TestFormatICSDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"zero", 0, "PT0S"},
		{"15 minutes", 15 * time.Minute, "PT15M"},
		{"negative 15 minutes", -15 * time.Minute, "-PT15M"},
		{"1 hour", 1 * time.Hour, "PT1H"},
		{"1 hour 30 minutes", 90 * time.Minute, "PT1H30M"},
		{"negative 1 hour", -1 * time.Hour, "-PT1H"},
		{"1 day", 24 * time.Hour, "P1D"},
		{"1 day 2 hours", 26 * time.Hour, "P1DT2H"},
		{"negative 1 day", -24 * time.Hour, "-P1D"},
		{"30 seconds", 30 * time.Second, "PT30S"},
		{"1 hour 30 min 45 sec", 1*time.Hour + 30*time.Minute + 45*time.Second, "PT1H30M45S"},
		{"2 days 3 hours 15 min", 2*24*time.Hour + 3*time.Hour + 15*time.Minute, "P2DT3H15M"},
		{"7 days", 7 * 24 * time.Hour, "P7D"},
		{"negative complex", -(2*24*time.Hour + 5*time.Hour + 30*time.Minute), "-P2DT5H30M"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatICSDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("formatICSDuration(%v) = %q, want %q", tt.duration, result, tt.expected)
			}
		})
	}
}

// ========================================
// Test ParseDateTime function
// ========================================

func TestParseDateTime(t *testing.T) {
	tests := []struct {
		name     string
		dateStr  string
		timeStr  string
		timezone string
		wantErr  bool
	}{
		{"date only", "2025-11-15", "", "", false},
		{"date and time", "2025-11-15", "14:30", "", false},
		{"with timezone", "2025-11-15", "14:30", "America/New_York", false},
		{"invalid timezone", "2025-11-15", "14:30", "Invalid/Zone", true},
		{"invalid date", "2025-13-32", "", "", true},
		{"invalid time", "2025-11-15", "25:99", "", true},
		{"UTC timezone", "2025-11-15", "14:30", "UTC", false},
		{"Europe/Madrid", "2025-11-15", "14:30", "Europe/Madrid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDateTime(tt.dateStr, tt.timeStr, tt.timezone)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDateTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result.IsZero() {
				t.Errorf("ParseDateTime() returned zero time")
			}
		})
	}
}

// ========================================
// Test CommonTimezones function
// ========================================

func TestCommonTimezones(t *testing.T) {
	timezones := CommonTimezones()

	if len(timezones) == 0 {
		t.Error("CommonTimezones() returned empty map")
	}

	// Check for expected timezones
	expectedTZs := map[string]string{
		"madrid":   "Europe/Madrid",
		"utc":      "UTC",
		"new_york": "America/New_York",
		"tokyo":    "Asia/Tokyo",
	}

	for key, expected := range expectedTZs {
		if tz, ok := timezones[key]; !ok {
			t.Errorf("CommonTimezones() missing key %q", key)
		} else if tz != expected {
			t.Errorf("CommonTimezones()[%q] = %q, want %q", key, tz, expected)
		}
	}
}

// ========================================
// Test SetDefaultTimezone
// ========================================

func TestSetDefaultTimezone(t *testing.T) {
	cal := NewCalendar()

	cal.SetDefaultTimezone("Europe/Madrid")
	if cal.DefaultTZ != "Europe/Madrid" {
		t.Errorf("SetDefaultTimezone() = %q, want %q", cal.DefaultTZ, "Europe/Madrid")
	}

	// Test trimming
	cal.SetDefaultTimezone("  America/New_York  ")
	if cal.DefaultTZ != "America/New_York" {
		t.Errorf("SetDefaultTimezone() should trim spaces, got %q", cal.DefaultTZ)
	}

	// Test empty
	cal.SetDefaultTimezone("")
	if cal.DefaultTZ != "" {
		t.Errorf("SetDefaultTimezone() with empty string should set empty, got %q", cal.DefaultTZ)
	}
}

// ========================================
// Test SetTimezone (both start and end)
// ========================================

func TestSetTimezone(t *testing.T) {
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))

	event.SetTimezone("Europe/Madrid")

	if event.StartTZ != "Europe/Madrid" {
		t.Errorf("SetTimezone() StartTZ = %q, want %q", event.StartTZ, "Europe/Madrid")
	}
	if event.EndTZ != "Europe/Madrid" {
		t.Errorf("SetTimezone() EndTZ = %q, want %q", event.EndTZ, "Europe/Madrid")
	}
}

// ========================================
// Test AddEvent with timezone inference
// ========================================

func TestAddEventInfersDefaultTimezone(t *testing.T) {
	cal := NewCalendar()
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))
	event.SetTimezone("Europe/Madrid")

	cal.AddEvent(event)

	if cal.DefaultTZ != "Europe/Madrid" {
		t.Errorf("AddEvent() should infer DefaultTZ = %q, got %q", "Europe/Madrid", cal.DefaultTZ)
	}

	// Second event with different TZ should not override
	event2 := NewEvent("Test2", time.Now(), time.Now().Add(1*time.Hour))
	event2.SetTimezone("America/New_York")
	cal.AddEvent(event2)

	if cal.DefaultTZ != "Europe/Madrid" {
		t.Errorf("AddEvent() should keep first DefaultTZ, got %q", cal.DefaultTZ)
	}
}

func TestAddEventDoesNotInferMismatchedTimezones(t *testing.T) {
	cal := NewCalendar()
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))
	event.StartTZ = "Europe/Madrid"
	event.EndTZ = "America/New_York" // Different

	cal.AddEvent(event)

	if cal.DefaultTZ != "" {
		t.Errorf("AddEvent() should not infer DefaultTZ for mismatched timezones, got %q", cal.DefaultTZ)
	}
}

// ========================================
// Test VALARM with absolute time trigger
// ========================================

func TestEventWithAbsoluteTimeAlarm(t *testing.T) {
	cal := NewCalendar()
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))

	triggerTime := time.Date(2025, 11, 15, 9, 0, 0, 0, time.UTC)
	alarm := Alarm{
		Action:            "DISPLAY",
		Description:       "Wake up call",
		TriggerIsRelative: false,
		TriggerTime:       triggerTime,
	}
	event.Alarms = append(event.Alarms, alarm)
	cal.AddEvent(event)

	ics := cal.ToICS()

	if !strings.Contains(ics, "TRIGGER;VALUE=DATE-TIME:20251115T090000Z") {
		t.Error("ICS should contain absolute TRIGGER with DATE-TIME value")
	}
	if !strings.Contains(ics, "DESCRIPTION:Wake up call") {
		t.Error("ICS should contain alarm description")
	}
}

// ========================================
// Test VALARM with repeat
// ========================================

func TestEventWithRepeatingAlarm(t *testing.T) {
	cal := NewCalendar()
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))

	alarm := Alarm{
		Action:            "DISPLAY",
		Description:       "Reminder",
		TriggerIsRelative: true,
		TriggerDuration:   -15 * time.Minute,
		Repeat:            3,
		RepeatDuration:    5 * time.Minute,
	}
	event.Alarms = append(event.Alarms, alarm)
	cal.AddEvent(event)

	ics := cal.ToICS()

	requiredFields := []string{
		"REPEAT:3",
		"DURATION:PT5M",
	}

	for _, field := range requiredFields {
		if !strings.Contains(ics, field) {
			t.Errorf("Repeating alarm missing field: %s", field)
		}
	}
}

// ========================================
// Test VALARM with empty action (default DISPLAY)
// ========================================

func TestEventWithAlarmDefaultAction(t *testing.T) {
	cal := NewCalendar()
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))

	alarm := Alarm{
		Action:            "", // Should default to DISPLAY
		TriggerIsRelative: true,
		TriggerDuration:   -10 * time.Minute,
	}
	event.Alarms = append(event.Alarms, alarm)
	cal.AddEvent(event)

	ics := cal.ToICS()

	if !strings.Contains(ics, "ACTION:DISPLAY") {
		t.Error("Empty action should default to DISPLAY")
	}
	if !strings.Contains(ics, "DESCRIPTION:Reminder") {
		t.Error("DISPLAY alarm with no description should default to 'Reminder'")
	}
}

// ========================================
// Test VALARM with EMAIL action
// ========================================

func TestEventWithEmailAlarm(t *testing.T) {
	cal := NewCalendar()
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))

	alarm := Alarm{
		Action:            "email", // Should be uppercase
		Summary:           "Meeting Reminder",
		Description:       "Don't forget the meeting",
		TriggerIsRelative: true,
		TriggerDuration:   -30 * time.Minute,
	}
	event.Alarms = append(event.Alarms, alarm)
	cal.AddEvent(event)

	ics := cal.ToICS()

	if !strings.Contains(ics, "ACTION:EMAIL") {
		t.Error("Email action should be uppercase")
	}
	if !strings.Contains(ics, "SUMMARY:Meeting Reminder") {
		t.Error("Email alarm should include summary")
	}
}

// ========================================
// Test VALARM with SUMMARY field
// ========================================

func TestEventWithAlarmSummary(t *testing.T) {
	cal := NewCalendar()
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))

	alarm := Alarm{
		Action:            "DISPLAY",
		Summary:           "Custom Summary",
		Description:       "Custom Description",
		TriggerIsRelative: true,
		TriggerDuration:   -5 * time.Minute,
	}
	event.Alarms = append(event.Alarms, alarm)
	cal.AddEvent(event)

	ics := cal.ToICS()

	if !strings.Contains(ics, "SUMMARY:Custom Summary") {
		t.Error("Alarm should include custom summary")
	}
}

// ========================================
// Test VALARM - multiple alarms
// ========================================

func TestEventWithMultipleAlarms(t *testing.T) {
	cal := NewCalendar()
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))

	alarm1 := Alarm{
		Action:            "DISPLAY",
		Description:       "15 min before",
		TriggerIsRelative: true,
		TriggerDuration:   -15 * time.Minute,
	}
	alarm2 := Alarm{
		Action:            "DISPLAY",
		Description:       "5 min before",
		TriggerIsRelative: true,
		TriggerDuration:   -5 * time.Minute,
	}
	alarm3 := Alarm{
		Action:            "DISPLAY",
		Description:       "At start time",
		TriggerIsRelative: true,
		TriggerDuration:   0,
	}

	event.Alarms = append(event.Alarms, alarm1, alarm2, alarm3)
	cal.AddEvent(event)

	ics := cal.ToICS()

	alarmCount := strings.Count(ics, "BEGIN:VALARM")
	if alarmCount != 3 {
		t.Errorf("Should have 3 VALARM blocks, got %d", alarmCount)
	}
}

// ========================================
// Test EXDATE handling - all three formats
// ========================================

func TestEventWithExDatesAllDay(t *testing.T) {
	cal := NewCalendar()
	start := time.Date(2025, 11, 15, 0, 0, 0, 0, time.UTC)
	event := NewEvent("Recurring", start, start.Add(24*time.Hour))
	event.AllDay = true
	event.RRule = "FREQ=DAILY;COUNT=10"

	exDate1 := time.Date(2025, 11, 16, 0, 0, 0, 0, time.UTC)
	exDate2 := time.Date(2025, 11, 18, 0, 0, 0, 0, time.UTC)
	event.ExDates = []time.Time{exDate1, exDate2}

	cal.AddEvent(event)
	ics := cal.ToICS()

	if !strings.Contains(ics, "EXDATE;VALUE=DATE:20251116,20251118") {
		t.Error("All-day event EXDATE should use VALUE=DATE format")
	}
}

func TestEventWithExDatesWithTimezone(t *testing.T) {
	cal := NewCalendar()
	loc, _ := time.LoadLocation("America/New_York")
	start := time.Date(2025, 11, 15, 10, 0, 0, 0, loc)
	event := NewEvent("Recurring", start, start.Add(1*time.Hour))
	event.SetTimezone("America/New_York")
	event.RRule = "FREQ=WEEKLY;BYDAY=MO"

	exDate := time.Date(2025, 11, 22, 10, 0, 0, 0, loc)
	event.ExDates = []time.Time{exDate}

	cal.AddEvent(event)
	ics := cal.ToICS()

	if !strings.Contains(ics, "EXDATE;TZID=America/New_York:") {
		t.Error("Event with timezone should have EXDATE with TZID")
	}
	if !strings.Contains(ics, "20251122T100000") {
		t.Error("EXDATE should contain correct date in local format")
	}
}

func TestEventWithExDatesUTC(t *testing.T) {
	cal := NewCalendar()
	start := time.Date(2025, 11, 15, 14, 0, 0, 0, time.UTC)
	event := NewEvent("Recurring", start, start.Add(1*time.Hour))
	event.RRule = "FREQ=DAILY"

	exDate := time.Date(2025, 11, 17, 14, 0, 0, 0, time.UTC)
	event.ExDates = []time.Time{exDate}

	cal.AddEvent(event)
	ics := cal.ToICS()

	if !strings.Contains(ics, "EXDATE:20251117T140000Z") {
		t.Error("UTC event EXDATE should use UTC format with Z suffix")
	}
}

// ========================================
// Test multi-timezone events (flights)
// ========================================

func TestEventWithDifferentStartEndTimezones(t *testing.T) {
	cal := NewCalendar()

	nyLoc, _ := time.LoadLocation("America/New_York")
	londonLoc, _ := time.LoadLocation("Europe/London")

	start := time.Date(2025, 11, 15, 18, 0, 0, 0, nyLoc)
	end := time.Date(2025, 11, 16, 6, 0, 0, 0, londonLoc)

	event := NewEvent("Flight NYC to London", start, end)
	event.SetStartTimezone("America/New_York")
	event.SetEndTimezone("Europe/London")

	cal.AddEvent(event)
	ics := cal.ToICS()

	if !strings.Contains(ics, "DTSTART;TZID=America/New_York:20251115T180000") {
		t.Error("Flight should have start time in departure timezone")
	}
	if !strings.Contains(ics, "DTEND;TZID=Europe/London:20251116T060000") {
		t.Error("Flight should have end time in arrival timezone")
	}
}

// ========================================
// Test SEQUENCE field
// ========================================

func TestEventWithSequence(t *testing.T) {
	cal := NewCalendar()
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))
	event.Sequence = 3
	cal.AddEvent(event)

	ics := cal.ToICS()

	if !strings.Contains(ics, "SEQUENCE:3") {
		t.Error("ICS should contain SEQUENCE field")
	}
}

func TestEventWithZeroSequence(t *testing.T) {
	cal := NewCalendar()
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))
	event.Sequence = 0
	cal.AddEvent(event)

	ics := cal.ToICS()

	if strings.Contains(ics, "SEQUENCE:") {
		t.Error("ICS should not contain SEQUENCE field when value is 0")
	}
}

// ========================================
// Test STATUS field variations
// ========================================

func TestEventStatusVariations(t *testing.T) {
	tests := []struct {
		name           string
		status         string
		expectedInICS  string
	}{
		{"confirmed", "CONFIRMED", "STATUS:CONFIRMED"},
		{"tentative", "TENTATIVE", "STATUS:TENTATIVE"},
		{"cancelled", "CANCELLED", "STATUS:CANCELLED"},
		{"empty defaults to confirmed", "", "STATUS:CONFIRMED"},
		{"whitespace defaults to confirmed", "   ", "STATUS:CONFIRMED"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cal := NewCalendar()
			event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))
			event.Status = tt.status
			cal.AddEvent(event)

			ics := cal.ToICS()

			if !strings.Contains(ics, tt.expectedInICS) {
				t.Errorf("Expected ICS to contain %q, got:\n%s", tt.expectedInICS, ics)
			}
		})
	}
}

// ========================================
// Test DTSTAMP handling when Created is zero
// ========================================

func TestEventDTSTAMPWithZeroCreated(t *testing.T) {
	cal := NewCalendar()
	event := &Event{
		UID:       "test-uid@tempus",
		Summary:   "Test",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(1 * time.Hour),
		Created:   time.Time{}, // Zero value
		LastMod:   time.Now(),
		Status:    "CONFIRMED",
	}
	cal.AddEvent(event)

	ics := cal.ToICS()

	// Should still have DTSTAMP (defaults to "now")
	if !strings.Contains(ics, "DTSTAMP:") {
		t.Error("ICS should contain DTSTAMP even when Created is zero")
	}
}

// ========================================
// Test attendee with empty/whitespace
// ========================================

func TestEventWithEmptyAttendees(t *testing.T) {
	cal := NewCalendar()
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))
	event.Attendees = []string{"alice@example.com", "  ", "", "bob@example.com"}
	cal.AddEvent(event)

	ics := cal.ToICS()

	// Should only have 2 attendees (empty and whitespace are skipped)
	attendeeCount := strings.Count(ics, "ATTENDEE:")
	if attendeeCount != 2 {
		t.Errorf("Expected 2 attendees (empty/whitespace skipped), got %d", attendeeCount)
	}
}

// ========================================
// Test description and location with newlines
// ========================================

func TestEventDescriptionWithUserNewlines(t *testing.T) {
	cal := NewCalendar()
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))
	event.Description = "Line 1\\nLine 2\\nLine 3"
	cal.AddEvent(event)

	ics := cal.ToICS()

	// Should convert \\n to actual newlines, then escape to \n
	if !strings.Contains(ics, "DESCRIPTION:Line 1\\nLine 2\\nLine 3") {
		t.Error("Description should normalize user newlines")
	}
}

func TestEventLocationWithUserNewlines(t *testing.T) {
	cal := NewCalendar()
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))
	event.Location = "Building A\\nFloor 2\\nRoom 101"
	cal.AddEvent(event)

	ics := cal.ToICS()

	if !strings.Contains(ics, "LOCATION:Building A\\nFloor 2\\nRoom 101") {
		t.Error("Location should normalize user newlines")
	}
}

// ========================================
// Test Calendar.ToICS with empty Method
// ========================================

func TestCalendarWithEmptyMethod(t *testing.T) {
	cal := NewCalendar()
	cal.Method = ""
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))
	cal.AddEvent(event)

	ics := cal.ToICS()

	// Empty method should not be included
	if strings.Contains(ics, "METHOD:") {
		t.Error("Empty METHOD should not be included in ICS")
	}
}

func TestCalendarWithWhitespaceMethod(t *testing.T) {
	cal := NewCalendar()
	cal.Method = "   "
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))
	cal.AddEvent(event)

	ics := cal.ToICS()

	// Whitespace method should not be included
	if strings.Contains(ics, "METHOD:") {
		t.Error("Whitespace METHOD should not be included in ICS")
	}
}

// ========================================
// Test Calendar.ToICS with empty Name
// ========================================

func TestCalendarWithEmptyName(t *testing.T) {
	cal := NewCalendar()
	cal.Name = ""
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))
	cal.AddEvent(event)

	ics := cal.ToICS()

	if strings.Contains(ics, "X-WR-CALNAME:") {
		t.Error("Empty Name should not be included in ICS")
	}
}

// ========================================
// Test Calendar.ToICS with empty DefaultTZ
// ========================================

func TestCalendarWithEmptyDefaultTZ(t *testing.T) {
	cal := NewCalendar()
	cal.DefaultTZ = ""
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))
	cal.AddEvent(event)

	ics := cal.ToICS()

	if strings.Contains(ics, "X-WR-TIMEZONE:") {
		t.Error("Empty DefaultTZ should not be included in ICS")
	}
}

// ========================================
// Test Calendar with IncludeVTZ (VTIMEZONE blocks)
// ========================================

func TestCalendarWithIncludeVTZ(t *testing.T) {
	cal := NewCalendar()
	cal.IncludeVTZ = true
	cal.DefaultTZ = "Europe/Madrid"

	madridLoc, _ := time.LoadLocation("Europe/Madrid")
	start := time.Date(2025, 11, 15, 10, 0, 0, 0, madridLoc)
	event := NewEvent("Test", start, start.Add(1*time.Hour))
	event.SetTimezone("Europe/Madrid")
	cal.AddEvent(event)

	ics := cal.ToICS()

	// Should include VTIMEZONE block
	if !strings.Contains(ics, "BEGIN:VTIMEZONE") {
		t.Error("ICS with IncludeVTZ should contain VTIMEZONE block")
	}
	if !strings.Contains(ics, "TZID:Europe/Madrid") {
		t.Error("VTIMEZONE should contain TZID:Europe/Madrid")
	}
	if !strings.Contains(ics, "END:VTIMEZONE") {
		t.Error("ICS should have closing END:VTIMEZONE")
	}
}

func TestCalendarWithIncludeVTZMultipleTimezones(t *testing.T) {
	cal := NewCalendar()
	cal.IncludeVTZ = true

	dublinLoc, _ := time.LoadLocation("Europe/Dublin")
	londonLoc, _ := time.LoadLocation("Europe/London")

	start1 := time.Date(2025, 11, 15, 10, 0, 0, 0, dublinLoc)
	event1 := NewEvent("Dublin Event", start1, start1.Add(1*time.Hour))
	event1.SetTimezone("Europe/Dublin")

	start2 := time.Date(2025, 11, 16, 14, 0, 0, 0, londonLoc)
	event2 := NewEvent("London Event", start2, start2.Add(1*time.Hour))
	event2.SetTimezone("Europe/London")

	cal.AddEvent(event1)
	cal.AddEvent(event2)

	ics := cal.ToICS()

	// Should include both VTIMEZONE blocks
	vtzCount := strings.Count(ics, "BEGIN:VTIMEZONE")
	if vtzCount != 2 {
		t.Errorf("Expected 2 VTIMEZONE blocks, got %d", vtzCount)
	}
	if !strings.Contains(ics, "TZID:Europe/Dublin") {
		t.Error("Should contain Dublin timezone")
	}
	if !strings.Contains(ics, "TZID:Europe/London") {
		t.Error("Should contain London timezone")
	}
}

func TestCalendarWithIncludeVTZAllDayEvents(t *testing.T) {
	cal := NewCalendar()
	cal.IncludeVTZ = true

	start := time.Date(2025, 11, 15, 0, 0, 0, 0, time.UTC)
	event := NewEvent("All Day", start, start.Add(24*time.Hour))
	event.AllDay = true

	cal.AddEvent(event)
	ics := cal.ToICS()

	// All-day events don't use timezones, so no VTIMEZONE should be included
	if strings.Contains(ics, "BEGIN:VTIMEZONE") {
		t.Error("All-day events should not trigger VTIMEZONE inclusion")
	}
}

func TestCalendarWithIncludeVTZUnknownTimezone(t *testing.T) {
	cal := NewCalendar()
	cal.IncludeVTZ = true

	start := time.Now()
	event := NewEvent("Test", start, start.Add(1*time.Hour))
	event.SetTimezone("America/Los_Angeles") // Not in knownVTZ

	cal.AddEvent(event)
	ics := cal.ToICS()

	// Unknown timezone should not produce VTIMEZONE block
	if strings.Contains(ics, "TZID:America/Los_Angeles") {
		t.Error("Unknown timezone should not generate VTIMEZONE block")
	}
}

// ========================================
// Test knownVTZ function
// ========================================

func TestKnownVTZ(t *testing.T) {
	tests := []struct {
		tzid     string
		hasVTZ   bool
	}{
		{"Europe/Madrid", true},
		{"Europe/Dublin", true},
		{"Europe/London", true},
		{"America/Sao_Paulo", true},
		{"Atlantic/Canary", true},
		{"America/New_York", false}, // Not in knownVTZ
		{"Invalid/Zone", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.tzid, func(t *testing.T) {
			result := knownVTZ(tt.tzid)
			hasResult := result != ""
			if hasResult != tt.hasVTZ {
				t.Errorf("knownVTZ(%q) returned data=%v, want=%v", tt.tzid, hasResult, tt.hasVTZ)
			}
			if tt.hasVTZ && !strings.Contains(result, "BEGIN:VTIMEZONE") {
				t.Errorf("knownVTZ(%q) should return valid VTIMEZONE block", tt.tzid)
			}
		})
	}
}

// ========================================
// Test uniqueTZIDs function
// ========================================

func TestUniqueTZIDs(t *testing.T) {
	events := []Event{
		{
			StartTZ: "Europe/Madrid",
			EndTZ:   "Europe/Madrid",
			AllDay:  false,
		},
		{
			StartTZ: "America/New_York",
			EndTZ:   "Europe/London",
			AllDay:  false,
		},
		{
			StartTZ: "Europe/Madrid", // Duplicate
			EndTZ:   "Europe/Madrid",
			AllDay:  false,
		},
		{
			StartTZ: "Europe/Madrid",
			EndTZ:   "Europe/Madrid",
			AllDay:  true, // All-day should be ignored
		},
	}

	result := uniqueTZIDs(events)

	// Should have 3 unique TZIDs (Madrid, New_York, London)
	if len(result) != 3 {
		t.Errorf("uniqueTZIDs() returned %d TZIDs, want 3", len(result))
	}

	// Check that all expected TZIDs are present
	tzMap := make(map[string]bool)
	for _, tz := range result {
		tzMap[tz] = true
	}

	expected := []string{"Europe/Madrid", "America/New_York", "Europe/London"}
	for _, tz := range expected {
		if !tzMap[tz] {
			t.Errorf("uniqueTZIDs() missing expected TZID: %s", tz)
		}
	}
}

func TestUniqueTZIDsWithEmptyTimezones(t *testing.T) {
	events := []Event{
		{
			StartTZ: "",
			EndTZ:   "",
			AllDay:  false,
		},
		{
			StartTZ: "   ",
			EndTZ:   "  ",
			AllDay:  false,
		},
	}

	result := uniqueTZIDs(events)

	if len(result) != 0 {
		t.Errorf("uniqueTZIDs() with empty timezones should return empty slice, got %d", len(result))
	}
}

// ========================================
// Test edge cases for all-day events
// ========================================

func TestAllDayEventDifferentFormats(t *testing.T) {
	cal := NewCalendar()

	// All-day event spanning multiple days
	start := time.Date(2025, 11, 15, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 11, 17, 0, 0, 0, 0, time.UTC) // 2 days

	event := NewEvent("Multi-day event", start, end)
	event.AllDay = true
	cal.AddEvent(event)

	ics := cal.ToICS()

	if !strings.Contains(ics, "DTSTART;VALUE=DATE:20251115") {
		t.Error("Multi-day event should have DATE format for start")
	}
	if !strings.Contains(ics, "DTEND;VALUE=DATE:20251117") {
		t.Error("Multi-day event should have DATE format for end")
	}
	// Should NOT have time component
	if strings.Contains(ics, "20251115T") {
		t.Error("All-day event should not have time component")
	}
}

// ========================================
// Test Priority = 0 (should not be included)
// ========================================

func TestEventWithZeroPriority(t *testing.T) {
	cal := NewCalendar()
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))
	event.Priority = 0
	cal.AddEvent(event)

	ics := cal.ToICS()

	if strings.Contains(ics, "PRIORITY:") {
		t.Error("Priority of 0 should not be included in ICS")
	}
}

// ========================================
// Test empty categories (should still be included)
// ========================================

func TestEventWithEmptyCategories(t *testing.T) {
	cal := NewCalendar()
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))
	event.Categories = []string{} // Empty
	cal.AddEvent(event)

	ics := cal.ToICS()

	if strings.Contains(ics, "CATEGORIES:") {
		t.Error("Empty categories should not be included in ICS")
	}
}

// ========================================
// Test empty RRULE (should not be included)
// ========================================

func TestEventWithEmptyRRule(t *testing.T) {
	cal := NewCalendar()
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))
	event.RRule = ""
	cal.AddEvent(event)

	ics := cal.ToICS()

	if strings.Contains(ics, "RRULE:") {
		t.Error("Empty RRULE should not be included in ICS")
	}
}

func TestEventWithWhitespaceRRule(t *testing.T) {
	cal := NewCalendar()
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))
	event.RRule = "   "
	cal.AddEvent(event)

	ics := cal.ToICS()

	if strings.Contains(ics, "RRULE:") {
		t.Error("Whitespace RRULE should not be included in ICS")
	}
}

// ========================================
// Test empty ExDates (should not be included)
// ========================================

func TestEventWithEmptyExDates(t *testing.T) {
	cal := NewCalendar()
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))
	event.ExDates = []time.Time{} // Empty
	cal.AddEvent(event)

	ics := cal.ToICS()

	if strings.Contains(ics, "EXDATE") {
		t.Error("Empty ExDates should not be included in ICS")
	}
}

// ========================================
// Test empty Summary (should not crash)
// ========================================

func TestEventWithEmptySummary(t *testing.T) {
	cal := NewCalendar()
	event := NewEvent("", time.Now(), time.Now().Add(1*time.Hour))
	cal.AddEvent(event)

	ics := cal.ToICS()

	// Should not contain SUMMARY line
	lines := strings.Split(ics, "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "SUMMARY:") &&
		   !strings.Contains(line, "SUMMARY: ") {
			t.Error("Empty summary should not produce SUMMARY: line")
		}
	}
}

func TestEventWithWhitespaceSummary(t *testing.T) {
	cal := NewCalendar()
	event := NewEvent("   ", time.Now(), time.Now().Add(1*time.Hour))
	cal.AddEvent(event)

	ics := cal.ToICS()

	// Should not contain SUMMARY line
	if strings.Contains(ics, "SUMMARY:   ") {
		t.Error("Whitespace summary should not be included")
	}
}

// ========================================
// Test empty Description (should not be included)
// ========================================

func TestEventWithEmptyDescription(t *testing.T) {
	cal := NewCalendar()
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))
	event.Description = ""
	cal.AddEvent(event)

	ics := cal.ToICS()

	if strings.Contains(ics, "DESCRIPTION:") {
		t.Error("Empty description should not be included in ICS")
	}
}

// ========================================
// Test empty Location (should not be included)
// ========================================

func TestEventWithEmptyLocation(t *testing.T) {
	cal := NewCalendar()
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))
	event.Location = ""
	cal.AddEvent(event)

	ics := cal.ToICS()

	if strings.Contains(ics, "LOCATION:") {
		t.Error("Empty location should not be included in ICS")
	}
}

// ========================================
// Test CRLF line endings
// ========================================

func TestICSUsesProperCRLFLineEndings(t *testing.T) {
	cal := NewCalendar()
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))
	cal.AddEvent(event)

	ics := cal.ToICS()

	// Should use CRLF (\r\n) not just LF (\n)
	if !strings.Contains(ics, "\r\n") {
		t.Error("ICS should use CRLF line endings")
	}

	// Every \n should be preceded by \r
	lines := strings.Split(ics, "\n")
	for i, line := range lines {
		if i < len(lines)-1 && !strings.HasSuffix(line, "\r") {
			t.Errorf("Line %d does not end with \\r before \\n", i)
			break
		}
	}
}

// ========================================
// Test generateUID uniqueness
// ========================================

func TestGenerateUIDUniqueness(t *testing.T) {
	uids := make(map[string]bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		uid := generateUID()
		if uids[uid] {
			t.Errorf("Duplicate UID generated: %s", uid)
		}
		uids[uid] = true
		if !strings.HasSuffix(uid, "@tempus") {
			t.Errorf("UID should end with @tempus, got: %s", uid)
		}
	}

	if len(uids) != iterations {
		t.Errorf("Expected %d unique UIDs, got %d", iterations, len(uids))
	}
}

// ========================================
// Test escaping in calendar name
// ========================================

func TestCalendarNameWithSpecialCharacters(t *testing.T) {
	cal := NewCalendar()
	cal.Name = "My Calendar, with; special\\chars"
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))
	cal.AddEvent(event)

	ics := cal.ToICS()

	// Calendar name should be escaped
	if !strings.Contains(ics, "X-WR-CALNAME:My Calendar\\, with\\; special\\\\chars") {
		t.Error("Calendar name should escape special characters")
	}
}

// ========================================
// Test complex alarm scenarios
// ========================================

func TestAlarmWithZeroRepeat(t *testing.T) {
	cal := NewCalendar()
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))

	alarm := Alarm{
		Action:            "DISPLAY",
		Description:       "Test",
		TriggerIsRelative: true,
		TriggerDuration:   -15 * time.Minute,
		Repeat:            0, // Should not include REPEAT
		RepeatDuration:    5 * time.Minute,
	}
	event.Alarms = append(event.Alarms, alarm)
	cal.AddEvent(event)

	ics := cal.ToICS()

	if strings.Contains(ics, "REPEAT:") {
		t.Error("Alarm with Repeat=0 should not include REPEAT field")
	}
}

func TestAlarmWithRepeatButZeroDuration(t *testing.T) {
	cal := NewCalendar()
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))

	alarm := Alarm{
		Action:            "DISPLAY",
		Description:       "Test",
		TriggerIsRelative: true,
		TriggerDuration:   -15 * time.Minute,
		Repeat:            3,
		RepeatDuration:    0, // Should not include REPEAT/DURATION
	}
	event.Alarms = append(event.Alarms, alarm)
	cal.AddEvent(event)

	ics := cal.ToICS()

	if strings.Contains(ics, "REPEAT:") {
		t.Error("Alarm with RepeatDuration=0 should not include REPEAT field")
	}
}

// ========================================
// Test UTC vs timezone handling in DTSTART/DTEND
// ========================================

func TestEventUTCWithoutTimezone(t *testing.T) {
	cal := NewCalendar()
	start := time.Date(2025, 11, 15, 14, 30, 0, 0, time.UTC)
	end := start.Add(1 * time.Hour)

	event := NewEvent("UTC Event", start, end)
	// Don't set timezone - should use UTC with Z suffix
	cal.AddEvent(event)

	ics := cal.ToICS()

	if !strings.Contains(ics, "DTSTART:20251115T143000Z") {
		t.Error("UTC event without timezone should use Z suffix")
	}
	if !strings.Contains(ics, "DTEND:20251115T153000Z") {
		t.Error("UTC event end time should use Z suffix")
	}
}

func TestEventWithTimezoneUsesLocalFormat(t *testing.T) {
	cal := NewCalendar()
	loc, _ := time.LoadLocation("Europe/Madrid")
	start := time.Date(2025, 11, 15, 14, 30, 0, 0, loc)
	end := start.Add(1 * time.Hour)

	event := NewEvent("Madrid Event", start, end)
	event.SetTimezone("Europe/Madrid")
	cal.AddEvent(event)

	ics := cal.ToICS()

	// Should use local format (no Z suffix) with TZID
	if !strings.Contains(ics, "DTSTART;TZID=Europe/Madrid:20251115T143000") {
		t.Error("Event with timezone should use local format with TZID")
	}
	// Should NOT have Z suffix
	if strings.Contains(ics, "20251115T143000Z") {
		t.Error("Event with TZID should not use Z suffix")
	}
}

// ========================================
// Test VALARM trigger edge cases
// ========================================

func TestAlarmTriggerAfterEvent(t *testing.T) {
	cal := NewCalendar()
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))

	alarm := Alarm{
		Action:            "DISPLAY",
		Description:       "After event starts",
		TriggerIsRelative: true,
		TriggerDuration:   5 * time.Minute, // Positive = after
	}
	event.Alarms = append(event.Alarms, alarm)
	cal.AddEvent(event)

	ics := cal.ToICS()

	if !strings.Contains(ics, "TRIGGER:PT5M") {
		t.Error("Positive trigger duration should not have minus sign")
	}
}

func TestAlarmTriggerZeroDuration(t *testing.T) {
	cal := NewCalendar()
	event := NewEvent("Test", time.Now(), time.Now().Add(1*time.Hour))

	alarm := Alarm{
		Action:            "DISPLAY",
		Description:       "At event start",
		TriggerIsRelative: true,
		TriggerDuration:   0,
	}
	event.Alarms = append(event.Alarms, alarm)
	cal.AddEvent(event)

	ics := cal.ToICS()

	if !strings.Contains(ics, "TRIGGER:PT0S") {
		t.Error("Zero trigger duration should be PT0S")
	}
}

// ========================================
// Tests for alarms_parser.go
// ========================================

// ========================================
// Test ParseHumanDuration function
// ========================================

func TestAlarmsParser_ParseHumanDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
		wantErr  bool
	}{
		// Valid HH:MM format
		{"1:30 format", "1:30", 1*time.Hour + 30*time.Minute, false},
		{"10:45 format", "10:45", 10*time.Hour + 45*time.Minute, false},
		{"0:15 format", "0:15", 15 * time.Minute, false},
		{"23:59 format", "23:59", 23*time.Hour + 59*time.Minute, false},
		{"with spaces", "  2:30  ", 2*time.Hour + 30*time.Minute, false},
		{"single digit hour", "2:05", 2*time.Hour + 5*time.Minute, false},

		// Valid h/m format
		{"1h format", "1h", 1 * time.Hour, false},
		{"30m format", "30m", 30 * time.Minute, false},
		{"1h30m format", "1h30m", 1*time.Hour + 30*time.Minute, false},
		{"2h15m format", "2h15m", 2*time.Hour + 15*time.Minute, false},
		{"just hours", "5h", 5 * time.Hour, false},
		{"just minutes", "45m", 45 * time.Minute, false},
		{"with spaces", " 1h 30m ", 1*time.Hour + 30*time.Minute, false},

		// Valid plain minutes format
		{"10 minutes", "10", 10 * time.Minute, false},
		{"90 minutes", "90", 90 * time.Minute, false},
		{"1 minute", "1", 1 * time.Minute, false},
		{"with spaces", "  60  ", 60 * time.Minute, false},

		// Invalid cases
		{"empty string", "", 0, true},
		{"only spaces", "   ", 0, true},
		{"zero minutes", "0", 0, true},
		{"negative minutes", "-10", 0, true},
		{"0h0m", "0h0m", 0, true},
		{"invalid format", "abc", 0, true},
		{"invalid time", "25:00", 25 * time.Hour, false}, // atoiSafe doesn't validate hour range
		{"invalid minutes", "1:60", 0, true}, // Regex validates minutes must be 0-59
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

// ========================================
// Test SplitAlarmInput function
// ========================================

func TestAlarmsParser_SplitAlarmInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: nil,
		},
		{
			name:     "single value",
			input:    "15m",
			expected: []string{"15m"},
		},
		{
			name:     "comma separated",
			input:    "15m,30m,1h",
			expected: []string{"15m", "30m", "1h"},
		},
		{
			name:     "semicolon separated",
			input:    "15m;30m;1h",
			expected: []string{"15m", "30m", "1h"},
		},
		{
			name:     "pipe separated",
			input:    "15m|30m|1h",
			expected: []string{"15m", "30m", "1h"},
		},
		{
			name:     "double pipe separated",
			input:    "15m||30m||1h",
			expected: []string{"15m", "30m", "1h"},
		},
		{
			name:     "newline separated",
			input:    "15m\n30m\n1h",
			expected: []string{"15m", "30m", "1h"},
		},
		{
			name:     "CRLF separated",
			input:    "15m\r\n30m\r\n1h",
			expected: []string{"15m", "30m", "1h"},
		},
		{
			name:     "CR separated",
			input:    "15m\r30m\r1h",
			expected: []string{"15m", "30m", "1h"},
		},
		{
			name:     "mixed separators",
			input:    "15m,30m;1h|2h",
			expected: []string{"15m", "30m", "1h", "2h"},
		},
		{
			name:     "with key-value pair",
			input:    "15m,trigger=30m",
			expected: []string{"15m,trigger=30m"}, // Key-value pairs are not split on comma
		},
		{
			name:     "multiple key-value pairs",
			input:    "trigger=15m,action=DISPLAY",
			expected: []string{"trigger=15m,action=DISPLAY"},
		},
		{
			name:     "key-value with double pipe",
			input:    "trigger=15m||trigger=30m",
			expected: []string{"trigger=15m", "trigger=30m"},
		},
		{
			name:     "with extra whitespace",
			input:    "  15m  ,  30m  ,  1h  ",
			expected: []string{"15m", "30m", "1h"},
		},
		{
			name:     "empty values ignored",
			input:    "15m,,30m",
			expected: []string{"15m", "30m"},
		},
		{
			name:     "blank lines ignored",
			input:    "15m\n\n30m\n",
			expected: []string{"15m", "30m"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SplitAlarmInput(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("SplitAlarmInput(%q) returned %d items, want %d", tt.input, len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("SplitAlarmInput(%q)[%d] = %q, want %q", tt.input, i, result[i], tt.expected[i])
				}
			}
		})
	}
}

// ========================================
// Test ParseAlarmsFromString function
// ========================================

func TestAlarmsParser_ParseAlarmsFromString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		defaultTZ string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "empty string",
			input:     "",
			defaultTZ: "",
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "single alarm",
			input:     "15m",
			defaultTZ: "",
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "multiple alarms comma separated",
			input:     "15m,30m,1h",
			defaultTZ: "",
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "multiple alarms newline separated",
			input:     "15m\n30m\n1h",
			defaultTZ: "",
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "with absolute time",
			input:     "2025-11-15 10:00:00",
			defaultTZ: "UTC",
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "invalid alarm",
			input:     "invalid",
			defaultTZ: "",
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseAlarmsFromString(tt.input, tt.defaultTZ)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAlarmsFromString(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(result) != tt.wantCount {
				t.Errorf("ParseAlarmsFromString(%q) returned %d alarms, want %d", tt.input, len(result), tt.wantCount)
			}
		})
	}
}

// ========================================
// Test ParseAlarmSpecs function
// ========================================

func TestAlarmsParser_ParseAlarmSpecs(t *testing.T) {
	tests := []struct {
		name      string
		specs     []string
		defaultTZ string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "empty specs",
			specs:     []string{},
			defaultTZ: "",
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "single spec",
			specs:     []string{"15m"},
			defaultTZ: "",
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "multiple specs",
			specs:     []string{"15m", "30m", "1h"},
			defaultTZ: "",
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "skip empty strings",
			specs:     []string{"15m", "", "  ", "30m"},
			defaultTZ: "",
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "invalid spec",
			specs:     []string{"15m", "invalid"},
			defaultTZ: "",
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseAlarmSpecs(tt.specs, tt.defaultTZ)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAlarmSpecs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(result) != tt.wantCount {
				t.Errorf("ParseAlarmSpecs() returned %d alarms, want %d", len(result), tt.wantCount)
			}
		})
	}
}

// ========================================
// Test parseSimpleAlarmSpec function
// ========================================

func TestAlarmsParser_parseSimpleAlarmSpec(t *testing.T) {
	tests := []struct {
		name       string
		spec       string
		defaultTZ  string
		wantAction string
		wantDesc   string
		wantRel    bool
		wantErr    bool
	}{
		{
			name:       "relative duration 15m",
			spec:       "15m",
			defaultTZ:  "",
			wantAction: "DISPLAY",
			wantDesc:   "Reminder",
			wantRel:    true,
			wantErr:    false,
		},
		{
			name:       "relative duration 1h",
			spec:       "1h",
			defaultTZ:  "",
			wantAction: "DISPLAY",
			wantDesc:   "Reminder",
			wantRel:    true,
			wantErr:    false,
		},
		{
			name:       "absolute time",
			spec:       "2025-11-15 10:00:00",
			defaultTZ:  "UTC",
			wantAction: "DISPLAY",
			wantDesc:   "Reminder",
			wantRel:    false,
			wantErr:    false,
		},
		{
			name:       "empty spec",
			spec:       "",
			defaultTZ:  "",
			wantAction: "",
			wantDesc:   "",
			wantRel:    false,
			wantErr:    true,
		},
		{
			name:       "invalid spec",
			spec:       "invalid",
			defaultTZ:  "",
			wantAction: "",
			wantDesc:   "",
			wantRel:    false,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseSimpleAlarmSpec(tt.spec, tt.defaultTZ)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSimpleAlarmSpec(%q) error = %v, wantErr %v", tt.spec, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if result.Action != tt.wantAction {
					t.Errorf("Action = %q, want %q", result.Action, tt.wantAction)
				}
				if result.Description != tt.wantDesc {
					t.Errorf("Description = %q, want %q", result.Description, tt.wantDesc)
				}
				if result.TriggerIsRelative != tt.wantRel {
					t.Errorf("TriggerIsRelative = %v, want %v", result.TriggerIsRelative, tt.wantRel)
				}
			}
		})
	}
}

// ========================================
// Test parseKeyValueAlarmSpec function
// ========================================

func TestAlarmsParser_parseKeyValueAlarmSpec(t *testing.T) {
	tests := []struct {
		name    string
		spec    string
		wantErr bool
		check   func(*testing.T, Alarm)
	}{
		{
			name:    "simple trigger",
			spec:    "trigger=15m",
			wantErr: false,
			check: func(t *testing.T, a Alarm) {
				if a.TriggerDuration != -15*time.Minute {
					t.Errorf("TriggerDuration = %v, want -15m", a.TriggerDuration)
				}
				if !a.TriggerIsRelative {
					t.Error("TriggerIsRelative should be true")
				}
			},
		},
		{
			name:    "with action",
			spec:    "trigger=15m,action=EMAIL",
			wantErr: false,
			check: func(t *testing.T, a Alarm) {
				if a.Action != "EMAIL" {
					t.Errorf("Action = %q, want EMAIL", a.Action)
				}
			},
		},
		{
			name:    "with description",
			spec:    "trigger=15m,description=Custom reminder",
			wantErr: false,
			check: func(t *testing.T, a Alarm) {
				if a.Description != "Custom reminder" {
					t.Errorf("Description = %q, want 'Custom reminder'", a.Description)
				}
			},
		},
		{
			name:    "with summary",
			spec:    "trigger=15m,summary=Meeting",
			wantErr: false,
			check: func(t *testing.T, a Alarm) {
				if a.Summary != "Meeting" {
					t.Errorf("Summary = %q, want 'Meeting'", a.Summary)
				}
			},
		},
		{
			name:    "direction=after",
			spec:    "trigger=15m,direction=after",
			wantErr: false,
			check: func(t *testing.T, a Alarm) {
				if a.TriggerDuration != 15*time.Minute {
					t.Errorf("TriggerDuration = %v, want 15m (positive)", a.TriggerDuration)
				}
			},
		},
		{
			name:    "direction=before",
			spec:    "trigger=15m,direction=before",
			wantErr: false,
			check: func(t *testing.T, a Alarm) {
				if a.TriggerDuration != -15*time.Minute {
					t.Errorf("TriggerDuration = %v, want -15m", a.TriggerDuration)
				}
			},
		},
		{
			name:    "kind=relative",
			spec:    "trigger=15m,kind=relative",
			wantErr: false,
			check: func(t *testing.T, a Alarm) {
				if !a.TriggerIsRelative {
					t.Error("TriggerIsRelative should be true")
				}
			},
		},
		{
			name:    "kind=before",
			spec:    "trigger=15m,kind=before",
			wantErr: false,
			check: func(t *testing.T, a Alarm) {
				if a.TriggerDuration != -15*time.Minute {
					t.Errorf("TriggerDuration = %v, want -15m", a.TriggerDuration)
				}
			},
		},
		{
			name:    "kind=after",
			spec:    "trigger=15m,kind=after",
			wantErr: false,
			check: func(t *testing.T, a Alarm) {
				if a.TriggerDuration != 15*time.Minute {
					t.Errorf("TriggerDuration = %v, want 15m", a.TriggerDuration)
				}
			},
		},
		{
			name:    "with repeat",
			spec:    "trigger=15m,repeat=3,repeat_duration=5m",
			wantErr: false,
			check: func(t *testing.T, a Alarm) {
				if a.Repeat != 3 {
					t.Errorf("Repeat = %d, want 3", a.Repeat)
				}
				if a.RepeatDuration != 5*time.Minute {
					t.Errorf("RepeatDuration = %v, want 5m", a.RepeatDuration)
				}
			},
		},
		{
			name:    "relative hint true",
			spec:    "trigger=15m,relative=yes",
			wantErr: false,
			check: func(t *testing.T, a Alarm) {
				if !a.TriggerIsRelative {
					t.Error("TriggerIsRelative should be true")
				}
			},
		},
		{
			name:    "missing trigger",
			spec:    "action=DISPLAY",
			wantErr: true,
			check:   nil,
		},
		{
			name:    "invalid segment",
			spec:    "trigger=15m,invalid",
			wantErr: true,
			check:   nil,
		},
		{
			name:    "invalid repeat count",
			spec:    "trigger=15m,repeat=invalid",
			wantErr: true,
			check:   nil,
		},
		{
			name:    "negative repeat count",
			spec:    "trigger=15m,repeat=-1",
			wantErr: true,
			check:   nil,
		},
		{
			name:    "invalid repeat duration",
			spec:    "trigger=15m,repeat=3,repeat_duration=invalid",
			wantErr: true,
			check:   nil,
		},
		{
			name:    "repeat without duration",
			spec:    "trigger=15m,repeat=3",
			wantErr: true,
			check:   nil,
		},
		{
			name:    "absolute trigger with kind=absolute",
			spec:    "trigger=2025-11-15T10:00:00Z,kind=absolute",
			wantErr: false,
			check: func(t *testing.T, a Alarm) {
				if a.TriggerIsRelative {
					t.Error("TriggerIsRelative should be false")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseKeyValueAlarmSpec(tt.spec, "UTC")
			if (err != nil) != tt.wantErr {
				t.Errorf("parseKeyValueAlarmSpec(%q) error = %v, wantErr %v", tt.spec, err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

// ========================================
// Test parseAlarmAbsolute function
// ========================================

func TestAlarmsParser_parseAlarmAbsolute(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		defaultTZ string
		wantErr   bool
	}{
		{
			name:      "RFC3339 format",
			input:     "2025-11-15T10:00:00Z",
			defaultTZ: "",
			wantErr:   false,
		},
		{
			name:      "RFC3339 with space",
			input:     "2025-11-15 10:00:00Z",
			defaultTZ: "",
			wantErr:   false,
		},
		{
			name:      "date time with seconds",
			input:     "2025-11-15 10:00:00",
			defaultTZ: "UTC",
			wantErr:   false,
		},
		{
			name:      "date time without seconds",
			input:     "2025-11-15 10:00",
			defaultTZ: "UTC",
			wantErr:   false,
		},
		{
			name:      "T separator with seconds",
			input:     "2025-11-15T10:00:00",
			defaultTZ: "UTC",
			wantErr:   false,
		},
		{
			name:      "T separator without seconds",
			input:     "2025-11-15T10:00",
			defaultTZ: "UTC",
			wantErr:   false,
		},
		{
			name:      "with timezone",
			input:     "2025-11-15 10:00:00",
			defaultTZ: "America/New_York",
			wantErr:   false,
		},
		{
			name:      "empty string",
			input:     "",
			defaultTZ: "",
			wantErr:   true,
		},
		{
			name:      "invalid format",
			input:     "invalid",
			defaultTZ: "",
			wantErr:   true,
		},
		{
			name:      "invalid date",
			input:     "2025-13-45 10:00:00",
			defaultTZ: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseAlarmAbsolute(tt.input, tt.defaultTZ)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseAlarmAbsolute(%q, %q) error = %v, wantErr %v", tt.input, tt.defaultTZ, err, tt.wantErr)
				return
			}
			if !tt.wantErr && result.IsZero() {
				t.Error("parseAlarmAbsolute() returned zero time")
			}
		})
	}
}

// ========================================
// Test parseRelativeAlarmDuration function
// ========================================

func TestAlarmsParser_parseRelativeAlarmDuration(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		defaultDirection int
		expected         time.Duration
		wantErr          bool
	}{
		{
			name:             "positive sign",
			input:            "+15m",
			defaultDirection: -1,
			expected:         15 * time.Minute,
			wantErr:          false,
		},
		{
			name:             "negative sign",
			input:            "-15m",
			defaultDirection: 1,
			expected:         -15 * time.Minute,
			wantErr:          false,
		},
		{
			name:             "no sign, default negative",
			input:            "15m",
			defaultDirection: -1,
			expected:         -15 * time.Minute,
			wantErr:          false,
		},
		{
			name:             "no sign, default positive",
			input:            "15m",
			defaultDirection: 1,
			expected:         15 * time.Minute,
			wantErr:          false,
		},
		{
			name:             "no sign, no default",
			input:            "15m",
			defaultDirection: 0,
			expected:         -15 * time.Minute, // Falls back to -1
			wantErr:          false,
		},
		{
			name:             "empty string",
			input:            "",
			defaultDirection: -1,
			expected:         0,
			wantErr:          true,
		},
		{
			name:             "invalid duration",
			input:            "invalid",
			defaultDirection: -1,
			expected:         0,
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseRelativeAlarmDuration(tt.input, tt.defaultDirection)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRelativeAlarmDuration(%q, %d) error = %v, wantErr %v", tt.input, tt.defaultDirection, err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("parseRelativeAlarmDuration(%q, %d) = %v, want %v", tt.input, tt.defaultDirection, result, tt.expected)
			}
		})
	}
}

// ========================================
// Test parseAlarmDurationValue function
// ========================================

func TestAlarmsParser_parseAlarmDurationValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
		wantErr  bool
	}{
		// Human duration formats
		{"human 15m", "15m", 15 * time.Minute, false},
		{"human 1h", "1h", 1 * time.Hour, false},
		{"human 1h30m", "1h30m", 1*time.Hour + 30*time.Minute, false},

		// Go duration formats
		{"go 15m", "15m", 15 * time.Minute, false},
		{"go 1h30m", "1h30m", 1*time.Hour + 30*time.Minute, false},
		{"go with seconds", "1h30m45s", 1*time.Hour + 30*time.Minute + 45*time.Second, false},

		// ICS duration formats
		{"ICS PT15M", "PT15M", 15 * time.Minute, false},
		{"ICS PT1H", "PT1H", 1 * time.Hour, false},
		{"ICS PT1H30M", "PT1H30M", 1*time.Hour + 30*time.Minute, false},
		{"ICS P1D", "P1D", 24 * time.Hour, false},
		{"ICS P1W", "P1W", 7 * 24 * time.Hour, false},
		{"ICS lowercase", "pt15m", 15 * time.Minute, false},

		// With plus sign
		{"with plus", "+15m", 15 * time.Minute, false},

		// Invalid cases
		{"empty", "", 0, true},
		{"with minus", "-15m", 0, true},
		{"invalid", "invalid", 0, true},
		{"negative go duration", "-1h", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseAlarmDurationValue(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseAlarmDurationValue(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("parseAlarmDurationValue(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// ========================================
// Test parseICSDuration function
// ========================================

func TestAlarmsParser_parseICSDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
		wantErr  bool
	}{
		// Valid formats
		{"PT15M", "PT15M", 15 * time.Minute, false},
		{"PT1H", "PT1H", 1 * time.Hour, false},
		{"PT1H30M", "PT1H30M", 1*time.Hour + 30*time.Minute, false},
		{"PT1H30M45S", "PT1H30M45S", 1*time.Hour + 30*time.Minute + 45*time.Second, false},
		{"P1D", "P1D", 24 * time.Hour, false},
		{"P1DT2H", "P1DT2H", 24*time.Hour + 2*time.Hour, false},
		{"P1W", "P1W", 7 * 24 * time.Hour, false},
		{"P1W2D", "P1W2D", 7*24*time.Hour + 2*24*time.Hour, false},
		{"P1W2DT3H4M5S", "P1W2DT3H4M5S", 7*24*time.Hour + 2*24*time.Hour + 3*time.Hour + 4*time.Minute + 5*time.Second, false},
		{"lowercase", "pt15m", 15 * time.Minute, false},
		{"with spaces", "  PT15M  ", 15 * time.Minute, false},

		// Invalid formats
		{"empty", "", 0, true},
		{"no P prefix", "T15M", 0, true},
		{"P0", "P0", 0, true},
		{"PT0M", "PT0M", 0, true},
		{"invalid", "INVALID", 0, true},
		{"with minus", "-PT15M", 0, true},
		{"with plus", "+PT15M", 15 * time.Minute, false},
		{"malformed", "P1X", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseICSDuration(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseICSDuration(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("parseICSDuration(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// ========================================
// Test parseBoolish function
// ========================================

func TestAlarmsParser_parseBoolish(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"1", "1", true},
		{"true", "true", true},
		{"TRUE", "TRUE", true},
		{"True", "True", true},
		{"yes", "yes", true},
		{"YES", "YES", true},
		{"y", "y", true},
		{"Y", "Y", true},
		{"on", "on", true},
		{"ON", "ON", true},
		{"with spaces", "  yes  ", true},

		{"0", "0", false},
		{"false", "false", false},
		{"no", "no", false},
		{"n", "n", false},
		{"off", "off", false},
		{"empty", "", false},
		{"invalid", "invalid", false},
		{"spaces", "   ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseBoolish(tt.input)
			if result != tt.expected {
				t.Errorf("parseBoolish(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// ========================================
// Test firstNonEmpty function
// ========================================

func TestAlarmsParser_firstNonEmpty(t *testing.T) {
	tests := []struct {
		name     string
		inputs   []string
		expected string
	}{
		{
			name:     "first non-empty",
			inputs:   []string{"hello", "world"},
			expected: "hello",
		},
		{
			name:     "skip empty",
			inputs:   []string{"", "world"},
			expected: "world",
		},
		{
			name:     "skip whitespace",
			inputs:   []string{"  ", "world"},
			expected: "world",
		},
		{
			name:     "all empty",
			inputs:   []string{"", "  ", ""},
			expected: "",
		},
		{
			name:     "no inputs",
			inputs:   []string{},
			expected: "",
		},
		{
			name:     "middle value",
			inputs:   []string{"", "", "middle", "last"},
			expected: "middle",
		},
		{
			name:     "with trimming",
			inputs:   []string{"", "  value  "},
			expected: "  value  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := firstNonEmpty(tt.inputs...)
			if result != tt.expected {
				t.Errorf("firstNonEmpty(%v) = %q, want %q", tt.inputs, result, tt.expected)
			}
		})
	}
}

// ========================================
// Test atoiSafe function
// ========================================

func TestAlarmsParser_atoiSafe(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"zero", "0", 0},
		{"single digit", "5", 5},
		{"two digits", "42", 42},
		{"three digits", "123", 123},
		{"large number", "999999", 999999},
		{"with spaces", "  42  ", 42},
		{"empty", "", 0},
		{"only spaces", "   ", 0},
		{"invalid chars", "12a34", 0},
		{"letters", "abc", 0},
		{"negative", "-5", 0},
		{"with plus", "+5", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := atoiSafe(tt.input)
			if result != tt.expected {
				t.Errorf("atoiSafe(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

// ========================================
// Integration tests for alarm parsing
// ========================================

func TestAlarmsParser_Integration_RelativeAlarms(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantDuration time.Duration
	}{
		{"15 minutes before", "15m", -15 * time.Minute},
		{"1 hour before", "1h", -1 * time.Hour},
		{"90 minutes before", "90", -90 * time.Minute},
		{"1:30 before", "1:30", -(1*time.Hour + 30*time.Minute)},
		{"explicit negative", "-30m", -30 * time.Minute},
		{"explicit positive", "+30m", 30 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alarms, err := ParseAlarmsFromString(tt.input, "UTC")
			if err != nil {
				t.Fatalf("ParseAlarmsFromString(%q) error: %v", tt.input, err)
			}
			if len(alarms) != 1 {
				t.Fatalf("Expected 1 alarm, got %d", len(alarms))
			}
			if !alarms[0].TriggerIsRelative {
				t.Error("Expected relative alarm")
			}
			if alarms[0].TriggerDuration != tt.wantDuration {
				t.Errorf("TriggerDuration = %v, want %v", alarms[0].TriggerDuration, tt.wantDuration)
			}
			if alarms[0].Action != "DISPLAY" {
				t.Errorf("Action = %q, want DISPLAY", alarms[0].Action)
			}
			if alarms[0].Description != "Reminder" {
				t.Errorf("Description = %q, want Reminder", alarms[0].Description)
			}
		})
	}
}

func TestAlarmsParser_Integration_AbsoluteAlarms(t *testing.T) {
	input := "2025-11-15T10:00:00Z"
	alarms, err := ParseAlarmsFromString(input, "UTC")
	if err != nil {
		t.Fatalf("ParseAlarmsFromString(%q) error: %v", input, err)
	}
	if len(alarms) != 1 {
		t.Fatalf("Expected 1 alarm, got %d", len(alarms))
	}
	if alarms[0].TriggerIsRelative {
		t.Error("Expected absolute alarm")
	}
	if alarms[0].TriggerTime.IsZero() {
		t.Error("TriggerTime should not be zero")
	}
	expected := time.Date(2025, 11, 15, 10, 0, 0, 0, time.UTC)
	if !alarms[0].TriggerTime.Equal(expected) {
		t.Errorf("TriggerTime = %v, want %v", alarms[0].TriggerTime, expected)
	}
}

func TestAlarmsParser_Integration_ComplexKeyValue(t *testing.T) {
	input := "trigger=15m,action=EMAIL,description=Meeting reminder,summary=Important Meeting,repeat=3,repeat_duration=5m"
	alarms, err := ParseAlarmsFromString(input, "UTC")
	if err != nil {
		t.Fatalf("ParseAlarmsFromString(%q) error: %v", input, err)
	}
	if len(alarms) != 1 {
		t.Fatalf("Expected 1 alarm, got %d", len(alarms))
	}

	alarm := alarms[0]
	if alarm.Action != "EMAIL" {
		t.Errorf("Action = %q, want EMAIL", alarm.Action)
	}
	if alarm.Description != "Meeting reminder" {
		t.Errorf("Description = %q, want 'Meeting reminder'", alarm.Description)
	}
	if alarm.Summary != "Important Meeting" {
		t.Errorf("Summary = %q, want 'Important Meeting'", alarm.Summary)
	}
	if alarm.TriggerDuration != -15*time.Minute {
		t.Errorf("TriggerDuration = %v, want -15m", alarm.TriggerDuration)
	}
	if alarm.Repeat != 3 {
		t.Errorf("Repeat = %d, want 3", alarm.Repeat)
	}
	if alarm.RepeatDuration != 5*time.Minute {
		t.Errorf("RepeatDuration = %v, want 5m", alarm.RepeatDuration)
	}
}

func TestAlarmsParser_Integration_MultipleAlarms(t *testing.T) {
	input := "15m,30m,1h,2h"
	alarms, err := ParseAlarmsFromString(input, "UTC")
	if err != nil {
		t.Fatalf("ParseAlarmsFromString(%q) error: %v", input, err)
	}
	if len(alarms) != 4 {
		t.Fatalf("Expected 4 alarms, got %d", len(alarms))
	}

	expected := []time.Duration{-15 * time.Minute, -30 * time.Minute, -1 * time.Hour, -2 * time.Hour}
	for i, alarm := range alarms {
		if alarm.TriggerDuration != expected[i] {
			t.Errorf("Alarm[%d] TriggerDuration = %v, want %v", i, alarm.TriggerDuration, expected[i])
		}
	}
}

func TestAlarmsParser_Integration_AlternativeKeys(t *testing.T) {
	tests := []struct {
		name  string
		input string
		field string
		want  string
	}{
		{"offset instead of trigger", "offset=15m", "trigger", "15m"},
		{"message instead of description", "trigger=15m,message=Test", "description", "Test"},
		{"text instead of description", "trigger=15m,text=Test", "description", "Test"},
		{"title instead of summary", "trigger=15m,title=Test", "summary", "Test"},
		{"when instead of direction", "trigger=15m,when=after", "direction", "after"},
		{"repetitions instead of repeat", "trigger=15m,repetitions=3,repeat_interval=5m", "repeat", "3"},
		{"is_relative instead of relative", "trigger=15m,is_relative=yes", "relative", "yes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alarms, err := ParseAlarmsFromString(tt.input, "UTC")
			if err != nil {
				t.Errorf("ParseAlarmsFromString(%q) error: %v", tt.input, err)
			}
			if len(alarms) != 1 {
				t.Errorf("Expected 1 alarm, got %d", len(alarms))
			}
		})
	}
}

func TestAlarmsParser_EdgeCase_ZeroRelativeDuration(t *testing.T) {
	input := "trigger=0m"
	_, err := ParseAlarmsFromString(input, "UTC")
	if err == nil {
		t.Error("Expected error for zero relative duration")
	}
}

func TestAlarmsParser_EdgeCase_DirectionHints(t *testing.T) {
	tests := []struct {
		direction string
		wantSign  int // 1 for positive, -1 for negative
	}{
		{"after", 1},
		{"post", 1},
		{"later", 1},
		{"follow", 1},
		{"following", 1},
		{"plus", 1},
		{"before", -1},
		{"prior", -1},
		{"pre", -1},
		{"minus", -1},
	}

	for _, tt := range tests {
		t.Run(tt.direction, func(t *testing.T) {
			input := fmt.Sprintf("trigger=15m,direction=%s", tt.direction)
			alarms, err := ParseAlarmsFromString(input, "UTC")
			if err != nil {
				t.Fatalf("ParseAlarmsFromString(%q) error: %v", input, err)
			}
			if len(alarms) != 1 {
				t.Fatalf("Expected 1 alarm, got %d", len(alarms))
			}

			expected := time.Duration(tt.wantSign) * 15 * time.Minute
			if alarms[0].TriggerDuration != expected {
				t.Errorf("TriggerDuration = %v, want %v (direction=%s)", alarms[0].TriggerDuration, expected, tt.direction)
			}
		})
	}
}

// ========================================
// Additional edge case tests for higher coverage
// ========================================

func TestAlarmsParser_EdgeCase_SplitAlarmInput_Recursive(t *testing.T) {
	// Test recursive splitting with double pipe containing another double pipe
	input := "15m||30m||trigger=1h"
	result := SplitAlarmInput(input)
	if len(result) != 3 {
		t.Errorf("Expected 3 items, got %d: %v", len(result), result)
	}
}

func TestAlarmsParser_EdgeCase_KeyValueAlarmSpec_ForceRelativeWithError(t *testing.T) {
	// Force relative but trigger is absolute
	input := "trigger=2025-11-15T10:00:00Z,kind=relative"
	_, err := ParseAlarmsFromString(input, "UTC")
	if err == nil {
		t.Error("Expected error when forcing relative with absolute trigger")
	}
}

func TestAlarmsParser_EdgeCase_KeyValueAlarmSpec_ForceAbsoluteWithError(t *testing.T) {
	// Force absolute but trigger is relative and invalid as absolute
	input := "trigger=15m,kind=absolute"
	_, err := ParseAlarmsFromString(input, "UTC")
	if err == nil {
		t.Error("Expected error when forcing absolute with invalid absolute trigger")
	}
}

func TestAlarmsParser_EdgeCase_KeyValueAlarmSpec_NegativeRepeatDuration(t *testing.T) {
	// Negative repeat duration should fail
	input := "trigger=15m,repeat=3,repeat_duration=-5m"
	_, err := ParseAlarmsFromString(input, "UTC")
	if err == nil {
		t.Error("Expected error for negative repeat duration")
	}
}

func TestAlarmsParser_EdgeCase_KeyValueAlarmSpec_ZeroRepeatDuration(t *testing.T) {
	// Zero repeat duration should fail
	input := "trigger=15m,repeat=3,repeat_duration=0m"
	_, err := ParseAlarmsFromString(input, "UTC")
	if err == nil {
		t.Error("Expected error for zero repeat duration")
	}
}

func TestAlarmsParser_EdgeCase_KeyValueAlarmSpec_DurationOnly(t *testing.T) {
	// Duration without repeat count should fail
	input := "trigger=15m,repeat_duration=5m"
	_, err := ParseAlarmsFromString(input, "UTC")
	if err == nil {
		t.Error("Expected error for repeat_duration without repeat count")
	}
}

func TestAlarmsParser_EdgeCase_KeyValueAlarmSpec_RelativeFalse(t *testing.T) {
	// relative=false should force absolute
	input := "trigger=2025-11-15T10:00:00Z,relative=no"
	alarms, err := ParseAlarmsFromString(input, "UTC")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if len(alarms) != 1 {
		t.Fatalf("Expected 1 alarm, got %d", len(alarms))
	}
	if alarms[0].TriggerIsRelative {
		t.Error("Expected absolute trigger with relative=no")
	}
}

func TestAlarmsParser_EdgeCase_KeyValueAlarmSpec_EmptyKeySegment(t *testing.T) {
	// Empty segment with just spaces should be ignored
	input := "trigger=15m;   ;action=DISPLAY"
	alarms, err := ParseAlarmsFromString(input, "UTC")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if len(alarms) != 1 {
		t.Fatalf("Expected 1 alarm, got %d", len(alarms))
	}
}

func TestAlarmsParser_EdgeCase_ParseAlarmAbsolute_InvalidTimezone(t *testing.T) {
	// Invalid timezone should still work by falling back to UTC
	_, err := parseAlarmAbsolute("2025-11-15 10:00:00", "Invalid/Timezone")
	if err != nil {
		t.Errorf("Should not error with invalid timezone, should fall back: %v", err)
	}
}

func TestAlarmsParser_EdgeCase_ParseAlarmDurationValue_NegativeGoDuration(t *testing.T) {
	// Go duration that parses as negative should be rejected
	_, err := parseAlarmDurationValue("-1h30m")
	if err == nil {
		t.Error("Expected error for negative Go duration")
	}
}

func TestAlarmsParser_EdgeCase_ParseICSDuration_OnlyWeeks(t *testing.T) {
	// Test week-only duration
	result, err := parseICSDuration("P2W")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	expected := 2 * 7 * 24 * time.Hour
	if result != expected {
		t.Errorf("Result = %v, want %v", result, expected)
	}
}

func TestAlarmsParser_EdgeCase_ParseICSDuration_OnlyDays(t *testing.T) {
	// Test day-only duration
	result, err := parseICSDuration("P5D")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	expected := 5 * 24 * time.Hour
	if result != expected {
		t.Errorf("Result = %v, want %v", result, expected)
	}
}

func TestAlarmsParser_EdgeCase_ParseICSDuration_OnlyHours(t *testing.T) {
	// Test hour-only duration
	result, err := parseICSDuration("PT3H")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	expected := 3 * time.Hour
	if result != expected {
		t.Errorf("Result = %v, want %v", result, expected)
	}
}

func TestAlarmsParser_EdgeCase_ParseICSDuration_OnlySeconds(t *testing.T) {
	// Test second-only duration
	result, err := parseICSDuration("PT30S")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	expected := 30 * time.Second
	if result != expected {
		t.Errorf("Result = %v, want %v", result, expected)
	}
}

func TestAlarmsParser_EdgeCase_KeyValueAlarmSpec_EmptyDescription(t *testing.T) {
	// Empty description should default to "Reminder" for DISPLAY action
	input := "trigger=15m,action=DISPLAY,description="
	alarms, err := ParseAlarmsFromString(input, "UTC")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if len(alarms) != 1 {
		t.Fatalf("Expected 1 alarm, got %d", len(alarms))
	}
	if alarms[0].Description != "Reminder" {
		t.Errorf("Description = %q, want 'Reminder'", alarms[0].Description)
	}
}

func TestAlarmsParser_EdgeCase_ParseAlarmAbsolute_WithLocalTimezone(t *testing.T) {
	// Test with a specific timezone
	result, err := parseAlarmAbsolute("2025-11-15 10:00:00", "Europe/Madrid")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if result.IsZero() {
		t.Error("Result should not be zero")
	}
}

func TestAlarmsParser_EdgeCase_SplitAlarmInput_OnlyPipes(t *testing.T) {
	// Only separators should return empty
	input := "|||"
	result := SplitAlarmInput(input)
	if len(result) != 0 {
		t.Errorf("Expected 0 items for only separators, got %d: %v", len(result), result)
	}
}

func TestAlarmsParser_EdgeCase_SplitAlarmInput_ComplexNested(t *testing.T) {
	// Complex nested with key-value and double pipe
	input := "trigger=15m,action=DISPLAY||30m||trigger=1h,action=EMAIL"
	result := SplitAlarmInput(input)
	// Should split on || but keep key-value pairs together
	if len(result) != 3 {
		t.Errorf("Expected 3 items, got %d: %v", len(result), result)
	}
}
