package main

import (
	"tempus/internal/testutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateSupportsAlarms(t *testing.T) {
	cmd := newCreateCmd()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "alarm.ics")

	set := func(name, value string) {
		if err := cmd.Flags().Set(name, value); err != nil {
			t.Fatalf("failed to set flag %s: %v", name, err)
		}
	}

	set("start", "2025-03-01 10:00")
	set("end", "2025-03-01 11:00")
	set("start-tz", testutil.TZEuropeMadrid)
	set("output", outputPath)
	set("alarm", "15m")
	set("alarm", "trigger=+10m,description=Wrap up")
	set("alarm", "trigger=2025-03-01 09:15,description=Airport check-in")

	if err := runCreate(cmd, []string{"Flight Reminder"}); err != nil {
		t.Fatalf("runCreate returned error: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read generated ICS: %v", err)
	}
	ics := string(data)

	if strings.Count(ics, "BEGIN:VALARM") != 3 {
		t.Fatalf("expected 3 VALARM blocks, got:\n%s", ics)
	}
	if !strings.Contains(ics, "TRIGGER:-PT15M") {
		t.Fatalf("expected 15 minute reminder before start:\n%s", ics)
	}
	if !strings.Contains(ics, "TRIGGER:PT10M") {
		t.Fatalf("expected post-start reminder:\n%s", ics)
	}
	if !strings.Contains(ics, "TRIGGER;VALUE=DATE-TIME:20250301T081500Z") {
		t.Fatalf("expected absolute trigger converted to UTC:\n%s", ics)
	}
	if !strings.Contains(ics, "DESCRIPTION:Wrap up") {
		t.Fatalf("expected custom description for second alarm:\n%s", ics)
	}
	if !strings.Contains(ics, "DESCRIPTION:Airport check-in") {
		t.Fatalf("expected custom description for absolute alarm:\n%s", ics)
	}
}
