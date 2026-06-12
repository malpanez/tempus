package cli

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"tempus/internal/calendar"
	"tempus/internal/prompts"
	"tempus/internal/testutil"
	"testing"
	"time"
)

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
				t.Fatalf(testutil.ErrMsgFailedToWriteTestFile, err)
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
				t.Fatalf(testutil.ErrMsgFailedToWriteTestFile, err)
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

func TestRunTemplateList(t *testing.T) {
	app := TestApp()
	cmd := NewTemplateCmd(app)
	listCmd := findSubcommand(cmd, "list")
	if listCmd == nil {
		t.Fatal("list subcommand not found")
	}
	if listCmd.RunE == nil {
		t.Error("list command should have RunE function")
	}
}

func TestRunTemplateDescribe(t *testing.T) {
	app := TestApp()
	cmd := NewTemplateCmd(app)
	descCmd := findSubcommand(cmd, "describe")
	if descCmd == nil {
		t.Fatal("describe subcommand not found")
	}
	if descCmd.RunE == nil {
		t.Error("describe command should have RunE function")
	}
}

func TestRunTemplateValidate(t *testing.T) {
	app := TestApp()
	cmd := NewTemplateCmd(app)
	validateCmd := findSubcommand(cmd, "validate")
	if validateCmd == nil {
		t.Fatal("validate subcommand not found")
	}
	if validateCmd.RunE == nil {
		t.Error("validate command should have RunE function")
	}

	tmpDir := t.TempDir()
	if err := cmd.PersistentFlags().Set(testutil.TemplatesDir, tmpDir); err != nil {
		t.Fatalf("failed to set templates-dir flag: %v", err)
	}

	err := runTemplateValidate(app, validateCmd, nil)
	if err != nil {
		t.Errorf("runTemplateValidate() with empty dir error = %v, want nil", err)
	}
}

