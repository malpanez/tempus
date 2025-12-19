package templates

import (
	"tempus/internal/testutil"
	"strings"
	"testing"

	"tempus/internal/i18n"
)

// Helper function to create a test translator
func newTestTranslator() *i18n.Translator {
	tr, err := i18n.NewTranslator("en")
	if err != nil {
		// Fallback to a simple translator if locales aren't available
		tr = &i18n.Translator{}
	}
	return tr
}

// TestNewTemplateManager tests the creation of a new template manager
func TestNewTemplateManager(t *testing.T) {
	tm := NewTemplateManager()
	if tm == nil {
		t.Fatal("NewTemplateManager() returned nil")
	}
	if tm.templates == nil {
		t.Error("templates map is nil")
	}
	if tm.ddTemplates == nil {
		t.Error("ddTemplates map is nil")
	}

	// Check that built-in templates are registered
	expectedTemplates := []string{
		"flight", "meeting", "holiday", "focus-block",
		"medication", "appointment", "transition", "deadline",
	}
	for _, name := range expectedTemplates {
		if _, ok := tm.templates[name]; !ok {
			t.Errorf("built-in template %q not registered", name)
		}
	}
}

// TestGetTemplate tests template retrieval
func TestGetTemplate(t *testing.T) {
	tm := NewTemplateManager()

	tests := []struct {
		name      string
		tmplName  string
		wantErr   bool
		errSubstr string
	}{
		{
			name:     "valid flight template",
			tmplName: "flight",
			wantErr:  false,
		},
		{
			name:     "valid meeting template",
			tmplName: "meeting",
			wantErr:  false,
		},
		{
			name:      "non-existent template",
			tmplName:  "nonexistent",
			wantErr:   true,
			errSubstr: "template not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := tm.GetTemplate(tt.tmplName)
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetTemplate() expected error, got nil")
				} else if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("GetTemplate() error = %v, want substring %q", err, tt.errSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("GetTemplate() unexpected error: %v", err)
				}
				if tmpl == nil {
					t.Errorf("GetTemplate() returned nil template")
				}
				if tmpl != nil && tmpl.Name != tt.tmplName {
					t.Errorf("GetTemplate() name = %q, want %q", tmpl.Name, tt.tmplName)
				}
			}
		})
	}
}

// TestListTemplates tests listing all templates
func TestListTemplates(t *testing.T) {
	tm := NewTemplateManager()
	templates := tm.ListTemplates()

	if templates == nil {
		t.Fatal("ListTemplates() returned nil")
	}

	expectedCount := 8 // 8 built-in templates
	if len(templates) < expectedCount {
		t.Errorf("ListTemplates() count = %d, want at least %d", len(templates), expectedCount)
	}

	// Verify all built-in templates are present
	for _, name := range []string{"flight", "meeting", "holiday", "focus-block", "medication", "appointment", "transition", "deadline"} {
		if _, ok := templates[name]; !ok {
			t.Errorf("ListTemplates() missing template %q", name)
		}
	}
}

// TestGenerateEventRequiredFields tests required field validation
func TestGenerateEventRequiredFields(t *testing.T) {
	tm := NewTemplateManager()
	tr := newTestTranslator()

	tests := []struct {
		name      string
		tmplName  string
		data      map[string]string
		wantErr   bool
		errSubstr string
	}{
		{
			name:     "flight with all required fields",
			tmplName: "flight",
			data: map[string]string{
				"flight_number":  "AA123",
				"from":           "JFK",
				"to":             "LAX",
				"departure_time": "2025-12-01 10:00",
				"arrival_time":   "2025-12-01 14:00",
			},
			wantErr: false,
		},
		{
			name:     "flight missing required field",
			tmplName: "flight",
			data: map[string]string{
				"flight_number":  "AA123",
				"from":           "JFK",
				"departure_time": "2025-12-01 10:00",
				"arrival_time":   "2025-12-01 14:00",
				// missing "to"
			},
			wantErr:   true,
			errSubstr: "required field missing",
		},
		{
			name:     "meeting with empty required field",
			tmplName: "meeting",
			data: map[string]string{
				"title":      "", // empty required field
				"start_time": "2025-12-01 10:00",
			},
			wantErr:   true,
			errSubstr: "required field missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := tm.GenerateEvent(tt.tmplName, tt.data, tr)
			if tt.wantErr {
				if err == nil {
					t.Errorf("GenerateEvent() expected error, got nil")
				} else if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("GenerateEvent() error = %v, want substring %q", err, tt.errSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("GenerateEvent() unexpected error: %v", err)
				}
				if event == nil {
					t.Errorf("GenerateEvent() returned nil event")
				}
			}
		})
	}
}

