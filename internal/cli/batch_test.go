package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tempus/internal/calendar"
	"tempus/internal/nd"
	"tempus/internal/testutil"

	"github.com/spf13/cobra"
)

func TestDetectBatchFormat(t *testing.T) {
	tests := []struct {
		name    string
		flag    string
		path    string
		want    batchFormat
		wantErr bool
	}{
		{"auto csv", "auto", testutil.FilenameEventsCSV, batchFormatCSV, false},
		{"auto json", "auto", "events.json", batchFormatJSON, false},
		{"empty auto csv", "", testutil.FilenameEventsCSV, batchFormatCSV, false},
		{"empty auto json", "", "events.json", batchFormatJSON, false},
		{"explicit csv", "csv", testutil.FilenameEventsTXT, batchFormatCSV, false},
		{"explicit json", "json", testutil.FilenameEventsTXT, batchFormatJSON, false},
		{"CSV uppercase", "CSV", testutil.FilenameEventsTXT, batchFormatCSV, false},
		{"JSON uppercase", "JSON", testutil.FilenameEventsTXT, batchFormatJSON, false},
		{"auto unknown", "auto", testutil.FilenameEventsTXT, "", true},
		{"invalid format", "xml", testutil.FilenameEventsCSV, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := detectBatchFormat(tt.flag, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("detectBatchFormat(%q, %q) error = %v, wantErr %v", tt.flag, tt.path, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("detectBatchFormat(%q, %q) = %v, want %v", tt.flag, tt.path, got, tt.want)
			}
		})
	}
}

func TestLoadBatchFromCSV(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    int
		wantErr bool
	}{
		{
			name:    "valid csv",
			content: "summary,start,end\nEvent 1,2025-05-01 10:00,2025-05-01 11:00\nEvent 2,2025-05-02 14:00,2025-05-02 15:00",
			want:    2,
			wantErr: false,
		},
		{
			name:    testutil.TestNameEmptyFile,
			content: "",
			want:    0,
			wantErr: false,
		},
		{
			name:    "header only",
			content: "summary,start,end",
			want:    0,
			wantErr: false,
		},
		{
			name:    "with all_day",
			content: "summary,start,end,all_day\nEvent,2025-05-01,2025-05-02,true",
			want:    1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, testutil.FilenameTestCSV)
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatalf(testutil.ErrMsgFailedToWriteTestFile, err)
			}

			got, err := loadBatchFromCSV(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadBatchFromCSV() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("loadBatchFromCSV() returned %d records, want %d", len(got), tt.want)
			}
		})
	}

	t.Run("validates all_day parsing", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, testutil.FilenameTestCSV)
		content := "summary,start,all_day\nEvent1,2025-05-01,true\nEvent2,2025-05-02,false\nEvent3,2025-05-03,1"
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf(testutil.ErrMsgFailedToWriteTestFile, err)
		}

		records, err := loadBatchFromCSV(path)
		if err != nil {
			t.Fatalf("loadBatchFromCSV() error = %v", err)
		}
		if len(records) != 3 {
			t.Fatalf("expected 3 records, got %d", len(records))
		}
		if !records[0].AllDay {
			t.Errorf("record 0 AllDay = false, want true")
		}
		if records[1].AllDay {
			t.Errorf("record 1 AllDay = true, want false")
		}
		if !records[2].AllDay {
			t.Errorf("record 2 AllDay = false, want true")
		}
	})
}

func TestLoadBatchFromJSON(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    int
		wantErr bool
	}{
		{
			name:    "valid json",
			content: `[{"summary":"Event 1","start":"2025-05-01 10:00","end":"2025-05-01 11:00"}]`,
			want:    1,
			wantErr: false,
		},
		{
			name:    "empty array",
			content: `[]`,
			want:    0,
			wantErr: false,
		},
		{
			name:    testutil.TestNameEmptyFile,
			content: "",
			want:    0,
			wantErr: false,
		},
		{
			name:    "invalid json",
			content: `{invalid}`,
			want:    0,
			wantErr: true,
		},
		{
			name:    "with all_day bool",
			content: `[{"summary":"Event","start":"2025-05-01","all_day":true}]`,
			want:    1,
			wantErr: false,
		},
		{
			name:    "with all_day string",
			content: `[{"summary":"Event","start":"2025-05-01","all_day":"yes"}]`,
			want:    1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, testutil.FilenameTestJSON)
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatalf(testutil.ErrMsgFailedToWriteTestFile, err)
			}

			got, err := loadBatchFromJSON(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadBatchFromJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("loadBatchFromJSON() returned %d records, want %d", len(got), tt.want)
			}
		})
	}
}

