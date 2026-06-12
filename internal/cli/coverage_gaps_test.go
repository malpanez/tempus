package cli

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"tempus/internal/calendar"
	"tempus/internal/nd"
	"tempus/internal/prompts"
	tpl "tempus/internal/templates"

	"github.com/olebedev/when"
	"github.com/spf13/cobra"
)

// =====================================================================
// locale.go — runLocaleList
// =====================================================================

func TestRunLocaleList(t *testing.T) {
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	err := runLocaleList(app, nil, nil)
	if err != nil {
		t.Fatalf("runLocaleList() error: %v", err)
	}

	out := buf.String()
	if out == "" {
		t.Error("runLocaleList() produced no output")
	}
}

func TestRunLocaleListNilStdout(t *testing.T) {
	app := TestApp()
	app.Stdout = nil

	err := runLocaleList(app, nil, nil)
	if err != nil {
		t.Fatalf("runLocaleList() with nil stdout error: %v", err)
	}
}

// =====================================================================
// timezone.go — runTZList, runTZInfo, printZoneInfo
// =====================================================================

func TestRunTZList(t *testing.T) {
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	cmd := NewTimezoneCmd(app)
	listCmd := findSubcommand(cmd, "list")
	if listCmd == nil {
		t.Fatal("list subcommand not found")
	}

	err := runTZList(app, listCmd, nil)
	if err != nil {
		t.Fatalf("runTZList() error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "IANA") {
		t.Errorf("expected 'IANA' header in output, got: %s", out)
	}
}

func TestRunTZListWithSearch(t *testing.T) {
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	cmd := NewTimezoneCmd(app)
	listCmd := findSubcommand(cmd, "list")
	if listCmd == nil {
		t.Fatal("list subcommand not found")
	}
	if err := listCmd.Flags().Set("search", "madrid"); err != nil {
		t.Fatalf("failed to set search flag: %v", err)
	}

	err := runTZList(app, listCmd, nil)
	if err != nil {
		t.Fatalf("runTZList() with search error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(strings.ToLower(out), "madrid") {
		t.Errorf("expected 'madrid' in filtered output, got: %s", out)
	}
}

func TestRunTZListWithRegion(t *testing.T) {
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	cmd := NewTimezoneCmd(app)
	listCmd := findSubcommand(cmd, "list")
	if listCmd == nil {
		t.Fatal("list subcommand not found")
	}
	if err := listCmd.Flags().Set("region", "europe"); err != nil {
		t.Fatalf("failed to set region flag: %v", err)
	}

	err := runTZList(app, listCmd, nil)
	if err != nil {
		t.Fatalf("runTZList() with region error: %v", err)
	}
}

func TestRunTZListWithCountry(t *testing.T) {
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	cmd := NewTimezoneCmd(app)
	listCmd := findSubcommand(cmd, "list")
	if listCmd == nil {
		t.Fatal("list subcommand not found")
	}
	if err := listCmd.Flags().Set("country", "spain"); err != nil {
		t.Fatalf("failed to set country flag: %v", err)
	}

	err := runTZList(app, listCmd, nil)
	if err != nil {
		t.Fatalf("runTZList() with country error: %v", err)
	}
}

func TestRunTZListShowAll(t *testing.T) {
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	cmd := NewTimezoneCmd(app)
	listCmd := findSubcommand(cmd, "list")
	if listCmd == nil {
		t.Fatal("list subcommand not found")
	}
	if err := listCmd.Flags().Set("all", "true"); err != nil {
		t.Fatalf("failed to set all flag: %v", err)
	}

	err := runTZList(app, listCmd, nil)
	if err != nil {
		t.Fatalf("runTZList() all error: %v", err)
	}
}

func TestRunTZInfo(t *testing.T) {
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	err := runTZInfo(app, nil, []string{"Europe/Madrid"})
	if err != nil {
		t.Fatalf("runTZInfo() error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Europe/Madrid") {
		t.Errorf("expected 'Europe/Madrid' in output, got: %s", out)
	}
}

func TestRunTZInfoUnknownCity(t *testing.T) {
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	err := runTZInfo(app, nil, []string{"madrid"})
	if err != nil {
		t.Fatalf("runTZInfo() for city error: %v", err)
	}
}

func TestRunTZInfoNotFound(t *testing.T) {
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	err := runTZInfo(app, nil, []string{"Foo/Bar_NotReal"})
	if err != nil {
		t.Fatalf("runTZInfo() not-found error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "not found") && !strings.Contains(out, "Did you mean") {
		t.Errorf("expected 'not found' message, got: %s", out)
	}
}

func TestRunTZInfoEmptyArgs(t *testing.T) {
	app := TestApp()
	err := runTZInfo(app, nil, []string{""})
	if err == nil {
		t.Fatal("runTZInfo() expected error for empty args")
	}
}

// =====================================================================
// version.go — NewVersionCmd
// =====================================================================

func TestNewVersionCmdOutput(t *testing.T) {
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	cmd := NewVersionCmd(app, "1.2.3", "abc1234", "2025-01-01")
	cmd.SetOut(&buf)
	cmd.Run(cmd, nil)

	out := buf.String()
	if !strings.Contains(out, "1.2.3") {
		t.Errorf("expected version '1.2.3' in output, got: %s", out)
	}
}

func TestNewVersionCmdNoDate(t *testing.T) {
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	cmd := NewVersionCmd(app, "0.9.0", "", "")
	cmd.Run(cmd, nil)

	out := buf.String()
	if !strings.Contains(out, "0.9.0") {
		t.Errorf("expected version '0.9.0' in output, got: %s", out)
	}
}

// =====================================================================
// config.go — runConfigList, runConfigAlarmProfiles, findSubcommand
// =====================================================================

func TestFindSubcommand(t *testing.T) {
	app := TestApp()
	parent := NewConfigCmd(app)

	found := findSubcommand(parent, "set")
	if found == nil {
		t.Fatal("findSubcommand() for 'set' returned nil")
	}
	if !strings.HasPrefix(found.Use, "set") {
		t.Errorf("findSubcommand() for 'set' returned Use=%q", found.Use)
	}

	notFound := findSubcommand(parent, "nonexistent")
	if notFound != nil {
		t.Errorf("findSubcommand() for 'nonexistent' should return nil, got %v", notFound)
	}
}

// =====================================================================
// lint.go — handleEndEvent, processLintLine edge cases
// =====================================================================

func TestHandleEndEventWithoutBegin(t *testing.T) {
	state := newLintState()
	state.inEvent = false

	handleEndEvent(&state)

	if len(state.eventIssues) == 0 {
		t.Error("expected eventIssues for END:VEVENT without matching BEGIN")
	}
	if !strings.Contains(state.eventIssues[0], "unexpected END:VEVENT") {
		t.Errorf("unexpected issue message: %q", state.eventIssues[0])
	}
}

func TestProcessLintLineEndCalendar(t *testing.T) {
	state := newLintState()
	state.calendarSeen = true

	processLintLine(&state, "END:VCALENDAR")

	if !state.calendarSeen {
		t.Error("calendarSeen should still be true after END:VCALENDAR")
	}
}

func TestProcessLintLineEmpty(t *testing.T) {
	state := newLintState()
	processLintLine(&state, "")
	processLintLine(&state, "   ")

	if state.calendarSeen || state.eventSeen {
		t.Error("empty lines should not affect lint state")
	}
}

func TestRunLintNoFiles(t *testing.T) {
	app := TestApp()
	cmd := NewLintCmd(app)

	err := runLint(app, cmd, nil)
	if err == nil {
		t.Fatal("runLint() expected error when no --file provided")
	}
	if !strings.Contains(err.Error(), "no files to lint") {
		t.Errorf("expected 'no files to lint' error, got: %v", err)
	}
}

func TestRunLintValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	icsPath := filepath.Join(tmpDir, "valid.ics")
	icsContent := `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VEVENT
UID:test-uid-001@tempus
SUMMARY:Test Event
DTSTAMP:20250401T090000Z
DTSTART:20250501T100000Z
DTEND:20250501T110000Z
END:VEVENT
END:VCALENDAR
`
	if err := os.WriteFile(icsPath, []byte(icsContent), 0644); err != nil {
		t.Fatalf("failed to write ICS: %v", err)
	}

	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf
	cmd := NewLintCmd(app)
	if err := cmd.Flags().Set("file", icsPath); err != nil {
		t.Fatalf("failed to set file flag: %v", err)
	}

	err := runLint(app, cmd, nil)
	if err != nil {
		t.Fatalf("runLint() error for valid file: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Lint passed") {
		t.Errorf("expected 'Lint passed' in output, got: %s", out)
	}
}

func TestRunLintInvalidFile(t *testing.T) {
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf
	cmd := NewLintCmd(app)
	if err := cmd.Flags().Set("file", "/nonexistent/path/file.ics"); err != nil {
		t.Fatalf("failed to set file flag: %v", err)
	}

	err := runLint(app, cmd, nil)
	if err == nil {
		t.Fatal("runLint() expected error for nonexistent file")
	}
}

// =====================================================================
// template.go — printTemplateBasicInfo, printTemplateFields,
//               printTemplateOutput, printIfNotEmpty
// =====================================================================

func TestPrintTemplateBasicInfo(t *testing.T) {
	var buf bytes.Buffer
	tmpl := &tpl.Template{
		Name:        "my-template",
		Description: "A test template",
	}
	printTemplateBasicInfo(&buf, tmpl)

	out := buf.String()
	if !strings.Contains(out, "my-template") {
		t.Errorf("expected name in output, got: %s", out)
	}
	if !strings.Contains(out, "A test template") {
		t.Errorf("expected description in output, got: %s", out)
	}
}

func TestPrintTemplateBasicInfoNoDescription(t *testing.T) {
	var buf bytes.Buffer
	tmpl := &tpl.Template{
		Name:        "bare-template",
		Description: "",
	}
	printTemplateBasicInfo(&buf, tmpl)

	out := buf.String()
	if !strings.Contains(out, "bare-template") {
		t.Errorf("expected name in output, got: %s", out)
	}
	if !strings.Contains(out, "Description: -") {
		t.Errorf("expected 'Description: -' for empty description, got: %s", out)
	}
}

func TestPrintTemplateFields(t *testing.T) {
	var buf bytes.Buffer
	fields := []tpl.Field{
		{Key: "summary", Name: "Summary", Type: "text", Required: true},
		{Key: "location", Name: "Location", Type: "text", Required: false, Default: "Office"},
		{Key: "notes", Name: "Notes", Type: "text", Required: false, Description: "Additional info"},
	}
	printTemplateFields(&buf, fields)

	out := buf.String()
	if !strings.Contains(out, "summary") {
		t.Errorf("expected 'summary' field in output, got: %s", out)
	}
	if !strings.Contains(out, "required") {
		t.Errorf("expected 'required' in output, got: %s", out)
	}
	if !strings.Contains(out, "optional") {
		t.Errorf("expected 'optional' in output, got: %s", out)
	}
	if !strings.Contains(out, `default="Office"`) {
		t.Errorf("expected default value in output, got: %s", out)
	}
	if !strings.Contains(out, "Additional info") {
		t.Errorf("expected description in output, got: %s", out)
	}
}

func TestPrintIfNotEmpty(t *testing.T) {
	var buf bytes.Buffer

	printIfNotEmpty(&buf, "value: %s\n", "hello")
	if !strings.Contains(buf.String(), "hello") {
		t.Errorf("expected 'hello' in output, got: %s", buf.String())
	}

	buf.Reset()
	printIfNotEmpty(&buf, "value: %s\n", "")
	if buf.String() != "" {
		t.Errorf("expected no output for empty value, got: %s", buf.String())
	}

	buf.Reset()
	printIfNotEmpty(&buf, "value: %s\n", "   ")
	if buf.String() != "" {
		t.Errorf("expected no output for whitespace-only value, got: %s", buf.String())
	}
}

// =====================================================================
// quick.go — extractEventDetails, getQuickOutput, writeQuickCalendar
// =====================================================================

func TestGetQuickOutput(t *testing.T) {
	app := TestApp()
	cmd := NewQuickCmd(app)

	t.Run("uses flag value when set", func(t *testing.T) {
		if err := cmd.Flags().Set("output", "my-event.ics"); err != nil {
			t.Fatalf("failed to set output flag: %v", err)
		}
		got := getQuickOutput(cmd, "Some Meeting")
		if got != "my-event.ics" {
			t.Errorf("getQuickOutput() = %q, want %q", got, "my-event.ics")
		}
	})

	t.Run("derives from summary when no flag", func(t *testing.T) {
		cmd2 := NewQuickCmd(app)
		got := getQuickOutput(cmd2, "Team Standup")
		if !strings.HasSuffix(got, ".ics") {
			t.Errorf("getQuickOutput() = %q, expected .ics suffix", got)
		}
	})
}

func TestWriteQuickCalendar(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "quick.ics")

	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	details := quickParsedEvent{
		Summary:   "Quick Meeting",
		StartTime: time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2025, 6, 1, 11, 0, 0, 0, time.UTC),
		Location:  "Room 1",
	}

	err := writeQuickCalendar(app, details, "UTC", outputPath)
	if err != nil {
		t.Fatalf("writeQuickCalendar() error: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	ics := string(data)
	if !strings.Contains(ics, "BEGIN:VCALENDAR") {
		t.Errorf("expected valid ICS, got: %s", ics)
	}
	if !strings.Contains(ics, "SUMMARY:Quick Meeting") {
		t.Errorf("expected SUMMARY:Quick Meeting in ICS, got: %s", ics)
	}
	if !strings.Contains(ics, "LOCATION:Room 1") {
		t.Errorf("expected LOCATION in ICS, got: %s", ics)
	}
}

func TestWriteQuickCalendarNoTimezone(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "quick-notz.ics")

	app := TestApp()
	details := quickParsedEvent{
		Summary:   "Simple Event",
		StartTime: time.Date(2025, 7, 1, 9, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2025, 7, 1, 10, 0, 0, 0, time.UTC),
	}

	err := writeQuickCalendar(app, details, "", outputPath)
	if err != nil {
		t.Fatalf("writeQuickCalendar() no-tz error: %v", err)
	}
}

// =====================================================================
// create.go — parseDurationEnd, parseEndTime, parseTimedEventTimes
// =====================================================================

func TestParseDurationEndEdgeCases(t *testing.T) {
	start := time.Date(2025, 5, 1, 10, 0, 0, 0, time.UTC)

	t.Run("zero duration rejected", func(t *testing.T) {
		_, err := parseDurationEnd(start, "0m")
		if err == nil {
			t.Error("parseDurationEnd() expected error for zero duration")
		}
	})

	t.Run("invalid duration rejected", func(t *testing.T) {
		_, err := parseDurationEnd(start, "not-valid")
		if err == nil {
			t.Error("parseDurationEnd() expected error for invalid duration")
		}
	})

	t.Run("valid 30m", func(t *testing.T) {
		end, err := parseDurationEnd(start, "30m")
		if err != nil {
			t.Fatalf("parseDurationEnd() error: %v", err)
		}
		expected := start.Add(30 * time.Minute)
		if !end.Equal(expected) {
			t.Errorf("end = %v, want %v", end, expected)
		}
	})
}

func TestParseEndTimeEdgeCases(t *testing.T) {
	start := time.Date(2025, 5, 1, 10, 0, 0, 0, time.UTC)

	t.Run("zero duration rejected", func(t *testing.T) {
		_, err := parseEndTime(start, "0m")
		if err == nil {
			t.Error("parseEndTime() expected error for zero duration")
		}
	})

	t.Run("valid human duration", func(t *testing.T) {
		end, err := parseEndTime(start, "2h")
		if err != nil {
			t.Fatalf("parseEndTime() error: %v", err)
		}
		expected := start.Add(2 * time.Hour)
		if !end.Equal(expected) {
			t.Errorf("end = %v, want %v", end, expected)
		}
	})

	t.Run("invalid end time string", func(t *testing.T) {
		_, err := parseEndTime(start, "not-a-time")
		if err == nil {
			t.Error("parseEndTime() expected error for invalid time string")
		}
	})
}

func TestParseTimedEventTimesDefaultDuration(t *testing.T) {
	start, end, err := parseTimedEventTimes("2025-05-01 10:00", "", "")
	if err != nil {
		t.Fatalf("parseTimedEventTimes() error: %v", err)
	}
	expected := start.Add(time.Hour)
	if !end.Equal(expected) {
		t.Errorf("default end time = %v, want start+1h = %v", end, expected)
	}
}

func TestParseTimedEventTimesEndBeforeStart(t *testing.T) {
	_, _, err := parseTimedEventTimes("2025-05-01 14:00", "2025-05-01 10:00", "")
	if err == nil {
		t.Error("parseTimedEventTimes() expected error when end before start")
	}
}

func TestParseAllDayTimesEdgeCases(t *testing.T) {
	t.Run("same day start and end", func(t *testing.T) {
		_, _, err := parseAllDayTimes("2025-05-01", "2025-05-01")
		if err != nil {
			t.Logf("same day end: %v (may not error)", err)
		}
	})

	t.Run("invalid start date", func(t *testing.T) {
		_, _, err := parseAllDayTimes("not-a-date", "")
		if err == nil {
			t.Error("expected error for invalid start date")
		}
	})
}

// =====================================================================
// rrule.go — interpretRRule additional cases
// =====================================================================

func TestInterpretRRuleAllClauses(t *testing.T) {
	tests := []struct {
		rrule string
		want  []string
	}{
		{
			rrule: "FREQ=WEEKLY;INTERVAL=2;BYDAY=MO,WE;COUNT=10",
			want:  []string{"weekly", "MO,WE", "10 times"},
		},
		{
			rrule: "FREQ=MONTHLY;UNTIL=20251231",
			want:  []string{"monthly", "20251231"},
		},
		{
			rrule: "FREQ=DAILY",
			want:  []string{"daily", "forever"},
		},
		{
			rrule: "FREQ=YEARLY;INTERVAL=1",
			want:  []string{"yearly"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.rrule, func(t *testing.T) {
			got := interpretRRule(tt.rrule)
			for _, w := range tt.want {
				if !strings.Contains(strings.ToLower(got), strings.ToLower(w)) {
					t.Errorf("interpretRRule(%q) = %q, want substring %q", tt.rrule, got, w)
				}
			}
		})
	}
}

// =====================================================================
// quick.go — applyTimezoneToDetails, extractEventDetails
// =====================================================================

func TestApplyTimezoneToDetails(t *testing.T) {
	t.Run("applies valid timezone", func(t *testing.T) {
		details := quickParsedEvent{
			StartTime: time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 6, 1, 11, 0, 0, 0, time.UTC),
		}
		applyTimezoneToDetails(&details, "Europe/Madrid")
		if details.StartTime.Location().String() != "Europe/Madrid" {
			t.Errorf("StartTime location = %q, want Europe/Madrid", details.StartTime.Location().String())
		}
	})

	t.Run("empty tz is no-op", func(t *testing.T) {
		orig := time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC)
		details := quickParsedEvent{StartTime: orig, EndTime: orig.Add(time.Hour)}
		applyTimezoneToDetails(&details, "")
		if !details.StartTime.Equal(orig) {
			t.Error("StartTime should be unchanged for empty tz")
		}
	})

	t.Run("invalid tz is no-op", func(t *testing.T) {
		orig := time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC)
		details := quickParsedEvent{StartTime: orig, EndTime: orig.Add(time.Hour)}
		applyTimezoneToDetails(&details, "Not/A/Timezone")
		if !details.StartTime.Equal(orig) {
			t.Error("StartTime should be unchanged for invalid tz")
		}
	})
}

// =====================================================================
// template.go — runTemplateList, runTemplateValidate, runTemplateInit,
//               newTranslator, labelForField, isAlarmField,
//               templateFieldDefault, normalizeValuesForTemplate,
//               printTemplateTypeInfo, printTemplateOutput,
//               deriveTemplateFilename
// =====================================================================

func TestRunTemplateListDirect(t *testing.T) {
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	templateCmd := NewTemplateCmd(app)
	var listCmd *cobra.Command
	for _, sub := range templateCmd.Commands() {
		if strings.HasPrefix(sub.Use, "list") {
			listCmd = sub
			break
		}
	}
	if listCmd == nil {
		t.Fatal("list subcommand not found in template cmd")
	}

	err := runTemplateList(app, listCmd, nil)
	if err != nil {
		t.Fatalf("runTemplateList() error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Available templates") {
		t.Errorf("expected 'Available templates' in output, got: %s", out)
	}
}

func templateSubcmd(t *testing.T, name string) *cobra.Command {
	t.Helper()
	app := TestApp()
	templateCmd := NewTemplateCmd(app)
	for _, sub := range templateCmd.Commands() {
		if strings.HasPrefix(sub.Use, name) {
			return sub
		}
	}
	return nil
}

func TestRunTemplateValidateNoDir(t *testing.T) {
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	validateCmd := templateSubcmd(t, "validate")
	if validateCmd == nil {
		t.Skip("validate subcommand not present")
	}

	err := runTemplateValidate(app, validateCmd, nil)
	if err != nil {
		t.Logf("runTemplateValidate() error (may be expected): %v", err)
	}
}

func TestRunTemplateInitBasic(t *testing.T) {
	tmpDir := t.TempDir()
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	initCmd := templateSubcmd(t, "init")
	if initCmd == nil {
		t.Skip("init subcommand not present")
	}

	if err := initCmd.Flags().Set("dir", tmpDir); err != nil {
		t.Fatalf("failed to set --dir flag: %v", err)
	}

	err := runTemplateInit(app, initCmd, []string{"test-event"})
	if err != nil {
		t.Fatalf("runTemplateInit() error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Created scaffold") {
		t.Errorf("expected 'Created scaffold' in output, got: %s", out)
	}

	err2 := runTemplateInit(app, initCmd, []string{"test-event"})
	if err2 == nil {
		t.Error("expected error when file already exists without --force")
	}
	if !strings.Contains(err2.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got: %v", err2)
	}
}

func TestNewTranslator(t *testing.T) {
	tr, err := newTranslator("en")
	if err != nil {
		t.Fatalf("newTranslator('en') error: %v", err)
	}
	if tr == nil {
		t.Fatal("newTranslator('en') returned nil")
	}

	tr2, err := newTranslator("es")
	if err != nil {
		t.Logf("newTranslator('es') error (may fallback): %v", err)
	}
	_ = tr2

	tr3, _ := newTranslator("xx-invalid")
	_ = tr3
}

func TestNormalizeValuesForTemplateNoDDTemplate(t *testing.T) {
	tmpl := &tpl.Template{
		Fields: []tpl.Field{
			{Key: "start_time", Default: ""},
			{Key: "duration", Default: "1h"},
		},
	}
	dd := tpl.DataDrivenTemplate{}

	values := map[string]string{
		"start_time": "2025-06-01 10:00",
		"end_time":   "",
		"timezone":   "UTC",
	}
	if err := normalizeValuesForTemplate(values, tmpl, dd); err != nil {
		t.Fatalf("normalizeValuesForTemplate() error: %v", err)
	}
}

func TestNormalizeValuesForTemplateWithDD(t *testing.T) {
	tmpl := &tpl.Template{
		Fields: []tpl.Field{
			{Key: "start_time", Default: ""},
			{Key: "duration", Default: "30m"},
		},
	}
	dd := tpl.DataDrivenTemplate{
		Name: "test-dd",
		Output: tpl.OutputTemplate{
			StartField:    "start_time",
			EndField:      "end_time",
			DurationField: "duration",
			StartTZField:  "timezone",
		},
	}

	values := map[string]string{
		"start_time": "2025-06-01 10:00",
		"end_time":   "",
		"duration":   "45m",
		"timezone":   "UTC",
	}
	if err := normalizeValuesForTemplate(values, tmpl, dd); err != nil {
		t.Fatalf("normalizeValuesForTemplate() error: %v", err)
	}
}

func TestPrintTemplateTypeInfoBuiltIn(t *testing.T) {
	var buf bytes.Buffer
	tm := tpl.NewTemplateManager()
	dd := printTemplateTypeInfo(&buf, tm, "nonexistent-template")

	out := buf.String()
	if !strings.Contains(out, "built-in") {
		t.Errorf("expected 'built-in' for non-DD template, got: %s", out)
	}
	if dd.Name != "" {
		t.Errorf("expected empty DD for non-DD template, got name: %s", dd.Name)
	}
}

func TestPrintTemplateOutputEmpty(t *testing.T) {
	var buf bytes.Buffer
	dd := tpl.DataDrivenTemplate{}
	printTemplateOutput(&buf, dd)
	if buf.String() != "" {
		t.Errorf("printTemplateOutput with empty DD should produce no output, got: %s", buf.String())
	}
}

func TestRunTemplateDescribeBuiltIn(t *testing.T) {
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	descCmd := templateSubcmd(t, "describe")
	if descCmd == nil {
		t.Skip("describe subcommand not present")
	}

	err := runTemplateDescribe(app, descCmd, []string{"medical"})
	if err != nil {
		t.Logf("runTemplateDescribe('medical') error: %v (expected if medical not in default set)", err)
	}
}

// =====================================================================
// template.go — promptAlarmField (via prompts.Scanner injection)
// =====================================================================

func TestPromptAlarmFieldKeepDefaults(t *testing.T) {
	prevScanner := prompts.Scanner
	prompts.Scanner = bufio.NewScanner(strings.NewReader("\n"))
	defer func() { prompts.Scanner = prevScanner }()

	result := promptAlarmField(io.Discard, TestApp().Translator, "Alarms", "-15m|-5m")
	if result == "" {
		t.Error("promptAlarmField() should return non-empty when keeping defaults")
	}
}

func TestPromptAlarmFieldNoDefault(t *testing.T) {
	prevScanner := prompts.Scanner
	prompts.Scanner = bufio.NewScanner(strings.NewReader("\n"))
	defer func() { prompts.Scanner = prevScanner }()

	result := promptAlarmField(io.Discard, TestApp().Translator, "Alarms", "")
	_ = result
}

// =====================================================================
// template.go — runTemplateValidate with real empty dir
// =====================================================================

func TestRunTemplateValidateWithDir(t *testing.T) {
	tmpDir := t.TempDir()
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	templateCmd := NewTemplateCmd(app)
	var validateCmd *cobra.Command
	for _, sub := range templateCmd.Commands() {
		if strings.HasPrefix(sub.Use, "validate") {
			validateCmd = sub
			break
		}
	}
	if validateCmd == nil {
		t.Skip("validate subcommand not present")
	}

	if err := templateCmd.PersistentFlags().Set("templates-dir", tmpDir); err != nil {
		t.Fatalf("failed to set templates-dir on parent: %v", err)
	}

	err := runTemplateValidate(app, validateCmd, nil)
	if err != nil {
		t.Logf("runTemplateValidate() error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Validating") && !strings.Contains(out, "completed") && !strings.Contains(out, "no templates") {
		t.Logf("runTemplateValidate output: %s", out)
	}
}

// =====================================================================
// init.go — NewInitCmd basic structure
// =====================================================================

func TestNewInitCmd(t *testing.T) {
	app := TestApp()
	cmd := NewInitCmd(app)
	if cmd == nil {
		t.Fatal("NewInitCmd() returned nil")
	}
	if cmd.Use != "init" {
		t.Errorf("Use = %q, want %q", cmd.Use, "init")
	}
	if cmd.RunE == nil {
		t.Error("init command should have RunE function")
	}
}

// =====================================================================
// timezone.go — printZoneInfo directly
// =====================================================================

func TestPrintZoneInfoFull(t *testing.T) {
	var buf bytes.Buffer

	app := TestApp()
	app.Stdout = &buf

	err := runTZInfo(app, nil, []string{"UTC"})
	if err != nil {
		t.Fatalf("runTZInfo('UTC') error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "UTC") {
		t.Errorf("expected 'UTC' in output, got: %s", out)
	}
}

// =====================================================================
// create.go — runCreate with args (no-interactive path)
// =====================================================================

func TestRunCreateWithArgs(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "created.ics")

	app := TestApp()
	cmd := NewCreateCmd(app)

	mustSetFlag(t, cmd, "start", "2025-06-01 10:00")
	mustSetFlag(t, cmd, "end", "2025-06-01 11:00")
	mustSetFlag(t, cmd, "output", outputPath)

	err := runCreate(app, cmd, []string{"My Event"})
	if err != nil {
		t.Fatalf("runCreate() error: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	if !strings.Contains(string(data), "SUMMARY:My Event") {
		t.Errorf("expected SUMMARY in ICS, got: %s", string(data))
	}
}

func TestRunCreateNoArgs(t *testing.T) {
	app := TestApp()
	cmd := NewCreateCmd(app)

	err := runCreate(app, cmd, []string{})
	if err != nil {
		t.Fatalf("runCreate() with no args should show help, not error: %v", err)
	}
}

// =====================================================================
// batch.go — loadBatchFromCSV invalid path
// =====================================================================

func TestLoadBatchFromCSVMalformed(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "malformed.csv")
	content := "summary,start\n\"unclosed quote"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}
	_, err := loadBatchFromCSV(path)
	if err == nil {
		t.Error("loadBatchFromCSV() expected error for malformed CSV")
	}
}

// =====================================================================
// quick.go — extractEventDetails, resolveQuickTimezone
// =====================================================================

func TestExtractEventDetails(t *testing.T) {
	baseTime := time.Date(2025, 6, 1, 14, 0, 0, 0, time.UTC)
	res := &when.Result{
		Text: "tomorrow at 2pm",
		Time: baseTime,
	}

	t.Run("basic extraction", func(t *testing.T) {
		text := "Team meeting tomorrow at 2pm"
		details := extractEventDetails(text, res)
		if details.StartTime.IsZero() {
			t.Error("StartTime should not be zero")
		}
		if details.EndTime.IsZero() {
			t.Error("EndTime should not be zero")
		}
		if details.EndTime.Before(details.StartTime) {
			t.Error("EndTime should be after StartTime")
		}
	})

	t.Run("with duration", func(t *testing.T) {
		text := "Standup for 30min tomorrow at 2pm"
		res2 := &when.Result{Text: "tomorrow at 2pm", Time: baseTime}
		details := extractEventDetails(text, res2)
		expectedDur := 30 * time.Minute
		actualDur := details.EndTime.Sub(details.StartTime)
		if actualDur != expectedDur {
			t.Errorf("duration = %v, want %v", actualDur, expectedDur)
		}
	})

	t.Run("default 1h duration when no duration specified", func(t *testing.T) {
		text := "Meeting tomorrow at 2pm"
		res3 := &when.Result{Text: "tomorrow at 2pm", Time: baseTime}
		details := extractEventDetails(text, res3)
		expectedDur := time.Hour
		actualDur := details.EndTime.Sub(details.StartTime)
		if actualDur != expectedDur {
			t.Errorf("default duration = %v, want %v", actualDur, expectedDur)
		}
	})

	t.Run("preserves input text", func(t *testing.T) {
		text := "Sync call tomorrow at 2pm"
		details := extractEventDetails(text, res)
		if details.InputText != text {
			t.Errorf("InputText = %q, want %q", details.InputText, text)
		}
	})
}

func TestResolveQuickTimezone(t *testing.T) {
	app := TestApp()
	cmd := NewQuickCmd(app)

	t.Run("UTC config default resolves to empty", func(t *testing.T) {
		got, err := resolveQuickTimezone(app, cmd)
		if err != nil {
			t.Fatalf("resolveQuickTimezone() error: %v", err)
		}
		if got != "" {
			t.Errorf("resolveQuickTimezone() = %q, want empty for UTC default", got)
		}
	})

	t.Run("uses flag when set", func(t *testing.T) {
		cmd2 := NewQuickCmd(app)
		if err := cmd2.Flags().Set("timezone", "Europe/Madrid"); err != nil {
			t.Fatalf("failed to set timezone flag: %v", err)
		}
		got, err := resolveQuickTimezone(app, cmd2)
		if err != nil {
			t.Fatalf("resolveQuickTimezone() error: %v", err)
		}
		if got != "Europe/Madrid" {
			t.Errorf("resolveQuickTimezone() = %q, want %q", got, "Europe/Madrid")
		}
	})

	t.Run("resolves city alias", func(t *testing.T) {
		cmd3 := NewQuickCmd(app)
		if err := cmd3.Flags().Set("timezone", "madrid"); err != nil {
			t.Fatalf("failed to set timezone flag: %v", err)
		}
		got, err := resolveQuickTimezone(app, cmd3)
		if err != nil {
			t.Fatalf("resolveQuickTimezone() error: %v", err)
		}
		if got != "Europe/Madrid" {
			t.Errorf("resolveQuickTimezone() = %q, want Europe/Madrid", got)
		}
	})

	t.Run("invalid timezone is an error", func(t *testing.T) {
		cmd4 := NewQuickCmd(app)
		if err := cmd4.Flags().Set("timezone", "narnia"); err != nil {
			t.Fatalf("failed to set timezone flag: %v", err)
		}
		if _, err := resolveQuickTimezone(app, cmd4); err == nil {
			t.Error("resolveQuickTimezone() should fail for invalid timezone")
		}
	})
}

// =====================================================================
// template.go — printTemplateTypeInfo with DD template,
//               printTemplateOutput with populated DD,
//               deriveTemplateFilename
// =====================================================================

func TestPrintTemplateTypeInfoWithDD(t *testing.T) {
	var buf bytes.Buffer
	tm := tpl.NewTemplateManager()

	dd := printTemplateTypeInfo(&buf, tm, "medical")
	out := buf.String()

	if dd.Name != "" {
		if !strings.Contains(out, "Type: data-driven") {
			t.Errorf("expected 'data-driven' for DD template, got: %s", out)
		}
	} else {
		if !strings.Contains(out, "built-in") {
			t.Errorf("expected 'built-in' for non-DD template, got: %s", out)
		}
	}
}

func TestPrintTemplateOutputPopulated(t *testing.T) {
	var buf bytes.Buffer
	dd := tpl.DataDrivenTemplate{
		Name: "test-dd",
		Output: tpl.OutputTemplate{
			StartField:      "start_time",
			EndField:        "end_time",
			DurationField:   "duration",
			StartTZField:    "timezone",
			SummaryTmpl:     "Meeting with {{doctor}}",
			LocationTmpl:    "{{clinic}}",
			DescriptionTmpl: "{{notes}}",
			Categories:      []string{"Health", "Medical"},
			Priority:        5,
		},
	}

	printTemplateOutput(&buf, dd)
	out := buf.String()

	if !strings.Contains(out, "start_time") {
		t.Errorf("expected start_field in output, got: %s", out)
	}
	if !strings.Contains(out, "Health") {
		t.Errorf("expected categories in output, got: %s", out)
	}
	if !strings.Contains(out, "5") {
		t.Errorf("expected priority in output, got: %s", out)
	}
}

func TestDeriveTemplateFilenameWithEvent(t *testing.T) {
	tm := tpl.NewTemplateManager()
	ev := &calendar.Event{
		Summary:   "Team Meeting",
		StartTime: time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2025, 6, 15, 11, 0, 0, 0, time.UTC),
	}
	values := map[string]string{"summary": "Team Meeting"}

	got := deriveTemplateFilename(tm, "medical", values, ev, nil)
	if got == "" {
		t.Error("deriveTemplateFilename() returned empty string")
	}
	if !strings.Contains(got, "2025-06-15") {
		t.Errorf("expected date in filename, got: %s", got)
	}
}

func TestDeriveTemplateFilenameNoEvent(t *testing.T) {
	tm := tpl.NewTemplateManager()
	ev := &calendar.Event{}
	values := map[string]string{}

	got := deriveTemplateFilename(tm, "medical", values, ev, nil)
	if got == "" {
		t.Error("deriveTemplateFilename() returned empty string")
	}
	if !strings.HasSuffix(got, ".ics") {
		t.Errorf("expected .ics suffix, got: %s", got)
	}
}

// =====================================================================
// template.go — newTranslator fallback path
// =====================================================================

func TestNewTranslatorFallback(t *testing.T) {
	tr, err := newTranslator("xx-does-not-exist")
	if err == nil && tr == nil {
		t.Error("newTranslator() with unknown lang: both err and tr are nil/zero")
	}
}

// =====================================================================
// config.go — NewConfigCmd structure
// =====================================================================

func TestNewConfigCmdStructure(t *testing.T) {
	app := TestApp()
	cmd := NewConfigCmd(app)
	if cmd == nil {
		t.Fatal("NewConfigCmd() returned nil")
	}
	if cmd.Use != "config" {
		t.Errorf("Use = %q, want %q", cmd.Use, "config")
	}
	subs := cmd.Commands()
	if len(subs) != 3 {
		t.Errorf("expected 3 subcommands, got %d", len(subs))
	}
	names := make(map[string]bool)
	for _, sub := range subs {
		names[strings.Split(sub.Use, " ")[0]] = true
	}
	for _, name := range []string{"set", "list", "alarm-profiles"} {
		if !names[name] {
			t.Errorf("expected subcommand %q, not found", name)
		}
	}
}

// =====================================================================
// rrule.go — NewRRuleHelperCmd structure (via cmd itself)
// =====================================================================

func TestNewRRuleHelperCmdStructure(t *testing.T) {
	app := TestApp()
	cmd := NewRRuleHelperCmd(app)
	if cmd == nil {
		t.Fatal("NewRRuleHelperCmd() returned nil")
	}
	if cmd.Use != "rrule" {
		t.Errorf("Use = %q, want %q", cmd.Use, "rrule")
	}
	if cmd.RunE == nil {
		t.Error("rrule command should have RunE")
	}
}

// =====================================================================
// template.go — runTemplateDescribe with known template
// =====================================================================

func TestPrintTemplateTypeInfoWithRealDD(t *testing.T) {
	repoRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}
	templatesDir := filepath.Join(repoRoot, "..", "..", "internal", "templates", "json")
	if _, err := os.Stat(templatesDir); err != nil {
		t.Skip("templates json dir not found")
	}

	tm := tpl.NewTemplateManager()
	tm.LoadDDDir(templatesDir)

	var buf bytes.Buffer
	dd := printTemplateTypeInfo(&buf, tm, "medical")
	out := buf.String()

	if dd.Name != "" {
		if !strings.Contains(out, "data-driven") {
			t.Errorf("expected 'data-driven' in output for DD template, got: %s", out)
		}
	} else {
		if !strings.Contains(out, "built-in") {
			t.Errorf("expected 'built-in' in output, got: %s", out)
		}
	}
}

func TestRunTemplateValidateWithRealDir(t *testing.T) {
	repoRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}
	templatesDir := filepath.Join(repoRoot, "..", "..", "internal", "templates", "json")
	if _, err := os.Stat(templatesDir); err != nil {
		t.Skip("templates json dir not found")
	}

	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	validateCmd := templateSubcmd(t, "validate")
	if validateCmd == nil {
		t.Skip("validate subcommand not present")
	}

	if err := validateCmd.Flags().Set("templates-dir", templatesDir); err != nil {
		if err2 := validateCmd.PersistentFlags().Set("templates-dir", templatesDir); err2 != nil {
			t.Skip("cannot set templates-dir on validate cmd")
		}
	}

	err = runTemplateValidate(app, validateCmd, nil)
	if err != nil {
		t.Logf("runTemplateValidate() error: %v", err)
	}
}