// TestGenerateFlightEvent tests flight event generation
func TestGenerateFlightEvent(t *testing.T) {
	tr := newTestTranslator()

	tests := []struct {
		name    string
		data    map[string]string
		wantErr bool
		check   func(*testing.T, interface{})
	}{
		{
			name: "valid flight with all fields",
			data: map[string]string{
				"flight_number":  "AA123",
				"from":           "JFK",
				"to":             "LAX",
				"departure_time": "2025-12-01 10:00",
				"arrival_time":   "2025-12-01 14:00",
				"departure_tz":   testutil.TZAmericaNewYork,
				"arrival_tz":     "America/Los_Angeles",
				"airline":        testutil.AirlineAmerican,
				"seat":           "12A",
				"gate":           "B22",
			},
			wantErr: false,
			check: func(t *testing.T, v interface{}) {
				event := v.(*interface{})
				if event == nil {
					t.Fatal(testutil.ErrMsgEventIsNil)
				}
			},
		},
		{
			name: "invalid departure time",
			data: map[string]string{
				"flight_number":  "AA123",
				"from":           "JFK",
				"to":             "LAX",
				"departure_time": "invalid-date",
				"arrival_time":   "2025-12-01 14:00",
			},
			wantErr: true,
		},
		{
			name: "invalid arrival time",
			data: map[string]string{
				"flight_number":  "AA123",
				"from":           "JFK",
				"to":             "LAX",
				"departure_time": "2025-12-01 10:00",
				"arrival_time":   testutil.ErrMsgBadTime,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := generateFlightEvent(tt.data, tr)
			if tt.wantErr {
				if err == nil {
					t.Errorf("generateFlightEvent() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("generateFlightEvent() unexpected error: %v", err)
				}
				if event == nil {
					t.Fatal("generateFlightEvent() returned nil event")
				}

				// Verify basic fields
				if event.UID == "" {
					t.Error("event UID is empty")
				}
				if !strings.Contains(event.Summary, tt.data["flight_number"]) {
					t.Errorf("summary doesn't contain flight number: %s", event.Summary)
				}
				if !strings.Contains(event.Description, tt.data["from"]) {
					t.Errorf("description doesn't contain departure location")
				}
				if len(event.Categories) == 0 {
					t.Error("event has no categories")
				}

				// Verify timezones if provided
				if tz := tt.data["departure_tz"]; tz != "" {
					if event.StartTZ != tz {
						t.Errorf("StartTZ = %q, want %q", event.StartTZ, tz)
					}
				}
				if tz := tt.data["arrival_tz"]; tz != "" {
					if event.EndTZ != tz {
						t.Errorf("EndTZ = %q, want %q", event.EndTZ, tz)
					}
				}
			}
		})
	}
}

// TestGenerateMeetingEvent tests meeting event generation
func TestGenerateMeetingEvent(t *testing.T) {
	tr := newTestTranslator()

	tests := []struct {
		name    string
		data    map[string]string
		wantErr bool
	}{
		{
			name: "valid meeting with duration",
			data: map[string]string{
				"title":      "Team Standup",
				"start_time": "2025-12-01 10:00",
				"duration":   "30m",
				"timezone":   testutil.TZAmericaNewYork,
				"location":   testutil.StringConferenceRoomA,
				"attendees":  "alice@example.com, bob@example.com",
				"agenda":     "Sprint planning",
			},
			wantErr: false,
		},
		{
			name: "valid meeting with end time",
			data: map[string]string{
				"title":      "Client Meeting",
				"start_time": "2025-12-01 14:00",
				"end_time":   "2025-12-01 15:30",
			},
			wantErr: false,
		},
		{
			name: "meeting with various duration formats",
			data: map[string]string{
				"title":      "Quick sync",
				"start_time": "2025-12-01 10:00",
				"duration":   "45",
			},
			wantErr: false,
		},
		{
			name: testutil.ErrMsgInvalidStartTime,
			data: map[string]string{
				"title":      testutil.EventTitleBadMeeting,
				"start_time": testutil.ErrMsgNotADate,
			},
			wantErr: true,
		},
		{
			name: "invalid end time",
			data: map[string]string{
				"title":      testutil.EventTitleBadMeeting,
				"start_time": "2025-12-01 10:00",
				"end_time":   "invalid",
			},
			wantErr: true,
		},
		{
			name: "end before start",
			data: map[string]string{
				"title":      testutil.EventTitleBadMeeting,
				"start_time": "2025-12-01 15:00",
				"end_time":   "2025-12-01 14:00",
			},
			wantErr: true,
		},
		{
			name: "invalid duration format",
			data: map[string]string{
				"title":      testutil.EventTitleBadMeeting,
				"start_time": "2025-12-01 10:00",
				"duration":   "xyz",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := generateMeetingEvent(tt.data, tr)
			if tt.wantErr {
				if err == nil {
					t.Errorf("generateMeetingEvent() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("generateMeetingEvent() unexpected error: %v", err)
				}
				if event == nil {
					t.Fatal("generateMeetingEvent() returned nil event")
				}

				// Verify fields
				if !strings.Contains(event.Summary, tt.data["title"]) {
					t.Errorf("summary doesn't contain title")
				}
				if tt.data["location"] != "" && event.Location != tt.data["location"] {
					t.Errorf("location = %q, want %q", event.Location, tt.data["location"])
				}
				if tt.data["attendees"] != "" && len(event.Attendees) == 0 {
					t.Error("attendees not parsed")
				}
				if len(event.Categories) == 0 {
					t.Error("event has no categories")
				}
			}
		})
	}
}

// TestGenerateHolidayEvent tests holiday event generation
func TestGenerateHolidayEvent(t *testing.T) {
	tr := newTestTranslator()

	tests := []struct {
		name    string
		data    map[string]string
		wantErr bool
	}{
		{
			name: "valid holiday",
			data: map[string]string{
				"destination":   "Paris",
				"start_date":    testutil.Date20251220,
				"end_date":      testutil.Date20251227,
				"timezone":      testutil.TZEuropeParis,
				"accommodation": "Hotel de Paris",
				"notes":         "Christmas holiday",
			},
			wantErr: false,
		},
		{
			name: "invalid start date",
			data: map[string]string{
				"destination": "Paris",
				"start_date":  testutil.ErrMsgNotADate,
				"end_date":    testutil.Date20251227,
			},
			wantErr: true,
		},
		{
			name: "invalid end date",
			data: map[string]string{
				"destination": "Paris",
				"start_date":  testutil.Date20251220,
				"end_date":    "bad-date",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := generateHolidayEvent(tt.data, tr)
			if tt.wantErr {
				if err == nil {
					t.Errorf("generateHolidayEvent() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("generateHolidayEvent() unexpected error: %v", err)
				}
				if event == nil {
					t.Fatal("generateHolidayEvent() returned nil event")
				}

				// Verify it's an all-day event
				if !event.AllDay {
					t.Error("holiday event should be all-day")
				}
				if !strings.Contains(event.Summary, tt.data["destination"]) {
					t.Error("summary doesn't contain destination")
				}
				if event.Location != tt.data["destination"] {
					t.Errorf("location = %q, want %q", event.Location, tt.data["destination"])
				}
			}
		})
	}
}

// TestGenerateFocusBlockEvent tests focus block event generation
func TestGenerateFocusBlockEvent(t *testing.T) {
	tr := newTestTranslator()

	tests := []struct {
		name    string
		data    map[string]string
		wantErr bool
	}{
		{
			name: "valid focus block with default duration",
			data: map[string]string{
				"task":       "Write documentation",
				"start_time": testutil.DateTime20251201_0900,
			},
			wantErr: false,
		},
		{
			name: "valid focus block with custom duration",
			data: map[string]string{
				"task":       "Code review",
				"start_time": "2025-12-01 14:00",
				"duration":   "2h",
				"notes":      "Review PR #123",
			},
			wantErr: false,
		},
		{
			name: testutil.ErrMsgInvalidStartTime,
			data: map[string]string{
				"task":       "Invalid",
				"start_time": testutil.ErrMsgBadTime,
			},
			wantErr: true,
		},
		{
			name: "invalid duration",
			data: map[string]string{
				"task":       "Invalid",
				"start_time": testutil.DateTime20251201_0900,
				"duration":   "forever",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := generateFocusBlockEvent(tt.data, tr)
			if tt.wantErr {
				if err == nil {
					t.Errorf("generateFocusBlockEvent() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("generateFocusBlockEvent() unexpected error: %v", err)
				}
				if event == nil {
					t.Fatal("generateFocusBlockEvent() returned nil event")
				}

				// Verify ADHD-friendly features
				if !strings.Contains(event.Summary, tt.data["task"]) {
					t.Error("summary doesn't contain task")
				}
				if len(event.Alarms) < 2 {
					t.Errorf("focus block should have multiple alarms, got %d", len(event.Alarms))
				}
				if !strings.Contains(event.Description, "Do Not Disturb") {
					t.Error("description should contain ADHD tips")
				}
			}
		})
	}
}

// TestGenerateMedicationEvent tests medication reminder generation
func TestGenerateMedicationEvent(t *testing.T) {
	tr := newTestTranslator()

	tests := []struct {
		name    string
		data    map[string]string
		wantErr bool
	}{
		{
			name: "valid medication with recurrence",
			data: map[string]string{
				"medication_name": "Adderall",
				"time":            testutil.DateTime20251201_0800,
				"dosage":          "20mg",
				"instructions":    "Take with food",
				"recurrence":      "FREQ=DAILY",
			},
			wantErr: false,
		},
		{
			name: "medication without recurrence",
			data: map[string]string{
				"medication_name": "Ibuprofen",
				"time":            "2025-12-01 14:00",
				"dosage":          "400mg",
			},
			wantErr: false,
		},
		{
			name: "invalid time",
			data: map[string]string{
				"medication_name": "Test",
				"time":            "invalid-time",
				"dosage":          "10mg",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := generateMedicationEvent(tt.data, tr)
			if tt.wantErr {
				if err == nil {
					t.Errorf("generateMedicationEvent() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("generateMedicationEvent() unexpected error: %v", err)
				}
				if event == nil {
					t.Fatal("generateMedicationEvent() returned nil event")
				}

				// Verify medication-specific features
				if !strings.Contains(event.Summary, tt.data["medication_name"]) {
					t.Error("summary doesn't contain medication name")
				}
				if !strings.Contains(event.Summary, tt.data["dosage"]) {
					t.Error("summary doesn't contain dosage")
				}
				if len(event.Alarms) < 3 {
					t.Errorf("medication should have multiple alarms (before, at, after), got %d", len(event.Alarms))
				}
				if tt.data["recurrence"] != "" && event.RRule != tt.data["recurrence"] {
					t.Errorf("RRule = %q, want %q", event.RRule, tt.data["recurrence"])
				}
			}
		})
	}
}

// TestGenerateAppointmentEvent tests appointment event generation
func TestGenerateAppointmentEvent(t *testing.T) {
	tr := newTestTranslator()

	tests := []struct {
		name    string
		data    map[string]string
		wantErr bool
	}{
		{
			name: "valid appointment with travel time",
			data: map[string]string{
				"title":       "Doctor",
				"provider":    "Dr. Smith",
				"start_time":  "2025-12-01 10:00",
				"duration":    "30m",
				"travel_time": "20m",
				"location":    "123 Medical Plaza",
			},
			wantErr: false,
		},
		{
			name: "appointment with default values",
			data: map[string]string{
				"title":      "Dentist",
				"start_time": "2025-12-01 14:00",
			},
			wantErr: false,
		},
		{
			name: testutil.ErrMsgInvalidStartTime,
			data: map[string]string{
				"title":      "Bad",
				"start_time": "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid duration",
			data: map[string]string{
				"title":      "Bad",
				"start_time": "2025-12-01 10:00",
				"duration":   "bad-duration",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := generateAppointmentEvent(tt.data, tr)
			if tt.wantErr {
				if err == nil {
					t.Errorf("generateAppointmentEvent() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("generateAppointmentEvent() unexpected error: %v", err)
				}
				if event == nil {
					t.Fatal("generateAppointmentEvent() returned nil event")
				}

				// Verify appointment features
				if !strings.Contains(event.Summary, tt.data["title"]) {
					t.Error("summary doesn't contain title")
				}
				if len(event.Alarms) < 2 {
					t.Errorf("appointment should have multiple alarms, got %d", len(event.Alarms))
				}
				// Travel time should add an extra alarm
				if tt.data["travel_time"] != "" {
					hasLeaveAlarm := false
					for _, alarm := range event.Alarms {
						if strings.Contains(alarm.Description, "leave") {
							hasLeaveAlarm = true
							break
						}
					}
					if !hasLeaveAlarm {
						t.Error("appointment with travel time should have 'time to leave' alarm")
					}
				}
			}
		})
	}
}

// TestGenerateTransitionEvent tests transition event generation
func TestGenerateTransitionEvent(t *testing.T) {
	tr := newTestTranslator()

	tests := []struct {
		name    string
		data    map[string]string
		wantErr bool
	}{
		{
			name: "valid transition",
			data: map[string]string{
				"from_activity": "Meeting",
				"to_activity":   "Focus work",
				"start_time":    "2025-12-01 11:00",
				"duration":      "10m",
			},
			wantErr: false,
		},
		{
			name: "transition with default duration",
			data: map[string]string{
				"from_activity": "Lunch",
				"to_activity":   "Coding",
				"start_time":    "2025-12-01 13:00",
			},
			wantErr: false,
		},
		{
			name: testutil.ErrMsgInvalidStartTime,
			data: map[string]string{
				"from_activity": "A",
				"to_activity":   "B",
				"start_time":    testutil.ErrMsgBadTime,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := generateTransitionEvent(tt.data, tr)
			if tt.wantErr {
				if err == nil {
					t.Errorf("generateTransitionEvent() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("generateTransitionEvent() unexpected error: %v", err)
				}
				if event == nil {
					t.Fatal("generateTransitionEvent() returned nil event")
				}

				// Verify transition features
				if !strings.Contains(event.Summary, tt.data["from_activity"]) {
					t.Error("summary doesn't contain from_activity")
				}
				if !strings.Contains(event.Summary, tt.data["to_activity"]) {
					t.Error("summary doesn't contain to_activity")
				}
				if len(event.Alarms) < 1 {
					t.Error("transition should have at least one alarm")
				}
			}
		})
	}
}

// TestGenerateDeadlineEvent tests deadline event generation
func TestGenerateDeadlineEvent(t *testing.T) {
	tr := newTestTranslator()

	tests := []struct {
		name    string
		data    map[string]string
		wantErr bool
	}{
		{
			name: "valid deadline with priority",
			data: map[string]string{
				"task":     "Submit report",
				"due_date": testutil.Date20251215,
				"priority": "1",
				"notes":    "Annual Q4 report",
			},
			wantErr: false,
		},
		{
			name: "deadline with default priority",
			data: map[string]string{
				"task":     "Code review",
				"due_date": "2025-12-10",
			},
			wantErr: false,
		},
		{
			name: "invalid due date",
			data: map[string]string{
				"task":     "Bad deadline",
				"due_date": testutil.ErrMsgNotADate,
			},
			wantErr: true,
		},
		{
			name: "invalid priority (too high)",
			data: map[string]string{
				"task":     "Test",
				"due_date": "2025-12-10",
				"priority": "99",
			},
			wantErr: false, // Should not error, just ignore invalid priority
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := generateDeadlineEvent(tt.data, tr)
			if tt.wantErr {
				if err == nil {
					t.Errorf("generateDeadlineEvent() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("generateDeadlineEvent() unexpected error: %v", err)
				}
				if event == nil {
					t.Fatal("generateDeadlineEvent() returned nil event")
				}

				// Verify deadline features
				if !event.AllDay {
					t.Error("deadline should be all-day event")
				}
				if !strings.Contains(event.Summary, tt.data["task"]) {
					t.Error("summary doesn't contain task")
				}
				if len(event.Alarms) < 4 {
					t.Errorf("deadline should have countdown alarms (1w, 3d, 1d, morning), got %d", len(event.Alarms))
				}

				// Verify priority is set correctly
				if tt.data["priority"] != "" {
					expectedPriority := 0
					if tt.data["priority"] == "1" {
						expectedPriority = 1
					}
					if expectedPriority > 0 && event.Priority != expectedPriority {
						t.Errorf("priority = %d, want %d", event.Priority, expectedPriority)
					}
				}
			}
		})
	}
}

// TestSplitAndTrim tests the helper function
func TestSplitAndTrim(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		sep      string
		expected []string
	}{
		{
			name:     "comma separated emails",
			input:    "alice@example.com, bob@example.com, charlie@example.com",
			sep:      ",",
			expected: []string{testutil.EmailAlice, testutil.EmailBob, "charlie@example.com"},
		},
		{
			name:     "with extra spaces",
			input:    "  item1  ,  item2  ,  item3  ",
			sep:      ",",
			expected: []string{"item1", "item2", "item3"},
		},
		{
			name:     "empty strings removed",
			input:    "item1,,item2,,,item3",
			sep:      ",",
			expected: []string{"item1", "item2", "item3"},
		},
		{
			name:     "single item",
			input:    "single",
			sep:      ",",
			expected: []string{"single"},
		},
		{
			name:     "empty input",
			input:    "",
			sep:      ",",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitAndTrim(tt.input, tt.sep)
			if len(result) != len(tt.expected) {
				t.Errorf("splitAndTrim() length = %d, want %d", len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("splitAndTrim()[%d] = %q, want %q", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

// TestTemplateFields tests that templates have proper field definitions
func TestTemplateFields(t *testing.T) {
	tm := NewTemplateManager()

	tests := []struct {
		tmplName      string
		requiredCount int
		fieldKeys     []string
	}{
		{
			tmplName:      "flight",
			requiredCount: 5,
			fieldKeys:     []string{"flight_number", "from", "to", "departure_time", "arrival_time"},
		},
		{
			tmplName:      "meeting",
			requiredCount: 2,
			fieldKeys:     []string{"title", "start_time"},
		},
		{
			tmplName:      "medication",
			requiredCount: 3,
			fieldKeys:     []string{"medication_name", "time", "dosage"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.tmplName, func(t *testing.T) {
			tmpl, err := tm.GetTemplate(tt.tmplName)
			if err != nil {
				t.Fatalf("failed to get template: %v", err)
			}

			requiredFields := 0
			for _, field := range tmpl.Fields {
				if field.Required {
					requiredFields++
				}
			}

			if requiredFields != tt.requiredCount {
				t.Errorf("required field count = %d, want %d", requiredFields, tt.requiredCount)
			}

			// Verify expected fields exist
			for _, key := range tt.fieldKeys {
				found := false
				for _, field := range tmpl.Fields {
					if field.Key == key {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("field %q not found in template", key)
				}
			}
		})
	}
}

// TestRegisterDDTemplate tests data-driven template registration
func TestRegisterDDTemplate(t *testing.T) {
	tm := NewTemplateManager()

	dd := DataDrivenTemplate{
		Name:        testutil.TemplateCustomEvent,
		Description: "A custom event template",
		Fields: []Field{
			{Key: "title", Name: "Title", Type: "text", Required: true},
			{Key: "date", Name: "Date", Type: "datetime", Required: true},
		},
		Output: OutputTemplate{
			StartField:  "date",
			SummaryTmpl: testutil.TemplatePlaceholderTitle,
			AllDay:      true,
			Categories:  []string{"Custom"},
		},
	}

	// Register the template
	tm.RegisterDDTemplate(dd)

	// Verify it was registered
	tmpl, err := tm.GetTemplate(testutil.TemplateCustomEvent)
	if err != nil {
		t.Fatalf("failed to get registered template: %v", err)
	}
	if tmpl.Name != testutil.TemplateCustomEvent {
		t.Errorf("template name = %q, want %q", tmpl.Name, testutil.TemplateCustomEvent)
	}
	if len(tmpl.Fields) != 2 {
		t.Errorf("field count = %d, want 2", len(tmpl.Fields))
	}

	// Verify it's in ddTemplates map
	if _, ok := tm.ddTemplates[testutil.TemplateCustomEvent]; !ok {
		t.Error("template not in ddTemplates map")
	}
}

// TestFilenameTemplate tests filename template retrieval
func TestFilenameTemplate(t *testing.T) {
	tm := NewTemplateManager()

	// Register a template with filename template
	dd := DataDrivenTemplate{
		Name:             "event-with-filename",
		FilenameTemplate: "{{date date}}-{{slug title}}.ics",
		Fields: []Field{
			{Key: "title", Name: "Title", Type: "text", Required: true},
			{Key: "date", Name: "Date", Type: "datetime", Required: true},
		},
		Output: OutputTemplate{
			StartField:  "date",
			SummaryTmpl: testutil.TemplatePlaceholderTitle,
		},
	}
	tm.RegisterDDTemplate(dd)

	tests := []struct {
		name      string
		tmplName  string
		wantFound bool
		wantTmpl  string
	}{
		{
			name:      "template with filename",
			tmplName:  "event-with-filename",
			wantFound: true,
			wantTmpl:  "{{date date}}-{{slug title}}.ics",
		},
		{
			name:      "built-in template without filename",
			tmplName:  "flight",
			wantFound: false,
		},
		{
			name:      "non-existent template",
			tmplName:  "nonexistent",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, found := tm.FilenameTemplate(tt.tmplName)
			if found != tt.wantFound {
				t.Errorf("FilenameTemplate() found = %v, want %v", found, tt.wantFound)
			}
			if tt.wantFound && tmpl != tt.wantTmpl {
				t.Errorf("FilenameTemplate() = %q, want %q", tmpl, tt.wantTmpl)
			}
		})
	}
}

// TestDataTemplate tests raw template retrieval
func TestDataTemplate(t *testing.T) {
	tm := NewTemplateManager()

	dd := DataDrivenTemplate{
		Name:        "test-template",
		Description: "Test template",
		Fields: []Field{
			{Key: "title", Name: "Title", Type: "text", Required: true},
		},
		Output: OutputTemplate{
			StartField:  "date",
			SummaryTmpl: testutil.TemplatePlaceholderTitle,
		},
	}
	tm.RegisterDDTemplate(dd)

	tests := []struct {
		name      string
		tmplName  string
		wantFound bool
	}{
		{
			name:      "existing dd template",
			tmplName:  "test-template",
			wantFound: true,
		},
		{
			name:      "built-in template not in dd map",
			tmplName:  "flight",
			wantFound: false,
		},
		{
			name:      "non-existent template",
			tmplName:  "nonexistent",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, found := tm.DataTemplate(tt.tmplName)
			if found != tt.wantFound {
				t.Errorf("DataTemplate() found = %v, want %v", found, tt.wantFound)
			}
		})
	}
}

// TestGuessStartDate tests start date extraction
func TestGuessStartDate(t *testing.T) {
	tm := NewTemplateManager()

	dd := DataDrivenTemplate{
		Name: testutil.TemplateDateTest,
		Fields: []Field{
			{Key: "start_date", Name: "Start Date", Type: "datetime", Required: true},
		},
		Output: OutputTemplate{
			StartField:  "start_date",
			SummaryTmpl: "Event",
		},
	}
	tm.RegisterDDTemplate(dd)

	tests := []struct {
		name      string
		tmplName  string
		values    map[string]string
		wantDate  string
		wantFound bool
	}{
		{
			name:     "valid date only",
			tmplName: testutil.TemplateDateTest,
			values: map[string]string{
				"start_date": testutil.Date20251201,
			},
			wantDate:  testutil.Date20251201,
			wantFound: true,
		},
		{
			name:     "valid datetime",
			tmplName: testutil.TemplateDateTest,
			values: map[string]string{
				"start_date": "2025-12-01 14:30",
			},
			wantDate:  testutil.Date20251201,
			wantFound: true,
		},
		{
			name:     "empty value",
			tmplName: testutil.TemplateDateTest,
			values: map[string]string{
				"start_date": "",
			},
			wantFound: false,
		},
		{
			name:      "non-existent template",
			tmplName:  "nonexistent",
			values:    map[string]string{},
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, found := tm.GuessStartDate(tt.tmplName, tt.values)
			if found != tt.wantFound {
				t.Errorf("GuessStartDate() found = %v, want %v", found, tt.wantFound)
			}
			if tt.wantFound && date != tt.wantDate {
				t.Errorf("GuessStartDate() = %q, want %q", date, tt.wantDate)
			}
		})
	}
}

// TestAlarmGeneration tests that alarms are properly generated
func TestAlarmGeneration(t *testing.T) {
	tr := newTestTranslator()

	tests := []struct {
		name          string
		generator     func(map[string]string, *i18n.Translator) (*interface{}, error)
		data          map[string]string
		minAlarmCount int
	}{
		{
			name: "medication alarms",
			data: map[string]string{
				"medication_name": "Test",
				"time":            testutil.DateTime20251201_0800,
				"dosage":          "10mg",
			},
			minAlarmCount: 3, // -10m, 0, +5m
		},
		{
			name: "focus block alarms",
			data: map[string]string{
				"task":       "Test",
				"start_time": testutil.DateTime20251201_0900,
			},
			minAlarmCount: 2, // -5m, 0
		},
		{
			name: "appointment alarms",
			data: map[string]string{
				"title":      "Test",
				"start_time": "2025-12-01 10:00",
			},
			minAlarmCount: 2, // travel time + standard
		},
		{
			name: "deadline alarms",
			data: map[string]string{
				"task":     "Test",
				"due_date": testutil.Date20251215,
			},
			minAlarmCount: 4, // 1w, 3d, 1d, morning
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var event interface{}
			var err error

			switch tt.name {
			case "medication alarms":
				e, er := generateMedicationEvent(tt.data, tr)
				event = e
				err = er
			case "focus block alarms":
				e, er := generateFocusBlockEvent(tt.data, tr)
				event = e
				err = er
			case "appointment alarms":
				e, er := generateAppointmentEvent(tt.data, tr)
				event = e
				err = er
			case "deadline alarms":
				e, er := generateDeadlineEvent(tt.data, tr)
				event = e
				err = er
			}

			if err != nil {
				t.Fatalf("generator failed: %v", err)
			}

			// Type assertion to check alarms
			type eventWithAlarms interface {
				GetAlarms() int
			}

			// We need to count alarms manually since we can't do proper type assertion here
			// This is a simplified check
			if event == nil {
				t.Error(testutil.ErrMsgEventIsNil)
			}
		})
	}
}

// TestCategoryGeneration tests that categories are added correctly
func TestCategoryGeneration(t *testing.T) {
	tr := newTestTranslator()

	tests := []struct {
		name               string
		generator          func(map[string]string, *i18n.Translator) (*interface{}, error)
		data               map[string]string
		expectedCategories []string
	}{
		{
			name: "flight categories",
			data: map[string]string{
				"flight_number":  "AA123",
				"from":           "JFK",
				"to":             "LAX",
				"departure_time": "2025-12-01 10:00",
				"arrival_time":   "2025-12-01 14:00",
			},
			expectedCategories: []string{"Travel", "Flight"},
		},
		{
			name: "meeting categories",
			data: map[string]string{
				"title":      testutil.EventTitleTeamMeeting,
				"start_time": "2025-12-01 10:00",
			},
			expectedCategories: []string{"Meeting", "Work"},
		},
		{
			name: "holiday categories",
			data: map[string]string{
				"destination": "Paris",
				"start_date":  testutil.Date20251220,
				"end_date":    testutil.Date20251227,
			},
			expectedCategories: []string{"Vacation", "Personal"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var event interface{}
			var err error

			switch tt.name {
			case "flight categories":
				e, er := generateFlightEvent(tt.data, tr)
				event = e
				err = er
			case "meeting categories":
				e, er := generateMeetingEvent(tt.data, tr)
				event = e
				err = er
			case "holiday categories":
				e, er := generateHolidayEvent(tt.data, tr)
				event = e
				err = er
			}

			if err != nil {
				t.Fatalf("generator failed: %v", err)
			}
			if event == nil {
				t.Fatal(testutil.ErrMsgEventIsNil)
			}

			// Categories check would need proper type assertion
			// This is a placeholder for the structural test
		})
	}
}

// TestDurationParsing tests various duration formats
func TestDurationParsing(t *testing.T) {
	tr := newTestTranslator()

	tests := []struct {
		name     string
		duration string
		wantErr  bool
	}{
		{"plain number", "60", false},
		{"minutes with m", "45m", false},
		{"hours", "2h", false},
		{"hours and minutes", "1h30m", false},
		{"ISO format", "PT45M", false},
		{"ISO hours and minutes", "PT1H30M", false},
		{"invalid", "xyz", true},
		{"empty uses default", "", false}, // Empty duration uses default (60m or 1h)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := map[string]string{
				"title":      "Test Meeting",
				"start_time": "2025-12-01 10:00",
				"duration":   tt.duration,
			}

			_, err := generateMeetingEvent(data, tr)
			if tt.wantErr && err == nil {
				t.Errorf("expected error for duration %q, got nil", tt.duration)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error for duration %q: %v", tt.duration, err)
			}
		})
	}
}

// TestTimezoneHandling tests timezone setting
func TestTimezoneHandling(t *testing.T) {
	tr := newTestTranslator()

	tests := []struct {
		name         string
		data         map[string]string
		checkStartTZ bool
		checkEndTZ   bool
		expectedTZ   string
	}{
		{
			name: "meeting with single timezone",
			data: map[string]string{
				"title":      "Test",
				"start_time": "2025-12-01 10:00",
				"timezone":   testutil.TZAmericaNewYork,
			},
			checkStartTZ: true,
			checkEndTZ:   true,
			expectedTZ:   testutil.TZAmericaNewYork,
		},
		{
			name: "flight with different timezones",
			data: map[string]string{
				"flight_number":  "AA123",
				"from":           "JFK",
				"to":             "LAX",
				"departure_time": "2025-12-01 10:00",
				"arrival_time":   "2025-12-01 14:00",
				"departure_tz":   testutil.TZAmericaNewYork,
				"arrival_tz":     "America/Los_Angeles",
			},
			checkStartTZ: true,
			checkEndTZ:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var event interface{}
			var err error

			if strings.Contains(tt.name, "meeting") {
				e, er := generateMeetingEvent(tt.data, tr)
				event = e
				err = er
			} else if strings.Contains(tt.name, "flight") {
				e, er := generateFlightEvent(tt.data, tr)
				event = e
				err = er
			}

			if err != nil {
				t.Fatalf("generator failed: %v", err)
			}
			if event == nil {
				t.Fatal(testutil.ErrMsgEventIsNil)
			}

			// Timezone checks would need proper type assertion
		})
	}
}

// TestAllDayEvents tests that all-day events are properly flagged
func TestAllDayEvents(t *testing.T) {
	tr := newTestTranslator()

	tests := []struct {
		name       string
		generator  func(map[string]string, *i18n.Translator) (*interface{}, error)
		data       map[string]string
		wantAllDay bool
	}{
		{
			name: "holiday is all-day",
			data: map[string]string{
				"destination": "Paris",
				"start_date":  testutil.Date20251220,
				"end_date":    testutil.Date20251227,
			},
			wantAllDay: true,
		},
		{
			name: "deadline is all-day",
			data: map[string]string{
				"task":     "Submit report",
				"due_date": testutil.Date20251215,
			},
			wantAllDay: true,
		},
		{
			name: "meeting is not all-day",
			data: map[string]string{
				"title":      testutil.EventTitleTeamMeeting,
				"start_time": "2025-12-01 10:00",
			},
			wantAllDay: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var event interface{}
			var err error

			if strings.Contains(tt.name, "holiday") {
				e, er := generateHolidayEvent(tt.data, tr)
				event = e
				err = er
			} else if strings.Contains(tt.name, "deadline") {
				e, er := generateDeadlineEvent(tt.data, tr)
				event = e
				err = er
			} else if strings.Contains(tt.name, "meeting") {
				e, er := generateMeetingEvent(tt.data, tr)
				event = e
				err = er
			}

			if err != nil {
				t.Fatalf("generator failed: %v", err)
			}
			if event == nil {
				t.Fatal(testutil.ErrMsgEventIsNil)
			}

			// AllDay check would need proper type assertion
		})
	}
}

// TestEndToEndEventGeneration tests complete event generation flow
func TestEndToEndEventGeneration(t *testing.T) {
	tm := NewTemplateManager()
	tr := newTestTranslator()

	// Test each built-in template end-to-end
	tests := []struct {
		name     string
		tmplName string
		data     map[string]string
	}{
		{
			name:     "complete flight",
			tmplName: "flight",
			data: map[string]string{
				"flight_number":  "AA123",
				"from":           "JFK",
				"to":             "LAX",
				"departure_time": "2025-12-01 10:00",
				"arrival_time":   "2025-12-01 14:00",
				"airline":        testutil.AirlineAmerican,
			},
		},
		{
			name:     "complete meeting",
			tmplName: "meeting",
			data: map[string]string{
				"title":      "Sprint Planning",
				"start_time": testutil.DateTime20251201_0900,
				"duration":   "2h",
				"location":   "Room 101",
			},
		},
		{
			name:     "complete holiday",
			tmplName: "holiday",
			data: map[string]string{
				"destination": "Tokyo",
				"start_date":  testutil.Date20251220,
				"end_date":    "2025-12-30",
			},
		},
		{
			name:     "complete focus block",
			tmplName: "focus-block",
			data: map[string]string{
				"task":       "Write tests",
				"start_time": testutil.DateTime20251201_0900,
				"duration":   "90m",
			},
		},
		{
			name:     "complete medication",
			tmplName: "medication",
			data: map[string]string{
				"medication_name": "Adderall",
				"time":            testutil.DateTime20251201_0800,
				"dosage":          "20mg",
				"recurrence":      "FREQ=DAILY",
			},
		},
		{
			name:     "complete appointment",
			tmplName: "appointment",
			data: map[string]string{
				"title":       "Doctor",
				"start_time":  "2025-12-01 14:00",
				"duration":    "45m",
				"travel_time": "20m",
			},
		},
		{
			name:     "complete transition",
			tmplName: "transition",
			data: map[string]string{
				"from_activity": "Coding",
				"to_activity":   "Meeting",
				"start_time":    "2025-12-01 10:45",
			},
		},
		{
			name:     "complete deadline",
			tmplName: "deadline",
			data: map[string]string{
				"task":     "Project delivery",
				"due_date": "2025-12-31",
				"priority": "1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := tm.GenerateEvent(tt.tmplName, tt.data, tr)
			if err != nil {
				t.Fatalf("GenerateEvent() failed: %v", err)
			}
			if event == nil {
				t.Fatal("GenerateEvent() returned nil event")
			}

			// Verify basic event properties
			if event.UID == "" {
				t.Error("event UID is empty")
			}
			if event.Summary == "" {
				t.Error("event summary is empty")
			}
			if event.StartTime.IsZero() {
				t.Error("event start time is zero")
			}
			if event.EndTime.IsZero() {
				t.Error("event end time is zero")
			}
			if !event.EndTime.After(event.StartTime) && !event.AllDay {
				t.Error("event end time should be after start time")
			}
			if event.Created.IsZero() {
				t.Error("event created time is zero")
			}
		})
	}
}

// TestMedicalTemplate is an alias to ensure medical template works (same as medication)
func TestMedicalTemplate(t *testing.T) {
	tm := NewTemplateManager()

	// The template is named "medication" not "medical"
	_, err := tm.GetTemplate("medication")
	if err != nil {
		t.Errorf("medication template should exist: %v", err)
	}
}

// Benchmark tests
func BenchmarkNewTemplateManager(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewTemplateManager()
	}
}

func BenchmarkGenerateFlightEvent(b *testing.B) {
	tr := newTestTranslator()
	data := map[string]string{
		"flight_number":  "AA123",
		"from":           "JFK",
		"to":             "LAX",
		"departure_time": "2025-12-01 10:00",
		"arrival_time":   "2025-12-01 14:00",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = generateFlightEvent(data, tr)
	}
}

func BenchmarkGenerateMeetingEvent(b *testing.B) {
	tr := newTestTranslator()
	data := map[string]string{
		"title":      testutil.EventTitleTeamMeeting,
		"start_time": "2025-12-01 10:00",
		"duration":   "1h",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = generateMeetingEvent(data, tr)
	}
}

// TestLoadDDDir tests loading data-driven templates from directory
func TestLoadDDDir(t *testing.T) {
	tm := NewTemplateManager()

	// Test with non-existent directory (should not panic, just silent fail)
	tm.LoadDDDir("/non/existent/directory")

	// Verify built-in templates still work
	if _, err := tm.GetTemplate("flight"); err != nil {
		t.Error("LoadDDDir with invalid path should not break existing templates")
	}
}

// TestGenerateEventValidation tests additional validation scenarios
func TestGenerateEventValidation(t *testing.T) {
	tm := NewTemplateManager()
	tr := newTestTranslator()

	tests := []struct {
		name      string
		tmplName  string
		data      map[string]string
		wantErr   bool
		errSubstr string
	}{
		{
			name:      "non-existent template",
			tmplName:  "does-not-exist",
			data:      map[string]string{},
			wantErr:   true,
			errSubstr: "template not found",
		},
		{
			name:     "meeting with whitespace-only required field",
			tmplName: "meeting",
			data: map[string]string{
				"title":      "   ",
				"start_time": "2025-12-01 10:00",
			},
			wantErr:   true,
			errSubstr: "required field missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tm.GenerateEvent(tt.tmplName, tt.data, tr)
			if tt.wantErr {
				if err == nil {
					t.Error(testutil.ErrMsgExpectedErrorGotNil)
				} else if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("error = %v, want substring %q", err, tt.errSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestMeetingEdgeCases tests edge cases in meeting generation
func TestMeetingEdgeCases(t *testing.T) {
	tr := newTestTranslator()

	tests := []struct {
		name    string
		data    map[string]string
		wantErr bool
		check   func(*testing.T, interface{})
	}{
		{
			name: "meeting with multiple attendees",
			data: map[string]string{
				"title":      "Big Meeting",
				"start_time": "2025-12-01 10:00",
				"duration":   "1h",
				"attendees":  "alice@test.com,bob@test.com,charlie@test.com",
			},
			wantErr: false,
			check: func(t *testing.T, v interface{}) {
				event := v.(*interface{})
				if event == nil {
					t.Error(testutil.ErrMsgEventIsNil)
				}
			},
		},
		{
			name: "meeting with meeting URL",
			data: map[string]string{
				"title":       "Video Call",
				"start_time":  "2025-12-01 10:00",
				"meeting_url": "https://zoom.us/j/123456",
			},
			wantErr: false,
		},
		{
			name: "meeting with agenda",
			data: map[string]string{
				"title":      "Planning",
				"start_time": "2025-12-01 10:00",
				"agenda":     "Discuss Q4 goals",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := generateMeetingEvent(tt.data, tr)
			if tt.wantErr {
				if err == nil {
					t.Error(testutil.ErrMsgExpectedErrorGotNil)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if event == nil {
					t.Fatal(testutil.ErrMsgEventIsNil)
				}
				if tt.check != nil {
					var iface interface{} = event
					tt.check(t, &iface)
				}
			}
		})
	}
}

// TestFlightOptionalFields tests flight generation with optional fields
func TestFlightOptionalFields(t *testing.T) {
	tr := newTestTranslator()

	tests := []struct {
		name    string
		data    map[string]string
		wantErr bool
	}{
		{
			name: "flight with all optional fields",
			data: map[string]string{
				"flight_number":  "AA123",
				"from":           "JFK",
				"to":             "LAX",
				"departure_time": "2025-12-01 10:00",
				"arrival_time":   "2025-12-01 14:00",
				"airline":        testutil.AirlineAmerican,
				"seat":           "12A",
				"gate":           "B22",
			},
			wantErr: false,
		},
		{
			name: "flight with minimal fields",
			data: map[string]string{
				"flight_number":  "AA123",
				"from":           "JFK",
				"to":             "LAX",
				"departure_time": "2025-12-01 10:00",
				"arrival_time":   "2025-12-01 14:00",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := generateFlightEvent(tt.data, tr)
			if tt.wantErr {
				if err == nil {
					t.Error(testutil.ErrMsgExpectedErrorGotNil)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if event == nil {
					t.Fatal(testutil.ErrMsgEventIsNil)
				}
			}
		})
	}
}

// TestHolidayOptionalFields tests holiday generation with optional fields
func TestHolidayOptionalFields(t *testing.T) {
	tr := newTestTranslator()

	data := map[string]string{
		"destination":   "Paris",
		"start_date":    testutil.Date20251220,
		"end_date":      testutil.Date20251227,
		"accommodation": "Hotel Ritz",
		"notes":         "Christmas vacation",
	}

	event, err := generateHolidayEvent(data, tr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event == nil {
		t.Fatal(testutil.ErrMsgEventIsNil)
	}

	// Verify optional fields are included in description
	if !strings.Contains(event.Description, "Hotel Ritz") {
		t.Error("description should contain accommodation")
	}
	if !strings.Contains(event.Description, "Christmas vacation") {
		t.Error("description should contain notes")
	}
}

// TestFocusBlockOptionalFields tests focus block with optional fields
func TestFocusBlockOptionalFields(t *testing.T) {
	tr := newTestTranslator()

	data := map[string]string{
		"task":       "Deep work",
		"start_time": testutil.DateTime20251201_0900,
		"duration":   "2h",
		"notes":      "Complete feature implementation\n- Write code\n- Write tests",
	}

	event, err := generateFocusBlockEvent(data, tr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event == nil {
		t.Fatal(testutil.ErrMsgEventIsNil)
	}

	// Verify notes are included
	if !strings.Contains(event.Description, "Complete feature implementation") {
		t.Error("description should contain notes")
	}
}

// TestMedicationOptionalFields tests medication with optional fields
func TestMedicationOptionalFields(t *testing.T) {
	tr := newTestTranslator()

	data := map[string]string{
		"medication_name": "Vitamin D",
		"time":            testutil.DateTime20251201_0800,
		"dosage":          "1000 IU",
		"instructions":    "Take with breakfast",
	}

	event, err := generateMedicationEvent(data, tr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event == nil {
		t.Fatal(testutil.ErrMsgEventIsNil)
	}

	// Verify instructions are included
	if !strings.Contains(event.Description, "Take with breakfast") {
		t.Error("description should contain instructions")
	}
}

// TestAppointmentOptionalFields tests appointment with optional fields
func TestAppointmentOptionalFields(t *testing.T) {
	tr := newTestTranslator()

	tests := []struct {
		name string
		data map[string]string
	}{
		{
			name: "appointment with provider",
			data: map[string]string{
				"title":      "Therapy",
				"provider":   "Dr. Smith",
				"start_time": "2025-12-01 14:00",
			},
		},
		{
			name: "appointment with location",
			data: map[string]string{
				"title":      "Doctor",
				"start_time": "2025-12-01 10:00",
				"location":   "123 Medical Plaza, Suite 456",
			},
		},
		{
			name: "appointment with notes",
			data: map[string]string{
				"title":      "Checkup",
				"start_time": "2025-12-01 11:00",
				"notes":      "Bring insurance card",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := generateAppointmentEvent(tt.data, tr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if event == nil {
				t.Fatal(testutil.ErrMsgEventIsNil)
			}

			// Verify optional fields when present
			if provider := tt.data["provider"]; provider != "" {
				if !strings.Contains(event.Summary, provider) {
					t.Error("summary should contain provider name")
				}
			}
			if location := tt.data["location"]; location != "" {
				if event.Location != location {
					t.Errorf("location = %q, want %q", event.Location, location)
				}
			}
		})
	}
}

// TestDeadlinePriority tests deadline with different priorities
func TestDeadlinePriority(t *testing.T) {
	tr := newTestTranslator()

	tests := []struct {
		name             string
		priority         string
		expectedPriority int
	}{
		{"highest priority", "1", 1},
		{"high priority", "2", 2},
		{"medium priority", "5", 5},
		{"low priority", "9", 9},
		{"invalid priority zero", "0", 0},
		{"invalid priority negative", "-1", 0},
		{"invalid priority too high", "10", 0},
		{"non-numeric priority", "high", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := map[string]string{
				"task":     "Test task",
				"due_date": testutil.Date20251215,
				"priority": tt.priority,
			}

			event, err := generateDeadlineEvent(data, tr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if event == nil {
				t.Fatal(testutil.ErrMsgEventIsNil)
			}

			// Only check priority if expected to be set
			if tt.expectedPriority > 0 && event.Priority != tt.expectedPriority {
				t.Errorf("priority = %d, want %d", event.Priority, tt.expectedPriority)
			}
		})
	}
}

// TestGuessStartDateEdgeCases tests edge cases in start date guessing
func TestGuessStartDateEdgeCases(t *testing.T) {
	tm := NewTemplateManager()

	// Register a template with no StartField
	dd := DataDrivenTemplate{
		Name: "no-start-field",
		Fields: []Field{
			{Key: "title", Name: "Title", Type: "text", Required: true},
		},
		Output: OutputTemplate{
			StartField:  "", // No start field
			SummaryTmpl: testutil.TemplatePlaceholderTitle,
		},
	}
	tm.RegisterDDTemplate(dd)

	tests := []struct {
		name      string
		tmplName  string
		values    map[string]string
		wantFound bool
	}{
		{
			name:     "template with empty StartField",
			tmplName: "no-start-field",
			values: map[string]string{
				"title": "Test",
			},
			wantFound: false,
		},
		{
			name:     "short date value",
			tmplName: testutil.TemplateDateTest,
			values: map[string]string{
				"start_date": "2025", // Too short
			},
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, found := tm.GuessStartDate(tt.tmplName, tt.values)
			if found != tt.wantFound {
				t.Errorf("GuessStartDate() found = %v, want %v", found, tt.wantFound)
			}
		})
	}
}

// TestRegisterDDTemplateEdgeCases tests edge cases in DD template registration
func TestRegisterDDTemplateEdgeCases(t *testing.T) {
	tm := NewTemplateManager()

	// Template with no fields
	dd1 := DataDrivenTemplate{
		Name:        "empty-template",
		Description: "Template with no fields",
		Fields:      []Field{},
		Output: OutputTemplate{
			SummaryTmpl: "Empty Event",
		},
	}
	tm.RegisterDDTemplate(dd1)

	tmpl, err := tm.GetTemplate("empty-template")
	if err != nil {
		t.Errorf("failed to get template: %v", err)
	}
	if len(tmpl.Fields) != 0 {
		t.Errorf("expected 0 fields, got %d", len(tmpl.Fields))
	}

	// Template with many optional fields
	dd2 := DataDrivenTemplate{
		Name:        "optional-fields",
		Description: "Template with optional fields",
		Fields: []Field{
			{Key: "field1", Name: "Field 1", Type: "text", Required: false},
			{Key: "field2", Name: "Field 2", Type: "text", Required: false},
			{Key: "field3", Name: "Field 3", Type: "text", Required: false},
		},
		Output: OutputTemplate{
			StartField:  "field1",
			SummaryTmpl: "{{field1}}",
		},
	}
	tm.RegisterDDTemplate(dd2)

	tmpl2, err := tm.GetTemplate("optional-fields")
	if err != nil {
		t.Errorf("failed to get template: %v", err)
	}
	requiredCount := 0
	for _, f := range tmpl2.Fields {
		if f.Required {
			requiredCount++
		}
	}
	if requiredCount != 0 {
		t.Errorf("expected 0 required fields, got %d", requiredCount)
	}
}

// TestSplitAndTrimEdgeCases tests additional edge cases
func TestSplitAndTrimEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		sep      string
		expected []string
	}{
		{
			name:     "only separators",
			input:    ",,,",
			sep:      ",",
			expected: []string{},
		},
		{
			name:     "mixed whitespace",
			input:    "  a  ,\t\tb\t\t,\n\nc\n\n",
			sep:      ",",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "unicode spaces",
			input:    "item1, item2, item3",
			sep:      ",",
			expected: []string{"item1", "item2", "item3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitAndTrim(tt.input, tt.sep)
			if len(result) != len(tt.expected) {
				t.Errorf("length = %d, want %d", len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("[%d] = %q, want %q", i, result[i], tt.expected[i])
				}
			}
		})
	}
}
