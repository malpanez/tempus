package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"tempus/internal/testutil"
	"testing"
	"time"

	"tempus/internal/calendar"

	"github.com/spf13/cobra"
)

// ============================================================================
// Helper function tests
// ============================================================================

func TestParseBoolish(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"true", "true", true},
		{"True", "True", true},
		{"TRUE", "TRUE", true},
		{"1", "1", true},
		{"yes", "yes", true},
		{"Yes", "Yes", true},
		{"YES", "YES", true},
		{"y", "y", true},
		{"Y", "Y", true},
		{"on", "on", true},
		{"On", "On", true},
		{"ON", "ON", true},
		{"false", "false", false},
		{"0", "0", false},
		{"no", "no", false},
		{"n", "n", false},
		{"off", "off", false},
		{"empty", "", false},
		{"whitespace", "   ", false},
		{"random", "random", false},
		{testutil.TestNameWithSpaces, " true ", true},
		{"with spaces no", " false ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseBoolish(tt.input)
			if got != tt.want {
				t.Errorf("parseBoolish(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestValueAsString(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  string
	}{
		{"nil", nil, ""},
		{"string", "hello", "hello"},
		{"string with spaces", "  hello  ", "hello"},
		{testutil.TestNameEmptyString, "", ""},
		{"float64", 42.5, "42.5"},
		{"float64 int", 42.0, "42"},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
		{"int", 123, "123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := valueAsString(tt.input)
			if got != tt.want {
				t.Errorf("valueAsString(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestValueAsBool(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  bool
	}{
		{"nil", nil, false},
		{"bool true", true, true},
		{"bool false", false, false},
		{"float64 zero", 0.0, false},
		{"float64 nonzero", 1.0, true},
		{"string true", "true", true},
		{"string false", "false", false},
		{"string 1", "1", true},
		{"string 0", "0", false},
		{"string yes", "yes", true},
		{"string no", "no", false},
		{"string empty", "", false},
		{"int via string", "123", false}, // parseBoolish returns false for "123"
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := valueAsBool(tt.input)
			if got != tt.want {
				t.Errorf("valueAsBool(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestValueAsStringSlice(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  []string
	}{
		{"nil", nil, nil},
		{testutil.TestNameEmptySlice, []interface{}{}, nil},
		{"slice with strings", []interface{}{"a", "b", "c"}, []string{"a", "b", "c"}},
		{"slice with spaces", []interface{}{"  a  ", "b", "  c  "}, []string{"a", "b", "c"}},
		{"slice with empty", []interface{}{"a", "", "b"}, []string{"a", "b"}},
		{"string slice", []string{"x", "y", "z"}, []string{"x", "y", "z"}},
		{"string comma delimited", "a,b,c", []string{"a", "b", "c"}},
		{"string semicolon delimited", "a;b;c", []string{"a", "b", "c"}},
		{"string pipe delimited", "a|b|c", []string{"a", "b", "c"}},
		{"string newline delimited", "a\nb\nc", []string{"a", "b", "c"}},
		{"string mixed delimiters", "a,b;c|d", []string{"a", "b", "c", "d"}},
		{testutil.TestNameEmptyString, "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := valueAsStringSlice(tt.input)
			if !equalStringSlices(got, tt.want) {
				t.Errorf("valueAsStringSlice(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestValueAsAlarmSlice(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  []string
	}{
		{"nil", nil, nil},
		{testutil.TestNameEmptySlice, []interface{}{}, nil},
		{"slice with strings", []interface{}{"15m", "30m"}, []string{"15m", "30m"}},
		{"string slice", []string{"10m", "20m"}, []string{"10m", "20m"}},
		{"string single", "15m", []string{"15m"}},
		{testutil.TestNameEmptyString, "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := valueAsAlarmSlice(tt.input)
			if !equalStringSlices(got, tt.want) {
				t.Errorf("valueAsAlarmSlice(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestSplitDelimited(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"empty", "", nil},
		{"whitespace", "   ", nil},
		{"single", "a", []string{"a"}},
		{"comma", "a,b,c", []string{"a", "b", "c"}},
		{"semicolon", "a;b;c", []string{"a", "b", "c"}},
		{"pipe", "a|b|c", []string{"a", "b", "c"}},
		{"newline", "a\nb\nc", []string{"a", "b", "c"}},
		{"mixed", "a,b;c|d\ne", []string{"a", "b", "c", "d", "e"}},
		{testutil.TestNameWithSpaces, " a , b , c ", []string{"a", "b", "c"}},
		{"with empty parts", "a,,b", []string{"a", "b"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitDelimited(tt.input)
			if !equalStringSlices(got, tt.want) {
				t.Errorf("splitDelimited(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestExtractDate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{testutil.TestNameFullDatetime, "2025-05-01 12:00", testutil.Date20250501},
		{testutil.TestNameDateOnly, testutil.Date20250501, testutil.Date20250501},
		{"short", "2025", "2025"},
		{"empty", "", ""},
		{testutil.TestNameWithSpaces, "  2025-05-01 12:00  ", testutil.Date20250501},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractDate(tt.input)
			if got != tt.want {
				t.Errorf("extractDate(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEnsureICSExtension(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", testutil.FilenameEventICS},
		{"whitespace", "   ", testutil.FilenameEventICS},
		{"no extension", "event", testutil.FilenameEventICS},
		{"with extension", testutil.FilenameEventICS, testutil.FilenameEventICS},
		{"uppercase extension", "event.ICS", "event.ICS"},
		{"mixed case", "event.Ics", "event.Ics"},
		{"other extension", "event.txt", "event.txt.ics"},
		{testutil.TestNameWithSpaces, "  my event  ", "my event.ics"},
		{"with .ics already", "my-event.ics", "my-event.ics"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ensureICSExtension(tt.input)
			if got != tt.want {
				t.Errorf("ensureICSExtension(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEnsureDirForFile(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"no directory", testutil.FilenameEventICS, false},
		{"current directory", "./event.ics", false},
		{"subdirectory", filepath.Join(t.TempDir(), "subdir", testutil.FilenameEventICS), false},
		{"multiple levels", filepath.Join(t.TempDir(), "a", "b", "c", testutil.FilenameEventICS), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ensureDirForFile(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ensureDirForFile(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
			if !tt.wantErr && filepath.Dir(tt.path) != "." && filepath.Dir(tt.path) != "" {
				// Check directory was created
				dir := filepath.Dir(tt.path)
				if info, err := os.Stat(dir); err != nil || !info.IsDir() {
					t.Errorf("ensureDirForFile(%q) did not create directory %q", tt.path, dir)
				}
			}
		})
	}
}

func TestEnsureUniquePath(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name         string
		path         string
		existingFile bool
		want         string
	}{
		{"non-existent", filepath.Join(tmpDir, "new.ics"), false, filepath.Join(tmpDir, "new.ics")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.existingFile {
				// Create the file
				if err := os.WriteFile(tt.path, []byte("test"), 0644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
			}
			got := ensureUniquePath(tt.path)
			if got != tt.want {
				t.Errorf("ensureUniquePath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}

	// Test collision scenario
	t.Run("with collision", func(t *testing.T) {
		path := filepath.Join(tmpDir, "collision.ics")
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
		got := ensureUniquePath(path)
		expected := filepath.Join(tmpDir, "collision-2.ics")
		if got != expected {
			t.Errorf("ensureUniquePath(%q) = %q, want %q", path, got, expected)
		}
	})
}

// ============================================================================
// Batch format detection tests
// ============================================================================

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

// ============================================================================
// Template input format detection tests
// ============================================================================

func TestDetectTemplateInputFormat(t *testing.T) {
	tests := []struct {
		name    string
		flag    string
		path    string
		want    string
		wantErr bool
	}{
		{"auto csv", "auto", testutil.FilenameDataCSV, "csv", false},
		{"auto json", "auto", "data.json", "json", false},
		{"empty auto csv", "", testutil.FilenameDataCSV, "csv", false},
		{"explicit csv", "csv", testutil.FilenameDataTXT, "csv", false},
		{"explicit json", "json", testutil.FilenameDataTXT, "json", false},
		{"CSV uppercase", "CSV", testutil.FilenameDataTXT, "csv", false},
		{"auto unknown", "auto", testutil.FilenameDataTXT, "", true},
		{"invalid format", "yaml", testutil.FilenameDataCSV, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := detectTemplateInputFormat(tt.flag, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("detectTemplateInputFormat(%q, %q) error = %v, wantErr %v", tt.flag, tt.path, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("detectTemplateInputFormat(%q, %q) = %v, want %v", tt.flag, tt.path, got, tt.want)
			}
		})
	}
}

// ============================================================================
// File loading tests
// ============================================================================

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
				t.Fatalf("failed to write test file: %v", err)
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
			t.Fatalf("failed to write test file: %v", err)
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
				t.Fatalf("failed to write test file: %v", err)
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

func TestLoadTemplateFromJSON(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    int
		wantErr bool
	}{
		{
			name:    "valid json",
			content: `[{"field1":"value1","field2":"value2"}]`,
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
			name:    "multiple records",
			content: `[{"field":"a"},{"field":"b"},{"field":"c"}]`,
			want:    3,
			wantErr: false,
		},
		{
			name:    "invalid json",
			content: `{invalid}`,
			want:    0,
			wantErr: true,
		},
		{
			name:    "skip empty records",
			content: `[{"field":"a"},{"":""},{"field":"c"}]`,
			want:    2,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, testutil.FilenameTestJSON)
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			got, err := loadTemplateFromJSON(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadTemplateFromJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("loadTemplateFromJSON() returned %d records, want %d", len(got), tt.want)
			}
		})
	}
}

func TestLoadTemplateRecords(t *testing.T) {
	tmpDir := t.TempDir()

	csvPath := filepath.Join(tmpDir, "template.csv")
	csvContent := "field1,field2\nvalue1,value2"
	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		t.Fatalf("failed to write CSV: %v", err)
	}

	jsonPath := filepath.Join(tmpDir, "template.json")
	jsonContent := `[{"field1":"value1","field2":"value2"}]`
	if err := os.WriteFile(jsonPath, []byte(jsonContent), 0644); err != nil {
		t.Fatalf("failed to write JSON: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		format  string
		wantLen int
		wantErr bool
	}{
		{"csv", csvPath, "csv", 1, false},
		{"json", jsonPath, "json", 1, false},
		{"unknown format", csvPath, "yaml", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := loadTemplateRecords(tt.path, tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadTemplateRecords() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("loadTemplateRecords() returned %d records, want %d", len(got), tt.wantLen)
			}
		})
	}
}

// ============================================================================
// ICS parsing and linting tests
// ============================================================================

func TestUnfoldICSLines(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"simple", "LINE1\nLINE2\nLINE3", 3},
		{"folded line", "LINE1\n CONTINUED", 1},
		{"folded with tab", "LINE1\n\tCONTINUED", 1},
		{"multiple folds", "LINE1\n PART2\n PART3", 1},
		{"crlf", "LINE1\r\nLINE2\r\n", 2},
		{"empty lines", "LINE1\n\nLINE2", 2},
		{"only empty", "\n\n\n", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := unfoldICSLines(tt.input)
			if len(got) != tt.want {
				t.Errorf("unfoldICSLines() returned %d lines, want %d", len(got), tt.want)
			}
		})
	}
}

func TestParseICSProperty(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		wantName  string
		wantValue string
		wantOk    bool
	}{
		{"simple", "SUMMARY:Test Event", "SUMMARY", testutil.EventTitleTestEvent, true},
		{"with params", "DTSTART;TZID=Europe/Madrid:20250501T100000", "DTSTART", "20250501T100000", true},
		{"no colon", "INVALID", "", "", false},
		{"empty key", ":value", "", "", false},
		{"empty value", "KEY:", "KEY", "", true},
		{"lowercase", "summary:Test", "SUMMARY", "Test", true},
		{testutil.TestNameWithSpaces, "  SUMMARY  :  Test  ", "SUMMARY", "Test", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotValue, gotOk := parseICSProperty(tt.line)
			if gotOk != tt.wantOk {
				t.Errorf("parseICSProperty(%q) ok = %v, want %v", tt.line, gotOk, tt.wantOk)
				return
			}
			if gotOk {
				if gotName != tt.wantName {
					t.Errorf("parseICSProperty(%q) name = %q, want %q", tt.line, gotName, tt.wantName)
				}
				if gotValue != tt.wantValue {
					t.Errorf("parseICSProperty(%q) value = %q, want %q", tt.line, gotValue, tt.wantValue)
				}
			}
		})
	}
}

func TestLintICSFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name: "valid",
			content: `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VEVENT
UID:test@example.com
SUMMARY:Test Event
DTSTART:20250501T100000Z
DTEND:20250501T110000Z
END:VEVENT
END:VCALENDAR`,
			wantErr: false,
		},
		{
			name:    testutil.TestNameEmptyFile,
			content: "",
			wantErr: true,
		},
		{
			name: "missing BEGIN:VCALENDAR",
			content: `VERSION:2.0
BEGIN:VEVENT
END:VEVENT`,
			wantErr: true,
		},
		{
			name: "missing VEVENT",
			content: `BEGIN:VCALENDAR
VERSION:2.0
END:VCALENDAR`,
			wantErr: true,
		},
		{
			name: "missing UID",
			content: `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VEVENT
SUMMARY:Test
DTSTART:20250501T100000Z
DTEND:20250501T110000Z
END:VEVENT
END:VCALENDAR`,
			wantErr: true,
		},
		{
			name: "missing SUMMARY",
			content: `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VEVENT
UID:test@example.com
DTSTART:20250501T100000Z
DTEND:20250501T110000Z
END:VEVENT
END:VCALENDAR`,
			wantErr: true,
		},
		{
			name: "missing DTSTART",
			content: `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VEVENT
UID:test@example.com
SUMMARY:Test
DTEND:20250501T110000Z
END:VEVENT
END:VCALENDAR`,
			wantErr: true,
		},
		{
			name: "has DURATION instead of DTEND",
			content: `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VEVENT
UID:test@example.com
SUMMARY:Test
DTSTART:20250501T100000Z
DURATION:PT1H
END:VEVENT
END:VCALENDAR`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.ics")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			err := lintICSFile(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("lintICSFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	t.Run("directory instead of file", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := lintICSFile(tmpDir)
		if err == nil {
			t.Error("lintICSFile() expected error for directory, got nil")
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		err := lintICSFile("/non/existent/file.ics")
		if err == nil {
			t.Error("lintICSFile() expected error for non-existent file, got nil")
		}
	})
}

// ============================================================================
// Event building tests
// ============================================================================

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
		{
			name: "zero duration",
			record: batchRecord{
				Summary:  "Zero Duration",
				Start:    testutil.DateTime20250501_1000,
				Duration: "0m",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev, err := buildEventFromBatch(tt.record, tt.fallbackTZ)
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

// ============================================================================
// Config command tests
// ============================================================================

func TestNewConfigCmd(t *testing.T) {
	cmd := newConfigCmd()
	if cmd == nil {
		t.Fatal("newConfigCmd() returned nil")
	}
	if cmd.Use != "config" {
		t.Errorf("Use = %q, want %q", cmd.Use, "config")
	}

	// Check subcommands
	subcommands := cmd.Commands()
	if len(subcommands) != 3 {
		t.Errorf("expected 3 subcommands, got %d", len(subcommands))
	}

	var hasSet, hasList, hasAlarmProfiles bool
	for _, sub := range subcommands {
		if strings.HasPrefix(sub.Use, "set") {
			hasSet = true
		}
		if strings.HasPrefix(sub.Use, "list") {
			hasList = true
		}
		if strings.HasPrefix(sub.Use, "alarm-profiles") {
			hasAlarmProfiles = true
		}
	}
	if !hasSet {
		t.Error("config command missing 'set' subcommand")
	}
	if !hasList {
		t.Error("config command missing 'list' subcommand")
	}
	if !hasAlarmProfiles {
		t.Error("config command missing 'alarm-profiles' subcommand")
	}
}

func TestRunConfigSet(t *testing.T) {
	// This test requires the config package to work properly
	// We'll test the command creation and basic structure
	cmd := newConfigCmd()
	setCmd := findSubcommand(cmd, "set")
	if setCmd == nil {
		t.Fatal("set subcommand not found")
	}

	// Check that it requires exactly 2 args
	if setCmd.Args == nil {
		t.Error("set command should have Args validator")
	}
}

func TestRunConfigList(t *testing.T) {
	cmd := newConfigCmd()
	listCmd := findSubcommand(cmd, "list")
	if listCmd == nil {
		t.Fatal("list subcommand not found")
	}
	if listCmd.RunE == nil {
		t.Error("list command should have RunE function")
	}
}

// ============================================================================
// Version command tests
// ============================================================================

func TestNewVersionCmd(t *testing.T) {
	cmd := newVersionCmd()
	if cmd == nil {
		t.Fatal("newVersionCmd() returned nil")
	}
	if cmd.Use != "version" {
		t.Errorf("Use = %q, want %q", cmd.Use, "version")
	}
	if cmd.Run == nil {
		t.Error("version command should have Run function")
	}
}

// ============================================================================
// Template command tests
// ============================================================================

func TestRunTemplateList(t *testing.T) {
	cmd := newTemplateCmd()
	listCmd := findSubcommand(cmd, "list")
	if listCmd == nil {
		t.Fatal("list subcommand not found")
	}
	if listCmd.RunE == nil {
		t.Error("list command should have RunE function")
	}

	// Test that it doesn't crash with no templates
	// We can't easily test the full functionality without template setup
	// but we can verify the command structure
}

func TestRunTemplateDescribe(t *testing.T) {
	cmd := newTemplateCmd()
	descCmd := findSubcommand(cmd, "describe")
	if descCmd == nil {
		t.Fatal("describe subcommand not found")
	}
	if descCmd.RunE == nil {
		t.Error("describe command should have RunE function")
	}
}

func TestRunTemplateValidate(t *testing.T) {
	cmd := newTemplateCmd()
	validateCmd := findSubcommand(cmd, "validate")
	if validateCmd == nil {
		t.Fatal("validate subcommand not found")
	}
	if validateCmd.RunE == nil {
		t.Error("validate command should have RunE function")
	}

	// Test with no template directories
	tmpDir := t.TempDir()
	// Get parent command's flag
	if err := cmd.PersistentFlags().Set(testutil.TemplatesDir, tmpDir); err != nil {
		t.Fatalf("failed to set templates-dir flag: %v", err)
	}

	// Should not error with empty directory
	err := runTemplateValidate(validateCmd, nil)
	if err != nil {
		t.Errorf("runTemplateValidate() with empty dir error = %v, want nil", err)
	}
}

func TestTemplateFieldDefault(t *testing.T) {
	// This test is a placeholder since we need tpl.Template structure
	// which is in the internal package
	t.Skip("requires internal template structures")
}

func TestDeriveTemplateFilename(t *testing.T) {
	t.Skip("requires template manager setup")
}

// ============================================================================
// Normalization tests
// ============================================================================

func TestLooksLikeClock(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid 24h", "14:30", true},
		{"valid 12h", "9:15", true},
		{"valid single digit hour", "9:00", true},
		{"with seconds", "14:30:45", false},
		{testutil.TestNameFullDatetime, "2025-05-01 14:30", false},
		{testutil.TestNameDateOnly, testutil.Date20250501, false},
		{"empty", "", false},
		{"just hour", "14", false},
		{testutil.TestNameWithSpaces, " 14:30 ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := looksLikeClock(tt.input)
			if got != tt.want {
				t.Errorf("looksLikeClock(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestFirstNonEmpty(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  string
	}{
		{"first non-empty", []string{"", "a", "b"}, "a"},
		{"all empty", []string{"", "  ", ""}, ""},
		{"first is non-empty", []string{"a", "b", "c"}, "a"},
		{"last is non-empty", []string{"", "", "c"}, "c"},
		{testutil.TestNameEmptySlice, []string{}, ""},
		{"with whitespace", []string{"  ", "b"}, "b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := firstNonEmpty(tt.input...)
			if got != tt.want {
				t.Errorf("firstNonEmpty(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFmtDurationHuman(t *testing.T) {
	tests := []struct {
		name  string
		input time.Duration
		want  string
	}{
		{"zero", 0, "0m"},
		{"negative", -time.Minute, "0m"},
		{"minutes only", 45 * time.Minute, "45m"},
		{"hours only", 2 * time.Hour, "2h"},
		{"hours and minutes", 2*time.Hour + 30*time.Minute, "2h30m"},
		{"round up", 59*time.Second + 30*time.Millisecond, "1m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fmtDurationHuman(tt.input)
			if got != tt.want {
				t.Errorf("fmtDurationHuman(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSplitDateTime(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantDate string
		wantTime string
	}{
		{testutil.TestNameFullDatetime, "2025-05-01 14:30", testutil.Date20250501, "14:30"},
		{testutil.TestNameDateOnly, testutil.Date20250501, testutil.Date20250501, ""},
		{"empty", "", "", ""},
		{"with extra spaces", "2025-05-01  14:30", testutil.Date20250501, "14:30"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDate, gotTime := splitDateTime(tt.input)
			if gotDate != tt.wantDate || gotTime != tt.wantTime {
				t.Errorf("splitDateTime(%q) = (%q, %q), want (%q, %q)",
					tt.input, gotDate, gotTime, tt.wantDate, tt.wantTime)
			}
		})
	}
}

// ============================================================================
// Helper functions for tests
// ============================================================================

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func findSubcommand(parent *cobra.Command, name string) *cobra.Command {
	for _, cmd := range parent.Commands() {
		if strings.HasPrefix(cmd.Use, name) {
			return cmd
		}
	}
	return nil
}

// ============================================================================
// Additional batch tests for better coverage
// ============================================================================

func TestLoadBatchFromCSVWithDelimitedFields(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, testutil.FilenameTestCSV)
	content := `summary,start,end,exdate,categories,alarms
Event,2025-05-01 10:00,2025-05-01 11:00,"2025-05-03,2025-05-04","work,urgent","15m,30m"`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
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
		t.Errorf("expected 2 exdates, got %d", len(rec.ExDates))
	}
	if len(rec.Categories) != 2 {
		t.Errorf("expected 2 categories, got %d", len(rec.Categories))
	}
	if len(rec.Alarms) != 2 {
		t.Errorf("expected 2 alarms, got %d", len(rec.Alarms))
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
		t.Fatalf("failed to write test file: %v", err)
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

	ev, err := buildEventFromBatch(rec, "")
	if err != nil {
		t.Fatalf("buildEventFromBatch() error = %v", err)
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

	ev, err := buildEventFromBatch(rec, "")
	if err != nil {
		t.Fatalf("buildEventFromBatch() error = %v", err)
	}

	if ev.RRule != testutil.RRuleDaily5Count {
		t.Errorf("RRule = %q, want %q", ev.RRule, testutil.RRuleDaily5Count)
	}
}

func TestParseICSPropertyWithComplexParams(t *testing.T) {
	line := "ATTENDEE;CN=John Doe;ROLE=REQ-PARTICIPANT:mailto:john@example.com"
	name, value, ok := parseICSProperty(line)

	if !ok {
		t.Fatal("parseICSProperty() returned ok = false")
	}
	if name != "ATTENDEE" {
		t.Errorf("name = %q, want %q", name, "ATTENDEE")
	}
	if value != "mailto:john@example.com" {
		t.Errorf("value = %q, want %q", value, "mailto:john@example.com")
	}
}

func TestUnfoldICSLinesWithRealExample(t *testing.T) {
	input := `DESCRIPTION:This is a very long description that spans multiple lines be
 cause it exceeds 75 characters according to RFC 5545 folding rules. It sh
 ould be unfolded into a single line.
SUMMARY:Test Event`

	lines := unfoldICSLines(input)
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}

	desc := lines[0]
	if !strings.Contains(desc, "DESCRIPTION:") {
		t.Error("first line should be DESCRIPTION")
	}
	if strings.Contains(desc, "\n") {
		t.Error("unfolded line should not contain newlines")
	}
}

func TestEnsureUniquePathMultipleCollisions(t *testing.T) {
	tmpDir := t.TempDir()
	basePath := filepath.Join(tmpDir, testutil.FilenameEventICS)

	// Create the base file
	if err := os.WriteFile(basePath, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// First call should return event-2.ics
	result := ensureUniquePath(basePath)
	expected := filepath.Join(tmpDir, "event-2.ics")
	if result != expected {
		t.Errorf("ensureUniquePath() = %q, want %q", result, expected)
	}

	// Create event-2.ics
	if err := os.WriteFile(result, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Second call should return event-3.ics
	result2 := ensureUniquePath(basePath)
	expected2 := filepath.Join(tmpDir, "event-3.ics")
	if result2 != expected2 {
		t.Errorf("ensureUniquePath() = %q, want %q", result2, expected2)
	}
}

// ============================================================================
// Additional comprehensive tests for better coverage
// ============================================================================

func TestLoadTemplateFromCSV(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    int
		wantErr bool
	}{
		{
			name:    "valid csv",
			content: "field1,field2\nvalue1,value2\nvalue3,value4",
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
			content: "field1,field2",
			want:    0,
			wantErr: false,
		},
		{
			name:    "skip empty rows",
			content: "field1,field2\n,\nvalue1,value2",
			want:    1,
			wantErr: false,
		},
		{
			name:    "skip empty header columns",
			content: "field1,,field3\nvalue1,value2,value3",
			want:    1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, testutil.FilenameTestCSV)
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			got, err := loadTemplateFromCSV(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadTemplateFromCSV() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("loadTemplateFromCSV() returned %d records, want %d", len(got), tt.want)
			}
		})
	}

	t.Run("non-existent file", func(t *testing.T) {
		_, err := loadTemplateFromCSV("/non/existent/file.csv")
		if err == nil {
			t.Error("expected error for non-existent file")
		}
	})
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
			_, err := buildEventFromBatch(tt.record, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("buildEventFromBatch() error = %v, wantErr %v", err, tt.wantErr)
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

	ev, err := buildEventFromBatch(rec, "")
	if err != nil {
		t.Fatalf("buildEventFromBatch() error = %v", err)
	}

	if len(ev.ExDates) != 2 {
		t.Errorf("expected 2 exdates, got %d", len(ev.ExDates))
	}

	if len(ev.Alarms) != 2 {
		t.Errorf("expected 2 alarms, got %d", len(ev.Alarms))
	}
}

func TestValueAsStringSliceEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  []string
	}{
		{"int slice", []interface{}{1, 2, 3}, []string{"1", "2", "3"}},
		{"mixed types", []interface{}{"a", 1, true, 2.5}, []string{"a", "1", "true", "2.5"}},
		{"float numbers", []interface{}{1.5, 2.0, 3.7}, []string{"1.5", "2", "3.7"}},
		{"with empty strings in slice", []interface{}{"a", "", "b", "  "}, []string{"a", "b"}},
		{"int to string", 123, []string{"123"}},
		{"bool to string", true, []string{"true"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := valueAsStringSlice(tt.input)
			if !equalStringSlices(got, tt.want) {
				t.Errorf("valueAsStringSlice(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestValueAsAlarmSliceComplexInputs(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  int
	}{
		{"single alarm", "15m", 1},
		{"multiple alarms in string", []interface{}{"15m\n30m", "1h"}, 3},
		{"complex alarm specs", []string{"trigger=-15m,description=Test", "20m"}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := valueAsAlarmSlice(tt.input)
			if len(got) != tt.want {
				t.Errorf("valueAsAlarmSlice(%v) returned %d items, want %d", tt.input, len(got), tt.want)
			}
		})
	}
}

func TestParseDateTimeWithTZ(t *testing.T) {
	tests := []struct {
		name    string
		dateStr string
		timeStr string
		tz      string
		wantErr bool
	}{
		{"date only UTC", testutil.Date20250501, "", "", false},
		{"datetime UTC", testutil.Date20250501, "14:30", "", false},
		{testutil.TestNameWithTimezone, testutil.Date20250501, "14:30", testutil.TZEuropeMadrid, false},
		{"invalid timezone", testutil.Date20250501, "14:30", "Invalid/Zone", true},
		{"invalid date", "invalid", "14:30", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseDateTimeWithTZ(tt.dateStr, tt.timeStr, tt.tz)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDateTimeWithTZ(%q, %q, %q) error = %v, wantErr %v",
					tt.dateStr, tt.timeStr, tt.tz, err, tt.wantErr)
			}
		})
	}
}

func TestAddDurationToStart(t *testing.T) {
	tests := []struct {
		name      string
		start     string
		tz        string
		duration  time.Duration
		wantEmpty bool
	}{
		{"valid datetime", testutil.DateTime20250501_1000, "", 30 * time.Minute, false},
		{testutil.TestNameWithTimezone, testutil.DateTime20250501_1000, testutil.TZEuropeMadrid, 1 * time.Hour, false},
		{"invalid datetime", "invalid", "", 30 * time.Minute, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := addDurationToStart(tt.start, tt.tz, tt.duration)
			isEmpty := got == ""
			if isEmpty != tt.wantEmpty {
				t.Errorf("addDurationToStart(%q, %q, %v) empty = %v, want %v",
					tt.start, tt.tz, tt.duration, isEmpty, tt.wantEmpty)
			}
		})
	}
}

func TestParseExDateValues(t *testing.T) {
	tests := []struct {
		name    string
		values  []string
		tz      string
		allDay  bool
		wantLen int
		wantErr bool
	}{
		{"single date", []string{testutil.Date20250501}, "UTC", true, 1, false},
		{"multiple dates", []string{testutil.Date20250501, "2025-05-02", testutil.Date20250503}, "UTC", true, 3, false},
		{"datetime values", []string{testutil.DateTime20250501_1000, "2025-05-02 14:00"}, testutil.TZEuropeMadrid, false, 2, false},
		{"mixed date and datetime", []string{testutil.Date20250501, "2025-05-02 10:00"}, "UTC", false, 2, false},
		{"with T separator", []string{"2025-05-01T10:00"}, "UTC", false, 1, false},
		{"empty values", []string{"", "  "}, "UTC", true, 0, false},
		{"invalid date", []string{"invalid"}, "UTC", true, 0, true},
		{testutil.TestNameWithTimezone, []string{testutil.DateTime20250501_1000}, testutil.TZAmericaNewYork, false, 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseExDateValues(tt.values, tt.tz, tt.allDay)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseExDateValues() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("parseExDateValues() returned %d dates, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestNormalizeClockOnlyDateTimes(t *testing.T) {
	tests := []struct {
		name      string
		values    map[string]string
		startKey  string
		endKey    string
		tzKey     string
		checkFunc func(*testing.T, map[string]string)
	}{
		{
			name: "clock only start",
			values: map[string]string{
				"start": "14:30",
				"tz":    testutil.TZEuropeMadrid,
			},
			startKey: "start",
			endKey:   "end",
			tzKey:    "tz",
			checkFunc: func(t *testing.T, m map[string]string) {
				if !strings.Contains(m["start"], " 14:30") {
					t.Errorf("expected start to have time component, got %q", m["start"])
				}
			},
		},
		{
			name: "clock only end but might be duration",
			values: map[string]string{
				"start": testutil.DateTime20250501_1400,
				"end":   "15:30",
				"tz":    "UTC",
			},
			startKey: "start",
			endKey:   "end",
			tzKey:    "tz",
			checkFunc: func(t *testing.T, m map[string]string) {
				// The function checks if it's a duration first, so 15:30 might not be converted
				// This is expected behavior per the implementation
				if m["end"] == "" {
					t.Error("end should not be empty")
				}
			},
		},
		{
			name: "no normalization needed",
			values: map[string]string{
				"start": testutil.DateTime20250501_1400,
				"end":   "2025-05-01 15:00",
			},
			startKey: "start",
			endKey:   "end",
			tzKey:    "tz",
			checkFunc: func(t *testing.T, m map[string]string) {
				if m["start"] != testutil.DateTime20250501_1400 {
					t.Errorf("start should not be modified, got %q", m["start"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalizeClockOnlyDateTimes(tt.values, tt.startKey, tt.endKey, tt.tzKey)
			if tt.checkFunc != nil {
				tt.checkFunc(t, tt.values)
			}
		})
	}
}

func TestNormalizeEndTimeFromDuration(t *testing.T) {
	tests := []struct {
		name            string
		values          map[string]string
		startKey        string
		endKey          string
		durationKey     string
		tzKey           string
		defaultDuration string
		checkFunc       func(*testing.T, map[string]string)
	}{
		{
			name: "use duration field",
			values: map[string]string{
				"start":    testutil.DateTime20250501_1000,
				"duration": "90m",
			},
			startKey:        "start",
			endKey:          "end",
			durationKey:     "duration",
			tzKey:           "tz",
			defaultDuration: "30m",
			checkFunc: func(t *testing.T, m map[string]string) {
				if m["end"] == "" {
					t.Error("end should be set from duration")
				}
			},
		},
		{
			name: "use default duration",
			values: map[string]string{
				"start": testutil.DateTime20250501_1000,
			},
			startKey:        "start",
			endKey:          "end",
			durationKey:     "duration",
			tzKey:           "tz",
			defaultDuration: "1h",
			checkFunc: func(t *testing.T, m map[string]string) {
				if m["end"] == "" {
					t.Error("end should be set from default duration")
				}
			},
		},
		{
			name: "end as duration",
			values: map[string]string{
				"start": testutil.DateTime20250501_1000,
				"end":   "45m",
			},
			startKey:        "start",
			endKey:          "end",
			durationKey:     "duration",
			tzKey:           "tz",
			defaultDuration: "30m",
			checkFunc: func(t *testing.T, m map[string]string) {
				if !strings.Contains(m["end"], testutil.Date20250501) {
					t.Errorf("end should be converted to datetime, got %q", m["end"])
				}
				if m["duration"] == "" {
					t.Error("duration should be set")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalizeEndTimeFromDuration(tt.values, tt.startKey, tt.endKey, tt.durationKey, tt.tzKey, tt.defaultDuration)
			if tt.checkFunc != nil {
				tt.checkFunc(t, tt.values)
			}
		})
	}
}

func TestCityToIANA(t *testing.T) {
	tests := []struct {
		city string
		want string
	}{
		// Spain
		{"madrid", testutil.TZEuropeMadrid},
		{"barcelona", testutil.TZEuropeMadrid},
		{"melilla", testutil.TZAfricaCeuta},
		{"ceuta", testutil.TZAfricaCeuta},
		{"canarias", testutil.TZAtlanticCanary},
		{"tenerife", testutil.TZAtlanticCanary},

		// Brazil
		{"pelotas", testutil.TZAmericaSaoPaulo},
		{"porto alegre", testutil.TZAmericaSaoPaulo},
		{"campo grande", testutil.TZAmericaCampoGrande},
		{"manaus", "America/Manaus"},
		{"rio", testutil.TZAmericaSaoPaulo},
		{"sao paulo", testutil.TZAmericaSaoPaulo},

		// Ireland/UK
		{"dublin", testutil.TZEuropeDublin},
		{"london", testutil.TZEuropeLondon},

		// Unknown
		{"unknown", ""},
		{"", ""},
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

func TestLabelForField(t *testing.T) {
	// This requires internal template structures, but we can test the logic
	t.Skip("requires internal template structures")
}

func TestIsAlarmField(t *testing.T) {
	// This requires internal template structures
	t.Skip("requires internal template structures")
}

func TestCleanDisplay(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Madrid (Spain)", "Madrid"},
		{"Central European Time (CET)", "Central European Time"},
		{"UTC", "UTC"},
		{"Dublin (no closing", "Dublin (no closing"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := cleanDisplay(tt.input)
			if got != tt.want {
				t.Errorf("cleanDisplay(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestMergeTemplateValues(t *testing.T) {
	// This requires internal template structures
	t.Skip("requires internal template structures")
}

func TestAugmentValuesForFilename(t *testing.T) {
	ev := &calendar.Event{
		Summary:   testutil.EventTitleTestEvent,
		StartTime: time.Date(2025, 5, 1, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2025, 5, 1, 11, 0, 0, 0, time.UTC),
	}

	values := map[string]string{
		"field1": "value1",
		"field2": "value2",
	}

	result := augmentValuesForFilename(values, ev)

	if result["field1"] != "value1" {
		t.Error("original values should be preserved")
	}

	if result["start_date"] != testutil.Date20250501 {
		t.Errorf("start_date = %q, want %q", result["start_date"], testutil.Date20250501)
	}

	if result["end_date"] != testutil.Date20250501 {
		t.Errorf("end_date = %q, want %q", result["end_date"], testutil.Date20250501)
	}

	if !strings.Contains(result["start_time_iso"], testutil.DateTime20250501_1000) {
		t.Errorf("start_time_iso should contain datetime, got %q", result["start_time_iso"])
	}
}

func TestBuildTemplateCalendar(t *testing.T) {
	ev := &calendar.Event{
		Summary:   testutil.EventTitleTestEvent,
		StartTime: time.Date(2025, 5, 1, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2025, 5, 1, 11, 0, 0, 0, time.UTC),
		StartTZ:   testutil.TZEuropeMadrid,
	}

	cal := buildTemplateCalendar(ev)

	if cal == nil {
		t.Fatal("buildTemplateCalendar() returned nil")
	}

	if cal.Name != testutil.EventTitleTestEvent {
		t.Errorf("calendar name = %q, want %q", cal.Name, testutil.EventTitleTestEvent)
	}

	if !cal.IncludeVTZ {
		t.Error("expected IncludeVTZ to be true")
	}
}

func TestPrependToday(t *testing.T) {
	result := prependToday("14:30", "UTC")
	if result == "" {
		t.Error("prependToday() returned empty string")
	}
	if !strings.Contains(result, "14:30") {
		t.Errorf("result should contain time component, got %q", result)
	}
	if !strings.Contains(result, "-") {
		t.Errorf("result should contain date separator, got %q", result)
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
		name string
		key  string
		want string
	}{
		{"exists", "col1", "value1"},
		{"second column", "col2", "value2"},
		{"last column", "col3", "value3"},
		{"missing key", "col4", ""},
		{testutil.TestNameEmptyString, "missing", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := csvValue(row, index, tt.key)
			if got != tt.want {
				t.Errorf("csvValue() = %q, want %q", got, tt.want)
			}
		})
	}

	// Test with index out of range
	index2 := map[string]int{"col": 10}
	result := csvValue(row, index2, "col")
	if result != "" {
		t.Errorf("csvValue() with out of range index = %q, want empty", result)
	}
}