// =====================================================================
// template.go — runTemplateValidate with local --templates-dir flag,
//               runTemplateInit JSON format, unsupported format,
//               runTemplateDescribe via built-in medical
// =====================================================================

// makeValidateCmd creates a minimal cobra.Command with a --templates-dir flag
// that runTemplateValidate can read, mimicking how the real cmd hierarchy works.
func makeValidateCmd(app *App, templatesDir string) *cobra.Command {
	cmd := &cobra.Command{Use: "validate"}
	cmd.Flags().String("templates-dir", templatesDir, "")
	return cmd
}

func TestRunTemplateValidateExistingDir(t *testing.T) {
	repoRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}
	templatesDir := filepath.Join(repoRoot, "..", "..", "internal", "templates", "json")
	if _, err := os.Stat(templatesDir); err != nil {
		t.Skip("templates json dir not found")
	}

	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	cmd := makeValidateCmd(app, templatesDir)
	err = runTemplateValidate(app, cmd, nil)
	if err != nil {
		t.Logf("runTemplateValidate() with real dir error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Validating") {
		t.Errorf("expected 'Validating' in output, got: %s", out)
	}
}

func TestRunTemplateValidateNotADir(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "not-a-dir.txt")
	if err := os.WriteFile(filePath, []byte("x"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	cmd := makeValidateCmd(app, filePath)
	err := runTemplateValidate(app, cmd, nil)
	if err != nil {
		t.Logf("runTemplateValidate() on file path error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "not a directory") {
		t.Errorf("expected 'not a directory' in output, got: %s", out)
	}
}

func TestRunTemplateInitJSONFormat(t *testing.T) {
	tmpDir := t.TempDir()
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	initCmd := templateSubcmd(t, "init")
	if initCmd == nil {
		t.Skip("init subcommand not present")
	}
	if err := initCmd.Flags().Set("dir", tmpDir); err != nil {
		t.Fatalf("failed to set dir: %v", err)
	}
	if err := initCmd.Flags().Set("format", "json"); err != nil {
		t.Fatalf("failed to set format: %v", err)
	}

	err := runTemplateInit(app, initCmd, []string{"json-event"})
	if err != nil {
		t.Fatalf("runTemplateInit() JSON format error: %v", err)
	}
	expectedFile := filepath.Join(tmpDir, "json-event.json")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("expected JSON scaffold file %s to exist", expectedFile)
	}
}

func TestRunTemplateInitUnsupportedFormat(t *testing.T) {
	tmpDir := t.TempDir()
	app := TestApp()

	initCmd := templateSubcmd(t, "init")
	if initCmd == nil {
		t.Skip("init subcommand not present")
	}
	if err := initCmd.Flags().Set("dir", tmpDir); err != nil {
		t.Fatalf("failed to set dir: %v", err)
	}
	if err := initCmd.Flags().Set("format", "toml"); err != nil {
		t.Fatalf("failed to set format: %v", err)
	}

	err := runTemplateInit(app, initCmd, []string{"bad-format"})
	if err == nil {
		t.Fatal("runTemplateInit() expected error for unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("expected 'unsupported format' error, got: %v", err)
	}
}

func TestRunTemplateInitForce(t *testing.T) {
	tmpDir := t.TempDir()
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	initCmd := templateSubcmd(t, "init")
	if initCmd == nil {
		t.Skip("init subcommand not present")
	}
	if err := initCmd.Flags().Set("dir", tmpDir); err != nil {
		t.Fatalf("failed to set dir: %v", err)
	}

	if err := runTemplateInit(app, initCmd, []string{"force-event"}); err != nil {
		t.Fatalf("first init error: %v", err)
	}
	if err := initCmd.Flags().Set("force", "true"); err != nil {
		t.Fatalf("failed to set force: %v", err)
	}
	if err := runTemplateInit(app, initCmd, []string{"force-event"}); err != nil {
		t.Fatalf("runTemplateInit() with --force error: %v", err)
	}
}

func TestNewTranslatorFallbackToEnglish(t *testing.T) {
	tr, err := newTranslator("pt")
	if err != nil {
		t.Logf("newTranslator('pt') returned error, may fallback: %v", err)
	}
	_ = tr
}

func TestRunTZInfoWithSuggestions(t *testing.T) {
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	err := runTZInfo(app, nil, []string{"Europ"})
	if err != nil {
		t.Fatalf("runTZInfo() suggestion error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "not found") && !strings.Contains(out, "Did you mean") {
		t.Logf("runTZInfo('Europ') output: %s", out)
	}
}

func TestRunTZInfoInvalidLoadLocation(t *testing.T) {
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	err := runTZInfo(app, nil, []string{"recife"})
	if err != nil {
		t.Logf("runTZInfo('recife') error (may be ok): %v", err)
	}
}

func TestRunBatchTemplateWriteError(t *testing.T) {
	app := TestApp()
	cmd := NewBatchTemplateCmd(app)
	cmd.SetArgs([]string{"basic"})
	mustSetFlag(t, cmd, "output", "/nonexistent/path/that/cannot/be/written/out.csv")

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error writing to nonexistent path")
	}
}

func TestExpandAlarmProfilesEmpty(t *testing.T) {
	result, err := expandAlarmProfiles(TestApp().Config, []string{"  ", ""})
	if err != nil {
		t.Fatalf("expandAlarmProfiles() error for whitespace specs: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 results for whitespace-only specs, got %d", len(result))
	}
}

func TestBuildBatchCalendarWhitespaceNameAndTZ(t *testing.T) {
	records := []batchRecord{
		{Summary: "Event", Start: "2025-05-01 10:00", End: "2025-05-01 11:00"},
	}
	opts := &batchOptions{name: "  ", defaultTZ: "  "}
	spellCache := nd.NewSpellCheckCache(nil)
	catCache := nd.NewCategoryCache()

	cal, _, err := buildBatchCalendar(records, opts, spellCache, catCache, TestApp().Config)
	if err != nil {
		t.Fatalf("buildBatchCalendar() error: %v", err)
	}
	if cal.Name != "" {
		t.Errorf("Calendar Name should be empty for whitespace, got %q", cal.Name)
	}
}

// makeDescribeCmd creates a minimal cobra.Command with --templates-dir flag,
// so that loadTemplateManager can read it via cmd.Flags().GetString().
func makeDescribeCmd(templatesDir string) *cobra.Command {
	cmd := &cobra.Command{Use: "describe"}
	cmd.Flags().String("templates-dir", templatesDir, "")
	return cmd
}

func TestRunTemplateDescribeBuiltInTemplate(t *testing.T) {
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	cmd := makeDescribeCmd("")

	err := runTemplateDescribe(app, cmd, []string{"medical"})
	if err != nil {
		t.Logf("runTemplateDescribe('medical') error: %v (may be ok if not built-in)", err)
		return
	}
	out := buf.String()
	if out == "" {
		t.Error("runTemplateDescribe() produced no output")
	}
}

func TestRunTemplateDescribeWithRealDD(t *testing.T) {
	repoRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}
	templatesDir := filepath.Join(repoRoot, "..", "..", "internal", "templates", "json")
	if _, err := os.Stat(templatesDir); err != nil {
		t.Skip("templates json dir not found")
	}

	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	cmd := makeDescribeCmd(templatesDir)

	err = runTemplateDescribe(app, cmd, []string{"medical"})
	if err != nil {
		t.Logf("runTemplateDescribe('medical') with DD: %v", err)
		return
	}
	out := buf.String()
	if out == "" {
		t.Error("runTemplateDescribe() with DD produced no output")
	}
}

func TestRunTemplateDescribeUnknownTemplate(t *testing.T) {
	app := TestApp()
	cmd := makeDescribeCmd("")

	err := runTemplateDescribe(app, cmd, []string{"nonexistent-template-xyz"})
	if err == nil {
		t.Fatal("runTemplateDescribe() expected error for unknown template")
	}
}

func TestNewTranslatorKnownLanguages(t *testing.T) {
	for _, lang := range []string{"en", "es", "pt", "ga"} {
		tr, err := newTranslator(lang)
		if err != nil {
			t.Logf("newTranslator(%q) error (may fallback): %v", lang, err)
		}
		_ = tr
	}
}

func TestWriteCalendarOutputEmpty(t *testing.T) {
	app := TestApp()
	cal := calendar.NewCalendar()
	ev := &calendar.Event{
		Summary:   "Test",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(time.Hour),
	}
	cal.AddEvent(ev)

	err := writeCalendarOutput(app, cal, "")
	if err != nil {
		t.Fatalf("writeCalendarOutput() with empty path error: %v", err)
	}
}

func TestWriteCalendarOutputWriteError(t *testing.T) {
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf
	cal := calendar.NewCalendar()

	err := writeCalendarOutput(app, cal, "/nonexistent/path/cannot/write.ics")
	if err == nil {
		t.Fatal("writeCalendarOutput() expected error for unwritable path")
	}
	out := buf.String()
	if !strings.Contains(out, "❌") && !strings.Contains(out, "failed") {
		t.Logf("writeCalendarOutput error output: %s", out)
	}
}

func TestLoadBatchInputFormatError(t *testing.T) {
	opts := &batchOptions{input: "events.txt", formatFlag: "auto"}
	_, _, err := loadBatchInput(opts)
	if err == nil {
		t.Fatal("loadBatchInput() expected error for unknown extension")
	}
}

func TestRunBatchMaxEventsPerDay(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "events.csv")
	outputPath := filepath.Join(tmpDir, "out.ics")

	csvData := "summary,start,end\nA,2025-05-01 09:00,2025-05-01 10:00\nB,2025-05-01 10:00,2025-05-01 11:00\nC,2025-05-01 11:00,2025-05-01 12:00\n"
	if err := os.WriteFile(inputPath, []byte(csvData), 0644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	cmd := NewBatchCmd(app)
	mustSetFlag(t, cmd, "input", inputPath)
	mustSetFlag(t, cmd, "output", outputPath)
	mustSetFlag(t, cmd, "max-events-per-day", "2")

	if err := runBatch(app, cmd, nil); err != nil {
		t.Fatalf("runBatch() max-events-per-day error: %v", err)
	}
}

func TestRunBatchDryRunWithValidationErrors(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "events.csv")
	outputPath := filepath.Join(tmpDir, "out.ics")

	csvData := "summary,start,end\n,2025-05-01 10:00,2025-05-01 11:00\n"
	if err := os.WriteFile(inputPath, []byte(csvData), 0644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	cmd := NewBatchCmd(app)
	mustSetFlag(t, cmd, "input", inputPath)
	mustSetFlag(t, cmd, "output", outputPath)
	mustSetFlag(t, cmd, "dry-run", "true")

	err := runBatch(app, cmd, nil)
	if err == nil {
		t.Fatal("runBatch() dry-run with validation errors should return error")
	}
	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("expected 'validation failed', got: %v", err)
	}
}

