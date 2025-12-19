package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"tempus/internal/testutil"
	"testing"

	"tempus/internal/prompts"

	"github.com/spf13/cobra"
)

func findTemplateCreateCmd() *cobra.Command {
	templateCmd := newTemplateCmd()
	for _, cmd := range templateCmd.Commands() {
		if strings.HasPrefix(cmd.Use, "create") {
			return cmd
		}
	}
	return nil
}

func TestTemplateCreateMedicalSupportsAdvancedFeatures(t *testing.T) {
	createCmd := findTemplateCreateCmd()
	if createCmd == nil {
		t.Fatalf("create template command not found")
	}

	repoRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	templatesDir := filepath.Join(repoRoot, "internal", "templates", "json")
	if err := createCmd.Flags().Set(testutil.TemplatesDir, templatesDir); err != nil {
		t.Fatalf("failed to set templates-dir flag: %v", err)
	}

	outputDir := t.TempDir()
	if err := createCmd.Flags().Set("output-dir", outputDir); err != nil {
		t.Fatalf("failed to set output-dir flag: %v", err)
	}

	inputs := strings.Join([]string{
		"Dr. Jane Doe",                      // doctor
		"",                                  // specialty (blank)
		"City Clinic",                       // clinic
		"2025-10-15 09:30",                  // start_time
		"",                                  // end_time (blank -> use duration)
		"45m",                               // duration
		testutil.TZEuropeMadrid,             // timezone
		"Bring previous records",            // notes
		"FREQ=MONTHLY;COUNT=3",              // rrule
		"2025-10-20 09:30,2025-10-25 09:30", // exdates
		"-15m",                              // alarm #1
		"",                                  // finish alarms
		"",                                  // accept default filename
	}, "\n") + "\n"

	prevScanner := prompts.Scanner
	prompts.Scanner = bufio.NewScanner(strings.NewReader(inputs))
	defer func() {
		prompts.Scanner = prevScanner
	}()

	if err := runTemplateCreate(createCmd, []string{"medical"}); err != nil {
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
}

func TestTemplateCreateMedicalFromCSVGeneratesMultipleICS(t *testing.T) {
	createCmd := findTemplateCreateCmd()
	if createCmd == nil {
		t.Fatalf("create template command not found")
	}

	repoRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	templatesDir := filepath.Join(repoRoot, "internal", "templates", "json")
	if err := createCmd.Flags().Set(testutil.TemplatesDir, templatesDir); err != nil {
		t.Fatalf("failed to set templates-dir flag: %v", err)
	}

	outputDir := t.TempDir()
	if err := createCmd.Flags().Set("output-dir", outputDir); err != nil {
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

	prevScanner := scanner
	scanner = bufio.NewScanner(strings.NewReader(""))
	defer func() {
		scanner = prevScanner
	}()

	if err := createCmd.Flags().Set("input", csvPath); err != nil {
		t.Fatalf("failed to set input flag: %v", err)
	}

	if err := runTemplateCreate(createCmd, []string{"medical"}); err != nil {
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
}
