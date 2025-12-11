package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCreateWritesRecurrenceData(t *testing.T) {
	cmd := newCreateCmd()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "event.ics")

	set := func(name, value string) {
		if err := cmd.Flags().Set(name, value); err != nil {
			t.Fatalf("failed to set flag %s: %v", name, err)
		}
	}

	set("start", "2025-03-01 10:00")
	set("end", "2025-03-01 11:00")
	set("start-tz", "Europe/Madrid")
	set("output", outputPath)
	set("rrule", "FREQ=DAILY;COUNT=5")
	set("exdate", "2025-03-03 10:00")

	if err := runCreate(cmd, []string{"Recurrent Event"}); err != nil {
		t.Fatalf("runCreate returned error: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read generated ICS: %v", err)
	}
	ics := string(data)

	if !strings.Contains(ics, "RRULE:FREQ=DAILY;COUNT=5") {
		t.Fatalf("expected RRULE to be present, got:\n%s", ics)
	}

	if !strings.Contains(ics, "EXDATE;TZID=Europe/Madrid:20250303T100000") {
		t.Fatalf("expected EXDATE with timezone to be present, got:\n%s", ics)
	}
}