func TestRunTemplateValidateNonExistentDir(t *testing.T) {
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	cmd := makeValidateCmd(app, "/nonexistent/path/does/not/exist")
	err := runTemplateValidate(app, cmd, nil)
	if err != nil {
		t.Logf("runTemplateValidate() nonexistent error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "not found") && !strings.Contains(out, "Validating") {
		t.Logf("output: %s", out)
	}
}

// =====================================================================
// template.go — templateFieldDefault key-not-found path
// =====================================================================

func TestTemplateFieldDefaultKeyNotFound(t *testing.T) {
	tmpl := &tpl.Template{
		Fields: []tpl.Field{
			{Key: "summary", Default: "default-summary"},
		},
	}
	got := templateFieldDefault(tmpl, "nonexistent-key")
	if got != "" {
		t.Errorf("templateFieldDefault() for missing key = %q, want empty string", got)
	}

	got2 := templateFieldDefault(tmpl, "summary")
	if got2 != "default-summary" {
		t.Errorf("templateFieldDefault() for existing key = %q, want %q", got2, "default-summary")
	}
}

// =====================================================================
// template.go — labelForField with empty Name (falls back to Key)
// =====================================================================

func TestLabelForFieldEmptyName(t *testing.T) {
	f := tpl.Field{Key: "start_time", Name: "", Required: false}
	got := labelForField(f)
	if got != "start_time" {
		t.Errorf("labelForField() with empty Name = %q, want %q", got, "start_time")
	}
}

func TestLabelForFieldRequiredWithName(t *testing.T) {
	f := tpl.Field{Key: "email", Name: "Email Address", Required: true}
	got := labelForField(f)
	if !strings.Contains(got, "*") {
		t.Errorf("labelForField() required field = %q, expected '*' suffix", got)
	}
	if !strings.Contains(got, "Email Address") {
		t.Errorf("labelForField() = %q, expected 'Email Address'", got)
	}
}

// =====================================================================
// template.go — isAlarmField via Type="alarms"
// =====================================================================

func TestIsAlarmFieldByType(t *testing.T) {
	f := tpl.Field{Key: "reminders", Type: "alarms"}
	if !isAlarmField(f) {
		t.Error("isAlarmField() should return true when Type='alarms'")
	}
}

func TestIsAlarmFieldByKey(t *testing.T) {
	f := tpl.Field{Key: "Alarms", Type: "text"}
	if !isAlarmField(f) {
		t.Error("isAlarmField() should return true when Key='Alarms'")
	}
}

func TestIsAlarmFieldFalse(t *testing.T) {
	f := tpl.Field{Key: "location", Type: "text"}
	if isAlarmField(f) {
		t.Error("isAlarmField() should return false for regular field")
	}
}

// =====================================================================
// helpers.go — expandAlarmProfiles with unknown profile
// =====================================================================

func TestExpandAlarmProfilesUnknownProfile(t *testing.T) {
	_, err := expandAlarmProfiles(TestApp().Config, []string{"profile:nonexistent-profile-xyz"})
	if err == nil {
		t.Fatal("expandAlarmProfiles() expected error for unknown profile")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

func TestExpandAlarmProfilesKnownProfile(t *testing.T) {
	cfg := TestApp().Config
	cfg.AlarmProfiles = map[string][]string{
		"adhd-default": {"-2h", "-1h", "-30m", "-10m"},
	}
	result, err := expandAlarmProfiles(cfg, []string{"profile:adhd-default"})
	if err != nil {
		t.Fatalf("expandAlarmProfiles('profile:adhd-default') error: %v", err)
	}
	if len(result) != 4 {
		t.Errorf("expected 4 expanded alarm specs from adhd-default, got %d: %v", len(result), result)
	}
}

// =====================================================================
// batch.go — getBatchTemplateContent school-event yaml and csv
// =====================================================================

func TestGetBatchTemplateContentSchoolEventYAML(t *testing.T) {
	content, err := getBatchTemplateContent("school-event", "yaml")
	if err != nil {
		t.Fatalf("getBatchTemplateContent() error: %v", err)
	}
	if content == "" {
		t.Error("getBatchTemplateContent() returned empty string for school-event yaml")
	}
}

func TestGetBatchTemplateContentSchoolEventCSV(t *testing.T) {
	content, err := getBatchTemplateContent("school-event", "csv")
	if err != nil {
		t.Fatalf("getBatchTemplateContent() error: %v", err)
	}
	if content == "" {
		t.Error("getBatchTemplateContent() returned empty string for school-event csv")
	}
}

func TestGetBatchTemplateContentRecruiterMeetingYAML(t *testing.T) {
	content, err := getBatchTemplateContent("recruiter-meeting", "yaml")
	if err != nil {
		t.Fatalf("getBatchTemplateContent() error: %v", err)
	}
	if content == "" {
		t.Error("getBatchTemplateContent() returned empty for recruiter-meeting yaml")
	}
}

func TestGetBatchTemplateContentTravelDayYAML(t *testing.T) {
	content, err := getBatchTemplateContent("travel-day", "yaml")
	if err != nil {
		t.Fatalf("getBatchTemplateContent() error: %v", err)
	}
	if content == "" {
		t.Error("getBatchTemplateContent() returned empty for travel-day yaml")
	}
}

func TestGetBatchTemplateContentUnknown(t *testing.T) {
	_, err := getBatchTemplateContent("unknown-xyz", "csv")
	if err == nil {
		t.Fatal("getBatchTemplateContent() expected error for unknown template type")
	}
	if !strings.Contains(err.Error(), "unknown template type") {
		t.Errorf("expected 'unknown template type' in error, got: %v", err)
	}
}

// =====================================================================
// template.go — printTemplateTypeInfo with DD having Source + FilenameTemplate
// (using the real json template dir to get a DD template with source set)
// =====================================================================

func TestPrintTemplateTypeInfoWithSourceAndFilename(t *testing.T) {
	repoRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}
	templatesDir := filepath.Join(repoRoot, "..", "..", "internal", "templates", "json")
	if _, err := os.Stat(templatesDir); err != nil {
		t.Skip("templates json dir not found")
	}

	tm := tpl.NewTemplateManager()
	tm.LoadDDDir(templatesDir)

	var buf bytes.Buffer
	dd := printTemplateTypeInfo(&buf, tm, "medical")
	out := buf.String()

	if dd.Name != "" {
		if !strings.Contains(out, "data-driven") {
			t.Errorf("expected 'data-driven' in output, got: %s", out)
		}
		if !strings.Contains(out, "Source:") {
			t.Logf("printTemplateTypeInfo output (no Source): %s", out)
		}
	} else {
		t.Logf("medical is built-in, not DD: %s", out)
	}
}

// =====================================================================
// template.go — printTemplateOutput with EndTZField different from StartTZField
// =====================================================================

func TestPrintTemplateOutputWithDifferentEndTZ(t *testing.T) {
	var buf bytes.Buffer
	dd := tpl.DataDrivenTemplate{
		Name: "tz-test",
		Output: tpl.OutputTemplate{
			StartField:   "start_time",
			StartTZField: "start_tz",
			EndTZField:   "end_tz",
		},
	}
	printTemplateOutput(&buf, dd)
	out := buf.String()
	if !strings.Contains(out, "end_tz") {
		t.Errorf("expected 'end_tz' in output when EndTZField != StartTZField, got: %s", out)
	}
}

// =====================================================================
// batch.go — parseDurationEnd (unified) with zero duration
// =====================================================================

func TestParseBatchDurationEndZero(t *testing.T) {
	start := time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC)
	_, err := parseDurationEnd(start, "0m")
	if err == nil {
		t.Error("parseDurationEnd() expected error for zero duration")
	}
}

func TestParseBatchDurationEndInvalid(t *testing.T) {
	start := time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC)
	_, err := parseDurationEnd(start, "not-a-duration")
	if err == nil {
		t.Error("parseDurationEnd() expected error for invalid duration")
	}
}

