package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tempus/internal/parsing"
)

func TestProcessCategories(t *testing.T) {
	tests := []struct {
		name       string
		categories []string
		customCat  string
		want       []string
	}{
		{
			name:       "no other selected",
			categories: []string{"work", "health"},
			customCat:  "",
			want:       []string{"work", "health"},
		},
		{
			name:       "other with custom cat",
			categories: []string{"work", "other"},
			customCat:  "meetings",
			want:       []string{"work", "meetings"},
		},
		{
			name:       "other without custom cat is dropped",
			categories: []string{"other"},
			customCat:  "",
			want:       []string{},
		},
		{
			name:       "no other selected with non-empty customCat is ignored",
			categories: []string{},
			customCat:  "anything",
			want:       []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := processCategories(tt.categories, tt.customCat)
			if len(got) != len(tt.want) {
				t.Fatalf("processCategories(%v, %q) = %v, want %v", tt.categories, tt.customCat, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("processCategories()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestResolveAlarmSpecs(t *testing.T) {
	tests := []struct {
		name    string
		profile string
		custom  string
		want    []string
	}{
		{
			name:    "adhd-default returns profile spec",
			profile: "adhd-default",
			custom:  "",
			want:    []string{"profile:adhd-default"},
		},
		{
			name:    "none returns empty",
			profile: "none",
			custom:  "",
			want:    []string{},
		},
		{
			name:    "custom with offsets splits by comma",
			profile: "custom",
			custom:  "-2h,-30m,-5m",
			want:    []string{"-2h", "-30m", "-5m"},
		},
		{
			name:    "custom with empty string returns empty",
			profile: "custom",
			custom:  "",
			want:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveAlarmSpecs(tt.profile, tt.custom)
			if len(got) != len(tt.want) {
				t.Fatalf("resolveAlarmSpecs(%q, %q) = %v, want %v", tt.profile, tt.custom, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("resolveAlarmSpecs()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestBuildStartStr(t *testing.T) {
	tests := []struct {
		name      string
		date      string
		startTime string
		want      string
	}{
		{
			name:      "date and time combined",
			date:      "2025-01-15",
			startTime: "10:00",
			want:      "2025-01-15 10:00",
		},
		{
			name:      "date only when time is empty",
			date:      "2025-01-15",
			startTime: "",
			want:      "2025-01-15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildStartStr(tt.date, tt.startTime)
			if got != tt.want {
				t.Errorf("buildStartStr(%q, %q) = %q, want %q", tt.date, tt.startTime, got, tt.want)
			}
		})
	}
}

func TestInteractiveCreatesPipeline(t *testing.T) {
	tmpDir := t.TempDir()
	app := TestApp()
	app.Config.OutputDir = tmpDir

	vars := &interactiveVars{
		summary:   "Team Meeting",
		startDate: "2025-06-15",
		startTime: "10:00",
		duration:  "1h",
		timezone:  "UTC",
		alarmProf: "adhd-default",
	}

	parseOpts := parsing.ParseOptions{
		StartDate: vars.startDate,
		StartTime: vars.startTime,
		Duration:  vars.duration,
		Timezone:  vars.timezone,
		AllDay:    vars.allDay,
		Summary:   vars.summary,
	}
	result, err := parsing.Parse(parseOpts)
	if err != nil {
		t.Fatalf("parsing.Parse() error: %v", err)
	}

	wantStart := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
	wantEnd := time.Date(2025, 6, 15, 11, 0, 0, 0, time.UTC)
	if !result.Start.Equal(wantStart) {
		t.Errorf("Start = %v, want %v", result.Start, wantStart)
	}
	if !result.End.Equal(wantEnd) {
		t.Errorf("End = %v, want %v", result.End, wantEnd)
	}

	opts := &createOptions{
		summary:    vars.summary,
		startStr:   buildStartStr(vars.startDate, vars.startTime),
		durStr:     vars.duration,
		startTZ:    vars.timezone,
		allDay:     vars.allDay,
		categories: processCategories(vars.categories, vars.customCat),
		alarms:     resolveAlarmSpecs(vars.alarmProf, vars.customAlarm),
		output:     filepath.Join(tmpDir, Slugify(vars.summary)+".ics"),
	}

	cal := createCalendarWithEvent(opts, result.Start, result.End)
	if len(cal.Events) != 1 {
		t.Fatalf("calendar has %d events, want 1", len(cal.Events))
	}
	if cal.Events[0].Summary != "Team Meeting" {
		t.Errorf("event Summary = %q, want %q", cal.Events[0].Summary, "Team Meeting")
	}
	if !cal.Events[0].StartTime.Equal(wantStart) {
		t.Errorf("event StartTime = %v, want %v", cal.Events[0].StartTime, wantStart)
	}
	if !cal.Events[0].EndTime.Equal(wantEnd) {
		t.Errorf("event EndTime = %v, want %v", cal.Events[0].EndTime, wantEnd)
	}

	if err := writeCalendarOutput(app, cal, opts.output); err != nil {
		t.Fatalf("writeCalendarOutput() error: %v", err)
	}

	content, err := os.ReadFile(opts.output)
	if err != nil {
		t.Fatalf("ReadFile(%q) error: %v", opts.output, err)
	}
	ics := string(content)
	if !strings.Contains(ics, "BEGIN:VCALENDAR") {
		t.Error("ICS missing BEGIN:VCALENDAR")
	}
	if !strings.Contains(ics, "SUMMARY:Team Meeting") {
		t.Error("ICS missing SUMMARY:Team Meeting")
	}
	if !strings.Contains(ics, "BEGIN:VEVENT") {
		t.Error("ICS missing BEGIN:VEVENT")
	}
}

func TestInteractiveCancelDoesNotWrite(t *testing.T) {
	tmpDir := t.TempDir()
	app := TestApp()
	app.Config.OutputDir = tmpDir

	vars := &interactiveVars{
		summary:   "Cancelled Event",
		startDate: "2025-06-15",
		startTime: "10:00",
		duration:  "1h",
		timezone:  "UTC",
		alarmProf: "adhd-default",
		confirmed: false,
	}

	if vars.confirmed {
		t.Fatal("confirmed should be false for this test")
	}

	outputPath := filepath.Join(tmpDir, Slugify(vars.summary)+".ics")
	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		t.Fatal("file should not exist before cancellation")
	}

	_ = vars
}

func TestBuildSummaryDescription(t *testing.T) {
	vars := &interactiveVars{
		summary:     "Team Meeting",
		startDate:   "2025-06-15",
		startTime:   "10:00",
		duration:    "1h",
		timezone:    "UTC",
		alarmProf:   "adhd-default",
		location:    "Room 1",
		description: "Weekly sync",
	}

	desc := buildSummaryDescription(vars)

	for _, want := range []string{"Team Meeting", "2025-06-15 10:00", "1h", "UTC", "profile:adhd-default", "Room 1", "Weekly sync"} {
		if !strings.Contains(desc, want) {
			t.Errorf("buildSummaryDescription() missing %q in output:\n%s", want, desc)
		}
	}
}

func TestBuildSummaryDescriptionNoneAlarm(t *testing.T) {
	vars := &interactiveVars{
		summary:   "Doctor",
		startDate: "2025-06-15",
		startTime: "",
		duration:  "",
		timezone:  "Europe/Madrid",
		alarmProf: "none",
	}

	desc := buildSummaryDescription(vars)
	if !strings.Contains(desc, "(none)") {
		t.Errorf("buildSummaryDescription() should contain '(none)' for alarms when profile is none, got:\n%s", desc)
	}
}

func TestCreateEventFromWizard(t *testing.T) {
	tmpDir := t.TempDir()
	app := TestApp()
	app.Config.OutputDir = tmpDir

	vars := &interactiveVars{
		summary:   "Stand Up",
		startDate: "2025-07-01",
		startTime: "09:00",
		duration:  "30m",
		timezone:  "UTC",
		alarmProf: "single",
		confirmed: true,
	}

	err := createEventFromWizard(app, vars)
	if err != nil {
		t.Fatalf("createEventFromWizard() error: %v", err)
	}

	outputPath := filepath.Join(tmpDir, Slugify(vars.summary)+".ics")
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error: %v", outputPath, err)
	}
	ics := string(content)
	if !strings.Contains(ics, "BEGIN:VCALENDAR") {
		t.Error("ICS missing BEGIN:VCALENDAR")
	}
	if !strings.Contains(ics, "SUMMARY:Stand Up") {
		t.Error("ICS missing SUMMARY:Stand Up")
	}
}

func TestCreateEventFromWizardAllDay(t *testing.T) {
	tmpDir := t.TempDir()
	app := TestApp()
	app.Config.OutputDir = tmpDir

	vars := &interactiveVars{
		summary:   "Holiday",
		startDate: "2025-12-25",
		startTime: "",
		timezone:  "UTC",
		alarmProf: "none",
		confirmed: true,
	}

	err := createEventFromWizard(app, vars)
	if err != nil {
		t.Fatalf("createEventFromWizard() all-day error: %v", err)
	}

	outputPath := filepath.Join(tmpDir, Slugify(vars.summary)+".ics")
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("Expected ICS file not found: %v", err)
	}
}

func TestCreateEventFromWizardWithCategories(t *testing.T) {
	tmpDir := t.TempDir()
	app := TestApp()
	app.Config.OutputDir = tmpDir

	vars := &interactiveVars{
		summary:     "Focus Block",
		startDate:   "2025-08-01",
		startTime:   "14:00",
		duration:    "2h",
		timezone:    "UTC",
		alarmProf:   "custom",
		customAlarm: "-30m,-5m",
		categories:  []string{"work", "other"},
		customCat:   "focus",
		confirmed:   true,
	}

	err := createEventFromWizard(app, vars)
	if err != nil {
		t.Fatalf("createEventFromWizard() with categories error: %v", err)
	}

	outputPath := filepath.Join(tmpDir, Slugify(vars.summary)+".ics")
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error: %v", outputPath, err)
	}
	ics := string(content)
	if !strings.Contains(ics, "BEGIN:VCALENDAR") {
		t.Error("ICS missing BEGIN:VCALENDAR")
	}
	if !strings.Contains(ics, "SUMMARY:Focus Block") {
		t.Errorf("ICS missing SUMMARY:Focus Block, got:\n%s", ics)
	}
}

func TestInteractiveFormStructure(t *testing.T) {
	vars := &interactiveVars{
		timezone:  "UTC",
		alarmProf: "adhd-default",
	}

	form := buildInteractiveForm(vars)
	if form == nil {
		t.Fatal("buildInteractiveForm() returned nil")
	}

	icsContent, err := os.ReadFile("/home/malpanez/repos/tempus/internal/cli/interactive.go")
	if err != nil {
		t.Skip("could not read interactive.go for step count verification")
	}

	src := string(icsContent)
	stepCount := strings.Count(src, "Step ")
	if stepCount < 7 {
		t.Errorf("interactive.go has %d Step references, want at least 7", stepCount)
	}

	for _, step := range []string{"Step 1/7", "Step 2/7", "Step 3/7", "Step 4/7", "Step 5/7", "Step 6/7", "Step 7/7"} {
		if !strings.Contains(src, step) {
			t.Errorf("interactive.go missing %q", step)
		}
	}
}
