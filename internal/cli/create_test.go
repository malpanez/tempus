package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"tempus/internal/calendar"
	"tempus/internal/testutil"
	"testing"
	"time"
)

func TestNewCreateCmd(t *testing.T) {
	app := TestApp()
	cmd := NewCreateCmd(app)
	if cmd == nil {
		t.Fatal("NewCreateCmd() returned nil")
	}
	if cmd.Use != "create [event-name]" {
		t.Errorf("Use = %q, want %q", cmd.Use, "create [event-name]")
	}
	if cmd.RunE == nil {
		t.Error("create command should have RunE function")
	}
}

func TestWriteCalendarOutput(t *testing.T) {
	tmpDir := t.TempDir()

	cal := calendar.NewCalendar()
	ev := calendar.Event{
		Summary:   "Test Event",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(1 * time.Hour),
	}
	cal.AddEvent(&ev)

	tests := []struct {
		name       string
		output     string
		wantErr    bool
		wantStdout bool
	}{
		{
			name:       "write to file",
			output:     filepath.Join(tmpDir, "test.ics"),
			wantErr:    false,
			wantStdout: false,
		},
		{
			name:       "write to stdout (empty output)",
			output:     "",
			wantErr:    false,
			wantStdout: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := TestApp()

			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := writeCalendarOutput(app, cal, tt.output)

			w.Close()
			var buf bytes.Buffer
			io.Copy(&buf, r)
			os.Stdout = oldStdout
			output := buf.String()

			if (err != nil) != tt.wantErr {
				t.Errorf("writeCalendarOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantStdout {
				if !strings.Contains(output, "BEGIN:VCALENDAR") {
					t.Errorf("writeCalendarOutput() did not print ICS to stdout")
				}
			} else {
				if _, err := os.Stat(tt.output); os.IsNotExist(err) {
					t.Errorf("writeCalendarOutput() did not create file %s", tt.output)
				}

				content, err := os.ReadFile(tt.output)
				if err != nil {
					t.Errorf("writeCalendarOutput() failed to read file: %v", err)
				}
				if !strings.Contains(string(content), "BEGIN:VCALENDAR") {
					t.Errorf("writeCalendarOutput() file is not valid ICS")
				}
			}
		})
	}
}