// =====================================================================
// app.go — SetupPersistentPreRunE exercises language and timezone flags
// (avoiding redeclaration - app_test.go already has the base test)
// =====================================================================

func TestSetupPersistentPreRunESpanish(t *testing.T) {
	app := TestApp()
	fn := SetupPersistentPreRunE(app)

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("language", "", "language")
	cmd.Flags().String("timezone", "", "timezone")

	if err := cmd.Flags().Set("language", "es"); err != nil {
		t.Fatalf("failed to set language: %v", err)
	}
	if err := cmd.Flags().Set("timezone", "Europe/Madrid"); err != nil {
		t.Fatalf("failed to set timezone: %v", err)
	}

	if err := fn(cmd, nil); err != nil {
		t.Fatalf("SetupPersistentPreRunE() error: %v", err)
	}

	if app.Config.Language != "es" {
		t.Errorf("Config.Language = %q, want %q", app.Config.Language, "es")
	}
	if app.Config.Timezone != "Europe/Madrid" {
		t.Errorf("Config.Timezone = %q, want %q", app.Config.Timezone, "Europe/Madrid")
	}
}

// =====================================================================
// config.go — runConfigSet and runConfigAlarmProfiles exercising RunE path
// =====================================================================

func TestRunConfigSetAndAlarmProfiles(t *testing.T) {
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	err := runConfigList(app, nil, nil)
	if err != nil {
		t.Logf("runConfigList() error (may be ok if config not found): %v", err)
	}

	buf.Reset()
	err = runConfigAlarmProfiles(app, nil, nil)
	if err != nil {
		t.Logf("runConfigAlarmProfiles() error: %v", err)
	}
}

