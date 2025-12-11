package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateSupportsCategoriesAttendeesAndPriority(t *testing.T) {
	cmd := newCreateCmd()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "category.ics")

	set := func(name, value string) {
		if err := cmd.Flags().Set(name, value); err != nil {
			t.Fatalf("failed to set flag %s: %v", name, err)
		}
	}

	set("start", "2025-04-01 14:00")
	set("end", "2025-04-01 15:00")
	set("start-tz", "Europe/Madrid")
	set("output", outputPath)
	set("category", "Focus")
	set("category", "DeepWork")
	set("attendee", "alice@example.com")
	set("attendee", "bob@example.com")
	set("priority", "3")

	if err := runCreate(cmd, []string{"Focus Session"}); err != nil {
		t.Fatalf("runCreate returned error: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read generated ICS: %v", err)
	}
	ics := string(data)

	if !strings.Contains(ics, "CATEGORIES:Focus,DeepWork") {
		t.Fatalf("expected categories in ICS, got:\n%s", ics)
	}
	if strings.Count(ics, "ATTENDEE") != 2 {
		t.Fatalf("expected two attendees in ICS, got:\n%s", ics)
	}
	if !strings.Contains(ics, "ATTENDEE:mailto:alice@example.com") || !strings.Contains(ics, "ATTENDEE:mailto:bob@example.com") {
		t.Fatalf("expected attendee mailto entries, got:\n%s", ics)
	}
	if !strings.Contains(ics, "PRIORITY:3") {
		t.Fatalf("expected priority 3 in ICS, got:\n%s", ics)
	}
}

func TestCreateRejectsInvalidPriority(t *testing.T) {
	cmd := newCreateCmd()
	set := func(name, value string) {
		if err := cmd.Flags().Set(name, value); err != nil {
			t.Fatalf("failed to set flag %s: %v", name, err)
		}
	}

	set("start", "2025-04-01 09:00")
	set("end", "2025-04-01 10:00")
	set("priority", "10")

	err := runCreate(cmd, []string{"Invalid priority"})
	if err == nil || !strings.Contains(err.Error(), "priority must be between 0 and 9") {
		t.Fatalf("expected priority validation error, got %v", err)
	}
}
