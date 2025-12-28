package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tempus/internal/calendar"
	"tempus/internal/testutil"
)

// TestCollectBatchWarnings tests the collectBatchWarnings function
func TestCollectBatchWarnings(t *testing.T) {
	tz, _ := time.LoadLocation(testutil.TZEuropeMadrid)
	baseTime := time.Date(2025, 12, 28, 9, 0, 0, 0, tz)

	// Create overlapping events for conflict detection
	ev1 := calendar.Event{
		Summary:   "Meeting 1",
		StartTime: baseTime,
		EndTime:   baseTime.Add(1 * time.Hour),
	}
	ev2 := calendar.Event{
		Summary:   "Meeting 2",
		StartTime: baseTime.Add(30 * time.Minute), // Overlaps with ev1
		EndTime:   baseTime.Add(90 * time.Minute),
	}
	ev3 := calendar.Event{
		Summary:   "Meeting 3",
		StartTime: baseTime.Add(2 * time.Hour),
		EndTime:   baseTime.Add(3 * time.Hour),
	}

	tests := []struct {
		name          string
		events        []calendar.Event
		opts          *batchOptions
		wantConflicts bool
		wantOverwhelm bool
	}{
		{
			name:   "no warnings - check conflicts disabled",
			events: []calendar.Event{ev1, ev2},
			opts: &batchOptions{
				checkConflicts: false,
				maxEventsPerDay: 0,
				dryRun:          false,
			},
			wantConflicts: false,
			wantOverwhelm: false,
		},
		{
			name:   "detect conflicts when enabled",
			events: []calendar.Event{ev1, ev2},
			opts: &batchOptions{
				checkConflicts: true,
				maxEventsPerDay: 0,
				dryRun:          false,
			},
			wantConflicts: true,
			wantOverwhelm: false,
		},
		{
			name:   "detect conflicts in dry-run mode",
			events: []calendar.Event{ev1, ev2},
			opts: &batchOptions{
				checkConflicts: false,
				maxEventsPerDay: 0,
				dryRun:          true,
			},
			wantConflicts: true,
			wantOverwhelm: false,
		},
		{
			name:   "detect overwhelm when max events exceeded",
			events: []calendar.Event{ev1, ev2, ev3},
			opts: &batchOptions{
				checkConflicts: false,
				maxEventsPerDay: 2, // 3 events on same day
				dryRun:          false,
			},
			wantConflicts: false,
			wantOverwhelm: true,
		},
		{
			name:   "detect overwhelm in dry-run mode",
			events: []calendar.Event{ev3}, // Non-overlapping event
			opts: &batchOptions{
				checkConflicts: false,
				maxEventsPerDay: 0,
				dryRun:          true, // Auto-checks with default threshold
			},
			wantConflicts: false,
			wantOverwhelm: false, // Only 1 event, default threshold is 8
		},
		{
			name:   "both conflicts and overwhelm",
			events: []calendar.Event{ev1, ev2, ev3},
			opts: &batchOptions{
				checkConflicts: true,
				maxEventsPerDay: 2,
				dryRun:          false,
			},
			wantConflicts: true,
			wantOverwhelm: true,
		},
		{
			name:   "no events",
			events: []calendar.Event{},
			opts: &batchOptions{
				checkConflicts: true,
				maxEventsPerDay: 5,
				dryRun:          false,
			},
			wantConflicts: false,
			wantOverwhelm: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings := collectBatchWarnings(tt.events, tt.opts)

			hasConflicts := false
			hasOverwhelm := false
			for _, w := range warnings {
				if strings.Contains(w, "conflict") {
					hasConflicts = true
				}
				if strings.Contains(w, "event load") || strings.Contains(w, "Days with high") {
					hasOverwhelm = true
				}
			}

			if hasConflicts != tt.wantConflicts {
				t.Errorf("collectBatchWarnings() conflicts = %v, want %v", hasConflicts, tt.wantConflicts)
			}
			if hasOverwhelm != tt.wantOverwhelm {
				t.Errorf("collectBatchWarnings() overwhelm = %v, want %v", hasOverwhelm, tt.wantOverwhelm)
			}
		})
	}
}