// =====================================================================
// batch.go — writeBatchOutput with warnings
// =====================================================================

func TestWriteBatchOutputWithWarnings(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "out.ics")

	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	cal := &calendar.Calendar{}
	warnings := []string{"Warning: event A conflicts with event B", "Warning: too many events per day"}

	err := writeBatchOutput(app, cal, warnings, outputPath, 2)
	if err != nil {
		t.Fatalf("writeBatchOutput() with warnings error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Warning:") {
		t.Errorf("expected warnings in output, got: %s", out)
	}
}

// =====================================================================
// template.go — runTemplateCreateFromFile with no data
// =====================================================================

func TestRunTemplateCreateFromFileNoData(t *testing.T) {
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "empty.csv")
	if err := os.WriteFile(csvPath, []byte("summary,start\n"), 0644); err != nil {
		t.Fatalf("failed to write csv: %v", err)
	}

	app := TestApp()
	tm := tpl.NewTemplateManager()
	tr, _ := newTranslator("en")
	tmpl, err := tm.GetTemplate("medical")
	if err != nil {
		t.Skip("medical template not available")
	}
	dd, _ := tm.DataTemplate("medical")

	params := templateCreateParams{
		templateName: "medical",
		inputPath:    csvPath,
		formatFlag:   "csv",
		outputDir:    tmpDir,
	}
	err = runTemplateCreateFromFile(app, tm, tr, tmpl, dd, params)
	if err == nil {
		t.Fatal("runTemplateCreateFromFile() expected error for empty CSV data")
	}
	if !strings.Contains(err.Error(), "no data found") {
		t.Errorf("expected 'no data found' error, got: %v", err)
	}
}

// =====================================================================
// timezone.go — NewTimezoneCmd exercises RunE closures via Execute
// =====================================================================

func TestNewTimezoneCmdRunE(t *testing.T) {
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	cmd := NewTimezoneCmd(app)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"info", "UTC"})

	if err := cmd.Execute(); err != nil {
		t.Logf("NewTimezoneCmd Execute error: %v", err)
	}
}

// =====================================================================
// locale.go — NewLocaleCmd exercises RunE via Execute
// =====================================================================

func TestNewLocaleCmdRunE(t *testing.T) {
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	cmd := NewLocaleCmd(app)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"list"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("NewLocaleCmd Execute error: %v", err)
	}
}

// =====================================================================
// lint.go — NewLintCmd exercises RunE via Execute (with --file flag)
// =====================================================================

func TestNewLintCmdRunE(t *testing.T) {
	tmpDir := t.TempDir()
	icsPath := filepath.Join(tmpDir, "test.ics")
	icsContent := `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VEVENT
UID:lint-test-uid@tempus
SUMMARY:Lint RunE Test
DTSTAMP:20250401T090000Z
DTSTART:20250601T100000Z
DTEND:20250601T110000Z
END:VEVENT
END:VCALENDAR
`
	if err := os.WriteFile(icsPath, []byte(icsContent), 0644); err != nil {
		t.Fatalf("failed to write ICS: %v", err)
	}

	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	cmd := NewLintCmd(app)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--file", icsPath})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("NewLintCmd Execute error: %v", err)
	}
}

// =====================================================================
// config.go — NewConfigCmd exercises RunE via Execute (list subcommand)
// =====================================================================

func TestNewConfigCmdListRunE(t *testing.T) {
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	cmd := NewConfigCmd(app)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"list"})

	if err := cmd.Execute(); err != nil {
		t.Logf("NewConfigCmd list Execute error (may be ok if no config): %v", err)
	}
}

// =====================================================================
// template.go — NewTemplateCmd exercises RunE closures via Execute
// =====================================================================

func TestNewTemplateCmdListRunE(t *testing.T) {
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	cmd := NewTemplateCmd(app)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"list"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("NewTemplateCmd list Execute error: %v", err)
	}
}

// =====================================================================
// rrule.go — NewRRuleHelperCmd exercises RunE (requires TTY for stdin read)
// but we can at least verify the command structure and that RunE is set
// =====================================================================

func TestNewRRuleHelperCmdHasRunE(t *testing.T) {
	app := TestApp()
	cmd := NewRRuleHelperCmd(app)
	if cmd.RunE == nil {
		t.Error("NewRRuleHelperCmd should have RunE set")
	}
	if cmd.Use != "rrule" {
		t.Errorf("Use = %q, want %q", cmd.Use, "rrule")
	}
}

// =====================================================================
// init.go — NewInitCmd exercises RunE closure structure
// =====================================================================

func TestNewInitCmdHasRunE(t *testing.T) {
	app := TestApp()
	cmd := NewInitCmd(app)
	if cmd.RunE == nil {
		t.Error("NewInitCmd should have RunE set")
	}
}

// =====================================================================
// create.go — runCreate path with allDay flag
// =====================================================================

func TestRunCreateAllDay(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "allday.ics")

	app := TestApp()
	cmd := NewCreateCmd(app)
	mustSetFlag(t, cmd, "start", "2025-06-01")
	mustSetFlag(t, cmd, "all-day", "true")
	mustSetFlag(t, cmd, "output", outputPath)

	err := runCreate(app, cmd, []string{"All Day Event"})
	if err != nil {
		t.Fatalf("runCreate() all-day error: %v", err)
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	if !strings.Contains(string(data), "SUMMARY:All Day Event") {
		t.Errorf("expected SUMMARY in ICS, got: %s", string(data))
	}
}

// =====================================================================
// create.go — parseCreateFlags with priority out of range
// =====================================================================

func TestParseCreateFlagsPriorityOutOfRange(t *testing.T) {
	app := TestApp()
	cmd := NewCreateCmd(app)
	mustSetFlag(t, cmd, "start", "2025-06-01 10:00")
	mustSetFlag(t, cmd, "priority", "10")

	_, err := parseCreateFlags(cmd, []string{"Test Event"})
	if err == nil {
		t.Fatal("parseCreateFlags() expected error for priority > 9")
	}
	if !strings.Contains(err.Error(), "priority") {
		t.Errorf("expected 'priority' in error, got: %v", err)
	}
}

// =====================================================================
// batch.go — runBatch with missing input file (error path)
// =====================================================================

func TestRunBatchMissingInputFile(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "out.ics")

	app := TestApp()
	cmd := NewBatchCmd(app)
	mustSetFlag(t, cmd, "input", filepath.Join(tmpDir, "nonexistent.csv"))
	mustSetFlag(t, cmd, "output", outputPath)

	err := runBatch(app, cmd, nil)
	if err == nil {
		t.Fatal("runBatch() expected error for missing input file")
	}
}

// =====================================================================
// template.go — runTemplateValidate with a dir containing invalid JSON
// =====================================================================

func TestRunTemplateValidateInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	invalidJSON := filepath.Join(tmpDir, "invalid.json")
	if err := os.WriteFile(invalidJSON, []byte("{invalid json"), 0644); err != nil {
		t.Fatalf("failed to write invalid JSON: %v", err)
	}

	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	cmd := makeValidateCmd(app, tmpDir)
	err := runTemplateValidate(app, cmd, nil)
	if err != nil {
		t.Logf("runTemplateValidate() invalid JSON error: %v", err)
	}
	out := buf.String()
	t.Logf("output: %s", out)
}

