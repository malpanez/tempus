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
				checkConflicts:  false,
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
				checkConflicts:  true,
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
				checkConflicts:  false,
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
				checkConflicts:  false,
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
				checkConflicts:  false,
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
				checkConflicts:  true,
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
				checkConflicts:  true,
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

// TestWriteCalendarOutput, TestParseEndTime, TestParseDurationEnd moved to internal/cli/create_test.go
