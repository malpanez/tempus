package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLintSucceedsOnValidICS(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "valid.ics")
	content := `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Tempus//Test//EN
BEGIN:VEVENT
UID:test-1
SUMMARY:Valid event
DTSTART:20250101T100000Z
DTEND:20250101T110000Z
END:VEVENT
END:VCALENDAR
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write ICS: %v", err)
	}

	cmd := newLintCmd()
	mustSetFlag(t, cmd, "file", path)
	if err := runLint(cmd, nil); err != nil {
		t.Fatalf("expected lint to pass, got error: %v", err)
	}
}

func TestLintFailsWhenRequiredFieldsMissing(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "invalid.ics")
	content := `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Tempus//Test//EN
BEGIN:VEVENT
UID:test-2
SUMMARY:Missing start
END:VEVENT
END:VCALENDAR
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write ICS: %v", err)
	}

	cmd := newLintCmd()
	mustSetFlag(t, cmd, "file", path)
	err := runLint(cmd, nil)
	if err == nil {
		t.Fatal("expected lint error for missing DTSTART, got nil")
	}
}