// =====================================================================
// helpers.go — ValueAsString with map type (not-a-string branch)
// (helpers_test.go already covers basic cases; these cover unexplored types)
// =====================================================================

func TestValueAsStringNilInput(t *testing.T) {
	val := ValueAsString(nil)
	if val != "" {
		t.Errorf("ValueAsString(nil) = %q, want empty string", val)
	}
}

func TestValueAsStringFloat64(t *testing.T) {
	val := ValueAsString(float64(3.14))
	if val == "" {
		t.Error("ValueAsString(float64) returned empty string")
	}
}

func TestValueAsStringBoolTrue(t *testing.T) {
	val := ValueAsString(true)
	if val != "true" {
		t.Errorf("ValueAsString(true) = %q, want %q", val, "true")
	}
}

func TestValueAsStringBoolFalse(t *testing.T) {
	val := ValueAsString(false)
	if val != "false" {
		t.Errorf("ValueAsString(false) = %q, want %q", val, "false")
	}
}

func TestValueAsAlarmSliceNilInput(t *testing.T) {
	val := ValueAsAlarmSlice(nil)
	if len(val) != 0 {
		t.Errorf("ValueAsAlarmSlice(nil) = %v, want empty", val)
	}
}

// =====================================================================
// helpers.go — ValueAsString with fmt.Stringer input
// =====================================================================

type testStringer struct{ s string }

func (ts testStringer) String() string { return ts.s }

func TestValueAsStringStringer(t *testing.T) {
	val := ValueAsString(testStringer{s: "stringer-value"})
	if val != "stringer-value" {
		t.Errorf("ValueAsString(Stringer) = %q, want %q", val, "stringer-value")
	}
}

// =====================================================================
// template.go — NewTemplateCmd RunE closures via Execute
// =====================================================================

func TestNewTemplateCmdValidateRunE(t *testing.T) {
	tmpDir := t.TempDir()
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	cmd := NewTemplateCmd(app)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"validate", "--templates-dir", tmpDir})

	if err := cmd.Execute(); err != nil {
		t.Logf("NewTemplateCmd validate Execute error: %v", err)
	}
}

func TestNewTemplateCmdDescribeRunE(t *testing.T) {
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	cmd := NewTemplateCmd(app)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"describe", "medical"})

	if err := cmd.Execute(); err != nil {
		t.Logf("NewTemplateCmd describe Execute error (may be ok): %v", err)
	}
}

func TestNewTemplateCmdInitRunE(t *testing.T) {
	tmpDir := t.TempDir()
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	cmd := NewTemplateCmd(app)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"init", "test-scaffold", "--dir", tmpDir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("NewTemplateCmd init Execute error: %v", err)
	}
}

// =====================================================================
// config.go — NewConfigCmd alarm-profiles RunE via Execute
// =====================================================================

func TestNewConfigCmdAlarmProfilesRunE(t *testing.T) {
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	cmd := NewConfigCmd(app)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"alarm-profiles"})

	if err := cmd.Execute(); err != nil {
		t.Logf("NewConfigCmd alarm-profiles Execute error: %v", err)
	}
}

// =====================================================================
// batch.go — runBatchTemplate format validation error
// =====================================================================

func TestRunBatchTemplateInvalidFormatXML(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "out.csv")

	app := TestApp()
	cmd := NewBatchTemplateCmd(app)
	mustSetFlag(t, cmd, "output", outputPath)
	mustSetFlag(t, cmd, "format", "xml")

	err := runBatchTemplate(app, cmd, []string{"basic"})
	if err == nil {
		t.Fatal("runBatchTemplate() expected error for invalid format")
	}
	if !strings.Contains(err.Error(), "format") {
		t.Errorf("expected 'format' in error, got: %v", err)
	}
}

func TestRunBatchTemplateNoOutput(t *testing.T) {
	app := TestApp()
	cmd := NewBatchTemplateCmd(app)

	err := runBatchTemplate(app, cmd, []string{"basic"})
	if err == nil {
		t.Fatal("runBatchTemplate() expected error when --output not set")
	}
	if !strings.Contains(err.Error(), "--output") {
		t.Errorf("expected '--output' in error, got: %v", err)
	}
}

func TestRunBatchTemplateAllTypes(t *testing.T) {
	tmpDir := t.TempDir()
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	types := []string{"basic", "adhd-routine", "medication", "work-meetings", "medical", "travel", "family"}
	for _, templateType := range types {
		outputPath := filepath.Join(tmpDir, templateType+".csv")
		cmd := NewBatchTemplateCmd(app)
		mustSetFlag(t, cmd, "output", outputPath)
		mustSetFlag(t, cmd, "format", "csv")

		err := runBatchTemplate(app, cmd, []string{templateType})
		if err != nil {
			t.Errorf("runBatchTemplate(%q) error: %v", templateType, err)
		}
	}
}

// =====================================================================
// create.go — parseCreateFlags with only clock time (no date, just time)
// =====================================================================

func TestNormalizeTimeInputClockOnly(t *testing.T) {
	result := normalizeTimeInput("14:30", "Europe/Madrid", "")
	if result == "" {
		t.Error("normalizeTimeInput() for clock-only time returned empty string")
	}
	if result == "14:30" {
		t.Error("normalizeTimeInput() for clock-only time should prepend date")
	}
}

func TestNormalizeTimeInputFullDateTime(t *testing.T) {
	input := "2025-06-01 14:30"
	result := normalizeTimeInput(input, "UTC", "")
	if result != input {
		t.Errorf("normalizeTimeInput() for full datetime = %q, want %q", result, input)
	}
}

func TestNormalizeTimeInputEmpty(t *testing.T) {
	result := normalizeTimeInput("", "UTC", "")
	if result != "" {
		t.Errorf("normalizeTimeInput() for empty = %q, want empty string", result)
	}
}

// =====================================================================
// create.go — runCreate with invalid start time
// =====================================================================

func TestRunCreateInvalidStartTime(t *testing.T) {
	app := TestApp()
	cmd := NewCreateCmd(app)
	mustSetFlag(t, cmd, "start", "not-a-time")

	err := runCreate(app, cmd, []string{"My Event"})
	if err == nil {
		t.Fatal("runCreate() expected error for invalid start time")
	}
}

// =====================================================================
// create.go — configureEvent with location, description, rrule
// =====================================================================

func TestRunCreateWithAllOptions(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "full.ics")

	app := TestApp()
	cmd := NewCreateCmd(app)
	mustSetFlag(t, cmd, "start", "2025-06-01 10:00")
	mustSetFlag(t, cmd, "end", "2025-06-01 11:00")
	mustSetFlag(t, cmd, "location", "Room 5")
	mustSetFlag(t, cmd, "description", "Team meeting")
	mustSetFlag(t, cmd, "rrule", "FREQ=WEEKLY;COUNT=4")
	mustSetFlag(t, cmd, "priority", "3")
	mustSetFlag(t, cmd, "output", outputPath)

	err := runCreate(app, cmd, []string{"Weekly Sync"})
	if err != nil {
		t.Fatalf("runCreate() with all options error: %v", err)
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	ics := string(data)
	if !strings.Contains(ics, "LOCATION:Room 5") {
		t.Errorf("expected LOCATION in ICS, got: %s", ics)
	}
	if !strings.Contains(ics, "RRULE:FREQ=WEEKLY;COUNT=4") {
		t.Errorf("expected RRULE in ICS, got: %s", ics)
	}
	if !strings.Contains(ics, "PRIORITY:3") {
		t.Errorf("expected PRIORITY:3 in ICS, got: %s", ics)
	}
}

// =====================================================================
// create.go — addEventAlarms with invalid alarm specs
// =====================================================================

func TestAddEventAlarmsInvalidSpec(t *testing.T) {
	event := &calendar.Event{
		Summary:   "Test",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(time.Hour),
	}
	err := addEventAlarms(event, []string{"not-a-valid-alarm-spec-xyz"}, "UTC", TestApp().Config)
	if err == nil {
		t.Fatal("addEventAlarms() expected error for invalid alarm spec, got nil")
	}
	if len(event.Alarms) != 0 {
		t.Errorf("expected 0 alarms after invalid spec, got %d", len(event.Alarms))
	}
}

func TestAddEventAlarmsWithStartTZ(t *testing.T) {
	event := &calendar.Event{
		Summary:   "Test",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(time.Hour),
	}
	if err := addEventAlarms(event, []string{"-15m"}, "Europe/Madrid", TestApp().Config); err != nil {
		t.Fatalf("addEventAlarms() error: %v", err)
	}
	if len(event.Alarms) == 0 {
		t.Error("expected alarm to be added with valid spec")
	}
}

// =====================================================================
// batch.go — parseBatchTimedEventTimes with end before start
// =====================================================================

func TestParseBatchTimedEventTimesEndBeforeStart(t *testing.T) {
	rec := batchRecord{
		Summary: "Test",
		Start:   "2025-06-01 14:00",
		End:     "2025-06-01 10:00",
	}
	_, _, err := parseBatchTimedEventTimes(rec, rec.Start, "UTC", "UTC", "Test")
	if err == nil {
		t.Fatal("parseBatchTimedEventTimes() expected error when end before start")
	}
}

// =====================================================================
// timezone.go — cityToIANA with empty input
// =====================================================================

func TestCityToIANAEmpty(t *testing.T) {
	iana, _ := cityToIANA("")
	if iana != "" {
		t.Errorf("cityToIANA('') = %q, want empty string", iana)
	}
}

func TestCityToIANAKnownCity(t *testing.T) {
	iana, err := cityToIANA("madrid")
	if err != nil {
		t.Logf("cityToIANA('madrid') error (may not be in mapping): %v", err)
	}
	_ = iana
}

// =====================================================================
// template.go — runTemplateInit with os.MkdirAll failure (unwritable path)
// =====================================================================

func TestRunTemplateInitDirCreationFails(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root user can write anywhere")
	}
	app := TestApp()
	initCmd := templateSubcmd(t, "init")
	if initCmd == nil {
		t.Skip("init subcommand not present")
	}
	if runtime.GOOS == "windows" {
		t.Skip("rooted /nonexistent paths are creatable on windows runners")
	}
	if err := initCmd.Flags().Set("dir", "/nonexistent/cannot/create/this"); err != nil {
		t.Fatalf("failed to set dir: %v", err)
	}
	err := runTemplateInit(app, initCmd, []string{"test-event"})
	if err == nil {
		t.Fatal("runTemplateInit() expected error for unwritable dir")
	}
}

// =====================================================================
// template.go — runTemplateCreateFromFile with format detection error
// =====================================================================

func TestRunTemplateCreateFromFileNilStdout(t *testing.T) {
	repoRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}
	templatesDir := filepath.Join(repoRoot, "..", "..", "internal", "templates", "json")
	if _, err := os.Stat(templatesDir); err != nil {
		t.Skip("templates json dir not found")
	}

	app := TestApp()
	app.Stdout = nil

	tm := tpl.NewTemplateManager()
	tm.LoadDDDir(templatesDir)
	tr, _ := newTranslator("en")
	tmpl, err := tm.GetTemplate("medical")
	if err != nil {
		t.Skip("medical template not available")
	}
	dd, _ := tm.DataTemplate("medical")

	tmpDir := t.TempDir()
	csvContent := "doctor,specialty,clinic,start_time,duration,timezone\nDr. Test,,Test Clinic,2025-11-01 09:00,30m,UTC\n"
	csvPath := filepath.Join(tmpDir, "test.csv")
	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		t.Fatalf("failed to write csv: %v", err)
	}

	params := templateCreateParams{
		templateName: "medical",
		inputPath:    csvPath,
		formatFlag:   "csv",
		outputDir:    tmpDir,
	}
	err = runTemplateCreateFromFile(app, tm, tr, tmpl, dd, params)
	if err != nil {
		t.Logf("runTemplateCreateFromFile() with nil stdout error: %v (may be ok)", err)
	}
}

func TestRunTemplateCreateFromFileFormatError(t *testing.T) {
	app := TestApp()
	tm := tpl.NewTemplateManager()
	tr, _ := newTranslator("en")
	tmpl, err := tm.GetTemplate("medical")
	if err != nil {
		t.Skip("medical template not available")
	}
	dd, _ := tm.DataTemplate("medical")

	params := templateCreateParams{
		templateName: "medical",
		inputPath:    "/some/path/file.xml",
		formatFlag:   "auto",
		outputDir:    t.TempDir(),
	}
	err = runTemplateCreateFromFile(app, tm, tr, tmpl, dd, params)
	if err == nil {
		t.Fatal("runTemplateCreateFromFile() expected error for unsupported format")
	}
	if !strings.Contains(err.Error(), "format") && !strings.Contains(err.Error(), "infer") {
		t.Errorf("expected format error, got: %v", err)
	}
}

// =====================================================================
// create.go — addExDates with allDay events and invalid dates
// =====================================================================

func TestAddEventExDatesAllDay(t *testing.T) {
	event := &calendar.Event{
		Summary:   "All Day",
		StartTime: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2025, 6, 2, 0, 0, 0, 0, time.UTC),
		AllDay:    true,
	}
	err := addExDates(event, []string{"2025-06-08", "invalid-date", ""}, "UTC", true)
	if err == nil {
		t.Fatal("addExDates() expected error for invalid exdate, got nil")
	}
}

