package cli

import (
	"os"
	"path/filepath"
	"strings"
	"tempus/internal/testutil"
	"testing"
)

func TestInterpretRRule(t *testing.T) {
	tests := []struct {
		name      string
		rrule     string
		wantEmpty bool
	}{
		{"daily", "FREQ=DAILY", false},
		{"weekly", "FREQ=WEEKLY;BYDAY=MO,WE,FR", false},
		{"monthly", "FREQ=MONTHLY;BYMONTHDAY=15", false},
		{"yearly", "FREQ=YEARLY", false},
		{"with count", "FREQ=DAILY;COUNT=10", false},
		{"with until", "FREQ=WEEKLY;UNTIL=20251231", false},
		{"complex", "FREQ=MONTHLY;BYDAY=1MO;COUNT=12", false},
		{"empty", "", false},
		{"invalid", "INVALID", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := interpretRRule(tt.rrule)
			isEmpty := got == ""
			if isEmpty != tt.wantEmpty {
				t.Errorf("interpretRRule(%q) isEmpty = %v, want %v, got %q",
					tt.rrule, isEmpty, tt.wantEmpty, got)
			}
		})
	}
}

func TestNewRRuleHelperCmd(t *testing.T) {
	app := TestApp()
	cmd := NewRRuleHelperCmd(app)
	if cmd == nil {
		t.Fatal("NewRRuleHelperCmd() returned nil")
	}
	if cmd.Use != "rrule" {
		t.Errorf(testutil.ErrMsgUseMismatch, cmd.Use, "rrule")
	}
	if cmd.RunE == nil {
		t.Error("rrule command should have RunE function")
	}
}

func TestRunCreateWritesRecurrenceData(t *testing.T) {
	app := TestApp()
	cmd := NewCreateCmd(app)

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, testutil.FilenameEventICS)

	set := func(name, value string) {
		if err := cmd.Flags().Set(name, value); err != nil {
			t.Fatalf("failed to set flag %s: %v", name, err)
		}
	}

	set("start", "2025-03-01 10:00")
	set("end", "2025-03-01 11:00")
	set("start-tz", testutil.TZEuropeMadrid)
	set("output", outputPath)
	set("rrule", testutil.RRuleDaily5Count)
	set("exdate", "2025-03-03 10:00")

	if err := cmd.RunE(cmd, []string{"Recurrent Event"}); err != nil {
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

	if !strings.Contains(ics, "EXDATE") {
		t.Fatalf("expected EXDATE to be present, got:\n%s", ics)
	}
}