func TestNewTemplateCmd(t *testing.T) {
	app := TestApp()
	cmd := NewTemplateCmd(app)
	if cmd == nil {
		t.Fatal("NewTemplateCmd() returned nil")
	}
	if cmd.Use != "template" {
		t.Errorf(testutil.ErrMsgUseMismatch, cmd.Use, "template")
	}

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

func TestTemplateFieldDefault(t *testing.T) {
	t.Skip(testutil.ErrMsgRequiresInternalStructures)
}

func TestDeriveTemplateFilename(t *testing.T) {
	t.Skip("requires template manager setup")
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

func TestMergeTemplateValues(t *testing.T) {
	t.Skip(testutil.ErrMsgRequiresInternalStructures)
}

func TestLabelForField(t *testing.T) {
	t.Skip(testutil.ErrMsgRequiresInternalStructures)
}

func TestIsAlarmField(t *testing.T) {
	t.Skip(testutil.ErrMsgRequiresInternalStructures)
}

func TestTemplateCreateMedicalSupportsAdvancedFeatures(t *testing.T) {
	app := TestApp()
	templateCmd := NewTemplateCmd(app)

	repoRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	templatesDir := filepath.Join(repoRoot, "..", "..", "internal", "templates", "json")

	for _, cmd := range templateCmd.Commands() {
		if !strings.HasPrefix(cmd.Use, "create") {
			continue
		}
		if err := cmd.Flags().Set(testutil.TemplatesDir, templatesDir); err != nil {
			t.Fatalf("failed to set templates-dir flag: %v", err)
		}

		outputDir := t.TempDir()
		if err := cmd.Flags().Set("output-dir", outputDir); err != nil {
			t.Fatalf("failed to set output-dir flag: %v", err)
		}

		inputs := strings.Join([]string{
			"Dr. Jane Doe",
			"",
			"City Clinic",
			"2025-10-15 09:30",
			"",
			"45m",
			testutil.TZEuropeMadrid,
			"Bring previous records",
			"FREQ=MONTHLY;COUNT=3",
			"2025-10-20 09:30,2025-10-25 09:30",
			"-15m",
			"",
			"",
		}, "\n") + "\n"

		prevScanner := prompts.Scanner
		prompts.Scanner = bufio.NewScanner(strings.NewReader(inputs))
		defer func() {
			prompts.Scanner = prevScanner
		}()

		if err := runTemplateCreate(app, cmd, []string{"medical"}); err != nil {
			t.Fatalf("runTemplateCreate returned error: %v", err)
		}

		expectedFile := filepath.Join(outputDir, "medical-dr-jane-doe-2025-10-15.ics")
		data, err := os.ReadFile(expectedFile)
		if err != nil {
			t.Fatalf("failed to read generated ICS: %v", err)
		}
		ics := string(data)

		if !strings.Contains(ics, "X-WR-TIMEZONE:Europe/Madrid") {
			t.Fatalf("expected ICS to include X-WR-TIMEZONE for Google Calendar compatibility:\n%s", ics)
		}
		if !strings.Contains(ics, "RRULE:FREQ=MONTHLY;COUNT=3") {
			t.Fatalf("expected ICS to include RRULE:\n%s", ics)
		}
		if !strings.Contains(ics, "EXDATE;TZID=Europe/Madrid:20251020T093000,20251025T093000") {
			t.Fatalf("expected ICS to include EXDATE line:\n%s", ics)
		}
		if strings.Count(ics, "BEGIN:VALARM") != 1 {
			t.Fatalf("expected one VALARM block, got:\n%s", ics)
		}
		if !strings.Contains(ics, "TRIGGER:-PT15M") {
			t.Fatalf("expected VALARM with -15m trigger:\n%s", ics)
		}
		break
	}
}

func TestTemplateCreateMedicalFromCSVGeneratesMultipleICS(t *testing.T) {
	app := TestApp()
	templateCmd := NewTemplateCmd(app)

	repoRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	templatesDir := filepath.Join(repoRoot, "..", "..", "internal", "templates", "json")

	for _, cmd := range templateCmd.Commands() {
		if !strings.HasPrefix(cmd.Use, "create") {
			continue
		}
		if err := cmd.Flags().Set(testutil.TemplatesDir, templatesDir); err != nil {
			t.Fatalf("failed to set templates-dir flag: %v", err)
		}

		outputDir := t.TempDir()
		if err := cmd.Flags().Set("output-dir", outputDir); err != nil {
			t.Fatalf("failed to set output-dir flag: %v", err)
		}

		csvContent := strings.Join([]string{
			"doctor,specialty,clinic,start_time,duration,timezone,notes,rrule,exdates,alarms",
			"Dr. Alice Smith,,Downtown Clinic,2025-11-01 08:00,30m,Europe/Madrid,,FREQ=WEEKLY;COUNT=4,,15m",
			"Dr. Bob Lee,,North Hospital,2025-11-02 09:15,45m,Europe/Madrid,Bring MRI results,,2025-11-09 09:15,-30m",
		}, "\n")

		csvPath := filepath.Join(t.TempDir(), "appointments.csv")
		if err := os.WriteFile(csvPath, []byte(csvContent), 0o644); err != nil {
			t.Fatalf("failed to write CSV: %v", err)
		}

		prevScanner := prompts.Scanner
		prompts.Scanner = bufio.NewScanner(strings.NewReader(""))
		defer func() {
			prompts.Scanner = prevScanner
		}()

		if err := cmd.Flags().Set("input", csvPath); err != nil {
			t.Fatalf("failed to set input flag: %v", err)
		}

		if err := runTemplateCreate(app, cmd, []string{"medical"}); err != nil {
			t.Fatalf("runTemplateCreate with CSV returned error: %v", err)
		}

		files, err := os.ReadDir(outputDir)
		if err != nil {
			t.Fatalf("failed to read output dir: %v", err)
		}
		if len(files) != 2 {
			t.Fatalf("expected 2 ICS files, got %d", len(files))
		}

		firstICS, err := os.ReadFile(filepath.Join(outputDir, files[0].Name()))
		if err != nil {
			t.Fatalf("failed to read first ICS: %v", err)
		}
		if !strings.Contains(string(firstICS), "BEGIN:VEVENT") {
			t.Fatalf("expected ICS content to contain VEVENT")
		}
		break
	}
}