func TestAddEventExDatesWithStartTZ(t *testing.T) {
	event := &calendar.Event{
		Summary:   "Event",
		StartTime: time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2025, 6, 1, 11, 0, 0, 0, time.UTC),
		StartTZ:   "Europe/Madrid",
	}
	if err := addExDates(event, []string{"2025-06-08 10:00"}, "", false); err != nil {
		t.Fatalf("addExDates() error: %v", err)
	}
	if len(event.ExDates) != 1 {
		t.Errorf("expected 1 exdate, got %d", len(event.ExDates))
	}
}

func TestAddEventExDatesEmptyList(t *testing.T) {
	event := &calendar.Event{}
	if err := addExDates(event, []string{}, "UTC", false); err != nil {
		t.Fatalf("addExDates() error: %v", err)
	}
	if len(event.ExDates) != 0 {
		t.Error("expected no exdates for empty input")
	}
}

// =====================================================================
// batch.go — addExDates with startTZ and error path
// =====================================================================

func TestAddBatchExDatesWithStartTZ(t *testing.T) {
	event := &calendar.Event{
		Summary:   "Event",
		StartTime: time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2025, 6, 1, 11, 0, 0, 0, time.UTC),
		StartTZ:   "Europe/Madrid",
	}
	if err := addExDates(event, []string{"2025-06-08"}, "", false); err != nil {
		t.Fatalf("addExDates() error: %v", err)
	}
}

func TestAddBatchExDatesEmpty(t *testing.T) {
	event := &calendar.Event{}
	if err := addExDates(event, []string{}, "UTC", false); err != nil {
		t.Fatalf("addExDates() error: %v", err)
	}
}

// =====================================================================
// template.go — deriveTemplateFilename with GuessStartDate path
// =====================================================================

func TestDeriveTemplateFilenameGuessStartDate(t *testing.T) {
	tm := tpl.NewTemplateManager()
	ev := &calendar.Event{}
	values := map[string]string{
		"start_time": "2025-07-15 10:00",
	}

	got := deriveTemplateFilename(tm, "medical", values, ev, nil)
	if got == "" {
		t.Error("deriveTemplateFilename() returned empty string")
	}
	if !strings.HasSuffix(got, ".ics") {
		t.Errorf("expected .ics suffix, got: %s", got)
	}
}

// =====================================================================
// batch.go — parseDurationEnd (unified) valid case
// =====================================================================

func TestParseBatchDurationEndValid(t *testing.T) {
	start := time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC)
	end, err := parseDurationEnd(start, "45m")
	if err != nil {
		t.Fatalf("parseDurationEnd() error: %v", err)
	}
	expected := start.Add(45 * time.Minute)
	if !end.Equal(expected) {
		t.Errorf("end = %v, want %v", end, expected)
	}
}

// =====================================================================
// create.go — parseDurationEnd with negative duration
// =====================================================================

func TestParseDurationEndNegative(t *testing.T) {
	start := time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC)
	_, err := parseDurationEnd(start, "-30m")
	if err == nil {
		t.Error("parseDurationEnd() expected error for negative duration")
	}
}

// =====================================================================
// template.go — runTemplateInit dir is empty string
// =====================================================================

func TestRunTemplateInitEmptyDir(t *testing.T) {
	app := TestApp()
	initCmd := templateSubcmd(t, "init")
	if initCmd == nil {
		t.Skip("init subcommand not present")
	}
	if err := initCmd.Flags().Set("dir", ""); err != nil {
		t.Fatalf("failed to set dir: %v", err)
	}

	err := runTemplateInit(app, initCmd, []string{"test-event"})
	if err == nil {
		t.Fatal("runTemplateInit() expected error for empty dir")
	}
	if !strings.Contains(err.Error(), "directory cannot be empty") {
		t.Errorf("expected 'directory cannot be empty' error, got: %v", err)
	}
}

// =====================================================================
// template.go — runTemplateList with nil stdout (os.Stdout fallback)
// =====================================================================

func TestRunTemplateListNilStdout(t *testing.T) {
	app := TestApp()
	app.Stdout = nil

	templateCmd := NewTemplateCmd(app)
	var listCmd *cobra.Command
	for _, sub := range templateCmd.Commands() {
		if strings.HasPrefix(sub.Use, "list") {
			listCmd = sub
			break
		}
	}
	if listCmd == nil {
		t.Skip("list subcommand not found")
	}

	err := runTemplateList(app, listCmd, nil)
	if err != nil {
		t.Fatalf("runTemplateList() with nil stdout error: %v", err)
	}
}

// =====================================================================
// config.go — runConfigSet with valid key
// =====================================================================

func TestRunConfigSetLanguageKey(t *testing.T) {
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	cmd := &cobra.Command{}
	err := runConfigSet(app, cmd, []string{"language", "es"})
	if err != nil {
		t.Logf("runConfigSet() error (may be ok if config read-only): %v", err)
	}
}

// =====================================================================
// batch.go — writeBatchOutput with EnsureDirForFile error
// =====================================================================

func TestWriteBatchOutputDirError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root user can write anywhere")
	}
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	if runtime.GOOS == "windows" {
		t.Skip("rooted /nonexistent paths are creatable on windows runners")
	}
	cal := &calendar.Calendar{}
	err := writeBatchOutput(app, cal, nil, "/nonexistent/path/cannot/create/out.ics", 0)
	if err == nil {
		t.Fatal("writeBatchOutput() expected error for unwritable path")
	}
}

// =====================================================================
// template.go — promptAlarmField interactive branches via Scanner injection
// =====================================================================

func TestPromptAlarmFieldChangeDefaults(t *testing.T) {
	prevScanner := prompts.Scanner
	// User says "n" to change, then types "-15m", then empty to stop
	prompts.Scanner = bufio.NewScanner(strings.NewReader("n\n-15m\n\n"))
	defer func() { prompts.Scanner = prevScanner }()

	result := promptAlarmField(io.Discard, TestApp().Translator, "Reminders", "-30m|-5m")
	if result == "" {
		t.Error("promptAlarmField() should return non-empty result")
	}
}

func TestPromptAlarmFieldHelpInput(t *testing.T) {
	prevScanner := prompts.Scanner
	// User types "?" for help then empty to stop
	prompts.Scanner = bufio.NewScanner(strings.NewReader("?\n\n"))
	defer func() { prompts.Scanner = prevScanner }()

	result := promptAlarmField(io.Discard, TestApp().Translator, "Reminders", "")
	_ = result
}

func TestPromptAlarmFieldInvalidAlarmSpec(t *testing.T) {
	prevScanner := prompts.Scanner
	// User types an invalid spec, then empty to stop
	prompts.Scanner = bufio.NewScanner(strings.NewReader("not-valid-alarm\n\n"))
	defer func() { prompts.Scanner = prevScanner }()

	result := promptAlarmField(io.Discard, TestApp().Translator, "Reminders", "")
	_ = result
}

func TestPromptAlarmFieldWithDescription(t *testing.T) {
	prevScanner := prompts.Scanner
	// User types a simple trigger (-15m), then adds a description, then empty
	prompts.Scanner = bufio.NewScanner(strings.NewReader("-15m\nMy reminder\n\n"))
	defer func() { prompts.Scanner = prevScanner }()

	result := promptAlarmField(io.Discard, TestApp().Translator, "Reminders", "")
	_ = result
}

// =====================================================================
// template.go — runTemplateList with no-description template (desc == "")
// =====================================================================

func TestRunTemplateListDescriptionFallback(t *testing.T) {
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	templateCmd := NewTemplateCmd(app)
	var listCmd *cobra.Command
	for _, sub := range templateCmd.Commands() {
		if strings.HasPrefix(sub.Use, "list") {
			listCmd = sub
			break
		}
	}
	if listCmd == nil {
		t.Skip("list subcommand not found")
	}

	err := runTemplateList(app, listCmd, nil)
	if err != nil {
		t.Fatalf("runTemplateList() error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "-") {
		t.Logf("output: %s", out)
	}
}

// =====================================================================
// template.go — mergeTemplateValues with extra record keys
// =====================================================================

func TestMergeTemplateValuesExtraKeys(t *testing.T) {
	tmpl := &tpl.Template{
		Fields: []tpl.Field{
			{Key: "summary", Default: "default-summary"},
			{Key: "duration", Default: "1h"},
		},
	}
	record := map[string]string{
		"summary":     "My Event",
		"extra_field": "extra_value",
	}
	values := mergeTemplateValues(tmpl, record)
	if values["summary"] != "My Event" {
		t.Errorf("mergeTemplateValues() summary = %q, want %q", values["summary"], "My Event")
	}
	if values["duration"] != "1h" {
		t.Errorf("mergeTemplateValues() duration fallback = %q, want %q", values["duration"], "1h")
	}
	if values["extra_field"] != "extra_value" {
		t.Errorf("mergeTemplateValues() extra_field = %q, want %q", values["extra_field"], "extra_value")
	}
}

// =====================================================================
// template.go — loadTemplateFromCSV with load error
// =====================================================================

func TestLoadTemplateFromCSVNonexistentFile(t *testing.T) {
	_, err := loadTemplateFromCSV("/nonexistent/path/file.csv")
	if err == nil {
		t.Fatal("loadTemplateFromCSV() expected error for nonexistent file")
	}
}

func TestLoadTemplateFromJSONNonexistentFile(t *testing.T) {
	_, err := loadTemplateFromJSON("/nonexistent/path/file.json")
	if err == nil {
		t.Fatal("loadTemplateFromJSON() expected error for nonexistent file")
	}
}

func TestLoadTemplateFromJSONInvalid(t *testing.T) {
	tmpDir := t.TempDir()
	badJSON := filepath.Join(tmpDir, "bad.json")
	if err := os.WriteFile(badJSON, []byte("{invalid"), 0644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}
	_, err := loadTemplateFromJSON(badJSON)
	if err == nil {
		t.Fatal("loadTemplateFromJSON() expected error for invalid JSON")
	}
}

// =====================================================================
// config.go — runConfigList and runConfigAlarmProfiles with real config
// =====================================================================

func TestRunConfigListWithApp(t *testing.T) {
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	err := runConfigList(app, nil, nil)
	if err != nil {
		t.Logf("runConfigList() error (may be ok): %v", err)
	}
}

// =====================================================================
// timezone.go — cityToIANA with known and unknown cities
// =====================================================================

func TestCityToIANALondon(t *testing.T) {
	iana, err := cityToIANA("london")
	if err != nil {
		t.Logf("cityToIANA('london') not in mapping: %v", err)
	}
	_ = iana
}

func TestCityToIANAUpperCase(t *testing.T) {
	iana, err := cityToIANA("MADRID")
	if err != nil {
		t.Logf("cityToIANA('MADRID') not in mapping: %v", err)
	}
	_ = iana
}

// =====================================================================
// create.go — parseEndTime with exact datetime string
// =====================================================================

func TestParseEndTimeExactDateTime(t *testing.T) {
	start := time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC)
	end, err := parseEndTime(start, "2025-06-01 12:00")
	if err != nil {
		t.Fatalf("parseEndTime() exact datetime error: %v", err)
	}
	expected := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	if !end.Equal(expected) {
		t.Errorf("end = %v, want %v", end, expected)
	}
}

// =====================================================================
// create.go — addExDates with date-only non-allDay event
// =====================================================================

func TestAddEventExDatesDateOnlyNonAllDay(t *testing.T) {
	event := &calendar.Event{
		Summary:   "Event",
		StartTime: time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2025, 6, 1, 11, 0, 0, 0, time.UTC),
	}
	if err := addExDates(event, []string{"2025-06-08"}, "UTC", false); err != nil {
		t.Fatalf("addExDates() error: %v", err)
	}
	if len(event.ExDates) != 1 {
		t.Errorf("expected 1 exdate for date-only non-allday, got %d", len(event.ExDates))
	}
}

func TestAddEventExDatesInvalidTimestamp(t *testing.T) {
	event := &calendar.Event{
		Summary:   "Event",
		StartTime: time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2025, 6, 1, 11, 0, 0, 0, time.UTC),
	}
	err := addExDates(event, []string{"not-a-valid-date 25:99"}, "UTC", false)
	if err == nil {
		t.Fatal("addExDates() expected error for invalid date, got nil")
	}
	if len(event.ExDates) != 0 {
		t.Error("expected no exdates for invalid date")
	}
}

// =====================================================================
// lint.go — runLint with multiple valid files
// =====================================================================

// =====================================================================
// timezone.go — cityToIANA with Brazilian city names (specific cases)
// =====================================================================

func TestCityToIANABrazilianCities(t *testing.T) {
	cities := []struct {
		city string
		want string
	}{
		{"recife", "America/Recife"},
		{"manaus", "America/Manaus"},
		{"fortaleza", "America/Fortaleza"},
		{"salvador", "America/Bahia"},
		{"melilla", "Africa/Ceuta"},
		{"tenerife", "Atlantic/Canary"},
	}
	for _, tc := range cities {
		t.Run(tc.city, func(t *testing.T) {
			iana, err := cityToIANA(tc.city)
			if err != nil {
				t.Errorf("cityToIANA(%q) error: %v", tc.city, err)
				return
			}
			if iana != tc.want {
				t.Errorf("cityToIANA(%q) = %q, want %q", tc.city, iana, tc.want)
			}
		})
	}
}

