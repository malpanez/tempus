package cli

import (
	"os"
	"path/filepath"
	"strings"
	"tempus/internal/testutil"
	"testing"
)

func TestUnfoldICSLines(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"simple", "LINE1\nLINE2\nLINE3", 3},
		{"folded line", "LINE1\n CONTINUED", 1},
		{"folded with tab", "LINE1\n\tCONTINUED", 1},
		{"multiple folds", "LINE1\n PART2\n PART3", 1},
		{"crlf", "LINE1\r\nLINE2\r\n", 2},
		{"empty lines", "LINE1\n\nLINE2", 2},
		{"only empty", "\n\n\n", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := unfoldICSLines(tt.input)
			if len(got) != tt.want {
				t.Errorf("unfoldICSLines() returned %d lines, want %d", len(got), tt.want)
			}
		})
	}
}

func TestParseICSProperty(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		wantName  string
		wantValue string
		wantOk    bool
	}{
		{"simple", "SUMMARY:Test Event", "SUMMARY", testutil.EventTitleTestEvent, true},
		{"with params", "DTSTART;TZID=Europe/Madrid:20250501T100000", "DTSTART", "20250501T100000", true},
		{"no colon", "INVALID", "", "", false},
		{"empty key", ":value", "", "", false},
		{"empty value", "KEY:", "KEY", "", true},
		{"lowercase", "summary:Test", "SUMMARY", "Test", true},
		{testutil.TestNameWithSpaces, "  SUMMARY  :  Test  ", "SUMMARY", "Test", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotValue, gotOk := parseICSProperty(tt.line)
			if gotOk != tt.wantOk {
				t.Errorf("parseICSProperty(%q) ok = %v, want %v", tt.line, gotOk, tt.wantOk)
				return
			}
			if gotOk {
				if gotName != tt.wantName {
					t.Errorf("parseICSProperty(%q) name = %q, want %q", tt.line, gotName, tt.wantName)
				}
				if gotValue != tt.wantValue {
					t.Errorf("parseICSProperty(%q) value = %q, want %q", tt.line, gotValue, tt.wantValue)
				}
			}
		})
	}
}

func TestParseICSPropertyWithComplexParams(t *testing.T) {
	line := "ATTENDEE;CN=John Doe;ROLE=REQ-PARTICIPANT:mailto:john@example.com"
	name, value, ok := parseICSProperty(line)

	if !ok {
		t.Fatal("parseICSProperty() returned ok = false")
	}
	if name != "ATTENDEE" {
		t.Errorf("name = %q, want %q", name, "ATTENDEE")
	}
	if value != "mailto:john@example.com" {
		t.Errorf("value = %q, want %q", value, "mailto:john@example.com")
	}
}

func TestUnfoldICSLinesWithRealExample(t *testing.T) {
	input := `DESCRIPTION:This is a very long description that spans multiple lines be
 cause it exceeds 75 characters according to RFC 5545 folding rules. It sh
 ould be unfolded into a single line.
SUMMARY:Test Event`

	lines := unfoldICSLines(input)
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}

	desc := lines[0]
	if !strings.Contains(desc, "DESCRIPTION:") {
		t.Error("first line should be DESCRIPTION")
	}
	if strings.Contains(desc, "\n") {
		t.Error("unfolded line should not contain newlines")
	}
}

func TestLintICSFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name: "valid",
			content: `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VEVENT
UID:test@example.com
DTSTAMP:20250401T090000Z
SUMMARY:Test Event
DTSTART:20250501T100000Z
DTEND:20250501T110000Z
END:VEVENT
END:VCALENDAR`,
			wantErr: false,
		},
		{
			name:    testutil.TestNameEmptyFile,
			content: "",
			wantErr: true,
		},
		{
			name: "missing BEGIN:VCALENDAR",
			content: `VERSION:2.0
BEGIN:VEVENT
END:VEVENT`,
			wantErr: true,
		},
		{
			name: "missing VEVENT",
			content: `BEGIN:VCALENDAR
VERSION:2.0
END:VCALENDAR`,
			wantErr: true,
		},
		{
			name: "missing UID",
			content: `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VEVENT
DTSTAMP:20250401T090000Z
SUMMARY:Test
DTSTART:20250501T100000Z
DTEND:20250501T110000Z
END:VEVENT
END:VCALENDAR`,
			wantErr: true,
		},
		{
			name: "missing SUMMARY is valid per RFC 5545",
			content: `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VEVENT
UID:test@example.com
DTSTAMP:20250401T090000Z
DTSTART:20250501T100000Z
DTEND:20250501T110000Z
END:VEVENT
END:VCALENDAR`,
			wantErr: false,
		},
		{
			name: "missing DTSTAMP",
			content: `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VEVENT
UID:test@example.com
SUMMARY:Test
DTSTART:20250501T100000Z
DTEND:20250501T110000Z
END:VEVENT
END:VCALENDAR`,
			wantErr: true,
		},
		{
			name: "unclosed VEVENT",
			content: `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VEVENT
UID:test@example.com
DTSTAMP:20250401T090000Z
DTSTART:20250501T100000Z
END:VCALENDAR`,
			wantErr: true,
		},
		{
			name: "UNTIL date with timed DTSTART",
			content: `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VEVENT
UID:test@example.com
DTSTAMP:20250401T090000Z
DTSTART:20250501T100000Z
RRULE:FREQ=WEEKLY;UNTIL=20251231
END:VEVENT
END:VCALENDAR`,
			wantErr: true,
		},
		{
			name: "unescaped semicolon in CATEGORIES",
			content: `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VEVENT
UID:test@example.com
DTSTAMP:20250401T090000Z
DTSTART:20250501T100000Z
CATEGORIES:work;personal
END:VEVENT
END:VCALENDAR`,
			wantErr: true,
		},
		{
			name: "missing DTSTART",
			content: `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VEVENT
UID:test@example.com
DTSTAMP:20250401T090000Z
SUMMARY:Test
DTEND:20250501T110000Z
END:VEVENT
END:VCALENDAR`,
			wantErr: true,
		},
		{
			name: "has DURATION instead of DTEND",
			content: `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VEVENT
UID:test@example.com
DTSTAMP:20250401T090000Z
SUMMARY:Test
DTSTART:20250501T100000Z
DURATION:PT1H
END:VEVENT
END:VCALENDAR`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.ics")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatalf(testutil.ErrMsgFailedToWriteTestFile, err)
			}

			_, err := lintICSFile(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("lintICSFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	t.Run("directory instead of file", func(t *testing.T) {
		tmpDir := t.TempDir()
		_, err := lintICSFile(tmpDir)
		if err == nil {
			t.Error("lintICSFile() expected error for directory, got nil")
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		_, err := lintICSFile("/non/existent/file.ics")
		if err == nil {
			t.Error("lintICSFile() expected error for non-existent file, got nil")
		}
	})
}

func TestLintSucceedsOnValidICS(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "valid.ics")
	content := `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Tempus//Test//EN
BEGIN:VEVENT
UID:test-1
DTSTAMP:20250101T090000Z
SUMMARY:Valid event
DTSTART:20250101T100000Z
DTEND:20250101T110000Z
END:VEVENT
END:VCALENDAR
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write ICS: %v", err)
	}

	app := TestApp()
	cmd := NewLintCmd(app)
	mustSetFlag(t, cmd, "file", path)
	if err := runLint(app, cmd, nil); err != nil {
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

	app := TestApp()
	cmd := NewLintCmd(app)
	mustSetFlag(t, cmd, "file", path)
	err := runLint(app, cmd, nil)
	if err == nil {
		t.Fatal("expected lint error for missing DTSTART, got nil")
	}
}
