package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestBatchCSVGeneratesCalendarWithMultipleEvents(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "events.csv")
	outputPath := filepath.Join(tmpDir, "batch.ics")

	csvData := strings.Join([]string{
		"summary,start,end,start_tz,end_tz,location,description,all_day,duration,rrule,exdate,alarms",
		`"Daily Standup","2025-05-01 09:00","2025-05-01 09:15","Europe/Madrid","","Zoom link","Team sync",false,,"FREQ=DAILY;COUNT=5","2025-05-03 09:00","15m"`,
		`"All Hands","2025-05-04 10:00","2025-05-04 11:30","Europe/Madrid","","Auditorium","Company-wide update",false,,"","","trigger=+5m,description=Wrap up"`,
	}, "\n")

	if err := os.WriteFile(inputPath, []byte(csvData), 0644); err != nil {
		t.Fatalf("failed to write csv: %v", err)
	}

	cmd := newBatchCmd()
	mustSetFlag(t, cmd, "input", inputPath)
	mustSetFlag(t, cmd, "output", outputPath)
	mustSetFlag(t, cmd, "format", "csv")
	mustSetFlag(t, cmd, "name", "Team Events")

	if err := runBatch(cmd, nil); err != nil {
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

	cmd := newBatchCmd()
	mustSetFlag(t, cmd, "input", inputPath)
	mustSetFlag(t, cmd, "output", outputPath)
	mustSetFlag(t, cmd, "format", "json")
	mustSetFlag(t, cmd, "name", "Autumn Plan")

	if err := runBatch(cmd, nil); err != nil {
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

func mustSetFlag(t *testing.T, cmd *cobra.Command, name, value string) {
	t.Helper()
	if err := cmd.Flags().Set(name, value); err != nil {
		t.Fatalf("failed to set flag %s: %v", name, err)
	}
}