func TestLoadBatchRecords(t *testing.T) {
	tmpDir := t.TempDir()

	csvPath := filepath.Join(tmpDir, testutil.FilenameTestCSV)
	csvContent := "summary,start,end\nCSV Event,2025-05-01 10:00,2025-05-01 11:00"
	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		t.Fatalf("failed to write CSV: %v", err)
	}

	jsonPath := filepath.Join(tmpDir, testutil.FilenameTestJSON)
	jsonContent := `[{"summary":"JSON Event","start":"2025-05-01 10:00","end":"2025-05-01 11:00"}]`
	if err := os.WriteFile(jsonPath, []byte(jsonContent), 0644); err != nil {
		t.Fatalf("failed to write JSON: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		format  batchFormat
		wantLen int
		wantErr bool
	}{
		{"csv", csvPath, batchFormatCSV, 1, false},
		{"json", jsonPath, batchFormatJSON, 1, false},
		{"unknown format", csvPath, "xml", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := loadBatchRecords(tt.path, tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadBatchRecords() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("loadBatchRecords() returned %d records, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestBuildEventFromBatch(t *testing.T) {
	tests := []struct {
		name       string
		record     batchRecord
		fallbackTZ string
		wantErr    bool
		checkFunc  func(*testing.T, *calendar.Event)
	}{
		{
			name: "basic event",
			record: batchRecord{
				Summary: testutil.EventTitleTestEvent,
				Start:   testutil.DateTime20250501_1000,
				End:     testutil.DateTime20250501_1100,
				StartTZ: testutil.TZEuropeMadrid,
			},
			fallbackTZ: "",
			wantErr:    false,
			checkFunc: func(t *testing.T, ev *calendar.Event) {
				if ev.Summary != testutil.EventTitleTestEvent {
					t.Errorf("Summary = %q, want %q", ev.Summary, testutil.EventTitleTestEvent)
				}
			},
		},
		{
			name: "missing summary",
			record: batchRecord{
				Start: testutil.DateTime20250501_1000,
			},
			wantErr: true,
		},
		{
			name: "missing start",
			record: batchRecord{
				Summary: "Test",
			},
			wantErr: true,
		},
		{
			name: "all day event",
			record: batchRecord{
				Summary: "All Day",
				Start:   testutil.Date20250501,
				AllDay:  true,
			},
			wantErr: false,
			checkFunc: func(t *testing.T, ev *calendar.Event) {
				if !ev.AllDay {
					t.Error("expected AllDay to be true")
				}
			},
		},
		{
			name: "with duration",
			record: batchRecord{
				Summary:  "Duration Event",
				Start:    testutil.DateTime20250501_1000,
				Duration: "90m",
				StartTZ:  "UTC",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, ev *calendar.Event) {
				duration := ev.EndTime.Sub(ev.StartTime)
				expected := 90 * time.Minute
				if duration != expected {
					t.Errorf("duration = %v, want %v", duration, expected)
				}
			},
		},
		{
			name: "with location and description",
			record: batchRecord{
				Summary:     "Detailed Event",
				Start:       testutil.DateTime20250501_1000,
				End:         testutil.DateTime20250501_1100,
				Location:    "Office",
				Description: testutil.DescriptionMeetingNotes,
			},
			wantErr: false,
			checkFunc: func(t *testing.T, ev *calendar.Event) {
				if ev.Location != "Office" {
					t.Errorf("Location = %q, want %q", ev.Location, "Office")
				}
				if ev.Description != testutil.DescriptionMeetingNotes {
					t.Errorf("Description = %q, want %q", ev.Description, testutil.DescriptionMeetingNotes)
				}
			},
		},
		{
			name: "invalid duration",
			record: batchRecord{
				Summary:  "Bad Duration",
				Start:    testutil.DateTime20250501_1000,
				Duration: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev, err := buildEventFromBatch(tt.record, tt.fallbackTZ, nd.NewSpellCheckCache(nil), nd.NewCategoryCache())
			if (err != nil) != tt.wantErr {
				t.Errorf("buildEventFromBatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, ev)
			}
		})
	}
}

func TestLoadBatchFromCSVWithDelimitedFields(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, testutil.FilenameTestCSV)
	content := `summary,start,end,exdate,categories,alarms
Event,2025-05-01 10:00,2025-05-01 11:00,"2025-05-03,2025-05-04","work,urgent","15m,30m"`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf(testutil.ErrMsgFailedToWriteTestFile, err)
	}

	records, err := loadBatchFromCSV(path)
	if err != nil {
		t.Fatalf("loadBatchFromCSV() error = %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	rec := records[0]
	if len(rec.ExDates) != 2 {
		t.Errorf("expected 2 exdates, got %d: %v", len(rec.ExDates), rec.ExDates)
	}
	if len(rec.Categories) != 2 {
		t.Errorf("expected 2 categories, got %d: %v", len(rec.Categories), rec.Categories)
	}
	if len(rec.Alarms) != 2 {
		t.Errorf("expected 2 alarms, got %d: %v", len(rec.Alarms), rec.Alarms)
	}
}

func TestLoadBatchFromJSONWithComplexTypes(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, testutil.FilenameTestJSON)

	data := []map[string]interface{}{
		{
			"summary":    testutil.EventTitleTestEvent,
			"start":      testutil.DateTime20250501_1000,
			"end":        testutil.DateTime20250501_1100,
			"all_day":    false,
			"exdate":     []interface{}{testutil.Date20250503, "2025-05-04"},
			"categories": []interface{}{"work", "urgent"},
			"alarms":     []interface{}{"15m", "30m"},
		},
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}

	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		t.Fatalf(testutil.ErrMsgFailedToWriteTestFile, err)
	}

	records, err := loadBatchFromJSON(path)
	if err != nil {
		t.Fatalf("loadBatchFromJSON() error = %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	rec := records[0]
	if len(rec.ExDates) != 2 {
		t.Errorf("expected 2 exdates, got %d: %v", len(rec.ExDates), rec.ExDates)
	}
	if len(rec.Categories) != 2 {
		t.Errorf("expected 2 categories, got %d: %v", len(rec.Categories), rec.Categories)
	}
	if len(rec.Alarms) != 2 {
		t.Errorf("expected 2 alarms, got %d: %v", len(rec.Alarms), rec.Alarms)
	}
}

func TestBuildEventFromBatchWithCategories(t *testing.T) {
	rec := batchRecord{
		Summary:    "Categorized Event",
		Start:      testutil.DateTime20250501_1000,
		End:        testutil.DateTime20250501_1100,
		Categories: []string{"work", "urgent", "meeting"},
	}

	ev, err := buildEventFromBatch(rec, "", nd.NewSpellCheckCache(nil), nd.NewCategoryCache())
	if err != nil {
		t.Fatalf(testutil.ErrMsgBuildEventFromBatchError, err)
	}

	if len(ev.Categories) != 3 {
		t.Errorf("expected 3 categories, got %d", len(ev.Categories))
	}
}

func TestBuildEventFromBatchWithRRule(t *testing.T) {
	rec := batchRecord{
		Summary: "Recurring Event",
		Start:   testutil.DateTime20250501_1000,
		End:     testutil.DateTime20250501_1100,
		RRule:   testutil.RRuleDaily5Count,
	}

	ev, err := buildEventFromBatch(rec, "", nd.NewSpellCheckCache(nil), nd.NewCategoryCache())
	if err != nil {
		t.Fatalf(testutil.ErrMsgBuildEventFromBatchError, err)
	}

	if ev.RRule != testutil.RRuleDaily5Count {
		t.Errorf("RRule = %q, want %q", ev.RRule, testutil.RRuleDaily5Count)
	}
}

func TestBuildEventFromBatchAllDayEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		record  batchRecord
		wantErr bool
	}{
		{
			name: "all day with end date",
			record: batchRecord{
				Summary: "Multi-day Event",
				Start:   testutil.Date20250501,
				End:     testutil.Date20250503,
				AllDay:  true,
			},
			wantErr: false,
		},
		{
			name: "all day end before start",
			record: batchRecord{
				Summary: "Invalid Range",
				Start:   testutil.Date20250503,
				End:     testutil.Date20250501,
				AllDay:  true,
			},
			wantErr: true,
		},
		{
			name: "all day with time component in start",
			record: batchRecord{
				Summary: "All Day with Time",
				Start:   testutil.DateTime20250501_1000,
				AllDay:  true,
			},
			wantErr: false,
		},
		{
			name: "clock only time",
			record: batchRecord{
				Summary: "Clock Time",
				Start:   "14:30",
				StartTZ: testutil.TZEuropeMadrid,
			},
			wantErr: false,
		},
		{
			name: "end time as duration string",
			record: batchRecord{
				Summary: "Duration in End",
				Start:   testutil.DateTime20250501_1000,
				End:     "1h30m",
			},
			wantErr: false,
		},
		{
			name: "end time before start time",
			record: batchRecord{
				Summary: "Invalid Time Range",
				Start:   testutil.DateTime20250501_1400,
				End:     testutil.DateTime20250501_1000,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := buildEventFromBatch(tt.record, "", nd.NewSpellCheckCache(nil), nd.NewCategoryCache())
			if (err != nil) != tt.wantErr {
				t.Errorf(testutil.ErrMsgBuildEventFromBatchError+", wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBuildEventFromBatchWithExDatesAndAlarms(t *testing.T) {
	rec := batchRecord{
		Summary: "Event with ExDates and Alarms",
		Start:   testutil.DateTime20250501_1000,
		End:     testutil.DateTime20250501_1100,
		StartTZ: testutil.TZEuropeMadrid,
		RRule:   testutil.RRuleDaily5Count,
		ExDates: []string{"2025-05-03 10:00", "2025-05-04 10:00"},
		Alarms:  []string{"15m", "30m"},
	}

	ev, err := buildEventFromBatch(rec, "", nd.NewSpellCheckCache(nil), nd.NewCategoryCache())
	if err != nil {
		t.Fatalf(testutil.ErrMsgBuildEventFromBatchError, err)
	}

	if len(ev.ExDates) != 2 {
		t.Errorf("expected 2 exdates, got %d", len(ev.ExDates))
	}

	if len(ev.Alarms) != 2 {
		t.Errorf("expected 2 alarms, got %d", len(ev.Alarms))
	}
}

func TestBatchTemplateFormatFlag(t *testing.T) {
	csvContent, err := getBatchTemplateContent("school-event", "csv")
	if err != nil {
		t.Fatalf("getBatchTemplateContent(school-event, csv) error: %v", err)
	}
	if csvContent == "" {
		t.Fatal("getBatchTemplateContent(school-event, csv) returned empty")
	}

	yamlContent, err := getBatchTemplateContent("school-event", "yaml")
	if err != nil {
		t.Fatalf("getBatchTemplateContent(school-event, yaml) error: %v", err)
	}
	if yamlContent == "" {
		t.Fatal("getBatchTemplateContent(school-event, yaml) returned empty")
	}

	if csvContent == yamlContent {
		t.Error("CSV and YAML content should differ")
	}
}

func TestRunBatchTemplateWithFormatFlag(t *testing.T) {
	tmpDir := t.TempDir()
	app := TestApp()

	tests := []struct {
		name     string
		tmpl     string
		format   string
		contains string
	}{
		{"school-event-csv", "school-event", "csv", "summary,start_date"},
		{"school-event-yaml", "school-event", "yaml", "summary:"},
		{"recruiter-meeting-csv", "recruiter-meeting", "csv", "company,role"},
		{"travel-day-yaml", "travel-day", "yaml", "destination_timezone:"},
		{"basic-default", "basic", "csv", "summary,start"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outPath := filepath.Join(tmpDir, tt.name+".out")
			cmd := NewBatchTemplateCmd(app)
			cmd.SetArgs([]string{tt.tmpl})
			mustSetFlag(t, cmd, "output", outPath)
			mustSetFlag(t, cmd, "format", tt.format)

			if err := cmd.Execute(); err != nil {
				t.Fatalf("execute failed: %v", err)
			}

			data, err := os.ReadFile(outPath)
			if err != nil {
				t.Fatalf("failed to read output: %v", err)
			}
			if !strings.Contains(string(data), tt.contains) {
				t.Errorf("output missing %q:\n%s", tt.contains, string(data))
			}
		})
	}
}

func TestRunBatchTemplateInvalidFormat(t *testing.T) {
	app := TestApp()
	cmd := NewBatchTemplateCmd(app)
	cmd.SetArgs([]string{"basic"})
	mustSetFlag(t, cmd, "output", filepath.Join(t.TempDir(), "out.csv"))
	mustSetFlag(t, cmd, "format", "xml")

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
	if !strings.Contains(err.Error(), "--format must be") {
		t.Errorf("expected format error, got: %v", err)
	}
}

func TestBatchTemplateFormatInvalid(t *testing.T) {
	_, err := getBatchTemplateContent("unknown-type", "csv")
	if err == nil {
		t.Fatal("expected error for unknown template type")
	}
	if !strings.Contains(err.Error(), "unknown template type") {
		t.Errorf("expected 'unknown template type' in error, got: %v", err)
	}
}

func TestCollectBatchWarnings(t *testing.T) {
	tz, _ := time.LoadLocation(testutil.TZEuropeMadrid)
	baseTime := time.Date(2025, 12, 28, 9, 0, 0, 0, tz)

	ev1 := calendar.Event{
		Summary:   "Meeting 1",
		StartTime: baseTime,
		EndTime:   baseTime.Add(1 * time.Hour),
	}
	ev2 := calendar.Event{
		Summary:   "Meeting 2",
		StartTime: baseTime.Add(30 * time.Minute),
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
				maxEventsPerDay: 2,
				dryRun:          false,
			},
			wantConflicts: false,
			wantOverwhelm: true,
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
			app := TestApp()
			var buf bytes.Buffer
			app.Stdout = &buf

			err := writeBatchOutput(app, cal, tt.warnings, tt.output, tt.eventCount)
			output := buf.String()

			if (err != nil) != tt.wantErr {
				t.Errorf("writeBatchOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if _, err := os.Stat(tt.output); os.IsNotExist(err) {
					t.Errorf("writeBatchOutput() did not create file %s", tt.output)
				}

				if !strings.Contains(output, "Created:") {
					t.Errorf("writeBatchOutput() output missing 'Created:' message")
				}

				for _, warning := range tt.warnings {
					if !strings.Contains(output, warning) {
						t.Errorf("writeBatchOutput() output missing warning %q", warning)
					}
				}

				if tt.checkContent {
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

func TestBatchCSVGeneratesCalendarWithMultipleEvents(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, testutil.FilenameEventsCSV)
	outputPath := filepath.Join(tmpDir, "batch.ics")

	csvData := strings.Join([]string{
		"summary,start,end,start_tz,end_tz,location,description,all_day,duration,rrule,exdate,alarms",
		`"Daily Standup","2025-05-01 09:00","2025-05-01 09:15","Europe/Madrid","","Zoom link","Team sync",false,,"FREQ=DAILY;COUNT=5","2025-05-03 09:00","15m"`,
		`"All Hands","2025-05-04 10:00","2025-05-04 11:30","Europe/Madrid","","Auditorium","Company-wide update",false,,"","","trigger=+5m,description=Wrap up"`,
	}, "\n")

	if err := os.WriteFile(inputPath, []byte(csvData), 0644); err != nil {
		t.Fatalf("failed to write csv: %v", err)
	}

	app := TestApp()
	cmd := NewBatchCmd(app)
	mustSetFlag(t, cmd, "input", inputPath)
	mustSetFlag(t, cmd, "output", outputPath)
	mustSetFlag(t, cmd, "format", "csv")
	mustSetFlag(t, cmd, "name", "Team Events")

	if err := runBatch(app, cmd, nil); err != nil {
		t.Fatalf("runBatch returned error: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	ics := string(data)

	if !strings.Contains(ics, "SUMMARY:Daily Standup") || !strings.Contains(ics, "SUMMARY:All Hands") {
		t.Fatalf("expected both event summaries in ICS:\n%s", ics)
	}
	if !strings.Contains(ics, "RRULE:FREQ=DAILY;COUNT=5") {
		t.Fatalf("expected RRULE block:\n%s", ics)
	}
	if !strings.Contains(ics, "EXDATE;TZID=Europe/Madrid:20250503T090000") {
		t.Fatalf("expected EXDATE block:\n%s", ics)
	}
	if !strings.Contains(ics, "TRIGGER:-PT15M") || !strings.Contains(ics, "TRIGGER:PT5M") {
		t.Fatalf("expected alarm triggers in ICS:\n%s", ics)
	}
	if !strings.Contains(ics, "DESCRIPTION:Wrap up") {
		t.Fatalf("expected custom alarm description in ICS:\n%s", ics)
	}
	if !strings.Contains(ics, "X-WR-CALNAME:Team Events") {
		t.Fatalf("expected calendar name header:\n%s", ics)
	}
}

func TestBatchJSONSupportsAllDayAndDuration(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "events.json")
	outputPath := filepath.Join(tmpDir, "batch.ics")

	jsonData := `[
		{
			"summary": "Conference Day",
			"start": "2025-09-10",
			"end": "2025-09-12",
			"all_day": true,
			"start_tz": "Europe/Dublin",
			"location": "Dublin Convention Centre",
			"description": "Annual conference"
		},
		{
			"summary": "Retro",
			"start": "2025-09-15 16:00",
			"duration": "45m",
			"start_tz": "Europe/Dublin",
			"rrule": "FREQ=WEEKLY;COUNT=4",
			"exdate": ["2025-09-29 16:00"],
			"alarms": ["20m", "trigger=+5m,description=Team wrap"]
		}
	]`

	if err := os.WriteFile(inputPath, []byte(jsonData), 0644); err != nil {
		t.Fatalf("failed to write json: %v", err)
	}

	app := TestApp()
	cmd := NewBatchCmd(app)
	mustSetFlag(t, cmd, "input", inputPath)
	mustSetFlag(t, cmd, "output", outputPath)
	mustSetFlag(t, cmd, "format", "json")
	mustSetFlag(t, cmd, "name", "Autumn Plan")

	if err := runBatch(app, cmd, nil); err != nil {
		t.Fatalf("runBatch returned error: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	ics := string(data)

	if !strings.Contains(ics, "SUMMARY:Conference Day") || !strings.Contains(ics, "SUMMARY:Retro") {
		t.Fatalf("expected both event summaries in ICS:\n%s", ics)
	}
	if !strings.Contains(ics, "DTSTART;VALUE=DATE:20250910") || !strings.Contains(ics, "DTEND;VALUE=DATE:20250913") {
		t.Fatalf("expected all-day date range:\n%s", ics)
	}
	if !strings.Contains(ics, "DTSTART;TZID=Europe/Dublin:20250915T160000") {
		t.Fatalf("expected timezone-aware DTSTART:\n%s", ics)
	}
	if !strings.Contains(ics, "RRULE:FREQ=WEEKLY;COUNT=4") {
		t.Fatalf("expected RRULE block:\n%s", ics)
	}
	if !strings.Contains(ics, "EXDATE;TZID=Europe/Dublin:20250929T160000") {
		t.Fatalf("expected EXDATE block:\n%s", ics)
	}
	if !strings.Contains(ics, "TRIGGER:-PT20M") || !strings.Contains(ics, "TRIGGER:PT5M") {
		t.Fatalf("expected JSON alarms to be rendered:\n%s", ics)
	}
	if !strings.Contains(ics, "DESCRIPTION:Team wrap") {
		t.Fatalf("expected custom JSON alarm description:\n%s", ics)
	}
}

func TestExpandAlarmProfilesWithError(t *testing.T) {
	tests := []struct {
		name      string
		specs     []string
		wantErr   bool
		errSubstr string
	}{
		{"empty input", []string{}, false, ""},
		{"non-profile spec passthrough", []string{"-15m", "+5m"}, false, ""},
		{"unknown profile", []string{"profile:nonexistent"}, true, "not found"},
		{"unknown profile lists available", []string{"profile:nonexistent"}, true, "Available:"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := expandAlarmProfilesWithError(tt.specs)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expandAlarmProfilesWithError(%v) expected error, got nil", tt.specs)
				} else if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("expandAlarmProfilesWithError(%v) error = %q, want substring %q", tt.specs, err.Error(), tt.errSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("expandAlarmProfilesWithError(%v) unexpected error: %v", tt.specs, err)
				}
				_ = result
			}
		})
	}
}

func TestCSVValue(t *testing.T) {
	row := []string{"value1", "value2", "value3"}
	index := map[string]int{
		"col1": 0,
		"col2": 1,
		"col3": 2,
	}

	tests := []struct {
		name  string
		key   string
		want  string
	}{
		{"existing key", "col1", "value1"},
		{"second key", "col2", "value2"},
		{"third key", "col3", "value3"},
		{"non-existing key", "col4", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := csvValue(row, index, tt.key)
			if got != tt.want {
				t.Errorf("csvValue(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func newBatchCmdForTest() *cobra.Command {
	return NewBatchCmd(TestApp())
}

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
			got, err := getBatchTemplateContent(tt.templateKey, "csv")
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

func TestSchoolEventTemplateCSV(t *testing.T) {
	content := getSchoolEventTemplateCSV()
	if content == "" {
		t.Fatal("getSchoolEventTemplateCSV() returned empty string")
	}
	requiredHeaders := []string{"summary", "start_date", "end_date", "category", "location", "alarm", "notes"}
	for _, h := range requiredHeaders {
		if !strings.Contains(content, h) {
			t.Errorf("getSchoolEventTemplateCSV() missing header %q", h)
		}
	}
	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) < 4 {
		t.Errorf("getSchoolEventTemplateCSV() expected at least 3 example rows + header, got %d lines", len(lines))
	}
}

func TestSchoolEventTemplateYAML(t *testing.T) {
	content := getSchoolEventTemplateYAML()
	if content == "" {
		t.Fatal("getSchoolEventTemplateYAML() returned empty string")
	}
	for _, key := range []string{"summary:", "start_date:"} {
		if !strings.Contains(content, key) {
			t.Errorf("getSchoolEventTemplateYAML() missing key %q", key)
		}
	}
}

func TestRecruiterMeetingTemplateCSV(t *testing.T) {
	content := getRecruiterMeetingTemplateCSV()
	if content == "" {
		t.Fatal("getRecruiterMeetingTemplateCSV() returned empty string")
	}
	requiredHeaders := []string{"summary", "start_date", "time", "duration", "timezone", "alarm", "add_prep_time", "company", "role", "recruiter", "notes"}
	for _, h := range requiredHeaders {
		if !strings.Contains(content, h) {
			t.Errorf("getRecruiterMeetingTemplateCSV() missing header %q", h)
		}
	}
	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) < 2 {
		t.Errorf("getRecruiterMeetingTemplateCSV() expected at least 1 example row + header, got %d lines", len(lines))
	}
}

func TestRecruiterMeetingTemplateYAML(t *testing.T) {
	content := getRecruiterMeetingTemplateYAML()
	if content == "" {
		t.Fatal("getRecruiterMeetingTemplateYAML() returned empty string")
	}
	for _, key := range []string{"summary:", "company:"} {
		if !strings.Contains(content, key) {
			t.Errorf("getRecruiterMeetingTemplateYAML() missing key %q", key)
		}
	}
}

func TestTravelDayTemplateCSV(t *testing.T) {
	content := getTravelDayTemplateCSV()
	if content == "" {
		t.Fatal("getTravelDayTemplateCSV() returned empty string")
	}
	requiredHeaders := []string{"summary", "start_date", "time", "end_time", "timezone", "destination_timezone", "category", "location", "add_prep_time", "alarm", "notes"}
	for _, h := range requiredHeaders {
		if !strings.Contains(content, h) {
			t.Errorf("getTravelDayTemplateCSV() missing header %q", h)
		}
	}
	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) < 5 {
		t.Errorf("getTravelDayTemplateCSV() expected at least 4 example rows + header, got %d lines", len(lines))
	}
}

func TestTravelDayTemplateYAML(t *testing.T) {
	content := getTravelDayTemplateYAML()
	if content == "" {
		t.Fatal("getTravelDayTemplateYAML() returned empty string")
	}
	for _, key := range []string{"summary:", "destination_timezone:"} {
		if !strings.Contains(content, key) {
			t.Errorf("getTravelDayTemplateYAML() missing key %q", key)
		}
	}
}