func TestParseEndTime(t *testing.T) {
	startTime := time.Date(2025, 12, 28, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		endStr  string
		want    time.Time
		wantErr bool
	}{
		{"parse duration 1h", "1h", startTime.Add(1 * time.Hour), false},
		{"parse duration 30m", "30m", startTime.Add(30 * time.Minute), false},
		{"parse duration 1h30m", "1h30m", startTime.Add(90 * time.Minute), false},
		{"parse absolute time", "2025-12-28 15:00", time.Date(2025, 12, 28, 15, 0, 0, 0, time.UTC), false},
		{"zero duration error", "0m", time.Time{}, true},
		{"negative duration error", "-30m", time.Time{}, true},
		{"invalid time format", "invalid", time.Time{}, true},
		{"invalid date format", "28-12-2025 15:00", time.Time{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseEndTime(startTime, tt.endStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseEndTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !got.Equal(tt.want) {
				t.Errorf("parseEndTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseDurationEnd(t *testing.T) {
	startTime := time.Date(2025, 12, 28, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		durStr  string
		want    time.Time
		wantErr bool
	}{
		{"parse 1 hour", "1h", startTime.Add(1 * time.Hour), false},
		{"parse 45 minutes", "45m", startTime.Add(45 * time.Minute), false},
		{"parse 2h30m", "2h30m", startTime.Add(150 * time.Minute), false},
		{"zero duration", "0h", time.Time{}, true},
		{"negative duration", "-1h", time.Time{}, true},
		{"invalid format", "not-a-duration", time.Time{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDurationEnd(startTime, tt.durStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDurationEnd() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !got.Equal(tt.want) {
				t.Errorf("parseDurationEnd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseTimedEventTimesBasic(t *testing.T) {
	tests := []struct {
		name     string
		startStr string
		durStr   string
		wantYear int
		wantMon  time.Month
		wantDay  int
		wantHour int
		wantMin  int
	}{
		{"standard format", "2025-12-16 14:00", "1h", 2025, time.December, 16, 14, 0},
		{"different duration", "2025-12-16 09:00", "30m", 2025, time.December, 16, 9, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, _, err := parseTimedEventTimes(tt.startStr, "", tt.durStr)
			if err != nil {
				t.Fatalf("parseTimedEventTimes(%q, \"\", %q) error: %v", tt.startStr, tt.durStr, err)
			}
			if start.Year() != tt.wantYear || start.Month() != tt.wantMon || start.Day() != tt.wantDay ||
				start.Hour() != tt.wantHour || start.Minute() != tt.wantMin {
				t.Errorf("parseTimedEventTimes(%q) start = %v, want %d-%02d-%02d %02d:%02d",
					tt.startStr, start, tt.wantYear, tt.wantMon, tt.wantDay, tt.wantHour, tt.wantMin)
			}
		})
	}
}

func TestParseAllDayTimesBasic(t *testing.T) {
	tests := []struct {
		name     string
		startStr string
		wantYear int
		wantMon  time.Month
		wantDay  int
	}{
		{"standard format", "2025-12-16", 2025, time.December, 16},
		{"january", "2025-01-05", 2025, time.January, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, _, err := parseAllDayTimes(tt.startStr, "")
			if err != nil {
				t.Fatalf("parseAllDayTimes(%q, \"\") error: %v", tt.startStr, err)
			}
			if start.Year() != tt.wantYear || start.Month() != tt.wantMon || start.Day() != tt.wantDay {
				t.Errorf("parseAllDayTimes(%q) start = %v, want %d-%02d-%02d",
					tt.startStr, start, tt.wantYear, tt.wantMon, tt.wantDay)
			}
		})
	}
}

func TestParseAllDayTimesEndDate(t *testing.T) {
	tests := []struct {
		name     string
		startStr string
		endStr   string
		wantErr  bool
	}{
		{"empty end → next day", "2025-12-16", "", false},
		{"valid end date", "2025-12-16", "2025-12-20", false},
		{"end before start", "2025-12-20", "2025-12-16", true},
		{"invalid start", "not-a-date", "", true},
		{"invalid end", "2025-12-16", "not-a-date", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := parseAllDayTimes(tt.startStr, tt.endStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseAllDayTimes(%q, %q) error = %v, wantErr %v", tt.startStr, tt.endStr, err, tt.wantErr)
			}
		})
	}
}

func TestParseTimedEventTimesEndStr(t *testing.T) {
	tests := []struct {
		name    string
		startStr string
		endStr   string
		durStr   string
		wantErr  bool
	}{
		{"end time specified", "2025-12-16 10:00", "2025-12-16 11:00", "", false},
		{"duration specified", "2025-12-16 10:00", "", "1h", false},
		{"no end defaults to 1h", "2025-12-16 10:00", "", "", false},
		{"invalid start", "not-a-time", "", "1h", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := parseTimedEventTimes(tt.startStr, tt.endStr, tt.durStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTimedEventTimes(%q, %q, %q) error = %v, wantErr %v",
					tt.startStr, tt.endStr, tt.durStr, err, tt.wantErr)
			}
		})
	}
}

func TestParseCreateTimes(t *testing.T) {
	tests := []struct {
		name    string
		opts    createOptions
		wantErr bool
	}{
		{
			name:    "timed event",
			opts:    createOptions{startStr: "2025-12-16 10:00", durStr: "1h"},
			wantErr: false,
		},
		{
			name:    "all-day event",
			opts:    createOptions{allDay: true, startStr: "2025-12-16"},
			wantErr: false,
		},
		{
			name:    "invalid timed",
			opts:    createOptions{startStr: "bad-input", durStr: "1h"},
			wantErr: true,
		},
		{
			name:    "invalid all-day",
			opts:    createOptions{allDay: true, startStr: "bad-input"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := parseCreateTimes(&tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseCreateTimes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNormalizeTimeInput(t *testing.T) {
	tests := []struct {
		name    string
		timeStr string
		wantNil bool
	}{
		{"empty string", "", false},
		{"already has date", "2025-12-16 10:00", false},
		{"clock time only", "10:00", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeTimeInput(tt.timeStr, "UTC", "")
			_ = result
		})
	}
}

func TestCreateCalendarWithEvent(t *testing.T) {
	opts := &createOptions{
		summary:  "Test Event",
		startTZ:  "UTC",
		location: "Room 1",
	}
	start := time.Date(2025, 12, 16, 10, 0, 0, 0, time.UTC)
	end := time.Date(2025, 12, 16, 11, 0, 0, 0, time.UTC)

	cal := createCalendarWithEvent(opts, start, end)
	if cal == nil {
		t.Fatal("createCalendarWithEvent() returned nil")
	}
	if len(cal.Events) != 1 {
		t.Errorf("expected 1 event, got %d", len(cal.Events))
	}
}

func TestConfigureEvent(t *testing.T) {
	event := calendar.NewEvent("Test", time.Now(), time.Now().Add(time.Hour))
	opts := &createOptions{
		location:    "Room A",
		description: "A test event",
		startTZ:     "UTC",
		endTZ:       "UTC",
		rrule:       "FREQ=DAILY;COUNT=5",
		categories:  []string{"Work", "Meeting"},
		attendees:   []string{"alice@example.com"},
		priority:    3,
	}
	configureEvent(event, opts)

	if event.Location != "Room A" {
		t.Errorf("Location = %q, want %q", event.Location, "Room A")
	}
	if event.Description != "A test event" {
		t.Errorf("Description = %q, want %q", event.Description, "A test event")
	}
	if event.RRule != "FREQ=DAILY;COUNT=5" {
		t.Errorf("RRule = %q, want %q", event.RRule, "FREQ=DAILY;COUNT=5")
	}
	if event.Priority != 3 {
		t.Errorf("Priority = %d, want 3", event.Priority)
	}
}

func TestSetEventTimezones(t *testing.T) {
	event := calendar.NewEvent("Test", time.Now(), time.Now().Add(time.Hour))
	setEventTimezones(event, "Europe/Madrid", "America/New_York")
	if event.StartTZ != "Europe/Madrid" {
		t.Errorf("StartTZ = %q, want %q", event.StartTZ, "Europe/Madrid")
	}
	if event.EndTZ != "America/New_York" {
		t.Errorf("EndTZ = %q, want %q", event.EndTZ, "America/New_York")
	}
}

func TestSetEventTimezonesEndInheritsStart(t *testing.T) {
	event := calendar.NewEvent("Test", time.Now(), time.Now().Add(time.Hour))
	setEventTimezones(event, "Europe/Madrid", "")
	if event.StartTZ != "Europe/Madrid" {
		t.Errorf("StartTZ = %q, want %q", event.StartTZ, "Europe/Madrid")
	}
	if event.EndTZ != "Europe/Madrid" {
		t.Errorf("EndTZ = %q, want %q (should inherit start)", event.EndTZ, "Europe/Madrid")
	}
}

func TestAddEventExDates(t *testing.T) {
	event := calendar.NewEvent("Test", time.Now(), time.Now().Add(time.Hour))
	exdates := []string{"2025-12-25", "2026-01-01", "invalid-date"}
	addEventExDates(event, exdates, "UTC", true)
	if len(event.ExDates) != 2 {
		t.Errorf("expected 2 valid exdates, got %d", len(event.ExDates))
	}
}

func TestAddEventExDatesEmpty(t *testing.T) {
	event := calendar.NewEvent("Test", time.Now(), time.Now().Add(time.Hour))
	addEventExDates(event, []string{}, "UTC", false)
	if len(event.ExDates) != 0 {
		t.Errorf("expected 0 exdates, got %d", len(event.ExDates))
	}
}

func TestAddEventAlarms(t *testing.T) {
	event := calendar.NewEvent("Test", time.Now(), time.Now().Add(time.Hour))
	addEventAlarms(event, []string{"-15m", "-1h"}, "UTC")
	if len(event.Alarms) != 2 {
		t.Errorf("expected 2 alarms, got %d", len(event.Alarms))
	}
}

func TestAddEventAlarmsEmpty(t *testing.T) {
	event := calendar.NewEvent("Test", time.Now(), time.Now().Add(time.Hour))
	addEventAlarms(event, []string{}, "UTC")
	if len(event.Alarms) != 0 {
		t.Errorf("expected 0 alarms, got %d", len(event.Alarms))
	}
}

func TestAddEventCategories(t *testing.T) {
	event := calendar.NewEvent("Test", time.Now(), time.Now().Add(time.Hour))
	addEventCategories(event, []string{"Work", "Meeting", ""})
	if len(event.Categories) != 2 {
		t.Errorf("expected 2 categories (empty filtered), got %d", len(event.Categories))
	}
}

func TestAddEventAttendees(t *testing.T) {
	event := calendar.NewEvent("Test", time.Now(), time.Now().Add(time.Hour))
	addEventAttendees(event, []string{"alice@example.com", "bob@example.com", " "})
	if len(event.Attendees) != 2 {
		t.Errorf("expected 2 attendees (whitespace filtered), got %d", len(event.Attendees))
	}
}

func TestNewVersionCmd(t *testing.T) {
	app := TestApp()
	cmd := NewVersionCmd(app, "1.0.0", "abc123", "2025-01-01")
	if cmd == nil {
		t.Fatal("NewVersionCmd() returned nil")
	}
	if cmd.Use != "version" {
		t.Errorf(testutil.ErrMsgUseMismatch, cmd.Use, "version")
	}
	if cmd.Run == nil {
		t.Error("version command should have Run function")
	}
}

func TestNewQuickCmd(t *testing.T) {
	app := TestApp()
	cmd := NewQuickCmd(app)
	if cmd == nil {
		t.Fatal("NewQuickCmd() returned nil")
	}
	if cmd.Use != "quick [natural language event description]" {
		t.Errorf("Use = %q, want quick command use string", cmd.Use)
	}

	outputFlag := cmd.Flags().Lookup("output")
	if outputFlag == nil {
		t.Error("quick command missing 'output' flag")
	}
	tzFlag := cmd.Flags().Lookup("timezone")
	if tzFlag == nil {
		t.Error("quick command missing 'timezone' flag")
	}
}
