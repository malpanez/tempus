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