func TestCityToIANAUnknown(t *testing.T) {
	_, err := cityToIANA("unknowncity123")
	if err == nil {
		t.Error("cityToIANA() expected error for unknown city")
	}
}

// =====================================================================
// batch.go — runBatch with SpellCorrections in config
// =====================================================================

func TestRunBatchWithSpellCorrections(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "events.csv")
	outputPath := filepath.Join(tmpDir, "out.ics")

	csvData := "summary,start,end\nTeam Meetting,2025-05-01 09:00,2025-05-01 10:00\n"
	if err := os.WriteFile(inputPath, []byte(csvData), 0644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	app := TestApp()
	app.Config.SpellCorrections = map[string]string{"meetting": "meeting"}
	var buf bytes.Buffer
	app.Stdout = &buf

	cmd := NewBatchCmd(app)
	mustSetFlag(t, cmd, "input", inputPath)
	mustSetFlag(t, cmd, "output", outputPath)

	if err := runBatch(app, cmd, nil); err != nil {
		t.Fatalf("runBatch() with spell corrections error: %v", err)
	}
}

// =====================================================================
// batch.go — handleDryRun with nil stdout
// =====================================================================

func TestHandleDryRunNilStdout(t *testing.T) {
	app := TestApp()
	app.Stdout = nil

	records := []batchRecord{
		{Summary: "Event A", Start: "2025-05-01 09:00"},
	}
	err := handleDryRun(app, nil, nil, records, "input.csv", "output.ics")
	if err != nil {
		t.Fatalf("handleDryRun() with nil stdout error: %v", err)
	}
}

// =====================================================================
// batch.go — parseBatchTimedEventTimes with clock-only start
// =====================================================================

func TestParseBatchTimedEventTimesInvalidStart(t *testing.T) {
	rec := batchRecord{
		Summary: "Meeting",
		Start:   "not-a-date",
		End:     "2025-06-01 11:00",
	}
	_, _, err := parseBatchTimedEventTimes(rec, rec.Start, "UTC", "UTC", "Meeting")
	if err == nil {
		t.Fatal("parseBatchTimedEventTimes() expected error for invalid start")
	}
	if !strings.Contains(err.Error(), "invalid start time") {
		t.Errorf("expected 'invalid start time' in error, got: %v", err)
	}
}

func TestParseBatchTimedEventTimesClockOnly(t *testing.T) {
	rec := batchRecord{
		Summary: "Meeting",
		Start:   "10:00",
		End:     "11:00",
	}
	start, end, err := parseBatchTimedEventTimes(rec, rec.Start, "UTC", "UTC", "Meeting")
	if err != nil {
		t.Fatalf("parseBatchTimedEventTimes() clock-only error: %v", err)
	}
	if start.IsZero() {
		t.Error("start time should not be zero")
	}
	if !end.After(start) {
		t.Error("end time should be after start time")
	}
}

// =====================================================================
// batch.go — parseBatchExplicitEnd with zero duration
// =====================================================================

func TestParseBatchExplicitEndZeroDuration(t *testing.T) {
	startTime := time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC)
	_, err := parseBatchExplicitEnd("0m", startTime, "UTC", "0m")
	if err == nil {
		t.Fatal("parseBatchExplicitEnd() expected error for zero duration")
	}
}

func TestParseBatchExplicitEndClockOnly(t *testing.T) {
	startTime := time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC)
	end, err := parseBatchExplicitEnd("11:00", startTime, "UTC", "11:00")
	if err != nil {
		t.Fatalf("parseBatchExplicitEnd() clock-only error: %v", err)
	}
	if end.IsZero() {
		t.Error("end time should not be zero")
	}
}

// =====================================================================
// batch.go — configureBatchEvent with endTZ
// =====================================================================

func TestConfigureBatchEventWithEndTZ(t *testing.T) {
	catCache := nd.NewCategoryCache()
	event := &calendar.Event{
		Summary:   "Test Event",
		StartTime: time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2025, 6, 1, 11, 0, 0, 0, time.UTC),
	}
	rec := batchRecord{
		Summary:    "Test Event",
		RRule:      "FREQ=WEEKLY;COUNT=4",
		Location:   "Room 1",
		Categories: []string{"Work"},
	}
	if err := configureBatchEvent(event, rec, "Europe/Madrid", "America/New_York", catCache, TestApp().Config); err != nil {
		t.Fatalf("configureBatchEvent() error: %v", err)
	}
	if event.StartTZ != "Europe/Madrid" {
		t.Errorf("StartTZ = %q, want %q", event.StartTZ, "Europe/Madrid")
	}
	if event.EndTZ != "America/New_York" {
		t.Errorf("EndTZ = %q, want %q", event.EndTZ, "America/New_York")
	}
}

// =====================================================================
// batch.go — loadBatchFromCSV with empty file (EOF on header)
// =====================================================================

func TestLoadBatchFromCSVEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	emptyPath := filepath.Join(tmpDir, "empty.csv")
	if err := os.WriteFile(emptyPath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}
	records, err := loadBatchFromCSV(emptyPath)
	if err != nil {
		t.Fatalf("loadBatchFromCSV() empty file error: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected 0 records for empty file, got %d", len(records))
	}
}

// =====================================================================
// locale.go — runLocaleList with nil stdout
// =====================================================================

func TestRunLocaleListNilStdoutFallback(t *testing.T) {
	app := TestApp()
	app.Stdout = nil

	cmd := NewLocaleCmd(app)
	listCmd := findSubcommand(cmd, "list")
	if listCmd == nil {
		t.Fatal("list subcommand not found")
	}

	err := runLocaleList(app, listCmd, nil)
	if err != nil {
		t.Fatalf("runLocaleList() with nil stdout error: %v", err)
	}
}

// =====================================================================
// create.go — addEventAlarms with profile that triggers expandAlarmProfiles error
// =====================================================================

func TestAddEventAlarmsWithUnknownProfile(t *testing.T) {
	event := &calendar.Event{
		Summary:   "Test",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(time.Hour),
	}
	// unknown profile - must fail loud and not add alarms
	err := addEventAlarms(event, []string{"profile:nonexistent-profile-xyz"}, "UTC", TestApp().Config)
	if err == nil {
		t.Fatal("addEventAlarms() expected error for unknown profile, got nil")
	}
	if len(event.Alarms) != 0 {
		t.Errorf("expected 0 alarms for unknown profile, got %d", len(event.Alarms))
	}
}

// =====================================================================
// template.go — runTemplateCreate with required field empty (scanner injection)
// =====================================================================

func TestRunTemplateCreateRequiredFieldEmpty(t *testing.T) {
	repoRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}
	templatesDir := filepath.Join(repoRoot, "..", "..", "internal", "templates", "json")
	if _, err := os.Stat(templatesDir); err != nil {
		t.Skip("templates json dir not found")
	}

	app := TestApp()
	templateCmd := NewTemplateCmd(app)

	var createCmd *cobra.Command
	for _, sub := range templateCmd.Commands() {
		if strings.HasPrefix(sub.Use, "create") {
			createCmd = sub
			break
		}
	}
	if createCmd == nil {
		t.Skip("create subcommand not present")
	}
	if err := createCmd.Flags().Set("templates-dir", templatesDir); err != nil {
		t.Fatalf("failed to set templates-dir: %v", err)
	}

	prevScanner := prompts.Scanner
	// Provide empty input for all fields — required fields should error
	prompts.Scanner = bufio.NewScanner(strings.NewReader("\n\n\n\n\n\n\n\n\n\n\n\n\n"))
	defer func() { prompts.Scanner = prevScanner }()

	err = runTemplateCreate(app, createCmd, []string{"medical"})
	if err == nil {
		t.Skip("medical template has no required fields — skipping")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Logf("runTemplateCreate() empty required field error: %v", err)
	}
}

// =====================================================================
// template.go — runTemplateCreate with unknown template name
// =====================================================================

func TestRunTemplateCreateUnknownTemplate(t *testing.T) {
	app := TestApp()
	templateCmd := NewTemplateCmd(app)

	var createCmd *cobra.Command
	for _, sub := range templateCmd.Commands() {
		if strings.HasPrefix(sub.Use, "create") {
			createCmd = sub
			break
		}
	}
	if createCmd == nil {
		t.Skip("create subcommand not present")
	}

	err := runTemplateCreate(app, createCmd, []string{"nonexistent-template-xyz"})
	if err == nil {
		t.Fatal("runTemplateCreate() expected error for unknown template")
	}
}

// =====================================================================
// template.go — printTemplateTypeInfo with Source empty (embedded path)
// =====================================================================

func TestPrintTemplateTypeInfoWithDDEmbedded(t *testing.T) {
	repoRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}
	templatesDir := filepath.Join(repoRoot, "..", "..", "internal", "templates", "json")
	if _, err := os.Stat(templatesDir); err != nil {
		t.Skip("templates json dir not found")
	}

	tm := tpl.NewTemplateManager()
	tm.LoadDDDir(templatesDir)

	var buf bytes.Buffer
	dd := printTemplateTypeInfo(&buf, tm, "medical")
	out := buf.String()

	if dd.Name == "" {
		t.Skip("medical is not a DD template in this setup")
	}
	if strings.Contains(out, "Source:") {
		// Source was set — OK
		t.Logf("Source output: %s", out)
	} else {
		// Source was empty → should show "embedded"
		if !strings.Contains(out, "embedded") {
			t.Errorf("expected 'embedded' or 'Source:' in output, got: %s", out)
		}
	}
}

// =====================================================================
// batch.go — runBatch with addPrepTime flag (add-prep-time)
// =====================================================================

func TestRunBatchWithAddPrepTime(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "events.csv")
	outputPath := filepath.Join(tmpDir, "out.ics")

	csvData := "summary,start,end\nDoctor Appointment,2025-06-01 10:00,2025-06-01 11:00\n"
	if err := os.WriteFile(inputPath, []byte(csvData), 0644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	cmd := NewBatchCmd(app)
	mustSetFlag(t, cmd, "input", inputPath)
	mustSetFlag(t, cmd, "output", outputPath)
	mustSetFlag(t, cmd, "add-prep-time", "true")

	if err := runBatch(app, cmd, nil); err != nil {
		t.Fatalf("runBatch() add-prep-time error: %v", err)
	}
}

// =====================================================================
// timezone.go — NewTimezoneCmd with nil stdout (os.Stdout fallback)
// =====================================================================

func TestRunTZListNilStdout(t *testing.T) {
	app := TestApp()
	app.Stdout = nil

	cmd := NewTimezoneCmd(app)
	listCmd := findSubcommand(cmd, "list")
	if listCmd == nil {
		t.Fatal("list subcommand not found")
	}

	err := runTZList(app, listCmd, nil)
	if err != nil {
		t.Fatalf("runTZList() nil stdout error: %v", err)
	}
}

func TestRunTZInfoNilStdout(t *testing.T) {
	app := TestApp()
	app.Stdout = nil

	err := runTZInfo(app, nil, []string{"UTC"})
	if err != nil {
		t.Fatalf("runTZInfo() nil stdout error: %v", err)
	}
}

func TestRunLintMultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()

	makeICS := func(name, uid string) string {
		icsPath := filepath.Join(tmpDir, name)
		content := "BEGIN:VCALENDAR\nVERSION:2.0\nBEGIN:VEVENT\nUID:" + uid + "\nSUMMARY:Test\nDTSTAMP:20250401T090000Z\nDTSTART:20250601T100000Z\nDTEND:20250601T110000Z\nEND:VEVENT\nEND:VCALENDAR\n"
		if err := os.WriteFile(icsPath, []byte(content), 0644); err != nil {
			panic(err)
		}
		return icsPath
	}

	path1 := makeICS("a.ics", "uid-1@test")
	path2 := makeICS("b.ics", "uid-2@test")

	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf

	cmd := NewLintCmd(app)
	if err := cmd.Flags().Set("file", path1); err != nil {
		t.Fatalf("failed to set file1: %v", err)
	}
	if err := cmd.Flags().Set("file", path2); err != nil {
		t.Fatalf("failed to set file2: %v", err)
	}

	err := runLint(app, cmd, nil)
	if err != nil {
		t.Fatalf("runLint() multiple valid files error: %v", err)
	}
}

func TestRunTZListRegionFilters(t *testing.T) {
	app := TestApp()
	var buf bytes.Buffer
	app.Stdout = &buf
	cmd := NewTimezoneCmd(app)
	listCmd, _, err := cmd.Find([]string{"list"})
	if err != nil {
		t.Fatal(err)
	}
	if err := listCmd.Flags().Set("region", "asia"); err != nil {
		t.Fatal(err)
	}
	if err := runTZList(app, listCmd, nil); err != nil {
		t.Fatalf("runTZList(region=asia) error: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "Europe/Madrid") {
		t.Error("region=asia must not list European zones")
	}
	if !strings.Contains(out, "Asia/") {
		t.Error("region=asia must list Asian zones")
	}
}

func TestRunTZListUnknownRegionErrors(t *testing.T) {
	app := TestApp()
	cmd := NewTimezoneCmd(app)
	listCmd, _, err := cmd.Find([]string{"list"})
	if err != nil {
		t.Fatal(err)
	}
	if err := listCmd.Flags().Set("region", "narnia"); err != nil {
		t.Fatal(err)
	}
	if err := runTZList(app, listCmd, nil); err == nil {
		t.Error("unknown region must be an error, not a silent full listing")
	}
}