// TestWriteBatchOutput tests the writeBatchOutput function
func TestWriteBatchOutput(t *testing.T) {
	tmpDir := t.TempDir()

	cal := calendar.NewCalendar()
	ev := calendar.Event{
		Summary:   "Test Event",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(1 * time.Hour),
	}
	cal.AddEvent(&ev)

	tests := []struct {
		name         string
		warnings     []string
		output       string
		eventCount   int
		wantErr      bool
		checkContent bool
	}{
		{
			name:         "write with no warnings",
			warnings:     []string{},
			output:       filepath.Join(tmpDir, "test1.ics"),
			eventCount:   1,
			wantErr:      false,
			checkContent: true,
		},
		{
			name:         "write with warnings",
			warnings:     []string{"⚠️  Warning 1", "⚠️  Warning 2"},
			output:       filepath.Join(tmpDir, "test2.ics"),
			eventCount:   1,
			wantErr:      false,
			checkContent: true,
		},
		{
			name:       "write to subdirectory",
			warnings:   []string{},
			output:     filepath.Join(tmpDir, "subdir", "test3.ics"),
			eventCount: 1,
			wantErr:    false,
		},
		{
			name:       "multiple events",
			warnings:   []string{},
			output:     filepath.Join(tmpDir, "test4.ics"),
			eventCount: 5,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := writeBatchOutput(cal, tt.warnings, tt.output, tt.eventCount)

			w.Close()
			var buf bytes.Buffer
			io.Copy(&buf, r)
			os.Stdout = oldStdout
			output := buf.String()

			if (err != nil) != tt.wantErr {
				t.Errorf("writeBatchOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Check file was created
				if _, err := os.Stat(tt.output); os.IsNotExist(err) {
					t.Errorf("writeBatchOutput() did not create file %s", tt.output)
				}

				// Check output message
				if !strings.Contains(output, "Created:") {
					t.Errorf("writeBatchOutput() output missing 'Created:' message")
				}

				// Check warnings were printed
				for _, warning := range tt.warnings {
					if !strings.Contains(output, warning) {
						t.Errorf("writeBatchOutput() output missing warning %q", warning)
					}
				}

				if tt.checkContent {
					// Check file content is valid ICS
					content, err := os.ReadFile(tt.output)
					if err != nil {
						t.Errorf("writeBatchOutput() failed to read created file: %v", err)
					}
					if !strings.Contains(string(content), "BEGIN:VCALENDAR") {
						t.Errorf("writeBatchOutput() created file is not valid ICS")
					}
				}
			}
		})
	}
}

// TestWriteCalendarOutput tests the writeCalendarOutput function
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
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := writeCalendarOutput(cal, tt.output)

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
				// Check ICS was printed to stdout
				if !strings.Contains(output, "BEGIN:VCALENDAR") {
					t.Errorf("writeCalendarOutput() did not print ICS to stdout")
				}
			} else {
				// Check file was created
				if _, err := os.Stat(tt.output); os.IsNotExist(err) {
					t.Errorf("writeCalendarOutput() did not create file %s", tt.output)
				}

				// Check success message
				if !strings.Contains(output, "Created:") {
					t.Errorf("writeCalendarOutput() output missing success message")
				}

				// Verify file content
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

// TestParseEndTime tests the parseEndTime function
func TestParseEndTime(t *testing.T) {
	startTime := time.Date(2025, 12, 28, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		endStr  string
		want    time.Time
		wantErr bool
	}{
		{
			name:    "parse duration 1h",
			endStr:  "1h",
			want:    startTime.Add(1 * time.Hour),
			wantErr: false,
		},
		{
			name:    "parse duration 30m",
			endStr:  "30m",
			want:    startTime.Add(30 * time.Minute),
			wantErr: false,
		},
		{
			name:    "parse duration 1h30m",
			endStr:  "1h30m",
			want:    startTime.Add(90 * time.Minute),
			wantErr: false,
		},
		{
			name:    "parse absolute time",
			endStr:  "2025-12-28 15:00",
			want:    time.Date(2025, 12, 28, 15, 0, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "zero duration error",
			endStr:  "0m",
			wantErr: true,
		},
		{
			name:    "negative duration error",
			endStr:  "-30m",
			wantErr: true,
		},
		{
			name:    "invalid time format",
			endStr:  "invalid",
			wantErr: true,
		},
		{
			name:    "invalid date format",
			endStr:  "28-12-2025 15:00",
			wantErr: true,
		},
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

// TestParseDurationEnd tests the parseDurationEnd function
func TestParseDurationEnd(t *testing.T) {
	startTime := time.Date(2025, 12, 28, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		durStr  string
		want    time.Time
		wantErr bool
	}{
		{
			name:    "parse 1 hour",
			durStr:  "1h",
			want:    startTime.Add(1 * time.Hour),
			wantErr: false,
		},
		{
			name:    "parse 45 minutes",
			durStr:  "45m",
			want:    startTime.Add(45 * time.Minute),
			wantErr: false,
		},
		{
			name:    "parse 2h30m",
			durStr:  "2h30m",
			want:    startTime.Add(150 * time.Minute),
			wantErr: false,
		},
		{
			name:    "zero duration",
			durStr:  "0h",
			wantErr: true,
		},
		{
			name:    "negative duration",
			durStr:  "-1h",
			wantErr: true,
		},
		{
			name:    "invalid format",
			durStr:  "not-a-duration",
			wantErr: true,
		},
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

